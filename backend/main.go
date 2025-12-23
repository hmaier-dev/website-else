package main

import (
	"context"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hmaier-dev/website-else/contact-formular/mailbox"
	"github.com/jedib0t/go-pretty/v6/table"

	// For future me:
	// Use 'go get -u github.com/mattn/go-sqlite3' to download it
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

type Server struct {
	Router  *mux.Router
	Mailbox *mailbox.Queries
	API 		*API
}

type API struct {
	Username string
	Password string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	const port = "3000"
	host := fmt.Sprintf("0.0.0.0:%s", port)
	srv := NewServer()
	log.Printf("Starting tool on %s \n", host)
	err := http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}

func NewServer() *Server {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		log.Fatal(err)
	}

	u := os.Getenv("API_USERNAME")
	p := os.Getenv("API_PASSWORD")
	s := &Server{
		Router:  mux.NewRouter(),
		Mailbox: mailbox.New(db),
		API: &API{Username: u, Password: p,},
	}
	mailbox := s.Router.PathPrefix("/mailbox").Subrouter()
	mailbox.HandleFunc("/contact", s.HandlerContactRequest).Methods("POST")
	if u != "" && p != ""{
		mailbox.HandleFunc("/all", s.HandlerAllMessages).Methods("GET")
		mailbox.HandleFunc("/unread", s.HandlerUnreadMessages).Methods("GET")
	}else{
		log.Println("Didn't start the api because username/password weren't set.")
	}


	// logs all routes when starting after they go defined
	s.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		method, _ := route.GetMethods()
		log.Println(method, path)
		return nil
	})

	return s
}

// Incoming POST request from the html formular
// TODO: Implement rate limiting at traefik
func (s *Server) HandlerContactRequest(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	approval := r.FormValue("approval")
	log.Println(name, email, message, approval)
	// TODO: Sanatizing and validating the input
	ctx := context.Background()
	if approval == "on" {
		err := s.Mailbox.AddMessage(ctx, mailbox.AddMessageParams{
			Name:     name,
			Email:    email,
			Message:  message,
			Date:     time.Now().Unix(),
			Approval: sql.NullInt64{Int64: 1, Valid: true},
			Isread:   sql.NullInt64{Int64: 0, Valid: false},
		})
		if err != nil {
			http.Error(w, "Something wen't wrong storing the data...", http.StatusInternalServerError)
			log.Fatalf("Err: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "We need approval to store user data", http.StatusForbidden)
}

// Serving GET request to return all messages
func (s *Server) HandlerAllMessages(w http.ResponseWriter, r *http.Request) {

	// Refactor this into one function
	// Here is how to do it
	// https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
	u, p, ok := r.BasicAuth()
	realm := "Einmal bitte Credentials eingeben: "
	if !ok || subtle.ConstantTimeCompare([]byte(u), []byte(s.API.Username)) != 1 || subtle.ConstantTimeCompare([]byte(p), []byte(s.API.Password)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
		return
	}

	ctx := context.Background()
	all, err := s.Mailbox.GetAllMessages(ctx)
	if err != nil {
		http.Error(w, "Something wen't wrong querying all messages...", http.StatusInternalServerError)
	}
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{"ID", "Name", "Email", "Message", "Date", "Approval", "IsRead"})
	for _, a := range all {
		d := time.Unix(a.Date, 0).Format("02.01.2006 15:04:05")
		t.AppendRow(table.Row{a.ID, a.Name, a.Email, a.Message, d, a.Approval.Int64, a.Isread.Int64})
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}


func (s *Server) HandlerUnreadMessages(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	unread, err := s.Mailbox.GetUnreadMessage(ctx)

	// Refactor this into one function
	// Here is how to do it
	// https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
	u, p, ok := r.BasicAuth()
	realm := "Einmal bitte Credentials eingeben: "
	if !ok || subtle.ConstantTimeCompare([]byte(u), []byte(s.API.Username)) != 1 || subtle.ConstantTimeCompare([]byte(p), []byte(s.API.Password)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
		return
	}

	if err != nil {
		http.Error(w, "Something wen't wrong querying unread messages...", http.StatusInternalServerError)
	}
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{"ID", "Name", "Email", "Message", "Date", "Approval", "IsRead"})
	for _, u := range unread {
		d := time.Unix(u.Date, 0).Format("02.01.2006 15:04:05")
		t.AppendRow(table.Row{u.ID, u.Name, u.Email, u.Message, d, u.Approval.Int64, u.Isread.Int64})
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}

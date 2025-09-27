package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	// For future me:
	// Use 'go get -u github.com/mattn/go-sqlite3' to download it
	"github.com/hmaier-dev/website-else/contact-formular/mailbox"
	"github.com/jedib0t/go-pretty/v6/table"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

type Server struct {
	Router  *mux.Router
	Mailbox *mailbox.Queries
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
	s := &Server{
		Router:  mux.NewRouter(),
		Mailbox: mailbox.New(db),
	}
	s.Router.HandleFunc("/contact", s.HandlerContactRequest).Methods("POST")
	mailbox := s.Router.PathPrefix("/mailbox").Subrouter()
	mailbox.HandleFunc("/all", s.HandlerAllMessages).Methods("GET")
	mailbox.HandleFunc("/unread", s.HandlerUnreadMessages).Methods("GET")

	return s
}

// Incoming POST request from the html formular
func (s *Server) HandlerContactRequest(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	approval := r.FormValue("approval")
	log.Println(name, email, message, approval)
	ctx := context.Background()
	if approval != "false" {
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
		}
		w.WriteHeader(http.StatusOK)
	}
	http.Error(w, "We need approval to store user data", http.StatusForbidden)
}

// Serving GET request to return all messages
func (s *Server) HandlerAllMessages(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	all, err := s.Mailbox.GetAllMessages(ctx)
	if err != nil {
		http.Error(w, "Something wen't wrong querying all messages...", http.StatusInternalServerError)
	}
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{"ID","Name","Email","Message","Date","Approval","IsRead"})
	for _, a := range all {
		d := time.Unix(a.Date, 0).Format("02.01.2006 15:04:05")
		t.AppendRow(table.Row{a.ID,a.Name,a.Email,a.Message,d,a.Approval.Int64,a.Isread.Int64})
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}
func (s *Server) HandlerUnreadMessages(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	unread, err := s.Mailbox.GetUnreadMessage(ctx)

	if err != nil {
		http.Error(w, "Something wen't wrong querying unread messages...", http.StatusInternalServerError)
	}
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{"ID","Name","Email","Message","Date","Approval","IsRead"})
	for _, u := range unread {
		d := time.Unix(u.Date, 0).Format("02.01.2006 15:04:05")
		t.AppendRow(table.Row{u.ID,u.Name,u.Email,u.Message,d,u.Approval.Int64,u.Isread.Int64})
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}

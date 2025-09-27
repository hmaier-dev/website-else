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
	_ "github.com/mattn/go-sqlite3"
	"github.com/hmaier-dev/website-else/contact-formular/mailbox"
)

//go:embed schema.sql
var ddl string

type Server struct {
	Router *mux.Router
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
		Router: mux.NewRouter(),
		Mailbox: mailbox.New(db),
	}
  s.Router.HandleFunc("/contact", s.HandlerContactRequest).Methods("POST")
	mailbox := s.Router.PathPrefix("/mailbox").Subrouter()
	mailbox.HandleFunc("/all", s.HandlerAllMessages).Methods("GET")

	return s
}

func (s *Server) HandlerContactRequest(w http.ResponseWriter, r *http.Request){
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	approval := r.FormValue("approval")
	log.Println(name, email, message, approval)
	ctx  := context.Background()
	if approval != "false"{
		err := s.Mailbox.AddMessage(ctx,mailbox.AddMessageParams{
			Name: name,	
			Email: email,
			Message: message,
			Date: time.Now().Unix(),
			Approval: sql.NullInt64{Int64: 1, Valid: true,},
			Isread: sql.NullInt64{Int64: 0, Valid: false},
		})
		if err != nil{
			http.Error(w, "Something wen't wrong storing the data.", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
	http.Error(w,"We need approval to store user data", http.StatusForbidden)
}
func (s *Server) HandlerAllMessages(w http.ResponseWriter, r *http.Request){
	ctx  := context.Background()
	all, err := s.Mailbox.GetAllMessages(ctx)
	if err != nil{
		http.Error(w, "Something wen't wrong querying all messages...", http.StatusInternalServerError)
	}
	for _, msg := range all{
		log.Printf("%#v \n", msg)
	}
}

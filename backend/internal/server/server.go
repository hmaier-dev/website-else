package server

import (
	"database/sql"
	"embed"
	_ "embed"
	"log"
	"os"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/events"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/mailbox"

	"github.com/gorilla/mux"

	// For future me:
	// Use 'go get -u github.com/mattn/go-sqlite3' to download it
	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	Router  *mux.Router
}


func New(sqlite *sql.DB, internalFS embed.FS) *Server {
	u := os.Getenv("API_USERNAME")
	p := os.Getenv("API_PASSWORD")

	s := &Server{
		Router:  mux.NewRouter(),
	}

	//----------------------------------------------------------
	// Mailbox endpoints 
	m := &mailbox.Handler{
		DB: sqlite,
		API: mailbox.API{Username: u,Password: p},
	}
	mbr := s.Router.PathPrefix("/mailbox").Subrouter()
	mbr.HandleFunc("/contact", m.HandlerContactRequest).Methods("POST")
	if u != "" && p != ""{
		mbr.HandleFunc("/all", m.HandlerAllMessages).Methods("GET")
		mbr.HandleFunc("/unread", m.HandlerUnreadMessages).Methods("GET")
	}else{
		log.Println("Didn't start the api because username/password weren't set.")
	}
	//----------------------------------------------------------
	// Events endpoints 

	e := &events.Handler{
		DB: sqlite,
		InternalFS: internalFS,
	}
	
	// This stuff is public
	events := s.Router.PathPrefix("/events/").Subrouter()
	events.HandleFunc("/load", e.Load).Methods("GET")
	
	// and this stuff is like /events/manage behind an auth
	manage := events.PathPrefix("/manage/api").Subrouter()
	manage.HandleFunc("/load/mutable", e.LoadMutable).Methods("GET")
	manage.HandleFunc("/create", e.Create).Methods("POST")
	manage.HandleFunc(`/delete/{id:\d*}`, e.Delete).Methods("POST")
	manage.HandleFunc("/create/bulk", e.Bulk).Methods("POST")
	manage.HandleFunc(`/change/{id:\d*}`, e.Change).Methods("POST")


	// logs all routes when starting after they go defined
	s.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		method, _ := route.GetMethods()
		log.Println(method, path)
		return nil
	})

	return s
}

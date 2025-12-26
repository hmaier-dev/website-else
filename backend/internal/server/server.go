package server

import (
	_ "embed"
	"log"
	"os"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/database"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/mailbox"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/events"

	"github.com/gorilla/mux"

	// For future me:
	// Use 'go get -u github.com/mattn/go-sqlite3' to download it
	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	Router  *mux.Router
}


func New(db *database.Queries) *Server {
	u := os.Getenv("API_USERNAME")
	p := os.Getenv("API_PASSWORD")

	s := &Server{
		Router:  mux.NewRouter(),
	}

	//----------------------------------------------------------
	// Mailbox endpoints 
	m := &mailbox.Handler{
		Queries: db,
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
	// Mailbox endpoints 

	e := &events.Handler{
		Queries: db,
	}
	er := s.Router.PathPrefix("/events/change").Subrouter()

	// logs all routes when starting after they go defined
	s.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		method, _ := route.GetMethods()
		log.Println(method, path)
		return nil
	})

	return s
}

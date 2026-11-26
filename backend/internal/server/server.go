package server

import (
	"database/sql"
	"embed"
	_ "embed"
	"log"
	"os"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/auth"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/events"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/mailbox"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/shop"

	"github.com/gorilla/mux"

	// For future me:
	// Use 'go get -u github.com/mattn/go-sqlite3' to download it
	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	Router *mux.Router
}

func New(sqlite *sql.DB, internalFS embed.FS) *Server {
	u := os.Getenv("API_USERNAME")
	p := os.Getenv("API_PASSWORD")

	s := &Server{
		Router: mux.NewRouter(),
	}

	//----------------------------------------------------------
	// Mailbox endpoints
	m := &mailbox.Handler{
		DB:  sqlite,
		API: mailbox.API{Username: u, Password: p},
	}
	mbr := s.Router.PathPrefix("/mailbox").Subrouter()
	mbr.HandleFunc("/contact", m.HandlerContactRequest).Methods("POST")
	if u != "" && p != "" {
		mbr.HandleFunc("/all", m.HandlerAllMessages).Methods("GET")
		mbr.HandleFunc("/unread", m.HandlerUnreadMessages).Methods("GET")
	} else {
		log.Println("Didn't start the api because username/password weren't set.")
	}
	//----------------------------------------------------------
	// Events endpoints

	e := &events.Handler{
		DB:         sqlite,
		InternalFS: internalFS,
	}

	// This stuff is public
	events := s.Router.PathPrefix("/events/").Subrouter()
	events.HandleFunc("/load", e.PublicLoad).Methods("GET")
	events.HandleFunc(`/cal/ics/{id:\d+:}`, e.ProvideICSFile).Methods("GET")

	//----------------------------------------------------------
	// Shop endpoints (öffentlich)
	sh := shop.FromEnv(sqlite, os.Getenv)
	shopAPI := s.Router.PathPrefix("/shop/api").Subrouter()
	shopAPI.HandleFunc("/config", sh.Config).Methods("GET")
	shopAPI.HandleFunc("/orders", sh.CreateOrder).Methods("POST")
	shopAPI.HandleFunc("/orders/{id:[A-Za-z0-9-]+}/capture", sh.CaptureOrder).Methods("POST")

	//----------------------------------------------------------
	// Verwaltung: alles unter /manage/api.
	//
	// Traefik schützt /manage (statische Seite aus dem website-Container) und
	// /manage/api (dieses Backend) mit Authentik/ForwardAuth. Zusätzlich lässt
	// das Backend nur Anfragen durch, bei denen Authentik den Benutzer-Header setzt.
	if h, ok := os.LookupEnv("MANAGE_AUTH_HEADER"); ok {
		auth.Header = h
	}
	if auth.Header == "" {
		log.Println("WARNUNG: MANAGE_AUTH_HEADER ist leer – /manage/api ist ohne Anmeldung erreichbar. Nur für lokale Entwicklung!")
	}
	manage := s.Router.PathPrefix("/manage/api").Subrouter()
	manage.Use(auth.Require)

	manage.HandleFunc("/events", e.AdminPanel).Methods("GET")
	manage.HandleFunc("/events", e.AdminCreate).Methods("POST")
	manage.HandleFunc("/events/bulk", e.BulkCreate).Methods("POST")
	manage.HandleFunc(`/events/{id:\d+}`, e.AdminRow).Methods("GET")
	manage.HandleFunc(`/events/{id:\d+}`, e.AdminUpdate).Methods("PUT")
	manage.HandleFunc(`/events/{id:\d+}`, e.AdminDelete).Methods("DELETE")
	manage.HandleFunc(`/events/{id:\d+}/edit`, e.AdminEdit).Methods("GET")

	manage.HandleFunc("/orders", sh.AdminOrders).Methods("GET")
	manage.HandleFunc(`/orders/{id:\d+}/shipped`, sh.AdminMarkShipped).Methods("POST")

	manage.HandleFunc("/messages", m.AdminMessages).Methods("GET")
	manage.HandleFunc(`/messages/{id:\d+}/read`, m.AdminMarkRead).Methods("POST")

	// logs all routes when starting after they go defined
	s.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		method, _ := route.GetMethods()
		log.Println(method, path)
		return nil
	})

	return s
}

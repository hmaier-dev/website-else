package events

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"html/template"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

type Handler struct{
	DB *sql.DB
}


// Renders all events in html.
// TODO: set ?sort={asc|desc|random} to sort the events by date or randomize.
// TODO: set ?head={nth} just load the nth fetched elems. This combines with ?sort.
// TODO: set ?highlight={true|false} to highlight the nearest future event.
// TODO: set ?raw={true|false} when set to true, events are returned as csv (maybe I could return this as vcard, so the people can import it directly into their calendar....)
func (h *Handler) Load (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	sortQ := query.Get("sort")
	headQ := query.Get("head")
	highlightQ := query.Get("highlight")

	q := db.New(h.DB)
	// Default values
	var defaultHead int64
	defaultHead, err := q.CountEvents(ctx) // To get all events by default, we need the exact number of these
	if err != nil{
		log.Println("Could count events...")
	}
	sortBy := "desc" // default sort order
	highlight := false
	raw := false

	if sortQ == "asc" || sortQ == "random" {
		sortBy = sortQ
	}
	
	// With head we decided how many entry we are gonna load
	// if head is not set in the query, all events are loaded.
	var head int64
	if headQ != "" {
		headInt, err := strconv.Atoi(headQ)
		if err == nil && headInt > 0 {
			head = int64(headInt)
		}
	}else{
		head = defaultHead	
	}

	var events = make([]db.Event, head)
	switch sortBy{
		case "asc":
			events, err = q.LoadEventsAsc(ctx, head)
			if err != nil{
				log.Println("help help help")
			}
		case "desc":
			events, err = q.LoadEventsDesc(ctx, head)
			if err != nil{
				log.Println("help help help")
			}
		case "random":
			events, err = q.LoadEventsRandom(ctx, head)
			if err != nil{
				log.Println("help help help")
			}
	}

	if highlightQ == "true" {
		highlight = true
	}

	tmpl, err := template.ParseFiles("./templates/events-list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	return
}

func (h *Handler) Change(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	return
}

func (h *Handler) RemoveEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	return
}

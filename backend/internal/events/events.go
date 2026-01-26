package events

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

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
	log.Println("events were requested...")
	ctx := r.Context()
	query := r.URL.Query()
	sortQ := query.Get("sort")
	headQ := query.Get("head")
	// highlightQ := query.Get("highlight")

	// if highlightQ == "true" {
	// 	highlight = true
	// }

	q := db.New(h.DB)
	// Default values
	var defaultHead int64
	defaultHead, err := q.CountEvents(ctx) // To get all events by default, we need the exact number of these
	if err != nil{
		log.Println("Could count events...")
	}
	sortBy := "desc" // default sort order
	// highlight := false
	// raw := false

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


	tmpl, err := template.ParseFiles("internal/events/templates/events-list.html")
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
	q := db.New(h.DB)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract form values
	datum := r.FormValue("datum")
	event := r.FormValue("event")
	ort := r.FormValue("ort")
	stadt := r.FormValue("stadt")
	link := r.FormValue("link")

	// Parse the date
	date, err := time.Parse("2006-01-02", datum)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	doorsTime := sql.NullTime{Time: date, Valid: true}
	beginTime := sql.NullTime{Time: date, Valid: true}
	linkString := sql.NullString{String: link, Valid: link != ""}

	err = q.AddSingleEvent(ctx, db.AddSingleEventParams{
		Name:  event,
		Date:  date,
		Doors: doorsTime,
		Begin: beginTime,
		Place: ort,
		City:  stadt,
		Link:  linkString,
		Pic:   nil, // Assuming no picture is uploaded
	})
	if err != nil {
		log.Printf("Failed to add event: %v", err)
		http.Error(w, "Failed to add event", http.StatusInternalServerError)
		return
	}
	// Return a success response
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Event added successfully"))
}

func (h *Handler) Bulk(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the YAML file from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	type Event struct {
		Datum string `yaml:"datum"`
		Event string `yaml:"event"`
		Doors string `yaml:"doors"`
		Begin string `yaml:"begin"`
		Ort   string `yaml:"ort"`
		Stadt string `yaml:"stadt"`
		Link  string `yaml:"link"`
	}

	var yamlEvents []Event
	err = yaml.Unmarshal(body, &yamlEvents)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse YAML: \n%v", err.Error())
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Insert each event into the database
	for _, e := range yamlEvents {
		date, err := time.Parse("02.01.2006", e.Datum)
		if err != nil {
			log.Printf("Invalid date format for event %s: %v", e.Event, err)
			continue
		}
		doorsTime := sql.NullTime{Time: date, Valid: true}
		beginTime := sql.NullTime{Time: date, Valid: true}
		linkString := sql.NullString{String: e.Link, Valid: e.Link != ""}

		err = q.AddSingleEvent(ctx, db.AddSingleEventParams{
			Name:  e.Event,
			Date:  date,
			Doors: doorsTime,
			Begin: beginTime,
			Place: e.Ort,
			City:  e.Stadt,
			Link:  linkString,
			Pic:   nil,
		})
		if err != nil {
			log.Printf("Failed to add event %s: %v", e.Event, err)
			continue
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Events added successfully"))
}

func (h *Handler) Change(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	return
}

func (h *Handler) RemoveEvent(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	return
}

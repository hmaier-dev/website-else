package events

import (
	"context"
	"database/sql"
	"embed"
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

type Handler struct {
	DB         *sql.DB
	InternalFS embed.FS
}

func formatTime(s time.Time) template.HTML {
	return template.HTML(s.Format("02.01.2006"))
}

type Event struct {
	Name  string
	Date  time.Time
	Doors time.Time
	Begin time.Time
	Place string
	City  string
	Link  *string
}

// This interface is needed for accessing the different Event*-structs
type EventRow interface {
    GetName() string
    GetDate() time.Time
    GetDoors() time.Time
    GetBegin() time.Time
    GetPlace() string
    GetCity() string
    GetLink() *string
}

type EventAscRow struct{
	r db.LoadEventsAscRow
}

func (a EventAscRow ) GetName() string {return a.r.Name}
func (a EventAscRow ) GetDate() time.Time {return a.r.Date}
func (a EventAscRow ) GetDoors() time.Time {return a.r.Doors.Time}
func (a EventAscRow ) GetBegin() time.Time {return a.r.Begin.Time}
func (a EventAscRow ) GetPlace() string {return a.r.Place}
func (a EventAscRow ) GetCity() string {return a.r.City}
func (a EventAscRow ) GetLink() *string {
    if !a.r.Link.Valid || a.r.Link.String == "" {
        return nil
    }
    return &a.r.Link.String
}
// This maps the row to the struct which has methods to access it
func adaptAscRow(rows []db.LoadEventsAscRow) []EventRow{
	out := make([]EventRow, len(rows))
	for i, r := range rows{
		out[i] = EventAscRow{r: r}
	}
	return out
}

type EventDescRow struct{
	r db.LoadEventsDescRow
}
func (d EventDescRow ) GetName() string {return d.r.Name}
func (d EventDescRow ) GetDate() time.Time {return d.r.Date}
func (d EventDescRow ) GetDoors() time.Time {return d.r.Doors.Time}
func (d EventDescRow ) GetBegin() time.Time {return d.r.Begin.Time}
func (d EventDescRow ) GetPlace() string {return d.r.Place}
func (d EventDescRow ) GetCity() string {return d.r.City}
func (d EventDescRow ) GetLink() *string {
    if !d.r.Link.Valid || d.r.Link.String == "" {
        return nil
    }
    return &d.r.Link.String
}

func adaptDescRow(rows []db.LoadEventsDescRow) []EventRow{
	out := make([]EventRow, len(rows))
	for i, r := range rows{
		out[i] = EventDescRow{r: r}
	}
	return out
}

type EventRandRow struct{
	r db.LoadEventsRandomRow
}
func (rand EventRandRow ) GetName() string {return rand.r.Name}
func (rand EventRandRow ) GetDate() time.Time {return rand.r.Date}
func (rand EventRandRow ) GetDoors() time.Time {return rand.r.Doors.Time}
func (rand EventRandRow ) GetBegin() time.Time {return rand.r.Begin.Time}
func (rand EventRandRow ) GetPlace() string {return rand.r.Place}
func (rand EventRandRow ) GetCity() string {return rand.r.City}
func (rand EventRandRow ) GetLink() *string {
    if !rand.r.Link.Valid || rand.r.Link.String == "" {
        return nil
    }
    return &rand.r.Link.String
}

func adaptRandRow(rows []db.LoadEventsRandomRow) []EventRow{
	out := make([]EventRow, len(rows))
	for i, r := range rows{
		out[i] = EventRandRow{r: r}
	}
	return out
}
// ##########################################################################################################
// A new query with a new db.Struct is generated, you need to add the same boilerplate from above here
// ###########################################################################################################

// In the future I may gonna need to add custom-values
// This way
func constructEvents(rows []EventRow) []Event {
	var events = make([]Event, len(rows))	
	for i, r := range rows {
		events[i] = Event{
			Name:  r.GetName(),
			Date:  r.GetDate(),
			Doors: r.GetDoors(),
			Begin: r.GetBegin(),
			Place: r.GetPlace(),
			City:  r.GetCity(),
			Link:  r.GetLink(),
		}
		log.Printf("%v \n", events[i])
	}
	return events
}

type urlArgs struct {
	Head string
	Sort string
	Out  string
}

// Returns []db.Event
// set ?sort={asc|desc|random|next|prev} to sort the events by date or randomize.
// set ?head={nth} just load the nth fetched elems. This combines with ?sort.
func (h *Handler) eventsByArgs(ctx context.Context, args urlArgs) []Event {
	q := db.New(h.DB)

	// By default all events get rendered
	var defaultHead int64
	defaultHead, err := q.CountEvents(ctx) // To get all events by default, we need the exact number of these
	if err != nil {
		log.Println("Could not count events...")
	}

	// With head we decided how many entry we are gonna load
	// if head is not set in the query, all events are loaded.
	var head int64
	if args.Head != "" {
		headInt, err := strconv.Atoi(args.Head)
		if err == nil && headInt > 0 {
			head = int64(headInt)
		}
	} else {
		head = defaultHead
	}

	sortBy := "desc" // default sort order
	if args.Sort == "asc" || args.Sort == "random" {
		sortBy = args.Sort
	}

	// Because 'sqlc generate' does a struct per query, we need to reformat from e.g. db.LoadEventsAscRow into db.Events
	// event though they hold the same data
	var events = make([]db.Event, head)
	// Generate this very important number to make API-Calls very senseful...
	var funID int64
	switch sortBy {
	case "asc":
		rows, err := q.LoadEventsAsc(ctx, head)
		if err != nil {
			log.Println("help help help")
		}
		for i, r := range rows {
			funID = rand.Int63()
			events[i] = db.Event{
				ID:    funID,
				Name:  r.Name,
				Date:  r.Date,
				Doors: r.Doors,
				Begin: r.Begin,
				Place: r.Place,
				City:  r.City,
				Link:  r.Link,
			}
		}
	case "desc":
		rows, err := q.LoadEventsDesc(ctx, head)
		if err != nil {
			log.Println("help help help")
		}
		for i, r := range rows {
			funID = rand.Int63()
			events[i] = db.Event{
				ID:    funID,
				Name:  r.Name,
				Date:  r.Date,
				Doors: r.Doors,
				Begin: r.Begin,
				Place: r.Place,
				City:  r.City,
				Link:  r.Link,
			}
		}
	case "random":
		rows, err := q.LoadEventsRandom(ctx, head)
		if err != nil {
			log.Println("help help help")
		}
		for i, r := range rows {
			funID = rand.Int63()
			events[i] = db.Event{
				ID:    funID,
				Name:  r.Name,
				Date:  r.Date,
				Doors: r.Doors,
				Begin: r.Begin,
				Place: r.Place,
				City:  r.City,
				Link:  r.Link,
			}
		}
	}
	return events	
}

// TODO: set ?highlight={true|false} to highlight the nearest future event.
// TODO: set ?render={html|json}
func (h *Handler) Load(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()
	args := urlArgs{
		Sort: query.Get("sort"),
		Head: query.Get("head"),
		Out:  query.Get("out"),
	}
	events := h.eventsByArgs(ctx, args)
	
	outOpt := "json"	
	if args.Out != ""{
		outOpt = args.Out
	}
	var err error
	switch outOpt{
		case "json":
			bytes, err := json.Marshal(events)
			if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(bytes)
			return
		case "html":
			// this can be done prettier => see checklist-tool
			tmpl  := template.Must(template.New("events-range.html").Funcs(template.FuncMap{
				"safe":           safe,
				"formatTime":     formatTime,
				"formatNulltime": formatNullTime,
			}).ParseFS(h.InternalFS, "internal/events/templates/events-range.html"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tmpl.Execute(w, map[string]any{
				"Events": events,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		default:
			w.Write([]byte("no valid out="))
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
	date, err := time.Parse("02.01.2006", datum)
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

// Load /events/manage/api/load/mutable
func (h *Handler) LoadMutable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	args := urlArgs{
		Sort: query.Get("sort"),
		Head: query.Get("head"),
		Out:  "html",
	}
	events := h.eventsByArgs(ctx, args)
	// this can be done prettier => see checklist-tool
	tmpl  := template.Must(template.New("events-mutable.html").Funcs(template.FuncMap{
		"safe":           safe,
		"formatTime":     formatTime,
		"formatNulltime": formatNullTime,
	}).ParseFS(h.InternalFS, "internal/events/templates/events-mutable.html"))
	err := tmpl.Execute(w, map[string]any{
		"Events": events,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Change(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	return
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	return
}

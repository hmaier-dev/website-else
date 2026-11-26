package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"
	"sort"

	"github.com/btcsuite/btcutil/base58"

	ics "github.com/arran4/golang-ical"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

type Handler struct {
	DB         *sql.DB
	InternalFS fs.FS // in Produktion das eingebettete embed.FS aus main.go
}

func formatTime(s time.Time) template.HTML {
	return template.HTML(s.Format("02.01.2006"))
}

// This interface is needed for accessing the different Event*-structs.
// sqlc generates per query a different struct, which got the same field as the other ones.
// The following interface is made to serialize the data
type EventRow interface {
	GetName() string
	GetDate() time.Time
	GetDoors() time.Time
	GetBegin() time.Time
	GetPlace() string
	GetCity() string
	GetCoordinates() string
	GetLink() *string
	GetHash() string
}

type EventAscRow struct {
	r db.LoadEventsAscRow
}

func (a EventAscRow) GetName() string        { return a.r.Name }
func (a EventAscRow) GetDate() time.Time     { return a.r.Date }
func (a EventAscRow) GetDoors() time.Time    { return a.r.Doors.Time }
func (a EventAscRow) GetBegin() time.Time    { return a.r.Begin.Time }
func (a EventAscRow) GetPlace() string       { return a.r.Place }
func (a EventAscRow) GetCity() string        { return a.r.City }
func (a EventAscRow) GetCoordinates() string { return a.r.Coordinates.String }
func (a EventAscRow) GetLink() *string {
	if !a.r.Link.Valid || a.r.Link.String == "" {
		return nil
	}
	return &a.r.Link.String
}
func (a EventAscRow) GetHash() string { return a.r.Hash }

// This maps the row to the struct which has methods to access it
func adaptAscRow(rows []db.LoadEventsAscRow) []EventRow {
	out := make([]EventRow, len(rows))
	for i, r := range rows {
		out[i] = EventAscRow{r: r}
	}
	return out
}

type EventDescRow struct {
	r db.LoadEventsDescRow
}

func (d EventDescRow) GetName() string        { return d.r.Name }
func (d EventDescRow) GetDate() time.Time     { return d.r.Date }
func (d EventDescRow) GetDoors() time.Time    { return d.r.Doors.Time }
func (d EventDescRow) GetBegin() time.Time    { return d.r.Begin.Time }
func (d EventDescRow) GetPlace() string       { return d.r.Place }
func (d EventDescRow) GetCoordinates() string { return d.r.Coordinates.String }
func (d EventDescRow) GetCity() string        { return d.r.City }
func (d EventDescRow) GetLink() *string {
	if !d.r.Link.Valid || d.r.Link.String == "" {
		return nil
	}
	return &d.r.Link.String
}
func (d EventDescRow) GetHash() string { return d.r.Hash }

func adaptDescRow(rows []db.LoadEventsDescRow) []EventRow {
	out := make([]EventRow, len(rows))
	for i, r := range rows {
		out[i] = EventDescRow{r: r}
	}
	return out
}

//#######################################################################################################
// RANDOM

type EventRandRow struct {
	r db.LoadEventsRandomRow
}

func (rand EventRandRow) GetName() string        { return rand.r.Name }
func (rand EventRandRow) GetDate() time.Time     { return rand.r.Date }
func (rand EventRandRow) GetDoors() time.Time    { return rand.r.Doors.Time }
func (rand EventRandRow) GetBegin() time.Time    { return rand.r.Begin.Time }
func (rand EventRandRow) GetPlace() string       { return rand.r.Place }
func (rand EventRandRow) GetCity() string        { return rand.r.City }
func (rand EventRandRow) GetCoordinates() string { return rand.r.Coordinates.String }
func (rand EventRandRow) GetLink() *string {
	if !rand.r.Link.Valid || rand.r.Link.String == "" {
		return nil
	}
	return &rand.r.Link.String
}
func (rand EventRandRow) GetHash() string { return rand.r.Hash }

func adaptRandRow(rows []db.LoadEventsRandomRow) []EventRow {
	out := make([]EventRow, len(rows))
	for i, r := range rows {
		out[i] = EventRandRow{r: r}
	}
	return out
}

//#######################################################################################################
// NEXT

type EventNextRow struct {
	r db.LoadEventsNextRow
}

func (rand EventNextRow) GetName() string        { return rand.r.Name }
func (rand EventNextRow) GetDate() time.Time     { return rand.r.Date }
func (rand EventNextRow) GetDoors() time.Time    { return rand.r.Doors.Time }
func (rand EventNextRow) GetBegin() time.Time    { return rand.r.Begin.Time }
func (rand EventNextRow) GetPlace() string       { return rand.r.Place }
func (rand EventNextRow) GetCity() string        { return rand.r.City }
func (rand EventNextRow) GetCoordinates() string { return rand.r.Coordinates.String }
func (rand EventNextRow) GetLink() *string {
	if !rand.r.Link.Valid || rand.r.Link.String == "" {
		return nil
	}
	return &rand.r.Link.String
}
func (rand EventNextRow) GetHash() string { return rand.r.Hash }

func adaptNextRow(rows []db.LoadEventsNextRow) []EventRow {
	out := make([]EventRow, len(rows))
	for i, r := range rows {
		out[i] = EventNextRow{r: r}
	}
	return out
}

//#######################################################################################################
// PREVIOUS

type EventPrevRow struct {
	r db.LoadEventsPrevRow
}

func (rand EventPrevRow) GetName() string        { return rand.r.Name }
func (rand EventPrevRow) GetDate() time.Time     { return rand.r.Date }
func (rand EventPrevRow) GetDoors() time.Time    { return rand.r.Doors.Time }
func (rand EventPrevRow) GetBegin() time.Time    { return rand.r.Begin.Time }
func (rand EventPrevRow) GetPlace() string       { return rand.r.Place }
func (rand EventPrevRow) GetCity() string        { return rand.r.City }
func (rand EventPrevRow) GetCoordinates() string { return rand.r.Coordinates.String }
func (rand EventPrevRow) GetLink() *string {
	if !rand.r.Link.Valid || rand.r.Link.String == "" {
		return nil
	}
	return &rand.r.Link.String
}
func (rand EventPrevRow) GetHash() string { return rand.r.Hash }

func adaptPrevRow(rows []db.LoadEventsPrevRow) []EventRow {
	out := make([]EventRow, len(rows))
	for i, r := range rows {
		out[i] = EventPrevRow{r: r}
	}
	return out
}

// ############################################################################################################################
// Note to future me: When a new query with a new db.Struct is generated, you need to add the same boilerplate from above here
// ############################################################################################################################

type Event struct {
	Name  string
	Date  time.Time
	FDate string
	// Doors       time.Time
	// Begin       time.Time
	Place string
	City  string
	// Coordinates string
	Link *string
	// Hinweis statt Link (z. B. "Eintritt: Spendenbasis"), siehe links.go
	Note *string `json:",omitempty"`
	// Hash        string
	Separator *string
	Upcomming *bool
}

// formatEventDate gibt z. B. "05.12.2026, 20:00 Uhr" zurück, ohne Uhrzeit nur das Datum.
func formatEventDate(date, begin time.Time) string {
	d := date.Format("02.01.2006")
	if !begin.IsZero() && (begin.Hour() != 0 || begin.Minute() != 0) {
		d += ", " + begin.Format("15:04") + " Uhr"
	}
	return d
}

// Exposes different value to the API
func constructEvents(rows []EventRow) []Event {
	var events = make([]Event, len(rows))
	for i, r := range rows {
		events[i] = Event{
			Name:  r.GetName(),
			Date:  r.GetDate(),
			FDate: formatEventDate(r.GetDate(), r.GetBegin()),
			// Doors:       r.GetDoors(),
			// Begin:       r.GetBegin(),
			Place: r.GetPlace(),
			City:  r.GetCity(),
			// Coordinates: r.GetCoordinates(),
			Link: nil,
			// Hash:        r.GetHash(),
		}
	}
	// Link-Feld aufteilen: echte Links bleiben Link, Text wird Hinweis, "null" verschwindet
	for i, r := range rows {
		if l := r.GetLink(); l != nil {
			link, note := splitLink(*l)
			if link != "" {
				events[i].Link = &link
			}
			if note != "" {
				events[i].Note = &note
			}
		}
	}
	return events
}

// set ?sort={asc|desc|random|next|prev} to sort the events by date or randomize.
// set ?head={nth} just load the nth fetched elems. This combines with ?sort.
// set ?out={json|html} output format, by default json.
// set ?sep={year|month} add separators to hold events apart. e.g. set a marker in html
// set ?highlight={true} put emphasis on the next comming event
type urlArgs struct {
	Head      string
	Sort      string
	Out       string
	Sep       string
	Upcomming string
}

// Handles the different options receives in the API call
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
	if args.Sort == "asc" || args.Sort == "random" || args.Sort == "next" || args.Sort == "prev" {
		sortBy = args.Sort
	}

	// Loading the events in a certain order
	var events = make([]Event, head)
	switch sortBy {
	case "asc":
		rows, err := q.LoadEventsAsc(ctx, head)
		if err != nil {
			log.Println("Problem loading events from database...")
		}
		wrap := adaptAscRow(rows)
		events = constructEvents(wrap)
	case "desc":
		rows, err := q.LoadEventsDesc(ctx, head)
		if err != nil {
			log.Println("Problem loading events from database...")
		}
		wrap := adaptDescRow(rows)
		events = constructEvents(wrap)
	case "random":
		rows, err := q.LoadEventsRandom(ctx, head)
		if err != nil {
			log.Println("Problem loading events from database...")
		}
		wrap := adaptRandRow(rows)
		events = constructEvents(wrap)
	case "next":
		rows, err := q.LoadEventsNext(ctx, head)
		if err != nil {
			log.Println("Problem loading events from database...")
		}
		wrap := adaptNextRow(rows)
		events = constructEvents(wrap)
	case "prev":
		rows, err := q.LoadEventsPrev(ctx, head)
		if err != nil {
			log.Println("Problem loading events from database...")
		}
		wrap := adaptPrevRow(rows)
		events = constructEvents(wrap)
	}

	if args.Sep == "year" && sortBy != "random" {
		for i := range events {
			currYear := events[i].Date.Format("2006")
			// Add a separator at the top of the slice
			if i == 0 {
				events[i].Separator = &currYear
				continue
			}
			// Aren't these both statements are the same?
			// Merge them in one?
			if i != 0 && (sortBy == "desc" || sortBy == "prev") && events[i-1].Date.Format("2006") > currYear {
				events[i].Separator = &currYear
				continue
			}
			if i != 0 && (sortBy == "asc" || sortBy == "next") && events[i-1].Date.Format("2006") < currYear {
				events[i].Separator = &currYear
				continue
			}
		}
	}

	if args.Upcomming == "true" {
		markUpcoming(events, sortBy, time.Now())
	}

	return events

}

// markUpcoming hebt den nächsten anstehenden Termin hervor ("Als nächstes").
// Termine von heute zählen noch als anstehend.
// (Die alte Version griff auf events[i+1] zu und stürzte bei nur einem Termin ab.)
func markUpcoming(events []Event, sortBy string, now time.Time) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	isUpcoming := func(e Event) bool { return !e.Date.Before(today) }
	idx := -1
	switch sortBy {
	case "next":
		if len(events) > 0 {
			idx = 0
		}
	case "asc":
		// aufsteigend: der erste Termin ab heute
		for i, e := range events {
			if isUpcoming(e) {
				idx = i
				break
			}
		}
	case "desc":
		// absteigend: der letzte Termin ab heute (direkt vor dem ersten vergangenen)
		for i, e := range events {
			if isUpcoming(e) {
				idx = i
			} else {
				break
			}
		}
	}
	if idx >= 0 {
		val := true
		events[idx].Upcomming = &val
	}
}

func (h *Handler) PublicLoad(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()
	args := urlArgs{
		Sort:      query.Get("sort"),
		Head:      query.Get("head"),
		Out:       query.Get("out"),
		Sep:       query.Get("sep"),
		Upcomming: query.Get("upcomming"),
	}
	events := h.eventsByArgs(ctx, args)

	outOpt := "json"
	if args.Out != "" {
		outOpt = args.Out
	}
	var err error
	switch outOpt {
	case "json":
		bytes, err := json.Marshal(events)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bytes)
		return
	case "html":
		// this can be done prettier => see checklist-tool
		tmpl := template.Must(template.New("events-range.html").Funcs(template.FuncMap{
			"formatTime": formatTime,
		}).ParseFS(h.InternalFS, "internal/events/templates/events-range.html"))
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

func makeHash(datum string, event string, ort string, stadt string) string {
	// We take these values to generate a unique hash
	data := make([]string, 5)
	data[0] = datum
	data[1] = event
	data[2] = ort
	data[3] = stadt
	data[4] = strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	sort.Strings(data)
	var chars []byte
	for _, c := range data {
		chars = append(chars, []byte(c)...)
	}
	algo := sha256.New()
	algo.Write(chars)
	// base58 is used, because it misses certain chars which makes it more human-readable.
	// It's also used for bitcoin-address and stuff like that
	hash := base58.Encode(algo.Sum(nil))
	return hash
}

// Takes a yaml file with multiple new events
func (h *Handler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(1 << 20)
	file, header, err := r.FormFile("yaml")
	if err != nil {
		msg := fmt.Sprintf("Failed reading the request: %v \n %s \n", header, file)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Parse the YAML file from the request body
	body, err := io.ReadAll(file)
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
			log.Printf("Invalid date format for event %s: %v\n Right format is dd.mm.yyy", e.Event, err)
			continue
		}
		doorsTime := sql.NullTime{Time: date, Valid: true}
		beginTime := sql.NullTime{Time: date, Valid: true}
		linkString := sql.NullString{String: e.Link, Valid: e.Link != ""}
		hash := makeHash(e.Datum, e.Event, e.Ort, e.Stadt)

		err = q.AddSingleEvent(ctx, db.AddSingleEventParams{
			Name:  e.Event,
			Date:  date,
			Doors: doorsTime,
			Begin: beginTime,
			Place: e.Ort,
			City:  e.Stadt,
			Link:  linkString,
			Pic:   nil,
			Hash:  hash,
		})
		if err != nil {
			log.Printf("Failed to add event %s: %v", e.Event, err)
			continue
		}
	}
	http.Redirect(w, r, "/manage/", http.StatusSeeOther)
}

func (h *Handler) ProvideICSFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hash := mux.Vars(r)["id"]

	q := db.New(h.DB)

	event, err := q.GetEventByHash(ctx, hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)
	e := cal.AddEvent(hash)
	e.SetSummary("")
	e.SetLocation("Address")
	if event.Link.Valid {
		e.SetURL(event.Link.String)
	}
	card := cal.Serialize()

	fileName := fmt.Sprintf("Elisa_Adam_am_%s_in_%s.ics", event.Date.Format("02.01.2006"), event.City)

	w.Header().Set("Content-Type", "text/calendar")
	disposition := fmt.Sprintf("attachment; filename=%s", fileName)
	w.Header().Set("Content-Disposition", disposition)

	_, err = io.Copy(w, strings.NewReader(card))
	if err != nil {
		w.Write([]byte("unfortunatly no ics for you :( server problem..."))
	}
}

func logErr(w http.ResponseWriter, msg string, err error, code int) {
	log.Printf("%s\nError: %q \n", msg, err)
	http.Error(w, msg, code)
}

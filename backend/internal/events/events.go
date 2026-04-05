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
	"net/url"
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
	InternalFS embed.FS
}

func formatTime(s time.Time) template.HTML {
	return template.HTML(s.Format("02.01.2006"))
}

type Event struct {
	Name        string
	Date        time.Time
	Doors       time.Time
	Begin       time.Time
	Place       string
	City        string
	Coordinates string
	Link        *string
	Hash        string
	Separator   *string
	Upcomming   *bool
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

// ##########################################################################################################
// Single event

type EventByIDRow struct {
	r db.GetEventByIDRow
}

func (id EventByIDRow) GetName() string        { return id.r.Name }
func (id EventByIDRow) GetDate() time.Time     { return id.r.Date }
func (id EventByIDRow) GetDoors() time.Time    { return id.r.Doors.Time }
func (id EventByIDRow) GetBegin() time.Time    { return id.r.Begin.Time }
func (id EventByIDRow) GetPlace() string       { return id.r.Place }
func (id EventByIDRow) GetCity() string        { return id.r.City }
func (id EventByIDRow) GetCoordinates() string { return id.r.Coordinates.String }
func (id EventByIDRow) GetLink() *string {
	if !id.r.Link.Valid || id.r.Link.String == "" {
		return nil
	}
	return &id.r.Link.String
}
func (id EventByIDRow) GetHash() string { return id.r.Hash }

func adaptByIDRow(row db.GetEventByIDRow) EventByIDRow {
	return EventByIDRow{r: row}
}

// ##########################################################################################################
// When a new query with a new db.Struct is generated, you need to add the same boilerplate from above here
// ###########################################################################################################

// In the future I may gonna need to add custom-values
// This way
func constructEvents(rows []EventRow) []Event {
	var events = make([]Event, len(rows))
	for i, r := range rows {
		events[i] = Event{
			Name:        r.GetName(),
			Date:        r.GetDate(),
			Doors:       r.GetDoors(),
			Begin:       r.GetBegin(),
			Place:       r.GetPlace(),
			City:        r.GetCity(),
			Coordinates: r.GetCoordinates(),
			Link:        r.GetLink(),
			Hash:        r.GetHash(),
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

// Returns []db.Event
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
		for i := range events {
			curr := events[i].Date
			// If the event is today, still highlight it!
			yesterday := time.Now().AddDate(0, 0, -1)
			val := true
			if i == len(events) {
				events[i].Upcomming = &val
				break
			}
			// This can produce a runtime error.... if the if-condition above didnt existed...
			next := events[i+1].Date
			// Just set the next one as highlighted.
			if sortBy == "next" && i == 0 {
				events[i].Upcomming = &val
				break
			}
			if sortBy == "desc" && curr.Unix() > yesterday.Unix() && next.Unix() < yesterday.Unix() && i != len(events)-1 {
				events[i].Upcomming = &val
				break
			}
			if sortBy == "asc" && curr.Unix() > yesterday.Unix() && next.Unix() > yesterday.Unix() && i != len(events)-1 {
				events[i].Upcomming = &val
				break
			}

		}
	}

	return events

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

// Create a single new Event
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	datum := r.FormValue("datum")
	event := r.FormValue("event")
	ort := r.FormValue("ort")
	stadt := r.FormValue("stadt")
	link := r.FormValue("link")

	date, err := time.Parse("2006-01-02", datum)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	doorsTime := sql.NullTime{Time: date, Valid: true}
	beginTime := sql.NullTime{Time: date, Valid: true}
	linkString := sql.NullString{String: link, Valid: link != ""}

	hash := makeHash(datum, event, ort, stadt)

	err = q.AddSingleEvent(ctx, db.AddSingleEventParams{
		Name:  event,
		Date:  date,
		Doors: doorsTime,
		Begin: beginTime,
		Place: ort,
		City:  stadt,
		Link:  linkString,
		Pic:   nil,
		Hash:  hash,
	})
	if err != nil {
		log.Printf("Failed to add event: %v", err)
		http.Error(w, "Failed to add event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	args := urlArgs{
		Sort: "desc",
		Out:  "html",
	}
	events := h.eventsByArgs(ctx, args)
	// this can be done prettier => see checklist-tool
	tmpl := template.Must(template.New("events-mutable.html").Funcs(template.FuncMap{
		"formatTime": formatTime,
	}).ParseFS(h.InternalFS, "internal/events/templates/events-mutable.html"))
	err = tmpl.Execute(w, map[string]any{
		"Events": events,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	http.Redirect(w, r, "/events/manage", http.StatusSeeOther)
}

type EventSmooth struct {
	ID    int64
	Name  string
	Date  string
	Place string
	City  string
	Link  *string
}

// Loads all events descending, with all fields to alter events
func (h *Handler) PrivateLoad(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)

	all, err := q.GetAllEventsDesc(ctx)
	if err != nil {
		logErr(w, "Load all events desc", err, http.StatusInternalServerError)
		return
	}
	// Transform the data, so the frontend can work with the values as they are
	resp := make([]EventSmooth, len(all))
	for i, e := range all {
		// Just giving out the pure link if it is there
		var link *string
		if e.Link.Valid {
			link = &e.Link.String
		}
		// Format the Date to German style
		var date string
		date = e.Date.Format("02.01.2006")
		resp[i] = EventSmooth{
			ID:    e.ID,
			Name:  e.Name,
			Date:  date,
			Place: e.Place,
			City:  e.City,
			Link:  link,
		}
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(bytes)
}

// Handles changes to a single cell of an event
// Idea: Instead of receiving just one value at the time, we could
// insert the complete row and skip a lot of the logic down below
func (h *Handler) Change(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)
	err := r.ParseForm()
	if err != nil {
		logErr(w, "Couldn't parse the form.", err, http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("The id:'%s' is invalid.", idStr)
		logErr(w, msg, err, http.StatusBadRequest)
		return
	}
	field := r.FormValue("field")
	value := r.FormValue("value")

	// This humongous switch-case-statement is done, because I refuse to pass in the complete row.
	// Because of this infantility I need to declare for each field a separate SQL-Query
	switch field {
	case "Name":
		args := db.UpdateNameByIDParams{Name: value, ID: id}
		q.UpdateNameByID(ctx, args)
	case "Date":
		date, err := time.Parse("02.01.2006", value)
		args := db.UpdateDateByIDParams{Date: date, ID: id}
		if err != nil {
			logErr(w, "Wrong time-data format. We want 'dd.mm.yyyy'/'02.01.2006'", err, http.StatusBadRequest)
			return
		}
		q.UpdateDateByID(ctx, args)
	case "Doors":
		if value == "" {
			return
		}
		// 15:04 is hh:mm
		t, err := time.Parse("15:04", value)
		if err != nil {
			logErr(w, "Wrong time-data format. We want 'hh:mm'/'15:04'", err, http.StatusBadRequest)
			return
		}
		doors := sql.NullTime{Time: t, Valid: true}
		args := db.UpdateDoorsByIDParams{Doors: doors, ID: id}
		q.UpdateDoorsByID(ctx, args)
	case "Begin":
		if value == "" {
			return
		}
		// 15:04 is hh:mm
		t, err := time.Parse("15:04", value)
		if err != nil {
			logErr(w, "Wrong time-data format. We want 'hh:mm'/'15:04'", err, http.StatusBadRequest)
			return
		}
		begin := sql.NullTime{Time: t, Valid: true}
		args := db.UpdateBeginByIDParams{Begin: begin, ID: id}
		q.UpdateBeginByID(ctx, args)
	case "Place":
		args := db.UpdatePlaceByIDParams{Place: value, ID: id}
		q.UpdatePlaceByID(ctx, args)
	case "City":
		args := db.UpdateCityByIDParams{City: value, ID: id}
		q.UpdateCityByID(ctx, args)
	case "Coordinates":
		if value == "" {
			return
		}
		coords := sql.NullString{String: value, Valid: true}
		args := db.UpdateCoordinatesByIDParams{Coordinates: coords, ID: id}
		q.UpdateCoordinatesByID(ctx, args)
	case "Link":
		if value == "" {
			return
		}
		link := sql.NullString{String: value, Valid: true}
		args := db.UpdateLinkByIDParams{Link: link, ID: id}
		q.UpdateLinkByID(ctx, args)
	case "Hash":
		args := db.UpdateHashByIDParams{Hash: value, ID: id}
		q.UpdateHashByID(ctx, args)
	}
}

func (h *Handler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	datum, err := time.Parse("02.01.2006", r.FormValue("date"))
	if err != nil {
		msg := "Wrong time date format."
		intern := msg + fmt.Sprintf(" Error happend: %v \n", err.Error())
		log.Println(intern)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	event := r.FormValue("name")
	ort := r.FormValue("place")
	stadt := r.FormValue("city")

	argsID := db.GetIDsFromEventParams{
		Name:  event,
		Date:  datum,
		Place: ort,
		City:  stadt,
	}
	// There could be event duplicates, so fetch all ids
	ids, err := q.GetIDsFromEvent(ctx, argsID)
	if err != nil {
		http.Error(w, "Database error.", http.StatusBadRequest)
		return
	}
	if len(ids) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No ids found."))
		return
	}
	// This logging is ugly and needs to be redone...
	row, err := q.GetEventByID(ctx, ids[0])
	msg := `The event will get deleted:
	Event:	%s 
	Datum:	%s
	Ort:  %s
	Stadt:	%s
	Link:		%s
	`
	log.Printf(msg, event, datum, ort, stadt, row.Link.String)

	// If there was an unique id for an entry, we could delete more precise.
	// This way we ONLY match form-data which also could come from a different event, which holds the same data.
	// Indeed this is dirty, but user shouldn't detect it.... (maybe)
	err = q.DeleteEventByID(ctx, ids[0])
	if err != nil {
		http.Error(w, "Couldn't delete.", http.StatusNotModified)
		return
	}
	args := urlArgs{
		Sort: "desc",
	}
	events := h.eventsByArgs(ctx, args)
	tmpl := template.Must(template.New("events-mutable.html").Funcs(template.FuncMap{
		"formatTime": formatTime,
	}).ParseFS(h.InternalFS, "internal/events/templates/events-mutable.html"))
	err = tmpl.Execute(w, map[string]any{
		"Events": events,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

type Coordinates struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// Uses the Openstreetmap API
func FetchCoordinates(ort string, stadt string) (*Coordinates, error) {

	base := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", fmt.Sprintf("%s,+%s", ort, stadt))
	params.Add("format", "json")
	req, err := http.NewRequest("GET", base+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "elses-website-backend/1.0 (hendrik_maier@protonmail.com)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var coords []Coordinates
	decoder.Decode(&coords)
	if err != nil {
		return nil, err
	}
	if len(coords) > 0 {
		return &coords[0], nil
	}
	return nil, fmt.Errorf("no results found")
}

func logErr(w http.ResponseWriter, msg string, err error, code int) {
	log.Printf("%s\nError: %q \n", msg, err)
	http.Error(w, msg, code)
}

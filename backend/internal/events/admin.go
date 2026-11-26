package events

// Verwaltung der Termine unter /manage/api/events.
// Alle Antworten sind HTML-Schnipsel für htmx (siehe templates/admin-events.html).

import (
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/auth"
	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

//go:embed templates/admin-events.html
var adminFS embed.FS

var weekdays = [...]string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"}
var months = [...]string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}

var adminTmpl = template.Must(template.New("admin").Funcs(template.FuncMap{
	"day":     func(t time.Time) string { return t.Format("02") },
	"month":   func(t time.Time) string { return months[t.Month()-1] },
	"weekday": func(t time.Time) string { return weekdays[t.Weekday()] },
	"year":    func(t time.Time) int { return t.Year() },
}).ParseFS(adminFS, "templates/admin-events.html"))

// AdminEvent ist ein Termin, wie ihn die Verwaltung anzeigt und bearbeitet.
type AdminEvent struct {
	ID    int64
	Name  string
	Date  time.Time
	Time  string // "20:00" oder ""
	Place string
	City  string
	Link  string // gültiger http(s)-Link
	Note  string // Hinweistext statt Link, z. B. "Eintritt: Spendenbasis"
	Raw   string // Feldinhalt fürs Bearbeiten
}

// formValues sind die Eingaben im Formular (auch bei Fehlern wieder angezeigt).
type formValues struct {
	Name, Date, Time, Place, City, Link string
}

func (e AdminEvent) form() formValues {
	return formValues{e.Name, e.Date.Format("2006-01-02"), e.Time, e.Place, e.City, e.Raw}
}

func toAdminEvent(id int64, name string, date time.Time, begin sql.NullTime, place, city string, link sql.NullString) AdminEvent {
	l, note := splitLink(link.String)
	raw := l
	if raw == "" {
		raw = note
	}
	return AdminEvent{
		ID: id, Name: name, Date: date, Time: clockTime(begin),
		Place: place, City: city, Link: l, Note: note, Raw: raw,
	}
}

// clockTime liefert "20:00", oder "" wenn keine Uhrzeit gesetzt ist (00:00 gilt als "keine").
func clockTime(t sql.NullTime) string {
	if !t.Valid || (t.Time.Hour() == 0 && t.Time.Minute() == 0) {
		return ""
	}
	return t.Time.Format("15:04")
}

type parsedEvent struct {
	Name, Place, City string
	Date              time.Time
	Begin             sql.NullTime
	Link              sql.NullString
}

// parseEventForm prüft die Formulareingaben und liefert eine verständliche Fehlermeldung.
func parseEventForm(r *http.Request) (parsedEvent, formValues, string) {
	f := formValues{
		Name:  strings.TrimSpace(r.FormValue("name")),
		Date:  strings.TrimSpace(r.FormValue("date")),
		Time:  strings.TrimSpace(r.FormValue("time")),
		Place: strings.TrimSpace(r.FormValue("place")),
		City:  strings.TrimSpace(r.FormValue("city")),
		Link:  strings.TrimSpace(r.FormValue("link")),
	}
	var p parsedEvent
	if f.Name == "" || f.Date == "" || f.Place == "" || f.City == "" {
		return p, f, "Bitte Name, Datum, Ort und Stadt ausfüllen."
	}
	date, err := time.Parse("2006-01-02", f.Date)
	if err != nil {
		return p, f, "Das Datum ist ungültig."
	}
	begin := date
	if f.Time != "" {
		t, err := time.Parse("15:04", f.Time)
		if err != nil {
			return p, f, "Die Uhrzeit ist ungültig (Format 20:00)."
		}
		begin = date.Add(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute)
	}
	if msg := validLinkInput(f.Link); msg != "" {
		return p, f, msg
	}
	p = parsedEvent{
		Name: f.Name, Place: f.Place, City: f.City, Date: date,
		Begin: sql.NullTime{Time: begin, Valid: true},
		Link:  sql.NullString{String: f.Link, Valid: f.Link != ""},
	}
	return p, f, ""
}

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminTmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("admin: template %s: %v", name, err)
	}
}

func eventID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	return id, err == nil
}

type panelData struct {
	Scope     string // welche Liste zuerst sichtbar ist
	Upcoming  []AdminEvent
	Past      []AdminEvent
	Form      formValues
	Error     string
	Flash     string
	NewFormID string
}

func (h *Handler) panel(r *http.Request, scope string) (panelData, error) {
	q := db.New(h.DB)
	ctx := r.Context()
	up, err := q.ListEventsUpcoming(ctx)
	if err != nil {
		return panelData{}, err
	}
	past, err := q.ListEventsPast(ctx)
	if err != nil {
		return panelData{}, err
	}
	d := panelData{Scope: scope}
	if d.Scope != "past" {
		d.Scope = "upcoming"
	}
	for _, e := range up {
		d.Upcoming = append(d.Upcoming, toAdminEvent(e.ID, e.Name, e.Date, e.Begin, e.Place, e.City, e.Link))
	}
	for _, e := range past {
		d.Past = append(d.Past, toAdminEvent(e.ID, e.Name, e.Date, e.Begin, e.Place, e.City, e.Link))
	}
	return d, nil
}

// AdminPanel: GET /manage/api/events?scope=upcoming|past
func (h *Handler) AdminPanel(w http.ResponseWriter, r *http.Request) {
	d, err := h.panel(r, r.URL.Query().Get("scope"))
	if err != nil {
		logErr(w, "Termine laden", err, http.StatusInternalServerError)
		return
	}
	render(w, "events-panel", d)
}

// AdminCreate: POST /manage/api/events
func (h *Handler) AdminCreate(w http.ResponseWriter, r *http.Request) {
	p, f, msg := parseEventForm(r)
	if msg == "" {
		err := db.New(h.DB).AddSingleEvent(r.Context(), db.AddSingleEventParams{
			Name: p.Name, Date: p.Date, Doors: p.Begin, Begin: p.Begin,
			Place: p.Place, City: p.City, Link: p.Link,
			Hash: makeHash(p.Date.Format("02.01.2006"), p.Name, p.Place, p.City),
		})
		if err != nil {
			logErr(w, "Termin speichern", err, http.StatusInternalServerError)
			return
		}
		log.Printf("admin: %s hat Termin angelegt: %s am %s", auth.User(r), p.Name, p.Date.Format("02.01.2006"))
	}

	// Ansicht passend zum neuen Termin (kommend/vergangen) zeigen
	scope := "upcoming"
	if msg == "" && p.Date.Before(time.Now().Truncate(24*time.Hour)) {
		scope = "past"
	}
	d, err := h.panel(r, scope)
	if err != nil {
		logErr(w, "Termine laden", err, http.StatusInternalServerError)
		return
	}
	if msg != "" {
		d.Form, d.Error = f, msg
	} else {
		d.Flash = "„" + p.Name + "“ wurde gespeichert."
	}
	render(w, "events-panel", d)
}

func (h *Handler) loadAdminEvent(r *http.Request) (AdminEvent, error) {
	id, ok := eventID(r)
	if !ok {
		return AdminEvent{}, sql.ErrNoRows
	}
	e, err := db.New(h.DB).GetEventFull(r.Context(), id)
	if err != nil {
		return AdminEvent{}, err
	}
	return toAdminEvent(e.ID, e.Name, e.Date, e.Begin, e.Place, e.City, e.Link), nil
}

// AdminRow: GET /manage/api/events/{id} – eine Zeile (z. B. nach "Abbrechen")
func (h *Handler) AdminRow(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadAdminEvent(r)
	if err != nil {
		notFound(w, err)
		return
	}
	render(w, "event-row", e)
}

// AdminEdit: GET /manage/api/events/{id}/edit – Bearbeiten-Formular
func (h *Handler) AdminEdit(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadAdminEvent(r)
	if err != nil {
		notFound(w, err)
		return
	}
	render(w, "event-edit", map[string]any{"ID": e.ID, "Form": e.form()})
}

// AdminUpdate: PUT /manage/api/events/{id}
func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := eventID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	p, f, msg := parseEventForm(r)
	if msg != "" {
		render(w, "event-edit", map[string]any{"ID": id, "Form": f, "Error": msg})
		return
	}
	err := db.New(h.DB).UpdateEvent(r.Context(), db.UpdateEventParams{
		Name: p.Name, Date: p.Date, Doors: p.Begin, Begin: p.Begin,
		Place: p.Place, City: p.City, Link: p.Link, ID: id,
	})
	if err != nil {
		logErr(w, "Termin ändern", err, http.StatusInternalServerError)
		return
	}
	log.Printf("admin: %s hat Termin %d geändert", auth.User(r), id)
	h.AdminRow(w, r)
}

// AdminDelete: DELETE /manage/api/events/{id}
func (h *Handler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	e, err := h.loadAdminEvent(r)
	if err != nil {
		notFound(w, err)
		return
	}
	if err := db.New(h.DB).DeleteEventByID(r.Context(), e.ID); err != nil {
		logErr(w, "Termin löschen", err, http.StatusInternalServerError)
		return
	}
	log.Printf("admin: %s hat Termin %d gelöscht: %s am %s", auth.User(r), e.ID, e.Name, e.Date.Format("02.01.2006"))
	// leere Antwort → htmx entfernt die Zeile
	w.WriteHeader(http.StatusOK)
}

func notFound(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Termin nicht gefunden", http.StatusNotFound)
		return
	}
	logErr(w, "Termin laden", err, http.StatusInternalServerError)
}

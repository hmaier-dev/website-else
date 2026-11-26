package mailbox

// Nachrichten aus dem Kontaktformular für die Verwaltung unter /manage/api/messages.
//
// Ansichten:  Posteingang (archived = 0) und Erledigt (archived = 1)
// Aktionen:   gelesen/ungelesen, erledigt/zurück, löschen

import (
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/auth"
	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

//go:embed templates/admin-messages.html
var adminFS embed.FS

var adminTmpl = template.Must(template.New("admin").Funcs(template.FuncMap{
	"date": func(unix int64) string { return time.Unix(unix, 0).Local().Format("02.01.2006, 15:04") },
	"ago":  ago,
}).ParseFS(adminFS, "templates/admin-messages.html"))

// ago gibt eine grobe Zeitangabe wie "vor 3 Stunden" zurück.
func ago(unix int64) string {
	d := time.Since(time.Unix(unix, 0))
	switch {
	case d < time.Minute:
		return "gerade eben"
	case d < time.Hour:
		return plural(int(d.Minutes()), "Minute", "Minuten")
	case d < 24*time.Hour:
		return plural(int(d.Hours()), "Stunde", "Stunden")
	case d < 30*24*time.Hour:
		return plural(int(d.Hours()/24), "Tag", "Tagen")
	default:
		return "am " + time.Unix(unix, 0).Local().Format("02.01.2006")
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "vor 1 " + one
	}
	return "vor " + strconv.Itoa(n) + " " + many
}

type message struct {
	ID      int64
	Name    string
	Email   string
	Message string
	Date    int64
	Unread  bool
}

type messagesPanel struct {
	View     string // "inbox" | "archived"
	Messages []message
	Counts   db.CountMessagesRow
}

func view(r *http.Request) string {
	if r.URL.Query().Get("view") == "archived" {
		return "archived"
	}
	return "inbox"
}

func (h *Handler) renderMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)
	p := messagesPanel{View: view(r)}

	var err error
	if p.Counts, err = q.CountMessages(ctx); err == nil {
		if p.View == "archived" {
			var rows []db.ListMessagesArchivedRow
			rows, err = q.ListMessagesArchived(ctx)
			for _, m := range rows {
				p.Messages = append(p.Messages, message{m.ID, m.Name, m.Email, m.Message, m.Date, m.Isread.Int64 == 0})
			}
		} else {
			var rows []db.ListMessagesInboxRow
			rows, err = q.ListMessagesInbox(ctx)
			for _, m := range rows {
				p.Messages = append(p.Messages, message{m.ID, m.Name, m.Email, m.Message, m.Date, m.Isread.Int64 == 0})
			}
		}
	}
	if err != nil {
		log.Printf("admin: Nachrichten laden: %v", err)
		http.Error(w, "Nachrichten konnten nicht geladen werden", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminTmpl.ExecuteTemplate(w, "messages-panel", p); err != nil {
		log.Printf("admin: template messages-panel: %v", err)
	}
}

// AdminMessages: GET /manage/api/messages?view=inbox|archived
func (h *Handler) AdminMessages(w http.ResponseWriter, r *http.Request) {
	h.renderMessages(w, r)
}

// AdminBadge: GET /manage/api/messages/badge – Zahl der ungelesenen Nachrichten für den Reiter
func (h *Handler) AdminBadge(w http.ResponseWriter, r *http.Request) {
	c, err := db.New(h.DB).CountMessages(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil || c.Unread == 0 {
		return
	}
	adminTmpl.ExecuteTemplate(w, "messages-badge", c.Unread)
}

// AdminMessageAction: POST /manage/api/messages/{id}/{action}
// action: read | unread | archive | unarchive
func (h *Handler) AdminMessageAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := db.New(h.DB)
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch mux.Vars(r)["action"] {
	case "read":
		err = q.SetMessageRead(ctx, db.SetMessageReadParams{Isread: sql.NullInt64{Int64: 1, Valid: true}, ID: id})
	case "unread":
		err = q.SetMessageRead(ctx, db.SetMessageReadParams{Isread: sql.NullInt64{Int64: 0, Valid: true}, ID: id})
	case "archive":
		err = q.SetMessageArchived(ctx, db.SetMessageArchivedParams{Archived: 1, ID: id})
	case "unarchive":
		err = q.SetMessageArchived(ctx, db.SetMessageArchivedParams{Archived: 0, ID: id})
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("admin: Nachricht %d %s: %v", id, mux.Vars(r)["action"], err)
		http.Error(w, "Konnte nicht gespeichert werden", http.StatusInternalServerError)
		return
	}
	// Zähler im Reiter aktualisieren
	w.Header().Set("HX-Trigger", "messages-changed")
	h.renderMessages(w, r)
}

// AdminDeleteMessage: DELETE /manage/api/messages/{id}
// Wichtig für den Datenschutz: Anfragen nach Erledigung löschen.
func (h *Handler) AdminDeleteMessage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := db.New(h.DB).DeleteMessage(r.Context(), id); err != nil {
		log.Printf("admin: Nachricht %d löschen: %v", id, err)
		http.Error(w, "Konnte nicht gelöscht werden", http.StatusInternalServerError)
		return
	}
	log.Printf("admin: %s hat Nachricht %d gelöscht", auth.User(r), id)
	w.Header().Set("HX-Trigger", "messages-changed")
	h.renderMessages(w, r)
}

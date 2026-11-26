package mailbox

// Nachrichten aus dem Kontaktformular für die Verwaltung unter /manage/api/messages.

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

//go:embed templates/admin-messages.html
var adminFS embed.FS

var adminTmpl = template.Must(template.New("admin").Funcs(template.FuncMap{
	"date": func(unix int64) string { return time.Unix(unix, 0).Local().Format("02.01.2006, 15:04") },
}).ParseFS(adminFS, "templates/admin-messages.html"))

type messagesPanel struct {
	Messages []db.Mailbox
	Unread   int
}

func (h *Handler) renderMessages(w http.ResponseWriter, r *http.Request) {
	all, err := db.New(h.DB).GetCompleteMailbox(r.Context())
	if err != nil {
		log.Printf("admin: Nachrichten laden: %v", err)
		http.Error(w, "Nachrichten konnten nicht geladen werden", http.StatusInternalServerError)
		return
	}
	p := messagesPanel{}
	// neueste zuerst
	for i := len(all) - 1; i >= 0; i-- {
		p.Messages = append(p.Messages, all[i])
		if all[i].Isread.Int64 == 0 {
			p.Unread++
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminTmpl.ExecuteTemplate(w, "messages-panel", p); err != nil {
		log.Printf("admin: template messages-panel: %v", err)
	}
}

// AdminMessages: GET /manage/api/messages
func (h *Handler) AdminMessages(w http.ResponseWriter, r *http.Request) {
	h.renderMessages(w, r)
}

// AdminMarkRead: POST /manage/api/messages/{id}/read
func (h *Handler) AdminMarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := db.New(h.DB).MarkMessageRead(r.Context(), id); err != nil {
		log.Printf("admin: Nachricht %d gelesen: %v", id, err)
		http.Error(w, "Konnte nicht gespeichert werden", http.StatusInternalServerError)
		return
	}
	h.renderMessages(w, r)
}

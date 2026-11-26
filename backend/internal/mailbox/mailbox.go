package mailbox

import (
	"context"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
	"github.com/jedib0t/go-pretty/v6/table"
)

type API struct {
	Username string
	Password string
}

// Notifier wird bei jeder neuen Nachricht aufgerufen (z. B. E-Mail an Else).
// nil = keine Benachrichtigung.
type Notifier interface {
	NewContactMessage(name, email, message string)
}

type Handler struct {
	DB     *sql.DB
	API    API
	Notify Notifier
}

const (
	maxName    = 200
	maxEmail   = 254
	maxMessage = 5000
)

// validate prüft die Eingaben und gibt eine Fehlermeldung für die Absenderin zurück.
func validate(name, email, message string) string {
	switch {
	case name == "" || email == "" || message == "":
		return "Bitte Name, E-Mail und Nachricht ausfüllen."
	case utf8.RuneCountInString(name) > maxName:
		return "Der Name ist zu lang."
	case len(email) > maxEmail || !validEmail(email):
		return "Die E-Mail-Adresse sieht nicht richtig aus."
	case utf8.RuneCountInString(message) > maxMessage:
		return fmt.Sprintf("Die Nachricht ist zu lang (höchstens %d Zeichen).", maxMessage)
	}
	return ""
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	// nur die nackte Adresse erlauben, keine "Name <adresse>"-Form oder Zeilenumbrüche
	return err == nil && a.Address == s && !strings.ContainsAny(s, "\r\n")
}

// Incoming POST request from the html formular
// TODO: Implement rate limiting at traefik
func (h *Handler) HandlerContactRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Spam-Falle: Das Feld "website" ist für Menschen unsichtbar. Bots füllen es aus.
	// Wir tun so, als hätte alles geklappt, speichern aber nichts.
	if r.FormValue("website") != "" {
		log.Println("mailbox: Spam-Falle ausgelöst, Nachricht verworfen")
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.FormValue("approval") != "on" {
		http.Error(w, "Bitte stimme der Datenschutzerklärung zu.", http.StatusForbidden)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	message := strings.TrimSpace(r.FormValue("message"))
	if msg := validate(name, email, message); msg != "" {
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return
	}

	q := db.New(h.DB)
	err := q.AddMessageToMailbox(ctx, db.AddMessageToMailboxParams{
		Name:     name,
		Email:    email,
		Message:  message,
		Date:     time.Now().Unix(),
		Approval: sql.NullInt64{Int64: 1, Valid: true},
		Isread:   sql.NullInt64{Int64: 0, Valid: false},
	})
	if err != nil {
		// Nicht log.Fatal: ein Fehler beim Speichern darf nicht das ganze Backend beenden.
		log.Printf("mailbox: Nachricht speichern fehlgeschlagen: %v", err)
		http.Error(w, "Die Nachricht konnte nicht gespeichert werden.", http.StatusInternalServerError)
		return
	}
	// Keine Inhalte ins Log schreiben (Datenschutz), nur dass etwas angekommen ist.
	log.Println("mailbox: neue Nachricht gespeichert")
	if h.Notify != nil {
		h.Notify.NewContactMessage(name, email, message)
	}
	w.WriteHeader(http.StatusOK)
}

// Serving GET request to return all messages
func (h *Handler) HandlerAllMessages(w http.ResponseWriter, r *http.Request) {

	// Refactor this into one function
	// Here is how to do it
	// https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
	u, p, ok := r.BasicAuth()
	realm := "Einmal bitte Credentials eingeben: "
	if !ok || subtle.ConstantTimeCompare([]byte(u), []byte(h.API.Username)) != 1 || subtle.ConstantTimeCompare([]byte(p), []byte(h.API.Password)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
		return
	}

	ctx := context.Background()
	q := db.New(h.DB)
	all, err := q.GetCompleteMailbox(ctx)
	if err != nil {
		http.Error(w, "Something wen't wrong querying all messages...", http.StatusInternalServerError)
	}
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{"ID", "Name", "Email", "Message", "Date", "Approval", "IsRead"})
	for _, a := range all {
		d := time.Unix(a.Date, 0).Format("02.01.2006 15:04:05")
		t.AppendRow(table.Row{a.ID, a.Name, a.Email, a.Message, d, a.Approval.Int64, a.Isread.Int64})
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}


func (h *Handler) HandlerUnreadMessages(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	q := db.New(h.DB)
	unread, err := q.GetUnreadMailboxItems(ctx)

	// Refactor this into one function
	// Here is how to do it
	// https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
	u, p, ok := r.BasicAuth()
	realm := "Einmal bitte Credentials eingeben: "
	if !ok || subtle.ConstantTimeCompare([]byte(u), []byte(h.API.Username)) != 1 || subtle.ConstantTimeCompare([]byte(p), []byte(h.API.Password)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
		return
	}

	if err != nil {
		http.Error(w, "Something wen't wrong querying unread messages...", http.StatusInternalServerError)
	}
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.AppendHeader(table.Row{"ID", "Name", "Email", "Message", "Date", "Approval", "IsRead"})
	for _, u := range unread {
		d := time.Unix(u.Date, 0).Format("02.01.2006 15:04:05")
		t.AppendRow(table.Row{u.ID, u.Name, u.Email, u.Message, d, u.Approval.Int64, u.Isread.Int64})
	}
	t.SetStyle(table.StyleLight)
	t.Render()
}

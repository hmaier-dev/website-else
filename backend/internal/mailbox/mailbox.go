package mailbox

import (
	"context"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"time"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
	"github.com/jedib0t/go-pretty/v6/table"
)

type API struct {
	Username string
	Password string
}

type Handler struct{
	DB *sql.DB
	API API
}

// Incoming POST request from the html formular
// TODO: Implement rate limiting at traefik
func (h *Handler) HandlerContactRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	approval := r.FormValue("approval")
	log.Println(name, email, message, approval)

	// TODO: Sanatizing and validating the input
	if approval == "on" {
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
			http.Error(w, "Something wen't wrong storing the data...", http.StatusInternalServerError)
			log.Fatalf("Err: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "We need approval to store user data", http.StatusForbidden)
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

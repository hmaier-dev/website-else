package mailbox

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

type recorder struct{ names []string }

func (r *recorder) NewContactMessage(name, email, message string) { r.names = append(r.names, name) }

func setup(t *testing.T) (*mux.Router, *sql.DB, *recorder) {
	t.Helper()
	sqlite, _ := sql.Open("sqlite3", ":memory:")
	sqlite.SetMaxOpenConns(1)
	ddl, _ := os.ReadFile("../../schema.sql")
	if _, err := sqlite.Exec(string(ddl)); err != nil {
		t.Fatal(err)
	}
	rec := &recorder{}
	h := &Handler{DB: sqlite, Notify: rec}
	r := mux.NewRouter()
	r.HandleFunc("/mailbox/contact", h.HandlerContactRequest).Methods("POST")
	r.HandleFunc("/manage/api/messages", h.AdminMessages).Methods("GET")
	r.HandleFunc("/manage/api/messages/badge", h.AdminBadge).Methods("GET")
	r.HandleFunc(`/manage/api/messages/{id:\d+}/{action:read|unread|archive|unarchive}`, h.AdminMessageAction).Methods("POST")
	r.HandleFunc(`/manage/api/messages/{id:\d+}`, h.AdminDeleteMessage).Methods("DELETE")
	return r, sqlite, rec
}

func send(r *mux.Router, method, path string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func count(t *testing.T, sqlite *sql.DB) db.CountMessagesRow {
	c, err := db.New(sqlite).CountMessages(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func valid() url.Values {
	return url.Values{"name": {"Max Muster"}, "email": {"max@example.com"}, "message": {"Hallo Else!"}, "approval": {"on"}}
}

func TestContactValidation(t *testing.T) {
	r, sqlite, rec := setup(t)
	cases := map[string]func(url.Values){
		"ohne Zustimmung":    func(v url.Values) { v.Del("approval") },
		"leerer Name":        func(v url.Values) { v.Set("name", "  ") },
		"kaputte E-Mail":     func(v url.Values) { v.Set("email", "keine-mail") },
		"E-Mail mit Name":    func(v url.Values) { v.Set("email", "Max <max@example.com>") },
		"zu lange Nachricht": func(v url.Values) { v.Set("message", strings.Repeat("a", 5001)) },
	}
	for name, mod := range cases {
		v := valid()
		mod(v)
		if w := send(r, "POST", "/mailbox/contact", v); w.Code < 400 {
			t.Errorf("%s: HTTP %d, erwartet Fehler", name, w.Code)
		}
	}
	if c := count(t, sqlite); c.Inbox != 0 || len(rec.names) != 0 {
		t.Errorf("ungültige Nachrichten wurden gespeichert/gemeldet: %+v %v", c, rec.names)
	}
}

func TestHoneypot(t *testing.T) {
	r, sqlite, rec := setup(t)
	v := valid()
	v.Set("website", "http://spam.example")
	if w := send(r, "POST", "/mailbox/contact", v); w.Code != http.StatusOK {
		t.Errorf("Bot soll 200 sehen, bekam %d", w.Code)
	}
	if c := count(t, sqlite); c.Inbox != 0 || len(rec.names) != 0 {
		t.Error("Spam wurde gespeichert oder gemeldet")
	}
}

func TestMessageLifecycle(t *testing.T) {
	r, sqlite, rec := setup(t)
	if w := send(r, "POST", "/mailbox/contact", valid()); w.Code != http.StatusOK {
		t.Fatalf("HTTP %d: %s", w.Code, w.Body.String())
	}
	if len(rec.names) != 1 || rec.names[0] != "Max Muster" {
		t.Errorf("Benachrichtigung: %v", rec.names)
	}
	if c := count(t, sqlite); c.Unread != 1 || c.Inbox != 1 {
		t.Fatalf("counts nach Eingang: %+v", c)
	}
	if b := send(r, "GET", "/manage/api/messages/badge", nil).Body.String(); !strings.Contains(b, ">1<") {
		t.Errorf("Badge: %q", b)
	}

	w := send(r, "POST", "/manage/api/messages/1/read?view=inbox", nil)
	if w.Header().Get("HX-Trigger") != "messages-changed" {
		t.Error("HX-Trigger fehlt")
	}
	if c := count(t, sqlite); c.Unread != 0 {
		t.Errorf("nach gelesen: %+v", c)
	}
	send(r, "POST", "/manage/api/messages/1/archive?view=inbox", nil)
	if c := count(t, sqlite); c.Inbox != 0 || c.Archived != 1 {
		t.Errorf("nach erledigt: %+v", c)
	}
	if b := send(r, "GET", "/manage/api/messages?view=archived", nil).Body.String(); !strings.Contains(b, "Max Muster") {
		t.Error("erledigte Nachricht fehlt in der Ansicht Erledigt")
	}
	send(r, "DELETE", "/manage/api/messages/1?view=archived", nil)
	if c := count(t, sqlite); c.Archived != 0 || c.Inbox != 0 {
		t.Errorf("nach löschen: %+v", c)
	}
}

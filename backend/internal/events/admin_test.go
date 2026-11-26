package events

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/auth"
)

func adminRouter(t *testing.T) *mux.Router {
	t.Helper()
	sqlite, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	sqlite.SetMaxOpenConns(1)
	ddl, _ := os.ReadFile("../../schema.sql")
	if _, err := sqlite.Exec(string(ddl)); err != nil {
		t.Fatal(err)
	}
	h := &Handler{DB: sqlite}
	r := mux.NewRouter()
	m := r.PathPrefix("/manage/api").Subrouter()
	m.Use(auth.Require)
	m.HandleFunc("/events", h.AdminPanel).Methods("GET")
	m.HandleFunc("/events", h.AdminCreate).Methods("POST")
	m.HandleFunc(`/events/{id:\d+}`, h.AdminUpdate).Methods("PUT")
	m.HandleFunc(`/events/{id:\d+}`, h.AdminDelete).Methods("DELETE")
	m.HandleFunc(`/events/{id:\d+}/edit`, h.AdminEdit).Methods("GET")
	return r
}

func do(r *mux.Router, method, path string, form url.Values, user string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if user != "" {
		req.Header.Set(auth.DefaultHeader, user)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestAdminRequiresLogin(t *testing.T) {
	r := adminRouter(t)
	if rec := do(r, "GET", "/manage/api/events", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("ohne Authentik-Header: HTTP %d, erwartet 401", rec.Code)
	}
	if rec := do(r, "GET", "/manage/api/events", nil, "else"); rec.Code != http.StatusOK {
		t.Errorf("mit Authentik-Header: HTTP %d, erwartet 200", rec.Code)
	}
}

func TestAdminEventLifecycle(t *testing.T) {
	r := adminRouter(t)
	future := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

	// Fehlende Pflichtfelder → Formular mit Fehlermeldung, nichts gespeichert
	rec := do(r, "POST", "/manage/api/events", url.Values{"name": {"Slam"}}, "else")
	if !strings.Contains(rec.Body.String(), "Bitte Name, Datum, Ort und Stadt") || !strings.Contains(rec.Body.String(), `value="Slam"`) {
		t.Fatalf("Fehlermeldung/Formularwerte fehlen:\n%s", rec.Body.String())
	}

	// Ungültiger Link
	rec = do(r, "POST", "/manage/api/events", url.Values{"name": {"Slam"}, "date": {future}, "place": {"Werk 2"}, "city": {"Leipzig"}, "link": {"javascript:alert(1)"}}, "else")
	if !strings.Contains(rec.Body.String(), "https://") {
		t.Fatalf("javascript:-Link wurde akzeptiert")
	}

	// Gültig anlegen
	rec = do(r, "POST", "/manage/api/events", url.Values{"name": {"Queer Slam"}, "date": {future}, "time": {"20:00"}, "place": {"Werk 2"}, "city": {"Leipzig"}}, "else")
	body := rec.Body.String()
	if !strings.Contains(body, "wurde gespeichert") || !strings.Contains(body, "Queer Slam") || !strings.Contains(body, "20:00 Uhr") {
		t.Fatalf("Termin nicht angelegt:\n%s", body)
	}

	// Bearbeiten
	rec = do(r, "PUT", "/manage/api/events/1", url.Values{"name": {"Queer Slam Leipzig"}, "date": {future}, "time": {""}, "place": {"Werk 2"}, "city": {"Leipzig"}, "link": {"https://example.org"}}, "else")
	if !strings.Contains(rec.Body.String(), "Queer Slam Leipzig") || strings.Contains(rec.Body.String(), "Uhr") {
		t.Fatalf("Änderung nicht übernommen:\n%s", rec.Body.String())
	}

	// Löschen
	if rec = do(r, "DELETE", "/manage/api/events/1", nil, "else"); rec.Code != http.StatusOK {
		t.Fatalf("löschen: HTTP %d", rec.Code)
	}
	if rec = do(r, "GET", "/manage/api/events/1/edit", nil, "else"); rec.Code != http.StatusNotFound {
		t.Errorf("gelöschter Termin noch da: HTTP %d", rec.Code)
	}
}

func TestFormatEventDate(t *testing.T) {
	d := time.Date(2026, 12, 5, 0, 0, 0, 0, time.UTC)
	if got := formatEventDate(d, d); got != "05.12.2026" {
		t.Errorf("ohne Uhrzeit: %q", got)
	}
	if got := formatEventDate(d, d.Add(20*time.Hour)); got != "05.12.2026, 20:00 Uhr" {
		t.Errorf("mit Uhrzeit: %q", got)
	}
}

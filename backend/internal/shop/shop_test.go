package shop

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/shop/paypalfake"
)

func testCatalog() *Catalog {
	return &Catalog{
		Currency:              "EUR",
		ShippingCents:         490,
		FreeShippingFromCents: 5000,
		ShippingCountries:     []string{"DE"},
		Products: []Product{
			{ID: "buch", Name: "Schlicht und Dicht Kunst", PriceCents: 1400, Available: true},
			{ID: "poster", Name: "Poster A2", PriceCents: 2500, Available: true},
			{ID: "ausverkauft", Name: "Sticker", PriceCents: 300, Available: false},
		},
	}
}

type env struct {
	h      *Handler
	fake   *paypalfake.Fake
	router *mux.Router
	db     *sql.DB
}

func setup(t *testing.T) *env {
	t.Helper()
	sqlite, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	sqlite.SetMaxOpenConns(1) // :memory: gilt pro Verbindung
	ddl, err := os.ReadFile("../../schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqlite.Exec(string(ddl)); err != nil {
		t.Fatal(err)
	}

	fake := paypalfake.New("cid", "secret")
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)

	h := &Handler{
		DB:       sqlite,
		Catalog:  StaticCatalog{C: testCatalog()},
		PayPal:   NewPayPal(srv.URL, "cid", "secret"),
		ClientID: "cid",
	}
	r := mux.NewRouter()
	r.HandleFunc("/shop/api/config", h.Config).Methods("GET")
	r.HandleFunc("/shop/api/orders", h.CreateOrder).Methods("POST")
	r.HandleFunc("/shop/api/orders/{id}/capture", h.CaptureOrder).Methods("POST")
	return &env{h: h, fake: fake, router: r, db: sqlite}
}

func (e *env) call(t *testing.T, method, path, body string) (int, map[string]string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	var out map[string]string
	json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestCheckoutHappyPath(t *testing.T) {
	e := setup(t)

	// Der Browser schickt einen Preis mit – der muss ignoriert werden.
	code, res := e.call(t, "POST", "/shop/api/orders", `{"items":[{"id":"buch","quantity":2,"price_cents":1}]}`)
	if code != http.StatusCreated {
		t.Fatalf("create: HTTP %d %v", code, res)
	}
	id := res["id"]

	sent := e.fake.Orders[id]["purchase_units"].([]any)[0].(map[string]any)["amount"].(map[string]any)
	if sent["value"] != "32.90" { // 2 × 14,00 + 4,90 Versand
		t.Errorf("an PayPal gesendeter Betrag = %v, erwartet 32.90", sent["value"])
	}

	code, res = e.call(t, "POST", "/shop/api/orders/"+id+"/capture", "")
	if code != http.StatusOK || res["status"] != "paid" || res["orderNumber"] != "EE-00001" {
		t.Fatalf("capture: HTTP %d %v", code, res)
	}

	o, err := db.New(e.db).GetOrderByPayPalID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != "paid" || o.TotalCents != 3290 || o.ShippingCents != 490 {
		t.Errorf("gespeicherte Bestellung falsch: %+v", o)
	}
	if !strings.Contains(o.ShippingAddress, "Musterstraße 1") || o.PayerEmail != "kaeuferin@example.com" {
		t.Errorf("Adresse/E-Mail nicht gespeichert: %+v", o)
	}

	// Zweiter Capture-Aufruf (z. B. Doppelklick) darf nicht erneut abbuchen.
	code, res = e.call(t, "POST", "/shop/api/orders/"+id+"/capture", "")
	if code != http.StatusOK || res["status"] != "paid" {
		t.Fatalf("zweiter capture: HTTP %d %v", code, res)
	}
	if n := e.fake.Capture[id]; n != 1 {
		t.Errorf("PayPal-Capture %d-mal aufgerufen, erwartet 1", n)
	}
}

func TestFreeShipping(t *testing.T) {
	e := setup(t)
	code, res := e.call(t, "POST", "/shop/api/orders", `{"items":[{"id":"poster","quantity":2}]}`)
	if code != http.StatusCreated {
		t.Fatalf("HTTP %d %v", code, res)
	}
	amount := e.fake.Orders[res["id"]]["purchase_units"].([]any)[0].(map[string]any)["amount"].(map[string]any)
	if amount["value"] != "50.00" {
		t.Errorf("Betrag = %v, erwartet 50.00 (versandkostenfrei ab 50 €)", amount["value"])
	}
}

func TestInvalidCarts(t *testing.T) {
	e := setup(t)
	for name, body := range map[string]string{
		"leer":          `{"items":[]}`,
		"unbekannt":     `{"items":[{"id":"gibtsnicht","quantity":1}]}`,
		"ausverkauft":   `{"items":[{"id":"ausverkauft","quantity":1}]}`,
		"menge 0":       `{"items":[{"id":"buch","quantity":0}]}`,
		"menge zu groß": `{"items":[{"id":"buch","quantity":11}]}`,
		"doppelt":       `{"items":[{"id":"buch","quantity":1},{"id":"buch","quantity":1}]}`,
		"negativ":       `{"items":[{"id":"buch","quantity":-3}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			code, res := e.call(t, "POST", "/shop/api/orders", body)
			if code != http.StatusUnprocessableEntity || res["error"] == "" {
				t.Errorf("HTTP %d %v, erwartet 422 mit Fehlermeldung", code, res)
			}
		})
	}
	if len(e.fake.Orders) != 0 {
		t.Errorf("bei ungültigem Warenkorb wurde trotzdem eine PayPal-Order angelegt")
	}
}

func TestCaptureWrongCountry(t *testing.T) {
	e := setup(t)
	e.fake.Country = "AT"
	_, res := e.call(t, "POST", "/shop/api/orders", `{"items":[{"id":"buch","quantity":1}]}`)
	id := res["id"]
	code, res := e.call(t, "POST", "/shop/api/orders/"+id+"/capture", "")
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("HTTP %d %v, erwartet 422", code, res)
	}
	if e.fake.Capture[id] != 0 {
		t.Errorf("Zahlung wurde trotz falschem Lieferland eingezogen")
	}
}

func TestCaptureUnknownOrder(t *testing.T) {
	e := setup(t)
	code, _ := e.call(t, "POST", "/shop/api/orders/NIX/capture", "")
	if code != http.StatusNotFound {
		t.Errorf("HTTP %d, erwartet 404", code)
	}
}

func TestDisabledShop(t *testing.T) {
	e := setup(t)
	e.h.PayPal = nil
	e.h.ClientID = ""
	code, _ := e.call(t, "POST", "/shop/api/orders", `{"items":[{"id":"buch","quantity":1}]}`)
	if code != http.StatusServiceUnavailable {
		t.Errorf("HTTP %d, erwartet 503", code)
	}
	req := httptest.NewRequest("GET", "/shop/api/config", nil)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), `"enabled":false`) {
		t.Errorf("config: %s", rec.Body.String())
	}
}

func TestCents(t *testing.T) {
	for in, want := range map[string]int64{"14.90": 1490, "5": 500, "0.5": 50, "100.00": 10000} {
		if got, err := parseCents(in); err != nil || got != want {
			t.Errorf("parseCents(%q) = %d, %v; erwartet %d", in, got, err, want)
		}
	}
	if FormatCents(1490) != "14.90" || FormatEuro(505) != "5,05 €" {
		t.Error("FormatCents/FormatEuro falsch")
	}
}

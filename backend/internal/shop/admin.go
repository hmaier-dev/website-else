package shop

// Bestellübersicht für die Verwaltung unter /manage/api/orders.

import (
	"database/sql"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/auth"
	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

//go:embed templates/admin-orders.html
var adminFS embed.FS

var adminTmpl = template.Must(template.New("admin").Funcs(template.FuncMap{
	"euro":        FormatEuro,
	"date":        func(t time.Time) string { return t.Local().Format("02.01.2006, 15:04") },
	"orderNumber": OrderNumber,
	"mul":         func(a, b int64) int64 { return a * b },
}).ParseFS(adminFS, "templates/admin-orders.html"))

type adminOrder struct {
	db.Order
	Lines []OrderLine
	When  time.Time
}

type ordersPanel struct {
	Status string
	Orders []adminOrder
	Open   int
}

func (h *Handler) ordersPanel(r *http.Request, status string) (ordersPanel, error) {
	if status != "shipped" {
		status = "paid"
	}
	q := db.New(h.DB)
	rows, err := q.ListOrdersByStatus(r.Context(), status)
	if err != nil {
		return ordersPanel{}, err
	}
	open, err := q.ListOrdersByStatus(r.Context(), "paid")
	if err != nil {
		return ordersPanel{}, err
	}
	p := ordersPanel{Status: status, Open: len(open)}
	for _, o := range rows {
		a := adminOrder{Order: o, When: o.CreatedAt}
		if o.PaidAt.Valid {
			a.When = o.PaidAt.Time
		}
		json.Unmarshal([]byte(o.Items), &a.Lines)
		p.Orders = append(p.Orders, a)
	}
	return p, nil
}

func (h *Handler) renderOrders(w http.ResponseWriter, r *http.Request, status string) {
	p, err := h.ordersPanel(r, status)
	if err != nil {
		log.Printf("admin: Bestellungen laden: %v", err)
		http.Error(w, "Bestellungen konnten nicht geladen werden", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminTmpl.ExecuteTemplate(w, "orders-panel", p); err != nil {
		log.Printf("admin: template orders-panel: %v", err)
	}
}

// AdminOrders: GET /manage/api/orders?status=paid|shipped
func (h *Handler) AdminOrders(w http.ResponseWriter, r *http.Request) {
	h.renderOrders(w, r, r.URL.Query().Get("status"))
}

// AdminMarkShipped: POST /manage/api/orders/{id}/shipped
func (h *Handler) AdminMarkShipped(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	err = db.New(h.DB).MarkOrderShipped(r.Context(), db.MarkOrderShippedParams{
		ShippedAt: sql.NullTime{Time: h.now(), Valid: true},
		ID:        id,
	})
	if err != nil {
		log.Printf("admin: Bestellung %d versendet: %v", id, err)
		http.Error(w, "Konnte nicht gespeichert werden", http.StatusInternalServerError)
		return
	}
	log.Printf("admin: %s hat %s als versendet markiert", auth.User(r), OrderNumber(id))
	h.renderOrders(w, r, "paid")
}

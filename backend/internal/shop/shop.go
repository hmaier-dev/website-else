// Package shop stellt die Endpunkte für den Checkout mit PayPal bereit.
//
// Ablauf:
//  1. Browser schickt den Warenkorb (nur IDs + Mengen) an POST /shop/api/orders.
//  2. Backend rechnet die Preise mit dem Katalog nach, legt die Bestellung bei
//     PayPal an und speichert sie mit Status "created".
//  3. Kund:in bestätigt im PayPal-Fenster.
//  4. Browser ruft POST /shop/api/orders/{id}/capture auf. Das Backend prüft
//     Lieferland und Betrag, zieht die Zahlung ein und setzt den Status auf "paid".
package shop

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	db "codeberg.org/schnelles_geraet/website-else/backend/internal/database"
)

const (
	maxLines    = 20
	maxQuantity = 10
)

type Handler struct {
	DB       *sql.DB
	Catalog  CatalogSource
	PayPal   *PayPal // nil → Shop deaktiviert
	ClientID string  // öffentliche PayPal-Client-ID für das JS-SDK
	Now      func() time.Time
}

func (h *Handler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

// OrderLine ist eine Position, wie sie in der Datenbank gespeichert wird.
type OrderLine struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

type cartRequest struct {
	Items []struct {
		ID       string `json:"id"`
		Quantity int64  `json:"quantity"`
	} `json:"items"`
}

// --- Hilfsfunktionen ----------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// userError gibt eine Meldung zurück, die im Shop angezeigt werden darf.
func userError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// FormatCents wandelt 1490 in "14.90" (Format der PayPal-API).
func FormatCents(c int64) string {
	return fmt.Sprintf("%d.%02d", c/100, c%100)
}

// FormatEuro wandelt 1490 in "14,90 €" (Anzeige).
func FormatEuro(c int64) string {
	return fmt.Sprintf("%d,%02d €", c/100, c%100)
}

// parseCents wandelt "14.90" in 1490.
func parseCents(v string) (int64, error) {
	whole, frac, _ := strings.Cut(v, ".")
	if len(frac) > 2 {
		return 0, fmt.Errorf("ungültiger Betrag %q", v)
	}
	for len(frac) < 2 {
		frac += "0"
	}
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, err
	}
	return w*100 + f, nil
}

// OrderNumber ist die Bestellnummer, die Kund:innen sehen.
func OrderNumber(id int64) string {
	return fmt.Sprintf("EE-%05d", id)
}

// --- Endpunkte ---------------------------------------------------------------

// Config liefert dem Browser, ob der Shop aktiv ist und welche Client-ID das PayPal-SDK nutzen soll.
func (h *Handler) Config(w http.ResponseWriter, r *http.Request) {
	enabled := h.PayPal != nil && h.ClientID != ""
	currency := "EUR"
	if c, err := h.Catalog.Get(r.Context()); err == nil {
		currency = c.Currency
	} else {
		enabled = false
		log.Printf("shop: Katalog nicht erreichbar: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":  enabled,
		"clientId": h.ClientID,
		"currency": currency,
	})
}

// priceCart rechnet den Warenkorb mit den Katalogpreisen nach.
func priceCart(c *Catalog, req cartRequest) ([]OrderLine, int64, error) {
	if len(req.Items) == 0 {
		return nil, 0, errors.New("Dein Warenkorb ist leer.")
	}
	if len(req.Items) > maxLines {
		return nil, 0, errors.New("Zu viele Positionen im Warenkorb.")
	}
	var lines []OrderLine
	var total int64
	seen := map[string]bool{}
	for _, it := range req.Items {
		p, ok := c.Product(it.ID)
		if !ok {
			return nil, 0, errors.New("Ein Artikel in deinem Warenkorb existiert nicht mehr. Bitte lade die Seite neu.")
		}
		if !p.Available {
			return nil, 0, fmt.Errorf("%s ist leider gerade nicht verfügbar.", p.Name)
		}
		if it.Quantity < 1 || it.Quantity > maxQuantity {
			return nil, 0, fmt.Errorf("Ungültige Menge für %s.", p.Name)
		}
		if seen[p.ID] {
			return nil, 0, errors.New("Ein Artikel ist doppelt im Warenkorb.")
		}
		seen[p.ID] = true
		lines = append(lines, OrderLine{ID: p.ID, Name: p.Name, Quantity: it.Quantity, PriceCents: p.PriceCents})
		total += p.PriceCents * it.Quantity
	}
	return lines, total, nil
}

// CreateOrder legt die Bestellung bei PayPal an.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if h.PayPal == nil {
		userError(w, http.StatusServiceUnavailable, "Der Shop ist gerade nicht verfügbar.")
		return
	}
	ctx := r.Context()

	var req cartRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&req); err != nil {
		userError(w, http.StatusBadRequest, "Ungültige Anfrage.")
		return
	}
	cat, err := h.Catalog.Get(ctx)
	if err != nil {
		log.Printf("shop: Katalog: %v", err)
		userError(w, http.StatusServiceUnavailable, "Der Shop ist gerade nicht verfügbar.")
		return
	}
	lines, itemTotal, err := priceCart(cat, req)
	if err != nil {
		userError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	shipping := cat.ShippingFor(itemTotal)
	total := itemTotal + shipping

	items := make([]ppItem, len(lines))
	for i, l := range lines {
		items[i] = ppItem{
			Name:       l.Name,
			Quantity:   strconv.FormatInt(l.Quantity, 10),
			UnitAmount: Money{CurrencyCode: cat.Currency, Value: FormatCents(l.PriceCents)},
			SKU:        l.ID,
			Category:   "PHYSICAL_GOODS",
		}
	}
	ppOrder := ppCreateOrder{
		Intent: "CAPTURE",
		PurchaseUnits: []ppPurchaseUnit{{
			Description: "Einfach Else Shop",
			Items:       items,
			Amount: ppAmount{
				CurrencyCode: cat.Currency,
				Value:        FormatCents(total),
				Breakdown: ppBreakdown{
					ItemTotal: Money{CurrencyCode: cat.Currency, Value: FormatCents(itemTotal)},
					Shipping:  Money{CurrencyCode: cat.Currency, Value: FormatCents(shipping)},
				},
			},
		}},
		PaymentSource: map[string]any{
			"paypal": map[string]any{
				"experience_context": map[string]any{
					"brand_name":          "Einfach Else",
					"locale":              "de-DE",
					"shipping_preference": "GET_FROM_FILE",
					"user_action":         "PAY_NOW",
				},
			},
		},
	}

	id, err := h.PayPal.CreateOrder(ctx, ppOrder)
	if err != nil {
		log.Printf("shop: PayPal CreateOrder: %v", err)
		userError(w, http.StatusBadGateway, "PayPal ist gerade nicht erreichbar. Bitte versuch es gleich nochmal.")
		return
	}

	linesJSON, _ := json.Marshal(lines)
	err = db.New(h.DB).CreateOrder(ctx, db.CreateOrderParams{
		PaypalOrderID:  id,
		Items:          string(linesJSON),
		ItemTotalCents: itemTotal,
		ShippingCents:  shipping,
		TotalCents:     total,
		Currency:       cat.Currency,
		CreatedAt:      h.now(),
	})
	if err != nil {
		log.Printf("shop: Bestellung %s speichern: %v", id, err)
		userError(w, http.StatusInternalServerError, "Die Bestellung konnte nicht angelegt werden.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// CaptureOrder zieht die Zahlung ein, nachdem Kund:innen bei PayPal bestätigt haben.
func (h *Handler) CaptureOrder(w http.ResponseWriter, r *http.Request) {
	if h.PayPal == nil {
		userError(w, http.StatusServiceUnavailable, "Der Shop ist gerade nicht verfügbar.")
		return
	}
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	q := db.New(h.DB)

	order, err := q.GetOrderByPayPalID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		userError(w, http.StatusNotFound, "Diese Bestellung kennen wir nicht.")
		return
	} else if err != nil {
		log.Printf("shop: Bestellung %s laden: %v", id, err)
		userError(w, http.StatusInternalServerError, "Interner Fehler.")
		return
	}
	// Doppelklick / erneuter Aufruf: einfach Erfolg melden
	if order.Status != "created" {
		writeJSON(w, http.StatusOK, map[string]string{"orderNumber": OrderNumber(order.ID), "status": order.Status})
		return
	}

	// Lieferland vor dem Einziehen prüfen, damit wir nichts abbuchen, was wir nicht liefern können.
	cat, err := h.Catalog.Get(ctx)
	if err != nil {
		log.Printf("shop: Katalog: %v", err)
		userError(w, http.StatusServiceUnavailable, "Der Shop ist gerade nicht verfügbar.")
		return
	}
	ppOrder, err := h.PayPal.GetOrder(ctx, id)
	if err != nil {
		log.Printf("shop: PayPal GetOrder %s: %v", id, err)
		userError(w, http.StatusBadGateway, "PayPal ist gerade nicht erreichbar. Bitte versuch es gleich nochmal.")
		return
	}
	if len(ppOrder.PurchaseUnits) > 0 {
		cc := ppOrder.PurchaseUnits[0].Shipping.Address.CountryCode
		if cc != "" && !cat.ShipsTo(cc) {
			userError(w, http.StatusUnprocessableEntity,
				fmt.Sprintf("Wir liefern leider nur nach: %s. Es wurde nichts abgebucht.", strings.Join(cat.ShippingCountries, ", ")))
			return
		}
	}

	captured, err := h.PayPal.CaptureOrder(ctx, id)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && strings.Contains(apiErr.Body, "INSTRUMENT_DECLINED") {
			userError(w, http.StatusPaymentRequired, "Die Zahlung wurde von PayPal abgelehnt. Bitte wähle eine andere Zahlungsart.")
			return
		}
		log.Printf("shop: PayPal Capture %s: %v", id, err)
		userError(w, http.StatusBadGateway, "Die Zahlung konnte nicht abgeschlossen werden. Es wurde nichts abgebucht.")
		return
	}

	if captured.Status != "COMPLETED" || len(captured.PurchaseUnits) == 0 || len(captured.PurchaseUnits[0].Payments.Captures) == 0 {
		log.Printf("shop: Capture %s unerwarteter Status %q", id, captured.Status)
		userError(w, http.StatusBadGateway, "Die Zahlung ist noch nicht abgeschlossen. Wir melden uns bei dir.")
		return
	}
	pu := captured.PurchaseUnits[0]
	capt := pu.Payments.Captures[0]
	paid, err := parseCents(capt.Amount.Value)
	if err != nil || paid != order.TotalCents || capt.Amount.CurrencyCode != order.Currency {
		// Sollte nie passieren – trotzdem laut loggen, damit es auffällt.
		log.Printf("shop: ACHTUNG Betrag passt nicht! Bestellung %s erwartet %d %s, bezahlt %s %s",
			id, order.TotalCents, order.Currency, capt.Amount.Value, capt.Amount.CurrencyCode)
	}

	name := strings.TrimSpace(pu.Shipping.Name.FullName)
	if name == "" {
		name = strings.TrimSpace(captured.Payer.Name.GivenName + " " + captured.Payer.Name.Surname)
	}
	err = q.MarkOrderPaid(ctx, db.MarkOrderPaidParams{
		PayerName:       name,
		PayerEmail:      captured.Payer.EmailAddress,
		ShippingAddress: formatAddress(name, pu.Shipping.Address),
		CaptureID:       capt.ID,
		PaidAt:          sql.NullTime{Time: h.now(), Valid: true},
		PaypalOrderID:   id,
	})
	if err != nil {
		// Geld ist eingezogen, nur das Speichern hat nicht geklappt → laut loggen.
		log.Printf("shop: ACHTUNG Bestellung %s bezahlt (Capture %s), aber Speichern fehlgeschlagen: %v", id, capt.ID, err)
	}
	log.Printf("shop: Bestellung %s (%s) bezahlt: %s", OrderNumber(order.ID), id, FormatEuro(order.TotalCents))
	writeJSON(w, http.StatusOK, map[string]string{"orderNumber": OrderNumber(order.ID), "status": "paid"})
}

func formatAddress(name string, a Address) string {
	var parts []string
	for _, s := range []string{name, a.Line1, a.Line2, strings.TrimSpace(a.PostalCode + " " + a.City), a.State, a.CountryCode} {
		if s = strings.TrimSpace(s); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n")
}

// FromEnv baut den Handler aus Umgebungsvariablen. Ohne Zugangsdaten bleibt der Shop aus.
func FromEnv(sqlite *sql.DB, getenv func(string) string) *Handler {
	h := &Handler{DB: sqlite}

	catalogURL := getenv("SHOP_CATALOG_URL")
	if catalogURL == "" {
		catalogURL = "http://website/shop/index.json"
	}
	h.Catalog = NewHTTPCatalog(catalogURL)

	id, secret := getenv("PAYPAL_CLIENT_ID"), getenv("PAYPAL_CLIENT_SECRET")
	if id == "" || secret == "" {
		log.Println("shop: PAYPAL_CLIENT_ID/PAYPAL_CLIENT_SECRET nicht gesetzt – Shop-Checkout ist deaktiviert.")
		return h
	}
	base := PayPalSandboxURL
	if strings.ToLower(getenv("PAYPAL_MODE")) == "live" {
		base = PayPalLiveURL
	}
	// Nur für lokale Tests: eigene API-URL (z. B. ein Fake-Server)
	if u := getenv("PAYPAL_API_URL"); u != "" {
		base = u
	}
	h.PayPal = NewPayPal(base, id, secret)
	h.ClientID = id
	log.Printf("shop: PayPal aktiv (%s), Katalog: %s", base, catalogURL)
	return h
}

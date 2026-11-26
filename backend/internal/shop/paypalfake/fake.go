// Package paypalfake simuliert die wenigen PayPal-API-Aufrufe, die der Shop nutzt.
// Nur für Tests und lokale Entwicklung – niemals in Produktion verwenden.
package paypalfake

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Fake struct {
	ClientID, Secret string
	// Lieferland, das die "Käuferin" bei PayPal angibt
	Country string

	mu      sync.Mutex
	nextID  int
	Orders  map[string]map[string]any // Original-Request je Order-ID
	Capture map[string]int            // Anzahl Capture-Aufrufe je Order-ID
}

func New(clientID, secret string) *Fake {
	return &Fake{ClientID: clientID, Secret: secret, Country: "DE", Orders: map[string]map[string]any{}, Capture: map[string]int{}}
}

func (f *Fake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path == "/v1/oauth2/token" {
		id, secret, ok := r.BasicAuth()
		if !ok || id != f.ClientID || secret != f.Secret {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"invalid_client"}`)
			return
		}
		fmt.Fprint(w, `{"access_token":"fake-token","expires_in":3600}`)
		return
	}
	if r.Header.Get("Authorization") != "Bearer fake-token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	rest, ok := strings.CutPrefix(r.URL.Path, "/v2/checkout/orders")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch {
	case rest == "" && r.Method == http.MethodPost:
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		f.nextID++
		id := fmt.Sprintf("FAKE%06d", f.nextID)
		f.Orders[id] = body
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id":%q,"status":"CREATED"}`, id)

	case strings.HasSuffix(rest, "/capture") && r.Method == http.MethodPost:
		id := strings.TrimSuffix(strings.TrimPrefix(rest, "/"), "/capture")
		body, ok := f.Orders[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f.Capture[id]++
		amount := body["purchase_units"].([]any)[0].(map[string]any)["amount"].(map[string]any)
		json.NewEncoder(w).Encode(f.order(id, "COMPLETED", amount["currency_code"].(string), amount["value"].(string)))

	case r.Method == http.MethodGet:
		id := strings.TrimPrefix(rest, "/")
		if _, ok := f.Orders[id]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(f.order(id, "APPROVED", "", ""))

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *Fake) order(id, status, currency, value string) map[string]any {
	pu := map[string]any{
		"shipping": map[string]any{
			"name": map[string]any{"full_name": "Erika Mustermann"},
			"address": map[string]any{
				"address_line_1": "Musterstraße 1",
				"admin_area_2":   "Leipzig",
				"postal_code":    "04109",
				"country_code":   f.Country,
			},
		},
	}
	if status == "COMPLETED" {
		pu["payments"] = map[string]any{"captures": []any{map[string]any{
			"id": "CAP-" + id, "status": "COMPLETED",
			"amount": map[string]any{"currency_code": currency, "value": value},
		}}}
	}
	return map[string]any{
		"id":     id,
		"status": status,
		"payer": map[string]any{
			"name":          map[string]any{"given_name": "Erika", "surname": "Mustermann"},
			"email_address": "kaeuferin@example.com",
		},
		"purchase_units": []any{pu},
	}
}

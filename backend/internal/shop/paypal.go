package shop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// PayPal-Basis-URLs für die REST-API (Orders v2).
const (
	PayPalSandboxURL = "https://api-m.sandbox.paypal.com"
	PayPalLiveURL    = "https://api-m.paypal.com"
)

// PayPal ist ein minimaler Client für die PayPal Orders API v2.
// Es werden nur die drei Aufrufe benötigt, die ein einfacher Checkout braucht:
// Bestellung anlegen, Bestellung abfragen, Zahlung einziehen (capture).
type PayPal struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	HTTP         *http.Client

	mu       sync.Mutex
	token    string
	tokenExp time.Time
}

func NewPayPal(baseURL, clientID, secret string) *PayPal {
	return &PayPal{
		BaseURL:      strings.TrimRight(baseURL, "/"),
		ClientID:     clientID,
		ClientSecret: secret,
		HTTP:         &http.Client{Timeout: 20 * time.Second},
	}
}

// --- Datentypen (nur die Felder, die wir brauchen) ---------------------------

type Money struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type ppItem struct {
	Name       string `json:"name"`
	Quantity   string `json:"quantity"`
	UnitAmount Money  `json:"unit_amount"`
	SKU        string `json:"sku,omitempty"`
	Category   string `json:"category,omitempty"`
}

type ppBreakdown struct {
	ItemTotal Money `json:"item_total"`
	Shipping  Money `json:"shipping"`
}

type ppAmount struct {
	CurrencyCode string      `json:"currency_code"`
	Value        string      `json:"value"`
	Breakdown    ppBreakdown `json:"breakdown"`
}

type ppPurchaseUnit struct {
	Description string   `json:"description,omitempty"`
	Items       []ppItem `json:"items"`
	Amount      ppAmount `json:"amount"`
}

type ppCreateOrder struct {
	Intent        string           `json:"intent"`
	PurchaseUnits []ppPurchaseUnit `json:"purchase_units"`
	PaymentSource map[string]any   `json:"payment_source,omitempty"`
}

// Order ist die Antwort von PayPal auf GET/Capture.
type Order struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Payer  struct {
		Name struct {
			GivenName string `json:"given_name"`
			Surname   string `json:"surname"`
		} `json:"name"`
		EmailAddress string `json:"email_address"`
	} `json:"payer"`
	PurchaseUnits []struct {
		Shipping struct {
			Name struct {
				FullName string `json:"full_name"`
			} `json:"name"`
			Address Address `json:"address"`
		} `json:"shipping"`
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount Money  `json:"amount"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

type Address struct {
	Line1       string `json:"address_line_1"`
	Line2       string `json:"address_line_2"`
	City        string `json:"admin_area_2"`
	State       string `json:"admin_area_1"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// APIError enthält die Fehlermeldung von PayPal.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("paypal: HTTP %d: %s", e.Status, e.Body)
}

// --- Aufrufe -----------------------------------------------------------------

// accessToken holt (und cached) ein OAuth2-Token per Client-Credentials.
func (p *PayPal) accessToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.token != "" && time.Now().Before(p.tokenExp) {
		return p.token, nil
	}

	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/v1/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(p.ClientID, p.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", &APIError{Status: resp.StatusCode, Body: string(body)}
	}

	var t struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &t); err != nil {
		return "", err
	}
	p.token = t.AccessToken
	// eine Minute Puffer, damit das Token nicht mitten im Aufruf abläuft
	p.tokenExp = time.Now().Add(time.Duration(t.ExpiresIn)*time.Second - time.Minute)
	return p.token, nil
}

func (p *PayPal) do(ctx context.Context, method, path string, payload any, requestID string, out any) error {
	token, err := p.accessToken(ctx)
	if err != nil {
		return err
	}

	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		// Idempotenz: wiederholte Aufrufe mit derselben ID werden nur einmal ausgeführt
		req.Header.Set("PayPal-Request-Id", requestID)
	}

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &APIError{Status: resp.StatusCode, Body: string(respBody)}
	}
	if out != nil {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func (p *PayPal) CreateOrder(ctx context.Context, order ppCreateOrder) (string, error) {
	var res struct {
		ID string `json:"id"`
	}
	if err := p.do(ctx, http.MethodPost, "/v2/checkout/orders", order, "", &res); err != nil {
		return "", err
	}
	if res.ID == "" {
		return "", fmt.Errorf("paypal: keine Order-ID erhalten")
	}
	return res.ID, nil
}

func (p *PayPal) GetOrder(ctx context.Context, id string) (*Order, error) {
	var o Order
	if err := p.do(ctx, http.MethodGet, "/v2/checkout/orders/"+url.PathEscape(id), nil, "", &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (p *PayPal) CaptureOrder(ctx context.Context, id string) (*Order, error) {
	var o Order
	if err := p.do(ctx, http.MethodPost, "/v2/checkout/orders/"+url.PathEscape(id)+"/capture", struct{}{}, "capture-"+id, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

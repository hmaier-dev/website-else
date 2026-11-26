package shop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Catalog ist das, was Hugo aus content/shop.md als /shop/index.json erzeugt.
// Die Preise kommen ausschließlich von hier, niemals aus dem Browser.
type Catalog struct {
	Currency              string    `json:"currency"`
	ShippingCents         int64     `json:"shipping_cents"`
	FreeShippingFromCents int64     `json:"free_shipping_from_cents"`
	ShippingCountries     []string  `json:"shipping_countries"`
	Products              []Product `json:"products"`
}

type Product struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Available  bool   `json:"available"`
}

func (c *Catalog) Product(id string) (Product, bool) {
	for _, p := range c.Products {
		if p.ID == id {
			return p, true
		}
	}
	return Product{}, false
}

// ShippingFor berechnet die Versandkosten für eine Warenkorbsumme.
func (c *Catalog) ShippingFor(itemTotal int64) int64 {
	if c.FreeShippingFromCents > 0 && itemTotal >= c.FreeShippingFromCents {
		return 0
	}
	return c.ShippingCents
}

func (c *Catalog) ShipsTo(country string) bool {
	if len(c.ShippingCountries) == 0 {
		return true
	}
	for _, cc := range c.ShippingCountries {
		if cc == country {
			return true
		}
	}
	return false
}

func (c *Catalog) validate() error {
	if c.Currency == "" {
		return fmt.Errorf("katalog: currency fehlt")
	}
	seen := map[string]bool{}
	for _, p := range c.Products {
		if p.ID == "" || p.Name == "" {
			return fmt.Errorf("katalog: Produkt ohne id oder name")
		}
		if p.PriceCents <= 0 {
			return fmt.Errorf("katalog: Produkt %q hat keinen gültigen Preis", p.ID)
		}
		if seen[p.ID] {
			return fmt.Errorf("katalog: Produkt-ID %q doppelt", p.ID)
		}
		seen[p.ID] = true
	}
	return nil
}

// CatalogSource liefert den aktuellen Katalog.
type CatalogSource interface {
	Get(ctx context.Context) (*Catalog, error)
}

// HTTPCatalog lädt den Katalog per HTTP vom Website-Container und cached ihn kurz.
// So wirkt eine Preisänderung in shop.md nach dem nächsten Deploy automatisch.
type HTTPCatalog struct {
	URL    string
	TTL    time.Duration
	Client *http.Client

	mu        sync.Mutex
	cached    *Catalog
	fetchedAt time.Time
}

func NewHTTPCatalog(url string) *HTTPCatalog {
	return &HTTPCatalog{URL: url, TTL: 5 * time.Minute, Client: &http.Client{Timeout: 10 * time.Second}}
}

func (h *HTTPCatalog) Get(ctx context.Context) (*Catalog, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cached != nil && time.Since(h.fetchedAt) < h.TTL {
		return h.cached, nil
	}

	c, err := h.fetch(ctx)
	if err != nil {
		// Lieber einen etwas älteren Katalog verwenden als gar keinen.
		if h.cached != nil {
			return h.cached, nil
		}
		return nil, err
	}
	h.cached, h.fetchedAt = c, time.Now()
	return c, nil
}

func (h *HTTPCatalog) fetch(ctx context.Context) (*Catalog, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("katalog: %s antwortet mit HTTP %d", h.URL, resp.StatusCode)
	}
	var c Catalog
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&c); err != nil {
		return nil, fmt.Errorf("katalog: ungültiges JSON: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// StaticCatalog ist für Tests.
type StaticCatalog struct{ C *Catalog }

func (s StaticCatalog) Get(context.Context) (*Catalog, error) { return s.C, nil }

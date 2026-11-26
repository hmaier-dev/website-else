package events

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMarkUpcoming(t *testing.T) {
	now := time.Date(2026, 11, 26, 15, 0, 0, 0, time.UTC)
	day := func(d int) Event { return Event{Date: now.AddDate(0, 0, d).Truncate(24 * time.Hour)} }
	which := func(ev []Event) int {
		for i, e := range ev {
			if e.Upcomming != nil {
				return i
			}
		}
		return -1
	}
	cases := []struct {
		name   string
		sort   string
		events []Event
		want   int
	}{
		{"nur ein Termin (stürzte früher ab)", "next", []Event{day(3)}, 0},
		{"leer", "next", nil, -1},
		{"asc: erster ab heute", "asc", []Event{day(-5), day(-1), day(0), day(4)}, 2},
		{"desc: letzter ab heute", "desc", []Event{day(9), day(4), day(-2)}, 1},
		{"desc: nur vergangene", "desc", []Event{day(-2), day(-9)}, -1},
		{"asc: nur ein vergangener", "asc", []Event{day(-2)}, -1},
	}
	for _, c := range cases {
		markUpcoming(c.events, c.sort, now)
		if got := which(c.events); got != c.want {
			t.Errorf("%s: hervorgehoben %d, erwartet %d", c.name, got, c.want)
		}
	}
}

// Ende-zu-Ende über die öffentliche API: genau ein kommender Termin, HTML mit Hervorhebung.
func TestPublicLoadSingleEvent(t *testing.T) {
	r := adminRouter(t)
	future := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	do(r, "POST", "/manage/api/events", url.Values{"name": {"Einziger Termin"}, "date": {future}, "place": {"Werk 2"}, "city": {"Leipzig"}}, "else")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/events/load?sort=next&out=html&sep=year&upcomming=true", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Einziger Termin") || !strings.Contains(rec.Body.String(), "Als nächstes") {
		t.Fatalf("HTTP %d:\n%s", rec.Code, rec.Body.String())
	}
}

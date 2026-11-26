package events

import (
	"net/url"
	"regexp"
	"strings"
)

// Das Link-Feld eines Termins enthält in der Praxis nicht nur Links, sondern auch
// Hinweise wie "Eintritt: Spendenbasis" oder "Anmeldung per Mail" (Altdaten aus der
// früheren Tabellen-Verwaltung, teils auch wörtlich "null"). splitLink trennt das sauber.

const maxNoteLen = 120

// sieht aus wie "schema:etwas" ohne Leerzeichen nach dem Doppelpunkt (javascript:, mailto:, ftp:, …)
var schemeLike = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:\S`)

// splitLink liefert entweder einen gültigen http(s)-Link oder einen Hinweistext (oder beides leer).
func splitLink(raw string) (link, note string) {
	s := strings.TrimSpace(raw)
	if s == "" || strings.EqualFold(s, "null") {
		return "", ""
	}
	if u, err := url.Parse(s); err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return s, ""
	}
	if schemeLike.MatchString(s) {
		return "", "" // andere Schemata (z. B. javascript:) nie anzeigen
	}
	return "", s
}

// validLinkInput prüft eine Eingabe im Formular. Erlaubt: http(s)-Link oder kurzer Hinweis.
func validLinkInput(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	link, note := splitLink(s)
	switch {
	case link != "":
		return ""
	case note == "":
		return "Links müssen mit https:// beginnen."
	case len([]rune(note)) > maxNoteLen:
		return "Der Hinweis ist zu lang (höchstens 120 Zeichen)."
	}
	return ""
}

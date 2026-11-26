// Package auth schützt die Verwaltungs-Endpunkte.
//
// Die eigentliche Anmeldung macht Authentik (ForwardAuth in Traefik). Authentik
// schickt danach den Benutzernamen im Header X-authentik-username mit.
// Das Backend lässt Anfragen an /manage/api nur durch, wenn dieser Header
// gesetzt ist. Wird Traefik/Authentik falsch konfiguriert, bleibt die
// Verwaltung also zu, statt offen zu sein.
package auth

import (
	"log"
	"net/http"
)

// DefaultHeader setzt Authentik nach erfolgreicher Anmeldung.
const DefaultHeader = "X-authentik-username"

// Header ist der Header mit dem Benutzernamen. Leer = keine Prüfung (nur für lokale Entwicklung!).
var Header = DefaultHeader

// User gibt den angemeldeten Benutzer zurück (für Log-Ausgaben).
func User(r *http.Request) string {
	if Header != "" {
		if u := r.Header.Get(Header); u != "" {
			return u
		}
	}
	return "unbekannt"
}

// Require lässt nur Anfragen mit gesetztem Benutzer-Header durch.
func Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Header != "" && r.Header.Get(Header) == "" {
			log.Printf("auth: Anfrage an %s ohne %s abgelehnt", r.URL.Path, Header)
			http.Error(w, "Nicht angemeldet", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

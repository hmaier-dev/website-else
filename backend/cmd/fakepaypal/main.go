// fakepaypal simuliert die PayPal-API für lokale Tests des Shops.
//
//	go run ./cmd/fakepaypal            # lauscht auf :9999
//	PAYPAL_CLIENT_ID=test PAYPAL_CLIENT_SECRET=test PAYPAL_API_URL=http://localhost:9999 ./backend
//
// Achtung: Der Bezahl-Button im Browser kommt trotzdem von paypal.com und
// braucht dafür eine echte Sandbox-Client-ID. Mit diesem Fake lassen sich die
// Backend-Endpunkte per curl testen.
package main

import (
	"flag"
	"log"
	"net/http"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/shop/paypalfake"
)

func main() {
	addr := flag.String("addr", ":9999", "Adresse")
	id := flag.String("client-id", "test", "erwartete Client-ID")
	secret := flag.String("client-secret", "test", "erwartetes Secret")
	country := flag.String("country", "DE", "Lieferland der Test-Käuferin")
	flag.Parse()

	f := paypalfake.New(*id, *secret)
	f.Country = *country
	log.Printf("Fake-PayPal auf %s (Client-ID %q)", *addr, *id)
	log.Fatal(http.ListenAndServe(*addr, f))
}

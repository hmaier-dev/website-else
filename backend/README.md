While the main part of the website is static html, for receiving booking-request we need an HTTP-endpoint.
This is the purpose this small go-program serves.

Start the binary:
```bash
go build -x -v -p 4 -o ./bin/backend main.go && ./bin/backend -db=sqlite.db
```
And test the endpoints:

## Mailbox
```bash
## POST to /contact
## This inserts a new entry into the database
curl -d 'email=test@mail.xyz&name=Mensch&message=Hello to Golang&approval=true' localhost:3000/contact
## GET to /mailbox/all
## gets all messages from database
curl localhost:3000/mailbox/all
```

## Events
```bash
## Öffentlich: die nächsten Termine
curl "localhost:3000/events/load?sort=next&head=3&out=json" | jq '.'
curl "localhost:3000/events/load?sort=desc&sep=year&out=html"
```
Termine werden über die Verwaltungsseite `/manage` gepflegt (siehe unten).

## Verwaltung (`/manage`)
Die Seite `/manage` (Hugo, `layouts/manage.html`) lädt ihre Inhalte per htmx vom Backend unter `/manage/api/...`:

| Bereich | Endpunkte |
|---|---|
| Termine | `GET/POST /manage/api/events`, `GET/PUT/DELETE /manage/api/events/{id}`, `GET /manage/api/events/{id}/edit`, `POST /manage/api/events/bulk` (YAML-Upload) |
| Bestellungen | `GET /manage/api/orders?status=paid\|shipped`, `POST /manage/api/orders/{id}/shipped` |
| Nachrichten | `GET /manage/api/messages?view=inbox\|archived`, `GET /manage/api/messages/badge`, `POST /manage/api/messages/{id}/{read\|unread\|archive\|unarchive}`, `DELETE /manage/api/messages/{id}` |

**Anmeldung:** Traefik schützt `/manage` und `/manage/api` per ForwardAuth mit Authentik (siehe `compose.example.yml`).
Zusätzlich lässt das Backend Anfragen an `/manage/api` nur durch, wenn der Header `X-authentik-username` gesetzt ist
(den setzt Authentik nach dem Login). Fehlt die Middleware in Traefik, bleibt die Verwaltung also zu (HTTP 401) statt offen zu sein.
Wichtig: In der Authentik-Middleware `authResponseHeaders=X-authentik-username,...` setzen, damit Traefik den Header weiterreicht.

Lokal ohne Authentik:
```bash
MANAGE_AUTH_HEADER= ./bin/backend        # Prüfung aus (nur lokal!)
curl -H "X-authentik-username: else" localhost:3000/manage/api/events   # oder Header mitschicken
```

## Kontaktformular
- Anfragen landen in der Tabelle `mailbox` und erscheinen in `/manage` → „Nachrichten“
  (Posteingang / Erledigt, gelesen/ungelesen, löschen). Der Reiter zeigt die Zahl der ungelesenen Nachrichten.
- Das Backend prüft die Eingaben (Pflichtfelder, gültige E-Mail, max. 5000 Zeichen).
- Spam-Falle: Das unsichtbare Feld `website` füllen nur Bots aus. Solche Anfragen werden still verworfen.
- Nachrichteninhalte werden nicht mehr ins Log geschrieben.

## E-Mail-Benachrichtigung (optional)
Bei jeder neuen Anfrage und jeder bezahlten Bestellung kann das Backend eine Mail verschicken.
Bei Anfragen ist `Reply-To` die Absenderin – „Antworten“ im Mailprogramm schreibt ihr also direkt.
Ohne `SMTP_HOST` ist die Funktion aus.

| Variable | Beispiel | |
|---|---|---|
| `SMTP_HOST` | `smtp.example.org` | Pflicht, sonst keine Mails |
| `SMTP_PORT` | `587` | 587 = STARTTLS (Standard), 465 = TLS |
| `SMTP_USERNAME` | `website@elisa-adam.com` | Login beim Mailanbieter |
| `SMTP_PASSWORD` | | Passwort bzw. App-Passwort, **nicht ins Git** |
| `SMTP_FROM` | `website@elisa-adam.com` | Absender (Standard: `SMTP_USERNAME`) |
| `NOTIFY_EMAIL` | `else@example.org, hendrik@example.org` | Empfänger, mit Komma getrennt |
| `SITE_URL` | `https://elisa-adam.com` | für den Link zur Verwaltung in der Mail |

Ohne Verschlüsselung wird nichts verschickt (`SMTP_INSECURE=true` nur für lokale Tests mit z. B. Mailpit).
Schlägt der Versand fehl, steht das im Log; die Nachricht selbst ist trotzdem gespeichert.

## Shop mit PayPal
Die Produkte stehen in `content/shop.md`. Hugo erzeugt daraus zusätzlich `/shop/index.json`.
Das Backend lädt diesen Katalog (`SHOP_CATALOG_URL`) und rechnet **jeden Preis selbst nach** – der Browser schickt nur Produkt-IDs und Mengen.

Ablauf einer Bestellung:
1. Browser → `POST /shop/api/orders` mit `{"items":[{"id":"schlicht-und-dicht","quantity":2}]}`
2. Backend prüft den Warenkorb, legt die Bestellung bei PayPal an und speichert sie in der Tabelle `orders` (Status `created`).
3. Kund:in bestätigt im PayPal-Fenster (Lieferadresse kommt von PayPal).
4. Browser → `POST /shop/api/orders/{id}/capture`. Das Backend prüft das Lieferland, zieht das Geld ein und setzt den Status auf `paid`.
5. In `/manage` → „Bestellungen“ erscheinen Adresse und Artikel. Nach dem Verschicken „Als versendet markieren“.
   PayPal schickt außerdem bei jeder Zahlung eine E-Mail an das Händlerkonto.

### PayPal einrichten (einmalig)
1. Auf <https://developer.paypal.com> mit dem PayPal-**Geschäftskonto** anmelden.
2. „Apps & Credentials“ → oben **Sandbox** wählen → „Create App“. Client-ID und Secret kopieren.
3. Unter „Sandbox accounts“ gibt es fertige Test-Käufer:innen (E-Mail + Passwort) zum Ausprobieren.
4. Im Backend-Container setzen:
   ```
   PAYPAL_MODE=sandbox
   PAYPAL_CLIENT_ID=...
   PAYPAL_CLIENT_SECRET=...
   ```
5. In `content/shop.md` vorübergehend `draft: false` (z. B. auf dem preview-Branch) und eine Testbestellung mit dem Sandbox-Käuferkonto machen.
6. Läuft alles: dasselbe unter „Live“ wiederholen, `PAYPAL_MODE=live` und die Live-Zugangsdaten setzen.

Das Secret gehört **nur** in die Umgebung des Backend-Containers (z. B. `.env` neben der Compose-Datei, nicht ins Git).
Ohne Zugangsdaten startet das Backend trotzdem, der Shop zeigt dann „Bezahlen ist gerade nicht möglich“.

### Lokal testen ohne PayPal-Konto
```bash
go run ./cmd/fakepaypal &      # simuliert die PayPal-API auf :9999
PAYPAL_CLIENT_ID=test PAYPAL_CLIENT_SECRET=test PAYPAL_API_URL=http://localhost:9999 \
  SHOP_CATALOG_URL=http://localhost:1313/shop/index.json go run .
# in einem anderen Terminal: hugo server -D  (Shop ist noch Entwurf)
curl -s -XPOST localhost:3000/shop/api/orders -d '{"items":[{"id":"schlicht-und-dicht","quantity":1}]}'
curl -s -XPOST localhost:3000/shop/api/orders/FAKE000001/capture
```
Der Bezahl-Button im Browser kommt immer von paypal.com und braucht dafür eine echte Sandbox-Client-ID.

### Datenbank-Code
Die Queries stehen in `query.sql`, das Schema in `schema.sql`. Nach Änderungen:
```bash
sqlc generate
go test ./...
```

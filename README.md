## Develop locally
Benötigt **Hugo ≥ 0.158** (getestet mit 0.158, 0.160 und 0.170; das Docker-Image nutzt 0.160.1).
Alles zum Entwickeln findet man in der `shell.nix`. Damit begibt man sich in eine passende Entwicklungsumgebung.

### Frontend Only
If you want to develop the website further, do it locally. You can start a server with hugo:
```bash
hugo server --bind 0.0.0.0 --baseURL http://$(ip addr | grep -w "inet 192.168" | awk '{print $2}' | cut -d/ -f1):1313/
```
Hugo generate a website which reloads itself on changes.
### Backend Only
To run the HTTP-Endpoints for the backend, just build the binary and execute it.
```bash
cd ./backend
go build -x -v -p 4 -o ./bin/backend main.go && ./bin/backend -db=sqlite.db
```
### Run both together locally
If you want try out both parts together, you can run a traefik-instance which mimics the server-setup:
```bash
docker compose -f compose.example.yml up
docker compose -f compose.example.yml down
```
Make sure you build current images:
```bash
## ./
earthly +run --tag="local"
## ./backend
cd ./backend
earthly +run --tag="v2"
```

## Publish a new image
If you do any changes to the hugo project, run the following command:
```bash
earthly --push +run --tag="latest"
```
Before running this check which `baseURL` is set!
```toml
# For prod usage
baseURL = "https://elisa-adam.com/"
# For testing usage
# baseURL = "http://localhost/"
```
If you have a valid auth-token for the container registry, the `--push`-flag will do push the image upstream.
Beware: no tag is set and the new image will overwrite the old on.



## Bilder: Fotograf:innen und Beschreibungen
Alle Angaben zu Fotos stehen zentral in `data/bilder.yaml`. Pro Foto gibt es einen Eintrag:
```yaml
- datei: media/kreis4.webp          # Pfad unter assets/, wie im Inhalt
  fotograf: "Name"                  # klein in der Bildecke + im Impressum unter "Bildnachweise"
  link: "https://…"                 # optional
  beschreibung: "Was zu sehen ist"  # Alternativtext; bei Angeboten/Über mich auch als Bildunterschrift
```
Das gilt automatisch überall, wo das Foto vorkommt. Fehlt bei einem genutzten Foto der Fotograf, weist `hugo` beim Bauen darauf hin.
Einzelne Stellen lassen sich überschreiben: in Shortcodes mit `fotograf="…" fotolink="…" beschreibung="…"`,
im Shop pro Produkt mit `fotograf:` und `bildbeschreibung:`.

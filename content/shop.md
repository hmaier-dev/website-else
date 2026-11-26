---
title: Shop
description: Bücher, Merch und Kleinkram von Einfach Else – direkt nach Hause geliefert.
kicker: Merch & Bücher
layout: shop

# ⚠️ Der Shop ist noch nicht öffentlich. Erst auf "false" setzen, wenn:
#   1. PayPal-Zugangsdaten im Backend gesetzt sind (siehe backend/README.md)
#   2. content/agb.md und content/widerruf.md echte Texte enthalten (und dort draft: false)
#   3. die Datenschutzerklärung einen Abschnitt zu PayPal hat
#   4. die Beispielprodukte unten durch echte ersetzt sind
draft: true

# Erzeugt zusätzlich /shop/index.json – daraus liest das Backend die Preise.
outputs: [html, json]

menus:
  main:
    name: Shop
    weight: 35

# ---------------------------------------------------------------------------
# Einstellungen
# ---------------------------------------------------------------------------
shop:
  waehrung: EUR
  versandkosten: 4.90          # pro Bestellung
  versandkostenfrei_ab: 50     # 0 = nie versandkostenfrei
  lieferlaender: [DE]          # ISO-Ländercodes, z. B. [DE, AT, CH]
  lieferzeit: "Lieferzeit 3–5 Werktage innerhalb Deutschlands."
  # Kleinunternehmerin? Dann diesen Satz behalten. Sonst z. B. "Alle Preise inkl. MwSt."
  steuerhinweis: "Gemäß § 19 UStG wird keine Umsatzsteuer berechnet."

# ---------------------------------------------------------------------------
# Produkte
#   id:           eindeutig, nur a-z, 0-9 und Bindestrich. Nach dem ersten Verkauf nicht mehr ändern!
#   name:         wird im Shop, bei PayPal und in der Bestellübersicht angezeigt
#   preis:        in Euro, mit Punkt: 14.00
#   preis_min /   Preis frei wählbar zwischen preis_min und preis_max (z. B. Soli-Preis).
#   preis_max:    "preis" ist dann der Vorschlag, mit dem das Feld startet (optional).
#   bild:         Pfad unter assets/
#   verfuegbar:   false → "Ausverkauft", kann nicht bestellt werden
#   beschreibung: Markdown erlaubt
# ---------------------------------------------------------------------------
produkte:
  - id: schlicht-und-dicht
    name: "Schlicht und Dicht Kunst (Buch)"
    preis: 14.00
    bild: media/book-poetry.webp
    verfuegbar: true
    beschreibung: |
      **Beispielprodukt** – Texte und Gedichte von Einfach Else, erschienen 2022.
      Auf Wunsch mit Widmung: einfach nach der Bestellung eine Nachricht schreiben.

  - id: postkarten-set
    name: "Postkarten-Set (5 Stück)"
    preis_min: 5
    preis_max: 20
    preis: 8.50
    bild: media/lachen1.webp
    verfuegbar: true
    beschreibung: |
      **Beispielprodukt mit frei wählbarem Preis** – fünf Postkarten mit Zitaten aus Elses Texten.
      Zahl, was es dir wert ist.

  - id: jutebeutel
    name: "Jutebeutel „Einfach Else“"
    preis: 12.00
    bild: media/spread-the-word-loud.webp
    verfuegbar: false
    beschreibung: |
      **Beispielprodukt** – zeigt, wie ein ausverkaufter Artikel aussieht.
---

Hier findest du Bücher und Merch direkt von mir. Jede Bestellung packe ich selbst ein. ♥

CREATE TABLE IF NOT EXISTS mailbox (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  message TEXT NOT NULL,
  date INT NOT NULL,
  approval INT DEFAULT 0,
  IsRead INT,
  -- 1 = erledigt (in der Verwaltung unter "Erledigt")
  -- Bei bestehenden Datenbanken legt main.go die Spalte per Migration an.
  archived INT NOT NULL DEFAULT 0
);


CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  date DATE NOT NULL,
  doors DATETIME,
  begin DATETIME,
  place TEXT NOT NULL,
  city TEXT NOT NULL,
  coordinates TEXT,
  link TEXT DEFAULT '',
  pic BLOB,
  hash TEXT NOT NULL
);

-- Bestellungen aus dem Shop (PayPal)
-- Beträge immer in Cent, um Rundungsfehler zu vermeiden.
CREATE TABLE IF NOT EXISTS orders (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  paypal_order_id TEXT NOT NULL UNIQUE,
  -- created: bei PayPal angelegt, noch nicht bezahlt
  -- paid: bezahlt (Capture erfolgreich)
  -- shipped: versendet
  status TEXT NOT NULL DEFAULT 'created',
  items TEXT NOT NULL,            -- JSON-Liste der Positionen
  item_total_cents INTEGER NOT NULL,
  shipping_cents INTEGER NOT NULL,
  total_cents INTEGER NOT NULL,
  currency TEXT NOT NULL,
  payer_name TEXT NOT NULL DEFAULT '',
  payer_email TEXT NOT NULL DEFAULT '',
  shipping_address TEXT NOT NULL DEFAULT '',
  capture_id TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  paid_at DATETIME,
  shipped_at DATETIME
);

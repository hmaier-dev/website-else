package migrate

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Eine "alte" Datenbank ohne archived-Spalte bekommt sie, Daten bleiben erhalten.
func TestAddsMissingColumn(t *testing.T) {
	ctx := context.Background()
	db, _ := sql.Open("sqlite3", ":memory:")
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE mailbox (id INTEGER PRIMARY KEY, name TEXT, IsRead INT)`)
	db.Exec(`INSERT INTO mailbox (name) VALUES ('Max')`)

	for i := 0; i < 2; i++ { // zweimal: muss idempotent sein
		if err := Run(ctx, db); err != nil {
			t.Fatal(err)
		}
	}
	var name string
	var archived int
	if err := db.QueryRow(`SELECT name, archived FROM mailbox`).Scan(&name, &archived); err != nil {
		t.Fatal(err)
	}
	if name != "Max" || archived != 0 {
		t.Errorf("got %q %d", name, archived)
	}
}

// Package migrate ergänzt Spalten in bestehenden Datenbanken.
//
// schema.sql legt Tabellen nur an, wenn es sie noch nicht gibt
// (CREATE TABLE IF NOT EXISTS). Neue Spalten in bereits vorhandenen
// Tabellen kommen deshalb hierher.
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type column struct {
	Table, Name, Definition string
}

// Neue Spalten immer unten anhängen.
var columns = []column{
	{"mailbox", "archived", "INT NOT NULL DEFAULT 0"},
}

// Run legt fehlende Spalten an. Mehrfaches Ausführen ist unbedenklich.
func Run(ctx context.Context, db *sql.DB) error {
	for _, c := range columns {
		ok, err := hasColumn(ctx, db, c.Table, c.Name)
		if err != nil {
			return err
		}
		if ok {
			continue
		}
		stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", c.Table, c.Name, c.Definition)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration %s.%s: %w", c.Table, c.Name, err)
		}
		log.Printf("migrate: Spalte %s.%s angelegt", c.Table, c.Name)
	}
	return nil
}

func hasColumn(ctx context.Context, db *sql.DB, table, name string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid          int
			colName, typ string
			notNull, pk  int
			dflt         sql.NullString
		)
		if err := rows.Scan(&cid, &colName, &typ, &notNull, &dflt, &pk); err != nil {
			return false, err
		}
		if colName == name {
			return true, nil
		}
	}
	return false, rows.Err()
}

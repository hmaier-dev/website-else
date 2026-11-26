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

// Datenbereinigungen, die bei jedem Start laufen dürfen (idempotent).
var dataFixes = []struct{ Name, SQL string }{
	// Altdaten aus der früheren Tabellen-Verwaltung: wörtlich "null" oder nur Leerzeichen als Link
	{"Termine: leere/\"null\"-Links entfernt", `UPDATE events SET link = NULL WHERE link IS NOT NULL AND (trim(link) = '' OR lower(trim(link)) = 'null')`},
	{"Termine: Leerzeichen um Links entfernt", `UPDATE events SET link = trim(link) WHERE link IS NOT NULL AND link <> trim(link)`},
}

// Run legt fehlende Spalten an und bereinigt Altdaten. Mehrfaches Ausführen ist unbedenklich.
func Run(ctx context.Context, db *sql.DB) error {
	defer func() {
		for _, f := range dataFixes {
			res, err := db.ExecContext(ctx, f.SQL)
			if err != nil {
				log.Printf("migrate: %s fehlgeschlagen: %v", f.Name, err)
				continue
			}
			if n, _ := res.RowsAffected(); n > 0 {
				log.Printf("migrate: %s (%d)", f.Name, n)
			}
		}
	}()
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

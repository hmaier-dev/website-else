package main

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"context"

	"codeberg.org/schnelles_geraet/website-else/backend/internal/server"

	_ "codeberg.org/schnelles_geraet/website-else/backend/internal/mailbox"
	_ "codeberg.org/schnelles_geraet/website-else/backend/internal/events"

	// For future me:
	// Use 'go get -u github.com/mattn/go-sqlite3' to download it
	_ "github.com/mattn/go-sqlite3"
	_ "embed"
)

//go:embed schema.sql
var ddl string
const port = "3000"

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ctx := context.Background()

	sqlite, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	err = sqlite.Ping()
	if err != nil{
		log.Fatal(err)
	}
	if _, err := sqlite.ExecContext(ctx, ddl); err != nil {
		log.Fatal(err)
	}

	if _, err := sqlite.ExecContext(ctx, ddl); err != nil {
		log.Fatal(err)
	}
	srv := server.New(sqlite)

	host := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Starting tool on %s \n", host)
	err = http.ListenAndServe(host, srv.Router)
	if err != nil {
		log.Fatal("cannot listen and server", err)
	}
}



package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

type Migration struct {
	Version int
	UpSQL   string
	DownSQL string
}

/*
prepareTablesContext create tables for metrice storage if not exist

Args:

	ctx context.Context
	db *sql.DB: pointer to sql.DB instance

Returns:

	error: nil or an error that occurred while processing the request
*/
func prepareTablesContext(ctx context.Context, db *sql.DB) error {
	files, _ := migrationFiles.ReadDir("migrations")
	for _, file := range files {
		content, _ := migrationFiles.ReadFile("migrations/" + file.Name())
		// fmt.Println("File:migrations/" + file.Name())
		// fmt.Println("SQL>>>>>>>>>>>>>>>>>")
		// fmt.Println(string(content))
		// fmt.Println(">>>>>>>>>>>>>>>>>SQL")
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("preapring database error: %w", err)
		}
	}
	return nil
}

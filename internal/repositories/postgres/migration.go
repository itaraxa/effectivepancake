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
	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migration files error: %w", err)
	}
	for _, file := range files {
		content, err := migrationFiles.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("reading migration file %s error: %w", file.Name(), err)
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("preapring database error: %w", err)
		}
	}
	return nil
}

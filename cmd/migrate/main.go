package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://cry076:cry076@localhost:5432/cry076?sslmode=disable"
	}
	directory := "migrations"
	if len(os.Args) > 1 {
		directory = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(name text PRIMARY KEY,applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		panic(err)
	}
	entries, err := filepath.Glob(filepath.Join(directory, "*.sql"))
	if err != nil {
		panic(err)
	}
	sort.Strings(entries)
	for _, path := range entries {
		name := filepath.Base(path)
		var applied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name=$1)`, name).Scan(&applied); err != nil {
			panic(err)
		}
		if applied {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			panic(err)
		}
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			panic(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(name) VALUES($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			panic(err)
		}
		if err := tx.Commit(ctx); err != nil {
			panic(err)
		}
		fmt.Printf("applied %s\n", name)
	}
}

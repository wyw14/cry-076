package tests

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/wyw14/cry-076/internal/domain"
	"github.com/wyw14/cry-076/internal/repository/postgres"
)

func TestPostgresProfileRepositoryOptimisticConcurrency(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	migrations, err := filepath.Glob("../migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("discover migrations: paths=%v err=%v", migrations, err)
	}
	sort.Strings(migrations)
	for _, migrationPath := range migrations {
		migration, readErr := os.ReadFile(migrationPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, execErr := db.Pool.Exec(ctx, string(migration)); execErr != nil {
			t.Fatalf("apply %s: %v", filepath.Base(migrationPath), execErr)
		}
	}
	repo := postgres.ProfileRepository{DB: db}
	owner := "integration-owner-" + time.Now().Format("150405.000000")
	saved, err := repo.Save(ctx, domain.Profile{OwnerID: owner, Contact: domain.ContactProfile{FullName: "集成测试"}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Version != 1 {
		t.Fatalf("version=%d", saved.Version)
	}
	if _, err := repo.Save(ctx, saved, 0); err == nil {
		t.Fatal("stale save should conflict")
	}
	loaded, err := repo.Get(ctx, owner)
	if err != nil || loaded.Contact.FullName != "集成测试" {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
}

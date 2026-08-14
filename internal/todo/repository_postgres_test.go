package todo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositoryCRUD(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, "drop table if exists todos"); err != nil {
		t.Fatalf("drop table: %v", err)
	}
	migrationSQL, err := os.ReadFile("../../migrations/0001_create_todos.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(migrationSQL)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	repository := NewPostgresRepository(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	created, err := repository.Create(ctx, Todo{
		ID:          "018f1a54-3c5b-4ef3-9ec4-8ff8e32180b8",
		Title:       "learn postgres",
		Description: "",
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Title != created.Title {
		t.Fatalf("Title = %q, want %q", got.Title, created.Title)
	}

	updated, err := repository.Update(ctx, UpdateInput{
		ID:          created.ID,
		Title:       "learn optimistic locking",
		Description: "",
		Status:      StatusInProgress,
		Version:     created.Version,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != 2 {
		t.Fatalf("Version = %d, want 2", updated.Version)
	}

	_, err = repository.Update(ctx, UpdateInput{
		ID:          created.ID,
		Title:       "stale",
		Description: "",
		Status:      StatusCompleted,
		Version:     created.Version,
	})
	if err != ErrVersionConflict {
		t.Fatalf("Update() error = %v, want ErrVersionConflict", err)
	}

	if err := repository.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

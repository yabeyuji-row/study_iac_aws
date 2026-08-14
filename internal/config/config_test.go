package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_SECRET_JSON", "")
	t.Setenv("HTTP_ADDR", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadUsesDatabaseURLAndDefaultHTTPAddr(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://todo:secret@localhost:5432/todo_api")
	t.Setenv("DATABASE_SECRET_JSON", "")
	t.Setenv("HTTP_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabaseURL != "postgres://todo:secret@localhost:5432/todo_api" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
}

func TestLoadUsesHTTPAddrOverride(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://todo:secret@localhost:5432/todo_api")
	t.Setenv("DATABASE_SECRET_JSON", "")
	t.Setenv("HTTP_ADDR", ":9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
}

func TestLoadBuildsDatabaseURLFromSecretJSON(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_SECRET_JSON", `{
		"host": "db.internal",
		"port": 5432,
		"dbname": "todo_api",
		"username": "todo",
		"password": "secret with symbols:/@",
		"sslmode": "require"
	}`)
	t.Setenv("HTTP_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := "postgres://todo:secret%20with%20symbols%3A%2F%40@db.internal:5432/todo_api?sslmode=require"
	if cfg.DatabaseURL != want {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, want)
	}
}

func TestLoadRejectsDatabaseSecretWithoutPassword(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_SECRET_JSON", `{
		"host": "db.internal",
		"port": 5432,
		"dbname": "todo_api",
		"username": "todo",
		"sslmode": "require"
	}`)
	t.Setenv("HTTP_ADDR", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

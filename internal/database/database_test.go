package database

import (
	"context"
	"strings"
	"testing"
)

func TestOpenRejectsInvalidDatabaseURL(t *testing.T) {
	pool, err := Open(context.Background(), "not-a-postgres-url")
	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Fatal("Open() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "parse database url") {
		t.Fatalf("Open() error = %v, want parse database url context", err)
	}
}

package main

import (
	"context"
	"strings"
	"testing"
)

func TestRunRequiresDirection(t *testing.T) {
	err := run(context.Background(), []string{"migrate"})
	if err == nil {
		t.Fatal("run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "usage: migrate up|down") {
		t.Fatalf("run() error = %v, want usage error", err)
	}
}

func TestRunRejectsUnknownDirection(t *testing.T) {
	err := run(context.Background(), []string{"migrate", "sideways"})
	if err == nil {
		t.Fatal("run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), `unknown migration direction "sideways"`) {
		t.Fatalf("run() error = %v, want unknown direction error", err)
	}
}

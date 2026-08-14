package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"study_iac_aws/internal/config"
	"study_iac_aws/internal/database"
)

func main() {
	if err := run(context.Background(), os.Args); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) (err error) {
	if len(args) != 2 {
		return fmt.Errorf("usage: migrate up|down")
	}

	direction := args[1]
	if direction != "up" && direction != "down" {
		return fmt.Errorf("unknown migration direction %q", direction)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	files, err := filepath.Glob(filepath.Join("migrations", "*."+direction+".sql"))
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(files)
	if direction == "down" {
		for leftIndex, rightIndex := 0, len(files)-1; leftIndex < rightIndex; {
			files[leftIndex], files[rightIndex] = files[rightIndex], files[leftIndex]
			leftIndex++
			rightIndex--
		}
	}

	migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		sqlText := strings.TrimSpace(string(sqlBytes))
		if sqlText == "" {
			continue
		}
		if _, err := pool.Exec(migrationCtx, sqlText); err != nil {
			return fmt.Errorf("execute migration %s: %w", file, err)
		}
		slog.Info("executed migration", "file", file)
	}

	return nil
}

package database_migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMigrateCancelsLongRunningMigration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewMigrationDatabase(context.Background(), "migration-cancel")
	if err != nil {
		t.Fatalf("create migration database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropMigrationDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop migration database: %v", err)
		}
	})

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open migration database: %v", err)
	}
	defer db.Close()

	migrationDir := t.TempDir()
	migrationPath := filepath.Join(migrationDir, "00001_sleep.sql")
	if err := os.WriteFile(migrationPath, []byte(`-- +goose Up
SELECT pg_sleep(10);
-- +goose Down
SELECT 1;
`), 0o600); err != nil {
		t.Fatalf("write sleep migration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = postgres.Apply(ctx, db, postgres.NewMigrationSource(migrationDir))
	if err == nil {
		t.Fatal("expected migration cancellation error")
	}
	if !isContextCancellationMigrationError(err) {
		t.Fatalf("expected context cancellation migration error, got %v", err)
	}
}

func isContextCancellationMigrationError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "context canceled") ||
		strings.Contains(lower, "context deadline exceeded") ||
		strings.Contains(lower, "canceling statement due to user request")
}

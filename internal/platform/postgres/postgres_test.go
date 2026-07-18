package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSchemaBootstrap_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "phase0-i-0-01", "up")
	source := dbmigrations.Source()

	assertCount(t, db, `SELECT COUNT(*) FROM pg_extension WHERE extname IN ('pgcrypto', 'citext')`, 2)
	assertCount(t, db, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'deployment_bootstrap_state', 'deployment_admin_audit_events')`, 3)

	if _, err := postgres.Migrate(context.Background(), db, source, "up"); err != nil {
		t.Fatalf("run second schema bootstrap: %v", err)
	}

	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto'`)
	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'citext'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_bootstrap_state'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_admin_audit_events'`)
	assertCount(t, db, `SELECT COUNT(*) FROM pg_extension WHERE extname IN ('pgcrypto', 'citext')`, 2)
	assertCount(t, db, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'deployment_bootstrap_state', 'deployment_admin_audit_events')`, 3)
}

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

	_, err = postgres.Migrate(ctx, db, postgres.NewMigrationSource(migrationDir), "up")
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

func assertExists(t testing.TB, db *sql.DB, query string) {
	t.Helper()

	var marker int
	if err := db.QueryRowContext(context.Background(), query).Scan(&marker); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
}

func assertCount(t testing.TB, db *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}

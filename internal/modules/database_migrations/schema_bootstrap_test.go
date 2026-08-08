package database_migrations_test

import (
	"context"
	"database/sql"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSchemaBootstrap_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "bootstrap-i-0-01")
	source := dbmigrations.Source()

	assertCount(t, db, `SELECT COUNT(*) FROM pg_extension WHERE extname IN ('pgcrypto', 'citext')`, 2)
	assertCount(t, db, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'deployment_bootstrap_state', 'deployment_admin_audit_events')`, 3)

	if _, err := database_migrations.Apply(context.Background(), db, source); err != nil {
		t.Fatalf("run second schema bootstrap: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("borrowed database handle was closed by migration apply: %v", err)
	}

	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto'`)
	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'citext'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_bootstrap_state'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_admin_audit_events'`)
	assertCount(t, db, `SELECT COUNT(*) FROM pg_extension WHERE extname IN ('pgcrypto', 'citext')`, 2)
	assertCount(t, db, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'deployment_bootstrap_state', 'deployment_admin_audit_events')`, 3)
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

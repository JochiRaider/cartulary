package postgres_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPhase0_SchemaBootstrap_I_0_01(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewDatabase(context.Background(), "phase0-i-0-01")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	}()

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	defer db.Close()

	source := dbmigrations.Source()

	if _, err := postgres.Migrate(db, source, "up"); err != nil {
		t.Fatalf("run first schema bootstrap: %v", err)
	}

	assertCount(t, db, `SELECT COUNT(*) FROM pg_extension WHERE extname IN ('pgcrypto', 'citext')`, 2)
	assertCount(t, db, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'deployment_bootstrap_state', 'deployment_admin_audit_events')`, 3)

	if _, err := postgres.Migrate(db, source, "up"); err != nil {
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

package postgres_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPhase0_SchemaBootstrap_U_0_06(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewDatabase(context.Background(), "phase0-u-0-06")
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

	status, err := postgres.Migrate(db, phase0MigrationsDir(), "up")
	if err != nil {
		t.Fatalf("run first schema bootstrap: %v", err)
	}
	if status.Empty {
		t.Fatal("expected numbered phase 0 migrations to exist")
	}

	status, err = postgres.Migrate(db, phase0MigrationsDir(), "up")
	if err != nil {
		t.Fatalf("run second schema bootstrap: %v", err)
	}
	if status.Empty {
		t.Fatal("expected migration rerun to use the numbered phase 0 migrations")
	}
}

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

	if _, err := postgres.Migrate(db, phase0MigrationsDir(), "up"); err != nil {
		t.Fatalf("run first schema bootstrap: %v", err)
	}
	if _, err := postgres.Migrate(db, phase0MigrationsDir(), "up"); err != nil {
		t.Fatalf("run second schema bootstrap: %v", err)
	}

	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto'`)
	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'citext'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_bootstrap_state'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_admin_audit_events'`)
}

func assertExists(t testing.TB, db *sql.DB, query string) {
	t.Helper()

	var marker int
	if err := db.QueryRowContext(context.Background(), query).Scan(&marker); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
}

func phase0MigrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "db", "migrations")
}

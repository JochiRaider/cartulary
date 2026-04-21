package pgtest

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestHarnessStartsPostgresAndRunsCurrentMigrationPath(t *testing.T) {
	harness := Start(t)

	testDB, status, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database: %v", err)
	}
	defer func() {
		if err := harness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop database: %v", err)
		}
	}()

	if testDB.DSN == "" {
		t.Fatal("expected database dsn")
	}
	if status.Empty {
		t.Fatal("expected the current bootstrap migration set to include numbered migrations")
	}
}

func TestPrepareDatabaseTReturnsMigratedDatabase(t *testing.T) {
	harness := Start(t)
	testDB := harness.PrepareDatabaseT(t, "bootstrap_t")

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var exists bool
	if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'users'
)`).Scan(&exists); err != nil {
		t.Fatalf("query users table: %v", err)
	}
	if !exists {
		t.Fatal("expected PrepareDatabaseT to return a migrated database")
	}
}

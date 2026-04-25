// Package phase1storetest contains DB-backed Phase 1 helpers that stay free of
// runtime and HTTP test dependencies so auth-package unit tests can import them
// without a package cycle.
package phase1storetest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type StoreHarness struct {
	Pool *pgxpool.Pool
	DB   *sql.DB
}

func OpenStore(t testing.TB, dsn string) *StoreHarness {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return &StoreHarness{
		Pool: pool,
		DB:   db,
	}
}

func StartStore(t testing.TB, prefix string) *StoreHarness {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, prefix)
	return OpenStore(t, testDB.DSN)
}

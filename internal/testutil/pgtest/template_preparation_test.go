package pgtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestTemplatePreparationCompletesCopyBeforePublication(t *testing.T) {
	// A separate physical cluster makes checkpoint provenance independent of
	// other test processes and the suite's own template preparation.
	h, err := StartOwned(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := h.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	ctx := t.Context()
	if err := h.ensureLocalTemplateDatabase(ctx); err != nil {
		t.Fatal(err)
	}
	admin, err := sql.Open("pgx", h.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	if _, err := admin.ExecContext(ctx, "CHECKPOINT"); err != nil {
		t.Fatal(err)
	}
	var before string
	if err := admin.QueryRowContext(ctx, "SELECT pg_current_wal_insert_lsn()::text").Scan(&before); err != nil {
		t.Fatal(err)
	}
	first := h.PrepareIsolatedDatabaseT(t, "prepared_first")
	var checkpointCompleted bool
	if err := admin.QueryRowContext(ctx, "SELECT checkpoint_lsn > $1::pg_lsn FROM pg_control_checkpoint()", before).Scan(&checkpointCompleted); err != nil {
		t.Fatal(err)
	}
	if !checkpointCompleted {
		t.Fatal("template copy was published before its preparation checkpoint completed")
	}
	second := h.PrepareIsolatedDatabaseT(t, "prepared_second")
	firstDB, err := sql.Open("pgx", first.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = firstDB.Close() })
	if _, err := firstDB.ExecContext(ctx, "CREATE TABLE public.clone_isolation_probe (id integer PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	secondDB, err := sql.Open("pgx", second.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = secondDB.Close() })
	var isolated, migrated bool
	if err := secondDB.QueryRowContext(ctx, "SELECT to_regclass('public.clone_isolation_probe') IS NULL, to_regclass('public.goose_db_version') IS NOT NULL").Scan(&isolated, &migrated); err != nil {
		t.Fatal(err)
	}
	if !isolated || !migrated {
		t.Fatal("prepared clones lost template schema or shared mutable state")
	}
	if err := h.createDatabase(ctx, second.Name, h.templateDB); err == nil {
		t.Fatal("template copy replaced an existing database")
	}
	if err := firstDB.Close(); err != nil {
		t.Fatal(err)
	}
	if err := h.DropDatabase(ctx, first.Name); err != nil {
		t.Fatal(err)
	}
	if err := secondDB.PingContext(ctx); err != nil {
		t.Fatalf("dropping one clone disrupted its live peer: %v", err)
	}
	var templatePresent bool
	if err := admin.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1 AND datistemplate AND NOT datallowconn)", h.templateDB).Scan(&templatePresent); err != nil || !templatePresent {
		t.Fatalf("template changed after clone cleanup: present=%v error=%v", templatePresent, err)
	}
	// A failed copy must not publish a database or acquire ownership of it.
	failedName := h.nextDatabaseName("failed_copy")
	if err := h.createDatabase(ctx, failedName, "ct_missing_preparation_template"); err == nil {
		t.Fatal("missing template was accepted")
	}
	var exists bool
	if err := admin.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1)", failedName).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists || h.ownsDatabase(failedName) {
		t.Fatalf("failed template copy was published: exists=%v owned=%v", exists, h.ownsDatabase(failedName))
	}
}

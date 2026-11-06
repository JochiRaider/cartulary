package database_migrations_test

import (
	"context"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestIncidentPortabilityCutoverRejectsPredecessorBeforeDDL_Integration(t *testing.T) {
	h := pgtest.Start(t)
	previous := h.MigrationDatabaseThroughT(t, 38)
	db := previous.SQL()
	const snapshot = `SELECT jsonb_build_object('ledger',(SELECT jsonb_agg(to_jsonb(v) ORDER BY id) FROM public.goose_db_version v),'relations',(SELECT jsonb_agg(c.relname ORDER BY c.relname) FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public'))::text`
	var before, after string
	if err := db.QueryRowContext(t.Context(), snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	err = database_migrations.Apply(context.Background(), db, source)
	requireExternalMigrationFailureReason(t, err, "schema_fresh_state_required")
	if err := db.QueryRowContext(t.Context(), snapshot).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("fresh-state rejection mutated the schema or migration ledger")
	}
}

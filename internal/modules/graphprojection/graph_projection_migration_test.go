package graphprojection_test

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestGraphProjectionHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "graph-projection-head-contract", postgres.PurposeRuntime)
	ctx := context.Background()

	var columns int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND ((table_name = 'graph_projection_runs' AND column_name IN (
            'started_at', 'generated_at', 'replaced_at', 'invalidated_at', 'retention_policy_json'
        ))
     OR (table_name = 'graph_projection_views' AND column_name IN (
            'invalidation_json', 'selected_projection_run_id'
        ))
     OR (table_name = 'graph_projection_idempotency' AND column_name = 'scope_key'))
`).Scan(&columns); err != nil {
		t.Fatalf("inspect graph projection columns: %v", err)
	}
	if columns != 8 {
		t.Fatalf("graph projection contract columns = %d, want 8", columns)
	}

	var selectedRunFK bool
	if err := db.QueryRowContext(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM pg_catalog.pg_constraint
     WHERE connamespace = 'public'::regnamespace
       AND conname = 'graph_projection_views_selected_run_fkey'
       AND convalidated
       AND condeferrable
       AND condeferred
)
`).Scan(&selectedRunFK); err != nil {
		t.Fatalf("inspect selected-run foreign key: %v", err)
	}
	if !selectedRunFK {
		t.Fatal("graph projection selected-run foreign key is not validated and initially deferred")
	}
}

package graphprojection_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestGraphProjectionV2HeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "graph-projection-v2-head-contract", postgres.PurposeRuntime)
	ctx := context.Background()

	assertGraphProjectionV2TableCount(t, ctx, db, 5)
	assertLegacyGraphProjectionTableCount(t, ctx, db, 0)
	assertCleanupCandidateIndex(t, ctx, db, true)

	var declarationResultForeignKeys int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM pg_catalog.pg_constraint constraint_row
  JOIN pg_catalog.pg_class source_table ON source_table.oid = constraint_row.conrelid
  JOIN pg_catalog.pg_class target_table ON target_table.oid = constraint_row.confrelid
 WHERE constraint_row.contype = 'f'
   AND source_table.relname = 'network_flow_graph_views'
   AND target_table.relname LIKE 'graph_projection_result%'
`).Scan(&declarationResultForeignKeys); err != nil {
		t.Fatalf("inspect declaration/result foreign keys: %v", err)
	}
	if declarationResultForeignKeys != 0 {
		t.Fatalf("authoritative declaration has %d forbidden derived-result foreign keys", declarationResultForeignKeys)
	}
}

func assertGraphProjectionV2TableCount(t *testing.T, ctx context.Context, db *sql.DB, want int) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.tables
 WHERE table_schema = 'public'
   AND table_name IN (
       'network_flow_graph_views',
       'graph_projection_results',
       'graph_projection_result_vertices',
       'graph_projection_result_edges',
       'graph_projection_result_leases'
   )
`).Scan(&got); err != nil {
		t.Fatalf("count Graph Projection v2 tables: %v", err)
	}
	if got != want {
		t.Fatalf("Graph Projection v2 table count = %d; want %d", got, want)
	}
}

func assertLegacyGraphProjectionTableCount(t *testing.T, ctx context.Context, db *sql.DB, want int) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.tables
 WHERE table_schema = 'public'
   AND table_name IN (
       'graph_projection_edges',
       'graph_projection_idempotency',
       'graph_projection_runs',
       'graph_projection_vertices',
       'graph_projection_views'
   )
`).Scan(&got); err != nil {
		t.Fatalf("count legacy Graph Projection tables: %v", err)
	}
	if got != want {
		t.Fatalf("legacy Graph Projection table count = %d; want %d", got, want)
	}
}

func assertCleanupCandidateIndex(t testing.TB, ctx context.Context, db *sql.DB, want bool) {
	t.Helper()
	var definition sql.NullString
	if err := db.QueryRowContext(ctx, `
SELECT indexdef
  FROM pg_catalog.pg_indexes
 WHERE schemaname = 'public'
   AND indexname = 'graph_projection_results_cleanup_candidate_idx'
`).Scan(&definition); err != nil && err != sql.ErrNoRows {
		t.Fatalf("inspect cleanup candidate index: %v", err)
	}
	if definition.Valid != want {
		t.Fatalf("cleanup candidate index present=%t want=%t", definition.Valid, want)
	}
	if want && (!strings.Contains(definition.String, "source_owner_id") ||
		!strings.Contains(definition.String, "published_at") ||
		!strings.Contains(definition.String, "projection_result_id")) {
		t.Fatalf("cleanup candidate index definition drifted: %q", definition.String)
	}
}

package graphprojection_test

import (
	"context"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestGraphProjectionMigration32ResetsUnreferencedDerivedState(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, "graph-projection-32-reset", 31)
	db := migrationDB.SQL()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id, graph_view_key, state, latest_projection_run_id,
    latest_source_snapshot_id, projection_version, updated_at, validation_status
) VALUES ('gv_legacy', 'legacy', 'available', NULL, 'snapshot', 'v1', now(), 'valid')
`); err != nil {
		t.Fatalf("seed legacy graph view: %v", err)
	}
	if err := migrationDB.ApplyThrough(ctx, 32); err != nil {
		t.Fatalf("migrate graph projection state to 32: %v", err)
	}
	var viewCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM graph_projection_views`).Scan(&viewCount); err != nil {
		t.Fatalf("count reset graph views: %v", err)
	}
	if viewCount != 0 {
		t.Fatalf("legacy graph views retained after conformance reset: %d", viewCount)
	}
	var addedColumns int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND ((table_name = 'graph_projection_runs' AND column_name IN ('started_at', 'generated_at', 'replaced_at', 'invalidated_at', 'retention_policy_json'))
     OR (table_name = 'graph_projection_views' AND column_name = 'invalidation_json')
	 OR (table_name = 'graph_projection_views' AND column_name = 'selected_projection_run_id')
     OR (table_name = 'graph_projection_idempotency' AND column_name = 'scope_key'))
`).Scan(&addedColumns); err != nil {
		t.Fatalf("inspect graph projection migration columns: %v", err)
	}
	if addedColumns != 8 {
		t.Fatalf("graph projection migration columns = %d want 8", addedColumns)
	}
}

func TestGraphProjectionMigration32BlocksReferencedDerivedState(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, "graph-projection-32-referenced", 31)
	db := migrationDB.SQL()
	ctx := context.Background()
	digest := strings.Repeat("a", 64)
	if _, err := db.ExecContext(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id, graph_view_key, state, latest_projection_run_id,
    latest_source_snapshot_id, projection_version, updated_at, validation_status
) VALUES ('gv_referenced', 'referenced', 'available', 'gpr_referenced', 'snapshot', 'v1', now(), 'valid')
`); err != nil {
		t.Fatalf("seed referenced graph view: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO graph_projection_runs (
    projection_run_id, graph_view_id, source_snapshot_id, projection_version,
    state, projection_run_nonce, projection_config_digest, projection_source_digest,
    projection_output_digest, accepted_at, completed_at, validation_summary_json
) VALUES ('gpr_referenced', 'gv_referenced', 'snapshot', 'v1', 'available', 'nonce', $1, $1, $1, now(), now(), '{}'::jsonb)
`, digest); err != nil {
		t.Fatalf("seed referenced projection run: %v", err)
	}
	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatalf("disable fixture foreign-key triggers: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO reporting_releases (
    release_id, incident_id, snapshot_id, created_by_user_id, client_txn_id,
    release_scope, release_state, snapshot_at, source_change_set_high_watermark,
    derivation_version, export_model_sha256, template_id, template_version,
    redaction_profile_id, redaction_profile_version, redaction_profile_sha256,
    output_kind, create_job_id, graph_projection_refs, render_admitted_at
) VALUES (
    '32000000-0000-4000-8000-000000000001',
    '32000000-0000-4000-8000-000000000002',
    '32000000-0000-4000-8000-000000000003',
    '32000000-0000-4000-8000-000000000004',
    'migration-fixture', 'internal_draft', 'pending_approval', now(), 'boundary',
    'v1', $1, 'template', 'v1', 'redaction', 'v1', $1, 'slidev',
    '32000000-0000-4000-8000-000000000005',
    '[{"projection_run_id":"gpr_referenced"}]'::jsonb,
    now()
)
`, digest); err != nil {
		t.Fatalf("seed release projection reference: %v", err)
	}
	if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
		t.Fatalf("restore fixture foreign-key triggers: %v", err)
	}
	err := migrationDB.ApplyThrough(ctx, 32)
	if err == nil || !strings.Contains(err.Error(), "referenced_projection_run_count=1") {
		t.Fatalf("migration reference preflight err = %v", err)
	}
}

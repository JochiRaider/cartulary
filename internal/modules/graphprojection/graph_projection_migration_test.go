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

func TestGraphProjectionV1RemovalEmptyPreflightAndMechanicalRollback_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 32)
	ctx := context.Background()

	assertLegacyGraphProjectionTableCount(t, ctx, migrationDB.SQL(), 5)
	if err := migrationDB.ApplyThrough(ctx, 33); err != nil {
		t.Fatalf("apply Graph Projection v1 removal: %v", err)
	}
	assertGraphProjectionV2TableCount(t, ctx, migrationDB.SQL(), 5)
	assertLegacyGraphProjectionTableCount(t, ctx, migrationDB.SQL(), 0)

	if err := migrationDB.RollbackThrough(ctx, 32); err != nil {
		t.Fatalf("mechanically roll back Graph Projection v1 removal: %v", err)
	}
	assertLegacyGraphProjectionTableCount(t, ctx, migrationDB.SQL(), 5)

	if err := migrationDB.ApplyThrough(ctx, 33); err != nil {
		t.Fatalf("reapply Graph Projection v1 removal: %v", err)
	}
	assertLegacyGraphProjectionTableCount(t, ctx, migrationDB.SQL(), 0)
}

func TestGraphProjectionV1RemovalRejectsLegacyTableRows_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 32)
	ctx := context.Background()

	if _, err := migrationDB.SQL().ExecContext(ctx, `
INSERT INTO graph_projection_idempotency (
    operation, scope_key, idempotency_key, request_fingerprint,
    response_json, created_at, expires_at
) VALUES ('create', 'incident:test', 'legacy-key', 'legacy-fingerprint', '{}', now(), now() + interval '1 hour')
`); err != nil {
		t.Fatalf("seed legacy Graph Projection row: %v", err)
	}

	requireV1CutoverPreflightFailure(t, migrationDB.ApplyThrough(ctx, 33), "idempotency=1")
	assertLegacyGraphProjectionTableCount(t, ctx, migrationDB.SQL(), 5)
}

func TestGraphProjectionV1RemovalRejectsReportingReleaseV1References_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 32)
	ctx := context.Background()
	fixture := seedCutoverReportingBase(t, ctx, migrationDB.SQL())

	if _, err := migrationDB.SQL().ExecContext(ctx, `
INSERT INTO reporting_snapshots (
    snapshot_id, incident_id, created_by_user_id, client_txn_id, snapshot_at,
    source_change_set_high_watermark, source_boundary_json, derivation_version,
    export_model_sha256, export_model_json, create_job_id
) VALUES ($1, $2, $3, 'snapshot-v1-ref', now(), 'watermark', '{}', 'v1', $4, '{}', $5)
`, "30000000-0000-0000-0000-000000000001", fixture.incidentID, fixture.userID, cutoverSHA256, fixture.jobID); err != nil {
		t.Fatalf("seed Reporting snapshot: %v", err)
	}
	if _, err := migrationDB.SQL().ExecContext(ctx, `
INSERT INTO reporting_releases (
    release_id, incident_id, snapshot_id, created_by_user_id, client_txn_id,
    release_scope, release_state, snapshot_at, source_change_set_high_watermark,
    derivation_version, export_model_sha256, template_id, template_version,
    redaction_profile_id, redaction_profile_version, redaction_profile_sha256,
    output_kind, create_job_id, graph_projection_refs, render_admitted_at
) VALUES (
    $1, $2, $3, $4, 'release-v1-ref', 'internal_draft', 'pending_approval', now(),
    'watermark', 'v1', $5, 'template', 'v1', 'internal_passthrough', 'v1', $5,
    'mermaid', $6, $7::jsonb, now()
)
`, "40000000-0000-0000-0000-000000000001", fixture.incidentID, "30000000-0000-0000-0000-000000000001", fixture.userID, cutoverSHA256, fixture.jobID, `[{
  "projection_schema_id":"graph_projection.v1",
  "graph_view_id":"legacy-view",
  "projection_run_id":"legacy-run",
  "source_snapshot_id":"legacy-snapshot",
  "projection_version":"v1",
  "projection_config_digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "projection_source_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  "projection_output_digest":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
}]`); err != nil {
		t.Fatalf("seed Reporting release v1 reference: %v", err)
	}

	requireV1CutoverPreflightFailure(t, migrationDB.ApplyThrough(ctx, 33), "reporting_release_refs=1")
}

func TestGraphProjectionV1RemovalRejectsReportingJobV1AndMalformedReferences_Integration(t *testing.T) {
	for _, test := range []struct {
		name        string
		requestJSON string
		wantDetail  string
	}{
		{
			name: "v1 reference",
			requestJSON: `{"graph_projection_refs":[{
  "projection_schema_id":"graph_projection.v1",
  "graph_view_id":"legacy-view",
  "projection_run_id":"legacy-run"
}]}`,
			wantDetail: "reporting_job_refs=1",
		},
		{
			name:        "malformed reference collection",
			requestJSON: `{"graph_projection_refs":{"projection_run_id":"legacy-run"}}`,
			wantDetail:  "malformed_reporting_job_refs=1",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			harness := pgtest.Start(t)
			migrationDB := harness.MigrationDatabaseThroughT(t, 32)
			ctx := context.Background()
			fixture := seedCutoverReportingBase(t, ctx, migrationDB.SQL())

			if _, err := migrationDB.SQL().ExecContext(ctx, `
INSERT INTO reporting_job_payloads (
    job_id, job_kind, incident_id, actor_user_id, request_json
) VALUES ($1, 'release_create', $2, $3, $4::jsonb)
`, fixture.jobID, fixture.incidentID, fixture.userID, test.requestJSON); err != nil {
				t.Fatalf("seed Reporting job reference: %v", err)
			}

			requireV1CutoverPreflightFailure(t, migrationDB.ApplyThrough(ctx, 33), test.wantDetail)
		})
	}
}

type cutoverReportingFixture struct {
	userID     string
	incidentID string
	jobID      string
}

const cutoverSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func seedCutoverReportingBase(t *testing.T, ctx context.Context, db *sql.DB) cutoverReportingFixture {
	t.Helper()
	fixture := cutoverReportingFixture{
		userID:     "10000000-0000-0000-0000-000000000001",
		incidentID: "20000000-0000-0000-0000-000000000001",
		jobID:      "50000000-0000-0000-0000-000000000001",
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required)
VALUES ($1, 'graph-v1-cutover@example.test', 'Graph v1 cutover', 'hash', false)
`, fixture.userID); err != nil {
		t.Fatalf("seed cutover user: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, 'GP-V1-CUTOVER', 'gp-v1-cutover', 'Graph v1 cutover', 'active', $2, $2)
`, fixture.incidentID, fixture.userID); err != nil {
		t.Fatalf("seed cutover incident: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, auth_policy, handler_name,
    job_kind, progress_unit_id
) VALUES (
    $1, 'incident', $2, 'queued', true, $3, now(), now(), 0,
    'incident_membership', 'reporting.release_create', 'reporting.release_create',
    'reporting.release.create.items.v1'
)
`, fixture.jobID, fixture.incidentID, fixture.userID); err != nil {
		t.Fatalf("seed cutover job: %v", err)
	}
	return fixture
}

func requireV1CutoverPreflightFailure(t *testing.T, err error, wantDetail string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected Graph Projection v1 cutover preflight failure")
	}
	message := err.Error()
	if !strings.Contains(message, "graph_projection_v1_cutover_preflight_failed") {
		t.Fatalf("cutover failure = %q; want stable reason for fixture %q", message, wantDetail)
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

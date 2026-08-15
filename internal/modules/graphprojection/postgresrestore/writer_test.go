package postgresrestore

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestGraphRestoreAcceptanceGPRA01And16ClearsAllDerivedHistory_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-clear-only")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	writer, err := New(db)
	if err != nil {
		t.Fatalf("construct restore writer: %v", err)
	}
	plan := restorePublicationPlan(nil)
	proof, err := writer.ReplaceAll(ctx, plan)
	if err != nil {
		t.Fatalf("clear-only restore publication: %v (cause: %v)", err, errors.Unwrap(err))
	}
	if proof.PostconditionSHA256 != plan.PostconditionSHA256 || len(proof.RebuiltViews) != 0 {
		t.Fatalf("clear-only proof mismatch: %#v", proof)
	}
	assertGraphTableCounts(t, ctx, db, []int{0, 0, 0, 0, 0})
}

func TestGraphRestoreAcceptanceGPRA02PublishesOneFreshAvailableRun_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-publish")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	writer, err := New(db)
	if err != nil {
		t.Fatalf("construct restore writer: %v", err)
	}
	run := restoreTestRun("gv_restore_new", "run_restore_new")
	rebuilt := restoreRebuiltView(run)
	plan := restorePublicationPlan([]graphprojection.RestoreStagedProjection{{
		SourceRegistrationID: "source.test", CandidateID: "candidate.test", Run: run,
	}})
	plan.RebuiltViews = []graphprojection.RestoreRebuiltView{rebuilt}
	proof, err := writer.ReplaceAll(ctx, plan)
	if err != nil {
		t.Fatalf("publish restored Graph run: %v (cause: %v)", err, errors.Unwrap(err))
	}
	if len(proof.RebuiltViews) != 1 || proof.RebuiltViews[0].ProjectionRunID != run.ProjectionRunID {
		t.Fatalf("publication proof mismatch: %#v", proof)
	}
	assertGraphTableCounts(t, ctx, db, []int{0, 0, 1, 0, 1})
	var state, selectedRunID string
	if err := db.QueryRow(ctx, `SELECT state, selected_projection_run_id FROM graph_projection_views WHERE graph_view_id = $1`, run.GraphViewID).Scan(&state, &selectedRunID); err != nil {
		t.Fatalf("read restored Graph view: %v", err)
	}
	if state != "available" || selectedRunID != run.ProjectionRunID {
		t.Fatalf("restored Graph selection got state=%q run=%q", state, selectedRunID)
	}
}

func TestGraphRestoreAcceptanceGPRA12PublicationFailureRollsBackClear_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-rollback")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	writer, err := New(db)
	if err != nil {
		t.Fatalf("construct restore writer: %v", err)
	}
	run := restoreTestRun("gv_restore_duplicate", "run_restore_duplicate")
	staged := []graphprojection.RestoreStagedProjection{
		{SourceRegistrationID: "source.a", CandidateID: "candidate.a", Run: run},
		{SourceRegistrationID: "source.b", CandidateID: "candidate.b", Run: run},
	}
	plan := restorePublicationPlan(staged)
	plan.RebuiltViews = []graphprojection.RestoreRebuiltView{restoreRebuiltView(run), restoreRebuiltView(run)}
	if _, err := writer.ReplaceAll(ctx, plan); err == nil {
		t.Fatal("duplicate staged publication unexpectedly succeeded")
	}
	assertGraphTableCounts(t, ctx, db, []int{1, 1, 1, 1, 1})
}

func seedStaleGraphState(t *testing.T, ctx context.Context, db postgres.DB) {
	t.Helper()
	commands := []string{
		`INSERT INTO graph_projection_views (graph_view_id, graph_view_key, state, updated_at, validation_status) VALUES ('gv_stale', 'stale', 'creating', '2026-05-30T00:00:00Z', 'pending')`,
		`INSERT INTO graph_projection_runs (projection_run_id, graph_view_id, source_snapshot_id, projection_version, state, projection_run_nonce, projection_config_digest, projection_source_digest, accepted_at, validation_summary_json, retention_policy_json) VALUES ('run_stale', 'gv_stale', 'snap_stale', 'v1', 'accepted', 'nonce', '` + strings.Repeat("a", 64) + `', '` + strings.Repeat("b", 64) + `', '2026-05-30T00:00:00Z', 'null', '{}')`,
		`UPDATE graph_projection_views SET latest_projection_run_id = 'run_stale', selected_projection_run_id = 'run_stale' WHERE graph_view_id = 'gv_stale'`,
		`INSERT INTO graph_projection_vertices (projection_run_id, graph_view_id, vertex_id, vertex_kind, sort_key, vertex_json) VALUES ('run_stale', 'gv_stale', 'vertex_stale', 'host', 'a', '{}')`,
		`INSERT INTO graph_projection_edges (projection_run_id, graph_view_id, edge_id, edge_kind, src_vertex_id, dst_vertex_id, direction, sort_key, edge_json) VALUES ('run_stale', 'gv_stale', 'edge_stale', 'link', 'vertex_stale', 'vertex_stale', 'directed', 'a', '{}')`,
		`INSERT INTO graph_projection_idempotency (operation, scope_key, idempotency_key, request_fingerprint, response_json, created_at, expires_at) VALUES ('create_projection', 'stale', 'key', '` + strings.Repeat("c", 64) + `', '{}', '2026-05-30T00:00:00Z', '2026-05-31T00:00:00Z')`,
	}
	for _, command := range commands {
		if _, err := db.Exec(ctx, command); err != nil {
			t.Fatalf("seed stale Graph state: %v", err)
		}
	}
	if _, err := db.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
		t.Fatalf("settle stale Graph fixture constraints: %v", err)
	}
}

func assertGraphTableCounts(t *testing.T, ctx context.Context, db postgres.DB, want []int) {
	t.Helper()
	var edges, idempotency, runs, vertices, views int
	if err := db.QueryRow(ctx, `
SELECT (SELECT COUNT(*) FROM graph_projection_edges),
       (SELECT COUNT(*) FROM graph_projection_idempotency),
       (SELECT COUNT(*) FROM graph_projection_runs),
       (SELECT COUNT(*) FROM graph_projection_vertices),
       (SELECT COUNT(*) FROM graph_projection_views)
`).Scan(&edges, &idempotency, &runs, &vertices, &views); err != nil {
		t.Fatalf("read Graph table counts: %v", err)
	}
	got := []int{edges, idempotency, runs, vertices, views}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Graph table counts got %v want %v", got, want)
		}
	}
}

func restoreTestRun(graphViewID string, runID string) graphprojection.ProjectionRun {
	now := time.Date(2026, 5, 30, 1, 0, 0, 0, time.UTC)
	return graphprojection.ProjectionRun{
		Request: graphprojection.ProjectionRequest{
			SourceSnapshotID: "snap_restore", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "recovery_restore",
			ProjectionConfig: graphprojection.ProjectionConfig{
				GraphViewKey: "restore", ProjectionVersion: "v1", AllowEmptyKindRegistry: true,
				RetentionPolicy: graphprojection.RetentionPolicy{},
			},
		},
		GraphViewID: graphViewID, ProjectionRunID: runID, ProjectionRunNonce: "server-nonce",
		ProjectionConfigDigest: strings.Repeat("a", 64), ProjectionSourceDigest: strings.Repeat("b", 64),
		ProjectionOutputDigest: strings.Repeat("c", 64), AcceptedAt: now, StartedAt: &now, GeneratedAt: &now, CompletedAt: &now,
		State: graphprojection.RunStateAvailable, ValidationSummary: graphprojection.ValidationSummary{Status: "passed", Issues: []graphprojection.ValidationIssue{}},
		GraphView: &graphprojection.GraphView{
			ProjectionSchemaID: graphprojection.ProjectionSchemaID, GraphViewID: graphViewID, GraphViewKey: "restore",
			ProjectionRunID: runID, SourceSnapshotID: "snap_restore", ProjectionVersion: "v1", GeneratedAt: "2026-05-30T01:00:00Z",
			State: graphprojection.RunStateAvailable, Vertices: []graphprojection.Vertex{}, Edges: []graphprojection.Edge{},
			ValidationSummary: graphprojection.ValidationSummary{Status: "passed", Issues: []graphprojection.ValidationIssue{}},
		},
	}
}

func restoreRebuiltView(run graphprojection.ProjectionRun) graphprojection.RestoreRebuiltView {
	return graphprojection.RestoreRebuiltView{
		SourceRegistrationID: "source.test", CandidateID: "candidate.test", GraphViewID: run.GraphViewID,
		ProjectionRunID: run.ProjectionRunID, SourceSnapshotID: run.Request.SourceSnapshotID,
		ProjectionVersion:             run.Request.ProjectionConfig.ProjectionVersion,
		NormalizedConfigurationSHA256: run.ProjectionConfigDigest, NormalizedSourceSHA256: run.ProjectionSourceDigest,
		CanonicalOutputSHA256: run.ProjectionOutputDigest,
	}
}

func restorePublicationPlan(staged []graphprojection.RestoreStagedProjection) graphprojection.RestorePublicationPlan {
	return graphprojection.RestorePublicationPlan{
		RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000008001"),
		TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000008002"),
		ClearedTableIDs:    graphprojection.RestoreGraphTableIDs(),
		Projections:        staged, RebuiltViews: []graphprojection.RestoreRebuiltView{}, SkippedCandidates: []graphprojection.RestoreSkippedCandidate{},
		PostconditionSHA256: strings.Repeat("d", 64),
	}
}

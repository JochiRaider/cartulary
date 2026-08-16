package postgresrestore

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const (
	restoreResultIDA = "gpres_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	restoreResultIDB = "gpres_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	restoreVertexIDA = "vx_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	restoreVertexIDB = "vx_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	restoreEdgeIDA   = "ed_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestGraphRestoreAcceptanceGPRA01And16ClearsAllDerivedHistory_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-v2-clear-only")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	writer, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	plan := restorePublicationPlan(nil)
	proof, err := writer.ReplaceAll(ctx, plan)
	if err != nil || proof.PostconditionSHA256 != plan.PostconditionSHA256 || len(proof.RebuiltViews) != 0 {
		t.Fatalf("clear current v2 tables: proof=%#v err=%v cause=%v", proof, err, errors.Unwrap(err))
	}
	assertV2GraphTableCounts(t, ctx, db, []int{0, 0, 0, 0})
}

func TestGraphRestoreAcceptanceGPRA02PublishesOneFreshAvailableRun_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-v2-publish")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	writer, _ := New(db)
	result := restoreCompletedResult(restoreResultIDB)
	staged := graphprojection.RestoreStagedProjection{SourceRegistrationID: "network_flow_activity.graph_views.v1", CandidateID: result.Binding.GraphViewID, Result: result}
	plan := restorePublicationPlan([]graphprojection.RestoreStagedProjection{staged})
	plan.RebuiltViews = []graphprojection.RestoreRebuiltView{restoreRebuiltView(staged)}
	proof, err := writer.ReplaceAll(ctx, plan)
	if err != nil || len(proof.RebuiltViews) != 1 || proof.RebuiltViews[0].ProjectionResultID != restoreResultIDB {
		t.Fatalf("publish exact restored result: proof=%#v err=%v cause=%v", proof, err, errors.Unwrap(err))
	}
	assertV2GraphTableCounts(t, ctx, db, []int{1, 0, 2, 1})
}

func TestGraphRestoreAcceptanceGPRA12PublicationFailureRollsBackClear_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-v2-rollback")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	writer, _ := New(db)
	result := restoreCompletedResult(restoreResultIDB)
	first := graphprojection.RestoreStagedProjection{SourceRegistrationID: "network_flow_activity.graph_views.v1", CandidateID: "candidate-a", Result: result}
	second := first
	second.CandidateID = "candidate-b"
	plan := restorePublicationPlan([]graphprojection.RestoreStagedProjection{first, second})
	plan.RebuiltViews = []graphprojection.RestoreRebuiltView{restoreRebuiltView(first), restoreRebuiltView(second)}
	if _, err := writer.ReplaceAll(ctx, plan); err == nil {
		t.Fatal("duplicate immutable result plan unexpectedly succeeded")
	}
	assertV2GraphTableCounts(t, ctx, db, []int{1, 1, 2, 1})
}

func seedStaleGraphState(t *testing.T, ctx context.Context, db postgres.DB) {
	t.Helper()
	commands := []string{
		`INSERT INTO graph_projection_results (projection_result_id, graph_view_id, source_owner_id, source_snapshot_id, projection_schema_id, projection_version, normalized_configuration_sha256, normalized_source_sha256, canonical_output_sha256, vertex_count, edge_count, result_json, published_at) VALUES ('` + restoreResultIDA + `', 'nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 'network_flow_activity', 'snapshot-stale', 'graph_projection.v2', 'network_flow_activity.v1', '` + strings.Repeat("a", 64) + `', '` + strings.Repeat("b", 64) + `', '` + strings.Repeat("c", 64) + `', 2, 1, '{}', '2026-05-30T00:00:00Z')`,
		`INSERT INTO graph_projection_result_vertices (projection_result_id, vertex_id, vertex_kind, sort_ordinal, sort_key, vertex_json) VALUES ('` + restoreResultIDA + `', '` + restoreVertexIDA + `', 'endpoint', 0, 'a', '{}'), ('` + restoreResultIDA + `', '` + restoreVertexIDB + `', 'endpoint', 1, 'b', '{}')`,
		`INSERT INTO graph_projection_result_edges (projection_result_id, edge_id, edge_kind, src_vertex_id, dst_vertex_id, direction, sort_ordinal, sort_key, edge_json) VALUES ('` + restoreResultIDA + `', '` + restoreEdgeIDA + `', 'flow', '` + restoreVertexIDA + `', '` + restoreVertexIDB + `', 'directed', 0, 'a', '{}')`,
		`INSERT INTO graph_projection_result_leases (lease_id, projection_result_id, lease_owner_id, lease_owner_resource_id, lease_purpose, leased_until, created_at, renewed_at) VALUES ('00000000-0000-0000-0000-000000009999', '` + restoreResultIDA + `', 'snapshot_reporting', 'release-stale', 'render', '2026-06-01T00:00:00Z', '2026-05-30T00:00:00Z', '2026-05-30T00:00:00Z')`,
	}
	for _, command := range commands {
		if _, err := db.Exec(ctx, command); err != nil {
			t.Fatalf("seed stale Graph state: %v", err)
		}
	}
	if _, err := db.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
		t.Fatal(err)
	}
}

func assertV2GraphTableCounts(t *testing.T, ctx context.Context, db postgres.DB, want []int) {
	t.Helper()
	var edges, leases, vertices, results int
	if err := db.QueryRow(ctx, `SELECT (SELECT COUNT(*) FROM graph_projection_result_edges), (SELECT COUNT(*) FROM graph_projection_result_leases), (SELECT COUNT(*) FROM graph_projection_result_vertices), (SELECT COUNT(*) FROM graph_projection_results)`).Scan(&edges, &leases, &vertices, &results); err != nil {
		t.Fatal(err)
	}
	got := []int{edges, leases, vertices, results}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("v2 Graph counts got %v want %v", got, want)
		}
	}
}

func restoreCompletedResult(resultID string) graphprojection.CompletedResultV2 {
	return graphprojection.CompletedResultV2{
		Binding:     graphprojection.ResultBindingV2{ProjectionResultID: resultID, GraphViewID: "nfgv_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", SourceOwnerID: "network_flow_activity", SourceSnapshotID: "snapshot-restored", ProjectionSchemaID: graphprojection.ProjectionSchemaIDV2, ProjectionVersion: "network_flow_activity.v1", NormalizedConfigurationSHA256: strings.Repeat("d", 64), NormalizedSourceSHA256: strings.Repeat("e", 64), CanonicalOutputSHA256: strings.Repeat("f", 64)},
		ResultJSON:  []byte(`{"projection_schema_id":"graph_projection.v2"}`),
		Vertices:    []graphprojection.ResultVertexV2{{VertexID: restoreVertexIDA, VertexKind: "endpoint", SortKey: "a", JSON: []byte(`{"vertex_id":"` + restoreVertexIDA + `"}`)}, {VertexID: restoreVertexIDB, VertexKind: "endpoint", SortKey: "b", JSON: []byte(`{"vertex_id":"` + restoreVertexIDB + `"}`)}},
		Edges:       []graphprojection.ResultEdgeV2{{EdgeID: restoreEdgeIDA, EdgeKind: "flow", SrcVertexID: restoreVertexIDA, DstVertexID: restoreVertexIDB, Direction: "directed", SortKey: "a", JSON: []byte(`{"edge_id":"` + restoreEdgeIDA + `"}`)}},
		PublishedAt: time.Date(2026, 5, 30, 1, 0, 0, 0, time.UTC),
	}
}

func restoreRebuiltView(staged graphprojection.RestoreStagedProjection) graphprojection.RestoreRebuiltView {
	binding := staged.Result.Binding
	return graphprojection.RestoreRebuiltView{SourceRegistrationID: staged.SourceRegistrationID, CandidateID: staged.CandidateID, GraphViewID: binding.GraphViewID, ProjectionResultID: binding.ProjectionResultID, SourceSnapshotID: binding.SourceSnapshotID, ProjectionVersion: binding.ProjectionVersion, NormalizedConfigurationSHA256: binding.NormalizedConfigurationSHA256, NormalizedSourceSHA256: binding.NormalizedSourceSHA256, VertexCount: len(staged.Result.Vertices), EdgeCount: len(staged.Result.Edges), CanonicalOutputSHA256: binding.CanonicalOutputSHA256}
}

func restorePublicationPlan(staged []graphprojection.RestoreStagedProjection) graphprojection.RestorePublicationPlan {
	return graphprojection.RestorePublicationPlan{RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000008001"), TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000008002"), ClearedTableIDs: graphprojection.RestoreGraphTableIDs(), Projections: staged, RebuiltViews: []graphprojection.RestoreRebuiltView{}, PostconditionSHA256: strings.Repeat("d", 64)}
}

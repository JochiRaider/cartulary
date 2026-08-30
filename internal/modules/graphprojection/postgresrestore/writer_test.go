package postgresrestore

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
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

type noOpDerivedStateReconciler struct{}

func (noOpDerivedStateReconciler) ReconcileGraphProjectionDerivedState(context.Context, pgx.Tx, []graphprojection.ResultBindingV2) (int, int, error) {
	return 0, 0, nil
}

type derivedStateReconcilerFunc func(context.Context, pgx.Tx, []graphprojection.ResultBindingV2) (int, int, error)

func (reconcile derivedStateReconcilerFunc) ReconcileGraphProjectionDerivedState(ctx context.Context, tx pgx.Tx, bindings []graphprojection.ResultBindingV2) (int, int, error) {
	return reconcile(ctx, tx, bindings)
}

type restoreDBWrapper struct {
	postgres.DB
	wrapTx   func(pgx.Tx) pgx.Tx
	queryErr error
}

func (db restoreDBWrapper) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil || db.wrapTx == nil {
		return tx, err
	}
	return db.wrapTx(tx), nil
}

func (db restoreDBWrapper) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if db.queryErr != nil {
		return restoreErrorRow{err: db.queryErr}
	}
	return db.DB.QueryRow(ctx, sql, args...)
}

type restoreErrorRow struct{ err error }

func (row restoreErrorRow) Scan(...any) error { return row.err }

type restoreFailureTx struct {
	pgx.Tx
	failExecContains string
	commitAfterWrite bool
	commitErr        error
}

func (tx *restoreFailureTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx.failExecContains != "" && strings.Contains(sql, tx.failExecContains) {
		return pgconn.CommandTag{}, errors.New("injected restore phase failure")
	}
	return tx.Tx.Exec(ctx, sql, args...)
}

func (tx *restoreFailureTx) Commit(ctx context.Context) error {
	if tx.commitErr == nil {
		return tx.Tx.Commit(ctx)
	}
	if tx.commitAfterWrite {
		if err := tx.Tx.Commit(ctx); err != nil {
			return err
		}
	}
	return tx.commitErr
}

func TestGraphRestoreWriterRequiresDatabaseAndReconciler_Unit(t *testing.T) {
	if writer, err := New(nil, noOpDerivedStateReconciler{}); err == nil || writer != nil {
		t.Fatalf("nil database constructor result = %#v, %v", writer, err)
	}
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-required-reconciler")
	if writer, err := New(db, nil); err == nil || writer != nil {
		t.Fatalf("nil reconciler constructor result = %#v, %v", writer, err)
	}
}

func TestGraphRestoreAcceptanceGPRA01And16ClearsAllDerivedHistory_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-v2-clear-only")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	writer, err := New(db, noOpDerivedStateReconciler{})
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
	writer, _ := New(db, noOpDerivedStateReconciler{})
	result := restoreCompletedResult(restoreResultIDB)
	staged := graphrestore.RestoreStagedProjection{SourceRegistrationID: "network_flow_activity.graph_views.v1", CandidateID: result.Binding.GraphViewID, Result: result}
	plan := restorePublicationPlan([]graphrestore.RestoreStagedProjection{staged})
	plan.RebuiltViews = []graphrestore.RestoreRebuiltView{restoreRebuiltView(staged)}
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
	writer, _ := New(db, noOpDerivedStateReconciler{})
	result := restoreCompletedResult(restoreResultIDB)
	first := graphrestore.RestoreStagedProjection{SourceRegistrationID: "network_flow_activity.graph_views.v1", CandidateID: "candidate-a", Result: result}
	second := first
	second.CandidateID = "candidate-b"
	plan := restorePublicationPlan([]graphrestore.RestoreStagedProjection{first, second})
	plan.RebuiltViews = []graphrestore.RestoreRebuiltView{restoreRebuiltView(first), restoreRebuiltView(second)}
	if _, err := writer.ReplaceAll(ctx, plan); err == nil {
		t.Fatal("duplicate immutable result plan unexpectedly succeeded")
	}
	assertV2GraphTableCounts(t, ctx, db, []int{1, 1, 2, 1})
}

func TestGraphRestoreWriterRollsBackEveryPrecommitPhase_Integration(t *testing.T) {
	tests := []struct {
		name       string
		wrapTx     func(pgx.Tx) pgx.Tx
		reconciler DerivedStateReconciler
	}{
		{
			name: "clear",
			wrapTx: func(tx pgx.Tx) pgx.Tx {
				return &restoreFailureTx{Tx: tx, failExecContains: "TRUNCATE TABLE"}
			},
			reconciler: noOpDerivedStateReconciler{},
		},
		{
			name: "reconciliation",
			reconciler: derivedStateReconcilerFunc(func(context.Context, pgx.Tx, []graphprojection.ResultBindingV2) (int, int, error) {
				return 0, 0, errors.New("injected reconciliation failure")
			}),
		},
		{
			name: "invalid reconciliation counts",
			reconciler: derivedStateReconcilerFunc(func(context.Context, pgx.Tx, []graphprojection.ResultBindingV2) (int, int, error) {
				return -1, 0, nil
			}),
		},
		{
			name: "postcondition mismatch",
			reconciler: derivedStateReconcilerFunc(func(context.Context, pgx.Tx, []graphprojection.ResultBindingV2) (int, int, error) {
				return 0, 1, nil
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-phase-"+strings.ReplaceAll(test.name, " ", "-"))
			ctx := context.Background()
			seedStaleGraphState(t, ctx, db)
			wrapped := restoreDBWrapper{DB: db, wrapTx: test.wrapTx}
			writer, err := New(wrapped, test.reconciler)
			if err != nil {
				t.Fatalf("construct phase writer: %v", err)
			}
			plan := restorePlanWithFreshResult()
			if proof, err := writer.ReplaceAll(ctx, plan); err == nil || len(proof.RebuiltViews) != 0 {
				t.Fatalf("phase failure returned proof=%#v err=%v", proof, err)
			}
			assertRestoreRollbackState(t, ctx, db)
		})
	}
}

func TestGraphRestoreWriterReconcilesBeforeCommitAndReturnsExactCounts_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-reconcile-order")
	ctx := context.Background()
	seedStaleGraphState(t, ctx, db)
	reconciled := false
	reconciler := derivedStateReconcilerFunc(func(ctx context.Context, tx pgx.Tx, bindings []graphprojection.ResultBindingV2) (int, int, error) {
		if len(bindings) != 1 || bindings[0].ProjectionResultID != restoreResultIDB {
			return 0, 0, errors.New("reconciler did not receive exact rebuilt binding")
		}
		var staged int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, restoreResultIDB).Scan(&staged); err != nil || staged != 1 {
			return 0, 0, errors.New("reconciler ran before immutable publication")
		}
		now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
		if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_result_leases (
    lease_id, projection_result_id, lease_owner_id, lease_owner_resource_id,
    lease_purpose, leased_until, created_at, renewed_at
) VALUES ($1, $2, 'snapshot_reporting', 'release-restored', 'render', $3, $4, $4)
`, uuid.MustParse("00000000-0000-0000-0000-000000008099"), restoreResultIDB, now.Add(time.Hour), now); err != nil {
			return 0, 0, err
		}
		reconciled = true
		return 2, 1, nil
	})
	writer, err := New(db, reconciler)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := writer.ReplaceAll(ctx, restorePlanWithFreshResult())
	if err != nil || !reconciled || proof.ReconciledNonterminalJobCount != 2 || proof.ReconciledLeaseCount != 1 || len(proof.RebuiltViews) != 1 {
		t.Fatalf("reconciled publication proof=%#v called=%t err=%v", proof, reconciled, err)
	}
	assertV2GraphTableCounts(t, ctx, db, []int{1, 1, 2, 1})
}

func TestGraphRestoreWriterCommitAndPostcommitFailuresAreIndeterminate_Integration(t *testing.T) {
	tests := []struct {
		name         string
		wrapper      func(postgres.DB) restoreDBWrapper
		wantRestored bool
	}{
		{
			name: "commit before outcome",
			wrapper: func(db postgres.DB) restoreDBWrapper {
				return restoreDBWrapper{DB: db, wrapTx: func(tx pgx.Tx) pgx.Tx {
					return &restoreFailureTx{Tx: tx, commitErr: errors.New("injected commit failure")}
				}}
			},
		},
		{
			name: "commit acknowledged as failure after write",
			wrapper: func(db postgres.DB) restoreDBWrapper {
				return restoreDBWrapper{DB: db, wrapTx: func(tx pgx.Tx) pgx.Tx {
					return &restoreFailureTx{Tx: tx, commitAfterWrite: true, commitErr: errors.New("injected ambiguous commit result")}
				}}
			},
			wantRestored: true,
		},
		{
			name: "postcommit verification",
			wrapper: func(db postgres.DB) restoreDBWrapper {
				return restoreDBWrapper{DB: db, queryErr: errors.New("injected postcommit verification failure")}
			},
			wantRestored: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := pgtest.Start(t).BeginRollbackDBT(t, "graph-restore-indeterminate-"+strings.ReplaceAll(test.name, " ", "-"))
			ctx := context.Background()
			seedStaleGraphState(t, ctx, db)
			writer, err := New(test.wrapper(db), noOpDerivedStateReconciler{})
			if err != nil {
				t.Fatal(err)
			}
			proof, err := writer.ReplaceAll(ctx, restorePlanWithFreshResult())
			var publicationErr *graphrestore.RestorePublicationError
			if !errors.As(err, &publicationErr) || !publicationErr.Indeterminate || len(proof.RebuiltViews) != 0 {
				t.Fatalf("indeterminate phase proof=%#v err=%#v", proof, err)
			}
			var stale, restored int
			if err := db.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, restoreResultIDA).Scan(&stale); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, restoreResultIDB).Scan(&restored); err != nil {
				t.Fatal(err)
			}
			if test.wantRestored && (stale != 0 || restored != 1) || !test.wantRestored && (stale != 1 || restored != 0) {
				t.Fatalf("indeterminate database state stale/restored=%d/%d want_restored=%t", stale, restored, test.wantRestored)
			}
		})
	}
}

func restorePlanWithFreshResult() graphrestore.RestorePublicationPlan {
	result := restoreCompletedResult(restoreResultIDB)
	staged := graphrestore.RestoreStagedProjection{SourceRegistrationID: "network_flow_activity.graph_views.v1", CandidateID: result.Binding.GraphViewID, Result: result}
	plan := restorePublicationPlan([]graphrestore.RestoreStagedProjection{staged})
	plan.RebuiltViews = []graphrestore.RestoreRebuiltView{restoreRebuiltView(staged)}
	return plan
}

func assertRestoreRollbackState(t testing.TB, ctx context.Context, db postgres.DB) {
	t.Helper()
	var stale, restored int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, restoreResultIDA).Scan(&stale); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, restoreResultIDB).Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if stale != 1 || restored != 0 {
		t.Fatalf("failed restore mutated target state: stale/restored=%d/%d", stale, restored)
	}
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

func restoreRebuiltView(staged graphrestore.RestoreStagedProjection) graphrestore.RestoreRebuiltView {
	binding := staged.Result.Binding
	return graphrestore.RestoreRebuiltView{SourceRegistrationID: staged.SourceRegistrationID, CandidateID: staged.CandidateID, GraphViewID: binding.GraphViewID, ProjectionResultID: binding.ProjectionResultID, SourceSnapshotID: binding.SourceSnapshotID, ProjectionVersion: binding.ProjectionVersion, NormalizedConfigurationSHA256: binding.NormalizedConfigurationSHA256, NormalizedSourceSHA256: binding.NormalizedSourceSHA256, VertexCount: len(staged.Result.Vertices), EdgeCount: len(staged.Result.Edges), CanonicalOutputSHA256: binding.CanonicalOutputSHA256}
}

func restorePublicationPlan(staged []graphrestore.RestoreStagedProjection) graphrestore.RestorePublicationPlan {
	return graphrestore.RestorePublicationPlan{RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000008001"), TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000008002"), ClearedTableIDs: graphrestore.RestoreGraphTableIDs(), Projections: staged, RebuiltViews: []graphrestore.RestoreRebuiltView{}, PostconditionSHA256: strings.Repeat("d", 64)}
}

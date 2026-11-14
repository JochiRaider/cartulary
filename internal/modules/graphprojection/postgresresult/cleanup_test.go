package postgresresult_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/testutil/graphfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestResultV2CleanupBoundIsolationLockingAndRollback_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "graph-result-v2-cleanup-locking")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("open cleanup test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	now := time.Date(2026, 8, 16, 15, 0, 0, 0, time.UTC)

	first := completedResult(t, "view-a", now.Add(-3*time.Hour))
	second := completedResult(t, "view-b", now.Add(-2*time.Hour))
	otherOwner := graphfixture.Result(t, graphfixture.ResultOptions{GraphViewID: "view-c", SourceOwnerID: "other_projection_owner", SourceSnapshotID: "fixture-snapshot", ProjectionVersion: "fixture-v2", VertexKind: "endpoint", EdgeKind: "flow", VertexCount: 2, PublishedAt: now.Add(-4 * time.Hour)})
	resultIDA, resultIDB, resultIDC := first.Binding.ProjectionResultID, second.Binding.ProjectionResultID, otherOwner.Binding.ProjectionResultID

	for _, result := range []struct {
		name    string
		publish func()
	}{
		{name: "first", publish: func() { publishInNestedTransaction(t, ctx, pool, first) }},
		{name: "second", publish: func() { publishInNestedTransaction(t, ctx, pool, second) }},
		{name: "other owner", publish: func() { publishInNestedTransaction(t, ctx, pool, otherOwner) }},
	} {
		t.Run("publish "+result.name, func(t *testing.T) { result.publish() })
	}

	firstTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer firstTx.Rollback(ctx)
	firstCleaner, err := postgresresult.NewCleaner(firstTx)
	if err != nil {
		t.Fatal(err)
	}
	firstCandidate, err := firstCleaner.LockCleanupCandidate(ctx, "network_flow_activity", nil)
	if err != nil || firstCandidate == nil || firstCandidate.ProjectionResultID != resultIDA {
		t.Fatalf("first candidate = %#v err=%v", firstCandidate, err)
	}

	secondTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer secondTx.Rollback(ctx)
	secondCleaner, err := postgresresult.NewCleaner(secondTx)
	if err != nil {
		t.Fatal(err)
	}
	secondCandidate, err := secondCleaner.LockCleanupCandidate(ctx, "network_flow_activity", nil)
	if err != nil || secondCandidate == nil || secondCandidate.ProjectionResultID != resultIDB {
		t.Fatalf("skip-locked candidate = %#v err=%v", secondCandidate, err)
	}
	if deleted, err := secondCleaner.DeleteLockedResult(ctx, resultIDB); err != nil || !deleted {
		t.Fatalf("delete independently locked candidate: deleted=%t err=%v", deleted, err)
	}
	if err := secondTx.Commit(ctx); err != nil {
		t.Fatalf("commit second cleanup transaction: %v", err)
	}

	if deleted, err := firstCleaner.DeleteLockedResult(ctx, resultIDA); err != nil || !deleted {
		t.Fatalf("stage first candidate deletion: deleted=%t err=%v", deleted, err)
	}
	if err := firstTx.Rollback(ctx); err != nil {
		t.Fatalf("roll back first cleanup transaction: %v", err)
	}
	requireResultPresence(t, ctx, pool, resultIDA, true)
	requireResultPresence(t, ctx, pool, resultIDB, false)
	requireResultPresence(t, ctx, pool, resultIDC, true)

	finalTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer finalTx.Rollback(ctx)
	finalCleaner, _ := postgresresult.NewCleaner(finalTx)
	finalCandidate, err := finalCleaner.LockCleanupCandidate(ctx, "network_flow_activity", nil)
	if err != nil || finalCandidate == nil || finalCandidate.ProjectionResultID != resultIDA {
		t.Fatalf("candidate after rollback = %#v err=%v", finalCandidate, err)
	}
	if err := finalTx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}

	large := graphfixture.Result(t, graphfixture.ResultOptions{GraphViewID: "view-d", SourceOwnerID: "network_flow_activity", SourceSnapshotID: "fixture-snapshot", ProjectionVersion: "fixture-v2", VertexKind: "endpoint", EdgeKind: "flow", VertexCount: 2048, PublishedAt: now.Add(-time.Hour)})
	resultIDD := large.Binding.ProjectionResultID

	publishInNestedTransaction(t, ctx, pool, large)
	largeTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer largeTx.Rollback(ctx)
	largeCleaner, _ := postgresresult.NewCleaner(largeTx)
	largeCandidate, err := largeCleaner.LockCleanupCandidate(ctx, "network_flow_activity", finalCandidate)
	if err != nil || largeCandidate == nil || largeCandidate.ProjectionResultID != resultIDD {
		t.Fatalf("large cascade candidate = %#v err=%v", largeCandidate, err)
	}
	if deleted, err := largeCleaner.DeleteLockedResult(ctx, resultIDD); err != nil || !deleted {
		t.Fatalf("delete large cascade: deleted=%t err=%v", deleted, err)
	}
	if err := largeTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var retainedVertices int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM graph_projection_result_vertices WHERE projection_result_id = $1`, resultIDD).Scan(&retainedVertices); err != nil || retainedVertices != 0 {
		t.Fatalf("large result cascade retained vertices=%d err=%v", retainedVertices, err)
	}
}

func requireResultPresence(t testing.TB, ctx context.Context, pool *pgxpool.Pool, resultID string, want bool) {
	t.Helper()
	var present bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM graph_projection_results WHERE projection_result_id = $1)`, resultID).Scan(&present); err != nil {
		t.Fatalf("inspect result %s: %v", resultID, err)
	}
	if present != want {
		t.Fatalf("result %s presence=%t want=%t", resultID, present, want)
	}
}

package postgresresult_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/testutil/graphfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestResultV2PublicationReadTraversalAndCleanup_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.BeginRollbackDBT(t, "graph-result-v2")
	ctx := context.Background()
	now := time.Date(2026, 8, 15, 19, 0, 0, 0, time.UTC)

	t.Run("current producer content proofs", func(t *testing.T) { checkCurrentResultProofs(t, ctx, db) })

	result := completedResult(t, "view-a", now)
	second := completedResult(t, "view-b", now)
	resultIDA, resultIDB := result.Binding.ProjectionResultID, second.Binding.ProjectionResultID
	vertexIDA, vertexIDB, edgeIDA := result.Edges[0].SrcVertexID, result.Edges[0].DstVertexID, result.Edges[0].EdgeID
	publishInNestedTransaction(t, ctx, db, result)

	reader, err := postgresresult.NewReader(db)
	if err != nil {
		t.Fatalf("construct exact result reader: %v", err)
	}
	envelope, err := reader.ReadResultEnvelope(ctx, result.Binding.ProjectionResultID)
	if err != nil || envelope != result.Binding {
		t.Fatalf("read immutable envelope = %#v %v", envelope, err)
	}
	loaded, err := reader.ReadExactResult(ctx, result.Binding)
	if err != nil {
		t.Fatalf("read exact result: %v", err)
	}
	if len(loaded.Vertices) != 2 || len(loaded.Edges) != 1 || loaded.Vertices[0].VertexID != vertexIDA || loaded.Edges[0].EdgeID != edgeIDA {
		t.Fatalf("exact ordered result drifted: %#v", loaded)
	}
	mismatchedBinding := result.Binding
	mismatchedBinding.CanonicalOutputSHA256 = strings.Repeat("a", 64)
	if _, err := reader.ReadExactResult(ctx, mismatchedBinding); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatalf("exact read binding mismatch = %v", err)
	}
	missingBinding := second.Binding
	if _, err := reader.ReadExactResult(ctx, missingBinding); !errors.Is(err, graphprojection.ErrResultV2NotFound) {
		t.Fatalf("exact read missing result = %v", err)
	}
	if _, err := reader.ReadVertices(ctx, resultIDA, 1); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatalf("bounded vertex read = %v", err)
	}

	// Retried publication of byte-identical semantic output is idempotent even
	// when the operational publication time differs.
	retry := result
	retry.PublishedAt = now.Add(time.Minute)
	publishInNestedTransaction(t, ctx, db, retry)

	conflict := result
	conflict.ResultJSON = []byte(`{"projection_schema_id":"graph_projection.v2","changed":true}`)
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin identity conflict transaction: %v", err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct identity conflict publisher: %v", err)
	}
	if err := publisher.PublishResult(ctx, conflict); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatalf("same identity with different bytes got %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback identity conflict transaction: %v", err)
	}

	invalid := second
	invalid.Edges[0].EdgeKind = strings.Repeat("x", 256)
	tx, err = db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin atomicity transaction: %v", err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	publisher, err = postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct atomicity publisher: %v", err)
	}
	if err := publisher.PublishResult(ctx, invalid); err == nil {
		t.Fatal("database-inadmissible child unexpectedly published")
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback failed publication: %v", err)
	}
	var partialCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, resultIDB).Scan(&partialCount); err != nil || partialCount != 0 {
		t.Fatalf("failed publication left partial result count=%d err=%v", partialCount, err)
	}

	traversal, err := reader.Traverse(ctx, graphprojection.TraversalRequestV2{
		ProjectionResultID: resultIDA,
		SeedVertexIDs:      []string{vertexIDA},
		Direction:          graphprojection.TraversalOutgoingV2,
		MaximumDepth:       1,
		MaximumVertices:    2,
		MaximumEdges:       1,
	})
	if err != nil || len(traversal.Vertices) != 2 || len(traversal.Edges) != 1 || !reflect.DeepEqual(traversal.Vertices, result.Vertices) || !reflect.DeepEqual(traversal.Edges, result.Edges) {
		t.Fatalf("bounded traversal got %#v err=%v", traversal, err)
	}
	incoming, err := reader.Traverse(ctx, graphprojection.TraversalRequestV2{
		ProjectionResultID: resultIDA,
		SeedVertexIDs:      []string{vertexIDB},
		Direction:          graphprojection.TraversalIncomingV2,
		MaximumDepth:       1,
		MaximumVertices:    2,
		MaximumEdges:       1,
	})
	if err != nil || len(incoming.Vertices) != 2 || len(incoming.Edges) != 1 || !reflect.DeepEqual(incoming.Vertices, result.Vertices) || !reflect.DeepEqual(incoming.Edges, result.Edges) {
		t.Fatalf("incoming traversal got %#v err=%v", incoming, err)
	}
	seedOnly, err := reader.Traverse(ctx, graphprojection.TraversalRequestV2{
		ProjectionResultID: resultIDA,
		SeedVertexIDs:      []string{vertexIDA},
		Direction:          graphprojection.TraversalBothV2,
		MaximumDepth:       0,
		MaximumVertices:    1,
		MaximumEdges:       0,
		VertexKinds:        []string{"endpoint"},
		EdgeKinds:          []string{"flow"},
	})
	if err != nil || len(seedOnly.Vertices) != 1 || len(seedOnly.Edges) != 0 || seedOnly.Vertices[0].VertexID != vertexIDA {
		t.Fatalf("zero-depth filtered traversal got %#v err=%v", seedOnly, err)
	}
	if _, err := reader.Traverse(ctx, graphprojection.TraversalRequestV2{
		ProjectionResultID: resultIDA,
		SeedVertexIDs:      []string{vertexIDA},
		Direction:          graphprojection.TraversalOutgoingV2,
		MaximumDepth:       1,
		MaximumVertices:    1,
		MaximumEdges:       1,
	}); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatalf("one-over traversal result = %v", err)
	}

	tx, err = db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	cleaner, err := postgresresult.NewCleaner(tx)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := cleaner.LockCleanupCandidate(ctx, "network_flow_activity", nil)
	if err != nil || candidate == nil || candidate.ProjectionResultID != resultIDA || !candidate.PublishedAt.Equal(now) {
		t.Fatalf("ordered candidate: %#v %v", candidate, err)
	}
	if deleted, err := cleaner.DeleteLockedResult(ctx, resultIDA); err != nil || !deleted {
		t.Fatalf("locked deletion: %t %v", deleted, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func publishInNestedTransaction(t testing.TB, ctx context.Context, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, result graphprojection.CompletedResultV2) {
	t.Helper()
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin publication transaction: %v", err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct result publisher: %v", err)
	}
	if err := publisher.PublishResult(ctx, result); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("publish result: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit result: %v", err)
	}
}

func completedResult(t testing.TB, view string, now time.Time) graphprojection.CompletedResultV2 {
	t.Helper()
	return graphfixture.Result(t, graphfixture.ResultOptions{GraphViewID: view, SourceOwnerID: "network_flow_activity", SourceSnapshotID: "fixture-snapshot", ProjectionVersion: "fixture-v2", VertexKind: "endpoint", EdgeKind: "flow", VertexCount: 2, PublishedAt: now})
}

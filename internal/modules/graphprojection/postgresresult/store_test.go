package postgresresult_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const (
	resultIDA = "gpres_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	resultIDB = "gpres_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	vertexIDA = "vx_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	vertexIDB = "vx_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	edgeIDA   = "ed_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	digestA   = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	digestB   = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	digestC   = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func TestResultV2PublicationReadTraversalLeaseAndCleanup_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.BeginRollbackDBT(t, "graph-result-v2")
	ctx := context.Background()
	now := time.Date(2026, 8, 15, 19, 0, 0, 0, time.UTC)

	result := completedResult(resultIDA, now)
	publishInNestedTransaction(t, ctx, db, result)

	reader, err := postgresresult.NewReader(db)
	if err != nil {
		t.Fatalf("construct exact result reader: %v", err)
	}
	loaded, err := reader.ReadExactResult(ctx, result.Binding)
	if err != nil {
		t.Fatalf("read exact result: %v", err)
	}
	if len(loaded.Vertices) != 2 || len(loaded.Edges) != 1 || loaded.Vertices[0].VertexID != vertexIDA || loaded.Edges[0].EdgeID != edgeIDA {
		t.Fatalf("exact ordered result drifted: %#v", loaded)
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
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct identity conflict publisher: %v", err)
	}
	if err := publisher.PublishResult(ctx, conflict); !errors.Is(err, graphprojection.ErrResultV2IdentityConflict) {
		t.Fatalf("same identity with different bytes got %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback identity conflict transaction: %v", err)
	}

	invalid := completedResult(resultIDB, now)
	invalid.Edges[0].EdgeKind = strings.Repeat("x", 256)
	tx, err = db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin atomicity transaction: %v", err)
	}
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
	if err != nil || len(traversal.Vertices) != 2 || len(traversal.Edges) != 1 {
		t.Fatalf("bounded traversal got %#v err=%v", traversal, err)
	}

	leaseID := uuid.NewString()
	tx, err = db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin lease transaction: %v", err)
	}
	leaseWriter, err := postgresresult.NewLeaseWriter(tx)
	if err != nil {
		t.Fatalf("construct lease writer: %v", err)
	}
	lease, err := leaseWriter.AcquireLease(ctx, graphprojection.ResultLeaseV2{
		LeaseID:              leaseID,
		ProjectionResultID:   resultIDA,
		LeaseOwnerID:         "snapshot_reporting",
		LeaseOwnerResourceID: "release-1",
		LeasePurpose:         "render",
		CreatedAt:            now,
		RenewedAt:            now,
		LeasedUntil:          now.Add(time.Hour),
	})
	if err != nil || lease.LeaseID != leaseID {
		t.Fatalf("acquire result lease got %#v err=%v", lease, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit result lease: %v", err)
	}

	tx, err = db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin protected cleanup transaction: %v", err)
	}
	cleaner, err := postgresresult.NewCleaner(tx)
	if err != nil {
		t.Fatalf("construct protected cleaner: %v", err)
	}
	deletedLeases, hasMore, err := cleaner.DeleteExpiredLeases(ctx, now.Add(30*time.Minute), 1000)
	if err != nil || deletedLeases != 0 || hasMore {
		t.Fatalf("active lease cleanup got deleted=%d has_more=%t err=%v", deletedLeases, hasMore, err)
	}
	candidate, err := cleaner.LockCleanupCandidate(ctx, "network_flow_activity", nil)
	if err != nil || candidate != nil {
		t.Fatalf("active lease did not exclude result candidate: candidate=%#v err=%v", candidate, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit protected cleanup: %v", err)
	}

	tx, err = db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin expired cleanup transaction: %v", err)
	}
	cleaner, err = postgresresult.NewCleaner(tx)
	if err != nil {
		t.Fatalf("construct expired cleaner: %v", err)
	}
	deletedLeases, hasMore, err = cleaner.DeleteExpiredLeases(ctx, now.Add(2*time.Hour), 1000)
	if err != nil || deletedLeases != 1 || hasMore {
		t.Fatalf("expired lease cleanup got deleted=%d has_more=%t err=%v", deletedLeases, hasMore, err)
	}
	candidate, err = cleaner.LockCleanupCandidate(ctx, "network_flow_activity", nil)
	if err != nil || candidate == nil || candidate.ProjectionResultID != resultIDA || !candidate.PublishedAt.Equal(now) {
		t.Fatalf("oldest source-owner cleanup candidate got %#v err=%v", candidate, err)
	}
	leased, err := cleaner.HasUnexpiredLease(ctx, resultIDA, now.Add(2*time.Hour))
	if err != nil || leased {
		t.Fatalf("expired result lease recheck got leased=%t err=%v", leased, err)
	}
	deleted, err := cleaner.DeleteLockedResult(ctx, resultIDA)
	if err != nil || !deleted {
		t.Fatalf("locked result deletion got deleted=%t err=%v", deleted, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit expired cleanup: %v", err)
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

func completedResult(resultID string, now time.Time) graphprojection.CompletedResultV2 {
	return graphprojection.CompletedResultV2{
		Binding: graphprojection.ResultBindingV2{
			ProjectionResultID:            resultID,
			GraphViewID:                   "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			SourceOwnerID:                 "network_flow_activity",
			SourceSnapshotID:              "network-flow-snapshot-1",
			ProjectionSchemaID:            graphprojection.ProjectionSchemaIDV2,
			ProjectionVersion:             "network-flow-graph-v1",
			NormalizedConfigurationSHA256: digestA,
			NormalizedSourceSHA256:        digestB,
			CanonicalOutputSHA256:         digestC,
		},
		ResultJSON: []byte(`{"projection_schema_id":"graph_projection.v2"}`),
		Vertices: []graphprojection.ResultVertexV2{
			{VertexID: vertexIDA, VertexKind: "endpoint", SortKey: "a", JSON: []byte(`{"vertex_id":"` + vertexIDA + `"}`)},
			{VertexID: vertexIDB, VertexKind: "endpoint", SortKey: "b", JSON: []byte(`{"vertex_id":"` + vertexIDB + `"}`)},
		},
		Edges: []graphprojection.ResultEdgeV2{
			{EdgeID: edgeIDA, EdgeKind: "flow", SrcVertexID: vertexIDA, DstVertexID: vertexIDB, Direction: "directed", SortKey: "a", JSON: []byte(`{"edge_id":"` + edgeIDA + `"}`)},
		},
		PublishedAt: now,
	}
}

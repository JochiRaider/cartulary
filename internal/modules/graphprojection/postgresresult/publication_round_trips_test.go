package postgresresult_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/graphfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

// TestResultV2PublicationRoundTripsAreBoundedByObjects_Integration: publishing
// a new result costs a fixed number of round trips, whatever its vertex and
// edge counts, because Network Flow publishes inside the import transaction.
func TestResultV2PublicationRoundTripsAreBoundedByObjects_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.BeginRollbackDBT(t, "graph-result-round-trips")
	ctx := context.Background()
	now := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	publish := func(vertices int) int {
		t.Helper()
		result := graphfixture.Result(t, graphfixture.ResultOptions{
			GraphViewID: fmt.Sprintf("round-trips-%d", vertices), SourceOwnerID: "network_flow_activity",
			SourceSnapshotID: "fixture-snapshot", ProjectionVersion: "fixture-v2",
			VertexKind: "endpoint", EdgeKind: "flow", VertexCount: vertices, PublishedAt: now,
		})
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
		counter := dbassert.CountRoundTrips(tx)
		publisher, err := postgresresult.NewPublisher(counter)
		if err != nil {
			t.Fatal(err)
		}
		if err := publisher.PublishResult(ctx, result); err != nil {
			t.Fatalf("publish %d-vertex result: %v", vertices, err)
		}
		trips := counter.RoundTrips()
		reader, err := postgresresult.NewReader(tx)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := reader.ReadExactResult(ctx, result.Binding)
		if err != nil || len(loaded.Vertices) != vertices || len(loaded.Edges) != len(result.Edges) {
			t.Fatalf("published %d-vertex result reads back as %d vertices, %d edges: %v", vertices, len(loaded.Vertices), len(loaded.Edges), err)
		}
		for index, vertex := range loaded.Vertices {
			if vertex.VertexID != result.Vertices[index].VertexID {
				t.Fatalf("vertex %d is %s, want %s in published order", index, vertex.VertexID, result.Vertices[index].VertexID)
			}
		}
		return trips
	}
	few, many := publish(2), publish(40)
	if few != many {
		t.Fatalf("result publication round trips grow with objects: %d for 2 vertices, %d for 40", few, many)
	}
}

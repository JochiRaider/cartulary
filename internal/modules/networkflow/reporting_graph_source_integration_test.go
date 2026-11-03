package networkflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/graphfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
)

func TestReportingGraphSourceRetainsSelectedInput_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "network-flow-reporting-frozen-input")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	actor := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: "frozen-graph-input@example.test", DisplayName: "Frozen Graph Input"})
	incident := appsupport.CreateIncidentInStore(t, pool, actor, "txn-frozen-graph", "IR-FROZENGRAPH", "Frozen Graph Input")
	incidentID := incident.ID
	store := NewStore(pool, DefaultEffectiveLimits(), WithResourceIntentAppender(networkflowsupport.NewResourceIntentAppender()))
	now := time.Date(2026, 8, 15, 21, 0, 0, 0, time.UTC)
	graphViewID := "nfgv_cccccccccccccccccccccccccccccccc"
	result := graphResultForReportingSource(t, graphViewID, now)
	declaration := graphViewDeclarationFixture(graphViewID, incidentID, actor.ID, now)
	declaration.DesiredSourceSnapshotID = result.Binding.SourceSnapshotID
	declaration.SelectedResult = &GraphViewSelectedResultBinding{
		ProjectionResultID:            result.Binding.ProjectionResultID,
		SourceSnapshotID:              result.Binding.SourceSnapshotID,
		ProjectionSchemaID:            result.Binding.ProjectionSchemaID,
		ProjectionVersion:             result.Binding.ProjectionVersion,
		NormalizedConfigurationSHA256: result.Binding.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        result.Binding.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         result.Binding.CanonicalOutputSHA256,
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin fixture transaction: %v", err)
	}
	defer tx.Rollback(ctx)
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatalf("construct result publisher: %v", err)
	}
	if err := publisher.PublishResult(ctx, result); err != nil {
		t.Fatalf("publish exact graph result: %v", err)
	}
	if err := store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
		t.Fatalf("insert selected graph declaration: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit graph source fixture: %v", err)
	}

	ready := true
	source, err := NewReportingGraphSource(store, func() bool { return ready })
	if err != nil {
		t.Fatalf("construct Reporting graph source: %v", err)
	}
	tx, err = pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	admitted, err := source.ReadSelectedResultTx(ctx, tx, incidentID, result.Binding)
	if err != nil {
		t.Fatal(err)
	}
	if admitted.Projection.Binding != result.Binding || len(admitted.Projection.Vertices) != 2 || len(admitted.Projection.Edges) != 1 || len(admitted.LabelCandidates.VertexLabelCandidates) != 2 || len(admitted.LabelCandidates.EdgeLabelCandidates) != 1 {
		t.Fatal("incomplete owner contribution")
	}
	expectedEndpoints := map[string]bool{"192.0.2.1": true, "192.0.2.2": true}
	for index, candidate := range admitted.LabelCandidates.VertexLabelCandidates {
		if candidate.ProjectedVertexID != result.Vertices[index].VertexID || !expectedEndpoints[candidate.Endpoint.StringValue] {
			t.Fatal("source labels lost current producer identity or ordering")
		}
		delete(expectedEndpoints, candidate.Endpoint.StringValue)
	}
	if len(expectedEndpoints) != 0 {
		t.Fatal("source omitted endpoint values")
	}
	retainedFirstEndpoint := admitted.LabelCandidates.VertexLabelCandidates[0].Endpoint.StringValue
	cleanup, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup.Rollback(ctx)
	cleaner, err := postgresresult.NewCleaner(cleanup)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := cleaner.LockCleanupCandidate(ctx, ProfileID, nil)
	if err != nil || candidate != nil {
		t.Fatalf("cleanup crossed admission lock: %#v %v", candidate, err)
	}
	if err := cleanup.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	for _, badIncident := range []uuid.UUID{uuid.Nil} {
		if _, err := source.ReadSelectedResultTx(ctx, tx, badIncident, result.Binding); !errors.Is(err, graphprojection.ErrResultV2BindingMismatch) {
			t.Fatalf("invalid selection: %v", err)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := source.ReadSelectedResultTx(cancelled, tx, incidentID, result.Binding); !errors.Is(err, context.Canceled) {
		t.Fatalf("caller cancellation: %v", err)
	}
	var stillUsable int
	if err := tx.QueryRow(ctx, "SELECT 1").Scan(&stillUsable); err != nil || stillUsable != 1 {
		t.Fatal("rejected selection changed caller transaction", err)
	}
	stale := result.Binding
	stale.SourceSnapshotID = "stale"
	if _, err := source.ReadSelectedResultTx(ctx, tx, incidentID, stale); !errors.Is(err, graphprojection.ErrResultV2BindingMismatch) {
		t.Fatalf("malformed tuple precedence: %v", err)
	}
	if _, err := source.ReadSelectedResultTx(ctx, tx, uuid.New(), result.Binding); !errors.Is(err, graphprojection.ErrResultV2NotFound) {
		t.Fatalf("cross-incident input: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	ready = false
	if _, err := pool.Exec(ctx, `DELETE FROM graph_projection_results WHERE projection_result_id=$1`, result.Binding.ProjectionResultID); err != nil {
		t.Fatal(err)
	}
	if admitted.Projection.Binding != result.Binding || len(admitted.Projection.Vertices) != 2 || admitted.LabelCandidates.VertexLabelCandidates[0].Endpoint.StringValue != retainedFirstEndpoint {
		t.Fatal("admitted contribution lost after derived-state removal")
	}
	tx, err = pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := source.ReadSelectedResultTx(ctx, tx, incidentID, result.Binding); err == nil {
		t.Fatal("new input admitted with source unavailable")
	}
}

func graphResultForReportingSource(t testing.TB, graphViewID string, now time.Time) graphprojection.CompletedResultV2 {
	t.Helper()
	return graphResultForOwner(t, graphViewID, ProfileID, now)
}

func graphResultForOwner(t testing.TB, graphViewID, owner string, now time.Time) graphprojection.CompletedResultV2 {
	t.Helper()
	return graphfixture.Result(t, graphfixture.ResultOptions{GraphViewID: graphViewID, SourceOwnerID: owner, SourceSnapshotID: "nfsnap_" + testSHA1, ProjectionVersion: "network_flow_activity.v1", VertexKind: "network_flow.ip_endpoint.v1", EdgeKind: "network_flow.flow_edge.v1", VertexCount: 2, PublishedAt: now})
}

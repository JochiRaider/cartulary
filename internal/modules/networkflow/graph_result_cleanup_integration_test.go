package networkflow_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestNetworkFlowGraphResultCleanupIsOwnerScopedSelectedSafeAndBounded_Integration(t *testing.T) {
	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-graph-result-cleanup")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	service, err := NewGraphResultCleanupService(harness.DB, store)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Date(2026, 8, 16, 16, 0, 0, 0, time.UTC)
	graphViewID := "nfgv_cccccccccccccccccccccccccccccccc"

	selectedResult := graphResultForReportingSource(graphViewID, now.Add(-3*time.Hour))
	unselectedResult := selectedResult
	unselectedResult.Binding.ProjectionResultID = "gpres_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	unselectedResult.PublishedAt = now.Add(-2 * time.Hour)
	otherOwnerResult := selectedResult
	otherOwnerResult.Binding.ProjectionResultID = "gpres_cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	otherOwnerResult.Binding.SourceOwnerID = "other_projection_owner"
	otherOwnerResult.PublishedAt = now.Add(-4 * time.Hour)

	declaration := graphViewDeclarationFixture(graphViewID, incidentID, actor.ID, now.Add(-4*time.Hour))
	declaration.SelectedResult = &GraphViewSelectedResultBinding{
		ProjectionResultID:            selectedResult.Binding.ProjectionResultID,
		SourceSnapshotID:              selectedResult.Binding.SourceSnapshotID,
		ProjectionSchemaID:            selectedResult.Binding.ProjectionSchemaID,
		ProjectionVersion:             selectedResult.Binding.ProjectionVersion,
		NormalizedConfigurationSHA256: selectedResult.Binding.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        selectedResult.Binding.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         selectedResult.Binding.CanonicalOutputSHA256,
	}
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []graphprojection.CompletedResultV2{selectedResult, unselectedResult, otherOwnerResult} {
		if err := publisher.PublishResult(ctx, result); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("publish cleanup fixture: %v", err)
		}
	}
	if err := store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("insert selected cleanup declaration: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	result, err := service.SweepGraphResults(ctx, now, nil)
	if err != nil || result.Examined != 2 || result.DeletedResults != 1 || !result.Exhausted || result.HasMore ||
		!result.HealthSnapshotValid || result.EligibleResultBacklog != 0 || result.OldestEligibleResultAge != nil {
		t.Fatalf("source-owner cleanup result = %#v err=%v", result, err)
	}
	requireGraphResultCount(t, harness.DB, selectedResult.Binding.ProjectionResultID, 1)
	requireGraphResultCount(t, harness.DB, unselectedResult.Binding.ProjectionResultID, 0)
	requireGraphResultCount(t, harness.DB, otherOwnerResult.Binding.ProjectionResultID, 1)

	bounded := make([]graphprojection.CompletedResultV2, 0, 9)
	for index := 0; index < 9; index++ {
		candidate := selectedResult
		candidate.Binding.ProjectionResultID = "gpres_" + fmt.Sprintf("%064x", index+16)
		candidate.PublishedAt = now.Add(time.Duration(index) * time.Second)
		bounded = append(bounded, candidate)
	}
	tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	publisher, _ = postgresresult.NewPublisher(tx)
	for _, candidate := range bounded {
		if err := publisher.PublishResult(ctx, candidate); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("publish bounded cleanup fixture: %v", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	cursor := &graphprojection.ResultCleanupCandidateV2{
		ProjectionResultID: selectedResult.Binding.ProjectionResultID,
		PublishedAt:        selectedResult.PublishedAt,
	}
	first, err := service.SweepGraphResults(ctx, now.Add(time.Minute), cursor)
	if err != nil || first.Examined != 8 || first.DeletedResults != 8 || !first.HasMore || first.NextCursor == nil ||
		!first.HealthSnapshotValid || first.EligibleResultBacklog != 1 || first.OldestEligibleResultAge == nil ||
		*first.OldestEligibleResultAge != 52*time.Second {
		t.Fatalf("first bounded cleanup = %#v err=%v", first, err)
	}
	second, err := service.SweepGraphResults(ctx, now.Add(time.Minute), first.NextCursor)
	if err != nil || second.Examined != 1 || second.DeletedResults != 1 || second.HasMore || !second.Exhausted ||
		!second.HealthSnapshotValid || second.EligibleResultBacklog != 0 || second.OldestEligibleResultAge != nil {
		t.Fatalf("continued bounded cleanup = %#v err=%v", second, err)
	}
	for _, candidate := range bounded {
		requireGraphResultCount(t, harness.DB, candidate.Binding.ProjectionResultID, 0)
	}

	SetGraphResultCleanupTestBounds(service, 8, time.Nanosecond)
	timeBounded, err := service.SweepGraphResults(ctx, now.Add(2*time.Minute), nil)
	if err != nil || timeBounded.Examined != 0 || !timeBounded.HasMore || timeBounded.Exhausted {
		t.Fatalf("time-bounded cleanup = %#v err=%v", timeBounded, err)
	}
	SetGraphResultCleanupTestBounds(service, 8, 30*time.Second)

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := service.SweepGraphResults(canceled, now, nil); err == nil {
		t.Fatal("canceled cleanup unexpectedly succeeded")
	}
	requireGraphResultCount(t, harness.DB, selectedResult.Binding.ProjectionResultID, 1)
}

func requireGraphResultCount(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, resultID string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM graph_projection_results WHERE projection_result_id = $1`, resultID).Scan(&got); err != nil {
		t.Fatalf("count graph result %s: %v", resultID, err)
	}
	if got != want {
		t.Fatalf("graph result %s count=%d want=%d", resultID, got, want)
	}
}

package graphprojection_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/fixturetest"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestGPFIX004Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-004")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix004"), fixedTime)
	input := invalidReverseEdgeInput(t, "gpfix004")
	run, err := store.CreateProjection(context.Background(), mustJSON(t, input), RetainedProjectionOptions{ProjectionRunNonce: "gpfix004", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if run.State != RunStateFailed || run.GraphView != nil || run.ValidationSummary.Status != "failed" {
		t.Fatalf("failed run = %#v", run)
	}
	reloaded, err := store.GetProjectionRun(context.Background(), run.ProjectionRunID)
	if err != nil || reloaded.State != RunStateFailed || reloaded.GraphView != nil {
		t.Fatalf("reloaded failed run = %#v err=%v", reloaded, err)
	}
}

func TestGPFIX018Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-018")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix018"), fixedTime)
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	first, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime(), IdempotencyKey: "idem"})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime(), IdempotencyKey: "idem"})
	if err != nil || replayed.ProjectionRunID != first.ProjectionRunID || replayed.AcceptedReplay == nil {
		t.Fatalf("replay = %#v err=%v", replayed, err)
	}
	other := incidentGraphInput(t)
	other["source_metadata"].(map[string]any)["case"] = "beta"
	_, err = store.CreateProjection(ctx, mustJSON(t, other), RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-b", AcceptedAt: fixedTime(), GeneratedAt: fixedTime(), IdempotencyKey: "idem"})
	assertStoreOperationError(t, err, "operation_conflict", "idempotency_key_conflict")
	_, err = store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-a", AcceptedAt: fixedTime().Add(24 * time.Hour), GeneratedAt: fixedTime().Add(24 * time.Hour), IdempotencyKey: "idem"})
	assertStoreOperationError(t, err, "invalid_operation", "graph_view_already_exists")
}

func TestGPFIX019Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-019")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix019"), fixedTime)
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	first, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix019-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RefreshProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix019-b", AcceptedAt: fixedTime().Add(time.Second), GeneratedAt: fixedTime().Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	summary, err := store.InvalidateGraphView(ctx, RetainedInvalidation{GraphViewID: first.GraphViewID, ReasonCode: "security_withdrawal", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "fixture", InvalidatedAt: fixedTime().Add(2 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if summary.GraphViewStateAfter != GraphViewStateInvalidated || len(summary.InvalidatedRunIDs) != 2 || summary.InvalidatedRunIDs[0] > summary.InvalidatedRunIDs[1] {
		t.Fatalf("invalidation summary = %#v", summary)
	}
}

func TestGPFIX020Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-020")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix020"), fixedTime)
	ctx := context.Background()
	available, err := store.CreateProjection(ctx, mustJSON(t, retargetGraphInput(t, incidentGraphInput(t), "gpfix020_available")), RetainedProjectionOptions{ProjectionRunNonce: "gpfix020-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	failed, err := store.CreateProjection(ctx, mustJSON(t, invalidReverseEdgeInput(t, "gpfix020_failed")), RetainedProjectionOptions{ProjectionRunNonce: "gpfix020-f", AcceptedAt: fixedTime().Add(time.Second), GeneratedAt: fixedTime().Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.InvalidateGraphView(ctx, RetainedInvalidation{GraphViewID: available.GraphViewID, ReasonCode: "operator_requested", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "fixture", InvalidatedAt: fixedTime().Add(2 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	summaries, _, err := store.ListGraphViews(ctx, ListGraphViewsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !summaryState(summaries, available.GraphViewID, GraphViewStateInvalidated) || !summaryState(summaries, failed.GraphViewID, GraphViewStateFailed) {
		t.Fatalf("summaries = %#v", summaries)
	}
}

func TestGPFIX021Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-021")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix021"), fixedTime)
	ctx := context.Background()
	input := incidentGraphInput(t)
	input["projection_config"].(map[string]any)["retention_policy"] = map[string]any{"retain_replaced_results": true, "retention_count": 0, "retention_duration_seconds": 2592000, "retain_failed_results": true, "failed_retention_count": 20, "failed_retention_duration_seconds": 2592000}
	first, err := store.CreateProjection(ctx, mustJSON(t, input), RetainedProjectionOptions{ProjectionRunNonce: "gpfix021-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RefreshProjection(ctx, mustJSON(t, input), RetainedProjectionOptions{ProjectionRunNonce: "gpfix021-b", AcceptedAt: fixedTime().Add(time.Second), GeneratedAt: fixedTime().Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetProjectionRun(ctx, first.ProjectionRunID); !errors.Is(err, ErrProjectionRunNotFound) {
		t.Fatalf("expired replaced lookup err = %v", err)
	}
}

func TestGPFIX029Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-029")
	db := pgtest.Start(t).BeginRollbackDBT(t, "gpfix029")
	baseStore := postgresstore.NewWithClock(db, fixedTime)
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	created, err := baseStore.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix029-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	paused := make(chan struct{})
	resume := make(chan struct{})
	refreshStore := postgresstore.NewWithClockAndHooks(db, fixedTime, postgresstore.Hooks{BeforePublication: func(ctx context.Context, run ProjectionRun) error {
		close(paused)
		select {
		case <-resume:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
	resultCh := make(chan ProjectionRun, 1)
	errCh := make(chan error, 1)
	go func() {
		run, err := refreshStore.RefreshProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix029-b", AcceptedAt: fixedTime().Add(time.Second), GeneratedAt: fixedTime().Add(time.Second)})
		resultCh <- run
		errCh <- err
	}()
	<-paused
	if _, err := baseStore.InvalidateGraphView(ctx, RetainedInvalidation{GraphViewID: created.GraphViewID, ReasonCode: "security_withdrawal", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "fixture", InvalidatedAt: fixedTime().Add(2 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	close(resume)
	refreshed := <-resultCh
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	if refreshed.State != RunStateInvalidated || refreshed.GraphView != nil {
		t.Fatalf("in-flight invalidated run = %#v", refreshed)
	}
}

func TestGPFIX030Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-030")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix030"), fixedTime)
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	created, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix030-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.InvalidateGraphView(ctx, RetainedInvalidation{GraphViewID: created.GraphViewID, ReasonCode: "security_withdrawal", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "fixture", InvalidatedAt: fixedTime().Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	failedInput := invalidReverseEdgeInput(t, "incident_graph")
	run, err := store.RefreshProjection(ctx, mustJSON(t, failedInput), RetainedProjectionOptions{ProjectionRunNonce: "gpfix030-b", AcceptedAt: fixedTime().Add(2 * time.Second), GeneratedAt: fixedTime().Add(2 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if run.State != RunStateFailed {
		t.Fatalf("refresh run = %#v", run)
	}
	if _, err := store.GetGraphView(ctx, created.GraphViewID, ""); !errors.Is(err, ErrGraphViewUnavailable) {
		t.Fatalf("invalidated view lookup err = %v", err)
	}
}

func TestGPFIX031Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-031")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix031"), fixedTime)
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	created, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix031-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.InvalidateGraphView(ctx, RetainedInvalidation{GraphViewID: created.GraphViewID, ReasonCode: "operator_requested", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "fixture", InvalidatedAt: fixedTime().Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix031-b", AcceptedAt: fixedTime().Add(2 * time.Second), GeneratedAt: fixedTime().Add(2 * time.Second)})
	assertStoreOperationError(t, err, "invalid_operation", "graph_view_already_exists")
}

func TestGPFIX032Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-032")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix032"), fixedTime)
	ctx := context.Background()
	input := incidentGraphInput(t)
	input["projection_config"].(map[string]any)["retention_policy"] = map[string]any{"retain_replaced_results": true, "retention_count": 1, "retention_duration_seconds": 2592000, "retain_failed_results": true, "failed_retention_count": 20, "failed_retention_duration_seconds": 2592000}
	first, err := store.CreateProjection(ctx, mustJSON(t, input), RetainedProjectionOptions{ProjectionRunNonce: "gpfix032-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.RefreshProjection(ctx, mustJSON(t, input), RetainedProjectionOptions{ProjectionRunNonce: "gpfix032-b", AcceptedAt: fixedTime().Add(time.Second), GeneratedAt: fixedTime().Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	summary, err := store.InvalidateGraphView(ctx, RetainedInvalidation{GraphViewID: first.GraphViewID, ReasonCode: "operator_requested", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "fixture", InvalidatedAt: fixedTime().Add(2 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.InvalidatedRunIDs) != 2 {
		t.Fatalf("invalidated run IDs = %#v", summary.InvalidatedRunIDs)
	}
	if selected, err := store.GetProjectionRun(ctx, second.ProjectionRunID); err != nil || selected.State != RunStateInvalidated || selected.RetentionExpiresAt != nil {
		t.Fatalf("selected invalidated run = %#v err=%v", selected, err)
	}
}

func TestGPFIX033Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-033")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix033"), fixedTime)
	ctx := context.Background()
	input := invalidReverseEdgeInput(t, "gpfix033")
	failed, err := store.CreateProjection(ctx, mustJSON(t, input), RetainedProjectionOptions{ProjectionRunNonce: "gpfix033-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.RefreshProjection(ctx, mustJSON(t, input), RetainedProjectionOptions{ProjectionRunNonce: "gpfix033-b", AcceptedAt: fixedTime().Add(time.Second), GeneratedAt: fixedTime().Add(time.Second)})
	assertStoreOperationError(t, err, "invalid_operation", "no_consumable_prior_run")
	reloaded, err := store.GetProjectionRun(ctx, failed.ProjectionRunID)
	if err != nil || reloaded.State != RunStateFailed {
		t.Fatalf("failed run changed = %#v err=%v", reloaded, err)
	}
}

func TestGPFIX035Remediation(t *testing.T) {
	assertStoreFixtureLoaded(t, "GP-FIX-035")
	store := postgresstore.NewWithClock(pgtest.Start(t).BeginRollbackDBT(t, "gpfix035"), fixedTime)
	if _, _, err := store.ListGraphViews(context.Background(), ListGraphViewsOptions{CursorToken: stringsRepeat("x", ResourceLimits().MaxCursorTokenLength+1)}); !IsOperationError(err, "cursor_invalid", "cursor_token_too_long") {
		t.Fatalf("oversized cursor err = %v", err)
	}
	if _, _, err := store.ListGraphViews(context.Background(), ListGraphViewsOptions{CursorToken: "not-a-valid-cursor"}); !IsOperationError(err, "cursor_invalid", "malformed") {
		t.Fatalf("malformed cursor err = %v", err)
	}
}

func assertStoreFixtureLoaded(t *testing.T, fixtureID string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := fixturetest.RepoRoot(wd)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := fixturetest.Load(root, fixtureID); err != nil {
		t.Fatalf("load %s: %v", fixtureID, err)
	}
}

func invalidReverseEdgeInput(t *testing.T, key string) map[string]any {
	t.Helper()
	input := retargetGraphInput(t, incidentGraphInput(t), key)
	relMappings := input["projection_config"].(map[string]any)["relationship_mappings"].([]any)
	relMappings[0].(map[string]any)["direction_policy"] = "preserve"
	relMappings[0].(map[string]any)["emit_reverse_edge"] = true
	return input
}

func assertStoreOperationError(t *testing.T, err error, code, reason string) {
	t.Helper()
	var operationError *OperationError
	if !errors.As(err, &operationError) || operationError.Code != code || operationError.ReasonCode != reason {
		t.Fatalf("operation error = %#v want %s/%s", err, code, reason)
	}
}

func summaryState(summaries []GraphViewSummary, graphViewID string, state GraphViewState) bool {
	for _, summary := range summaries {
		if summary.GraphViewID == graphViewID {
			return summary.State == state
		}
	}
	return false
}

func stringsRepeat(value string, count int) string {
	out := ""
	for i := 0; i < count; i++ {
		out += value
	}
	return out
}

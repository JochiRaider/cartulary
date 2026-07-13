package graphprojection_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/fixturetest"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestGPFIX004Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-004") }
func TestGPFIX018Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-018") }
func TestGPFIX019Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-019") }
func TestGPFIX020Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-020") }
func TestGPFIX021Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-021") }
func TestGPFIX029Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-029") }
func TestGPFIX030Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-030") }
func TestGPFIX031Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-031") }
func TestGPFIX032Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-032") }
func TestGPFIX033Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-033") }
func TestGPFIX035Remediation(t *testing.T) { verifyStoreFixture(t, "GP-FIX-035") }

func verifyStoreFixture(t *testing.T, fixtureID string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := fixturetest.RepoRoot(wd)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixturetest.Verify(root, fixtureID, storeFixtureExecutor{t: t}); err != nil {
		t.Fatal(err)
	}
}

type storeFixtureExecutor struct {
	t *testing.T
}

func (executor storeFixtureExecutor) ExecuteFixtureStep(manifest fixturetest.Manifest, step fixturetest.Step, input []byte) (fixturetest.StepExecution, error) {
	if manifest.ExecutionLayer != "backend_store" {
		return fixturetest.StepExecution{}, fmt.Errorf("fixture %s is not backend_store", manifest.FixtureID)
	}
	var declared struct {
		FixtureID string `json:"fixture_id"`
		Scenario  string `json:"scenario"`
	}
	if err := json.Unmarshal(input, &declared); err != nil {
		return fixturetest.StepExecution{}, err
	}
	if declared.FixtureID != manifest.FixtureID || declared.Scenario == "" {
		return fixturetest.StepExecution{}, fmt.Errorf("fixture %s input does not declare a semantic scenario", manifest.FixtureID)
	}
	state := executor.execute(manifest.FixtureID)
	artifact, err := json.Marshal(state)
	if err != nil {
		return fixturetest.StepExecution{}, err
	}
	return fixturetest.StepExecution{Artifact: artifact, StateEffectMode: "retained_state_change"}, nil
}

func (executor storeFixtureExecutor) execute(fixtureID string) map[string]any {
	switch fixtureID {
	case "GP-FIX-004":
		return executor.gpfix004()
	case "GP-FIX-018":
		return executor.gpfix018()
	case "GP-FIX-019":
		return executor.gpfix019()
	case "GP-FIX-020":
		return executor.gpfix020()
	case "GP-FIX-021":
		return executor.gpfix021()
	case "GP-FIX-029":
		return executor.gpfix029()
	case "GP-FIX-030":
		return executor.gpfix030()
	case "GP-FIX-031":
		return executor.gpfix031()
	case "GP-FIX-032":
		return executor.gpfix032()
	case "GP-FIX-033":
		return executor.gpfix033()
	case "GP-FIX-035":
		return executor.gpfix035()
	default:
		executor.t.Fatalf("unsupported store fixture %s", fixtureID)
		return nil
	}
}

func (executor storeFixtureExecutor) gpfix004() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix004"), fixedTime, postgresstore.Hooks{})
	run, err := store.CreateProjection(context.Background(), mustJSON(t, invalidReverseEdgeInput(t, "gpfix004")), RetainedProjectionOptions{ProjectionRunNonce: "gpfix004", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := store.GetProjectionRun(context.Background(), run.ProjectionRunID)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]any{
		"fixture_id":               "GP-FIX-004",
		"create_state":             run.State,
		"create_graph_retained":    run.GraphView != nil,
		"create_validation_status": run.ValidationSummary.Status,
		"reloaded_state":           reloaded.State,
		"reloaded_graph_retained":  reloaded.GraphView != nil,
	}
}

func (executor storeFixtureExecutor) gpfix018() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix018"), fixedTime, postgresstore.Hooks{})
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	first, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime(), IdempotencyKey: "idem"})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime(), IdempotencyKey: "idem"})
	if err != nil {
		t.Fatal(err)
	}
	other := incidentGraphInput(t)
	other["source_metadata"].(map[string]any)["case"] = "beta"
	_, conflictErr := store.CreateProjection(ctx, mustJSON(t, other), RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-b", AcceptedAt: fixedTime(), GeneratedAt: fixedTime(), IdempotencyKey: "idem"})
	_, expiredErr := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix018-c", AcceptedAt: fixedTime().Add(24 * time.Hour), GeneratedAt: fixedTime().Add(24 * time.Hour), IdempotencyKey: "idem"})
	return map[string]any{
		"fixture_id":              "GP-FIX-018",
		"initial_state":           first.State,
		"replay_same_run":         replayed.ProjectionRunID == first.ProjectionRunID,
		"replay_summary_returned": replayed.AcceptedReplay != nil,
		"conflict_error":          lifecycleCodeReason(conflictErr),
		"expired_key_error":       lifecycleCodeReason(expiredErr),
	}
}

func (executor storeFixtureExecutor) gpfix019() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix019"), fixedTime, postgresstore.Hooks{})
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
	return map[string]any{
		"fixture_id":               "GP-FIX-019",
		"graph_view_state_after":   summary.GraphViewStateAfter,
		"invalidated_run_count":    len(summary.InvalidatedRunIDs),
		"invalidated_ids_ordered":  idsOrdered(summary.InvalidatedRunIDs),
		"target_scope":             summary.TargetScope,
		"invalidation_reason_code": summary.ReasonCode,
	}
}

func (executor storeFixtureExecutor) gpfix020() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix020"), fixedTime, postgresstore.Hooks{})
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
	return map[string]any{
		"fixture_id":              "GP-FIX-020",
		"available_summary_state": summaryState(summaries, available.GraphViewID),
		"failed_summary_state":    summaryState(summaries, failed.GraphViewID),
		"summary_count":           len(summaries),
	}
}

func (executor storeFixtureExecutor) gpfix021() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix021"), fixedTime, postgresstore.Hooks{})
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
	_, err = store.GetProjectionRun(ctx, first.ProjectionRunID)
	return map[string]any{
		"fixture_id":                 "GP-FIX-021",
		"expired_replaced_not_found": errors.Is(err, ErrProjectionRunNotFound),
	}
}

func (executor storeFixtureExecutor) gpfix029() map[string]any {
	t := executor.t
	db := pgtest.Start(t).BeginRollbackDBT(t, "gpfix029")
	baseStore := mustPostgresStore(t, db, fixedTime, postgresstore.Hooks{})
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	created, err := baseStore.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix029-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	paused := make(chan struct{})
	resume := make(chan struct{})
	refreshStore := mustPostgresStore(t, db, fixedTime, postgresstore.Hooks{BeforePublication: func(ctx context.Context, run ProjectionRun) error {
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
	return map[string]any{
		"fixture_id":             "GP-FIX-029",
		"refresh_state":          refreshed.State,
		"refresh_graph_retained": refreshed.GraphView != nil,
		"selected_read_error":    queryCodeReason(func() error { _, err := baseStore.GetGraphView(ctx, created.GraphViewID, ""); return err }()),
	}
}

func (executor storeFixtureExecutor) gpfix030() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix030"), fixedTime, postgresstore.Hooks{})
	ctx := context.Background()
	input := mustJSON(t, incidentGraphInput(t))
	created, err := store.CreateProjection(ctx, input, RetainedProjectionOptions{ProjectionRunNonce: "gpfix030-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.InvalidateGraphView(ctx, RetainedInvalidation{GraphViewID: created.GraphViewID, ReasonCode: "security_withdrawal", RequestedAt: "2026-05-30T00:00:00Z", RequestedBy: "fixture", InvalidatedAt: fixedTime().Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	run, err := store.RefreshProjection(ctx, mustJSON(t, invalidReverseEdgeInput(t, "incident_graph")), RetainedProjectionOptions{ProjectionRunNonce: "gpfix030-b", AcceptedAt: fixedTime().Add(2 * time.Second), GeneratedAt: fixedTime().Add(2 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	_, readErr := store.GetGraphView(ctx, created.GraphViewID, "")
	return map[string]any{
		"fixture_id":          "GP-FIX-030",
		"refresh_state":       run.State,
		"selected_read_error": queryCodeReason(readErr),
	}
}

func (executor storeFixtureExecutor) gpfix031() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix031"), fixedTime, postgresstore.Hooks{})
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
	return map[string]any{
		"fixture_id": "GP-FIX-031",
		"error":      lifecycleCodeReason(err),
	}
}

func (executor storeFixtureExecutor) gpfix032() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix032"), fixedTime, postgresstore.Hooks{})
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
	selected, err := store.GetProjectionRun(ctx, second.ProjectionRunID)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]any{
		"fixture_id":                         "GP-FIX-032",
		"invalidated_run_count":              len(summary.InvalidatedRunIDs),
		"selected_state":                     selected.State,
		"selected_retention_expires_at_null": selected.RetentionExpiresAt == nil,
	}
}

func (executor storeFixtureExecutor) gpfix033() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix033"), fixedTime, postgresstore.Hooks{})
	ctx := context.Background()
	failed, err := store.CreateProjection(ctx, mustJSON(t, invalidReverseEdgeInput(t, "gpfix033")), RetainedProjectionOptions{ProjectionRunNonce: "gpfix033-a", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	_, refreshErr := store.RefreshProjection(ctx, mustJSON(t, invalidReverseEdgeInput(t, "gpfix033")), RetainedProjectionOptions{ProjectionRunNonce: "gpfix033-b", AcceptedAt: fixedTime().Add(time.Second), GeneratedAt: fixedTime().Add(time.Second)})
	reloaded, err := store.GetProjectionRun(ctx, failed.ProjectionRunID)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]any{
		"fixture_id":       "GP-FIX-033",
		"refresh_error":    lifecycleCodeReason(refreshErr),
		"failed_run_state": reloaded.State,
	}
}

func (executor storeFixtureExecutor) gpfix035() map[string]any {
	t := executor.t
	store := mustPostgresStore(t, pgtest.Start(t).BeginRollbackDBT(t, "gpfix035"), fixedTime, postgresstore.Hooks{})
	_, _, oversizedErr := store.ListGraphViews(context.Background(), ListGraphViewsOptions{CursorToken: strings.Repeat("x", ResourceLimits().MaxCursorTokenLength+1)})
	_, _, malformedErr := store.ListGraphViews(context.Background(), ListGraphViewsOptions{CursorToken: "not-a-valid-cursor"})
	return map[string]any{
		"fixture_id":      "GP-FIX-035",
		"oversized_error": queryCodeReason(oversizedErr),
		"malformed_error": queryCodeReason(malformedErr),
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

func lifecycleCodeReason(err error) map[string]any {
	var lifecycleErr *LifecycleError
	if errors.As(err, &lifecycleErr) {
		return map[string]any{"code": lifecycleErr.Code, "reason_code": lifecycleErr.ReasonCode}
	}
	return map[string]any{"code": fmt.Sprintf("%T", err), "reason_code": ""}
}

func queryCodeReason(err error) map[string]any {
	var queryErr *QueryError
	if errors.As(err, &queryErr) {
		return map[string]any{"code": queryErr.Code, "reason_code": queryErr.ReasonCode}
	}
	if errors.Is(err, ErrGraphViewUnavailable) {
		return map[string]any{"code": "projection_not_available", "reason_code": ""}
	}
	return map[string]any{"code": fmt.Sprintf("%T", err), "reason_code": ""}
}

func idsOrdered(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] > values[index] {
			return false
		}
	}
	return true
}

func summaryState(summaries []GraphViewSummary, graphViewID string) GraphViewState {
	for _, summary := range summaries {
		if summary.GraphViewID == graphViewID {
			return summary.State
		}
	}
	return ""
}

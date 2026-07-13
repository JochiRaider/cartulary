package graphprojection

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGraphProjectionErrorEnvelopesAreClosedAndOrdered(t *testing.T) {
	lifecycle := &LifecycleError{Code: "invalid_operation_request", ReasonCode: "invalid_value", Details: map[string]any{"reason_code": "invalid_value", "field": "requested_at", "operation": "invalidate_graph_view"}}
	body, err := lifecycle.EnvelopeJSON()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"status":"error","error":{"code":"invalid_operation_request","retryable":false,"details":{"operation":"invalidate_graph_view","field":"requested_at","reason_code":"invalid_value"}}}`
	if string(body) != want {
		t.Fatalf("lifecycle envelope = %s", body)
	}
	query := &QueryError{Code: "projection_not_available", ReasonCode: "computing", Details: map[string]any{"state": "computing", "graph_view_id": "gv_test"}}
	body, err = query.EnvelopeJSON()
	if err != nil {
		t.Fatal(err)
	}
	want = `{"status":"error","error":{"code":"projection_not_available","retryable":true,"details":{"graph_view_id":"gv_test","state":"computing"}}}`
	if string(body) != want {
		t.Fatalf("query envelope = %s", body)
	}
	query.Details["database_error"] = "private"
	if _, err := query.EnvelopeJSON(); err == nil {
		t.Fatal("unregistered query detail unexpectedly serialized")
	}
}

func TestValidationRegistryRejectsInvalidConstructionAndHashesRequiredDetailsOnly(t *testing.T) {
	run := ProjectionRun{GraphViewID: "gv_test", ProjectionRunID: "gpr_test"}
	base := map[string]any{"projected_key": "name", "expected_type": "string", "actual_type": "integer", "source_field_path": "properties.name", "output_object_id": "vx_test", "contributor_id": "first"}
	first := run.issue("error", "invalid_property_type", "property", "vx_test#name", "properties.name", base)
	secondDetails := cloneErrorDetails(base)
	secondDetails["contributor_id"] = "second"
	second := run.issue("error", "invalid_property_type", "property", "vx_test#name", "properties.name", secondDetails)
	if first.IssueID != second.IssueID {
		t.Fatalf("optional details changed issue identity: %s != %s", first.IssueID, second.IssueID)
	}
	invalid := run.issue("warning", "invalid_filter", "filter", "filter", nil, map[string]any{"field": "$.filters", "reason_code": "invalid_operator"})
	if invalid.Code != "projection_computation_failed" || invalid.Details["reason_code"] != "implementation_invariant_failed" {
		t.Fatalf("invalid issue construction = %#v", invalid)
	}
}

func TestProjectionCancellationNonceAndOutputValidationBoundaries(t *testing.T) {
	if _, err := AdmitRetainedProjection([]byte(`{}`), "", time.Time{}); err == nil {
		t.Fatal("empty retained nonce unexpectedly accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := projectWithContext(ctx, nil, projectOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled projection error = %v", err)
	}
	input := incidentGraphInput(t)
	admitted, err := admitProjectionInput(mustJSON(t, input), admitOptions{Operation: "project_ephemeral", ProjectionRunNonce: "cancel-during-loop", AcceptedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	cancelDuringDerivation := &cancelAfterChecksContext{Context: context.Background(), after: 3}
	canceledRun, err := projectAdmittedRunWithContext(cancelDuringDerivation, admitted, projectOptions{GeneratedAt: fixedTime()})
	if !errors.Is(err, context.Canceled) || canceledRun.GraphView != nil {
		t.Fatalf("mid-derivation cancellation run=%#v err=%v", canceledRun, err)
	}
	run := ProjectionRun{GraphViewID: "gv_test", ProjectionRunID: "gpr_test"}
	graph := &GraphView{GraphViewID: "gv_other", ProjectionRunID: "gpr_test"}
	if err := validateProjectedGraph(run, graph); err == nil || err.Error() != "id_mismatch" {
		t.Fatalf("output identity validation = %v", err)
	}
}

type cancelAfterChecksContext struct {
	context.Context
	checks int
	after  int
}

func (ctx *cancelAfterChecksContext) Err() error {
	ctx.checks++
	if ctx.checks >= ctx.after {
		return context.Canceled
	}
	return nil
}

func TestProjectionRunInspectionDoesNotExposeRetainedInternals(t *testing.T) {
	started := time.Date(2026, 5, 30, 0, 0, 1, 0, time.UTC)
	completed := started.Add(time.Second)
	run := ProjectionRun{
		Request:     ProjectionRequest{SourceSnapshotID: "snapshot", ProjectionConfig: ProjectionConfig{ProjectionVersion: "v1"}},
		GraphViewID: "gv_test", ProjectionRunID: "gpr_test", ProjectionRunNonce: "SECRET_NONCE",
		ProjectionConfigDigest: "SECRET_CONFIG", ProjectionSourceDigest: "SECRET_SOURCE",
		StartedAt: &started, CompletedAt: &completed, State: RunStateAvailable,
		ValidationSummary: ValidationSummary{Status: "valid"}, GraphView: &GraphView{GraphViewKey: "SECRET_GRAPH"},
	}
	inspection := projectionRunInspection(run)
	body, err := json.Marshal(inspection)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"SECRET_NONCE", "SECRET_CONFIG", "SECRET_SOURCE", "SECRET_GRAPH", "projection_run_nonce", "projection_config_digest", "projection_source_digest", "graph_view_key"} {
		if strings.Contains(string(body), secret) {
			t.Fatalf("inspection leaked %q: %s", secret, body)
		}
	}
	wantPrefix := `{"graph_view_id":"gv_test","projection_run_id":"gpr_test","source_snapshot_id":"snapshot","projection_version":"v1","state":"available"`
	if !strings.HasPrefix(string(body), wantPrefix) || !inspection.HasConsumableGraphView || inspection.ValidationSummary == nil {
		t.Fatalf("inspection = %#v json=%s", inspection, body)
	}
	listBody, err := json.Marshal(ListGraphViewsResult{GraphViews: []GraphViewSummary{}})
	if err != nil {
		t.Fatal(err)
	}
	if string(listBody) != `{"graph_views":[],"next_cursor_token":null}` {
		t.Fatalf("empty list resource = %s", listBody)
	}
}

func TestServiceWrapsRepositorySentinelsInContractErrors(t *testing.T) {
	input := minimalInput(t, "facade_sentinel")
	graphViewID := input["graph_view_id"].(string)
	projectionRunID := "gpr_" + strings.Repeat("a", 64)
	vertexID := "vx_" + strings.Repeat("b", 64)
	ctx := context.Background()

	lifecycleService := NewService(ServiceOptions{
		Repository: contractErrorRepository{lifecycleErr: ErrGraphViewNotFound},
		Now:        fixedTime,
		NewNonce:   func() (string, error) { return "facade-sentinel", nil },
	})
	_, err := lifecycleService.RefreshProjection(ctx, RefreshProjectionRequest{GraphViewID: graphViewID, ProjectionInput: mustJSON(t, input)})
	var lifecycleErr *LifecycleError
	if !errors.As(err, &lifecycleErr) || lifecycleErr.Code != "graph_view_not_found" || lifecycleErr.Details["operation"] != "refresh_projection" {
		t.Fatalf("refresh sentinel error = %#v", err)
	}

	queryService := NewService(ServiceOptions{Repository: contractErrorRepository{queryErr: ErrProjectionRunNotFound}})
	_, err = queryService.GetProjectionRun(ctx, GetProjectionRunRequest{GraphViewID: graphViewID, ProjectionRunID: projectionRunID})
	var queryErr *QueryError
	if !errors.As(err, &queryErr) || queryErr.Code != "projection_run_not_found" || queryErr.Details["projection_run_id"] != projectionRunID {
		t.Fatalf("run sentinel error = %#v", err)
	}

	queryService = NewService(ServiceOptions{Repository: contractErrorRepository{queryErr: ErrVertexNotFound}})
	_, err = queryService.GetVertex(ctx, GetVertexRequest{GraphViewID: graphViewID, ProjectionRunID: ValueOf(projectionRunID), VertexID: vertexID})
	if !errors.As(err, &queryErr) || queryErr.Code != "vertex_not_found" || queryErr.Details["vertex_id"] != vertexID {
		t.Fatalf("vertex sentinel error = %#v", err)
	}
}

type contractErrorRepository struct {
	lifecycleErr error
	queryErr     error
}

func (repo contractErrorRepository) CreateProjection(context.Context, []byte, RetainedProjectionOptions) (ProjectionRun, error) {
	return ProjectionRun{}, repo.lifecycleErr
}

func (repo contractErrorRepository) RefreshProjection(context.Context, []byte, RetainedProjectionOptions) (ProjectionRun, error) {
	return ProjectionRun{}, repo.lifecycleErr
}

func (repo contractErrorRepository) GetProjectionRun(context.Context, string) (ProjectionRun, error) {
	return ProjectionRun{}, repo.queryErr
}

func (repo contractErrorRepository) GetGraphView(context.Context, string, string) (GraphView, error) {
	return GraphView{}, repo.queryErr
}

func (repo contractErrorRepository) GetVertex(context.Context, string, string, string) (Vertex, error) {
	return Vertex{}, repo.queryErr
}

func (repo contractErrorRepository) GetEdge(context.Context, string, string, string) (Edge, error) {
	return Edge{}, repo.queryErr
}

func (repo contractErrorRepository) ListGraphViews(context.Context, ListGraphViewsOptions) ([]GraphViewSummary, string, error) {
	return nil, "", repo.queryErr
}

func (repo contractErrorRepository) Traverse(context.Context, TraverseRequest) (TraverseResult, error) {
	return TraverseResult{}, repo.queryErr
}

func (repo contractErrorRepository) InvalidateGraphView(context.Context, RetainedInvalidation) (InvalidationSummary, error) {
	return InvalidationSummary{}, repo.lifecycleErr
}

func (repo contractErrorRepository) InvalidateProjectionRun(context.Context, RetainedInvalidation) (InvalidationSummary, error) {
	return InvalidationSummary{}, repo.lifecycleErr
}

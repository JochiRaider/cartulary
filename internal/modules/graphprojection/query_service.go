package graphprojection

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

func (s *Service) GetGraphView(ctx context.Context, request GetGraphViewRequest) (GraphView, error) {
	projectionRunID, err := validateOptionalGeneratedID(request.ProjectionRunID, "projection_run_id", "gpr_")
	if err != nil {
		return GraphView{}, err
	}
	if err := validateGeneratedIDArgument(request.GraphViewID, "graph_view_id", "gv_"); err != nil {
		return GraphView{}, err
	}
	if s.repository == nil {
		return GraphView{}, errors.New("graphprojection: query service unavailable")
	}
	graphView, err := s.repository.GetGraphView(ctx, request.GraphViewID, projectionRunID)
	if err != nil {
		return GraphView{}, wrapQueryContractError(err, request.GraphViewID, projectionRunID, "", "")
	}
	return graphView, nil
}

func (s *Service) GetProjectionRun(ctx context.Context, request GetProjectionRunRequest) (ProjectionRunInspection, error) {
	if err := validateGeneratedIDArgument(request.GraphViewID, "graph_view_id", "gv_"); err != nil {
		return ProjectionRunInspection{}, err
	}
	if err := validateGeneratedIDArgument(request.ProjectionRunID, "projection_run_id", "gpr_"); err != nil {
		return ProjectionRunInspection{}, err
	}
	if s.repository == nil {
		return ProjectionRunInspection{}, errors.New("graphprojection: query service unavailable")
	}
	run, err := s.repository.GetProjectionRun(ctx, request.ProjectionRunID)
	if err != nil {
		return ProjectionRunInspection{}, wrapQueryContractError(err, request.GraphViewID, request.ProjectionRunID, "", "")
	}
	if run.GraphViewID != request.GraphViewID {
		return ProjectionRunInspection{}, NewQueryError("projection_run_not_found", "", map[string]any{"graph_view_id": request.GraphViewID, "projection_run_id": request.ProjectionRunID}, ErrProjectionRunNotFound)
	}
	return projectionRunInspection(run), nil
}

func projectionRunInspection(run ProjectionRun) ProjectionRunInspection {
	inspection := ProjectionRunInspection{
		GraphViewID:            run.GraphViewID,
		ProjectionRunID:        run.ProjectionRunID,
		SourceSnapshotID:       run.Request.SourceSnapshotID,
		ProjectionVersion:      run.Request.ProjectionConfig.ProjectionVersion,
		State:                  run.State,
		HasConsumableGraphView: run.State == RunStateAvailable || run.State == RunStateReplaced,
		Invalidation:           run.Invalidation,
	}
	if run.StartedAt != nil {
		value := formatLifecycleTimestamp(*run.StartedAt)
		inspection.StartedAt = &value
	}
	if run.CompletedAt != nil {
		value := formatLifecycleTimestamp(*run.CompletedAt)
		inspection.CompletedAt = &value
	}
	if run.State != RunStateAccepted && run.State != RunStateComputing {
		summary := run.ValidationSummary
		inspection.ValidationSummary = &summary
	}
	if run.State == RunStateFailed && run.FailureReason != "" {
		value := run.FailureReason
		inspection.FailureReason = &value
	}
	if run.RetentionExpiresAt != nil {
		value := formatLifecycleTimestamp(*run.RetentionExpiresAt)
		inspection.RetentionExpiresAt = &value
	}
	return inspection
}

func (s *Service) GetVertex(ctx context.Context, request GetVertexRequest) (Vertex, error) {
	projectionRunID, err := validateOptionalGeneratedID(request.ProjectionRunID, "projection_run_id", "gpr_")
	if err != nil {
		return Vertex{}, err
	}
	if err := validateGeneratedIDArgument(request.GraphViewID, "graph_view_id", "gv_"); err != nil {
		return Vertex{}, err
	}
	if err := validateGeneratedIDArgument(request.VertexID, "vertex_id", "vx_"); err != nil {
		return Vertex{}, err
	}
	if s.repository == nil {
		return Vertex{}, errors.New("graphprojection: query service unavailable")
	}
	vertex, err := s.repository.GetVertex(ctx, request.GraphViewID, projectionRunID, request.VertexID)
	if err != nil {
		return Vertex{}, wrapQueryContractError(err, request.GraphViewID, projectionRunID, request.VertexID, "")
	}
	return vertex, nil
}

func (s *Service) GetEdge(ctx context.Context, request GetEdgeRequest) (Edge, error) {
	projectionRunID, err := validateOptionalGeneratedID(request.ProjectionRunID, "projection_run_id", "gpr_")
	if err != nil {
		return Edge{}, err
	}
	if err := validateGeneratedIDArgument(request.GraphViewID, "graph_view_id", "gv_"); err != nil {
		return Edge{}, err
	}
	if err := validateGeneratedIDArgument(request.EdgeID, "edge_id", "ed_"); err != nil {
		return Edge{}, err
	}
	if s.repository == nil {
		return Edge{}, errors.New("graphprojection: query service unavailable")
	}
	edge, err := s.repository.GetEdge(ctx, request.GraphViewID, projectionRunID, request.EdgeID)
	if err != nil {
		return Edge{}, wrapQueryContractError(err, request.GraphViewID, projectionRunID, "", request.EdgeID)
	}
	return edge, nil
}

func validateOptionalGeneratedID(optional Optional[string], field, prefix string) (string, error) {
	if !optional.Present {
		return "", nil
	}
	if optional.Null {
		return "", &QueryError{Code: "invalid_argument", ReasonCode: "explicit_null_not_allowed", Field: field, Details: map[string]any{"field": field, "reason_code": "explicit_null_not_allowed"}}
	}
	if err := validateGeneratedIDArgument(optional.Value, field, prefix); err != nil {
		return "", err
	}
	return optional.Value, nil
}

func validateGeneratedIDArgument(value, field, prefix string) error {
	if value == "" {
		return &QueryError{Code: "invalid_argument", ReasonCode: "missing_required_parameter", Field: field, Details: map[string]any{"field": field, "reason_code": "missing_required_parameter"}}
	}
	if !generatedIDPattern.MatchString(value) || !strings.HasPrefix(value, prefix) {
		return &QueryError{Code: "invalid_argument", ReasonCode: "invalid_value", Field: field, Details: map[string]any{"field": field, "reason_code": "invalid_value"}}
	}
	return nil
}

func (s *Service) ListGraphViews(ctx context.Context, options ListGraphViewsOptions) (ListGraphViewsResult, error) {
	if s.repository == nil {
		return ListGraphViewsResult{}, errors.New("graphprojection: query service unavailable")
	}
	options.QueryShapeDigest = listGraphViewsQueryShapeDigest()
	views, cursor, err := s.repository.ListGraphViews(ctx, options)
	if err != nil {
		return ListGraphViewsResult{}, wrapQueryContractError(err, "", "", "", "")
	}
	var next *string
	if cursor != "" {
		next = &cursor
	}
	return ListGraphViewsResult{GraphViews: views, NextCursorToken: next}, nil
}

func listGraphViewsQueryShapeDigest() string {
	digest := sha256.Sum256([]byte("cartulary.graph_projection.query_shape.v1\nlist_graph_views"))
	return hex.EncodeToString(digest[:])
}

func (s *Service) Traverse(ctx context.Context, request TraverseRequest) (TraverseResult, error) {
	if err := validateGeneratedIDArgument(request.GraphViewID, "graph_view_id", "gv_"); err != nil {
		return TraverseResult{}, err
	}
	if request.ProjectionRunID != "" {
		if err := validateGeneratedIDArgument(request.ProjectionRunID, "projection_run_id", "gpr_"); err != nil {
			return TraverseResult{}, err
		}
	}
	for _, seed := range request.SeedVertexIDs {
		if err := validateGeneratedIDArgument(seed, "seed_vertex_ids", "vx_"); err != nil {
			return TraverseResult{}, err
		}
	}
	for _, kind := range request.VertexKinds {
		if !validIdentifier(kind) {
			return TraverseResult{}, &QueryError{Code: "invalid_argument", ReasonCode: "invalid_value", Field: "vertex_kinds", Details: map[string]any{"field": "vertex_kinds", "reason_code": "invalid_value"}}
		}
	}
	for _, kind := range request.EdgeKinds {
		if !validIdentifier(kind) {
			return TraverseResult{}, &QueryError{Code: "invalid_argument", ReasonCode: "invalid_value", Field: "edge_kinds", Details: map[string]any{"field": "edge_kinds", "reason_code": "invalid_value"}}
		}
	}
	if s.repository == nil {
		return TraverseResult{}, errors.New("graphprojection: query service unavailable")
	}
	result, err := s.repository.Traverse(ctx, request)
	if err != nil {
		return TraverseResult{}, wrapQueryContractError(err, request.GraphViewID, request.ProjectionRunID, "", "")
	}
	return result, nil
}

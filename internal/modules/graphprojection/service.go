package graphprojection

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type EphemeralProjectionRequest struct {
	ProjectionInput []byte
}

type EphemeralProjectionResult struct {
	GraphView             GraphView
	EphemeralProjectionID string
}

func (result EphemeralProjectionResult) Resource() map[string]any {
	data := graphViewResource(result.GraphView)
	delete(data, "projection_run_id")
	data["ephemeral_projection_id"] = result.EphemeralProjectionID
	data["state"] = "ephemeral_available"
	metadata := data["metadata"].(map[string]any)
	metadata["previous_projection_run_id"] = nil
	metadata["invalidation"] = nil
	return data
}

type ServiceOptions struct {
	Now        func() time.Time
	NewNonce   func() (string, error)
	Repository RetainedRepository
}

type Service struct {
	now        func() time.Time
	newNonce   func() (string, error)
	repository RetainedRepository
}

func NewService(options ServiceOptions) *Service {
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newNonce := options.NewNonce
	if newNonce == nil {
		newNonce = secureInvocationNonce
	}
	return &Service{now: now, newNonce: newNonce, repository: options.Repository}
}

type CreateProjectionRequest struct {
	ProjectionInput []byte
	IdempotencyKey  Optional[string]
}

type RefreshProjectionRequest struct {
	GraphViewID     string
	ProjectionInput []byte
	IdempotencyKey  Optional[string]
}

type AcceptedRunSummary struct {
	GraphViewID          string
	ProjectionRunID      string
	State                RunState
	SourceSnapshotID     string
	ProjectionVersion    string
	AcceptedAt           string
	IdempotencyExpiresAt *string
}

type InvalidateGraphViewRequest struct {
	GraphViewID    string
	ReasonCode     string
	RequestedAt    string
	RequestedBy    string
	IdempotencyKey Optional[string]
}

type InvalidateProjectionRunRequest struct {
	GraphViewID     string
	ProjectionRunID string
	ReasonCode      string
	RequestedAt     string
	RequestedBy     string
	IdempotencyKey  Optional[string]
}

func (s *Service) CreateProjection(ctx context.Context, request CreateProjectionRequest) (AcceptedRunSummary, error) {
	if s.repository == nil {
		return AcceptedRunSummary{}, errors.New("graphprojection: retained lifecycle unavailable")
	}
	return s.runRetained(ctx, "create_projection", "", request.ProjectionInput, request.IdempotencyKey)
}

func (s *Service) RefreshProjection(ctx context.Context, request RefreshProjectionRequest) (AcceptedRunSummary, error) {
	if s.repository == nil {
		return AcceptedRunSummary{}, errors.New("graphprojection: retained lifecycle unavailable")
	}
	return s.runRetained(ctx, "refresh_projection", request.GraphViewID, request.ProjectionInput, request.IdempotencyKey)
}

func (s *Service) runRetained(ctx context.Context, operation, graphViewID string, input []byte, idempotency Optional[string]) (AcceptedRunSummary, error) {
	idempotencyKey, err := resolveIdempotencyKey(operation, idempotency)
	if err != nil {
		return AcceptedRunSummary{}, err
	}
	derived, err := deriveGraphViewIDFromInput(input, operation)
	if err != nil {
		return AcceptedRunSummary{}, err
	}
	if operation == "refresh_projection" && (graphViewID == "" || graphViewID != derived) {
		return AcceptedRunSummary{}, &LifecycleError{Code: "invalid_projection_request", ReasonCode: "invalid_graph_view_id", Details: map[string]any{"operation": operation, "reason_code": "invalid_graph_view_id", "field": nil, "validation_code": "invalid_graph_view_id"}}
	}
	nonce, err := s.newNonce()
	if err != nil {
		return AcceptedRunSummary{}, err
	}
	now := s.now().UTC()
	options := RetainedProjectionOptions{ProjectionRunNonce: nonce, AcceptedAt: now, GeneratedAt: now, IdempotencyKey: idempotencyKey}
	var run ProjectionRun
	if operation == "refresh_projection" {
		run, err = s.repository.RefreshProjection(ctx, input, options)
	} else {
		run, err = s.repository.CreateProjection(ctx, input, options)
	}
	if err != nil {
		if errors.Is(err, ErrGraphViewNotFound) {
			return AcceptedRunSummary{}, lifecycleGraphViewNotFound(operation, graphViewID)
		}
		return AcceptedRunSummary{}, err
	}
	if run.AcceptedReplay != nil {
		return *run.AcceptedReplay, nil
	}
	return acceptedRunSummary(run, idempotencyKey != ""), nil
}

func acceptedRunSummary(run ProjectionRun, includeIdempotencyExpiry bool) AcceptedRunSummary {
	var expiresAt *string
	if includeIdempotencyExpiry {
		formatted := formatLifecycleTimestamp(run.AcceptedAt.Add(24 * time.Hour))
		expiresAt = &formatted
	}
	return AcceptedRunSummary{
		GraphViewID:          run.GraphViewID,
		ProjectionRunID:      run.ProjectionRunID,
		State:                RunStateAccepted,
		SourceSnapshotID:     run.Request.SourceSnapshotID,
		ProjectionVersion:    run.Request.ProjectionConfig.ProjectionVersion,
		AcceptedAt:           formatLifecycleTimestamp(run.AcceptedAt),
		IdempotencyExpiresAt: expiresAt,
	}
}

func deriveGraphViewIDFromInput(input []byte, operation string) (string, error) {
	run, err := admitProjectionInput(input, admitOptions{ProjectionRunNonce: "derive-only", AcceptedAt: time.Unix(0, 0).UTC(), Operation: operation})
	if err != nil {
		return "", err
	}
	return run.GraphViewID, nil
}

func (s *Service) GraphViewID(graphViewKey string) (string, error) {
	return deriveGraphViewID(graphViewKey)
}

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

func wrapQueryContractError(err error, graphViewID, projectionRunID, vertexID, edgeID string) error {
	if err == nil {
		return nil
	}
	var queryErr *QueryError
	if errors.As(err, &queryErr) {
		return err
	}
	switch {
	case errors.Is(err, ErrGraphViewNotFound):
		return NewQueryError("graph_view_not_found", "", map[string]any{"graph_view_id": graphViewID}, ErrGraphViewNotFound)
	case errors.Is(err, ErrProjectionRunNotFound):
		return NewQueryError("projection_run_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, ErrProjectionRunNotFound)
	case errors.Is(err, ErrVertexNotFound):
		return NewQueryError("vertex_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "vertex_id": vertexID}, ErrVertexNotFound)
	case errors.Is(err, ErrEdgeNotFound):
		return NewQueryError("edge_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "edge_id": edgeID}, ErrEdgeNotFound)
	case errors.Is(err, ErrGraphViewUnavailable):
		return NewQueryError("projection_not_available", "", map[string]any{"graph_view_id": graphViewID, "state": ""}, ErrGraphViewUnavailable)
	default:
		return err
	}
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

func (s *Service) InvalidateGraphView(ctx context.Context, request InvalidateGraphViewRequest) (InvalidationSummary, error) {
	if s.repository == nil {
		return InvalidationSummary{}, errors.New("graphprojection: retained lifecycle unavailable")
	}
	idempotencyKey, err := resolveIdempotencyKey("invalidate_graph_view", request.IdempotencyKey)
	if err != nil {
		return InvalidationSummary{}, err
	}
	if err := validateInvalidationRequest(request.GraphViewID, "", request.ReasonCode, request.RequestedAt, request.RequestedBy); err != nil {
		return InvalidationSummary{}, err
	}
	summary, err := s.repository.InvalidateGraphView(ctx, RetainedInvalidation{
		GraphViewID: request.GraphViewID, ReasonCode: request.ReasonCode, RequestedAt: request.RequestedAt,
		RequestedBy: request.RequestedBy, IdempotencyKey: idempotencyKey, InvalidatedAt: s.now().UTC(),
	})
	if err != nil {
		return InvalidationSummary{}, wrapLifecycleContractError(err, "invalidate_graph_view", request.GraphViewID, "")
	}
	return summary, nil
}

func (s *Service) InvalidateProjectionRun(ctx context.Context, request InvalidateProjectionRunRequest) (InvalidationSummary, error) {
	if s.repository == nil {
		return InvalidationSummary{}, errors.New("graphprojection: retained lifecycle unavailable")
	}
	idempotencyKey, err := resolveIdempotencyKey("invalidate_projection_run", request.IdempotencyKey)
	if err != nil {
		return InvalidationSummary{}, err
	}
	if err := validateInvalidationRequest(request.GraphViewID, request.ProjectionRunID, request.ReasonCode, request.RequestedAt, request.RequestedBy); err != nil {
		return InvalidationSummary{}, err
	}
	summary, err := s.repository.InvalidateProjectionRun(ctx, RetainedInvalidation{
		GraphViewID: request.GraphViewID, ProjectionRunID: request.ProjectionRunID, ReasonCode: request.ReasonCode,
		RequestedAt: request.RequestedAt, RequestedBy: request.RequestedBy, IdempotencyKey: idempotencyKey,
		InvalidatedAt: s.now().UTC(),
	})
	if err != nil {
		return InvalidationSummary{}, wrapLifecycleContractError(err, "invalidate_projection_run", request.GraphViewID, request.ProjectionRunID)
	}
	return summary, nil
}

func wrapLifecycleContractError(err error, operation, graphViewID, projectionRunID string) error {
	if err == nil {
		return nil
	}
	var lifecycleErr *LifecycleError
	if errors.As(err, &lifecycleErr) {
		return err
	}
	switch {
	case errors.Is(err, ErrGraphViewNotFound):
		return lifecycleGraphViewNotFound(operation, graphViewID)
	case errors.Is(err, ErrProjectionRunNotFound):
		return &LifecycleError{Code: "projection_run_not_found", Details: map[string]any{"operation": operation, "graph_view_id": graphViewID, "projection_run_id": projectionRunID}}
	default:
		return err
	}
}

func lifecycleGraphViewNotFound(operation, graphViewID string) *LifecycleError {
	return &LifecycleError{Code: "graph_view_not_found", Details: map[string]any{"operation": operation, "graph_view_id": graphViewID}}
}

func validateInvalidationRequest(graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy string) error {
	if !generatedIDPattern.MatchString(graphViewID) || !strings.HasPrefix(graphViewID, "gv_") {
		return invalidOperationRequest("invalidate_graph_view", "graph_view_id", "invalid_value")
	}
	if projectionRunID != "" && (!generatedIDPattern.MatchString(projectionRunID) || !strings.HasPrefix(projectionRunID, "gpr_")) {
		return invalidOperationRequest("invalidate_projection_run", "projection_run_id", "invalid_value")
	}
	validReason := map[string]bool{"operator_requested": true, "source_snapshot_withdrawn": true, "projection_config_retired": true, "security_withdrawal": true, "schema_version_retired": true}
	if !validReason[reasonCode] {
		operation := "invalidate_graph_view"
		if projectionRunID != "" {
			operation = "invalidate_projection_run"
		}
		return &LifecycleError{Code: "invalid_operation", ReasonCode: "invalid_reason_code", Details: map[string]any{"operation": operation, "reason_code": "invalid_reason_code"}}
	}
	if _, err := parseTimestamp(requestedAt); err != nil {
		operation := "invalidate_graph_view"
		if projectionRunID != "" {
			operation = "invalidate_projection_run"
		}
		return invalidOperationRequest(operation, "requested_at", "invalid_value")
	}
	if !validIdentifier(requestedBy) {
		operation := "invalidate_graph_view"
		if projectionRunID != "" {
			operation = "invalidate_projection_run"
		}
		return invalidOperationRequest(operation, "requested_by", "invalid_value")
	}
	return nil
}

func invalidOperationRequest(operation, field, reason string) *LifecycleError {
	return &LifecycleError{Code: "invalid_operation_request", ReasonCode: reason, Field: field, Details: map[string]any{"operation": operation, "field": field, "reason_code": reason}}
}

func resolveIdempotencyKey(operation string, optional Optional[string]) (string, error) {
	if !optional.Present {
		return "", nil
	}
	if optional.Null {
		return "", invalidOperationRequest(operation, "idempotency_key", "explicit_null_not_allowed")
	}
	if optional.Value == "" {
		return "", invalidOperationRequest(operation, "idempotency_key", "out_of_bounds")
	}
	if err := validateIdempotencyKey(operation, optional.Value); err != nil {
		return "", err
	}
	return optional.Value, nil
}

func validateIdempotencyKey(operation, value string) error {
	if value == "" {
		return nil
	}
	runes := []rune(value)
	valid := len(runes) <= 128 && !isSpecWhitespace(runes[0]) && !isSpecWhitespace(runes[len(runes)-1])
	for _, r := range runes {
		if r == 0 || r <= 0x1f || (r >= 0x80 && r <= 0x9f) || (r >= 0xd800 && r <= 0xdfff) {
			valid = false
			break
		}
	}
	if valid {
		return nil
	}
	reason := "invalid_value"
	if len(runes) > 128 {
		reason = "out_of_bounds"
	}
	return invalidOperationRequest(operation, "idempotency_key", reason)
}

func (s *Service) ProjectEphemeral(ctx context.Context, request EphemeralProjectionRequest) (EphemeralProjectionResult, error) {
	if err := ctx.Err(); err != nil {
		return EphemeralProjectionResult{}, err
	}
	if len(request.ProjectionInput) == 0 {
		return EphemeralProjectionResult{}, &LifecycleError{Code: "invalid_projection_request", ReasonCode: "missing_required_member", Field: "$.projection_input", Details: map[string]any{"operation": "project_ephemeral", "reason_code": "missing_required_member", "field": "$.projection_input", "validation_code": nil}}
	}
	nonce, err := s.newNonce()
	if err != nil {
		return EphemeralProjectionResult{}, err
	}
	if strings.TrimSpace(nonce) == "" {
		return EphemeralProjectionResult{}, errors.New("graphprojection: ephemeral projection nonce is required")
	}
	now := s.now().UTC()
	run, err := projectWithContext(ctx, request.ProjectionInput, projectOptions{ProjectionRunNonce: nonce, AcceptedAt: now, GeneratedAt: now, InvocationIDPrefix: "gpe_", InvocationDomain: "GPEPHEMERAL1\n", Operation: "project_ephemeral"})
	if err != nil {
		return EphemeralProjectionResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return EphemeralProjectionResult{}, err
	}
	if run.State != RunStateAvailable || run.GraphView == nil {
		return EphemeralProjectionResult{}, &LifecycleError{Code: "ephemeral_projection_failed", ReasonCode: "fatal_validation", Details: map[string]any{"operation": "project_ephemeral", "reason_code": "fatal_validation", "graph_view_id": run.GraphViewID, "ephemeral_projection_id": run.ProjectionRunID, "validation_summary": validationSummaryResource(run.ValidationSummary)}}
	}
	return EphemeralProjectionResult{GraphView: *run.GraphView, EphemeralProjectionID: run.ProjectionRunID}, nil
}

func secureInvocationNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

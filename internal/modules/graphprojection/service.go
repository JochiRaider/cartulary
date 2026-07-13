package graphprojection

import (
	"context"
	"crypto/rand"
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
		return AcceptedRunSummary{}, &OperationError{Code: "invalid_projection_request", ReasonCode: "invalid_graph_view_id", Details: map[string]any{"operation": operation, "reason_code": "invalid_graph_view_id", "field": nil, "validation_code": "invalid_graph_view_id"}}
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
	return s.repository.GetGraphView(ctx, request.GraphViewID, projectionRunID)
}

func (s *Service) GetProjectionRun(ctx context.Context, request GetProjectionRunRequest) (ProjectionRun, error) {
	if err := validateGeneratedIDArgument(request.GraphViewID, "graph_view_id", "gv_"); err != nil {
		return ProjectionRun{}, err
	}
	if err := validateGeneratedIDArgument(request.ProjectionRunID, "projection_run_id", "gpr_"); err != nil {
		return ProjectionRun{}, err
	}
	if s.repository == nil {
		return ProjectionRun{}, errors.New("graphprojection: query service unavailable")
	}
	run, err := s.repository.GetProjectionRun(ctx, request.ProjectionRunID)
	if err != nil {
		return ProjectionRun{}, err
	}
	if run.GraphViewID != request.GraphViewID {
		return ProjectionRun{}, ErrProjectionRunNotFound
	}
	return run, nil
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
	return s.repository.GetVertex(ctx, request.GraphViewID, projectionRunID, request.VertexID)
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
	return s.repository.GetEdge(ctx, request.GraphViewID, projectionRunID, request.EdgeID)
}

func validateOptionalGeneratedID(optional Optional[string], field, prefix string) (string, error) {
	if !optional.Present {
		return "", nil
	}
	if optional.Null {
		return "", &OperationError{Code: "invalid_argument", ReasonCode: "explicit_null_not_allowed", Field: field, Details: map[string]any{"field": field, "reason_code": "explicit_null_not_allowed"}}
	}
	if err := validateGeneratedIDArgument(optional.Value, field, prefix); err != nil {
		return "", err
	}
	return optional.Value, nil
}

func validateGeneratedIDArgument(value, field, prefix string) error {
	if value == "" {
		return &OperationError{Code: "invalid_argument", ReasonCode: "missing_required_parameter", Field: field, Details: map[string]any{"field": field, "reason_code": "missing_required_parameter"}}
	}
	if !generatedIDPattern.MatchString(value) || !strings.HasPrefix(value, prefix) {
		return &OperationError{Code: "invalid_argument", ReasonCode: "invalid_type", Field: field, Details: map[string]any{"field": field, "reason_code": "invalid_type"}}
	}
	return nil
}

func (s *Service) ListGraphViews(ctx context.Context, options ListGraphViewsOptions) ([]GraphViewSummary, string, error) {
	if s.repository == nil {
		return nil, "", errors.New("graphprojection: query service unavailable")
	}
	return s.repository.ListGraphViews(ctx, options)
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
	if s.repository == nil {
		return TraverseResult{}, errors.New("graphprojection: query service unavailable")
	}
	return s.repository.Traverse(ctx, request)
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
	return s.repository.InvalidateGraphView(ctx, RetainedInvalidation{
		GraphViewID: request.GraphViewID, ReasonCode: request.ReasonCode, RequestedAt: request.RequestedAt,
		RequestedBy: request.RequestedBy, IdempotencyKey: idempotencyKey, InvalidatedAt: s.now().UTC(),
	})
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
	return s.repository.InvalidateProjectionRun(ctx, RetainedInvalidation{
		GraphViewID: request.GraphViewID, ProjectionRunID: request.ProjectionRunID, ReasonCode: request.ReasonCode,
		RequestedAt: request.RequestedAt, RequestedBy: request.RequestedBy, IdempotencyKey: idempotencyKey,
		InvalidatedAt: s.now().UTC(),
	})
}

func validateInvalidationRequest(graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy string) error {
	if !generatedIDPattern.MatchString(graphViewID) || !strings.HasPrefix(graphViewID, "gv_") {
		return &OperationError{Code: "invalid_argument", ReasonCode: "invalid_graph_view_id", Field: "graph_view_id", Details: map[string]any{"field": "graph_view_id"}}
	}
	if projectionRunID != "" && (!generatedIDPattern.MatchString(projectionRunID) || !strings.HasPrefix(projectionRunID, "gpr_")) {
		return &OperationError{Code: "invalid_argument", ReasonCode: "invalid_projection_run_id", Field: "projection_run_id", Details: map[string]any{"field": "projection_run_id"}}
	}
	validReason := map[string]bool{"operator_requested": true, "source_snapshot_withdrawn": true, "projection_config_retired": true, "security_withdrawal": true, "schema_version_retired": true}
	if !validReason[reasonCode] {
		return &OperationError{Code: "invalid_argument", ReasonCode: "invalid_reason_code", Field: "reason_code", Details: map[string]any{"field": "reason_code"}}
	}
	if _, err := parseTimestamp(requestedAt); err != nil {
		return &OperationError{Code: "invalid_argument", ReasonCode: "invalid_timestamp", Field: "requested_at", Details: map[string]any{"field": "requested_at"}}
	}
	if !validIdentifier(requestedBy) {
		return &OperationError{Code: "invalid_argument", ReasonCode: "invalid_requested_by", Field: "requested_by", Details: map[string]any{"field": "requested_by"}}
	}
	return nil
}

func resolveIdempotencyKey(operation string, optional Optional[string]) (string, error) {
	if !optional.Present {
		return "", nil
	}
	if optional.Null {
		return "", &OperationError{Code: "invalid_operation", ReasonCode: "explicit_null_not_allowed", Details: map[string]any{"operation": operation, "reason_code": "explicit_null_not_allowed"}}
	}
	if optional.Value == "" {
		return "", &OperationError{Code: "invalid_operation", ReasonCode: "invalid_idempotency_key", Details: map[string]any{"operation": operation, "reason_code": "invalid_idempotency_key"}}
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
	return &OperationError{Code: "invalid_operation", ReasonCode: "invalid_idempotency_key", Details: map[string]any{"operation": operation, "reason_code": "invalid_idempotency_key"}}
}

func (s *Service) ProjectEphemeral(ctx context.Context, request EphemeralProjectionRequest) (EphemeralProjectionResult, error) {
	if err := ctx.Err(); err != nil {
		return EphemeralProjectionResult{}, err
	}
	if len(request.ProjectionInput) == 0 {
		return EphemeralProjectionResult{}, &OperationError{Code: "invalid_projection_request", ReasonCode: "missing_required_member", Field: "$.projection_input", Details: map[string]any{"operation": "project_ephemeral", "reason_code": "missing_required_member", "field": "$.projection_input", "validation_code": nil}}
	}
	nonce, err := s.newNonce()
	if err != nil {
		return EphemeralProjectionResult{}, err
	}
	now := s.now().UTC()
	run, err := project(request.ProjectionInput, projectOptions{ProjectionRunNonce: nonce, AcceptedAt: now, GeneratedAt: now, InvocationIDPrefix: "gpe_", InvocationDomain: "GPEPHEMERAL1\n", Operation: "project_ephemeral"})
	if err != nil {
		return EphemeralProjectionResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return EphemeralProjectionResult{}, err
	}
	if run.State != RunStateAvailable || run.GraphView == nil {
		return EphemeralProjectionResult{}, &OperationError{Code: "ephemeral_projection_failed", ReasonCode: "fatal_validation", Details: map[string]any{"operation": "project_ephemeral", "reason_code": "fatal_validation", "graph_view_id": run.GraphViewID, "ephemeral_projection_id": run.ProjectionRunID, "validation_summary": validationSummaryResource(run.ValidationSummary)}}
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

func IsOperationError(err error, code, reason string) bool {
	var operationErr *OperationError
	return errors.As(err, &operationErr) && operationErr.Code == code && operationErr.ReasonCode == reason
}

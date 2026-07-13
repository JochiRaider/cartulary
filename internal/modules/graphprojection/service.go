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
	Data map[string]any
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
	IdempotencyKey  string
}

type RefreshProjectionRequest struct {
	GraphViewID     string
	ProjectionInput []byte
	IdempotencyKey  string
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
	IdempotencyKey string
}

type InvalidateProjectionRunRequest struct {
	GraphViewID     string
	ProjectionRunID string
	ReasonCode      string
	RequestedAt     string
	RequestedBy     string
	IdempotencyKey  string
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

func (s *Service) runRetained(ctx context.Context, operation, graphViewID string, input []byte, idempotencyKey string) (AcceptedRunSummary, error) {
	nonce, err := s.newNonce()
	if err != nil {
		return AcceptedRunSummary{}, err
	}
	now := s.now().UTC()
	options := RetainedProjectionOptions{ProjectionRunNonce: nonce, AcceptedAt: now, GeneratedAt: now, IdempotencyKey: idempotencyKey}
	var run ProjectionRun
	if operation == "refresh_projection" {
		derived, deriveErr := deriveGraphViewIDFromInput(input)
		if deriveErr != nil {
			return AcceptedRunSummary{}, deriveErr
		}
		if graphViewID == "" || graphViewID != derived {
			return AcceptedRunSummary{}, &OperationError{Code: "invalid_projection_request", ReasonCode: "invalid_graph_view_id", Details: map[string]any{"operation": operation, "reason_code": "invalid_graph_view_id", "field": nil, "validation_code": "invalid_graph_view_id"}}
		}
		run, err = s.repository.RefreshProjection(ctx, input, options)
	} else {
		run, err = s.repository.CreateProjection(ctx, input, options)
	}
	if err != nil {
		return AcceptedRunSummary{}, err
	}
	var expiresAt *string
	if idempotencyKey != "" {
		formatted := formatLifecycleTimestamp(run.AcceptedAt.Add(24 * time.Hour))
		expiresAt = &formatted
	}
	return AcceptedRunSummary{GraphViewID: run.GraphViewID, ProjectionRunID: run.ProjectionRunID, State: RunStateAccepted, SourceSnapshotID: run.Request.SourceSnapshotID, ProjectionVersion: run.Request.ProjectionConfig.ProjectionVersion, AcceptedAt: formatLifecycleTimestamp(run.AcceptedAt), IdempotencyExpiresAt: expiresAt}, nil
}

func deriveGraphViewIDFromInput(input []byte) (string, error) {
	run, err := admitProjectionInput(input, admitOptions{ProjectionRunNonce: "derive-only", AcceptedAt: time.Unix(0, 0).UTC()})
	if err != nil {
		return "", err
	}
	return run.GraphViewID, nil
}

func (s *Service) GraphViewID(graphViewKey string) (string, error) {
	return deriveGraphViewID(graphViewKey)
}

func (s *Service) GetGraphView(ctx context.Context, graphViewID, projectionRunID string) (GraphView, error) {
	if s.repository == nil {
		return GraphView{}, errors.New("graphprojection: query service unavailable")
	}
	return s.repository.GetGraphView(ctx, graphViewID, projectionRunID)
}

func (s *Service) GetProjectionRun(ctx context.Context, graphViewID, projectionRunID string) (ProjectionRun, error) {
	if s.repository == nil {
		return ProjectionRun{}, errors.New("graphprojection: query service unavailable")
	}
	run, err := s.repository.GetProjectionRun(ctx, projectionRunID)
	if err != nil {
		return ProjectionRun{}, err
	}
	if run.GraphViewID != graphViewID {
		return ProjectionRun{}, ErrProjectionRunNotFound
	}
	return run, nil
}

func (s *Service) GetVertex(ctx context.Context, graphViewID, projectionRunID, vertexID string) (Vertex, error) {
	if s.repository == nil {
		return Vertex{}, errors.New("graphprojection: query service unavailable")
	}
	return s.repository.GetVertex(ctx, graphViewID, projectionRunID, vertexID)
}

func (s *Service) GetEdge(ctx context.Context, graphViewID, projectionRunID, edgeID string) (Edge, error) {
	if s.repository == nil {
		return Edge{}, errors.New("graphprojection: query service unavailable")
	}
	return s.repository.GetEdge(ctx, graphViewID, projectionRunID, edgeID)
}

func (s *Service) ListGraphViews(ctx context.Context, options ListGraphViewsOptions) ([]GraphViewSummary, string, error) {
	if s.repository == nil {
		return nil, "", errors.New("graphprojection: query service unavailable")
	}
	return s.repository.ListGraphViews(ctx, options)
}

func (s *Service) Traverse(ctx context.Context, request TraverseRequest) (TraverseResult, error) {
	if s.repository == nil {
		return TraverseResult{}, errors.New("graphprojection: query service unavailable")
	}
	return s.repository.Traverse(ctx, request)
}

func (s *Service) InvalidateGraphView(ctx context.Context, request InvalidateGraphViewRequest) (InvalidationSummary, error) {
	if s.repository == nil {
		return InvalidationSummary{}, errors.New("graphprojection: retained lifecycle unavailable")
	}
	if err := validateInvalidationRequest(request.GraphViewID, "", request.ReasonCode, request.RequestedAt, request.RequestedBy, request.IdempotencyKey); err != nil {
		return InvalidationSummary{}, err
	}
	return s.repository.InvalidateGraphView(ctx, RetainedInvalidation{
		GraphViewID: request.GraphViewID, ReasonCode: request.ReasonCode, RequestedAt: request.RequestedAt,
		RequestedBy: request.RequestedBy, IdempotencyKey: request.IdempotencyKey, InvalidatedAt: s.now().UTC(),
	})
}

func (s *Service) InvalidateProjectionRun(ctx context.Context, request InvalidateProjectionRunRequest) (InvalidationSummary, error) {
	if s.repository == nil {
		return InvalidationSummary{}, errors.New("graphprojection: retained lifecycle unavailable")
	}
	if err := validateInvalidationRequest(request.GraphViewID, request.ProjectionRunID, request.ReasonCode, request.RequestedAt, request.RequestedBy, request.IdempotencyKey); err != nil {
		return InvalidationSummary{}, err
	}
	return s.repository.InvalidateProjectionRun(ctx, RetainedInvalidation{
		GraphViewID: request.GraphViewID, ProjectionRunID: request.ProjectionRunID, ReasonCode: request.ReasonCode,
		RequestedAt: request.RequestedAt, RequestedBy: request.RequestedBy, IdempotencyKey: request.IdempotencyKey,
		InvalidatedAt: s.now().UTC(),
	})
}

func validateInvalidationRequest(graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy, idempotencyKey string) error {
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
	if len(idempotencyKey) > 256 {
		return &OperationError{Code: "invalid_argument", ReasonCode: "invalid_idempotency_key", Field: "idempotency_key", Details: map[string]any{"field": "idempotency_key"}}
	}
	return nil
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
	run, err := project(request.ProjectionInput, projectOptions{ProjectionRunNonce: nonce, AcceptedAt: now, GeneratedAt: now, InvocationIDPrefix: "gpe_", InvocationDomain: "GPEPHEMERAL1\n"})
	if err != nil {
		return EphemeralProjectionResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return EphemeralProjectionResult{}, err
	}
	if run.State != RunStateAvailable || run.GraphView == nil {
		return EphemeralProjectionResult{}, &OperationError{Code: "ephemeral_projection_failed", ReasonCode: "fatal_validation", Details: map[string]any{"operation": "project_ephemeral", "reason_code": "fatal_validation", "graph_view_id": run.GraphViewID, "ephemeral_projection_id": run.ProjectionRunID, "validation_summary": validationSummaryResource(run.ValidationSummary)}}
	}
	data := graphViewResource(*run.GraphView)
	delete(data, "projection_run_id")
	data["ephemeral_projection_id"] = run.ProjectionRunID
	data["state"] = "ephemeral_available"
	metadata := data["metadata"].(map[string]any)
	metadata["previous_projection_run_id"] = nil
	metadata["invalidation"] = nil
	return EphemeralProjectionResult{Data: data}, nil
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

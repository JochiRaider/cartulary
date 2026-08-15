package graphprojection

import (
	"context"
	"errors"
	"strings"
)

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

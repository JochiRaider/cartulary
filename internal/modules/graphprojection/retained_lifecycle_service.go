package graphprojection

import (
	"context"
	"errors"
	"time"
)

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

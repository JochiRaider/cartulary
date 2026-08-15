package graphprojection

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

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

package graphprojection

import "time"

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

package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ProgressIntent is the Jobs-owned producer contract. It contains only the
// public common-job resource needed by Collaboration; worker-private state is
// deliberately absent.
type ProgressIntent struct {
	IntentKey        string
	IncidentID       uuid.UUID
	CanonicalPayload json.RawMessage
	SourceIdentity   string
	CreatedAt        time.Time
}

// ProgressIntentAppender is the narrow consumer-owned port implemented by
// application composition. Jobs never receives Collaboration storage access.
type ProgressIntentAppender interface {
	AppendProgressIntentTx(context.Context, pgx.Tx, ProgressIntent) error
}

// TransactionService binds Jobs' transaction-scoped operations to its
// configured durable intent producer.
type TransactionService struct {
	progressIntents ProgressIntentAppender
}

func NewTransactionService(progressIntents ProgressIntentAppender) (*TransactionService, error) {
	if progressIntents == nil {
		return nil, errors.New("jobs transaction service requires a progress intent appender")
	}
	return &TransactionService{progressIntents: progressIntents}, nil
}

func (s *TransactionService) CreateQueuedTx(ctx context.Context, tx pgx.Tx, params CreateParams, now time.Time) (Resource, error) {
	if s == nil || s.progressIntents == nil {
		return Resource{}, ErrNotConfigured
	}
	resource, err := createQueuedTx(ctx, tx, params, now)
	if err != nil {
		return Resource{}, err
	}
	if err := s.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func (s *TransactionService) CompleteSucceededTx(ctx context.Context, tx pgx.Tx, params TransitionParams, now time.Time) (Resource, error) {
	if s == nil || s.progressIntents == nil {
		return Resource{}, ErrNotConfigured
	}
	resource, err := completeSucceededTx(ctx, tx, params, now)
	if err != nil {
		return Resource{}, err
	}
	if err := s.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func (s *TransactionService) CompleteFailedTx(ctx context.Context, tx pgx.Tx, params TransitionParams, now time.Time) (Resource, error) {
	return s.completeTerminalTx(ctx, tx, params, now, StatusFailed)
}

func (s *TransactionService) CompleteCanceledTx(ctx context.Context, tx pgx.Tx, params TransitionParams, now time.Time) (Resource, error) {
	return s.completeTerminalTx(ctx, tx, params, now, StatusCanceled)
}

func (s *TransactionService) completeTerminalTx(
	ctx context.Context,
	tx pgx.Tx,
	params TransitionParams,
	now time.Time,
	status string,
) (Resource, error) {
	if s == nil || s.progressIntents == nil {
		return Resource{}, ErrNotConfigured
	}
	resource, err := completeTerminalTx(ctx, tx, params, now, status)
	if err != nil {
		return Resource{}, err
	}
	if err := s.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func (s *TransactionService) appendProgressIntentTx(ctx context.Context, tx pgx.Tx, resource Resource) error {
	if s == nil || s.progressIntents == nil {
		return ErrNotConfigured
	}
	if resource.Scope.Kind != ScopeKindIncident || resource.Scope.IncidentID == nil {
		return nil
	}
	payload := map[string]any{
		"job_id": resource.JobID,
		"scope": map[string]any{
			"kind":        ScopeKindIncident,
			"incident_id": resource.Scope.IncidentID.String(),
		},
		"status": resource.Status,
		"progress": map[string]any{
			"completed": resource.Progress.Completed,
			"total":     resource.Progress.Total,
		},
		"updated_at": resource.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"cancelable": resource.Cancelable,
	}
	if resource.Message != nil {
		payload["message"] = *resource.Message
	}
	if resource.ResultSummary != nil {
		payload["result_summary"] = resource.ResultSummary
	}
	if resource.ErrorSummary != nil {
		payload["error_summary"] = resource.ErrorSummary
	}
	if resource.RetainedUntil != nil {
		payload["retained_until"] = resource.RetainedUntil.UTC().Format(time.RFC3339Nano)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal job progress intent: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return s.progressIntents.AppendProgressIntentTx(ctx, tx, ProgressIntent{
		IntentKey:        fmt.Sprintf("job_progress:v2:%s:%x", resource.JobID, digest[:]),
		IncidentID:       *resource.Scope.IncidentID,
		CanonicalPayload: encoded,
		SourceIdentity:   "job:" + resource.JobID,
		CreatedAt:        resource.UpdatedAt.UTC(),
	})
}

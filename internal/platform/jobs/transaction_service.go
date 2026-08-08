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
	progressIntents       ProgressIntentAppender
	routeIdempotency      RouteIdempotencyPort
	extensionCancellation ExtensionCancellationPort
	definitions           *jobDefinitionCatalog
}

type ExtensionFinalizationContext struct {
	Definition              ExtensionJobContract
	IdempotencyIdentity     json.RawMessage
	IdempotencyRouteKey     string
	IdempotencyScopeKey     string
	NormalizedRequestSHA256 string
	ActorUserID             uuid.UUID
	ClientTxnID             string
}

func NewTransactionService(progressIntents ProgressIntentAppender, ownerPorts OwnerTransactionPorts, definitions ...ExtensionJobContract) (*TransactionService, error) {
	if progressIntents == nil || ownerPorts.RouteIdempotency == nil || ownerPorts.ExtensionCancellation == nil {
		return nil, errors.New("jobs transaction service requires progress, route idempotency, and extension cancellation ports")
	}
	service := &TransactionService{
		progressIntents:       progressIntents,
		routeIdempotency:      ownerPorts.RouteIdempotency,
		extensionCancellation: ownerPorts.ExtensionCancellation,
	}
	if len(definitions) > 0 {
		if _, err := service.configureDefinitions(definitions); err != nil {
			return nil, err
		}
	}
	return service, nil
}

func (s *TransactionService) CreateQueuedTx(ctx context.Context, tx pgx.Tx, params CreateParams, now time.Time) (Resource, error) {
	if s == nil || s.progressIntents == nil {
		return Resource{}, ErrNotConfigured
	}
	jobKind := params.JobKind
	if jobKind == "" && params.Extension != nil {
		jobKind = params.Extension.JobKind
	}
	if s.definitions == nil {
		return Resource{}, fmt.Errorf("%w: definition catalog unavailable", ErrInvalidJobDefinition)
	}
	definition, present := s.definitions.byKind[jobKind]
	if !present {
		return Resource{}, fmt.Errorf("%w: unknown job kind", ErrInvalidJobDefinition)
	}
	params.JobKind = jobKind
	resource, err := createDefinedQueuedTx(ctx, tx, params, definition, now)
	if err != nil {
		return Resource{}, err
	}
	if err := s.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func (s *TransactionService) configureDefinitions(definitions []ExtensionJobContract) (*jobDefinitionCatalog, error) {
	if s == nil {
		return nil, ErrNotConfigured
	}
	if s.definitions != nil {
		if s.definitions.contains(definitions) {
			return s.definitions, nil
		}
		return nil, fmt.Errorf("%w: job definition catalog already configured", ErrInvalidJobDefinition)
	}
	catalog, err := newJobDefinitionCatalog(definitions)
	if err != nil {
		return nil, err
	}
	s.definitions = catalog
	return catalog, nil
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

// ValidateRunnableTx serializes an owner transaction with cancellation and
// terminal publication, then exposes only the semantic runnable result.
func (s *TransactionService) ValidateRunnableTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) error {
	if s == nil || tx == nil || jobID == uuid.Nil {
		return ErrNotConfigured
	}
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return err
	}
	resource, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		return err
	}
	switch resource.Status {
	case StatusRunning:
		return nil
	case StatusCancelRequested:
		return ErrCancellationRequested
	default:
		return ErrInvalidTransition
	}
}

// ExtensionFinalizationContextTx validates and locks a runnable extension job,
// then returns only the definition and replay facts required by its owner
// finalizer. Callers never query Jobs storage directly.
func (s *TransactionService) ExtensionFinalizationContextTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (ExtensionFinalizationContext, error) {
	if s == nil || s.definitions == nil || tx == nil || jobID == uuid.Nil {
		return ExtensionFinalizationContext{}, ErrNotConfigured
	}
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return ExtensionFinalizationContext{}, err
	}
	var metadata ExtensionFinalizationContext
	var ownerProfileID string
	var jobKind string
	var workerKind string
	err := tx.QueryRow(ctx, `
SELECT extension_owner_profile_id, job_kind, handler_name,
       extension_idempotency_identity, extension_idempotency_route_key,
       extension_idempotency_scope_key, extension_normalized_request_sha256,
       submitted_by_user_id
  FROM jobs
 WHERE job_id = $1
   AND status IN ('running', 'cancel_requested')
 FOR UPDATE
`, jobID).Scan(
		&ownerProfileID,
		&jobKind,
		&workerKind,
		&metadata.IdempotencyIdentity,
		&metadata.IdempotencyRouteKey,
		&metadata.IdempotencyScopeKey,
		&metadata.NormalizedRequestSHA256,
		&metadata.ActorUserID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ExtensionFinalizationContext{}, ErrInvalidTransition
	}
	if err != nil {
		return ExtensionFinalizationContext{}, err
	}
	definition, present := s.definitions.byKind[jobKind]
	if !present || !definition.ProofRequired || definition.OwnerProfileID != ownerProfileID || definition.WorkerKind != workerKind {
		return ExtensionFinalizationContext{}, ErrInvalidJobDefinition
	}
	var identity routeScopedIdempotencyIdentity
	if err := json.Unmarshal(metadata.IdempotencyIdentity, &identity); err != nil ||
		identity.SchemaID != "cartulary.route_scoped_idempotency_identity.v1" ||
		identity.ActorUserID != metadata.ActorUserID.String() ||
		identity.RouteIdentity != metadata.IdempotencyRouteKey+":"+metadata.IdempotencyScopeKey ||
		identity.ClientTxnID == "" {
		return ExtensionFinalizationContext{}, ErrInvalidJobDefinition
	}
	definition.ResourceRefs = append([]ExtensionResourceRefContract(nil), definition.ResourceRefs...)
	metadata.Definition = definition
	metadata.ClientTxnID = identity.ClientTxnID
	return metadata, nil
}

// ValidateInactiveJobTx acquires Jobs' transition lock and proves that the
// reconciliation candidate is still the same mutable, extension-owned job
// observed by the Extensions owner. The lock remains held by the caller's
// transaction while owner evidence is classified and the terminal state is
// committed.
func (s *TransactionService) ValidateInactiveJobTx(
	ctx context.Context,
	tx pgx.Tx,
	jobID uuid.UUID,
	ownerProfileID string,
	submittedAt time.Time,
) error {
	if s == nil || s.definitions == nil || tx == nil || jobID == uuid.Nil || ownerProfileID == "" || submittedAt.IsZero() {
		return ErrNotConfigured
	}
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return err
	}
	var storedOwnerProfileID string
	var jobKind string
	var storedSubmittedAt time.Time
	err := tx.QueryRow(ctx, `
SELECT extension_owner_profile_id, job_kind, submitted_at
  FROM jobs
 WHERE job_id = $1
   AND status IN ('queued', 'running', 'cancel_requested')
`, jobID).Scan(&storedOwnerProfileID, &jobKind, &storedSubmittedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidTransition
	}
	if err != nil {
		return err
	}
	definition, present := s.definitions.byKind[jobKind]
	if !present || definition.OwnerProfileID != ownerProfileID ||
		storedOwnerProfileID != ownerProfileID || !storedSubmittedAt.Equal(submittedAt) {
		return ErrInvalidJobDefinition
	}
	return nil
}

// CompleteInactiveJobTx is the narrow reconciliation operation used after an
// extension profile becomes inactive. ValidateInactiveJobTx must already have
// locked and validated the candidate in this transaction. This operation
// preserves the public state sequence and publishes every state change inside
// the caller transaction.
func (s *TransactionService) CompleteInactiveJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, status string, terminalJSON json.RawMessage, now time.Time) error {
	if s == nil || s.progressIntents == nil || tx == nil || jobID == uuid.Nil || now.IsZero() {
		return ErrNotConfigured
	}
	current, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		return err
	}
	if current.Status == StatusQueued {
		var mutation transitionMutation
		if status == StatusCanceled {
			mutation, _, err = transitionCancellationTx(ctx, tx, jobID, now)
		} else {
			mutation, err = transitionRunningTx(ctx, tx, jobID, current.Progress, current.Message, now)
		}
		if err != nil {
			return err
		}
		if mutation.changed {
			if err := s.appendProgressIntentTx(ctx, tx, mutation.resource); err != nil {
				return err
			}
		}
		current = mutation.resource
	}
	params := TransitionParams{JobID: jobID, Progress: current.Progress}
	switch status {
	case StatusSucceeded, StatusCanceled:
		var summary ResultSummary
		if err := json.Unmarshal(terminalJSON, &summary); err != nil {
			return ErrInvalidJobDefinition
		}
		params.ResultSummary = &summary
	case StatusFailed:
		var summary ErrorSummary
		if err := json.Unmarshal(terminalJSON, &summary); err != nil {
			return ErrInvalidJobDefinition
		}
		params.ErrorSummary = &summary
	default:
		return ErrInvalidTransition
	}
	mutation, err := transitionTerminalTx(ctx, tx, params, now, status)
	if err != nil {
		return err
	}
	return s.appendProgressIntentTx(ctx, tx, mutation.resource)
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

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
	catalog               *Catalog
	selection             *RuntimeSelection
}

type ExtensionFinalizationContext struct {
	Definition              Definition
	IdempotencyIdentity     json.RawMessage
	IdempotencyRouteKey     string
	IdempotencyScopeKey     string
	NormalizedRequestSHA256 string
	ActorUserID             uuid.UUID
	ClientTxnID             string
}

func NewTransactionService(progressIntents ProgressIntentAppender, ownerPorts OwnerTransactionPorts, catalog *Catalog, selection *RuntimeSelection) (*TransactionService, error) {
	if progressIntents == nil || ownerPorts.RouteIdempotency == nil || ownerPorts.ExtensionCancellation == nil || catalog == nil ||
		selection == nil || selection.catalog != catalog {
		return nil, errors.New("jobs transaction service requires progress, route idempotency, and extension cancellation ports")
	}
	return &TransactionService{
		progressIntents:       progressIntents,
		routeIdempotency:      ownerPorts.RouteIdempotency,
		extensionCancellation: ownerPorts.ExtensionCancellation,
		catalog:               catalog,
		selection:             selection,
	}, nil
}

func (s *TransactionService) CreateQueuedTx(ctx context.Context, tx pgx.Tx, params EnqueueParams, now time.Time) (Resource, error) {
	if s == nil || s.progressIntents == nil {
		return Resource{}, ErrNotConfigured
	}
	jobKind := params.JobKind
	if !s.selection.containsJobKind(jobKind) {
		return Resource{}, fmt.Errorf("%w: job kind is not admitted", ErrInvalidJobDefinition)
	}
	definition, present := s.catalog.definition(jobKind)
	if !present {
		return Resource{}, fmt.Errorf("%w: unknown job kind", ErrInvalidJobDefinition)
	}
	resource, err := createDefinedQueuedTx(ctx, tx, params, definition, now)
	if err != nil {
		return Resource{}, err
	}
	if err := s.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func (s *TransactionService) CompleteSucceededTx(ctx context.Context, tx pgx.Tx, execution Execution, completion SuccessCompletion, now time.Time) (Resource, error) {
	if s == nil || s.progressIntents == nil {
		return Resource{}, ErrNotConfigured
	}
	if err := s.validateExecutionTx(ctx, tx, execution, false); err != nil {
		return Resource{}, err
	}
	summary := completion.ResultSummary
	resource, err := completeSucceededTx(ctx, tx, terminalTransition{
		JobID: execution.jobID, Progress: completion.Progress,
		ResultSummary: &summary, Message: completion.Message,
	}, now)
	if err != nil {
		return Resource{}, err
	}
	if err := s.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func (s *TransactionService) CompleteFailedTx(ctx context.Context, tx pgx.Tx, execution Execution, completion FailureCompletion, now time.Time) (Resource, error) {
	summary := completion.ErrorSummary
	return s.completeExecutionTerminalTx(ctx, tx, execution, terminalTransition{
		JobID: execution.jobID, Progress: completion.Progress,
		ErrorSummary: &summary, Message: completion.Message,
	}, now, StatusFailed, false)
}

func (s *TransactionService) CompleteCanceledTx(ctx context.Context, tx pgx.Tx, execution Execution, completion CancellationCompletion, now time.Time) (Resource, error) {
	summary := completion.ResultSummary
	if summary.Code == "" && summary.Message == "" {
		summary = ResultSummary{Code: "job_canceled", Message: "Job canceled."}
	}
	return s.completeExecutionTerminalTx(ctx, tx, execution, terminalTransition{
		JobID: execution.jobID, Progress: completion.Progress,
		ResultSummary: &summary, Message: completion.Message,
	}, now, StatusCanceled, true)
}

// ValidateExecutionTx serializes owner work with cancellation and attempt
// replacement. The opaque execution must still own an unexpired lease.
func (s *TransactionService) ValidateExecutionTx(ctx context.Context, tx pgx.Tx, execution Execution) error {
	return s.validateExecutionTx(ctx, tx, execution, false)
}

// ValidateCancellationExecutionTx proves that an execution still owns its
// lease while permitting the cancel-requested state required by a cancellation
// finalizer. It must not be used to authorize owner mutations.
func (s *TransactionService) ValidateCancellationExecutionTx(ctx context.Context, tx pgx.Tx, execution Execution) error {
	return s.validateExecutionTx(ctx, tx, execution, true)
}

func (s *TransactionService) validateExecutionTx(
	ctx context.Context,
	tx pgx.Tx,
	execution Execution,
	allowCancellationRequested bool,
) error {
	if s == nil || tx == nil || !execution.valid() {
		return ErrNotConfigured
	}
	if err := lockTransitionTx(ctx, tx, execution.jobID); err != nil {
		return err
	}
	var status string
	err := tx.QueryRow(ctx, `
SELECT status
  FROM jobs
 WHERE job_id = $1
   AND handler_attempt_id = $2
   AND handler_lease_expires_at > now()
   AND status IN ('running', 'cancel_requested')
 FOR UPDATE
`, execution.jobID, execution.attemptID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrExecutionLost
	}
	if err != nil {
		return err
	}
	if status == StatusCancelRequested && !allowCancellationRequested {
		return ErrCancellationRequested
	}
	return nil
}

// ExtensionFinalizationContextTx validates and locks a runnable extension job,
// then returns only the definition and replay facts required by its owner
// finalizer. Callers never query Jobs storage directly.
func (s *TransactionService) ExtensionFinalizationContextTx(ctx context.Context, tx pgx.Tx, execution Execution) (ExtensionFinalizationContext, error) {
	if s == nil || s.catalog == nil || tx == nil || !execution.valid() {
		return ExtensionFinalizationContext{}, ErrNotConfigured
	}
	if err := s.validateExecutionTx(ctx, tx, execution, false); err != nil {
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
	AND handler_attempt_id = $2
	AND handler_lease_expires_at > now()
	AND status = 'running'
 FOR UPDATE
`, execution.jobID, execution.attemptID).Scan(
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
	definition, present := s.catalog.definition(jobKind)
	if !present || definition.Extension == nil || !definition.Extension.ProofRequired ||
		definition.Extension.OwnerProfileID != ownerProfileID || definition.HandlerName != workerKind {
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
	metadata.Definition = definition
	metadata.ClientTxnID = identity.ClientTxnID
	return metadata, nil
}

// ValidateInactiveJobTx acquires Jobs' transition lock and proves that the
// reconciliation candidate is still the same mutable, extension-owned job
// observed by the Extensions owner. The lock remains held by the caller's
// transaction while owner evidence is classified and the terminal state is
// committed.
type InactiveJobGrant struct {
	jobID             uuid.UUID
	ownerProfileID    string
	storedSubmittedAt time.Time
}

type InactiveTerminalOutcome struct {
	status  string
	result  *ResultSummary
	failure *ErrorSummary
}

func NewInactiveSuccessOutcome(summary ResultSummary) InactiveTerminalOutcome {
	return InactiveTerminalOutcome{status: StatusSucceeded, result: &summary}
}

func NewInactiveFailureOutcome(summary ErrorSummary) InactiveTerminalOutcome {
	return InactiveTerminalOutcome{status: StatusFailed, failure: &summary}
}

func NewInactiveCancellationOutcome() InactiveTerminalOutcome {
	summary := ResultSummary{Code: "job_canceled", Message: "Job canceled."}
	return InactiveTerminalOutcome{status: StatusCanceled, result: &summary}
}

func (s *TransactionService) ValidateInactiveJobTx(
	ctx context.Context,
	tx pgx.Tx,
	jobID uuid.UUID,
	ownerProfileID string,
	submittedAt time.Time,
) (InactiveJobGrant, error) {
	if s == nil || s.catalog == nil || tx == nil || jobID == uuid.Nil || ownerProfileID == "" || submittedAt.IsZero() {
		return InactiveJobGrant{}, ErrNotConfigured
	}
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return InactiveJobGrant{}, err
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
		return InactiveJobGrant{}, ErrInvalidTransition
	}
	if err != nil {
		return InactiveJobGrant{}, err
	}
	definition, present := s.catalog.definition(jobKind)
	if !present || definition.Extension == nil || definition.Extension.OwnerProfileID != ownerProfileID ||
		storedOwnerProfileID != ownerProfileID || !storedSubmittedAt.Equal(submittedAt) {
		return InactiveJobGrant{}, ErrInvalidJobDefinition
	}
	return InactiveJobGrant{jobID: jobID, ownerProfileID: ownerProfileID, storedSubmittedAt: storedSubmittedAt}, nil
}

// CompleteInactiveJobTx is the narrow reconciliation operation used after an
// extension profile becomes inactive. ValidateInactiveJobTx must already have
// locked and validated the candidate in this transaction. This operation
// preserves the public state sequence and publishes every state change inside
// the caller transaction.
func (s *TransactionService) CompleteInactiveJobTx(ctx context.Context, tx pgx.Tx, grant InactiveJobGrant, outcome InactiveTerminalOutcome, now time.Time) error {
	if s == nil || s.progressIntents == nil || tx == nil || grant.jobID == uuid.Nil ||
		grant.ownerProfileID == "" || grant.storedSubmittedAt.IsZero() || now.IsZero() {
		return ErrNotConfigured
	}
	current, err := getJobTx(ctx, tx, grant.jobID)
	if err != nil {
		return err
	}
	if current.Status == StatusQueued {
		var mutation transitionMutation
		if outcome.status == StatusCanceled {
			mutation, _, err = transitionCancellationTx(ctx, tx, grant.jobID, now)
		} else {
			mutation, err = transitionRunningTx(ctx, tx, grant.jobID, current.Progress, current.Message, now)
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
	params := terminalTransition{JobID: grant.jobID, Progress: current.Progress}
	switch outcome.status {
	case StatusSucceeded, StatusCanceled:
		if outcome.result == nil || outcome.failure != nil {
			return ErrInvalidJobDefinition
		}
		params.ResultSummary = outcome.result
	case StatusFailed:
		if outcome.failure == nil || outcome.result != nil {
			return ErrInvalidJobDefinition
		}
		params.ErrorSummary = outcome.failure
	default:
		return ErrInvalidTransition
	}
	mutation, err := transitionTerminalTx(ctx, tx, params, now, outcome.status)
	if err != nil {
		return err
	}
	return s.appendProgressIntentTx(ctx, tx, mutation.resource)
}

func (s *TransactionService) completeExecutionTerminalTx(
	ctx context.Context,
	tx pgx.Tx,
	execution Execution,
	params terminalTransition,
	now time.Time,
	status string,
	allowCancellationRequested bool,
) (Resource, error) {
	if s == nil || s.progressIntents == nil {
		return Resource{}, ErrNotConfigured
	}
	if err := s.validateExecutionTx(ctx, tx, execution, allowCancellationRequested); err != nil {
		return Resource{}, err
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

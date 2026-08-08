package jobs

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (m *Manager) Create(ctx context.Context, params CreateParams) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	jobKind := params.JobKind
	if jobKind == "" && params.Extension != nil {
		jobKind = params.Extension.JobKind
	}
	jobKind = m.catalogJobKind(jobKind)
	ctx, span := m.startJobSpan(ctx, "cartulary.jobs.enqueue", jobKind, "enqueue")
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		m.finishJobSpan(span, "enqueue", jobKind, "", resultForJobError(err), err)
		return Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	resource, err := m.transactions.CreateQueuedTx(ctx, tx, params, m.now().UTC())
	if err == nil {
		err = tx.Commit(ctx)
	}
	m.finishJobSpan(span, "enqueue", jobKind, "", resultForJobError(err), err)
	if err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func createDefinedQueuedTx(ctx context.Context, tx pgx.Tx, params CreateParams, definition ExtensionJobContract, now time.Time) (Resource, error) {
	if err := validateScope(params.Scope); err != nil {
		return Resource{}, err
	}
	if err := validateInitialProgress(params.Progress); err != nil {
		return Resource{}, err
	}
	authPolicy, err := normalizeAuthPolicy(params.Scope, params.AuthPolicy)
	if err != nil {
		return Resource{}, err
	}
	if params.SubmittedByUserID == uuid.Nil {
		return Resource{}, fmt.Errorf("%w: missing submitted_by_user_id", ErrInvalidJobDefinition)
	}
	if retiredProfileHandler(params.HandlerName) {
		return Resource{}, fmt.Errorf("%w: retired extension handler %q", ErrInvalidJobDefinition, params.HandlerName)
	}
	if err := validateExtensionAdmission(params); err != nil {
		return Resource{}, err
	}
	if params.JobKind != definition.JobKind || !validProgressUnitID(definition.ProgressUnitID) {
		return Resource{}, fmt.Errorf("%w: unknown job kind", ErrInvalidJobDefinition)
	}
	if params.Extension != nil && params.Extension.JobKind != params.JobKind {
		return Resource{}, fmt.Errorf("%w: extension job kind mismatch", ErrInvalidJobDefinition)
	}
	var handlerPayload []byte
	if len(params.HandlerPayload) > 0 {
		if params.HandlerName == "" {
			return Resource{}, fmt.Errorf("%w: handler payload without handler name", ErrInvalidJobDefinition)
		}
		if !json.Valid(params.HandlerPayload) {
			return Resource{}, fmt.Errorf("%w: invalid handler payload json", ErrInvalidJobDefinition)
		}
		handlerPayload = append([]byte(nil), params.HandlerPayload...)
	}
	record, err := scanJob(tx.QueryRow(ctx, `
INSERT INTO jobs (
    scope_kind, incident_id, status, cancelable, auth_policy, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, progress_total, message,
    handler_name, handler_payload_json, extension_owner_profile_id,
    job_kind, progress_unit_id, extension_idempotency_identity,
    extension_idempotency_route_key, extension_idempotency_scope_key,
    extension_normalized_request_sha256
)
VALUES ($1, $2, 'queued', $3, $4, $5, $6, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, params.Scope.Kind, params.Scope.IncidentID, params.Cancelable, authPolicy, params.SubmittedByUserID, now, params.Progress.Completed, params.Progress.Total, params.Message, nullableText(params.HandlerName), handlerPayload,
		extensionOwner(params.Extension), definition.JobKind, definition.ProgressUnitID, extensionIdempotencyIdentity(params.Extension),
		extensionIdempotencyRouteKey(params.Extension), extensionIdempotencyScopeKey(params.Extension), extensionRequestDigest(params.Extension)))
	if err != nil {
		return Resource{}, err
	}
	return record, nil
}

func validateExtensionAdmission(params CreateParams) error {
	if params.Extension == nil {
		return nil
	}
	admission := params.Extension
	if params.HandlerName == "" || admission.OwnerProfileID == "" || admission.JobKind == "" ||
		admission.IdempotencyRouteKey == "" || admission.IdempotencyScopeKey == "" ||
		len(admission.IdempotencyIdentity) == 0 || !json.Valid(admission.IdempotencyIdentity) ||
		len(admission.NormalizedRequestSHA256) != 64 {
		return fmt.Errorf("%w: incomplete extension job admission", ErrInvalidJobDefinition)
	}
	var identity routeScopedIdempotencyIdentity
	var identityMembers map[string]json.RawMessage
	if err := json.Unmarshal(admission.IdempotencyIdentity, &identityMembers); err != nil ||
		len(identityMembers) != 6 ||
		!hasExactIdentityMembers(identityMembers) {
		return fmt.Errorf("%w: invalid extension idempotency identity", ErrInvalidJobDefinition)
	}
	if err := json.Unmarshal(admission.IdempotencyIdentity, &identity); err != nil ||
		identity.SchemaID != "cartulary.route_scoped_idempotency_identity.v1" ||
		identity.ActorUserID != params.SubmittedByUserID.String() ||
		identity.ScopeKind != params.Scope.Kind ||
		identity.RouteIdentity != admission.IdempotencyRouteKey+":"+admission.IdempotencyScopeKey ||
		identity.ClientTxnID == "" {
		return fmt.Errorf("%w: invalid extension idempotency identity", ErrInvalidJobDefinition)
	}
	if params.Scope.Kind == ScopeKindIncident {
		if identity.ScopeID == nil || params.Scope.IncidentID == nil || *identity.ScopeID != params.Scope.IncidentID.String() {
			return fmt.Errorf("%w: extension incident scope mismatch", ErrInvalidJobDefinition)
		}
	} else if identity.ScopeID != nil {
		return fmt.Errorf("%w: extension deployment scope must omit scope_id", ErrInvalidJobDefinition)
	}
	for _, character := range admission.NormalizedRequestSHA256 {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return fmt.Errorf("%w: invalid extension request digest", ErrInvalidJobDefinition)
		}
	}
	return nil
}

func hasExactIdentityMembers(members map[string]json.RawMessage) bool {
	for _, key := range []string{
		"schema_id", "actor_user_id", "route_identity",
		"scope_kind", "scope_id", "client_txn_id",
	} {
		if _, present := members[key]; !present {
			return false
		}
	}
	return true
}

func retiredProfileHandler(handlerName string) bool {
	switch handlerName {
	case "imports.discovery", "imports.apply", "incident_bundles.execute", "reference_data.execute", "reporting.execute":
		return true
	default:
		return false
	}
}

func extensionOwner(admission *ExtensionJobAdmission) *string {
	if admission == nil {
		return nil
	}
	return &admission.OwnerProfileID
}

func extensionIdempotencyIdentity(admission *ExtensionJobAdmission) json.RawMessage {
	if admission == nil {
		return nil
	}
	return admission.IdempotencyIdentity
}

func extensionRequestDigest(admission *ExtensionJobAdmission) *string {
	if admission == nil {
		return nil
	}
	return &admission.NormalizedRequestSHA256
}

func extensionIdempotencyRouteKey(admission *ExtensionJobAdmission) *string {
	if admission == nil {
		return nil
	}
	return &admission.IdempotencyRouteKey
}

func extensionIdempotencyScopeKey(admission *ExtensionJobAdmission) *string {
	if admission == nil {
		return nil
	}
	return &admission.IdempotencyScopeKey
}

func (m *Manager) Get(ctx context.Context, jobID uuid.UUID) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	stored, err := scanStoredJob(m.pool.QueryRow(ctx, `
SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
       auth_policy,
       submitted_at, updated_at, progress_completed, progress_total, started_at,
       finished_at, retained_until, result_summary_json, error_summary_json, message,
       job_kind, progress_unit_id
  FROM jobs
 WHERE job_id = $1
`, jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrNotFound
	}
	return stored.publicResource(), err
}

func (m *Manager) MarkRunning(ctx context.Context, jobID uuid.UUID, progress Progress, message *string) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	now := m.now().UTC()
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	mutation, err := transitionRunningTx(ctx, tx, jobID, progress, message, now)
	if err != nil {
		return Resource{}, err
	}
	if mutation.changed {
		if err := m.transactions.appendProgressIntentTx(ctx, tx, mutation.resource); err != nil {
			return Resource{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Resource{}, err
	}
	return mutation.resource, nil
}

func (m *Manager) CompleteSucceeded(ctx context.Context, params TransitionParams) (Resource, error) {
	return m.completeTerminal(ctx, params, StatusSucceeded)
}

func (m *Manager) CompleteFailed(ctx context.Context, params TransitionParams) (Resource, error) {
	return m.completeTerminal(ctx, params, StatusFailed)
}

func (m *Manager) CompleteCanceled(ctx context.Context, params TransitionParams) (Resource, error) {
	if params.ResultSummary == nil {
		params.ResultSummary = &ResultSummary{Code: "job_canceled", Message: "Job canceled."}
	}
	return m.completeTerminal(ctx, params, StatusCanceled)
}

// CompleteSucceededTx joins terminal job-result publication to an
// owner-controlled final transaction. Callers publish the returned resource to
// live progress subscribers only after the enclosing commit is proven.
func completeSucceededTx(ctx context.Context, tx pgx.Tx, params TransitionParams, now time.Time) (Resource, error) {
	return completeTerminalTx(ctx, tx, params, now, StatusSucceeded)
}

func completeTerminalTx(
	ctx context.Context,
	tx pgx.Tx,
	params TransitionParams,
	now time.Time,
	status string,
) (Resource, error) {
	if status == StatusCanceled && params.ResultSummary == nil {
		params.ResultSummary = &ResultSummary{Code: "job_canceled", Message: "Job canceled."}
	}
	mutation, err := transitionTerminalTx(ctx, tx, params, now, status)
	if err != nil {
		return Resource{}, err
	}
	return mutation.resource, nil
}

func (m *Manager) Cancel(ctx context.Context, params CancelParams) (CancelResult, error) {
	if err := m.ensureConfigured(); err != nil {
		return CancelResult{}, err
	}
	if params.JobID == uuid.Nil || params.ActorUserID == uuid.Nil || params.ClientTxnID == "" {
		return CancelResult{}, fmt.Errorf("%w: missing cancel key", ErrInvalidJobDefinition)
	}
	sum := sha256.Sum256(params.NormalizedRequest)
	requestHash := sum[:]
	key := RouteIdempotencyKey{
		RouteKey:    "jobs.cancel",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.JobID.String(),
		ClientTxnID: params.ClientTxnID,
	}
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return CancelResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	existing, present, err := m.transactions.routeIdempotency.LookupRouteIdempotencyTx(ctx, tx, key)
	if err != nil {
		return CancelResult{}, err
	}
	if present {
		if subtle.ConstantTimeCompare(existing.RequestHash, requestHash) != 1 {
			return CancelResult{}, ErrClientTxnConflict
		}
		resource, err := getJobTx(ctx, tx, params.JobID)
		if err != nil {
			return CancelResult{}, err
		}
		return CancelResult{Resource: resource, Replayed: true}, tx.Commit(ctx)
	}

	mutation, reason, err := transitionCancellationTx(ctx, tx, params.JobID, m.now().UTC())
	if err != nil {
		return CancelResult{}, err
	}
	if reason != "" {
		return CancelResult{ReasonCode: reason}, ErrCancelRejected
	}
	resource := mutation.resource
	if err := m.transactions.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return CancelResult{}, err
	}
	ownerProfileID, err := extensionOwnerProfileIDTx(ctx, tx, params.JobID)
	if err != nil {
		return CancelResult{}, err
	}
	if ownerProfileID != nil {
		if err := m.transactions.extensionCancellation.AppendExtensionCancellationObservationTx(ctx, tx, ExtensionCancellationObservation{
			Key: key, JobID: params.JobID, ObservedAt: resource.UpdatedAt,
		}); err != nil {
			return CancelResult{}, err
		}
	}
	if err := m.transactions.routeIdempotency.CommitRouteIdempotencyTx(ctx, tx, key, requestHash, 200, resource); err != nil {
		return CancelResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CancelResult{}, err
	}
	return CancelResult{Resource: resource}, nil
}

func extensionOwnerProfileIDTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (*string, error) {
	var ownerProfileID *string
	if err := tx.QueryRow(ctx, `
SELECT extension_owner_profile_id
  FROM jobs
 WHERE job_id = $1
`, jobID).Scan(&ownerProfileID); err != nil {
		return nil, err
	}
	return ownerProfileID, nil
}

func (m *Manager) completeTerminal(ctx context.Context, params TransitionParams, status string) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	ctx, span := m.startJobSpan(ctx, "cartulary.jobs.run", "unknown", "run")
	now := m.now().UTC()
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if status == StatusCanceled && params.ResultSummary == nil {
		params.ResultSummary = &ResultSummary{Code: "job_canceled", Message: "Job canceled."}
	}
	mutation, err := transitionTerminalTx(ctx, tx, params, now, status)
	if err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	record := mutation.resource
	jobKind, kindErr := jobKindTx(ctx, tx, params.JobID)
	if kindErr != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", kindErr)
		return Resource{}, kindErr
	}
	jobKind = m.catalogJobKind(jobKind)
	if err := m.transactions.appendProgressIntentTx(ctx, tx, record); err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	result := resultForTerminalStatus(record.Status)
	m.finishJobSpan(span, "run", jobKind, record.Status, result, nil)
	m.recordJobDuration(ctx, record, jobKind, result)
	return record, nil
}

func nullableText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// lockTransitionTx serializes current-state checks and commits that can change
// a job's cancellation or terminal state. Owner transactions that commit work
// for a job must acquire this lock before revalidating the job status.
func lockTransitionTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) error {
	if tx == nil || jobID == uuid.Nil {
		return ErrInvalidJobDefinition
	}
	const transitionLockSeed int64 = 49006006
	_, err := tx.Exec(ctx, `
SELECT pg_advisory_xact_lock(hashtextextended($1::text, $2))
`, jobID, transitionLockSeed)
	return err
}

func validateScope(scope Scope) error {
	switch scope.Kind {
	case ScopeKindIncident:
		if scope.IncidentID == nil || *scope.IncidentID == uuid.Nil {
			return fmt.Errorf("%w: missing incident_id", ErrInvalidJobDefinition)
		}
	case ScopeKindDeployment:
		if scope.IncidentID != nil {
			return fmt.Errorf("%w: deployment job has incident_id", ErrInvalidJobDefinition)
		}
	default:
		return fmt.Errorf("%w: unknown scope kind", ErrInvalidJobDefinition)
	}
	return nil
}

func normalizeAuthPolicy(scope Scope, authPolicy string) (string, error) {
	if authPolicy == "" {
		switch scope.Kind {
		case ScopeKindIncident:
			return AuthPolicyIncidentMembership, nil
		case ScopeKindDeployment:
			return AuthPolicySubmitterOrDeploymentAdmin, nil
		default:
			return "", fmt.Errorf("%w: unknown scope kind", ErrInvalidJobDefinition)
		}
	}
	switch scope.Kind {
	case ScopeKindIncident:
		if authPolicy == AuthPolicyIncidentMembership || authPolicy == AuthPolicyDeploymentAdminIncidentMembership {
			return authPolicy, nil
		}
	case ScopeKindDeployment:
		if authPolicy == AuthPolicySubmitterOrDeploymentAdmin || authPolicy == AuthPolicyDeploymentAdmin {
			return authPolicy, nil
		}
	}
	return "", fmt.Errorf("%w: auth policy %q not valid for %s job", ErrInvalidJobDefinition, authPolicy, scope.Kind)
}

func marshalSummaries(result *ResultSummary, failure *ErrorSummary, status string) ([]byte, []byte, error) {
	switch status {
	case StatusSucceeded, StatusCanceled:
		if result == nil || failure != nil {
			return nil, nil, fmt.Errorf("%w: terminal success requires result summary only", ErrInvalidJobDefinition)
		}
	case StatusFailed:
		if failure == nil || result != nil {
			return nil, nil, fmt.Errorf("%w: terminal failure requires error summary only", ErrInvalidJobDefinition)
		}
	default:
		return nil, nil, fmt.Errorf("%w: unsupported terminal status", ErrInvalidTransition)
	}
	var resultJSON []byte
	var errorJSON []byte
	var err error
	if result != nil {
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return nil, nil, err
		}
	}
	if failure != nil {
		errorJSON, err = json.Marshal(failure)
		if err != nil {
			return nil, nil, err
		}
	}
	return resultJSON, errorJSON, nil
}

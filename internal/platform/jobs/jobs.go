package jobs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	ScopeKindIncident   = "incident"
	ScopeKindDeployment = "deployment"

	AuthPolicyIncidentMembership                = "incident_membership"
	AuthPolicyDeploymentAdminIncidentMembership = "deployment_admin_incident_membership"
	AuthPolicySubmitterOrDeploymentAdmin        = "submitter_or_deployment_admin"
	AuthPolicyDeploymentAdmin                   = "deployment_admin"

	StatusQueued          = "queued"
	StatusRunning         = "running"
	StatusCancelRequested = "cancel_requested"
	StatusSucceeded       = "succeeded"
	StatusFailed          = "failed"
	StatusCanceled        = "canceled"

	CancelReasonAlreadyCancelRequested = "already_cancel_requested"
	CancelReasonAlreadyTerminal        = "already_terminal"
	CancelReasonNotCancelable          = "not_cancelable"
)

var (
	ErrNotConfigured        = errors.New("jobs: manager not configured")
	ErrNotFound             = errors.New("jobs: not found")
	ErrClientTxnConflict    = errors.New("jobs: client transaction conflict")
	ErrCancelRejected       = errors.New("jobs: cancel rejected")
	ErrInvalidTransition    = errors.New("jobs: invalid transition")
	ErrInvalidJobDefinition = errors.New("jobs: invalid job definition")
)

type Manager struct {
	pool                  *pgxpool.Pool
	now                   func() time.Time
	transactions          *TransactionService
	serviceVersion        string
	activeGaugeRegistered bool
	extensionContracts    map[string]ExtensionJobContract
}

type Scope struct {
	Kind       string     `json:"kind"`
	IncidentID *uuid.UUID `json:"incident_id,omitempty"`
}

type Progress struct {
	Completed int  `json:"completed"`
	Total     *int `json:"total"`
}

type ResourceRef struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Route string `json:"route"`
}

type ResultSummary struct {
	Code         string        `json:"code"`
	Message      string        `json:"message"`
	ResourceRefs []ResourceRef `json:"resource_refs,omitempty"`
}

type ErrorSummary struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

type Resource struct {
	JobID             string         `json:"job_id"`
	Scope             Scope          `json:"scope"`
	StatusRoute       string         `json:"status_route"`
	Status            string         `json:"status"`
	Cancelable        bool           `json:"cancelable"`
	AuthPolicy        string         `json:"-"`
	SubmittedByUserID string         `json:"submitted_by_user_id"`
	SubmittedAt       time.Time      `json:"submitted_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	Progress          Progress       `json:"progress"`
	StartedAt         *time.Time     `json:"started_at"`
	FinishedAt        *time.Time     `json:"finished_at"`
	RetainedUntil     *time.Time     `json:"retained_until"`
	ResultSummary     *ResultSummary `json:"result_summary"`
	ErrorSummary      *ErrorSummary  `json:"error_summary"`
	Message           *string        `json:"message,omitempty"`
}

type CreateParams struct {
	Scope             Scope
	SubmittedByUserID uuid.UUID
	AuthPolicy        string
	Cancelable        bool
	Progress          Progress
	Message           *string
	HandlerName       string
	HandlerPayload    json.RawMessage
	Extension         *ExtensionJobAdmission
}

// ExtensionJobAdmission is durable internal ownership and replay evidence. It
// is intentionally absent from Resource and therefore never enters the public
// common-job envelope.
type ExtensionJobAdmission struct {
	OwnerProfileID          string
	JobKind                 string
	IdempotencyIdentity     json.RawMessage
	IdempotencyRouteKey     string
	IdempotencyScopeKey     string
	NormalizedRequestSHA256 string
}

type RouteScopedIdempotencyIdentity struct {
	SchemaID      string  `json:"schema_id"`
	ActorUserID   string  `json:"actor_user_id"`
	RouteIdentity string  `json:"route_identity"`
	ScopeKind     string  `json:"scope_kind"`
	ScopeID       *string `json:"scope_id"`
	ClientTxnID   string  `json:"client_txn_id"`
}

func NewExtensionJobAdmission(ownerProfileID string, jobKind string, key authn.RouteIdempotencyKey, scope Scope, normalizedRequest []byte) (*ExtensionJobAdmission, error) {
	if ownerProfileID == "" || jobKind == "" || key.ActorUserID == uuid.Nil ||
		key.RouteKey == "" || key.ScopeKey == "" || key.ClientTxnID == "" || len(normalizedRequest) == 0 {
		return nil, fmt.Errorf("%w: incomplete extension job admission", ErrInvalidJobDefinition)
	}
	if err := validateScope(scope); err != nil {
		return nil, err
	}
	var scopeID *string
	if scope.Kind == ScopeKindIncident {
		value := scope.IncidentID.String()
		scopeID = &value
	}
	identity, err := json.Marshal(RouteScopedIdempotencyIdentity{
		SchemaID:      "cartulary.route_scoped_idempotency_identity.v1",
		ActorUserID:   key.ActorUserID.String(),
		RouteIdentity: key.RouteKey + ":" + key.ScopeKey,
		ScopeKind:     scope.Kind,
		ScopeID:       scopeID,
		ClientTxnID:   key.ClientTxnID,
	})
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(normalizedRequest)
	return &ExtensionJobAdmission{
		OwnerProfileID:          ownerProfileID,
		JobKind:                 jobKind,
		IdempotencyIdentity:     identity,
		IdempotencyRouteKey:     key.RouteKey,
		IdempotencyScopeKey:     key.ScopeKey,
		NormalizedRequestSHA256: fmt.Sprintf("%x", digest[:]),
	}, nil
}

type TransitionParams struct {
	JobID         uuid.UUID
	Progress      Progress
	ResultSummary *ResultSummary
	ErrorSummary  *ErrorSummary
	Message       *string
}

type CancelParams struct {
	JobID             uuid.UUID
	ActorUserID       uuid.UUID
	ClientTxnID       string
	NormalizedRequest []byte
}

type CancelResult struct {
	Resource   Resource
	Replayed   bool
	ReasonCode string
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Configure(pool *pgxpool.Pool, transactions *TransactionService, now func() time.Time) {
	m.pool = pool
	m.transactions = transactions
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	m.now = now
	m.registerActiveGauge()
}

func (m *Manager) ConfigureTelemetry(serviceVersion string) {
	if m == nil {
		return
	}
	m.serviceVersion = serviceVersion
	m.registerActiveGauge()
}

func (m *Manager) Create(ctx context.Context, params CreateParams) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	ctx, span := m.startJobSpan(ctx, "cartulary.jobs.enqueue", jobKindFromScope(params.Scope), "enqueue")
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		m.finishJobSpan(span, "enqueue", jobKindFromScope(params.Scope), "", resultForJobError(err), err)
		return Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	resource, err := m.transactions.CreateQueuedTx(ctx, tx, params, m.now().UTC())
	if err == nil {
		err = tx.Commit(ctx)
	}
	m.finishJobSpan(span, "enqueue", jobKindFromScope(params.Scope), "", resultForJobError(err), err)
	if err != nil {
		return Resource{}, err
	}
	return resource, nil
}

func createQueuedTx(ctx context.Context, tx pgx.Tx, params CreateParams, now time.Time) (Resource, error) {
	if err := validateScope(params.Scope); err != nil {
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
    extension_job_kind, extension_idempotency_identity,
    extension_idempotency_route_key, extension_idempotency_scope_key,
    extension_normalized_request_sha256
)
VALUES ($1, $2, 'queued', $3, $4, $5, $6, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, params.Scope.Kind, params.Scope.IncidentID, params.Cancelable, authPolicy, params.SubmittedByUserID, now, params.Progress.Completed, params.Progress.Total, params.Message, nullableText(params.HandlerName), handlerPayload,
		extensionOwner(params.Extension), extensionJobKind(params.Extension), extensionIdempotencyIdentity(params.Extension),
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
	var identity RouteScopedIdempotencyIdentity
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

func extensionJobKind(admission *ExtensionJobAdmission) *string {
	if admission == nil {
		return nil
	}
	return &admission.JobKind
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
	record, err := scanJob(m.pool.QueryRow(ctx, `
SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
       auth_policy,
       submitted_at, updated_at, progress_completed, progress_total, started_at,
       finished_at, retained_until, result_summary_json, error_summary_json, message
  FROM jobs
 WHERE job_id = $1
`, jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrNotFound
	}
	return record, err
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
	record, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = 'running',
       started_at = COALESCE(started_at, $2),
       updated_at = $2,
       progress_completed = $3,
       progress_total = $4,
       message = $5
 WHERE job_id = $1
   AND status = 'queued'
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, now, progress.Completed, progress.Total, message))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrInvalidTransition
	}
	if err != nil {
		return Resource{}, err
	}
	if err := m.transactions.appendProgressIntentTx(ctx, tx, record); err != nil {
		return Resource{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Resource{}, err
	}
	return record, nil
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
	if tx == nil || params.JobID == uuid.Nil || now.IsZero() {
		return Resource{}, ErrInvalidJobDefinition
	}
	resultJSON, errorJSON, err := marshalSummaries(params.ResultSummary, params.ErrorSummary, StatusSucceeded)
	if err != nil {
		return Resource{}, err
	}
	now = now.UTC()
	record, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = 'succeeded',
       cancelable = false,
       updated_at = $2,
       finished_at = $2,
       retained_until = $3,
       progress_completed = $4,
       progress_total = $5,
       result_summary_json = $6,
       error_summary_json = $7,
       message = $8
 WHERE job_id = $1
   AND status IN ('queued', 'running', 'cancel_requested')
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, params.JobID, now, now.Add(7*24*time.Hour), params.Progress.Completed, params.Progress.Total, resultJSON, errorJSON, params.Message))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrInvalidTransition
	}
	if err != nil {
		return Resource{}, err
	}
	return record, nil
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
	key := authn.RouteIdempotencyKey{
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

	var existing authn.RouteIdempotencyRecord
	row := tx.QueryRow(ctx, `
SELECT route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID)
	if err := row.Scan(&existing.RouteKey, &existing.ScopeKey, &existing.ClientTxnID, &existing.ActorUserID, &existing.RequestHash, &existing.StatusCode, &existing.ResponseJSON); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return CancelResult{}, ErrClientTxnConflict
		}
		resource, err := getJobTx(ctx, tx, params.JobID)
		if err != nil {
			return CancelResult{}, err
		}
		return CancelResult{Resource: resource, Replayed: true}, tx.Commit(ctx)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return CancelResult{}, err
	}

	resource, reason, err := cancelJobTx(ctx, tx, params.JobID, m.now().UTC())
	if err != nil {
		return CancelResult{}, err
	}
	if reason != "" {
		return CancelResult{ReasonCode: reason}, ErrCancelRejected
	}
	if err := m.transactions.appendProgressIntentTx(ctx, tx, resource); err != nil {
		return CancelResult{}, err
	}
	if err := recordExtensionCancellationObservationTx(ctx, tx, key, params.JobID, resource.UpdatedAt); err != nil {
		return CancelResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, 200, resource); err != nil {
		return CancelResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CancelResult{}, err
	}
	return CancelResult{Resource: resource}, nil
}

func recordExtensionCancellationObservationTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, jobID uuid.UUID, observedAt time.Time) error {
	var ownerProfileID *string
	if err := tx.QueryRow(ctx, `
SELECT extension_owner_profile_id
  FROM jobs
 WHERE job_id = $1
`, jobID).Scan(&ownerProfileID); err != nil {
		return err
	}
	if ownerProfileID == nil {
		return nil
	}
	identity := key.RouteKey + "\x00" + key.ActorUserID.String() + "\x00" +
		key.ScopeKey + "\x00" + key.ClientTxnID
	digest := sha256.Sum256([]byte(identity))
	cancellationRequestID := fmt.Sprintf("cancel:%x", digest[:])
	_, err := tx.Exec(ctx, `
INSERT INTO extension_job_cancellation_observations (
    cancellation_request_id, job_id, observed_at, observed_before_final_commit
) VALUES ($1, $2, $3, TRUE)
`, cancellationRequestID, jobID, observedAt.UTC())
	return err
}

func (m *Manager) completeTerminal(ctx context.Context, params TransitionParams, status string) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	ctx, span := m.startJobSpan(ctx, "cartulary.jobs.run", "unknown", "run")
	now := m.now().UTC()
	retainedUntil := now.Add(7 * 24 * time.Hour)
	resultJSON, errorJSON, err := marshalSummaries(params.ResultSummary, params.ErrorSummary, status)
	if err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	record, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = $2,
       cancelable = false,
       updated_at = $3,
       finished_at = $3,
       retained_until = $4,
       progress_completed = $5,
       progress_total = $6,
       result_summary_json = $7,
       error_summary_json = $8,
       message = $9
 WHERE job_id = $1
   AND status IN ('queued', 'running', 'cancel_requested')
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, params.JobID, status, now, retainedUntil, params.Progress.Completed, params.Progress.Total, resultJSON, errorJSON, params.Message))
	if errors.Is(err, pgx.ErrNoRows) {
		m.finishJobSpan(span, "run", "unknown", "", "failed", ErrInvalidTransition)
		return Resource{}, ErrInvalidTransition
	}
	if err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	if err := m.transactions.appendProgressIntentTx(ctx, tx, record); err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		m.finishJobSpan(span, "run", "unknown", "", "failed", err)
		return Resource{}, err
	}
	result := resultForTerminalStatus(record.Status)
	jobKind := jobKindFromScope(record.Scope)
	m.finishJobSpan(span, "run", jobKind, record.Status, result, nil)
	m.recordJobDuration(ctx, record, jobKind, result)
	return record, nil
}

func (m *Manager) ensureConfigured() error {
	if m == nil || m.pool == nil || m.transactions == nil {
		return ErrNotConfigured
	}
	if m.now == nil {
		m.now = func() time.Time { return time.Now().UTC() }
	}
	return nil
}

func nullableText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func cancelJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, now time.Time) (Resource, string, error) {
	current, err := getJobTx(ctx, tx, jobID)
	if errors.Is(err, ErrNotFound) {
		return Resource{}, "", ErrNotFound
	}
	if err != nil {
		return Resource{}, "", err
	}
	switch current.Status {
	case StatusCancelRequested:
		return Resource{}, CancelReasonAlreadyCancelRequested, nil
	case StatusSucceeded, StatusFailed, StatusCanceled:
		return Resource{}, CancelReasonAlreadyTerminal, nil
	}
	if !current.Cancelable {
		return Resource{}, CancelReasonNotCancelable, nil
	}
	record, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = 'cancel_requested',
       cancelable = false,
       updated_at = $2
 WHERE job_id = $1
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, now.UTC()))
	return record, "", err
}

func getJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (Resource, error) {
	record, err := scanJob(tx.QueryRow(ctx, `
SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
       auth_policy,
       submitted_at, updated_at, progress_completed, progress_total, started_at,
       finished_at, retained_until, result_summary_json, error_summary_json, message
  FROM jobs
 WHERE job_id = $1
 FOR UPDATE
`, jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrNotFound
	}
	return record, err
}

func scanJob(row pgx.Row) (Resource, error) {
	var record Resource
	var jobID uuid.UUID
	var submittedBy uuid.UUID
	var incidentID *uuid.UUID
	var progressTotal *int
	var resultJSON []byte
	var errorJSON []byte
	if err := row.Scan(
		&jobID,
		&record.Scope.Kind,
		&incidentID,
		&record.Status,
		&record.Cancelable,
		&submittedBy,
		&record.AuthPolicy,
		&record.SubmittedAt,
		&record.UpdatedAt,
		&record.Progress.Completed,
		&progressTotal,
		&record.StartedAt,
		&record.FinishedAt,
		&record.RetainedUntil,
		&resultJSON,
		&errorJSON,
		&record.Message,
	); err != nil {
		return Resource{}, err
	}
	record.JobID = jobID.String()
	record.Scope.IncidentID = incidentID
	record.StatusRoute = "/api/v1/jobs/" + record.JobID
	record.SubmittedByUserID = submittedBy.String()
	record.Progress.Total = progressTotal
	if len(resultJSON) > 0 {
		var summary ResultSummary
		if err := json.Unmarshal(resultJSON, &summary); err != nil {
			return Resource{}, err
		}
		record.ResultSummary = &summary
	}
	if len(errorJSON) > 0 {
		var summary ErrorSummary
		if err := json.Unmarshal(errorJSON, &summary); err != nil {
			return Resource{}, err
		}
		record.ErrorSummary = &summary
	}
	return record, nil
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

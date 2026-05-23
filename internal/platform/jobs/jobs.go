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
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const (
	ScopeKindIncident   = "incident"
	ScopeKindDeployment = "deployment"

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
	pool *pgxpool.Pool
	now  func() time.Time
	hub  *platformws.Hub
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
	Cancelable        bool
	Progress          Progress
	Message           *string
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

func (m *Manager) Configure(pool *pgxpool.Pool, now func() time.Time) {
	m.pool = pool
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	m.now = now
}

func (m *Manager) ConfigureProgressHub(hub *platformws.Hub) {
	m.hub = hub
}

func (m *Manager) Create(ctx context.Context, params CreateParams) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	resource, err := CreateQueuedTx(ctx, m.pool, params, m.now().UTC())
	if err != nil {
		return Resource{}, err
	}
	m.PublishProgress(resource)
	return resource, nil
}

type queuedJobInserter interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func CreateQueuedTx(ctx context.Context, tx queuedJobInserter, params CreateParams, now time.Time) (Resource, error) {
	if err := validateScope(params.Scope); err != nil {
		return Resource{}, err
	}
	if params.SubmittedByUserID == uuid.Nil {
		return Resource{}, fmt.Errorf("%w: missing submitted_by_user_id", ErrInvalidJobDefinition)
	}
	record, err := scanJob(tx.QueryRow(ctx, `
INSERT INTO jobs (
    scope_kind, incident_id, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, progress_total, message
)
VALUES ($1, $2, 'queued', $3, $4, $5, $5, $6, $7, $8)
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, params.Scope.Kind, params.Scope.IncidentID, params.Cancelable, params.SubmittedByUserID, now, params.Progress.Completed, params.Progress.Total, params.Message))
	if err != nil {
		return Resource{}, err
	}
	return record, nil
}

func (m *Manager) Get(ctx context.Context, jobID uuid.UUID) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	record, err := scanJob(m.pool.QueryRow(ctx, `
SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
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
	record, err := scanJob(m.pool.QueryRow(ctx, `
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
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, now, progress.Completed, progress.Total, message))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrInvalidTransition
	}
	if err != nil {
		return Resource{}, err
	}
	m.PublishProgress(record)
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
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, 200, resource); err != nil {
		return CancelResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CancelResult{}, err
	}
	m.PublishProgress(resource)
	return CancelResult{Resource: resource}, nil
}

func (m *Manager) completeTerminal(ctx context.Context, params TransitionParams, status string) (Resource, error) {
	if err := m.ensureConfigured(); err != nil {
		return Resource{}, err
	}
	now := m.now().UTC()
	retainedUntil := now.Add(7 * 24 * time.Hour)
	resultJSON, errorJSON, err := marshalSummaries(params.ResultSummary, params.ErrorSummary, status)
	if err != nil {
		return Resource{}, err
	}
	record, err := scanJob(m.pool.QueryRow(ctx, `
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
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, params.JobID, status, now, retainedUntil, params.Progress.Completed, params.Progress.Total, resultJSON, errorJSON, params.Message))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrInvalidTransition
	}
	if err != nil {
		return Resource{}, err
	}
	m.PublishProgress(record)
	return record, nil
}

func (m *Manager) ensureConfigured() error {
	if m == nil || m.pool == nil {
		return ErrNotConfigured
	}
	if m.now == nil {
		m.now = func() time.Time { return time.Now().UTC() }
	}
	return nil
}

func (m *Manager) PublishProgress(resource Resource) {
	if m == nil || m.hub == nil || resource.Scope.Kind != ScopeKindIncident || resource.Scope.IncidentID == nil {
		return
	}
	payload := platformws.NewIncidentJobProgressPayload(resource.JobID, *resource.Scope.IncidentID, resource.Status, platformws.JobProgress{
		Completed: int64(resource.Progress.Completed),
		Total:     intPointerToInt64(resource.Progress.Total),
	}, resource.UpdatedAt)
	cancelable := resource.Cancelable
	payload.Cancelable = &cancelable
	if resource.Message != nil {
		payload.Message = *resource.Message
	}
	payload.ResultSummary = resource.ResultSummary
	payload.ErrorSummary = resource.ErrorSummary
	payload.RetainedUntil = resource.RetainedUntil
	_ = m.hub.PublishJobProgress(*resource.Scope.IncidentID, payload)
}

func intPointerToInt64(value *int) *int64 {
	if value == nil {
		return nil
	}
	converted := int64(*value)
	return &converted
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
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, now.UTC()))
	return record, "", err
}

func getJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (Resource, error) {
	record, err := scanJob(tx.QueryRow(ctx, `
SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
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

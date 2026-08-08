package jobs

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
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
	ErrNotConfigured         = errors.New("jobs: manager not configured")
	ErrNotFound              = errors.New("jobs: not found")
	ErrClientTxnConflict     = errors.New("jobs: client transaction conflict")
	ErrCancelRejected        = errors.New("jobs: cancel rejected")
	ErrInvalidTransition     = errors.New("jobs: invalid transition")
	ErrCancellationRequested = errors.New("jobs: cancellation requested")
	ErrInvalidJobDefinition  = errors.New("jobs: invalid job definition")
	ErrStorageIncompatible   = errors.New("jobs: storage incompatible")
	ErrExecutionLost         = errors.New("jobs: execution ownership lost")
)

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

type EnqueueParams struct {
	JobKind           string
	Scope             Scope
	SubmittedByUserID uuid.UUID
	AuthPolicy        string
	Cancelable        bool
	Progress          Progress
	Message           *string
	HandlerPayload    json.RawMessage
	Extension         *ExtensionJobAdmission
}

// Execution is an opaque lease-fenced handler attempt. Only Jobs can create a
// valid value; consumers may inspect the stable job ID but never the token.
type Execution struct {
	jobID     uuid.UUID
	attemptID uuid.UUID
}

func newExecution(jobID uuid.UUID, attemptID uuid.UUID) Execution {
	return Execution{jobID: jobID, attemptID: attemptID}
}

func (execution Execution) JobID() uuid.UUID {
	return execution.jobID
}

func (execution Execution) valid() bool {
	return execution.jobID != uuid.Nil && execution.attemptID != uuid.Nil
}

type SuccessCompletion struct {
	Progress      Progress
	ResultSummary ResultSummary
	Message       *string
}

type FailureCompletion struct {
	Progress     Progress
	ErrorSummary ErrorSummary
	Message      *string
}

type CancellationCompletion struct {
	Progress      Progress
	ResultSummary ResultSummary
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

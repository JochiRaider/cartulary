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

type CreateParams struct {
	JobKind           string
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

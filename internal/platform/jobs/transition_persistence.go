package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func transitionRunningTx(
	ctx context.Context,
	tx pgx.Tx,
	jobID uuid.UUID,
	progress Progress,
	message *string,
	now time.Time,
) (transitionMutation, error) {
	if tx == nil || jobID == uuid.Nil || now.IsZero() {
		return transitionMutation{}, ErrInvalidJobDefinition
	}
	if err := validateProgress(progress); err != nil {
		return transitionMutation{}, err
	}
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return transitionMutation{}, err
	}
	current, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		return transitionMutation{}, err
	}
	if current.Status != StatusQueued && current.Status != StatusRunning {
		return transitionMutation{}, invalidTransition(transitionReasonState)
	}
	if err := validateProgressAdvance(current.Progress, progress); err != nil {
		return transitionMutation{}, err
	}
	if current.Status == StatusRunning && progressEqual(current.Progress, progress) && optionalStringEqual(current.Message, message) {
		return transitionMutation{resource: current}, nil
	}

	record, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = 'running',
       started_at = COALESCE(started_at, $3),
       updated_at = $3,
       progress_completed = $4,
       progress_total = $5,
       message = $6
 WHERE job_id = $1
   AND status = $2
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, current.Status, now.UTC(), progress.Completed, progress.Total, message))
	if errors.Is(err, pgx.ErrNoRows) {
		return transitionMutation{}, invalidTransition(transitionReasonState)
	}
	if err != nil {
		return transitionMutation{}, err
	}
	return transitionMutation{resource: record, changed: true}, nil
}

type terminalTransition struct {
	JobID         uuid.UUID
	Progress      Progress
	ResultSummary *ResultSummary
	ErrorSummary  *ErrorSummary
	Message       *string
}

func transitionTerminalTx(
	ctx context.Context,
	tx pgx.Tx,
	params terminalTransition,
	now time.Time,
	status string,
) (transitionMutation, error) {
	if tx == nil || params.JobID == uuid.Nil || now.IsZero() {
		return transitionMutation{}, ErrInvalidJobDefinition
	}
	allowedSources, recognized := terminalSources[status]
	if !recognized {
		return transitionMutation{}, invalidTransition(transitionReasonState)
	}
	if err := validateProgress(params.Progress); err != nil {
		return transitionMutation{}, err
	}
	resultJSON, errorJSON, err := marshalSummaries(params.ResultSummary, params.ErrorSummary, status)
	if err != nil {
		return transitionMutation{}, err
	}
	if err := lockTransitionTx(ctx, tx, params.JobID); err != nil {
		return transitionMutation{}, err
	}
	current, err := getJobTx(ctx, tx, params.JobID)
	if err != nil {
		return transitionMutation{}, err
	}
	if _, allowed := allowedSources[current.Status]; !allowed {
		return transitionMutation{}, invalidTransition(transitionReasonState)
	}
	if err := validateProgressAdvance(current.Progress, params.Progress); err != nil {
		return transitionMutation{}, err
	}
	if status == StatusSucceeded && params.Progress.Total != nil && params.Progress.Completed != *params.Progress.Total {
		return transitionMutation{}, invalidTransition(transitionReasonProgress)
	}

	now = now.UTC()
	record, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = $3,
       cancelable = false,
       updated_at = $4,
       finished_at = $4,
       retained_until = $5,
       progress_completed = $6,
       progress_total = $7,
       result_summary_json = $8,
       error_summary_json = $9,
       message = $10,
       handler_attempt_id = NULL,
       handler_lease_expires_at = NULL,
       handler_next_attempt_at = NULL
 WHERE job_id = $1
   AND status = $2
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, params.JobID, current.Status, status, now, now.Add(7*24*time.Hour), params.Progress.Completed, params.Progress.Total, resultJSON, errorJSON, params.Message))
	if errors.Is(err, pgx.ErrNoRows) {
		return transitionMutation{}, invalidTransition(transitionReasonState)
	}
	if err != nil {
		return transitionMutation{}, err
	}
	return transitionMutation{resource: record, changed: true}, nil
}

// transitionCancellationLockedTx applies cancellation after the caller has
// acquired the per-job transition advisory lock. Keeping this precondition
// explicit prevents callers that already hold durable rows from introducing a
// row-before-transition lock-order inversion.
func transitionCancellationLockedTx(
	ctx context.Context,
	tx pgx.Tx,
	jobID uuid.UUID,
	now time.Time,
) (transitionMutation, string, error) {
	if tx == nil || jobID == uuid.Nil || now.IsZero() {
		return transitionMutation{}, "", ErrInvalidJobDefinition
	}
	current, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		return transitionMutation{}, "", err
	}
	switch current.Status {
	case StatusCancelRequested:
		return transitionMutation{}, CancelReasonAlreadyCancelRequested, nil
	case StatusSucceeded, StatusFailed, StatusCanceled:
		return transitionMutation{}, CancelReasonAlreadyTerminal, nil
	case StatusQueued, StatusRunning:
	default:
		return transitionMutation{}, CancelReasonNotCancelable, nil
	}
	if !current.Cancelable {
		return transitionMutation{}, CancelReasonNotCancelable, nil
	}
	record, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = 'cancel_requested',
       cancelable = false,
       updated_at = $3
 WHERE job_id = $1
   AND status = $2
   AND cancelable = true
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, current.Status, now.UTC()))
	if errors.Is(err, pgx.ErrNoRows) {
		return transitionMutation{}, CancelReasonNotCancelable, nil
	}
	if err != nil {
		return transitionMutation{}, "", err
	}
	return transitionMutation{resource: record, changed: true}, "", nil
}

func jobKindTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (string, error) {
	var jobKind *string
	if err := tx.QueryRow(ctx, `SELECT job_kind FROM jobs WHERE job_id = $1`, jobID).Scan(&jobKind); err != nil {
		return "", err
	}
	if jobKind == nil {
		return "unknown", nil
	}
	return *jobKind, nil
}

package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	DefaultHandlerMaxAttempts = 3
	HandlerAttemptsExhausted  = "job_handler_attempts_exhausted"
	HandlerExecutionFailed    = "job_handler_execution_failed"
	HandlerIncomplete         = "job_handler_incomplete"
)

func (m *Manager) HandlerPayload(ctx context.Context, jobID uuid.UUID) (json.RawMessage, error) {
	if err := m.ensureConfigured(); err != nil {
		return nil, err
	}
	var payload pgtype.Text
	err := m.pool.QueryRow(ctx, `
SELECT handler_payload_json
  FROM jobs
 WHERE job_id = $1
`, jobID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !payload.Valid {
		return nil, nil
	}
	return append(json.RawMessage(nil), payload.String...), nil
}

func (m *Manager) RecoverableHandlerJobs(ctx context.Context, handlerName string, limit int) ([]uuid.UUID, error) {
	if err := m.ensureConfigured(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	if err := m.failExhaustedHandlerJobs(ctx, handlerName); err != nil {
		return nil, err
	}
	rows, err := m.pool.Query(ctx, `
SELECT job_id
  FROM jobs
 WHERE handler_name = $1
   AND status IN ('queued', 'running', 'cancel_requested')
   AND handler_attempts < handler_max_attempts
   AND (handler_lease_expires_at IS NULL OR handler_lease_expires_at <= $2)
 ORDER BY submitted_at ASC, job_id ASC
 LIMIT $3
`, handlerName, m.now().UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobIDs []uuid.UUID
	for rows.Next() {
		var jobID uuid.UUID
		if err := rows.Scan(&jobID); err != nil {
			return nil, err
		}
		jobIDs = append(jobIDs, jobID)
	}
	return jobIDs, rows.Err()
}

func (m *Manager) ClaimHandlerJob(ctx context.Context, jobID uuid.UUID, handlerName string, owner string, leaseDuration time.Duration) (bool, error) {
	if err := m.ensureConfigured(); err != nil {
		return false, err
	}
	if jobID == uuid.Nil || handlerName == "" || owner == "" {
		return false, ErrInvalidJobDefinition
	}
	if leaseDuration <= 0 {
		leaseDuration = 30 * time.Second
	}
	now := m.now().UTC()
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return false, err
	}
	current, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		return false, err
	}
	if current.Status != StatusQueued && current.Status != StatusRunning && current.Status != StatusCancelRequested {
		return false, tx.Commit(ctx)
	}
	var storedHandlerName *string
	var attempts, maxAttempts int
	var leaseOwner *string
	var leaseExpiresAt *time.Time
	if err := tx.QueryRow(ctx, `
SELECT handler_name, handler_attempts, handler_max_attempts,
       handler_lease_owner, handler_lease_expires_at
  FROM jobs
 WHERE job_id = $1
 FOR UPDATE
`, jobID).Scan(&storedHandlerName, &attempts, &maxAttempts, &leaseOwner, &leaseExpiresAt); err != nil {
		return false, err
	}
	if storedHandlerName == nil || *storedHandlerName != handlerName || attempts >= maxAttempts ||
		(leaseOwner != nil && (leaseExpiresAt == nil || leaseExpiresAt.After(now))) {
		return false, tx.Commit(ctx)
	}
	resource, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = CASE WHEN status = 'queued' THEN 'running' ELSE status END,
       started_at = CASE WHEN status = 'queued' THEN COALESCE(started_at, $6) ELSE started_at END,
       updated_at = CASE WHEN status = 'queued' THEN $6 ELSE updated_at END,
       handler_attempts = handler_attempts + 1,
       handler_lease_owner = $4,
       handler_lease_expires_at = $5,
       handler_last_attempted_at = $6,
       handler_last_error = NULL
 WHERE job_id = $1
   AND handler_name = $2
   AND status = $3
   AND handler_attempts = $7
   AND (handler_lease_owner IS NULL OR handler_lease_expires_at <= $6)
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, handlerName, current.Status, owner, now.Add(leaseDuration), now, attempts))
	if errors.Is(err, pgx.ErrNoRows) {
		return false, tx.Commit(ctx)
	}
	if err != nil {
		return false, err
	}
	if current.Status == StatusQueued {
		if err := m.transactions.appendProgressIntentTx(ctx, tx, resource); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (m *Manager) ReleaseHandlerLease(ctx context.Context, jobID uuid.UUID, owner string) error {
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	_, err := m.pool.Exec(ctx, `
UPDATE jobs
   SET handler_lease_owner = NULL,
       handler_lease_expires_at = NULL
 WHERE job_id = $1
   AND handler_lease_owner = $2
`, jobID, owner)
	return err
}

func (m *Manager) RecordHandlerFailure(ctx context.Context, jobID uuid.UUID, owner string) error {
	return m.recordHandlerFailure(ctx, jobID, owner, HandlerExecutionFailed)
}

func (m *Manager) RecordHandlerIncomplete(ctx context.Context, jobID uuid.UUID, owner string) error {
	return m.recordHandlerFailure(ctx, jobID, owner, HandlerIncomplete)
}

func (m *Manager) recordHandlerFailure(ctx context.Context, jobID uuid.UUID, owner string, safeReason string) error {
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	if safeReason != HandlerExecutionFailed && safeReason != HandlerIncomplete {
		return ErrInvalidJobDefinition
	}
	now := m.now().UTC()
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return err
	}
	current, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		return err
	}
	if current.Status != StatusRunning && current.Status != StatusCancelRequested {
		return tx.Commit(ctx)
	}
	var attempts, maxAttempts int
	var leaseOwner *string
	if err := tx.QueryRow(ctx, `
SELECT handler_attempts, handler_max_attempts, handler_lease_owner
  FROM jobs
 WHERE job_id = $1
 FOR UPDATE
`, jobID).Scan(&attempts, &maxAttempts, &leaseOwner); err != nil {
		return err
	}
	if leaseOwner == nil || *leaseOwner != owner {
		return tx.Commit(ctx)
	}
	if attempts < maxAttempts {
		_, err := tx.Exec(ctx, `
UPDATE jobs
   SET handler_last_error = $3,
       handler_lease_owner = NULL,
       handler_lease_expires_at = NULL
 WHERE job_id = $1
   AND handler_lease_owner = $2
   AND status IN ('running', 'cancel_requested')
`, jobID, owner, safeReason)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	if _, err := tx.Exec(ctx, `UPDATE jobs SET handler_last_error = $2 WHERE job_id = $1`, jobID, HandlerAttemptsExhausted); err != nil {
		return err
	}
	mutation, err := transitionTerminalTx(ctx, tx, TransitionParams{
		JobID:    jobID,
		Progress: current.Progress,
		ErrorSummary: &ErrorSummary{
			Code:      HandlerAttemptsExhausted,
			Message:   "Job failed closed after exhausting durable handler attempts.",
			Retryable: false,
			Details:   map[string]any{"reason_code": HandlerAttemptsExhausted},
		},
	}, now, StatusFailed)
	if err != nil {
		return err
	}
	if err := m.transactions.appendProgressIntentTx(ctx, tx, mutation.resource); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (m *Manager) failExhaustedHandlerJobs(ctx context.Context, handlerName string) error {
	rows, err := m.pool.Query(ctx, `
SELECT job_id
  FROM jobs
 WHERE handler_name = $1
   AND status IN ('running', 'cancel_requested')
   AND handler_attempts >= handler_max_attempts
   AND (handler_lease_expires_at IS NULL OR handler_lease_expires_at <= $2)
 ORDER BY submitted_at, job_id
`, handlerName, m.now().UTC())
	if err != nil {
		return err
	}
	var jobIDs []uuid.UUID
	for rows.Next() {
		var jobID uuid.UUID
		if err := rows.Scan(&jobID); err != nil {
			return err
		}
		jobIDs = append(jobIDs, jobID)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()
	for _, jobID := range jobIDs {
		if err := m.failExpiredExhaustedJob(ctx, jobID); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) failExpiredExhaustedJob(ctx context.Context, jobID uuid.UUID) error {
	now := m.now().UTC()
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return err
	}
	current, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		return err
	}
	var attempts, maxAttempts int
	var leaseExpiresAt *time.Time
	if err := tx.QueryRow(ctx, `SELECT handler_attempts, handler_max_attempts, handler_lease_expires_at FROM jobs WHERE job_id = $1 FOR UPDATE`, jobID).Scan(&attempts, &maxAttempts, &leaseExpiresAt); err != nil {
		return err
	}
	if (current.Status != StatusRunning && current.Status != StatusCancelRequested) ||
		attempts < maxAttempts || (leaseExpiresAt != nil && leaseExpiresAt.After(now)) {
		return tx.Commit(ctx)
	}
	if _, err := tx.Exec(ctx, `UPDATE jobs SET handler_last_error = $2 WHERE job_id = $1`, jobID, HandlerAttemptsExhausted); err != nil {
		return err
	}
	mutation, err := transitionTerminalTx(ctx, tx, TransitionParams{
		JobID:    jobID,
		Progress: current.Progress,
		ErrorSummary: &ErrorSummary{
			Code: HandlerAttemptsExhausted, Message: "Job failed closed after exhausting durable handler attempts.",
			Retryable: false, Details: map[string]any{"reason_code": HandlerAttemptsExhausted},
		},
	}, now, StatusFailed)
	if err != nil {
		return err
	}
	if err := m.transactions.appendProgressIntentTx(ctx, tx, mutation.resource); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

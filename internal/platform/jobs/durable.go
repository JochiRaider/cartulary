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
	tag, err := m.pool.Exec(ctx, `
UPDATE jobs
   SET handler_attempts = handler_attempts + 1,
       handler_lease_owner = $3,
       handler_lease_expires_at = $4,
       handler_last_attempted_at = $5,
       handler_last_error = NULL,
       updated_at = $5
 WHERE job_id = $1
   AND handler_name = $2
   AND status IN ('queued', 'running', 'cancel_requested')
   AND handler_attempts < handler_max_attempts
   AND (handler_lease_expires_at IS NULL OR handler_lease_expires_at <= $5 OR handler_lease_owner = $3)
`, jobID, handlerName, owner, now.Add(leaseDuration), now)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
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

func (m *Manager) RecordHandlerError(ctx context.Context, jobID uuid.UUID, owner string, runErr error) error {
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	message := HandlerExecutionFailed
	if runErr != nil {
		message = runErr.Error()
	}
	failureJSON, err := handlerFailureJSON(HandlerAttemptsExhausted, message, true)
	if err != nil {
		return err
	}
	now := m.now().UTC()
	_, err = m.pool.Exec(ctx, `
UPDATE jobs
   SET handler_last_error = $3,
       handler_lease_owner = CASE WHEN handler_attempts >= handler_max_attempts THEN NULL ELSE handler_lease_owner END,
       handler_lease_expires_at = CASE WHEN handler_attempts >= handler_max_attempts THEN NULL ELSE handler_lease_expires_at END,
       status = CASE WHEN handler_attempts >= handler_max_attempts THEN 'failed' ELSE status END,
       cancelable = CASE WHEN handler_attempts >= handler_max_attempts THEN false ELSE cancelable END,
       updated_at = $4,
       finished_at = CASE WHEN handler_attempts >= handler_max_attempts THEN $4 ELSE finished_at END,
       retained_until = CASE WHEN handler_attempts >= handler_max_attempts THEN $5 ELSE retained_until END,
       result_summary_json = CASE WHEN handler_attempts >= handler_max_attempts THEN NULL ELSE result_summary_json END,
       error_summary_json = CASE WHEN handler_attempts >= handler_max_attempts THEN $6 ELSE error_summary_json END
 WHERE job_id = $1
   AND handler_lease_owner = $2
   AND status IN ('queued', 'running', 'cancel_requested')
`, jobID, owner, message, now, now.Add(7*24*time.Hour), failureJSON)
	return err
}

func (m *Manager) failExhaustedHandlerJobs(ctx context.Context, handlerName string) error {
	now := m.now().UTC()
	failureJSON, err := handlerFailureJSON(HandlerAttemptsExhausted, "Job failed closed after exhausting durable handler attempts.", false)
	if err != nil {
		return err
	}
	_, err = m.pool.Exec(ctx, `
UPDATE jobs
   SET status = 'failed',
       cancelable = false,
       updated_at = $2,
       finished_at = $2,
       retained_until = $3,
       result_summary_json = NULL,
       error_summary_json = $4,
       handler_lease_owner = NULL,
       handler_lease_expires_at = NULL,
       handler_last_error = $5
 WHERE handler_name = $1
   AND status IN ('queued', 'running', 'cancel_requested')
   AND handler_attempts >= handler_max_attempts
`, handlerName, now, now.Add(7*24*time.Hour), failureJSON, HandlerAttemptsExhausted)
	return err
}

func handlerFailureJSON(code string, message string, retryable bool) ([]byte, error) {
	_, failureJSON, err := marshalSummaries(nil, &ErrorSummary{
		Code:      code,
		Message:   message,
		Retryable: retryable,
		Details: map[string]any{
			"reason_code": code,
		},
	}, StatusFailed)
	return failureJSON, err
}

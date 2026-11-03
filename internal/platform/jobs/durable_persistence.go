package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	HandlerAbandoned          = "job_handler_abandoned"
)

func (m *Manager) HandlerPayload(ctx context.Context, execution Execution) (json.RawMessage, error) {
	if err := m.ensureConfigured(); err != nil {
		return nil, err
	}
	if !execution.valid() {
		return nil, ErrExecutionLost
	}
	var payload pgtype.Text
	err := m.db.QueryRow(ctx, `
SELECT handler_payload_json
  FROM jobs
 WHERE job_id = $1
   AND handler_attempt_id = $2
   AND handler_lease_expires_at > $3
   AND status IN ('running', 'cancel_requested')
`, execution.jobID, execution.attemptID, m.now().UTC()).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExecutionLost
	}
	if err != nil {
		return nil, err
	}
	if !payload.Valid {
		return nil, nil
	}
	return append(json.RawMessage(nil), payload.String...), nil
}

func (m *Manager) executionResource(ctx context.Context, execution Execution) (Resource, error) {
	if !execution.valid() {
		return Resource{}, ErrExecutionLost
	}
	stored, err := scanStoredJob(m.db.QueryRow(ctx, `
SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
       auth_policy,
       submitted_at, updated_at, progress_completed, progress_total, started_at,
       finished_at, retained_until, result_summary_json, error_summary_json, message,
       job_kind, progress_unit_id
  FROM jobs
 WHERE job_id = $1
   AND handler_attempt_id = $2
   AND handler_lease_expires_at > $3
   AND status IN ('running', 'cancel_requested')
`, execution.jobID, execution.attemptID, m.now().UTC()))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrExecutionLost
	}
	return stored.publicResource(), err
}

func (m *Manager) validateExecutionTx(ctx context.Context, tx pgx.Tx, execution Execution, now time.Time) error {
	return m.validateExecutionStatusTx(ctx, tx, execution, now, false)
}

func (m *Manager) validateCancellationExecutionTx(ctx context.Context, tx pgx.Tx, execution Execution, now time.Time) error {
	return m.validateExecutionStatusTx(ctx, tx, execution, now, true)
}

func (m *Manager) validateExecutionStatusTx(
	ctx context.Context,
	tx pgx.Tx,
	execution Execution,
	now time.Time,
	allowCancellationRequested bool,
) error {
	if tx == nil || !execution.valid() || now.IsZero() {
		return ErrExecutionLost
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
   AND handler_lease_expires_at > $3
   AND status IN ('running', 'cancel_requested')
 FOR UPDATE
`, execution.jobID, execution.attemptID, now.UTC()).Scan(&status)
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

func (m *Manager) RenewExecution(ctx context.Context, execution Execution) error {
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	if !execution.valid() {
		return ErrExecutionLost
	}
	now := m.now().UTC()
	result, err := m.db.Exec(ctx, `
UPDATE jobs
   SET handler_lease_expires_at = $3
 WHERE job_id = $1
   AND handler_attempt_id = $2
   AND handler_lease_expires_at > $4
   AND status IN ('running', 'cancel_requested')
`, execution.jobID, execution.attemptID, now.Add(m.policy.HandlerLease), now)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrExecutionLost
	}
	return nil
}

func (m *Manager) ReleaseExecution(ctx context.Context, execution Execution) error {
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	if !execution.valid() {
		return ErrExecutionLost
	}
	result, err := m.db.Exec(ctx, `
UPDATE jobs
   SET handler_attempt_id = NULL,
       handler_lease_expires_at = NULL
 WHERE job_id = $1
   AND handler_attempt_id = $2
   AND handler_lease_expires_at > $3
   AND status IN ('running', 'cancel_requested')
`, execution.jobID, execution.attemptID, m.now().UTC())
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrExecutionLost
	}
	return nil
}

type candidateReadiness func(context.Context, runnerCandidate) (bool, error)

type runnerCandidate struct {
	JobID       uuid.UUID
	JobKind     string
	HandlerName string
}

func (m *Manager) recoverableCandidatesForSelection(ctx context.Context, limit int, selection *RuntimeSelection, jobKinds []string, ready candidateReadiness) ([]runnerCandidate, error) {
	if err := m.ensureConfigured(); err != nil {
		return nil, err
	}
	if selection == nil || selection.catalog != m.catalog {
		return nil, ErrNotConfigured
	}
	for _, kind := range jobKinds {
		if !selection.containsJobKind(kind) {
			return nil, ErrInvalidJobDefinition
		}
	}
	if len(jobKinds) == 0 {
		return nil, nil
	}
	if limit <= 0 || limit > m.policy.RecoveryBatch {
		limit = m.policy.RecoveryBatch
	}
	if err := m.classifyExpiredAttempts(ctx, limit, jobKinds, ready); err != nil {
		return nil, err
	}
	var candidates []runnerCandidate
	var hasCursor bool
	var eligibleAt, submittedAt time.Time
	var afterID uuid.UUID
	observedAt := m.now().UTC()
	for {
		rows, err := m.db.Query(ctx, `
SELECT job_id,job_kind,COALESCE(handler_next_attempt_at,submitted_at),submitted_at
 FROM jobs
 WHERE status IN ('queued','running','cancel_requested')
 AND handler_failure_count < $1 AND handler_attempt_id IS NULL
 AND (handler_next_attempt_at IS NULL OR handler_next_attempt_at <= $2)
 AND job_kind=ANY($4::text[])
 AND (NOT $5::boolean OR (COALESCE(handler_next_attempt_at,submitted_at),submitted_at,job_id)>($6::timestamptz,$7::timestamptz,$8::uuid))
 ORDER BY COALESCE(handler_next_attempt_at,submitted_at),submitted_at,job_id
 LIMIT $3`, m.policy.MaximumFailures, observedAt, limit, jobKinds, hasCursor, eligibleAt, submittedAt, afterID)
		if err != nil {
			return nil, err
		}
		var page []runnerCandidate
		for rows.Next() {
			var candidate runnerCandidate
			if err := rows.Scan(&candidate.JobID, &candidate.JobKind, &eligibleAt, &submittedAt); err != nil {
				rows.Close()
				return nil, err
			}
			definition, present := m.catalog.definition(candidate.JobKind)
			workerKind, assigned := selection.workerKindForJob(candidate.JobKind)
			if !present || !assigned || definition.HandlerName != workerKind {
				rows.Close()
				return nil, ErrInvalidJobDefinition
			}
			candidate.HandlerName = definition.HandlerName
			page = append(page, candidate)
			afterID = candidate.JobID
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		// Release the cursor before asking an owner to inspect its immutable input.
		for _, candidate := range page {
			if ready != nil {
				allowed, err := ready(ctx, candidate)
				if err != nil {
					return nil, err
				}
				if !allowed {
					continue
				}
			}
			candidates = append(candidates, candidate)
			if len(candidates) == limit {
				return candidates, nil
			}
		}
		if len(page) < limit {
			return candidates, nil
		}
		hasCursor = true
	}
}

func (m *Manager) runnerCandidateForSelection(ctx context.Context, jobID uuid.UUID, selection *RuntimeSelection) (runnerCandidate, bool, error) {
	if err := m.ensureConfigured(); err != nil {
		return runnerCandidate{}, false, err
	}
	if selection == nil || selection.catalog != m.catalog || jobID == uuid.Nil {
		return runnerCandidate{}, false, ErrNotConfigured
	}
	var jobKind string
	err := m.db.QueryRow(ctx, `SELECT job_kind FROM jobs WHERE job_id = $1`, jobID).Scan(&jobKind)
	if errors.Is(err, pgx.ErrNoRows) {
		return runnerCandidate{}, false, nil
	}
	if err != nil {
		return runnerCandidate{}, false, err
	}
	definition, present := m.catalog.definition(jobKind)
	workerKind, assigned := selection.workerKindForJob(jobKind)
	if !present || !assigned || definition.HandlerName != workerKind {
		return runnerCandidate{}, false, nil
	}
	return runnerCandidate{JobID: jobID, JobKind: jobKind, HandlerName: definition.HandlerName}, true, nil
}

func (m *Manager) Claim(ctx context.Context, jobID uuid.UUID) (Execution, bool, error) {
	execution, _, _, claimed, err := m.claimForRunnerSelection(ctx, jobID, m.transactions.selection)
	return execution, claimed, err
}

func (m *Manager) claimForRunnerSelection(ctx context.Context, jobID uuid.UUID, selection *RuntimeSelection) (Execution, string, string, bool, error) {
	if err := m.ensureConfigured(); err != nil {
		return Execution{}, "", "", false, err
	}
	var jobKind string
	err := m.db.QueryRow(ctx, `SELECT job_kind FROM jobs WHERE job_id = $1`, jobID).Scan(&jobKind)
	if errors.Is(err, pgx.ErrNoRows) {
		return Execution{}, "", "", false, nil
	}
	if err != nil {
		return Execution{}, "", "", false, err
	}
	definition, present := m.catalog.definition(jobKind)
	if !present {
		return Execution{}, "", "", false, ErrInvalidJobDefinition
	}
	if selection != nil {
		if selection.catalog != m.catalog {
			return Execution{}, "", "", false, ErrNotConfigured
		}
		if !selection.containsJobKind(jobKind) {
			return Execution{}, "", "", false, nil
		}
	}
	execution, claimed, err := m.claimExecution(ctx, jobID, definition.JobKind, definition.HandlerName)
	return execution, definition.HandlerName, definition.JobKind, claimed, err
}

func (m *Manager) claimExecution(ctx context.Context, jobID uuid.UUID, jobKind string, handlerName string) (Execution, bool, error) {
	if err := m.ensureConfigured(); err != nil {
		return Execution{}, false, err
	}
	if jobID == uuid.Nil || jobKind == "" || handlerName == "" {
		return Execution{}, false, ErrInvalidJobDefinition
	}
	now := m.now().UTC()
	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Execution{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockTransitionTx(ctx, tx, jobID); err != nil {
		return Execution{}, false, err
	}
	current, err := getJobTx(ctx, tx, jobID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Execution{}, false, tx.Commit(ctx)
		}
		return Execution{}, false, err
	}
	if current.Status != StatusQueued && current.Status != StatusRunning && current.Status != StatusCancelRequested {
		return Execution{}, false, tx.Commit(ctx)
	}
	var storedHandlerName string
	var failureCount int
	var attemptID *uuid.UUID
	var leaseExpiresAt, nextAttemptAt *time.Time
	if err := tx.QueryRow(ctx, `
SELECT handler_name, handler_failure_count, handler_attempt_id,
       handler_lease_expires_at, handler_next_attempt_at
  FROM jobs
 WHERE job_id = $1
 FOR UPDATE
`, jobID).Scan(&storedHandlerName, &failureCount, &attemptID, &leaseExpiresAt, &nextAttemptAt); err != nil {
		return Execution{}, false, err
	}
	if storedHandlerName != handlerName {
		return Execution{}, false, tx.Commit(ctx)
	}
	if attemptID != nil {
		if leaseExpiresAt == nil || leaseExpiresAt.After(now) {
			return Execution{}, false, tx.Commit(ctx)
		}
		terminal, err := m.consumeFailureTx(ctx, tx, current, failureCount, HandlerAbandoned, now)
		if err != nil {
			return Execution{}, false, err
		}
		if terminal {
			return Execution{}, false, tx.Commit(ctx)
		}
		return Execution{}, false, tx.Commit(ctx)
	}
	if failureCount >= m.policy.MaximumFailures || (nextAttemptAt != nil && nextAttemptAt.After(now)) {
		return Execution{}, false, tx.Commit(ctx)
	}
	eligibleAt := current.SubmittedAt.UTC()
	if nextAttemptAt != nil {
		eligibleAt = nextAttemptAt.UTC()
	}
	if eligibleAt.After(now) {
		return Execution{}, false, fmt.Errorf("%w: job queue eligibility is after claim time", ErrInvalidTransition)
	}
	newAttemptID := uuid.New()
	resource, err := scanJob(tx.QueryRow(ctx, `
UPDATE jobs
   SET status = CASE WHEN status = 'queued' THEN 'running' ELSE status END,
       started_at = CASE WHEN status = 'queued' THEN COALESCE(started_at, $4) ELSE started_at END,
       updated_at = CASE WHEN status = 'queued' THEN $4 ELSE updated_at END,
       handler_attempt_id = $3,
       handler_lease_expires_at = $5,
       handler_next_attempt_at = NULL,
       handler_last_attempted_at = $4
 WHERE job_id = $1
   AND handler_name = $2
   AND status IN ('queued', 'running', 'cancel_requested')
   AND handler_attempt_id IS NULL
   AND handler_failure_count = $6
   AND (handler_next_attempt_at IS NULL OR handler_next_attempt_at <= $4)
RETURNING job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
          auth_policy,
          submitted_at, updated_at, progress_completed, progress_total, started_at,
          finished_at, retained_until, result_summary_json, error_summary_json, message
`, jobID, handlerName, newAttemptID, now, now.Add(m.policy.HandlerLease), failureCount))
	if errors.Is(err, pgx.ErrNoRows) {
		return Execution{}, false, tx.Commit(ctx)
	}
	if err != nil {
		return Execution{}, false, err
	}
	if current.Status == StatusQueued {
		if err := m.transactions.appendProgressIntentTx(ctx, tx, resource); err != nil {
			return Execution{}, false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Execution{}, false, err
	}
	m.recordQueueWait(ctx, jobKind, now.Sub(eligibleAt))
	return newExecution(jobID, newAttemptID), true, nil
}

func (m *Manager) RecordExecutionFailure(ctx context.Context, execution Execution, incomplete bool) error {
	reason := HandlerExecutionFailed
	if incomplete {
		reason = HandlerIncomplete
	}
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	if !execution.valid() {
		return ErrExecutionLost
	}
	now := m.now().UTC()
	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := m.validateExecutionTx(ctx, tx, execution, now); err != nil {
		return err
	}
	current, err := getJobTx(ctx, tx, execution.jobID)
	if err != nil {
		return err
	}
	var failureCount int
	if err := tx.QueryRow(ctx, `SELECT handler_failure_count FROM jobs WHERE job_id = $1 FOR UPDATE`, execution.jobID).Scan(&failureCount); err != nil {
		return err
	}
	if _, err := m.consumeFailureTx(ctx, tx, current, failureCount, reason, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (m *Manager) consumeFailureTx(
	ctx context.Context,
	tx pgx.Tx,
	current Resource,
	failureCount int,
	reason string,
	now time.Time,
) (bool, error) {
	nextFailureCount := failureCount + 1
	if nextFailureCount < m.policy.MaximumFailures {
		delay := m.policy.RetryDelays[nextFailureCount-1]
		_, err := tx.Exec(ctx, `
UPDATE jobs
   SET handler_failure_count = $2,
       handler_attempt_id = NULL,
       handler_lease_expires_at = NULL,
       handler_next_attempt_at = $3,
       handler_last_error = $4
 WHERE job_id = $1
   AND status IN ('running', 'cancel_requested')
`, uuid.MustParse(current.JobID), nextFailureCount, now.Add(delay), reason)
		return false, err
	}
	jobID := uuid.MustParse(current.JobID)
	if _, err := tx.Exec(ctx, `
UPDATE jobs
   SET handler_failure_count = $2,
       handler_attempt_id = NULL,
       handler_lease_expires_at = NULL,
       handler_next_attempt_at = NULL,
       handler_last_error = $3
 WHERE job_id = $1
`, jobID, m.policy.MaximumFailures, HandlerAttemptsExhausted); err != nil {
		return false, err
	}
	summary := ErrorSummary{
		Code:      HandlerAttemptsExhausted,
		Message:   "Job failed closed after exhausting durable handler attempts.",
		Retryable: false,
		Details:   map[string]any{"reason_code": HandlerAttemptsExhausted},
	}
	mutation, err := transitionTerminalTx(ctx, tx, terminalTransition{
		JobID: jobID, Progress: current.Progress, ErrorSummary: &summary,
	}, now, StatusFailed)
	if err != nil {
		return false, err
	}
	if err := m.transactions.appendProgressIntentTx(ctx, tx, mutation.resource); err != nil {
		return false, err
	}
	return true, nil
}

func (m *Manager) classifyExpiredAttempts(ctx context.Context, limit int, jobKinds []string, ready candidateReadiness) error {
	var hasCursor bool
	var expiresAt time.Time
	var afterID uuid.UUID
	observedAt := m.now().UTC()
	processed := 0
	for processed < limit {
		rows, err := m.db.Query(ctx, `
SELECT job_id,job_kind,handler_lease_expires_at
 FROM jobs
 WHERE status IN ('running','cancel_requested')
 AND handler_attempt_id IS NOT NULL AND handler_lease_expires_at <= $1
 AND job_kind=ANY($3::text[])
 AND (NOT $4::boolean OR (handler_lease_expires_at,job_id)>($5::timestamptz,$6::uuid))
 ORDER BY handler_lease_expires_at,job_id LIMIT $2`, observedAt, limit, jobKinds, hasCursor, expiresAt, afterID)
		if err != nil {
			return err
		}
		var page []runnerCandidate
		for rows.Next() {
			var candidate runnerCandidate
			if err := rows.Scan(&candidate.JobID, &candidate.JobKind, &expiresAt); err != nil {
				rows.Close()
				return err
			}
			definition, present := m.catalog.definition(candidate.JobKind)
			if !present {
				rows.Close()
				return ErrInvalidJobDefinition
			}
			candidate.HandlerName = definition.HandlerName
			page = append(page, candidate)
			afterID = candidate.JobID
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, candidate := range page {
			if ready != nil {
				allowed, err := ready(ctx, candidate)
				if err != nil {
					return err
				}
				if !allowed {
					continue
				}
			}
			if err := m.classifyExpiredAttempt(ctx, candidate.JobID); err != nil {
				return err
			}
			processed++
			if processed == limit {
				return nil
			}
		}
		if len(page) < limit {
			return nil
		}
		hasCursor = true
	}
	return nil
}

func (m *Manager) classifyExpiredAttempt(ctx context.Context, jobID uuid.UUID) error {
	now := m.now().UTC()
	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
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
	var failureCount int
	var attemptID *uuid.UUID
	var expiresAt *time.Time
	if err := tx.QueryRow(ctx, `
SELECT handler_failure_count, handler_attempt_id, handler_lease_expires_at
  FROM jobs WHERE job_id = $1 FOR UPDATE
`, jobID).Scan(&failureCount, &attemptID, &expiresAt); err != nil {
		return err
	}
	if attemptID == nil || expiresAt == nil || expiresAt.After(now) ||
		(current.Status != StatusRunning && current.Status != StatusCancelRequested) {
		return tx.Commit(ctx)
	}
	if _, err := m.consumeFailureTx(ctx, tx, current, failureCount, HandlerAbandoned, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

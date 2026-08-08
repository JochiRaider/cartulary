package jobs

import (
	"context"
)

const expirySweepLockKey int64 = 49006060

// compactExpiredJobs converts at most one ordered batch of logically expired
// terminal jobs into private tombstones. Durable owner outputs, proofs,
// identity, provenance, lifecycle timestamps, and progress remain untouched.
func (m *Manager) compactExpiredJobs(ctx context.Context, limit int) (int64, error) {
	if m == nil || m.pool == nil || m.now == nil {
		return 0, ErrNotConfigured
	}
	if limit <= 0 || limit > m.policy.ExpiryBatch {
		limit = m.policy.ExpiryBatch
	}
	if limit <= 0 {
		return 0, ErrNotConfigured
	}
	cutoff := m.now().UTC()
	if cutoff.IsZero() {
		return 0, ErrNotConfigured
	}
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var acquired bool
	if err := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, expirySweepLockKey).Scan(&acquired); err != nil {
		return 0, err
	}
	if !acquired {
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
		return 0, nil
	}
	rows, err := tx.Query(ctx, `
WITH candidates AS (
    SELECT job_id
      FROM jobs
     WHERE retained_until <= $1
       AND expired_at IS NULL
     ORDER BY retained_until, job_id
     FOR UPDATE SKIP LOCKED
     LIMIT $2
)
UPDATE jobs AS job
   SET expired_at = $1,
       handler_payload_json = NULL,
       handler_attempt_id = NULL,
       handler_lease_expires_at = NULL,
       handler_failure_count = 0,
       handler_next_attempt_at = NULL,
       handler_last_attempted_at = NULL,
       handler_last_error = NULL,
       message = NULL,
       result_summary_json = NULL,
       error_summary_json = NULL,
       extension_idempotency_identity = NULL,
       extension_idempotency_route_key = NULL,
       extension_idempotency_scope_key = NULL,
       extension_normalized_request_sha256 = NULL
  FROM candidates
 WHERE job.job_id = candidates.job_id
RETURNING job.job_kind
`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	counts := make(map[string]int64)
	var compacted int64
	for rows.Next() {
		var jobKind string
		if err := rows.Scan(&jobKind); err != nil {
			rows.Close()
			return 0, err
		}
		counts[jobKind]++
		compacted++
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	m.recordExpiredJobs(ctx, counts)
	return compacted, nil
}

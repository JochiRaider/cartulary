package jobs

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ReconcileRestoredNonterminalTx releases an execution attempt copied from a
// quiesced backup so the restored Common Job can be claimed by the target
// runtime. The durable job status, retry count, and retry schedule remain
// authoritative and are not rewritten by an owner-specific restore adapter.
func ReconcileRestoredNonterminalTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, expectedJobKind string) error {
	if ctx == nil || tx == nil || jobID == uuid.Nil || expectedJobKind == "" {
		return ErrInvalidJobDefinition
	}
	var status, jobKind string
	var failureCount int
	err := tx.QueryRow(ctx, `
SELECT status, job_kind, handler_failure_count
  FROM jobs
 WHERE job_id = $1
 FOR UPDATE
`, jobID).Scan(&status, &jobKind, &failureCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: restored job is missing", ErrInvalidJobDefinition)
	}
	if err != nil {
		return fmt.Errorf("lock restored Common Job: %w", err)
	}
	if jobKind != expectedJobKind ||
		(status != string(StatusQueued) && status != string(StatusRunning) && status != string(StatusCancelRequested)) ||
		failureCount < 0 || failureCount >= DefaultHandlerMaxAttempts {
		return fmt.Errorf("%w: restored nonterminal job state is not resumable", ErrInvalidJobDefinition)
	}
	if _, err := tx.Exec(ctx, `
UPDATE jobs
   SET handler_attempt_id = NULL,
       handler_lease_expires_at = NULL
 WHERE job_id = $1
`, jobID); err != nil {
		return fmt.Errorf("release restored Common Job execution attempt: %w", err)
	}
	return nil
}

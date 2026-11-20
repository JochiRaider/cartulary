package evidence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const (
	cleanupBatchSize         = 100
	cleanupClaimLease        = 5 * time.Minute
	cleanupDeleteTimeout     = time.Minute
	cleanupMetadataRetention = 7 * 24 * time.Hour
	cleanupOverdueThreshold  = 15 * time.Minute
)

// cleanupObjectDeleter is the only object-store capability used by the
// durable cleanup engine. The production adapter binds the governed purpose.
type cleanupObjectDeleter interface {
	DeleteObject(context.Context, string) error
}

type cleanupSweepResult struct {
	ExpiredPendingCount  int
	ClaimedBlobCount     int
	CleanedBlobCount     int
	RetryScheduledCount  int
	DeletedMetadataCount int
	HasMore              bool
	HealthSnapshotValid  bool
	OverdueBlobCount     int64
	OldestEligibleAge    time.Duration
	// InvariantViolationCount counts Blobs this sweep loaded with a
	// REQ-02-189a tuple violation (timeout, claim or completion). Each was left
	// unchanged while the sweep continued; a positive count reports the sweep
	// as failed with the invariant_violation error class.
	InvariantViolationCount int
}

type cleanupClaim struct {
	ObjectBlobID uuid.UUID
	StorageKey   string
	ClaimToken   uuid.UUID
	AttemptCount int
}

// SweepFailedUnattachedBlobs performs one bounded, restart-safe cleanup sweep.
// Claim acquisition and completion are transactional; object deletion is
// deliberately outside a database transaction.
func (s *cleanupService) SweepFailedUnattachedBlobs(
	ctx context.Context,
	deleter cleanupObjectDeleter,
	now time.Time,
) (cleanupSweepResult, error) {
	if deleter == nil {
		return cleanupSweepResult{}, errors.New("sweep Evidence cleanup: object deleter is required")
	}
	now = now.UTC()
	expired, expiryViolations, err := s.blobLifecycle.markExpiredPending(ctx, now)
	if err != nil {
		return cleanupSweepResult{}, fmt.Errorf("expire pending Evidence blobs: %w", err)
	}
	claims, claimViolations, err := s.claimCleanupBatch(ctx, now)
	if err != nil {
		return cleanupSweepResult{}, err
	}
	result := cleanupSweepResult{
		ExpiredPendingCount:     expired,
		ClaimedBlobCount:        len(claims),
		HasMore:                 len(claims) == cleanupBatchSize,
		InvariantViolationCount: expiryViolations + claimViolations,
	}
	for _, claim := range claims {
		deleteCtx, cancel := context.WithTimeout(ctx, cleanupDeleteTimeout)
		deleteErr := deleter.DeleteObject(deleteCtx, claim.StorageKey)
		deleteTimedOut := errors.Is(deleteCtx.Err(), context.DeadlineExceeded)
		cancel()
		if deleteErr != nil && !objectstore.IsObjectNotFound(deleteErr) {
			failureClass := "delete_failed"
			if deleteTimedOut || errors.Is(deleteErr, context.DeadlineExceeded) {
				failureClass = "delete_timeout"
			}
			if err := s.scheduleCleanupRetry(ctx, claim, failureClass, now); err != nil {
				return result, err
			}
			result.RetryScheduledCount++
			continue
		}
		completed, err := s.completeCleanupClaim(ctx, claim, now)
		if errors.Is(err, errBlobTupleViolation) {
			// REQ-02-189a: completion committed nothing for this Blob. Its
			// claim stays as written, exactly as a re-claim leaves a violating
			// candidate, and the sweep continues with the other Blobs.
			result.InvariantViolationCount++
			continue
		}
		if err != nil {
			return result, err
		}
		if completed {
			result.CleanedBlobCount++
			continue
		}
		if err := s.scheduleCleanupRetry(ctx, claim, "state_changed", now); err != nil {
			return result, err
		}
		result.RetryScheduledCount++
	}
	deleted, err := s.deleteExpiredCleanupMetadata(ctx, now)
	if err != nil {
		return result, err
	}
	result.DeletedMetadataCount = deleted
	overdue, oldest, err := s.cleanupHealthSnapshot(ctx, now)
	if err != nil {
		return result, err
	}
	result.HealthSnapshotValid = true
	result.OverdueBlobCount = overdue
	result.OldestEligibleAge = oldest
	return result, nil
}

func (s *cleanupService) cleanupHealthSnapshot(ctx context.Context, now time.Time) (int64, time.Duration, error) {
	var overdue int64
	var oldestSeconds float64
	if err := s.pool.QueryRow(ctx, `
SELECT COUNT(*) FILTER (
           WHERE b.cleanup_due_at <= $2
       ),
       COALESCE(MAX(EXTRACT(EPOCH FROM ($1 - b.cleanup_due_at))), 0)
  FROM object_blobs b
 WHERE b.upload_state = 'failed'
   AND b.cleaned_up_at IS NULL
   AND b.cleanup_due_at <= $1
   AND NOT EXISTS (
       SELECT 1
         FROM evidence e
        WHERE e.object_blob_id = b.object_blob_id
   )
`, now, now.Add(-cleanupOverdueThreshold)).Scan(&overdue, &oldestSeconds); err != nil {
		return 0, 0, fmt.Errorf("load Evidence cleanup health: %w", err)
	}
	if oldestSeconds < 0 {
		oldestSeconds = 0
	}
	return overdue, time.Duration(oldestSeconds * float64(time.Second)), nil
}

// claimCleanupBatch also returns the number of violating candidates it left
// unclaimed in this sweep.
func (s *cleanupService) claimCleanupBatch(ctx context.Context, now time.Time) ([]cleanupClaim, int, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("begin Evidence cleanup claim transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// REQ-02-189a: cleanup claims only Blobs whose lifecycle tuple is coherent.
	// A violating Blob is left unclaimed and unchanged for operator recovery,
	// counted as an invariant violation of this sweep, and stays visible as
	// overdue in the cleanup health snapshot. The scan continues past
	// violating rows in key order, so they never consume the batch.
	candidates := make([]cleanupClaim, 0, cleanupBatchSize)
	violations := 0
	var after *evidenceblob.CleanupKey
	for len(candidates) < cleanupBatchSize {
		page, last, err := evidenceblob.LockCleanupCandidatesTx(ctx, tx, now, after, cleanupBatchSize-len(candidates))
		if err != nil {
			return nil, 0, err
		}
		for _, candidate := range page {
			if !candidate.Coherent {
				violations++
				continue
			}
			candidates = append(candidates, cleanupClaim{
				ObjectBlobID: candidate.ObjectBlobID,
				StorageKey:   candidate.StorageKey,
				ClaimToken:   uuid.New(),
				AttemptCount: candidate.AttemptCount + 1,
			})
		}
		if last == nil {
			break
		}
		after = last
	}
	for _, candidate := range candidates {
		tag, err := tx.Exec(ctx, `
INSERT INTO evidence_blob_cleanup_claims (
    object_blob_id, claim_token, claim_state, attempt_count,
    claimed_at, claim_expires_at, created_at, updated_at
) VALUES ($1, $2, 'claimed', $3, $4, $5, $4, $4)
ON CONFLICT (object_blob_id) DO UPDATE
   SET claim_token = EXCLUDED.claim_token,
       claim_state = 'claimed',
       attempt_count = EXCLUDED.attempt_count,
       claimed_at = EXCLUDED.claimed_at,
       claim_expires_at = EXCLUDED.claim_expires_at,
       next_attempt_at = NULL,
       last_attempt_at = NULL,
       completed_at = NULL,
       last_failure_class = NULL,
       updated_at = EXCLUDED.updated_at
 WHERE (evidence_blob_cleanup_claims.claim_state = 'retry_wait'
            AND evidence_blob_cleanup_claims.next_attempt_at <= EXCLUDED.claimed_at)
    OR (evidence_blob_cleanup_claims.claim_state = 'claimed'
            AND evidence_blob_cleanup_claims.claim_expires_at <= EXCLUDED.claimed_at)
`, candidate.ObjectBlobID, candidate.ClaimToken, candidate.AttemptCount, now, now.Add(cleanupClaimLease))
		if err != nil {
			return nil, 0, fmt.Errorf("acquire Evidence cleanup claim: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return nil, 0, errors.New("acquire Evidence cleanup claim: candidate became unavailable")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("commit Evidence cleanup claims: %w", err)
	}
	return candidates, violations, nil
}

func (s *cleanupService) scheduleCleanupRetry(
	ctx context.Context,
	claim cleanupClaim,
	failureClass string,
	now time.Time,
) error {
	delay := cleanupRetryDelay(claim.AttemptCount)
	tag, err := s.pool.Exec(ctx, `
UPDATE evidence_blob_cleanup_claims
   SET claim_state = 'retry_wait',
       claim_expires_at = NULL,
       next_attempt_at = $3,
       last_attempt_at = $2,
       last_failure_class = $4,
       updated_at = $2
 WHERE object_blob_id = $1
   AND claim_token = $5
   AND claim_state = 'claimed'
`, claim.ObjectBlobID, now, now.Add(delay), failureClass, claim.ClaimToken)
	if err != nil {
		return fmt.Errorf("schedule Evidence cleanup retry: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("schedule Evidence cleanup retry: claim ownership lost")
	}
	return nil
}

func cleanupRetryDelay(attempt int) time.Duration {
	switch attempt {
	case 1:
		return time.Minute
	case 2:
		return 5 * time.Minute
	default:
		return 15 * time.Minute
	}
}

func (s *cleanupService) completeCleanupClaim(
	ctx context.Context,
	claim cleanupClaim,
	now time.Time,
) (bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin Evidence cleanup completion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// REQ-02-189a: completion verifies the tuple under the Blob lock. A
	// violation returns errBlobTupleViolation and commits nothing; the sweep
	// counts it and continues, leaving repair to operator recovery.
	locked, err := evidenceblob.LockTx(ctx, tx, claim.ObjectBlobID, evidenceblob.UpdateWait)
	if errors.Is(err, evidenceblob.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("lock Evidence cleanup blob: %w", err)
	}
	var associated bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM evidence e
     WHERE e.object_blob_id = $1
)
`, claim.ObjectBlobID).Scan(&associated); err != nil {
		return false, fmt.Errorf("check Evidence cleanup blob association: %w", err)
	}
	eligible := locked.UploadState == evidencepolicy.BlobFailed && locked.CleanedUpAt == nil && !associated
	var owned bool
	if err := tx.QueryRow(ctx, `
SELECT claim_token = $2 AND claim_state = 'claimed'
  FROM evidence_blob_cleanup_claims
 WHERE object_blob_id = $1
 FOR UPDATE
`, claim.ObjectBlobID, claim.ClaimToken).Scan(&owned); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lock Evidence cleanup claim: %w", err)
	}
	if !eligible || !owned {
		return false, nil
	}
	if _, err := tx.Exec(ctx, `
UPDATE object_blobs
   SET cleaned_up_at = $2,
       updated_at = $2
 WHERE object_blob_id = $1
`, claim.ObjectBlobID, now); err != nil {
		return false, fmt.Errorf("complete Evidence blob cleanup: %w", err)
	}
	if _, err := tx.Exec(ctx, `
UPDATE evidence_blob_cleanup_claims
   SET claim_state = 'completed',
       claim_expires_at = NULL,
       next_attempt_at = NULL,
       last_attempt_at = $2,
       completed_at = $2,
       last_failure_class = NULL,
       updated_at = $2
 WHERE object_blob_id = $1
   AND claim_token = $3
`, claim.ObjectBlobID, now, claim.ClaimToken); err != nil {
		return false, fmt.Errorf("complete Evidence cleanup claim: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit Evidence cleanup completion: %w", err)
	}
	return true, nil
}

func (s *cleanupService) deleteExpiredCleanupMetadata(ctx context.Context, now time.Time) (int, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin Evidence cleanup metadata retention: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	ids, err := evidenceblob.LockExpiredCleanupMetadataTx(ctx, tx, now.Add(-cleanupMetadataRetention), cleanupBatchSize)
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		if _, err := tx.Exec(ctx, `DELETE FROM object_blobs WHERE object_blob_id = $1`, id); err != nil {
			return 0, fmt.Errorf("delete expired Evidence cleanup metadata: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit Evidence cleanup metadata retention: %w", err)
	}
	return len(ids), nil
}

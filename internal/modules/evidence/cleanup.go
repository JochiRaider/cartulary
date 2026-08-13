package evidence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const (
	cleanupBatchSize         = 100
	cleanupClaimLease        = 5 * time.Minute
	cleanupDeleteTimeout     = time.Minute
	cleanupMetadataRetention = 7 * 24 * time.Hour
	cleanupOverdueThreshold  = 15 * time.Minute
)

// CleanupObjectDeleter is the only object-store capability used by the
// durable cleanup engine. S09 supplies the governed typed-purpose adapter.
type CleanupObjectDeleter interface {
	DeleteObject(context.Context, string) error
}

type CleanupSweepResult struct {
	ExpiredPendingCount  int
	ClaimedBlobCount     int
	CleanedBlobCount     int
	RetryScheduledCount  int
	DeletedMetadataCount int
	HasMore              bool
	HealthSnapshotValid  bool
	OverdueBlobCount     int64
	OldestEligibleAge    time.Duration
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
func (s *CleanupService) SweepFailedUnattachedBlobs(
	ctx context.Context,
	deleter CleanupObjectDeleter,
	now time.Time,
) (CleanupSweepResult, error) {
	if deleter == nil {
		return CleanupSweepResult{}, errors.New("sweep Evidence cleanup: object deleter is required")
	}
	now = now.UTC()
	expired, err := s.blobLifecycle.markExpiredPending(ctx, now)
	if err != nil {
		return CleanupSweepResult{}, fmt.Errorf("expire pending Evidence blobs: %w", err)
	}
	claims, err := s.claimCleanupBatch(ctx, now)
	if err != nil {
		return CleanupSweepResult{}, err
	}
	result := CleanupSweepResult{
		ExpiredPendingCount: expired,
		ClaimedBlobCount:    len(claims),
		HasMore:             len(claims) == cleanupBatchSize,
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

func (s *CleanupService) cleanupHealthSnapshot(ctx context.Context, now time.Time) (int64, time.Duration, error) {
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

func (s *CleanupService) claimCleanupBatch(ctx context.Context, now time.Time) ([]cleanupClaim, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin Evidence cleanup claim transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `
SELECT b.object_blob_id, b.storage_key, COALESCE(c.attempt_count, 0)
  FROM object_blobs b
  LEFT JOIN evidence_blob_cleanup_claims c
    ON c.object_blob_id = b.object_blob_id
 WHERE b.upload_state = 'failed'
   AND b.cleaned_up_at IS NULL
   AND b.cleanup_due_at <= $1
   AND NOT EXISTS (
       SELECT 1
         FROM evidence e
        WHERE e.object_blob_id = b.object_blob_id
   )
   AND (
       c.object_blob_id IS NULL
       OR (c.claim_state = 'retry_wait' AND c.next_attempt_at <= $1)
       OR (c.claim_state = 'claimed' AND c.claim_expires_at <= $1)
   )
 ORDER BY b.cleanup_due_at, b.object_blob_id
 FOR UPDATE OF b SKIP LOCKED
 LIMIT $2
`, now, cleanupBatchSize)
	if err != nil {
		return nil, fmt.Errorf("select Evidence cleanup claims: %w", err)
	}
	candidates := make([]cleanupClaim, 0, cleanupBatchSize)
	for rows.Next() {
		var candidate cleanupClaim
		if err := rows.Scan(&candidate.ObjectBlobID, &candidate.StorageKey, &candidate.AttemptCount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan Evidence cleanup claim candidate: %w", err)
		}
		candidate.AttemptCount++
		candidate.ClaimToken = uuid.New()
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate Evidence cleanup claim candidates: %w", err)
	}
	rows.Close()
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
			return nil, fmt.Errorf("acquire Evidence cleanup claim: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return nil, errors.New("acquire Evidence cleanup claim: candidate became unavailable")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit Evidence cleanup claims: %w", err)
	}
	return candidates, nil
}

func (s *CleanupService) scheduleCleanupRetry(
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

func (s *CleanupService) completeCleanupClaim(
	ctx context.Context,
	claim cleanupClaim,
	now time.Time,
) (bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin Evidence cleanup completion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var eligible bool
	if err := tx.QueryRow(ctx, `
SELECT b.upload_state = 'failed'
       AND b.cleaned_up_at IS NULL
       AND NOT EXISTS (
           SELECT 1
             FROM evidence e
            WHERE e.object_blob_id = b.object_blob_id
       )
  FROM object_blobs b
 WHERE b.object_blob_id = $1
 FOR UPDATE
`, claim.ObjectBlobID).Scan(&eligible); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lock Evidence cleanup blob: %w", err)
	}
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

func (s *CleanupService) deleteExpiredCleanupMetadata(ctx context.Context, now time.Time) (int, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin Evidence cleanup metadata retention: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `
SELECT b.object_blob_id
  FROM object_blobs b
  JOIN evidence_blob_cleanup_claims c
    ON c.object_blob_id = b.object_blob_id
 WHERE b.upload_state = 'failed'
   AND b.cleaned_up_at <= $1
   AND c.claim_state = 'completed'
   AND c.completed_at <= $1
   AND NOT EXISTS (
       SELECT 1
         FROM evidence e
        WHERE e.object_blob_id = b.object_blob_id
   )
 ORDER BY b.cleaned_up_at, b.object_blob_id
 FOR UPDATE OF b SKIP LOCKED
 LIMIT $2
`, now.Add(-cleanupMetadataRetention), cleanupBatchSize)
	if err != nil {
		return 0, fmt.Errorf("select expired Evidence cleanup metadata: %w", err)
	}
	ids := make([]uuid.UUID, 0, cleanupBatchSize)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan expired Evidence cleanup metadata: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate expired Evidence cleanup metadata: %w", err)
	}
	rows.Close()
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

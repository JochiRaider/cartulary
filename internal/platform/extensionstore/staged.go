package extensionstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type StagedState string
type DeleteState string

const (
	StagedAllocated StagedState = "allocated"
	StagedReady     StagedState = "ready"
	StagedPublished StagedState = "published"
	StagedAbandoned StagedState = "abandoned"

	DeleteNotApplicable DeleteState = "not_applicable"
	DeletePending       DeleteState = "pending"
	DeleteDeleted       DeleteState = "deleted"
)

type StagedObject struct {
	StagingID          string
	OwnerProfileID     string
	StorageIdentity    string
	ExpectedByteSize   int64
	ExpectedSHA256     string
	StagedAt           time.Time
	StagingExpiresAt   time.Time
	ReadyAt            *time.Time
	PublishedAt        *time.Time
	AbandonedAt        *time.Time
	State              StagedState
	DeleteState        DeleteState
	DeleteAttemptCount int32
	NextDeleteAttempt  *time.Time
	LastDeleteError    *string
}

func NewStagedObject(stagingID, profileID, storageIdentity string, byteSize int64, digest string, now time.Time) StagedObject {
	stagedAt := now.UTC()
	return StagedObject{
		StagingID:          stagingID,
		OwnerProfileID:     profileID,
		StorageIdentity:    storageIdentity,
		ExpectedByteSize:   byteSize,
		ExpectedSHA256:     digest,
		StagedAt:           stagedAt,
		StagingExpiresAt:   stagedAt.Add(24 * time.Hour),
		ReadyAt:            nil,
		PublishedAt:        nil,
		AbandonedAt:        nil,
		State:              StagedAllocated,
		DeleteState:        DeleteNotApplicable,
		DeleteAttemptCount: 0,
		NextDeleteAttempt:  nil,
		LastDeleteError:    nil,
	}
}

func (s *Store) AllocateStagedObject(ctx context.Context, object StagedObject) error {
	if object.State != StagedAllocated || object.DeleteState != DeleteNotApplicable || object.ReadyAt != nil || object.PublishedAt != nil || object.AbandonedAt != nil || object.DeleteAttemptCount != 0 || object.NextDeleteAttempt != nil || object.LastDeleteError != nil {
		return ErrInvalidTransition
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO extension_staged_objects (
    staging_id, owner_profile_id, storage_identity, expected_byte_size,
    expected_sha256, staged_at, staging_expires_at, ready_at, published_at,
    abandoned_at, state, delete_state, delete_attempt_count,
    next_delete_attempt_at, last_delete_error_code
) VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, NULL, NULL,
          'allocated', 'not_applicable', 0, NULL, NULL)
`, object.StagingID, object.OwnerProfileID, object.StorageIdentity, object.ExpectedByteSize,
		object.ExpectedSHA256, object.StagedAt.UTC(), object.StagingExpiresAt.UTC())
	return err
}

func (s *Store) MarkStagedReady(ctx context.Context, stagingID string, now time.Time) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE extension_staged_objects
   SET state = 'ready', ready_at = $2
 WHERE staging_id = $1 AND state = 'allocated' AND staging_expires_at > $2
`, stagingID, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrInvalidTransition
	}
	return nil
}

// MarkStagedPublished must be invoked on the caller's final transaction so the
// authoritative reference and publication timestamp become visible atomically.
func MarkStagedPublished(ctx context.Context, tx pgx.Tx, stagingID, resourceKind, resourceID string, byteSize int64, digest string, now time.Time) error {
	if tx == nil {
		return ErrInvalidTransition
	}
	tag, err := tx.Exec(ctx, `
UPDATE extension_staged_objects
   SET state = 'published', published_at = $2
 WHERE staging_id = $1
   AND state = 'ready'
   AND staging_expires_at > $2
   AND expected_byte_size = $3
   AND expected_sha256 = $4
`, stagingID, now.UTC(), byteSize, digest)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrIntegrity
	}
	_, err = tx.Exec(ctx, `
INSERT INTO extension_staged_object_references (
    staging_id, owner_resource_kind, owner_resource_id, expected_byte_size,
    expected_sha256, committed_at
) VALUES ($1, $2, $3, $4, $5, $6)
`, stagingID, resourceKind, resourceID, byteSize, digest, now.UTC())
	return err
}

func (s *Store) AbandonStagedObject(ctx context.Context, stagingID string, now time.Time) error {
	current := now.UTC()
	tag, err := s.pool.Exec(ctx, `
UPDATE extension_staged_objects
   SET state = 'abandoned', ready_at = NULL, abandoned_at = $2, delete_state = 'pending',
       next_delete_attempt_at = $2, last_delete_error_code = NULL
 WHERE staging_id = $1 AND state IN ('allocated', 'ready')
`, stagingID, current)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrInvalidTransition
	}
	return nil
}

func (s *Store) RedeemableStagedObject(ctx context.Context, stagingID string, now time.Time) (StagedObject, error) {
	object, err := s.StagedObject(ctx, stagingID)
	if err != nil {
		return StagedObject{}, err
	}
	if object.State != StagedPublished {
		if !now.UTC().Before(object.StagingExpiresAt) {
			return StagedObject{}, ErrNotFound
		}
		return StagedObject{}, ErrNotFound
	}
	return object, nil
}

func (s *Store) StagedObject(ctx context.Context, stagingID string) (StagedObject, error) {
	return scanStagedObject(s.pool.QueryRow(ctx, stagedObjectSelect+` WHERE staging_id = $1`, stagingID))
}

const stagedObjectSelect = `
SELECT staging_id, owner_profile_id, storage_identity, expected_byte_size,
       expected_sha256, staged_at, staging_expires_at, ready_at, published_at,
       abandoned_at, state, delete_state, delete_attempt_count,
       next_delete_attempt_at, last_delete_error_code
  FROM extension_staged_objects`

func scanStagedObject(row pgx.Row) (StagedObject, error) {
	var object StagedObject
	err := row.Scan(
		&object.StagingID,
		&object.OwnerProfileID,
		&object.StorageIdentity,
		&object.ExpectedByteSize,
		&object.ExpectedSHA256,
		&object.StagedAt,
		&object.StagingExpiresAt,
		&object.ReadyAt,
		&object.PublishedAt,
		&object.AbandonedAt,
		&object.State,
		&object.DeleteState,
		&object.DeleteAttemptCount,
		&object.NextDeleteAttempt,
		&object.LastDeleteError,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return StagedObject{}, ErrNotFound
	}
	return object, err
}

// PrepareCleanupBatch commits logical inaccessibility before returning physical
// deletion work. It never performs object-store I/O while its transaction is open.
func (s *Store) PrepareCleanupBatch(ctx context.Context, cutoff, now time.Time, limit int) ([]StagedObject, error) {
	if limit < 1 {
		return nil, errors.New("extension staged-object cleanup limit must be positive")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, stagedObjectSelect+`
 WHERE ((state IN ('allocated', 'ready') AND staging_expires_at <= $1)
     OR (state = 'abandoned' AND delete_state = 'pending' AND next_delete_attempt_at <= $1))
 ORDER BY CASE
            WHEN state IN ('allocated', 'ready') THEN staging_expires_at
            ELSE next_delete_attempt_at
          END,
          staging_id
 LIMIT $2
 FOR UPDATE
`, cutoff.UTC(), limit)
	if err != nil {
		return nil, err
	}
	var candidates []StagedObject
	for rows.Next() {
		object, err := scanStagedObject(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		candidates = append(candidates, object)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range candidates {
		object := &candidates[index]
		var referenceCount int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM extension_staged_object_references WHERE staging_id = $1`, object.StagingID).Scan(&referenceCount); err != nil {
			return nil, err
		}
		if referenceCount != 0 {
			return nil, fmt.Errorf("%w: nonpublished staged object has authoritative reference", ErrIntegrity)
		}
		if object.State == StagedAllocated || object.State == StagedReady {
			_, err := tx.Exec(ctx, `
UPDATE extension_staged_objects
   SET state = 'abandoned', ready_at = NULL, abandoned_at = $2, delete_state = 'pending',
       next_delete_attempt_at = $2, last_delete_error_code = NULL
 WHERE staging_id = $1
`, object.StagingID, now.UTC())
			if err != nil {
				return nil, err
			}
			object.State = StagedAbandoned
			object.ReadyAt = nil
			object.DeleteState = DeletePending
			abandonedAt := now.UTC()
			object.AbandonedAt = &abandonedAt
			object.NextDeleteAttempt = &abandonedAt
			object.LastDeleteError = nil
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: cleanup state commit indeterminate: %v", ErrIntegrity, err)
	}
	return candidates, nil
}

func (s *Store) RecordDeletionSuccess(ctx context.Context, stagingID string) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE extension_staged_objects
   SET delete_state = 'deleted', next_delete_attempt_at = NULL,
       last_delete_error_code = NULL
 WHERE staging_id = $1 AND state = 'abandoned' AND delete_state = 'pending'
`, stagingID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrIntegrity
	}
	return nil
}

func (s *Store) RecordDeletionFailure(ctx context.Context, stagingID string, attemptCount int32, safeErrorCode string, nextAttemptAt time.Time) error {
	object, err := s.StagedObject(ctx, stagingID)
	if err != nil {
		return err
	}
	if object.State != StagedAbandoned || object.DeleteState != DeletePending {
		return ErrIntegrity
	}
	if attemptCount < object.DeleteAttemptCount ||
		(object.DeleteAttemptCount < 1<<31-1 && attemptCount != object.DeleteAttemptCount+1) ||
		(object.DeleteAttemptCount == 1<<31-1 && attemptCount != object.DeleteAttemptCount) ||
		safeErrorCode == "" ||
		nextAttemptAt.IsZero() {
		return ErrInvalidTransition
	}
	tag, err := s.pool.Exec(ctx, `
UPDATE extension_staged_objects
   SET delete_attempt_count = $2,
       last_delete_error_code = $3,
       next_delete_attempt_at = $4
 WHERE staging_id = $1 AND state = 'abandoned' AND delete_state = 'pending'
   AND delete_attempt_count = $5
`, stagingID, attemptCount, safeErrorCode, nextAttemptAt.UTC(), object.DeleteAttemptCount)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrIntegrity
	}
	return nil
}

package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// blobSlotRepository owns pending-slot persistence inside a caller transaction.
type blobSlotRepository struct{}

func (repository blobSlotRepository) insertTx(
	ctx context.Context,
	tx pgx.Tx,
	params BlobSlotParams,
) error {
	_, err := tx.Exec(ctx, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, expected_sha256_hex,
    target_expires_at, pending_expires_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, 'pending', $5, $6, $7, $8, $9, $10, $11, $11)
`, params.ObjectBlobID, params.IncidentID, params.ActorUserID, params.StorageKey, params.ByteSize,
		params.FilenameHint, params.ContentTypeHint, params.ExpectedSHA256Hex,
		params.TargetExpiresAt.UTC(), params.PendingExpiresAt.UTC(), params.TargetExpiresAt.Add(-60*time.Minute).UTC())
	return err
}

// blobRepository owns blob reads and row locking. Lifecycle transitions remain
// with blobLifecycleRepository so storage identity and state changes are not
// conflated.
type blobRepository struct {
	db postgres.DB
}

func (repository blobRepository) load(ctx context.Context, objectBlobID uuid.UUID) (BlobRecord, error) {
	row := repository.db.QueryRow(ctx, `
SELECT object_blob_id, incident_id, storage_key, upload_state, byte_size,
       filename_hint, content_type_hint, expected_sha256_hex,
       observed_size, observed_content_type, observed_sha256_hex,
       target_expires_at, pending_expires_at
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID)
	var record BlobRecord
	if err := row.Scan(&record.ObjectBlobID, &record.IncidentID, &record.StorageKey, &record.UploadState, &record.ByteSize,
		&record.FilenameHint, &record.ContentTypeHint, &record.ExpectedSHA256Hex,
		&record.ObservedSize, &record.ObservedContentType, &record.ObservedSHA256Hex,
		&record.TargetExpiresAt, &record.PendingExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BlobRecord{}, ErrBlobNotFound
		}
		return BlobRecord{}, err
	}
	return record, nil
}

func (blobRepository) loadForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
) (BlobRecord, error) {
	return loadBlobForUpdateTx(ctx, tx, objectBlobID)
}

type evidenceRecordRepository struct{}

func (evidenceRecordRepository) loadForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (evidenceMeta, error) {
	return loadEvidenceMetaForUpdateTx(ctx, tx, recordID)
}

func (evidenceRecordRepository) associateBlobTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	objectBlobID uuid.UUID,
	storageRef string,
	sha256 *string,
	now time.Time,
) error {
	_, err := tx.Exec(ctx, `
UPDATE evidence
   SET object_blob_id = $2,
       lifecycle_state = CASE WHEN lifecycle_state IN ('requested', 'pending_receipt', 'received') THEN 'available' ELSE lifecycle_state END,
       upload_state = 'available',
       storage_ref = $3,
       blob_hash = $4,
       received_at = COALESCE(received_at, $5),
       updated_at = $5
 WHERE record_id = $1
`, recordID, objectBlobID, storageRef, sha256, now)
	return err
}

// blobLifecycleRepository owns blob-state transitions and cleanup selection.
// Callers retain transaction orchestration and object-store byte deletion.
type blobLifecycleRepository struct {
	db postgres.DB
}

func (blobLifecycleRepository) failTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
	reason string,
	now time.Time,
) error {
	return failBlobTx(ctx, tx, objectBlobID, reason, now)
}

func (blobLifecycleRepository) recordFinalizeFailureTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
	now time.Time,
) (bool, error) {
	return recordNonTerminalFinalizeFailureTx(ctx, tx, objectBlobID, now)
}

func (blobLifecycleRepository) markAvailableTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
	observed *ObservedObject,
	now time.Time,
) error {
	return markBlobAvailableTx(ctx, tx, objectBlobID, observed, now)
}

func (repository blobLifecycleRepository) markExpiredPending(
	ctx context.Context,
	now time.Time,
) (int, error) {
	tag, err := repository.db.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'failed',
       terminal_reason = 'pending_timeout',
       failed_at = $1,
       cleanup_due_at = $1::timestamptz + interval '1 hour',
       updated_at = $1
 WHERE upload_state = 'pending'
   AND pending_expires_at <= $1
`, now.UTC())
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (repository blobLifecycleRepository) failedUnattachedCleanupCandidates(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]cleanupCandidate, error) {
	rows, err := repository.db.Query(ctx, `
SELECT b.object_blob_id, b.storage_key
  FROM object_blobs b
 WHERE b.upload_state = 'failed'
   AND b.cleaned_up_at IS NULL
   AND b.cleanup_due_at <= $1
   AND NOT EXISTS (
       SELECT 1
         FROM evidence e
        WHERE e.object_blob_id = b.object_blob_id
   )
 ORDER BY b.cleanup_due_at, b.object_blob_id
 LIMIT $2
`, now.UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	candidates := make([]cleanupCandidate, 0)
	for rows.Next() {
		var candidate cleanupCandidate
		if err := rows.Scan(&candidate.ObjectBlobID, &candidate.StorageKey); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return candidates, nil
}

package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
)

type evidenceMeta struct {
	IncidentID uuid.UUID
	RowVersion int64
}

type rowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *rowVersionConflictError) Error() string { return errRowVersionConflict.Error() }
func (e *rowVersionConflictError) Unwrap() error { return errRowVersionConflict }

func loadEvidenceMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (evidenceMeta, error) {
	var meta evidenceMeta
	err := tx.QueryRow(ctx, `
SELECT r.incident_id, r.row_version
  FROM records r
  JOIN evidence e ON e.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'evidence'
   AND r.deleted_at IS NULL
 FOR UPDATE OF r, e
`, recordID).Scan(&meta.IncidentID, &meta.RowVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return evidenceMeta{}, ErrEvidenceNotFound
	}
	return meta, err
}

func loadBlobForUpdateTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID) (blobRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT b.object_blob_id, b.incident_id, b.storage_key, b.upload_state, b.byte_size,
       b.filename_hint, b.content_type_hint, b.expected_sha256_hex,
       b.observed_size, b.observed_content_type, b.observed_sha256_hex,
       b.target_expires_at, b.pending_expires_at,
       COALESCE(l.lease_state, '')
  FROM object_blobs b
  LEFT JOIN evidence_object_upload_leases l ON l.object_blob_id = b.object_blob_id
 WHERE b.object_blob_id = $1
 FOR UPDATE OF b
`, objectBlobID)
	var record blobRecord
	if err := row.Scan(&record.ObjectBlobID, &record.IncidentID, &record.StorageKey, &record.UploadState, &record.ByteSize,
		&record.FilenameHint, &record.ContentTypeHint, &record.ExpectedSHA256Hex,
		&record.ObservedSize, &record.ObservedContentType, &record.ObservedSHA256Hex,
		&record.TargetExpiresAt, &record.PendingExpiresAt, &record.UploadLeaseState); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return blobRecord{}, ErrBlobNotFound
		}
		return blobRecord{}, err
	}
	return record, nil
}

func failBlobTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, reason string, now time.Time) error {
	schedule := evidencepolicy.ScheduleFailure(now)
	tag, err := tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'failed',
       terminal_reason = $2,
       failed_at = $3,
       cleanup_due_at = $4,
       updated_at = $3
 WHERE object_blob_id = $1
   AND upload_state = 'pending'
`, objectBlobID, reason, schedule.FailedAt, schedule.CleanupDueAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errIllegalBlobTransition
	}
	return nil
}

func recordNonTerminalFinalizeFailureTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, now time.Time) (bool, error) {
	schedule := evidencepolicy.ScheduleFailure(now)
	var uploadState string
	err := tx.QueryRow(ctx, `
UPDATE object_blobs
   SET finalize_attempt_count = finalize_attempt_count + 1,
       upload_state = CASE WHEN finalize_attempt_count + 1 >= $4 THEN 'failed' ELSE upload_state END,
       terminal_reason = CASE WHEN finalize_attempt_count + 1 >= $4 THEN 'finalize_retry_exhausted' ELSE terminal_reason END,
       failed_at = CASE WHEN finalize_attempt_count + 1 >= $4 THEN $2 ELSE failed_at END,
       cleanup_due_at = CASE WHEN finalize_attempt_count + 1 >= $4 THEN $3 ELSE cleanup_due_at END,
       updated_at = $2
 WHERE object_blob_id = $1
   AND upload_state = 'pending'
 RETURNING upload_state
`, objectBlobID, schedule.FailedAt, schedule.CleanupDueAt, evidencepolicy.TerminalFinalizeAttempt).Scan(&uploadState)
	return uploadState == "failed", err
}

func markBlobAvailableTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, observed *observedObject, now time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'available',
       observed_size = $2,
       observed_content_type = $3,
       observed_sha256_hex = $4,
       finalized_at = $5,
       updated_at = $5
 WHERE object_blob_id = $1
   AND upload_state = 'pending'
`, objectBlobID, observed.Size, observed.ContentType, observed.SHA256Hex, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errIllegalBlobTransition
	}
	return nil
}

func ensureIncidentVisibleTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM incidents WHERE id = $1)`, incidentID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	return nil
}

func firstNonEmptyPtr(left *string, right *string, fallback string) string {
	if left != nil && *left != "" {
		return *left
	}
	if right != nil && *right != "" {
		return *right
	}
	return fallback
}

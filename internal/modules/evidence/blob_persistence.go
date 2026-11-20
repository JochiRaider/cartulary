package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type evidenceMeta struct {
	IncidentID uuid.UUID
	RowVersion int64
	// Deleted reports a soft-deleted Evidence record. The attach route returns
	// an exact committed replay before it reports the deleted state
	// (REQ-01-705).
	Deleted bool
}

type rowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *rowVersionConflictError) Error() string { return errRowVersionConflict.Error() }
func (e *rowVersionConflictError) Unwrap() error { return errRowVersionConflict }

// loadEvidenceMetaForUpdateTx locks the Evidence envelope and subtype row,
// including a soft-deleted record, so the caller can classify the target in
// REQ-01-705 order: a missing Evidence record is not found, and a deleted one
// is reported only after the exact-replay lookup.
func loadEvidenceMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (evidenceMeta, error) {
	var meta evidenceMeta
	var deletedAt *time.Time
	err := tx.QueryRow(ctx, `
SELECT r.incident_id, r.row_version, r.deleted_at
  FROM records r
  JOIN evidence e ON e.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'evidence'
 FOR UPDATE OF r, e
`, recordID).Scan(&meta.IncidentID, &meta.RowVersion, &deletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return evidenceMeta{}, ErrEvidenceNotFound
	}
	if err != nil {
		return evidenceMeta{}, err
	}
	meta.Deleted = errors.Is(records.ClassifyTargetPresence(true, deletedAt), records.ErrTargetDeleted)
	return meta, nil
}

// loadBlobForUpdateTx locks the Blob through the single Blob authority, which
// verifies the REQ-02-189a tuple under the lock. A violation fails as
// errBlobTupleViolation (internal_error) before any caller decision.
func loadBlobForUpdateTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID) (blobRecord, error) {
	locked, err := evidenceblob.LockTx(ctx, tx, objectBlobID, evidenceblob.UpdateWait)
	if errors.Is(err, evidenceblob.ErrNotFound) {
		return blobRecord{}, ErrBlobNotFound
	}
	if err != nil {
		return blobRecord{}, err
	}
	record := blobRecord{
		ObjectBlobID: locked.ObjectBlobID, IncidentID: locked.IncidentID, StorageKey: locked.StorageKey,
		UploadState: locked.UploadState, ByteSize: locked.ByteSize,
		FilenameHint: locked.FilenameHint, ContentTypeHint: locked.ContentTypeHint, ExpectedSHA256Hex: locked.ExpectedSHA256Hex,
		ObservedSize: locked.ObservedSize, ObservedContentType: locked.ObservedContentType, ObservedSHA256Hex: locked.ObservedSHA256Hex,
		TargetExpiresAt: locked.TargetExpiresAt, PendingExpiresAt: locked.PendingExpiresAt,
	}
	if err := tx.QueryRow(ctx, `
SELECT COALESCE((
           SELECT l.lease_state
             FROM evidence_object_upload_leases l
            WHERE l.object_blob_id = $1
       ), '')
`, objectBlobID).Scan(&record.UploadLeaseState); err != nil {
		return blobRecord{}, err
	}
	return record, nil
}

// blobTransitionError is a rejected Blob upload-state transition. The Blob
// machine is matrix-only (Core 01 §3.3.6.2), so it carries no guard.
type blobTransitionError struct {
	fromStatus string
	toStatus   string
}

func (e *blobTransitionError) Error() string { return errIllegalBlobTransition.Error() }
func (e *blobTransitionError) Unwrap() error { return errIllegalBlobTransition }

// illegalBlobTransitionTx reports the current state of a Blob whose guarded
// transition to toStatus matched no row.
func illegalBlobTransitionTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, toStatus string) error {
	var current string
	if err := tx.QueryRow(ctx, `SELECT upload_state FROM object_blobs WHERE object_blob_id = $1`, objectBlobID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBlobNotFound
		}
		return err
	}
	return &blobTransitionError{fromStatus: current, toStatus: toStatus}
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
		return illegalBlobTransitionTx(ctx, tx, objectBlobID, "failed")
	}
	return nil
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
		return illegalBlobTransitionTx(ctx, tx, objectBlobID, "available")
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

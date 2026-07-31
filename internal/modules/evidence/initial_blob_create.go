package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func (f *WorkbookFacade) observeInitialBlob(
	ctx context.Context,
	incidentID uuid.UUID,
	objectBlobID uuid.UUID,
) (*ObservedObject, error) {
	blob, err := f.store.GetBlob(ctx, objectBlobID)
	if errors.Is(err, ErrBlobNotFound) {
		return nil, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotFound}
	}
	if err != nil {
		return nil, err
	}
	if blob.IncidentID != incidentID {
		return nil, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrIncidentMismatch}
	}
	associated, err := isBlobAssociated(ctx, f.pool, objectBlobID)
	if err != nil {
		return nil, err
	}
	if associated {
		return nil, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
	}
	switch blob.UploadState {
	case "failed":
		return nil, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case "quarantined":
		return nil, AttachRejectedError{ReasonCode: AttachReasonBlobQuarantined, Cause: ErrBlobNotAttachable}
	case "available":
		return nil, nil
	case "pending":
	default:
		return nil, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
	}
	if f.objects == nil {
		return nil, ErrObjectStoreUnavailable
	}
	observed, err := (routeObjectStoreAdapter{store: f.objects}).observeUploadedObject(ctx, blob)
	if objectstore.IsObjectNotFound(err) {
		return nil, AttachRejectedError{ReasonCode: AttachReasonBlobPending, Cause: ErrBlobNotAttachable}
	}
	if err != nil {
		return nil, err
	}
	return observed, nil
}

// finalizeInitialBlobTx locks the slot and rechecks every mutable association
// precondition in the transaction that creates the Evidence row. The bool
// return requests a blob-only commit for a terminal failure disposition.
func (f *WorkbookFacade) finalizeInitialBlobTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	objectBlobID uuid.UUID,
	observed *ObservedObject,
	now time.Time,
) (*InitialBlobAssociation, bool, error) {
	blob, err := f.store.blobs.loadForUpdateTx(ctx, tx, objectBlobID)
	if errors.Is(err, ErrBlobNotFound) {
		return nil, false, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotFound}
	}
	if err != nil {
		return nil, false, err
	}
	if blob.IncidentID != incidentID {
		return nil, false, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrIncidentMismatch}
	}
	associated, err := isBlobAssociatedTx(ctx, tx, objectBlobID)
	if err != nil {
		return nil, false, err
	}
	if associated {
		return nil, false, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
	}
	switch blob.UploadState {
	case "failed":
		return nil, false, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case "quarantined":
		return nil, false, AttachRejectedError{ReasonCode: AttachReasonBlobQuarantined, Cause: ErrBlobNotAttachable}
	case "pending":
		if !blob.PendingExpiresAt.After(now) {
			if err := f.store.blobLifecycle.failTx(ctx, tx, objectBlobID, "pending_timeout", now); err != nil {
				return nil, false, err
			}
			return nil, true, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
		}
		if observed == nil {
			return nil, false, AttachRejectedError{ReasonCode: AttachReasonBlobPending, Cause: ErrBlobNotAttachable}
		}
		if observed.Size != blob.ByteSize {
			if err := f.store.blobLifecycle.failTx(ctx, tx, objectBlobID, "declared_size_mismatch", now); err != nil {
				return nil, false, err
			}
			return nil, true, AttachRejectedError{ReasonCode: AttachReasonAcceptedContractMismatch, Cause: ErrBlobNotAttachable}
		}
		if blob.ExpectedSHA256Hex != nil && observed.SHA256Hex != *blob.ExpectedSHA256Hex {
			if err := f.store.blobLifecycle.failTx(ctx, tx, objectBlobID, "expected_sha256_mismatch", now); err != nil {
				return nil, false, err
			}
			return nil, true, AttachRejectedError{ReasonCode: AttachReasonAcceptedContractMismatch, Cause: ErrBlobNotAttachable}
		}
		if err := f.store.blobLifecycle.markAvailableTx(ctx, tx, objectBlobID, observed, now); err != nil {
			return nil, false, err
		}
		blob.UploadState = "available"
		blob.ObservedSHA256Hex = &observed.SHA256Hex
	case "available":
	default:
		return nil, false, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
	}
	storageRef, err := blobref.ObjectBlobStorageRef(objectBlobID)
	if err != nil {
		return nil, false, err
	}
	return &InitialBlobAssociation{
		ObjectBlobID: objectBlobID,
		StorageRef:   storageRef,
		SHA256Hex:    blob.ObservedSHA256Hex,
	}, false, nil
}

type queryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func isBlobAssociated(ctx context.Context, db queryRower, objectBlobID uuid.UUID) (bool, error) {
	var associated bool
	if err := db.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM evidence
     WHERE object_blob_id = $1
)
`, objectBlobID).Scan(&associated); err != nil {
		return false, err
	}
	return associated, nil
}

func isBlobAssociatedTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID) (bool, error) {
	return isBlobAssociated(ctx, tx, objectBlobID)
}

func isEvidenceBlobUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == "evidence_object_blob_unique_idx"
}

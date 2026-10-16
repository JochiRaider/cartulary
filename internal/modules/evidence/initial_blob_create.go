package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
)

// finalizeInitialBlobTx locks the slot and rechecks every mutable association
// precondition in the transaction that creates the Evidence row. The bool
// return requests a blob-only commit for an admitted failure disposition.
func (f *mutationFacade) finalizeInitialBlobTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	objectBlobID uuid.UUID,
	now time.Time,
) (*initialBlobAssociation, bool, error) {
	blob, err := f.blobs.loadForUpdateTx(ctx, tx, objectBlobID)
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
		return nil, false, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
	}
	switch blob.UploadState {
	case "failed":
		return nil, false, AttachRejectedError{ReasonCode: attachReasonBlobFailed, Cause: errBlobNotAttachable}
	case "quarantined":
		return nil, false, AttachRejectedError{ReasonCode: attachReasonBlobQuarantined, Cause: errBlobNotAttachable}
	case "pending":
		if blob.UploadLeaseState != "completed" {
			return nil, false, AttachRejectedError{ReasonCode: attachReasonBlobPending, Cause: errBlobNotAttachable}
		}
		if !blob.PendingExpiresAt.After(now) {
			if err := f.blobLifecycle.failTx(ctx, tx, objectBlobID, "pending_timeout", now); err != nil {
				return nil, false, err
			}
			return nil, true, AttachRejectedError{ReasonCode: attachReasonBlobFailed, Cause: errBlobNotAttachable}
		}
		observed, err := retainedBlobObservation(ctx, tx, objectBlobID)
		if errors.Is(err, errBlobNotVerified) {
			return nil, false, AttachRejectedError{ReasonCode: attachReasonBlobPending, Cause: errBlobNotAttachable}
		}
		if err != nil {
			return nil, false, err
		}
		if observed.Size != blob.ByteSize {
			if err := f.blobLifecycle.failTx(ctx, tx, objectBlobID, "declared_size_mismatch", now); err != nil {
				return nil, false, err
			}
			return nil, true, AttachRejectedError{ReasonCode: attachReasonAcceptedContractMismatch, Cause: errBlobNotAttachable}
		}
		if blob.ExpectedSHA256Hex != nil && observed.SHA256Hex != *blob.ExpectedSHA256Hex {
			if err := f.blobLifecycle.failTx(ctx, tx, objectBlobID, "expected_sha256_mismatch", now); err != nil {
				return nil, false, err
			}
			return nil, true, AttachRejectedError{ReasonCode: attachReasonAcceptedContractMismatch, Cause: errBlobNotAttachable}
		}
		if err := f.blobLifecycle.markAvailableTx(ctx, tx, objectBlobID, observed, now); err != nil {
			return nil, false, err
		}
		blob.UploadState = "available"
		blob.ObservedSHA256Hex = &observed.SHA256Hex
	case "available":
	default:
		return nil, false, AttachRejectedError{ReasonCode: attachReasonEvidenceInconsistent, Cause: errBlobNotAttachable}
	}
	storageRef, err := blobref.ObjectBlobStorageRef(objectBlobID)
	if err != nil {
		return nil, false, err
	}
	return &initialBlobAssociation{
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

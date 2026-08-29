package evidence

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

// PreflightAttachBlob evaluates every durable authorization, lifecycle,
// contract, version, and idempotency gate that must precede object-store
// observation. AttachBlob repeats these checks inside its mutation transaction.
func (s *blobLifecycleService) PreflightAttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request attachBlobRequest, requestHash []byte, now time.Time) (attachBlobPreflightResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return attachBlobPreflightResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := s.evidenceRows.loadForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobPreflightResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return attachBlobPreflightResult{}, err
	}
	row, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobPreflightResult{}, err
	}
	if evidenceRowCellValue(row, "evidence.lifecycle_state") == "quarantined" {
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: attachReasonEvidenceQuarantined, Cause: errEvidenceQuarantined}
	}
	blob, err := s.blobs.loadForUpdateTx(ctx, tx, request.ObjectBlobID)
	if err != nil {
		return attachBlobPreflightResult{}, err
	}
	if blob.IncidentID != meta.IncidentID {
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrIncidentMismatch}
	}
	var associatedRecordID uuid.UUID
	associationErr := tx.QueryRow(ctx, `
SELECT record_id
  FROM evidence
 WHERE object_blob_id = $1
`, request.ObjectBlobID).Scan(&associatedRecordID)
	if associationErr != nil && !errors.Is(associationErr, pgx.ErrNoRows) {
		return attachBlobPreflightResult{}, associationErr
	}
	associated := associationErr == nil
	if associated && associatedRecordID != recordID {
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
	}
	switch evidencepolicy.ClassifyBlobForAssociation(blob.UploadState, blob.PendingExpiresAt, now) {
	case evidencepolicy.AssociationBlobNeedsFinalization:
		if blob.UploadLeaseState != "completed" {
			return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: attachReasonBlobPending, Cause: errBlobNotAttachable}
		}
	case evidencepolicy.AssociationBlobAvailable:
	case evidencepolicy.AssociationBlobExpired, evidencepolicy.AssociationBlobFailed:
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: attachReasonBlobFailed, Cause: errBlobNotAttachable}
	case evidencepolicy.AssociationBlobQuarantined:
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: attachReasonBlobQuarantined, Cause: errBlobNotAttachable}
	case evidencepolicy.AssociationBlobInconsistent:
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: attachReasonEvidenceInconsistent, Cause: errBlobNotAttachable}
	}
	key := LifecycleIdempotencyKey{
		OperationID: LifecycleOperationBlobAttach, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	if payload, found, err := s.idempotency.GetTx(ctx, tx, key, requestHash); err != nil {
		return attachBlobPreflightResult{}, err
	} else if found {
		replay := attachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: request.ClientTxnID}
		if err := tx.Commit(ctx); err != nil {
			return attachBlobPreflightResult{}, err
		}
		return attachBlobPreflightResult{Blob: blob, Replay: &replay}, nil
	}
	if associated {
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
	}
	if meta.RowVersion != request.BaseRowVersion {
		return attachBlobPreflightResult{}, &rowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
	}
	if err := tx.Commit(ctx); err != nil {
		return attachBlobPreflightResult{}, err
	}
	return attachBlobPreflightResult{Blob: blob}, nil
}

func (s *blobLifecycleService) AttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request attachBlobRequest, requestHash []byte, observed *observedObject, requestID string, now time.Time) (attachBlobResult, error) {
	key := LifecycleIdempotencyKey{
		OperationID: LifecycleOperationBlobAttach, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	if payload, found, err := s.idempotency.Get(ctx, key, requestHash); err != nil {
		return attachBlobResult{}, err
	} else if found {
		return attachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: request.ClientTxnID}, nil
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return attachBlobResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := s.evidenceRows.loadForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return attachBlobResult{}, err
	}
	if meta.RowVersion != request.BaseRowVersion {
		return attachBlobResult{}, &rowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
	}
	beforeRow, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobResult{}, err
	}
	beforeSnapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobResult{}, err
	}
	if evidenceRowCellValue(beforeRow, "evidence.lifecycle_state") == "quarantined" {
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonEvidenceQuarantined, Cause: errEvidenceQuarantined}
	}
	blob, err := s.blobs.loadForUpdateTx(ctx, tx, request.ObjectBlobID)
	if err != nil {
		return attachBlobResult{}, err
	}
	if blob.IncidentID != meta.IncidentID {
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrIncidentMismatch}
	}
	associated, err := isBlobAssociatedTx(ctx, tx, request.ObjectBlobID)
	if err != nil {
		return attachBlobResult{}, err
	}
	if associated {
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
	}
	switch evidencepolicy.ClassifyBlobForAssociation(blob.UploadState, blob.PendingExpiresAt, now) {
	case evidencepolicy.AssociationBlobQuarantined:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobQuarantined, Cause: errBlobNotAttachable}
	case evidencepolicy.AssociationBlobFailed:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobFailed, Cause: errBlobNotAttachable}
	case evidencepolicy.AssociationBlobExpired:
		if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "pending_timeout", now); err != nil {
			return attachBlobResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return attachBlobResult{}, err
		}
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobFailed, Cause: errBlobNotAttachable}
	case evidencepolicy.AssociationBlobNeedsFinalization:
		if observed == nil {
			failed, err := s.blobLifecycle.recordFinalizeFailureTx(ctx, tx, request.ObjectBlobID, now)
			if err != nil {
				return attachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return attachBlobResult{}, err
			}
			reason := attachReasonBlobPending
			if failed {
				reason = attachReasonBlobFailed
			}
			return attachBlobResult{}, AttachRejectedError{ReasonCode: reason, Cause: errBlobNotAttachable}
		}
		if observed.Size != blob.ByteSize {
			if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "declared_size_mismatch", now); err != nil {
				return attachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return attachBlobResult{}, err
			}
			return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonAcceptedContractMismatch, Cause: errBlobNotAttachable}
		}
		if blob.ExpectedSHA256Hex != nil && observed.SHA256Hex != *blob.ExpectedSHA256Hex {
			if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "expected_sha256_mismatch", now); err != nil {
				return attachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return attachBlobResult{}, err
			}
			return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonAcceptedContractMismatch, Cause: errBlobNotAttachable}
		}
		if err := s.blobLifecycle.markAvailableTx(ctx, tx, request.ObjectBlobID, observed, now); err != nil {
			return attachBlobResult{}, err
		}
		blob.UploadState = "available"
		blob.ObservedSize = &observed.Size
		blob.ObservedContentType = &observed.ContentType
		blob.ObservedSHA256Hex = &observed.SHA256Hex
	case evidencepolicy.AssociationBlobAvailable:
	case evidencepolicy.AssociationBlobInconsistent:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonEvidenceInconsistent, Cause: errBlobNotAttachable}
	}
	if blob.UploadState != "available" {
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonEvidenceInconsistent, Cause: errBlobNotAttachable}
	}
	sha := blob.ObservedSHA256Hex
	if sha == nil && observed != nil {
		sha = &observed.SHA256Hex
	}
	storageRef, err := blobref.ObjectBlobStorageRef(request.ObjectBlobID)
	if err != nil {
		return attachBlobResult{}, err
	}
	evidenceLifecycle, ok := evidenceRowCellValue(beforeRow, "evidence.lifecycle_state").(string)
	if !ok || !evidencepolicy.ValidEvidenceLifecycle(evidenceLifecycle) ||
		evidencepolicy.ViolatesEvidenceBlobBridge(evidenceLifecycle, true, evidencepolicy.BlobAvailable) {
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonEvidenceInconsistent, Cause: errBlobNotAttachable}
	}
	if err := s.evidenceRows.associateBlobTx(
		ctx,
		tx,
		recordID,
		request.ObjectBlobID,
		storageRef,
		sha,
		now.UTC(),
	); err != nil {
		if isEvidenceBlobUniqueViolation(err) {
			return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
		}
		return attachBlobResult{}, err
	}
	if err := insertEvidenceCustodyEventTx(ctx, tx, evidenceCustodyEventParams{
		IncidentID:       meta.IncidentID,
		EvidenceRecordID: recordID,
		CustodyEventType: "made_available",
		ActorUserID:      &actor.ID,
		OccurredAt:       now.UTC(),
		Metadata:         map[string]any{"object_blob_id": request.ObjectBlobID.String()},
	}); err != nil {
		return attachBlobResult{}, err
	}
	rowVersion, err := s.records.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return attachBlobResult{}, err
	}
	if err := s.projections.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
		return attachBlobResult{}, err
	}
	afterRow, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobResult{}, err
	}
	afterSnapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobResult{}, err
	}
	changeSetID, err := s.revisionStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: meta.IncidentID, ActorUserID: actor.ID, Source: blobAttachRouteKey,
		ClientTxnID: &request.ClientTxnID, RequestID: &requestID, CreatedAt: now.UTC(),
	})
	if err != nil {
		return attachBlobResult{}, err
	}
	beforeVersionID := fmt.Sprintf("%s:%d", recordID, request.BaseRowVersion)
	afterVersionID := fmt.Sprintf("%s:%d", recordID, rowVersion)
	if err := s.revisionStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "record", RecordID: recordID,
		OperationKind: "patch", BeforeVersionID: &beforeVersionID, AfterVersionID: &afterVersionID,
		BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
	}); err != nil {
		return attachBlobResult{}, err
	}
	changedFieldKeys := sortedChangedKeys(beforeRow, afterRow)
	if err := s.revisionStore.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID: changeSetID, RecordID: recordID, RowVersion: rowVersion,
		BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
		ConflictFacts: evidenceRevisionFacts(beforeRow, afterRow, changedFieldKeys),
	}); err != nil {
		return attachBlobResult{}, err
	}
	affectedChanges, err := s.refreshEvidenceSupportProjectionsTx(ctx, tx, meta.IncidentID, recordID)
	if err != nil {
		return attachBlobResult{}, err
	}
	if err := appendEvidenceRecordChangeIntentsTx(
		ctx,
		tx,
		s.collaboration,
		meta.IncidentID,
		actor.ID,
		request.ClientTxnID,
		changeSetID,
		attachRecordChange{
			RecordID:         recordID,
			RowVersion:       rowVersion,
			ViewSchemaID:     ViewSchemaID,
			ChangedFieldKeys: changedFieldKeys,
		},
		afterRow,
		affectedChanges,
		now,
	); err != nil {
		return attachBlobResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": ViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"object_blob_id": request.ObjectBlobID.String(),
		"row":            afterRow,
	}
	if err := s.idempotency.PutTx(ctx, tx, key, requestHash, payload); err != nil {
		return attachBlobResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return attachBlobResult{}, err
	}
	return attachBlobResult{
		Payload: payload, StatusCode: http.StatusOK, IncidentID: meta.IncidentID, RecordID: recordID,
		ChangeSetID: changeSetID, ClientTxnID: request.ClientTxnID, RowVersion: rowVersion,
		ChangedFieldKeys: changedFieldKeys, AffectedRecordChanges: affectedChanges,
	}, nil
}

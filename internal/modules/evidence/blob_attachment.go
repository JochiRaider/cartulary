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
	"github.com/JochiRaider/cartulary/internal/modules/records"
	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

// errAttachTargetDeleted reports a soft-deleted attach target, which must be
// restored first (REQ-01-705 step 2). It matches records.ErrTargetDeleted, so
// the route can map it without depending on Records.
var errAttachTargetDeleted = fmt.Errorf("evidence: attach target is deleted: %w", records.ErrTargetDeleted)

// evidenceAttachTarget is the unlocked envelope of the record an attach request
// addresses. Its incident decides visibility before anything else about the
// record is classified (REQ-01-704), so it includes soft-deleted records and
// records of every type.
type evidenceAttachTarget struct {
	IncidentID uuid.UUID
	RecordType string
	Deleted    bool
}

// LoadAttachTarget reads the addressed record envelope through Records. A
// missing record is ErrEvidenceNotFound; any other read failure is returned
// unclassified (REQ-01-706).
func (s *blobLifecycleService) LoadAttachTarget(ctx context.Context, recordID uuid.UUID) (evidenceAttachTarget, error) {
	envelope, err := recordenvelope.ScanRow(s.pool.QueryRow(ctx, recordenvelope.SelectSQL+"\n WHERE record_id = $1", recordID))
	if errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
		return evidenceAttachTarget{}, ErrEvidenceNotFound
	}
	if err != nil {
		return evidenceAttachTarget{}, err
	}
	return evidenceAttachTarget{
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		Deleted:    errors.Is(records.ClassifyTargetPresence(true, envelope.DeletedAt), records.ErrTargetDeleted),
	}, nil
}

// PreflightAttachBlob classifies an attach under the Evidence row lock before
// any object-store observation. An exact committed replay returns its original
// receipt first (REQ-01-245, REQ-01-705). A fresh request then passes the
// deleted, incident-lifecycle, Evidence, Blob-eligibility and version gates,
// which AttachBlob repeats inside its mutation transaction.
func (s *blobLifecycleService) PreflightAttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request attachBlobRequest, requestHash []byte, now time.Time) (attachBlobPreflightResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return attachBlobPreflightResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	key := LifecycleIdempotencyKey{
		OperationID: LifecycleOperationBlobAttach, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	meta, err := s.evidenceRows.loadForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return attachBlobPreflightResult{}, err
	}
	replay, err := s.committedAttachReplayTx(ctx, tx, key, requestHash, recordID)
	if err != nil {
		return attachBlobPreflightResult{}, err
	}
	if replay != nil {
		if err := tx.Commit(ctx); err != nil {
			return attachBlobPreflightResult{}, err
		}
		return attachBlobPreflightResult{Replay: replay}, nil
	}
	if meta.Deleted {
		return attachBlobPreflightResult{}, errAttachTargetDeleted
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

// committedAttachReplayTx returns the original committed attach result for an
// exact replay of key, or nil when no receipt exists (REQ-01-245). It runs
// first under the Evidence row lock, before the deleted, incident-lifecycle,
// Blob-eligibility and version gates (REQ-01-705), and the same request under a
// divergent hash fails with ErrClientTxnConflict.
//
// Every attach commits its receipt while holding this Evidence row lock, so
// the lookup also sees a competing exact request that committed while this
// one waited. A replay only reads: it takes no incident or Blob lock, so it
// cannot wait behind or deadlock with a quarantine, which locks the Evidence
// before the Blob (REQ-02-190a), and it commits no attach, change set, custody
// event or Collaboration publication.
func (s *blobLifecycleService) committedAttachReplayTx(ctx context.Context, tx pgx.Tx, key LifecycleIdempotencyKey, requestHash []byte, recordID uuid.UUID) (*attachBlobResult, error) {
	payload, found, err := s.idempotency.GetTx(ctx, tx, key, requestHash)
	if err != nil || !found {
		return nil, err
	}
	return &attachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: key.ClientTxnID}, nil
}

func (s *blobLifecycleService) AttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request attachBlobRequest, requestHash []byte, requestID string, now time.Time) (attachBlobResult, error) {
	key := LifecycleIdempotencyKey{
		OperationID: LifecycleOperationBlobAttach, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
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
	// A competing exact request may have committed between preflight and this
	// lock; its receipt still precedes every other gate.
	replay, err := s.committedAttachReplayTx(ctx, tx, key, requestHash, recordID)
	if err != nil {
		return attachBlobResult{}, err
	}
	if replay != nil {
		return *replay, nil
	}
	if meta.Deleted {
		return attachBlobResult{}, errAttachTargetDeleted
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return attachBlobResult{}, err
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
	var associatedRecordID uuid.UUID
	associationErr := tx.QueryRow(ctx, `SELECT record_id FROM evidence WHERE object_blob_id = $1`, request.ObjectBlobID).Scan(&associatedRecordID)
	if associationErr != nil && !errors.Is(associationErr, pgx.ErrNoRows) {
		return attachBlobResult{}, associationErr
	}
	associated := associationErr == nil
	if associated && associatedRecordID != recordID {
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
	}
	// Repeat current eligibility: observation may have taken time.
	switch evidencepolicy.ClassifyBlobForAssociation(blob.UploadState, blob.PendingExpiresAt, now) {
	case evidencepolicy.AssociationBlobNeedsFinalization:
		if blob.UploadLeaseState != "completed" {
			return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobPending, Cause: errBlobNotAttachable}
		}
	case evidencepolicy.AssociationBlobFailed:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobFailed, Cause: errBlobNotAttachable}
	case evidencepolicy.AssociationBlobQuarantined:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobQuarantined, Cause: errBlobNotAttachable}
	case evidencepolicy.AssociationBlobInconsistent:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonEvidenceInconsistent, Cause: errBlobNotAttachable}
	}
	if meta.RowVersion != request.BaseRowVersion {
		return attachBlobResult{}, &rowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
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
		if blob.UploadLeaseState != "completed" {
			return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobPending, Cause: errBlobNotAttachable}
		}
		observed, err := retainedBlobObservation(ctx, tx, request.ObjectBlobID)
		if errors.Is(err, errBlobNotVerified) {
			return attachBlobResult{}, AttachRejectedError{ReasonCode: attachReasonBlobPending, Cause: errBlobNotAttachable}
		}
		if err != nil {
			return attachBlobResult{}, err
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

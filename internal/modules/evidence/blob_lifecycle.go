package evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrBlobNotFound           = errors.New("evidence: blob not found")
	ErrEvidenceNotFound       = errors.New("evidence: evidence not found")
	ErrBlobNotAttachable      = errors.New("evidence: blob not attachable")
	ErrEvidenceQuarantined    = errors.New("evidence: quarantined")
	ErrIncidentMismatch       = errors.New("evidence: incident mismatch")
	ErrRowVersionConflict     = errors.New("evidence: row version conflict")
	ErrIllegalBlobTransition  = errors.New("evidence: illegal blob transition")
	ErrObjectStoreUnavailable = errors.New("evidence: object store unavailable")
	errUploadLeaseNotFound    = errors.New("evidence: upload lease not found")
	errUploadLeaseUnavailable = errors.New("evidence: upload lease unavailable")
)

const (
	AttachReasonBlobNotVisible           = "blob_not_visible"
	AttachReasonBlobPending              = "blob_pending"
	AttachReasonBlobFailed               = "blob_failed"
	AttachReasonBlobQuarantined          = "blob_quarantined"
	AttachReasonAcceptedContractMismatch = "accepted_contract_mismatch"
	AttachReasonEvidenceQuarantined      = "evidence_quarantined"
	AttachReasonEvidenceInconsistent     = "evidence_inconsistent"
)

type AttachRejectedError struct {
	ReasonCode string
	Cause      error
}

func (e AttachRejectedError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return ErrBlobNotAttachable.Error()
}

func (e AttachRejectedError) Unwrap() error {
	return e.Cause
}

// blobLifecycleService owns object-blob slot, lease, attach, finalization, and
// quarantine behavior. It has no access-handle or cleanup surface.
type blobLifecycleService struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess *admission.Checker
	revisionStore  revisionAppendPort
	projections    evidenceprojection.Rows
	supportEffects evidenceprojection.SupportProjectionEffectsTx
	collaboration  collaboration.RecordChangedAppender
	blobSlots      blobSlotRepository
	blobs          blobRepository
	evidenceRows   evidenceRecordRepository
	blobLifecycle  blobLifecycleRepository
	uploadLeases   uploadLeaseRepository
}

type blobLifecycleDependencies struct {
	Postgres       postgres.DB
	Revisions      *revisions.Appender
	Projections    evidenceprojection.Rows
	SupportEffects evidenceprojection.SupportProjectionEffectsTx
	Collaboration  collaboration.RecordChangedAppender
}

type blobSlotParams struct {
	ObjectBlobID      uuid.UUID
	IncidentID        uuid.UUID
	ActorUserID       uuid.UUID
	StorageKey        string
	ByteSize          int64
	FilenameHint      *string
	ContentTypeHint   *string
	ExpectedSHA256Hex *string
	TargetExpiresAt   time.Time
	PendingExpiresAt  time.Time
	UploadTarget      map[string]any
	AcceptedContract  map[string]any
	RequestHash       []byte
	ClientTxnID       string
	UploadLease       uploadLeaseCreateParams
}

type uploadLeaseCreateParams struct {
	LeaseID                uuid.UUID
	CapabilityHash         []byte
	IssuingUserID          uuid.UUID
	IssuingSessionID       uuid.UUID
	IssuedAt               time.Time
	ExpiresAt              time.Time
	RequiredMethod         string
	RequiredHeaders        map[string]string
	AcceptedContractSHA256 []byte
}

type uploadLeaseRecord struct {
	LeaseID                uuid.UUID
	ObjectBlobID           uuid.UUID
	IncidentID             uuid.UUID
	CapabilityHash         []byte
	IssuingUserID          uuid.UUID
	IssuingSessionID       uuid.UUID
	IssuedAt               time.Time
	ExpiresAt              time.Time
	RequiredMethod         string
	RequiredHeaders        map[string]string
	AcceptedContractSHA256 []byte
	LeaseState             string
	Blob                   blobRecord
}

type blobSlotResult struct {
	Payload    map[string]any
	StatusCode int
	Replayed   bool
}

type blobRecord struct {
	ObjectBlobID        uuid.UUID
	IncidentID          uuid.UUID
	StorageKey          string
	UploadState         string
	ByteSize            int64
	FilenameHint        *string
	ContentTypeHint     *string
	ExpectedSHA256Hex   *string
	ObservedSize        *int64
	ObservedContentType *string
	ObservedSHA256Hex   *string
	TargetExpiresAt     time.Time
	PendingExpiresAt    time.Time
	UploadLeaseState    string
}

type observedObject struct {
	Size        int64
	ContentType string
	SHA256Hex   string
}

type attachBlobResult struct {
	Payload               map[string]any
	StatusCode            int
	Replayed              bool
	IncidentID            uuid.UUID
	RecordID              uuid.UUID
	ChangeSetID           uuid.UUID
	ClientTxnID           string
	RowVersion            int64
	ChangedFieldKeys      []string
	AffectedRecordChanges []attachRecordChange
}

type attachBlobPreflightResult struct {
	Blob   blobRecord
	Replay *attachBlobResult
}

type attachRecordChange struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
	AffectedViews    []evidenceprojection.EvidenceAffectedViewChange
}

type quarantineBlobResult struct {
	IncidentID            uuid.UUID
	ObjectBlobID          uuid.UUID
	ChangeSetID           uuid.UUID
	ChangedEvidenceRows   []attachRecordChange
	ChangedEvidenceRecord int
}

func newBlobLifecycleService(dependencies blobLifecycleDependencies) (*blobLifecycleService, error) {
	if dependencies.Postgres == nil {
		return nil, errors.New("compose Evidence blob lifecycle: Postgres is required")
	}
	if dependencies.Revisions == nil {
		return nil, errors.New("compose Evidence blob lifecycle: Revisions is required")
	}
	if dependencies.Projections == nil {
		return nil, errors.New("compose Evidence blob lifecycle: Projections is required")
	}
	if dependencies.SupportEffects == nil {
		return nil, errors.New("compose Evidence blob lifecycle: support projection effects are required")
	}
	if dependencies.Collaboration == nil {
		return nil, errors.New("compose Evidence blob lifecycle: Collaboration is required")
	}
	return &blobLifecycleService{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: admission.NewChecker(dependencies.Postgres),
		revisionStore:  newRevisionAppendAdapter(dependencies.Revisions),
		projections:    dependencies.Projections,
		supportEffects: dependencies.SupportEffects,
		collaboration:  dependencies.Collaboration,
		blobSlots:      blobSlotRepository{},
		blobs:          blobRepository{db: dependencies.Postgres},
		evidenceRows:   evidenceRecordRepository{},
		blobLifecycle:  blobLifecycleRepository{db: dependencies.Postgres},
		uploadLeases:   uploadLeaseRepository{db: dependencies.Postgres},
	}, nil
}

func (s *blobLifecycleService) CreateBlobSlot(ctx context.Context, params blobSlotParams) (blobSlotResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey: blobCreateRouteKey, ActorUserID: params.ActorUserID,
		ScopeKey: params.IncidentID.String(), ClientTxnID: params.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, params.RequestHash) {
			return blobSlotResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return blobSlotResult{}, err
		}
		return blobSlotResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return blobSlotResult{}, err
	}

	payload := map[string]any{
		"incident_id":        params.IncidentID.String(),
		"object_blob_id":     params.ObjectBlobID.String(),
		"upload_state":       "pending",
		"target_expires_at":  formatHTTPTime(params.TargetExpiresAt),
		"pending_expires_at": formatHTTPTime(params.PendingExpiresAt),
		"upload_target":      params.UploadTarget,
		"accepted_contract":  params.AcceptedContract,
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return blobSlotResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ensureIncidentVisibleTx(ctx, tx, params.IncidentID); err != nil {
		return blobSlotResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, params.IncidentID); err != nil {
		return blobSlotResult{}, err
	}
	if err := s.blobSlots.insertTx(ctx, tx, params); err != nil {
		return blobSlotResult{}, err
	}
	if err := s.uploadLeases.insertTx(ctx, tx, params.ObjectBlobID, params.IncidentID, params.UploadLease); err != nil {
		return blobSlotResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, params.RequestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return blobSlotResult{}, authn.ErrClientTxnConflict
		}
		return blobSlotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blobSlotResult{}, err
	}
	return blobSlotResult{Payload: payload, StatusCode: http.StatusCreated}, nil
}

func (s *blobLifecycleService) GetBlob(ctx context.Context, objectBlobID uuid.UUID) (blobRecord, error) {
	return s.blobs.load(ctx, objectBlobID)
}

func (s *blobLifecycleService) GetUploadLease(ctx context.Context, leaseID uuid.UUID) (uploadLeaseRecord, error) {
	return s.uploadLeases.load(ctx, leaseID)
}

func (s *blobLifecycleService) ClaimUploadLease(ctx context.Context, leaseID uuid.UUID, capabilityHash []byte, now time.Time) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	lease, err := s.uploadLeases.loadForUpdateTx(ctx, tx, leaseID)
	if err != nil {
		return err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, lease.IncidentID); err != nil {
		return err
	}
	if !bytes.Equal(lease.CapabilityHash, capabilityHash) || lease.LeaseState != "issued" || !lease.ExpiresAt.After(now) ||
		lease.Blob.UploadState != "pending" || !lease.Blob.TargetExpiresAt.After(now) || !lease.Blob.PendingExpiresAt.After(now) {
		return errUploadLeaseUnavailable
	}
	if err := s.uploadLeases.claimTx(ctx, tx, leaseID, capabilityHash, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *blobLifecycleService) CompleteUploadLease(ctx context.Context, leaseID uuid.UUID, now time.Time) error {
	return s.uploadLeases.complete(ctx, leaseID, now)
}

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
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceQuarantined, Cause: ErrEvidenceQuarantined}
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
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
	}
	switch evidencepolicy.ClassifyBlobForAssociation(blob.UploadState, blob.PendingExpiresAt, now) {
	case evidencepolicy.AssociationBlobNeedsFinalization:
		if blob.UploadLeaseState != "completed" {
			return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobPending, Cause: ErrBlobNotAttachable}
		}
	case evidencepolicy.AssociationBlobAvailable:
	case evidencepolicy.AssociationBlobExpired, evidencepolicy.AssociationBlobFailed:
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobQuarantined:
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobQuarantined, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobInconsistent:
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
	}
	key := authn.RouteIdempotencyKey{
		RouteKey: blobAttachRouteKey, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	if existing, err := authn.GetRouteIdempotencyTx(ctx, tx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return attachBlobPreflightResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return attachBlobPreflightResult{}, err
		}
		replay := attachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: request.ClientTxnID}
		if err := tx.Commit(ctx); err != nil {
			return attachBlobPreflightResult{}, err
		}
		return attachBlobPreflightResult{Blob: blob, Replay: &replay}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return attachBlobPreflightResult{}, err
	}
	if associated {
		return attachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
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
	key := authn.RouteIdempotencyKey{
		RouteKey: blobAttachRouteKey, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return attachBlobResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return attachBlobResult{}, err
		}
		return attachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return attachBlobResult{}, err
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
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceQuarantined, Cause: ErrEvidenceQuarantined}
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
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
	}
	switch evidencepolicy.ClassifyBlobForAssociation(blob.UploadState, blob.PendingExpiresAt, now) {
	case evidencepolicy.AssociationBlobQuarantined:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobQuarantined, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobFailed:
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobExpired:
		if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "pending_timeout", now); err != nil {
			return attachBlobResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return attachBlobResult{}, err
		}
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobNeedsFinalization:
		if observed == nil {
			failed, err := s.blobLifecycle.recordFinalizeFailureTx(ctx, tx, request.ObjectBlobID, now)
			if err != nil {
				return attachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return attachBlobResult{}, err
			}
			reason := AttachReasonBlobPending
			if failed {
				reason = AttachReasonBlobFailed
			}
			return attachBlobResult{}, AttachRejectedError{ReasonCode: reason, Cause: ErrBlobNotAttachable}
		}
		if observed.Size != blob.ByteSize {
			if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "declared_size_mismatch", now); err != nil {
				return attachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return attachBlobResult{}, err
			}
			return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonAcceptedContractMismatch, Cause: ErrBlobNotAttachable}
		}
		if blob.ExpectedSHA256Hex != nil && observed.SHA256Hex != *blob.ExpectedSHA256Hex {
			if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "expected_sha256_mismatch", now); err != nil {
				return attachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return attachBlobResult{}, err
			}
			return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonAcceptedContractMismatch, Cause: ErrBlobNotAttachable}
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
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
	}
	if blob.UploadState != "available" {
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
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
		return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
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
			return attachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
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
	rowVersion, err := records.NewStore().AdvanceVersionTx(ctx, tx, recordID, actor.ID, now)
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
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return attachBlobResult{}, authn.ErrClientTxnConflict
		}
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

func (s *blobLifecycleService) QuarantineBlob(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID, trigger string, requestID string, now time.Time) (quarantineBlobResult, error) {
	if !evidencepolicy.ValidQuarantineEntryTrigger(trigger) {
		return quarantineBlobResult{}, ErrIllegalBlobTransition
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return quarantineBlobResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	blob, err := loadBlobForUpdateTx(ctx, tx, objectBlobID)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, blob.IncidentID); err != nil {
		return quarantineBlobResult{}, err
	}
	if !evidencepolicy.LegalBlobTransition(blob.UploadState, evidencepolicy.BlobQuarantined, trigger) {
		return quarantineBlobResult{}, ErrIllegalBlobTransition
	}

	rows, err := tx.Query(ctx, `
SELECT e.record_id
  FROM evidence e
  JOIN records r ON r.record_id = e.record_id
 WHERE e.object_blob_id = $1
   AND e.lifecycle_state IN ('available', 'released')
   AND r.deleted_at IS NULL
 ORDER BY e.record_id
 FOR UPDATE OF e, r
`, objectBlobID)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	recordIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var recordID uuid.UUID
		if err := rows.Scan(&recordID); err != nil {
			rows.Close()
			return quarantineBlobResult{}, err
		}
		recordIDs = append(recordIDs, recordID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return quarantineBlobResult{}, err
	}
	rows.Close()

	beforeRows := make(map[uuid.UUID]map[string]any, len(recordIDs))
	beforeSnapshots := make(map[uuid.UUID]revisions.RecordSnapshot, len(recordIDs))
	beforeVersions := make(map[uuid.UUID]int64, len(recordIDs))
	for _, recordID := range recordIDs {
		row, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		beforeRows[recordID] = row
		snapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		beforeSnapshots[recordID] = snapshot
		beforeVersions[recordID] = int64FromAny(row["row_version"])
	}

	_, err = tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'quarantined',
       updated_at = $2
 WHERE object_blob_id = $1
   AND upload_state = 'available'
`, objectBlobID, now.UTC())
	if err != nil {
		return quarantineBlobResult{}, err
	}

	var changeSetID uuid.UUID
	changedRows := make([]attachRecordChange, 0, len(recordIDs))
	if len(recordIDs) > 0 {
		reason := trigger
		requestIDPtr := &requestID
		if requestID == "" {
			requestIDPtr = nil
		}
		changeSetID, err = s.revisionStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
			IncidentID: blob.IncidentID, ActorUserID: actorUserID, Source: "evidence.blob.quarantine",
			Reason: &reason, RequestID: requestIDPtr, CreatedAt: now.UTC(),
		})
		if err != nil {
			return quarantineBlobResult{}, err
		}
	}

	for idx, recordID := range recordIDs {
		_, err := tx.Exec(ctx, `
UPDATE evidence
   SET lifecycle_state = 'quarantined',
       upload_state = 'quarantined',
       updated_at = $2
 WHERE record_id = $1
`, recordID, now.UTC())
		if err != nil {
			return quarantineBlobResult{}, err
		}
		if err := insertEvidenceCustodyEventTx(ctx, tx, evidenceCustodyEventParams{
			IncidentID:       blob.IncidentID,
			EvidenceRecordID: recordID,
			CustodyEventType: "quarantined",
			ActorUserID:      &actorUserID,
			OccurredAt:       now.UTC(),
			Metadata:         map[string]any{"object_blob_id": objectBlobID.String(), "trigger": trigger},
		}); err != nil {
			return quarantineBlobResult{}, err
		}
		rowVersion, err := records.NewStore().AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		if err := s.projections.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
			return quarantineBlobResult{}, err
		}
		afterRow, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		afterSnapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		beforeVersionID := fmt.Sprintf("%s:%d", recordID, beforeVersions[recordID])
		afterVersionID := fmt.Sprintf("%s:%d", recordID, rowVersion)
		beforeSnapshot := beforeSnapshots[recordID]
		if err := s.revisionStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
			ChangeSetID: changeSetID, SequenceNo: idx + 1, TargetKind: "record", RecordID: recordID,
			OperationKind: "patch", BeforeVersionID: &beforeVersionID, AfterVersionID: &afterVersionID,
			BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
		}); err != nil {
			return quarantineBlobResult{}, err
		}
		changedFieldKeys := sortedChangedKeys(beforeRows[recordID], afterRow)
		if err := s.revisionStore.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
			ChangeSetID: changeSetID, RecordID: recordID, RowVersion: rowVersion,
			BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
			ConflictFacts: evidenceRevisionFacts(beforeRows[recordID], afterRow, changedFieldKeys),
		}); err != nil {
			return quarantineBlobResult{}, err
		}
		projectionChanges, err := s.refreshEvidenceSupportProjectionsTx(ctx, tx, blob.IncidentID, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		primaryChange := attachRecordChange{
			RecordID: recordID, RowVersion: rowVersion, ViewSchemaID: ViewSchemaID,
			ChangedFieldKeys: changedFieldKeys,
		}
		if err := appendEvidenceRecordChangeIntentsTx(
			ctx,
			tx,
			s.collaboration,
			blob.IncidentID,
			actorUserID,
			changeSetID.String(),
			changeSetID,
			primaryChange,
			afterRow,
			projectionChanges,
			now,
		); err != nil {
			return quarantineBlobResult{}, err
		}
		changedRows = append(changedRows, primaryChange)
		changedRows = append(changedRows, projectionChanges...)
	}

	if err := tx.Commit(ctx); err != nil {
		return quarantineBlobResult{}, err
	}
	return quarantineBlobResult{
		IncidentID: blob.IncidentID, ObjectBlobID: objectBlobID, ChangeSetID: changeSetID,
		ChangedEvidenceRows: changedRows, ChangedEvidenceRecord: len(recordIDs),
	}, nil
}

type evidenceCustodyEventParams struct {
	IncidentID       uuid.UUID
	EvidenceRecordID uuid.UUID
	CustodyEventType string
	ActorUserID      *uuid.UUID
	OccurredAt       time.Time
	LocationText     *string
	Note             *string
	Metadata         map[string]any
}

func insertEvidenceCustodyEventTx(ctx context.Context, tx pgx.Tx, params evidenceCustodyEventParams) error {
	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO evidence_custody_events (
    incident_id,
    evidence_record_id,
    custody_event_type,
    actor_user_id,
    occurred_at,
    location_text,
    note,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
`, params.IncidentID, params.EvidenceRecordID, params.CustodyEventType, params.ActorUserID, params.OccurredAt.UTC(), params.LocationText, params.Note, metadataJSON)
	return err
}

type evidenceMeta struct {
	IncidentID uuid.UUID
	RowVersion int64
}

type rowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *rowVersionConflictError) Error() string { return ErrRowVersionConflict.Error() }
func (e *rowVersionConflictError) Unwrap() error { return ErrRowVersionConflict }

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
		return ErrIllegalBlobTransition
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
		return ErrIllegalBlobTransition
	}
	return nil
}

func evidenceRowCellValue(row map[string]any, fieldKey string) any {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return nil
	}
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil
	}
	return cell["value"]
}

func int64FromAny(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func (s *blobLifecycleService) refreshEvidenceSupportProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) ([]attachRecordChange, error) {
	return refreshEvidenceSupportProjectionsTx(ctx, tx, s.supportEffects, incidentID, evidenceRecordID)
}

func refreshEvidenceSupportProjectionsTx(
	ctx context.Context,
	tx pgx.Tx,
	effects evidenceprojection.SupportProjectionEffectsTx,
	incidentID uuid.UUID,
	evidenceRecordID uuid.UUID,
) ([]attachRecordChange, error) {
	subjects, err := loadEvidenceAssociationSubjectsTx(ctx, tx, incidentID, evidenceRecordID)
	if err != nil {
		return nil, err
	}
	if len(subjects) == 0 {
		return nil, nil
	}
	result, err := effects.RefreshEvidenceAssociationEffects(ctx, tx, evidenceprojection.EvidenceAssociationEffectsInput{
		IncidentID: incidentID,
		Subjects:   subjects,
	})
	if err != nil {
		return nil, err
	}
	return attachRecordChangesFromSupportEffects(result), nil
}

func loadEvidenceAssociationSubjectsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) ([]evidenceprojection.EvidenceAssociationSubject, error) {
	rows, err := tx.Query(ctx, `
SELECT r.record_id, r.record_type
  FROM active_record_links_v1 rl
  JOIN records r
    ON r.incident_id = rl.incident_id
   AND r.record_id = rl.src_record_id
   AND r.deleted_at IS NULL
 WHERE rl.incident_id = $1
   AND rl.dst_record_id = $2
   AND rl.link_type = 'attached_evidence'
 ORDER BY r.record_id, r.record_type
`, incidentID, evidenceRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subjects := make([]evidenceprojection.EvidenceAssociationSubject, 0)
	for rows.Next() {
		var subject evidenceprojection.EvidenceAssociationSubject
		if err := rows.Scan(&subject.RecordID, &subject.RecordType); err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}
	return subjects, rows.Err()
}

func attachRecordChangesFromSupportEffects(result evidenceprojection.EvidenceAssociationEffectsResult) []attachRecordChange {
	changes := make([]attachRecordChange, 0, len(result.Changes))
	for _, effect := range result.Changes {
		changedFieldKeys := make([]string, 0)
		for _, view := range effect.AffectedViews {
			changedFieldKeys = append(changedFieldKeys, view.ChangedFieldKeys...)
		}
		slices.Sort(changedFieldKeys)
		changedFieldKeys = slices.Compact(changedFieldKeys)
		changes = append(changes, attachRecordChange{
			RecordID:         effect.RecordID,
			RowVersion:       effect.RowVersion,
			ChangedFieldKeys: changedFieldKeys,
			AffectedViews:    append([]evidenceprojection.EvidenceAffectedViewChange(nil), effect.AffectedViews...),
		})
	}
	return changes
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

func decodeStoredPayload(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
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

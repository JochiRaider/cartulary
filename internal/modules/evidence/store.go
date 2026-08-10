package evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
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

type Store struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidents.Access
	revisionStore  revisionAppendPort
	projections    evidenceprojection.Rows
	collaboration  collaboration.IntentAppender
	source         evidenceSourceKernel
	blobSlots      blobSlotRepository
	blobs          blobRepository
	evidenceRows   evidenceRecordRepository
	blobLifecycle  blobLifecycleRepository
	accessHandles  accessHandleRepository
}

type StoreOption func(*Store)

func WithWorkbookProjections(rows evidenceprojection.Rows) StoreOption {
	return func(store *Store) {
		store.projections = rows
		store.source.projections = rows
	}
}

func WithCollaborationIntents(appender collaboration.IntentAppender) StoreOption {
	return func(store *Store) {
		store.collaboration = appender
	}
}

func WithRevisionAppender(appender *revisions.Appender) StoreOption {
	return func(store *Store) {
		store.revisionStore = newRevisionAppendAdapter(appender)
	}
}

type BlobSlotParams struct {
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
}

type BlobSlotResult struct {
	Payload    map[string]any
	StatusCode int
	Replayed   bool
}

type BlobRecord struct {
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
}

type ObservedObject struct {
	Size        int64
	ContentType string
	SHA256Hex   string
}

type AttachBlobResult struct {
	Payload               map[string]any
	StatusCode            int
	Replayed              bool
	IncidentID            uuid.UUID
	RecordID              uuid.UUID
	ChangeSetID           uuid.UUID
	ClientTxnID           string
	RowVersion            int64
	ChangedFieldKeys      []string
	AffectedRecordChanges []AttachRecordChange
}

type AttachRecordChange struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
}

type QuarantineBlobResult struct {
	IncidentID            uuid.UUID
	ObjectBlobID          uuid.UUID
	ChangeSetID           uuid.UUID
	ChangedEvidenceRows   []AttachRecordChange
	ChangedEvidenceRecord int
}

type CleanupFailedBlobResult struct {
	ExpiredPendingCount int
	CleanedBlobCount    int
}

type EvidenceAccessRecord struct {
	IncidentID             uuid.UUID
	RecordID               uuid.UUID
	RecordRowVersion       int64
	ObjectBlobID           *uuid.UUID
	BlobMetadataVisible    bool
	StorageKey             *string
	EvidenceLifecycleState string
	UploadState            string
	FilenameSource         string
	ContentType            string
	SizeBytes              int64
	SHA256                 *string
	MediaClass             string
	PreviewKind            *string
}

type HandleRecord struct {
	Token                  string
	IncidentID             uuid.UUID
	RecordID               uuid.UUID
	RecordRowVersion       int64
	ObjectBlobID           uuid.UUID
	StorageKey             string
	SessionID              uuid.UUID
	HandleKind             string
	MediaClass             string
	PreviewKind            *string
	Disposition            string
	Filename               string
	ContentType            string
	SizeBytes              int64
	SHA256                 *string
	EvidenceLifecycleState string
	UploadState            string
	ExpiresAt              time.Time
	ConsumedAt             *time.Time
}

func NewStore(pool postgres.DB, options ...StoreOption) *Store {
	projectionRows := missingProjectionRows{}
	store := &Store{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		incidentAccess: incidents.NewAccess(pool),
		projections:    projectionRows,
		blobSlots:      blobSlotRepository{},
		blobs:          blobRepository{db: pool},
		evidenceRows:   evidenceRecordRepository{},
		blobLifecycle:  blobLifecycleRepository{db: pool},
		accessHandles:  accessHandleRepository{db: pool},
	}
	store.source = evidenceSourceKernel{
		records:     records.NewStore(),
		rows:        store,
		projections: projectionRows,
	}
	for _, option := range options {
		if option != nil {
			option(store)
		}
	}
	return store
}

func (s *Store) CreateBlobSlot(ctx context.Context, params BlobSlotParams) (BlobSlotResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey: blobCreateRouteKey, ActorUserID: params.ActorUserID,
		ScopeKey: params.IncidentID.String(), ClientTxnID: params.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, params.RequestHash) {
			return BlobSlotResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return BlobSlotResult{}, err
		}
		return BlobSlotResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return BlobSlotResult{}, err
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
		return BlobSlotResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ensureIncidentVisibleTx(ctx, tx, params.IncidentID); err != nil {
		return BlobSlotResult{}, err
	}
	if err := s.blobSlots.insertTx(ctx, tx, params); err != nil {
		return BlobSlotResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, params.RequestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return BlobSlotResult{}, authn.ErrClientTxnConflict
		}
		return BlobSlotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return BlobSlotResult{}, err
	}
	return BlobSlotResult{Payload: payload, StatusCode: http.StatusCreated}, nil
}

func (s *Store) GetBlob(ctx context.Context, objectBlobID uuid.UUID) (BlobRecord, error) {
	return s.blobs.load(ctx, objectBlobID)
}

func (s *Store) AttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request AttachBlobRequest, requestHash []byte, observed *ObservedObject, requestID string, now time.Time) (AttachBlobResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey: blobAttachRouteKey, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AttachBlobResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return AttachBlobResult{}, err
		}
		return AttachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return AttachBlobResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AttachBlobResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := s.evidenceRows.loadForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return AttachBlobResult{}, err
	}
	if meta.RowVersion != request.BaseRowVersion {
		return AttachBlobResult{}, &rowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
	}
	beforeRow, err := loadEvidenceRowTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	beforeSnapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	if evidenceRowCellValue(beforeRow, "evidence.lifecycle_state") == "quarantined" {
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceQuarantined, Cause: ErrEvidenceQuarantined}
	}
	blob, err := s.blobs.loadForUpdateTx(ctx, tx, request.ObjectBlobID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	if blob.IncidentID != meta.IncidentID {
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrIncidentMismatch}
	}
	associated, err := isBlobAssociatedTx(ctx, tx, request.ObjectBlobID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	if associated {
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
	}
	if blob.UploadState == "quarantined" {
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobQuarantined, Cause: ErrBlobNotAttachable}
	}
	if blob.UploadState == "failed" {
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	}
	if blob.UploadState == "pending" {
		if now.After(blob.PendingExpiresAt) {
			if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "pending_timeout", now); err != nil {
				return AttachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return AttachBlobResult{}, err
			}
			return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
		}
		if observed == nil {
			failed, err := s.blobLifecycle.recordFinalizeFailureTx(ctx, tx, request.ObjectBlobID, now)
			if err != nil {
				return AttachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return AttachBlobResult{}, err
			}
			reason := AttachReasonBlobPending
			if failed {
				reason = AttachReasonBlobFailed
			}
			return AttachBlobResult{}, AttachRejectedError{ReasonCode: reason, Cause: ErrBlobNotAttachable}
		}
		if observed.Size != blob.ByteSize {
			if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "declared_size_mismatch", now); err != nil {
				return AttachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return AttachBlobResult{}, err
			}
			return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonAcceptedContractMismatch, Cause: ErrBlobNotAttachable}
		}
		if blob.ExpectedSHA256Hex != nil && observed.SHA256Hex != *blob.ExpectedSHA256Hex {
			if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "expected_sha256_mismatch", now); err != nil {
				return AttachBlobResult{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return AttachBlobResult{}, err
			}
			return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonAcceptedContractMismatch, Cause: ErrBlobNotAttachable}
		}
		if err := s.blobLifecycle.markAvailableTx(ctx, tx, request.ObjectBlobID, observed, now); err != nil {
			return AttachBlobResult{}, err
		}
		blob.UploadState = "available"
		blob.ObservedSize = &observed.Size
		blob.ObservedContentType = &observed.ContentType
		blob.ObservedSHA256Hex = &observed.SHA256Hex
	}
	if blob.UploadState != "available" {
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
	}
	sha := blob.ObservedSHA256Hex
	if sha == nil && observed != nil {
		sha = &observed.SHA256Hex
	}
	storageRef, err := blobref.ObjectBlobStorageRef(request.ObjectBlobID)
	if err != nil {
		return AttachBlobResult{}, err
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
			return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
		}
		return AttachBlobResult{}, err
	}
	if err := insertEvidenceCustodyEventTx(ctx, tx, evidenceCustodyEventParams{
		IncidentID:       meta.IncidentID,
		EvidenceRecordID: recordID,
		CustodyEventType: "made_available",
		ActorUserID:      &actor.ID,
		OccurredAt:       now.UTC(),
		Metadata:         map[string]any{"object_blob_id": request.ObjectBlobID.String()},
	}); err != nil {
		return AttachBlobResult{}, err
	}
	rowVersion, err := records.NewStore().AdvanceVersionTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return AttachBlobResult{}, err
	}
	if err := s.projections.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
		return AttachBlobResult{}, err
	}
	afterRow, err := loadEvidenceRowTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	afterSnapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	changeSetID, err := s.revisionStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: meta.IncidentID, ActorUserID: actor.ID, Source: blobAttachRouteKey,
		ClientTxnID: &request.ClientTxnID, RequestID: &requestID, CreatedAt: now.UTC(),
	})
	if err != nil {
		return AttachBlobResult{}, err
	}
	beforeVersionID := fmt.Sprintf("%s:%d", recordID, request.BaseRowVersion)
	afterVersionID := fmt.Sprintf("%s:%d", recordID, rowVersion)
	if err := s.revisionStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "record", RecordID: recordID,
		OperationKind: "patch", BeforeVersionID: &beforeVersionID, AfterVersionID: &afterVersionID,
		BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
	}); err != nil {
		return AttachBlobResult{}, err
	}
	if err := s.revisionStore.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID, RecordID: recordID, RowVersion: rowVersion,
		BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
		LiveChange: revisions.LiveRecordChange{BeforeValue: beforeRow, AfterValue: afterRow},
	}); err != nil {
		return AttachBlobResult{}, err
	}
	affectedChanges, err := s.refreshEvidenceSupportProjectionsTx(ctx, tx, meta.IncidentID, recordID)
	if err != nil {
		return AttachBlobResult{}, err
	}
	changedFieldKeys := sortedChangedKeys(beforeRow, afterRow)
	if err := appendEvidenceRecordChangeIntentsTx(
		ctx,
		tx,
		s.collaboration,
		meta.IncidentID,
		actor.ID,
		request.ClientTxnID,
		changeSetID,
		AttachRecordChange{
			RecordID:         recordID,
			RowVersion:       rowVersion,
			ViewSchemaID:     ViewSchemaID,
			ChangedFieldKeys: changedFieldKeys,
		},
		afterRow,
		affectedChanges,
		now,
	); err != nil {
		return AttachBlobResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": ViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"object_blob_id": request.ObjectBlobID.String(),
		"row":            afterRow,
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return AttachBlobResult{}, authn.ErrClientTxnConflict
		}
		return AttachBlobResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AttachBlobResult{}, err
	}
	return AttachBlobResult{
		Payload: payload, StatusCode: http.StatusOK, IncidentID: meta.IncidentID, RecordID: recordID,
		ChangeSetID: changeSetID, ClientTxnID: request.ClientTxnID, RowVersion: rowVersion,
		ChangedFieldKeys: changedFieldKeys, AffectedRecordChanges: affectedChanges,
	}, nil
}

func (s *Store) QuarantineBlob(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID, trigger string, requestID string, now time.Time) (QuarantineBlobResult, error) {
	if trigger != "content_inspection_quarantine" && trigger != "admin_quarantine" {
		return QuarantineBlobResult{}, ErrIllegalBlobTransition
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return QuarantineBlobResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	blob, err := loadBlobForUpdateTx(ctx, tx, objectBlobID)
	if err != nil {
		return QuarantineBlobResult{}, err
	}
	if blob.UploadState != "available" {
		return QuarantineBlobResult{}, ErrIllegalBlobTransition
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
		return QuarantineBlobResult{}, err
	}
	recordIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var recordID uuid.UUID
		if err := rows.Scan(&recordID); err != nil {
			rows.Close()
			return QuarantineBlobResult{}, err
		}
		recordIDs = append(recordIDs, recordID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return QuarantineBlobResult{}, err
	}
	rows.Close()

	beforeRows := make(map[uuid.UUID]map[string]any, len(recordIDs))
	beforeSnapshots := make(map[uuid.UUID]revisions.RecordSnapshot, len(recordIDs))
	beforeVersions := make(map[uuid.UUID]int64, len(recordIDs))
	for _, recordID := range recordIDs {
		row, err := loadEvidenceRowTx(ctx, tx, recordID)
		if err != nil {
			return QuarantineBlobResult{}, err
		}
		beforeRows[recordID] = row
		snapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return QuarantineBlobResult{}, err
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
		return QuarantineBlobResult{}, err
	}

	var changeSetID uuid.UUID
	changedRows := make([]AttachRecordChange, 0, len(recordIDs))
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
			return QuarantineBlobResult{}, err
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
			return QuarantineBlobResult{}, err
		}
		if err := insertEvidenceCustodyEventTx(ctx, tx, evidenceCustodyEventParams{
			IncidentID:       blob.IncidentID,
			EvidenceRecordID: recordID,
			CustodyEventType: "quarantined",
			ActorUserID:      &actorUserID,
			OccurredAt:       now.UTC(),
			Metadata:         map[string]any{"object_blob_id": objectBlobID.String(), "trigger": trigger},
		}); err != nil {
			return QuarantineBlobResult{}, err
		}
		rowVersion, err := records.NewStore().AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
		if err != nil {
			return QuarantineBlobResult{}, err
		}
		if err := s.projections.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
			return QuarantineBlobResult{}, err
		}
		afterRow, err := loadEvidenceRowTx(ctx, tx, recordID)
		if err != nil {
			return QuarantineBlobResult{}, err
		}
		afterSnapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return QuarantineBlobResult{}, err
		}
		beforeVersionID := fmt.Sprintf("%s:%d", recordID, beforeVersions[recordID])
		afterVersionID := fmt.Sprintf("%s:%d", recordID, rowVersion)
		beforeSnapshot := beforeSnapshots[recordID]
		if err := s.revisionStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
			ChangeSetID: changeSetID, SequenceNo: idx + 1, TargetKind: "record", RecordID: recordID,
			OperationKind: "patch", BeforeVersionID: &beforeVersionID, AfterVersionID: &afterVersionID,
			BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
		}); err != nil {
			return QuarantineBlobResult{}, err
		}
		if err := s.revisionStore.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
			ChangeSetID: changeSetID, RecordID: recordID, RowVersion: rowVersion,
			BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
			LiveChange: revisions.LiveRecordChange{BeforeValue: beforeRows[recordID], AfterValue: afterRow},
		}); err != nil {
			return QuarantineBlobResult{}, err
		}
		projectionChanges, err := s.refreshEvidenceSupportProjectionsTx(ctx, tx, blob.IncidentID, recordID)
		if err != nil {
			return QuarantineBlobResult{}, err
		}
		primaryChange := AttachRecordChange{
			RecordID: recordID, RowVersion: rowVersion, ViewSchemaID: ViewSchemaID,
			ChangedFieldKeys: sortedChangedKeys(beforeRows[recordID], afterRow),
		}
		if err := appendEvidenceRecordChangeIntentsTx(
			ctx,
			tx,
			s.collaboration,
			blob.IncidentID,
			actorUserID,
			"",
			changeSetID,
			primaryChange,
			afterRow,
			projectionChanges,
			now,
		); err != nil {
			return QuarantineBlobResult{}, err
		}
		changedRows = append(changedRows, primaryChange)
		changedRows = append(changedRows, projectionChanges...)
	}

	if err := tx.Commit(ctx); err != nil {
		return QuarantineBlobResult{}, err
	}
	return QuarantineBlobResult{
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

func (s *Store) CleanupFailedUnattachedBlobBytes(ctx context.Context, objectStore objectstore.Store, now time.Time, limit int) (CleanupFailedBlobResult, error) {
	if limit <= 0 {
		limit = 100
	}
	expiredCount, err := s.blobLifecycle.markExpiredPending(ctx, now)
	if err != nil {
		return CleanupFailedBlobResult{}, err
	}
	candidates, err := s.blobLifecycle.failedUnattachedCleanupCandidates(ctx, now, limit)
	if err != nil {
		return CleanupFailedBlobResult{}, err
	}
	cleaned := 0
	for _, candidate := range candidates {
		if err := objectStore.DeleteObject(ctx, candidate.StorageKey); err != nil {
			return CleanupFailedBlobResult{}, err
		}
		tag, err := s.pool.Exec(ctx, `
UPDATE object_blobs b
   SET cleaned_up_at = $2,
       updated_at = $2
 WHERE b.object_blob_id = $1
   AND b.upload_state = 'failed'
   AND b.cleaned_up_at IS NULL
   AND NOT EXISTS (
       SELECT 1
         FROM evidence e
        WHERE e.object_blob_id = b.object_blob_id
   )
`, candidate.ObjectBlobID, now.UTC())
		if err != nil {
			return CleanupFailedBlobResult{}, err
		}
		if tag.RowsAffected() > 0 {
			cleaned++
		}
	}
	return CleanupFailedBlobResult{ExpiredPendingCount: expiredCount, CleanedBlobCount: cleaned}, nil
}

func (s *Store) LoadEvidenceAccess(ctx context.Context, recordID uuid.UUID) (EvidenceAccessRecord, error) {
	return s.accessHandles.loadEvidence(ctx, recordID)
}

func classifyEvidenceAccess(access EvidenceAccessRecord, boundObjectBlobID *uuid.UUID) string {
	if access.ObjectBlobID == nil {
		return "no_visible_blob"
	}
	if boundObjectBlobID != nil && *access.ObjectBlobID != *boundObjectBlobID {
		return "evidence_inconsistent"
	}
	if !access.BlobMetadataVisible {
		return "blob_missing"
	}
	if access.EvidenceLifecycleState == "quarantined" || access.UploadState == "quarantined" {
		return "evidence_quarantined"
	}
	switch access.UploadState {
	case "pending":
		return "blob_pending"
	case "failed":
		return "blob_failed"
	}
	if (access.EvidenceLifecycleState != "available" && access.EvidenceLifecycleState != "released") || access.UploadState != "available" {
		return "evidence_inconsistent"
	}
	return ""
}

func (s *Store) InsertHandle(ctx context.Context, handle HandleRecord, issuedByUserID uuid.UUID) error {
	return s.accessHandles.insert(ctx, handle, issuedByUserID)
}

func (s *Store) LoadHandle(ctx context.Context, token string) (HandleRecord, error) {
	return s.accessHandles.load(ctx, token)
}

func (s *Store) ConsumeDownloadHandle(ctx context.Context, token string, now time.Time) error {
	return s.accessHandles.consumeDownload(ctx, token, now)
}

func (s *Store) CheckHandleAccess(ctx context.Context, handle HandleRecord) (string, error) {
	return s.accessHandles.checkCurrent(ctx, handle)
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

func loadBlobForUpdateTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID) (BlobRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT object_blob_id, incident_id, storage_key, upload_state, byte_size,
       filename_hint, content_type_hint, expected_sha256_hex,
       observed_size, observed_content_type, observed_sha256_hex,
       target_expires_at, pending_expires_at
  FROM object_blobs
 WHERE object_blob_id = $1
 FOR UPDATE
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

func failBlobTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, reason string, now time.Time) error {
	_, err := tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'failed',
       terminal_reason = $2,
       failed_at = $3,
       cleanup_due_at = $3::timestamptz + interval '1 hour',
       updated_at = $3
 WHERE object_blob_id = $1
`, objectBlobID, reason, now.UTC())
	return err
}

func recordNonTerminalFinalizeFailureTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, now time.Time) (bool, error) {
	var uploadState string
	err := tx.QueryRow(ctx, `
UPDATE object_blobs
   SET finalize_attempt_count = finalize_attempt_count + 1,
       upload_state = CASE WHEN finalize_attempt_count + 1 >= 4 THEN 'failed' ELSE upload_state END,
       terminal_reason = CASE WHEN finalize_attempt_count + 1 >= 4 THEN 'finalize_retry_exhausted' ELSE terminal_reason END,
       failed_at = CASE WHEN finalize_attempt_count + 1 >= 4 THEN $2 ELSE failed_at END,
       cleanup_due_at = CASE WHEN finalize_attempt_count + 1 >= 4 THEN $2::timestamptz + interval '1 hour' ELSE cleanup_due_at END,
       updated_at = $2
 WHERE object_blob_id = $1
   AND upload_state = 'pending'
 RETURNING upload_state
`, objectBlobID, now.UTC()).Scan(&uploadState)
	return uploadState == "failed", err
}

func markBlobAvailableTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, observed *ObservedObject, now time.Time) error {
	_, err := tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'available',
       observed_size = $2,
       observed_content_type = $3,
       observed_sha256_hex = $4,
       finalized_at = $5,
       updated_at = $5
 WHERE object_blob_id = $1
`, objectBlobID, observed.Size, observed.ContentType, observed.SHA256Hex, now.UTC())
	return err
}

type cleanupCandidate struct {
	ObjectBlobID uuid.UUID
	StorageKey   string
}

func loadEvidenceRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	row := tx.QueryRow(ctx, `
SELECT e.record_id, r.row_version,
       e.title, e.lifecycle_state, e.requested_at, e.received_at, e.storage_ref, e.blob_hash,
       e.collector_party_text, e.collector_party_id, e.source_party_text, e.source_party_id,
       COALESCE(b.upload_state, e.upload_state), 0::integer, e.updated_at
  FROM evidence e
  JOIN records r ON r.record_id = e.record_id
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
   AND r.deleted_at IS NULL
`, recordID)
	var data evidenceRowData
	if err := row.Scan(
		&data.RecordID,
		&data.RowVersion,
		&data.Title,
		&data.LifecycleState,
		&data.RequestedAt,
		&data.ReceivedAt,
		&data.StorageRef,
		&data.BlobHash,
		&data.CollectorPartyText,
		&data.CollectorPartyID,
		&data.SourcePartyText,
		&data.SourcePartyID,
		&data.UploadState,
		&data.LinkedRecordCount,
		&data.EditedAt,
	); err != nil {
		return nil, err
	}
	return evidenceRowFromData(data)
}

type evidenceRowData struct {
	RecordID           any
	RowVersion         any
	Title              any
	LifecycleState     any
	RequestedAt        any
	ReceivedAt         any
	StorageRef         any
	BlobHash           any
	CollectorPartyText any
	CollectorPartyID   any
	SourcePartyText    any
	SourcePartyID      any
	UploadState        any
	LinkedRecordCount  any
	EditedAt           any
}

type evidenceFieldCell struct {
	key   string
	value any
}

func (row evidenceRowData) fieldCells() []evidenceFieldCell {
	return []evidenceFieldCell{
		{key: "evidence.title", value: row.Title},
		{key: "evidence.lifecycle_state", value: row.LifecycleState},
		{key: "evidence.requested_at", value: row.RequestedAt},
		{key: "evidence.received_at", value: row.ReceivedAt},
		{key: "evidence.storage_ref", value: row.StorageRef},
		{key: "evidence.blob_hash", value: row.BlobHash},
		{key: "evidence.collector_party_text", value: row.CollectorPartyText},
		{key: "evidence.collector_party_id", value: row.CollectorPartyID},
		{key: "evidence.source_party_text", value: row.SourcePartyText},
		{key: "evidence.source_party_id", value: row.SourcePartyID},
		{key: "evidence.upload_state", value: row.UploadState},
		{key: "evidence.linked_record_count", value: row.LinkedRecordCount},
		{key: "evidence.edited_at", value: row.EditedAt},
	}
}

func evidenceRowFromData(data evidenceRowData) (map[string]any, error) {
	recordID, err := uuidFromDBValue(data.RecordID)
	if err != nil {
		return nil, err
	}
	fieldCells := data.fieldCells()
	cells := make(map[string]any, len(fieldCells))
	for _, cell := range fieldCells {
		cells[cell.key] = map[string]any{"value": publicCellValue(cell.value)}
	}
	return map[string]any{
		"record_id":   recordID.String(),
		"row_version": data.RowVersion,
		"cells":       cells,
	}, nil
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

func uuidFromDBValue(value any) (uuid.UUID, error) {
	switch typed := value.(type) {
	case uuid.UUID:
		return typed, nil
	case [16]byte:
		return uuid.UUID(typed), nil
	case []byte:
		if len(typed) == 16 {
			return uuid.FromBytes(typed)
		}
		return uuid.Parse(string(typed))
	case string:
		return uuid.Parse(typed)
	default:
		return uuid.UUID{}, fmt.Errorf("evidence row record_id was %T", value)
	}
}

func publicCellValue(value any) any {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano)
	case uuid.UUID:
		return typed.String()
	case []byte:
		if len(typed) == 16 {
			if id, err := uuid.FromBytes(typed); err == nil {
				return id.String()
			}
		}
		return string(typed)
	default:
		return typed
	}
}

func (s *Store) refreshEvidenceSupportProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) ([]AttachRecordChange, error) {
	changes, err := loadEvidenceSupportRecordChangesTx(ctx, tx, incidentID, evidenceRecordID)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, nil
	}
	if err := s.projections.RefreshEvidenceSupportTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}
	return changes, nil
}

type missingProjectionRows struct{}

func (missingProjectionRows) RefreshEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error {
	return errors.New("evidence projection rows are required")
}

func (missingProjectionRows) LoadEvidenceTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return nil, errors.New("evidence projection rows are required")
}

func (missingProjectionRows) RefreshEvidenceSupportTx(context.Context, pgx.Tx, uuid.UUID) error {
	return errors.New("evidence projection rows are required")
}

func (missingProjectionRows) RebuildEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error {
	return errors.New("evidence projection rows are required")
}

func loadEvidenceSupportRecordChangesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) ([]AttachRecordChange, error) {
	rows, err := tx.Query(ctx, `
SELECT r.record_id, r.row_version, r.record_type
  FROM active_record_links_v1 rl
  JOIN records r
    ON r.incident_id = rl.incident_id
   AND r.record_id = rl.src_record_id
   AND r.deleted_at IS NULL
 WHERE rl.incident_id = $1
   AND rl.dst_record_id = $2
   AND rl.link_type = 'attached_evidence'
   AND r.record_type IN ('timeline_event', 'host', 'identity')
 ORDER BY r.record_id
`, incidentID, evidenceRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var changes []AttachRecordChange
	for rows.Next() {
		var change AttachRecordChange
		var recordType string
		if err := rows.Scan(&change.RecordID, &change.RowVersion, &recordType); err != nil {
			return nil, err
		}
		switch recordType {
		case "timeline_event":
			change.ViewSchemaID = "cartulary.view.timeline.v2"
			change.ChangedFieldKeys = []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"}
		case "host":
			change.ViewSchemaID = "cartulary.view.hosts.v1"
			change.ChangedFieldKeys = []string{"host.evidence_count"}
		case "identity":
			change.ViewSchemaID = "cartulary.view.identities.v1"
			change.ChangedFieldKeys = []string{"identity.evidence_count"}
		default:
			continue
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
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

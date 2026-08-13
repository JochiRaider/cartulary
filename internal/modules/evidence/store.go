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
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
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
	ErrUploadLeaseNotFound    = errors.New("evidence: upload lease not found")
	ErrUploadLeaseUnavailable = errors.New("evidence: upload lease unavailable")
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

// BlobLifecycleService owns object-blob slot, lease, attach, finalization, and
// quarantine behavior. It has no access-handle or cleanup surface.
type BlobLifecycleService struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidents.Access
	revisionStore  revisionAppendPort
	projections    evidenceprojection.Rows
	supportEffects evidenceprojection.SupportProjectionEffectsTx
	collaboration  collaboration.IntentAppender
	blobSlots      blobSlotRepository
	blobs          blobRepository
	evidenceRows   evidenceRecordRepository
	blobLifecycle  blobLifecycleRepository
	uploadLeases   uploadLeaseRepository
}

type BlobLifecycleDependencies struct {
	Postgres       postgres.DB
	Revisions      *revisions.Appender
	Projections    evidenceprojection.Rows
	SupportEffects evidenceprojection.SupportProjectionEffectsTx
	Collaboration  collaboration.IntentAppender
}

// AccessHandleService owns Evidence access snapshots and opaque handle state.
// It has no blob mutation, projection, revision, or object-store dependency.
type AccessHandleService struct {
	accessHandles accessHandleRepository
}

// CleanupService owns durable failed-blob claim coordination. It remains inert
// until the private dispatcher is activated by application assembly in S09.
type CleanupService struct {
	pool          postgres.DB
	blobLifecycle blobLifecycleRepository
}

// RouteOperations is the narrow transport facade composed from the blob and
// access capabilities. It does not expose cleanup or source mutation.
type RouteOperations struct {
	*BlobLifecycleService
	*AccessHandleService
}

type SourceMutationService struct {
	source    evidenceSourceKernel
	mutations evidenceSourceMutationKernel
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
	UploadLease       UploadLeaseCreateParams
}

type UploadLeaseCreateParams struct {
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

type UploadLeaseRecord struct {
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
	Blob                   BlobRecord
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
	UploadLeaseState    string
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

type AttachBlobPreflightResult struct {
	Blob   BlobRecord
	Replay *AttachBlobResult
}

type AttachRecordChange struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
	AffectedViews    []evidenceprojection.EvidenceAffectedViewChange
}

type QuarantineBlobResult struct {
	IncidentID            uuid.UUID
	ObjectBlobID          uuid.UUID
	ChangeSetID           uuid.UUID
	ChangedEvidenceRows   []AttachRecordChange
	ChangedEvidenceRecord int
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

func NewBlobLifecycleService(dependencies BlobLifecycleDependencies) (*BlobLifecycleService, error) {
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
	return &BlobLifecycleService{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: incidents.NewAccess(dependencies.Postgres),
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

func NewAccessHandleService(pool postgres.DB) (*AccessHandleService, error) {
	if pool == nil {
		return nil, errors.New("compose Evidence access handles: Postgres is required")
	}
	return &AccessHandleService{accessHandles: accessHandleRepository{db: pool}}, nil
}

func NewCleanupService(pool postgres.DB) (*CleanupService, error) {
	if pool == nil {
		return nil, errors.New("compose Evidence cleanup: Postgres is required")
	}
	return &CleanupService{pool: pool, blobLifecycle: blobLifecycleRepository{db: pool}}, nil
}

func NewRouteOperations(blobs *BlobLifecycleService, access *AccessHandleService) (*RouteOperations, error) {
	if blobs == nil {
		return nil, errors.New("compose Evidence routes: blob lifecycle is required")
	}
	if access == nil {
		return nil, errors.New("compose Evidence routes: access handles are required")
	}
	return &RouteOperations{BlobLifecycleService: blobs, AccessHandleService: access}, nil
}

func newSourceMutationService(
	pool postgres.DB,
	projectionRows evidenceprojection.Rows,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) (*SourceMutationService, error) {
	if pool == nil {
		return nil, errors.New("compose Evidence source mutations: Postgres is required")
	}
	if projectionRows == nil {
		return nil, errors.New("compose Evidence source mutations: Projections is required")
	}
	if appender == nil {
		return nil, errors.New("compose Evidence source mutations: Revisions is required")
	}
	if intents == nil {
		return nil, errors.New("compose Evidence source mutations: Collaboration is required")
	}
	service := &SourceMutationService{}
	service.source = evidenceSourceKernel{
		records:     records.NewStore(),
		rows:        service,
		projections: projectionRows,
	}
	service.mutations = evidenceSourceMutationKernel{
		incidents:     incidents.NewAccess(pool),
		source:        service.source,
		revisions:     newRevisionAppendAdapter(appender),
		collaboration: intents,
	}
	return service, nil
}

func (s *BlobLifecycleService) CreateBlobSlot(ctx context.Context, params BlobSlotParams) (BlobSlotResult, error) {
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
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, params.IncidentID); err != nil {
		return BlobSlotResult{}, err
	}
	if err := s.blobSlots.insertTx(ctx, tx, params); err != nil {
		return BlobSlotResult{}, err
	}
	if err := s.uploadLeases.insertTx(ctx, tx, params.ObjectBlobID, params.IncidentID, params.UploadLease); err != nil {
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

func (s *BlobLifecycleService) GetBlob(ctx context.Context, objectBlobID uuid.UUID) (BlobRecord, error) {
	return s.blobs.load(ctx, objectBlobID)
}

func (s *BlobLifecycleService) GetUploadLease(ctx context.Context, leaseID uuid.UUID) (UploadLeaseRecord, error) {
	return s.uploadLeases.load(ctx, leaseID)
}

func (s *BlobLifecycleService) ClaimUploadLease(ctx context.Context, leaseID uuid.UUID, capabilityHash []byte, now time.Time) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	lease, err := s.uploadLeases.loadForUpdateTx(ctx, tx, leaseID)
	if err != nil {
		return err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, lease.IncidentID); err != nil {
		return err
	}
	if !bytes.Equal(lease.CapabilityHash, capabilityHash) || lease.LeaseState != "issued" || !lease.ExpiresAt.After(now) ||
		lease.Blob.UploadState != "pending" || !lease.Blob.TargetExpiresAt.After(now) || !lease.Blob.PendingExpiresAt.After(now) {
		return ErrUploadLeaseUnavailable
	}
	if err := s.uploadLeases.claimTx(ctx, tx, leaseID, capabilityHash, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *BlobLifecycleService) CompleteUploadLease(ctx context.Context, leaseID uuid.UUID, now time.Time) error {
	return s.uploadLeases.complete(ctx, leaseID, now)
}

// PreflightAttachBlob evaluates every durable authorization, lifecycle,
// contract, version, and idempotency gate that must precede object-store
// observation. AttachBlob repeats these checks inside its mutation transaction.
func (s *BlobLifecycleService) PreflightAttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request AttachBlobRequest, requestHash []byte, now time.Time) (AttachBlobPreflightResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AttachBlobPreflightResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := s.evidenceRows.loadForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobPreflightResult{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return AttachBlobPreflightResult{}, err
	}
	row, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
	if err != nil {
		return AttachBlobPreflightResult{}, err
	}
	if evidenceRowCellValue(row, "evidence.lifecycle_state") == "quarantined" {
		return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceQuarantined, Cause: ErrEvidenceQuarantined}
	}
	blob, err := s.blobs.loadForUpdateTx(ctx, tx, request.ObjectBlobID)
	if err != nil {
		return AttachBlobPreflightResult{}, err
	}
	if blob.IncidentID != meta.IncidentID {
		return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrIncidentMismatch}
	}
	var associatedRecordID uuid.UUID
	associationErr := tx.QueryRow(ctx, `
SELECT record_id
  FROM evidence
 WHERE object_blob_id = $1
`, request.ObjectBlobID).Scan(&associatedRecordID)
	if associationErr != nil && !errors.Is(associationErr, pgx.ErrNoRows) {
		return AttachBlobPreflightResult{}, associationErr
	}
	associated := associationErr == nil
	if associated && associatedRecordID != recordID {
		return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
	}
	switch evidencepolicy.ClassifyBlobForAssociation(blob.UploadState, blob.PendingExpiresAt, now) {
	case evidencepolicy.AssociationBlobNeedsFinalization:
		if blob.UploadLeaseState != "completed" {
			return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobPending, Cause: ErrBlobNotAttachable}
		}
	case evidencepolicy.AssociationBlobAvailable:
	case evidencepolicy.AssociationBlobExpired, evidencepolicy.AssociationBlobFailed:
		return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobQuarantined:
		return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobQuarantined, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobInconsistent:
		return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
	}
	key := authn.RouteIdempotencyKey{
		RouteKey: blobAttachRouteKey, ActorUserID: actor.ID,
		ScopeKey: recordID.String(), ClientTxnID: request.ClientTxnID,
	}
	if existing, err := authn.GetRouteIdempotencyTx(ctx, tx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return AttachBlobPreflightResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return AttachBlobPreflightResult{}, err
		}
		replay := AttachBlobResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ClientTxnID: request.ClientTxnID}
		if err := tx.Commit(ctx); err != nil {
			return AttachBlobPreflightResult{}, err
		}
		return AttachBlobPreflightResult{Blob: blob, Replay: &replay}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return AttachBlobPreflightResult{}, err
	}
	if associated {
		return AttachBlobPreflightResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
	}
	if meta.RowVersion != request.BaseRowVersion {
		return AttachBlobPreflightResult{}, &rowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
	}
	if err := tx.Commit(ctx); err != nil {
		return AttachBlobPreflightResult{}, err
	}
	return AttachBlobPreflightResult{Blob: blob}, nil
}

func (s *BlobLifecycleService) AttachBlob(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request AttachBlobRequest, requestHash []byte, observed *ObservedObject, requestID string, now time.Time) (AttachBlobResult, error) {
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
	beforeRow, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
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
	switch evidencepolicy.ClassifyBlobForAssociation(blob.UploadState, blob.PendingExpiresAt, now) {
	case evidencepolicy.AssociationBlobQuarantined:
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobQuarantined, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobFailed:
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobExpired:
		if err := s.blobLifecycle.failTx(ctx, tx, request.ObjectBlobID, "pending_timeout", now); err != nil {
			return AttachBlobResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return AttachBlobResult{}, err
		}
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobFailed, Cause: ErrBlobNotAttachable}
	case evidencepolicy.AssociationBlobNeedsFinalization:
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
	case evidencepolicy.AssociationBlobAvailable:
	case evidencepolicy.AssociationBlobInconsistent:
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
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
	evidenceLifecycle, ok := evidenceRowCellValue(beforeRow, "evidence.lifecycle_state").(string)
	if !ok || !evidencepolicy.ValidEvidenceLifecycle(evidenceLifecycle) ||
		evidencepolicy.ViolatesEvidenceBlobBridge(evidenceLifecycle, true, evidencepolicy.BlobAvailable) {
		return AttachBlobResult{}, AttachRejectedError{ReasonCode: AttachReasonEvidenceInconsistent, Cause: ErrBlobNotAttachable}
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
	afterRow, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
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

func (s *BlobLifecycleService) QuarantineBlob(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID, trigger string, requestID string, now time.Time) (QuarantineBlobResult, error) {
	if !evidencepolicy.ValidQuarantineEntryTrigger(trigger) {
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
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, blob.IncidentID); err != nil {
		return QuarantineBlobResult{}, err
	}
	if !evidencepolicy.LegalBlobTransition(blob.UploadState, evidencepolicy.BlobQuarantined, trigger) {
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
		row, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
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
		afterRow, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
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

func (s *AccessHandleService) LoadEvidenceAccess(ctx context.Context, recordID uuid.UUID) (EvidenceAccessRecord, error) {
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

func (s *AccessHandleService) InsertHandle(ctx context.Context, handle HandleRecord, issuedByUserID uuid.UUID) error {
	return s.accessHandles.insert(ctx, handle, issuedByUserID)
}

func (s *AccessHandleService) LoadHandle(ctx context.Context, token string) (HandleRecord, error) {
	return s.accessHandles.load(ctx, token)
}

func (s *AccessHandleService) ConsumeDownloadHandle(ctx context.Context, token string, now time.Time) error {
	return s.accessHandles.consumeDownload(ctx, token, now)
}

func (s *AccessHandleService) CheckHandleAccess(ctx context.Context, handle HandleRecord) (string, error) {
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
	var record BlobRecord
	if err := row.Scan(&record.ObjectBlobID, &record.IncidentID, &record.StorageKey, &record.UploadState, &record.ByteSize,
		&record.FilenameHint, &record.ContentTypeHint, &record.ExpectedSHA256Hex,
		&record.ObservedSize, &record.ObservedContentType, &record.ObservedSHA256Hex,
		&record.TargetExpiresAt, &record.PendingExpiresAt, &record.UploadLeaseState); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BlobRecord{}, ErrBlobNotFound
		}
		return BlobRecord{}, err
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

func markBlobAvailableTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, observed *ObservedObject, now time.Time) error {
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

func (s *BlobLifecycleService) refreshEvidenceSupportProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) ([]AttachRecordChange, error) {
	return refreshEvidenceSupportProjectionsTx(ctx, tx, s.supportEffects, incidentID, evidenceRecordID)
}

func refreshEvidenceSupportProjectionsTx(
	ctx context.Context,
	tx pgx.Tx,
	effects evidenceprojection.SupportProjectionEffectsTx,
	incidentID uuid.UUID,
	evidenceRecordID uuid.UUID,
) ([]AttachRecordChange, error) {
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

func attachRecordChangesFromSupportEffects(result evidenceprojection.EvidenceAssociationEffectsResult) []AttachRecordChange {
	changes := make([]AttachRecordChange, 0, len(result.Changes))
	for _, effect := range result.Changes {
		changedFieldKeys := make([]string, 0)
		for _, view := range effect.AffectedViews {
			changedFieldKeys = append(changedFieldKeys, view.ChangedFieldKeys...)
		}
		slices.Sort(changedFieldKeys)
		changedFieldKeys = slices.Compact(changedFieldKeys)
		changes = append(changes, AttachRecordChange{
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

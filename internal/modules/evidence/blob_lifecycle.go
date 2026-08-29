package evidence

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrBlobNotFound           = errors.New("evidence: blob not found")
	ErrEvidenceNotFound       = errors.New("evidence: evidence not found")
	errBlobNotAttachable      = errors.New("evidence: blob not attachable")
	errEvidenceQuarantined    = errors.New("evidence: quarantined")
	ErrIncidentMismatch       = errors.New("evidence: incident mismatch")
	errRowVersionConflict     = errors.New("evidence: row version conflict")
	errIllegalBlobTransition  = errors.New("evidence: illegal blob transition")
	ErrObjectStoreUnavailable = errors.New("evidence: object store unavailable")
	errUploadLeaseNotFound    = errors.New("evidence: upload lease not found")
	errUploadLeaseUnavailable = errors.New("evidence: upload lease unavailable")
)

const (
	AttachReasonBlobNotVisible           = "blob_not_visible"
	attachReasonBlobPending              = "blob_pending"
	attachReasonBlobFailed               = "blob_failed"
	attachReasonBlobQuarantined          = "blob_quarantined"
	attachReasonAcceptedContractMismatch = "accepted_contract_mismatch"
	attachReasonEvidenceQuarantined      = "evidence_quarantined"
	attachReasonEvidenceInconsistent     = "evidence_inconsistent"
)

type AttachRejectedError struct {
	ReasonCode string
	Cause      error
}

func (e AttachRejectedError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return errBlobNotAttachable.Error()
}

func (e AttachRejectedError) Unwrap() error {
	return e.Cause
}

// blobLifecycleService owns object-blob slot, lease, attach, finalization, and
// quarantine behavior. It has no access-handle or cleanup surface.
type blobLifecycleService struct {
	pool           postgres.DB
	idempotency    LifecycleIdempotencyCapability
	incidentAccess evidenceIncidentAdmissionPort
	records        evidenceRecordEnvelopePort
	revisionStore  revisionAppendPort
	projections    evidenceprojection.MutationRows
	supportEffects evidenceprojection.AssociationEffects
	collaboration  collaboration.RecordChangedAppender
	blobSlots      blobSlotRepository
	blobs          blobRepository
	evidenceRows   evidenceRecordRepository
	blobLifecycle  blobLifecycleRepository
	uploadLeases   uploadLeaseRepository
}

type blobLifecycleDependencies struct {
	Postgres        postgres.DB
	Revisions       revisionAppendPort
	Projections     evidenceprojection.MutationRows
	SupportEffects  evidenceprojection.AssociationEffects
	Collaboration   collaboration.RecordChangedAppender
	IncidentState   evidenceIncidentAdmissionPort
	RecordEnvelopes evidenceRecordEnvelopePort
	Idempotency     LifecycleIdempotencyCapability
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
	if isNilMutationCapability(dependencies.Revisions) {
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
	if isNilMutationCapability(dependencies.IncidentState) {
		return nil, errors.New("compose Evidence blob lifecycle: incident state is required")
	}
	if isNilMutationCapability(dependencies.RecordEnvelopes) {
		return nil, errors.New("compose Evidence blob lifecycle: record envelopes are required")
	}
	if isNilMutationCapability(dependencies.Idempotency) {
		return nil, errors.New("compose Evidence blob lifecycle: lifecycle idempotency is required")
	}
	return &blobLifecycleService{
		pool:           dependencies.Postgres,
		idempotency:    dependencies.Idempotency,
		incidentAccess: dependencies.IncidentState,
		records:        dependencies.RecordEnvelopes,
		revisionStore:  dependencies.Revisions,
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

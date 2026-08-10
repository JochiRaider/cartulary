package artifacts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

var (
	ErrIdempotencyNotFound        = errors.New("artifacts: idempotency result not found")
	ErrClientTxnConflict          = errors.New("artifacts: client transaction conflict")
	ErrStoredMutationKindMismatch = errors.New("artifacts: stored mutation kind mismatch")
)

// OperationID is the closed set of Workbook-coordinated operations implemented
// by the Artifacts source owner. Its values remain the durable route identities.
type OperationID string

const (
	OperationWorkbookCreate   OperationID = "workbook.rows.create"
	OperationWorkbookPatch    OperationID = "workbook.records.patch"
	OperationConflictResolve  OperationID = "workbook.records.conflicts.resolve"
	OperationLinkedNoteCreate OperationID = "workbook.records.linked_notes.create"
)

type IdempotencyKey struct {
	OperationID OperationID
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

type IdempotencyRecord struct {
	RequestHash []byte
	Result      StoredMutationResult
}

type StoredMutationKind string

const (
	StoredMutationCreate     StoredMutationKind = "create"
	StoredMutationPatch      StoredMutationKind = "patch"
	StoredMutationLinkedNote StoredMutationKind = "linked_note"
)

type StoredWorkbookResult struct {
	ViewSchemaID   string
	RecordID       uuid.UUID
	ChangeSetID    uuid.UUID
	Row            map[string]any
	SourceRecordID *uuid.UUID
	LinkType       string
}

// StoredMutationResult is a closed operation-tagged union used by the
// application idempotency adapter. It carries source facts, never HTTP status
// codes or transport response envelopes.
type StoredMutationResult struct {
	kind     StoredMutationKind
	workbook StoredWorkbookResult
}

func NewStoredCreateResult(result StoredWorkbookResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationCreate, workbook: result}
}

func NewStoredPatchResult(result StoredWorkbookResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationPatch, workbook: result}
}

func NewStoredLinkedNoteResult(result StoredWorkbookResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationLinkedNote, workbook: result}
}

func (r StoredMutationResult) Kind() StoredMutationKind { return r.kind }

func (r StoredMutationResult) WorkbookResult() (StoredWorkbookResult, bool) {
	switch r.kind {
	case StoredMutationCreate, StoredMutationPatch, StoredMutationLinkedNote:
		return r.workbook, true
	default:
		return StoredWorkbookResult{}, false
	}
}

type IncidentStateCapability interface {
	EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type MemberReferenceCapability interface {
	ValidateIncidentMemberUserTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) error
}

type IdempotencyCapability interface {
	Get(context.Context, IdempotencyKey, []byte) (IdempotencyRecord, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error
}

type RecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (records.Envelope, error)
}

type LinkCapability interface {
	ValidatePartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionValidation) error
	ValidateRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionValidation) error
	ValidateTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionValidation) error
	ApplyPartyRefCollectionWithMutationValuesTx(context.Context, pgx.Tx, links.PartyRefCollectionCommand) (links.CollectionMutationResult, error)
	ApplyRecordRefCollectionWithMutationValuesTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (links.CollectionMutationResult, error)
	ApplyTagCollectionWithMutationValuesTx(context.Context, pgx.Tx, links.TagCollectionCommand) (links.CollectionMutationResult, error)
	InsertLinkedNoteReferenceTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) (links.RecordLink, bool, error)
	LoadRecordLinkValueTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
}

type RevisionCapability interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendNonRowMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendRecordRevisionAndIntentTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicts.RevisionWindowRow, error)
}

type artifactSourceMutationPort interface {
	InsertRowTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, CreateParams, time.Time) error
	ApplyDirectChangeTx(context.Context, pgx.Tx, uuid.UUID, string, FieldValue, time.Time) (bool, error)
	ApplyHandoffRiskRefPayloadTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, RiskRefActionPayload, time.Time) (bool, error)
	NormalizeFindingLifecycleTx(context.Context, pgx.Tx, uuid.UUID, time.Time) (bool, error)
	TouchRowTx(context.Context, pgx.Tx, uuid.UUID, time.Time) error
}

// MutationDependencies is assembled at the application composition root. The
// Artifacts facade never constructs peer-owner or platform stores.
type MutationDependencies struct {
	IncidentState        IncidentStateCapability
	MemberReferences     MemberReferenceCapability
	Idempotency          IdempotencyCapability
	RecordEnvelopes      RecordEnvelopeCapability
	Links                LinkCapability
	Projections          artifactprojection.Rows
	Revisions            RevisionCapability
	ConflictFields       conflicts.FieldResolver
	KeepSavedIdempotency conflicts.IdempotencyPort
}

func (d MutationDependencies) validate() error {
	required := []struct {
		name  string
		value any
	}{
		{name: "Incident admission", value: d.IncidentState},
		{name: "Member validation", value: d.MemberReferences},
		{name: "Route idempotency", value: d.Idempotency},
		{name: "Record envelopes", value: d.RecordEnvelopes},
		{name: "Links", value: d.Links},
		{name: "Projections", value: d.Projections},
		{name: "Revisions/history", value: d.Revisions},
		{name: "Conflict fields", value: d.ConflictFields},
		{name: "Keep-saved idempotency", value: d.KeepSavedIdempotency},
	}
	for _, dependency := range required {
		if dependency.value == nil {
			return fmt.Errorf("artifacts mutation dependencies: %s is required", dependency.name)
		}
	}
	return nil
}

type memberReferenceValidator struct{}

func NewMemberReferenceCapability() MemberReferenceCapability {
	return memberReferenceValidator{}
}

func (memberReferenceValidator) ValidateIncidentMemberUserTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	userID uuid.UUID,
	field string,
) error {
	return validateIncidentMemberUserTx(ctx, tx, incidentID, userID, field)
}

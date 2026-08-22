package tasksdecisions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
)

// ErrIdempotencyNotFound is returned when no committed route result exists.
var ErrIdempotencyNotFound = errors.New("tasksdecisions: idempotency result not found")

// ErrClientTxnConflict identifies reuse of an idempotency key with different
// request content. Workbook translates this owner error at its transport edge.
var ErrClientTxnConflict = errors.New("tasksdecisions: client transaction conflict")

// ErrStoredMutationKindMismatch rejects replay data bound to a different
// operation before any source transaction begins.
var ErrStoredMutationKindMismatch = errors.New("tasksdecisions: stored mutation kind mismatch")

type IdempotencyKey struct {
	RouteKey    string
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
	StoredMutationCreate               StoredMutationKind = "create"
	StoredMutationPatch                StoredMutationKind = "patch"
	StoredMutationDecisionSupersession StoredMutationKind = "decision_supersession"
)

type StoredRowMutationResult struct {
	ViewSchemaID string
	RecordID     uuid.UUID
	ChangeSetID  uuid.UUID
	Row          map[string]any
}

type StoredDecisionSupersessionResult struct {
	ViewSchemaID string
	ChangeSetID  uuid.UUID
	Facts        SupersedeFacts
}

// StoredMutationResult is a closed operation-tagged union. Constructors are
// the only way application adapters can create a valid stored result.
type StoredMutationResult struct {
	kind         StoredMutationKind
	row          StoredRowMutationResult
	supersession StoredDecisionSupersessionResult
}

func NewStoredCreateResult(result StoredRowMutationResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationCreate, row: result}
}

func NewStoredPatchResult(result StoredRowMutationResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationPatch, row: result}
}

func NewStoredDecisionSupersessionResult(result StoredDecisionSupersessionResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationDecisionSupersession, supersession: result}
}

func (r StoredMutationResult) Kind() StoredMutationKind { return r.kind }

func (r StoredMutationResult) RowMutationResult() (StoredRowMutationResult, bool) {
	if r.kind != StoredMutationCreate && r.kind != StoredMutationPatch {
		return StoredRowMutationResult{}, false
	}
	return r.row, true
}

func (r StoredMutationResult) DecisionSupersessionResult() (StoredDecisionSupersessionResult, bool) {
	if r.kind != StoredMutationDecisionSupersession {
		return StoredDecisionSupersessionResult{}, false
	}
	return r.supersession, true
}

// IncidentStateCapability owns only incident lifecycle admission in a caller
// supplied transaction.
type IncidentStateCapability interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

// MemberReferenceCapability validates the stable active same-incident user
// reference required by Task/Decision source semantics.
type MemberReferenceCapability interface {
	ValidateIncidentMemberUserTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) error
}

// IdempotencyCapability preserves the existing lookup-before-validation and
// transaction-bound write order without exposing the platform auth store.
type IdempotencyCapability interface {
	Get(context.Context, IdempotencyKey, []byte) (IdempotencyRecord, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error
}

// RecordEnvelopeCapability is the exact Records-owned persistence used by the
// Task/Decision aggregate. Its methods never begin or commit transactions.
type RecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (records.Envelope, error)
}

// LinkCapability groups only the fixed link operations consumed by
// Task/Decision mutations.
type LinkCapability interface {
	ValidateRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionValidation) error
	ApplyRecordRefCollectionWithMutationValuesTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (links.CollectionMutationResult, error)
	SyncFieldReferenceWithMutationValuesTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (links.CollectionMutationResult, error)
	InsertSupersedesCommandTx(context.Context, pgx.Tx, links.InsertSupersedesCommand) (links.RecordLinkCommandResult, error)
}

// RevisionCapability is the exact revision append/history surface consumed by
// Task/Decision mutations. It does not publish, retry, or commit.
type RevisionCapability interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendRecordRevisionAndIntentTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicts.RevisionWindowRow, error)
}

// MutationDependencies is assembled at the application composition root. The
// facade does not construct platform or peer-owner stores.
type MutationDependencies struct {
	IncidentState        IncidentStateCapability
	MemberReferences     MemberReferenceCapability
	Idempotency          IdempotencyCapability
	RecordEnvelopes      RecordEnvelopeCapability
	Links                LinkCapability
	Projections          taskdecisionprojection.Rows
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
			return fmt.Errorf("tasks/decisions mutation dependencies: %s is required", dependency.name)
		}
	}
	return nil
}

type memberReferenceValidator struct{}

// NewMemberReferenceCapability constructs the source-owned reference policy;
// application composition decides where it is used.
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

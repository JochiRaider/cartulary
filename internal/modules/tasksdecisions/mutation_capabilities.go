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
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflictresolution"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
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

type StoredWorkbookResult struct {
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
	workbook     StoredWorkbookResult
	supersession StoredDecisionSupersessionResult
}

func NewStoredCreateResult(result StoredWorkbookResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationCreate, workbook: result}
}

func NewStoredPatchResult(result StoredWorkbookResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationPatch, workbook: result}
}

func NewStoredDecisionSupersessionResult(result StoredDecisionSupersessionResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationDecisionSupersession, supersession: result}
}

func (r StoredMutationResult) Kind() StoredMutationKind { return r.kind }

func (r StoredMutationResult) WorkbookResult() (StoredWorkbookResult, bool) {
	if r.kind != StoredMutationCreate && r.kind != StoredMutationPatch {
		return StoredWorkbookResult{}, false
	}
	return r.workbook, true
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
	EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error
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
	ApplyRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (bool, error)
	SyncFieldReferenceCommandTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (bool, error)
	InsertSupersedesCommandTx(context.Context, pgx.Tx, links.InsertSupersedesCommand) (links.SupersedesLink, error)
}

// ProjectionCapability supplies authoritative projection refresh/load facts in
// the transaction owned by the facade.
type ProjectionCapability interface {
	RefreshTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) error
	RefreshDecisionTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	LoadDecisionTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	LoadTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

// RevisionCapability is the exact revision append/history surface consumed by
// Task/Decision mutations. It does not publish, retry, or commit.
type RevisionCapability interface {
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, revisions.AppendMutationParams) error
	AppendRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]historyquery.RevisionWindowRow, error)
}

// MutationDependencies is assembled at the application composition root. The
// facade does not construct platform or peer-owner stores.
type MutationDependencies struct {
	IncidentState        IncidentStateCapability
	MemberReferences     MemberReferenceCapability
	Idempotency          IdempotencyCapability
	RecordEnvelopes      RecordEnvelopeCapability
	Links                LinkCapability
	Projections          ProjectionCapability
	Revisions            RevisionCapability
	KeepSavedIdempotency conflictresolution.IdempotencyPort
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

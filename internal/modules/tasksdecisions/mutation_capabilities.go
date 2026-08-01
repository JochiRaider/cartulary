package tasksdecisions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

// ErrIdempotencyNotFound is returned when no committed route result exists.
var ErrIdempotencyNotFound = errors.New("tasksdecisions: idempotency result not found")

// ErrClientTxnConflict identifies reuse of an idempotency key with different
// request content. Workbook translates this owner error at its transport edge.
var ErrClientTxnConflict = errors.New("tasksdecisions: client transaction conflict")

type IdempotencyKey struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

type IdempotencyRecord struct {
	RequestHash  []byte
	ResponseJSON []byte
}

type IdempotencyOutcome string

const (
	IdempotencyOutcomeCreated IdempotencyOutcome = "created"
	IdempotencyOutcomeUpdated IdempotencyOutcome = "updated"
)

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
	Get(context.Context, IdempotencyKey) (IdempotencyRecord, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, IdempotencyOutcome, any) error
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

// MutationCapabilities is assembled at the application composition root. The
// two facades do not construct platform or peer-owner stores.
type MutationCapabilities struct {
	IncidentState       IncidentStateCapability
	MemberReferences    MemberReferenceCapability
	Idempotency         IdempotencyCapability
	RecordEnvelopes     RecordEnvelopeCapability
	Links               LinkCapability
	Projections         ProjectionCapability
	Revisions           RevisionCapability
	ConflictIdempotency *authn.Store
}

func (c MutationCapabilities) validate() {
	if c.IncidentState == nil || c.MemberReferences == nil || c.Idempotency == nil ||
		c.RecordEnvelopes == nil || c.Links == nil || c.Projections == nil ||
		c.Revisions == nil || c.ConflictIdempotency == nil {
		panic("tasksdecisions mutation capabilities are incomplete")
	}
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

// ClassifyIdempotencyWriteError keeps PostgreSQL classification in the source
// adapter while allowing application composition to translate the conflict.
func ClassifyIdempotencyWriteError(err error) error {
	if tasksource.IsUniqueViolation(err) {
		return ErrClientTxnConflict
	}
	return err
}

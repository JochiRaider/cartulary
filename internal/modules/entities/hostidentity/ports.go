package hostidentity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type entityStorePorts struct {
	records     entityRecordPort
	revisions   entityRevisionPort
	projections workbookprojection.Writer
}

type entityRecordPort interface {
	InsertTx(context.Context, pgx.Tx, entityRecordInsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type entityRevisionPort interface {
	AppendChangeSetTx(context.Context, pgx.Tx, entityChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, entityMutationParams) error
	AppendRecordRevisionTx(context.Context, pgx.Tx, entityRecordRevisionParams) error
}

type entityChangeSetParams struct {
	ChangeSetID *uuid.UUID
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

type entityRecordInsertParams struct {
	RecordID        *uuid.UUID
	IncidentID      uuid.UUID
	RecordType      string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedByUserID uuid.UUID
	UpdatedAt       time.Time
	RowVersion      int64
}

type entityMutationParams struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type entityRecordRevisionParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue any
	AfterValue  any
}

func newEntityStorePorts(
	pool postgres.DB,
	appender *revisions.Appender,
	projectionWriter workbookprojection.Writer,
) entityStorePorts {
	return entityStorePorts{
		records:     entityRecordAdapter{store: records.NewStore()},
		revisions:   entityRevisionAdapter{appender: appender},
		projections: projectionWriter,
	}
}

type entityRecordAdapter struct {
	store *records.Store
}

func (a entityRecordAdapter) InsertTx(ctx context.Context, tx pgx.Tx, params entityRecordInsertParams) (uuid.UUID, error) {
	return a.store.InsertTx(ctx, tx, records.InsertParams(params))
}

func (a entityRecordAdapter) AdvanceVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	return a.store.AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
}

func (a entityRecordAdapter) LoadRowVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, error) {
	return a.store.LoadRowVersionTx(ctx, tx, recordID)
}

type entityRevisionAdapter struct {
	appender *revisions.Appender
}

func (a entityRevisionAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params entityChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams(params))
}

func (a entityRevisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params entityMutationParams) error {
	return a.appender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams(params))
}

func (a entityRevisionAdapter) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params entityRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams(params))
}

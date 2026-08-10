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
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, entityChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, entityMutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendRecordRevisionAndIntentTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
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

func (a entityRevisionAdapter) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (a entityRevisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params entityMutationParams) error {
	return a.appender.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams(params))
}

func (a entityRevisionAdapter) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return a.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (a entityRevisionAdapter) AppendRecordRevisionAndIntentTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionAndIntentTx(ctx, tx, params)
}

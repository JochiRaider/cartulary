package hostidentity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type entityStorePorts struct {
	records     entityRecordPort
	revisions   entityRevisionPort
	projections entityProjectionPort
}

type entityRecordPort interface {
	InsertTx(context.Context, pgx.Tx, entityRecordInsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type entityRevisionPort interface {
	InsertChangeSetTx(context.Context, pgx.Tx, entityChangeSetParams) (uuid.UUID, error)
	InsertMutationTx(context.Context, pgx.Tx, entityMutationParams) error
	InsertRecordRevisionTx(context.Context, pgx.Tx, entityRecordRevisionParams) error
}

type entityProjectionPort interface {
	RefreshEntityRowTx(context.Context, pgx.Tx, uuid.UUID, string) error
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

func newEntityStorePorts(pool postgres.DB) entityStorePorts {
	return entityStorePorts{
		records:     entityRecordAdapter{store: records.NewStore()},
		revisions:   entityRevisionAdapter{store: revisions.NewStore()},
		projections: entityProjectionAdapter{projector: projectionadapters.NewRowProjector(pool)},
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
	store *revisions.Store
}

func (a entityRevisionAdapter) InsertChangeSetTx(ctx context.Context, tx pgx.Tx, params entityChangeSetParams) (uuid.UUID, error) {
	return a.store.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams(params))
}

func (a entityRevisionAdapter) InsertMutationTx(ctx context.Context, tx pgx.Tx, params entityMutationParams) error {
	return a.store.InsertMutationTx(ctx, tx, revisions.MutationParams(params))
}

func (a entityRevisionAdapter) InsertRecordRevisionTx(ctx context.Context, tx pgx.Tx, params entityRecordRevisionParams) error {
	return a.store.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams(params))
}

type entityProjectionAdapter struct {
	projector *projectionadapters.RowProjector
}

func (a entityProjectionAdapter) RefreshEntityRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) error {
	switch entityType {
	case "host":
		return a.projector.RefreshRowTx(ctx, tx, projectionadapters.HostsViewSchemaID, recordID)
	case "identity":
		return a.projector.RefreshRowTx(ctx, tx, projectionadapters.IdentitiesViewSchemaID, recordID)
	default:
		return nil
	}
}

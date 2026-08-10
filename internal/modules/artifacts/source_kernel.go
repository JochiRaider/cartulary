package artifacts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type recordEnvelopeInserter interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
}

type artifactProjectionRows interface {
	RefreshArtifactTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadArtifactTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

// artifactSourceKernel is the caller-transaction source boundary shared by
// interactive mutation and import adapters. It owns record/source persistence
// and projection refresh, but never begins, commits, or rolls back a transaction.
type artifactSourceKernel struct {
	records     recordEnvelopeInserter
	rows        artifactSourceMutationPort
	projections artifactProjectionRows
}

func (k artifactSourceKernel) createRecordTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	params createParams,
	now time.Time,
) (uuid.UUID, error) {
	recordID, err := k.records.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      "artifact",
		CreatedByUserID: actorID,
		CreatedAt:       now,
		UpdatedByUserID: actorID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := k.rows.insertRowTx(
		ctx,
		tx,
		recordID,
		incidentID,
		actorID,
		params,
		now,
	); err != nil {
		return uuid.UUID{}, err
	}
	return recordID, nil
}

func (k artifactSourceKernel) refreshRowTx(
	ctx context.Context,
	tx pgx.Tx,
	viewSchemaID string,
	recordID uuid.UUID,
) (map[string]any, error) {
	if err := k.projections.RefreshArtifactTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return k.projections.LoadArtifactTx(ctx, tx, viewSchemaID, recordID)
}

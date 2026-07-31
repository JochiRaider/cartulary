package evidence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
)

// evidenceSourceKernel owns transaction-supplied Evidence record/source writes
// and row refreshes. It never begins, commits, or rolls back a transaction.
type evidenceSourceKernel struct {
	records     evidenceRecordEnvelopePort
	rows        evidenceSourceMutationPort
	projections evidenceProjectionRowPort
}

func (kernel evidenceSourceKernel) createRecordTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	params WorkbookCreateParams,
	now time.Time,
) (uuid.UUID, error) {
	recordID, err := kernel.records.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      "evidence",
		CreatedByUserID: actorID,
		CreatedAt:       now,
		UpdatedByUserID: actorID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return uuid.Nil, err
	}
	if err := kernel.rows.InsertWorkbookRowTx(ctx, tx, recordID, incidentID, params, now); err != nil {
		return uuid.Nil, err
	}
	return recordID, nil
}

func (kernel evidenceSourceKernel) refreshRowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	if err := kernel.projections.RefreshTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return kernel.projections.LoadTx(ctx, tx, recordID)
}

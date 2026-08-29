package evidence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

// evidenceSourceKernel owns transaction-supplied Evidence record/source writes
// and row refreshes. It never begins, commits, or rolls back a transaction.
type evidenceSourceKernel struct {
	records     evidenceRecordEnvelopePort
	rows        evidenceSourceMutationPort
	projections evidenceprojection.MutationRows
}

func (kernel evidenceSourceKernel) createRecordTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	params createParams,
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
	if err := kernel.rows.insertRowTx(ctx, tx, recordID, incidentID, params, now); err != nil {
		return uuid.Nil, err
	}
	return recordID, nil
}

func (kernel evidenceSourceKernel) refreshRowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	if err := kernel.projections.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return kernel.projections.LoadEvidenceTx(ctx, tx, recordID)
}

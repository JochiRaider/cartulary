package deleterestore

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

type source struct{}

var _ deleterestorecontract.DeleteRestoreSource = source{}

func NewSource() deleterestorecontract.DeleteRestoreSource {
	return source{}
}

func (source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(e))
  FROM records r
  JOIN evidence e
    ON e.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (source) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (source) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.evidence.v1", nil
}

func (source) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

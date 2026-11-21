package deleterestore

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

const timelineViewSchemaID = "cartulary.view.timeline.v2"

type Source struct{}

var _ deleterestorecontract.DeleteRestoreSource = Source{}

func NewSource() Source {
	return Source{}
}

func (s Source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	snapshots, err := s.SnapshotsTx(ctx, tx, []uuid.UUID{recordID})
	return deleterestorecontract.SnapshotOf(snapshots, recordID, err)
}

func (Source) SnapshotsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	return deleterestorecontract.CollectSnapshots(tx.Query(ctx, `
SELECT r.record_id, jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(t))
  FROM records r
  JOIN timeline_events t
    ON t.record_id = r.record_id
 WHERE r.record_id = ANY($1)
`, recordIDs))
}

func (Source) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (Source) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return timelineViewSchemaID, nil
}

func (Source) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func (Source) ValidateStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

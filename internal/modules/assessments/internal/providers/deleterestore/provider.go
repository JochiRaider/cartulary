package deleterestore

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

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
SELECT r.record_id, jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(a))
  FROM records r
  JOIN assessments a
    ON a.record_id = r.record_id
 WHERE r.record_id = ANY($1)
`, recordIDs))
}

func (Source) UpdateSourceDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, deleting bool) error {
	var (
		tag pgconn.CommandTag
		err error
	)
	if deleting {
		tag, err = tx.Exec(ctx, `
UPDATE assessments
   SET deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2
 WHERE record_id = $1
`, recordID, now.UTC(), actorUserID)
	} else {
		tag, err = tx.Exec(ctx, `
UPDATE assessments
   SET deleted_at = NULL,
       deleted_by_user_id = NULL,
       updated_at = $2
 WHERE record_id = $1
`, recordID, now.UTC())
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("update assessment delete state affected %d rows", tag.RowsAffected())
	}
	return nil
}

func (Source) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.assessments.v1", nil
}

func (Source) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func (Source) ValidateStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

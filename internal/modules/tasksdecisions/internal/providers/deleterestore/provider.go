package deleterestore

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

type TaskRequestSource struct{}

var _ deleterestorecontract.DeleteRestoreSource = TaskRequestSource{}

func NewTaskRequestSource() TaskRequestSource {
	return TaskRequestSource{}
}

func (TaskRequestSource) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(t))
  FROM records r
  JOIN task_requests t
    ON t.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (TaskRequestSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (TaskRequestSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.task_requests.v1", nil
}

func (TaskRequestSource) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

type DecisionSource struct{}

var _ deleterestorecontract.DeleteRestoreSource = DecisionSource{}

func NewDecisionSource() DecisionSource {
	return DecisionSource{}
}

func (DecisionSource) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(d))
  FROM records r
  JOIN decisions d
    ON d.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (DecisionSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (DecisionSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.decisions.v1", nil
}

func (DecisionSource) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

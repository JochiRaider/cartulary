package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
)

const taskRequestProjectionRowSQL = `
INSERT INTO task_request_grid_projection (
    record_id, incident_id, row_version, title, status, owner_user_id,
    priority, task_kind, workstream, due_at, requester_party_text,
    requester_party_id, blocked_reason, completed_at, external_ticket_ref,
    closure_summary, decision_record_id, linked_record_count, updated_at,
    no_owner
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
`

func taskRequestProjectionRowArguments(input taskprojection.TaskRequestProjectionInput) ([]any, error) {
	return []any{
		input.RecordID, input.IncidentID, input.RowVersion, input.Title,
		input.Status, input.OwnerUserID, input.Priority, input.TaskKind,
		input.Workstream, input.DueAt, input.RequesterPartyText,
		input.RequesterPartyID, input.BlockedReason, input.CompletedAt,
		input.ExternalTicketRef, input.ClosureSummary, input.DecisionRecordID,
		input.LinkedRecordCount, input.UpdatedAt.UTC(), input.NoOwner,
	}, nil
}

func (store *Store) InsertTaskRequestTx(ctx context.Context, tx pgx.Tx, input taskprojection.TaskRequestProjectionInput) error {
	args, err := taskRequestProjectionRowArguments(input)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, taskRequestProjectionRowSQL, args...); err != nil {
		return fmt.Errorf("insert Task-request projection row: %w", err)
	}
	return nil
}

// InsertTaskRequestRowsTx writes a page of rows in one round trip.
func (store *Store) InsertTaskRequestRowsTx(ctx context.Context, tx pgx.Tx, inputs []taskprojection.TaskRequestProjectionInput) error {
	if len(inputs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, input := range inputs {
		args, err := taskRequestProjectionRowArguments(input)
		if err != nil {
			return err
		}
		batch.Queue(taskRequestProjectionRowSQL, args...)
	}
	return projectionRowsResult(tx.SendBatch(ctx, batch), len(inputs), "insert Task-request projection row")
}

func (store *Store) DeleteTaskRequestRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM task_request_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete Task-request projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteTaskRequestIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM task_request_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear Task-request projection rows: %w", err)
	}
	return nil
}

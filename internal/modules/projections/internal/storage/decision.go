package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	decisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
)

func (store *Store) InsertDecisionTx(ctx context.Context, tx pgx.Tx, input decisionprojection.DecisionProjectionInput) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO decision_grid_projection (
    record_id, incident_id, row_version, summary, status, owner_user_id,
    decision_type, decided_at, rationale, affected_record_count,
    supersedes_record_id, updated_at, is_superseded
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
`, input.RecordID, input.IncidentID, input.RowVersion, input.Summary,
		input.Status, input.OwnerUserID, input.DecisionType, input.DecidedAt,
		input.Rationale, input.AffectedRecordCount, input.SupersedesRecordID,
		input.UpdatedAt.UTC(), input.IsSuperseded); err != nil {
		return fmt.Errorf("insert Decision projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteDecisionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM decision_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete Decision projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteDecisionIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM decision_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear Decision projection rows: %w", err)
	}
	return nil
}

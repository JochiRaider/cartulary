package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entitycontract "github.com/JochiRaider/cartulary/internal/modules/entities/projectioncontract"
)

func (store *Store) UpsertIdentityTx(
	ctx context.Context,
	tx pgx.Tx,
	input entitycontract.IdentityProjectionInput,
) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO identity_grid_projection (
    record_id, incident_id, row_version, display_name, upn, email,
    sam_account_name, identity_state, linked_event_count, evidence_count,
    privilege_level, mfa_state, reset_status, edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    display_name = EXCLUDED.display_name,
    upn = EXCLUDED.upn,
    email = EXCLUDED.email,
    sam_account_name = EXCLUDED.sam_account_name,
    identity_state = EXCLUDED.identity_state,
    linked_event_count = EXCLUDED.linked_event_count,
    evidence_count = EXCLUDED.evidence_count,
    privilege_level = EXCLUDED.privilege_level,
    mfa_state = EXCLUDED.mfa_state,
    reset_status = EXCLUDED.reset_status,
    edited_at = EXCLUDED.edited_at
`, input.RecordID, input.IncidentID, input.RowVersion, input.DisplayName,
		input.UPN, input.Email, input.SamAccountName, input.IdentityState,
		input.LinkedEventCount, input.EvidenceCount, input.PrivilegeLevel,
		input.MFAState, input.ResetStatus, input.EditedAt.UTC()); err != nil {
		return fmt.Errorf("upsert identity projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteIdentityRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM identity_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete identity projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteIdentityIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM identity_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear identity projection rows: %w", err)
	}
	return nil
}

func (store *Store) LoadIdentityEvidenceAssociationStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, int64, error) {
	var rowVersion int64
	var evidenceCount int64
	if err := tx.QueryRow(ctx, `
SELECT row_version, evidence_count
  FROM identity_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&rowVersion, &evidenceCount); err != nil {
		return 0, 0, err
	}
	return rowVersion, evidenceCount, nil
}

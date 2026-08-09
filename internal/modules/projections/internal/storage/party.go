package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
)

func (store *Store) InsertPartyTx(ctx context.Context, tx pgx.Tx, input partyprojection.ProjectionInput) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO party_grid_projection (
    record_id, incident_id, row_version, display_name, party_kind,
    organization_name, role_title, primary_email, timezone_name,
    external_ref, notes, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`, input.RecordID, input.IncidentID, input.RowVersion, input.DisplayName,
		input.PartyKind, input.OrganizationName, input.RoleTitle,
		input.PrimaryEmail, input.TimezoneName, input.ExternalRef, input.Notes,
		input.UpdatedAt.UTC()); err != nil {
		return fmt.Errorf("insert Party projection row: %w", err)
	}
	return nil
}

func (store *Store) DeletePartyRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM party_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete Party projection row: %w", err)
	}
	return nil
}

func (store *Store) DeletePartyIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM party_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear Party projection rows: %w", err)
	}
	return nil
}

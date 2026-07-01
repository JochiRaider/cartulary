package projectionprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const partyProjectionColumns = `
    record_id,
    incident_id,
    row_version,
    display_name,
    party_kind,
    organization_name,
    role_title,
    primary_email,
    timezone_name,
    external_ref,
    notes,
    updated_at`

const partyProjectionSelect = `
SELECT
    p.record_id,
    p.incident_id,
    r.row_version,
    p.display_name,
    p.party_kind,
    p.organization_name,
    p.role_title,
    p.primary_email,
    p.timezone_name,
    p.external_ref,
    p.notes,
    p.updated_at
  FROM parties p
  JOIN records r
    ON r.incident_id = p.incident_id
   AND r.record_id = p.record_id
   AND r.deleted_at IS NULL`

func RefreshPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM party_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("clear party projection row: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO party_grid_projection (`+partyProjectionColumns+`) `+partyProjectionSelect+` WHERE p.record_id = $1`, recordID); err != nil {
		return fmt.Errorf("refresh party projection: %w", err)
	}
	return nil
}

func RebuildIncidentPartiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM party_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear party projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO party_grid_projection (`+partyProjectionColumns+`) `+partyProjectionSelect+` WHERE p.incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("insert party projection rows: %w", err)
	}
	return nil
}

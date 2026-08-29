package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entitycontract "github.com/JochiRaider/cartulary/internal/modules/entities/projectioncontract"
)

func (store *Store) UpsertHostTx(
	ctx context.Context,
	tx pgx.Tx,
	input entitycontract.HostProjectionInput,
) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO host_grid_projection (
    record_id, incident_id, row_version, display_name, hostname, host_state,
    linked_event_count, evidence_count, location, os_platform, business_owner,
    criticality, containment_status, edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    display_name = EXCLUDED.display_name,
    hostname = EXCLUDED.hostname,
    host_state = EXCLUDED.host_state,
    linked_event_count = EXCLUDED.linked_event_count,
    evidence_count = EXCLUDED.evidence_count,
    location = EXCLUDED.location,
    os_platform = EXCLUDED.os_platform,
    business_owner = EXCLUDED.business_owner,
    criticality = EXCLUDED.criticality,
    containment_status = EXCLUDED.containment_status,
    edited_at = EXCLUDED.edited_at
`, input.RecordID, input.IncidentID, input.RowVersion, input.DisplayName,
		input.Hostname, input.HostState, input.LinkedEventCount, input.EvidenceCount,
		input.Location, input.OSPlatform, input.BusinessOwner, input.Criticality,
		input.ContainmentStatus, input.EditedAt.UTC()); err != nil {
		return fmt.Errorf("upsert host projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteHostRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM host_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete host projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteHostIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM host_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear host projection rows: %w", err)
	}
	return nil
}

func (store *Store) LoadHostEvidenceAssociationStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, int64, error) {
	var rowVersion int64
	var evidenceCount int64
	if err := tx.QueryRow(ctx, `
SELECT row_version, evidence_count
  FROM host_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&rowVersion, &evidenceCount); err != nil {
		return 0, 0, err
	}
	return rowVersion, evidenceCount, nil
}

package projectionprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const evidenceProjectionColumns = `
    record_id,
    incident_id,
    row_version,
    title,
    lifecycle_state,
    requested_at,
    received_at,
    storage_ref,
    blob_hash,
    collector_party_text,
    collector_party_id,
    source_party_text,
    source_party_id,
    upload_state,
    linked_record_count,
    edited_at`

const evidenceProjectionSelect = `
SELECT
    e.record_id,
    e.incident_id,
    r.row_version,
    e.title,
    e.lifecycle_state,
    e.requested_at,
    e.received_at,
    e.storage_ref,
    COALESCE(b.observed_sha256_hex, e.blob_hash),
    e.collector_party_text,
    e.collector_party_id,
    e.source_party_text,
    e.source_party_id,
    COALESCE(b.upload_state, e.upload_state),
    0::integer,
    e.updated_at
  FROM evidence e
  JOIN records r
    ON r.incident_id = e.incident_id
   AND r.record_id = e.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN object_blobs b
    ON b.object_blob_id = e.object_blob_id`

func RefreshEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM evidence_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("clear evidence projection row: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO evidence_grid_projection (`+evidenceProjectionColumns+`) `+evidenceProjectionSelect+` WHERE e.record_id = $1`, recordID); err != nil {
		return fmt.Errorf("refresh evidence projection: %w", err)
	}
	return nil
}

func RebuildIncidentEvidenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM evidence_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear evidence projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO evidence_grid_projection (`+evidenceProjectionColumns+`) `+evidenceProjectionSelect+` WHERE e.incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("insert evidence projection rows: %w", err)
	}
	return nil
}

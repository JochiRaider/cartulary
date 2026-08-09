package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
)

func (store *Store) InsertEvidenceTx(
	ctx context.Context,
	tx pgx.Tx,
	input evidenceprojection.ProjectionInput,
) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO evidence_grid_projection (
    record_id, incident_id, row_version, title, lifecycle_state,
    requested_at, received_at, storage_ref, blob_hash,
    collector_party_text, collector_party_id, source_party_text,
    source_party_id, upload_state, linked_record_count, edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
`, input.RecordID, input.IncidentID, input.RowVersion, input.Title,
		input.LifecycleState, input.RequestedAt, input.ReceivedAt,
		input.StorageRef, input.BlobHash, input.CollectorPartyText,
		input.CollectorPartyID, input.SourcePartyText, input.SourcePartyID,
		input.UploadState, input.LinkedRecordCount, input.EditedAt.UTC()); err != nil {
		return fmt.Errorf("insert Evidence projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteEvidenceRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM evidence_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete Evidence projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteEvidenceIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM evidence_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear Evidence projection rows: %w", err)
	}
	return nil
}

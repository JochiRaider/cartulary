package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
)

func (store *Store) UpsertIndicatorTx(
	ctx context.Context,
	tx pgx.Tx,
	input indicatorprojection.ProjectionInput,
) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO indicator_grid_projection (
    record_id, incident_id, row_version, indicator_type, value_kind,
    display_value, normalized_value, dedupe_key, defanged_value,
    hash_algorithm, hash_value, stix_pattern, first_observed_at,
    last_observed_at, observation_count, lifecycle_summary,
    supporting_link_count, edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    indicator_type = EXCLUDED.indicator_type,
    value_kind = EXCLUDED.value_kind,
    display_value = EXCLUDED.display_value,
    normalized_value = EXCLUDED.normalized_value,
    dedupe_key = EXCLUDED.dedupe_key,
    defanged_value = EXCLUDED.defanged_value,
    hash_algorithm = EXCLUDED.hash_algorithm,
    hash_value = EXCLUDED.hash_value,
    stix_pattern = EXCLUDED.stix_pattern,
    first_observed_at = EXCLUDED.first_observed_at,
    last_observed_at = EXCLUDED.last_observed_at,
    observation_count = EXCLUDED.observation_count,
    lifecycle_summary = EXCLUDED.lifecycle_summary,
    supporting_link_count = EXCLUDED.supporting_link_count,
    edited_at = EXCLUDED.edited_at
`, input.RecordID, input.IncidentID, input.RowVersion, input.IndicatorType,
		input.ValueKind, input.DisplayValue, input.NormalizedValue, input.DedupeKey,
		input.DefangedValue, input.HashAlgorithm, input.HashValue, input.STIXPattern,
		input.FirstObservedAt, input.LastObservedAt, input.ObservationCount,
		input.LifecycleSummary, input.SupportingLinkCount, input.EditedAt.UTC()); err != nil {
		return fmt.Errorf("upsert Indicator projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteIndicatorRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM indicator_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("delete Indicator projection row: %w", err)
	}
	return nil
}

func (store *Store) DeleteIndicatorIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM indicator_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear Indicator projection rows: %w", err)
	}
	return nil
}

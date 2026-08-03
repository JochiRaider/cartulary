package indicators

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (sourceRepository) loadByDedupeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, indicatorType string, dedupeKey string) (IndicatorRecord, bool, error) {
	record, err := scanIndicatorRecord(tx.QueryRow(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id,
    r.deleted_at,
    r.deleted_by_user_id
  FROM indicators i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.incident_id = $1
   AND i.indicator_type = $2
   AND i.dedupe_key = $3
   AND r.deleted_at IS NULL
 LIMIT 1
 FOR UPDATE OF i, r
`, incidentID, indicatorType, dedupeKey))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorRecord{}, false, nil
	}
	if err != nil {
		return IndicatorRecord{}, false, fmt.Errorf("load indicator by dedupe: %w", err)
	}
	return record, true, nil
}

func (sourceRepository) insertTx(ctx context.Context, tx pgx.Tx, record *IndicatorRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO indicators (
    record_id,
    incident_id,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13, $14, $14)
RETURNING record_id
`, record.RecordID, record.IncidentID, record.IndicatorType, record.ValueKind, record.DisplayValue, record.NormalizedValue, record.DedupeKey, record.DefangedValue, record.HashAlgorithm, record.HashValue, record.STIXPattern, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func (sourceRepository) updateTx(ctx context.Context, tx pgx.Tx, record IndicatorRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE indicators
   SET defanged_value = $2,
       hash_algorithm = $3,
       hash_value = $4,
       stix_pattern = $5,
       row_version = $6,
       updated_at = $7,
       updated_by_user_id = $8
 WHERE record_id = $1
`, record.RecordID, record.DefangedValue, record.HashAlgorithm, record.HashValue, record.STIXPattern, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update indicator: %w", err)
	}
	return nil
}

func (sourceRepository) loadSupportingLinkCountTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int, error) {
	var count int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*)
  FROM active_record_links_v1
 WHERE dst_record_id = $1
`, recordID).Scan(&count); err != nil {
		return 0, fmt.Errorf("load indicator supporting link count: %w", err)
	}
	return count, nil
}

func (sourceRepository) loadTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IndicatorRecord, error) {
	record, err := scanIndicatorRecord(tx.QueryRow(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id,
    r.deleted_at,
    r.deleted_by_user_id
  FROM indicators i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorRecord{}, ErrIndicatorNotFound
	}
	if err != nil {
		return IndicatorRecord{}, fmt.Errorf("load indicator record: %w", err)
	}
	return record, nil
}

func (sourceRepository) validateIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE record_id = $1
       AND incident_id = $2
       AND record_type = 'indicator'
       AND deleted_at IS NULL
)
`, recordID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate indicator record incident: %w", err)
	}
	if !exists {
		return ErrIndicatorNotFound
	}
	return nil
}

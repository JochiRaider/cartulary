package indicators

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (sourceRepository) loadByDedupeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, indicatorType string, dedupeKey string) (indicatorRecord, bool, error) {
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
  FROM indicator_active_identities active_identity
  JOIN indicators i
    ON i.record_id = active_identity.indicator_record_id
  JOIN records r
    ON r.record_id = i.record_id
 WHERE active_identity.incident_id = $1
   AND active_identity.indicator_type = $2
   AND active_identity.dedupe_key = $3
   AND r.deleted_at IS NULL
 LIMIT 1
 FOR UPDATE OF active_identity, i, r
`, incidentID, indicatorType, dedupeKey))
	if errors.Is(err, pgx.ErrNoRows) {
		return indicatorRecord{}, false, nil
	}
	if err != nil {
		return indicatorRecord{}, false, fmt.Errorf("load indicator by dedupe: %w", err)
	}
	return record, true, nil
}

func (sourceRepository) insertTx(ctx context.Context, tx pgx.Tx, record *indicatorRecord) error {
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
    stix_pattern
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING record_id
`, record.RecordID, record.IncidentID, record.IndicatorType, record.ValueKind, record.DisplayValue, record.NormalizedValue, record.DedupeKey, record.DefangedValue, record.HashAlgorithm, record.HashValue, record.STIXPattern).Scan(&record.RecordID)
}

func (sourceRepository) updateTx(ctx context.Context, tx pgx.Tx, record indicatorRecord) error {
	tag, err := tx.Exec(ctx, `
UPDATE indicators
   SET defanged_value = $2,
       hash_algorithm = $3,
       hash_value = $4,
       stix_pattern = $5
 WHERE record_id = $1
`, record.RecordID, record.DefangedValue, record.HashAlgorithm, record.HashValue, record.STIXPattern)
	if err != nil {
		return fmt.Errorf("update indicator: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("update indicator affected %d rows", tag.RowsAffected())
	}
	return nil
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

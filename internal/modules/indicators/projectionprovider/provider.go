package projectionprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func RebuildIncidentIndicatorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM indicator_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear indicator projection rows: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO indicator_grid_projection (
    record_id,
    incident_id,
    row_version,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    first_observed_at,
    last_observed_at,
    observation_count,
    lifecycle_summary,
    supporting_link_count,
    edited_at
)
SELECT
    i.record_id,
    i.incident_id,
    r.row_version,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    obs.first_observed_at,
    obs.last_observed_at,
    COALESCE(obs.observation_count, 0),
    lifecycle.lifecycle_summary,
    COALESCE(links.supporting_link_count, 0),
    r.updated_at
  FROM indicators i
  JOIN records r
    ON r.record_id = i.record_id
  LEFT JOIN (
        SELECT
            resolved_indicator_record_id,
            MIN(created_at) AS first_observed_at,
            MAX(created_at) AS last_observed_at,
            COUNT(*) AS observation_count
          FROM indicator_observations
         WHERE resolution_status = 'resolved'
           AND resolved_indicator_record_id IS NOT NULL
           AND deleted_at IS NULL
         GROUP BY resolved_indicator_record_id
  ) obs
    ON obs.resolved_indicator_record_id = i.record_id
  LEFT JOIN (
        SELECT DISTINCT ON (indicator_record_id)
            indicator_record_id,
            lifecycle_state AS lifecycle_summary
          FROM indicator_state_intervals
         WHERE incident_id = $1
           AND deleted_at IS NULL
         ORDER BY indicator_record_id, CASE WHEN valid_to IS NULL THEN 0 ELSE 1 END ASC, valid_from DESC, indicator_state_interval_id DESC
  ) lifecycle
    ON lifecycle.indicator_record_id = i.record_id
  LEFT JOIN (
        SELECT dst_record_id, COUNT(*) AS supporting_link_count
          FROM active_record_links_v1
         GROUP BY dst_record_id
  ) links
    ON links.dst_record_id = i.record_id
 WHERE i.incident_id = $1
   AND r.deleted_at IS NULL
`, incidentID); err != nil {
		return fmt.Errorf("insert indicator projection rows: %w", err)
	}
	return nil
}

package projections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) RebuildIncidentHosts(ctx context.Context, incidentID uuid.UUID) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin host projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentHostsTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit host projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM host_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear host projection rows: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO host_grid_projection (
    record_id,
    incident_id,
    row_version,
    display_name,
    hostname,
    host_state,
    linked_event_count,
    evidence_count,
    location,
    os_platform,
    business_owner,
    criticality,
    containment_status,
    edited_at
)
SELECT
    h.record_id,
    h.incident_id,
    r.row_version,
    h.display_name,
    h.hostname,
    h.host_state,
    0,
    (
        SELECT COUNT(*)::integer
          FROM record_links l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = h.incident_id
           AND l.src_record_id = h.record_id
           AND l.link_type = 'supported_by'
           AND l.deleted_at IS NULL
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    NULL,
    NULL,
    NULL,
    NULL,
    NULL,
    r.updated_at
  FROM hosts h
  JOIN records r
    ON r.record_id = h.record_id
 WHERE h.incident_id = $1
   AND h.host_state IN ('stub', 'canonical')
`, incidentID); err != nil {
		return fmt.Errorf("insert host projection rows: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentIdentities(ctx context.Context, incidentID uuid.UUID) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin identity projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentIdentitiesTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit identity projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM identity_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear identity projection rows: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO identity_grid_projection (
    record_id,
    incident_id,
    row_version,
    display_name,
    upn,
    email,
    sam_account_name,
    identity_state,
    linked_event_count,
    evidence_count,
    privilege_level,
    mfa_state,
    reset_status,
    edited_at
)
SELECT
    i.record_id,
    i.incident_id,
    r.row_version,
    i.display_name,
    i.upn,
    i.email,
    i.sam_account_name,
    i.identity_state,
    0,
    (
        SELECT COUNT(*)::integer
          FROM record_links l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = i.incident_id
           AND l.src_record_id = i.record_id
           AND l.link_type = 'supported_by'
           AND l.deleted_at IS NULL
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    NULL,
    NULL,
    NULL,
    r.updated_at
  FROM identities i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.incident_id = $1
   AND i.identity_state IN ('stub', 'canonical')
`, incidentID); err != nil {
		return fmt.Errorf("insert identity projection rows: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentIndicators(ctx context.Context, incidentID uuid.UUID) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin indicator projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentIndicatorsTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit indicator projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentIndicatorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
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
         GROUP BY resolved_indicator_record_id
  ) obs
    ON obs.resolved_indicator_record_id = i.record_id
  LEFT JOIN (
        SELECT DISTINCT ON (indicator_record_id)
            indicator_record_id,
            lifecycle_state AS lifecycle_summary
          FROM indicator_state_intervals
         WHERE incident_id = $1
         ORDER BY indicator_record_id, CASE WHEN valid_to IS NULL THEN 0 ELSE 1 END ASC, valid_from DESC, indicator_state_interval_id DESC
  ) lifecycle
    ON lifecycle.indicator_record_id = i.record_id
  LEFT JOIN (
        SELECT dst_record_id, COUNT(*) AS supporting_link_count
          FROM record_links
         WHERE deleted_at IS NULL
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

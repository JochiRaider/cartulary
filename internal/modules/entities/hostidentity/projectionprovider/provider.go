package projectionprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func RefreshHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM host_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("clear host projection row: %w", err)
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
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN records source_record
            ON source_record.record_id = l.src_record_id
           AND source_record.record_type = 'timeline_event'
           AND source_record.deleted_at IS NULL
         WHERE l.incident_id = h.incident_id
           AND l.dst_record_id = h.record_id
           AND l.link_type = 'observed_on_host'
           AND l.deleted_at IS NULL
    ),
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = h.incident_id
           AND l.src_record_id = h.record_id
           AND l.link_type = 'attached_evidence'
           AND l.deleted_at IS NULL
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    h.location,
    h.os_platform,
    h.business_owner,
    h.criticality,
    h.containment_status,
    r.updated_at
  FROM hosts h
  JOIN records r
    ON r.record_id = h.record_id
 WHERE h.record_id = $1
   AND r.deleted_at IS NULL
   AND h.host_state IN ('stub', 'canonical')
`, recordID); err != nil {
		return fmt.Errorf("refresh host projection row: %w", err)
	}
	return nil
}

func RebuildIncidentHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
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
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN records source_record
            ON source_record.record_id = l.src_record_id
           AND source_record.record_type = 'timeline_event'
           AND source_record.deleted_at IS NULL
         WHERE l.incident_id = h.incident_id
           AND l.dst_record_id = h.record_id
           AND l.link_type = 'observed_on_host'
           AND l.deleted_at IS NULL
    ),
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = h.incident_id
           AND l.src_record_id = h.record_id
           AND l.link_type = 'attached_evidence'
           AND l.deleted_at IS NULL
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    h.location,
    h.os_platform,
    h.business_owner,
    h.criticality,
    h.containment_status,
    r.updated_at
  FROM hosts h
  JOIN records r
    ON r.record_id = h.record_id
 WHERE h.incident_id = $1
   AND r.deleted_at IS NULL
   AND h.host_state IN ('stub', 'canonical')
`, incidentID); err != nil {
		return fmt.Errorf("insert host projection rows: %w", err)
	}
	return nil
}

func RefreshIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM identity_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("clear identity projection row: %w", err)
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
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN records source_record
            ON source_record.record_id = l.src_record_id
           AND source_record.record_type = 'timeline_event'
           AND source_record.deleted_at IS NULL
         WHERE l.incident_id = i.incident_id
           AND l.dst_record_id = i.record_id
           AND l.link_type = 'observed_as_identity'
           AND l.deleted_at IS NULL
    ),
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = i.incident_id
           AND l.src_record_id = i.record_id
           AND l.link_type = 'attached_evidence'
           AND l.deleted_at IS NULL
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    i.privilege_level,
    i.mfa_state,
    i.reset_status,
    r.updated_at
  FROM identities i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.record_id = $1
   AND r.deleted_at IS NULL
   AND i.identity_state IN ('stub', 'canonical')
`, recordID); err != nil {
		return fmt.Errorf("refresh identity projection row: %w", err)
	}
	return nil
}

func RebuildIncidentIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
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
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN records source_record
            ON source_record.record_id = l.src_record_id
           AND source_record.record_type = 'timeline_event'
           AND source_record.deleted_at IS NULL
         WHERE l.incident_id = i.incident_id
           AND l.dst_record_id = i.record_id
           AND l.link_type = 'observed_as_identity'
           AND l.deleted_at IS NULL
    ),
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 l
          JOIN evidence ev
            ON ev.incident_id = l.incident_id
           AND ev.record_id = l.dst_record_id
          JOIN object_blobs b
            ON b.object_blob_id = ev.object_blob_id
         WHERE l.incident_id = i.incident_id
           AND l.src_record_id = i.record_id
           AND l.link_type = 'attached_evidence'
           AND l.deleted_at IS NULL
           AND ev.lifecycle_state IN ('available', 'released')
           AND b.upload_state = 'available'
    ),
    i.privilege_level,
    i.mfa_state,
    i.reset_status,
    r.updated_at
  FROM identities i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.incident_id = $1
   AND r.deleted_at IS NULL
   AND i.identity_state IN ('stub', 'canonical')
`, incidentID); err != nil {
		return fmt.Errorf("insert identity projection rows: %w", err)
	}
	return nil
}

package projections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const artifactProjectionColumns = `
    record_id,
    incident_id,
    row_version,
    artifact_type,
    title,
    body,
    timestamp_utc,
    updated_at,
    created_at,
    created_by_user_id,
    comm_id,
    comm_type,
    audience,
    channel_or_meeting,
    summary,
    next_report_at,
    privilege_tag,
    handoff_id,
    outgoing_owner_user_id,
    incoming_owner_user_id,
    current_state_summary,
    next_checks,
    acknowledged_at,
    status_review_id,
    review_owner_user_id,
    active_risks_summary,
    lesson_id,
    owner_user_id,
    closure_state,
    finding_statement,
    finding_kind,
    finding_state,
    finding_owner_user_id,
    finding_confidence_score,
    finding_closed_at,
    finding_updated_at,
    finding_confidence_band,
    investigative_query_query_id,
    investigative_query_platform,
    investigative_query_purpose,
    investigative_query_query_text,
    investigative_query_created_by_user_id,
    investigative_query_created_at,
    investigative_query_created_day,
    forensic_keyword_keyword_id,
    forensic_keyword_pattern,
    forensic_keyword_reason,
    forensic_keyword_match_mode,
    forensic_keyword_case_sensitive,
    forensic_keyword_created_at,
    forensic_keyword_created_day,
    timestamp_day,
    next_report_day,
    ack_state,
    linked_record_count`

const artifactProjectionSelect = `
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.artifact_type,
    a.title,
    a.body,
    a.timestamp_utc,
    a.updated_at,
    a.created_at,
    a.created_by_user_id,
    a.comm_id,
    a.comm_type,
    a.audience,
    a.channel_or_meeting,
    a.summary,
    a.next_report_at,
    a.privilege_tag,
    a.handoff_id,
    a.outgoing_owner_user_id,
    a.incoming_owner_user_id,
    a.current_state_summary,
    a.next_checks,
    a.acknowledged_at,
    a.status_review_id,
    a.review_owner_user_id,
    a.active_risks_summary,
    a.lesson_id,
    a.owner_user_id,
    a.closure_state,
    f.statement AS finding_statement,
    f.kind AS finding_kind,
    f.state AS finding_state,
    f.owner_user_id AS finding_owner_user_id,
    f.confidence_score AS finding_confidence_score,
    f.closed_at AS finding_closed_at,
    GREATEST(a.updated_at, f.updated_at) AS finding_updated_at,
    cartulary_confidence_band(f.confidence_score) AS finding_confidence_band,
    iq.query_id AS investigative_query_query_id,
    iq.platform AS investigative_query_platform,
    iq.purpose AS investigative_query_purpose,
    iq.query_text AS investigative_query_query_text,
    iq.created_by_user_id AS investigative_query_created_by_user_id,
    iq.created_at AS investigative_query_created_at,
    iq.created_at::date AS investigative_query_created_day,
    fk.keyword_id AS forensic_keyword_keyword_id,
    fk.pattern AS forensic_keyword_pattern,
    fk.reason AS forensic_keyword_reason,
    fk.match_mode AS forensic_keyword_match_mode,
    fk.case_sensitive AS forensic_keyword_case_sensitive,
    fk.created_at AS forensic_keyword_created_at,
    fk.created_at::date AS forensic_keyword_created_day,
    a.timestamp_utc::date AS timestamp_day,
    a.next_report_at::date AS next_report_day,
    CASE WHEN a.acknowledged_at IS NULL THEN 'pending' ELSE 'acknowledged' END AS ack_state,
    COALESCE(links.linked_record_count, 0)::integer AS linked_record_count
  FROM artifacts a
  JOIN records r
    ON r.incident_id = a.incident_id
   AND r.record_id = a.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN artifact_findings f
    ON f.incident_id = a.incident_id
   AND f.record_id = a.record_id
  LEFT JOIN artifact_investigative_queries iq
    ON iq.incident_id = a.incident_id
   AND iq.record_id = a.record_id
  LEFT JOIN artifact_forensic_keywords fk
    ON fk.incident_id = a.incident_id
   AND fk.record_id = a.record_id
  LEFT JOIN (
        SELECT incident_id, record_id, COUNT(*) AS linked_record_count
          FROM (
                SELECT incident_id, src_record_id AS record_id
                  FROM record_links
                 WHERE deleted_at IS NULL
                UNION ALL
                SELECT rl.incident_id, rl.dst_record_id AS record_id
                  FROM record_links rl
                  JOIN artifacts note_artifact
                    ON note_artifact.incident_id = rl.incident_id
                   AND note_artifact.record_id = rl.dst_record_id
                   AND note_artifact.artifact_type = 'note'
                 WHERE rl.deleted_at IS NULL
                   AND rl.link_type = 'references_artifact'
            ) counted_links
         GROUP BY incident_id, record_id
    ) links
    ON links.incident_id = a.incident_id
   AND links.record_id = a.record_id`

func (s *Store) RefreshArtifactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM artifact_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("clear artifact projection row: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO artifact_grid_projection (`+artifactProjectionColumns+`) `+artifactProjectionSelect+` WHERE a.record_id = $1`, recordID); err != nil {
		return fmt.Errorf("refresh artifact projection: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentArtifacts(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, notesViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentArtifactsTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentArtifactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM artifact_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear artifact projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO artifact_grid_projection (`+artifactProjectionColumns+`) `+artifactProjectionSelect+` WHERE a.incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("insert artifact projection rows: %w", err)
	}
	return nil
}

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

func (s *Store) RefreshEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM evidence_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("clear evidence projection row: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO evidence_grid_projection (`+evidenceProjectionColumns+`) `+evidenceProjectionSelect+` WHERE e.record_id = $1`, recordID); err != nil {
		return fmt.Errorf("refresh evidence projection: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentEvidence(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, evidenceViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentEvidenceTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentEvidenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM evidence_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear evidence projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO evidence_grid_projection (`+evidenceProjectionColumns+`) `+evidenceProjectionSelect+` WHERE e.incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("insert evidence projection rows: %w", err)
	}
	return nil
}

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

func (s *Store) RefreshPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM party_grid_projection WHERE record_id = $1`, recordID); err != nil {
		return fmt.Errorf("clear party projection row: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO party_grid_projection (`+partyProjectionColumns+`) `+partyProjectionSelect+` WHERE p.record_id = $1`, recordID); err != nil {
		return fmt.Errorf("refresh party projection: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentParties(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, partiesViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentPartiesTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentPartiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM party_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear party projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO party_grid_projection (`+partyProjectionColumns+`) `+partyProjectionSelect+` WHERE p.incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("insert party projection rows: %w", err)
	}
	return nil
}

func (s *Store) rebuildIncidentHotProjection(ctx context.Context, viewSchemaID string, rebuild func(context.Context, pgx.Tx) error) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, viewSchemaID)
	defer func() { finishTelemetry(err) }()

	if s == nil || s.pool == nil {
		return fmt.Errorf("rebuild hot projection: projection store is required")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin hot projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := rebuild(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit hot projection rebuild: %w", err)
	}
	return nil
}

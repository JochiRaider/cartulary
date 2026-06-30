-- +goose Up
DROP VIEW IF EXISTS artifact_grid_projection;

CREATE TABLE IF NOT EXISTS artifact_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES artifacts (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    artifact_type text NOT NULL,
    title text,
    body text,
    timestamp_utc timestamptz,
    updated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    created_by_user_id uuid REFERENCES users (id),
    comm_id text,
    comm_type text,
    audience text,
    channel_or_meeting text,
    summary text,
    next_report_at timestamptz,
    privilege_tag text,
    handoff_id text,
    outgoing_owner_user_id uuid REFERENCES users (id),
    incoming_owner_user_id uuid REFERENCES users (id),
    current_state_summary text,
    next_checks text,
    acknowledged_at timestamptz,
    status_review_id text,
    review_owner_user_id uuid REFERENCES users (id),
    active_risks_summary text,
    lesson_id text,
    owner_user_id uuid REFERENCES users (id),
    closure_state text,
    finding_statement text,
    finding_kind text,
    finding_state text,
    finding_owner_user_id uuid REFERENCES users (id),
    finding_confidence_score integer,
    finding_closed_at timestamptz,
    finding_updated_at timestamptz,
    finding_confidence_band text,
    investigative_query_query_id text,
    investigative_query_platform text,
    investigative_query_purpose text,
    investigative_query_query_text text,
    investigative_query_created_by_user_id uuid REFERENCES users (id),
    investigative_query_created_at timestamptz,
    investigative_query_created_day date,
    forensic_keyword_keyword_id text,
    forensic_keyword_pattern text,
    forensic_keyword_reason text,
    forensic_keyword_match_mode text,
    forensic_keyword_case_sensitive boolean,
    forensic_keyword_created_at timestamptz,
    forensic_keyword_created_day date,
    timestamp_day date,
    next_report_day date,
    ack_state text NOT NULL,
    linked_record_count integer NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS artifact_grid_projection_incident_type_updated_idx
    ON artifact_grid_projection (incident_id, artifact_type, updated_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_grid_projection_note_linked_idx
    ON artifact_grid_projection (incident_id, artifact_type, linked_record_count DESC, updated_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_grid_projection_finding_state_idx
    ON artifact_grid_projection (incident_id, artifact_type, finding_state, finding_kind, finding_owner_user_id, finding_updated_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_grid_projection_query_platform_idx
    ON artifact_grid_projection (incident_id, artifact_type, investigative_query_platform, investigative_query_created_by_user_id, investigative_query_created_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_grid_projection_keyword_mode_idx
    ON artifact_grid_projection (incident_id, artifact_type, forensic_keyword_match_mode, forensic_keyword_case_sensitive, forensic_keyword_created_at DESC, record_id ASC);

INSERT INTO artifact_grid_projection (
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
    linked_record_count
)
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
   AND links.record_id = a.record_id
ON CONFLICT (record_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS evidence_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES evidence (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    title text,
    lifecycle_state text NOT NULL,
    requested_at timestamptz,
    received_at timestamptz,
    storage_ref text,
    blob_hash text,
    collector_party_text text,
    collector_party_id uuid,
    source_party_text text,
    source_party_id uuid,
    upload_state text NOT NULL,
    linked_record_count integer NOT NULL DEFAULT 0,
    edited_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS evidence_grid_projection_incident_requested_idx
    ON evidence_grid_projection (incident_id, requested_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS evidence_grid_projection_lifecycle_idx
    ON evidence_grid_projection (incident_id, lifecycle_state, upload_state, edited_at DESC, record_id ASC);

INSERT INTO evidence_grid_projection (
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
    edited_at
)
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
    ON b.object_blob_id = e.object_blob_id
ON CONFLICT (record_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS party_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES parties (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    display_name text,
    party_kind text,
    organization_name text,
    role_title text,
    primary_email text,
    timezone_name text,
    external_ref text,
    notes text,
    updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS party_grid_projection_incident_display_name_idx
    ON party_grid_projection (incident_id, display_name ASC, record_id ASC);

CREATE INDEX IF NOT EXISTS party_grid_projection_kind_idx
    ON party_grid_projection (incident_id, party_kind, updated_at DESC, record_id ASC);

INSERT INTO party_grid_projection (
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
    updated_at
)
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
   AND r.deleted_at IS NULL
ON CONFLICT (record_id) DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS party_grid_projection_kind_idx;
DROP INDEX IF EXISTS party_grid_projection_incident_display_name_idx;
DROP TABLE IF EXISTS party_grid_projection;

DROP INDEX IF EXISTS evidence_grid_projection_lifecycle_idx;
DROP INDEX IF EXISTS evidence_grid_projection_incident_requested_idx;
DROP TABLE IF EXISTS evidence_grid_projection;

DROP INDEX IF EXISTS artifact_grid_projection_keyword_mode_idx;
DROP INDEX IF EXISTS artifact_grid_projection_query_platform_idx;
DROP INDEX IF EXISTS artifact_grid_projection_finding_state_idx;
DROP INDEX IF EXISTS artifact_grid_projection_note_linked_idx;
DROP INDEX IF EXISTS artifact_grid_projection_incident_type_updated_idx;
DROP TABLE IF EXISTS artifact_grid_projection;

CREATE OR REPLACE VIEW artifact_grid_projection AS
SELECT
    a.record_id,
    a.incident_id,
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
   AND links.record_id = a.record_id;

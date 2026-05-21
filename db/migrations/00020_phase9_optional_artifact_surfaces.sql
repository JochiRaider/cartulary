-- +goose Up
ALTER TABLE artifacts
    DROP CONSTRAINT IF EXISTS artifacts_artifact_type_check;

ALTER TABLE artifacts
    ADD CONSTRAINT artifacts_artifact_type_check CHECK (artifact_type IN (
        'note',
        'comm_log',
        'handoff',
        'status_review',
        'lesson',
        'finding',
        'investigative_query',
        'forensic_keyword'
    ));

CREATE TABLE IF NOT EXISTS artifact_findings (
    record_id uuid PRIMARY KEY REFERENCES artifacts (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('finding', 'hypothesis')),
    statement text NOT NULL,
    state text NOT NULL CHECK (state IN ('open', 'closed')),
    confidence_score integer CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 100)),
    owner_user_id uuid NOT NULL REFERENCES users (id),
    closed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT artifact_findings_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS artifact_investigative_queries (
    record_id uuid PRIMARY KEY REFERENCES artifacts (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    query_id text NOT NULL,
    platform text NOT NULL,
    purpose text NOT NULL,
    query_text text NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT artifact_investigative_queries_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE,
    CONSTRAINT artifact_investigative_queries_query_id_unique UNIQUE (incident_id, query_id)
);

CREATE TABLE IF NOT EXISTS artifact_forensic_keywords (
    record_id uuid PRIMARY KEY REFERENCES artifacts (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    keyword_id text NOT NULL,
    pattern text NOT NULL,
    reason text NOT NULL,
    match_mode text NOT NULL CHECK (match_mode IN ('literal', 'regex')),
    case_sensitive boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT artifact_forensic_keywords_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE,
    CONSTRAINT artifact_forensic_keywords_keyword_id_unique UNIQUE (incident_id, keyword_id)
);

CREATE INDEX IF NOT EXISTS artifact_findings_incident_state_idx
    ON artifact_findings (incident_id, state, kind, owner_user_id, updated_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_findings_incident_confidence_idx
    ON artifact_findings (incident_id, confidence_score, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_findings_incident_closed_idx
    ON artifact_findings (incident_id, closed_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_investigative_queries_incident_created_idx
    ON artifact_investigative_queries (incident_id, created_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_investigative_queries_incident_platform_idx
    ON artifact_investigative_queries (incident_id, platform, created_by_user_id, created_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_forensic_keywords_incident_created_idx
    ON artifact_forensic_keywords (incident_id, created_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS artifact_forensic_keywords_incident_mode_idx
    ON artifact_forensic_keywords (incident_id, match_mode, case_sensitive, created_at DESC, record_id ASC);

CREATE OR REPLACE FUNCTION cartulary_confidence_band(confidence_score integer)
RETURNS text
LANGUAGE sql
IMMUTABLE
AS $$
    SELECT CASE
        WHEN confidence_score IS NULL THEN 'unset'
        WHEN confidence_score BETWEEN 0 AND 39 THEN 'low'
        WHEN confidence_score BETWEEN 40 AND 69 THEN 'medium'
        WHEN confidence_score BETWEEN 70 AND 100 THEN 'high'
        ELSE NULL
    END
$$;

DROP VIEW IF EXISTS artifact_grid_projection;

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

-- +goose Down
DELETE FROM records
 WHERE record_type = 'artifact'
   AND record_id IN (
       SELECT record_id
         FROM artifacts
        WHERE artifact_type IN ('investigative_query', 'forensic_keyword')
   );

DROP VIEW IF EXISTS artifact_grid_projection;

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
    a.timestamp_utc::date AS timestamp_day,
    a.next_report_at::date AS next_report_day,
    CASE WHEN a.acknowledged_at IS NULL THEN 'pending' ELSE 'acknowledged' END AS ack_state,
    COALESCE(links.linked_record_count, 0)::integer AS linked_record_count
  FROM artifacts a
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

DROP TABLE IF EXISTS artifact_forensic_keywords;
DROP TABLE IF EXISTS artifact_investigative_queries;
DROP TABLE IF EXISTS artifact_findings;

ALTER TABLE artifacts
    DROP CONSTRAINT IF EXISTS artifacts_artifact_type_check;

ALTER TABLE artifacts
    ADD CONSTRAINT artifacts_artifact_type_check CHECK (artifact_type IN (
        'note',
        'comm_log',
        'handoff',
        'status_review',
        'lesson',
        'finding'
    ));

DROP FUNCTION IF EXISTS cartulary_confidence_band(integer);

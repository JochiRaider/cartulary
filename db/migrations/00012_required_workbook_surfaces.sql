-- +goose Up
CREATE TABLE IF NOT EXISTS evidence (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    title text,
    lifecycle_state text NOT NULL DEFAULT 'requested',
    requested_at timestamptz,
    received_at timestamptz,
    storage_ref text,
    blob_hash text,
    collector_party_text text,
    collector_party_id uuid,
    source_party_text text,
    source_party_id uuid,
    upload_state text NOT NULL DEFAULT 'pending',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT evidence_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS evidence_incident_sort_idx
    ON evidence (incident_id, requested_at DESC, record_id ASC);

CREATE TABLE IF NOT EXISTS parties (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    display_name text,
    party_kind text,
    organization_name text,
    role_title text,
    primary_email text,
    timezone_name text,
    external_ref text,
    notes text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT parties_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS parties_incident_display_name_idx
    ON parties (incident_id, display_name ASC, record_id ASC);

CREATE TABLE IF NOT EXISTS task_requests (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    title text,
    status text NOT NULL DEFAULT 'open',
    owner_user_id uuid REFERENCES users (id),
    priority text,
    task_kind text,
    workstream text,
    due_at timestamptz,
    requester_party_text text,
    requester_party_id uuid,
    blocked_reason text,
    completed_at timestamptz,
    external_ticket_ref text,
    closure_summary text,
    decision_record_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT task_requests_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS task_requests_incident_updated_idx
    ON task_requests (incident_id, updated_at DESC, record_id ASC);

CREATE TABLE IF NOT EXISTS decisions (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    summary text,
    status text NOT NULL DEFAULT 'proposed',
    owner_user_id uuid REFERENCES users (id),
    decision_type text,
    decided_at timestamptz,
    rationale text,
    supersedes_record_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT decisions_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS decisions_incident_decided_idx
    ON decisions (incident_id, decided_at DESC, record_id ASC);

CREATE TABLE IF NOT EXISTS artifacts (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    artifact_type text NOT NULL CHECK (artifact_type IN ('note', 'comm_log', 'handoff', 'status_review', 'lesson', 'finding')),
    title text,
    body text,
    timestamp_utc timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
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
    created_by_user_id uuid REFERENCES users (id),
    CONSTRAINT artifacts_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS artifacts_incident_type_updated_idx
    ON artifacts (incident_id, artifact_type, updated_at DESC, record_id ASC);

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
        SELECT incident_id, src_record_id AS record_id, COUNT(*) AS linked_record_count
          FROM record_links
         GROUP BY incident_id, src_record_id
    ) links
    ON links.incident_id = a.incident_id
   AND links.record_id = a.record_id;

-- +goose Down
DROP VIEW IF EXISTS artifact_grid_projection;
DROP INDEX IF EXISTS artifacts_incident_type_updated_idx;
DROP TABLE IF EXISTS artifacts;
DROP INDEX IF EXISTS decisions_incident_decided_idx;
DROP TABLE IF EXISTS decisions;
DROP INDEX IF EXISTS task_requests_incident_updated_idx;
DROP TABLE IF EXISTS task_requests;
DROP INDEX IF EXISTS parties_incident_display_name_idx;
DROP TABLE IF EXISTS parties;
DROP INDEX IF EXISTS evidence_incident_sort_idx;
DROP TABLE IF EXISTS evidence;

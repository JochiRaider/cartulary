-- +goose Up
CREATE TABLE IF NOT EXISTS task_request_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES task_requests (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    title text,
    status text NOT NULL,
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
    linked_record_count integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    no_owner boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS task_request_grid_projection_incident_updated_idx
    ON task_request_grid_projection (incident_id, updated_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS task_request_grid_projection_queue_idx
    ON task_request_grid_projection (incident_id, status, owner_user_id, priority, due_at, record_id);

CREATE INDEX IF NOT EXISTS task_request_grid_projection_due_idx
    ON task_request_grid_projection (incident_id, due_at, record_id);

CREATE INDEX IF NOT EXISTS task_request_grid_projection_no_owner_idx
    ON task_request_grid_projection (incident_id, no_owner, updated_at DESC, record_id ASC);

INSERT INTO task_request_grid_projection (
    record_id,
    incident_id,
    row_version,
    title,
    status,
    owner_user_id,
    priority,
    task_kind,
    workstream,
    due_at,
    requester_party_text,
    requester_party_id,
    blocked_reason,
    completed_at,
    external_ticket_ref,
    closure_summary,
    decision_record_id,
    linked_record_count,
    updated_at,
    no_owner
)
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.title,
    t.status,
    t.owner_user_id,
    t.priority,
    t.task_kind,
    t.workstream,
    t.due_at,
    t.requester_party_text,
    t.requester_party_id,
    t.blocked_reason,
    t.completed_at,
    t.external_ticket_ref,
    t.closure_summary,
    t.decision_record_id,
    COALESCE(linked.linked_record_count, 0)::integer,
    t.updated_at,
    t.owner_user_id IS NULL
  FROM task_requests t
  JOIN records r
    ON r.incident_id = t.incident_id
   AND r.record_id = t.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN (
        SELECT incident_id, src_record_id, COUNT(*) AS linked_record_count
          FROM record_links
         WHERE deleted_at IS NULL
           AND link_type = 'references_record'
           AND field_key = 'task.linked_record_ids'
         GROUP BY incident_id, src_record_id
  ) linked
    ON linked.incident_id = t.incident_id
   AND linked.src_record_id = t.record_id
ON CONFLICT (record_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS decision_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES decisions (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    summary text,
    status text NOT NULL,
    owner_user_id uuid REFERENCES users (id),
    decision_type text,
    decided_at timestamptz,
    rationale text,
    affected_record_count integer NOT NULL DEFAULT 0,
    supersedes_record_id uuid,
    updated_at timestamptz NOT NULL,
    is_superseded boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS decision_grid_projection_incident_decided_idx
    ON decision_grid_projection (incident_id, decided_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS decision_grid_projection_review_queue_idx
    ON decision_grid_projection (incident_id, status, owner_user_id, decision_type, decided_at, record_id);

CREATE INDEX IF NOT EXISTS decision_grid_projection_superseded_idx
    ON decision_grid_projection (incident_id, is_superseded, decided_at DESC, record_id ASC);

INSERT INTO decision_grid_projection (
    record_id,
    incident_id,
    row_version,
    summary,
    status,
    owner_user_id,
    decision_type,
    decided_at,
    rationale,
    affected_record_count,
    supersedes_record_id,
    updated_at,
    is_superseded
)
SELECT
    d.record_id,
    d.incident_id,
    r.row_version,
    d.summary,
    d.status,
    d.owner_user_id,
    d.decision_type,
    d.decided_at,
    d.rationale,
    COALESCE(affected.affected_record_count, 0)::integer,
    supersedes.supersedes_record_id,
    d.updated_at,
    COALESCE(incoming.is_superseded, false)
  FROM decisions d
  JOIN records r
    ON r.incident_id = d.incident_id
   AND r.record_id = d.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN (
        SELECT incident_id, src_record_id, COUNT(*) AS affected_record_count
          FROM record_links
         WHERE deleted_at IS NULL
           AND link_type = 'references_record'
           AND field_key = 'decision.affected_record_ids'
         GROUP BY incident_id, src_record_id
  ) affected
    ON affected.incident_id = d.incident_id
   AND affected.src_record_id = d.record_id
  LEFT JOIN LATERAL (
        SELECT rl.dst_record_id AS supersedes_record_id
          FROM record_links rl
          JOIN records dst
            ON dst.incident_id = rl.incident_id
           AND dst.record_id = rl.dst_record_id
           AND dst.record_type = 'decision'
           AND dst.deleted_at IS NULL
         WHERE rl.incident_id = d.incident_id
           AND rl.src_record_id = d.record_id
           AND rl.link_type = 'supersedes'
           AND rl.deleted_at IS NULL
         ORDER BY rl.created_at DESC, rl.record_link_id DESC
         LIMIT 1
  ) supersedes ON true
  LEFT JOIN LATERAL (
        SELECT true AS is_superseded
          FROM record_links rl
          JOIN records src
            ON src.incident_id = rl.incident_id
           AND src.record_id = rl.src_record_id
           AND src.record_type = 'decision'
           AND src.deleted_at IS NULL
         WHERE rl.incident_id = d.incident_id
           AND rl.dst_record_id = d.record_id
           AND rl.link_type = 'supersedes'
           AND rl.deleted_at IS NULL
         LIMIT 1
  ) incoming ON true
ON CONFLICT (record_id) DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS decision_grid_projection_superseded_idx;
DROP INDEX IF EXISTS decision_grid_projection_review_queue_idx;
DROP INDEX IF EXISTS decision_grid_projection_incident_decided_idx;
DROP TABLE IF EXISTS decision_grid_projection;

DROP INDEX IF EXISTS task_request_grid_projection_no_owner_idx;
DROP INDEX IF EXISTS task_request_grid_projection_due_idx;
DROP INDEX IF EXISTS task_request_grid_projection_queue_idx;
DROP INDEX IF EXISTS task_request_grid_projection_incident_updated_idx;
DROP TABLE IF EXISTS task_request_grid_projection;

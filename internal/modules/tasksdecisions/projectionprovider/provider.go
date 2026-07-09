package projectionprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func RefreshTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
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
          FROM active_record_links_v1
         WHERE deleted_at IS NULL
           AND link_type = 'references_record'
           AND field_key = 'task.linked_record_ids'
           AND src_record_id = $1
         GROUP BY incident_id, src_record_id
  ) linked
    ON linked.incident_id = t.incident_id
   AND linked.src_record_id = t.record_id
 WHERE t.record_id = $1
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    title = EXCLUDED.title,
    status = EXCLUDED.status,
    owner_user_id = EXCLUDED.owner_user_id,
    priority = EXCLUDED.priority,
    task_kind = EXCLUDED.task_kind,
    workstream = EXCLUDED.workstream,
    due_at = EXCLUDED.due_at,
    requester_party_text = EXCLUDED.requester_party_text,
    requester_party_id = EXCLUDED.requester_party_id,
    blocked_reason = EXCLUDED.blocked_reason,
    completed_at = EXCLUDED.completed_at,
    external_ticket_ref = EXCLUDED.external_ticket_ref,
    closure_summary = EXCLUDED.closure_summary,
    decision_record_id = EXCLUDED.decision_record_id,
    linked_record_count = EXCLUDED.linked_record_count,
    updated_at = EXCLUDED.updated_at,
    no_owner = EXCLUDED.no_owner
`, recordID); err != nil {
		return fmt.Errorf("refresh task request projection: %w", err)
	}
	return nil
}

func RefreshDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
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
          FROM active_record_links_v1
         WHERE deleted_at IS NULL
           AND link_type = 'references_record'
           AND field_key = 'decision.affected_record_ids'
           AND src_record_id = $1
         GROUP BY incident_id, src_record_id
  ) affected
    ON affected.incident_id = d.incident_id
   AND affected.src_record_id = d.record_id
  LEFT JOIN LATERAL (
        SELECT rl.dst_record_id AS supersedes_record_id
          FROM active_record_links_v1 rl
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
          FROM active_record_links_v1 rl
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
 WHERE d.record_id = $1
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    summary = EXCLUDED.summary,
    status = EXCLUDED.status,
    owner_user_id = EXCLUDED.owner_user_id,
    decision_type = EXCLUDED.decision_type,
    decided_at = EXCLUDED.decided_at,
    rationale = EXCLUDED.rationale,
    affected_record_count = EXCLUDED.affected_record_count,
    supersedes_record_id = EXCLUDED.supersedes_record_id,
    updated_at = EXCLUDED.updated_at,
    is_superseded = EXCLUDED.is_superseded
`, recordID); err != nil {
		return fmt.Errorf("refresh decision projection: %w", err)
	}
	return nil
}

func RebuildIncidentTaskRequestsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM task_request_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear task request projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `
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
          FROM active_record_links_v1
         WHERE deleted_at IS NULL
           AND link_type = 'references_record'
           AND field_key = 'task.linked_record_ids'
           AND incident_id = $1
         GROUP BY incident_id, src_record_id
  ) linked
    ON linked.incident_id = t.incident_id
   AND linked.src_record_id = t.record_id
 WHERE t.incident_id = $1
`, incidentID); err != nil {
		return fmt.Errorf("insert task request projection rows: %w", err)
	}
	return nil
}

func RebuildIncidentDecisionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM decision_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear decision projection rows: %w", err)
	}
	if _, err := tx.Exec(ctx, `
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
          FROM active_record_links_v1
         WHERE deleted_at IS NULL
           AND link_type = 'references_record'
           AND field_key = 'decision.affected_record_ids'
           AND incident_id = $1
         GROUP BY incident_id, src_record_id
  ) affected
    ON affected.incident_id = d.incident_id
   AND affected.src_record_id = d.record_id
  LEFT JOIN LATERAL (
        SELECT rl.dst_record_id AS supersedes_record_id
          FROM active_record_links_v1 rl
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
          FROM active_record_links_v1 rl
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
 WHERE d.incident_id = $1
`, incidentID); err != nil {
		return fmt.Errorf("insert decision projection rows: %w", err)
	}
	return nil
}

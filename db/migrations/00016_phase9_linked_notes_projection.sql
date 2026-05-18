-- +goose Up
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

-- +goose Down
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
         WHERE deleted_at IS NULL
         GROUP BY incident_id, src_record_id
    ) links
    ON links.incident_id = a.incident_id
   AND links.record_id = a.record_id;

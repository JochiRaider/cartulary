-- name: ListTimelineProjectionRows :many
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.occurred_at,
    t.summary,
    t.details,
    t.source_text,
    t.recorded_at,
    t.edited_at,
    t.sort_ts,
    t.capture_state,
    t.replacement_record_id,
    t.occurred_day,
    t.recorded_day,
    t.evidence_count,
    t.has_evidence,
    t.has_unresolved_mentions
FROM timeline_grid_projection t
JOIN records r ON r.record_id = t.record_id
WHERE t.incident_id = $1
ORDER BY t.sort_ts ASC, t.record_id ASC;

-- name: GetTimelineProjectionRow :one
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.occurred_at,
    t.summary,
    t.details,
    t.source_text,
    t.recorded_at,
    t.edited_at,
    t.sort_ts,
    t.capture_state,
    t.replacement_record_id,
    t.occurred_day,
    t.recorded_day,
    t.evidence_count,
    t.has_evidence,
    t.has_unresolved_mentions
FROM timeline_grid_projection t
JOIN records r ON r.record_id = t.record_id
WHERE t.record_id = $1;

-- name: ListTimelineProjectionSourceRows :many
SELECT
    e.record_id,
    e.incident_id,
    r.row_version,
    e.occurred_at,
    e.summary,
    e.details,
    e.source_text,
    e.recorded_at,
    e.edited_at,
    COALESCE(e.occurred_at, e.recorded_at) AS sort_ts,
    e.capture_state,
    (
        SELECT l.src_record_id
        FROM record_links l
        WHERE l.dst_record_id = e.record_id
          AND l.link_type = 'supersedes'
          AND l.deleted_at IS NULL
        ORDER BY l.created_at DESC, l.record_link_id DESC
        LIMIT 1
    ) AS replacement_record_id,
    e.occurred_at::date AS occurred_day,
    e.recorded_at::date AS recorded_day,
    0::integer AS evidence_count,
    false::boolean AS has_evidence,
    EXISTS (
        SELECT 1
        FROM entity_mentions em
        WHERE em.source_record_id = e.record_id
          AND em.resolution_status = 'unresolved'
    ) AS has_unresolved_mentions
FROM timeline_events e
JOIN records r ON r.record_id = e.record_id
WHERE e.incident_id = $1
ORDER BY COALESCE(e.occurred_at, e.recorded_at) ASC, e.record_id ASC;

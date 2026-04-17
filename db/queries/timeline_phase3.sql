-- name: ListTimelineProjectionRows :many
SELECT
    record_id,
    incident_id,
    row_version,
    occurred_at,
    summary,
    details,
    source_text,
    recorded_at,
    edited_at,
    sort_ts,
    capture_state,
    replacement_record_id,
    occurred_day,
    recorded_day,
    evidence_count,
    has_evidence,
    has_unresolved_mentions
FROM timeline_grid_projection
WHERE incident_id = $1
ORDER BY sort_ts ASC, record_id ASC;

-- name: GetTimelineProjectionRow :one
SELECT
    record_id,
    incident_id,
    row_version,
    occurred_at,
    summary,
    details,
    source_text,
    recorded_at,
    edited_at,
    sort_ts,
    capture_state,
    replacement_record_id,
    occurred_day,
    recorded_day,
    evidence_count,
    has_evidence,
    has_unresolved_mentions
FROM timeline_grid_projection
WHERE record_id = $1;

-- name: ListTimelineProjectionSourceRows :many
SELECT
    e.record_id,
    e.incident_id,
    e.row_version,
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
    false::boolean AS has_unresolved_mentions
FROM timeline_events e
WHERE e.incident_id = $1
ORDER BY COALESCE(e.occurred_at, e.recorded_at) ASC, e.record_id ASC;

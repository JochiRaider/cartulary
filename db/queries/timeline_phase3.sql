-- name: ListTimelineProjectionRows :many
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.date_entered_text,
    t.analyst_text,
    t.mitre_stage_text,
    t.device_object_text,
    t.ip_address_text,
    t.activity_utc_text,
    t.activity_local_text,
    t.raw_activity_text,
    t.activity_synopsis_text,
    t.data_source_text,
    t.recorded_at,
    t.edited_at,
    t.activity_sort_ts,
    t.date_entered_sort_day,
    t.activity_time_pair_state,
    t.capture_state,
    t.replacement_record_id,
    t.evidence_count,
    t.has_evidence,
    t.has_unresolved_mentions
FROM timeline_grid_projection t
JOIN records r ON r.record_id = t.record_id
WHERE t.incident_id = $1
ORDER BY t.activity_sort_ts ASC NULLS LAST, t.record_id ASC;

-- name: GetTimelineProjectionRow :one
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.date_entered_text,
    t.analyst_text,
    t.mitre_stage_text,
    t.device_object_text,
    t.ip_address_text,
    t.activity_utc_text,
    t.activity_local_text,
    t.raw_activity_text,
    t.activity_synopsis_text,
    t.data_source_text,
    t.recorded_at,
    t.edited_at,
    t.activity_sort_ts,
    t.date_entered_sort_day,
    t.activity_time_pair_state,
    t.capture_state,
    t.replacement_record_id,
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
    e.date_entered_text,
    e.analyst_text,
    e.mitre_stage_text,
    e.device_object_text,
    e.ip_address_text,
    e.activity_utc_text,
    e.activity_local_text,
    e.raw_activity_text,
    e.activity_synopsis_text,
    e.data_source_text,
    e.recorded_at,
    e.edited_at,
    NULL::timestamptz AS activity_sort_ts,
    NULL::date AS date_entered_sort_day,
    e.activity_time_pair_state,
    e.capture_state,
    (
        SELECT l.src_record_id
        FROM active_record_links_v1 l
        WHERE l.dst_record_id = e.record_id
          AND l.link_type = 'supersedes'
        ORDER BY l.created_at DESC, l.record_link_id DESC
        LIMIT 1
    ) AS replacement_record_id,
    (
        SELECT COUNT(*)::integer
        FROM active_record_links_v1 l
        JOIN evidence ev
          ON ev.incident_id = l.incident_id
         AND ev.record_id = l.dst_record_id
        JOIN object_blobs b
          ON b.object_blob_id = ev.object_blob_id
        WHERE l.incident_id = e.incident_id
          AND l.src_record_id = e.record_id
          AND l.link_type = 'attached_evidence'
          AND ev.lifecycle_state IN ('available', 'released')
          AND b.upload_state = 'available'
    ) AS evidence_count,
    EXISTS (
        SELECT 1
        FROM active_record_links_v1 l
        JOIN evidence ev
          ON ev.incident_id = l.incident_id
         AND ev.record_id = l.dst_record_id
        JOIN object_blobs b
          ON b.object_blob_id = ev.object_blob_id
        WHERE l.incident_id = e.incident_id
          AND l.src_record_id = e.record_id
          AND l.link_type = 'attached_evidence'
          AND ev.lifecycle_state IN ('available', 'released')
          AND b.upload_state = 'available'
    ) AS has_evidence,
    EXISTS (
        SELECT 1
        FROM entity_mentions em
        WHERE em.source_record_id = e.record_id
          AND em.resolution_status = 'unresolved'
    ) AS has_unresolved_mentions
FROM timeline_events e
JOIN records r ON r.record_id = e.record_id
WHERE e.incident_id = $1
  AND r.deleted_at IS NULL
ORDER BY e.recorded_at ASC, e.record_id ASC;

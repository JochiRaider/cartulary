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

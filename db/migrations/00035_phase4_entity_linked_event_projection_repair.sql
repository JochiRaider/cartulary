-- +goose Up
WITH host_link_counts AS (
    SELECT
        h.record_id,
        COUNT(source_record.record_id)::integer AS linked_event_count
      FROM hosts h
      JOIN records target_record
        ON target_record.record_id = h.record_id
       AND target_record.deleted_at IS NULL
      LEFT JOIN record_links l
        ON l.incident_id = h.incident_id
       AND l.dst_record_id = h.record_id
       AND l.link_type = 'observed_on_host'
       AND l.deleted_at IS NULL
      LEFT JOIN records source_record
        ON source_record.record_id = l.src_record_id
       AND source_record.record_type = 'timeline_event'
       AND source_record.deleted_at IS NULL
     WHERE h.host_state IN ('stub', 'canonical')
     GROUP BY h.record_id
)
UPDATE host_grid_projection projection
   SET linked_event_count = host_link_counts.linked_event_count
  FROM host_link_counts
 WHERE projection.record_id = host_link_counts.record_id
   AND projection.linked_event_count IS DISTINCT FROM host_link_counts.linked_event_count;

WITH identity_link_counts AS (
    SELECT
        i.record_id,
        COUNT(source_record.record_id)::integer AS linked_event_count
      FROM identities i
      JOIN records target_record
        ON target_record.record_id = i.record_id
       AND target_record.deleted_at IS NULL
      LEFT JOIN record_links l
        ON l.incident_id = i.incident_id
       AND l.dst_record_id = i.record_id
       AND l.link_type = 'observed_as_identity'
       AND l.deleted_at IS NULL
      LEFT JOIN records source_record
        ON source_record.record_id = l.src_record_id
       AND source_record.record_type = 'timeline_event'
       AND source_record.deleted_at IS NULL
     WHERE i.identity_state IN ('stub', 'canonical')
     GROUP BY i.record_id
)
UPDATE identity_grid_projection projection
   SET linked_event_count = identity_link_counts.linked_event_count
  FROM identity_link_counts
 WHERE projection.record_id = identity_link_counts.record_id
   AND projection.linked_event_count IS DISTINCT FROM identity_link_counts.linked_event_count;

-- +goose Down
-- The previous cached projection values are not recoverable or desirable.
SELECT 1;

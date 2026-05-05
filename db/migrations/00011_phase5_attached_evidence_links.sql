-- +goose Up

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity', 'supported_by', 'references_record', 'attached_evidence'));

WITH candidates AS (
    SELECT
        rl.record_link_id,
        rl.incident_id,
        rl.src_record_id,
        rl.dst_record_id,
        CASE
            WHEN src.record_type = 'timeline_event' THEN 'timeline.attached_evidence_ids'
            ELSE rl.field_key
        END AS next_field_key
      FROM record_links rl
      JOIN records src
        ON src.incident_id = rl.incident_id
       AND src.record_id = rl.src_record_id
       AND src.deleted_at IS NULL
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.link_type = 'supported_by'
       AND rl.deleted_at IS NULL
       AND src.record_type IN ('timeline_event', 'host', 'identity')
       AND dst.record_type = 'evidence'
),
safe_candidates AS (
    SELECT c.*
      FROM candidates c
     WHERE NOT EXISTS (
        SELECT 1
          FROM record_links existing
         WHERE existing.incident_id = c.incident_id
           AND existing.src_record_id = c.src_record_id
           AND existing.dst_record_id = c.dst_record_id
           AND existing.link_type = 'attached_evidence'
           AND existing.deleted_at IS NULL
           AND existing.field_key IS NOT DISTINCT FROM c.next_field_key
     )
       AND 1 = (
        SELECT COUNT(*)
          FROM candidates peer
         WHERE peer.incident_id = c.incident_id
           AND peer.src_record_id = c.src_record_id
           AND peer.dst_record_id = c.dst_record_id
           AND peer.next_field_key IS NOT DISTINCT FROM c.next_field_key
       )
)
UPDATE record_links rl
   SET link_type = 'attached_evidence',
       field_key = sc.next_field_key
  FROM safe_candidates sc
 WHERE rl.record_link_id = sc.record_link_id;

-- +goose Down

UPDATE record_links rl
   SET link_type = 'supported_by',
       field_key = NULL
  FROM records src,
       records dst
 WHERE rl.incident_id = src.incident_id
   AND rl.src_record_id = src.record_id
   AND rl.incident_id = dst.incident_id
   AND rl.dst_record_id = dst.record_id
   AND rl.link_type = 'attached_evidence'
   AND rl.deleted_at IS NULL
   AND src.record_type IN ('timeline_event', 'host', 'identity')
   AND dst.record_type = 'evidence'
   AND NOT EXISTS (
        SELECT 1
          FROM record_links existing
         WHERE existing.incident_id = rl.incident_id
           AND existing.src_record_id = rl.src_record_id
           AND existing.dst_record_id = rl.dst_record_id
           AND existing.link_type = 'supported_by'
           AND existing.deleted_at IS NULL
           AND existing.field_key IS NULL
   )
   AND 1 = (
        SELECT COUNT(*)
          FROM record_links peer
         WHERE peer.incident_id = rl.incident_id
           AND peer.src_record_id = rl.src_record_id
           AND peer.dst_record_id = rl.dst_record_id
           AND peer.link_type = 'attached_evidence'
           AND peer.deleted_at IS NULL
   );

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity', 'supported_by', 'references_record'));

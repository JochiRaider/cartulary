-- +goose Up
ALTER TABLE decisions
    DROP COLUMN IF EXISTS supersedes_record_id;

-- +goose Down
ALTER TABLE decisions
    ADD COLUMN IF NOT EXISTS supersedes_record_id uuid;

UPDATE decisions d
   SET supersedes_record_id = supersedes.dst_record_id
  FROM (
        SELECT DISTINCT ON (rl.src_record_id)
               rl.src_record_id,
               rl.dst_record_id
          FROM record_links rl
          JOIN records dst
            ON dst.incident_id = rl.incident_id
           AND dst.record_id = rl.dst_record_id
           AND dst.record_type = 'decision'
           AND dst.deleted_at IS NULL
         WHERE rl.link_type = 'supersedes'
           AND rl.deleted_at IS NULL
         ORDER BY rl.src_record_id, rl.created_at DESC, rl.record_link_id DESC
  ) supersedes
 WHERE d.record_id = supersedes.src_record_id;

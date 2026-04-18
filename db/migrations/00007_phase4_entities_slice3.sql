-- +goose Up
ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check,
    DROP CONSTRAINT IF EXISTS record_links_dst_record_id_fkey;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity'));

CREATE INDEX IF NOT EXISTS record_links_active_src_lookup_idx
    ON record_links (incident_id, src_record_id, link_type, dst_record_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS record_links_active_src_lookup_idx;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes'));

ALTER TABLE record_links
    ADD CONSTRAINT record_links_dst_record_id_fkey FOREIGN KEY (dst_record_id) REFERENCES timeline_events (record_id) ON DELETE CASCADE;

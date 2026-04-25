-- +goose Up
ALTER TABLE record_links
    ADD COLUMN IF NOT EXISTS field_key text;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity', 'supported_by', 'references_record'));

DROP INDEX IF EXISTS record_links_active_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS record_links_active_unique_no_field_idx
    ON record_links (incident_id, src_record_id, dst_record_id, link_type)
    WHERE deleted_at IS NULL AND field_key IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS record_links_active_unique_field_idx
    ON record_links (incident_id, src_record_id, dst_record_id, link_type, field_key)
    WHERE deleted_at IS NULL AND field_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS handoff_risk_refs (
    risk_ref_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    handoff_record_id uuid NOT NULL,
    risk_ref_text text NOT NULL,
    normalized_risk_ref_text text NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users (id),
    CONSTRAINT handoff_risk_refs_handoff_fkey
        FOREIGN KEY (incident_id, handoff_record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE,
    CONSTRAINT handoff_risk_refs_delete_state_ck CHECK (
        (deleted_at IS NULL AND deleted_by_user_id IS NULL)
        OR (deleted_at IS NOT NULL AND deleted_by_user_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS handoff_risk_refs_active_text_idx
    ON handoff_risk_refs (handoff_record_id, normalized_risk_ref_text)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS handoff_risk_refs_active_text_idx;
DROP TABLE IF EXISTS handoff_risk_refs;

DROP INDEX IF EXISTS record_links_active_unique_field_idx;
DROP INDEX IF EXISTS record_links_active_unique_no_field_idx;

CREATE UNIQUE INDEX IF NOT EXISTS record_links_active_unique_idx
    ON record_links (incident_id, src_record_id, dst_record_id, link_type)
    WHERE deleted_at IS NULL;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity', 'supported_by'));

ALTER TABLE record_links
    DROP COLUMN IF EXISTS field_key;

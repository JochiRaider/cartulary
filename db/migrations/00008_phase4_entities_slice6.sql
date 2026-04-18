-- +goose Up
ALTER TABLE hosts
    ADD COLUMN IF NOT EXISTS merged_into_record_id uuid REFERENCES hosts (record_id);

ALTER TABLE hosts
    DROP CONSTRAINT IF EXISTS hosts_merge_lineage_ck;

ALTER TABLE hosts
    ADD CONSTRAINT hosts_merge_lineage_ck CHECK (
        (host_state IN ('stub', 'canonical') AND merged_into_record_id IS NULL)
        OR (host_state = 'merged' AND merged_into_record_id IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS hosts_incident_merged_into_idx
    ON hosts (incident_id, merged_into_record_id, record_id)
    WHERE merged_into_record_id IS NOT NULL;

ALTER TABLE identities
    ADD COLUMN IF NOT EXISTS merged_into_record_id uuid REFERENCES identities (record_id);

ALTER TABLE identities
    DROP CONSTRAINT IF EXISTS identities_merge_lineage_ck;

ALTER TABLE identities
    ADD CONSTRAINT identities_merge_lineage_ck CHECK (
        (identity_state IN ('stub', 'canonical') AND merged_into_record_id IS NULL)
        OR (identity_state = 'merged' AND merged_into_record_id IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS identities_incident_merged_into_idx
    ON identities (incident_id, merged_into_record_id, record_id)
    WHERE merged_into_record_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS record_tags (
    record_tag_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    record_id uuid NOT NULL,
    tag_name text NOT NULL,
    normalized_tag_name text NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS record_tags_active_unique_idx
    ON record_tags (incident_id, record_id, normalized_tag_name)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS record_tags_active_record_lookup_idx
    ON record_tags (incident_id, record_id, normalized_tag_name)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS compromise_assessments (
    compromise_assessment_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    subject_id uuid NOT NULL,
    subject_type text NOT NULL CHECK (subject_type IN ('host', 'identity')),
    state text NOT NULL,
    confidence integer CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 100)),
    assessed_by_user_id uuid REFERENCES users (id),
    assessed_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users (id)
);

CREATE INDEX IF NOT EXISTS compromise_assessments_active_subject_lookup_idx
    ON compromise_assessments (incident_id, subject_type, subject_id, compromise_assessment_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS compromise_assessments_active_subject_lookup_idx;
DROP TABLE IF EXISTS compromise_assessments;

DROP INDEX IF EXISTS record_tags_active_record_lookup_idx;
DROP INDEX IF EXISTS record_tags_active_unique_idx;
DROP TABLE IF EXISTS record_tags;

DROP INDEX IF EXISTS identities_incident_merged_into_idx;
ALTER TABLE identities
    DROP CONSTRAINT IF EXISTS identities_merge_lineage_ck;
ALTER TABLE identities
    DROP COLUMN IF EXISTS merged_into_record_id;

DROP INDEX IF EXISTS hosts_incident_merged_into_idx;
ALTER TABLE hosts
    DROP CONSTRAINT IF EXISTS hosts_merge_lineage_ck;
ALTER TABLE hosts
    DROP COLUMN IF EXISTS merged_into_record_id;

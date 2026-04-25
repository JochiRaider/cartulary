-- +goose Up
CREATE TABLE IF NOT EXISTS records (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    record_type text NOT NULL CHECK (record_type IN (
        'timeline_event',
        'host',
        'identity',
        'party',
        'indicator',
        'artifact',
        'task_request',
        'decision',
        'evidence',
        'assessment'
    )),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    row_version bigint NOT NULL DEFAULT 1 CHECK (row_version >= 1),
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users (id),
    CONSTRAINT records_delete_state_ck CHECK (
        (deleted_at IS NULL AND deleted_by_user_id IS NULL)
        OR (deleted_at IS NOT NULL AND deleted_by_user_id IS NOT NULL)
    ),
    UNIQUE (incident_id, record_id)
);

CREATE INDEX IF NOT EXISTS records_incident_lookup_idx
    ON records (incident_id, record_id);

CREATE INDEX IF NOT EXISTS records_active_incident_type_idx
    ON records (incident_id, record_type, updated_at DESC, record_id ASC)
    WHERE deleted_at IS NULL;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM compromise_assessments
        WHERE assessed_by_user_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot backfill assessment records without assessed_by_user_id attribution';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM compromise_assessments
        WHERE deleted_at IS NOT NULL
          AND deleted_by_user_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot backfill deleted assessment records without deleted_by_user_id attribution';
    END IF;
END $$;
-- +goose StatementEnd

INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version
)
SELECT
    record_id,
    incident_id,
    'timeline_event',
    created_by_user_id,
    recorded_at,
    updated_by_user_id,
    edited_at,
    row_version
FROM timeline_events
ON CONFLICT (record_id) DO NOTHING;

INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version
)
SELECT
    record_id,
    incident_id,
    'host',
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version
FROM hosts
ON CONFLICT (record_id) DO NOTHING;

INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version
)
SELECT
    record_id,
    incident_id,
    'identity',
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version
FROM identities
ON CONFLICT (record_id) DO NOTHING;

INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version,
    deleted_at,
    deleted_by_user_id
)
SELECT
    record_id,
    incident_id,
    'indicator',
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version,
    deleted_at,
    deleted_by_user_id
FROM indicators
ON CONFLICT (record_id) DO NOTHING;

INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version,
    deleted_at,
    deleted_by_user_id
)
SELECT
    compromise_assessment_id,
    incident_id,
    'assessment',
    assessed_by_user_id,
    created_at,
    assessed_by_user_id,
    updated_at,
    1,
    deleted_at,
    deleted_by_user_id
FROM compromise_assessments
ON CONFLICT (record_id) DO NOTHING;

ALTER TABLE record_links
    ADD COLUMN IF NOT EXISTS created_by_user_id uuid REFERENCES users (id);

UPDATE record_links
   SET created_by_user_id = owner_user_id
 WHERE created_by_user_id IS NULL;

ALTER TABLE record_links
    ALTER COLUMN created_by_user_id SET NOT NULL;

ALTER TABLE record_revisions
    DROP CONSTRAINT IF EXISTS record_revisions_record_id_fkey;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_src_record_id_fkey,
    DROP CONSTRAINT IF EXISTS record_links_dst_record_id_fkey;

ALTER TABLE entity_mentions
    DROP CONSTRAINT IF EXISTS entity_mentions_source_record_id_fkey;

ALTER TABLE indicator_observations
    DROP CONSTRAINT IF EXISTS indicator_observations_source_record_id_fkey;

ALTER TABLE timeline_events
    ADD CONSTRAINT timeline_events_record_envelope_fkey
    FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

ALTER TABLE hosts
    ADD CONSTRAINT hosts_record_envelope_fkey
    FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

ALTER TABLE identities
    ADD CONSTRAINT identities_record_envelope_fkey
    FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

ALTER TABLE indicators
    ADD CONSTRAINT indicators_record_envelope_fkey
    FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

ALTER TABLE compromise_assessments
    ADD CONSTRAINT compromise_assessments_record_envelope_fkey
    FOREIGN KEY (incident_id, compromise_assessment_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

ALTER TABLE record_revisions
    ADD CONSTRAINT record_revisions_record_id_fkey
    FOREIGN KEY (record_id) REFERENCES records (record_id) ON DELETE CASCADE;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_src_record_envelope_fkey
    FOREIGN KEY (incident_id, src_record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE,
    ADD CONSTRAINT record_links_dst_record_envelope_fkey
    FOREIGN KEY (incident_id, dst_record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

ALTER TABLE record_tags
    ADD CONSTRAINT record_tags_record_envelope_fkey
    FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

ALTER TABLE entity_mentions
    ADD CONSTRAINT entity_mentions_source_record_envelope_fkey
    FOREIGN KEY (source_record_id) REFERENCES records (record_id) ON DELETE CASCADE;

ALTER TABLE indicator_observations
    ADD CONSTRAINT indicator_observations_source_record_envelope_fkey
    FOREIGN KEY (incident_id, source_record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE indicator_observations
    DROP CONSTRAINT IF EXISTS indicator_observations_source_record_envelope_fkey;

ALTER TABLE entity_mentions
    DROP CONSTRAINT IF EXISTS entity_mentions_source_record_envelope_fkey;

ALTER TABLE record_tags
    DROP CONSTRAINT IF EXISTS record_tags_record_envelope_fkey;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_src_record_envelope_fkey,
    DROP CONSTRAINT IF EXISTS record_links_dst_record_envelope_fkey;

ALTER TABLE record_revisions
    DROP CONSTRAINT IF EXISTS record_revisions_record_id_fkey;

ALTER TABLE compromise_assessments
    DROP CONSTRAINT IF EXISTS compromise_assessments_record_envelope_fkey;

ALTER TABLE indicators
    DROP CONSTRAINT IF EXISTS indicators_record_envelope_fkey;

ALTER TABLE identities
    DROP CONSTRAINT IF EXISTS identities_record_envelope_fkey;

ALTER TABLE hosts
    DROP CONSTRAINT IF EXISTS hosts_record_envelope_fkey;

ALTER TABLE timeline_events
    DROP CONSTRAINT IF EXISTS timeline_events_record_envelope_fkey;

ALTER TABLE indicator_observations
    ADD CONSTRAINT indicator_observations_source_record_id_fkey
    FOREIGN KEY (source_record_id) REFERENCES timeline_events (record_id) ON DELETE CASCADE;

ALTER TABLE entity_mentions
    ADD CONSTRAINT entity_mentions_source_record_id_fkey
    FOREIGN KEY (source_record_id) REFERENCES timeline_events (record_id) ON DELETE CASCADE;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_src_record_id_fkey
    FOREIGN KEY (src_record_id) REFERENCES timeline_events (record_id) ON DELETE CASCADE;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_dst_record_id_fkey
    FOREIGN KEY (dst_record_id) REFERENCES timeline_events (record_id) ON DELETE CASCADE;

ALTER TABLE record_revisions
    ADD CONSTRAINT record_revisions_record_id_fkey
    FOREIGN KEY (record_id) REFERENCES timeline_events (record_id) ON DELETE CASCADE;

ALTER TABLE record_links
    DROP COLUMN IF EXISTS created_by_user_id;

DROP INDEX IF EXISTS records_active_incident_type_idx;
DROP INDEX IF EXISTS records_incident_lookup_idx;
DROP TABLE IF EXISTS records;

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM change_set_mutations LIMIT 1) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'revisions history association migration requires a pre-production database reset: change_set_mutations is not empty';
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION change_set_mutations_history_ids_are_canonical(candidate uuid[])
RETURNS boolean
LANGUAGE sql
IMMUTABLE
STRICT
PARALLEL SAFE
AS $$
    SELECT
        COALESCE(array_ndims(candidate), 1) = 1
        AND array_position(candidate, NULL::uuid) IS NULL
        AND array_position(candidate, '00000000-0000-0000-0000-000000000000'::uuid) IS NULL
        AND candidate = ARRAY(
            SELECT DISTINCT member
              FROM unnest(candidate) AS values_row(member)
             ORDER BY member
        );
$$;
-- +goose StatementEnd

ALTER TABLE change_set_mutations
    ADD COLUMN history_record_ids uuid[] NOT NULL,
    ADD COLUMN history_entry_record_ids uuid[] NOT NULL,
    ADD CONSTRAINT change_set_mutations_history_record_ids_ck
        CHECK (change_set_mutations_history_ids_are_canonical(history_record_ids)),
    ADD CONSTRAINT change_set_mutations_history_entry_record_ids_ck
        CHECK (change_set_mutations_history_ids_are_canonical(history_entry_record_ids)),
    ADD CONSTRAINT change_set_mutations_history_entry_subset_ck
        CHECK (history_entry_record_ids <@ history_record_ids);

CREATE INDEX change_set_mutations_history_record_ids_idx
    ON change_set_mutations USING gin (history_record_ids);

-- Conflict facts are a narrow concurrency index derived from the explicit
-- live-change input. They are not retained record snapshots and therefore do
-- not duplicate source-owned history or projection rows.
CREATE TABLE record_revision_conflict_facts (
    revision_id bigint NOT NULL
        REFERENCES record_revisions(revision_id) ON DELETE CASCADE,
    field_key text NOT NULL,
    before_present boolean NOT NULL,
    before_value jsonb,
    after_present boolean NOT NULL,
    after_value jsonb,
    PRIMARY KEY (revision_id, field_key),
    CONSTRAINT record_revision_conflict_facts_field_key_ck
        CHECK (field_key = btrim(field_key) AND field_key <> ''),
    CONSTRAINT record_revision_conflict_facts_before_value_ck
        CHECK (before_present = (before_value IS NOT NULL)),
    CONSTRAINT record_revision_conflict_facts_after_value_ck
        CHECK (after_present = (after_value IS NOT NULL)),
    CONSTRAINT record_revision_conflict_facts_presence_ck
        CHECK (before_present OR after_present)
);

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM change_set_mutations LIMIT 1) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'revisions history association downgrade is blocked while change_set_mutations is not empty';
    END IF;
END
$$;
-- +goose StatementEnd

DROP TABLE IF EXISTS record_revision_conflict_facts;

DROP INDEX IF EXISTS change_set_mutations_history_record_ids_idx;

ALTER TABLE change_set_mutations
    DROP CONSTRAINT IF EXISTS change_set_mutations_history_entry_subset_ck,
    DROP CONSTRAINT IF EXISTS change_set_mutations_history_entry_record_ids_ck,
    DROP CONSTRAINT IF EXISTS change_set_mutations_history_record_ids_ck,
    DROP COLUMN IF EXISTS history_entry_record_ids,
    DROP COLUMN IF EXISTS history_record_ids;

DROP FUNCTION IF EXISTS change_set_mutations_history_ids_are_canonical(uuid[]);

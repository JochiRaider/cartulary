-- +goose Up
CREATE TABLE IF NOT EXISTS import_apply_journal (
    import_apply_journal_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    import_session_id uuid NOT NULL REFERENCES import_sessions (import_session_id) ON DELETE CASCADE,
    import_unit_id uuid NOT NULL REFERENCES import_units (import_unit_id) ON DELETE CASCADE,
    mapping_fingerprint text NOT NULL,
    source_row_ref integer NOT NULL CHECK (source_row_ref > 0),
    target_view_schema_id text NOT NULL,
    owner_create_facade text NOT NULL,
    record_id uuid NOT NULL REFERENCES records (record_id) ON DELETE CASCADE,
    row_version bigint NOT NULL CHECK (row_version >= 1),
    change_set_id uuid NOT NULL REFERENCES change_sets (change_set_id) ON DELETE CASCADE,
    change_set_mutation_ref text NOT NULL,
    owner_result_code text NOT NULL,
    created_or_reused text NOT NULL,
    owner_response_json jsonb NOT NULL,
    row_refresh_json jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (import_unit_id, mapping_fingerprint, source_row_ref)
);

CREATE INDEX IF NOT EXISTS import_apply_journal_session_unit_idx
    ON import_apply_journal (import_session_id, import_unit_id, source_row_ref);

CREATE INDEX IF NOT EXISTS import_apply_journal_record_idx
    ON import_apply_journal (record_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS import_apply_journal_record_idx;
DROP INDEX IF EXISTS import_apply_journal_session_unit_idx;
DROP TABLE IF EXISTS import_apply_journal;

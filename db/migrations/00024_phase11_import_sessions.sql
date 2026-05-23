-- +goose Up
CREATE TABLE IF NOT EXISTS import_sessions (
    import_session_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    client_txn_id text NOT NULL,
    assistant_profile text NOT NULL,
    source_file_kind text NOT NULL CHECK (source_file_kind IN ('csv', 'xlsx')),
    original_filename text NOT NULL,
    source_content_sha256 text NOT NULL,
    source_media_type text NOT NULL,
    source_byte_size bigint NOT NULL CHECK (source_byte_size >= 0),
    parser_profile_id text NOT NULL,
    parser_version text NOT NULL,
    session_status text NOT NULL CHECK (session_status IN ('created', 'discovered', 'mapped', 'ready_to_apply', 'applying', 'applied', 'partially_applied', 'failed', 'canceled')),
    discovery_job_id uuid REFERENCES jobs (job_id),
    apply_job_id uuid REFERENCES jobs (job_id),
    selected_unit_ids uuid[] NOT NULL DEFAULT '{}',
    blocking_diagnostics_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    nonblocking_warning_codes text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS import_sessions_incident_lookup_idx
    ON import_sessions (incident_id, created_at DESC, import_session_id);

CREATE INDEX IF NOT EXISTS import_sessions_created_by_lookup_idx
    ON import_sessions (created_by_user_id, created_at DESC, import_session_id);

CREATE TABLE IF NOT EXISTS import_units (
    import_unit_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    import_session_id uuid NOT NULL REFERENCES import_sessions (import_session_id) ON DELETE CASCADE,
    unit_status text NOT NULL CHECK (unit_status IN ('discovered', 'selected', 'mapped', 'ready', 'skipped', 'applying', 'applied', 'rejected', 'failed')),
    locator_kind text NOT NULL,
    locator text NOT NULL,
    source_rect_a1 text NOT NULL,
    header_row_ref integer NOT NULL CHECK (header_row_ref > 0),
    data_start_row_ref integer NOT NULL CHECK (data_start_row_ref > 0),
    inferred_row_count integer NOT NULL CHECK (inferred_row_count >= 0),
    inferred_column_count integer NOT NULL CHECK (inferred_column_count >= 0),
    warning_codes text[] NOT NULL DEFAULT '{}',
    mapping_fingerprint text,
    approved_mapping_json jsonb,
    columns_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    source_rows_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    preview_rows_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS import_units_session_lookup_idx
    ON import_units (import_session_id, created_at ASC, import_unit_id);

-- +goose Down
DROP INDEX IF EXISTS import_units_session_lookup_idx;
DROP TABLE IF EXISTS import_units;

DROP INDEX IF EXISTS import_sessions_created_by_lookup_idx;
DROP INDEX IF EXISTS import_sessions_incident_lookup_idx;
DROP TABLE IF EXISTS import_sessions;

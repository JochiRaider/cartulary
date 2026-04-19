-- +goose Up
CREATE TABLE IF NOT EXISTS indicators (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    indicator_type text NOT NULL,
    value_kind text NOT NULL CHECK (value_kind IN ('atomic', 'pattern', 'reference')),
    display_value text NOT NULL,
    normalized_value text,
    dedupe_key text NOT NULL,
    defanged_value text,
    hash_algorithm text,
    hash_value text,
    stix_pattern text,
    row_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_by_user_id uuid NOT NULL REFERENCES users (id),
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users (id),
    CONSTRAINT indicators_hash_pair_ck CHECK (
        (hash_algorithm IS NULL AND hash_value IS NULL)
        OR (hash_algorithm IS NOT NULL AND hash_value IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS indicators_incident_dedupe_unique_idx
    ON indicators (incident_id, indicator_type, dedupe_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS indicators_incident_normalized_lookup_idx
    ON indicators (incident_id, indicator_type, normalized_value, record_id)
    WHERE deleted_at IS NULL AND normalized_value IS NOT NULL;

CREATE TABLE IF NOT EXISTS indicator_state_intervals (
    indicator_state_interval_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    indicator_record_id uuid NOT NULL REFERENCES indicators (record_id) ON DELETE CASCADE,
    lifecycle_state text NOT NULL,
    valid_from timestamptz NOT NULL,
    valid_to timestamptz,
    confidence integer CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 100)),
    rationale text,
    support_refs jsonb NOT NULL DEFAULT '[]'::jsonb,
    assessor text,
    assessed_at timestamptz NOT NULL DEFAULT now(),
    row_version bigint NOT NULL DEFAULT 1,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT indicator_state_intervals_validity_ck CHECK (valid_to IS NULL OR valid_to >= valid_from)
);

CREATE INDEX IF NOT EXISTS indicator_state_intervals_indicator_lookup_idx
    ON indicator_state_intervals (incident_id, indicator_record_id, valid_from DESC, indicator_state_interval_id DESC);

CREATE TABLE IF NOT EXISTS indicator_observations (
    indicator_observation_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    source_record_id uuid NOT NULL REFERENCES timeline_events (record_id) ON DELETE CASCADE,
    source_field_key text NOT NULL,
    origin_kind text NOT NULL,
    origin_locator text NOT NULL,
    observed_text text NOT NULL,
    parsed_indicator_type text,
    normalized_candidate text,
    resolution_status text NOT NULL CHECK (resolution_status IN ('unresolved', 'resolved', 'dismissed')),
    resolved_indicator_record_id uuid REFERENCES indicators (record_id) ON DELETE SET NULL,
    row_version bigint NOT NULL DEFAULT 1,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    resolved_by_user_id uuid REFERENCES users (id),
    resolved_at timestamptz,
    resolution_method text
);

CREATE INDEX IF NOT EXISTS indicator_observations_source_lookup_idx
    ON indicator_observations (source_record_id, source_field_key, created_at ASC, indicator_observation_id ASC);

CREATE INDEX IF NOT EXISTS indicator_observations_resolved_lookup_idx
    ON indicator_observations (incident_id, resolution_status, resolved_indicator_record_id, created_at ASC);

CREATE INDEX IF NOT EXISTS indicator_observations_candidate_lookup_idx
    ON indicator_observations (incident_id, parsed_indicator_type, normalized_candidate, indicator_observation_id)
    WHERE normalized_candidate IS NOT NULL;

CREATE TABLE IF NOT EXISTS indicator_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES indicators (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    indicator_type text NOT NULL,
    value_kind text NOT NULL,
    display_value text NOT NULL,
    normalized_value text,
    dedupe_key text NOT NULL,
    defanged_value text,
    hash_algorithm text,
    hash_value text,
    stix_pattern text,
    first_observed_at timestamptz,
    last_observed_at timestamptz,
    observation_count integer NOT NULL DEFAULT 0,
    lifecycle_summary text,
    supporting_link_count integer NOT NULL DEFAULT 0,
    edited_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS indicator_grid_projection_incident_sort_idx
    ON indicator_grid_projection (incident_id, indicator_type ASC, display_value ASC, record_id ASC);

CREATE INDEX IF NOT EXISTS indicator_grid_projection_incident_lifecycle_idx
    ON indicator_grid_projection (incident_id, lifecycle_summary, record_id ASC);

-- +goose Down
DROP INDEX IF EXISTS indicator_grid_projection_incident_lifecycle_idx;
DROP INDEX IF EXISTS indicator_grid_projection_incident_sort_idx;
DROP TABLE IF EXISTS indicator_grid_projection;

DROP INDEX IF EXISTS indicator_observations_candidate_lookup_idx;
DROP INDEX IF EXISTS indicator_observations_resolved_lookup_idx;
DROP INDEX IF EXISTS indicator_observations_source_lookup_idx;
DROP TABLE IF EXISTS indicator_observations;

DROP INDEX IF EXISTS indicator_state_intervals_indicator_lookup_idx;
DROP TABLE IF EXISTS indicator_state_intervals;

DROP INDEX IF EXISTS indicators_incident_normalized_lookup_idx;
DROP INDEX IF EXISTS indicators_incident_dedupe_unique_idx;
DROP TABLE IF EXISTS indicators;

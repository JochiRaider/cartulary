-- +goose Up
CREATE TABLE IF NOT EXISTS hosts (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    display_name text NOT NULL,
    hostname text,
    host_state text NOT NULL CHECK (host_state IN ('stub', 'canonical', 'merged')),
    row_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_by_user_id uuid NOT NULL REFERENCES users (id)
);

CREATE INDEX IF NOT EXISTS hosts_incident_display_name_idx
    ON hosts (incident_id, display_name, record_id);

CREATE INDEX IF NOT EXISTS hosts_incident_hostname_idx
    ON hosts (incident_id, hostname, record_id)
    WHERE hostname IS NOT NULL;

CREATE TABLE IF NOT EXISTS host_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES hosts (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    display_name text NOT NULL,
    hostname text,
    host_state text NOT NULL CHECK (host_state IN ('stub', 'canonical', 'merged')),
    linked_event_count integer NOT NULL DEFAULT 0,
    evidence_count integer NOT NULL DEFAULT 0,
    location text,
    os_platform text,
    business_owner text,
    criticality text,
    containment_status text,
    edited_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS host_grid_projection_incident_sort_idx
    ON host_grid_projection (incident_id, display_name ASC, record_id ASC);

CREATE TABLE IF NOT EXISTS identities (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    display_name text NOT NULL,
    upn text,
    email citext,
    sam_account_name text,
    identity_state text NOT NULL CHECK (identity_state IN ('stub', 'canonical', 'merged')),
    row_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_by_user_id uuid NOT NULL REFERENCES users (id)
);

CREATE INDEX IF NOT EXISTS identities_incident_display_name_idx
    ON identities (incident_id, display_name, record_id);

CREATE INDEX IF NOT EXISTS identities_incident_upn_idx
    ON identities (incident_id, upn, record_id)
    WHERE upn IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_email_idx
    ON identities (incident_id, email, record_id)
    WHERE email IS NOT NULL;

CREATE TABLE IF NOT EXISTS identity_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES identities (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    display_name text NOT NULL,
    upn text,
    email citext,
    sam_account_name text,
    identity_state text NOT NULL CHECK (identity_state IN ('stub', 'canonical', 'merged')),
    linked_event_count integer NOT NULL DEFAULT 0,
    evidence_count integer NOT NULL DEFAULT 0,
    privilege_level text,
    mfa_state text,
    reset_status text,
    edited_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS identity_grid_projection_incident_sort_idx
    ON identity_grid_projection (incident_id, display_name ASC, record_id ASC);

CREATE TABLE IF NOT EXISTS entity_mentions (
    entity_mention_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_record_id uuid NOT NULL REFERENCES timeline_events (record_id) ON DELETE CASCADE,
    entity_type text NOT NULL CHECK (entity_type IN ('host', 'identity')),
    source_field_key text NOT NULL,
    origin_kind text NOT NULL,
    origin_locator text NOT NULL,
    raw_text text NOT NULL,
    normalized_text text NOT NULL,
    resolution_status text NOT NULL CHECK (resolution_status IN ('unresolved', 'resolved', 'dismissed')),
    row_version bigint NOT NULL DEFAULT 1,
    ordinal integer NOT NULL CHECK (ordinal > 0),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    resolved_record_id uuid,
    resolved_by_user_id uuid REFERENCES users (id),
    resolved_at timestamptz,
    resolution_method text
);

CREATE INDEX IF NOT EXISTS entity_mentions_source_lookup_idx
    ON entity_mentions (source_record_id, source_field_key, ordinal ASC, entity_mention_id ASC);

CREATE INDEX IF NOT EXISTS entity_mentions_unresolved_lookup_idx
    ON entity_mentions (source_record_id, resolution_status, entity_type);

-- +goose Down
DROP INDEX IF EXISTS entity_mentions_unresolved_lookup_idx;
DROP INDEX IF EXISTS entity_mentions_source_lookup_idx;
DROP TABLE IF EXISTS entity_mentions;

DROP INDEX IF EXISTS identity_grid_projection_incident_sort_idx;
DROP TABLE IF EXISTS identity_grid_projection;

DROP INDEX IF EXISTS identities_incident_email_idx;
DROP INDEX IF EXISTS identities_incident_upn_idx;
DROP INDEX IF EXISTS identities_incident_display_name_idx;
DROP TABLE IF EXISTS identities;

DROP INDEX IF EXISTS host_grid_projection_incident_sort_idx;
DROP TABLE IF EXISTS host_grid_projection;

DROP INDEX IF EXISTS hosts_incident_hostname_idx;
DROP INDEX IF EXISTS hosts_incident_display_name_idx;
DROP TABLE IF EXISTS hosts;

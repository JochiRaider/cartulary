-- +goose Up
ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check,
    DROP CONSTRAINT IF EXISTS record_links_dst_record_id_fkey;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity'));

CREATE INDEX IF NOT EXISTS record_links_active_src_lookup_idx
    ON record_links (incident_id, src_record_id, link_type, dst_record_id)
    WHERE deleted_at IS NULL;

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

CREATE TABLE IF NOT EXISTS hosts (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    display_name text NOT NULL,
    hostname text,
    aad_device_id text,
    fqdn text,
    entity_origin text NOT NULL DEFAULT 'entity_sheet',
    seed_entity_mention_id uuid REFERENCES entity_mentions (entity_mention_id),
    host_state text NOT NULL CHECK (host_state IN ('stub', 'canonical', 'merged')),
    merged_into_record_id uuid REFERENCES hosts (record_id),
    row_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_by_user_id uuid NOT NULL REFERENCES users (id),
    CONSTRAINT hosts_entity_origin_core02_ck CHECK (entity_origin IN ('entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert')),
    CONSTRAINT hosts_merge_lineage_ck CHECK (
        (host_state IN ('stub', 'canonical') AND merged_into_record_id IS NULL)
        OR (host_state = 'merged' AND merged_into_record_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS hosts_incident_display_name_idx
    ON hosts (incident_id, display_name, record_id);

CREATE INDEX IF NOT EXISTS hosts_incident_hostname_idx
    ON hosts (incident_id, hostname, record_id)
    WHERE hostname IS NOT NULL;

CREATE INDEX IF NOT EXISTS hosts_incident_aad_device_id_idx
    ON hosts (incident_id, aad_device_id, record_id)
    WHERE aad_device_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS hosts_incident_fqdn_idx
    ON hosts (incident_id, fqdn, record_id)
    WHERE fqdn IS NOT NULL;

CREATE INDEX IF NOT EXISTS hosts_incident_merged_into_idx
    ON hosts (incident_id, merged_into_record_id, record_id)
    WHERE merged_into_record_id IS NOT NULL;

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
    aad_object_id text,
    sid text,
    entity_origin text NOT NULL DEFAULT 'entity_sheet',
    seed_entity_mention_id uuid REFERENCES entity_mentions (entity_mention_id),
    identity_state text NOT NULL CHECK (identity_state IN ('stub', 'canonical', 'merged')),
    merged_into_record_id uuid REFERENCES identities (record_id),
    row_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_by_user_id uuid NOT NULL REFERENCES users (id),
    CONSTRAINT identities_entity_origin_core02_ck CHECK (entity_origin IN ('entity_sheet', 'entity_import', 'created_from_mention', 'system_upsert')),
    CONSTRAINT identities_merge_lineage_ck CHECK (
        (identity_state IN ('stub', 'canonical') AND merged_into_record_id IS NULL)
        OR (identity_state = 'merged' AND merged_into_record_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS identities_incident_display_name_idx
    ON identities (incident_id, display_name, record_id);

CREATE INDEX IF NOT EXISTS identities_incident_upn_idx
    ON identities (incident_id, upn, record_id)
    WHERE upn IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_email_idx
    ON identities (incident_id, email, record_id)
    WHERE email IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_aad_object_id_idx
    ON identities (incident_id, aad_object_id, record_id)
    WHERE aad_object_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_sid_idx
    ON identities (incident_id, sid, record_id)
    WHERE sid IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_sam_account_name_idx
    ON identities (incident_id, sam_account_name, record_id)
    WHERE sam_account_name IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_merged_into_idx
    ON identities (incident_id, merged_into_record_id, record_id)
    WHERE merged_into_record_id IS NOT NULL;

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

CREATE TABLE IF NOT EXISTS entity_preserved_identifiers (
    entity_preserved_identifier_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    record_id uuid NOT NULL,
    entity_type text NOT NULL CHECK (entity_type IN ('host', 'identity')),
    identifier_type text NOT NULL,
    raw_value text NOT NULL,
    normalized_value text NOT NULL,
    classification text NOT NULL CHECK (classification IN ('exact_match_reuse', 'suggestion_only', 'provenance_only')),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS entity_preserved_identifiers_exact_lookup_idx
    ON entity_preserved_identifiers (incident_id, entity_type, identifier_type, normalized_value, record_id)
    WHERE deleted_at IS NULL AND classification = 'exact_match_reuse';

CREATE UNIQUE INDEX IF NOT EXISTS entity_preserved_identifiers_record_unique_idx
    ON entity_preserved_identifiers (record_id, entity_type, identifier_type, normalized_value, classification)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS entity_aliases (
    entity_alias_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    record_id uuid NOT NULL,
    entity_type text NOT NULL CHECK (entity_type IN ('host', 'identity')),
    raw_text text NOT NULL,
    normalized_text text NOT NULL,
    classification text NOT NULL CHECK (classification = 'suggestion_only'),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS entity_aliases_lookup_idx
    ON entity_aliases (incident_id, entity_type, normalized_text, record_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS entity_aliases_record_unique_idx
    ON entity_aliases (record_id, entity_type, normalized_text)
    WHERE deleted_at IS NULL;

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

DROP INDEX IF EXISTS compromise_assessments_active_subject_lookup_idx;
DROP TABLE IF EXISTS compromise_assessments;

DROP INDEX IF EXISTS record_tags_active_record_lookup_idx;
DROP INDEX IF EXISTS record_tags_active_unique_idx;
DROP TABLE IF EXISTS record_tags;

DROP INDEX IF EXISTS entity_aliases_record_unique_idx;
DROP INDEX IF EXISTS entity_aliases_lookup_idx;
DROP TABLE IF EXISTS entity_aliases;

DROP INDEX IF EXISTS entity_preserved_identifiers_record_unique_idx;
DROP INDEX IF EXISTS entity_preserved_identifiers_exact_lookup_idx;
DROP TABLE IF EXISTS entity_preserved_identifiers;

DROP INDEX IF EXISTS identity_grid_projection_incident_sort_idx;
DROP TABLE IF EXISTS identity_grid_projection;

DROP INDEX IF EXISTS identities_incident_merged_into_idx;
DROP INDEX IF EXISTS identities_incident_sam_account_name_idx;
DROP INDEX IF EXISTS identities_incident_sid_idx;
DROP INDEX IF EXISTS identities_incident_aad_object_id_idx;
DROP INDEX IF EXISTS identities_incident_email_idx;
DROP INDEX IF EXISTS identities_incident_upn_idx;
DROP INDEX IF EXISTS identities_incident_display_name_idx;
DROP TABLE IF EXISTS identities;

DROP INDEX IF EXISTS host_grid_projection_incident_sort_idx;
DROP TABLE IF EXISTS host_grid_projection;

DROP INDEX IF EXISTS hosts_incident_merged_into_idx;
DROP INDEX IF EXISTS hosts_incident_fqdn_idx;
DROP INDEX IF EXISTS hosts_incident_aad_device_id_idx;
DROP INDEX IF EXISTS hosts_incident_hostname_idx;
DROP INDEX IF EXISTS hosts_incident_display_name_idx;
DROP TABLE IF EXISTS hosts;

DROP INDEX IF EXISTS entity_mentions_unresolved_lookup_idx;
DROP INDEX IF EXISTS entity_mentions_source_lookup_idx;
DROP TABLE IF EXISTS entity_mentions;

DROP INDEX IF EXISTS record_links_active_src_lookup_idx;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes'));

ALTER TABLE record_links
    ADD CONSTRAINT record_links_dst_record_id_fkey FOREIGN KEY (dst_record_id) REFERENCES timeline_events (record_id) ON DELETE CASCADE;

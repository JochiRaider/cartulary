-- +goose Up
ALTER TABLE hosts
    ADD COLUMN IF NOT EXISTS aad_device_id text,
    ADD COLUMN IF NOT EXISTS fqdn text,
    ADD COLUMN IF NOT EXISTS entity_origin text NOT NULL DEFAULT 'direct_create' CHECK (entity_origin IN ('direct_create', 'created_from_mention')),
    ADD COLUMN IF NOT EXISTS seed_entity_mention_id uuid REFERENCES entity_mentions (entity_mention_id);

CREATE INDEX IF NOT EXISTS hosts_incident_aad_device_id_idx
    ON hosts (incident_id, aad_device_id, record_id)
    WHERE aad_device_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS hosts_incident_fqdn_idx
    ON hosts (incident_id, fqdn, record_id)
    WHERE fqdn IS NOT NULL;

ALTER TABLE identities
    ADD COLUMN IF NOT EXISTS aad_object_id text,
    ADD COLUMN IF NOT EXISTS sid text,
    ADD COLUMN IF NOT EXISTS entity_origin text NOT NULL DEFAULT 'direct_create' CHECK (entity_origin IN ('direct_create', 'created_from_mention')),
    ADD COLUMN IF NOT EXISTS seed_entity_mention_id uuid REFERENCES entity_mentions (entity_mention_id);

CREATE INDEX IF NOT EXISTS identities_incident_aad_object_id_idx
    ON identities (incident_id, aad_object_id, record_id)
    WHERE aad_object_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_sid_idx
    ON identities (incident_id, sid, record_id)
    WHERE sid IS NOT NULL;

CREATE INDEX IF NOT EXISTS identities_incident_sam_account_name_idx
    ON identities (incident_id, sam_account_name, record_id)
    WHERE sam_account_name IS NOT NULL;

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

-- +goose Down
DROP INDEX IF EXISTS entity_aliases_record_unique_idx;
DROP INDEX IF EXISTS entity_aliases_lookup_idx;
DROP TABLE IF EXISTS entity_aliases;

DROP INDEX IF EXISTS entity_preserved_identifiers_record_unique_idx;
DROP INDEX IF EXISTS entity_preserved_identifiers_exact_lookup_idx;
DROP TABLE IF EXISTS entity_preserved_identifiers;

DROP INDEX IF EXISTS identities_incident_sam_account_name_idx;
DROP INDEX IF EXISTS identities_incident_sid_idx;
DROP INDEX IF EXISTS identities_incident_aad_object_id_idx;
ALTER TABLE identities
    DROP COLUMN IF EXISTS seed_entity_mention_id,
    DROP COLUMN IF EXISTS entity_origin,
    DROP COLUMN IF EXISTS sid,
    DROP COLUMN IF EXISTS aad_object_id;

DROP INDEX IF EXISTS hosts_incident_fqdn_idx;
DROP INDEX IF EXISTS hosts_incident_aad_device_id_idx;
ALTER TABLE hosts
    DROP COLUMN IF EXISTS seed_entity_mention_id,
    DROP COLUMN IF EXISTS entity_origin,
    DROP COLUMN IF EXISTS fqdn,
    DROP COLUMN IF EXISTS aad_device_id;

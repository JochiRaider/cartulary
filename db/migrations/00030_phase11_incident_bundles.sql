-- +goose Up
CREATE TABLE incident_bundle_exports (
    bundle_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    export_job_id uuid NOT NULL UNIQUE REFERENCES jobs (job_id) ON DELETE CASCADE,
    exported_by_user_id uuid NOT NULL REFERENCES users (id),
    exported_at timestamptz NOT NULL,
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    reference_pack_mode text NOT NULL CHECK (reference_pack_mode IN ('refs_only', 'embedded')),
    history_mode text NOT NULL DEFAULT 'full' CHECK (history_mode = 'full'),
    blob_mode text NOT NULL DEFAULT 'full' CHECK (blob_mode = 'full'),
    optional_sections text[] NOT NULL DEFAULT '{}',
    required_capabilities text[] NOT NULL DEFAULT '{}',
    bundle_sha256 text NOT NULL CHECK (bundle_sha256 ~ '^[0-9a-f]{64}$'),
    bundle_byte_size bigint NOT NULL CHECK (bundle_byte_size >= 0),
    bundle_storage_path text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX incident_bundle_exports_incident_lookup_idx
    ON incident_bundle_exports (incident_id, exported_at DESC, bundle_id);

CREATE TABLE incident_bundle_manifest_files (
    bundle_id uuid NOT NULL REFERENCES incident_bundle_exports (bundle_id) ON DELETE CASCADE,
    path text NOT NULL,
    sha256 text NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
    required boolean NOT NULL DEFAULT true,
    PRIMARY KEY (bundle_id, path)
);

CREATE TABLE incident_bundle_job_payloads (
    job_id uuid PRIMARY KEY REFERENCES jobs (job_id) ON DELETE CASCADE,
    job_kind text NOT NULL CHECK (job_kind IN ('export', 'import')),
    actor_user_id uuid NOT NULL REFERENCES users (id),
    incident_id uuid REFERENCES incidents (id) ON DELETE CASCADE,
    bundle_id uuid,
    uploaded_sha256 text CHECK (uploaded_sha256 IS NULL OR uploaded_sha256 ~ '^[0-9a-f]{64}$'),
    bundle_staging_path text,
    imported_incident_id uuid,
    manifest_sha256 text CHECK (manifest_sha256 IS NULL OR manifest_sha256 ~ '^[0-9a-f]{64}$'),
    failure_reason text,
    request_json jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT incident_bundle_payload_kind_ck CHECK (
        (job_kind = 'export' AND incident_id IS NOT NULL AND uploaded_sha256 IS NULL)
        OR
        (job_kind = 'import' AND incident_id IS NULL AND uploaded_sha256 IS NOT NULL)
    )
);

CREATE INDEX incident_bundle_job_payloads_actor_idx
    ON incident_bundle_job_payloads (actor_user_id, created_at DESC);

CREATE TABLE incident_bundle_imported_actors (
    imported_actor_descriptor_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    source_actor_id text NOT NULL,
    display_name text,
    email_hint text,
    local_user_id uuid REFERENCES users (id),
    imported_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (incident_id, source_actor_id)
);

-- +goose Down
DROP TABLE IF EXISTS incident_bundle_imported_actors;
DROP INDEX IF EXISTS incident_bundle_job_payloads_actor_idx;
DROP TABLE IF EXISTS incident_bundle_job_payloads;
DROP TABLE IF EXISTS incident_bundle_manifest_files;
DROP INDEX IF EXISTS incident_bundle_exports_incident_lookup_idx;
DROP TABLE IF EXISTS incident_bundle_exports;

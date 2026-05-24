-- +goose Up
CREATE TABLE reference_packs (
    pack_key text NOT NULL,
    version text NOT NULL,
    pack_kind text NOT NULL,
    source_identifier text,
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    pack_contract_version text NOT NULL,
    verification_method text NOT NULL,
    signer_key_id text,
    status text NOT NULL DEFAULT 'staged' CHECK (
        status IN ('staged','available','disabled','failed','missing')
    ),
    imported_at timestamptz NOT NULL,
    imported_by_user_id uuid REFERENCES users(id),
    activated_at timestamptz,
    activated_by_user_id uuid REFERENCES users(id),
    previous_active_version text,
    verification_result text NOT NULL DEFAULT 'pending' CHECK (
        verification_result IN ('pending','passed','failed')
    ),
    bundle_sha256 text NOT NULL,
    bundle_storage_path text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (pack_key, version),
    FOREIGN KEY (pack_key, previous_active_version)
        REFERENCES reference_packs(pack_key, version)
);

CREATE TABLE reference_pack_activation_state (
    pack_key text PRIMARY KEY,
    active_version text,
    previous_active_version text,
    activated_at timestamptz,
    activated_by_user_id uuid REFERENCES users(id),
    operator_note text,
    CHECK (
        active_version IS NULL
        OR previous_active_version IS NULL
        OR previous_active_version <> active_version
    ),
    FOREIGN KEY (pack_key, active_version)
        REFERENCES reference_packs(pack_key, version),
    FOREIGN KEY (pack_key, previous_active_version)
        REFERENCES reference_packs(pack_key, version)
);

CREATE TABLE reference_pack_attestations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    pack_kind text NOT NULL,
    event_kind text NOT NULL CHECK (
        event_kind IN ('import','activate','disable','reverify','refresh')
    ),
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    source_identifier text,
    verification_method text NOT NULL,
    signer_key_id text,
    previous_active_version text,
    verification_result text NOT NULL CHECK (
        verification_result IN ('pending','passed','failed')
    ),
    actor_user_id uuid REFERENCES users(id),
    job_id uuid REFERENCES jobs(job_id) ON DELETE SET NULL,
    occurred_at timestamptz NOT NULL,
    operator_note text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    FOREIGN KEY (pack_key, pack_version)
        REFERENCES reference_packs(pack_key, version),
    FOREIGN KEY (pack_key, previous_active_version)
        REFERENCES reference_packs(pack_key, version)
);

CREATE TABLE reference_pack_job_payloads (
    job_id uuid PRIMARY KEY REFERENCES jobs(job_id) ON DELETE CASCADE,
    job_kind text NOT NULL CHECK (job_kind IN ('import','reverify','refresh')),
    actor_user_id uuid NOT NULL REFERENCES users(id),
    pack_key text,
    pack_version text,
    resolved_pack_keys text[] NOT NULL DEFAULT '{}'::text[],
    bundle_sha256 text,
    request_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL
);

CREATE INDEX reference_packs_status_idx
    ON reference_packs (status, verification_result);

CREATE INDEX reference_pack_attestations_pack_idx
    ON reference_pack_attestations (pack_key, pack_version, occurred_at);

CREATE INDEX reference_pack_job_payloads_actor_idx
    ON reference_pack_job_payloads (actor_user_id, created_at);

-- +goose Down
DROP INDEX IF EXISTS reference_pack_job_payloads_actor_idx;
DROP INDEX IF EXISTS reference_pack_attestations_pack_idx;
DROP INDEX IF EXISTS reference_packs_status_idx;
DROP TABLE IF EXISTS reference_pack_job_payloads;
DROP TABLE IF EXISTS reference_pack_attestations;
DROP TABLE IF EXISTS reference_pack_activation_state;
DROP TABLE IF EXISTS reference_packs;

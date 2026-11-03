-- +goose Up
-- Installation-local admission authority. This row is never backup content.
CREATE TABLE public.recovery_target_state (
    singleton boolean NOT NULL DEFAULT true,
    schema_id text NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    target_generation_id uuid,
    purpose text,
    operation text,
    operation_id uuid,
    attempt_id text,
    backup_set_id uuid,
    recovery_state_catalog_sha256 text,
    codec_registry_sha256 text,
    database_sha256 text,
    object_store_sha256 text,
    reference_pack_storage_sha256 text,
    export_outputs_sha256 text,
    completion_evidence_sha256 text,
    CONSTRAINT recovery_target_state_pkey PRIMARY KEY (singleton),
    CONSTRAINT recovery_target_state_singleton_check CHECK (singleton),
    CONSTRAINT recovery_target_state_schema_check CHECK (schema_id = 'cartulary.recovery_target_state.v1'),
    CONSTRAINT recovery_target_state_revision_check CHECK (revision > 0),
    CONSTRAINT recovery_target_state_state_check CHECK (
        state IN ('unproven', 'initialized', 'verification_ready', 'restoring', 'complete', 'resetting')
    ),
    CONSTRAINT recovery_target_state_purpose_check CHECK (
        purpose IS NULL OR purpose IN ('restore_target', 'restore_verification_target')
    ),
    CONSTRAINT recovery_target_state_operation_check CHECK (
        operation IS NULL OR operation IN ('restore_latest', 'restore_bundle', 'restore_verify_latest', 'restore_verify_due')
    ),
    CONSTRAINT recovery_target_state_uuid_check CHECK (
        (target_generation_id IS NULL OR target_generation_id <> '00000000-0000-0000-0000-000000000000'::uuid)
        AND (operation_id IS NULL OR operation_id <> '00000000-0000-0000-0000-000000000000'::uuid)
        AND (backup_set_id IS NULL OR backup_set_id <> '00000000-0000-0000-0000-000000000000'::uuid)
    ),
    CONSTRAINT recovery_target_state_hash_check CHECK (
        (recovery_state_catalog_sha256 IS NULL OR recovery_state_catalog_sha256 ~ '^[0-9a-f]{64}$')
        AND (codec_registry_sha256 IS NULL OR codec_registry_sha256 ~ '^[0-9a-f]{64}$')
        AND (database_sha256 IS NULL OR database_sha256 ~ '^[0-9a-f]{64}$')
        AND (object_store_sha256 IS NULL OR object_store_sha256 ~ '^[0-9a-f]{64}$')
        AND (reference_pack_storage_sha256 IS NULL OR reference_pack_storage_sha256 ~ '^[0-9a-f]{64}$')
        AND (export_outputs_sha256 IS NULL OR export_outputs_sha256 ~ '^[0-9a-f]{64}$')
        AND (completion_evidence_sha256 IS NULL OR completion_evidence_sha256 ~ '^[0-9a-f]{64}$')
    ),
    CONSTRAINT recovery_target_state_attempt_check CHECK (
        attempt_id IS NULL OR (length(attempt_id) <= 160 AND attempt_id ~ '^[A-Za-z0-9][A-Za-z0-9_.:-]*$')
    ),
    CONSTRAINT recovery_target_state_phase_check CHECK (
        (
            state IN ('unproven', 'initialized')
            AND target_generation_id IS NULL AND purpose IS NULL AND operation IS NULL
            AND operation_id IS NULL AND attempt_id IS NULL AND backup_set_id IS NULL
            AND recovery_state_catalog_sha256 IS NULL AND codec_registry_sha256 IS NULL
            AND database_sha256 IS NULL AND object_store_sha256 IS NULL
            AND reference_pack_storage_sha256 IS NULL AND export_outputs_sha256 IS NULL
            AND completion_evidence_sha256 IS NULL
        ) OR (
            state IN ('verification_ready', 'resetting')
            AND target_generation_id IS NOT NULL AND purpose IS NOT NULL
            AND purpose = 'restore_verification_target'
            AND database_sha256 IS NOT NULL AND object_store_sha256 IS NOT NULL
            AND reference_pack_storage_sha256 IS NOT NULL AND export_outputs_sha256 IS NOT NULL
            AND operation IS NULL AND operation_id IS NULL AND attempt_id IS NULL AND backup_set_id IS NULL
            AND recovery_state_catalog_sha256 IS NULL AND codec_registry_sha256 IS NULL
            AND completion_evidence_sha256 IS NULL
        ) OR (
            state IN ('restoring', 'complete')
            AND target_generation_id IS NOT NULL AND purpose IS NOT NULL
            AND operation IS NOT NULL AND operation_id IS NOT NULL AND backup_set_id IS NOT NULL
            AND recovery_state_catalog_sha256 IS NOT NULL AND codec_registry_sha256 IS NOT NULL
            AND database_sha256 IS NOT NULL AND object_store_sha256 IS NOT NULL
            AND reference_pack_storage_sha256 IS NOT NULL AND export_outputs_sha256 IS NOT NULL
            AND (
                (purpose = 'restore_target' AND operation IN ('restore_latest', 'restore_bundle') AND attempt_id IS NULL)
                OR (purpose = 'restore_verification_target' AND operation = 'restore_verify_latest' AND attempt_id IS NULL)
                OR (purpose = 'restore_verification_target' AND operation = 'restore_verify_due' AND attempt_id IS NOT NULL)
            )
            AND (
                (state = 'restoring' AND completion_evidence_sha256 IS NULL)
                OR (state = 'complete' AND completion_evidence_sha256 IS NOT NULL)
            )
        )
    )
);

REVOKE ALL ON TABLE public.recovery_target_state FROM PUBLIC;
REVOKE ALL ON TABLE public.recovery_target_state FROM cartulary_runtime, cartulary_recovery;
REVOKE ALL ON TYPE public.recovery_target_state FROM PUBLIC;
GRANT SELECT ON TABLE public.recovery_target_state TO cartulary_runtime;
GRANT SELECT, UPDATE ON TABLE public.recovery_target_state TO cartulary_recovery;

-- The initializer admits freshness before DDL and publishes crypto identity
-- only after the entire new schema and external dependencies are validated.
-- Existing initialized state is preserved, never inferred to be completed.
INSERT INTO public.recovery_target_state (schema_id, state, revision)
SELECT 'cartulary.recovery_target_state.v1',
       CASE WHEN EXISTS (SELECT 1 FROM public.application_crypto_format)
            THEN 'unproven' ELSE 'initialized' END,
       1;

-- +goose Down
DROP TABLE public.recovery_target_state;

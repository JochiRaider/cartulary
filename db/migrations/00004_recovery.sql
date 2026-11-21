-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: recovery.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.operator_recovery_journal (
    operator_recovery_journal_id uuid DEFAULT gen_random_uuid() NOT NULL,
    operation_id uuid NOT NULL,
    operation text NOT NULL,
    result text NOT NULL,
    backup_set_id uuid,
    error_code text,
    reason_code text,
    envelope_schema_id text NOT NULL,
    encryption_mode text NOT NULL,
    key_fingerprint_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    sealed_payload bytea NOT NULL,
    CONSTRAINT operator_recovery_journal_encryption_mode_check CHECK (encryption_mode = 'aes-256-gcm-envelope'::text),
    CONSTRAINT operator_recovery_journal_envelope_schema_id_check CHECK (envelope_schema_id = 'cartulary.operator_recovery_journal_envelope.v2'::text),
    CONSTRAINT operator_recovery_journal_key_fingerprint_sha256_check CHECK (key_fingerprint_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT operator_recovery_journal_operation_check CHECK (operation = ANY (ARRAY['backup_create'::text, 'backup_export_latest'::text, 'restore_latest'::text, 'restore_bundle'::text, 'restore_verify_latest'::text, 'restore_verify_due'::text])),
    CONSTRAINT operator_recovery_journal_payload_sha256_check CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT operator_recovery_journal_pkey PRIMARY KEY (operator_recovery_journal_id),
    CONSTRAINT operator_recovery_journal_result_check CHECK (result = ANY (ARRAY['started'::text, 'succeeded'::text, 'failed'::text, 'no_op'::text])),
    CONSTRAINT operator_recovery_journal_sealed_payload_check CHECK (octet_length(sealed_payload) >= 62 AND octet_length(sealed_payload) <= 16777277 AND get_byte(sealed_payload, 0) = 1)
);

REVOKE ALL ON TABLE public.operator_recovery_journal FROM PUBLIC;

REVOKE ALL ON TYPE public.operator_recovery_journal FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.operator_recovery_journal TO cartulary_recovery;

CREATE INDEX operator_recovery_journal_backup_set_idx ON public.operator_recovery_journal USING btree (backup_set_id, created_at DESC) WHERE (backup_set_id IS NOT NULL);

CREATE INDEX operator_recovery_journal_operation_created_idx ON public.operator_recovery_journal USING btree (operation, created_at DESC, operation_id);

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

INSERT INTO public.recovery_target_state (schema_id, state, revision) VALUES ('cartulary.recovery_target_state.v1', 'initialized', 1);

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TABLE public.recovery_target_state;
DROP TABLE public.operator_recovery_journal;

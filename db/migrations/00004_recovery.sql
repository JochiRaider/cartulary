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

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TABLE public.operator_recovery_journal;

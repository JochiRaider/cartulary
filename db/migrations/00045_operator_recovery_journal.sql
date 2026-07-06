-- +goose Up
CREATE TABLE IF NOT EXISTS operator_recovery_journal (
    operator_recovery_journal_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operation_id uuid NOT NULL,
    operation text NOT NULL CHECK (
        operation IN (
            'backup_create',
            'restore_latest',
            'restore_verify_latest',
            'restore_verify_due'
        )
    ),
    result text NOT NULL CHECK (result IN ('started', 'succeeded', 'failed', 'no_op')),
    backup_set_id uuid,
    error_code text,
    reason_code text,
    envelope_schema_id text NOT NULL CHECK (envelope_schema_id = 'cartulary.operator_recovery_journal_envelope.v1'),
    encryption_mode text NOT NULL CHECK (encryption_mode = 'aes-256-gcm-envelope'),
    key_fingerprint_sha256 text NOT NULL CHECK (key_fingerprint_sha256 ~ '^[0-9a-f]{64}$'),
    payload_sha256 text NOT NULL CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    nonce bytea NOT NULL,
    ciphertext bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS operator_recovery_journal_operation_created_idx
    ON operator_recovery_journal (operation, created_at DESC, operation_id);

CREATE INDEX IF NOT EXISTS operator_recovery_journal_backup_set_idx
    ON operator_recovery_journal (backup_set_id, created_at DESC)
    WHERE backup_set_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS operator_recovery_journal_backup_set_idx;
DROP INDEX IF EXISTS operator_recovery_journal_operation_created_idx;
DROP TABLE IF EXISTS operator_recovery_journal;

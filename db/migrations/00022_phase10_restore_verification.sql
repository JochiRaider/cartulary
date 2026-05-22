-- +goose Up
ALTER TABLE backup_sets
    ADD COLUMN IF NOT EXISTS last_verification_basis_sha256 text;

ALTER TABLE backup_sets
    ADD CONSTRAINT backup_sets_last_verification_basis_sha256_check
        CHECK (last_verification_basis_sha256 IS NULL OR last_verification_basis_sha256 ~ '^[0-9a-f]{64}$');

CREATE TABLE IF NOT EXISTS restore_verification_runs (
    restore_verification_run_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_set_id uuid NOT NULL REFERENCES backup_sets (backup_set_id) ON DELETE CASCADE,
    started_at timestamptz NOT NULL,
    completed_at timestamptz NOT NULL,
    verification_state text NOT NULL,
    verification_basis_sha256 text NOT NULL,
    failure_reason text,
    failure_message text,
    authoritative_rows_sha256 text,
    authoritative_row_count integer,
    change_sets_sha256 text,
    change_set_row_count integer,
    blob_hashes_sha256 text,
    blob_count integer,
    CONSTRAINT restore_verification_runs_state_check
        CHECK (verification_state IN ('verified', 'failed')),
    CONSTRAINT restore_verification_runs_basis_sha256_check
        CHECK (verification_basis_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT restore_verification_runs_completed_after_started_check
        CHECK (completed_at >= started_at),
    CONSTRAINT restore_verification_runs_failure_shape_check
        CHECK (
            (verification_state = 'failed' AND failure_reason IS NOT NULL AND failure_message IS NOT NULL)
            OR (verification_state = 'verified' AND failure_reason IS NULL AND failure_message IS NULL)
        )
);

CREATE INDEX IF NOT EXISTS restore_verification_runs_backup_set_started_idx
    ON restore_verification_runs (backup_set_id, started_at DESC, restore_verification_run_id ASC);

-- +goose Down
DROP TABLE IF EXISTS restore_verification_runs;

ALTER TABLE backup_sets
    DROP CONSTRAINT IF EXISTS backup_sets_last_verification_basis_sha256_check;

ALTER TABLE backup_sets
    DROP COLUMN IF EXISTS last_verification_basis_sha256;

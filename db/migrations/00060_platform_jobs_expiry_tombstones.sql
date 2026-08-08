-- +goose Up
DROP INDEX jobs_retention_lookup_idx;

ALTER TABLE jobs
    ADD COLUMN expired_at timestamptz,
    DROP CONSTRAINT jobs_terminal_summary_ck,
    DROP CONSTRAINT jobs_extension_ownership_ck,
    ADD CONSTRAINT jobs_terminal_summary_ck CHECK (
        (status IN ('queued', 'running', 'cancel_requested')
            AND finished_at IS NULL
            AND retained_until IS NULL
            AND expired_at IS NULL
            AND result_summary_json IS NULL
            AND error_summary_json IS NULL)
        OR (status IN ('succeeded', 'canceled')
            AND finished_at IS NOT NULL
            AND retained_until IS NOT NULL
            AND expired_at IS NULL
            AND result_summary_json IS NOT NULL
            AND error_summary_json IS NULL)
        OR (status = 'failed'
            AND finished_at IS NOT NULL
            AND retained_until IS NOT NULL
            AND expired_at IS NULL
            AND result_summary_json IS NULL
            AND error_summary_json IS NOT NULL)
        OR (status IN ('succeeded', 'failed', 'canceled')
            AND finished_at IS NOT NULL
            AND retained_until IS NOT NULL
            AND expired_at IS NOT NULL
            AND result_summary_json IS NULL
            AND error_summary_json IS NULL)
    ),
    ADD CONSTRAINT jobs_expiry_tombstone_ck CHECK (
        expired_at IS NULL
        OR (status IN ('succeeded', 'failed', 'canceled')
            AND expired_at >= retained_until
            AND handler_payload_json IS NULL
            AND handler_attempt_id IS NULL
            AND handler_lease_expires_at IS NULL
            AND handler_failure_count = 0
            AND handler_next_attempt_at IS NULL
            AND handler_last_attempted_at IS NULL
            AND handler_last_error IS NULL
            AND message IS NULL
            AND result_summary_json IS NULL
            AND error_summary_json IS NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
    ),
    ADD CONSTRAINT jobs_extension_ownership_ck CHECK (
        (extension_owner_profile_id IS NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
        OR (extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'
            AND expired_at IS NULL
            AND jsonb_typeof(extension_idempotency_identity) = 'object'
            AND octet_length(extension_idempotency_route_key) BETWEEN 1 AND 256
            AND octet_length(extension_idempotency_scope_key) BETWEEN 1 AND 512
            AND extension_normalized_request_sha256 ~ '^[0-9a-f]{64}$')
        OR (extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'
            AND expired_at IS NOT NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
    );

CREATE INDEX jobs_expiry_candidates_idx
    ON jobs (retained_until, job_id)
    WHERE retained_until IS NOT NULL AND expired_at IS NULL;

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    tombstone_count bigint;
BEGIN
    SELECT count(*)
      INTO tombstone_count
      FROM jobs
     WHERE expired_at IS NOT NULL;

    IF tombstone_count > 0 THEN
        RAISE EXCEPTION USING MESSAGE = format(
            'jobs expiry downgrade blocked: tombstone_count=%s; restore the pre-compaction backup or roll forward',
            tombstone_count
        );
    END IF;
END
$$;
-- +goose StatementEnd

DROP INDEX IF EXISTS jobs_expiry_candidates_idx;

ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS jobs_extension_ownership_ck,
    DROP CONSTRAINT IF EXISTS jobs_expiry_tombstone_ck,
    DROP CONSTRAINT IF EXISTS jobs_terminal_summary_ck,
    DROP COLUMN expired_at,
    ADD CONSTRAINT jobs_terminal_summary_ck CHECK (
        (status IN ('queued', 'running', 'cancel_requested')
            AND finished_at IS NULL
            AND retained_until IS NULL
            AND result_summary_json IS NULL
            AND error_summary_json IS NULL)
        OR (status IN ('succeeded', 'canceled')
            AND finished_at IS NOT NULL
            AND retained_until IS NOT NULL
            AND result_summary_json IS NOT NULL
            AND error_summary_json IS NULL)
        OR (status = 'failed'
            AND finished_at IS NOT NULL
            AND retained_until IS NOT NULL
            AND result_summary_json IS NULL
            AND error_summary_json IS NOT NULL)
    ),
    ADD CONSTRAINT jobs_extension_ownership_ck CHECK (
        (extension_owner_profile_id IS NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
        OR (extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'
            AND jsonb_typeof(extension_idempotency_identity) = 'object'
            AND octet_length(extension_idempotency_route_key) BETWEEN 1 AND 256
            AND octet_length(extension_idempotency_scope_key) BETWEEN 1 AND 512
            AND extension_normalized_request_sha256 ~ '^[0-9a-f]{64}$')
    );

CREATE INDEX jobs_retention_lookup_idx
    ON jobs (retained_until)
    WHERE retained_until IS NOT NULL;

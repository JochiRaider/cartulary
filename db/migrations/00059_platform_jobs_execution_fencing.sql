-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    missing_identity_count bigint;
    undrained_attempt_count bigint;
    bounded_tokens text;
BEGIN
    SELECT count(*)
      INTO missing_identity_count
      FROM jobs
     WHERE job_kind IS NULL
        OR progress_unit_id IS NULL
        OR handler_name IS NULL;

    SELECT count(*)
      INTO undrained_attempt_count
      FROM jobs
     WHERE handler_lease_owner IS NOT NULL
        OR handler_lease_expires_at IS NOT NULL;

    SELECT COALESCE(string_agg(token, ',' ORDER BY token), 'none')
      INTO bounded_tokens
      FROM (
          SELECT DISTINCT
                 left(CASE
                     WHEN job_kind ~ '^[a-z0-9_.-]+$' THEN job_kind
                     ELSE 'missing'
                 END, 63) || ':' || left(status, 31) AS token
            FROM jobs
           WHERE job_kind IS NULL
              OR progress_unit_id IS NULL
              OR handler_name IS NULL
              OR handler_lease_owner IS NOT NULL
              OR handler_lease_expires_at IS NOT NULL
           ORDER BY token
           LIMIT 10
      ) AS bounded;

    IF missing_identity_count > 0 OR undrained_attempt_count > 0 THEN
        RAISE EXCEPTION USING MESSAGE = format(
            'jobs execution preflight failed: missing_identity_count=%s undrained_attempt_count=%s kind_status_tokens=%s; stop and drain the old writer and use only the explicitly approved reset/reseed path for unsupported retained data',
            missing_identity_count,
            undrained_attempt_count,
            bounded_tokens
        );
    END IF;
END
$$;
-- +goose StatementEnd

DROP INDEX jobs_handler_recovery_idx;

ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS jobs_handler_attempts_ck,
    DROP COLUMN handler_attempts,
    DROP COLUMN handler_max_attempts,
    DROP COLUMN handler_lease_owner,
    ADD COLUMN handler_attempt_id uuid,
    ADD COLUMN handler_failure_count integer NOT NULL DEFAULT 0,
    ADD COLUMN handler_next_attempt_at timestamptz,
    ALTER COLUMN job_kind SET NOT NULL,
    ALTER COLUMN progress_unit_id SET NOT NULL,
    ALTER COLUMN handler_name SET NOT NULL,
    ADD CONSTRAINT jobs_handler_failure_count_ck CHECK (
        handler_failure_count BETWEEN 0 AND 3
    ),
    ADD CONSTRAINT jobs_handler_live_attempt_ck CHECK (
        (handler_attempt_id IS NULL AND handler_lease_expires_at IS NULL)
        OR (handler_attempt_id IS NOT NULL AND handler_lease_expires_at IS NOT NULL)
    ),
    ADD CONSTRAINT jobs_handler_attempt_eligibility_ck CHECK (
        handler_attempt_id IS NULL OR handler_next_attempt_at IS NULL
    ),
    ADD CONSTRAINT jobs_handler_terminal_attempt_ck CHECK (
        status NOT IN ('succeeded', 'failed', 'canceled')
        OR (handler_attempt_id IS NULL
            AND handler_lease_expires_at IS NULL
            AND handler_next_attempt_at IS NULL)
    ),
    ADD CONSTRAINT jobs_handler_retry_state_ck CHECK (
        handler_next_attempt_at IS NULL
        OR status IN ('running', 'cancel_requested')
    );

CREATE INDEX jobs_handler_recovery_idx
    ON jobs (handler_next_attempt_at, submitted_at, job_id)
    WHERE status IN ('queued', 'running', 'cancel_requested')
      AND handler_attempt_id IS NULL;

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    corrected_write_count bigint;
BEGIN
    SELECT count(*)
      INTO corrected_write_count
      FROM jobs
     WHERE handler_attempt_id IS NOT NULL
        OR handler_failure_count <> 0
        OR handler_next_attempt_at IS NOT NULL;

    IF corrected_write_count > 0 THEN
        RAISE EXCEPTION USING MESSAGE = format(
            'jobs execution downgrade blocked: corrected_write_count=%s; restore the pre-cutover backup or roll forward',
            corrected_write_count
        );
    END IF;
END
$$;
-- +goose StatementEnd

DROP INDEX IF EXISTS jobs_handler_recovery_idx;

ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS jobs_handler_retry_state_ck,
    DROP CONSTRAINT IF EXISTS jobs_handler_terminal_attempt_ck,
    DROP CONSTRAINT IF EXISTS jobs_handler_attempt_eligibility_ck,
    DROP CONSTRAINT IF EXISTS jobs_handler_live_attempt_ck,
    DROP CONSTRAINT IF EXISTS jobs_handler_failure_count_ck,
    DROP COLUMN handler_next_attempt_at,
    DROP COLUMN handler_failure_count,
    DROP COLUMN handler_attempt_id,
    ADD COLUMN handler_attempts integer NOT NULL DEFAULT 0,
    ADD COLUMN handler_max_attempts integer NOT NULL DEFAULT 3,
    ADD COLUMN handler_lease_owner text,
    ALTER COLUMN job_kind DROP NOT NULL,
    ALTER COLUMN progress_unit_id DROP NOT NULL,
    ALTER COLUMN handler_name DROP NOT NULL,
    ADD CONSTRAINT jobs_handler_attempts_ck CHECK (
        handler_attempts >= 0
        AND handler_max_attempts >= 1
        AND handler_attempts <= handler_max_attempts
    );

CREATE INDEX jobs_handler_recovery_idx ON jobs (
    handler_name,
    status,
    handler_lease_expires_at,
    submitted_at,
    job_id
) WHERE handler_name IS NOT NULL
    AND status IN ('queued', 'running', 'cancel_requested');

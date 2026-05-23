-- +goose Up
CREATE TABLE IF NOT EXISTS jobs (
    job_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_kind text NOT NULL CHECK (scope_kind IN ('incident', 'deployment')),
    incident_id uuid REFERENCES incidents (id) ON DELETE CASCADE,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'cancel_requested', 'succeeded', 'failed', 'canceled')),
    cancelable boolean NOT NULL,
    submitted_by_user_id uuid NOT NULL REFERENCES users (id),
    submitted_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    progress_completed integer NOT NULL CHECK (progress_completed >= 0),
    progress_total integer CHECK (progress_total IS NULL OR progress_total > 0),
    started_at timestamptz,
    finished_at timestamptz,
    retained_until timestamptz,
    result_summary_json jsonb,
    error_summary_json jsonb,
    message text,
    CONSTRAINT jobs_scope_incident_ck CHECK (
        (scope_kind = 'incident' AND incident_id IS NOT NULL) OR
        (scope_kind = 'deployment' AND incident_id IS NULL)
    ),
    CONSTRAINT jobs_terminal_summary_ck CHECK (
        (status IN ('queued', 'running', 'cancel_requested') AND finished_at IS NULL AND retained_until IS NULL AND result_summary_json IS NULL AND error_summary_json IS NULL) OR
        (status IN ('succeeded', 'canceled') AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND result_summary_json IS NOT NULL AND error_summary_json IS NULL) OR
        (status = 'failed' AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND result_summary_json IS NULL AND error_summary_json IS NOT NULL)
    ),
    CONSTRAINT jobs_terminal_cancelable_ck CHECK (
        (status IN ('cancel_requested', 'succeeded', 'failed', 'canceled') AND cancelable = false) OR
        (status IN ('queued', 'running'))
    ),
    CONSTRAINT jobs_progress_total_ck CHECK (
        progress_total IS NULL OR progress_completed <= progress_total
    )
);

CREATE INDEX IF NOT EXISTS jobs_submitted_by_lookup_idx
    ON jobs (submitted_by_user_id, submitted_at DESC, job_id);

CREATE INDEX IF NOT EXISTS jobs_incident_lookup_idx
    ON jobs (incident_id, submitted_at DESC, job_id)
    WHERE incident_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS jobs_retention_lookup_idx
    ON jobs (retained_until)
    WHERE retained_until IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS jobs_retention_lookup_idx;
DROP INDEX IF EXISTS jobs_incident_lookup_idx;
DROP INDEX IF EXISTS jobs_submitted_by_lookup_idx;
DROP TABLE IF EXISTS jobs;

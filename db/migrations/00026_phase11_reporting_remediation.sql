-- +goose Up
DELETE FROM reporting_release_approvals;
DELETE FROM reporting_releases;
DELETE FROM reporting_snapshots;

ALTER TABLE reporting_releases
    ADD COLUMN IF NOT EXISTS recipient_partition_refs jsonb NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_recipient_partition_refs_ck;

ALTER TABLE reporting_releases
    ADD CONSTRAINT reporting_releases_recipient_partition_refs_ck CHECK (
        jsonb_typeof(recipient_partition_refs) = 'array'
        AND (
            release_scope = 'external_release'
            OR recipient_partition_refs = '[]'::jsonb
        )
    );

ALTER TABLE reporting_releases
    ALTER COLUMN output_media_type DROP NOT NULL,
    ALTER COLUMN output_sha256 DROP NOT NULL,
    ALTER COLUMN redaction_manifest_sha256 DROP NOT NULL,
    ALTER COLUMN redaction_manifest_json DROP NOT NULL,
    ALTER COLUMN rendered_output DROP NOT NULL;

ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_render_failed_reason_ck;

ALTER TABLE reporting_releases
    ADD CONSTRAINT reporting_releases_render_failed_reason_ck CHECK (
        (
            release_state = 'render_failed'
            AND render_failed_reason_code IS NOT NULL
            AND output_media_type IS NULL
            AND output_sha256 IS NULL
            AND redaction_manifest_sha256 IS NULL
            AND redaction_manifest_json IS NULL
            AND rendered_output IS NULL
            AND approved_at IS NULL
            AND published_at IS NULL
        )
        OR
        (
            release_state <> 'render_failed'
            AND render_failed_reason_code IS NULL
            AND output_media_type IS NOT NULL
            AND output_sha256 IS NOT NULL
            AND redaction_manifest_sha256 IS NOT NULL
            AND redaction_manifest_json IS NOT NULL
            AND rendered_output IS NOT NULL
        )
    );

CREATE TABLE IF NOT EXISTS reporting_job_payloads (
    job_id uuid PRIMARY KEY REFERENCES jobs (job_id) ON DELETE CASCADE,
    job_kind text NOT NULL CHECK (job_kind IN ('snapshot_create', 'release_create')),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    actor_user_id uuid NOT NULL REFERENCES users (id),
    request_json jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS reporting_job_payloads_incident_lookup_idx
    ON reporting_job_payloads (incident_id, created_at DESC, job_id);

-- +goose Down
DROP INDEX IF EXISTS reporting_job_payloads_incident_lookup_idx;
DROP TABLE IF EXISTS reporting_job_payloads;

DELETE FROM reporting_release_approvals;
DELETE FROM reporting_releases;
DELETE FROM reporting_snapshots;

ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_render_failed_reason_ck;

ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_recipient_partition_refs_ck;

ALTER TABLE reporting_releases
    ALTER COLUMN output_media_type SET NOT NULL,
    ALTER COLUMN output_sha256 SET NOT NULL,
    ALTER COLUMN redaction_manifest_sha256 SET NOT NULL,
    ALTER COLUMN redaction_manifest_json SET NOT NULL,
    ALTER COLUMN rendered_output SET NOT NULL;

ALTER TABLE reporting_releases
    DROP COLUMN IF EXISTS recipient_partition_refs;

ALTER TABLE reporting_releases
    ADD CONSTRAINT reporting_releases_render_failed_reason_ck CHECK (
        (release_state = 'render_failed' AND render_failed_reason_code IS NOT NULL) OR
        (release_state <> 'render_failed' AND render_failed_reason_code IS NULL)
    );

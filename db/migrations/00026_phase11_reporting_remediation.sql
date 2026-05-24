-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM reporting_release_approvals LIMIT 1)
        OR EXISTS (SELECT 1 FROM reporting_releases LIMIT 1)
        OR EXISTS (SELECT 1 FROM reporting_snapshots LIMIT 1) THEN
        RAISE EXCEPTION 'phase11 reporting v3 remediation cannot run with existing reporting rows; reset reporting data before applying this migration';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE reporting_snapshots
    ADD COLUMN IF NOT EXISTS source_boundary_json jsonb;

ALTER TABLE reporting_snapshots
    ALTER COLUMN source_boundary_json SET NOT NULL;

ALTER TABLE reporting_snapshots
    DROP CONSTRAINT IF EXISTS reporting_snapshots_source_boundary_json_ck;

ALTER TABLE reporting_snapshots
    ADD CONSTRAINT reporting_snapshots_source_boundary_json_ck CHECK (
        jsonb_typeof(source_boundary_json) = 'object'
    );

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
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM reporting_release_approvals LIMIT 1)
        OR EXISTS (SELECT 1 FROM reporting_releases LIMIT 1)
        OR EXISTS (SELECT 1 FROM reporting_snapshots LIMIT 1) THEN
        RAISE EXCEPTION 'phase11 reporting v3 remediation rollback cannot run with existing reporting rows; reset reporting data before rolling back this migration';
    END IF;
END $$;
-- +goose StatementEnd

DROP INDEX IF EXISTS reporting_job_payloads_incident_lookup_idx;
DROP TABLE IF EXISTS reporting_job_payloads;

ALTER TABLE reporting_snapshots
    DROP CONSTRAINT IF EXISTS reporting_snapshots_source_boundary_json_ck;

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

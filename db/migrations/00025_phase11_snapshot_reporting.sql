-- +goose Up
CREATE TABLE IF NOT EXISTS reporting_snapshots (
    snapshot_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    client_txn_id text NOT NULL,
    snapshot_at timestamptz NOT NULL,
    source_change_set_high_watermark text NOT NULL,
    source_boundary_json jsonb NOT NULL CHECK (jsonb_typeof(source_boundary_json) = 'object'),
    derivation_version text NOT NULL,
    export_model_sha256 text NOT NULL,
    export_model_json jsonb NOT NULL,
    create_job_id uuid NOT NULL REFERENCES jobs (job_id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS reporting_snapshots_incident_lookup_idx
    ON reporting_snapshots (incident_id, created_at DESC, snapshot_id);

CREATE INDEX IF NOT EXISTS reporting_snapshots_created_by_lookup_idx
    ON reporting_snapshots (created_by_user_id, created_at DESC, snapshot_id);

CREATE TABLE IF NOT EXISTS reporting_releases (
    release_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    snapshot_id uuid NOT NULL REFERENCES reporting_snapshots (snapshot_id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    client_txn_id text NOT NULL,
    release_scope text NOT NULL CHECK (release_scope IN ('internal_draft', 'internal_review', 'external_release')),
    release_state text NOT NULL CHECK (release_state IN ('pending_approval', 'approved', 'published', 'invalidated', 'render_failed')),
    snapshot_at timestamptz NOT NULL,
    source_change_set_high_watermark text NOT NULL,
    derivation_version text NOT NULL,
    export_model_sha256 text NOT NULL,
    template_id text NOT NULL,
    template_version text NOT NULL,
    redaction_profile_id text NOT NULL,
    redaction_profile_version text NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    output_kind text NOT NULL CHECK (output_kind IN ('html', 'markdown', 'slidev', 'mermaid', 'reenactment')),
    output_media_type text NOT NULL,
    output_sha256 text NOT NULL,
    redaction_manifest_sha256 text NOT NULL,
    redaction_manifest_json jsonb NOT NULL,
    rendered_output text NOT NULL,
    create_job_id uuid NOT NULL REFERENCES jobs (job_id),
    render_failed_reason_code text,
    approved_at timestamptz,
    published_at timestamptz,
    invalidated_at timestamptz,
    invalidation_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT reporting_releases_render_failed_reason_ck CHECK (
        (release_state = 'render_failed' AND render_failed_reason_code IS NOT NULL) OR
        (release_state <> 'render_failed' AND render_failed_reason_code IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS reporting_releases_snapshot_lookup_idx
    ON reporting_releases (snapshot_id, created_at DESC, release_id);

CREATE INDEX IF NOT EXISTS reporting_releases_incident_lookup_idx
    ON reporting_releases (incident_id, created_at DESC, release_id);

CREATE INDEX IF NOT EXISTS reporting_releases_created_by_lookup_idx
    ON reporting_releases (created_by_user_id, created_at DESC, release_id);

CREATE TABLE IF NOT EXISTS reporting_release_approvals (
    approval_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    release_id uuid NOT NULL REFERENCES reporting_releases (release_id) ON DELETE CASCADE,
    actor_user_id uuid NOT NULL REFERENCES users (id),
    approval_role text NOT NULL CHECK (approval_role IN ('reviewer', 'admin')),
    reason text,
    approval_tuple_json jsonb NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    output_sha256 text NOT NULL,
    redaction_manifest_sha256 text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (release_id, actor_user_id, approval_role)
);

CREATE INDEX IF NOT EXISTS reporting_release_approvals_release_lookup_idx
    ON reporting_release_approvals (release_id, approval_role, created_at ASC, approval_id);

-- +goose Down
DROP INDEX IF EXISTS reporting_release_approvals_release_lookup_idx;
DROP TABLE IF EXISTS reporting_release_approvals;

DROP INDEX IF EXISTS reporting_releases_created_by_lookup_idx;
DROP INDEX IF EXISTS reporting_releases_incident_lookup_idx;
DROP INDEX IF EXISTS reporting_releases_snapshot_lookup_idx;
DROP TABLE IF EXISTS reporting_releases;

DROP INDEX IF EXISTS reporting_snapshots_created_by_lookup_idx;
DROP INDEX IF EXISTS reporting_snapshots_incident_lookup_idx;
DROP TABLE IF EXISTS reporting_snapshots;

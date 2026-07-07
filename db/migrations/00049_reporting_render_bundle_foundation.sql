-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM reporting_releases
         WHERE release_state <> 'render_failed'
         LIMIT 1
    ) THEN
        RAISE EXCEPTION 'reporting render-bundle migration cannot run with existing successful legacy releases; export or reset reporting release data before applying this migration';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE reporting_releases
    DROP COLUMN IF EXISTS rendered_output;

CREATE TABLE IF NOT EXISTS reporting_render_bundles (
    release_id uuid PRIMARY KEY REFERENCES reporting_releases (release_id) ON DELETE CASCADE,
    bundle_manifest_sha256 text NOT NULL CHECK (bundle_manifest_sha256 ~ '^[a-f0-9]{64}$'),
    bundle_manifest_json jsonb NOT NULL CHECK (
        jsonb_typeof(bundle_manifest_json) = 'object'
        AND bundle_manifest_json->>'schema_id' = 'cartulary.render_bundle_manifest.v1'
    ),
    primary_bundle_path text NOT NULL CHECK (primary_bundle_path <> ''),
    primary_media_type text NOT NULL CHECK (primary_media_type <> ''),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reporting_render_bundle_files (
    release_id uuid NOT NULL REFERENCES reporting_render_bundles (release_id) ON DELETE CASCADE,
    bundle_path text NOT NULL CHECK (bundle_path <> '' AND bundle_path !~ '(^/|(^|/)\.\.?(/|$))'),
    role text NOT NULL CHECK (role <> ''),
    media_type text NOT NULL CHECK (media_type <> ''),
    file_sha256 text NOT NULL CHECK (file_sha256 ~ '^[a-f0-9]{64}$'),
    size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
    storage_kind text NOT NULL CHECK (storage_kind IN ('database_inline', 'object_store')),
    object_ref text,
    inline_bytes bytea,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (release_id, bundle_path),
    CHECK (
        (
            storage_kind = 'database_inline'
            AND inline_bytes IS NOT NULL
            AND object_ref IS NULL
        )
        OR
        (
            storage_kind = 'object_store'
            AND inline_bytes IS NULL
            AND object_ref IS NOT NULL
            AND object_ref <> ''
        )
    )
);

CREATE INDEX IF NOT EXISTS reporting_render_bundle_files_release_idx
    ON reporting_render_bundle_files (release_id, role, bundle_path);

-- +goose Down
DROP INDEX IF EXISTS reporting_render_bundle_files_release_idx;
DROP TABLE IF EXISTS reporting_render_bundle_files;
DROP TABLE IF EXISTS reporting_render_bundles;

ALTER TABLE reporting_releases
    ADD COLUMN IF NOT EXISTS rendered_output text;

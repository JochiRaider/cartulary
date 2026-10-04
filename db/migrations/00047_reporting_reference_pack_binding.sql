-- +goose Up
-- Retained reports without an exact binding have already been rejected by the
-- coordinated preflight. Do not rewrite historical manifests or their digests.
ALTER TABLE public.reporting_render_bundles
    DROP CONSTRAINT reporting_render_bundles_bundle_manifest_json_check;
ALTER TABLE public.reporting_render_bundles
    ADD CONSTRAINT reporting_render_bundles_bundle_manifest_json_check CHECK ((
        jsonb_typeof(bundle_manifest_json) = 'object'
        AND bundle_manifest_json ->> 'schema_id' = 'cartulary.render_bundle_manifest.v2'
        AND bundle_manifest_json ->> 'manifest_version' = '2'
        AND jsonb_typeof(bundle_manifest_json -> 'reference_packs') = 'object'
        AND (bundle_manifest_json -> 'reference_packs' ->> 'pack_set_id') ~ '^rpset_[0-9a-f]{64}$'
        AND (bundle_manifest_json -> 'reference_packs' ->> 'pack_set_sha256') ~ '^[0-9a-f]{64}$'
        AND jsonb_typeof(bundle_manifest_json -> 'reference_packs' -> 'provenance') = 'array'
        AND bundle_manifest_json -> 'reference_packs' ->> 'pack_set_id' = 'rpset_' || (bundle_manifest_json -> 'reference_packs' ->> 'pack_set_sha256')
        AND bundle_manifest_json ? 'reference_packs'
        AND (bundle_manifest_json -> 'reference_packs') ?& ARRAY['pack_set_id', 'pack_set_sha256', 'provenance']
    ) IS TRUE);

-- +goose Down
-- The new binding cannot be discarded while retaining a truthful historical
-- render manifest. Restore a complete backup with its matching application.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'report reference binding cannot be downgraded; restore a complete matching backup' USING ERRCODE = '55000'; END $$;
-- +goose StatementEnd

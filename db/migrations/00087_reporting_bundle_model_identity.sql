-- +goose Up
-- Final assembled bytes and frozen admitted source have separate identities.
-- Never relabel or rehash old releases into the new bundle contract.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_render_bundles, public.reporting_imported_artifact_files,
        public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_render_bundles)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: retained history predates distinct admitted and assembled model identities; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_render_bundles
    DROP CONSTRAINT reporting_render_bundles_bundle_manifest_json_check;
ALTER TABLE public.reporting_render_bundles
    ADD CONSTRAINT reporting_bundle_model_identity_v3_ck CHECK ((
        jsonb_typeof(bundle_manifest_json) = 'object'
        AND bundle_manifest_json ->> 'schema_id' = 'cartulary.render_bundle_manifest.v3'
        AND bundle_manifest_json ->> 'manifest_version' = '3'
        AND bundle_manifest_json ->> 'admitted_export_model_sha256' ~ '^[0-9a-f]{64}$'
        AND bundle_manifest_json ->> 'export_model_sha256' ~ '^[0-9a-f]{64}$'
        AND jsonb_typeof(bundle_manifest_json -> 'reference_packs') = 'object'
        AND (bundle_manifest_json -> 'reference_packs' ->> 'pack_set_id') ~ '^rpset_[0-9a-f]{64}$'
        AND (bundle_manifest_json -> 'reference_packs' ->> 'pack_set_sha256') ~ '^[0-9a-f]{64}$'
        AND jsonb_typeof(bundle_manifest_json -> 'reference_packs' -> 'provenance') = 'array'
        AND bundle_manifest_json -> 'reference_packs' ->> 'pack_set_id' = 'rpset_' || (bundle_manifest_json -> 'reference_packs' ->> 'pack_set_sha256')
        AND (bundle_manifest_json -> 'reference_packs') ?& ARRAY['pack_set_id', 'pack_set_sha256', 'provenance']
    ) IS TRUE);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting bundle model identity cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

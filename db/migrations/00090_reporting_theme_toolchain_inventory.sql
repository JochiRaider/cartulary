-- +goose Up
-- Usable-font inventory cannot be inferred from an installed-font digest.
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
        RAISE EXCEPTION 'reporting_cutover_required: retained history lacks the complete toolchain font inventory; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_render_bundles ADD CONSTRAINT reporting_toolchain_inventory_v2_ck
    CHECK ((bundle_manifest_json ->> 'renderer_version' = '20') IS TRUE);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting theme/font inventory cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

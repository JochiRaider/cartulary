-- +goose Up
-- The built-in template now admits the declared manifest and binding semantics.
-- Retained releases are not reinterpreted through the replacement template.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: retained admissions lack the current template manifest; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_template_manifest_v2_ck CHECK (
    job_kind = 'snapshot_create' OR
    (request_json->'template_contract'->>'schema_id' = 'cartulary.reporting_template_pack_manifest.v2'
        AND jsonb_typeof(request_json->'template_contract'->'sections') = 'array') IS TRUE
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting template cutover cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

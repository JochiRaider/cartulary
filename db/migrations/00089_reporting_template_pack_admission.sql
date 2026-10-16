-- +goose Up
-- Template pack provenance cannot be inferred from historical job payloads.
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
        RAISE EXCEPTION 'reporting_cutover_required: retained history lacks independently admitted template manifests; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads ADD COLUMN admitted_template_manifest_sha256 text;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_template_manifest_admission_v1_ck CHECK (
    (job_kind = 'snapshot_create' AND admitted_template_manifest_sha256 IS NULL)
    OR (job_kind <> 'snapshot_create' AND (admitted_template_manifest_sha256 ~ '^[0-9a-f]{64}$') IS TRUE)
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting template pack admission cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

-- +goose Up
-- Rule decisions and consequential structural removal have distinct evidence.
-- Retained artifacts cannot acquire missing causal evidence by conversion.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: retained redaction artifacts lack structural removal evidence; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_releases DROP CONSTRAINT reporting_redaction_manifest_v3_ck;
ALTER TABLE public.reporting_composition_preview_outputs DROP CONSTRAINT reporting_preview_redaction_manifest_v3_ck;
ALTER TABLE public.reporting_releases ADD CONSTRAINT reporting_redaction_manifest_v4_ck CHECK (
    redaction_manifest_json IS NULL OR (redaction_manifest_json->>'schema_id' = 'cartulary.redaction_manifest.v4') IS TRUE
);
ALTER TABLE public.reporting_composition_preview_outputs ADD CONSTRAINT reporting_preview_redaction_manifest_v4_ck CHECK (
    (redaction_manifest_json->>'schema_id' = 'cartulary.redaction_manifest.v4') IS TRUE
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting structural removal evidence cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

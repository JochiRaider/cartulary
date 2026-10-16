-- +goose Up
-- Retained field-state validation and explicit null policy change the admitted
-- built-in profile tuple. Historical output is not converted or re-attested.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: retained outputs lack current field-state policy; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_render_profile_revision_v3_ck CHECK (
    job_kind = 'snapshot_create' OR
    (request_json->>'redaction_profile_version' = '3') IS TRUE
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting retained-field policy cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

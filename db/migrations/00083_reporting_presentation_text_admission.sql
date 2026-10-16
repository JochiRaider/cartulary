-- +goose Up
-- Composition prose now participates in redaction before presentation. Older
-- releases cannot be interpreted as evidence of the new admission contract.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: retained admissions predate presentation-text redaction; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_presentation_text_admission_v1_ck CHECK (
    job_kind = 'snapshot_create' OR
    (request_json->'export_model'->>'schema_id' = 'cartulary.reporting_export_model.v3'
        AND request_json->'export_model'->'presentation_texts' = '[]'::jsonb) IS TRUE
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting presentation-text admission cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

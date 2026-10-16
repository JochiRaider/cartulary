-- +goose Up
-- Private source selectors cannot establish retained public provenance.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: retained admissions predate canonical source references; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_snapshot_identity_v2_ck;
ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshot_identity_v3_ck CHECK ((
    export_model_json->>'schema_id' = 'cartulary.reporting_snapshot_model.v3'
    AND export_model_json->>'snapshot_id' = snapshot_id::text
    AND export_model_json->>'snapshot_model_id' ~ '^snapm_[a-f0-9]{64}$'
    AND export_model_json->'source_facts'->>'schema_id' = 'cartulary.reporting_source_facts.v1'
    AND jsonb_typeof(export_model_json->'source_facts'->'parties') = 'array'
    AND jsonb_typeof(export_model_json->'source_facts'->'assignments') = 'array'
    AND jsonb_typeof(export_model_json->'source_index') = 'array'
    AND NOT (export_model_json ?| ARRAY['release_id','preview_attempt_id','render_admitted_at','export_model_created_at','export_model_id'])
) IS TRUE);
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_presentation_text_admission_v1_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_source_index_admission_v1_ck CHECK (
    job_kind = 'snapshot_create' OR
    (request_json->'export_model'->>'schema_id' = 'cartulary.reporting_export_model.v4'
        AND request_json->'export_model'->'presentation_texts' = '[]'::jsonb
        AND jsonb_typeof(request_json->'export_model'->'source_index') = 'array') IS TRUE
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting source-reference admission cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

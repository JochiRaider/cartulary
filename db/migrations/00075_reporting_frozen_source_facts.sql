-- +goose Up
-- Source facts cannot be reconstructed from old rendered values or today's
-- mutable links. This cutover has no historical reader or provenance repair.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: frozen assignment and Party facts are unavailable; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_snapshot_identity_v1_ck;
ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshot_identity_v2_ck CHECK ((
    export_model_json->>'schema_id' = 'cartulary.reporting_snapshot_model.v2'
    AND export_model_json->>'snapshot_id' = snapshot_id::text
    AND export_model_json->>'snapshot_model_id' ~ '^snapm_[a-f0-9]{64}$'
    AND export_model_json->'source_facts'->>'schema_id' = 'cartulary.reporting_source_facts.v1'
    AND jsonb_typeof(export_model_json->'source_facts'->'parties') = 'array'
    AND jsonb_typeof(export_model_json->'source_facts'->'assignments') = 'array'
    AND NOT (export_model_json ?| ARRAY['release_id','preview_attempt_id','render_admitted_at','export_model_created_at','export_model_id'])
) IS TRUE);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting source facts cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

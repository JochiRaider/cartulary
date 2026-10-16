-- +goose Up
-- Template manifest v2 and canonical Slidev source v2 require exact admission
-- metadata. Pre-production history must not be reinterpreted under that template.
-- +goose StatementBegin
DO $$ DECLARE snapshots bigint; jobs bigint; imports bigint; previews bigint; BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    SELECT count(*) INTO snapshots FROM public.reporting_snapshots;
    SELECT count(*) INTO jobs FROM public.reporting_job_payloads;
    SELECT count(*) INTO imports FROM public.reporting_imported_artifact_files;
    SELECT count(*) INTO previews FROM public.report_composition_preview_attempts;
    IF snapshots + jobs + imports + previews > 0 THEN
        RAISE EXCEPTION 'reporting_cutover_required: snapshots=%, jobs=%, imported_artifacts=%, previews=%; use a matching backup or explicit disposable initialization', snapshots, jobs, imports, previews USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads
 ADD CONSTRAINT reporting_render_template_metadata_ck CHECK (
  job_kind = 'snapshot_create' OR COALESCE(
   jsonb_typeof(request_json->'template_contract') = 'object'
   AND jsonb_typeof(request_json->'template_contract'->'author') = 'string'
   AND jsonb_typeof(request_json->'template_contract'->'theme') = 'string'
   AND request_json->'template_contract'->>'template_id' = request_json->>'template_id'
   AND request_json->'template_contract'->>'template_version' = request_json->>'template_version'
   AND request_json->'template_contract'->>'template_id' <> ''
   AND request_json->'template_contract'->>'template_version' <> ''
   AND request_json->'template_contract'->>'theme' <> '', false)
 );

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'template metadata admission cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd

-- +goose Up
-- Freeze preview render input at admission. Retired preview history is not
-- rebuilt from current profiles, live drafts, or inferred admission facts.
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
 DROP CONSTRAINT reporting_job_payloads_job_kind_check,
 DROP CONSTRAINT reporting_release_profile_admission_ck,
 ADD CONSTRAINT reporting_job_payloads_job_kind_check CHECK (job_kind IN ('snapshot_create','release_create','composition_preview')),
 ADD CONSTRAINT reporting_render_profile_admission_ck CHECK (
  (job_kind = 'snapshot_create' AND admitted_profile_sha256 IS NULL) OR
  (job_kind IN ('release_create','composition_preview') AND COALESCE(
   jsonb_typeof(request_json->'redaction_profile') = 'object'
   AND jsonb_typeof(request_json->'redaction_profile_sha256') = 'string'
   AND admitted_profile_sha256 ~ '^[a-f0-9]{64}$'
   AND request_json->>'redaction_profile_sha256' = admitted_profile_sha256
   AND request_json->'redaction_profile'->>'schema_id' = 'cartulary.redaction_profile.v1'
   AND request_json->'redaction_profile'->>'profile_id' = request_json->>'redaction_profile_id'
   AND request_json->'redaction_profile'->>'version' = request_json->>'redaction_profile_version',false))
 );

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'frozen preview admission cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd

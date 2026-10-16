-- +goose Up
-- Profile scope/lifecycle policy must be admitted explicitly. Retired profile
-- bytes and outputs are never assigned inferred status or regenerated hashes.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: explicit frozen profile policy is unavailable; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_render_profile_admission_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_render_profile_admission_v2_ck CHECK (
    (job_kind = 'snapshot_create' AND admitted_profile_sha256 IS NULL) OR
    (job_kind IN ('release_create','composition_preview') AND (
        jsonb_typeof(request_json->'redaction_profile') = 'object'
        AND jsonb_typeof(request_json->'redaction_profile_sha256') = 'string'
        AND admitted_profile_sha256 ~ '^[a-f0-9]{64}$'
        AND request_json->>'redaction_profile_sha256' = admitted_profile_sha256
        AND request_json->'redaction_profile'->>'schema_id' = 'cartulary.redaction_profile.v2'
        AND request_json->'redaction_profile'->>'profile_id' = request_json->>'redaction_profile_id'
        AND request_json->'redaction_profile'->>'version' = request_json->>'redaction_profile_version'
        AND request_json->'redaction_profile'->>'profile_status' IN ('adopted_current','adopted_deprecated','future_only')
        AND jsonb_typeof(request_json->'redaction_profile'->'allowed_release_scopes') = 'array'
        AND request_json->>'release_scope' IN ('internal_draft','internal_review','external_release')
        AND (request_json->'redaction_profile'->'allowed_release_scopes') ? (request_json->>'release_scope')
        AND (request_json->>'release_scope' <> 'external_release' OR request_json->'redaction_profile'->>'profile_status' = 'adopted_current')
    ) IS TRUE)
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting profile policy cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

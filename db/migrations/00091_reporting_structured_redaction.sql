-- +goose Up
-- Retired split identities cannot be reconstructed as ordinary source blocks.
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
        RAISE EXCEPTION 'reporting_cutover_required: retained history predates the closed structural redaction contract; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_snapshot_identity_v5_ck;
ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshot_identity_v6_ck CHECK ((
    export_model_json->>'schema_id' = 'cartulary.reporting_snapshot_model.v6'
    AND export_model_json->>'snapshot_id' = snapshot_id::text
    AND export_model_json->>'snapshot_model_id' ~ '^snapm_[a-f0-9]{64}$'
    AND export_model_json->'source_facts'->>'schema_id' = 'cartulary.reporting_source_facts.v2'
    AND jsonb_typeof(export_model_json->'source_facts'->'parties') = 'array'
    AND jsonb_typeof(export_model_json->'source_facts'->'assignments') = 'array'
    AND jsonb_typeof(export_model_json->'source_facts'->'deleted_records') = 'array'
    AND jsonb_typeof(export_model_json->'source_index') = 'array'
    AND NOT (export_model_json ?| ARRAY['release_id','preview_attempt_id','render_admitted_at','export_model_created_at','export_model_id'])
) IS TRUE);
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_lifecycle_admission_v1_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_structured_input_v1_ck CHECK (
    job_kind = 'snapshot_create' OR
    (request_json->'export_model'->>'schema_id' = 'cartulary.reporting_export_model.v7'
        AND request_json->'source_facts'->>'schema_id' = 'cartulary.reporting_source_facts.v2'
        AND jsonb_typeof(request_json->'source_facts'->'deleted_records') = 'array'
        AND request_json->'export_model'->'presentation_texts' = '[]'::jsonb
        AND jsonb_typeof(request_json->'export_model'->'source_index') = 'array') IS TRUE
);
ALTER TABLE public.reporting_render_bundles DROP CONSTRAINT reporting_toolchain_inventory_v2_ck;
ALTER TABLE public.reporting_render_bundles ADD CONSTRAINT reporting_structured_redaction_v1_ck
    CHECK ((bundle_manifest_json->>'renderer_version' = '21') IS TRUE);
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_render_profile_revision_v4_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_render_profile_revision_v5_ck CHECK (
    job_kind = 'snapshot_create' OR (request_json->>'redaction_profile_version' = '5') IS TRUE
);
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_template_manifest_v2_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_template_manifest_v3_ck CHECK (
    job_kind = 'snapshot_create' OR (
        request_json->'template_contract'->>'schema_id' = 'cartulary.reporting_template_pack_manifest.v3'
        AND jsonb_typeof(request_json->'template_contract'->'sections') = 'array'
        AND NOT (request_json->'template_contract' ? 'aggregate_category_allowlists')
    ) IS TRUE
);
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_render_profile_admission_v2_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_render_profile_admission_v3_ck CHECK (
    (job_kind = 'snapshot_create' AND admitted_profile_sha256 IS NULL) OR
    (job_kind IN ('release_create','composition_preview') AND (
        jsonb_typeof(request_json->'redaction_profile') = 'object'
        AND jsonb_typeof(request_json->'redaction_profile_sha256') = 'string'
        AND admitted_profile_sha256 ~ '^[a-f0-9]{64}$'
        AND request_json->>'redaction_profile_sha256' = admitted_profile_sha256
        AND request_json->'redaction_profile'->>'schema_id' = 'cartulary.redaction_profile.v3'
        AND request_json->'redaction_profile'->>'profile_id' = request_json->>'redaction_profile_id'
        AND request_json->'redaction_profile'->>'version' = request_json->>'redaction_profile_version'
        AND request_json->'redaction_profile'->>'profile_status' IN ('adopted_current','adopted_deprecated','future_only')
        AND jsonb_typeof(request_json->'redaction_profile'->'allowed_release_scopes') = 'array'
        AND request_json->>'release_scope' IN ('internal_draft','internal_review','external_release')
        AND (request_json->'redaction_profile'->'allowed_release_scopes') ? (request_json->>'release_scope')
        AND (request_json->>'release_scope' <> 'external_release' OR request_json->'redaction_profile'->>'profile_status' = 'adopted_current')
        AND NOT (request_json->'redaction_profile' ? 'time_bucket_granularity')
    ) IS TRUE)
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting structural redaction cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

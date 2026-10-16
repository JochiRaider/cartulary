-- +goose Up
-- Retired render payloads cannot establish the new admission/preparation split.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_render_bundles, public.reporting_imported_artifact_files,
        public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_render_bundles)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts)
       OR EXISTS (SELECT 1 FROM public.jobs WHERE job_kind IN ('snapshot_reporting.release_create_v3','snapshot_reporting.composition_preview_v3')) THEN
        RAISE EXCEPTION 'reporting_cutover_required: render recipes require fresh initialization; preserve a matching application/data pair' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.reporting_job_payloads
    DROP CONSTRAINT reporting_render_template_metadata_ck,
    DROP CONSTRAINT reporting_structured_input_v1_ck,
    DROP CONSTRAINT reporting_render_profile_revision_v5_ck,
    DROP CONSTRAINT reporting_template_manifest_v3_ck,
    DROP CONSTRAINT reporting_render_profile_admission_v3_ck,
    ADD COLUMN render_admission_sha256 text,
    ADD COLUMN render_preparation_json jsonb,
    ADD COLUMN render_preparation_sha256 text;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_render_recipe_v1_ck CHECK (
    (job_kind = 'snapshot_create' AND admitted_profile_sha256 IS NULL
        AND render_admission_sha256 IS NULL AND render_preparation_json IS NULL AND render_preparation_sha256 IS NULL)
    OR (job_kind IN ('release_create','composition_preview') AND (
        request_json->>'schema_id' = 'cartulary.reporting_render_admission.v1'
        AND request_json->>'preparation_algorithm' = 'materialize_retained_reporting_render_v1'
        AND render_admission_sha256 ~ '^[a-f0-9]{64}$'
        AND request_json->>'source_capture_sha256' ~ '^[a-f0-9]{64}$'
        AND request_json->>'authored_redaction_profile_sha256' = admitted_profile_sha256
        AND admitted_profile_sha256 ~ '^[a-f0-9]{64}$'
        AND request_json->'authored_redaction_profile'->>'schema_id' = 'cartulary.redaction_profile.v3'
        AND request_json->'authored_redaction_profile'->>'version' = '5'
        AND (request_json->'authored_redaction_profile'->'allowed_release_scopes') ? (request_json->>'release_scope')
        AND (request_json->>'release_scope' <> 'external_release' OR request_json->'authored_redaction_profile'->>'profile_status' = 'adopted_current')
        AND request_json->'template_contract'->>'schema_id' = 'cartulary.reporting_template_pack_manifest.v3'
        AND request_json->>'template_sha256' = admitted_template_manifest_sha256
        AND request_json->>'graph_sources_sha256' = admitted_graph_sources_sha256
        AND NOT (request_json ?| ARRAY['export_model','redaction_profile','source_facts'])
        AND ((job_kind = 'release_create' AND jsonb_typeof(request_json->'identity'->'release_id') = 'string' AND request_json->'identity'->'preview_attempt_id' = 'null'::jsonb)
          OR (job_kind = 'composition_preview' AND jsonb_typeof(request_json->'identity'->'preview_attempt_id') = 'string' AND request_json->'identity'->'release_id' = 'null'::jsonb))
        AND ((render_preparation_json IS NULL AND render_preparation_sha256 IS NULL)
          OR (render_preparation_sha256 ~ '^[a-f0-9]{64}$'
            AND render_preparation_json->>'schema_id' = 'cartulary.reporting_render_preparation.v1'
            AND render_preparation_json->>'admission_sha256' = render_admission_sha256
            AND render_preparation_json->>'source_capture_sha256' = request_json->>'source_capture_sha256'
            AND render_preparation_json->>'export_model_sha256' ~ '^[a-f0-9]{64}$'
            AND render_preparation_json->>'source_facts_sha256' ~ '^[a-f0-9]{64}$'
            AND render_preparation_json->>'redaction_profile_sha256' ~ '^[a-f0-9]{64}$') IS TRUE)
    ) IS TRUE)
);
-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting render recipes cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

-- +goose Up
-- Container rule decisions cannot be inferred from old field-only evidence.
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
        RAISE EXCEPTION 'reporting_cutover_required: retained history lacks container redaction decisions; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_render_profile_revision_v3_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_render_profile_revision_v4_ck CHECK (
    job_kind = 'snapshot_create' OR (request_json->>'redaction_profile_version' = '4') IS TRUE
);

ALTER TABLE public.reporting_releases DROP CONSTRAINT reporting_redaction_manifest_v4_ck;
ALTER TABLE public.reporting_composition_preview_outputs DROP CONSTRAINT reporting_preview_redaction_manifest_v4_ck;
ALTER TABLE public.reporting_releases ADD CONSTRAINT reporting_redaction_manifest_v5_ck CHECK (
    redaction_manifest_json IS NULL OR (
        redaction_manifest_json->>'schema_id' = 'cartulary.redaction_manifest.v5'
        AND jsonb_typeof(redaction_manifest_json->'container_decisions') = 'array'
    ) IS TRUE
);
ALTER TABLE public.reporting_composition_preview_outputs ADD CONSTRAINT reporting_preview_redaction_manifest_v5_ck CHECK ((
    redaction_manifest_json->>'schema_id' = 'cartulary.redaction_manifest.v5'
    AND jsonb_typeof(redaction_manifest_json->'container_decisions') = 'array'
) IS TRUE);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting container redaction evidence cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

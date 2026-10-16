-- +goose Up
-- Canonical subject identities and display digests cannot be inferred for
-- historical jobs or artifacts. Preserve the matching application/data pair.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: canonical subject identity is unavailable; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_subject_identity_v1_ck CHECK ((
    jsonb_typeof(export_model_json->'subjects') = 'array'
    AND NOT jsonb_path_exists(export_model_json, '$.subjects[*] ? (exists(@.subject_kind) || exists(@.source_record_ref) || exists(@.display_name))')
    AND NOT jsonb_path_exists(export_model_json, '$.subjects[*] ? (!(exists(@.subject_family) && exists(@.source_record_id) && exists(@.source_record_type) && exists(@.entity_mention_id) && exists(@.canonical_display_value_sha256) && exists(@.source_refs) && exists(@.support_refs)))')
) IS TRUE);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting subject identity cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

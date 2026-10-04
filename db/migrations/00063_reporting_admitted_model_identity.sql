-- +goose Up
-- Snapshot materialization and admitted rendering have distinct identities.
-- Historical render tuples cannot be inferred safely during a cutover.
-- +goose StatementBegin
DO $$ DECLARE snapshots bigint; jobs bigint; imports bigint; previews bigint; BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    SELECT count(*) INTO snapshots FROM public.reporting_snapshots;
    SELECT count(*) INTO jobs FROM public.reporting_job_payloads;
    SELECT count(*) INTO imports FROM public.reporting_imported_artifact_files;
    SELECT count(*) INTO previews FROM public.report_composition_preview_attempts;
    IF snapshots + jobs + imports + previews > 0 THEN
        RAISE EXCEPTION 'reporting_cutover_required: snapshots=%, jobs=%, imported_artifacts=%, previews=%; preserve a matching backup, then use the explicit development reset procedure', snapshots, jobs, imports, previews USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshot_identity_v1_ck CHECK ((
    export_model_json->>'schema_id' = 'cartulary.reporting_snapshot_model.v1'
    AND export_model_json->>'snapshot_id' = snapshot_id::text
    AND export_model_json->>'snapshot_model_id' ~ '^snapm_[a-f0-9]{64}$'
    AND NOT (export_model_json ?| ARRAY['release_id','preview_attempt_id','render_admitted_at','export_model_created_at','export_model_id'])
) IS TRUE);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'admitted Reporting identities cannot be discarded by downgrade; restore a complete matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd

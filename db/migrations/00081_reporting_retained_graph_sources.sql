-- +goose Up
-- A lease is not retained render input. Do not reconstruct a completed graph
-- or source-owner labels from mutable state for historical render admissions.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: retained admissions lack frozen graph sources; preserve matching application/data history and explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads ADD COLUMN admitted_graph_sources_sha256 text;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_graph_input_v1_ck CHECK (
    (job_kind = 'snapshot_create' AND admitted_graph_sources_sha256 IS NULL) OR
    (job_kind <> 'snapshot_create'
        AND admitted_graph_sources_sha256 ~ '^[0-9a-f]{64}$'
        AND request_json->'graph_sources'->>'schema_id' = 'cartulary.reporting_graph_input.v1'
        AND jsonb_typeof(request_json->'graph_sources'->'results') = 'array') IS TRUE
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting graph retention cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

-- +goose Up
-- Failed/canceled attempts retain their original diagnostic bytes. Historical
-- attempts without that evidence cannot be reconstructed from a reason token.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads,
        public.reporting_imported_artifact_files, public.report_composition_preview_attempts IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files)
       OR EXISTS (SELECT 1 FROM public.report_composition_preview_attempts) THEN
        RAISE EXCEPTION 'reporting_cutover_required: preserve matching application/data history; explicitly initialize a fresh disposable environment' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads
    ADD COLUMN render_failure_summary bytea,
    ADD COLUMN render_failure_sha256 text,
    ADD CONSTRAINT reporting_render_failure_summary_ck CHECK (
        (render_failure_summary IS NULL AND render_failure_sha256 IS NULL)
        OR (job_kind IN ('release_create','composition_preview')
            AND render_failure_summary IS NOT NULL
            AND octet_length(render_failure_summary) BETWEEN 1 AND 1048576
            AND render_failure_sha256 IS NOT NULL
            AND render_failure_sha256 ~ '^[a-f0-9]{64}$')
    );

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting attempt evidence cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

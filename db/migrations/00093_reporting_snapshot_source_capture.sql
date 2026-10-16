-- +goose Up
-- Render projections are derived from an immutable source capture. There is
-- no lossless historical provenance upgrade from the retired model payload.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.jobs WHERE job_kind='snapshot_reporting.snapshot_create_v3') THEN
        RAISE EXCEPTION 'reporting_cutover_required: source capture requires fresh initialization' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_snapshot_identity_v6_ck;
ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_subject_identity_v1_ck;
ALTER TABLE public.reporting_snapshots DROP COLUMN export_model_json;
ALTER TABLE public.reporting_snapshots RENAME COLUMN export_model_sha256 TO source_capture_sha256;
ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshot_capture_v1_ck CHECK (source_capture_sha256 ~ '^[a-f0-9]{64}$');

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting source capture cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

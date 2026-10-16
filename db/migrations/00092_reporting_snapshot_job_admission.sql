-- +goose Up
-- The request is durable before source capture begins. Historical Jobs used
-- admission time as capture time and cannot be reinterpreted as this protocol.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.reporting_snapshots, public.reporting_job_payloads IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.reporting_snapshots)
       OR EXISTS (SELECT 1 FROM public.reporting_job_payloads)
       OR EXISTS (SELECT 1 FROM public.jobs WHERE job_kind = 'snapshot_reporting.snapshot_create_v2') THEN
        RAISE EXCEPTION 'reporting_cutover_required: snapshot admission requires fresh initialization; restore a matching application/data pair for historical work' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reporting_job_payloads ALTER COLUMN request_json DROP NOT NULL;
ALTER TABLE public.reporting_job_payloads ADD COLUMN snapshot_admission_json jsonb;
ALTER TABLE public.reporting_job_payloads ADD COLUMN snapshot_admission_sha256 text;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_snapshot_admission_v1_ck CHECK (
    (job_kind = 'snapshot_create' AND (
        jsonb_typeof(snapshot_admission_json) = 'object'
        AND snapshot_admission_json->>'schema_id' = 'cartulary.reporting_snapshot_admission.v1'
        AND snapshot_admission_sha256 ~ '^[a-f0-9]{64}$'
        AND (request_json IS NULL OR jsonb_typeof(request_json) = 'object')
    ) IS TRUE)
    OR (job_kind <> 'snapshot_create' AND request_json IS NOT NULL
        AND snapshot_admission_json IS NULL AND snapshot_admission_sha256 IS NULL)
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Reporting asynchronous snapshot admission cannot be downgraded; restore a matching application/data backup' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

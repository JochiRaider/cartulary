-- +goose Up
-- A private recipe is not an authority for recording its own rejection.
-- Reject retained retired diagnostics; never convert or delete them.
LOCK TABLE public.reporting_job_payloads IN ACCESS EXCLUSIVE MODE;
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM public.reporting_job_payloads
        WHERE render_failure_summary IS NOT NULL OR render_failure_sha256 IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'reporting failure journal cutover requires empty retired diagnostics';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE public.reporting_render_failure_records (
    job_id uuid NOT NULL,
    record_json bytea NOT NULL,
    record_sha256 text NOT NULL,
    CONSTRAINT reporting_render_failure_records_pkey PRIMARY KEY (job_id),
    CONSTRAINT reporting_render_failure_records_job_fk FOREIGN KEY (job_id) REFERENCES public.jobs(job_id) ON DELETE RESTRICT ON UPDATE NO ACTION,
    CONSTRAINT reporting_render_failure_records_bytes_ck CHECK (octet_length(record_json) BETWEEN 1 AND 2097152),
    CONSTRAINT reporting_render_failure_records_digest_ck CHECK (record_sha256 ~ '^[0-9a-f]{64}$'::text)
);

REVOKE ALL ON TABLE public.reporting_render_failure_records FROM PUBLIC;
REVOKE ALL ON TYPE public.reporting_render_failure_records FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reporting_render_failure_records TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_render_failure_records TO cartulary_recovery;
GRANT USAGE ON TYPE public.reporting_render_failure_records TO cartulary_runtime;
GRANT USAGE ON TYPE public.reporting_render_failure_records TO cartulary_recovery;

ALTER TABLE public.reporting_job_payloads
    DROP CONSTRAINT reporting_render_failure_summary_ck,
    DROP COLUMN render_failure_summary,
    DROP COLUMN render_failure_sha256;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or historical reader.
DROP TABLE public.reporting_render_failure_records;
ALTER TABLE public.reporting_job_payloads
    ADD COLUMN render_failure_summary bytea,
    ADD COLUMN render_failure_sha256 text,
    ADD CONSTRAINT reporting_render_failure_summary_ck CHECK (render_failure_summary IS NULL AND render_failure_sha256 IS NULL OR (job_kind = ANY (ARRAY['release_create'::text, 'composition_preview'::text])) AND render_failure_summary IS NOT NULL AND octet_length(render_failure_summary) >= 1 AND octet_length(render_failure_summary) <= 1048576 AND render_failure_sha256 IS NOT NULL AND render_failure_sha256 ~ '^[a-f0-9]{64}$'::text);

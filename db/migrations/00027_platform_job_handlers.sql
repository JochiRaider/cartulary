-- +goose Up
--
-- Name: jobs durable handler metadata; Type: TABLE EXTENSION; Schema: public; Owner: -
--

ALTER TABLE public.jobs
    ADD COLUMN handler_name text,
    ADD COLUMN handler_payload_json jsonb,
    ADD COLUMN handler_attempts integer DEFAULT 0 NOT NULL,
    ADD COLUMN handler_max_attempts integer DEFAULT 3 NOT NULL,
    ADD COLUMN handler_lease_owner text,
    ADD COLUMN handler_lease_expires_at timestamp with time zone,
    ADD COLUMN handler_last_attempted_at timestamp with time zone,
    ADD COLUMN handler_last_error text;

UPDATE public.jobs
   SET status = 'failed',
       cancelable = false,
       updated_at = now(),
       finished_at = now(),
       retained_until = now() + interval '7 days',
       progress_total = CASE
           WHEN progress_total IS NULL AND progress_completed > 0 THEN progress_completed
           WHEN progress_total IS NULL THEN NULL
           WHEN progress_total < progress_completed THEN progress_completed
           ELSE progress_total
       END,
       result_summary_json = NULL,
       error_summary_json = jsonb_build_object(
           'code', 'job_handler_missing',
           'message', 'Job failed closed because it has no durable handler metadata.',
           'retryable', false,
           'details', jsonb_build_object('reason_code', 'legacy_nonterminal_without_handler')
       )
 WHERE handler_name IS NULL
   AND status IN ('queued', 'running', 'cancel_requested');

ALTER TABLE public.jobs
    ADD CONSTRAINT jobs_handler_name_nonempty_ck CHECK ((handler_name IS NULL) OR (handler_name <> '')),
    ADD CONSTRAINT jobs_handler_attempts_ck CHECK ((handler_attempts >= 0) AND (handler_max_attempts > 0) AND (handler_attempts <= handler_max_attempts)),
    ADD CONSTRAINT jobs_handler_lease_owner_nonempty_ck CHECK ((handler_lease_owner IS NULL) OR (handler_lease_owner <> ''));

CREATE INDEX jobs_handler_recovery_idx ON public.jobs USING btree (
    handler_name,
    status,
    handler_lease_expires_at,
    submitted_at,
    job_id
) WHERE ((handler_name IS NOT NULL) AND (status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text])));

-- +goose Down
DROP INDEX IF EXISTS public.jobs_handler_recovery_idx;

ALTER TABLE public.jobs
    DROP CONSTRAINT IF EXISTS jobs_handler_lease_owner_nonempty_ck,
    DROP CONSTRAINT IF EXISTS jobs_handler_attempts_ck,
    DROP CONSTRAINT IF EXISTS jobs_handler_name_nonempty_ck,
    DROP COLUMN IF EXISTS handler_last_error,
    DROP COLUMN IF EXISTS handler_last_attempted_at,
    DROP COLUMN IF EXISTS handler_lease_expires_at,
    DROP COLUMN IF EXISTS handler_lease_owner,
    DROP COLUMN IF EXISTS handler_max_attempts,
    DROP COLUMN IF EXISTS handler_attempts,
    DROP COLUMN IF EXISTS handler_payload_json,
    DROP COLUMN IF EXISTS handler_name;

-- +goose Up
CREATE TABLE public.jobs (
    job_id uuid DEFAULT gen_random_uuid() NOT NULL,
    scope_kind text NOT NULL,
    incident_id uuid,
    status text NOT NULL,
    cancelable boolean NOT NULL,
    submitted_by_user_id uuid NOT NULL,
    submitted_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    progress_completed integer NOT NULL,
    progress_total integer,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    retained_until timestamp with time zone,
    result_summary_json jsonb,
    error_summary_json jsonb,
    message text,
    auth_policy text DEFAULT 'submitter_or_deployment_admin'::text NOT NULL,
    handler_name text NOT NULL,
    handler_payload_json jsonb,
    handler_lease_expires_at timestamp with time zone,
    handler_last_attempted_at timestamp with time zone,
    handler_last_error text,
    extension_owner_profile_id text,
    job_kind text NOT NULL,
    extension_idempotency_identity jsonb,
    extension_idempotency_route_key text,
    extension_idempotency_scope_key text,
    extension_normalized_request_sha256 text,
    progress_unit_id text NOT NULL,
    handler_attempt_id uuid,
    handler_failure_count integer DEFAULT 0 NOT NULL,
    handler_next_attempt_at timestamp with time zone,
    expired_at timestamp with time zone,
    CONSTRAINT jobs_pkey PRIMARY KEY (job_id),
    CONSTRAINT jobs_auth_policy_ck CHECK (
        (scope_kind = 'incident' AND auth_policy IN ('incident_membership', 'deployment_admin_incident_membership'))
        OR (scope_kind = 'deployment' AND auth_policy IN ('submitter_or_deployment_admin', 'deployment_admin'))
    ),
    CONSTRAINT jobs_definition_pair_ck CHECK (
        (job_kind IS NULL AND progress_unit_id IS NULL)
        OR (job_kind IS NOT NULL AND progress_unit_id IS NOT NULL)
    ),
    CONSTRAINT jobs_expiry_tombstone_ck CHECK (
        expired_at IS NULL
        OR (status IN ('succeeded', 'failed', 'canceled')
            AND expired_at >= retained_until
            AND handler_payload_json IS NULL
            AND handler_attempt_id IS NULL
            AND handler_lease_expires_at IS NULL
            AND handler_failure_count = 0
            AND handler_next_attempt_at IS NULL
            AND handler_last_attempted_at IS NULL
            AND handler_last_error IS NULL
            AND message IS NULL
            AND result_summary_json IS NULL
            AND error_summary_json IS NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
    ),
    CONSTRAINT jobs_extension_ownership_ck CHECK (
        (extension_owner_profile_id IS NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
        OR (extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'
            AND expired_at IS NULL
            AND jsonb_typeof(extension_idempotency_identity) = 'object'
            AND octet_length(extension_idempotency_route_key) BETWEEN 1 AND 256
            AND octet_length(extension_idempotency_scope_key) BETWEEN 1 AND 512
            AND extension_normalized_request_sha256 ~ '^[0-9a-f]{64}$')
        OR (extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'
            AND expired_at IS NOT NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
    ),
    CONSTRAINT jobs_handler_attempt_eligibility_ck CHECK (
        handler_attempt_id IS NULL OR handler_next_attempt_at IS NULL
    ),
    CONSTRAINT jobs_handler_failure_count_ck CHECK (handler_failure_count BETWEEN 0 AND 3),
    CONSTRAINT jobs_handler_live_attempt_ck CHECK (
        (handler_attempt_id IS NULL AND handler_lease_expires_at IS NULL)
        OR (handler_attempt_id IS NOT NULL AND handler_lease_expires_at IS NOT NULL)
    ),
    CONSTRAINT jobs_handler_name_nonempty_ck CHECK (handler_name <> ''),
    CONSTRAINT jobs_handler_retry_state_ck CHECK (
        handler_next_attempt_at IS NULL OR status IN ('running', 'cancel_requested')
    ),
    CONSTRAINT jobs_handler_terminal_attempt_ck CHECK (
        status NOT IN ('succeeded', 'failed', 'canceled')
        OR (handler_attempt_id IS NULL AND handler_lease_expires_at IS NULL AND handler_next_attempt_at IS NULL)
    ),
    CONSTRAINT jobs_job_kind_nonempty_ck CHECK (job_kind <> '' AND length(job_kind) <= 191),
    CONSTRAINT jobs_nonterminal_definition_ck CHECK (
        status IN ('succeeded', 'failed', 'canceled')
        OR (job_kind IS NOT NULL AND progress_unit_id IS NOT NULL)
    ),
    CONSTRAINT jobs_progress_completed_check CHECK (progress_completed >= 0),
    CONSTRAINT jobs_progress_total_check CHECK (progress_total IS NULL OR progress_total > 0),
    CONSTRAINT jobs_progress_total_ck CHECK (progress_total IS NULL OR progress_completed <= progress_total),
    CONSTRAINT jobs_progress_unit_id_shape_ck CHECK (
        progress_unit_id ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*){2,}\.v[1-9][0-9]*$'
    ),
    CONSTRAINT jobs_scope_incident_ck CHECK (
        (scope_kind = 'incident' AND incident_id IS NOT NULL)
        OR (scope_kind = 'deployment' AND incident_id IS NULL)
    ),
    CONSTRAINT jobs_scope_kind_check CHECK (scope_kind IN ('incident', 'deployment')),
    CONSTRAINT jobs_status_check CHECK (
        status IN ('queued', 'running', 'cancel_requested', 'succeeded', 'failed', 'canceled')
    ),
    CONSTRAINT jobs_succeeded_progress_ck CHECK (
        status <> 'succeeded' OR progress_total IS NULL OR progress_completed = progress_total
    ),
    CONSTRAINT jobs_terminal_cancelable_ck CHECK (
        (status IN ('cancel_requested', 'succeeded', 'failed', 'canceled') AND cancelable = false)
        OR status IN ('queued', 'running')
    ),
    CONSTRAINT jobs_terminal_summary_ck CHECK (
        (status IN ('queued', 'running', 'cancel_requested')
            AND finished_at IS NULL AND retained_until IS NULL AND expired_at IS NULL
            AND result_summary_json IS NULL AND error_summary_json IS NULL)
        OR (status IN ('succeeded', 'canceled')
            AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND expired_at IS NULL
            AND result_summary_json IS NOT NULL AND error_summary_json IS NULL)
        OR (status = 'failed'
            AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND expired_at IS NULL
            AND result_summary_json IS NULL AND error_summary_json IS NOT NULL)
        OR (status IN ('succeeded', 'failed', 'canceled')
            AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND expired_at IS NOT NULL
            AND result_summary_json IS NULL AND error_summary_json IS NULL)
    ),
    CONSTRAINT jobs_incident_id_fkey FOREIGN KEY (incident_id)
        REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT jobs_submitted_by_user_id_fkey FOREIGN KEY (submitted_by_user_id)
        REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX jobs_expiry_candidates_idx
    ON public.jobs USING btree (retained_until, job_id)
    WHERE retained_until IS NOT NULL AND expired_at IS NULL;
CREATE INDEX jobs_extension_nonterminal_idx
    ON public.jobs USING btree (extension_owner_profile_id, submitted_at, job_id)
    WHERE extension_owner_profile_id IS NOT NULL
      AND status IN ('queued', 'running', 'cancel_requested');
CREATE INDEX jobs_handler_recovery_idx
    ON public.jobs USING btree (handler_next_attempt_at, submitted_at, job_id)
    WHERE status IN ('queued', 'running', 'cancel_requested') AND handler_attempt_id IS NULL;
CREATE INDEX jobs_incident_lookup_idx
    ON public.jobs USING btree (incident_id, submitted_at DESC, job_id)
    WHERE incident_id IS NOT NULL;
CREATE INDEX jobs_submitted_by_lookup_idx
    ON public.jobs USING btree (submitted_by_user_id, submitted_at DESC, job_id);

CREATE INDEX jobs_incident_id_fk_idx ON public.jobs (incident_id);

-- +goose Down
DROP TABLE public.jobs;

-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: platform jobs.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.jobs (
    job_id uuid DEFAULT gen_random_uuid() NOT NULL,
    scope_kind text NOT NULL,
    incident_id uuid,
    status text NOT NULL,
    cancelable boolean NOT NULL,
    submitted_by_user_id uuid,
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
    submitting_operator_operation_id uuid,
    CONSTRAINT jobs_auth_policy_ck CHECK (scope_kind = 'incident'::text AND (auth_policy = ANY (ARRAY['incident_membership'::text, 'deployment_admin_incident_membership'::text])) OR scope_kind = 'deployment'::text AND (auth_policy = ANY (ARRAY['submitter_or_deployment_admin'::text, 'deployment_admin'::text]))),
    CONSTRAINT jobs_definition_pair_ck CHECK (job_kind IS NULL AND progress_unit_id IS NULL OR job_kind IS NOT NULL AND progress_unit_id IS NOT NULL),
    CONSTRAINT jobs_expiry_tombstone_ck CHECK (expired_at IS NULL OR (status = ANY (ARRAY['succeeded'::text, 'failed'::text, 'canceled'::text])) AND expired_at >= retained_until AND handler_payload_json IS NULL AND handler_attempt_id IS NULL AND handler_lease_expires_at IS NULL AND handler_failure_count = 0 AND handler_next_attempt_at IS NULL AND handler_last_attempted_at IS NULL AND handler_last_error IS NULL AND message IS NULL AND result_summary_json IS NULL AND error_summary_json IS NULL AND extension_idempotency_identity IS NULL AND extension_idempotency_route_key IS NULL AND extension_idempotency_scope_key IS NULL AND extension_normalized_request_sha256 IS NULL),
    CONSTRAINT jobs_extension_ownership_ck CHECK (extension_owner_profile_id IS NULL AND extension_idempotency_identity IS NULL AND extension_idempotency_route_key IS NULL AND extension_idempotency_scope_key IS NULL AND extension_normalized_request_sha256 IS NULL OR extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'::text AND expired_at IS NULL AND jsonb_typeof(extension_idempotency_identity) = 'object'::text AND octet_length(extension_idempotency_route_key) >= 1 AND octet_length(extension_idempotency_route_key) <= 256 AND octet_length(extension_idempotency_scope_key) >= 1 AND octet_length(extension_idempotency_scope_key) <= 512 AND extension_normalized_request_sha256 ~ '^[0-9a-f]{64}$'::text OR extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'::text AND expired_at IS NOT NULL AND extension_idempotency_identity IS NULL AND extension_idempotency_route_key IS NULL AND extension_idempotency_scope_key IS NULL AND extension_normalized_request_sha256 IS NULL),
    CONSTRAINT jobs_handler_attempt_eligibility_ck CHECK (handler_attempt_id IS NULL OR handler_next_attempt_at IS NULL),
    CONSTRAINT jobs_handler_failure_count_ck CHECK (handler_failure_count >= 0 AND handler_failure_count <= 3),
    CONSTRAINT jobs_handler_live_attempt_ck CHECK (handler_attempt_id IS NULL AND handler_lease_expires_at IS NULL OR handler_attempt_id IS NOT NULL AND handler_lease_expires_at IS NOT NULL),
    CONSTRAINT jobs_handler_name_nonempty_ck CHECK (handler_name <> ''::text),
    CONSTRAINT jobs_handler_retry_state_ck CHECK (handler_next_attempt_at IS NULL OR (status = ANY (ARRAY['running'::text, 'cancel_requested'::text]))),
    CONSTRAINT jobs_handler_terminal_attempt_ck CHECK ((status <> ALL (ARRAY['succeeded'::text, 'failed'::text, 'canceled'::text])) OR handler_attempt_id IS NULL AND handler_lease_expires_at IS NULL AND handler_next_attempt_at IS NULL),
    CONSTRAINT jobs_job_kind_nonempty_ck CHECK (job_kind <> ''::text AND length(job_kind) <= 191),
    CONSTRAINT jobs_nonterminal_definition_ck CHECK ((status = ANY (ARRAY['succeeded'::text, 'failed'::text, 'canceled'::text])) OR job_kind IS NOT NULL AND progress_unit_id IS NOT NULL),
    CONSTRAINT jobs_pkey PRIMARY KEY (job_id),
    CONSTRAINT jobs_progress_completed_check CHECK (progress_completed >= 0),
    CONSTRAINT jobs_progress_total_check CHECK (progress_total IS NULL OR progress_total > 0),
    CONSTRAINT jobs_progress_total_ck CHECK (progress_total IS NULL OR progress_completed <= progress_total),
    CONSTRAINT jobs_progress_unit_id_shape_ck CHECK (progress_unit_id ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*){2,}\.v[1-9][0-9]*$'::text),
    CONSTRAINT jobs_scope_incident_ck CHECK (scope_kind = 'incident'::text AND incident_id IS NOT NULL OR scope_kind = 'deployment'::text AND incident_id IS NULL),
    CONSTRAINT jobs_scope_kind_check CHECK (scope_kind = ANY (ARRAY['incident'::text, 'deployment'::text])),
    CONSTRAINT jobs_status_check CHECK (status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text, 'succeeded'::text, 'failed'::text, 'canceled'::text])),
    CONSTRAINT jobs_submitting_actor_ck CHECK (submitted_by_user_id IS NOT NULL AND submitting_operator_operation_id IS NULL OR (submitted_by_user_id IS NULL AND submitting_operator_operation_id IS NOT NULL AND submitting_operator_operation_id <> '00000000-0000-0000-0000-000000000000'::uuid AND scope_kind = 'deployment'::text AND incident_id IS NULL AND auth_policy = 'deployment_admin'::text AND extension_owner_profile_id IS NOT NULL AND (expired_at IS NOT NULL OR (extension_idempotency_identity ->> 'schema_id'::text) = 'cartulary.route_scoped_idempotency_identity.v2'::text AND (extension_idempotency_identity ->> 'actor_kind'::text) = 'local_operator'::text AND (extension_idempotency_identity -> 'actor_user_id'::text) = 'null'::jsonb AND (extension_idempotency_identity ->> 'operator_operation_id'::text) = submitting_operator_operation_id::text AND (extension_idempotency_identity ->> 'client_txn_id'::text) = submitting_operator_operation_id::text)) IS TRUE),
    CONSTRAINT jobs_succeeded_progress_ck CHECK (status <> 'succeeded'::text OR progress_total IS NULL OR progress_completed = progress_total),
    CONSTRAINT jobs_terminal_cancelable_ck CHECK ((status = ANY (ARRAY['cancel_requested'::text, 'succeeded'::text, 'failed'::text, 'canceled'::text])) AND cancelable = false OR (status = ANY (ARRAY['queued'::text, 'running'::text]))),
    CONSTRAINT jobs_terminal_summary_ck CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text])) AND finished_at IS NULL AND retained_until IS NULL AND expired_at IS NULL AND result_summary_json IS NULL AND error_summary_json IS NULL OR (status = ANY (ARRAY['succeeded'::text, 'canceled'::text])) AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND expired_at IS NULL AND result_summary_json IS NOT NULL AND error_summary_json IS NULL OR status = 'failed'::text AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND expired_at IS NULL AND result_summary_json IS NULL AND error_summary_json IS NOT NULL OR (status = ANY (ARRAY['succeeded'::text, 'failed'::text, 'canceled'::text])) AND finished_at IS NOT NULL AND retained_until IS NOT NULL AND expired_at IS NOT NULL AND result_summary_json IS NULL AND error_summary_json IS NULL),
    CONSTRAINT jobs_incident_identity_key UNIQUE (incident_id, job_id)
);

REVOKE ALL ON TABLE public.jobs FROM PUBLIC;

REVOKE ALL ON TYPE public.jobs FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.jobs TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.jobs TO cartulary_recovery;

CREATE INDEX jobs_expiry_candidates_idx ON public.jobs USING btree (retained_until, job_id) WHERE ((retained_until IS NOT NULL) AND (expired_at IS NULL));

CREATE INDEX jobs_extension_nonterminal_idx ON public.jobs USING btree (extension_owner_profile_id, submitted_at, job_id) WHERE ((extension_owner_profile_id IS NOT NULL) AND (status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text])));

CREATE INDEX jobs_handler_recovery_idx ON public.jobs USING btree (handler_next_attempt_at, submitted_at, job_id) WHERE ((status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text])) AND (handler_attempt_id IS NULL));

CREATE INDEX jobs_incident_id_fk_idx ON public.jobs USING btree (incident_id);

CREATE INDEX jobs_incident_lookup_idx ON public.jobs USING btree (incident_id, submitted_at DESC, job_id) WHERE (incident_id IS NOT NULL);

CREATE INDEX jobs_restored_nonterminal_order_idx ON public.jobs USING btree (extension_owner_profile_id, job_kind, job_id) WHERE (status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text]));

CREATE INDEX jobs_submitted_by_lookup_idx ON public.jobs USING btree (submitted_by_user_id, submitted_at DESC, job_id);

ALTER TABLE public.jobs ADD CONSTRAINT jobs_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.jobs ADD CONSTRAINT jobs_submitted_by_user_id_fkey FOREIGN KEY (submitted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.jobs DROP CONSTRAINT jobs_submitted_by_user_id_fkey;
ALTER TABLE public.jobs DROP CONSTRAINT jobs_incident_id_fkey;
DROP TABLE public.jobs;

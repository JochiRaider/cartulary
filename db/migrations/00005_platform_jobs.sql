-- +goose Up
--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

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
    CONSTRAINT jobs_auth_policy_ck CHECK ((((scope_kind = 'incident'::text) AND (auth_policy = ANY (ARRAY['incident_membership'::text, 'deployment_admin_incident_membership'::text]))) OR ((scope_kind = 'deployment'::text) AND (auth_policy = ANY (ARRAY['submitter_or_deployment_admin'::text, 'deployment_admin'::text]))))),
    CONSTRAINT jobs_progress_completed_check CHECK ((progress_completed >= 0)),
    CONSTRAINT jobs_progress_total_check CHECK (((progress_total IS NULL) OR (progress_total > 0))),
    CONSTRAINT jobs_progress_total_ck CHECK (((progress_total IS NULL) OR (progress_completed <= progress_total))),
    CONSTRAINT jobs_scope_incident_ck CHECK ((((scope_kind = 'incident'::text) AND (incident_id IS NOT NULL)) OR ((scope_kind = 'deployment'::text) AND (incident_id IS NULL)))),
    CONSTRAINT jobs_scope_kind_check CHECK ((scope_kind = ANY (ARRAY['incident'::text, 'deployment'::text]))),
    CONSTRAINT jobs_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text, 'succeeded'::text, 'failed'::text, 'canceled'::text]))),
    CONSTRAINT jobs_terminal_cancelable_ck CHECK ((((status = ANY (ARRAY['cancel_requested'::text, 'succeeded'::text, 'failed'::text, 'canceled'::text])) AND (cancelable = false)) OR (status = ANY (ARRAY['queued'::text, 'running'::text])))),
    CONSTRAINT jobs_terminal_summary_ck CHECK ((((status = ANY (ARRAY['queued'::text, 'running'::text, 'cancel_requested'::text])) AND (finished_at IS NULL) AND (retained_until IS NULL) AND (result_summary_json IS NULL) AND (error_summary_json IS NULL)) OR ((status = ANY (ARRAY['succeeded'::text, 'canceled'::text])) AND (finished_at IS NOT NULL) AND (retained_until IS NOT NULL) AND (result_summary_json IS NOT NULL) AND (error_summary_json IS NULL)) OR ((status = 'failed'::text) AND (finished_at IS NOT NULL) AND (retained_until IS NOT NULL) AND (result_summary_json IS NULL) AND (error_summary_json IS NOT NULL))))
);

--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (job_id);

--
-- Name: jobs_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX jobs_incident_lookup_idx ON public.jobs USING btree (incident_id, submitted_at DESC, job_id) WHERE (incident_id IS NOT NULL);

--
-- Name: jobs_retention_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX jobs_retention_lookup_idx ON public.jobs USING btree (retained_until) WHERE (retained_until IS NOT NULL);

--
-- Name: jobs_submitted_by_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX jobs_submitted_by_lookup_idx ON public.jobs USING btree (submitted_by_user_id, submitted_at DESC, job_id);

--
-- Name: jobs jobs_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: jobs jobs_submitted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_submitted_by_user_id_fkey FOREIGN KEY (submitted_by_user_id) REFERENCES public.users(id);

-- +goose Down
DROP TABLE IF EXISTS public.jobs CASCADE;

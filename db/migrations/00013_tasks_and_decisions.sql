-- +goose Up
--
-- Name: decisions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.decisions (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    summary text,
    status text DEFAULT 'proposed'::text NOT NULL,
    owner_user_id uuid,
    decision_type text,
    decided_at timestamp with time zone,
    rationale text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: task_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_requests (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    title text,
    status text DEFAULT 'open'::text NOT NULL,
    owner_user_id uuid,
    priority text,
    task_kind text,
    workstream text,
    due_at timestamp with time zone,
    requester_party_text text,
    requester_party_id uuid,
    blocked_reason text,
    completed_at timestamp with time zone,
    external_ticket_ref text,
    closure_summary text,
    decision_record_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: decisions decisions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.decisions
    ADD CONSTRAINT decisions_pkey PRIMARY KEY (record_id);

--
-- Name: task_requests task_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_requests
    ADD CONSTRAINT task_requests_pkey PRIMARY KEY (record_id);

--
-- Name: decisions_incident_decided_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX decisions_incident_decided_idx ON public.decisions USING btree (incident_id, decided_at DESC, record_id);

--
-- Name: task_requests_incident_updated_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX task_requests_incident_updated_idx ON public.task_requests USING btree (incident_id, updated_at DESC, record_id);

--
-- Name: task_requests_requester_party_ref_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX task_requests_requester_party_ref_lookup_idx ON public.task_requests USING btree (incident_id, requester_party_id) WHERE (requester_party_id IS NOT NULL);

--
-- Name: decisions decisions_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.decisions
    ADD CONSTRAINT decisions_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: decisions decisions_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.decisions
    ADD CONSTRAINT decisions_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id);

--
-- Name: decisions decisions_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.decisions
    ADD CONSTRAINT decisions_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: task_requests task_requests_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_requests
    ADD CONSTRAINT task_requests_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: task_requests task_requests_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_requests
    ADD CONSTRAINT task_requests_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id);

--
-- Name: task_requests task_requests_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_requests
    ADD CONSTRAINT task_requests_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.task_requests CASCADE;
DROP TABLE IF EXISTS public.decisions CASCADE;

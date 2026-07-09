-- +goose Up
--
-- Name: incident_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_memberships (
    incident_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text NOT NULL,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    added_by_user_id uuid NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by_user_id uuid NOT NULL,
    membership_version bigint DEFAULT 1 NOT NULL,
    CONSTRAINT incident_memberships_role_check CHECK ((role = ANY (ARRAY['viewer'::text, 'editor'::text, 'reviewer'::text, 'admin'::text])))
);

--
-- Name: incident_workbook_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_workbook_preferences (
    incident_id uuid NOT NULL,
    default_sheet_ref jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by_user_id uuid NOT NULL
);

--
-- Name: incidents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incidents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_key text NOT NULL,
    incident_key_canonical text NOT NULL,
    title text NOT NULL,
    description text,
    status text NOT NULL,
    severity text,
    tlp text,
    current_phase text,
    primary_external_case_ref text,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by_user_id uuid NOT NULL,
    incident_version bigint DEFAULT 1 NOT NULL,
    closed_at timestamp with time zone,
    CONSTRAINT incidents_closed_at_status_ck CHECK ((((status = 'active'::text) AND (closed_at IS NULL)) OR ((status = 'closed'::text) AND (closed_at IS NOT NULL)))),
    CONSTRAINT incidents_current_phase_contract_ck CHECK (((current_phase IS NULL) OR (((char_length(current_phase) >= 1) AND (char_length(current_phase) <= 128)) AND (current_phase !~ '[[:cntrl:]]'::text)))),
    CONSTRAINT incidents_description_contract_ck CHECK (((description IS NULL) OR (((char_length(description) >= 1) AND (char_length(description) <= 16384)) AND (regexp_replace(description, '[\n\t]'::text, ''::text, 'g'::text) !~ '[[:cntrl:]]'::text)))),
    CONSTRAINT incidents_primary_external_case_ref_contract_ck CHECK (((primary_external_case_ref IS NULL) OR (((char_length(primary_external_case_ref) >= 1) AND (char_length(primary_external_case_ref) <= 128)) AND (primary_external_case_ref !~ '[[:cntrl:]]'::text)))),
    CONSTRAINT incidents_severity_contract_ck CHECK (((severity IS NULL) OR (((char_length(severity) >= 1) AND (char_length(severity) <= 128)) AND (severity !~ '[[:cntrl:]]'::text)))),
    CONSTRAINT incidents_status_ck CHECK ((status = ANY (ARRAY['active'::text, 'closed'::text]))),
    CONSTRAINT incidents_tlp_ck CHECK (((tlp IS NULL) OR (tlp = ANY (ARRAY['TLP:CLEAR'::text, 'TLP:GREEN'::text, 'TLP:AMBER'::text, 'TLP:AMBER+STRICT'::text, 'TLP:RED'::text]))))
);

--
-- Name: user_workbook_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_workbook_preferences (
    incident_id uuid NOT NULL,
    user_id uuid NOT NULL,
    home_sheet_ref jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: incident_memberships incident_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_memberships
    ADD CONSTRAINT incident_memberships_pkey PRIMARY KEY (incident_id, user_id);

--
-- Name: incident_workbook_preferences incident_workbook_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_workbook_preferences
    ADD CONSTRAINT incident_workbook_preferences_pkey PRIMARY KEY (incident_id);

--
-- Name: incidents incidents_incident_key_canonical_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents
    ADD CONSTRAINT incidents_incident_key_canonical_key UNIQUE (incident_key_canonical);

--
-- Name: incidents incidents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents
    ADD CONSTRAINT incidents_pkey PRIMARY KEY (id);

--
-- Name: user_workbook_preferences user_workbook_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_workbook_preferences
    ADD CONSTRAINT user_workbook_preferences_pkey PRIMARY KEY (incident_id, user_id);

--
-- Name: incident_memberships_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX incident_memberships_incident_lookup_idx ON public.incident_memberships USING btree (incident_id, joined_at, user_id);

--
-- Name: incident_memberships_user_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX incident_memberships_user_lookup_idx ON public.incident_memberships USING btree (user_id, incident_id);

--
-- Name: incidents_status_updated_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX incidents_status_updated_lookup_idx ON public.incidents USING btree (status, updated_at DESC, id);

--
-- Name: incidents_updated_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX incidents_updated_lookup_idx ON public.incidents USING btree (updated_at DESC, id);

--
-- Name: deployment_admin_audit_events deployment_admin_audit_events_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.deployment_admin_audit_events
    ADD CONSTRAINT deployment_admin_audit_events_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id);

--
-- Name: incident_memberships incident_memberships_added_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_memberships
    ADD CONSTRAINT incident_memberships_added_by_user_id_fkey FOREIGN KEY (added_by_user_id) REFERENCES public.users(id);

--
-- Name: incident_memberships incident_memberships_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_memberships
    ADD CONSTRAINT incident_memberships_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: incident_memberships incident_memberships_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_memberships
    ADD CONSTRAINT incident_memberships_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

--
-- Name: incident_memberships incident_memberships_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_memberships
    ADD CONSTRAINT incident_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: incident_workbook_preferences incident_workbook_preferences_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_workbook_preferences
    ADD CONSTRAINT incident_workbook_preferences_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: incident_workbook_preferences incident_workbook_preferences_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_workbook_preferences
    ADD CONSTRAINT incident_workbook_preferences_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

--
-- Name: incidents incidents_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents
    ADD CONSTRAINT incidents_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: incidents incidents_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents
    ADD CONSTRAINT incidents_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

--
-- Name: user_workbook_preferences user_workbook_preferences_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_workbook_preferences
    ADD CONSTRAINT user_workbook_preferences_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: user_workbook_preferences user_workbook_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_workbook_preferences
    ADD CONSTRAINT user_workbook_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.incidents CASCADE;
DROP TABLE IF EXISTS public.incident_memberships CASCADE;
DROP TABLE IF EXISTS public.incident_workbook_preferences CASCADE;
DROP TABLE IF EXISTS public.user_workbook_preferences CASCADE;

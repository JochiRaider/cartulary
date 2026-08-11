-- +goose Up
--
-- Name: parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.parties (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    display_name text,
    party_kind text,
    organization_name text,
    role_title text,
    primary_email text,
    timezone_name text,
    external_ref text,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: parties parties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.parties
    ADD CONSTRAINT parties_pkey PRIMARY KEY (record_id);

--
-- Name: parties_incident_display_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX parties_incident_display_name_idx ON public.parties USING btree (incident_id, display_name, record_id);

--
-- Name: parties parties_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.parties
    ADD CONSTRAINT parties_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: parties parties_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.parties
    ADD CONSTRAINT parties_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX parties_incident_id_record_id_fk_idx ON public.parties (incident_id, record_id);

-- +goose Down
DROP TABLE public.parties ;

-- +goose Up
--
-- Name: timeline_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.timeline_events (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    capture_state text NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    recorded_at timestamp with time zone DEFAULT now() NOT NULL,
    edited_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    updated_by_user_id uuid NOT NULL,
    reviewed_by_user_id uuid,
    reviewed_at timestamp with time zone,
    superseded_by_user_id uuid,
    superseded_at timestamp with time zone,
    raw_capture jsonb DEFAULT '{}'::jsonb NOT NULL,
    date_entered_text text,
    analyst_text text,
    mitre_stage_text text,
    device_object_text text,
    ip_address_text text,
    activity_utc_text text,
    activity_local_text text,
    raw_activity_text text,
    activity_synopsis_text text,
    data_source_text text,
    activity_utc_generated boolean DEFAULT false NOT NULL,
    activity_local_generated boolean DEFAULT false NOT NULL,
    activity_time_pair_state text DEFAULT 'disabled'::text NOT NULL,
    CONSTRAINT timeline_events_activity_time_pair_state_check CHECK ((activity_time_pair_state = ANY (ARRAY['disabled'::text, 'empty'::text, 'paired_generated'::text, 'paired_user_preserved'::text, 'paired_mismatch'::text, 'conversion_unavailable'::text]))),
    CONSTRAINT timeline_events_capture_state_check CHECK ((capture_state = ANY (ARRAY['rough'::text, 'enriched'::text, 'reviewed'::text, 'superseded'::text])))
);

--
-- Name: timeline_time_conversion_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.timeline_time_conversion_profiles (
    incident_id uuid NOT NULL,
    enabled boolean DEFAULT false NOT NULL,
    local_offset_minutes integer,
    local_label text,
    profile_version bigint DEFAULT 1 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by_user_id uuid,
    CONSTRAINT timeline_time_conversion_profiles_enabled_offset_ck CHECK (((enabled = false) OR (local_offset_minutes IS NOT NULL))),
    CONSTRAINT timeline_time_conversion_profiles_local_offset_minutes_check CHECK (((local_offset_minutes IS NULL) OR ((local_offset_minutes >= '-840'::integer) AND (local_offset_minutes <= 840))))
);

--
-- Name: timeline_events timeline_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_events
    ADD CONSTRAINT timeline_events_pkey PRIMARY KEY (record_id);

--
-- Name: timeline_time_conversion_profiles timeline_time_conversion_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_time_conversion_profiles
    ADD CONSTRAINT timeline_time_conversion_profiles_pkey PRIMARY KEY (incident_id);

--
-- Name: timeline_events_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX timeline_events_incident_lookup_idx ON public.timeline_events USING btree (incident_id, edited_at DESC, record_id);

--
-- Name: timeline_events timeline_events_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_events
    ADD CONSTRAINT timeline_events_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: timeline_events timeline_events_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_events
    ADD CONSTRAINT timeline_events_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: timeline_events timeline_events_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_events
    ADD CONSTRAINT timeline_events_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: timeline_events timeline_events_reviewed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_events
    ADD CONSTRAINT timeline_events_reviewed_by_user_id_fkey FOREIGN KEY (reviewed_by_user_id) REFERENCES public.users(id);

--
-- Name: timeline_events timeline_events_superseded_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_events
    ADD CONSTRAINT timeline_events_superseded_by_user_id_fkey FOREIGN KEY (superseded_by_user_id) REFERENCES public.users(id);

--
-- Name: timeline_events timeline_events_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_events
    ADD CONSTRAINT timeline_events_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

--
-- Name: timeline_time_conversion_profiles timeline_time_conversion_profiles_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_time_conversion_profiles
    ADD CONSTRAINT timeline_time_conversion_profiles_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: timeline_time_conversion_profiles timeline_time_conversion_profiles_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.timeline_time_conversion_profiles
    ADD CONSTRAINT timeline_time_conversion_profiles_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

-- +goose Down
DROP TABLE IF EXISTS public.timeline_events CASCADE;
DROP TABLE IF EXISTS public.timeline_time_conversion_profiles CASCADE;

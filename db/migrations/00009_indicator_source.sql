-- +goose Up
--
-- Name: indicator_observations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.indicator_observations (
    indicator_observation_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    source_record_id uuid NOT NULL,
    source_field_key text NOT NULL,
    origin_kind text NOT NULL,
    origin_locator text NOT NULL,
    observed_text text NOT NULL,
    parsed_indicator_type text,
    normalized_candidate text,
    resolution_status text NOT NULL,
    resolved_indicator_record_id uuid,
    row_version bigint DEFAULT 1 NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    resolved_by_user_id uuid,
    resolved_at timestamp with time zone,
    resolution_method text,
    CONSTRAINT indicator_observations_resolution_status_check CHECK ((resolution_status = ANY (ARRAY['unresolved'::text, 'resolved'::text, 'dismissed'::text])))
);

--
-- Name: indicator_state_intervals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.indicator_state_intervals (
    indicator_state_interval_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    indicator_record_id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    valid_from timestamp with time zone NOT NULL,
    valid_to timestamp with time zone,
    confidence integer,
    rationale text,
    support_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    assessor text,
    assessed_at timestamp with time zone DEFAULT now() NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT indicator_state_intervals_confidence_check CHECK (((confidence IS NULL) OR ((confidence >= 0) AND (confidence <= 100)))),
    CONSTRAINT indicator_state_intervals_validity_ck CHECK (((valid_to IS NULL) OR (valid_to >= valid_from)))
);

--
-- Name: indicators; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.indicators (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    indicator_type text NOT NULL,
    value_kind text NOT NULL,
    display_value text NOT NULL,
    normalized_value text,
    dedupe_key text NOT NULL,
    defanged_value text,
    hash_algorithm text,
    hash_value text,
    stix_pattern text,
    row_version bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    updated_by_user_id uuid NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    CONSTRAINT indicators_hash_pair_ck CHECK ((((hash_algorithm IS NULL) AND (hash_value IS NULL)) OR ((hash_algorithm IS NOT NULL) AND (hash_value IS NOT NULL)))),
    CONSTRAINT indicators_value_kind_check CHECK ((value_kind = ANY (ARRAY['atomic'::text, 'pattern'::text, 'reference'::text])))
);

--
-- Name: indicator_observations indicator_observations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_observations
    ADD CONSTRAINT indicator_observations_pkey PRIMARY KEY (indicator_observation_id);

--
-- Name: indicator_state_intervals indicator_state_intervals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_pkey PRIMARY KEY (indicator_state_interval_id);

--
-- Name: indicators indicators_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicators
    ADD CONSTRAINT indicators_pkey PRIMARY KEY (record_id);

--
-- Name: indicator_observations_candidate_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX indicator_observations_candidate_lookup_idx ON public.indicator_observations USING btree (incident_id, parsed_indicator_type, normalized_candidate, indicator_observation_id) WHERE (normalized_candidate IS NOT NULL);

--
-- Name: indicator_observations_resolved_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX indicator_observations_resolved_lookup_idx ON public.indicator_observations USING btree (incident_id, resolution_status, resolved_indicator_record_id, created_at);

--
-- Name: indicator_observations_source_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX indicator_observations_source_lookup_idx ON public.indicator_observations USING btree (source_record_id, source_field_key, created_at, indicator_observation_id);

--
-- Name: indicator_state_intervals_indicator_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX indicator_state_intervals_indicator_lookup_idx ON public.indicator_state_intervals USING btree (incident_id, indicator_record_id, valid_from DESC, indicator_state_interval_id DESC);

--
-- Name: indicators_incident_dedupe_unique_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX indicators_incident_dedupe_unique_idx ON public.indicators USING btree (incident_id, indicator_type, dedupe_key) WHERE (deleted_at IS NULL);

--
-- Name: indicators_incident_normalized_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX indicators_incident_normalized_lookup_idx ON public.indicators USING btree (incident_id, indicator_type, normalized_value, record_id) WHERE ((deleted_at IS NULL) AND (normalized_value IS NOT NULL));

--
-- Name: indicator_observations indicator_observations_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_observations
    ADD CONSTRAINT indicator_observations_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: indicator_observations indicator_observations_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_observations
    ADD CONSTRAINT indicator_observations_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: indicator_observations indicator_observations_resolved_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_observations
    ADD CONSTRAINT indicator_observations_resolved_by_user_id_fkey FOREIGN KEY (resolved_by_user_id) REFERENCES public.users(id);

--
-- Name: indicator_observations indicator_observations_resolved_indicator_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_observations
    ADD CONSTRAINT indicator_observations_resolved_indicator_record_id_fkey FOREIGN KEY (resolved_indicator_record_id) REFERENCES public.indicators(record_id) ON DELETE SET NULL;

--
-- Name: indicator_observations indicator_observations_source_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_observations
    ADD CONSTRAINT indicator_observations_source_record_envelope_fkey FOREIGN KEY (incident_id, source_record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: indicator_state_intervals indicator_state_intervals_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: indicator_state_intervals indicator_state_intervals_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: indicator_state_intervals indicator_state_intervals_indicator_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_indicator_record_id_fkey FOREIGN KEY (indicator_record_id) REFERENCES public.indicators(record_id) ON DELETE CASCADE;

--
-- Name: indicators indicators_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicators
    ADD CONSTRAINT indicators_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: indicators indicators_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicators
    ADD CONSTRAINT indicators_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id);

--
-- Name: indicators indicators_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicators
    ADD CONSTRAINT indicators_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: indicators indicators_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicators
    ADD CONSTRAINT indicators_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: indicators indicators_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.indicators
    ADD CONSTRAINT indicators_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

-- +goose Down
DROP TABLE IF EXISTS public.indicators CASCADE;
DROP TABLE IF EXISTS public.indicator_state_intervals CASCADE;
DROP TABLE IF EXISTS public.indicator_observations CASCADE;

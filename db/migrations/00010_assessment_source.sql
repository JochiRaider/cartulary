-- +goose Up
--
-- Name: assessments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assessments (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    subject_record_id uuid NOT NULL,
    subject_type text NOT NULL,
    assessment_state text NOT NULL,
    confidence_score integer,
    assessor_user_id uuid NOT NULL,
    assessed_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    rationale text NOT NULL,
    CONSTRAINT assessments_assessment_state_ck CHECK ((assessment_state = ANY (ARRAY['unknown'::text, 'suspected'::text, 'confirmed'::text, 'disproven'::text, 'cleared'::text]))),
    CONSTRAINT assessments_confidence_score_ck CHECK (((confidence_score IS NULL) OR ((confidence_score >= 0) AND (confidence_score <= 100)))),
    CONSTRAINT assessments_subject_type_ck CHECK ((subject_type = ANY (ARRAY['host'::text, 'identity'::text])))
);

--
-- Name: assessments compromise_assessments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assessments
    ADD CONSTRAINT compromise_assessments_pkey PRIMARY KEY (record_id);

--
-- Name: assessments_active_subject_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX assessments_active_subject_lookup_idx ON public.assessments USING btree (incident_id, subject_type, subject_record_id, assessed_at DESC, record_id) WHERE (deleted_at IS NULL);

--
-- Name: assessments assessments_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assessments
    ADD CONSTRAINT assessments_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: assessments compromise_assessments_assessed_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assessments
    ADD CONSTRAINT compromise_assessments_assessed_by_user_id_fkey FOREIGN KEY (assessor_user_id) REFERENCES public.users(id);

--
-- Name: assessments compromise_assessments_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assessments
    ADD CONSTRAINT compromise_assessments_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id);

--
-- Name: assessments compromise_assessments_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assessments
    ADD CONSTRAINT compromise_assessments_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.assessments CASCADE;

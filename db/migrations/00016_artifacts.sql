-- +goose Up
--
-- Name: public.cartulary_confidence_band(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.cartulary_confidence_band(confidence_score integer) RETURNS text
    LANGUAGE sql IMMUTABLE
    SET search_path = pg_catalog, public
    AS $$
    SELECT CASE
        WHEN confidence_score IS NULL THEN 'unset'
        WHEN confidence_score BETWEEN 0 AND 39 THEN 'low'
        WHEN confidence_score BETWEEN 40 AND 69 THEN 'medium'
        WHEN confidence_score BETWEEN 70 AND 100 THEN 'high'
        ELSE NULL
    END
$$;

REVOKE ALL ON FUNCTION public.cartulary_confidence_band(integer) FROM PUBLIC;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: artifact_findings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.artifact_findings (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    kind text NOT NULL,
    statement text NOT NULL,
    state text NOT NULL,
    confidence_score integer,
    owner_user_id uuid NOT NULL,
    closed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT artifact_findings_confidence_score_check CHECK (((confidence_score IS NULL) OR ((confidence_score >= 0) AND (confidence_score <= 100)))),
    CONSTRAINT artifact_findings_kind_check CHECK ((kind = ANY (ARRAY['finding'::text, 'hypothesis'::text]))),
    CONSTRAINT artifact_findings_state_check CHECK ((state = ANY (ARRAY['open'::text, 'closed'::text])))
);

--
-- Name: artifact_forensic_keywords; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.artifact_forensic_keywords (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    keyword_id text NOT NULL,
    pattern text NOT NULL,
    reason text NOT NULL,
    match_mode text NOT NULL,
    case_sensitive boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT artifact_forensic_keywords_match_mode_check CHECK ((match_mode = ANY (ARRAY['literal'::text, 'regex'::text])))
);

--
-- Name: artifact_investigative_queries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.artifact_investigative_queries (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    query_id text NOT NULL,
    platform text NOT NULL,
    purpose text NOT NULL,
    query_text text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: artifacts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.artifacts (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    artifact_type text NOT NULL,
    title text,
    body text,
    timestamp_utc timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    comm_id text,
    comm_type text,
    audience text,
    channel_or_meeting text,
    summary text,
    next_report_at timestamp with time zone,
    privilege_tag text,
    handoff_id text,
    outgoing_owner_user_id uuid,
    incoming_owner_user_id uuid,
    current_state_summary text,
    next_checks text,
    acknowledged_at timestamp with time zone,
    status_review_id text,
    review_owner_user_id uuid,
    active_risks_summary text,
    lesson_id text,
    owner_user_id uuid,
    closure_state text,
    created_by_user_id uuid,
    CONSTRAINT artifacts_artifact_type_check CHECK ((artifact_type = ANY (ARRAY['note'::text, 'comm_log'::text, 'handoff'::text, 'status_review'::text, 'lesson'::text, 'finding'::text, 'investigative_query'::text, 'forensic_keyword'::text])))
);

--
-- Name: handoff_risk_refs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.handoff_risk_refs (
    risk_ref_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    handoff_record_id uuid NOT NULL,
    risk_ref_text text NOT NULL,
    normalized_risk_ref_text text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    CONSTRAINT handoff_risk_refs_delete_state_ck CHECK ((((deleted_at IS NULL) AND (deleted_by_user_id IS NULL)) OR ((deleted_at IS NOT NULL) AND (deleted_by_user_id IS NOT NULL))))
);

--
-- Name: artifact_findings artifact_findings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_findings
    ADD CONSTRAINT artifact_findings_pkey PRIMARY KEY (record_id);

--
-- Name: artifact_forensic_keywords artifact_forensic_keywords_keyword_id_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_forensic_keywords
    ADD CONSTRAINT artifact_forensic_keywords_keyword_id_unique UNIQUE (incident_id, keyword_id);

--
-- Name: artifact_forensic_keywords artifact_forensic_keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_forensic_keywords
    ADD CONSTRAINT artifact_forensic_keywords_pkey PRIMARY KEY (record_id);

--
-- Name: artifact_investigative_queries artifact_investigative_queries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_investigative_queries
    ADD CONSTRAINT artifact_investigative_queries_pkey PRIMARY KEY (record_id);

--
-- Name: artifact_investigative_queries artifact_investigative_queries_query_id_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_investigative_queries
    ADD CONSTRAINT artifact_investigative_queries_query_id_unique UNIQUE (incident_id, query_id);

--
-- Name: artifacts artifacts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_pkey PRIMARY KEY (record_id);

--
-- Name: handoff_risk_refs handoff_risk_refs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.handoff_risk_refs
    ADD CONSTRAINT handoff_risk_refs_pkey PRIMARY KEY (risk_ref_id);

--
-- Name: artifact_findings_incident_closed_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifact_findings_incident_closed_idx ON public.artifact_findings USING btree (incident_id, closed_at DESC, record_id);

--
-- Name: artifact_findings_incident_confidence_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifact_findings_incident_confidence_idx ON public.artifact_findings USING btree (incident_id, confidence_score, record_id);

--
-- Name: artifact_findings_incident_state_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifact_findings_incident_state_idx ON public.artifact_findings USING btree (incident_id, state, kind, owner_user_id, updated_at DESC, record_id);

--
-- Name: artifact_forensic_keywords_incident_created_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifact_forensic_keywords_incident_created_idx ON public.artifact_forensic_keywords USING btree (incident_id, created_at DESC, record_id);

--
-- Name: artifact_forensic_keywords_incident_mode_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifact_forensic_keywords_incident_mode_idx ON public.artifact_forensic_keywords USING btree (incident_id, match_mode, case_sensitive, created_at DESC, record_id);

--
-- Name: artifact_investigative_queries_incident_created_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifact_investigative_queries_incident_created_idx ON public.artifact_investigative_queries USING btree (incident_id, created_at DESC, record_id);

--
-- Name: artifact_investigative_queries_incident_platform_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifact_investigative_queries_incident_platform_idx ON public.artifact_investigative_queries USING btree (incident_id, platform, created_by_user_id, created_at DESC, record_id);

--
-- Name: artifacts_incident_type_updated_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX artifacts_incident_type_updated_idx ON public.artifacts USING btree (incident_id, artifact_type, updated_at DESC, record_id);

--
-- Name: handoff_risk_refs_active_text_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX handoff_risk_refs_active_text_idx ON public.handoff_risk_refs USING btree (handoff_record_id, normalized_risk_ref_text) WHERE (deleted_at IS NULL);

--
-- Name: artifact_findings artifact_findings_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_findings
    ADD CONSTRAINT artifact_findings_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_findings artifact_findings_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_findings
    ADD CONSTRAINT artifact_findings_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: artifact_findings artifact_findings_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_findings
    ADD CONSTRAINT artifact_findings_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_findings artifact_findings_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_findings
    ADD CONSTRAINT artifact_findings_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.artifacts(record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_forensic_keywords artifact_forensic_keywords_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_forensic_keywords
    ADD CONSTRAINT artifact_forensic_keywords_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_forensic_keywords artifact_forensic_keywords_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_forensic_keywords
    ADD CONSTRAINT artifact_forensic_keywords_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_forensic_keywords artifact_forensic_keywords_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_forensic_keywords
    ADD CONSTRAINT artifact_forensic_keywords_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.artifacts(record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_investigative_queries artifact_investigative_queries_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_investigative_queries
    ADD CONSTRAINT artifact_investigative_queries_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: artifact_investigative_queries artifact_investigative_queries_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_investigative_queries
    ADD CONSTRAINT artifact_investigative_queries_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_investigative_queries artifact_investigative_queries_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_investigative_queries
    ADD CONSTRAINT artifact_investigative_queries_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifact_investigative_queries artifact_investigative_queries_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifact_investigative_queries
    ADD CONSTRAINT artifact_investigative_queries_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.artifacts(record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifacts artifacts_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: artifacts artifacts_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifacts artifacts_incoming_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_incoming_owner_user_id_fkey FOREIGN KEY (incoming_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: artifacts artifacts_outgoing_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_outgoing_owner_user_id_fkey FOREIGN KEY (outgoing_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: artifacts artifacts_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: artifacts artifacts_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: artifacts artifacts_review_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_review_owner_user_id_fkey FOREIGN KEY (review_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: handoff_risk_refs handoff_risk_refs_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.handoff_risk_refs
    ADD CONSTRAINT handoff_risk_refs_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: handoff_risk_refs handoff_risk_refs_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.handoff_risk_refs
    ADD CONSTRAINT handoff_risk_refs_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: handoff_risk_refs handoff_risk_refs_handoff_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.handoff_risk_refs
    ADD CONSTRAINT handoff_risk_refs_handoff_fkey FOREIGN KEY (incident_id, handoff_record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: handoff_risk_refs handoff_risk_refs_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.handoff_risk_refs
    ADD CONSTRAINT handoff_risk_refs_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX artifact_findings_incident_id_record_id_fk_idx ON public.artifact_findings (incident_id, record_id);
CREATE INDEX artifact_findings_owner_user_id_fk_idx ON public.artifact_findings (owner_user_id);
CREATE INDEX artifact_forensic_keywords_incident_id_record_id_fk_idx ON public.artifact_forensic_keywords (incident_id, record_id);
CREATE INDEX artifact_investigative_queries_created_by_user_id_fk_idx ON public.artifact_investigative_queries (created_by_user_id);
CREATE INDEX artifact_investigative_queries_incident_id_record_id_fk_idx ON public.artifact_investigative_queries (incident_id, record_id);
CREATE INDEX artifacts_created_by_user_id_fk_idx ON public.artifacts (created_by_user_id);
CREATE INDEX artifacts_incident_id_record_id_fk_idx ON public.artifacts (incident_id, record_id);
CREATE INDEX artifacts_incoming_owner_user_id_fk_idx ON public.artifacts (incoming_owner_user_id);
CREATE INDEX artifacts_outgoing_owner_user_id_fk_idx ON public.artifacts (outgoing_owner_user_id);
CREATE INDEX artifacts_owner_user_id_fk_idx ON public.artifacts (owner_user_id);
CREATE INDEX artifacts_review_owner_user_id_fk_idx ON public.artifacts (review_owner_user_id);
CREATE INDEX handoff_risk_refs_created_by_user_id_fk_idx ON public.handoff_risk_refs (created_by_user_id);
CREATE INDEX handoff_risk_refs_deleted_by_user_id_fk_idx ON public.handoff_risk_refs (deleted_by_user_id);
CREATE INDEX handoff_risk_refs_incident_id_handoff_record_id_fk_idx ON public.handoff_risk_refs (incident_id, handoff_record_id);
CREATE INDEX handoff_risk_refs_incident_id_fk_idx ON public.handoff_risk_refs (incident_id);

-- +goose Down
DROP FUNCTION public.cartulary_confidence_band(integer);
DROP TABLE public.artifact_findings ;
DROP TABLE public.artifact_investigative_queries ;
DROP TABLE public.artifact_forensic_keywords ;
DROP TABLE public.artifacts ;
DROP TABLE public.handoff_risk_refs ;

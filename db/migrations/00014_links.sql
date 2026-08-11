-- +goose Up
--
-- Name: record_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.record_links (
    record_link_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    src_record_id uuid NOT NULL,
    dst_record_id uuid NOT NULL,
    link_type text NOT NULL,
    provenance text NOT NULL,
    confidence integer,
    owner_user_id uuid NOT NULL,
    decided_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    created_by_user_id uuid NOT NULL,
    field_key text,
    CONSTRAINT record_links_confidence_check CHECK (((confidence IS NULL) OR ((confidence >= 0) AND (confidence <= 100)))),
    CONSTRAINT record_links_distinct_endpoints_ck CHECK ((src_record_id <> dst_record_id)),
    CONSTRAINT record_links_link_type_check CHECK ((link_type = ANY (ARRAY['observed_on_host'::text, 'observed_as_identity'::text, 'references_indicator'::text, 'attached_evidence'::text, 'references_artifact'::text, 'derived_from'::text, 'merged_into'::text, 'supported_by'::text, 'references_record'::text, 'supersedes'::text]))),
    CONSTRAINT record_links_provenance_check CHECK ((provenance = ANY (ARRAY['manual'::text, 'auto_match'::text, 'import'::text, 'rollback'::text, 'system'::text])))
);

--
-- Name: record_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.record_tags (
    record_tag_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_id uuid NOT NULL,
    tag_name text NOT NULL,
    normalized_tag_name text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid
);

--
-- Name: record_links record_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_links
    ADD CONSTRAINT record_links_pkey PRIMARY KEY (record_link_id);

--
-- Name: record_tags record_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_tags
    ADD CONSTRAINT record_tags_pkey PRIMARY KEY (record_tag_id);

--
-- Name: record_links_active_party_ref_dst_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX record_links_active_party_ref_dst_lookup_idx ON public.record_links USING btree (incident_id, dst_record_id, link_type, field_key) WHERE ((deleted_at IS NULL) AND (field_key = ANY (ARRAY['comm_log.audience_party_ids'::text, 'comm_log.attendee_party_ids'::text])));

--
-- Name: record_links_active_src_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX record_links_active_src_lookup_idx ON public.record_links USING btree (incident_id, src_record_id, link_type, dst_record_id) WHERE (deleted_at IS NULL);

--
-- Name: record_links_active_timeline_supersedes_dst_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX record_links_active_timeline_supersedes_dst_idx ON public.record_links USING btree (incident_id, dst_record_id, link_type) WHERE ((deleted_at IS NULL) AND (link_type = 'supersedes'::text));

--
-- Name: record_links_active_unique_field_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX record_links_active_unique_field_idx ON public.record_links USING btree (incident_id, src_record_id, dst_record_id, link_type, field_key) WHERE ((deleted_at IS NULL) AND (field_key IS NOT NULL));

--
-- Name: record_links_active_unique_no_field_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX record_links_active_unique_no_field_idx ON public.record_links USING btree (incident_id, src_record_id, dst_record_id, link_type) WHERE ((deleted_at IS NULL) AND (field_key IS NULL));

--
-- Name: record_tags_active_record_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX record_tags_active_record_lookup_idx ON public.record_tags USING btree (incident_id, record_id, normalized_tag_name) WHERE (deleted_at IS NULL);

--
-- Name: record_tags_active_unique_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX record_tags_active_unique_idx ON public.record_tags USING btree (incident_id, record_id, normalized_tag_name) WHERE (deleted_at IS NULL);

--
-- Name: record_links record_links_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_links
    ADD CONSTRAINT record_links_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: record_links record_links_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_links
    ADD CONSTRAINT record_links_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: record_links record_links_dst_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_links
    ADD CONSTRAINT record_links_dst_record_envelope_fkey FOREIGN KEY (incident_id, dst_record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_links record_links_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_links
    ADD CONSTRAINT record_links_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_links record_links_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_links
    ADD CONSTRAINT record_links_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: record_links record_links_src_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_links
    ADD CONSTRAINT record_links_src_record_envelope_fkey FOREIGN KEY (incident_id, src_record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_tags record_tags_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_tags
    ADD CONSTRAINT record_tags_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: record_tags record_tags_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_tags
    ADD CONSTRAINT record_tags_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: record_tags record_tags_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_tags
    ADD CONSTRAINT record_tags_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: record_tags record_tags_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.record_tags
    ADD CONSTRAINT record_tags_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Links-owned projection read contracts. These views expose active link and tag
-- semantics without making projection providers depend on source table details.
--

CREATE VIEW public.active_record_links_v1 AS
SELECT
    rl.record_link_id,
    rl.incident_id,
    rl.src_record_id,
    src.record_type AS src_record_type,
    rl.dst_record_id,
    dst.record_type AS dst_record_type,
    rl.link_type,
    rl.provenance,
    rl.confidence,
    rl.owner_user_id,
    rl.decided_at,
    rl.created_at,
    rl.deleted_at,
    rl.deleted_by_user_id,
    rl.created_by_user_id,
    rl.field_key
  FROM public.record_links rl
  JOIN public.records src
    ON src.incident_id = rl.incident_id
   AND src.record_id = rl.src_record_id
   AND src.deleted_at IS NULL
  JOIN public.records dst
    ON dst.incident_id = rl.incident_id
   AND dst.record_id = rl.dst_record_id
   AND dst.deleted_at IS NULL
 WHERE rl.deleted_at IS NULL;

CREATE VIEW public.active_record_tags_v1 AS
SELECT
    rt.record_tag_id,
    rt.incident_id,
    rt.record_id,
    r.record_type,
    rt.tag_name,
    rt.normalized_tag_name,
    rt.created_by_user_id,
    rt.created_at,
    rt.updated_at,
    rt.deleted_at,
    rt.deleted_by_user_id
  FROM public.record_tags rt
  JOIN public.records r
    ON r.incident_id = rt.incident_id
   AND r.record_id = rt.record_id
   AND r.deleted_at IS NULL
 WHERE rt.deleted_at IS NULL;

CREATE INDEX record_links_created_by_user_id_fk_idx ON public.record_links (created_by_user_id);
CREATE INDEX record_links_deleted_by_user_id_fk_idx ON public.record_links (deleted_by_user_id);
CREATE INDEX record_links_owner_user_id_fk_idx ON public.record_links (owner_user_id);
CREATE INDEX record_tags_created_by_user_id_fk_idx ON public.record_tags (created_by_user_id);
CREATE INDEX record_tags_deleted_by_user_id_fk_idx ON public.record_tags (deleted_by_user_id);

CREATE INDEX record_links_incident_dst_record_fk_idx
    ON public.record_links (incident_id, dst_record_id);
CREATE INDEX record_links_incident_src_record_fk_idx
    ON public.record_links (incident_id, src_record_id);
CREATE INDEX record_tags_incident_record_fk_idx
    ON public.record_tags (incident_id, record_id);

-- +goose Down
DROP VIEW public.active_record_tags_v1;
DROP VIEW public.active_record_links_v1;

DROP TABLE public.record_links ;
DROP TABLE public.record_tags ;

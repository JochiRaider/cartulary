-- +goose Up
--
-- Name: entity_aliases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entity_aliases (
    entity_alias_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_id uuid NOT NULL,
    entity_type text NOT NULL,
    raw_text text NOT NULL,
    normalized_text text NOT NULL,
    classification text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT entity_aliases_classification_check CHECK ((classification = 'suggestion_only'::text)),
    CONSTRAINT entity_aliases_entity_type_check CHECK ((entity_type = ANY (ARRAY['host'::text, 'identity'::text])))
);

--
-- Name: entity_mentions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entity_mentions (
    entity_mention_id uuid DEFAULT gen_random_uuid() NOT NULL,
    source_record_id uuid NOT NULL,
    entity_type text NOT NULL,
    source_field_key text NOT NULL,
    origin_kind text NOT NULL,
    origin_locator text NOT NULL,
    raw_text text NOT NULL,
    normalized_text text NOT NULL,
    resolution_status text NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    ordinal integer NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    resolved_record_id uuid,
    resolved_by_user_id uuid,
    resolved_at timestamp with time zone,
    resolution_method text,
    CONSTRAINT entity_mentions_entity_type_check CHECK ((entity_type = ANY (ARRAY['host'::text, 'identity'::text]))),
    CONSTRAINT entity_mentions_ordinal_check CHECK ((ordinal > 0)),
    CONSTRAINT entity_mentions_resolution_status_check CHECK ((resolution_status = ANY (ARRAY['unresolved'::text, 'resolved'::text, 'dismissed'::text])))
);

--
-- Name: entity_preserved_identifiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entity_preserved_identifiers (
    entity_preserved_identifier_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_id uuid NOT NULL,
    entity_type text NOT NULL,
    identifier_type text NOT NULL,
    raw_value text NOT NULL,
    normalized_value text NOT NULL,
    classification text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT entity_preserved_identifiers_classification_check CHECK ((classification = ANY (ARRAY['exact_match_reuse'::text, 'suggestion_only'::text, 'provenance_only'::text]))),
    CONSTRAINT entity_preserved_identifiers_entity_type_check CHECK ((entity_type = ANY (ARRAY['host'::text, 'identity'::text])))
);

--
-- Name: hosts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.hosts (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    hostname text,
    aad_device_id text,
    fqdn text,
    entity_origin text DEFAULT 'entity_sheet'::text NOT NULL,
    seed_entity_mention_id uuid,
    host_state text NOT NULL,
    merged_into_record_id uuid,
    row_version bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    updated_by_user_id uuid NOT NULL,
    location text,
    os_platform text,
    business_owner text,
    criticality text,
    containment_status text,
    CONSTRAINT hosts_entity_origin_core02_ck CHECK ((entity_origin = ANY (ARRAY['entity_sheet'::text, 'entity_import'::text, 'created_from_mention'::text, 'system_upsert'::text]))),
    CONSTRAINT hosts_host_state_check CHECK ((host_state = ANY (ARRAY['stub'::text, 'canonical'::text, 'merged'::text]))),
    CONSTRAINT hosts_merge_lineage_ck CHECK ((((host_state = ANY (ARRAY['stub'::text, 'canonical'::text])) AND (merged_into_record_id IS NULL)) OR ((host_state = 'merged'::text) AND (merged_into_record_id IS NOT NULL))))
);

--
-- Name: identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.identities (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    upn text,
    email citext,
    sam_account_name text,
    aad_object_id text,
    sid text,
    entity_origin text DEFAULT 'entity_sheet'::text NOT NULL,
    seed_entity_mention_id uuid,
    identity_state text NOT NULL,
    merged_into_record_id uuid,
    row_version bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    updated_by_user_id uuid NOT NULL,
    privilege_level text,
    mfa_state text,
    reset_status text,
    CONSTRAINT identities_entity_origin_core02_ck CHECK ((entity_origin = ANY (ARRAY['entity_sheet'::text, 'entity_import'::text, 'created_from_mention'::text, 'system_upsert'::text]))),
    CONSTRAINT identities_identity_state_check CHECK ((identity_state = ANY (ARRAY['stub'::text, 'canonical'::text, 'merged'::text]))),
    CONSTRAINT identities_merge_lineage_ck CHECK ((((identity_state = ANY (ARRAY['stub'::text, 'canonical'::text])) AND (merged_into_record_id IS NULL)) OR ((identity_state = 'merged'::text) AND (merged_into_record_id IS NOT NULL))))
);

--
-- Name: entity_aliases entity_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_aliases
    ADD CONSTRAINT entity_aliases_pkey PRIMARY KEY (entity_alias_id);

--
-- Name: entity_mentions entity_mentions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_mentions
    ADD CONSTRAINT entity_mentions_pkey PRIMARY KEY (entity_mention_id);

--
-- Name: entity_preserved_identifiers entity_preserved_identifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_preserved_identifiers
    ADD CONSTRAINT entity_preserved_identifiers_pkey PRIMARY KEY (entity_preserved_identifier_id);

--
-- Name: hosts hosts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_pkey PRIMARY KEY (record_id);

--
-- Name: identities identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_pkey PRIMARY KEY (record_id);

--
-- Name: entity_aliases_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX entity_aliases_lookup_idx ON public.entity_aliases USING btree (incident_id, entity_type, normalized_text, record_id) WHERE (deleted_at IS NULL);

--
-- Name: entity_aliases_record_unique_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX entity_aliases_record_unique_idx ON public.entity_aliases USING btree (record_id, entity_type, normalized_text) WHERE (deleted_at IS NULL);

--
-- Name: entity_mentions_source_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX entity_mentions_source_lookup_idx ON public.entity_mentions USING btree (source_record_id, source_field_key, ordinal, entity_mention_id);

--
-- Name: entity_mentions_unresolved_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX entity_mentions_unresolved_lookup_idx ON public.entity_mentions USING btree (source_record_id, resolution_status, entity_type);

--
-- Name: entity_preserved_identifiers_exact_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX entity_preserved_identifiers_exact_lookup_idx ON public.entity_preserved_identifiers USING btree (incident_id, entity_type, identifier_type, normalized_value, record_id) WHERE ((deleted_at IS NULL) AND (classification = 'exact_match_reuse'::text));

--
-- Name: entity_preserved_identifiers_record_unique_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX entity_preserved_identifiers_record_unique_idx ON public.entity_preserved_identifiers USING btree (record_id, entity_type, identifier_type, normalized_value, classification) WHERE (deleted_at IS NULL);

--
-- Name: hosts_incident_aad_device_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX hosts_incident_aad_device_id_idx ON public.hosts USING btree (incident_id, aad_device_id, record_id) WHERE (aad_device_id IS NOT NULL);

--
-- Name: hosts_incident_display_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX hosts_incident_display_name_idx ON public.hosts USING btree (incident_id, display_name, record_id);

--
-- Name: hosts_incident_fqdn_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX hosts_incident_fqdn_idx ON public.hosts USING btree (incident_id, fqdn, record_id) WHERE (fqdn IS NOT NULL);

--
-- Name: hosts_incident_hostname_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX hosts_incident_hostname_idx ON public.hosts USING btree (incident_id, hostname, record_id) WHERE (hostname IS NOT NULL);

--
-- Name: hosts_incident_merged_into_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX hosts_incident_merged_into_idx ON public.hosts USING btree (incident_id, merged_into_record_id, record_id) WHERE (merged_into_record_id IS NOT NULL);

--
-- Name: identities_incident_aad_object_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX identities_incident_aad_object_id_idx ON public.identities USING btree (incident_id, aad_object_id, record_id) WHERE (aad_object_id IS NOT NULL);

--
-- Name: identities_incident_display_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX identities_incident_display_name_idx ON public.identities USING btree (incident_id, display_name, record_id);

--
-- Name: identities_incident_email_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX identities_incident_email_idx ON public.identities USING btree (incident_id, email, record_id) WHERE (email IS NOT NULL);

--
-- Name: identities_incident_merged_into_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX identities_incident_merged_into_idx ON public.identities USING btree (incident_id, merged_into_record_id, record_id) WHERE (merged_into_record_id IS NOT NULL);

--
-- Name: identities_incident_sam_account_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX identities_incident_sam_account_name_idx ON public.identities USING btree (incident_id, sam_account_name, record_id) WHERE (sam_account_name IS NOT NULL);

--
-- Name: identities_incident_sid_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX identities_incident_sid_idx ON public.identities USING btree (incident_id, sid, record_id) WHERE (sid IS NOT NULL);

--
-- Name: identities_incident_upn_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX identities_incident_upn_idx ON public.identities USING btree (incident_id, upn, record_id) WHERE (upn IS NOT NULL);

--
-- Name: entity_aliases entity_aliases_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_aliases
    ADD CONSTRAINT entity_aliases_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: entity_aliases entity_aliases_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_aliases
    ADD CONSTRAINT entity_aliases_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: entity_mentions entity_mentions_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_mentions
    ADD CONSTRAINT entity_mentions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: entity_mentions entity_mentions_resolved_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_mentions
    ADD CONSTRAINT entity_mentions_resolved_by_user_id_fkey FOREIGN KEY (resolved_by_user_id) REFERENCES public.users(id);

--
-- Name: entity_mentions entity_mentions_source_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_mentions
    ADD CONSTRAINT entity_mentions_source_record_envelope_fkey FOREIGN KEY (source_record_id) REFERENCES public.records(record_id) ON DELETE CASCADE;

--
-- Name: entity_preserved_identifiers entity_preserved_identifiers_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_preserved_identifiers
    ADD CONSTRAINT entity_preserved_identifiers_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: entity_preserved_identifiers entity_preserved_identifiers_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entity_preserved_identifiers
    ADD CONSTRAINT entity_preserved_identifiers_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: hosts hosts_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: hosts hosts_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: hosts hosts_merged_into_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_merged_into_record_id_fkey FOREIGN KEY (merged_into_record_id) REFERENCES public.hosts(record_id);

--
-- Name: hosts hosts_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: hosts hosts_seed_entity_mention_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_seed_entity_mention_id_fkey FOREIGN KEY (seed_entity_mention_id) REFERENCES public.entity_mentions(entity_mention_id);

--
-- Name: hosts hosts_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

--
-- Name: identities identities_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: identities identities_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: identities identities_merged_into_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_merged_into_record_id_fkey FOREIGN KEY (merged_into_record_id) REFERENCES public.identities(record_id);

--
-- Name: identities identities_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: identities identities_seed_entity_mention_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_seed_entity_mention_id_fkey FOREIGN KEY (seed_entity_mention_id) REFERENCES public.entity_mentions(entity_mention_id);

--
-- Name: identities identities_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

-- +goose Down
DROP TABLE IF EXISTS public.entity_mentions CASCADE;
DROP TABLE IF EXISTS public.hosts CASCADE;
DROP TABLE IF EXISTS public.identities CASCADE;
DROP TABLE IF EXISTS public.entity_preserved_identifiers CASCADE;
DROP TABLE IF EXISTS public.entity_aliases CASCADE;

-- +goose Up
--
-- Name: report_composition_preview_attempts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.report_composition_preview_attempts (
    preview_attempt_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    composition_id uuid NOT NULL,
    source_kind text NOT NULL,
    draft_version bigint,
    composition_version bigint,
    preview_source_sha256 text NOT NULL,
    composition_sha256 text,
    preview_source_json jsonb NOT NULL,
    snapshot_id text NOT NULL,
    derivation_version text NOT NULL,
    template_id text NOT NULL,
    template_version text NOT NULL,
    redaction_profile_id text NOT NULL,
    redaction_profile_version text NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    render_environment_profile_id text NOT NULL,
    output_kind text NOT NULL,
    output_options jsonb DEFAULT '{}'::jsonb NOT NULL,
    recipient_partition_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    graph_projection_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    render_attempt_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    preview_source_bytes bytea NOT NULL,
    CONSTRAINT report_composition_preview_attem_recipient_partition_refs_check CHECK ((jsonb_typeof(recipient_partition_refs) = 'array'::text)),
    CONSTRAINT report_composition_preview_attem_redaction_profile_sha256_check CHECK ((redaction_profile_sha256 ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT report_composition_preview_attempts_composition_sha256_check CHECK (((composition_sha256 IS NULL) OR (composition_sha256 ~ '^[a-f0-9]{64}$'::text))),
    CONSTRAINT report_composition_preview_attempts_composition_version_check CHECK (((composition_version IS NULL) OR (composition_version > 0))),
    CONSTRAINT report_composition_preview_attempts_draft_version_check CHECK (((draft_version IS NULL) OR (draft_version > 0))),
    CONSTRAINT report_composition_preview_attempts_graph_projection_refs_check CHECK ((jsonb_typeof(graph_projection_refs) = 'array'::text)),
    CONSTRAINT report_composition_preview_attempts_output_kind_check CHECK ((output_kind = ANY (ARRAY['slidev'::text, 'mermaid'::text]))),
    CONSTRAINT report_composition_preview_attempts_output_options_check CHECK ((jsonb_typeof(output_options) = 'object'::text)),
    CONSTRAINT report_composition_preview_attempts_preview_source_json_check CHECK ((jsonb_typeof(preview_source_json) = 'object'::text)),
    CONSTRAINT report_composition_preview_attempts_preview_source_sha256_check CHECK ((preview_source_sha256 ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT report_composition_preview_attempts_source_kind_check CHECK ((source_kind = ANY (ARRAY['draft'::text, 'version'::text]))),
    CONSTRAINT report_composition_preview_source_shape_ck CHECK ((((source_kind = 'draft'::text) AND (draft_version IS NOT NULL) AND (composition_version IS NULL) AND (composition_sha256 IS NULL)) OR ((source_kind = 'version'::text) AND (draft_version IS NULL) AND (composition_version IS NOT NULL) AND (composition_sha256 IS NOT NULL))))
);

--
-- Name: report_composition_release_bindings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.report_composition_release_bindings (
    composition_id uuid NOT NULL,
    composition_version bigint NOT NULL,
    composition_sha256 text NOT NULL,
    release_id uuid NOT NULL,
    release_scope text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT report_composition_release_bindings_composition_sha256_check CHECK ((composition_sha256 ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT report_composition_release_bindings_release_scope_check CHECK ((release_scope = ANY (ARRAY['internal_draft'::text, 'internal_review'::text, 'external_release'::text])))
);

--
-- Name: report_composition_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.report_composition_versions (
    composition_id uuid NOT NULL,
    composition_version bigint NOT NULL,
    composition_sha256 text NOT NULL,
    canonical_composition jsonb NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    canonical_composition_bytes bytea NOT NULL,
    CONSTRAINT report_composition_versions_canonical_bytes_ck CHECK ((octet_length(canonical_composition_bytes) > 0)),
    CONSTRAINT report_composition_versions_canonical_composition_check CHECK ((jsonb_typeof(canonical_composition) = 'object'::text)),
    CONSTRAINT report_composition_versions_composition_sha256_check CHECK ((composition_sha256 ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT report_composition_versions_composition_version_check CHECK ((composition_version > 0))
);

--
-- Name: report_compositions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.report_compositions (
    composition_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    client_txn_id text NOT NULL,
    template_id text NOT NULL,
    template_version text NOT NULL,
    draft_version bigint DEFAULT 1 NOT NULL,
    authored_against_snapshot_id text,
    deck_ops jsonb DEFAULT '[]'::jsonb NOT NULL,
    diagram_decls jsonb DEFAULT '[]'::jsonb NOT NULL,
    authored_texts jsonb DEFAULT '[]'::jsonb NOT NULL,
    latest_composition_version bigint,
    retired_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT report_compositions_authored_texts_array_ck CHECK ((jsonb_typeof(authored_texts) = 'array'::text)),
    CONSTRAINT report_compositions_deck_ops_array_ck CHECK ((jsonb_typeof(deck_ops) = 'array'::text)),
    CONSTRAINT report_compositions_diagram_decls_array_ck CHECK ((jsonb_typeof(diagram_decls) = 'array'::text)),
    CONSTRAINT report_compositions_draft_version_check CHECK ((draft_version > 0)),
    CONSTRAINT report_compositions_latest_composition_version_check CHECK (((latest_composition_version IS NULL) OR (latest_composition_version > 0)))
);

--
-- Name: report_composition_preview_attempts report_composition_preview_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_preview_attempts
    ADD CONSTRAINT report_composition_preview_attempts_pkey PRIMARY KEY (preview_attempt_id);

--
-- Name: report_composition_release_bindings report_composition_release_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_release_bindings
    ADD CONSTRAINT report_composition_release_bindings_pkey PRIMARY KEY (composition_id, composition_version, release_id);

--
-- Name: report_composition_versions report_composition_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_versions
    ADD CONSTRAINT report_composition_versions_pkey PRIMARY KEY (composition_id, composition_version);

--
-- Name: report_compositions report_compositions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_compositions
    ADD CONSTRAINT report_compositions_pkey PRIMARY KEY (composition_id);

--
-- Name: report_composition_preview_attempts_composition_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX report_composition_preview_attempts_composition_lookup_idx ON public.report_composition_preview_attempts USING btree (composition_id, created_at DESC, preview_attempt_id);

--
-- Name: report_composition_release_bindings_release_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX report_composition_release_bindings_release_lookup_idx ON public.report_composition_release_bindings USING btree (release_id, composition_id, composition_version);

--
-- Name: report_composition_versions_digest_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX report_composition_versions_digest_idx ON public.report_composition_versions USING btree (composition_id, composition_version, composition_sha256);

--
-- Name: report_compositions_created_by_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX report_compositions_created_by_lookup_idx ON public.report_compositions USING btree (created_by_user_id, created_at DESC, composition_id);

--
-- Name: report_compositions_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX report_compositions_incident_lookup_idx ON public.report_compositions USING btree (incident_id, retired_at NULLS FIRST, template_id, template_version, composition_id);

--
-- Name: report_composition_preview_attempts report_composition_preview_attempts_composition_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_preview_attempts
    ADD CONSTRAINT report_composition_preview_attempts_composition_id_fkey FOREIGN KEY (composition_id) REFERENCES public.report_compositions(composition_id) ON DELETE CASCADE;

--
-- Name: report_composition_preview_attempts report_composition_preview_attempts_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_preview_attempts
    ADD CONSTRAINT report_composition_preview_attempts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: report_composition_preview_attempts report_composition_preview_attempts_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_preview_attempts
    ADD CONSTRAINT report_composition_preview_attempts_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: report_composition_preview_attempts report_composition_preview_attempts_render_attempt_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_preview_attempts
    ADD CONSTRAINT report_composition_preview_attempts_render_attempt_fk FOREIGN KEY (render_attempt_id) REFERENCES public.jobs(job_id) ON DELETE RESTRICT;

--
-- Name: report_composition_release_bindings report_composition_release_bi_composition_id_composition_v_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_release_bindings
    ADD CONSTRAINT report_composition_release_bi_composition_id_composition_v_fkey FOREIGN KEY (composition_id, composition_version) REFERENCES public.report_composition_versions(composition_id, composition_version) ON DELETE RESTRICT;

--
-- Name: report_composition_release_bindings report_composition_release_bindings_release_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_release_bindings
    ADD CONSTRAINT report_composition_release_bindings_release_fk FOREIGN KEY (release_id) REFERENCES public.reporting_releases(release_id) ON DELETE RESTRICT;

--
-- Name: report_composition_versions report_composition_versions_composition_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_versions
    ADD CONSTRAINT report_composition_versions_composition_id_fkey FOREIGN KEY (composition_id) REFERENCES public.report_compositions(composition_id) ON DELETE CASCADE;

--
-- Name: report_composition_versions report_composition_versions_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_composition_versions
    ADD CONSTRAINT report_composition_versions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: report_compositions report_compositions_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_compositions
    ADD CONSTRAINT report_compositions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: report_compositions report_compositions_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_compositions
    ADD CONSTRAINT report_compositions_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.report_compositions CASCADE;
DROP TABLE IF EXISTS public.report_composition_versions CASCADE;
DROP TABLE IF EXISTS public.report_composition_release_bindings CASCADE;
DROP TABLE IF EXISTS public.report_composition_preview_attempts CASCADE;

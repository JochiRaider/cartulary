-- +goose Up
--
-- Name: incident_bundle_exports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_bundle_exports (
    bundle_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    export_job_id uuid NOT NULL,
    exported_by_user_id uuid NOT NULL,
    exported_at timestamp with time zone NOT NULL,
    manifest_sha256 text NOT NULL,
    reference_pack_mode text NOT NULL,
    history_mode text DEFAULT 'full'::text NOT NULL,
    blob_mode text DEFAULT 'full'::text NOT NULL,
    optional_sections text[] DEFAULT '{}'::text[] NOT NULL,
    required_capabilities text[] DEFAULT '{}'::text[] NOT NULL,
    bundle_sha256 text NOT NULL,
    bundle_byte_size bigint NOT NULL,
    bundle_storage_ref text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT incident_bundle_exports_blob_mode_check CHECK ((blob_mode = 'full'::text)),
    CONSTRAINT incident_bundle_exports_bundle_byte_size_check CHECK ((bundle_byte_size >= 0)),
    CONSTRAINT incident_bundle_exports_bundle_sha256_check CHECK ((bundle_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT incident_bundle_exports_history_mode_check CHECK ((history_mode = 'full'::text)),
    CONSTRAINT incident_bundle_exports_manifest_sha256_check CHECK ((manifest_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT incident_bundle_exports_reference_pack_mode_check CHECK ((reference_pack_mode = ANY (ARRAY['refs_only'::text, 'embedded'::text]))),
    CONSTRAINT incident_bundle_exports_storage_ref_ck CHECK (
        bundle_storage_ref <> ''
        AND bundle_storage_ref !~ '^/'
        AND strpos(bundle_storage_ref, E'\\') = 0
        AND bundle_storage_ref !~ '(^|/)\.\.?(/|$)'
        AND bundle_storage_ref !~ '//'
        AND bundle_storage_ref !~ '/$'
    )
);

--
-- Name: incident_bundle_imported_actors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_bundle_imported_actors (
    imported_actor_descriptor_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    source_actor_id text NOT NULL,
    display_name text,
    email_hint text,
    local_user_id uuid,
    imported_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: incident_bundle_imported_attributions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_bundle_imported_attributions (
    imported_attribution_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    source_table text NOT NULL,
    source_row_id text NOT NULL,
    source_column text NOT NULL,
    source_actor_id text NOT NULL,
    local_user_id uuid NOT NULL,
    imported_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: incident_bundle_job_payloads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_bundle_job_payloads (
    job_id uuid NOT NULL,
    job_kind text NOT NULL,
    actor_user_id uuid NOT NULL,
    incident_id uuid,
    bundle_id uuid,
    uploaded_sha256 text,
    bundle_staging_ref text,
    imported_incident_id uuid,
    manifest_sha256 text,
    failure_reason text,
    request_json jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT incident_bundle_job_payloads_job_kind_check CHECK ((job_kind = ANY (ARRAY['export'::text, 'import'::text]))),
    CONSTRAINT incident_bundle_job_payloads_manifest_sha256_check CHECK (((manifest_sha256 IS NULL) OR (manifest_sha256 ~ '^[0-9a-f]{64}$'::text))),
    CONSTRAINT incident_bundle_job_payloads_uploaded_sha256_check CHECK (((uploaded_sha256 IS NULL) OR (uploaded_sha256 ~ '^[0-9a-f]{64}$'::text))),
    CONSTRAINT incident_bundle_payload_kind_ck CHECK ((((job_kind = 'export'::text) AND (incident_id IS NOT NULL) AND (uploaded_sha256 IS NULL)) OR ((job_kind = 'import'::text) AND (incident_id IS NULL) AND (uploaded_sha256 IS NOT NULL)))),
    CONSTRAINT incident_bundle_job_payloads_staging_ref_ck CHECK (
        bundle_staging_ref IS NULL
        OR (
            bundle_staging_ref <> ''
            AND bundle_staging_ref !~ '^/'
            AND strpos(bundle_staging_ref, E'\\') = 0
            AND bundle_staging_ref !~ '(^|/)\.\.?(/|$)'
            AND bundle_staging_ref !~ '//'
            AND bundle_staging_ref !~ '/$'
        )
    )
);

--
-- Name: incident_bundle_manifest_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_bundle_manifest_files (
    bundle_id uuid NOT NULL,
    path text NOT NULL,
    sha256 text NOT NULL,
    size_bytes bigint NOT NULL,
    required boolean DEFAULT true NOT NULL,
    CONSTRAINT incident_bundle_manifest_files_sha256_check CHECK ((sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT incident_bundle_manifest_files_size_bytes_check CHECK ((size_bytes >= 0))
);

--
-- Name: incident_bundle_exports incident_bundle_exports_export_job_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_exports
    ADD CONSTRAINT incident_bundle_exports_export_job_id_key UNIQUE (export_job_id);

--
-- Name: incident_bundle_exports incident_bundle_exports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_exports
    ADD CONSTRAINT incident_bundle_exports_pkey PRIMARY KEY (bundle_id);

--
-- Name: incident_bundle_imported_actors incident_bundle_imported_actors_incident_id_source_actor_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_actors
    ADD CONSTRAINT incident_bundle_imported_actors_incident_id_source_actor_id_key UNIQUE (incident_id, source_actor_id);

--
-- Name: incident_bundle_imported_actors incident_bundle_imported_actors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_actors
    ADD CONSTRAINT incident_bundle_imported_actors_pkey PRIMARY KEY (imported_actor_descriptor_id);

--
-- Name: incident_bundle_imported_attributions incident_bundle_imported_attr_incident_id_source_table_sour_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_attributions
    ADD CONSTRAINT incident_bundle_imported_attr_incident_id_source_table_sour_key UNIQUE (incident_id, source_table, source_row_id, source_column);

--
-- Name: incident_bundle_imported_attributions incident_bundle_imported_attributions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_attributions
    ADD CONSTRAINT incident_bundle_imported_attributions_pkey PRIMARY KEY (imported_attribution_id);

--
-- Name: incident_bundle_job_payloads incident_bundle_job_payloads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_job_payloads
    ADD CONSTRAINT incident_bundle_job_payloads_pkey PRIMARY KEY (job_id);

--
-- Name: incident_bundle_manifest_files incident_bundle_manifest_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_manifest_files
    ADD CONSTRAINT incident_bundle_manifest_files_pkey PRIMARY KEY (bundle_id, path);

--
-- Name: incident_bundle_exports_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX incident_bundle_exports_incident_lookup_idx ON public.incident_bundle_exports USING btree (incident_id, exported_at DESC, bundle_id);

--
-- Name: incident_bundle_imported_attributions_actor_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX incident_bundle_imported_attributions_actor_idx ON public.incident_bundle_imported_attributions USING btree (incident_id, source_actor_id);

--
-- Name: incident_bundle_job_payloads_actor_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX incident_bundle_job_payloads_actor_idx ON public.incident_bundle_job_payloads USING btree (actor_user_id, created_at DESC);

--
-- Name: incident_bundle_exports incident_bundle_exports_export_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_exports
    ADD CONSTRAINT incident_bundle_exports_export_job_id_fkey FOREIGN KEY (export_job_id) REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: incident_bundle_exports incident_bundle_exports_exported_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_exports
    ADD CONSTRAINT incident_bundle_exports_exported_by_user_id_fkey FOREIGN KEY (exported_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: incident_bundle_exports incident_bundle_exports_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_exports
    ADD CONSTRAINT incident_bundle_exports_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: incident_bundle_imported_actors incident_bundle_imported_actors_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_actors
    ADD CONSTRAINT incident_bundle_imported_actors_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: incident_bundle_imported_actors incident_bundle_imported_actors_local_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_actors
    ADD CONSTRAINT incident_bundle_imported_actors_local_user_id_fkey FOREIGN KEY (local_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: incident_bundle_imported_attributions incident_bundle_imported_attributions_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_attributions
    ADD CONSTRAINT incident_bundle_imported_attributions_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: incident_bundle_imported_attributions incident_bundle_imported_attributions_local_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_imported_attributions
    ADD CONSTRAINT incident_bundle_imported_attributions_local_user_id_fkey FOREIGN KEY (local_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: incident_bundle_job_payloads incident_bundle_job_payloads_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_job_payloads
    ADD CONSTRAINT incident_bundle_job_payloads_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: incident_bundle_job_payloads incident_bundle_job_payloads_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_job_payloads
    ADD CONSTRAINT incident_bundle_job_payloads_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: incident_bundle_job_payloads incident_bundle_job_payloads_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_job_payloads
    ADD CONSTRAINT incident_bundle_job_payloads_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: incident_bundle_manifest_files incident_bundle_manifest_files_bundle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_bundle_manifest_files
    ADD CONSTRAINT incident_bundle_manifest_files_bundle_id_fkey FOREIGN KEY (bundle_id) REFERENCES public.incident_bundle_exports(bundle_id) ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX incident_bundle_exports_exported_by_user_id_fk_idx ON public.incident_bundle_exports (exported_by_user_id);
CREATE INDEX incident_bundle_imported_actors_local_user_id_fk_idx ON public.incident_bundle_imported_actors (local_user_id);
CREATE INDEX incident_bundle_imported_attributions_local_user_id_fk_idx ON public.incident_bundle_imported_attributions (local_user_id);
CREATE INDEX incident_bundle_job_payloads_incident_id_fk_idx ON public.incident_bundle_job_payloads (incident_id);

-- +goose Down
DROP TABLE public.incident_bundle_manifest_files ;
DROP TABLE public.incident_bundle_exports ;
DROP TABLE public.incident_bundle_job_payloads ;
DROP TABLE public.incident_bundle_imported_actors ;
DROP TABLE public.incident_bundle_imported_attributions ;

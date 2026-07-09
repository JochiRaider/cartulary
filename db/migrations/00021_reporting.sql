-- +goose Up
--
-- Name: reporting_job_payloads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reporting_job_payloads (
    job_id uuid NOT NULL,
    job_kind text NOT NULL,
    incident_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    request_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_job_payloads_job_kind_check CHECK ((job_kind = ANY (ARRAY['snapshot_create'::text, 'release_create'::text])))
);

--
-- Name: reporting_release_approvals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reporting_release_approvals (
    approval_id uuid DEFAULT gen_random_uuid() NOT NULL,
    release_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    approval_role text NOT NULL,
    reason text,
    approval_tuple_json jsonb NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    output_sha256 text NOT NULL,
    redaction_manifest_sha256 text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_release_approvals_approval_role_check CHECK ((approval_role = ANY (ARRAY['reviewer'::text, 'admin'::text])))
);

--
-- Name: reporting_releases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reporting_releases (
    release_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    snapshot_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    client_txn_id text NOT NULL,
    release_scope text NOT NULL,
    release_state text NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    source_change_set_high_watermark text NOT NULL,
    derivation_version text NOT NULL,
    export_model_sha256 text NOT NULL,
    template_id text NOT NULL,
    template_version text NOT NULL,
    redaction_profile_id text NOT NULL,
    redaction_profile_version text NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    output_kind text NOT NULL,
    output_media_type text,
    output_sha256 text,
    redaction_manifest_sha256 text,
    redaction_manifest_json jsonb,
    create_job_id uuid NOT NULL,
    render_failed_reason_code text,
    approved_at timestamp with time zone,
    published_at timestamp with time zone,
    invalidated_at timestamp with time zone,
    invalidation_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    recipient_partition_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    output_options jsonb DEFAULT '{}'::jsonb NOT NULL,
    graph_projection_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    composition_id uuid,
    composition_version text,
    composition_sha256 text,
    render_admitted_at timestamp with time zone NOT NULL,
    CONSTRAINT reporting_releases_composition_tuple_ck CHECK ((((composition_id IS NULL) AND (composition_version IS NULL) AND (composition_sha256 IS NULL)) OR ((composition_id IS NOT NULL) AND (composition_version ~ '^v[1-9][0-9]*$'::text) AND (composition_sha256 ~ '^[a-f0-9]{64}$'::text)))),
    CONSTRAINT reporting_releases_graph_projection_refs_ck CHECK ((jsonb_typeof(graph_projection_refs) = 'array'::text)),
    CONSTRAINT reporting_releases_output_kind_check CHECK ((output_kind = ANY (ARRAY['slidev'::text, 'mermaid'::text]))),
    CONSTRAINT reporting_releases_output_options_ck CHECK ((jsonb_typeof(output_options) = 'object'::text)),
    CONSTRAINT reporting_releases_recipient_partition_refs_ck CHECK (((jsonb_typeof(recipient_partition_refs) = 'array'::text) AND ((release_scope = 'external_release'::text) OR (recipient_partition_refs = '[]'::jsonb)))),
    CONSTRAINT reporting_releases_release_scope_check CHECK ((release_scope = ANY (ARRAY['internal_draft'::text, 'internal_review'::text, 'external_release'::text]))),
    CONSTRAINT reporting_releases_release_state_check CHECK ((release_state = ANY (ARRAY['pending_approval'::text, 'approved'::text, 'published'::text, 'invalidated'::text, 'render_failed'::text])))
);

--
-- Name: reporting_render_bundle_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reporting_render_bundle_files (
    release_id uuid NOT NULL,
    bundle_path text NOT NULL,
    role text NOT NULL,
    media_type text NOT NULL,
    file_sha256 text NOT NULL,
    size_bytes bigint NOT NULL,
    storage_kind text NOT NULL,
    object_ref text,
    inline_bytes bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_render_bundle_files_bundle_path_check CHECK (((bundle_path <> ''::text) AND (bundle_path !~ '(^/|(^|/)\.\.?(/|$))'::text))),
    CONSTRAINT reporting_render_bundle_files_check CHECK ((((storage_kind = 'database_inline'::text) AND (inline_bytes IS NOT NULL) AND (object_ref IS NULL)) OR ((storage_kind = 'object_store'::text) AND (inline_bytes IS NULL) AND (object_ref IS NOT NULL) AND (object_ref <> ''::text)))),
    CONSTRAINT reporting_render_bundle_files_file_sha256_check CHECK ((file_sha256 ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT reporting_render_bundle_files_media_type_check CHECK ((media_type <> ''::text)),
    CONSTRAINT reporting_render_bundle_files_role_check CHECK ((role <> ''::text)),
    CONSTRAINT reporting_render_bundle_files_size_bytes_check CHECK ((size_bytes >= 0)),
    CONSTRAINT reporting_render_bundle_files_storage_kind_check CHECK ((storage_kind = ANY (ARRAY['database_inline'::text, 'object_store'::text])))
);

--
-- Name: reporting_render_bundles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reporting_render_bundles (
    release_id uuid NOT NULL,
    bundle_manifest_sha256 text NOT NULL,
    bundle_manifest_json jsonb NOT NULL,
    primary_bundle_path text NOT NULL,
    primary_media_type text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_render_bundles_bundle_manifest_json_check CHECK (((jsonb_typeof(bundle_manifest_json) = 'object'::text) AND ((bundle_manifest_json ->> 'schema_id'::text) = 'cartulary.render_bundle_manifest.v1'::text))),
    CONSTRAINT reporting_render_bundles_bundle_manifest_sha256_check CHECK ((bundle_manifest_sha256 ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT reporting_render_bundles_primary_bundle_path_check CHECK ((primary_bundle_path <> ''::text)),
    CONSTRAINT reporting_render_bundles_primary_media_type_check CHECK ((primary_media_type <> ''::text))
);

--
-- Name: reporting_snapshots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reporting_snapshots (
    snapshot_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    client_txn_id text NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    source_change_set_high_watermark text NOT NULL,
    source_boundary_json jsonb NOT NULL,
    derivation_version text NOT NULL,
    export_model_sha256 text NOT NULL,
    export_model_json jsonb NOT NULL,
    create_job_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_snapshots_source_boundary_json_check CHECK ((jsonb_typeof(source_boundary_json) = 'object'::text)),
    CONSTRAINT reporting_snapshots_source_boundary_json_ck CHECK ((jsonb_typeof(source_boundary_json) = 'object'::text))
);

--
-- Name: reporting_job_payloads reporting_job_payloads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_job_payloads
    ADD CONSTRAINT reporting_job_payloads_pkey PRIMARY KEY (job_id);

--
-- Name: reporting_release_approvals reporting_release_approvals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_release_approvals
    ADD CONSTRAINT reporting_release_approvals_pkey PRIMARY KEY (approval_id);

--
-- Name: reporting_release_approvals reporting_release_approvals_release_id_actor_user_id_approv_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_release_approvals
    ADD CONSTRAINT reporting_release_approvals_release_id_actor_user_id_approv_key UNIQUE (release_id, actor_user_id, approval_role);

--
-- Name: reporting_releases reporting_releases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_releases
    ADD CONSTRAINT reporting_releases_pkey PRIMARY KEY (release_id);

--
-- Name: reporting_render_bundle_files reporting_render_bundle_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_render_bundle_files
    ADD CONSTRAINT reporting_render_bundle_files_pkey PRIMARY KEY (release_id, bundle_path);

--
-- Name: reporting_render_bundles reporting_render_bundles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_render_bundles
    ADD CONSTRAINT reporting_render_bundles_pkey PRIMARY KEY (release_id);

--
-- Name: reporting_snapshots reporting_snapshots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_snapshots
    ADD CONSTRAINT reporting_snapshots_pkey PRIMARY KEY (snapshot_id);

--
-- Name: reporting_job_payloads_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_job_payloads_incident_lookup_idx ON public.reporting_job_payloads USING btree (incident_id, created_at DESC, job_id);

--
-- Name: reporting_release_approvals_release_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_release_approvals_release_lookup_idx ON public.reporting_release_approvals USING btree (release_id, approval_role, created_at, approval_id);

--
-- Name: reporting_releases_created_by_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_releases_created_by_lookup_idx ON public.reporting_releases USING btree (created_by_user_id, created_at DESC, release_id);

--
-- Name: reporting_releases_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_releases_incident_lookup_idx ON public.reporting_releases USING btree (incident_id, created_at DESC, release_id);

--
-- Name: reporting_releases_snapshot_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_releases_snapshot_lookup_idx ON public.reporting_releases USING btree (snapshot_id, created_at DESC, release_id);

--
-- Name: reporting_render_bundle_files_release_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_render_bundle_files_release_idx ON public.reporting_render_bundle_files USING btree (release_id, role, bundle_path);

--
-- Name: reporting_snapshots_created_by_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_snapshots_created_by_lookup_idx ON public.reporting_snapshots USING btree (created_by_user_id, created_at DESC, snapshot_id);

--
-- Name: reporting_snapshots_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reporting_snapshots_incident_lookup_idx ON public.reporting_snapshots USING btree (incident_id, created_at DESC, snapshot_id);

--
-- Name: reporting_job_payloads reporting_job_payloads_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_job_payloads
    ADD CONSTRAINT reporting_job_payloads_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id);

--
-- Name: reporting_job_payloads reporting_job_payloads_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_job_payloads
    ADD CONSTRAINT reporting_job_payloads_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: reporting_job_payloads reporting_job_payloads_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_job_payloads
    ADD CONSTRAINT reporting_job_payloads_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(job_id) ON DELETE CASCADE;

--
-- Name: reporting_release_approvals reporting_release_approvals_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_release_approvals
    ADD CONSTRAINT reporting_release_approvals_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id);

--
-- Name: reporting_release_approvals reporting_release_approvals_release_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_release_approvals
    ADD CONSTRAINT reporting_release_approvals_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.reporting_releases(release_id) ON DELETE CASCADE;

--
-- Name: reporting_releases reporting_releases_create_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_releases
    ADD CONSTRAINT reporting_releases_create_job_id_fkey FOREIGN KEY (create_job_id) REFERENCES public.jobs(job_id);

--
-- Name: reporting_releases reporting_releases_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_releases
    ADD CONSTRAINT reporting_releases_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: reporting_releases reporting_releases_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_releases
    ADD CONSTRAINT reporting_releases_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: reporting_releases reporting_releases_snapshot_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_releases
    ADD CONSTRAINT reporting_releases_snapshot_id_fkey FOREIGN KEY (snapshot_id) REFERENCES public.reporting_snapshots(snapshot_id) ON DELETE CASCADE;

--
-- Name: reporting_render_bundle_files reporting_render_bundle_files_release_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_render_bundle_files
    ADD CONSTRAINT reporting_render_bundle_files_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.reporting_render_bundles(release_id) ON DELETE CASCADE;

--
-- Name: reporting_render_bundles reporting_render_bundles_release_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_render_bundles
    ADD CONSTRAINT reporting_render_bundles_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.reporting_releases(release_id) ON DELETE CASCADE;

--
-- Name: reporting_snapshots reporting_snapshots_create_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_snapshots
    ADD CONSTRAINT reporting_snapshots_create_job_id_fkey FOREIGN KEY (create_job_id) REFERENCES public.jobs(job_id);

--
-- Name: reporting_snapshots reporting_snapshots_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_snapshots
    ADD CONSTRAINT reporting_snapshots_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: reporting_snapshots reporting_snapshots_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reporting_snapshots
    ADD CONSTRAINT reporting_snapshots_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.reporting_snapshots CASCADE;
DROP TABLE IF EXISTS public.reporting_releases CASCADE;
DROP TABLE IF EXISTS public.reporting_release_approvals CASCADE;
DROP TABLE IF EXISTS public.reporting_job_payloads CASCADE;
DROP TABLE IF EXISTS public.reporting_render_bundles CASCADE;
DROP TABLE IF EXISTS public.reporting_render_bundle_files CASCADE;

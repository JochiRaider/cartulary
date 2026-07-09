-- +goose Up
--
-- Name: import_apply_journal; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.import_apply_journal (
    import_apply_journal_id uuid DEFAULT gen_random_uuid() NOT NULL,
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    mapping_fingerprint text NOT NULL,
    source_row_ref integer NOT NULL,
    target_view_schema_id text NOT NULL,
    owner_create_facade text NOT NULL,
    record_id uuid NOT NULL,
    row_version bigint NOT NULL,
    change_set_id uuid NOT NULL,
    change_set_mutation_ref text NOT NULL,
    owner_result_code text NOT NULL,
    created_or_reused text NOT NULL,
    owner_response_json jsonb NOT NULL,
    row_refresh_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_apply_journal_row_version_check CHECK ((row_version >= 1)),
    CONSTRAINT import_apply_journal_source_row_ref_check CHECK ((source_row_ref > 0))
);

--
-- Name: import_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.import_sessions (
    import_session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    client_txn_id text NOT NULL,
    assistant_profile text NOT NULL,
    source_file_kind text NOT NULL,
    original_filename text NOT NULL,
    source_content_sha256 text NOT NULL,
    source_media_type text NOT NULL,
    source_byte_size bigint NOT NULL,
    parser_profile_id text NOT NULL,
    parser_version text NOT NULL,
    session_status text NOT NULL,
    discovery_job_id uuid,
    apply_job_id uuid,
    selected_unit_ids uuid[] DEFAULT '{}'::uuid[] NOT NULL,
    blocking_diagnostics_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    nonblocking_warning_codes text[] DEFAULT '{}'::text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_sessions_session_status_check CHECK ((session_status = ANY (ARRAY['created'::text, 'discovered'::text, 'mapped'::text, 'ready_to_apply'::text, 'applying'::text, 'applied'::text, 'partially_applied'::text, 'failed'::text, 'canceled'::text]))),
    CONSTRAINT import_sessions_source_byte_size_check CHECK ((source_byte_size >= 0)),
    CONSTRAINT import_sessions_source_file_kind_check CHECK ((source_file_kind = ANY (ARRAY['csv'::text, 'xlsx'::text])))
);

--
-- Name: import_units; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.import_units (
    import_unit_id uuid DEFAULT gen_random_uuid() NOT NULL,
    import_session_id uuid NOT NULL,
    unit_status text NOT NULL,
    locator_kind text NOT NULL,
    locator text NOT NULL,
    source_rect_a1 text NOT NULL,
    header_row_ref integer NOT NULL,
    data_start_row_ref integer NOT NULL,
    inferred_row_count integer NOT NULL,
    inferred_column_count integer NOT NULL,
    warning_codes text[] DEFAULT '{}'::text[] NOT NULL,
    mapping_fingerprint text,
    approved_mapping_json jsonb,
    columns_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    source_rows_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    preview_rows_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_units_data_start_row_ref_check CHECK ((data_start_row_ref > 0)),
    CONSTRAINT import_units_header_row_ref_check CHECK ((header_row_ref > 0)),
    CONSTRAINT import_units_inferred_column_count_check CHECK ((inferred_column_count >= 0)),
    CONSTRAINT import_units_inferred_row_count_check CHECK ((inferred_row_count >= 0)),
    CONSTRAINT import_units_unit_status_check CHECK ((unit_status = ANY (ARRAY['discovered'::text, 'selected'::text, 'mapped'::text, 'ready'::text, 'skipped'::text, 'applying'::text, 'applied'::text, 'rejected'::text, 'failed'::text])))
);

--
-- Name: import_apply_journal import_apply_journal_import_unit_id_mapping_fingerprint_sou_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_apply_journal
    ADD CONSTRAINT import_apply_journal_import_unit_id_mapping_fingerprint_sou_key UNIQUE (import_unit_id, mapping_fingerprint, source_row_ref);

--
-- Name: import_apply_journal import_apply_journal_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_apply_journal
    ADD CONSTRAINT import_apply_journal_pkey PRIMARY KEY (import_apply_journal_id);

--
-- Name: import_sessions import_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_sessions
    ADD CONSTRAINT import_sessions_pkey PRIMARY KEY (import_session_id);

--
-- Name: import_units import_units_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_units
    ADD CONSTRAINT import_units_pkey PRIMARY KEY (import_unit_id);

--
-- Name: import_apply_journal_record_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX import_apply_journal_record_idx ON public.import_apply_journal USING btree (record_id, created_at DESC);

--
-- Name: import_apply_journal_session_unit_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX import_apply_journal_session_unit_idx ON public.import_apply_journal USING btree (import_session_id, import_unit_id, source_row_ref);

--
-- Name: import_sessions_created_by_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX import_sessions_created_by_lookup_idx ON public.import_sessions USING btree (created_by_user_id, created_at DESC, import_session_id);

--
-- Name: import_sessions_incident_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX import_sessions_incident_lookup_idx ON public.import_sessions USING btree (incident_id, created_at DESC, import_session_id);

--
-- Name: import_units_session_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX import_units_session_lookup_idx ON public.import_units USING btree (import_session_id, created_at, import_unit_id);

--
-- Name: import_apply_journal import_apply_journal_change_set_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_apply_journal
    ADD CONSTRAINT import_apply_journal_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON DELETE CASCADE;

--
-- Name: import_apply_journal import_apply_journal_import_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_apply_journal
    ADD CONSTRAINT import_apply_journal_import_session_id_fkey FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE;

--
-- Name: import_apply_journal import_apply_journal_import_unit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_apply_journal
    ADD CONSTRAINT import_apply_journal_import_unit_id_fkey FOREIGN KEY (import_unit_id) REFERENCES public.import_units(import_unit_id) ON DELETE CASCADE;

--
-- Name: import_apply_journal import_apply_journal_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_apply_journal
    ADD CONSTRAINT import_apply_journal_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.records(record_id) ON DELETE CASCADE;

--
-- Name: import_sessions import_sessions_apply_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_sessions
    ADD CONSTRAINT import_sessions_apply_job_id_fkey FOREIGN KEY (apply_job_id) REFERENCES public.jobs(job_id);

--
-- Name: import_sessions import_sessions_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_sessions
    ADD CONSTRAINT import_sessions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: import_sessions import_sessions_discovery_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_sessions
    ADD CONSTRAINT import_sessions_discovery_job_id_fkey FOREIGN KEY (discovery_job_id) REFERENCES public.jobs(job_id);

--
-- Name: import_sessions import_sessions_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_sessions
    ADD CONSTRAINT import_sessions_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: import_units import_units_import_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_units
    ADD CONSTRAINT import_units_import_session_id_fkey FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.import_sessions CASCADE;
DROP TABLE IF EXISTS public.import_units CASCADE;
DROP TABLE IF EXISTS public.import_apply_journal CASCADE;

-- +goose Up
--
-- Name: evidence; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.evidence (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    title text,
    lifecycle_state text DEFAULT 'requested'::text NOT NULL,
    requested_at timestamp with time zone,
    received_at timestamp with time zone,
    storage_ref text,
    blob_hash text,
    collector_party_text text,
    collector_party_id uuid,
    source_party_text text,
    source_party_id uuid,
    upload_state text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    object_blob_id uuid
);

--
-- Name: evidence_access_handles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.evidence_access_handles (
    handle_token text NOT NULL,
    incident_id uuid NOT NULL,
    record_id uuid NOT NULL,
    object_blob_id uuid NOT NULL,
    issued_by_user_id uuid NOT NULL,
    issuing_session_id uuid NOT NULL,
    handle_kind text NOT NULL,
    media_class text NOT NULL,
    preview_kind text,
    disposition text NOT NULL,
    filename text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL,
    sha256 text,
    evidence_lifecycle_state text NOT NULL,
    upload_state text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    record_row_version bigint NOT NULL,
    CONSTRAINT evidence_access_handles_disposition_check CHECK ((disposition = ANY (ARRAY['inline'::text, 'attachment'::text]))),
    CONSTRAINT evidence_access_handles_handle_kind_check CHECK ((handle_kind = ANY (ARRAY['preview'::text, 'download'::text]))),
    CONSTRAINT evidence_access_handles_preview_kind_ck CHECK ((((handle_kind = 'preview'::text) AND (preview_kind IS NOT NULL)) OR ((handle_kind = 'download'::text) AND (preview_kind IS NULL)))),
    CONSTRAINT evidence_access_handles_size_bytes_check CHECK ((size_bytes >= 0))
);

--
-- Name: evidence_custody_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.evidence_custody_events (
    custody_event_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    evidence_record_id uuid NOT NULL,
    custody_event_type text NOT NULL,
    actor_user_id uuid,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    location_text text,
    note text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT evidence_custody_events_custody_event_type_check CHECK ((custody_event_type = ANY (ARRAY['requested'::text, 'received'::text, 'made_available'::text, 'transferred'::text, 'quarantined'::text, 'released'::text])))
);

--
-- Name: object_blobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.object_blobs (
    object_blob_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    storage_key text NOT NULL,
    upload_state text DEFAULT 'pending'::text NOT NULL,
    byte_size bigint NOT NULL,
    filename_hint text,
    content_type_hint text,
    expected_sha256_hex text,
    observed_size bigint,
    observed_content_type text,
    observed_sha256_hex text,
    target_expires_at timestamp with time zone NOT NULL,
    pending_expires_at timestamp with time zone NOT NULL,
    finalized_at timestamp with time zone,
    terminal_reason text,
    failed_at timestamp with time zone,
    finalize_attempt_count integer DEFAULT 0 NOT NULL,
    cleanup_due_at timestamp with time zone,
    cleaned_up_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT object_blobs_byte_size_check CHECK ((byte_size >= 0)),
    CONSTRAINT object_blobs_expected_sha256_hex_check CHECK (((expected_sha256_hex IS NULL) OR (expected_sha256_hex ~ '^[0-9a-f]{64}$'::text))),
    CONSTRAINT object_blobs_finalize_attempt_count_check CHECK ((finalize_attempt_count >= 0)),
    CONSTRAINT object_blobs_observed_sha256_hex_check CHECK (((observed_sha256_hex IS NULL) OR (observed_sha256_hex ~ '^[0-9a-f]{64}$'::text))),
    CONSTRAINT object_blobs_terminal_reason_check CHECK (((terminal_reason IS NULL) OR (terminal_reason = ANY (ARRAY['pending_timeout'::text, 'finalize_retry_exhausted'::text, 'declared_size_mismatch'::text, 'expected_sha256_mismatch'::text])))),
    CONSTRAINT object_blobs_terminal_state_ck CHECK ((((upload_state = 'failed'::text) AND (terminal_reason IS NOT NULL) AND (failed_at IS NOT NULL)) OR ((upload_state <> 'failed'::text) AND (terminal_reason IS NULL)))),
    CONSTRAINT object_blobs_upload_state_check CHECK ((upload_state = ANY (ARRAY['pending'::text, 'available'::text, 'failed'::text, 'quarantined'::text])))
);

--
-- Name: evidence_access_handles evidence_access_handles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_access_handles
    ADD CONSTRAINT evidence_access_handles_pkey PRIMARY KEY (handle_token);

--
-- Name: evidence_custody_events evidence_custody_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_custody_events
    ADD CONSTRAINT evidence_custody_events_pkey PRIMARY KEY (custody_event_id);

--
-- Name: evidence evidence_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence
    ADD CONSTRAINT evidence_pkey PRIMARY KEY (record_id);

--
-- Name: object_blobs object_blobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_blobs
    ADD CONSTRAINT object_blobs_pkey PRIMARY KEY (object_blob_id);

--
-- Name: object_blobs object_blobs_storage_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_blobs
    ADD CONSTRAINT object_blobs_storage_key_key UNIQUE (storage_key);

--
-- Name: evidence_access_handles_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX evidence_access_handles_lookup_idx ON public.evidence_access_handles USING btree (incident_id, record_id, created_at DESC);

--
-- Name: evidence_collector_party_ref_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX evidence_collector_party_ref_lookup_idx ON public.evidence USING btree (incident_id, collector_party_id) WHERE (collector_party_id IS NOT NULL);

--
-- Name: evidence_custody_events_record_time_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX evidence_custody_events_record_time_idx ON public.evidence_custody_events USING btree (evidence_record_id, occurred_at DESC, custody_event_id DESC);

--
-- Name: evidence_incident_record_unique_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX evidence_incident_record_unique_idx ON public.evidence USING btree (incident_id, record_id);

--
-- Name: evidence_incident_sort_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX evidence_incident_sort_idx ON public.evidence USING btree (incident_id, requested_at DESC, record_id);

--
-- Name: evidence_object_blob_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX evidence_object_blob_idx ON public.evidence USING btree (object_blob_id) WHERE (object_blob_id IS NOT NULL);

--
-- Name: evidence_source_party_ref_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX evidence_source_party_ref_lookup_idx ON public.evidence USING btree (incident_id, source_party_id) WHERE (source_party_id IS NOT NULL);

--
-- Name: object_blobs_incident_state_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX object_blobs_incident_state_idx ON public.object_blobs USING btree (incident_id, upload_state, created_at DESC);

--
-- Name: evidence_access_handles evidence_access_handles_evidence_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_access_handles
    ADD CONSTRAINT evidence_access_handles_evidence_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.evidence(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: evidence_access_handles evidence_access_handles_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_access_handles
    ADD CONSTRAINT evidence_access_handles_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: evidence_access_handles evidence_access_handles_issued_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_access_handles
    ADD CONSTRAINT evidence_access_handles_issued_by_user_id_fkey FOREIGN KEY (issued_by_user_id) REFERENCES public.users(id);

--
-- Name: evidence_access_handles evidence_access_handles_object_blob_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_access_handles
    ADD CONSTRAINT evidence_access_handles_object_blob_id_fkey FOREIGN KEY (object_blob_id) REFERENCES public.object_blobs(object_blob_id) ON DELETE CASCADE;

--
-- Name: evidence_custody_events evidence_custody_events_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_custody_events
    ADD CONSTRAINT evidence_custody_events_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id);

--
-- Name: evidence_custody_events evidence_custody_events_evidence_record_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_custody_events
    ADD CONSTRAINT evidence_custody_events_evidence_record_id_fkey FOREIGN KEY (evidence_record_id) REFERENCES public.evidence(record_id) ON DELETE CASCADE;

--
-- Name: evidence_custody_events evidence_custody_events_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence_custody_events
    ADD CONSTRAINT evidence_custody_events_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: evidence evidence_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence
    ADD CONSTRAINT evidence_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

--
-- Name: evidence evidence_object_blob_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence
    ADD CONSTRAINT evidence_object_blob_fkey FOREIGN KEY (object_blob_id) REFERENCES public.object_blobs(object_blob_id);

--
-- Name: evidence evidence_record_envelope_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidence
    ADD CONSTRAINT evidence_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE;

--
-- Name: object_blobs object_blobs_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_blobs
    ADD CONSTRAINT object_blobs_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: object_blobs object_blobs_incident_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_blobs
    ADD CONSTRAINT object_blobs_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.object_blobs CASCADE;
DROP TABLE IF EXISTS public.evidence CASCADE;
DROP TABLE IF EXISTS public.evidence_access_handles CASCADE;
DROP TABLE IF EXISTS public.evidence_custody_events CASCADE;

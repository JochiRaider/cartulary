-- +goose Up
--
-- Name: reference_pack_activation_state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reference_pack_activation_state (
    pack_key text NOT NULL,
    active_version text,
    previous_active_version text,
    activated_at timestamp with time zone,
    activated_by_user_id uuid,
    operator_note text,
    CONSTRAINT reference_pack_activation_state_check CHECK (((active_version IS NULL) OR (previous_active_version IS NULL) OR (previous_active_version <> active_version)))
);

--
-- Name: reference_pack_attestations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reference_pack_attestations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    pack_kind text NOT NULL,
    event_kind text NOT NULL,
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    source_identifier text,
    verification_method text NOT NULL,
    signer_key_id text,
    previous_active_version text,
    verification_result text NOT NULL,
    actor_user_id uuid,
    job_id uuid,
    occurred_at timestamp with time zone NOT NULL,
    operator_note text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT reference_pack_attestations_event_kind_check CHECK ((event_kind = ANY (ARRAY['import'::text, 'activate'::text, 'disable'::text, 'reverify'::text, 'refresh'::text]))),
    CONSTRAINT reference_pack_attestations_verification_result_check CHECK ((verification_result = ANY (ARRAY['pending'::text, 'passed'::text, 'failed'::text])))
);

--
-- Name: reference_pack_job_payloads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reference_pack_job_payloads (
    job_id uuid NOT NULL,
    job_kind text NOT NULL,
    actor_user_id uuid NOT NULL,
    pack_key text,
    pack_version text,
    resolved_pack_keys text[] DEFAULT '{}'::text[] NOT NULL,
    bundle_sha256 text,
    request_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    bundle_staging_ref text,
    CONSTRAINT reference_pack_job_payloads_job_kind_check CHECK ((job_kind = ANY (ARRAY['import'::text, 'reverify'::text, 'refresh'::text]))),
    CONSTRAINT reference_pack_job_payloads_bundle_staging_ref_relative_check CHECK (
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
-- Name: reference_packs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reference_packs (
    pack_key text NOT NULL,
    version text NOT NULL,
    pack_kind text NOT NULL,
    source_identifier text,
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    pack_contract_version text NOT NULL,
    verification_method text NOT NULL,
    signer_key_id text,
    status text DEFAULT 'staged'::text NOT NULL,
    imported_at timestamp with time zone NOT NULL,
    imported_by_user_id uuid,
    activated_at timestamp with time zone,
    activated_by_user_id uuid,
    previous_active_version text,
    verification_result text DEFAULT 'pending'::text NOT NULL,
    bundle_sha256 text NOT NULL,
    bundle_storage_ref text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT reference_packs_status_check CHECK ((status = ANY (ARRAY['staged'::text, 'available'::text, 'disabled'::text, 'failed'::text, 'missing'::text]))),
    CONSTRAINT reference_packs_verification_result_check CHECK ((verification_result = ANY (ARRAY['pending'::text, 'passed'::text, 'failed'::text]))),
    CONSTRAINT reference_packs_bundle_storage_ref_relative_check CHECK (
        bundle_storage_ref <> ''
        AND bundle_storage_ref !~ '^/'
        AND strpos(bundle_storage_ref, E'\\') = 0
        AND bundle_storage_ref !~ '(^|/)\.\.?(/|$)'
        AND bundle_storage_ref !~ '//'
        AND bundle_storage_ref !~ '/$'
    )
);

--
-- Name: reference_pack_activation_state reference_pack_activation_state_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_activation_state
    ADD CONSTRAINT reference_pack_activation_state_pkey PRIMARY KEY (pack_key);

--
-- Name: reference_pack_attestations reference_pack_attestations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_attestations
    ADD CONSTRAINT reference_pack_attestations_pkey PRIMARY KEY (id);

--
-- Name: reference_pack_job_payloads reference_pack_job_payloads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_job_payloads
    ADD CONSTRAINT reference_pack_job_payloads_pkey PRIMARY KEY (job_id);

--
-- Name: reference_packs reference_packs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_packs
    ADD CONSTRAINT reference_packs_pkey PRIMARY KEY (pack_key, version);

--
-- Name: reference_pack_attestations_pack_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reference_pack_attestations_pack_idx ON public.reference_pack_attestations USING btree (pack_key, pack_version, occurred_at);

--
-- Name: reference_pack_job_payloads_actor_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reference_pack_job_payloads_actor_idx ON public.reference_pack_job_payloads USING btree (actor_user_id, created_at);

--
-- Name: reference_packs_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reference_packs_status_idx ON public.reference_packs USING btree (status, verification_result);

--
-- Name: reference_pack_activation_state reference_pack_activation_sta_pack_key_previous_active_ver_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_activation_state
    ADD CONSTRAINT reference_pack_activation_sta_pack_key_previous_active_ver_fkey FOREIGN KEY (pack_key, previous_active_version) REFERENCES public.reference_packs(pack_key, version) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_pack_activation_state reference_pack_activation_state_activated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_activation_state
    ADD CONSTRAINT reference_pack_activation_state_activated_by_user_id_fkey FOREIGN KEY (activated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_pack_activation_state reference_pack_activation_state_pack_key_active_version_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_activation_state
    ADD CONSTRAINT reference_pack_activation_state_pack_key_active_version_fkey FOREIGN KEY (pack_key, active_version) REFERENCES public.reference_packs(pack_key, version) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_pack_attestations reference_pack_attestations_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_attestations
    ADD CONSTRAINT reference_pack_attestations_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_pack_attestations reference_pack_attestations_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_attestations
    ADD CONSTRAINT reference_pack_attestations_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE SET NULL;

--
-- Name: reference_pack_attestations reference_pack_attestations_pack_key_pack_version_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_attestations
    ADD CONSTRAINT reference_pack_attestations_pack_key_pack_version_fkey FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_packs(pack_key, version) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_pack_attestations reference_pack_attestations_pack_key_previous_active_versi_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_attestations
    ADD CONSTRAINT reference_pack_attestations_pack_key_previous_active_versi_fkey FOREIGN KEY (pack_key, previous_active_version) REFERENCES public.reference_packs(pack_key, version) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_pack_job_payloads reference_pack_job_payloads_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_job_payloads
    ADD CONSTRAINT reference_pack_job_payloads_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_pack_job_payloads reference_pack_job_payloads_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_pack_job_payloads
    ADD CONSTRAINT reference_pack_job_payloads_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE CASCADE;

--
-- Name: reference_packs reference_packs_activated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_packs
    ADD CONSTRAINT reference_packs_activated_by_user_id_fkey FOREIGN KEY (activated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_packs reference_packs_imported_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_packs
    ADD CONSTRAINT reference_packs_imported_by_user_id_fkey FOREIGN KEY (imported_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

--
-- Name: reference_packs reference_packs_pack_key_previous_active_version_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reference_packs
    ADD CONSTRAINT reference_packs_pack_key_previous_active_version_fkey FOREIGN KEY (pack_key, previous_active_version) REFERENCES public.reference_packs(pack_key, version) ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE INDEX reference_pack_activation_state_activated_by_user_id_fk_idx ON public.reference_pack_activation_state (activated_by_user_id);
CREATE INDEX reference_pack_activation_state_pack_key_active_version_fk_idx ON public.reference_pack_activation_state (pack_key, active_version);
CREATE INDEX reference_pack_activation_state_pack_key_previous_acti_eb58f173 ON public.reference_pack_activation_state (pack_key, previous_active_version);
CREATE INDEX reference_pack_attestations_actor_user_id_fk_idx ON public.reference_pack_attestations (actor_user_id);
CREATE INDEX reference_pack_attestations_job_id_fk_idx ON public.reference_pack_attestations (job_id);
CREATE INDEX reference_pack_attestations_pack_key_previous_active_v_75bb245c ON public.reference_pack_attestations (pack_key, previous_active_version);
CREATE INDEX reference_packs_activated_by_user_id_fk_idx ON public.reference_packs (activated_by_user_id);
CREATE INDEX reference_packs_imported_by_user_id_fk_idx ON public.reference_packs (imported_by_user_id);
CREATE INDEX reference_packs_pack_key_previous_active_version_fk_idx ON public.reference_packs (pack_key, previous_active_version);

-- +goose Down
DROP TABLE public.reference_pack_job_payloads ;
DROP TABLE public.reference_pack_attestations ;
DROP TABLE public.reference_pack_activation_state ;
DROP TABLE public.reference_packs ;

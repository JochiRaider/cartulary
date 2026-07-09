-- +goose Up
--
-- Name: account_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.account_preferences (
    user_id uuid NOT NULL,
    density_mode text,
    preferences_version bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT account_preferences_density_mode_check CHECK ((density_mode = ANY (ARRAY['compact'::text, 'default'::text, 'comfortable'::text])))
);

--
-- Name: bootstrap_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bootstrap_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_fingerprint bytea NOT NULL,
    issued_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    superseded_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: enterprise_auth_bindings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.enterprise_auth_bindings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    provider_id uuid NOT NULL,
    provider_key text NOT NULL,
    provider_type text NOT NULL,
    provider_subject text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    last_auth_at timestamp with time zone,
    retired_at timestamp with time zone,
    retired_by_user_id uuid,
    retire_reason text,
    replaced_by_auth_binding_id uuid,
    CONSTRAINT enterprise_auth_bindings_provider_subject_check CHECK ((provider_subject <> ''::text)),
    CONSTRAINT enterprise_auth_bindings_provider_type_check CHECK ((provider_type = ANY (ARRAY['oidc'::text, 'saml'::text])))
);

--
-- Name: enterprise_auth_providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.enterprise_auth_providers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider_key text NOT NULL,
    provider_type text NOT NULL,
    display_name text NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    is_interactive boolean DEFAULT true NOT NULL,
    authorization_endpoint text,
    issuer text,
    audience text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT enterprise_auth_providers_display_name_check CHECK (((char_length(display_name) >= 1) AND (char_length(display_name) <= 256))),
    CONSTRAINT enterprise_auth_providers_provider_key_check CHECK ((provider_key ~ '^[a-z0-9][a-z0-9._-]{1,126}[a-z0-9]$'::text)),
    CONSTRAINT enterprise_auth_providers_provider_type_check CHECK ((provider_type = ANY (ARRAY['oidc'::text, 'saml'::text])))
);

--
-- Name: enterprise_auth_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.enterprise_auth_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider_id uuid NOT NULL,
    provider_key text NOT NULL,
    provider_type text NOT NULL,
    return_to text NOT NULL,
    state text,
    nonce text,
    pkce_verifier_hash bytea,
    relay_state text,
    browser_binding_hash bytea NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    saml_completion_hash bytea,
    saml_subject text,
    saml_staged_at timestamp with time zone,
    CONSTRAINT enterprise_auth_transactions_correlation_ck CHECK ((((provider_type = 'oidc'::text) AND (state IS NOT NULL) AND (nonce IS NOT NULL) AND (relay_state IS NULL)) OR ((provider_type = 'saml'::text) AND (relay_state IS NOT NULL) AND (state IS NULL) AND (nonce IS NULL)))),
    CONSTRAINT enterprise_auth_transactions_provider_type_check CHECK ((provider_type = ANY (ARRAY['oidc'::text, 'saml'::text]))),
    CONSTRAINT enterprise_auth_transactions_saml_staging_ck CHECK ((((provider_type <> 'saml'::text) AND (saml_completion_hash IS NULL) AND (saml_subject IS NULL) AND (saml_staged_at IS NULL)) OR ((provider_type = 'saml'::text) AND (((saml_completion_hash IS NULL) AND (saml_subject IS NULL) AND (saml_staged_at IS NULL)) OR ((saml_completion_hash IS NOT NULL) AND (saml_subject IS NOT NULL) AND (saml_subject <> ''::text) AND (saml_staged_at IS NOT NULL))))))
);

--
-- Name: pending_totp_enrollments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pending_totp_enrollments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    auth_scope_kind text NOT NULL,
    auth_scope_session_id uuid,
    auth_scope_bootstrap_token_id uuid,
    client_txn_id text NOT NULL,
    secret_ciphertext bytea NOT NULL,
    secret_nonce bytea NOT NULL,
    replaces_active boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    CONSTRAINT pending_totp_enrollments_auth_scope_ck CHECK ((((auth_scope_kind = 'session'::text) AND (auth_scope_session_id IS NOT NULL) AND (auth_scope_bootstrap_token_id IS NULL)) OR ((auth_scope_kind = 'bootstrap_token'::text) AND (auth_scope_bootstrap_token_id IS NOT NULL) AND (auth_scope_session_id IS NULL)))),
    CONSTRAINT pending_totp_enrollments_auth_scope_kind_check CHECK ((auth_scope_kind = ANY (ARRAY['session'::text, 'bootstrap_token'::text])))
);

--
-- Name: route_idempotency; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.route_idempotency (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    route_key text NOT NULL,
    scope_key text NOT NULL,
    client_txn_id text NOT NULL,
    actor_user_id uuid NOT NULL,
    target_user_id uuid,
    request_hash bytea NOT NULL,
    status_code integer NOT NULL,
    response_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_fingerprint bytea NOT NULL,
    authenticated_at timestamp with time zone NOT NULL,
    last_qualifying_activity_at timestamp with time zone NOT NULL,
    idle_expires_at timestamp with time zone NOT NULL,
    absolute_expires_at timestamp with time zone NOT NULL,
    session_expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    revoke_reason_code text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    provider_type text DEFAULT 'local'::text NOT NULL,
    auth_binding_id uuid,
    CONSTRAINT user_sessions_provider_type_check CHECK ((provider_type = ANY (ARRAY['local'::text, 'oidc'::text, 'saml'::text])))
);

--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email citext NOT NULL,
    display_name text NOT NULL,
    password_hash text NOT NULL,
    password_changed_at timestamp with time zone DEFAULT now() NOT NULL,
    mfa_required boolean DEFAULT true NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_deployment_admin boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    last_login_at timestamp with time zone,
    user_version bigint DEFAULT 1 NOT NULL,
    updated_by_user_id uuid,
    totp_enrolled_at timestamp with time zone,
    totp_secret_ciphertext bytea,
    totp_secret_nonce bytea,
    CONSTRAINT users_display_name_check CHECK ((char_length(display_name) <= 256))
);

--
-- Name: account_preferences account_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_preferences
    ADD CONSTRAINT account_preferences_pkey PRIMARY KEY (user_id);

--
-- Name: bootstrap_tokens bootstrap_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bootstrap_tokens
    ADD CONSTRAINT bootstrap_tokens_pkey PRIMARY KEY (id);

--
-- Name: bootstrap_tokens bootstrap_tokens_token_fingerprint_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bootstrap_tokens
    ADD CONSTRAINT bootstrap_tokens_token_fingerprint_key UNIQUE (token_fingerprint);

--
-- Name: enterprise_auth_bindings enterprise_auth_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_bindings
    ADD CONSTRAINT enterprise_auth_bindings_pkey PRIMARY KEY (id);

--
-- Name: enterprise_auth_providers enterprise_auth_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_providers
    ADD CONSTRAINT enterprise_auth_providers_pkey PRIMARY KEY (id);

--
-- Name: enterprise_auth_providers enterprise_auth_providers_provider_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_providers
    ADD CONSTRAINT enterprise_auth_providers_provider_key_key UNIQUE (provider_key);

--
-- Name: enterprise_auth_transactions enterprise_auth_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_transactions
    ADD CONSTRAINT enterprise_auth_transactions_pkey PRIMARY KEY (id);

--
-- Name: pending_totp_enrollments pending_totp_enrollments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pending_totp_enrollments
    ADD CONSTRAINT pending_totp_enrollments_pkey PRIMARY KEY (id);

--
-- Name: route_idempotency route_idempotency_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.route_idempotency
    ADD CONSTRAINT route_idempotency_pkey PRIMARY KEY (id);

--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);

--
-- Name: user_sessions user_sessions_token_fingerprint_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_token_fingerprint_key UNIQUE (token_fingerprint);

--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

--
-- Name: bootstrap_tokens_user_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX bootstrap_tokens_user_lookup_idx ON public.bootstrap_tokens USING btree (user_id, consumed_at, superseded_at, expires_at);

--
-- Name: enterprise_auth_bindings_active_provider_subject_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX enterprise_auth_bindings_active_provider_subject_idx ON public.enterprise_auth_bindings USING btree (provider_id, provider_subject) WHERE (retired_at IS NULL);

--
-- Name: enterprise_auth_bindings_active_user_provider_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX enterprise_auth_bindings_active_user_provider_idx ON public.enterprise_auth_bindings USING btree (user_id, provider_id) WHERE (retired_at IS NULL);

--
-- Name: enterprise_auth_bindings_user_active_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX enterprise_auth_bindings_user_active_idx ON public.enterprise_auth_bindings USING btree (user_id, provider_type, provider_key, created_at) WHERE (retired_at IS NULL);

--
-- Name: enterprise_auth_providers_discovery_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX enterprise_auth_providers_discovery_idx ON public.enterprise_auth_providers USING btree (display_name, provider_key) WHERE ((is_enabled = true) AND (is_interactive = true));

--
-- Name: enterprise_auth_transactions_provider_expiry_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX enterprise_auth_transactions_provider_expiry_idx ON public.enterprise_auth_transactions USING btree (provider_id, expires_at, consumed_at);

--
-- Name: enterprise_auth_transactions_relay_state_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX enterprise_auth_transactions_relay_state_idx ON public.enterprise_auth_transactions USING btree (relay_state) WHERE (relay_state IS NOT NULL);

--
-- Name: enterprise_auth_transactions_saml_completion_hash_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX enterprise_auth_transactions_saml_completion_hash_idx ON public.enterprise_auth_transactions USING btree (saml_completion_hash) WHERE (saml_completion_hash IS NOT NULL);

--
-- Name: enterprise_auth_transactions_state_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX enterprise_auth_transactions_state_idx ON public.enterprise_auth_transactions USING btree (state) WHERE (state IS NOT NULL);

--
-- Name: pending_totp_enrollments_one_current_per_user_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX pending_totp_enrollments_one_current_per_user_idx ON public.pending_totp_enrollments USING btree (user_id) WHERE (consumed_at IS NULL);

--
-- Name: pending_totp_enrollments_scope_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pending_totp_enrollments_scope_lookup_idx ON public.pending_totp_enrollments USING btree (auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id);

--
-- Name: route_idempotency_actor_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX route_idempotency_actor_lookup_idx ON public.route_idempotency USING btree (actor_user_id, created_at DESC);

--
-- Name: route_idempotency_route_actor_scope_client_txn_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX route_idempotency_route_actor_scope_client_txn_idx ON public.route_idempotency USING btree (route_key, actor_user_id, scope_key, client_txn_id);

--
-- Name: user_sessions_user_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_sessions_user_lookup_idx ON public.user_sessions USING btree (user_id, revoked_at, session_expires_at, last_qualifying_activity_at);

--
-- Name: account_preferences account_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.account_preferences
    ADD CONSTRAINT account_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: bootstrap_tokens bootstrap_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bootstrap_tokens
    ADD CONSTRAINT bootstrap_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: enterprise_auth_bindings enterprise_auth_bindings_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_bindings
    ADD CONSTRAINT enterprise_auth_bindings_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

--
-- Name: enterprise_auth_bindings enterprise_auth_bindings_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_bindings
    ADD CONSTRAINT enterprise_auth_bindings_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.enterprise_auth_providers(id) ON DELETE RESTRICT;

--
-- Name: enterprise_auth_bindings enterprise_auth_bindings_replaced_by_auth_binding_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_bindings
    ADD CONSTRAINT enterprise_auth_bindings_replaced_by_auth_binding_id_fkey FOREIGN KEY (replaced_by_auth_binding_id) REFERENCES public.enterprise_auth_bindings(id);

--
-- Name: enterprise_auth_bindings enterprise_auth_bindings_retired_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_bindings
    ADD CONSTRAINT enterprise_auth_bindings_retired_by_user_id_fkey FOREIGN KEY (retired_by_user_id) REFERENCES public.users(id);

--
-- Name: enterprise_auth_bindings enterprise_auth_bindings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_bindings
    ADD CONSTRAINT enterprise_auth_bindings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: enterprise_auth_transactions enterprise_auth_transactions_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.enterprise_auth_transactions
    ADD CONSTRAINT enterprise_auth_transactions_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.enterprise_auth_providers(id) ON DELETE CASCADE;

--
-- Name: pending_totp_enrollments pending_totp_enrollments_auth_scope_bootstrap_token_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pending_totp_enrollments
    ADD CONSTRAINT pending_totp_enrollments_auth_scope_bootstrap_token_id_fkey FOREIGN KEY (auth_scope_bootstrap_token_id) REFERENCES public.bootstrap_tokens(id) ON DELETE CASCADE;

--
-- Name: pending_totp_enrollments pending_totp_enrollments_auth_scope_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pending_totp_enrollments
    ADD CONSTRAINT pending_totp_enrollments_auth_scope_session_id_fkey FOREIGN KEY (auth_scope_session_id) REFERENCES public.user_sessions(id) ON DELETE CASCADE;

--
-- Name: pending_totp_enrollments pending_totp_enrollments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pending_totp_enrollments
    ADD CONSTRAINT pending_totp_enrollments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: route_idempotency route_idempotency_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.route_idempotency
    ADD CONSTRAINT route_idempotency_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id);

--
-- Name: route_idempotency route_idempotency_target_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.route_idempotency
    ADD CONSTRAINT route_idempotency_target_user_id_fkey FOREIGN KEY (target_user_id) REFERENCES public.users(id);

--
-- Name: user_sessions user_sessions_auth_binding_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_auth_binding_id_fkey FOREIGN KEY (auth_binding_id) REFERENCES public.enterprise_auth_bindings(id);

--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

--
-- Name: users users_updated_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id);

-- +goose Down
DROP TABLE IF EXISTS public.users CASCADE;
DROP TABLE IF EXISTS public.user_sessions CASCADE;
DROP TABLE IF EXISTS public.bootstrap_tokens CASCADE;
DROP TABLE IF EXISTS public.pending_totp_enrollments CASCADE;
DROP TABLE IF EXISTS public.route_idempotency CASCADE;
DROP TABLE IF EXISTS public.account_preferences CASCADE;
DROP TABLE IF EXISTS public.enterprise_auth_providers CASCADE;
DROP TABLE IF EXISTS public.enterprise_auth_transactions CASCADE;
DROP TABLE IF EXISTS public.enterprise_auth_bindings CASCADE;

-- +goose Up
--
-- Name: backup_sets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.backup_sets (
    backup_set_id uuid DEFAULT gen_random_uuid() NOT NULL,
    consistency_point_at timestamp with time zone NOT NULL,
    postgres_restore_anchor text NOT NULL,
    object_store_restore_anchor text NOT NULL,
    postgres_artifact_key text NOT NULL,
    postgres_artifact_sha256 text NOT NULL,
    postgres_artifact_size_bytes bigint NOT NULL,
    object_store_artifact_key text NOT NULL,
    object_store_artifact_sha256 text NOT NULL,
    object_store_artifact_size_bytes bigint NOT NULL,
    integrity_manifest_key text NOT NULL,
    integrity_manifest_sha256 text NOT NULL,
    integrity_manifest_size_bytes bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    retained_until timestamp with time zone NOT NULL,
    postgres_restore_anchor_retained_until timestamp with time zone NOT NULL,
    object_store_restore_anchor_retained_until timestamp with time zone NOT NULL,
    verification_state text DEFAULT 'unverified'::text NOT NULL,
    last_verified_restore_at timestamp with time zone,
    last_verification_basis_sha256 text,
    CONSTRAINT backup_sets_integrity_manifest_key_non_empty CHECK ((btrim(integrity_manifest_key) <> ''::text)),
    CONSTRAINT backup_sets_integrity_manifest_sha256_check CHECK ((integrity_manifest_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT backup_sets_integrity_manifest_size_check CHECK ((integrity_manifest_size_bytes > 0)),
    CONSTRAINT backup_sets_last_verification_basis_sha256_check CHECK (((last_verification_basis_sha256 IS NULL) OR (last_verification_basis_sha256 ~ '^[0-9a-f]{64}$'::text))),
    CONSTRAINT backup_sets_object_store_anchor_retained_until_floor_check CHECK ((object_store_restore_anchor_retained_until >= (created_at + '30 days'::interval))),
    CONSTRAINT backup_sets_object_store_artifact_key_non_empty CHECK ((btrim(object_store_artifact_key) <> ''::text)),
    CONSTRAINT backup_sets_object_store_artifact_sha256_check CHECK ((object_store_artifact_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT backup_sets_object_store_artifact_size_check CHECK ((object_store_artifact_size_bytes > 0)),
    CONSTRAINT backup_sets_object_store_restore_anchor_non_empty CHECK ((btrim(object_store_restore_anchor) <> ''::text)),
    CONSTRAINT backup_sets_postgres_anchor_retained_until_floor_check CHECK ((postgres_restore_anchor_retained_until >= (created_at + '30 days'::interval))),
    CONSTRAINT backup_sets_postgres_artifact_key_non_empty CHECK ((btrim(postgres_artifact_key) <> ''::text)),
    CONSTRAINT backup_sets_postgres_artifact_sha256_check CHECK ((postgres_artifact_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT backup_sets_postgres_artifact_size_check CHECK ((postgres_artifact_size_bytes > 0)),
    CONSTRAINT backup_sets_postgres_restore_anchor_non_empty CHECK ((btrim(postgres_restore_anchor) <> ''::text)),
    CONSTRAINT backup_sets_retained_until_floor_check CHECK ((retained_until >= (created_at + '30 days'::interval))),
    CONSTRAINT backup_sets_verification_state_check CHECK ((verification_state = ANY (ARRAY['unverified'::text, 'verified'::text, 'failed'::text]))),
    CONSTRAINT backup_sets_verification_timestamp_check CHECK ((((verification_state = 'unverified'::text) AND (last_verified_restore_at IS NULL)) OR ((verification_state = ANY (ARRAY['verified'::text, 'failed'::text])) AND (last_verified_restore_at IS NOT NULL))))
);

--
-- Name: operator_recovery_journal; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.operator_recovery_journal (
    operator_recovery_journal_id uuid DEFAULT gen_random_uuid() NOT NULL,
    operation_id uuid NOT NULL,
    operation text NOT NULL,
    result text NOT NULL,
    backup_set_id uuid,
    error_code text,
    reason_code text,
    envelope_schema_id text NOT NULL,
    encryption_mode text NOT NULL,
    key_fingerprint_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    nonce bytea NOT NULL,
    ciphertext bytea NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT operator_recovery_journal_encryption_mode_check CHECK ((encryption_mode = 'aes-256-gcm-envelope'::text)),
    CONSTRAINT operator_recovery_journal_envelope_schema_id_check CHECK ((envelope_schema_id = 'cartulary.operator_recovery_journal_envelope.v1'::text)),
    CONSTRAINT operator_recovery_journal_key_fingerprint_sha256_check CHECK ((key_fingerprint_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT operator_recovery_journal_operation_check CHECK ((operation = ANY (ARRAY['backup_create'::text, 'restore_latest'::text, 'restore_verify_latest'::text, 'restore_verify_due'::text]))),
    CONSTRAINT operator_recovery_journal_payload_sha256_check CHECK ((payload_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT operator_recovery_journal_result_check CHECK ((result = ANY (ARRAY['started'::text, 'succeeded'::text, 'failed'::text, 'no_op'::text])))
);

--
-- Name: restore_verification_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.restore_verification_runs (
    restore_verification_run_id uuid DEFAULT gen_random_uuid() NOT NULL,
    backup_set_id uuid NOT NULL,
    started_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone NOT NULL,
    verification_state text NOT NULL,
    verification_basis_sha256 text NOT NULL,
    failure_reason text,
    failure_message text,
    authoritative_rows_sha256 text,
    authoritative_row_count integer,
    change_sets_sha256 text,
    change_set_row_count integer,
    blob_hashes_sha256 text,
    blob_count integer,
    CONSTRAINT restore_verification_runs_basis_sha256_check CHECK ((verification_basis_sha256 ~ '^[0-9a-f]{64}$'::text)),
    CONSTRAINT restore_verification_runs_completed_after_started_check CHECK ((completed_at >= started_at)),
    CONSTRAINT restore_verification_runs_failure_shape_check CHECK ((((verification_state = 'failed'::text) AND (failure_reason IS NOT NULL) AND (failure_message IS NOT NULL)) OR ((verification_state = 'verified'::text) AND (failure_reason IS NULL) AND (failure_message IS NULL)))),
    CONSTRAINT restore_verification_runs_state_check CHECK ((verification_state = ANY (ARRAY['verified'::text, 'failed'::text])))
);

--
-- Name: backup_sets backup_sets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_sets
    ADD CONSTRAINT backup_sets_pkey PRIMARY KEY (backup_set_id);

--
-- Name: operator_recovery_journal operator_recovery_journal_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operator_recovery_journal
    ADD CONSTRAINT operator_recovery_journal_pkey PRIMARY KEY (operator_recovery_journal_id);

--
-- Name: restore_verification_runs restore_verification_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restore_verification_runs
    ADD CONSTRAINT restore_verification_runs_pkey PRIMARY KEY (restore_verification_run_id);

--
-- Name: backup_sets_latest_retained_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX backup_sets_latest_retained_idx ON public.backup_sets USING btree (consistency_point_at DESC, backup_set_id, retained_until);

--
-- Name: operator_recovery_journal_backup_set_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX operator_recovery_journal_backup_set_idx ON public.operator_recovery_journal USING btree (backup_set_id, created_at DESC) WHERE (backup_set_id IS NOT NULL);

--
-- Name: operator_recovery_journal_operation_created_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX operator_recovery_journal_operation_created_idx ON public.operator_recovery_journal USING btree (operation, created_at DESC, operation_id);

--
-- Name: restore_verification_runs_backup_set_started_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX restore_verification_runs_backup_set_started_idx ON public.restore_verification_runs USING btree (backup_set_id, started_at DESC, restore_verification_run_id);

--
-- Name: restore_verification_runs restore_verification_runs_backup_set_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restore_verification_runs
    ADD CONSTRAINT restore_verification_runs_backup_set_id_fkey FOREIGN KEY (backup_set_id) REFERENCES public.backup_sets(backup_set_id) ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
DROP TABLE public.restore_verification_runs ;
DROP TABLE public.backup_sets ;
DROP TABLE public.operator_recovery_journal ;

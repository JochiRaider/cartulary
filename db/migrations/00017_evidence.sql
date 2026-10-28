-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: evidence.
-- Administrator provisioning and final admission are separate lifecycle steps.

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
    object_blob_id uuid,
    CONSTRAINT evidence_pkey PRIMARY KEY (record_id),
    CONSTRAINT evidence_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.evidence FROM PUBLIC;

REVOKE ALL ON TYPE public.evidence FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.evidence TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence TO cartulary_recovery;

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
    CONSTRAINT evidence_access_handles_disposition_check CHECK (disposition = ANY (ARRAY['inline'::text, 'attachment'::text])),
    CONSTRAINT evidence_access_handles_handle_kind_check CHECK (handle_kind = ANY (ARRAY['preview'::text, 'download'::text])),
    CONSTRAINT evidence_access_handles_pkey PRIMARY KEY (handle_token),
    CONSTRAINT evidence_access_handles_preview_kind_ck CHECK (handle_kind = 'preview'::text AND preview_kind IS NOT NULL OR handle_kind = 'download'::text AND preview_kind IS NULL),
    CONSTRAINT evidence_access_handles_size_bytes_check CHECK (size_bytes >= 0)
);

REVOKE ALL ON TABLE public.evidence_access_handles FROM PUBLIC;

REVOKE ALL ON TYPE public.evidence_access_handles FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.evidence_access_handles TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_access_handles TO cartulary_recovery;

CREATE TABLE public.evidence_blob_cleanup_claims (
    object_blob_id uuid NOT NULL,
    claim_token uuid NOT NULL,
    claim_state text NOT NULL,
    attempt_count integer NOT NULL,
    claimed_at timestamp with time zone NOT NULL,
    claim_expires_at timestamp with time zone,
    next_attempt_at timestamp with time zone,
    last_attempt_at timestamp with time zone,
    completed_at timestamp with time zone,
    last_failure_class text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT evidence_blob_cleanup_claims_attempt_count_ck CHECK (attempt_count >= 1),
    CONSTRAINT evidence_blob_cleanup_claims_claim_token_key UNIQUE (claim_token),
    CONSTRAINT evidence_blob_cleanup_claims_failure_class_ck CHECK (last_failure_class IS NULL OR (last_failure_class = ANY (ARRAY['delete_failed'::text, 'delete_timeout'::text, 'state_changed'::text]))),
    CONSTRAINT evidence_blob_cleanup_claims_pkey PRIMARY KEY (object_blob_id),
    CONSTRAINT evidence_blob_cleanup_claims_state_ck CHECK (claim_state = ANY (ARRAY['claimed'::text, 'retry_wait'::text, 'completed'::text])),
    CONSTRAINT evidence_blob_cleanup_claims_time_ck CHECK (updated_at >= created_at AND (last_attempt_at IS NULL OR last_attempt_at >= claimed_at) AND (completed_at IS NULL OR completed_at >= claimed_at)),
    CONSTRAINT evidence_blob_cleanup_claims_transition_ck CHECK (claim_state = 'claimed'::text AND claim_expires_at IS NOT NULL AND next_attempt_at IS NULL AND completed_at IS NULL OR claim_state = 'retry_wait'::text AND claim_expires_at IS NULL AND next_attempt_at IS NOT NULL AND completed_at IS NULL AND last_failure_class IS NOT NULL OR claim_state = 'completed'::text AND claim_expires_at IS NULL AND next_attempt_at IS NULL AND completed_at IS NOT NULL)
);

REVOKE ALL ON TABLE public.evidence_blob_cleanup_claims FROM PUBLIC;

REVOKE ALL ON TYPE public.evidence_blob_cleanup_claims FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.evidence_blob_cleanup_claims TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_blob_cleanup_claims TO cartulary_recovery;

CREATE TABLE public.evidence_blob_verifications (
    object_blob_id uuid NOT NULL,
    upload_lease_id uuid NOT NULL,
    job_id uuid NOT NULL,
    verified_at timestamp with time zone NOT NULL,
    observed_size bigint NOT NULL,
    observed_content_type text NOT NULL,
    observed_sha256_hex text NOT NULL,
    CONSTRAINT evidence_blob_verifications_content_type_ck CHECK (length(observed_content_type) >= 1 AND length(observed_content_type) <= 1024),
    CONSTRAINT evidence_blob_verifications_pkey PRIMARY KEY (object_blob_id),
    CONSTRAINT evidence_blob_verifications_sha256_ck CHECK (observed_sha256_hex ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT evidence_blob_verifications_size_ck CHECK (observed_size >= 0)
);

REVOKE ALL ON TABLE public.evidence_blob_verifications FROM PUBLIC;

REVOKE ALL ON TYPE public.evidence_blob_verifications FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.evidence_blob_verifications TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_blob_verifications TO cartulary_recovery;

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
    CONSTRAINT evidence_custody_events_custody_event_type_check CHECK (custody_event_type = ANY (ARRAY['requested'::text, 'received'::text, 'made_available'::text, 'transferred'::text, 'quarantined'::text, 'released'::text])),
    CONSTRAINT evidence_custody_events_pkey PRIMARY KEY (custody_event_id)
);

REVOKE ALL ON TABLE public.evidence_custody_events FROM PUBLIC;

REVOKE ALL ON TYPE public.evidence_custody_events FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.evidence_custody_events TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_custody_events TO cartulary_recovery;

CREATE TABLE public.evidence_object_upload_leases (
    object_blob_id uuid NOT NULL,
    lease_id uuid NOT NULL,
    capability_hash bytea NOT NULL,
    incident_id uuid NOT NULL,
    issuing_user_id uuid NOT NULL,
    issuing_session_id uuid NOT NULL,
    issued_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    required_method text NOT NULL,
    required_headers jsonb NOT NULL,
    accepted_contract_sha256 bytea NOT NULL,
    lease_state text DEFAULT 'issued'::text NOT NULL,
    claimed_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT evidence_object_upload_leases_capability_digest_ck CHECK (octet_length(capability_hash) = 32),
    CONSTRAINT evidence_object_upload_leases_capability_hash_key UNIQUE (capability_hash),
    CONSTRAINT evidence_object_upload_leases_contract_digest_ck CHECK (octet_length(accepted_contract_sha256) = 32),
    CONSTRAINT evidence_object_upload_leases_headers_ck CHECK (jsonb_typeof(required_headers) = 'object'::text),
    CONSTRAINT evidence_object_upload_leases_lease_id_key UNIQUE (lease_id),
    CONSTRAINT evidence_object_upload_leases_method_ck CHECK (required_method = 'PUT'::text),
    CONSTRAINT evidence_object_upload_leases_pkey PRIMARY KEY (object_blob_id),
    CONSTRAINT evidence_object_upload_leases_state_ck CHECK (lease_state = ANY (ARRAY['issued'::text, 'claimed'::text, 'completed'::text])),
    CONSTRAINT evidence_object_upload_leases_time_ck CHECK (expires_at > issued_at AND updated_at >= created_at),
    CONSTRAINT evidence_object_upload_leases_transition_ck CHECK (lease_state = 'issued'::text AND claimed_at IS NULL AND completed_at IS NULL OR lease_state = 'claimed'::text AND claimed_at IS NOT NULL AND completed_at IS NULL OR lease_state = 'completed'::text AND claimed_at IS NOT NULL AND completed_at IS NOT NULL)
);

REVOKE ALL ON TABLE public.evidence_object_upload_leases FROM PUBLIC;

REVOKE ALL ON TYPE public.evidence_object_upload_leases FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.evidence_object_upload_leases TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_object_upload_leases TO cartulary_recovery;

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
    cleanup_due_at timestamp with time zone,
    cleaned_up_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT object_blobs_byte_size_check CHECK (byte_size >= 0),
    CONSTRAINT object_blobs_expected_sha256_hex_check CHECK (expected_sha256_hex IS NULL OR expected_sha256_hex ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT object_blobs_observed_sha256_hex_check CHECK (observed_sha256_hex IS NULL OR observed_sha256_hex ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT object_blobs_pkey PRIMARY KEY (object_blob_id),
    CONSTRAINT object_blobs_storage_key_key UNIQUE (storage_key),
    CONSTRAINT object_blobs_terminal_reason_check CHECK (terminal_reason IS NULL OR (terminal_reason = ANY (ARRAY['pending_timeout'::text, 'declared_size_mismatch'::text, 'expected_sha256_mismatch'::text]))),
    CONSTRAINT object_blobs_terminal_state_ck CHECK (upload_state = 'failed'::text AND terminal_reason IS NOT NULL AND failed_at IS NOT NULL OR upload_state <> 'failed'::text AND terminal_reason IS NULL),
    CONSTRAINT object_blobs_upload_state_check CHECK (upload_state = ANY (ARRAY['pending'::text, 'available'::text, 'failed'::text, 'quarantined'::text])),
    CONSTRAINT object_blobs_incident_identity_key UNIQUE (incident_id, object_blob_id)
);

REVOKE ALL ON TABLE public.object_blobs FROM PUBLIC;

REVOKE ALL ON TYPE public.object_blobs FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.object_blobs TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.object_blobs TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.evidence_reject_cleanup_claimed_association()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF NEW.object_blob_id IS NOT NULL AND EXISTS (
        SELECT 1
          FROM public.evidence_blob_cleanup_claims c
         WHERE c.object_blob_id = NEW.object_blob_id
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'object blob has an active durable cleanup claim';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.evidence_reject_cleanup_claimed_association() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.evidence_reject_cleanup_claimed_blob_update()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.evidence_blob_cleanup_claims c
         WHERE c.object_blob_id = OLD.object_blob_id
    ) AND (
        NEW.incident_id IS DISTINCT FROM OLD.incident_id
        OR NEW.created_by_user_id IS DISTINCT FROM OLD.created_by_user_id
        OR NEW.storage_key IS DISTINCT FROM OLD.storage_key
        OR NEW.upload_state IS DISTINCT FROM OLD.upload_state
        OR NEW.byte_size IS DISTINCT FROM OLD.byte_size
        OR NEW.filename_hint IS DISTINCT FROM OLD.filename_hint
        OR NEW.content_type_hint IS DISTINCT FROM OLD.content_type_hint
        OR NEW.expected_sha256_hex IS DISTINCT FROM OLD.expected_sha256_hex
        OR NEW.observed_size IS DISTINCT FROM OLD.observed_size
        OR NEW.observed_content_type IS DISTINCT FROM OLD.observed_content_type
        OR NEW.observed_sha256_hex IS DISTINCT FROM OLD.observed_sha256_hex
        OR NEW.target_expires_at IS DISTINCT FROM OLD.target_expires_at
        OR NEW.pending_expires_at IS DISTINCT FROM OLD.pending_expires_at
        OR NEW.finalized_at IS DISTINCT FROM OLD.finalized_at
        OR NEW.terminal_reason IS DISTINCT FROM OLD.terminal_reason
        OR NEW.failed_at IS DISTINCT FROM OLD.failed_at
        OR NEW.cleanup_due_at IS DISTINCT FROM OLD.cleanup_due_at
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'object blob has an active durable cleanup claim';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.evidence_reject_cleanup_claimed_blob_update() FROM PUBLIC;

CREATE INDEX evidence_access_handles_issued_by_user_id_fk_idx ON public.evidence_access_handles USING btree (issued_by_user_id);

CREATE INDEX evidence_access_handles_lookup_idx ON public.evidence_access_handles USING btree (incident_id, record_id, created_at DESC);

CREATE INDEX evidence_access_handles_object_blob_id_fk_idx ON public.evidence_access_handles USING btree (object_blob_id);

CREATE INDEX evidence_access_handles_object_blob_id_scope_idx ON public.evidence_access_handles (incident_id, object_blob_id);

CREATE INDEX evidence_blob_cleanup_claims_completed_idx ON public.evidence_blob_cleanup_claims USING btree (completed_at, object_blob_id) WHERE (claim_state = 'completed'::text);

CREATE INDEX evidence_blob_cleanup_claims_due_idx ON public.evidence_blob_cleanup_claims USING btree (next_attempt_at, object_blob_id) WHERE (claim_state = 'retry_wait'::text);

CREATE INDEX evidence_blob_cleanup_claims_lease_expiry_idx ON public.evidence_blob_cleanup_claims USING btree (claim_expires_at, object_blob_id) WHERE (claim_state = 'claimed'::text);

CREATE INDEX evidence_blob_verifications_lease_fk_idx ON public.evidence_blob_verifications USING btree (upload_lease_id);

CREATE INDEX evidence_collector_party_ref_lookup_idx ON public.evidence USING btree (incident_id, collector_party_id) WHERE (collector_party_id IS NOT NULL);

CREATE INDEX evidence_custody_events_actor_user_id_fk_idx ON public.evidence_custody_events USING btree (actor_user_id);

CREATE INDEX evidence_custody_events_evidence_record_id_scope_idx ON public.evidence_custody_events (incident_id, evidence_record_id);

CREATE INDEX evidence_custody_events_incident_id_fk_idx ON public.evidence_custody_events USING btree (incident_id);

CREATE INDEX evidence_custody_events_record_time_idx ON public.evidence_custody_events USING btree (evidence_record_id, occurred_at DESC, custody_event_id DESC);

CREATE UNIQUE INDEX evidence_incident_record_unique_idx ON public.evidence USING btree (incident_id, record_id);

CREATE INDEX evidence_incident_sort_idx ON public.evidence USING btree (incident_id, requested_at DESC, record_id);

CREATE INDEX evidence_object_blob_id_fk_idx ON public.evidence USING btree (object_blob_id);

CREATE INDEX evidence_object_blob_id_scope_idx ON public.evidence (incident_id, object_blob_id);

CREATE UNIQUE INDEX evidence_object_blob_unique_idx ON public.evidence USING btree (object_blob_id) WHERE (object_blob_id IS NOT NULL);

CREATE INDEX evidence_object_upload_leases_expiry_idx ON public.evidence_object_upload_leases USING btree (expires_at, lease_id) WHERE (lease_state = 'issued'::text);

CREATE INDEX evidence_object_upload_leases_incident_id_fk_idx ON public.evidence_object_upload_leases USING btree (incident_id);

CREATE INDEX evidence_object_upload_leases_issuing_session_id_fk_idx ON public.evidence_object_upload_leases USING btree (issuing_session_id);

CREATE INDEX evidence_object_upload_leases_issuing_user_id_fk_idx ON public.evidence_object_upload_leases USING btree (issuing_user_id);

CREATE INDEX evidence_object_upload_leases_object_blob_id_scope_idx ON public.evidence_object_upload_leases (incident_id, object_blob_id);

CREATE INDEX evidence_source_party_ref_lookup_idx ON public.evidence USING btree (incident_id, source_party_id) WHERE (source_party_id IS NOT NULL);

CREATE INDEX object_blobs_created_by_user_id_fk_idx ON public.object_blobs USING btree (created_by_user_id);

CREATE INDEX object_blobs_incident_state_idx ON public.object_blobs USING btree (incident_id, upload_state, created_at DESC);

ALTER TABLE public.evidence ADD CONSTRAINT evidence_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence ADD CONSTRAINT evidence_object_blob_fkey FOREIGN KEY (incident_id, object_blob_id) REFERENCES public.object_blobs(incident_id, object_blob_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.evidence ADD CONSTRAINT evidence_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_access_handles ADD CONSTRAINT evidence_access_handles_evidence_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.evidence(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_access_handles ADD CONSTRAINT evidence_access_handles_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_access_handles ADD CONSTRAINT evidence_access_handles_issued_by_user_id_fkey FOREIGN KEY (issued_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.evidence_access_handles ADD CONSTRAINT evidence_access_handles_object_blob_id_fkey FOREIGN KEY (incident_id, object_blob_id) REFERENCES public.object_blobs(incident_id, object_blob_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_blob_cleanup_claims ADD CONSTRAINT evidence_blob_cleanup_claims_blob_fkey FOREIGN KEY (object_blob_id) REFERENCES public.object_blobs(object_blob_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_blob_verifications ADD CONSTRAINT evidence_blob_verifications_blob_fkey FOREIGN KEY (object_blob_id) REFERENCES public.object_blobs(object_blob_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_blob_verifications ADD CONSTRAINT evidence_blob_verifications_lease_fkey FOREIGN KEY (upload_lease_id) REFERENCES public.evidence_object_upload_leases(lease_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_custody_events ADD CONSTRAINT evidence_custody_events_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.evidence_custody_events ADD CONSTRAINT evidence_custody_events_evidence_record_id_fkey FOREIGN KEY (incident_id, evidence_record_id) REFERENCES public.evidence(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_custody_events ADD CONSTRAINT evidence_custody_events_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_object_upload_leases ADD CONSTRAINT evidence_object_upload_leases_blob_fkey FOREIGN KEY (incident_id, object_blob_id) REFERENCES public.object_blobs(incident_id, object_blob_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_object_upload_leases ADD CONSTRAINT evidence_object_upload_leases_incident_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_object_upload_leases ADD CONSTRAINT evidence_object_upload_leases_session_fkey FOREIGN KEY (issuing_session_id) REFERENCES public.user_sessions(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_object_upload_leases ADD CONSTRAINT evidence_object_upload_leases_user_fkey FOREIGN KEY (issuing_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.object_blobs ADD CONSTRAINT object_blobs_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.object_blobs ADD CONSTRAINT object_blobs_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

CREATE TRIGGER evidence_reject_cleanup_claimed_association BEFORE INSERT OR UPDATE OF object_blob_id ON public.evidence FOR EACH ROW EXECUTE FUNCTION public.evidence_reject_cleanup_claimed_association();

CREATE TRIGGER evidence_reject_cleanup_claimed_blob_update BEFORE UPDATE ON public.object_blobs FOR EACH ROW EXECUTE FUNCTION public.evidence_reject_cleanup_claimed_blob_update();

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER evidence_reject_cleanup_claimed_blob_update ON public.object_blobs;
DROP TRIGGER evidence_reject_cleanup_claimed_association ON public.evidence;
ALTER TABLE public.object_blobs DROP CONSTRAINT object_blobs_incident_id_fkey;
ALTER TABLE public.object_blobs DROP CONSTRAINT object_blobs_created_by_user_id_fkey;
ALTER TABLE public.evidence_object_upload_leases DROP CONSTRAINT evidence_object_upload_leases_user_fkey;
ALTER TABLE public.evidence_object_upload_leases DROP CONSTRAINT evidence_object_upload_leases_session_fkey;
ALTER TABLE public.evidence_object_upload_leases DROP CONSTRAINT evidence_object_upload_leases_incident_fkey;
ALTER TABLE public.evidence_object_upload_leases DROP CONSTRAINT evidence_object_upload_leases_blob_fkey;
ALTER TABLE public.evidence_custody_events DROP CONSTRAINT evidence_custody_events_incident_id_fkey;
ALTER TABLE public.evidence_custody_events DROP CONSTRAINT evidence_custody_events_evidence_record_id_fkey;
ALTER TABLE public.evidence_custody_events DROP CONSTRAINT evidence_custody_events_actor_user_id_fkey;
ALTER TABLE public.evidence_blob_verifications DROP CONSTRAINT evidence_blob_verifications_lease_fkey;
ALTER TABLE public.evidence_blob_verifications DROP CONSTRAINT evidence_blob_verifications_blob_fkey;
ALTER TABLE public.evidence_blob_cleanup_claims DROP CONSTRAINT evidence_blob_cleanup_claims_blob_fkey;
ALTER TABLE public.evidence_access_handles DROP CONSTRAINT evidence_access_handles_object_blob_id_fkey;
ALTER TABLE public.evidence_access_handles DROP CONSTRAINT evidence_access_handles_issued_by_user_id_fkey;
ALTER TABLE public.evidence_access_handles DROP CONSTRAINT evidence_access_handles_incident_id_fkey;
ALTER TABLE public.evidence_access_handles DROP CONSTRAINT evidence_access_handles_evidence_fkey;
ALTER TABLE public.evidence DROP CONSTRAINT evidence_record_envelope_fkey;
ALTER TABLE public.evidence DROP CONSTRAINT evidence_object_blob_fkey;
ALTER TABLE public.evidence DROP CONSTRAINT evidence_incident_id_fkey;
DROP FUNCTION public.evidence_reject_cleanup_claimed_blob_update();
DROP FUNCTION public.evidence_reject_cleanup_claimed_association();
DROP TABLE public.object_blobs;
DROP TABLE public.evidence_object_upload_leases;
DROP TABLE public.evidence_custody_events;
DROP TABLE public.evidence_blob_verifications;
DROP TABLE public.evidence_blob_cleanup_claims;
DROP TABLE public.evidence_access_handles;
DROP TABLE public.evidence;

-- +goose Up
-- The old synchronous retry budget has no safe translation to retained Job
-- evidence. Reject used history before changing the schema.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.object_blobs IN SHARE MODE;
    IF EXISTS (SELECT 1 FROM public.object_blobs WHERE finalize_attempt_count <> 0 OR terminal_reason = 'finalize_retry_exhausted') THEN
        RAISE EXCEPTION 'Evidence verification requires fresh initialization for legacy finalization history' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.object_blobs DROP CONSTRAINT object_blobs_finalize_attempt_count_check;
ALTER TABLE public.object_blobs DROP COLUMN finalize_attempt_count;
ALTER TABLE public.object_blobs DROP CONSTRAINT object_blobs_terminal_reason_check;
ALTER TABLE public.object_blobs ADD CONSTRAINT object_blobs_terminal_reason_check CHECK
    (terminal_reason IS NULL OR terminal_reason IN ('pending_timeout','declared_size_mismatch','expected_sha256_mismatch'));
DROP TRIGGER evidence_reject_cleanup_claimed_blob_update ON public.object_blobs;
DROP FUNCTION public.evidence_reject_cleanup_claimed_blob_update();
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
CREATE TRIGGER evidence_reject_cleanup_claimed_blob_update
BEFORE UPDATE ON public.object_blobs
FOR EACH ROW EXECUTE FUNCTION public.evidence_reject_cleanup_claimed_blob_update();
-- An observation is published only with successful fenced Job completion.
-- Job expiry does not erase an already consumed source verification fact.
CREATE TABLE public.evidence_blob_verifications (
    object_blob_id uuid CONSTRAINT evidence_blob_verifications_pkey PRIMARY KEY
        CONSTRAINT evidence_blob_verifications_blob_fkey REFERENCES public.object_blobs(object_blob_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    upload_lease_id uuid NOT NULL CONSTRAINT evidence_blob_verifications_lease_fkey
        REFERENCES public.evidence_object_upload_leases(lease_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    job_id uuid NOT NULL,
    verified_at timestamptz NOT NULL,
    observed_size bigint NOT NULL CONSTRAINT evidence_blob_verifications_size_ck CHECK (observed_size >= 0),
    observed_content_type text NOT NULL CONSTRAINT evidence_blob_verifications_content_type_ck CHECK (length(observed_content_type) BETWEEN 1 AND 1024),
    observed_sha256_hex text NOT NULL CONSTRAINT evidence_blob_verifications_sha256_ck CHECK (observed_sha256_hex ~ '^[a-f0-9]{64}$')
);
CREATE INDEX evidence_blob_verifications_lease_fk_idx ON public.evidence_blob_verifications(upload_lease_id);
REVOKE USAGE ON TYPE public.evidence_blob_verifications FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.evidence_blob_verifications TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_blob_verifications TO cartulary_recovery;
-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'Evidence background verification cannot be downgraded; restore a matching application/data pair' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

-- +goose Up
CREATE TABLE public.evidence_blob_cleanup_claims (
    object_blob_id uuid
        CONSTRAINT evidence_blob_cleanup_claims_pkey PRIMARY KEY
        CONSTRAINT evidence_blob_cleanup_claims_blob_fkey
        REFERENCES public.object_blobs(object_blob_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    claim_token uuid NOT NULL
        CONSTRAINT evidence_blob_cleanup_claims_claim_token_key UNIQUE,
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
    CONSTRAINT evidence_blob_cleanup_claims_state_ck CHECK (
        claim_state IN ('claimed', 'retry_wait', 'completed')
    ),
    CONSTRAINT evidence_blob_cleanup_claims_attempt_count_ck CHECK (attempt_count >= 1),
    CONSTRAINT evidence_blob_cleanup_claims_failure_class_ck CHECK (
        last_failure_class IS NULL
        OR last_failure_class IN ('delete_failed', 'delete_timeout', 'state_changed')
    ),
    CONSTRAINT evidence_blob_cleanup_claims_time_ck CHECK (
        updated_at >= created_at
        AND (last_attempt_at IS NULL OR last_attempt_at >= claimed_at)
        AND (completed_at IS NULL OR completed_at >= claimed_at)
    ),
    CONSTRAINT evidence_blob_cleanup_claims_transition_ck CHECK (
        (claim_state = 'claimed'
            AND claim_expires_at IS NOT NULL
            AND next_attempt_at IS NULL
            AND completed_at IS NULL)
        OR (claim_state = 'retry_wait'
            AND claim_expires_at IS NULL
            AND next_attempt_at IS NOT NULL
            AND completed_at IS NULL
            AND last_failure_class IS NOT NULL)
        OR (claim_state = 'completed'
            AND claim_expires_at IS NULL
            AND next_attempt_at IS NULL
            AND completed_at IS NOT NULL)
    )
);

CREATE INDEX evidence_blob_cleanup_claims_due_idx
    ON public.evidence_blob_cleanup_claims (next_attempt_at, object_blob_id)
    WHERE claim_state = 'retry_wait';

CREATE INDEX evidence_blob_cleanup_claims_lease_expiry_idx
    ON public.evidence_blob_cleanup_claims (claim_expires_at, object_blob_id)
    WHERE claim_state = 'claimed';

CREATE INDEX evidence_blob_cleanup_claims_completed_idx
    ON public.evidence_blob_cleanup_claims (completed_at, object_blob_id)
    WHERE claim_state = 'completed';

REVOKE USAGE ON TYPE public.evidence_blob_cleanup_claims FROM PUBLIC;

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
        OR NEW.finalize_attempt_count IS DISTINCT FROM OLD.finalize_attempt_count
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

CREATE TRIGGER evidence_reject_cleanup_claimed_association
BEFORE INSERT OR UPDATE OF object_blob_id ON public.evidence
FOR EACH ROW EXECUTE FUNCTION public.evidence_reject_cleanup_claimed_association();

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.evidence_blob_cleanup_claims
    TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_blob_cleanup_claims
    TO cartulary_recovery;

-- +goose Down
DROP TRIGGER evidence_reject_cleanup_claimed_association ON public.evidence;
DROP FUNCTION public.evidence_reject_cleanup_claimed_association();
DROP TRIGGER evidence_reject_cleanup_claimed_blob_update ON public.object_blobs;
DROP FUNCTION public.evidence_reject_cleanup_claimed_blob_update();
DROP TABLE public.evidence_blob_cleanup_claims;

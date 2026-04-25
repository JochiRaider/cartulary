-- +goose Up
CREATE TABLE IF NOT EXISTS object_blobs (
    object_blob_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    storage_key text NOT NULL UNIQUE,
    upload_state text NOT NULL DEFAULT 'pending' CHECK (upload_state IN ('pending', 'available', 'failed', 'quarantined')),
    byte_size bigint NOT NULL CHECK (byte_size >= 0),
    filename_hint text,
    content_type_hint text,
    expected_sha256_hex text CHECK (expected_sha256_hex IS NULL OR expected_sha256_hex ~ '^[0-9a-f]{64}$'),
    observed_size bigint,
    observed_content_type text,
    observed_sha256_hex text CHECK (observed_sha256_hex IS NULL OR observed_sha256_hex ~ '^[0-9a-f]{64}$'),
    target_expires_at timestamptz NOT NULL,
    pending_expires_at timestamptz NOT NULL,
    finalized_at timestamptz,
    terminal_reason text CHECK (terminal_reason IS NULL OR terminal_reason IN ('pending_timeout', 'finalize_retry_exhausted', 'declared_size_mismatch', 'expected_sha256_mismatch')),
    failed_at timestamptz,
    finalize_attempt_count integer NOT NULL DEFAULT 0 CHECK (finalize_attempt_count >= 0),
    cleanup_due_at timestamptz,
    cleaned_up_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT object_blobs_terminal_state_ck CHECK (
        (upload_state = 'failed' AND terminal_reason IS NOT NULL AND failed_at IS NOT NULL)
        OR (upload_state <> 'failed' AND terminal_reason IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS object_blobs_incident_state_idx
    ON object_blobs (incident_id, upload_state, created_at DESC);

ALTER TABLE evidence
    ADD COLUMN IF NOT EXISTS object_blob_id uuid;

ALTER TABLE evidence
    ADD CONSTRAINT evidence_object_blob_fkey
    FOREIGN KEY (object_blob_id) REFERENCES object_blobs (object_blob_id);

CREATE INDEX IF NOT EXISTS evidence_object_blob_idx
    ON evidence (object_blob_id)
    WHERE object_blob_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS evidence_incident_record_unique_idx
    ON evidence (incident_id, record_id);

CREATE TABLE IF NOT EXISTS evidence_access_handles (
    handle_token text PRIMARY KEY,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    record_id uuid NOT NULL,
    object_blob_id uuid NOT NULL REFERENCES object_blobs (object_blob_id) ON DELETE CASCADE,
    issued_by_user_id uuid NOT NULL REFERENCES users (id),
    issuing_session_id uuid NOT NULL,
    handle_kind text NOT NULL CHECK (handle_kind IN ('preview', 'download')),
    media_class text NOT NULL,
    preview_kind text,
    disposition text NOT NULL CHECK (disposition IN ('inline', 'attachment')),
    filename text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
    sha256 text,
    evidence_lifecycle_state text NOT NULL,
    upload_state text NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT evidence_access_handles_evidence_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES evidence (incident_id, record_id) ON DELETE CASCADE,
    CONSTRAINT evidence_access_handles_preview_kind_ck CHECK (
        (handle_kind = 'preview' AND preview_kind IS NOT NULL)
        OR (handle_kind = 'download' AND preview_kind IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS evidence_access_handles_lookup_idx
    ON evidence_access_handles (incident_id, record_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS evidence_access_handles_lookup_idx;
DROP TABLE IF EXISTS evidence_access_handles;

DROP INDEX IF EXISTS evidence_object_blob_idx;
DROP INDEX IF EXISTS evidence_incident_record_unique_idx;
ALTER TABLE evidence
    DROP CONSTRAINT IF EXISTS evidence_object_blob_fkey;
ALTER TABLE evidence
    DROP COLUMN IF EXISTS object_blob_id;

DROP INDEX IF EXISTS object_blobs_incident_state_idx;
DROP TABLE IF EXISTS object_blobs;

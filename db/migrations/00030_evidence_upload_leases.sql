-- +goose Up
CREATE TABLE public.evidence_object_upload_leases (
    object_blob_id uuid
        CONSTRAINT evidence_object_upload_leases_pkey PRIMARY KEY
        CONSTRAINT evidence_object_upload_leases_blob_fkey
        REFERENCES public.object_blobs(object_blob_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    lease_id uuid NOT NULL
        CONSTRAINT evidence_object_upload_leases_lease_id_key UNIQUE,
    capability_hash bytea NOT NULL
        CONSTRAINT evidence_object_upload_leases_capability_hash_key UNIQUE,
    incident_id uuid NOT NULL
        CONSTRAINT evidence_object_upload_leases_incident_fkey
        REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    issuing_user_id uuid NOT NULL
        CONSTRAINT evidence_object_upload_leases_user_fkey
        REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    issuing_session_id uuid NOT NULL
        CONSTRAINT evidence_object_upload_leases_session_fkey
        REFERENCES public.user_sessions(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    issued_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    required_method text NOT NULL,
    required_headers jsonb NOT NULL,
    accepted_contract_sha256 bytea NOT NULL,
    lease_state text NOT NULL DEFAULT 'issued',
    claimed_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT evidence_object_upload_leases_method_ck CHECK (required_method = 'PUT'),
    CONSTRAINT evidence_object_upload_leases_headers_ck CHECK (jsonb_typeof(required_headers) = 'object'),
    CONSTRAINT evidence_object_upload_leases_contract_digest_ck CHECK (octet_length(accepted_contract_sha256) = 32),
    CONSTRAINT evidence_object_upload_leases_capability_digest_ck CHECK (octet_length(capability_hash) = 32),
    CONSTRAINT evidence_object_upload_leases_state_ck CHECK (lease_state IN ('issued', 'claimed', 'completed')),
    CONSTRAINT evidence_object_upload_leases_time_ck CHECK (
        expires_at > issued_at AND updated_at >= created_at
    ),
    CONSTRAINT evidence_object_upload_leases_transition_ck CHECK (
        (lease_state = 'issued' AND claimed_at IS NULL AND completed_at IS NULL)
        OR (lease_state = 'claimed' AND claimed_at IS NOT NULL AND completed_at IS NULL)
        OR (lease_state = 'completed' AND claimed_at IS NOT NULL AND completed_at IS NOT NULL)
    )
);

CREATE INDEX evidence_object_upload_leases_expiry_idx
    ON public.evidence_object_upload_leases (expires_at, lease_id)
    WHERE lease_state = 'issued';

CREATE INDEX evidence_object_upload_leases_incident_id_fk_idx
    ON public.evidence_object_upload_leases (incident_id);

CREATE INDEX evidence_object_upload_leases_issuing_user_id_fk_idx
    ON public.evidence_object_upload_leases (issuing_user_id);

CREATE INDEX evidence_object_upload_leases_issuing_session_id_fk_idx
    ON public.evidence_object_upload_leases (issuing_session_id);

REVOKE USAGE ON TYPE public.evidence_object_upload_leases FROM PUBLIC;

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.evidence_object_upload_leases
    TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_object_upload_leases
    TO cartulary_recovery;

-- +goose Down
DROP TABLE public.evidence_object_upload_leases;

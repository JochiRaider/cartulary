-- +goose Up
CREATE TABLE public.collaboration_event_intents (
    intent_id uuid DEFAULT gen_random_uuid()
        CONSTRAINT collaboration_event_intents_pkey PRIMARY KEY,
    intent_key text NOT NULL
        CONSTRAINT collaboration_event_intents_intent_key_key UNIQUE,
    incident_id uuid NOT NULL
        CONSTRAINT collaboration_event_intents_incident_id_fkey
        REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    event_family text NOT NULL,
    canonical_payload jsonb NOT NULL,
    source_change_set_id uuid,
    source_record_id uuid,
    source_row_version bigint,
    source_identity text NOT NULL,
    mutation_ordinal integer NOT NULL,
    dispatch_state text NOT NULL DEFAULT 'pending',
    sequenced_event_id uuid,
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamp with time zone NOT NULL,
    sequenced_at timestamp with time zone,
    last_error_code text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_event_intents_key_ck CHECK (
        octet_length(intent_key) BETWEEN 1 AND 512 AND intent_key !~ '[[:cntrl:]]'
    ),
    CONSTRAINT collaboration_event_intents_family_ck CHECK (
        event_family IN ('record_changed', 'job_progress', 'extension_resource_changed')
    ),
    CONSTRAINT collaboration_event_intents_payload_ck CHECK (
        jsonb_typeof(canonical_payload) = 'object'
    ),
    CONSTRAINT collaboration_event_intents_payload_size_ck CHECK (
        octet_length(canonical_payload::text) <= 262144
    ),
    CONSTRAINT collaboration_event_intents_source_identity_ck CHECK (
        octet_length(source_identity) BETWEEN 1 AND 512 AND source_identity !~ '[[:cntrl:]]'
    ),
    CONSTRAINT collaboration_event_intents_row_version_ck CHECK (
        source_row_version IS NULL OR source_row_version >= 1
    ),
    CONSTRAINT collaboration_event_intents_ordinal_ck CHECK (mutation_ordinal >= 0),
    CONSTRAINT collaboration_event_intents_dispatch_state_ck CHECK (
        dispatch_state IN ('pending', 'sequenced')
    ),
    CONSTRAINT collaboration_event_intents_attempt_ck CHECK (
        attempt_count BETWEEN 0 AND 2147483647
    ),
    CONSTRAINT collaboration_event_intents_sequence_ck CHECK (
        (dispatch_state = 'pending' AND sequenced_event_id IS NULL AND sequenced_at IS NULL)
        OR (dispatch_state = 'sequenced' AND sequenced_event_id IS NOT NULL AND sequenced_at IS NOT NULL)
    ),
    CONSTRAINT collaboration_event_intents_time_ck CHECK (
        updated_at >= created_at AND next_attempt_at >= created_at
    ),
    CONSTRAINT collaboration_event_intents_error_ck CHECK (
        last_error_code IS NULL OR last_error_code ~ '^[a-z][a-z0-9_.-]{0,127}$'
    )
);

CREATE INDEX idx_collaboration_event_intents_dispatch
    ON public.collaboration_event_intents (dispatch_state, next_attempt_at, created_at, intent_id)
    WHERE dispatch_state = 'pending';

CREATE TABLE public.collaboration_incident_stream_cursors (
    incident_id uuid CONSTRAINT collaboration_incident_stream_cursors_pkey PRIMARY KEY
        CONSTRAINT collaboration_incident_stream_cursors_incident_id_fkey
        REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    high_water_stream_seq bigint NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    failure_count integer NOT NULL DEFAULT 0,
    quarantined_at timestamp with time zone,
    quarantine_reason text,
    CONSTRAINT collaboration_incident_stream_cursors_high_water_ck CHECK (
        high_water_stream_seq >= 0
    ),
    CONSTRAINT collaboration_incident_stream_cursors_failure_count_ck CHECK (
        failure_count BETWEEN 0 AND 12
    ),
    CONSTRAINT collaboration_incident_stream_cursors_quarantine_ck CHECK (
        (quarantined_at IS NULL AND quarantine_reason IS NULL)
        OR (quarantined_at IS NOT NULL AND quarantine_reason ~ '^[a-z][a-z0-9_.-]{0,127}$')
    )
);

CREATE INDEX idx_collaboration_incident_stream_quarantine
    ON public.collaboration_incident_stream_cursors (quarantined_at, incident_id)
    WHERE quarantined_at IS NOT NULL;

CREATE TABLE public.collaboration_replay_events (
    event_id uuid CONSTRAINT collaboration_replay_events_pkey PRIMARY KEY,
    incident_id uuid NOT NULL
        CONSTRAINT collaboration_replay_events_incident_id_fkey
        REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    stream_seq bigint NOT NULL,
    intent_key text NOT NULL
        CONSTRAINT collaboration_replay_events_intent_key_key UNIQUE,
    event_family text NOT NULL,
    canonical_payload jsonb NOT NULL,
    emitted_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_replay_events_incident_sequence_uq UNIQUE (incident_id, stream_seq),
    CONSTRAINT collaboration_replay_events_sequence_ck CHECK (stream_seq >= 1),
    CONSTRAINT collaboration_replay_events_family_ck CHECK (
        event_family IN ('record_changed', 'job_progress', 'extension_resource_changed')
    ),
    CONSTRAINT collaboration_replay_events_payload_ck CHECK (
        jsonb_typeof(canonical_payload) = 'object'
    ),
    CONSTRAINT collaboration_replay_events_payload_size_ck CHECK (
        octet_length(canonical_payload::text) <= 262144
    )
);

CREATE INDEX idx_collaboration_replay_events_retention
    ON public.collaboration_replay_events (incident_id, emitted_at, stream_seq);

CREATE TABLE public.collaboration_resume_tokens (
    token_hash bytea CONSTRAINT collaboration_resume_tokens_pkey PRIMARY KEY,
    session_id uuid NOT NULL
        CONSTRAINT collaboration_resume_tokens_session_id_fkey
        REFERENCES public.user_sessions(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    incident_id uuid NOT NULL
        CONSTRAINT collaboration_resume_tokens_incident_id_fkey
        REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    client_instance_id text NOT NULL,
    issued_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_resume_tokens_hash_ck CHECK (octet_length(token_hash) = 32),
    CONSTRAINT collaboration_resume_tokens_client_ck CHECK (
        octet_length(client_instance_id) BETWEEN 1 AND 256
        AND client_instance_id !~ '[[:cntrl:]]'
    ),
    CONSTRAINT collaboration_resume_tokens_expiry_ck CHECK (expires_at > issued_at)
);

CREATE INDEX idx_collaboration_resume_tokens_expiry
    ON public.collaboration_resume_tokens (expires_at);

REVOKE CREATE ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC;
REVOKE EXECUTE ON ALL FUNCTIONS IN SCHEMA public FROM PUBLIC;
-- +goose StatementBegin
DO $cartulary_owned_type_acl$
DECLARE
    managed_type pg_catalog.regtype;
BEGIN
    FOR managed_type IN
        SELECT candidate.oid::pg_catalog.regtype
          FROM pg_catalog.pg_type AS candidate
          JOIN pg_catalog.pg_namespace AS namespace
            ON namespace.oid = candidate.typnamespace
          JOIN pg_catalog.pg_roles AS owner_role
            ON owner_role.oid = candidate.typowner
         WHERE namespace.nspname = 'public'
           AND owner_role.rolname = 'cartulary_schema_owner'
           AND candidate.typelem = 0
         ORDER BY candidate.oid::pg_catalog.regtype::text
    LOOP
        EXECUTE pg_catalog.format(
            'REVOKE USAGE ON TYPE %s FROM PUBLIC',
            managed_type
        );
    END LOOP;
END
$cartulary_owned_type_acl$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO cartulary_runtime, cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO cartulary_runtime;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON ALL TABLES IN SCHEMA public TO cartulary_recovery;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO cartulary_recovery;

REVOKE ALL ON TABLE public.active_record_links_v1, public.active_record_tags_v1
    FROM cartulary_runtime, cartulary_recovery;
GRANT SELECT ON TABLE public.active_record_links_v1, public.active_record_tags_v1
    TO cartulary_runtime, cartulary_recovery;

REVOKE ALL ON TABLE public.goose_db_version FROM cartulary_runtime, cartulary_recovery;
GRANT SELECT ON TABLE public.goose_db_version TO cartulary_runtime, cartulary_recovery;
REVOKE INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.schema_migration_lineage
    FROM cartulary_runtime, cartulary_recovery;
GRANT SELECT ON TABLE public.schema_migration_lineage TO cartulary_runtime, cartulary_recovery;

REVOKE ALL ON TABLE public.backup_sets, public.restore_verification_runs,
    public.operator_recovery_journal FROM cartulary_runtime;
GRANT SELECT, INSERT ON TABLE public.deployment_admin_audit_events,
    public.administrative_audit_projections TO cartulary_runtime;
REVOKE UPDATE, DELETE, TRUNCATE ON TABLE public.deployment_admin_audit_events,
    public.administrative_audit_projections FROM cartulary_runtime;

GRANT EXECUTE ON FUNCTION
    public.cartulary_confidence_band(integer),
    public.change_set_mutations_history_ids_are_canonical(uuid[]),
    public.enforce_indicator_support_ref_incident(),
    public.indicator_support_refs_are_valid(jsonb),
    public.network_flow_reject_immutable_update(),
    public.reject_administrative_audit_mutation(),
    public.revisions_incident_bundle_sequence_begin_v1(),
    public.revisions_incident_bundle_sequence_finish_v1(bigint),
    public.sync_indicator_active_identity_from_indicator(),
    public.sync_indicator_active_identity_from_record()
TO cartulary_runtime, cartulary_recovery;
GRANT EXECUTE ON FUNCTION public.rebuild_indicator_active_identities() TO cartulary_recovery;
GRANT EXECUTE ON FUNCTION public.indicator_active_identities_are_valid() TO cartulary_recovery;

ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public
    REVOKE ALL ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public
    REVOKE ALL ON SEQUENCES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner
    REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner
    REVOKE USAGE ON TYPES FROM PUBLIC;
CREATE INDEX collaboration_event_intents_incident_id_fk_idx ON public.collaboration_event_intents (incident_id);
CREATE INDEX collaboration_resume_tokens_incident_id_fk_idx ON public.collaboration_resume_tokens (incident_id);
CREATE INDEX collaboration_resume_tokens_session_id_fk_idx ON public.collaboration_resume_tokens (session_id);

-- +goose Down
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM cartulary_runtime, cartulary_recovery;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM cartulary_runtime, cartulary_recovery;
REVOKE EXECUTE ON ALL FUNCTIONS IN SCHEMA public FROM cartulary_runtime, cartulary_recovery;

DROP TABLE public.collaboration_resume_tokens;
DROP TABLE public.collaboration_replay_events;
DROP TABLE public.collaboration_incident_stream_cursors;
DROP TABLE public.collaboration_event_intents;

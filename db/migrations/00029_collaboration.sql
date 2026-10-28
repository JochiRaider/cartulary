-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: collaboration.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.collaboration_event_intents (
    intent_id uuid DEFAULT gen_random_uuid() NOT NULL,
    intent_key text NOT NULL,
    incident_id uuid NOT NULL,
    event_family text NOT NULL,
    canonical_payload jsonb NOT NULL,
    source_change_set_id uuid,
    source_record_id uuid,
    source_row_version bigint,
    source_identity text NOT NULL,
    mutation_ordinal integer NOT NULL,
    dispatch_state text DEFAULT 'pending'::text NOT NULL,
    sequenced_event_id uuid,
    attempt_count integer DEFAULT 0 NOT NULL,
    next_attempt_at timestamp with time zone NOT NULL,
    sequenced_at timestamp with time zone,
    last_error_code text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_event_intents_attempt_ck CHECK (attempt_count >= 0 AND attempt_count <= 2147483647),
    CONSTRAINT collaboration_event_intents_dispatch_state_ck CHECK (dispatch_state = ANY (ARRAY['pending'::text, 'sequenced'::text])),
    CONSTRAINT collaboration_event_intents_error_ck CHECK (last_error_code IS NULL OR last_error_code ~ '^[a-z][a-z0-9_.-]{0,127}$'::text),
    CONSTRAINT collaboration_event_intents_family_ck CHECK (event_family = ANY (ARRAY['record_changed'::text, 'job_progress'::text, 'extension_resource_changed'::text])),
    CONSTRAINT collaboration_event_intents_intent_key_key UNIQUE (intent_key),
    CONSTRAINT collaboration_event_intents_key_ck CHECK (octet_length(intent_key) >= 1 AND octet_length(intent_key) <= 512 AND intent_key !~ '[[:cntrl:]]'::text),
    CONSTRAINT collaboration_event_intents_ordinal_ck CHECK (mutation_ordinal >= 0),
    CONSTRAINT collaboration_event_intents_payload_ck CHECK (jsonb_typeof(canonical_payload) = 'object'::text),
    CONSTRAINT collaboration_event_intents_payload_size_ck CHECK (octet_length(canonical_payload::text) <= 262144),
    CONSTRAINT collaboration_event_intents_pkey PRIMARY KEY (intent_id),
    CONSTRAINT collaboration_event_intents_row_version_ck CHECK (source_row_version IS NULL OR source_row_version >= 1),
    CONSTRAINT collaboration_event_intents_sequence_ck CHECK (dispatch_state = 'pending'::text AND sequenced_event_id IS NULL AND sequenced_at IS NULL OR dispatch_state = 'sequenced'::text AND sequenced_event_id IS NOT NULL AND sequenced_at IS NOT NULL),
    CONSTRAINT collaboration_event_intents_source_identity_ck CHECK (octet_length(source_identity) >= 1 AND octet_length(source_identity) <= 512 AND source_identity !~ '[[:cntrl:]]'::text),
    CONSTRAINT collaboration_event_intents_time_ck CHECK (updated_at >= created_at AND next_attempt_at >= created_at)
);

REVOKE ALL ON TABLE public.collaboration_event_intents FROM PUBLIC;

REVOKE ALL ON TYPE public.collaboration_event_intents FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.collaboration_event_intents TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.collaboration_event_intents TO cartulary_recovery;

CREATE TABLE public.collaboration_incident_stream_cursors (
    incident_id uuid NOT NULL,
    high_water_stream_seq bigint NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    failure_count integer DEFAULT 0 NOT NULL,
    quarantined_at timestamp with time zone,
    quarantine_reason text,
    CONSTRAINT collaboration_incident_stream_cursors_failure_count_ck CHECK (failure_count >= 0 AND failure_count <= 12),
    CONSTRAINT collaboration_incident_stream_cursors_high_water_ck CHECK (high_water_stream_seq >= 0),
    CONSTRAINT collaboration_incident_stream_cursors_pkey PRIMARY KEY (incident_id),
    CONSTRAINT collaboration_incident_stream_cursors_quarantine_ck CHECK (quarantined_at IS NULL AND quarantine_reason IS NULL OR quarantined_at IS NOT NULL AND quarantine_reason ~ '^[a-z][a-z0-9_.-]{0,127}$'::text)
);

REVOKE ALL ON TABLE public.collaboration_incident_stream_cursors FROM PUBLIC;

REVOKE ALL ON TYPE public.collaboration_incident_stream_cursors FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.collaboration_incident_stream_cursors TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.collaboration_incident_stream_cursors TO cartulary_recovery;

CREATE TABLE public.collaboration_replay_events (
    event_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    stream_seq bigint NOT NULL,
    intent_key text NOT NULL,
    event_family text NOT NULL,
    canonical_payload jsonb NOT NULL,
    emitted_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_replay_events_family_ck CHECK (event_family = ANY (ARRAY['record_changed'::text, 'job_progress'::text, 'extension_resource_changed'::text])),
    CONSTRAINT collaboration_replay_events_incident_sequence_uq UNIQUE (incident_id, stream_seq),
    CONSTRAINT collaboration_replay_events_intent_key_key UNIQUE (intent_key),
    CONSTRAINT collaboration_replay_events_payload_ck CHECK (jsonb_typeof(canonical_payload) = 'object'::text),
    CONSTRAINT collaboration_replay_events_payload_size_ck CHECK (octet_length(canonical_payload::text) <= 262144),
    CONSTRAINT collaboration_replay_events_pkey PRIMARY KEY (event_id),
    CONSTRAINT collaboration_replay_events_sequence_ck CHECK (stream_seq >= 1)
);

REVOKE ALL ON TABLE public.collaboration_replay_events FROM PUBLIC;

REVOKE ALL ON TYPE public.collaboration_replay_events FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.collaboration_replay_events TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.collaboration_replay_events TO cartulary_recovery;

CREATE TABLE public.collaboration_resume_tokens (
    token_hash bytea NOT NULL,
    session_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    client_instance_id text NOT NULL,
    issued_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_resume_tokens_client_ck CHECK (octet_length(client_instance_id) >= 1 AND octet_length(client_instance_id) <= 256 AND client_instance_id !~ '[[:cntrl:]]'::text),
    CONSTRAINT collaboration_resume_tokens_expiry_ck CHECK (expires_at > issued_at),
    CONSTRAINT collaboration_resume_tokens_hash_ck CHECK (octet_length(token_hash) = 32),
    CONSTRAINT collaboration_resume_tokens_pkey PRIMARY KEY (token_hash)
);

REVOKE ALL ON TABLE public.collaboration_resume_tokens FROM PUBLIC;

REVOKE ALL ON TYPE public.collaboration_resume_tokens FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.collaboration_resume_tokens TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.collaboration_resume_tokens TO cartulary_recovery;

CREATE INDEX collaboration_event_intents_incident_id_fk_idx ON public.collaboration_event_intents USING btree (incident_id);

CREATE INDEX collaboration_resume_tokens_incident_id_fk_idx ON public.collaboration_resume_tokens USING btree (incident_id);

CREATE INDEX collaboration_resume_tokens_session_id_fk_idx ON public.collaboration_resume_tokens USING btree (session_id);

CREATE INDEX idx_collaboration_event_intents_dispatch ON public.collaboration_event_intents USING btree (dispatch_state, next_attempt_at, created_at, intent_id) WHERE (dispatch_state = 'pending'::text);

CREATE INDEX idx_collaboration_incident_stream_quarantine ON public.collaboration_incident_stream_cursors USING btree (quarantined_at, incident_id) WHERE (quarantined_at IS NOT NULL);

CREATE INDEX idx_collaboration_replay_events_retention ON public.collaboration_replay_events USING btree (incident_id, emitted_at, stream_seq);

CREATE INDEX idx_collaboration_resume_tokens_expiry ON public.collaboration_resume_tokens USING btree (expires_at);

ALTER TABLE public.collaboration_event_intents ADD CONSTRAINT collaboration_event_intents_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.collaboration_incident_stream_cursors ADD CONSTRAINT collaboration_incident_stream_cursors_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.collaboration_replay_events ADD CONSTRAINT collaboration_replay_events_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.collaboration_resume_tokens ADD CONSTRAINT collaboration_resume_tokens_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.collaboration_resume_tokens ADD CONSTRAINT collaboration_resume_tokens_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.user_sessions(id) ON DELETE CASCADE ON UPDATE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.collaboration_resume_tokens DROP CONSTRAINT collaboration_resume_tokens_session_id_fkey;
ALTER TABLE public.collaboration_resume_tokens DROP CONSTRAINT collaboration_resume_tokens_incident_id_fkey;
ALTER TABLE public.collaboration_replay_events DROP CONSTRAINT collaboration_replay_events_incident_id_fkey;
ALTER TABLE public.collaboration_incident_stream_cursors DROP CONSTRAINT collaboration_incident_stream_cursors_incident_id_fkey;
ALTER TABLE public.collaboration_event_intents DROP CONSTRAINT collaboration_event_intents_incident_id_fkey;
DROP TABLE public.collaboration_resume_tokens;
DROP TABLE public.collaboration_replay_events;
DROP TABLE public.collaboration_incident_stream_cursors;
DROP TABLE public.collaboration_event_intents;

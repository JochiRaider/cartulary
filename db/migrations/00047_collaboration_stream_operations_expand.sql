-- +goose Up
ALTER TABLE public.collaboration_incident_stream_cursors
    ADD COLUMN failure_count integer NOT NULL DEFAULT 0,
    ADD COLUMN quarantined_at timestamp with time zone,
    ADD COLUMN quarantine_reason text,
    ADD CONSTRAINT collaboration_incident_stream_cursors_failure_count_ck
        CHECK (failure_count BETWEEN 0 AND 12),
    ADD CONSTRAINT collaboration_incident_stream_cursors_quarantine_ck
        CHECK (
            (quarantined_at IS NULL AND quarantine_reason IS NULL)
            OR
            (
                quarantined_at IS NOT NULL
                AND quarantine_reason ~ '^[a-z][a-z0-9_.-]{0,127}$'
            )
        );

ALTER TABLE public.collaboration_event_intents
    ADD CONSTRAINT collaboration_event_intents_payload_size_ck
        CHECK (octet_length(canonical_payload::text) <= 262144);

ALTER TABLE public.collaboration_replay_events
    ADD CONSTRAINT collaboration_replay_events_payload_size_ck
        CHECK (octet_length(canonical_payload::text) <= 262144);

CREATE INDEX idx_collaboration_incident_stream_quarantine
    ON public.collaboration_incident_stream_cursors (quarantined_at, incident_id)
    WHERE quarantined_at IS NOT NULL;

-- +goose Down
DROP INDEX public.idx_collaboration_incident_stream_quarantine;

ALTER TABLE public.collaboration_replay_events
    DROP CONSTRAINT collaboration_replay_events_payload_size_ck;

ALTER TABLE public.collaboration_event_intents
    DROP CONSTRAINT collaboration_event_intents_payload_size_ck;

ALTER TABLE public.collaboration_incident_stream_cursors
    DROP CONSTRAINT collaboration_incident_stream_cursors_quarantine_ck,
    DROP CONSTRAINT collaboration_incident_stream_cursors_failure_count_ck,
    DROP COLUMN quarantine_reason,
    DROP COLUMN quarantined_at,
    DROP COLUMN failure_count;

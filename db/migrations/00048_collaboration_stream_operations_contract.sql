-- +goose Up
DROP INDEX public.idx_collaboration_event_intents_lease;
DROP INDEX public.idx_collaboration_event_intents_dispatch;

ALTER TABLE public.collaboration_event_intents
    DROP CONSTRAINT collaboration_event_intents_delivery_ck,
    DROP CONSTRAINT collaboration_event_intents_lease_ck,
    DROP CONSTRAINT collaboration_event_intents_sequence_ck,
    DROP COLUMN delivered_at,
    DROP COLUMN lease_owner,
    DROP COLUMN lease_expires_at,
    ADD CONSTRAINT collaboration_event_intents_sequence_ck
        CHECK (
            (
                dispatch_state = 'pending'
                AND sequenced_event_id IS NULL
                AND sequenced_at IS NULL
            )
            OR
            (
                dispatch_state = 'sequenced'
                AND sequenced_event_id IS NOT NULL
                AND sequenced_at IS NOT NULL
            )
        );

CREATE INDEX idx_collaboration_event_intents_dispatch
    ON public.collaboration_event_intents (
        dispatch_state,
        next_attempt_at,
        created_at,
        intent_id
    )
    WHERE dispatch_state = 'pending';

-- +goose Down
DROP INDEX public.idx_collaboration_event_intents_dispatch;

ALTER TABLE public.collaboration_event_intents
    DROP CONSTRAINT collaboration_event_intents_sequence_ck,
    ADD COLUMN lease_owner uuid,
    ADD COLUMN lease_expires_at timestamp with time zone,
    ADD COLUMN delivered_at timestamp with time zone,
    ADD CONSTRAINT collaboration_event_intents_lease_ck
        CHECK ((lease_owner IS NULL) = (lease_expires_at IS NULL)),
    ADD CONSTRAINT collaboration_event_intents_sequence_ck
        CHECK (
            (
                dispatch_state = 'pending'
                AND sequenced_event_id IS NULL
                AND sequenced_at IS NULL
                AND delivered_at IS NULL
            )
            OR
            (
                dispatch_state = 'sequenced'
                AND sequenced_event_id IS NOT NULL
                AND sequenced_at IS NOT NULL
            )
        ),
    ADD CONSTRAINT collaboration_event_intents_delivery_ck
        CHECK (delivered_at IS NULL OR delivered_at >= sequenced_at);

CREATE INDEX idx_collaboration_event_intents_dispatch
    ON public.collaboration_event_intents (
        dispatch_state,
        next_attempt_at,
        created_at,
        intent_id
    )
    WHERE delivered_at IS NULL;

CREATE INDEX idx_collaboration_event_intents_lease
    ON public.collaboration_event_intents (lease_expires_at, intent_id)
    WHERE lease_owner IS NOT NULL;

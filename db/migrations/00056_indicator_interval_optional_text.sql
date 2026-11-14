-- +goose Up
ALTER TABLE public.indicator_state_intervals
    DROP CONSTRAINT indicator_state_intervals_required_text_ck;

-- +goose Down
-- Refuse incompatible downgrade transactionally; never rewrite retained text.
ALTER TABLE public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_required_text_ck CHECK (
        (rationale IS NULL OR octet_length(rationale) > 0)
        AND (assessor IS NULL OR octet_length(assessor) > 0)
    );

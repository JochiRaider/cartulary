-- +goose Up

ALTER TABLE public.indicator_observations
    ADD COLUMN deleted_at timestamp with time zone,
    ADD COLUMN deleted_by_user_id uuid,
    ADD CONSTRAINT indicator_observations_tombstone_pair_ck
        CHECK ((deleted_at IS NULL) = (deleted_by_user_id IS NULL)),
    ADD CONSTRAINT indicator_observations_deleted_by_user_id_fkey
        FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id);

ALTER TABLE public.indicator_state_intervals
    ADD COLUMN deleted_at timestamp with time zone,
    ADD COLUMN deleted_by_user_id uuid,
    ADD CONSTRAINT indicator_state_intervals_tombstone_pair_ck
        CHECK ((deleted_at IS NULL) = (deleted_by_user_id IS NULL)),
    ADD CONSTRAINT indicator_state_intervals_deleted_by_user_id_fkey
        FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id);

DROP INDEX public.indicator_observations_candidate_lookup_idx;
DROP INDEX public.indicator_observations_resolved_lookup_idx;
DROP INDEX public.indicator_observations_source_lookup_idx;
DROP INDEX public.indicator_state_intervals_indicator_lookup_idx;

CREATE INDEX indicator_observations_candidate_lookup_idx
    ON public.indicator_observations (incident_id, parsed_indicator_type, normalized_candidate, indicator_observation_id)
    WHERE normalized_candidate IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX indicator_observations_resolved_lookup_idx
    ON public.indicator_observations (incident_id, resolution_status, resolved_indicator_record_id, created_at)
    WHERE deleted_at IS NULL;

CREATE INDEX indicator_observations_source_lookup_idx
    ON public.indicator_observations (source_record_id, source_field_key, created_at, indicator_observation_id)
    WHERE deleted_at IS NULL;

CREATE INDEX indicator_state_intervals_indicator_lookup_idx
    ON public.indicator_state_intervals (incident_id, indicator_record_id, valid_from DESC, indicator_state_interval_id DESC)
    WHERE deleted_at IS NULL;

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.indicator_observations WHERE deleted_at IS NOT NULL)
       OR EXISTS (SELECT 1 FROM public.indicator_state_intervals WHERE deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'cannot remove indicator child rollback tombstones while retained tombstones exist';
    END IF;
END
$$;
-- +goose StatementEnd

DROP INDEX public.indicator_observations_candidate_lookup_idx;
DROP INDEX public.indicator_observations_resolved_lookup_idx;
DROP INDEX public.indicator_observations_source_lookup_idx;
DROP INDEX public.indicator_state_intervals_indicator_lookup_idx;

CREATE INDEX indicator_observations_candidate_lookup_idx
    ON public.indicator_observations (incident_id, parsed_indicator_type, normalized_candidate, indicator_observation_id)
    WHERE normalized_candidate IS NOT NULL;

CREATE INDEX indicator_observations_resolved_lookup_idx
    ON public.indicator_observations (incident_id, resolution_status, resolved_indicator_record_id, created_at);

CREATE INDEX indicator_observations_source_lookup_idx
    ON public.indicator_observations (source_record_id, source_field_key, created_at, indicator_observation_id);

CREATE INDEX indicator_state_intervals_indicator_lookup_idx
    ON public.indicator_state_intervals (incident_id, indicator_record_id, valid_from DESC, indicator_state_interval_id DESC);

ALTER TABLE public.indicator_observations
    DROP CONSTRAINT indicator_observations_deleted_by_user_id_fkey,
    DROP CONSTRAINT indicator_observations_tombstone_pair_ck,
    DROP COLUMN deleted_by_user_id,
    DROP COLUMN deleted_at;

ALTER TABLE public.indicator_state_intervals
    DROP CONSTRAINT indicator_state_intervals_deleted_by_user_id_fkey,
    DROP CONSTRAINT indicator_state_intervals_tombstone_pair_ck,
    DROP COLUMN deleted_by_user_id,
    DROP COLUMN deleted_at;

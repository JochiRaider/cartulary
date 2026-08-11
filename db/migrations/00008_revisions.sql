-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION public.revisions_incident_bundle_sequence_begin_v1()
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
    original_next bigint;
BEGIN
    SELECT sequence_state.last_value +
           CASE WHEN sequence_state.is_called THEN sequence_catalog.seqincrement ELSE 0 END
      INTO STRICT original_next
      FROM public.record_revisions_revision_id_seq AS sequence_state
      JOIN pg_catalog.pg_sequence AS sequence_catalog
        ON sequence_catalog.seqrelid =
           'public.record_revisions_revision_id_seq'::pg_catalog.regclass;

    IF original_next < 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '22003',
            MESSAGE = 'record revision sequence has an invalid next value';
    END IF;

    EXECUTE pg_catalog.format(
        'ALTER SEQUENCE public.record_revisions_revision_id_seq RESTART WITH %s',
        original_next
    );
    RETURN original_next;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.revisions_incident_bundle_sequence_finish_v1(
    original_next bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
    imported_max bigint;
    repaired_next bigint;
BEGIN
    IF original_next < 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '22003',
            MESSAGE = 'record revision sequence original next value is invalid';
    END IF;

    SELECT max(revision_id)
      INTO imported_max
      FROM public.record_revisions;

    IF imported_max = 9223372036854775807 THEN
        RAISE EXCEPTION USING
            ERRCODE = '22003',
            MESSAGE = 'record revision sequence is exhausted';
    END IF;

    repaired_next := greatest(
        original_next,
        CASE WHEN imported_max IS NULL THEN 1::bigint ELSE imported_max + 1 END
    );
    EXECUTE pg_catalog.format(
        'ALTER SEQUENCE public.record_revisions_revision_id_seq RESTART WITH %s',
        repaired_next
    );
    RETURN repaired_next;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.revisions_incident_bundle_sequence_begin_v1()
    FROM PUBLIC;
REVOKE ALL ON FUNCTION public.revisions_incident_bundle_sequence_finish_v1(bigint)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.revisions_incident_bundle_sequence_begin_v1()
    TO CURRENT_USER;
GRANT EXECUTE ON FUNCTION public.revisions_incident_bundle_sequence_finish_v1(bigint)
    TO CURRENT_USER;

-- +goose StatementBegin
CREATE FUNCTION public.change_set_mutations_history_ids_are_canonical(candidate uuid[])
RETURNS boolean
LANGUAGE sql
IMMUTABLE
STRICT
PARALLEL SAFE
SET search_path = pg_catalog, public
AS $$
    SELECT
        COALESCE(array_ndims(candidate), 1) = 1
        AND array_position(candidate, NULL::uuid) IS NULL
        AND array_position(candidate, '00000000-0000-0000-0000-000000000000'::uuid) IS NULL
        AND candidate = ARRAY(
            SELECT DISTINCT member
              FROM unnest(candidate) AS values_row(member)
             ORDER BY member
        );
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.change_set_mutations_history_ids_are_canonical(uuid[])
    FROM PUBLIC;

ALTER TABLE public.change_set_mutations
    ADD COLUMN history_record_ids uuid[] NOT NULL,
    ADD COLUMN history_entry_record_ids uuid[] NOT NULL,
    ADD CONSTRAINT change_set_mutations_history_record_ids_ck
        CHECK (public.change_set_mutations_history_ids_are_canonical(history_record_ids)),
    ADD CONSTRAINT change_set_mutations_history_entry_record_ids_ck
        CHECK (public.change_set_mutations_history_ids_are_canonical(history_entry_record_ids)),
    ADD CONSTRAINT change_set_mutations_history_entry_subset_ck
        CHECK (history_entry_record_ids <@ history_record_ids);

CREATE INDEX change_set_mutations_history_record_ids_idx
    ON public.change_set_mutations USING gin (history_record_ids);

-- Conflict facts are a narrow concurrency index derived from the explicit
-- live-change input. They are not retained record snapshots and therefore do
-- not duplicate source-owned history or projection rows.
CREATE TABLE public.record_revision_conflict_facts (
    revision_id bigint NOT NULL
        CONSTRAINT record_revision_conflict_facts_revision_id_fkey
        REFERENCES public.record_revisions(revision_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    field_key text NOT NULL,
    before_present boolean NOT NULL,
    before_value jsonb,
    after_present boolean NOT NULL,
    after_value jsonb,
    CONSTRAINT record_revision_conflict_facts_pkey PRIMARY KEY (revision_id, field_key),
    CONSTRAINT record_revision_conflict_facts_field_key_ck
        CHECK (field_key = btrim(field_key) AND field_key <> ''),
    CONSTRAINT record_revision_conflict_facts_before_value_ck
        CHECK (before_present = (before_value IS NOT NULL)),
    CONSTRAINT record_revision_conflict_facts_after_value_ck
        CHECK (after_present = (after_value IS NOT NULL)),
    CONSTRAINT record_revision_conflict_facts_presence_ck
        CHECK (before_present OR after_present)
);

-- +goose Down
DROP TABLE public.record_revision_conflict_facts;

DROP INDEX public.change_set_mutations_history_record_ids_idx;

ALTER TABLE public.change_set_mutations
    DROP CONSTRAINT change_set_mutations_history_entry_subset_ck,
    DROP CONSTRAINT change_set_mutations_history_entry_record_ids_ck,
    DROP CONSTRAINT change_set_mutations_history_record_ids_ck,
    DROP COLUMN history_entry_record_ids,
    DROP COLUMN history_record_ids;

DROP FUNCTION public.change_set_mutations_history_ids_are_canonical(uuid[]);

DROP FUNCTION public.revisions_incident_bundle_sequence_finish_v1(bigint);
DROP FUNCTION public.revisions_incident_bundle_sequence_begin_v1();

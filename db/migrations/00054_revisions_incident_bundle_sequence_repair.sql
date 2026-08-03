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

-- +goose Down
DROP FUNCTION public.revisions_incident_bundle_sequence_finish_v1(bigint);
DROP FUNCTION public.revisions_incident_bundle_sequence_begin_v1();

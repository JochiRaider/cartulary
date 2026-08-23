-- +goose Up

-- Prove every historical row against the stricter semantics before replacing
-- the helper and its CHECK dependency; never choose or discard a duplicate
-- reference on an operator's behalf.
-- +goose StatementBegin
DO $$
DECLARE
    invalid_interval_count bigint;
BEGIN
    SELECT count(*)
      INTO invalid_interval_count
      FROM public.indicator_state_intervals
     WHERE NOT public.indicator_support_refs_are_valid(support_refs)
        OR EXISTS (
            SELECT 1
              FROM jsonb_array_elements(support_refs) member
             GROUP BY member.value
            HAVING count(*) > 1
        );

    IF invalid_interval_count <> 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'indicators_lifecycle_integrity_preflight_failed',
            DETAIL = format(
                'invalid_interval_rows=%s',
                invalid_interval_count
            ),
            HINT = 'Keep the pre-upgrade binary active and explicitly correct duplicate or malformed lifecycle support references before retrying; this migration never deduplicates or rewrites lifecycle history.';
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE public.indicator_state_intervals
    DROP CONSTRAINT indicator_state_intervals_support_refs_ck;

DROP FUNCTION public.indicator_support_refs_are_valid(jsonb);

-- +goose StatementBegin
CREATE FUNCTION public.indicator_support_refs_are_valid(candidate jsonb)
RETURNS boolean
LANGUAGE plpgsql
IMMUTABLE
STRICT
SET search_path = pg_catalog, public
AS $$
DECLARE
    member jsonb;
    raw text;
    support_record_id uuid;
    seen_support_record_ids uuid[] := ARRAY[]::uuid[];
BEGIN
    IF jsonb_typeof(candidate) <> 'array' THEN
        RETURN false;
    END IF;
    FOR member IN SELECT value FROM jsonb_array_elements(candidate)
    LOOP
        IF jsonb_typeof(member) <> 'string' THEN
            RETURN false;
        END IF;
        raw := member #>> '{}';
        BEGIN
            support_record_id := raw::uuid;
            IF raw <> support_record_id::text
                OR support_record_id = ANY(seen_support_record_ids) THEN
                RETURN false;
            END IF;
        EXCEPTION WHEN invalid_text_representation THEN
            RETURN false;
        END;
        seen_support_record_ids := array_append(
            seen_support_record_ids,
            support_record_id
        );
    END LOOP;
    RETURN true;
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.indicator_support_refs_are_valid(jsonb)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.indicator_support_refs_are_valid(jsonb)
    TO cartulary_runtime, cartulary_recovery;

ALTER TABLE public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_support_refs_ck CHECK (
        public.indicator_support_refs_are_valid(support_refs)
    );

-- +goose Down

-- Restore the version-12 helper semantics. Existing data is intentionally
-- untouched; a disposable rollback relaxes only future validation.
ALTER TABLE public.indicator_state_intervals
    DROP CONSTRAINT indicator_state_intervals_support_refs_ck;

DROP FUNCTION public.indicator_support_refs_are_valid(jsonb);

-- +goose StatementBegin
CREATE FUNCTION public.indicator_support_refs_are_valid(candidate jsonb)
RETURNS boolean
LANGUAGE plpgsql
IMMUTABLE
STRICT
SET search_path = pg_catalog, public
AS $$
DECLARE
    member jsonb;
    raw text;
BEGIN
    IF jsonb_typeof(candidate) <> 'array' THEN
        RETURN false;
    END IF;
    FOR member IN SELECT value FROM jsonb_array_elements(candidate)
    LOOP
        IF jsonb_typeof(member) <> 'string' THEN
            RETURN false;
        END IF;
        raw := member #>> '{}';
        BEGIN
            IF raw <> (raw::uuid)::text THEN
                RETURN false;
            END IF;
        EXCEPTION WHEN invalid_text_representation THEN
            RETURN false;
        END;
    END LOOP;
    RETURN true;
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.indicator_support_refs_are_valid(jsonb)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.indicator_support_refs_are_valid(jsonb)
    TO cartulary_runtime, cartulary_recovery;

ALTER TABLE public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_support_refs_ck CHECK (
        public.indicator_support_refs_are_valid(support_refs)
    );

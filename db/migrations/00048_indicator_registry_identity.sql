-- +goose Up
-- Migration 46 rejects retained pre-cutover identities. This migration never
-- guesses a type-specific identity or merges incident records.
-- +goose StatementBegin
DO $$ BEGIN
    LOCK TABLE public.indicators, public.indicator_observations IN SHARE ROW EXCLUSIVE MODE;
    IF EXISTS (SELECT 1 FROM public.indicators) OR EXISTS (SELECT 1 FROM public.indicator_observations) THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: retained indicator identities; preserve a matching backup and use the explicit development reset procedure' USING ERRCODE = '55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.indicators DROP CONSTRAINT indicators_dedupe_key_ck;
ALTER TABLE public.indicators ALTER COLUMN normalized_value SET NOT NULL;
ALTER TABLE public.indicators ADD CONSTRAINT indicators_dedupe_key_ck CHECK (
    dedupe_key = CASE WHEN indicator_type = 'text'
        THEN 'text:' || encode(public.digest(normalized_value, 'sha256'), 'hex')
        ELSE indicator_type || ':' || normalized_value END
    AND octet_length(dedupe_key) BETWEEN 1 AND 32784
);
ALTER TABLE public.indicators ADD CONSTRAINT indicators_registry_value_kind_ck CHECK (
    value_kind = 'atomic' OR
    (value_kind = 'reference' AND indicator_type IN ('url','registry_key','process_name','text')) OR
    (value_kind = 'pattern' AND indicator_type = 'text')
);

-- Index digests bound physical key size. Every lookup also compares exact
-- logical bytes; a digest collision can reject an insert, never alias records.
DROP INDEX public.indicators_incident_normalized_lookup_idx;
CREATE INDEX indicators_incident_normalized_lookup_idx ON public.indicators
    (incident_id, indicator_type, public.digest(normalized_value, 'sha256'), record_id);
ALTER TABLE public.indicator_active_identities DROP CONSTRAINT indicator_active_identities_pkey;
ALTER TABLE public.indicator_active_identities DROP CONSTRAINT indicator_active_identities_exact_dedupe_ck;
ALTER TABLE public.indicator_active_identities ADD COLUMN dedupe_sha256 bytea
    GENERATED ALWAYS AS (public.digest(dedupe_key, 'sha256')) STORED;
ALTER TABLE public.indicator_active_identities ADD CONSTRAINT indicator_active_identities_pkey
    PRIMARY KEY (incident_id, indicator_type, dedupe_sha256);
ALTER TABLE public.indicator_active_identities ADD CONSTRAINT indicator_active_identities_exact_dedupe_ck
    CHECK (left(dedupe_key, length(indicator_type)+1) = indicator_type || ':' AND octet_length(dedupe_key) BETWEEN 1 AND 32784);

DROP FUNCTION public.indicator_active_identities_are_valid();
-- +goose StatementBegin
CREATE FUNCTION public.indicator_active_identities_are_valid()
RETURNS boolean
LANGUAGE sql
STABLE
SET search_path = pg_catalog, public
AS $$
    WITH expected AS (
        SELECT indicator.incident_id, indicator.indicator_type, indicator.dedupe_key,
               indicator.record_id AS indicator_record_id
          FROM public.indicators AS indicator
          JOIN public.records AS envelope
            ON envelope.record_id = indicator.record_id
           AND envelope.incident_id = indicator.incident_id
           AND envelope.record_type = 'indicator'
         WHERE envelope.deleted_at IS NULL
    ), difference AS (
        (SELECT * FROM expected EXCEPT SELECT incident_id, indicator_type, dedupe_key, indicator_record_id FROM public.indicator_active_identities)
        UNION ALL
        (SELECT incident_id, indicator_type, dedupe_key, indicator_record_id FROM public.indicator_active_identities EXCEPT SELECT * FROM expected)
    )
    SELECT NOT EXISTS (SELECT 1 FROM difference)
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.indicator_active_identities_are_valid() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.indicator_active_identities_are_valid() TO cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'indicator registry identities cannot be downgraded; restore a complete matching backup' USING ERRCODE = '55000'; END $$;
-- +goose StatementEnd

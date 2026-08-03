-- +goose Up

CREATE TABLE public.indicator_active_identities (
    incident_id uuid NOT NULL,
    indicator_type text NOT NULL,
    dedupe_key text NOT NULL,
    indicator_record_id uuid NOT NULL,
    CONSTRAINT indicator_active_identities_pkey
        PRIMARY KEY (incident_id, indicator_type, dedupe_key),
    CONSTRAINT indicator_active_identities_record_key
        UNIQUE (indicator_record_id),
    CONSTRAINT indicator_active_identities_type_ck
        CHECK (octet_length(indicator_type) > 0 AND indicator_type !~ '[[:cntrl:]]')
        NOT VALID,
    CONSTRAINT indicator_active_identities_dedupe_ck
        CHECK (octet_length(dedupe_key) > 0 AND dedupe_key !~ '[[:cntrl:]]')
        NOT VALID,
    CONSTRAINT indicator_active_identities_indicator_fkey
        FOREIGN KEY (indicator_record_id)
        REFERENCES public.indicators(record_id)
        ON DELETE CASCADE
        NOT VALID,
    CONSTRAINT indicator_active_identities_record_envelope_fkey
        FOREIGN KEY (incident_id, indicator_record_id)
        REFERENCES public.records(incident_id, record_id)
        ON DELETE CASCADE
        NOT VALID
);

-- Records owns active/deleted state. Refuse to build claims from a data set that
-- already violates the one-active-record-per-canonical-identity invariant.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.indicators AS indicator
          JOIN public.records AS envelope
            ON envelope.record_id = indicator.record_id
           AND envelope.incident_id = indicator.incident_id
           AND envelope.record_type = 'indicator'
         WHERE envelope.deleted_at IS NULL
         GROUP BY indicator.incident_id, indicator.indicator_type, indicator.dedupe_key
        HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'active indicator identities are not unique according to records';
    END IF;
END
$$;
-- +goose StatementEnd

INSERT INTO public.indicator_active_identities (
    incident_id,
    indicator_type,
    dedupe_key,
    indicator_record_id
)
SELECT
    indicator.incident_id,
    indicator.indicator_type,
    indicator.dedupe_key,
    indicator.record_id
  FROM public.indicators AS indicator
  JOIN public.records AS envelope
    ON envelope.record_id = indicator.record_id
   AND envelope.incident_id = indicator.incident_id
   AND envelope.record_type = 'indicator'
 WHERE envelope.deleted_at IS NULL
 ORDER BY indicator.incident_id, indicator.indicator_type, indicator.dedupe_key;

-- The two synchronization triggers are the expand-window compatibility bridge.
-- They make Records deletion state authoritative while allowing both old writers
-- and the new claim-based reader to operate during deployment. ENABLE ALWAYS is
-- required because recovery restores authoritative tables with ordinary triggers
-- disabled and may load Records and Indicator subtype rows in either order.
-- +goose StatementBegin
CREATE FUNCTION public.sync_indicator_active_identity_from_indicator()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        DELETE FROM public.indicator_active_identities
         WHERE indicator_record_id = OLD.record_id;
        RETURN OLD;
    END IF;

    DELETE FROM public.indicator_active_identities
     WHERE indicator_record_id = NEW.record_id;

    INSERT INTO public.indicator_active_identities (
        incident_id,
        indicator_type,
        dedupe_key,
        indicator_record_id
    )
    SELECT
        NEW.incident_id,
        NEW.indicator_type,
        NEW.dedupe_key,
        NEW.record_id
      FROM public.records AS envelope
     WHERE envelope.record_id = NEW.record_id
       AND envelope.incident_id = NEW.incident_id
       AND envelope.record_type = 'indicator'
       AND envelope.deleted_at IS NULL;

    RETURN NEW;
END
$$;
-- +goose StatementEnd

CREATE TRIGGER indicators_sync_active_identity
AFTER INSERT OR DELETE OR UPDATE
ON public.indicators
FOR EACH ROW
EXECUTE FUNCTION public.sync_indicator_active_identity_from_indicator();

ALTER TABLE public.indicators
    ENABLE ALWAYS TRIGGER indicators_sync_active_identity;

-- +goose StatementBegin
CREATE FUNCTION public.sync_indicator_active_identity_from_record()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.record_type = 'indicator' THEN
            DELETE FROM public.indicator_active_identities
             WHERE indicator_record_id = OLD.record_id;
        END IF;
        RETURN OLD;
    END IF;

    IF NEW.record_type <> 'indicator' THEN
        IF TG_OP = 'UPDATE' AND OLD.record_type = 'indicator' THEN
            DELETE FROM public.indicator_active_identities
             WHERE indicator_record_id = OLD.record_id;
        END IF;
        RETURN NEW;
    END IF;

    DELETE FROM public.indicator_active_identities
     WHERE indicator_record_id = NEW.record_id;

    IF NEW.deleted_at IS NULL THEN
        INSERT INTO public.indicator_active_identities (
            incident_id,
            indicator_type,
            dedupe_key,
            indicator_record_id
        )
        SELECT
            indicator.incident_id,
            indicator.indicator_type,
            indicator.dedupe_key,
            indicator.record_id
          FROM public.indicators AS indicator
         WHERE indicator.record_id = NEW.record_id
           AND indicator.incident_id = NEW.incident_id;
    END IF;

    RETURN NEW;
END
$$;
-- +goose StatementEnd

CREATE TRIGGER records_sync_indicator_active_identity
AFTER INSERT OR DELETE OR UPDATE
ON public.records
FOR EACH ROW
EXECUTE FUNCTION public.sync_indicator_active_identity_from_record();

ALTER TABLE public.records
    ENABLE ALWAYS TRIGGER records_sync_indicator_active_identity;

-- This owner algorithm is safe to run after restore or operator repair. The
-- table locks make its result a deterministic snapshot rather than a best-effort
-- reconciliation racing ordinary writes.
-- +goose StatementBegin
CREATE FUNCTION public.rebuild_indicator_active_identities()
RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
    rebuilt_count bigint;
BEGIN
    LOCK TABLE public.records IN SHARE MODE;
    LOCK TABLE public.indicators IN SHARE MODE;
    LOCK TABLE public.indicator_active_identities IN EXCLUSIVE MODE;

    DELETE FROM public.indicator_active_identities;
    INSERT INTO public.indicator_active_identities (
        incident_id,
        indicator_type,
        dedupe_key,
        indicator_record_id
    )
    SELECT
        indicator.incident_id,
        indicator.indicator_type,
        indicator.dedupe_key,
        indicator.record_id
      FROM public.indicators AS indicator
      JOIN public.records AS envelope
        ON envelope.record_id = indicator.record_id
       AND envelope.incident_id = indicator.incident_id
       AND envelope.record_type = 'indicator'
     WHERE envelope.deleted_at IS NULL
     ORDER BY indicator.incident_id, indicator.indicator_type, indicator.dedupe_key;

    GET DIAGNOSTICS rebuilt_count = ROW_COUNT;
    RETURN rebuilt_count;
END
$$;
-- +goose StatementEnd

CREATE FUNCTION public.indicator_active_identities_are_valid()
RETURNS boolean
LANGUAGE sql
STABLE
AS $$
    WITH expected AS (
        SELECT
            indicator.incident_id,
            indicator.indicator_type,
            indicator.dedupe_key,
            indicator.record_id AS indicator_record_id
          FROM public.indicators AS indicator
          JOIN public.records AS envelope
            ON envelope.record_id = indicator.record_id
           AND envelope.incident_id = indicator.incident_id
           AND envelope.record_type = 'indicator'
         WHERE envelope.deleted_at IS NULL
    ), difference AS (
        (SELECT * FROM expected EXCEPT SELECT * FROM public.indicator_active_identities)
        UNION ALL
        (SELECT * FROM public.indicator_active_identities EXCEPT SELECT * FROM expected)
    )
    SELECT NOT EXISTS (SELECT 1 FROM difference)
$$;

-- +goose Down

DROP FUNCTION public.indicator_active_identities_are_valid();
DROP FUNCTION public.rebuild_indicator_active_identities();

DROP TRIGGER records_sync_indicator_active_identity ON public.records;
DROP FUNCTION public.sync_indicator_active_identity_from_record();

DROP TRIGGER indicators_sync_active_identity ON public.indicators;
DROP FUNCTION public.sync_indicator_active_identity_from_indicator();

DROP TABLE public.indicator_active_identities;

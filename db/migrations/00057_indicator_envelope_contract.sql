-- +goose Up

-- This is a contract migration. Deployments must drain every binary that can
-- still write the Indicator envelope mirrors before applying it.
LOCK TABLE public.records IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.indicators IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.indicator_active_identities IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.indicator_observations IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.indicator_state_intervals IN ACCESS EXCLUSIVE MODE;

-- Refuse to discard a mirror that disagrees with the Records authority.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.indicators AS indicator
          LEFT JOIN public.records AS envelope
            ON envelope.incident_id = indicator.incident_id
           AND envelope.record_id = indicator.record_id
           AND envelope.record_type = 'indicator'
         WHERE envelope.record_id IS NULL
            OR indicator.row_version <> envelope.row_version
            OR indicator.created_at <> envelope.created_at
            OR indicator.updated_at <> envelope.updated_at
            OR indicator.created_by_user_id <> envelope.created_by_user_id
            OR indicator.updated_by_user_id <> envelope.updated_by_user_id
            OR indicator.deleted_at IS DISTINCT FROM envelope.deleted_at
            OR indicator.deleted_by_user_id IS DISTINCT FROM envelope.deleted_by_user_id
    ) THEN
        RAISE EXCEPTION 'indicator contract migration blocked by envelope drift';
    END IF;

    IF NOT public.indicator_active_identities_are_valid() THEN
        RAISE EXCEPTION 'indicator contract migration blocked by active identity drift';
    END IF;
END
$$;
-- +goose StatementEnd

-- Stored support references remain JSON in the source-major-1 contract, but
-- storage admits only a JSON array of canonical UUID strings.
-- +goose StatementBegin
CREATE FUNCTION public.indicator_support_refs_are_valid(candidate jsonb)
RETURNS boolean
LANGUAGE plpgsql
IMMUTABLE
STRICT
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

ALTER TABLE public.indicators
    ADD CONSTRAINT indicators_incident_record_key
        UNIQUE (incident_id, record_id),
    ADD CONSTRAINT indicators_indicator_type_ck
        CHECK (indicator_type IN (
            'ipv4_addr', 'ipv6_addr', 'domain_name', 'url', 'sha256',
            'email_addr', 'registry_key', 'process_name', 'text'
        )) NOT VALID,
    ADD CONSTRAINT indicators_required_text_ck
        CHECK (
            octet_length(display_value) > 0
            AND octet_length(dedupe_key) > 0
            AND (normalized_value IS NULL OR octet_length(normalized_value) > 0)
            AND (defanged_value IS NULL OR octet_length(defanged_value) > 0)
            AND (hash_algorithm IS NULL OR octet_length(hash_algorithm) > 0)
            AND (hash_value IS NULL OR octet_length(hash_value) > 0)
            AND (stix_pattern IS NULL OR octet_length(stix_pattern) > 0)
        ) NOT VALID,
    ADD CONSTRAINT indicators_dedupe_key_ck
        CHECK (dedupe_key ~ '^[0-9a-f]{64}$') NOT VALID,
    ADD CONSTRAINT indicators_hash_value_ck
        CHECK (hash_value IS NULL OR hash_value ~ '^[0-9a-f]+$') NOT VALID,
    ADD CONSTRAINT indicators_ip_value_kind_ck
        CHECK (indicator_type NOT IN ('ipv4_addr', 'ipv6_addr') OR value_kind = 'atomic') NOT VALID;

ALTER TABLE public.indicator_active_identities
    ADD CONSTRAINT indicator_active_identities_exact_type_ck
        CHECK (indicator_type IN (
            'ipv4_addr', 'ipv6_addr', 'domain_name', 'url', 'sha256',
            'email_addr', 'registry_key', 'process_name', 'text'
        )) NOT VALID,
    ADD CONSTRAINT indicator_active_identities_exact_dedupe_ck
        CHECK (dedupe_key ~ '^[0-9a-f]{64}$') NOT VALID;

ALTER TABLE public.indicator_observations
    ADD CONSTRAINT indicator_observations_required_text_ck
        CHECK (
            octet_length(source_field_key) > 0
            AND octet_length(origin_locator) > 0
            AND octet_length(observed_text) > 0
            AND (normalized_candidate IS NULL OR octet_length(normalized_candidate) > 0)
            AND (resolution_method IS NULL OR octet_length(resolution_method) > 0)
        ) NOT VALID,
    ADD CONSTRAINT indicator_observations_parsed_type_ck
        CHECK (parsed_indicator_type IS NULL OR parsed_indicator_type IN (
            'ipv4_addr', 'ipv6_addr', 'domain_name', 'url', 'sha256',
            'email_addr', 'registry_key', 'process_name', 'text'
        )) NOT VALID,
    ADD CONSTRAINT indicator_observations_candidate_pair_ck
        CHECK ((parsed_indicator_type IS NULL) = (normalized_candidate IS NULL)) NOT VALID,
    ADD CONSTRAINT indicator_observations_row_version_ck
        CHECK (row_version >= 1) NOT VALID,
    ADD CONSTRAINT indicator_observations_resolution_tuple_ck
        CHECK (
            (resolution_status = 'unresolved'
                AND resolved_indicator_record_id IS NULL
                AND resolved_by_user_id IS NULL
                AND resolved_at IS NULL
                AND resolution_method IS NULL)
            OR
            (resolution_status = 'resolved'
                AND resolved_indicator_record_id IS NOT NULL
                AND resolved_by_user_id IS NOT NULL
                AND resolved_at IS NOT NULL
                AND resolution_method IS NOT NULL)
            OR
            (resolution_status = 'dismissed'
                AND resolved_indicator_record_id IS NULL
                AND resolved_by_user_id IS NOT NULL
                AND resolved_at IS NOT NULL
                AND resolution_method IS NOT NULL)
        ) NOT VALID;

ALTER TABLE public.indicator_state_intervals
    ADD CONSTRAINT indicator_state_intervals_lifecycle_state_ck
        CHECK (lifecycle_state IN ('active', 'benign', 'false_positive', 'retired')) NOT VALID,
    ADD CONSTRAINT indicator_state_intervals_required_text_ck
        CHECK (
            (rationale IS NULL OR octet_length(rationale) > 0)
            AND (assessor IS NULL OR octet_length(assessor) > 0)
        ) NOT VALID,
    ADD CONSTRAINT indicator_state_intervals_row_version_ck
        CHECK (row_version >= 1) NOT VALID,
    ADD CONSTRAINT indicator_state_intervals_support_refs_ck
        CHECK (public.indicator_support_refs_are_valid(support_refs)) NOT VALID;

ALTER TABLE public.indicator_observations
    DROP CONSTRAINT indicator_observations_resolved_indicator_record_id_fkey,
    ADD CONSTRAINT indicator_observations_resolved_indicator_incident_fkey
        FOREIGN KEY (incident_id, resolved_indicator_record_id)
        REFERENCES public.indicators(incident_id, record_id)
        NOT VALID;

ALTER TABLE public.indicator_state_intervals
    DROP CONSTRAINT indicator_state_intervals_indicator_record_id_fkey,
    ADD CONSTRAINT indicator_state_intervals_indicator_incident_fkey
        FOREIGN KEY (incident_id, indicator_record_id)
        REFERENCES public.indicators(incident_id, record_id)
        NOT VALID;

-- A support UUID is not a physical FK column, so enforce incident membership
-- through an owner trigger after the JSON shape check has made casts total.
-- +goose StatementBegin
CREATE FUNCTION public.enforce_indicator_support_ref_incident()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    member jsonb;
    support_record_id uuid;
BEGIN
    IF NOT public.indicator_support_refs_are_valid(NEW.support_refs) THEN
        RAISE EXCEPTION 'indicator lifecycle support references are malformed'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'indicator_state_intervals_support_refs_ck';
    END IF;
    FOR member IN SELECT value FROM jsonb_array_elements(NEW.support_refs)
    LOOP
        support_record_id := (member #>> '{}')::uuid;
        IF NOT EXISTS (
            SELECT 1
              FROM public.records
             WHERE incident_id = NEW.incident_id
               AND record_id = support_record_id
        ) THEN
            RAISE EXCEPTION 'indicator lifecycle support reference is outside the incident'
                USING ERRCODE = '23514';
        END IF;
    END LOOP;
    RETURN NEW;
END
$$;
-- +goose StatementEnd

CREATE TRIGGER indicator_state_intervals_support_ref_incident
BEFORE INSERT OR UPDATE
ON public.indicator_state_intervals
FOR EACH ROW
EXECUTE FUNCTION public.enforce_indicator_support_ref_incident();

ALTER TABLE public.indicator_active_identities
    VALIDATE CONSTRAINT indicator_active_identities_type_ck,
    VALIDATE CONSTRAINT indicator_active_identities_dedupe_ck,
    VALIDATE CONSTRAINT indicator_active_identities_indicator_fkey,
    VALIDATE CONSTRAINT indicator_active_identities_record_envelope_fkey,
    VALIDATE CONSTRAINT indicator_active_identities_exact_type_ck,
    VALIDATE CONSTRAINT indicator_active_identities_exact_dedupe_ck;

ALTER TABLE public.indicators
    VALIDATE CONSTRAINT indicators_indicator_type_ck,
    VALIDATE CONSTRAINT indicators_required_text_ck,
    VALIDATE CONSTRAINT indicators_dedupe_key_ck,
    VALIDATE CONSTRAINT indicators_hash_value_ck,
    VALIDATE CONSTRAINT indicators_ip_value_kind_ck;

ALTER TABLE public.indicator_observations
    VALIDATE CONSTRAINT indicator_observations_required_text_ck,
    VALIDATE CONSTRAINT indicator_observations_parsed_type_ck,
    VALIDATE CONSTRAINT indicator_observations_candidate_pair_ck,
    VALIDATE CONSTRAINT indicator_observations_row_version_ck,
    VALIDATE CONSTRAINT indicator_observations_resolution_tuple_ck,
    VALIDATE CONSTRAINT indicator_observations_resolved_indicator_incident_fkey;

ALTER TABLE public.indicator_state_intervals
    VALIDATE CONSTRAINT indicator_state_intervals_lifecycle_state_ck,
    VALIDATE CONSTRAINT indicator_state_intervals_required_text_ck,
    VALIDATE CONSTRAINT indicator_state_intervals_row_version_ck,
    VALIDATE CONSTRAINT indicator_state_intervals_support_refs_ck,
    VALIDATE CONSTRAINT indicator_state_intervals_indicator_incident_fkey;

-- Validate support-record incident membership for rows that predate the
-- trigger. The migration remains atomic when this or any prior gate fails.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.indicator_state_intervals AS interval
          CROSS JOIN LATERAL jsonb_array_elements_text(interval.support_refs) AS support(record_id)
          LEFT JOIN public.records AS envelope
            ON envelope.incident_id = interval.incident_id
           AND envelope.record_id = support.record_id::uuid
         WHERE envelope.record_id IS NULL
    ) THEN
        RAISE EXCEPTION 'indicator contract migration blocked by cross-incident support reference';
    END IF;
END
$$;
-- +goose StatementEnd

DROP INDEX public.indicators_incident_dedupe_unique_idx;
DROP INDEX public.indicators_incident_normalized_lookup_idx;

CREATE INDEX indicators_incident_normalized_lookup_idx
    ON public.indicators (incident_id, indicator_type, normalized_value, record_id)
    WHERE normalized_value IS NOT NULL;

ALTER TABLE public.indicators
    DROP CONSTRAINT indicators_created_by_user_id_fkey,
    DROP CONSTRAINT indicators_updated_by_user_id_fkey,
    DROP CONSTRAINT indicators_deleted_by_user_id_fkey,
    DROP COLUMN row_version,
    DROP COLUMN created_at,
    DROP COLUMN updated_at,
    DROP COLUMN created_by_user_id,
    DROP COLUMN updated_by_user_id,
    DROP COLUMN deleted_at,
    DROP COLUMN deleted_by_user_id;

-- Avoid claim churn on every Records version advance or source enrichment.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.sync_indicator_active_identity_from_indicator()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        DELETE FROM public.indicator_active_identities
         WHERE indicator_record_id = OLD.record_id;
        RETURN OLD;
    END IF;
    IF TG_OP = 'UPDATE'
       AND NEW.incident_id = OLD.incident_id
       AND NEW.indicator_type = OLD.indicator_type
       AND NEW.dedupe_key = OLD.dedupe_key THEN
        RETURN NEW;
    END IF;

    DELETE FROM public.indicator_active_identities
     WHERE indicator_record_id = NEW.record_id;
    INSERT INTO public.indicator_active_identities (
        incident_id, indicator_type, dedupe_key, indicator_record_id
    )
    SELECT NEW.incident_id, NEW.indicator_type, NEW.dedupe_key, NEW.record_id
      FROM public.records AS envelope
     WHERE envelope.record_id = NEW.record_id
       AND envelope.incident_id = NEW.incident_id
       AND envelope.record_type = 'indicator'
       AND envelope.deleted_at IS NULL;
    RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.sync_indicator_active_identity_from_record()
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
    IF TG_OP = 'UPDATE'
       AND NEW.incident_id = OLD.incident_id
       AND NEW.record_type = OLD.record_type
       AND NEW.deleted_at IS NOT DISTINCT FROM OLD.deleted_at THEN
        RETURN NEW;
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
            incident_id, indicator_type, dedupe_key, indicator_record_id
        )
        SELECT indicator.incident_id, indicator.indicator_type,
               indicator.dedupe_key, indicator.record_id
          FROM public.indicators AS indicator
         WHERE indicator.record_id = NEW.record_id
           AND indicator.incident_id = NEW.incident_id;
    END IF;
    RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose Down

LOCK TABLE public.records IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.indicators IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.indicator_active_identities IN ACCESS EXCLUSIVE MODE;

-- Reconstruct every legacy mirror from Records before an old binary can run.
ALTER TABLE public.indicators
    ADD COLUMN row_version bigint,
    ADD COLUMN created_at timestamp with time zone,
    ADD COLUMN updated_at timestamp with time zone,
    ADD COLUMN created_by_user_id uuid,
    ADD COLUMN updated_by_user_id uuid,
    ADD COLUMN deleted_at timestamp with time zone,
    ADD COLUMN deleted_by_user_id uuid;

UPDATE public.indicators AS indicator
   SET row_version = envelope.row_version,
       created_at = envelope.created_at,
       updated_at = envelope.updated_at,
       created_by_user_id = envelope.created_by_user_id,
       updated_by_user_id = envelope.updated_by_user_id,
       deleted_at = envelope.deleted_at,
       deleted_by_user_id = envelope.deleted_by_user_id
  FROM public.records AS envelope
 WHERE envelope.incident_id = indicator.incident_id
   AND envelope.record_id = indicator.record_id
   AND envelope.record_type = 'indicator';

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.indicators
         WHERE row_version IS NULL
            OR created_at IS NULL
            OR updated_at IS NULL
            OR created_by_user_id IS NULL
            OR updated_by_user_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot reconstruct Indicator envelope mirrors from Records';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.indicators
    ALTER COLUMN row_version SET DEFAULT 1,
    ALTER COLUMN row_version SET NOT NULL,
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET DEFAULT now(),
    ALTER COLUMN updated_at SET NOT NULL,
    ALTER COLUMN created_by_user_id SET NOT NULL,
    ALTER COLUMN updated_by_user_id SET NOT NULL,
    ADD CONSTRAINT indicators_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id) REFERENCES public.users(id),
    ADD CONSTRAINT indicators_updated_by_user_id_fkey
        FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id),
    ADD CONSTRAINT indicators_deleted_by_user_id_fkey
        FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id);

DROP INDEX public.indicators_incident_normalized_lookup_idx;
CREATE UNIQUE INDEX indicators_incident_dedupe_unique_idx
    ON public.indicators (incident_id, indicator_type, dedupe_key)
    WHERE deleted_at IS NULL;
CREATE INDEX indicators_incident_normalized_lookup_idx
    ON public.indicators (incident_id, indicator_type, normalized_value, record_id)
    WHERE deleted_at IS NULL AND normalized_value IS NOT NULL;

DROP TRIGGER indicator_state_intervals_support_ref_incident
    ON public.indicator_state_intervals;
DROP FUNCTION public.enforce_indicator_support_ref_incident();

ALTER TABLE public.indicator_observations
    DROP CONSTRAINT indicator_observations_resolved_indicator_incident_fkey,
    ADD CONSTRAINT indicator_observations_resolved_indicator_record_id_fkey
        FOREIGN KEY (resolved_indicator_record_id)
        REFERENCES public.indicators(record_id)
        ON DELETE SET NULL,
    DROP CONSTRAINT indicator_observations_required_text_ck,
    DROP CONSTRAINT indicator_observations_parsed_type_ck,
    DROP CONSTRAINT indicator_observations_candidate_pair_ck,
    DROP CONSTRAINT indicator_observations_row_version_ck,
    DROP CONSTRAINT indicator_observations_resolution_tuple_ck;

ALTER TABLE public.indicator_state_intervals
    DROP CONSTRAINT indicator_state_intervals_indicator_incident_fkey,
    ADD CONSTRAINT indicator_state_intervals_indicator_record_id_fkey
        FOREIGN KEY (indicator_record_id)
        REFERENCES public.indicators(record_id)
        ON DELETE CASCADE,
    DROP CONSTRAINT indicator_state_intervals_lifecycle_state_ck,
    DROP CONSTRAINT indicator_state_intervals_required_text_ck,
    DROP CONSTRAINT indicator_state_intervals_row_version_ck,
    DROP CONSTRAINT indicator_state_intervals_support_refs_ck;

ALTER TABLE public.indicator_active_identities
    DROP CONSTRAINT indicator_active_identities_exact_type_ck,
    DROP CONSTRAINT indicator_active_identities_exact_dedupe_ck;

ALTER TABLE public.indicators
    DROP CONSTRAINT indicators_incident_record_key,
    DROP CONSTRAINT indicators_indicator_type_ck,
    DROP CONSTRAINT indicators_required_text_ck,
    DROP CONSTRAINT indicators_dedupe_key_ck,
    DROP CONSTRAINT indicators_hash_value_ck,
    DROP CONSTRAINT indicators_ip_value_kind_ck;

DROP FUNCTION public.indicator_support_refs_are_valid(jsonb);

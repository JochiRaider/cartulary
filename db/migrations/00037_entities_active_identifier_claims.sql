-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION public.entities_normalize_identifier_v1(
    identifier_type text,
    raw_value text
)
RETURNS text
LANGUAGE plpgsql
IMMUTABLE
PARALLEL SAFE
SET search_path = pg_catalog, public
AS $$
DECLARE
    normalized_value text;
BEGIN
    IF raw_value IS NULL OR identifier_type NOT IN (
        'aad_device_id', 'fqdn', 'hostname', 'aad_object_id',
        'sid', 'upn', 'email', 'sam_account_name'
    ) THEN
        RETURN NULL;
    END IF;
    normalized_value := normalize(
        public.entities_trim_unicode_space_v1(raw_value),
        NFC
    );
    IF normalized_value = ''
        OR NOT public.entities_source_codepoints_admitted_v1(
            normalized_value,
            true
        ) THEN
        RETURN NULL;
    END IF;
    IF identifier_type = 'sid' THEN
        RETURN upper(normalized_value);
    END IF;
    RETURN lower(normalized_value);
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_normalize_identifier_v1(text, text)
    FROM PUBLIC;

-- Reject invalid canonical values and competing active owners before any
-- claim relation is created. The transaction rolls this function back with
-- the rest of the migration when the preflight fails.
-- +goose StatementBegin
DO $$
DECLARE
    invalid_identifier_count bigint;
    duplicate_claim_group_count bigint;
BEGIN
    WITH canonical_values AS (
        SELECT h.record_id, h.incident_id, 'host'::text AS entity_type,
               candidate.identifier_type, candidate.raw_value
          FROM public.hosts h
          JOIN public.records r
            ON r.record_id = h.record_id
           AND r.incident_id = h.incident_id
           AND r.record_type = 'host'
          CROSS JOIN LATERAL (VALUES
              ('aad_device_id'::text, h.aad_device_id),
              ('fqdn'::text, h.fqdn),
              ('hostname'::text, h.hostname)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND h.host_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
        UNION ALL
        SELECT i.record_id, i.incident_id, 'identity'::text,
               candidate.identifier_type, candidate.raw_value
          FROM public.identities i
          JOIN public.records r
            ON r.record_id = i.record_id
           AND r.incident_id = i.incident_id
           AND r.record_type = 'identity'
          CROSS JOIN LATERAL (VALUES
              ('aad_object_id'::text, i.aad_object_id),
              ('sid'::text, i.sid),
              ('upn'::text, i.upn),
              ('email'::text, i.email::text),
              ('sam_account_name'::text, i.sam_account_name)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND i.identity_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
    )
    SELECT count(*)
      INTO invalid_identifier_count
      FROM canonical_values
     WHERE public.entities_normalize_identifier_v1(
        identifier_type,
        raw_value
     ) IS NULL;

    WITH expected_claims AS (
        SELECT h.record_id, h.incident_id, 'host'::text AS entity_type,
               candidate.identifier_type,
               public.entities_normalize_identifier_v1(
                   candidate.identifier_type,
                   candidate.raw_value
               ) AS normalized_value
          FROM public.hosts h
          JOIN public.records r
            ON r.record_id = h.record_id
           AND r.incident_id = h.incident_id
           AND r.record_type = 'host'
          CROSS JOIN LATERAL (VALUES
              ('aad_device_id'::text, h.aad_device_id),
              ('fqdn'::text, h.fqdn),
              ('hostname'::text, h.hostname)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND h.host_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
        UNION
        SELECT i.record_id, i.incident_id, 'identity'::text,
               candidate.identifier_type,
               public.entities_normalize_identifier_v1(
                   candidate.identifier_type,
                   candidate.raw_value
               )
          FROM public.identities i
          JOIN public.records r
            ON r.record_id = i.record_id
           AND r.incident_id = i.incident_id
           AND r.record_type = 'identity'
          CROSS JOIN LATERAL (VALUES
              ('aad_object_id'::text, i.aad_object_id),
              ('sid'::text, i.sid),
              ('upn'::text, i.upn),
              ('email'::text, i.email::text),
              ('sam_account_name'::text, i.sam_account_name)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND i.identity_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
        UNION
        SELECT p.record_id, p.incident_id, p.entity_type,
               p.identifier_type, p.normalized_value
          FROM public.entity_preserved_identifiers p
          JOIN public.records r
            ON r.record_id = p.record_id
           AND r.incident_id = p.incident_id
           AND r.record_type = p.entity_type
          LEFT JOIN public.hosts h
            ON p.entity_type = 'host'
           AND h.record_id = p.record_id
          LEFT JOIN public.identities i
            ON p.entity_type = 'identity'
           AND i.record_id = p.record_id
         WHERE r.deleted_at IS NULL
           AND p.deleted_at IS NULL
           AND p.classification = 'exact_match_reuse'
           AND (
               (p.entity_type = 'host'
                   AND h.host_state IN ('stub', 'canonical'))
               OR (p.entity_type = 'identity'
                   AND i.identity_state IN ('stub', 'canonical'))
           )
    ), duplicate_groups AS (
        SELECT incident_id, entity_type, identifier_type, normalized_value
          FROM expected_claims
         WHERE normalized_value IS NOT NULL
         GROUP BY incident_id, entity_type, identifier_type, normalized_value
        HAVING count(DISTINCT record_id) > 1
    )
    SELECT count(*)
      INTO duplicate_claim_group_count
      FROM duplicate_groups;

    IF invalid_identifier_count + duplicate_claim_group_count <> 0 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23505',
            MESSAGE = 'entities_active_identifier_claims_preflight_failed',
            DETAIL = format(
                'invalid_identifiers=%s duplicate_claim_groups=%s',
                invalid_identifier_count,
                duplicate_claim_group_count
            ),
            HINT = 'Keep the pre-cutover binary active and explicitly disposition invalid or competing Entities identifiers before retrying; this migration never merges, discards, or chooses a winner.';
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE public.entity_active_identifier_claims (
    incident_id uuid NOT NULL,
    entity_type text NOT NULL,
    identifier_type text NOT NULL,
    normalized_value text NOT NULL,
    record_id uuid NOT NULL,
    CONSTRAINT entity_active_identifier_claims_pkey PRIMARY KEY (
        incident_id,
        entity_type,
        identifier_type,
        normalized_value
    ),
    CONSTRAINT entity_active_identifier_claims_entity_type_ck CHECK (
        entity_type IN ('host', 'identity')
    ),
    CONSTRAINT entity_active_identifier_claims_identifier_class_ck CHECK (
        (entity_type = 'host' AND identifier_type IN (
            'aad_device_id', 'fqdn', 'hostname'
        ))
        OR (entity_type = 'identity' AND identifier_type IN (
            'aad_object_id', 'sid', 'upn', 'email', 'sam_account_name'
        ))
    ),
    CONSTRAINT entity_active_identifier_claims_value_ck CHECK (
        normalized_value <> ''
        AND public.entities_normalize_identifier_v1(
            identifier_type,
            normalized_value
        ) = normalized_value
    ),
    CONSTRAINT entity_active_identifier_claims_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id)
        REFERENCES public.records(incident_id, record_id)
        ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX entity_active_identifier_claims_record_lookup_idx
    ON public.entity_active_identifier_claims (
        record_id,
        entity_type,
        identifier_type,
        normalized_value
    );

CREATE INDEX entity_active_identifier_claims_record_envelope_idx
    ON public.entity_active_identifier_claims (incident_id, record_id);

-- +goose StatementBegin
CREATE FUNCTION public.entities_expected_active_identifier_claims_v1(
    target_record_id uuid
)
RETURNS TABLE (
    incident_id uuid,
    entity_type text,
    identifier_type text,
    normalized_value text,
    record_id uuid
)
LANGUAGE sql
STABLE
SET search_path = pg_catalog, public
AS $$
    SELECT DISTINCT expected.incident_id, expected.entity_type,
           expected.identifier_type, expected.normalized_value,
           expected.record_id
      FROM (
        SELECT h.incident_id, 'host'::text AS entity_type,
               candidate.identifier_type,
               public.entities_normalize_identifier_v1(
                   candidate.identifier_type,
                   candidate.raw_value
               ) AS normalized_value,
               h.record_id
          FROM public.hosts h
          JOIN public.records r
            ON r.record_id = h.record_id
           AND r.incident_id = h.incident_id
           AND r.record_type = 'host'
          CROSS JOIN LATERAL (VALUES
              ('aad_device_id'::text, h.aad_device_id),
              ('fqdn'::text, h.fqdn),
              ('hostname'::text, h.hostname)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND h.host_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
           AND (target_record_id IS NULL OR h.record_id = target_record_id)
        UNION ALL
        SELECT i.incident_id, 'identity'::text,
               candidate.identifier_type,
               public.entities_normalize_identifier_v1(
                   candidate.identifier_type,
                   candidate.raw_value
               ),
               i.record_id
          FROM public.identities i
          JOIN public.records r
            ON r.record_id = i.record_id
           AND r.incident_id = i.incident_id
           AND r.record_type = 'identity'
          CROSS JOIN LATERAL (VALUES
              ('aad_object_id'::text, i.aad_object_id),
              ('sid'::text, i.sid),
              ('upn'::text, i.upn),
              ('email'::text, i.email::text),
              ('sam_account_name'::text, i.sam_account_name)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND i.identity_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
           AND (target_record_id IS NULL OR i.record_id = target_record_id)
        UNION ALL
        SELECT p.incident_id, p.entity_type, p.identifier_type,
               p.normalized_value, p.record_id
          FROM public.entity_preserved_identifiers p
          JOIN public.records r
            ON r.record_id = p.record_id
           AND r.incident_id = p.incident_id
           AND r.record_type = p.entity_type
          LEFT JOIN public.hosts h
            ON p.entity_type = 'host'
           AND h.record_id = p.record_id
          LEFT JOIN public.identities i
            ON p.entity_type = 'identity'
           AND i.record_id = p.record_id
         WHERE r.deleted_at IS NULL
           AND p.deleted_at IS NULL
           AND p.classification = 'exact_match_reuse'
           AND (target_record_id IS NULL OR p.record_id = target_record_id)
           AND (
               (p.entity_type = 'host'
                   AND h.host_state IN ('stub', 'canonical'))
               OR (p.entity_type = 'identity'
                   AND i.identity_state IN ('stub', 'canonical'))
           )
      ) AS expected
     WHERE expected.normalized_value IS NOT NULL
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_refresh_active_identifier_claims_v1(
    target_record_id uuid
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
    refreshed_count bigint;
BEGIN
    DELETE FROM public.entity_active_identifier_claims
     WHERE record_id = target_record_id;
    INSERT INTO public.entity_active_identifier_claims (
        incident_id,
        entity_type,
        identifier_type,
        normalized_value,
        record_id
    )
    SELECT expected.incident_id, expected.entity_type,
           expected.identifier_type, expected.normalized_value,
           expected.record_id
      FROM public.entities_expected_active_identifier_claims_v1(
          target_record_id
      ) expected
     ORDER BY expected.incident_id, expected.entity_type,
              expected.identifier_type, expected.normalized_value;
    GET DIAGNOSTICS refreshed_count = ROW_COUNT;
    RETURN refreshed_count;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_release_active_identifier_claims_v1(
    target_record_id uuid
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
    released_count bigint;
BEGIN
    DELETE FROM public.entity_active_identifier_claims
     WHERE record_id = target_record_id;
    GET DIAGNOSTICS released_count = ROW_COUNT;
    RETURN released_count;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_sync_active_identifier_claims_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
BEGIN
	IF pg_catalog.current_setting(
		'cartulary.entities_defer_active_identifier_claims',
		true
	) = 'on' THEN
		RETURN COALESCE(NEW, OLD);
	END IF;
    IF TG_OP = 'DELETE' THEN
        PERFORM public.entities_refresh_active_identifier_claims_v1(
            OLD.record_id
        );
        RETURN OLD;
    END IF;
    IF TG_OP = 'UPDATE' AND OLD.record_id IS DISTINCT FROM NEW.record_id THEN
        PERFORM public.entities_refresh_active_identifier_claims_v1(
            OLD.record_id
        );
    END IF;
    PERFORM public.entities_refresh_active_identifier_claims_v1(
        NEW.record_id
    );
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_rebuild_active_identifier_claims_v1()
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
    rebuilt_count bigint;
BEGIN
    LOCK TABLE public.records IN SHARE MODE;
    LOCK TABLE public.hosts IN SHARE MODE;
    LOCK TABLE public.identities IN SHARE MODE;
    LOCK TABLE public.entity_preserved_identifiers IN SHARE MODE;
    LOCK TABLE public.entity_active_identifier_claims IN EXCLUSIVE MODE;
    DELETE FROM public.entity_active_identifier_claims;
    INSERT INTO public.entity_active_identifier_claims (
        incident_id,
        entity_type,
        identifier_type,
        normalized_value,
        record_id
    )
    SELECT incident_id, entity_type, identifier_type,
           normalized_value, record_id
      FROM public.entities_expected_active_identifier_claims_v1(NULL::uuid)
     ORDER BY incident_id, entity_type, identifier_type, normalized_value;
    GET DIAGNOSTICS rebuilt_count = ROW_COUNT;
    RETURN rebuilt_count;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.entities_active_identifier_claims_are_valid_v1()
RETURNS boolean
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    WITH expected AS (
        SELECT *
          FROM public.entities_expected_active_identifier_claims_v1(NULL::uuid)
    ), difference AS (
        (SELECT * FROM expected
         EXCEPT
         SELECT incident_id, entity_type, identifier_type,
                normalized_value, record_id
           FROM public.entity_active_identifier_claims)
        UNION ALL
        (SELECT incident_id, entity_type, identifier_type,
                normalized_value, record_id
           FROM public.entity_active_identifier_claims
         EXCEPT
         SELECT * FROM expected)
    )
    SELECT NOT EXISTS (SELECT 1 FROM difference)
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_expected_active_identifier_claims_v1(uuid)
    FROM PUBLIC;
REVOKE ALL ON FUNCTION
    public.entities_refresh_active_identifier_claims_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION
    public.entities_release_active_identifier_claims_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION public.entities_sync_active_identifier_claims_v1()
    FROM PUBLIC;
REVOKE ALL ON FUNCTION public.entities_rebuild_active_identifier_claims_v1()
    FROM PUBLIC;
REVOKE ALL ON FUNCTION
    public.entities_active_identifier_claims_are_valid_v1() FROM PUBLIC;

CREATE TRIGGER hosts_sync_active_identifier_claims
AFTER INSERT OR DELETE OR UPDATE ON public.hosts
FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();
ALTER TABLE public.hosts
    ENABLE ALWAYS TRIGGER hosts_sync_active_identifier_claims;

CREATE TRIGGER identities_sync_active_identifier_claims
AFTER INSERT OR DELETE OR UPDATE ON public.identities
FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();
ALTER TABLE public.identities
    ENABLE ALWAYS TRIGGER identities_sync_active_identifier_claims;

CREATE TRIGGER entity_preserved_identifiers_sync_active_claims
AFTER INSERT OR DELETE OR UPDATE ON public.entity_preserved_identifiers
FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();
ALTER TABLE public.entity_preserved_identifiers
    ENABLE ALWAYS TRIGGER entity_preserved_identifiers_sync_active_claims;

CREATE TRIGGER records_sync_entity_active_identifier_claims
AFTER INSERT OR DELETE OR UPDATE OF record_id, incident_id, record_type, deleted_at
ON public.records
FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();
ALTER TABLE public.records
    ENABLE ALWAYS TRIGGER records_sync_entity_active_identifier_claims;

SELECT public.entities_rebuild_active_identifier_claims_v1();

-- Exact matching is now served exclusively by the active-claim primary key.
-- These partial source indexes existed only for the superseded incident-wide
-- matcher; retained display, lineage, foreign-key, and record-local indexes
-- continue to serve their independent query shapes.
DROP INDEX public.entity_preserved_identifiers_exact_lookup_idx;
DROP INDEX public.hosts_incident_aad_device_id_idx;
DROP INDEX public.hosts_incident_fqdn_idx;
DROP INDEX public.hosts_incident_hostname_idx;
DROP INDEX public.identities_incident_aad_object_id_idx;
DROP INDEX public.identities_incident_email_idx;
DROP INDEX public.identities_incident_sam_account_name_idx;
DROP INDEX public.identities_incident_sid_idx;
DROP INDEX public.identities_incident_upn_idx;

REVOKE ALL ON TABLE public.entity_active_identifier_claims
    FROM cartulary_runtime, cartulary_recovery;
REVOKE ALL ON TYPE public.entity_active_identifier_claims FROM PUBLIC;
GRANT SELECT ON TABLE public.entity_active_identifier_claims
    TO cartulary_runtime, cartulary_recovery;
GRANT TRUNCATE ON TABLE public.entity_active_identifier_claims
    TO cartulary_recovery;
GRANT EXECUTE ON FUNCTION
	public.entities_release_active_identifier_claims_v1(uuid),
	public.entities_refresh_active_identifier_claims_v1(uuid)
TO cartulary_runtime;
GRANT EXECUTE ON FUNCTION
    public.entities_rebuild_active_identifier_claims_v1(),
    public.entities_active_identifier_claims_are_valid_v1()
TO cartulary_recovery;

-- +goose Down

DROP TRIGGER records_sync_entity_active_identifier_claims ON public.records;
DROP TRIGGER entity_preserved_identifiers_sync_active_claims
    ON public.entity_preserved_identifiers;
DROP TRIGGER identities_sync_active_identifier_claims ON public.identities;
DROP TRIGGER hosts_sync_active_identifier_claims ON public.hosts;

DROP FUNCTION public.entities_active_identifier_claims_are_valid_v1();
DROP FUNCTION public.entities_rebuild_active_identifier_claims_v1();
DROP FUNCTION public.entities_sync_active_identifier_claims_v1();
DROP FUNCTION public.entities_release_active_identifier_claims_v1(uuid);
DROP FUNCTION public.entities_refresh_active_identifier_claims_v1(uuid);
DROP FUNCTION public.entities_expected_active_identifier_claims_v1(uuid);
DROP TABLE public.entity_active_identifier_claims;
DROP FUNCTION public.entities_normalize_identifier_v1(text, text);

CREATE INDEX entity_preserved_identifiers_exact_lookup_idx
    ON public.entity_preserved_identifiers (
        incident_id,
        entity_type,
        identifier_type,
        normalized_value,
        record_id
    )
    WHERE deleted_at IS NULL
      AND classification = 'exact_match_reuse';
CREATE INDEX hosts_incident_aad_device_id_idx
    ON public.hosts (incident_id, aad_device_id, record_id)
    WHERE aad_device_id IS NOT NULL;
CREATE INDEX hosts_incident_fqdn_idx
    ON public.hosts (incident_id, fqdn, record_id)
    WHERE fqdn IS NOT NULL;
CREATE INDEX hosts_incident_hostname_idx
    ON public.hosts (incident_id, hostname, record_id)
    WHERE hostname IS NOT NULL;
CREATE INDEX identities_incident_aad_object_id_idx
    ON public.identities (incident_id, aad_object_id, record_id)
    WHERE aad_object_id IS NOT NULL;
CREATE INDEX identities_incident_email_idx
    ON public.identities (incident_id, email, record_id)
    WHERE email IS NOT NULL;
CREATE INDEX identities_incident_sam_account_name_idx
    ON public.identities (incident_id, sam_account_name, record_id)
    WHERE sam_account_name IS NOT NULL;
CREATE INDEX identities_incident_sid_idx
    ON public.identities (incident_id, sid, record_id)
    WHERE sid IS NOT NULL;
CREATE INDEX identities_incident_upn_idx
    ON public.identities (incident_id, upn, record_id)
    WHERE upn IS NOT NULL;

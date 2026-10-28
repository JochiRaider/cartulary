-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: parties.
-- Administrator provisioning and final admission are separate lifecycle steps.

-- +goose StatementBegin
CREATE FUNCTION public.parties_trim_unicode_space_v1(candidate text)
 RETURNS text
 LANGUAGE sql
 IMMUTABLE PARALLEL SAFE STRICT
 SET search_path = pg_catalog, public
RETURN btrim(candidate, ((((((((((((((((((((((((chr(9) || chr(10)) || chr(11)) || chr(12)) || chr(13)) || chr(32)) || chr(133)) || chr(160)) || chr(5760)) || chr(8192)) || chr(8193)) || chr(8194)) || chr(8195)) || chr(8196)) || chr(8197)) || chr(8198)) || chr(8199)) || chr(8200)) || chr(8201)) || chr(8202)) || chr(8232)) || chr(8233)) || chr(8239)) || chr(8287)) || chr(12288)));
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_trim_unicode_space_v1(candidate text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.parties_trim_unicode_space_v1(candidate text) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.parties_trim_unicode_space_v1(candidate text) TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.parties_normalize_active_key_v1(key_kind text, raw_value text)
 RETURNS text
 LANGUAGE plpgsql
 IMMUTABLE PARALLEL SAFE
 SET search_path = pg_catalog, public
AS $$
DECLARE
    candidate text;
    codepoint integer;
    character_ordinal integer;
BEGIN
    IF raw_value IS NULL OR key_kind NOT IN ('primary_email', 'external_ref') THEN
        RETURN NULL;
    END IF;
    candidate := normalize(public.parties_trim_unicode_space_v1(raw_value), NFC);
    IF candidate = '' THEN
        RETURN NULL;
    END IF;
    IF (key_kind = 'primary_email' AND char_length(candidate) > 320)
        OR (key_kind = 'external_ref' AND char_length(candidate) > 1024) THEN
        RETURN NULL;
    END IF;
    FOR character_ordinal IN 1..char_length(candidate) LOOP
        codepoint := ascii(substring(candidate FROM character_ordinal FOR 1));
        IF codepoint BETWEEN 0 AND 31 OR codepoint BETWEEN 127 AND 159 THEN
            RETURN NULL;
        END IF;
    END LOOP;
    IF key_kind = 'primary_email' THEN
        IF candidate !~ '^[^[:space:]@]+@[^[:space:]@]+$' THEN
            RETURN NULL;
        END IF;
        RETURN lower(candidate COLLATE "C");
    END IF;
    RETURN candidate;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_normalize_active_key_v1(key_kind text, raw_value text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.parties_normalize_active_key_v1(key_kind text, raw_value text) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.parties_normalize_active_key_v1(key_kind text, raw_value text) TO cartulary_recovery;

CREATE TABLE public.parties (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    display_name text,
    party_kind text,
    organization_name text,
    role_title text,
    primary_email text,
    timezone_name text,
    external_ref text,
    notes text,
    CONSTRAINT parties_external_ref_claim_value_ck CHECK (external_ref IS NULL OR public.parties_normalize_active_key_v1('external_ref'::text, external_ref) IS NOT NULL),
    CONSTRAINT parties_pkey PRIMARY KEY (record_id),
    CONSTRAINT parties_primary_email_claim_value_ck CHECK (primary_email IS NULL OR public.parties_normalize_active_key_v1('primary_email'::text, primary_email) IS NOT NULL),
    CONSTRAINT parties_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.parties FROM PUBLIC;

REVOKE ALL ON TYPE public.parties FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.parties TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.parties TO cartulary_recovery;

CREATE TABLE public.party_active_key_claims (
    incident_id uuid NOT NULL,
    key_kind text NOT NULL,
    normalized_value text NOT NULL,
    party_record_id uuid NOT NULL,
    CONSTRAINT party_active_key_claims_key_kind_ck CHECK (key_kind = ANY (ARRAY['primary_email'::text, 'external_ref'::text])),
    CONSTRAINT party_active_key_claims_pkey PRIMARY KEY (incident_id, key_kind, normalized_value),
    CONSTRAINT party_active_key_claims_record_key_unique UNIQUE (incident_id, party_record_id, key_kind),
    CONSTRAINT party_active_key_claims_value_ck CHECK (normalized_value <> ''::text AND public.parties_normalize_active_key_v1(key_kind, normalized_value) = normalized_value)
);

REVOKE ALL ON TABLE public.party_active_key_claims FROM PUBLIC;

REVOKE ALL ON TYPE public.party_active_key_claims FROM PUBLIC;

GRANT SELECT ON TABLE public.party_active_key_claims TO cartulary_runtime;

GRANT SELECT, TRUNCATE ON TABLE public.party_active_key_claims TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.parties_expected_active_key_claims_v1(target_party_record_id uuid)
 RETURNS TABLE(incident_id uuid, key_kind text, normalized_value text, party_record_id uuid)
 LANGUAGE sql
 STABLE
 SET search_path = pg_catalog, public
AS $$
    SELECT p.incident_id,
           candidate.key_kind,
           public.parties_normalize_active_key_v1(
               candidate.key_kind,
               candidate.raw_value
           ) AS normalized_value,
           p.record_id AS party_record_id
      FROM public.parties p
      JOIN public.records r
        ON r.record_id = p.record_id
       AND r.incident_id = p.incident_id
       AND r.record_type = 'party'
      CROSS JOIN LATERAL (VALUES
          ('primary_email'::text, p.primary_email),
          ('external_ref'::text, p.external_ref)
      ) AS candidate(key_kind, raw_value)
     WHERE r.deleted_at IS NULL
       AND candidate.raw_value IS NOT NULL
       AND (target_party_record_id IS NULL
            OR p.record_id = target_party_record_id)
     ORDER BY p.incident_id,
              candidate.key_kind COLLATE "C",
              public.parties_normalize_active_key_v1(
                  candidate.key_kind,
                  candidate.raw_value
              ) COLLATE "C",
              p.record_id
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_expected_active_key_claims_v1(target_party_record_id uuid) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.parties_release_active_key_claims_v1(target_party_record_id uuid)
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    released_count bigint;
BEGIN
    DELETE FROM public.party_active_key_claims
     WHERE party_record_id = target_party_record_id;
    GET DIAGNOSTICS released_count = ROW_COUNT;
    RETURN released_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_release_active_key_claims_v1(target_party_record_id uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.parties_release_active_key_claims_v1(target_party_record_id uuid) TO cartulary_runtime;

-- +goose StatementBegin
CREATE FUNCTION public.parties_active_key_claims_are_valid_v1()
 RETURNS boolean
 LANGUAGE sql
 STABLE SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
    WITH expected AS (
        SELECT *
          FROM public.parties_expected_active_key_claims_v1(NULL::uuid)
    ), difference AS (
        (SELECT * FROM expected
         EXCEPT
         SELECT incident_id, key_kind, normalized_value, party_record_id
           FROM public.party_active_key_claims)
        UNION ALL
        (SELECT incident_id, key_kind, normalized_value, party_record_id
           FROM public.party_active_key_claims
         EXCEPT
         SELECT * FROM expected)
    )
    SELECT NOT EXISTS (SELECT 1 FROM difference)
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_active_key_claims_are_valid_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.parties_active_key_claims_are_valid_v1() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.parties_rebuild_active_key_claims_v1()
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    rebuilt_count bigint;
BEGIN
    LOCK TABLE public.records IN SHARE MODE;
    LOCK TABLE public.parties IN SHARE MODE;
    LOCK TABLE public.party_active_key_claims IN EXCLUSIVE MODE;
    DELETE FROM public.party_active_key_claims;
    INSERT INTO public.party_active_key_claims (
        incident_id,
        key_kind,
        normalized_value,
        party_record_id
    )
    SELECT incident_id, key_kind, normalized_value, party_record_id
      FROM public.parties_expected_active_key_claims_v1(NULL::uuid)
     ORDER BY incident_id, key_kind COLLATE "C", normalized_value COLLATE "C";
    GET DIAGNOSTICS rebuilt_count = ROW_COUNT;
    RETURN rebuilt_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_rebuild_active_key_claims_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.parties_rebuild_active_key_claims_v1() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.parties_refresh_active_key_claims_v1(target_party_record_id uuid)
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    refreshed_count bigint;
BEGIN
    DELETE FROM public.party_active_key_claims
     WHERE party_record_id = target_party_record_id;
    INSERT INTO public.party_active_key_claims (
        incident_id,
        key_kind,
        normalized_value,
        party_record_id
    )
    SELECT expected.incident_id,
           expected.key_kind,
           expected.normalized_value,
           expected.party_record_id
      FROM public.parties_expected_active_key_claims_v1(
          target_party_record_id
      ) expected
     ORDER BY expected.incident_id, expected.key_kind COLLATE "C",
              expected.normalized_value COLLATE "C";
    GET DIAGNOSTICS refreshed_count = ROW_COUNT;
    RETURN refreshed_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_refresh_active_key_claims_v1(target_party_record_id uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.parties_refresh_active_key_claims_v1(target_party_record_id uuid) TO cartulary_runtime;

-- +goose StatementBegin
CREATE FUNCTION public.parties_sync_active_key_claims_v1()
 RETURNS trigger
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF pg_catalog.current_setting(
        'cartulary.parties_defer_active_key_claims',
        true
    ) = 'on' THEN
        RETURN COALESCE(NEW, OLD);
    END IF;
    IF TG_OP = 'DELETE' THEN
        PERFORM public.parties_refresh_active_key_claims_v1(OLD.record_id);
        RETURN OLD;
    END IF;
    IF TG_OP = 'UPDATE' AND OLD.record_id IS DISTINCT FROM NEW.record_id THEN
        PERFORM public.parties_refresh_active_key_claims_v1(OLD.record_id);
    END IF;
    PERFORM public.parties_refresh_active_key_claims_v1(NEW.record_id);
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.parties_sync_active_key_claims_v1() FROM PUBLIC;

CREATE INDEX parties_incident_display_name_idx ON public.parties USING btree (incident_id, display_name, record_id);

CREATE INDEX parties_incident_id_record_id_fk_idx ON public.parties USING btree (incident_id, record_id);

CREATE INDEX party_active_key_claims_record_lookup_idx ON public.party_active_key_claims USING btree (party_record_id, key_kind, normalized_value);

ALTER TABLE public.parties ADD CONSTRAINT parties_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.parties ADD CONSTRAINT parties_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.party_active_key_claims ADD CONSTRAINT party_active_key_claims_record_envelope_fkey FOREIGN KEY (incident_id, party_record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

CREATE TRIGGER parties_sync_active_key_claims AFTER INSERT OR DELETE OR UPDATE OF record_id, incident_id, primary_email, external_ref ON public.parties FOR EACH ROW EXECUTE FUNCTION public.parties_sync_active_key_claims_v1();

ALTER TABLE public.parties ENABLE ALWAYS TRIGGER parties_sync_active_key_claims;

CREATE TRIGGER records_sync_party_active_key_claims AFTER INSERT OR DELETE OR UPDATE OF record_id, incident_id, record_type, deleted_at ON public.records FOR EACH ROW EXECUTE FUNCTION public.parties_sync_active_key_claims_v1();

ALTER TABLE public.records ENABLE ALWAYS TRIGGER records_sync_party_active_key_claims;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER records_sync_party_active_key_claims ON public.records;
DROP TRIGGER parties_sync_active_key_claims ON public.parties;
ALTER TABLE public.party_active_key_claims DROP CONSTRAINT party_active_key_claims_record_envelope_fkey;
ALTER TABLE public.parties DROP CONSTRAINT parties_record_envelope_fkey;
ALTER TABLE public.parties DROP CONSTRAINT parties_incident_id_fkey;
DROP FUNCTION public.parties_sync_active_key_claims_v1();
DROP FUNCTION public.parties_refresh_active_key_claims_v1(target_party_record_id uuid);
DROP FUNCTION public.parties_rebuild_active_key_claims_v1();
DROP FUNCTION public.parties_active_key_claims_are_valid_v1();
DROP FUNCTION public.parties_release_active_key_claims_v1(target_party_record_id uuid);
DROP FUNCTION public.parties_expected_active_key_claims_v1(target_party_record_id uuid);
DROP TABLE public.party_active_key_claims;
DROP TABLE public.parties;
DROP FUNCTION public.parties_normalize_active_key_v1(key_kind text, raw_value text);
DROP FUNCTION public.parties_trim_unicode_space_v1(candidate text);

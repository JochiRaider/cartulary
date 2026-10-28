-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: indicators.
-- Administrator provisioning and final admission are separate lifecycle steps.

-- +goose StatementBegin
CREATE FUNCTION public.indicator_support_refs_are_valid(candidate jsonb)
 RETURNS boolean
 LANGUAGE plpgsql
 IMMUTABLE STRICT
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

REVOKE ALL ON FUNCTION public.indicator_support_refs_are_valid(candidate jsonb) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.indicator_support_refs_are_valid(candidate jsonb) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.indicator_support_refs_are_valid(candidate jsonb) TO cartulary_recovery;

CREATE TABLE public.indicator_active_identities (
    incident_id uuid NOT NULL,
    indicator_type text NOT NULL,
    dedupe_key text NOT NULL,
    indicator_record_id uuid NOT NULL,
    dedupe_sha256 bytea GENERATED ALWAYS AS (public.digest(dedupe_key, 'sha256'::text)) STORED NOT NULL,
    CONSTRAINT indicator_active_identities_dedupe_ck CHECK (octet_length(dedupe_key) > 0 AND dedupe_key !~ '[[:cntrl:]]'::text),
    CONSTRAINT indicator_active_identities_exact_dedupe_ck CHECK ("left"(dedupe_key, length(indicator_type) + 1) = (indicator_type || ':'::text) AND octet_length(dedupe_key) >= 1 AND octet_length(dedupe_key) <= 32784),
    CONSTRAINT indicator_active_identities_exact_type_ck CHECK (indicator_type = ANY (ARRAY['ipv4_addr'::text, 'ipv6_addr'::text, 'domain_name'::text, 'url'::text, 'sha256'::text, 'email_addr'::text, 'registry_key'::text, 'process_name'::text, 'text'::text])),
    CONSTRAINT indicator_active_identities_pkey PRIMARY KEY (incident_id, indicator_type, dedupe_sha256),
    CONSTRAINT indicator_active_identities_record_key UNIQUE (indicator_record_id),
    CONSTRAINT indicator_active_identities_type_ck CHECK (octet_length(indicator_type) > 0 AND indicator_type !~ '[[:cntrl:]]'::text)
);

REVOKE ALL ON TABLE public.indicator_active_identities FROM PUBLIC;

REVOKE ALL ON TYPE public.indicator_active_identities FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.indicator_active_identities TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.indicator_active_identities TO cartulary_recovery;

CREATE TABLE public.indicator_observations (
    indicator_observation_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    source_record_id uuid NOT NULL,
    source_field_key text NOT NULL,
    origin_kind text NOT NULL,
    origin_locator text NOT NULL,
    observed_text text NOT NULL,
    parsed_indicator_type text,
    normalized_candidate text,
    resolution_status text NOT NULL,
    resolved_indicator_record_id uuid,
    row_version bigint DEFAULT 1 NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    resolved_by_user_id uuid,
    resolved_at timestamp with time zone,
    resolution_method text,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    CONSTRAINT indicator_observations_candidate_pair_ck CHECK ((parsed_indicator_type IS NULL) = (normalized_candidate IS NULL)),
    CONSTRAINT indicator_observations_origin_kind_ck CHECK (origin_kind = ANY (ARRAY['manual_entry'::text, 'clipboard_paste'::text, 'csv_import'::text, 'xlsx_import'::text, 'api_import'::text, 'extraction'::text, 'system'::text])),
    CONSTRAINT indicator_observations_parsed_type_ck CHECK (parsed_indicator_type IS NULL OR (parsed_indicator_type = ANY (ARRAY['ipv4_addr'::text, 'ipv6_addr'::text, 'domain_name'::text, 'url'::text, 'sha256'::text, 'email_addr'::text, 'registry_key'::text, 'process_name'::text, 'text'::text]))),
    CONSTRAINT indicator_observations_pkey PRIMARY KEY (indicator_observation_id),
    CONSTRAINT indicator_observations_required_text_ck CHECK (octet_length(source_field_key) > 0 AND octet_length(origin_locator) > 0 AND octet_length(observed_text) > 0 AND (normalized_candidate IS NULL OR octet_length(normalized_candidate) > 0) AND (resolution_method IS NULL OR octet_length(resolution_method) > 0)),
    CONSTRAINT indicator_observations_resolution_status_check CHECK (resolution_status = ANY (ARRAY['unresolved'::text, 'resolved'::text, 'dismissed'::text])),
    CONSTRAINT indicator_observations_resolution_tuple_ck CHECK (resolution_status = 'unresolved'::text AND resolved_indicator_record_id IS NULL AND resolved_by_user_id IS NULL AND resolved_at IS NULL AND resolution_method IS NULL OR resolution_status = 'resolved'::text AND resolved_indicator_record_id IS NOT NULL AND resolved_by_user_id IS NOT NULL AND resolved_at IS NOT NULL AND resolution_method IS NOT NULL OR resolution_status = 'dismissed'::text AND resolved_indicator_record_id IS NULL AND resolved_by_user_id IS NOT NULL AND resolved_at IS NOT NULL AND resolution_method IS NOT NULL),
    CONSTRAINT indicator_observations_row_version_ck CHECK (row_version >= 1),
    CONSTRAINT indicator_observations_tombstone_pair_ck CHECK ((deleted_at IS NULL) = (deleted_by_user_id IS NULL))
);

REVOKE ALL ON TABLE public.indicator_observations FROM PUBLIC;

REVOKE ALL ON TYPE public.indicator_observations FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.indicator_observations TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.indicator_observations TO cartulary_recovery;

CREATE TABLE public.indicator_state_intervals (
    indicator_state_interval_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    indicator_record_id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    valid_from timestamp with time zone NOT NULL,
    valid_to timestamp with time zone,
    confidence integer,
    rationale text,
    support_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    assessor text,
    assessed_at timestamp with time zone DEFAULT now() NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    CONSTRAINT indicator_state_intervals_confidence_check CHECK (confidence IS NULL OR confidence >= 0 AND confidence <= 100),
    CONSTRAINT indicator_state_intervals_lifecycle_state_ck CHECK (lifecycle_state = ANY (ARRAY['active'::text, 'benign'::text, 'false_positive'::text, 'retired'::text])),
    CONSTRAINT indicator_state_intervals_pkey PRIMARY KEY (indicator_state_interval_id),
    CONSTRAINT indicator_state_intervals_required_text_ck CHECK ((rationale IS NULL OR octet_length(rationale) > 0) AND (assessor IS NULL OR octet_length(assessor) > 0)),
    CONSTRAINT indicator_state_intervals_row_version_ck CHECK (row_version >= 1),
    CONSTRAINT indicator_state_intervals_support_refs_ck CHECK (public.indicator_support_refs_are_valid(support_refs)),
    CONSTRAINT indicator_state_intervals_tombstone_pair_ck CHECK ((deleted_at IS NULL) = (deleted_by_user_id IS NULL)),
    CONSTRAINT indicator_state_intervals_validity_ck CHECK (valid_to IS NULL OR valid_to >= valid_from)
);

REVOKE ALL ON TABLE public.indicator_state_intervals FROM PUBLIC;

REVOKE ALL ON TYPE public.indicator_state_intervals FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.indicator_state_intervals TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.indicator_state_intervals TO cartulary_recovery;

CREATE TABLE public.indicators (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    indicator_type text NOT NULL,
    value_kind text NOT NULL,
    display_value text NOT NULL,
    normalized_value text NOT NULL,
    dedupe_key text NOT NULL,
    defanged_value text,
    hash_algorithm text,
    hash_value text,
    stix_pattern text,
    CONSTRAINT indicators_dedupe_key_ck CHECK (dedupe_key =
CASE
    WHEN indicator_type = 'text'::text THEN 'text:'::text || encode(public.digest(normalized_value, 'sha256'::text), 'hex'::text)
    ELSE (indicator_type || ':'::text) || normalized_value
END AND octet_length(dedupe_key) >= 1 AND octet_length(dedupe_key) <= 32784),
    CONSTRAINT indicators_hash_pair_ck CHECK (hash_algorithm IS NULL AND hash_value IS NULL OR hash_algorithm IS NOT NULL AND hash_value IS NOT NULL),
    CONSTRAINT indicators_hash_value_ck CHECK (hash_value IS NULL OR hash_value ~ '^[0-9a-f]+$'::text),
    CONSTRAINT indicators_incident_record_key UNIQUE (incident_id, record_id),
    CONSTRAINT indicators_indicator_type_ck CHECK (indicator_type = ANY (ARRAY['ipv4_addr'::text, 'ipv6_addr'::text, 'domain_name'::text, 'url'::text, 'sha256'::text, 'email_addr'::text, 'registry_key'::text, 'process_name'::text, 'text'::text])),
    CONSTRAINT indicators_ip_value_kind_ck CHECK ((indicator_type <> ALL (ARRAY['ipv4_addr'::text, 'ipv6_addr'::text])) OR value_kind = 'atomic'::text),
    CONSTRAINT indicators_pkey PRIMARY KEY (record_id),
    CONSTRAINT indicators_registry_value_kind_ck CHECK (value_kind = 'atomic'::text OR value_kind = 'reference'::text AND (indicator_type = ANY (ARRAY['url'::text, 'registry_key'::text, 'process_name'::text, 'text'::text])) OR value_kind = 'pattern'::text AND indicator_type = 'text'::text),
    CONSTRAINT indicators_required_text_ck CHECK (octet_length(display_value) > 0 AND octet_length(dedupe_key) > 0 AND (normalized_value IS NULL OR octet_length(normalized_value) > 0) AND (defanged_value IS NULL OR octet_length(defanged_value) > 0) AND (hash_algorithm IS NULL OR octet_length(hash_algorithm) > 0) AND (hash_value IS NULL OR octet_length(hash_value) > 0) AND (stix_pattern IS NULL OR octet_length(stix_pattern) > 0)),
    CONSTRAINT indicators_value_kind_check CHECK (value_kind = ANY (ARRAY['atomic'::text, 'pattern'::text, 'reference'::text]))
);

REVOKE ALL ON TABLE public.indicators FROM PUBLIC;

REVOKE ALL ON TYPE public.indicators FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.indicators TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.indicators TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.enforce_indicator_support_ref_incident()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
DECLARE
    member jsonb;
    support_record_id uuid;
BEGIN
    IF NOT public.indicator_support_refs_are_valid(NEW.support_refs) THEN
        RAISE EXCEPTION 'indicator lifecycle support references are malformed'
            USING ERRCODE = '23514', CONSTRAINT = 'indicator_state_intervals_support_refs_ck';
    END IF;
    FOR member IN SELECT value FROM jsonb_array_elements(NEW.support_refs)
    LOOP
        support_record_id := (member #>> '{}')::uuid;
        IF NOT (EXISTS (
            SELECT 1 FROM public.records
             WHERE incident_id = NEW.incident_id AND record_id = support_record_id
        )) THEN
            RAISE EXCEPTION 'indicator lifecycle support reference is outside the incident'
                USING ERRCODE = '23514';
        END IF;
    END LOOP;
    RETURN NEW;
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.enforce_indicator_support_ref_incident() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.enforce_indicator_support_ref_incident() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.enforce_indicator_support_ref_incident() TO cartulary_recovery;

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
    AND NOT EXISTS (
        SELECT 1 FROM public.indicator_state_intervals i,
            LATERAL jsonb_array_elements_text(i.support_refs) ref(id)
        WHERE NOT EXISTS (SELECT 1 FROM public.records r
            WHERE r.incident_id=i.incident_id AND r.record_id=ref.id::uuid)
    )
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.indicator_active_identities_are_valid() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.indicator_active_identities_are_valid() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.rebuild_indicator_active_identities()
 RETURNS bigint
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
DECLARE
    rebuilt_count bigint;
BEGIN
    LOCK TABLE public.records IN SHARE MODE;
    LOCK TABLE public.indicators IN SHARE MODE;
    LOCK TABLE public.indicator_active_identities IN EXCLUSIVE MODE;
    DELETE FROM public.indicator_active_identities;
    INSERT INTO public.indicator_active_identities (
        incident_id, indicator_type, dedupe_key, indicator_record_id
    )
    SELECT indicator.incident_id, indicator.indicator_type, indicator.dedupe_key, indicator.record_id
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

REVOKE ALL ON FUNCTION public.rebuild_indicator_active_identities() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.rebuild_indicator_active_identities() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.sync_indicator_active_identity_from_indicator()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        DELETE FROM public.indicator_active_identities WHERE indicator_record_id = OLD.record_id;
        RETURN OLD;
    END IF;
    IF TG_OP = 'UPDATE'
       AND NEW.incident_id = OLD.incident_id
       AND NEW.indicator_type = OLD.indicator_type
       AND NEW.dedupe_key = OLD.dedupe_key THEN
        RETURN NEW;
    END IF;
    DELETE FROM public.indicator_active_identities WHERE indicator_record_id = NEW.record_id;
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

REVOKE ALL ON FUNCTION public.sync_indicator_active_identity_from_indicator() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.sync_indicator_active_identity_from_indicator() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.sync_indicator_active_identity_from_indicator() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.sync_indicator_active_identity_from_record()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.record_type = 'indicator' THEN
            DELETE FROM public.indicator_active_identities WHERE indicator_record_id = OLD.record_id;
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
            DELETE FROM public.indicator_active_identities WHERE indicator_record_id = OLD.record_id;
        END IF;
        RETURN NEW;
    END IF;
    DELETE FROM public.indicator_active_identities WHERE indicator_record_id = NEW.record_id;
    IF NEW.deleted_at IS NULL THEN
        INSERT INTO public.indicator_active_identities (
            incident_id, indicator_type, dedupe_key, indicator_record_id
        )
        SELECT indicator.incident_id, indicator.indicator_type, indicator.dedupe_key, indicator.record_id
          FROM public.indicators AS indicator
         WHERE indicator.record_id = NEW.record_id AND indicator.incident_id = NEW.incident_id;
    END IF;
    RETURN NEW;
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.sync_indicator_active_identity_from_record() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.sync_indicator_active_identity_from_record() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.sync_indicator_active_identity_from_record() TO cartulary_recovery;

CREATE INDEX indicator_active_identities_incident_id_indicator_reco_f4132796 ON public.indicator_active_identities USING btree (incident_id, indicator_record_id);

CREATE INDEX indicator_observations_candidate_lookup_idx ON public.indicator_observations USING btree (incident_id, parsed_indicator_type, normalized_candidate, indicator_observation_id) WHERE ((normalized_candidate IS NOT NULL) AND (deleted_at IS NULL));

CREATE INDEX indicator_observations_created_by_user_id_fk_idx ON public.indicator_observations USING btree (created_by_user_id);

CREATE INDEX indicator_observations_deleted_by_user_id_fk_idx ON public.indicator_observations USING btree (deleted_by_user_id);

CREATE INDEX indicator_observations_incident_id_resolved_indicator__f4dab861 ON public.indicator_observations USING btree (incident_id, resolved_indicator_record_id);

CREATE INDEX indicator_observations_incident_id_source_record_id_fk_idx ON public.indicator_observations USING btree (incident_id, source_record_id);

CREATE INDEX indicator_observations_resolved_by_user_id_fk_idx ON public.indicator_observations USING btree (resolved_by_user_id);

CREATE INDEX indicator_observations_resolved_lookup_idx ON public.indicator_observations USING btree (incident_id, resolution_status, resolved_indicator_record_id, created_at) WHERE (deleted_at IS NULL);

CREATE INDEX indicator_observations_source_lookup_idx ON public.indicator_observations USING btree (source_record_id, source_field_key, created_at, indicator_observation_id) WHERE (deleted_at IS NULL);

CREATE INDEX indicator_state_intervals_created_by_user_id_fk_idx ON public.indicator_state_intervals USING btree (created_by_user_id);

CREATE INDEX indicator_state_intervals_deleted_by_user_id_fk_idx ON public.indicator_state_intervals USING btree (deleted_by_user_id);

CREATE INDEX indicator_state_intervals_incident_id_fk_idx ON public.indicator_state_intervals USING btree (incident_id);

CREATE INDEX indicator_state_intervals_indicator_incident_fk_idx ON public.indicator_state_intervals USING btree (incident_id, indicator_record_id);

CREATE INDEX indicator_state_intervals_indicator_lookup_idx ON public.indicator_state_intervals USING btree (incident_id, indicator_record_id, valid_from DESC, indicator_state_interval_id DESC) WHERE (deleted_at IS NULL);

CREATE INDEX indicators_incident_normalized_lookup_idx ON public.indicators USING btree (incident_id, indicator_type, public.digest(normalized_value, 'sha256'::text), record_id);

ALTER TABLE public.indicator_active_identities ADD CONSTRAINT indicator_active_identities_indicator_fkey FOREIGN KEY (incident_id, indicator_record_id) REFERENCES public.indicators(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicator_active_identities ADD CONSTRAINT indicator_active_identities_record_envelope_fkey FOREIGN KEY (incident_id, indicator_record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicator_observations ADD CONSTRAINT indicator_observations_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.indicator_observations ADD CONSTRAINT indicator_observations_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.indicator_observations ADD CONSTRAINT indicator_observations_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicator_observations ADD CONSTRAINT indicator_observations_resolved_by_user_id_fkey FOREIGN KEY (resolved_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.indicator_observations ADD CONSTRAINT indicator_observations_resolved_indicator_incident_fkey FOREIGN KEY (incident_id, resolved_indicator_record_id) REFERENCES public.indicators(incident_id, record_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.indicator_observations ADD CONSTRAINT indicator_observations_source_record_envelope_fkey FOREIGN KEY (incident_id, source_record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicator_state_intervals ADD CONSTRAINT indicator_state_intervals_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.indicator_state_intervals ADD CONSTRAINT indicator_state_intervals_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.indicator_state_intervals ADD CONSTRAINT indicator_state_intervals_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicator_state_intervals ADD CONSTRAINT indicator_state_intervals_indicator_incident_fkey FOREIGN KEY (incident_id, indicator_record_id) REFERENCES public.indicators(incident_id, record_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.indicators ADD CONSTRAINT indicators_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicators ADD CONSTRAINT indicators_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

CREATE TRIGGER indicator_state_intervals_support_ref_incident BEFORE INSERT OR UPDATE ON public.indicator_state_intervals FOR EACH ROW EXECUTE FUNCTION public.enforce_indicator_support_ref_incident();

CREATE TRIGGER indicators_sync_active_identity AFTER INSERT OR DELETE OR UPDATE ON public.indicators FOR EACH ROW EXECUTE FUNCTION public.sync_indicator_active_identity_from_indicator();

ALTER TABLE public.indicators ENABLE ALWAYS TRIGGER indicators_sync_active_identity;

CREATE TRIGGER records_sync_indicator_active_identity AFTER INSERT OR DELETE OR UPDATE ON public.records FOR EACH ROW EXECUTE FUNCTION public.sync_indicator_active_identity_from_record();

ALTER TABLE public.records ENABLE ALWAYS TRIGGER records_sync_indicator_active_identity;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER records_sync_indicator_active_identity ON public.records;
DROP TRIGGER indicators_sync_active_identity ON public.indicators;
DROP TRIGGER indicator_state_intervals_support_ref_incident ON public.indicator_state_intervals;
ALTER TABLE public.indicators DROP CONSTRAINT indicators_record_envelope_fkey;
ALTER TABLE public.indicators DROP CONSTRAINT indicators_incident_id_fkey;
ALTER TABLE public.indicator_state_intervals DROP CONSTRAINT indicator_state_intervals_indicator_incident_fkey;
ALTER TABLE public.indicator_state_intervals DROP CONSTRAINT indicator_state_intervals_incident_id_fkey;
ALTER TABLE public.indicator_state_intervals DROP CONSTRAINT indicator_state_intervals_deleted_by_user_id_fkey;
ALTER TABLE public.indicator_state_intervals DROP CONSTRAINT indicator_state_intervals_created_by_user_id_fkey;
ALTER TABLE public.indicator_observations DROP CONSTRAINT indicator_observations_source_record_envelope_fkey;
ALTER TABLE public.indicator_observations DROP CONSTRAINT indicator_observations_resolved_indicator_incident_fkey;
ALTER TABLE public.indicator_observations DROP CONSTRAINT indicator_observations_resolved_by_user_id_fkey;
ALTER TABLE public.indicator_observations DROP CONSTRAINT indicator_observations_incident_id_fkey;
ALTER TABLE public.indicator_observations DROP CONSTRAINT indicator_observations_deleted_by_user_id_fkey;
ALTER TABLE public.indicator_observations DROP CONSTRAINT indicator_observations_created_by_user_id_fkey;
ALTER TABLE public.indicator_active_identities DROP CONSTRAINT indicator_active_identities_record_envelope_fkey;
ALTER TABLE public.indicator_active_identities DROP CONSTRAINT indicator_active_identities_indicator_fkey;
DROP FUNCTION public.sync_indicator_active_identity_from_record();
DROP FUNCTION public.sync_indicator_active_identity_from_indicator();
DROP FUNCTION public.rebuild_indicator_active_identities();
DROP FUNCTION public.indicator_active_identities_are_valid();
DROP FUNCTION public.enforce_indicator_support_ref_incident();
DROP TABLE public.indicators;
DROP TABLE public.indicator_state_intervals;
DROP TABLE public.indicator_observations;
DROP TABLE public.indicator_active_identities;
DROP FUNCTION public.indicator_support_refs_are_valid(candidate jsonb);

-- +goose Up
-- Network Flow row inserts evaluate every row CHECK. PostgreSQL's regex engine
-- expands a bounded repetition such as {64} into one state per repetition, so
-- the SHA-256 patterns dominated the per-row insert cost. Each rewritten
-- predicate pairs an unbounded pattern over the same character class with an
-- exact character count, which accepts exactly the same values, NULL included.
-- The direction-evidence predicate is versioned because routines are not
-- replaced in place; it changes only those two hash tests, and Recovery's
-- orientation check moves to it so CHECK and Recovery still share one predicate.
-- The incident-only foreign key is implied by the composite table key, whose
-- referenced network_flow_tables row references the same incident with the same
-- cascade; direction_idx covers every query of the removed two-column index.
-- A row's creator is always its table's creator (one import actor writes the
-- table and all of its immutable rows), so the creator joins the table key and
-- the row-to-user key is implied: the table still holds the user, and table or
-- incident removal still cascades to rows. The creator key index replaces the
-- user-key index, which served only the removed key. The evidence predicate
-- reads a stored, compressed argument once rather than once per operator, and
-- new evidence and provenance values use lz4. Logical values are unchanged.
-- No data is transformed.
-- +goose StatementBegin
CREATE FUNCTION public.network_flow_direction_evidence_is_valid_v2(e jsonb, kind text, origin text, reason text, identical boolean)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE SET search_path = pg_catalog, public AS $$
DECLARE
    role_a text; role_b text; first_slot text; algorithm text;
    item jsonb; input_key text; expected text[]; ordinals text[] := '{}'; expected_origin text;
BEGIN
    IF jsonb_typeof(e) IS DISTINCT FROM 'object' OR kind IS NULL OR origin IS NULL
       OR reason IS NULL OR identical IS NULL THEN RETURN false; END IF;
    -- e is a non-null object here, and #> '{}' returns it unchanged. Binding the
    -- result reads a stored, compressed argument once instead of once per operator.
    e := e #> '{}';
    IF e - ARRAY['derivation_id','first_endpoint','first_role','second_role','inputs'] <> '{}'::jsonb
       OR NOT (e ?& ARRAY['derivation_id','first_endpoint','first_role','second_role','inputs']) THEN RETURN false; END IF;
    algorithm := e->>'derivation_id'; first_slot := e->>'first_endpoint'; role_a := e->>'first_role'; role_b := e->>'second_role';
    IF jsonb_typeof(e->'derivation_id') <> 'string' OR jsonb_typeof(e->'first_endpoint') <> 'string'
       OR jsonb_typeof(e->'first_role') <> 'string' OR jsonb_typeof(e->'second_role') <> 'string'
       OR first_slot NOT IN ('a','b') OR role_a NOT IN ('','client','server') OR role_b NOT IN ('','client','server')
       OR jsonb_typeof(e->'inputs') <> 'array' THEN RETURN false; END IF;
    IF identical AND first_slot <> 'a' THEN RETURN false; END IF;
    expected := ARRAY['network_flow.input_first_ip','network_flow.input_first_port','network_flow.input_second_ip','network_flow.input_second_port'];
    IF algorithm = 'cisco_flow_search_orientation_v1' THEN
        expected := ARRAY['network_flow.input_first_ip','network_flow.input_first_port','network_flow.input_first_orientation','network_flow.input_second_ip','network_flow.input_second_port','network_flow.input_second_orientation'];
        IF role_a = '' OR role_b = '' THEN
            IF kind <> 'unknown' OR origin <> '' OR reason <> (CASE WHEN role_a = '' AND role_b = '' THEN 'missing_orientation' ELSE 'incomplete_orientation' END) THEN RETURN false; END IF;
        ELSE
            IF role_a = role_b OR kind <> 'client_to_server' OR reason <> '' THEN RETURN false; END IF;
            expected_origin := CASE WHEN identical THEN 'a' WHEN role_a = 'client' THEN first_slot WHEN first_slot = 'a' THEN 'b' ELSE 'a' END;
            IF origin <> expected_origin THEN RETURN false; END IF;
        END IF;
    ELSIF algorithm = 'declared_source_destination_v1' THEN
        IF role_a <> '' OR role_b <> '' OR kind <> 'source_to_destination' OR reason <> '' OR origin <> first_slot THEN RETURN false; END IF;
    ELSE RETURN false;
    END IF;
    IF jsonb_array_length(e->'inputs') <> cardinality(expected) THEN RETURN false; END IF;
    -- The derivation fixes this array at four or six inputs. Avoid per-row
    -- aggregate/iterator queries in the bulk import constraint path.
    FOR input_index IN 1..cardinality(expected) LOOP
        item := e->'inputs'->(input_index-1);
        IF jsonb_typeof(item) IS DISTINCT FROM 'object' THEN RETURN false; END IF;
        IF item - ARRAY['input_key','source_column_ordinal','raw_header_sha256','raw_value_sha256'] <> '{}'::jsonb
           OR NOT (item ?& ARRAY['input_key','source_column_ordinal','raw_header_sha256','raw_value_sha256']) THEN RETURN false; END IF;
        input_key := item->>'input_key';
        IF jsonb_typeof(item->'input_key') <> 'string' OR input_key <> expected[input_index]
           OR jsonb_typeof(item->'source_column_ordinal') <> 'number' OR (item->>'source_column_ordinal') !~ '^[1-9][0-9]*$'
           OR NOT pg_input_is_valid(item->>'source_column_ordinal', 'bigint')
           OR (item->>'source_column_ordinal') = ANY(ordinals)
           OR jsonb_typeof(item->'raw_header_sha256') <> 'string' OR NOT (char_length(item->>'raw_header_sha256') = 64 AND (item->>'raw_header_sha256') ~ '^[0-9a-f]+$')
           OR jsonb_typeof(item->'raw_value_sha256') <> 'string' OR NOT (char_length(item->>'raw_value_sha256') = 64 AND (item->>'raw_value_sha256') ~ '^[0-9a-f]+$') THEN RETURN false; END IF;
        ordinals := array_append(ordinals,item->>'source_column_ordinal');
    END LOOP;
    RETURN true;
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_direction_evidence_is_valid_v2(jsonb,text,text,text,boolean) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_direction_evidence_is_valid_v2(jsonb,text,text,text,boolean) TO cartulary_runtime, cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.network_flow_orientation_state_is_valid_v2()
RETURNS boolean LANGUAGE sql STABLE SET search_path = pg_catalog, public AS $$
 SELECT NOT EXISTS (SELECT 1 FROM public.network_flow_tables WHERE source_profile_id NOT IN ('cisco_sna_netflow_csv_v2','cisco_sna_flow_search_csv_v2'))
 AND NOT EXISTS (
   SELECT 1 FROM public.network_flow_rows r JOIN public.network_flow_tables t USING (network_flow_table_id)
   WHERE r.incident_id <> t.incident_id
      OR public.network_flow_endpoints_are_valid_v1(r.endpoint_a_ip,r.endpoint_a_port,r.endpoint_b_ip,r.endpoint_b_port) IS NOT TRUE
      OR public.network_flow_direction_evidence_is_valid_v2(r.direction_evidence,r.direction_kind,r.from_endpoint,r.direction_reason,r.endpoint_a_ip=r.endpoint_b_ip AND r.endpoint_a_port IS NOT DISTINCT FROM r.endpoint_b_port) IS NOT TRUE
      OR r.direction_evidence->>'derivation_id' <> CASE t.source_profile_id WHEN 'cisco_sna_flow_search_csv_v2' THEN 'cisco_flow_search_orientation_v1' ELSE 'declared_source_destination_v1' END
 );
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_orientation_state_is_valid_v2() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_orientation_state_is_valid_v2() TO cartulary_runtime, cartulary_recovery;

ALTER TABLE public.network_flow_tables
    ADD CONSTRAINT network_flow_tables_creator_identity_key UNIQUE (incident_id, network_flow_table_id, created_by_user_id);

CREATE INDEX network_flow_rows_table_creator_scope_idx ON public.network_flow_rows (incident_id, network_flow_table_id, created_by_user_id);

ALTER TABLE public.network_flow_rows
    DROP CONSTRAINT network_flow_rows_incident_id_fkey,
    DROP CONSTRAINT network_flow_rows_created_by_user_id_fkey,
    DROP CONSTRAINT network_flow_rows_table_id_fkey,
    DROP CONSTRAINT network_flow_rows_direction_evidence_ck,
    DROP CONSTRAINT network_flow_rows_id_ck,
    DROP CONSTRAINT network_flow_rows_sha_ck,
    ADD CONSTRAINT network_flow_rows_direction_evidence_ck CHECK (public.network_flow_direction_evidence_is_valid_v2(direction_evidence,direction_kind,from_endpoint,direction_reason,endpoint_a_ip = endpoint_b_ip AND endpoint_a_port IS NOT DISTINCT FROM endpoint_b_port) IS TRUE),
    ADD CONSTRAINT network_flow_rows_id_ck CHECK (
        char_length(network_flow_row_id) = 68 AND network_flow_row_id ~ '^nfr_[a-f0-9]+$'::text),
    ADD CONSTRAINT network_flow_rows_sha_ck CHECK (
        char_length(source_row_digest_sha256) = 64 AND source_row_digest_sha256 ~ '^[a-f0-9]+$'::text
        AND char_length(normalized_row_digest_sha256) = 64 AND normalized_row_digest_sha256 ~ '^[a-f0-9]+$'::text
        AND char_length(mapping_fingerprint) = 64 AND mapping_fingerprint ~ '^[a-f0-9]+$'::text);

ALTER TABLE public.network_flow_rows ADD CONSTRAINT network_flow_rows_table_creator_fkey FOREIGN KEY (incident_id, network_flow_table_id, created_by_user_id) REFERENCES public.network_flow_tables(incident_id, network_flow_table_id, created_by_user_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_rows
    ALTER COLUMN direction_evidence SET COMPRESSION lz4,
    ALTER COLUMN observation_source_ref SET COMPRESSION lz4;

DROP FUNCTION public.network_flow_orientation_state_is_valid_v1();
DROP FUNCTION public.network_flow_direction_evidence_is_valid_v1(jsonb,text,text,text,boolean);
DROP INDEX public.network_flow_rows_network_flow_table_id_scope_idx;
DROP INDEX public.network_flow_rows_created_by_user_id_fk_idx;

-- +goose Down
-- Restore the previous definitions exactly; no data is transformed.
ALTER TABLE public.network_flow_rows
    ALTER COLUMN direction_evidence SET COMPRESSION DEFAULT,
    ALTER COLUMN observation_source_ref SET COMPRESSION DEFAULT;

CREATE INDEX network_flow_rows_created_by_user_id_fk_idx ON public.network_flow_rows USING btree (created_by_user_id);
CREATE INDEX network_flow_rows_network_flow_table_id_scope_idx ON public.network_flow_rows (incident_id, network_flow_table_id);

-- +goose StatementBegin
CREATE FUNCTION public.network_flow_direction_evidence_is_valid_v1(e jsonb, kind text, origin text, reason text, identical boolean)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE SET search_path = pg_catalog, public AS $$
DECLARE
    role_a text; role_b text; first_slot text; algorithm text;
    item jsonb; input_key text; expected text[]; ordinals text[] := '{}'; expected_origin text;
BEGIN
    IF jsonb_typeof(e) IS DISTINCT FROM 'object' OR kind IS NULL OR origin IS NULL
       OR reason IS NULL OR identical IS NULL THEN RETURN false; END IF;
    IF e - ARRAY['derivation_id','first_endpoint','first_role','second_role','inputs'] <> '{}'::jsonb
       OR NOT (e ?& ARRAY['derivation_id','first_endpoint','first_role','second_role','inputs']) THEN RETURN false; END IF;
    algorithm := e->>'derivation_id'; first_slot := e->>'first_endpoint'; role_a := e->>'first_role'; role_b := e->>'second_role';
    IF jsonb_typeof(e->'derivation_id') <> 'string' OR jsonb_typeof(e->'first_endpoint') <> 'string'
       OR jsonb_typeof(e->'first_role') <> 'string' OR jsonb_typeof(e->'second_role') <> 'string'
       OR first_slot NOT IN ('a','b') OR role_a NOT IN ('','client','server') OR role_b NOT IN ('','client','server')
       OR jsonb_typeof(e->'inputs') <> 'array' THEN RETURN false; END IF;
    IF identical AND first_slot <> 'a' THEN RETURN false; END IF;
    expected := ARRAY['network_flow.input_first_ip','network_flow.input_first_port','network_flow.input_second_ip','network_flow.input_second_port'];
    IF algorithm = 'cisco_flow_search_orientation_v1' THEN
        expected := ARRAY['network_flow.input_first_ip','network_flow.input_first_port','network_flow.input_first_orientation','network_flow.input_second_ip','network_flow.input_second_port','network_flow.input_second_orientation'];
        IF role_a = '' OR role_b = '' THEN
            IF kind <> 'unknown' OR origin <> '' OR reason <> (CASE WHEN role_a = '' AND role_b = '' THEN 'missing_orientation' ELSE 'incomplete_orientation' END) THEN RETURN false; END IF;
        ELSE
            IF role_a = role_b OR kind <> 'client_to_server' OR reason <> '' THEN RETURN false; END IF;
            expected_origin := CASE WHEN identical THEN 'a' WHEN role_a = 'client' THEN first_slot WHEN first_slot = 'a' THEN 'b' ELSE 'a' END;
            IF origin <> expected_origin THEN RETURN false; END IF;
        END IF;
    ELSIF algorithm = 'declared_source_destination_v1' THEN
        IF role_a <> '' OR role_b <> '' OR kind <> 'source_to_destination' OR reason <> '' OR origin <> first_slot THEN RETURN false; END IF;
    ELSE RETURN false;
    END IF;
    IF jsonb_array_length(e->'inputs') <> cardinality(expected) THEN RETURN false; END IF;
    -- The derivation fixes this array at four or six inputs. Avoid per-row
    -- aggregate/iterator queries in the bulk import constraint path.
    FOR input_index IN 1..cardinality(expected) LOOP
        item := e->'inputs'->(input_index-1);
        IF jsonb_typeof(item) IS DISTINCT FROM 'object' THEN RETURN false; END IF;
        IF item - ARRAY['input_key','source_column_ordinal','raw_header_sha256','raw_value_sha256'] <> '{}'::jsonb
           OR NOT (item ?& ARRAY['input_key','source_column_ordinal','raw_header_sha256','raw_value_sha256']) THEN RETURN false; END IF;
        input_key := item->>'input_key';
        IF jsonb_typeof(item->'input_key') <> 'string' OR input_key <> expected[input_index]
           OR jsonb_typeof(item->'source_column_ordinal') <> 'number' OR (item->>'source_column_ordinal') !~ '^[1-9][0-9]*$'
           OR NOT pg_input_is_valid(item->>'source_column_ordinal', 'bigint')
           OR (item->>'source_column_ordinal') = ANY(ordinals)
           OR jsonb_typeof(item->'raw_header_sha256') <> 'string' OR (item->>'raw_header_sha256') !~ '^[0-9a-f]{64}$'
           OR jsonb_typeof(item->'raw_value_sha256') <> 'string' OR (item->>'raw_value_sha256') !~ '^[0-9a-f]{64}$' THEN RETURN false; END IF;
        ordinals := array_append(ordinals,item->>'source_column_ordinal');
    END LOOP;
    RETURN true;
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_direction_evidence_is_valid_v1(jsonb,text,text,text,boolean) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_direction_evidence_is_valid_v1(jsonb,text,text,text,boolean) TO cartulary_runtime, cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.network_flow_orientation_state_is_valid_v1()
RETURNS boolean LANGUAGE sql STABLE SET search_path = pg_catalog, public AS $$
 SELECT NOT EXISTS (SELECT 1 FROM public.network_flow_tables WHERE source_profile_id NOT IN ('cisco_sna_netflow_csv_v2','cisco_sna_flow_search_csv_v2'))
 AND NOT EXISTS (
   SELECT 1 FROM public.network_flow_rows r JOIN public.network_flow_tables t USING (network_flow_table_id)
   WHERE r.incident_id <> t.incident_id
      OR public.network_flow_endpoints_are_valid_v1(r.endpoint_a_ip,r.endpoint_a_port,r.endpoint_b_ip,r.endpoint_b_port) IS NOT TRUE
      OR public.network_flow_direction_evidence_is_valid_v1(r.direction_evidence,r.direction_kind,r.from_endpoint,r.direction_reason,r.endpoint_a_ip=r.endpoint_b_ip AND r.endpoint_a_port IS NOT DISTINCT FROM r.endpoint_b_port) IS NOT TRUE
      OR r.direction_evidence->>'derivation_id' <> CASE t.source_profile_id WHEN 'cisco_sna_flow_search_csv_v2' THEN 'cisco_flow_search_orientation_v1' ELSE 'declared_source_destination_v1' END
 );
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_orientation_state_is_valid_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_orientation_state_is_valid_v1() TO cartulary_runtime, cartulary_recovery;

ALTER TABLE public.network_flow_rows
    DROP CONSTRAINT network_flow_rows_table_creator_fkey,
    DROP CONSTRAINT network_flow_rows_direction_evidence_ck,
    DROP CONSTRAINT network_flow_rows_id_ck,
    DROP CONSTRAINT network_flow_rows_sha_ck,
    ADD CONSTRAINT network_flow_rows_direction_evidence_ck CHECK (public.network_flow_direction_evidence_is_valid_v1(direction_evidence,direction_kind,from_endpoint,direction_reason,endpoint_a_ip = endpoint_b_ip AND endpoint_a_port IS NOT DISTINCT FROM endpoint_b_port) IS TRUE),
    ADD CONSTRAINT network_flow_rows_id_ck CHECK (network_flow_row_id ~ '^nfr_[a-f0-9]{64}$'::text),
    ADD CONSTRAINT network_flow_rows_sha_ck CHECK (source_row_digest_sha256 ~ '^[a-f0-9]{64}$'::text AND normalized_row_digest_sha256 ~ '^[a-f0-9]{64}$'::text AND mapping_fingerprint ~ '^[a-f0-9]{64}$'::text);

ALTER TABLE public.network_flow_rows ADD CONSTRAINT network_flow_rows_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.network_flow_rows ADD CONSTRAINT network_flow_rows_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_rows ADD CONSTRAINT network_flow_rows_table_id_fkey FOREIGN KEY (incident_id, network_flow_table_id) REFERENCES public.network_flow_tables(incident_id, network_flow_table_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_tables DROP CONSTRAINT network_flow_tables_creator_identity_key;

DROP INDEX public.network_flow_rows_table_creator_scope_idx;

DROP FUNCTION public.network_flow_orientation_state_is_valid_v2();
DROP FUNCTION public.network_flow_direction_evidence_is_valid_v2(jsonb,text,text,text,boolean);

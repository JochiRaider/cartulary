-- +goose Up
-- Repair NF-031/032 closed-object order and scalar encoding, and NF-042
-- numeric tuple ordering. Run offline after draining Network Flow work.
-- Transaction-local helpers implement this historical conversion only; they do not
-- become a runtime compatibility codec. Goose owns the entire transaction.

-- +goose StatementBegin
CREATE FUNCTION public.nf34_json(value jsonb, member_order text[] DEFAULT NULL, legacy boolean DEFAULT false)
RETURNS text LANGUAGE plpgsql IMMUTABLE CALLED ON NULL INPUT SET search_path = pg_catalog, public AS $$
DECLARE result text; name text; item jsonb; first_member boolean := true;
BEGIN
    IF value IS NULL THEN RAISE EXCEPTION 'missing Network Flow canonical value'; END IF;
    CASE jsonb_typeof(value)
    WHEN 'object' THEN
        result := '{';
        IF member_order IS NULL OR legacy THEN
            SELECT array_agg(key ORDER BY key COLLATE "C") INTO member_order FROM jsonb_object_keys(value) key;
        ELSIF cardinality(member_order) <> (SELECT count(*) FROM jsonb_object_keys(value)) THEN
            RAISE EXCEPTION 'incomplete Network Flow closed canonical object';
        END IF;
        FOREACH name IN ARRAY coalesce(member_order, ARRAY[]::text[]) LOOP
            IF NOT value ? name THEN RAISE EXCEPTION 'missing Network Flow canonical member'; END IF;
            IF NOT first_member THEN result := result || ','; END IF;
            result := result || public.nf34_json(to_jsonb(name), NULL, legacy) || ':' || public.nf34_json(value->name, NULL, legacy);
            first_member := false;
        END LOOP;
        RETURN result || '}';
    WHEN 'array' THEN
        result := '[';
        FOR item IN SELECT element FROM jsonb_array_elements(value) element LOOP
            IF NOT first_member THEN result := result || ','; END IF;
            result := result || public.nf34_json(item, NULL, legacy);
            first_member := false;
        END LOOP;
        RETURN result || ']';
    WHEN 'number' THEN
        IF value::text !~ '^-?(0|[1-9][0-9]*)$' THEN RAISE EXCEPTION 'noninteger Network Flow canonical number'; END IF;
        RETURN value::text;
    WHEN 'string' THEN
        result := value::text;
        IF legacy THEN
            result := replace(replace(replace(replace(replace(result, '<', '\u003c'), '>', '\u003e'), '&', '\u0026'), chr(8232), '\u2028'), chr(8233), '\u2029');
        END IF;
        RETURN result;
    ELSE RETURN value::text;
    END CASE;
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_json(jsonb,text[],boolean) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_part(value text) RETURNS bytea LANGUAGE sql IMMUTABLE STRICT SET search_path = pg_catalog, public
AS $$ SELECT convert_to(value, 'UTF8') || decode('00', 'hex') $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_part(text) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_mapping(mapping jsonb, source_sha text, legacy boolean DEFAULT false)
RETURNS text LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog, public AS $$
DECLARE data bytea := public.nf34_part('cartulary.network_flow_mapping_fingerprint.v1');
    name text; item jsonb; keys text[]; profile jsonb;
BEGIN
    IF mapping IS NULL OR mapping->>'target_kind' <> 'network_flow_table'
       OR mapping->>'parser_profile_id' <> 'rfc4180_headered_csv_v1'
       OR jsonb_typeof(mapping->'source_columns') IS DISTINCT FROM 'array'
       OR jsonb_typeof(mapping->'field_mappings') IS DISTINCT FROM 'array'
       OR source_sha !~ '^[a-f0-9]{64}$' THEN
        RAISE EXCEPTION 'incomplete Network Flow approved mapping context';
    END IF;
    FOREACH name IN ARRAY ARRAY['target_kind','target_table_schema_id','source_profile_id','parser_profile_id'] LOOP
        IF coalesce(mapping->>name, '') = '' THEN RAISE EXCEPTION 'missing Network Flow mapping context'; END IF;
        data := data || public.nf34_part(mapping->>name);
    END LOOP;
    data := data || public.nf34_part(source_sha) || public.nf34_part(mapping->>'unknown_column_policy');
    profile := mapping->'timestamp_profile';
    CASE profile->>'mode'
    WHEN 'rfc3339' THEN keys := ARRAY['schema_id','mode','precision','timezone','timezone_ruleset_id','ambiguous_local_time_policy','local_time_gap_policy'];
    WHEN 'epoch_seconds', 'epoch_milliseconds' THEN keys := ARRAY['schema_id','mode','precision'];
    WHEN 'netflow_sys_uptime_milliseconds' THEN keys := ARRAY['schema_id','mode','precision','netflow_export_time_column_ordinal','netflow_export_time_mode','netflow_exporter_uptime_at_export_column_ordinal'];
    ELSE RAISE EXCEPTION 'unknown Network Flow timestamp profile';
    END CASE;
    data := data || public.nf34_part(public.nf34_json(profile, keys, legacy));
    FOR item IN SELECT value FROM jsonb_array_elements(mapping->'source_columns') ORDER BY (value->>'source_column_ordinal')::integer LOOP
        FOREACH name IN ARRAY ARRAY['source_column_ordinal','raw_header_text','normalized_header_for_suggestion','raw_header_sha256'] LOOP
            IF item->>name IS NULL THEN RAISE EXCEPTION 'missing Network Flow source descriptor'; END IF;
            data := data || public.nf34_part(item->>name);
        END LOOP;
    END LOOP;
    FOR item IN SELECT value FROM jsonb_array_elements(mapping->'field_mappings')
        ORDER BY coalesce(value->>'field_key','') COLLATE "C", value->>'mapping_kind' COLLATE "C",
            CASE WHEN legacy THEN coalesce(value->>'source_column_ordinal','0') ELSE '' END COLLATE "C",
            CASE WHEN legacy THEN 0 ELSE coalesce((value->>'source_column_ordinal')::integer,0) END,
            coalesce(value->>'derivation_id','') COLLATE "C", coalesce(value->>'ignore_reason','') COLLATE "C"
    LOOP
        CASE item->>'mapping_kind'
        WHEN 'source_column' THEN keys := ARRAY['mapping_kind','field_key','source_column_ordinal','transform_id','empty_value_policy','combinability'];
        WHEN 'ignored_source_column' THEN keys := ARRAY['mapping_kind','source_column_ordinal','ignore_reason'];
        WHEN 'system_derivation' THEN keys := ARRAY['mapping_kind','field_key','derivation_id','combinability'];
        ELSE RAISE EXCEPTION 'unknown Network Flow field mapping';
        END CASE;
        data := data || public.nf34_part(public.nf34_json(item, keys, legacy));
    END LOOP;
    IF data IS NULL THEN RAISE EXCEPTION 'incomplete Network Flow fingerprint'; END IF;
    RETURN encode(sha256(data),'hex');
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_mapping(jsonb,text,boolean) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_timestamp(value timestamptz) RETURNS jsonb LANGUAGE plpgsql IMMUTABLE STRICT SET search_path = pg_catalog, public AS $$
DECLARE local_value timestamp := value AT TIME ZONE 'UTC'; fraction text;
BEGIN
    IF extract(year FROM local_value) NOT BETWEEN 1 AND 9999 THEN RAISE EXCEPTION 'invalid Network Flow timestamp'; END IF;
    fraction := rtrim(to_char(local_value, 'US'), '0');
    RETURN to_jsonb(to_char(local_value, 'YYYY-MM-DD"T"HH24:MI:SS') || CASE WHEN fraction = '' THEN '' ELSE '.' || fraction END || 'Z');
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_timestamp(timestamptz) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_unmapped(value jsonb) RETURNS text LANGUAGE plpgsql IMMUTABLE STRICT SET search_path = pg_catalog, public AS $$
DECLARE result text := '{'; item record; first_member boolean := true;
BEGIN
    FOR item IN SELECT key, val FROM jsonb_each(value) AS entries(key,val) ORDER BY key::numeric LOOP
        IF item.key !~ '^[1-9][0-9]*$' OR item.val->>'source_column_ordinal' IS DISTINCT FROM item.key THEN
            RAISE EXCEPTION 'invalid Network Flow retained ordinal';
        END IF;
        IF NOT first_member THEN result := result || ','; END IF;
        result := result || public.nf34_json(to_jsonb(item.key)) || ':' || public.nf34_json(item.val,
            ARRAY['source_column_ordinal','raw_header_text','raw_header_sha256','decoded_value','decoded_value_sha256']);
        first_member := false;
    END LOOP;
    RETURN result || '}';
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_unmapped(jsonb) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_row(value public.network_flow_rows, fingerprint text, legacy boolean DEFAULT false)
RETURNS text LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog, public AS $$
DECLARE data bytea := public.nf34_part('cartulary.network_flow.normalized_row_digest.v1') || public.nf34_part(fingerprint);
    name text; item jsonb; fields jsonb := to_jsonb(value);
BEGIN
    fields := fields || jsonb_build_object('flow_start_utc', public.nf34_timestamp(value.flow_start_utc),
        'flow_end_utc', public.nf34_timestamp(value.flow_end_utc), 'observation_source_ref', '{}'::jsonb);
    FOREACH name IN ARRAY ARRAY['flow_start_utc','flow_end_utc','src_ip','dst_ip','src_port','dst_port','ip_protocol','bytes_count','packets_count','exporter_id','input_interface','output_interface','tcp_flags','application_label','observation_source_ref'] LOOP
        data := data || public.nf34_part('network_flow.' || name);
        item := fields->name;
        IF item IS NULL OR item = 'null'::jsonb THEN
            data := data || public.nf34_part('n') || public.nf34_part('null');
        ELSE
            data := data || public.nf34_part('p') || public.nf34_part(public.nf34_json(item,NULL,legacy));
        END IF;
    END LOOP;
    data := data || public.nf34_part('unmapped_raw') || public.nf34_part(CASE WHEN legacy THEN public.nf34_json(value.unmapped_raw,NULL,true) ELSE public.nf34_unmapped(value.unmapped_raw) END);
    RETURN encode(sha256(data),'hex');
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_row(public.network_flow_rows,text,boolean) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_bound(value jsonb) RETURNS void
LANGUAGE plpgsql IMMUTABLE STRICT SET search_path = pg_catalog, public AS $$
DECLARE scalar text := value #>> '{}';
BEGIN
    IF value = 'null'::jsonb THEN RETURN; END IF;
    IF jsonb_typeof(value) <> 'string'
       OR scalar !~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,6})?Z$'
       OR substring(scalar,1,4)::integer NOT BETWEEN 1 AND 9999
       OR public.nf34_timestamp(scalar::timestamptz) IS DISTINCT FROM value THEN
        RAISE EXCEPTION 'invalid Network Flow retained UTC bound';
    END IF;
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_bound(jsonb) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_filters(value jsonb) RETURNS jsonb
LANGUAGE plpgsql IMMUTABLE STRICT SET search_path = pg_catalog, public AS $$
DECLARE current_filter jsonb;
BEGIN
    FOR current_filter IN SELECT element FROM jsonb_array_elements(value) element
        WHERE element->>'field_key' IN ('network_flow.flow_start_utc','network_flow.flow_end_utc')
            AND element->>'op' = 'range' LOOP
        PERFORM public.nf34_bound(current_filter->'value'->'gte');
        PERFORM public.nf34_bound(current_filter->'value'->'lt');
    END LOOP;
    IF jsonb_typeof(value) IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION 'invalid Network Flow retained filters';
    END IF;
    IF EXISTS (SELECT 1 FROM jsonb_array_elements(value) item
        GROUP BY public.nf34_json(item) COLLATE "C" HAVING count(*) > 1) THEN
        RAISE EXCEPTION 'duplicate Network Flow retained canonical filter';
    END IF;
    -- Only permute complete objects. Nested arrays and scalar values are facts.
    RETURN (SELECT coalesce(jsonb_agg(item ORDER BY public.nf34_json(item) COLLATE "C"),'[]'::jsonb)
        FROM jsonb_array_elements(value) item);
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_filters(jsonb) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_semantic(value jsonb, legacy boolean DEFAULT false)
RETURNS text LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog, public AS $$
DECLARE aggregation_keys text[] := ARRAY['mode','include_example_row_refs'];
BEGIN
    IF legacy THEN RETURN public.nf34_json(value,NULL,true); END IF;
    IF value->>'schema_id' IS DISTINCT FROM 'cartulary.network_flow.graph_semantic_query.v2'
       OR jsonb_typeof(value->'selected_table_ids') IS DISTINCT FROM 'array'
       OR jsonb_typeof(value->'filters') IS DISTINCT FROM 'array'
       OR (SELECT count(*) FROM jsonb_object_keys(value)) <> 5 THEN
        RAISE EXCEPTION 'incomplete Network Flow graph semantic context';
    END IF;
    PERFORM public.nf34_bound(value->'time_range'->'start_utc');
    PERFORM public.nf34_bound(value->'time_range'->'end_utc');
    IF value->'aggregation'->>'mode' = 'time_bucket_v1' THEN
        aggregation_keys := aggregation_keys || ARRAY['bucket_width_seconds'];
    ELSIF value->'aggregation'->>'mode' IS DISTINCT FROM 'default_flow_edge_v1' THEN
        RAISE EXCEPTION 'invalid Network Flow graph aggregation';
    END IF;
    RETURN '{"schema_id":' || public.nf34_json(value->'schema_id') ||
        ',"selected_table_ids":' || public.nf34_json(value->'selected_table_ids') ||
        ',"filters":' || public.nf34_json(public.nf34_filters(value->'filters')) ||
        ',"time_range":' || public.nf34_json(value->'time_range',ARRAY['start_utc','end_utc']) ||
        ',"aggregation":' || public.nf34_json(value->'aggregation',aggregation_keys) || '}';
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_semantic(jsonb,boolean) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.nf34_mutation(value jsonb, legacy boolean DEFAULT false)
RETURNS text LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog, public AS $$
BEGIN
    IF legacy OR NOT value ? 'semantic_query' THEN RETURN public.nf34_json(value,NULL,legacy); END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(value)) <> 2 THEN RAISE EXCEPTION 'invalid Network Flow graph create context'; END IF;
    RETURN '{"display_name":' || public.nf34_json(value->'display_name') ||
        ',"semantic_query":' || public.nf34_semantic(value->'semantic_query') || '}';
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.nf34_mutation(jsonb,boolean) FROM PUBLIC;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.jobs j WHERE status IN ('queued','running','cancel_requested') AND
        (extension_owner_profile_id = 'network_flow_activity' OR job_kind = 'import.apply_v1'))
       OR EXISTS (SELECT 1 FROM public.import_units WHERE approved_target_kind = 'network_flow_table' AND unit_status = 'applying') THEN
        RAISE EXCEPTION 'drain Network Flow and all import apply jobs before canonical identity migration';
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE public.nf34_tables AS
SELECT t.network_flow_table_id, t.mapping_fingerprint AS old_fingerprint,
       public.nf34_mapping(u.approved_mapping_json->'owner_mapping',t.source_content_sha256) AS fingerprint,
       public.nf34_mapping(u.approved_mapping_json->'owner_mapping',t.source_content_sha256,true) AS legacy_fingerprint,
       ARRAY(SELECT field->>'source_column_ordinal' FROM jsonb_array_elements(u.approved_mapping_json->'owner_mapping'->'field_mappings') field
           WHERE field->>'mapping_kind'='ignored_source_column') AS ignored_ordinals
FROM public.network_flow_tables t
LEFT JOIN public.import_units u ON u.import_unit_id = t.source_import_unit_id AND u.import_session_id = t.source_import_session_id;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.nf34_tables WHERE old_fingerprint NOT IN (fingerprint,legacy_fingerprint)) THEN
        RAISE EXCEPTION 'Network Flow retained mapping does not establish committed identity';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_tables t JOIN public.import_units u ON u.import_unit_id=t.source_import_unit_id
        WHERE u.import_session_id<>t.source_import_session_id OR
            u.approved_mapping_json->'owner_mapping'->>'source_profile_id' IS DISTINCT FROM t.source_profile_id OR
            u.approved_mapping_json->'owner_mapping'->>'parser_profile_id' IS DISTINCT FROM t.parser_profile_id) THEN
        RAISE EXCEPTION 'inconsistent Network Flow source provenance';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_rows r
        JOIN public.network_flow_tables t USING (network_flow_table_id)
        JOIN public.import_units u ON u.import_unit_id=t.source_import_unit_id
        WHERE CASE u.approved_mapping_json->'owner_mapping'->'timestamp_profile'->>'precision'
            WHEN 'seconds' THEN date_trunc('second',r.flow_start_utc) <> r.flow_start_utc OR date_trunc('second',r.flow_end_utc) <> r.flow_end_utc
            WHEN 'milliseconds' THEN date_trunc('milliseconds',r.flow_start_utc) <> r.flow_start_utc OR date_trunc('milliseconds',r.flow_end_utc) <> r.flow_end_utc
            WHEN 'microseconds' THEN false
            ELSE true END) THEN
        RAISE EXCEPTION 'Network Flow retained timestamps violate approved precision; authoritative source repair required';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_rows r
        JOIN public.network_flow_tables t USING (network_flow_table_id)
        JOIN public.import_units u ON u.import_unit_id=t.source_import_unit_id,
        LATERAL jsonb_each(r.unmapped_raw) raw
        WHERE u.approved_mapping_json->'owner_mapping'->>'unknown_column_policy' <> 'preserve_unmapped_raw'
            OR raw.key !~ '^[1-9][0-9]*$' OR jsonb_typeof(raw.value) <> 'object'
            OR (SELECT count(*) FROM jsonb_object_keys(raw.value)) <> 5
            OR jsonb_typeof(raw.value->'source_column_ordinal') IS DISTINCT FROM 'number'
            OR raw.value->>'source_column_ordinal' IS DISTINCT FROM raw.key
            OR jsonb_typeof(raw.value->'raw_header_text') IS DISTINCT FROM 'string'
            OR jsonb_typeof(raw.value->'raw_header_sha256') IS DISTINCT FROM 'string'
            OR jsonb_typeof(raw.value->'decoded_value_sha256') IS DISTINCT FROM 'string'
            OR jsonb_typeof(raw.value->'decoded_value') IS DISTINCT FROM 'string'
            OR raw.value->>'decoded_value_sha256' IS DISTINCT FROM encode(sha256(convert_to(raw.value->>'decoded_value','UTF8')),'hex')
            OR NOT EXISTS (SELECT 1 FROM jsonb_array_elements(u.approved_mapping_json->'owner_mapping'->'source_columns') column_value
                WHERE column_value->>'source_column_ordinal'=raw.key
                    AND column_value->>'raw_header_text'=raw.value->>'raw_header_text'
                    AND column_value->>'raw_header_sha256'=raw.value->>'raw_header_sha256'
                    AND raw.value->>'raw_header_sha256'=encode(sha256(convert_to(raw.value->>'raw_header_text','UTF8')),'hex'))
            OR EXISTS (SELECT 1 FROM jsonb_array_elements(u.approved_mapping_json->'owner_mapping'->'field_mappings') field
                WHERE field->>'mapping_kind'='source_column' AND field->>'source_column_ordinal'=raw.key)) THEN
        RAISE EXCEPTION 'inconsistent Network Flow retained raw provenance';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_rows r JOIN public.nf34_tables t USING (network_flow_table_id)
        WHERE r.mapping_fingerprint <> t.old_fingerprint OR r.normalized_row_digest_sha256 NOT IN
            (public.nf34_row(r,t.old_fingerprint),public.nf34_row(r,t.old_fingerprint,true))) THEN
        RAISE EXCEPTION 'Network Flow retained values do not establish committed row identity';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_rows r JOIN public.network_flow_tables t USING (network_flow_table_id)
        WHERE r.incident_id <> t.incident_id OR r.source_row_number < 2 OR r.network_flow_row_id <> 'nfr_' || encode(sha256(public.nf34_part('cartulary.network_flow_row_id.v1') ||
            public.nf34_part(r.incident_id::text) || public.nf34_part(r.network_flow_table_id) || public.nf34_part(r.source_row_number::text) ||
            public.nf34_part(r.source_row_digest_sha256) || public.nf34_part(r.normalized_row_digest_sha256)),'hex') OR
            r.observation_source_ref IS DISTINCT FROM jsonb_build_object('import_session_id',t.source_import_session_id::text,
                'import_unit_id',t.source_import_unit_id::text,'source_content_sha256',t.source_content_sha256,
                'source_profile_id',t.source_profile_id,'parser_profile_id',t.parser_profile_id,'mapping_fingerprint',t.mapping_fingerprint,
                'source_row_number',r.source_row_number,'source_row_digest_sha256',r.source_row_digest_sha256)) THEN
        RAISE EXCEPTION 'inconsistent Network Flow committed identity or observation provenance';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_indicator_bindings b,
        LATERAL jsonb_array_elements(b.source_row_refs) ref
        WHERE jsonb_typeof(ref) IS DISTINCT FROM 'object'
            OR (SELECT count(*) FROM jsonb_object_keys(ref)) <> 4
            OR jsonb_typeof(ref->'network_flow_table_id') IS DISTINCT FROM 'string'
            OR jsonb_typeof(ref->'network_flow_row_id') IS DISTINCT FROM 'string'
            OR jsonb_typeof(ref->'source_row_number') IS DISTINCT FROM 'number'
            OR jsonb_typeof(ref->'mapping_fingerprint') IS DISTINCT FROM 'string'
            OR NOT EXISTS (SELECT 1 FROM public.network_flow_rows r WHERE
                r.network_flow_row_id=ref->>'network_flow_row_id' AND r.network_flow_table_id=ref->>'network_flow_table_id'
                AND r.incident_id=b.incident_id AND r.source_row_number::text=ref->>'source_row_number'
                AND r.mapping_fingerprint=ref->>'mapping_fingerprint')) OR
       EXISTS (SELECT 1 FROM public.network_flow_indicator_bindings b WHERE b.source_row_ref_row_ids IS DISTINCT FROM
        ARRAY(SELECT ref->>'network_flow_row_id' FROM jsonb_array_elements(b.source_row_refs) refs(ref) ORDER BY ref->>'network_flow_row_id' COLLATE "C")
        OR jsonb_array_length(b.source_row_refs) <> (SELECT count(DISTINCT ref->>'network_flow_row_id') FROM jsonb_array_elements(b.source_row_refs) ref)) THEN
        RAISE EXCEPTION 'inconsistent Network Flow binding source lineage';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_rejected_row_diagnostics d JOIN public.network_flow_tables t USING (network_flow_table_id)
        WHERE d.diagnostic_id <> 'nfd_' || encode(sha256(
            public.nf34_part('cartulary.network_flow.diagnostic_id.v1') || public.nf34_part(t.source_content_sha256) ||
            public.nf34_part(t.mapping_fingerprint) || public.nf34_part(d.source_row_number::text) ||
            public.nf34_part(CASE WHEN d.source_column_ordinal IS NULL THEN 'n' ELSE 'p:' || d.source_column_ordinal::text END) ||
            public.nf34_part(CASE WHEN d.field_key IS NULL THEN 'n' ELSE 'p:' || d.field_key END) ||
            public.nf34_part(d.error_code) || public.nf34_part(d.reason_code)), 'hex')) THEN
        RAISE EXCEPTION 'inconsistent Network Flow diagnostic identity';
    END IF;
    IF EXISTS (SELECT 1 FROM public.network_flow_graph_views WHERE semantic_query_sha256 NOT IN (
        encode(sha256(convert_to(public.nf34_semantic(semantic_query_json,true),'UTF8')),'hex'),
        encode(sha256(convert_to(public.nf34_semantic(semantic_query_json),'UTF8')),'hex'))) THEN
        RAISE EXCEPTION 'inconsistent Network Flow saved graph semantic identity';
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE public.nf34_rows AS
WITH retained AS (
    SELECT r.network_flow_row_id AS old_id, t.fingerprint,
        jsonb_populate_record(r,jsonb_build_object('unmapped_raw',r.unmapped_raw-t.ignored_ordinals)) AS value
    FROM public.network_flow_rows r JOIN public.nf34_tables t USING (network_flow_table_id)
)
SELECT old_id, (value).network_flow_table_id, fingerprint, (value).unmapped_raw,
       public.nf34_row(value,fingerprint) AS normalized_digest,
       'nfr_' || encode(sha256(public.nf34_part('cartulary.network_flow_row_id.v1') || public.nf34_part((value).incident_id::text) ||
           public.nf34_part((value).network_flow_table_id) || public.nf34_part((value).source_row_number::text) ||
           public.nf34_part((value).source_row_digest_sha256) || public.nf34_part(public.nf34_row(value,fingerprint))), 'hex') AS row_id
FROM retained;
CREATE UNIQUE INDEX nf34_rows_old_id ON public.nf34_rows(old_id);
CREATE UNIQUE INDEX nf34_rows_row_id ON public.nf34_rows(row_id);

ALTER TABLE public.network_flow_rows DISABLE TRIGGER network_flow_rows_immutable_update;
UPDATE public.network_flow_rows r SET network_flow_row_id = n.row_id, normalized_row_digest_sha256 = n.normalized_digest,
    mapping_fingerprint = n.fingerprint, unmapped_raw = n.unmapped_raw,
    observation_source_ref = jsonb_set(r.observation_source_ref,'{mapping_fingerprint}',to_jsonb(n.fingerprint),false)
FROM public.nf34_rows n WHERE r.network_flow_row_id = n.old_id;
ALTER TABLE public.network_flow_rows ENABLE TRIGGER network_flow_rows_immutable_update;

ALTER TABLE public.network_flow_rejected_row_diagnostics DISABLE TRIGGER network_flow_rejected_row_diagnostics_immutable_update;
UPDATE public.network_flow_rejected_row_diagnostics d SET diagnostic_id = 'nfd_' || encode(sha256(
    public.nf34_part('cartulary.network_flow.diagnostic_id.v1') || public.nf34_part(t.source_content_sha256) ||
    public.nf34_part(n.fingerprint) || public.nf34_part(d.source_row_number::text) ||
    public.nf34_part(CASE WHEN d.source_column_ordinal IS NULL THEN 'n' ELSE 'p:' || d.source_column_ordinal::text END) ||
    public.nf34_part(CASE WHEN d.field_key IS NULL THEN 'n' ELSE 'p:' || d.field_key END) ||
    public.nf34_part(d.error_code) || public.nf34_part(d.reason_code)), 'hex')
FROM public.network_flow_tables t JOIN public.nf34_tables n USING (network_flow_table_id) WHERE d.network_flow_table_id = t.network_flow_table_id;
ALTER TABLE public.network_flow_rejected_row_diagnostics ENABLE TRIGGER network_flow_rejected_row_diagnostics_immutable_update;

UPDATE public.network_flow_indicator_bindings b SET
    source_row_refs = (SELECT jsonb_agg(jsonb_set(jsonb_set(ref,'{network_flow_row_id}',to_jsonb(n.row_id),false),'{mapping_fingerprint}',to_jsonb(n.fingerprint),false) ORDER BY ordinal)
        FROM jsonb_array_elements(b.source_row_refs) WITH ORDINALITY refs(ref,ordinal)
        JOIN public.nf34_rows n ON n.old_id = ref->>'network_flow_row_id'),
    source_row_ref_row_ids = ARRAY(SELECT n.row_id FROM unnest(b.source_row_ref_row_ids) refs(ref)
        JOIN public.nf34_rows n ON n.old_id = ref ORDER BY n.row_id COLLATE "C");

UPDATE public.network_flow_tables t SET mapping_fingerprint = n.fingerprint, table_version = t.table_version + 1
FROM public.nf34_tables n WHERE t.network_flow_table_id = n.network_flow_table_id AND
    (n.old_fingerprint <> n.fingerprint OR EXISTS (SELECT 1 FROM public.nf34_rows r WHERE r.network_flow_table_id = t.network_flow_table_id AND r.old_id <> r.row_id));

-- Imports owns the original admission identity. Terminal unit approvals,
-- plans, outcomes and audit responses remain historical facts. Current unapplied
-- approvals require renewed admission; historical attempt plans stay unchanged.
CREATE TABLE public.nf34_unapplied AS
SELECT import_unit_id, import_session_id FROM public.import_units
WHERE approved_target_kind = 'network_flow_table' AND unit_status IN ('selected','mapped','ready');
UPDATE public.import_units u SET unit_status = 'selected', mapping_fingerprint = NULL, approved_mapping_json = NULL,
    approved_target_kind = NULL, approved_extension_profile_id = NULL, approved_target_view_schema_id = NULL
FROM public.nf34_unapplied n WHERE u.import_unit_id = n.import_unit_id;
UPDATE public.import_sessions s SET session_status = 'discovered'
WHERE session_status IN ('mapped','ready_to_apply') AND EXISTS (SELECT 1 FROM public.nf34_unapplied u WHERE u.import_session_id = s.import_session_id);

UPDATE public.network_flow_graph_views SET
    semantic_query_json = jsonb_set(semantic_query_json,'{filters}',public.nf34_filters(semantic_query_json->'filters'),false),
    semantic_query_sha256 = encode(sha256(convert_to(public.nf34_semantic(semantic_query_json), 'UTF8')), 'hex');

-- Reconstruct only receipts whose owner admission already reconstructs the
-- request. Other historical receipt hashes remain unchanged and fail closed.
-- +goose StatementBegin
DO $$
DECLARE receipt record; graph jsonb; framing bytea; comparison jsonb; semantic_hash text; old_hash bytea; new_hash bytea; graph_job uuid;
BEGIN
    FOR receipt IN SELECT * FROM public.route_idempotency WHERE route_key IN
        ('nf.graph_views.create','nf.graph_views.patch','nf.graph_views.refresh') LOOP
        graph := receipt.response_json->'graph_view';
        IF graph IS NULL THEN RAISE EXCEPTION 'missing Network Flow graph receipt context'; END IF;
        IF graph->'semantic_query' IS NULL OR graph->>'semantic_query_sha256' IS NULL OR graph->>'semantic_query_sha256' NOT IN (
            encode(sha256(convert_to(public.nf34_semantic(graph->'semantic_query',true),'UTF8')),'hex'),
            encode(sha256(convert_to(public.nf34_semantic(graph->'semantic_query'),'UTF8')),'hex')) THEN
            RAISE EXCEPTION 'inconsistent Network Flow graph receipt semantic identity';
        END IF;
        semantic_hash := encode(sha256(convert_to(public.nf34_semantic(graph->'semantic_query'),'UTF8')),'hex');
        IF receipt.route_key = 'nf.graph_views.create' THEN
            comparison := jsonb_build_object('display_name',graph->'display_name','semantic_query',graph->'semantic_query');
        ELSE
            comparison := jsonb_build_object('base_graph_view_version',(graph->>'graph_view_version')::bigint - 1);
            IF receipt.route_key = 'nf.graph_views.patch' THEN comparison := comparison || jsonb_build_object('display_name',graph->'display_name'); END IF;
        END IF;
        framing := public.nf34_part('cartulary.network_flow.mutation_request_digest.v1') || public.nf34_part(receipt.route_key) ||
            public.nf34_part(substring(receipt.scope_key FROM position(':' IN receipt.scope_key)+1));
        old_hash := sha256(framing || public.nf34_part(public.nf34_mutation(comparison,true)));
        new_hash := sha256(framing || public.nf34_part(public.nf34_mutation(comparison)));
        IF receipt.request_hash NOT IN (old_hash,new_hash) AND receipt.route_key = 'nf.graph_views.patch' THEN
            comparison := jsonb_set(comparison,'{base_graph_view_version}',graph->'graph_view_version');
            old_hash := sha256(framing || public.nf34_part(public.nf34_mutation(comparison,true)));
            new_hash := sha256(framing || public.nf34_part(public.nf34_mutation(comparison)));
        END IF;
        IF receipt.request_hash NOT IN (old_hash,new_hash) THEN RAISE EXCEPTION 'inconsistent Network Flow graph receipt hash'; END IF;
        IF receipt.route_key <> 'nf.graph_views.patch' THEN
            graph_job := (graph->>'latest_job_id')::uuid;
            IF graph_job IS NULL OR NOT EXISTS (SELECT 1 FROM public.jobs j WHERE j.job_id = graph_job
                AND j.extension_owner_profile_id = 'network_flow_activity'
                AND j.job_kind = 'network_flow_activity.graph_view_materialize_v1'
                AND j.handler_name = 'network_flow_activity.graph_view_worker_v1'
                AND j.progress_unit_id = 'network_flow_activity.graph_view_materialize.projection_result.v1'
                AND j.scope_kind = 'incident' AND j.auth_policy = 'incident_membership'
                AND j.incident_id::text = split_part(receipt.scope_key,':',1) AND j.submitted_by_user_id = receipt.actor_user_id
                AND (j.expired_at IS NOT NULL OR (j.extension_normalized_request_sha256 IN (encode(receipt.request_hash,'hex'),encode(new_hash,'hex'))
                    AND j.extension_idempotency_route_key = receipt.route_key AND j.extension_idempotency_scope_key = receipt.scope_key
                    AND j.extension_idempotency_identity = jsonb_build_object('schema_id','cartulary.route_scoped_idempotency_identity.v1',
                        'actor_user_id',receipt.actor_user_id::text,'route_identity',receipt.route_key || ':' || receipt.scope_key,
                        'client_txn_id',receipt.client_txn_id,'scope_kind','incident','scope_id',split_part(receipt.scope_key,':',1))))) THEN
                RAISE EXCEPTION 'inconsistent Network Flow retained graph job';
            END IF;
            IF EXISTS (SELECT 1 FROM public.jobs WHERE job_id = graph_job AND status = 'succeeded') THEN
                IF NOT EXISTS (SELECT 1 FROM public.extension_job_commit_proofs p WHERE p.job_id = graph_job
                    AND p.owner_profile_id = 'network_flow_activity' AND p.operation_kind = 'network_flow_activity.graph_view_materialize'
                    AND p.normalized_request_sha256 IN (encode(receipt.request_hash,'hex'),encode(new_hash,'hex'))
                    AND p.idempotency_identity = jsonb_build_object('schema_id','cartulary.route_scoped_idempotency_identity.v1',
                        'actor_user_id',receipt.actor_user_id::text,'route_identity',receipt.route_key || ':' || receipt.scope_key,
                        'client_txn_id',receipt.client_txn_id,'scope_kind','incident','scope_id',split_part(receipt.scope_key,':',1))) THEN
                    RAISE EXCEPTION 'inconsistent Network Flow graph commit proof';
                END IF;
            ELSIF EXISTS (SELECT 1 FROM public.extension_job_commit_proofs WHERE job_id=graph_job) THEN
                RAISE EXCEPTION 'unexpected Network Flow graph commit proof';
            END IF;
            UPDATE public.jobs SET extension_normalized_request_sha256 = encode(new_hash,'hex') WHERE job_id = graph_job AND expired_at IS NULL;
            UPDATE public.extension_job_commit_proofs SET normalized_request_sha256 = encode(new_hash,'hex')
                WHERE job_id = graph_job AND normalized_request_sha256 = encode(receipt.request_hash,'hex');
        END IF;
        UPDATE public.route_idempotency SET request_hash = new_hash,
            response_json = jsonb_set(jsonb_set(response_json,'{graph_view,semantic_query_sha256}',to_jsonb(semantic_hash),false),
                '{graph_view,semantic_query,filters}',public.nf34_filters(graph->'semantic_query'->'filters'),false) WHERE route_key = receipt.route_key AND scope_key = receipt.scope_key
            AND actor_user_id = receipt.actor_user_id AND client_txn_id = receipt.client_txn_id;
    END LOOP;
END;
$$;
-- +goose StatementEnd

DROP TABLE public.nf34_unapplied;
DROP TABLE public.nf34_rows;
DROP TABLE public.nf34_tables;

DROP FUNCTION public.nf34_row(public.network_flow_rows,text,boolean);
DROP FUNCTION public.nf34_bound(jsonb);
DROP FUNCTION public.nf34_timestamp(timestamptz);
DROP FUNCTION public.nf34_unmapped(jsonb);
DROP FUNCTION public.nf34_mapping(jsonb,text,boolean);
DROP FUNCTION public.nf34_part(text);
DROP FUNCTION public.nf34_mutation(jsonb,boolean);
DROP FUNCTION public.nf34_semantic(jsonb,boolean);
DROP FUNCTION public.nf34_filters(jsonb);
DROP FUNCTION public.nf34_json(jsonb,text[],boolean);

-- +goose Down
-- Canonical identity has no safe inverse: two serializers can establish the
-- same scalar bytes, and refreshed projections may already use new row IDs.
-- Restore the matching pre-upgrade database backup and binary together.
-- A disposable database with no retained affected state needs no data inverse.
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.network_flow_tables)
       OR EXISTS (SELECT 1 FROM public.network_flow_graph_views)
       OR EXISTS (SELECT 1 FROM public.network_flow_indicator_bindings)
       OR EXISTS (SELECT 1 FROM public.import_units)
       OR EXISTS (SELECT 1 FROM public.jobs WHERE extension_owner_profile_id='network_flow_activity')
       OR EXISTS (SELECT 1 FROM public.route_idempotency WHERE route_key LIKE 'nf.%') THEN
        RAISE EXCEPTION 'Network Flow canonical identity requires snapshot restore for downgrade';
    END IF;
END; $$;
-- +goose StatementEnd

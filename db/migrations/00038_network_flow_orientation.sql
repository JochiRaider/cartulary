-- +goose Up
-- No legacy state is translated. This lock and guard precede every DDL change.
LOCK TABLE public.network_flow_tables IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.network_flow_rows IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.extension_state_metadata IN SHARE ROW EXCLUSIVE MODE;
LOCK TABLE public.reporting_job_payloads IN SHARE ROW EXCLUSIVE MODE;
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.network_flow_tables)
       OR EXISTS (SELECT 1 FROM public.network_flow_rows)
       OR EXISTS (SELECT 1 FROM public.extension_state_metadata WHERE profile_id = 'network_flow_activity') THEN
        RAISE EXCEPTION 'orientation cutover requires fresh Network Flow state; original sources must be re-imported';
    END IF;
    IF EXISTS (SELECT 1 FROM public.reporting_job_payloads
               WHERE job_kind <> 'snapshot_create'
                 AND request_json->'graph_sources'->>'schema_id' IS DISTINCT FROM 'cartulary.reporting_graph_input.v2') THEN
        RAISE EXCEPTION 'orientation cutover requires fresh Reporting inputs; retain historical reports in the previous environment';
    END IF;
END
$$;
-- +goose StatementEnd

-- Reporting owns this coordinated retained-input boundary. No old frozen
-- input is converted, discarded, or relabeled during admission.
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_graph_input_v1_ck;
ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_graph_input_v2_ck CHECK (
    (job_kind = 'snapshot_create' AND admitted_graph_sources_sha256 IS NULL
     OR job_kind <> 'snapshot_create'
        AND admitted_graph_sources_sha256 ~ '^[0-9a-f]{64}$'
        AND request_json->'graph_sources'->>'schema_id' = 'cartulary.reporting_graph_input.v2'
        AND jsonb_typeof(request_json->'graph_sources'->'results') = 'array') IS TRUE
);

ALTER TABLE public.network_flow_rows RENAME COLUMN src_ip TO endpoint_a_ip;
ALTER TABLE public.network_flow_rows RENAME COLUMN dst_ip TO endpoint_b_ip;
ALTER TABLE public.network_flow_rows RENAME COLUMN src_port TO endpoint_a_port;
ALTER TABLE public.network_flow_rows RENAME COLUMN dst_port TO endpoint_b_port;

-- PostgreSQL validates addresses. The owner supplies the canonical IPv4 grammar
-- and formats IPv6 from bytes; PostgreSQL display conventions are not the contract.
-- +goose StatementBegin
CREATE FUNCTION public.network_flow_ip_literal_is_canonical_v1(value text)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE SET search_path = pg_catalog, public AS $$
DECLARE
    address inet; packed bytea; canonical text := ''; groups integer[] := '{}';
    i integer; run_start integer := 0; run_length integer := 0;
    best_start integer := 0; best_length integer := 0;
BEGIN
    IF value IS NULL OR octet_length(value) < 1 OR octet_length(value) > 45
       OR value COLLATE "C" ~ '[^0-9a-f:.]' THEN RETURN false; END IF;
    IF strpos(value, ':') = 0 THEN
        -- Four decimal octets without leading zeroes. PostgreSQL checks range;
        -- no byte-to-text round trip is needed for this unambiguous grammar.
        RETURN value COLLATE "C" ~ '^(0|[1-9][0-9]{0,2})([.](0|[1-9][0-9]{0,2})){3}$'
            AND pg_input_is_valid(value, 'inet');
    END IF;
    -- Avoid an exception subtransaction for every endpoint in bulk imports.
    -- PostgreSQL's soft-error input check preserves false-on-malformed input.
    IF NOT pg_input_is_valid(value, 'inet') THEN RETURN false; END IF;
    address := value::inet;
    packed := inet_send(address);
    IF substring(packed FROM 5 FOR 10) = decode(repeat('00',10),'hex')
          AND get_byte(packed,14) = 255 AND get_byte(packed,15) = 255 THEN
        canonical := '::ffff:';
        FOR i IN 16..19 LOOP
            IF i > 16 THEN canonical := canonical || '.'; END IF;
            canonical := canonical || get_byte(packed,i)::text;
        END LOOP;
    ELSE
        FOR i IN 1..8 LOOP
            groups := array_append(groups, get_byte(packed,2*i+2)*256 + get_byte(packed,2*i+3));
            IF groups[i] = 0 THEN
                IF run_length = 0 THEN run_start := i; END IF;
                run_length := run_length + 1;
                IF run_length >= 2 AND run_length > best_length THEN
                    best_start := run_start; best_length := run_length;
                END IF;
            ELSE run_length := 0;
            END IF;
        END LOOP;
        i := 1;
        WHILE i <= 8 LOOP
            IF i = best_start THEN
                canonical := canonical || '::'; i := i + best_length;
            ELSE
                IF canonical <> '' AND right(canonical,1) <> ':' THEN canonical := canonical || ':'; END IF;
                canonical := canonical || to_hex(groups[i]); i := i + 1;
            END IF;
        END LOOP;
    END IF;
    RETURN value COLLATE "C" = canonical COLLATE "C";
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.network_flow_endpoints_are_valid_v1(a_ip text, a_port integer, b_ip text, b_port integer)
RETURNS boolean LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE SET search_path = pg_catalog, public AS $$
BEGIN
    IF NOT public.network_flow_ip_literal_is_canonical_v1(a_ip)
       OR NOT public.network_flow_ip_literal_is_canonical_v1(b_ip)
       OR (a_port IS NOT NULL AND (a_port < 0 OR a_port > 65535))
       OR (b_port IS NOT NULL AND (b_port < 0 OR b_port > 65535)) THEN RETURN false; END IF;
    RETURN a_ip::inet < b_ip::inet OR
        (a_ip::inet = b_ip::inet AND
         (a_port IS NULL OR (b_port IS NOT NULL AND a_port <= b_port)));
END
$$;
-- +goose StatementEnd

-- The owner predicate is shared by CHECK constraints and Recovery validation.
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

ALTER TABLE public.network_flow_rows
    DROP CONSTRAINT network_flow_rows_ip_ck,
    DROP CONSTRAINT network_flow_rows_ports_ck,
    ADD COLUMN direction_kind text NOT NULL,
    ADD COLUMN from_endpoint text NOT NULL,
    ADD COLUMN direction_reason text NOT NULL,
    ADD COLUMN direction_evidence jsonb NOT NULL,
    ADD CONSTRAINT network_flow_rows_direction_ck CHECK (
        (direction_kind IN ('client_to_server','source_to_destination')
         AND from_endpoint IN ('a','b') AND direction_reason = ''
         AND (endpoint_a_ip <> endpoint_b_ip OR endpoint_a_port IS DISTINCT FROM endpoint_b_port OR from_endpoint = 'a'))
        OR (direction_kind = 'unknown' AND from_endpoint = ''
            AND direction_reason IN ('missing_orientation','incomplete_orientation'))
    ),
    ADD CONSTRAINT network_flow_rows_endpoints_ck CHECK (
        public.network_flow_endpoints_are_valid_v1(endpoint_a_ip,endpoint_a_port,endpoint_b_ip,endpoint_b_port) IS TRUE),
    ADD CONSTRAINT network_flow_rows_direction_evidence_ck CHECK (public.network_flow_direction_evidence_is_valid_v1(direction_evidence,direction_kind,from_endpoint,direction_reason,endpoint_a_ip = endpoint_b_ip AND endpoint_a_port IS NOT DISTINCT FROM endpoint_b_port) IS TRUE);

CREATE INDEX network_flow_rows_direction_idx
    ON public.network_flow_rows (incident_id, network_flow_table_id, direction_kind, direction_reason);

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

REVOKE ALL ON FUNCTION public.network_flow_direction_evidence_is_valid_v1(jsonb,text,text,text,boolean) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_direction_evidence_is_valid_v1(jsonb,text,text,text,boolean) TO cartulary_runtime, cartulary_recovery;
REVOKE ALL ON FUNCTION public.network_flow_orientation_state_is_valid_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_orientation_state_is_valid_v1() TO cartulary_runtime, cartulary_recovery;
REVOKE ALL ON FUNCTION public.network_flow_ip_literal_is_canonical_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_ip_literal_is_canonical_v1(text) TO cartulary_runtime, cartulary_recovery;
REVOKE ALL ON FUNCTION public.network_flow_endpoints_are_valid_v1(text,integer,text,integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.network_flow_endpoints_are_valid_v1(text,integer,text,integer) TO cartulary_runtime, cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'orientation cutover cannot be downgraded; restore the previous matching environment';
END
$$;
-- +goose StatementEnd

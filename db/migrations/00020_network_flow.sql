-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: network flow.
-- Administrator provisioning and final admission are separate lifecycle steps.

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

REVOKE ALL ON FUNCTION public.network_flow_ip_literal_is_canonical_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.network_flow_ip_literal_is_canonical_v1(text) TO cartulary_runtime, cartulary_recovery;

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

REVOKE ALL ON FUNCTION public.network_flow_endpoints_are_valid_v1(text,integer,text,integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.network_flow_endpoints_are_valid_v1(text,integer,text,integer) TO cartulary_runtime, cartulary_recovery;

-- The owner predicate is shared by CHECK constraints and Recovery validation.
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

CREATE TABLE public.network_flow_graph_views (
    graph_view_id text NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    normalized_display_name text NOT NULL,
    declaration_state text NOT NULL,
    semantic_query_json jsonb NOT NULL,
    semantic_query_sha256 text NOT NULL,
    desired_source_snapshot_id text NOT NULL,
    selected_projection_result_id text,
    selected_source_snapshot_id text,
    selected_projection_schema_id text,
    selected_projection_version text,
    selected_normalized_configuration_sha256 text,
    selected_normalized_source_sha256 text,
    selected_canonical_output_sha256 text,
    graph_view_version bigint NOT NULL,
    materialization_generation bigint NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    retired_at timestamp with time zone,
    latest_job_id uuid,
    last_failure_code text,
    last_failed_at timestamp with time zone,
    refresh_intent_json jsonb NOT NULL,
    selected_source_binding_json jsonb,
    CONSTRAINT network_flow_graph_views_display_name_ck CHECK (char_length(display_name) >= 1 AND char_length(display_name) <= 64 AND display_name !~ '[[:cntrl:]]'::text AND char_length(normalized_display_name) >= 1 AND char_length(normalized_display_name) <= 64 AND normalized_display_name !~ '[[:cntrl:]]'::text),
    CONSTRAINT network_flow_graph_views_failure_ck CHECK (last_failure_code IS NULL AND last_failed_at IS NULL OR last_failure_code ~ '^network_flow_[a-z0-9_]+$'::text AND last_failed_at IS NOT NULL),
    CONSTRAINT network_flow_graph_views_id_ck CHECK (graph_view_id ~ '^nfgv_[a-f0-9]{32}$'::text),
    CONSTRAINT network_flow_graph_views_pkey PRIMARY KEY (graph_view_id),
    CONSTRAINT network_flow_graph_views_refresh_intent_ck CHECK (
        jsonb_typeof(refresh_intent_json) = 'object'),
    CONSTRAINT network_flow_graph_views_selected_binding_ck CHECK (selected_projection_result_id IS NULL AND selected_source_snapshot_id IS NULL AND selected_projection_schema_id IS NULL AND selected_projection_version IS NULL AND selected_normalized_configuration_sha256 IS NULL AND selected_normalized_source_sha256 IS NULL AND selected_canonical_output_sha256 IS NULL OR selected_projection_result_id ~ '^gpres_[a-f0-9]{64}$'::text AND selected_source_snapshot_id IS NOT NULL AND selected_projection_schema_id = 'graph_projection.v2'::text AND selected_projection_version IS NOT NULL AND selected_normalized_configuration_sha256 ~ '^[a-f0-9]{64}$'::text AND selected_normalized_source_sha256 ~ '^[a-f0-9]{64}$'::text AND selected_canonical_output_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT network_flow_graph_views_selected_source_ck CHECK (
        (selected_projection_result_id IS NULL AND selected_source_binding_json IS NULL)
        OR (selected_projection_result_id IS NOT NULL AND selected_source_binding_json IS NOT NULL
            AND jsonb_typeof(selected_source_binding_json) = 'object')),
    CONSTRAINT network_flow_graph_views_semantic_query_ck CHECK (jsonb_typeof(semantic_query_json) = 'object'::text AND semantic_query_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT network_flow_graph_views_source_snapshot_ck CHECK (char_length(desired_source_snapshot_id) >= 1 AND char_length(desired_source_snapshot_id) <= 255 AND desired_source_snapshot_id !~ '[[:cntrl:]]'::text),
    CONSTRAINT network_flow_graph_views_state_ck CHECK (declaration_state = ANY (ARRAY['active'::text, 'retired'::text])),
    CONSTRAINT network_flow_graph_views_state_time_ck CHECK (declaration_state = 'active'::text AND retired_at IS NULL OR declaration_state = 'retired'::text AND retired_at IS NOT NULL),
    CONSTRAINT network_flow_graph_views_times_ck CHECK (updated_at >= created_at AND (retired_at IS NULL OR retired_at >= created_at) AND (last_failed_at IS NULL OR last_failed_at >= created_at)),
    CONSTRAINT network_flow_graph_views_versions_ck CHECK (graph_view_version >= 1 AND materialization_generation >= 1)
);

REVOKE ALL ON TABLE public.network_flow_graph_views FROM PUBLIC;

REVOKE ALL ON TYPE public.network_flow_graph_views FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.network_flow_graph_views TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.network_flow_graph_views TO cartulary_recovery;

GRANT USAGE ON TYPE public.network_flow_graph_views TO cartulary_runtime;

GRANT USAGE ON TYPE public.network_flow_graph_views TO cartulary_recovery;

CREATE TABLE public.network_flow_indicator_bindings (
    network_flow_indicator_binding_id text NOT NULL,
    incident_id uuid NOT NULL,
    target_indicator_record_id uuid NOT NULL,
    target_indicator_type text NOT NULL,
    target_indicator_value_kind text NOT NULL,
    target_indicator_normalized_value text NOT NULL,
    selector_kind text NOT NULL,
    candidate_value text NOT NULL,
    source_row_refs jsonb NOT NULL,
    source_row_ref_row_ids text[] NOT NULL,
    source_row_refs_truncated boolean DEFAULT false NOT NULL,
    source_row_refs_total_count bigint NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT network_flow_indicator_bindings_candidate_ck CHECK (char_length(candidate_value) >= 1 AND char_length(candidate_value) <= 45 AND candidate_value = target_indicator_normalized_value),
    CONSTRAINT network_flow_indicator_bindings_id_ck CHECK (network_flow_indicator_binding_id ~ '^nfb_[a-f0-9]{32}$'::text),
    CONSTRAINT network_flow_indicator_bindings_pkey PRIMARY KEY (network_flow_indicator_binding_id),
    CONSTRAINT network_flow_indicator_bindings_selector_kind_ck CHECK (selector_kind = ANY (ARRAY['row_field_value'::text, 'row_refs'::text, 'graph_vertex'::text, 'graph_edge'::text])),
    CONSTRAINT network_flow_indicator_bindings_source_count_ck CHECK (source_row_refs_total_count >= cardinality(source_row_ref_row_ids) AND source_row_refs_total_count > 0),
    CONSTRAINT network_flow_indicator_bindings_source_refs_ck CHECK (jsonb_typeof(source_row_refs) = 'array'::text AND jsonb_array_length(source_row_refs) >= 1 AND cardinality(source_row_ref_row_ids) >= 1),
    CONSTRAINT network_flow_indicator_bindings_target_ck CHECK ((target_indicator_type = ANY (ARRAY['ipv4_addr'::text, 'ipv6_addr'::text])) AND target_indicator_value_kind = 'atomic'::text)
);

REVOKE ALL ON TABLE public.network_flow_indicator_bindings FROM PUBLIC;

REVOKE ALL ON TYPE public.network_flow_indicator_bindings FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.network_flow_indicator_bindings TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.network_flow_indicator_bindings TO cartulary_recovery;

CREATE TABLE public.network_flow_rejected_row_diagnostics (
    diagnostic_id text NOT NULL,
    network_flow_table_id text NOT NULL,
    incident_id uuid NOT NULL,
    source_row_number bigint NOT NULL,
    source_column_ordinal bigint,
    raw_header_sha256 text,
    field_key text,
    error_code text NOT NULL,
    reason_code text NOT NULL,
    safe_sample text,
    raw_value_sha256 text,
    message_key text NOT NULL,
    message_args jsonb DEFAULT '{}'::jsonb NOT NULL,
    message text NOT NULL,
    limit_name text,
    limit_value bigint,
    actual_value bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT network_flow_rejected_row_diagnostics_error_code_ck CHECK (error_code ~ '^network_flow_[a-z0-9_]+$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_id_ck CHECK (diagnostic_id ~ '^nfd_[a-f0-9]{64}$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_limit_ck CHECK ((limit_value IS NULL OR limit_value >= 0) AND (actual_value IS NULL OR actual_value >= 0)),
    CONSTRAINT network_flow_rejected_row_diagnostics_message_args_ck CHECK (jsonb_typeof(message_args) = 'object'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_message_ck CHECK (char_length(message) <= 1024),
    CONSTRAINT network_flow_rejected_row_diagnostics_message_key_ck CHECK (message_key ~ '^network_flow\.diagnostic\.[a-z0-9_]+\.[a-z0-9_]+$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_pkey PRIMARY KEY (network_flow_table_id, diagnostic_id),
    CONSTRAINT network_flow_rejected_row_diagnostics_reason_code_ck CHECK (reason_code ~ '^[a-z][a-z0-9_]*$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_sample_ck CHECK (safe_sample IS NULL OR char_length(safe_sample) <= 64),
    CONSTRAINT network_flow_rejected_row_diagnostics_sha_ck CHECK ((raw_header_sha256 IS NULL OR raw_header_sha256 ~ '^[a-f0-9]{64}$'::text) AND (raw_value_sha256 IS NULL OR raw_value_sha256 ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT network_flow_rejected_row_diagnostics_source_column_ck CHECK (source_column_ordinal IS NULL OR source_column_ordinal > 0),
    CONSTRAINT network_flow_rejected_row_diagnostics_source_row_ck CHECK (source_row_number > 0)
);

REVOKE ALL ON TABLE public.network_flow_rejected_row_diagnostics FROM PUBLIC;

REVOKE ALL ON TYPE public.network_flow_rejected_row_diagnostics FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.network_flow_rejected_row_diagnostics TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.network_flow_rejected_row_diagnostics TO cartulary_recovery;

CREATE TABLE public.network_flow_rows (
    network_flow_row_id text NOT NULL,
    network_flow_table_id text NOT NULL,
    incident_id uuid NOT NULL,
    source_row_number bigint NOT NULL,
    source_row_digest_sha256 text NOT NULL,
    normalized_row_digest_sha256 text NOT NULL,
    mapping_fingerprint text NOT NULL,
    flow_start_utc timestamp with time zone NOT NULL,
    flow_end_utc timestamp with time zone NOT NULL,
    endpoint_a_ip text NOT NULL,
    endpoint_b_ip text NOT NULL,
    endpoint_a_port integer,
    endpoint_b_port integer,
    ip_protocol integer NOT NULL,
    bytes_count text NOT NULL,
    packets_count text NOT NULL,
    exporter_id text,
    input_interface text,
    output_interface text,
    tcp_flags integer,
    application_label text,
    unmapped_raw jsonb DEFAULT '{}'::jsonb NOT NULL,
    observation_source_ref jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    direction_kind text NOT NULL,
    from_endpoint text NOT NULL,
    direction_reason text NOT NULL,
    direction_evidence jsonb NOT NULL,
    CONSTRAINT network_flow_rows_direction_ck CHECK (
        (direction_kind IN ('client_to_server','source_to_destination')
         AND from_endpoint IN ('a','b') AND direction_reason = ''
         AND (endpoint_a_ip <> endpoint_b_ip OR endpoint_a_port IS DISTINCT FROM endpoint_b_port OR from_endpoint = 'a'))
        OR (direction_kind = 'unknown' AND from_endpoint = ''
            AND direction_reason IN ('missing_orientation','incomplete_orientation'))
    ),
    CONSTRAINT network_flow_rows_direction_evidence_ck CHECK (public.network_flow_direction_evidence_is_valid_v2(direction_evidence,direction_kind,from_endpoint,direction_reason,endpoint_a_ip = endpoint_b_ip AND endpoint_a_port IS NOT DISTINCT FROM endpoint_b_port) IS TRUE),
    CONSTRAINT network_flow_rows_endpoints_ck CHECK (
        public.network_flow_endpoints_are_valid_v1(endpoint_a_ip,endpoint_a_port,endpoint_b_ip,endpoint_b_port) IS TRUE),
    CONSTRAINT network_flow_rows_id_ck CHECK (
        char_length(network_flow_row_id) = 68 AND network_flow_row_id ~ '^nfr_[a-f0-9]+$'::text),
    CONSTRAINT network_flow_rows_nullable_text_ck CHECK ((exporter_id IS NULL OR char_length(exporter_id) <= 256 AND exporter_id !~ '[[:cntrl:]]'::text) AND (input_interface IS NULL OR char_length(input_interface) <= 256 AND input_interface !~ '[[:cntrl:]]'::text) AND (output_interface IS NULL OR char_length(output_interface) <= 256 AND output_interface !~ '[[:cntrl:]]'::text) AND (application_label IS NULL OR char_length(application_label) <= 256 AND application_label !~ '[[:cntrl:]]'::text)),
    CONSTRAINT network_flow_rows_observation_source_ref_object_ck CHECK (jsonb_typeof(observation_source_ref) = 'object'::text),
    CONSTRAINT network_flow_rows_pkey PRIMARY KEY (network_flow_row_id),
    CONSTRAINT network_flow_rows_positive_row_ck CHECK (source_row_number > 0),
    CONSTRAINT network_flow_rows_protocol_flags_ck CHECK (ip_protocol >= 0 AND ip_protocol <= 255 AND (tcp_flags IS NULL OR tcp_flags >= 0 AND tcp_flags <= 255)),
    CONSTRAINT network_flow_rows_sha_ck CHECK (
        char_length(source_row_digest_sha256) = 64 AND source_row_digest_sha256 ~ '^[a-f0-9]+$'::text
        AND char_length(normalized_row_digest_sha256) = 64 AND normalized_row_digest_sha256 ~ '^[a-f0-9]+$'::text
        AND char_length(mapping_fingerprint) = 64 AND mapping_fingerprint ~ '^[a-f0-9]+$'::text),
    CONSTRAINT network_flow_rows_table_source_row_key UNIQUE (network_flow_table_id, source_row_number),
    CONSTRAINT network_flow_rows_time_order_ck CHECK (flow_end_utc >= flow_start_utc),
    CONSTRAINT network_flow_rows_uint64_text_ck CHECK (bytes_count ~ '^(0|[1-9][0-9]{0,19})$'::text AND packets_count ~ '^(0|[1-9][0-9]{0,19})$'::text AND bytes_count::numeric <= '18446744073709551615'::numeric AND packets_count::numeric <= '18446744073709551615'::numeric),
    CONSTRAINT network_flow_rows_unmapped_raw_object_ck CHECK (jsonb_typeof(unmapped_raw) = 'object'::text)
);

REVOKE ALL ON TABLE public.network_flow_rows FROM PUBLIC;

REVOKE ALL ON TYPE public.network_flow_rows FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.network_flow_rows TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.network_flow_rows TO cartulary_recovery;

ALTER TABLE public.network_flow_rows ALTER COLUMN direction_evidence SET COMPRESSION lz4;

ALTER TABLE public.network_flow_rows ALTER COLUMN observation_source_ref SET COMPRESSION lz4;

CREATE TABLE public.network_flow_tables (
    network_flow_table_id text NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    table_version bigint DEFAULT 1 NOT NULL,
    table_status text NOT NULL,
    source_content_sha256 text NOT NULL,
    source_filename_display text NOT NULL,
    source_filename_digest text NOT NULL,
    source_filename_digest_key_id text NOT NULL,
    mapping_fingerprint text NOT NULL,
    source_profile_id text NOT NULL,
    parser_profile_id text NOT NULL,
    row_count_accepted bigint NOT NULL,
    row_count_rejected bigint NOT NULL,
    diagnostics_truncated boolean DEFAULT false NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    committed_source_id uuid NOT NULL,
    source_interpretation_sha256 text NOT NULL,
    source_raw_cell_scalar_limit bigint NOT NULL,
    CONSTRAINT network_flow_tables_committed_source_key UNIQUE (committed_source_id),
    CONSTRAINT network_flow_tables_counts_ck CHECK (row_count_accepted > 0 AND row_count_rejected >= 0),
    CONSTRAINT network_flow_tables_display_name_ck CHECK (char_length(display_name) >= 1 AND char_length(display_name) <= 64 AND display_name !~ '[[:cntrl:]]'::text),
    CONSTRAINT network_flow_tables_id_ck CHECK (network_flow_table_id ~ '^nft_[a-f0-9]{32}$'::text),
    CONSTRAINT network_flow_tables_interpretation_ck CHECK (source_interpretation_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT network_flow_tables_pkey PRIMARY KEY (network_flow_table_id),
    CONSTRAINT network_flow_tables_profile_ids_ck CHECK (char_length(source_profile_id) >= 1 AND char_length(source_profile_id) <= 128 AND source_profile_id !~ '[[:cntrl:]]'::text AND char_length(parser_profile_id) >= 1 AND char_length(parser_profile_id) <= 128 AND parser_profile_id !~ '[[:cntrl:]]'::text),
    CONSTRAINT network_flow_tables_safe_key_id_ck CHECK (source_filename_digest_key_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'::text),
    CONSTRAINT network_flow_tables_sha_ck CHECK (source_content_sha256 ~ '^[a-f0-9]{64}$'::text AND source_filename_digest ~ '^[a-f0-9]{64}$'::text AND mapping_fingerprint ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT network_flow_tables_source_filename_display_ck CHECK (char_length(source_filename_display) >= 1 AND char_length(source_filename_display) <= 256 AND source_filename_display !~ '[[:cntrl:]]'::text),
    CONSTRAINT network_flow_tables_source_raw_cell_scalar_limit_ck CHECK (source_raw_cell_scalar_limit BETWEEN 1 AND 16384),
    CONSTRAINT network_flow_tables_status_ck CHECK (table_status = ANY (ARRAY['active'::text, 'soft_deleted'::text])),
    CONSTRAINT network_flow_tables_status_deleted_at_ck CHECK (table_status = 'active'::text AND deleted_at IS NULL OR table_status = 'soft_deleted'::text AND deleted_at IS NOT NULL),
    CONSTRAINT network_flow_tables_timestamps_ck CHECK (updated_at >= created_at AND (deleted_at IS NULL OR deleted_at >= created_at)),
    CONSTRAINT network_flow_tables_version_ck CHECK (table_version >= 1),
    CONSTRAINT network_flow_tables_incident_identity_key UNIQUE (incident_id, network_flow_table_id),
    CONSTRAINT network_flow_tables_creator_identity_key UNIQUE (incident_id, network_flow_table_id, created_by_user_id)
);

REVOKE ALL ON TABLE public.network_flow_tables FROM PUBLIC;

REVOKE ALL ON TYPE public.network_flow_tables FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.network_flow_tables TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.network_flow_tables TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.network_flow_reject_immutable_update()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    RAISE EXCEPTION 'Network Flow committed analytical rows are immutable'
        USING ERRCODE = '23514';
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.network_flow_reject_immutable_update() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.network_flow_reject_immutable_update() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.network_flow_reject_immutable_update() TO cartulary_recovery;

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

CREATE INDEX network_flow_graph_views_active_order_idx ON public.network_flow_graph_views USING btree (incident_id, normalized_display_name, graph_view_id) WHERE (declaration_state = 'active'::text);

CREATE INDEX network_flow_graph_views_created_by_user_id_fk_idx ON public.network_flow_graph_views USING btree (created_by_user_id);

CREATE INDEX network_flow_graph_views_latest_job_id_fk_idx ON public.network_flow_graph_views USING btree (latest_job_id);

CREATE INDEX network_flow_graph_views_latest_job_id_scope_idx ON public.network_flow_graph_views (incident_id, latest_job_id);

CREATE INDEX network_flow_graph_views_retained_count_idx ON public.network_flow_graph_views USING btree (incident_id, declaration_state, graph_view_id);

CREATE INDEX network_flow_graph_views_selected_result_idx ON public.network_flow_graph_views USING btree (selected_projection_result_id) WHERE (selected_projection_result_id IS NOT NULL);

CREATE INDEX network_flow_indicator_bindings_created_by_user_id_fk_idx ON public.network_flow_indicator_bindings USING btree (created_by_user_id);

CREATE UNIQUE INDEX network_flow_indicator_bindings_identity_uidx ON public.network_flow_indicator_bindings USING btree (incident_id, target_indicator_record_id, candidate_value, source_row_ref_row_ids);

CREATE INDEX network_flow_indicator_bindings_incident_created_idx ON public.network_flow_indicator_bindings USING btree (incident_id, created_at DESC, network_flow_indicator_binding_id);

CREATE INDEX network_flow_indicator_bindings_target_indicator_recor_e8bba056 ON public.network_flow_indicator_bindings USING btree (target_indicator_record_id);

CREATE INDEX network_flow_rejected_diagnostics_keyset_idx ON public.network_flow_rejected_row_diagnostics USING btree (network_flow_table_id, source_row_number, source_column_ordinal, field_key, error_code, reason_code, diagnostic_id);

CREATE INDEX network_flow_rejected_row_diagnostics_incident_id_fk_idx ON public.network_flow_rejected_row_diagnostics USING btree (incident_id);

CREATE INDEX network_flow_rejected_row_diagnostics_table_id_scope_idx ON public.network_flow_rejected_row_diagnostics (incident_id, network_flow_table_id);

CREATE INDEX network_flow_rows_direction_idx ON public.network_flow_rows USING btree (incident_id, network_flow_table_id, direction_kind, direction_reason);

CREATE INDEX network_flow_rows_incident_default_keyset_idx ON public.network_flow_rows USING btree (incident_id, flow_start_utc, flow_end_utc, source_row_number, network_flow_row_id, network_flow_table_id);

CREATE INDEX network_flow_rows_table_creator_scope_idx ON public.network_flow_rows USING btree (incident_id, network_flow_table_id, created_by_user_id);

CREATE INDEX network_flow_rows_table_default_keyset_idx ON public.network_flow_rows USING btree (network_flow_table_id, flow_start_utc, flow_end_utc, source_row_number, network_flow_row_id);

CREATE UNIQUE INDEX network_flow_tables_active_display_name_uidx ON public.network_flow_tables USING btree (incident_id, display_name) WHERE (table_status = 'active'::text);

CREATE INDEX network_flow_tables_committed_source_scope_idx ON public.network_flow_tables USING btree (incident_id, committed_source_id);

CREATE INDEX network_flow_tables_active_order_idx ON public.network_flow_tables USING btree (incident_id, created_at, network_flow_table_id) WHERE (table_status = 'active'::text);

CREATE INDEX network_flow_tables_created_by_user_id_fk_idx ON public.network_flow_tables USING btree (created_by_user_id);

CREATE INDEX network_flow_tables_retained_count_idx ON public.network_flow_tables USING btree (incident_id, table_status);


ALTER TABLE public.network_flow_graph_views ADD CONSTRAINT network_flow_graph_views_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.network_flow_graph_views ADD CONSTRAINT network_flow_graph_views_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_graph_views ADD CONSTRAINT network_flow_graph_views_latest_job_id_fkey FOREIGN KEY (incident_id, latest_job_id) REFERENCES public.jobs(incident_id, job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.network_flow_indicator_bindings ADD CONSTRAINT network_flow_indicator_bindings_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.network_flow_indicator_bindings ADD CONSTRAINT network_flow_indicator_bindings_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_indicator_bindings ADD CONSTRAINT network_flow_indicator_bindings_target_indicator_record_id_fkey FOREIGN KEY (incident_id, target_indicator_record_id) REFERENCES public.indicators(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_rejected_row_diagnostics ADD CONSTRAINT network_flow_rejected_row_diagnostics_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_rejected_row_diagnostics ADD CONSTRAINT network_flow_rejected_row_diagnostics_table_id_fkey FOREIGN KEY (incident_id, network_flow_table_id) REFERENCES public.network_flow_tables(incident_id, network_flow_table_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_rows ADD CONSTRAINT network_flow_rows_table_creator_fkey FOREIGN KEY (incident_id, network_flow_table_id, created_by_user_id) REFERENCES public.network_flow_tables(incident_id, network_flow_table_id, created_by_user_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_tables ADD CONSTRAINT network_flow_tables_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.network_flow_tables ADD CONSTRAINT network_flow_tables_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.network_flow_tables ADD CONSTRAINT network_flow_tables_committed_source_fk FOREIGN KEY (incident_id, committed_source_id) REFERENCES public.import_committed_sources(incident_id, committed_source_id) ON UPDATE NO ACTION ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;

CREATE TRIGGER network_flow_rejected_row_diagnostics_immutable_update BEFORE UPDATE ON public.network_flow_rejected_row_diagnostics FOR EACH ROW EXECUTE FUNCTION public.network_flow_reject_immutable_update();

CREATE TRIGGER network_flow_rows_immutable_update BEFORE UPDATE ON public.network_flow_rows FOR EACH ROW EXECUTE FUNCTION public.network_flow_reject_immutable_update();

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER network_flow_rows_immutable_update ON public.network_flow_rows;
DROP TRIGGER network_flow_rejected_row_diagnostics_immutable_update ON public.network_flow_rejected_row_diagnostics;
ALTER TABLE public.network_flow_tables DROP CONSTRAINT network_flow_tables_committed_source_fk;
ALTER TABLE public.network_flow_tables DROP CONSTRAINT network_flow_tables_incident_id_fkey;
ALTER TABLE public.network_flow_tables DROP CONSTRAINT network_flow_tables_created_by_user_id_fkey;
ALTER TABLE public.network_flow_rows DROP CONSTRAINT network_flow_rows_table_creator_fkey;
ALTER TABLE public.network_flow_rejected_row_diagnostics DROP CONSTRAINT network_flow_rejected_row_diagnostics_table_id_fkey;
ALTER TABLE public.network_flow_rejected_row_diagnostics DROP CONSTRAINT network_flow_rejected_row_diagnostics_incident_id_fkey;
ALTER TABLE public.network_flow_indicator_bindings DROP CONSTRAINT network_flow_indicator_bindings_target_indicator_record_id_fkey;
ALTER TABLE public.network_flow_indicator_bindings DROP CONSTRAINT network_flow_indicator_bindings_incident_id_fkey;
ALTER TABLE public.network_flow_indicator_bindings DROP CONSTRAINT network_flow_indicator_bindings_created_by_user_id_fkey;
ALTER TABLE public.network_flow_graph_views DROP CONSTRAINT network_flow_graph_views_latest_job_id_fkey;
ALTER TABLE public.network_flow_graph_views DROP CONSTRAINT network_flow_graph_views_incident_id_fkey;
ALTER TABLE public.network_flow_graph_views DROP CONSTRAINT network_flow_graph_views_created_by_user_id_fkey;
DROP FUNCTION public.network_flow_orientation_state_is_valid_v2();
DROP FUNCTION public.network_flow_reject_immutable_update();
DROP TABLE public.network_flow_tables;
DROP TABLE public.network_flow_rows;
DROP TABLE public.network_flow_rejected_row_diagnostics;
DROP TABLE public.network_flow_indicator_bindings;
DROP TABLE public.network_flow_graph_views;
DROP FUNCTION public.network_flow_direction_evidence_is_valid_v2(jsonb,text,text,text,boolean);
DROP FUNCTION public.network_flow_endpoints_are_valid_v1(text,integer,text,integer);
DROP FUNCTION public.network_flow_ip_literal_is_canonical_v1(text);

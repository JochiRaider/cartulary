-- +goose Up
--
-- Name: Network Flow Activity storage; Type: TABLES; Schema: public; Owner: -
--

CREATE TABLE public.network_flow_tables (
    network_flow_table_id text NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    table_version bigint DEFAULT 1 NOT NULL,
    table_status text NOT NULL,
    source_import_session_id uuid NOT NULL,
    source_import_unit_id uuid NOT NULL,
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
    CONSTRAINT network_flow_tables_pkey PRIMARY KEY (network_flow_table_id),
    CONSTRAINT network_flow_tables_id_ck CHECK (network_flow_table_id ~ '^nft_[a-f0-9]{32}$'::text),
    CONSTRAINT network_flow_tables_display_name_ck CHECK ((char_length(display_name) >= 1) AND (char_length(display_name) <= 64) AND (display_name !~ '[[:cntrl:]]'::text)),
    CONSTRAINT network_flow_tables_status_ck CHECK (table_status = ANY (ARRAY['active'::text, 'soft_deleted'::text])),
    CONSTRAINT network_flow_tables_status_deleted_at_ck CHECK (((table_status = 'active'::text) AND (deleted_at IS NULL)) OR ((table_status = 'soft_deleted'::text) AND (deleted_at IS NOT NULL))),
    CONSTRAINT network_flow_tables_version_ck CHECK (table_version >= 1),
    CONSTRAINT network_flow_tables_sha_ck CHECK ((source_content_sha256 ~ '^[a-f0-9]{64}$'::text) AND (source_filename_digest ~ '^[a-f0-9]{64}$'::text) AND (mapping_fingerprint ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT network_flow_tables_safe_key_id_ck CHECK (source_filename_digest_key_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'::text),
    CONSTRAINT network_flow_tables_source_filename_display_ck CHECK ((char_length(source_filename_display) >= 1) AND (char_length(source_filename_display) <= 256) AND (source_filename_display !~ '[[:cntrl:]]'::text)),
    CONSTRAINT network_flow_tables_profile_ids_ck CHECK ((char_length(source_profile_id) >= 1) AND (char_length(source_profile_id) <= 128) AND (source_profile_id !~ '[[:cntrl:]]'::text) AND (char_length(parser_profile_id) >= 1) AND (char_length(parser_profile_id) <= 128) AND (parser_profile_id !~ '[[:cntrl:]]'::text)),
    CONSTRAINT network_flow_tables_counts_ck CHECK ((row_count_accepted > 0) AND (row_count_rejected >= 0)),
    CONSTRAINT network_flow_tables_timestamps_ck CHECK ((updated_at >= created_at) AND ((deleted_at IS NULL) OR (deleted_at >= created_at)))
);

ALTER TABLE ONLY public.network_flow_tables
    ADD CONSTRAINT network_flow_tables_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_tables_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    ADD CONSTRAINT network_flow_tables_source_import_session_id_fkey FOREIGN KEY (source_import_session_id) REFERENCES public.import_sessions(import_session_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    ADD CONSTRAINT network_flow_tables_source_import_unit_id_fkey FOREIGN KEY (source_import_unit_id) REFERENCES public.import_units(import_unit_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    ADD CONSTRAINT network_flow_tables_source_import_unit_id_key UNIQUE (source_import_unit_id);

CREATE UNIQUE INDEX network_flow_tables_active_display_name_uidx
    ON public.network_flow_tables USING btree (incident_id, display_name)
    WHERE (table_status = 'active'::text);

CREATE INDEX network_flow_tables_active_order_idx
    ON public.network_flow_tables USING btree (incident_id, created_at, network_flow_table_id)
    WHERE (table_status = 'active'::text);

CREATE INDEX network_flow_tables_retained_count_idx
    ON public.network_flow_tables USING btree (incident_id, table_status);

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
    src_ip text NOT NULL,
    dst_ip text NOT NULL,
    src_port integer,
    dst_port integer,
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
    CONSTRAINT network_flow_rows_pkey PRIMARY KEY (network_flow_row_id),
    CONSTRAINT network_flow_rows_id_ck CHECK (network_flow_row_id ~ '^nfr_[a-f0-9]{64}$'::text),
    CONSTRAINT network_flow_rows_positive_row_ck CHECK (source_row_number > 0),
    CONSTRAINT network_flow_rows_sha_ck CHECK ((source_row_digest_sha256 ~ '^[a-f0-9]{64}$'::text) AND (normalized_row_digest_sha256 ~ '^[a-f0-9]{64}$'::text) AND (mapping_fingerprint ~ '^[a-f0-9]{64}$'::text)),
    CONSTRAINT network_flow_rows_time_order_ck CHECK (flow_end_utc >= flow_start_utc),
    CONSTRAINT network_flow_rows_ip_ck CHECK ((char_length(src_ip) >= 1) AND (char_length(src_ip) <= 45) AND (char_length(dst_ip) >= 1) AND (char_length(dst_ip) <= 45)),
    CONSTRAINT network_flow_rows_ports_ck CHECK (((src_port IS NULL) OR ((src_port >= 0) AND (src_port <= 65535))) AND ((dst_port IS NULL) OR ((dst_port >= 0) AND (dst_port <= 65535)))),
    CONSTRAINT network_flow_rows_protocol_flags_ck CHECK ((ip_protocol >= 0) AND (ip_protocol <= 255) AND ((tcp_flags IS NULL) OR ((tcp_flags >= 0) AND (tcp_flags <= 255)))),
    CONSTRAINT network_flow_rows_uint64_text_ck CHECK ((bytes_count ~ '^(0|[1-9][0-9]{0,19})$'::text) AND (packets_count ~ '^(0|[1-9][0-9]{0,19})$'::text) AND (bytes_count::numeric <= 18446744073709551615::numeric) AND (packets_count::numeric <= 18446744073709551615::numeric)),
    CONSTRAINT network_flow_rows_nullable_text_ck CHECK (((exporter_id IS NULL) OR ((char_length(exporter_id) <= 256) AND (exporter_id !~ '[[:cntrl:]]'::text))) AND ((input_interface IS NULL) OR ((char_length(input_interface) <= 256) AND (input_interface !~ '[[:cntrl:]]'::text))) AND ((output_interface IS NULL) OR ((char_length(output_interface) <= 256) AND (output_interface !~ '[[:cntrl:]]'::text))) AND ((application_label IS NULL) OR ((char_length(application_label) <= 256) AND (application_label !~ '[[:cntrl:]]'::text)))),
    CONSTRAINT network_flow_rows_unmapped_raw_object_ck CHECK (jsonb_typeof(unmapped_raw) = 'object'::text),
    CONSTRAINT network_flow_rows_observation_source_ref_object_ck CHECK (jsonb_typeof(observation_source_ref) = 'object'::text)
);

ALTER TABLE ONLY public.network_flow_rows
    ADD CONSTRAINT network_flow_rows_table_id_fkey FOREIGN KEY (network_flow_table_id) REFERENCES public.network_flow_tables(network_flow_table_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_rows_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_rows_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    ADD CONSTRAINT network_flow_rows_table_source_row_key UNIQUE (network_flow_table_id, source_row_number);

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
    CONSTRAINT network_flow_rejected_row_diagnostics_pkey PRIMARY KEY (diagnostic_id),
    CONSTRAINT network_flow_rejected_row_diagnostics_id_ck CHECK (diagnostic_id ~ '^nfd_[a-f0-9]{64}$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_source_row_ck CHECK (source_row_number > 0),
    CONSTRAINT network_flow_rejected_row_diagnostics_source_column_ck CHECK ((source_column_ordinal IS NULL) OR (source_column_ordinal > 0)),
    CONSTRAINT network_flow_rejected_row_diagnostics_sha_ck CHECK (((raw_header_sha256 IS NULL) OR (raw_header_sha256 ~ '^[a-f0-9]{64}$'::text)) AND ((raw_value_sha256 IS NULL) OR (raw_value_sha256 ~ '^[a-f0-9]{64}$'::text))),
    CONSTRAINT network_flow_rejected_row_diagnostics_error_code_ck CHECK (error_code ~ '^network_flow_[a-z0-9_]+$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_reason_code_ck CHECK (reason_code ~ '^[a-z][a-z0-9_]*$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_sample_ck CHECK ((safe_sample IS NULL) OR (char_length(safe_sample) <= 64)),
    CONSTRAINT network_flow_rejected_row_diagnostics_message_key_ck CHECK (message_key ~ '^network_flow\.diagnostic\.[a-z0-9_]+\.[a-z0-9_]+$'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_message_args_ck CHECK (jsonb_typeof(message_args) = 'object'::text),
    CONSTRAINT network_flow_rejected_row_diagnostics_message_ck CHECK (char_length(message) <= 1024),
    CONSTRAINT network_flow_rejected_row_diagnostics_limit_ck CHECK (((limit_value IS NULL) OR (limit_value >= 0)) AND ((actual_value IS NULL) OR (actual_value >= 0)))
);

ALTER TABLE ONLY public.network_flow_rejected_row_diagnostics
    ADD CONSTRAINT network_flow_rejected_row_diagnostics_table_id_fkey FOREIGN KEY (network_flow_table_id) REFERENCES public.network_flow_tables(network_flow_table_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_rejected_row_diagnostics_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE;

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

CREATE TRIGGER network_flow_rows_immutable_update
    BEFORE UPDATE ON public.network_flow_rows
    FOR EACH ROW EXECUTE FUNCTION public.network_flow_reject_immutable_update();

CREATE TRIGGER network_flow_rejected_row_diagnostics_immutable_update
    BEFORE UPDATE ON public.network_flow_rejected_row_diagnostics
    FOR EACH ROW EXECUTE FUNCTION public.network_flow_reject_immutable_update();

--
-- Name: Network Flow Activity indicator bindings; Type: TABLES; Schema: public; Owner: -
--

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
    CONSTRAINT network_flow_indicator_bindings_pkey PRIMARY KEY (network_flow_indicator_binding_id),
    CONSTRAINT network_flow_indicator_bindings_id_ck CHECK (network_flow_indicator_binding_id ~ '^nfb_[a-f0-9]{32}$'::text),
    CONSTRAINT network_flow_indicator_bindings_selector_kind_ck CHECK (selector_kind = ANY (ARRAY['row_field_value'::text, 'row_refs'::text, 'graph_vertex'::text, 'graph_edge'::text])),
    CONSTRAINT network_flow_indicator_bindings_target_ck CHECK ((target_indicator_type = ANY (ARRAY['ipv4_addr'::text, 'ipv6_addr'::text])) AND (target_indicator_value_kind = 'atomic'::text)),
    CONSTRAINT network_flow_indicator_bindings_candidate_ck CHECK ((char_length(candidate_value) >= 1) AND (char_length(candidate_value) <= 45) AND (candidate_value = target_indicator_normalized_value)),
    CONSTRAINT network_flow_indicator_bindings_source_refs_ck CHECK ((jsonb_typeof(source_row_refs) = 'array'::text) AND (jsonb_array_length(source_row_refs) >= 1) AND (cardinality(source_row_ref_row_ids) >= 1)),
    CONSTRAINT network_flow_indicator_bindings_source_count_ck CHECK ((source_row_refs_total_count >= cardinality(source_row_ref_row_ids)) AND (source_row_refs_total_count > 0))
);

ALTER TABLE ONLY public.network_flow_indicator_bindings
    ADD CONSTRAINT network_flow_indicator_bindings_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_indicator_bindings_target_indicator_record_id_fkey FOREIGN KEY (target_indicator_record_id) REFERENCES public.indicators(record_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_indicator_bindings_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE UNIQUE INDEX network_flow_indicator_bindings_identity_uidx
    ON public.network_flow_indicator_bindings USING btree (incident_id, target_indicator_record_id, candidate_value, source_row_ref_row_ids);

CREATE INDEX network_flow_indicator_bindings_incident_created_idx
    ON public.network_flow_indicator_bindings USING btree (incident_id, created_at DESC, network_flow_indicator_binding_id);

--
-- Name: Network Flow keyset pagination indexes; Type: INDEXES; Schema: public; Owner: -
--

CREATE INDEX network_flow_rows_table_default_keyset_idx
    ON public.network_flow_rows USING btree
    (network_flow_table_id, flow_start_utc, flow_end_utc, source_row_number, network_flow_row_id);

CREATE INDEX network_flow_rows_incident_default_keyset_idx
    ON public.network_flow_rows USING btree
    (incident_id, flow_start_utc, flow_end_utc, source_row_number, network_flow_row_id, network_flow_table_id);

CREATE INDEX network_flow_rejected_diagnostics_keyset_idx
    ON public.network_flow_rejected_row_diagnostics USING btree
    (network_flow_table_id, source_row_number, source_column_ordinal ASC NULLS LAST,
     field_key ASC NULLS LAST, error_code, reason_code, diagnostic_id);

CREATE INDEX network_flow_indicator_bindings_created_by_user_id_fk_idx ON public.network_flow_indicator_bindings (created_by_user_id);
CREATE INDEX network_flow_indicator_bindings_target_indicator_recor_e8bba056 ON public.network_flow_indicator_bindings (target_indicator_record_id);
CREATE INDEX network_flow_rejected_row_diagnostics_incident_id_fk_idx ON public.network_flow_rejected_row_diagnostics (incident_id);
CREATE INDEX network_flow_rows_created_by_user_id_fk_idx ON public.network_flow_rows (created_by_user_id);
CREATE INDEX network_flow_tables_created_by_user_id_fk_idx ON public.network_flow_tables (created_by_user_id);
CREATE INDEX network_flow_tables_source_import_session_id_fk_idx ON public.network_flow_tables (source_import_session_id);

-- +goose Down
DROP INDEX public.network_flow_rejected_diagnostics_keyset_idx;
DROP INDEX public.network_flow_rows_incident_default_keyset_idx;
DROP INDEX public.network_flow_rows_table_default_keyset_idx;

DROP INDEX public.network_flow_indicator_bindings_incident_created_idx;
DROP INDEX public.network_flow_indicator_bindings_identity_uidx;
DROP TABLE public.network_flow_indicator_bindings;

DROP TRIGGER network_flow_rejected_row_diagnostics_immutable_update ON public.network_flow_rejected_row_diagnostics;
DROP TRIGGER network_flow_rows_immutable_update ON public.network_flow_rows;
DROP FUNCTION public.network_flow_reject_immutable_update();

DROP TABLE public.network_flow_rejected_row_diagnostics;

DROP TABLE public.network_flow_rows;

DROP INDEX public.network_flow_tables_retained_count_idx;
DROP INDEX public.network_flow_tables_active_order_idx;
DROP INDEX public.network_flow_tables_active_display_name_uidx;
DROP TABLE public.network_flow_tables;

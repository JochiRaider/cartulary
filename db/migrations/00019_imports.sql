-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: imports.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.import_apply_journal (
    import_apply_journal_id uuid DEFAULT gen_random_uuid() NOT NULL,
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    mapping_fingerprint text NOT NULL,
    source_row_ref integer NOT NULL,
    target_view_schema_id text NOT NULL,
    owner_create_facade text NOT NULL,
    record_id uuid NOT NULL,
    row_version bigint NOT NULL,
    change_set_id uuid NOT NULL,
    change_set_mutation_ref text NOT NULL,
    owner_result_code text NOT NULL,
    created_or_reused text NOT NULL,
    owner_response_json jsonb NOT NULL,
    row_refresh_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_apply_journal_import_unit_id_mapping_fingerprint_sou_key UNIQUE (import_unit_id, mapping_fingerprint, source_row_ref),
    CONSTRAINT import_apply_journal_pkey PRIMARY KEY (import_apply_journal_id),
    CONSTRAINT import_apply_journal_row_version_check CHECK (row_version >= 1),
    CONSTRAINT import_apply_journal_source_row_ref_check CHECK (source_row_ref > 0)
);

REVOKE ALL ON TABLE public.import_apply_journal FROM PUBLIC;

REVOKE ALL ON TYPE public.import_apply_journal FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.import_apply_journal TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_apply_journal TO cartulary_recovery;

CREATE TABLE public.import_apply_unit_plans (
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    apply_job_id uuid NOT NULL,
    discovery_sequence integer NOT NULL,
    source_file_kind text NOT NULL,
    source_content_sha256 text NOT NULL,
    source_stream_ref text NOT NULL,
    source_rows_sha256 text NOT NULL,
    parser_profile_id text NOT NULL,
    parser_version text NOT NULL,
    locator_kind text NOT NULL,
    locator text NOT NULL,
    source_rect_a1 text NOT NULL,
    mapping_fingerprint text NOT NULL,
    approved_mapping_json jsonb NOT NULL,
    approved_mapping_sha256 text NOT NULL,
    target_kind text NOT NULL,
    target_view_schema_id text,
    extension_profile_id text,
    owner_binding_id text NOT NULL,
    target_registry_sha256 text NOT NULL,
    admitted_at timestamp with time zone NOT NULL,
    CONSTRAINT import_apply_unit_plans_digest_ck CHECK (source_content_sha256 ~ '^[0-9a-f]{64}$'::text AND source_rows_sha256 ~ '^[0-9a-f]{64}$'::text AND mapping_fingerprint ~ '^[0-9a-f]{64}$'::text AND approved_mapping_sha256 ~ '^[0-9a-f]{64}$'::text AND target_registry_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT import_apply_unit_plans_mapping_json_ck CHECK (jsonb_typeof(approved_mapping_json) = 'object'::text),
    CONSTRAINT import_apply_unit_plans_pk PRIMARY KEY (import_session_id, import_unit_id, apply_job_id),
    CONSTRAINT import_apply_unit_plans_sequence_ck CHECK (discovery_sequence > 0),
    CONSTRAINT import_apply_unit_plans_session_sequence_uq UNIQUE (import_session_id, discovery_sequence),
    CONSTRAINT import_apply_unit_plans_target_kind_ck CHECK (target_kind = ANY (ARRAY['view_schema'::text, 'network_flow_table'::text])),
    CONSTRAINT import_apply_unit_plans_target_shape_ck CHECK (target_kind = 'view_schema'::text AND target_view_schema_id IS NOT NULL AND extension_profile_id IS NULL OR target_kind = 'network_flow_table'::text AND target_view_schema_id IS NULL AND extension_profile_id IS NOT NULL),
    CONSTRAINT import_apply_unit_plans_unit_uq UNIQUE (import_unit_id)
);

REVOKE ALL ON TABLE public.import_apply_unit_plans FROM PUBLIC;

REVOKE ALL ON TYPE public.import_apply_unit_plans FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.import_apply_unit_plans TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_apply_unit_plans TO cartulary_recovery;

CREATE TABLE public.import_sessions (
    import_session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    client_txn_id text NOT NULL,
    assistant_profile text NOT NULL,
    source_file_kind text NOT NULL,
    original_filename text NOT NULL,
    source_content_sha256 text NOT NULL,
    source_media_type text NOT NULL,
    source_byte_size bigint NOT NULL,
    parser_profile_id text NOT NULL,
    parser_version text NOT NULL,
    session_status text NOT NULL,
    discovery_job_id uuid,
    apply_job_id uuid,
    selected_unit_ids uuid[] DEFAULT '{}'::uuid[] NOT NULL,
    blocking_diagnostics_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    nonblocking_warning_codes text[] DEFAULT '{}'::text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_sessions_pkey PRIMARY KEY (import_session_id),
    CONSTRAINT import_sessions_session_status_check CHECK (session_status = ANY (ARRAY['created'::text, 'discovered'::text, 'mapped'::text, 'ready_to_apply'::text, 'applying'::text, 'applied'::text, 'partially_applied'::text, 'failed'::text, 'canceled'::text])),
    CONSTRAINT import_sessions_source_byte_size_check CHECK (source_byte_size >= 0),
    CONSTRAINT import_sessions_source_file_kind_check CHECK (source_file_kind = ANY (ARRAY['csv'::text, 'xlsx'::text])),
    CONSTRAINT import_sessions_incident_identity_key UNIQUE (incident_id, import_session_id)
);

REVOKE ALL ON TABLE public.import_sessions FROM PUBLIC;

REVOKE ALL ON TYPE public.import_sessions FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.import_sessions TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_sessions TO cartulary_recovery;

CREATE TABLE public.import_source_streams (
    source_stream_ref text NOT NULL,
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    source_content_sha256 text NOT NULL,
    source_media_type text NOT NULL,
    source_byte_size bigint NOT NULL,
    source_bytes bytea NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_source_streams_import_unit_id_key UNIQUE (import_unit_id),
    CONSTRAINT import_source_streams_pkey PRIMARY KEY (source_stream_ref),
    CONSTRAINT import_source_streams_ref_ck CHECK (source_stream_ref ~ '^impsrc_[0-9a-f]{32}$'::text),
    CONSTRAINT import_source_streams_sha_ck CHECK (source_content_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT import_source_streams_source_byte_size_ck CHECK (source_byte_size >= 0 AND octet_length(source_bytes) = source_byte_size)
);

REVOKE ALL ON TABLE public.import_source_streams FROM PUBLIC;

REVOKE ALL ON TYPE public.import_source_streams FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.import_source_streams TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_source_streams TO cartulary_recovery;

CREATE TABLE public.import_unit_apply_outcomes (
    import_unit_apply_outcome_id uuid DEFAULT gen_random_uuid() NOT NULL,
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    apply_job_id uuid NOT NULL,
    discovery_sequence integer NOT NULL,
    unit_commit_id text NOT NULL,
    outcome_status text NOT NULL,
    actor_user_id uuid NOT NULL,
    target_kind text NOT NULL,
    target_view_schema_id text,
    extension_profile_id text,
    owner_binding_id text NOT NULL,
    source_content_sha256 text NOT NULL,
    mapping_fingerprint text NOT NULL,
    owner_result_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    resource_refs_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    change_set_id uuid,
    error_code text,
    reason_code text,
    committed_at timestamp with time zone NOT NULL,
    error_retryable boolean DEFAULT false NOT NULL,
    error_details_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT import_unit_apply_outcomes_commit_id_ck CHECK (unit_commit_id ~ '^import-unit:[0-9a-f-]{36}:[0-9a-f-]{36}$'::text),
    CONSTRAINT import_unit_apply_outcomes_commit_uq UNIQUE (unit_commit_id),
    CONSTRAINT import_unit_apply_outcomes_error_details_ck CHECK (jsonb_typeof(error_details_json) = 'object'::text),
    CONSTRAINT import_unit_apply_outcomes_error_shape_ck CHECK (outcome_status = 'applied'::text AND error_code IS NULL AND reason_code IS NULL AND error_retryable = false AND error_details_json = '{}'::jsonb OR (outcome_status = ANY (ARRAY['failed'::text, 'canceled'::text])) AND error_code IS NOT NULL AND reason_code IS NOT NULL AND (error_details_json ->> 'reason_code'::text) = reason_code),
    CONSTRAINT import_unit_apply_outcomes_mapping_digest_ck CHECK (mapping_fingerprint ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT import_unit_apply_outcomes_owner_result_ck CHECK (jsonb_typeof(owner_result_json) = ANY (ARRAY['object'::text, 'array'::text])),
    CONSTRAINT import_unit_apply_outcomes_pkey PRIMARY KEY (import_unit_apply_outcome_id),
    CONSTRAINT import_unit_apply_outcomes_resource_refs_ck CHECK (jsonb_typeof(resource_refs_json) = 'array'::text),
    CONSTRAINT import_unit_apply_outcomes_sequence_ck CHECK (discovery_sequence > 0),
    CONSTRAINT import_unit_apply_outcomes_session_sequence_uq UNIQUE (import_session_id, discovery_sequence),
    CONSTRAINT import_unit_apply_outcomes_source_digest_ck CHECK (source_content_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT import_unit_apply_outcomes_status_ck CHECK (outcome_status = ANY (ARRAY['applied'::text, 'failed'::text, 'canceled'::text])),
    CONSTRAINT import_unit_apply_outcomes_target_kind_ck CHECK (target_kind = ANY (ARRAY['view_schema'::text, 'network_flow_table'::text])),
    CONSTRAINT import_unit_apply_outcomes_target_shape_ck CHECK (target_kind = 'view_schema'::text AND target_view_schema_id IS NOT NULL AND extension_profile_id IS NULL AND change_set_id IS NOT NULL OR target_kind = 'network_flow_table'::text AND target_view_schema_id IS NULL AND extension_profile_id IS NOT NULL AND change_set_id IS NULL OR outcome_status <> 'applied'::text AND (target_kind = 'view_schema'::text AND target_view_schema_id IS NOT NULL AND extension_profile_id IS NULL OR target_kind = 'network_flow_table'::text AND target_view_schema_id IS NULL AND extension_profile_id IS NOT NULL)),
    CONSTRAINT import_unit_apply_outcomes_unit_uq UNIQUE (import_unit_id)
);

REVOKE ALL ON TABLE public.import_unit_apply_outcomes FROM PUBLIC;

REVOKE ALL ON TYPE public.import_unit_apply_outcomes FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.import_unit_apply_outcomes TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_unit_apply_outcomes TO cartulary_recovery;

CREATE TABLE public.import_units (
    import_unit_id uuid DEFAULT gen_random_uuid() NOT NULL,
    import_session_id uuid NOT NULL,
    unit_status text NOT NULL,
    locator_kind text NOT NULL,
    locator text NOT NULL,
    source_rect_a1 text NOT NULL,
    header_row_ref integer NOT NULL,
    data_start_row_ref integer NOT NULL,
    inferred_row_count integer NOT NULL,
    inferred_column_count integer NOT NULL,
    warning_codes text[] DEFAULT '{}'::text[] NOT NULL,
    mapping_fingerprint text,
    approved_mapping_json jsonb,
    columns_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    source_rows_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    preview_rows_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    source_stream_ref text,
    approved_target_kind text,
    approved_extension_profile_id text,
    approved_target_view_schema_id text,
    discovery_sequence integer NOT NULL,
    base_import_unit_id uuid,
    operator_region_sequence integer,
    blocking_source_column_ordinals integer[] DEFAULT '{}'::integer[] NOT NULL,
    CONSTRAINT import_units_approved_target_kind_ck CHECK (approved_target_kind IS NULL OR (approved_target_kind = ANY (ARRAY['view_schema'::text, 'network_flow_table'::text]))),
    CONSTRAINT import_units_approved_target_shape_ck CHECK (approved_target_kind IS NULL AND approved_extension_profile_id IS NULL AND approved_target_view_schema_id IS NULL OR approved_target_kind = 'view_schema'::text AND approved_target_view_schema_id IS NOT NULL AND approved_extension_profile_id IS NULL OR approved_target_kind = 'network_flow_table'::text AND approved_target_view_schema_id IS NULL AND approved_extension_profile_id = 'network_flow_activity'::text),
    CONSTRAINT import_units_blocking_source_columns_ck CHECK (array_position(blocking_source_column_ordinals, 0) IS NULL),
    CONSTRAINT import_units_data_start_row_ref_check CHECK (data_start_row_ref > 0),
    CONSTRAINT import_units_discovery_sequence_ck CHECK (discovery_sequence >= 1 AND discovery_sequence <= 2147483647),
    CONSTRAINT import_units_header_row_ref_check CHECK (header_row_ref > 0),
    CONSTRAINT import_units_inferred_column_count_check CHECK (inferred_column_count >= 0),
    CONSTRAINT import_units_inferred_row_count_check CHECK (inferred_row_count >= 0),
    CONSTRAINT import_units_operator_region_shape_ck CHECK (locator_kind = 'operator_region'::text AND base_import_unit_id IS NOT NULL AND operator_region_sequence IS NOT NULL AND operator_region_sequence > 0 OR locator_kind <> 'operator_region'::text AND base_import_unit_id IS NULL AND operator_region_sequence IS NULL),
    CONSTRAINT import_units_pkey PRIMARY KEY (import_unit_id),
    CONSTRAINT import_units_session_discovery_sequence_uq UNIQUE (import_session_id, discovery_sequence),
    CONSTRAINT import_units_source_stream_ref_ck CHECK (source_stream_ref IS NULL OR source_stream_ref ~ '^impsrc_[0-9a-f]{32}$'::text),
    CONSTRAINT import_units_unit_status_check CHECK (unit_status = ANY (ARRAY['discovered'::text, 'selected'::text, 'mapped'::text, 'ready'::text, 'skipped'::text, 'applying'::text, 'applied'::text, 'rejected'::text, 'failed'::text, 'canceled'::text]))
);

REVOKE ALL ON TABLE public.import_units FROM PUBLIC;

REVOKE ALL ON TYPE public.import_units FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.import_units TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_units TO cartulary_recovery;

CREATE INDEX import_apply_journal_change_set_id_fk_idx ON public.import_apply_journal USING btree (change_set_id);

CREATE INDEX import_apply_journal_record_idx ON public.import_apply_journal USING btree (record_id, created_at DESC);

CREATE INDEX import_apply_journal_session_unit_idx ON public.import_apply_journal USING btree (import_session_id, import_unit_id, source_row_ref);

CREATE INDEX import_apply_unit_plans_apply_job_id_fk_idx ON public.import_apply_unit_plans USING btree (apply_job_id);

CREATE INDEX import_sessions_apply_job_id_fk_idx ON public.import_sessions USING btree (apply_job_id);

CREATE INDEX import_sessions_apply_job_id_scope_idx ON public.import_sessions (incident_id, apply_job_id);

CREATE INDEX import_sessions_created_by_lookup_idx ON public.import_sessions USING btree (created_by_user_id, created_at DESC, import_session_id);

CREATE INDEX import_sessions_discovery_job_id_fk_idx ON public.import_sessions USING btree (discovery_job_id);

CREATE INDEX import_sessions_discovery_job_id_scope_idx ON public.import_sessions (incident_id, discovery_job_id);

CREATE INDEX import_sessions_incident_lookup_idx ON public.import_sessions USING btree (incident_id, created_at DESC, import_session_id);

CREATE INDEX import_source_streams_session_idx ON public.import_source_streams USING btree (import_session_id, import_unit_id);

CREATE INDEX import_unit_apply_outcomes_actor_user_id_fk_idx ON public.import_unit_apply_outcomes USING btree (actor_user_id);

CREATE INDEX import_unit_apply_outcomes_apply_job_id_fk_idx ON public.import_unit_apply_outcomes USING btree (apply_job_id);

CREATE INDEX import_unit_apply_outcomes_change_set_id_fk_idx ON public.import_unit_apply_outcomes USING btree (change_set_id);

CREATE INDEX import_unit_apply_outcomes_finalize_idx ON public.import_unit_apply_outcomes USING btree (import_session_id, discovery_sequence, import_unit_id);

CREATE INDEX import_unit_apply_outcomes_import_session_id_import_un_71ce6c33 ON public.import_unit_apply_outcomes USING btree (import_session_id, import_unit_id, apply_job_id);

CREATE INDEX import_units_base_import_unit_id_fk_idx ON public.import_units USING btree (base_import_unit_id);

CREATE UNIQUE INDEX import_units_operator_region_identity_uq ON public.import_units USING btree (base_import_unit_id, source_rect_a1) WHERE (locator_kind = 'operator_region'::text);

CREATE UNIQUE INDEX import_units_operator_region_sequence_uq ON public.import_units USING btree (import_session_id, operator_region_sequence) WHERE (locator_kind = 'operator_region'::text);

CREATE INDEX import_units_session_lookup_idx ON public.import_units USING btree (import_session_id, created_at, import_unit_id);

CREATE UNIQUE INDEX import_units_source_stream_ref_idx ON public.import_units USING btree (source_stream_ref) WHERE (source_stream_ref IS NOT NULL);

ALTER TABLE public.import_apply_journal ADD CONSTRAINT import_apply_journal_change_set_id_fkey FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_apply_journal ADD CONSTRAINT import_apply_journal_import_session_id_fkey FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_apply_journal ADD CONSTRAINT import_apply_journal_import_unit_id_fkey FOREIGN KEY (import_unit_id) REFERENCES public.import_units(import_unit_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_apply_journal ADD CONSTRAINT import_apply_journal_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.records(record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_apply_unit_plans ADD CONSTRAINT import_apply_unit_plans_job_fk FOREIGN KEY (apply_job_id) REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.import_apply_unit_plans ADD CONSTRAINT import_apply_unit_plans_session_fk FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_apply_unit_plans ADD CONSTRAINT import_apply_unit_plans_unit_fk FOREIGN KEY (import_unit_id) REFERENCES public.import_units(import_unit_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_sessions ADD CONSTRAINT import_sessions_apply_job_id_fkey FOREIGN KEY (incident_id, apply_job_id) REFERENCES public.jobs(incident_id, job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.import_sessions ADD CONSTRAINT import_sessions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.import_sessions ADD CONSTRAINT import_sessions_discovery_job_id_fkey FOREIGN KEY (incident_id, discovery_job_id) REFERENCES public.jobs(incident_id, job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.import_sessions ADD CONSTRAINT import_sessions_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_source_streams ADD CONSTRAINT import_source_streams_import_session_id_fkey FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_source_streams ADD CONSTRAINT import_source_streams_import_unit_id_fkey FOREIGN KEY (import_unit_id) REFERENCES public.import_units(import_unit_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_unit_apply_outcomes ADD CONSTRAINT import_unit_apply_outcomes_actor_fk FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.import_unit_apply_outcomes ADD CONSTRAINT import_unit_apply_outcomes_change_set_fk FOREIGN KEY (change_set_id) REFERENCES public.change_sets(change_set_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_unit_apply_outcomes ADD CONSTRAINT import_unit_apply_outcomes_job_fk FOREIGN KEY (apply_job_id) REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.import_unit_apply_outcomes ADD CONSTRAINT import_unit_apply_outcomes_plan_fk FOREIGN KEY (import_session_id, import_unit_id, apply_job_id) REFERENCES public.import_apply_unit_plans(import_session_id, import_unit_id, apply_job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.import_unit_apply_outcomes ADD CONSTRAINT import_unit_apply_outcomes_session_fk FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_unit_apply_outcomes ADD CONSTRAINT import_unit_apply_outcomes_unit_fk FOREIGN KEY (import_unit_id) REFERENCES public.import_units(import_unit_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_units ADD CONSTRAINT import_units_base_import_unit_id_fkey FOREIGN KEY (base_import_unit_id) REFERENCES public.import_units(import_unit_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.import_units ADD CONSTRAINT import_units_import_session_id_fkey FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE ON UPDATE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.import_units DROP CONSTRAINT import_units_import_session_id_fkey;
ALTER TABLE public.import_units DROP CONSTRAINT import_units_base_import_unit_id_fkey;
ALTER TABLE public.import_unit_apply_outcomes DROP CONSTRAINT import_unit_apply_outcomes_unit_fk;
ALTER TABLE public.import_unit_apply_outcomes DROP CONSTRAINT import_unit_apply_outcomes_session_fk;
ALTER TABLE public.import_unit_apply_outcomes DROP CONSTRAINT import_unit_apply_outcomes_plan_fk;
ALTER TABLE public.import_unit_apply_outcomes DROP CONSTRAINT import_unit_apply_outcomes_job_fk;
ALTER TABLE public.import_unit_apply_outcomes DROP CONSTRAINT import_unit_apply_outcomes_change_set_fk;
ALTER TABLE public.import_unit_apply_outcomes DROP CONSTRAINT import_unit_apply_outcomes_actor_fk;
ALTER TABLE public.import_source_streams DROP CONSTRAINT import_source_streams_import_unit_id_fkey;
ALTER TABLE public.import_source_streams DROP CONSTRAINT import_source_streams_import_session_id_fkey;
ALTER TABLE public.import_sessions DROP CONSTRAINT import_sessions_incident_id_fkey;
ALTER TABLE public.import_sessions DROP CONSTRAINT import_sessions_discovery_job_id_fkey;
ALTER TABLE public.import_sessions DROP CONSTRAINT import_sessions_created_by_user_id_fkey;
ALTER TABLE public.import_sessions DROP CONSTRAINT import_sessions_apply_job_id_fkey;
ALTER TABLE public.import_apply_unit_plans DROP CONSTRAINT import_apply_unit_plans_unit_fk;
ALTER TABLE public.import_apply_unit_plans DROP CONSTRAINT import_apply_unit_plans_session_fk;
ALTER TABLE public.import_apply_unit_plans DROP CONSTRAINT import_apply_unit_plans_job_fk;
ALTER TABLE public.import_apply_journal DROP CONSTRAINT import_apply_journal_record_id_fkey;
ALTER TABLE public.import_apply_journal DROP CONSTRAINT import_apply_journal_import_unit_id_fkey;
ALTER TABLE public.import_apply_journal DROP CONSTRAINT import_apply_journal_import_session_id_fkey;
ALTER TABLE public.import_apply_journal DROP CONSTRAINT import_apply_journal_change_set_id_fkey;
DROP TABLE public.import_units;
DROP TABLE public.import_unit_apply_outcomes;
DROP TABLE public.import_source_streams;
DROP TABLE public.import_sessions;
DROP TABLE public.import_apply_unit_plans;
DROP TABLE public.import_apply_journal;

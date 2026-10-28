-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: reporting.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.reporting_composition_preview_output_files (
    preview_attempt_id uuid NOT NULL,
    bundle_path text NOT NULL,
    role text NOT NULL,
    media_type text NOT NULL,
    file_sha256 text NOT NULL,
    size_bytes bigint NOT NULL,
    storage_kind text NOT NULL,
    object_ref text,
    inline_bytes bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_composition_preview_output_files_path_ck CHECK (bundle_path <> ''::text AND bundle_path !~ '(^/|(^|/)\.\.?(/|$))'::text),
    CONSTRAINT reporting_composition_preview_output_files_pkey PRIMARY KEY (preview_attempt_id, bundle_path),
    CONSTRAINT reporting_composition_preview_output_files_sha_ck CHECK (file_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_composition_preview_output_files_size_ck CHECK (size_bytes >= 0),
    CONSTRAINT reporting_composition_preview_output_files_storage_ck CHECK (storage_kind = 'database_inline'::text AND inline_bytes IS NOT NULL AND object_ref IS NULL OR storage_kind = 'object_store'::text AND inline_bytes IS NULL AND object_ref IS NOT NULL AND object_ref <> ''::text),
    CONSTRAINT reporting_preview_public_members_ck CHECK (role <> 'sensitive_reveal_map'::text AND bundle_path !~~ 'internal/%'::text AND media_type <> 'application/vnd.cartulary.reporting-reveal-map+json'::text)
);

REVOKE ALL ON TABLE public.reporting_composition_preview_output_files FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_composition_preview_output_files FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_composition_preview_output_files TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_composition_preview_output_files TO cartulary_recovery;

CREATE TABLE public.reporting_composition_preview_outputs (
    preview_attempt_id uuid NOT NULL,
    render_attempt_id uuid NOT NULL,
    release_scope text DEFAULT 'internal_draft'::text NOT NULL,
    output_kind text NOT NULL,
    output_media_type text NOT NULL,
    output_sha256 text NOT NULL,
    redaction_profile_id text NOT NULL,
    redaction_profile_version text NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    redaction_manifest_sha256 text NOT NULL,
    redaction_manifest_json jsonb NOT NULL,
    bundle_manifest_sha256 text NOT NULL,
    bundle_manifest_json jsonb NOT NULL,
    primary_bundle_path text NOT NULL,
    primary_media_type text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_composition_preview_outputs_kind_ck CHECK (output_kind = ANY (ARRAY['slidev'::text, 'mermaid'::text])),
    CONSTRAINT reporting_composition_preview_outputs_manifest_ck CHECK (jsonb_typeof(redaction_manifest_json) = 'object'::text AND jsonb_typeof(bundle_manifest_json) = 'object'::text),
    CONSTRAINT reporting_composition_preview_outputs_pkey PRIMARY KEY (preview_attempt_id),
    CONSTRAINT reporting_composition_preview_outputs_render_attempt_key UNIQUE (render_attempt_id),
    CONSTRAINT reporting_composition_preview_outputs_scope_ck CHECK (release_scope = 'internal_draft'::text),
    CONSTRAINT reporting_composition_preview_outputs_sha_ck CHECK (output_sha256 ~ '^[a-f0-9]{64}$'::text AND redaction_profile_sha256 ~ '^[a-f0-9]{64}$'::text AND redaction_manifest_sha256 ~ '^[a-f0-9]{64}$'::text AND bundle_manifest_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_preview_redaction_manifest_v5_ck CHECK (((redaction_manifest_json ->> 'schema_id'::text) = 'cartulary.redaction_manifest.v5'::text AND jsonb_typeof(redaction_manifest_json -> 'container_decisions'::text) = 'array'::text) IS TRUE)
);

REVOKE ALL ON TABLE public.reporting_composition_preview_outputs FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_composition_preview_outputs FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_composition_preview_outputs TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_composition_preview_outputs TO cartulary_recovery;

CREATE TABLE public.reporting_imported_artifact_files (
    incident_id uuid NOT NULL,
    bundle_path text NOT NULL,
    operation_id uuid NOT NULL,
    size_bytes bigint NOT NULL,
    sha256 text NOT NULL,
    content bytea NOT NULL,
    CONSTRAINT reporting_imported_artifact_bytes_ck CHECK (octet_length(content) = size_bytes AND encode(sha256(content), 'hex'::text) = sha256),
    CONSTRAINT reporting_imported_artifact_files_pk PRIMARY KEY (incident_id, bundle_path),
    CONSTRAINT reporting_imported_artifact_path_ck CHECK (bundle_path ~~ 'ext/snapshots/%'::text AND octet_length(bundle_path) <= 640 AND bundle_path !~ '(^/|(^|/)\.\.?(/|$)|//)'::text AND POSITION((chr(92)) IN (bundle_path)) = 0),
    CONSTRAINT reporting_imported_artifact_sha256_ck CHECK (sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT reporting_imported_artifact_size_ck CHECK (size_bytes >= 0)
);

REVOKE ALL ON TABLE public.reporting_imported_artifact_files FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_imported_artifact_files FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reporting_imported_artifact_files TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_imported_artifact_files TO cartulary_recovery;

GRANT USAGE ON TYPE public.reporting_imported_artifact_files TO cartulary_runtime;

GRANT USAGE ON TYPE public.reporting_imported_artifact_files TO cartulary_recovery;

CREATE TABLE public.reporting_job_payloads (
    job_id uuid NOT NULL,
    job_kind text NOT NULL,
    incident_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    request_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    admitted_profile_sha256 text,
    render_failure_summary bytea,
    render_failure_sha256 text,
    admitted_graph_sources_sha256 text,
    admitted_template_manifest_sha256 text,
    snapshot_admission_json jsonb,
    snapshot_admission_sha256 text,
    render_admission_sha256 text,
    render_preparation_json jsonb,
    render_preparation_sha256 text,
    CONSTRAINT reporting_graph_input_v1_ck CHECK (job_kind = 'snapshot_create'::text AND admitted_graph_sources_sha256 IS NULL OR (job_kind <> 'snapshot_create'::text AND admitted_graph_sources_sha256 ~ '^[0-9a-f]{64}$'::text AND ((request_json -> 'graph_sources'::text) ->> 'schema_id'::text) = 'cartulary.reporting_graph_input.v1'::text AND jsonb_typeof((request_json -> 'graph_sources'::text) -> 'results'::text) = 'array'::text) IS TRUE),
    CONSTRAINT reporting_job_payloads_job_kind_check CHECK (job_kind = ANY (ARRAY['snapshot_create'::text, 'release_create'::text, 'composition_preview'::text])),
    CONSTRAINT reporting_job_payloads_pkey PRIMARY KEY (job_id),
    CONSTRAINT reporting_render_failure_summary_ck CHECK (render_failure_summary IS NULL AND render_failure_sha256 IS NULL OR (job_kind = ANY (ARRAY['release_create'::text, 'composition_preview'::text])) AND render_failure_summary IS NOT NULL AND octet_length(render_failure_summary) >= 1 AND octet_length(render_failure_summary) <= 1048576 AND render_failure_sha256 IS NOT NULL AND render_failure_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_render_recipe_v1_ck CHECK (job_kind = 'snapshot_create'::text AND admitted_profile_sha256 IS NULL AND render_admission_sha256 IS NULL AND render_preparation_json IS NULL AND render_preparation_sha256 IS NULL OR (job_kind = ANY (ARRAY['release_create'::text, 'composition_preview'::text])) AND ((request_json ->> 'schema_id'::text) = 'cartulary.reporting_render_admission.v1'::text AND (request_json ->> 'preparation_algorithm'::text) = 'materialize_retained_reporting_render_v1'::text AND render_admission_sha256 ~ '^[a-f0-9]{64}$'::text AND (request_json ->> 'source_capture_sha256'::text) ~ '^[a-f0-9]{64}$'::text AND (request_json ->> 'authored_redaction_profile_sha256'::text) = admitted_profile_sha256 AND admitted_profile_sha256 ~ '^[a-f0-9]{64}$'::text AND ((request_json -> 'authored_redaction_profile'::text) ->> 'schema_id'::text) = 'cartulary.redaction_profile.v3'::text AND ((request_json -> 'authored_redaction_profile'::text) ->> 'version'::text) = '5'::text AND ((request_json -> 'authored_redaction_profile'::text) -> 'allowed_release_scopes'::text) ? (request_json ->> 'release_scope'::text) AND ((request_json ->> 'release_scope'::text) <> 'external_release'::text OR ((request_json -> 'authored_redaction_profile'::text) ->> 'profile_status'::text) = 'adopted_current'::text) AND ((request_json -> 'template_contract'::text) ->> 'schema_id'::text) = 'cartulary.reporting_template_pack_manifest.v3'::text AND (request_json ->> 'template_sha256'::text) = admitted_template_manifest_sha256 AND (request_json ->> 'graph_sources_sha256'::text) = admitted_graph_sources_sha256 AND NOT request_json ?| ARRAY['export_model'::text, 'redaction_profile'::text, 'source_facts'::text] AND (job_kind = 'release_create'::text AND jsonb_typeof((request_json -> 'identity'::text) -> 'release_id'::text) = 'string'::text AND ((request_json -> 'identity'::text) -> 'preview_attempt_id'::text) = 'null'::jsonb OR job_kind = 'composition_preview'::text AND jsonb_typeof((request_json -> 'identity'::text) -> 'preview_attempt_id'::text) = 'string'::text AND ((request_json -> 'identity'::text) -> 'release_id'::text) = 'null'::jsonb) AND (render_preparation_json IS NULL AND render_preparation_sha256 IS NULL OR (render_preparation_sha256 ~ '^[a-f0-9]{64}$'::text AND (render_preparation_json ->> 'schema_id'::text) = 'cartulary.reporting_render_preparation.v1'::text AND (render_preparation_json ->> 'admission_sha256'::text) = render_admission_sha256 AND (render_preparation_json ->> 'source_capture_sha256'::text) = (request_json ->> 'source_capture_sha256'::text) AND (render_preparation_json ->> 'export_model_sha256'::text) ~ '^[a-f0-9]{64}$'::text AND (render_preparation_json ->> 'source_facts_sha256'::text) ~ '^[a-f0-9]{64}$'::text AND (render_preparation_json ->> 'redaction_profile_sha256'::text) ~ '^[a-f0-9]{64}$'::text) IS TRUE)) IS TRUE),
    CONSTRAINT reporting_snapshot_admission_v1_ck CHECK (job_kind = 'snapshot_create'::text AND (jsonb_typeof(snapshot_admission_json) = 'object'::text AND (snapshot_admission_json ->> 'schema_id'::text) = 'cartulary.reporting_snapshot_admission.v1'::text AND snapshot_admission_sha256 ~ '^[a-f0-9]{64}$'::text AND (request_json IS NULL OR jsonb_typeof(request_json) = 'object'::text)) IS TRUE OR job_kind <> 'snapshot_create'::text AND request_json IS NOT NULL AND snapshot_admission_json IS NULL AND snapshot_admission_sha256 IS NULL),
    CONSTRAINT reporting_template_manifest_admission_v1_ck CHECK (job_kind = 'snapshot_create'::text AND admitted_template_manifest_sha256 IS NULL OR job_kind <> 'snapshot_create'::text AND (admitted_template_manifest_sha256 ~ '^[0-9a-f]{64}$'::text) IS TRUE)
);

REVOKE ALL ON TABLE public.reporting_job_payloads FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_job_payloads FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_job_payloads TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_job_payloads TO cartulary_recovery;

CREATE TABLE public.reporting_preview_reveal_maps (
    preview_attempt_id uuid NOT NULL,
    reveal_map_sha256 text NOT NULL,
    canonical_bytes bytea NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT reporting_preview_reveal_maps_bytes_ck CHECK (octet_length(canonical_bytes) > 0),
    CONSTRAINT reporting_preview_reveal_maps_digest_ck CHECK (reveal_map_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_preview_reveal_maps_pk PRIMARY KEY (preview_attempt_id)
);

REVOKE ALL ON TABLE public.reporting_preview_reveal_maps FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_preview_reveal_maps FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reporting_preview_reveal_maps TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_preview_reveal_maps TO cartulary_recovery;

GRANT USAGE ON TYPE public.reporting_preview_reveal_maps TO cartulary_runtime;

GRANT USAGE ON TYPE public.reporting_preview_reveal_maps TO cartulary_recovery;

CREATE TABLE public.reporting_release_approvals (
    approval_id uuid DEFAULT gen_random_uuid() NOT NULL,
    release_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    approval_role text NOT NULL,
    reason text,
    approval_tuple_json jsonb NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    output_sha256 text NOT NULL,
    redaction_manifest_sha256 text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_release_approvals_approval_role_check CHECK (approval_role = ANY (ARRAY['reviewer'::text, 'admin'::text])),
    CONSTRAINT reporting_release_approvals_pkey PRIMARY KEY (approval_id),
    CONSTRAINT reporting_release_approvals_release_id_actor_user_id_approv_key UNIQUE (release_id, actor_user_id, approval_role)
);

REVOKE ALL ON TABLE public.reporting_release_approvals FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_release_approvals FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_release_approvals TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_release_approvals TO cartulary_recovery;

CREATE TABLE public.reporting_release_reveal_maps (
    release_id uuid NOT NULL,
    reveal_map_sha256 text NOT NULL,
    canonical_bytes bytea NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT reporting_release_reveal_maps_bytes_ck CHECK (octet_length(canonical_bytes) > 0),
    CONSTRAINT reporting_release_reveal_maps_digest_ck CHECK (reveal_map_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_release_reveal_maps_pk PRIMARY KEY (release_id)
);

REVOKE ALL ON TABLE public.reporting_release_reveal_maps FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_release_reveal_maps FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reporting_release_reveal_maps TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_release_reveal_maps TO cartulary_recovery;

GRANT USAGE ON TYPE public.reporting_release_reveal_maps TO cartulary_runtime;

GRANT USAGE ON TYPE public.reporting_release_reveal_maps TO cartulary_recovery;

CREATE TABLE public.reporting_releases (
    release_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    snapshot_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    client_txn_id text NOT NULL,
    release_scope text NOT NULL,
    release_state text NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    source_change_set_high_watermark text NOT NULL,
    derivation_version text NOT NULL,
    export_model_sha256 text NOT NULL,
    template_id text NOT NULL,
    template_version text NOT NULL,
    redaction_profile_id text NOT NULL,
    redaction_profile_version text NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    output_kind text NOT NULL,
    output_media_type text,
    output_sha256 text,
    redaction_manifest_sha256 text,
    redaction_manifest_json jsonb,
    create_job_id uuid NOT NULL,
    render_failed_reason_code text,
    approved_at timestamp with time zone,
    published_at timestamp with time zone,
    invalidated_at timestamp with time zone,
    invalidation_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    recipient_partition_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    output_options jsonb DEFAULT '{}'::jsonb NOT NULL,
    graph_projection_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    composition_id uuid,
    composition_version text,
    composition_sha256 text,
    render_admitted_at timestamp with time zone NOT NULL,
    CONSTRAINT reporting_redaction_manifest_v5_ck CHECK (redaction_manifest_json IS NULL OR ((redaction_manifest_json ->> 'schema_id'::text) = 'cartulary.redaction_manifest.v5'::text AND jsonb_typeof(redaction_manifest_json -> 'container_decisions'::text) = 'array'::text) IS TRUE),
    CONSTRAINT reporting_releases_composition_tuple_ck CHECK ((composition_id IS NULL AND composition_version IS NULL AND composition_sha256 IS NULL) OR (composition_id IS NOT NULL AND composition_version IS NOT NULL AND composition_sha256 IS NOT NULL AND composition_version ~ '^v[1-9][0-9]*$' AND composition_sha256 ~ '^[a-f0-9]{64}$')),
    CONSTRAINT reporting_releases_graph_projection_refs_ck CHECK (jsonb_typeof(graph_projection_refs) = 'array'::text),
    CONSTRAINT reporting_releases_output_kind_check CHECK (output_kind = ANY (ARRAY['slidev'::text, 'mermaid'::text])),
    CONSTRAINT reporting_releases_output_options_ck CHECK (jsonb_typeof(output_options) = 'object'::text),
    CONSTRAINT reporting_releases_pkey PRIMARY KEY (release_id),
    CONSTRAINT reporting_releases_recipient_partition_refs_ck CHECK (jsonb_typeof(recipient_partition_refs) = 'array'::text AND (release_scope = 'external_release'::text OR recipient_partition_refs = '[]'::jsonb)),
    CONSTRAINT reporting_releases_release_scope_check CHECK (release_scope = ANY (ARRAY['internal_draft'::text, 'internal_review'::text, 'external_release'::text])),
    CONSTRAINT reporting_releases_release_state_check CHECK (release_state = ANY (ARRAY['pending_approval'::text, 'approved'::text, 'published'::text, 'invalidated'::text, 'render_failed'::text]))
);

REVOKE ALL ON TABLE public.reporting_releases FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_releases FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_releases TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_releases TO cartulary_recovery;

CREATE TABLE public.reporting_render_bundle_files (
    release_id uuid NOT NULL,
    bundle_path text NOT NULL,
    role text NOT NULL,
    media_type text NOT NULL,
    file_sha256 text NOT NULL,
    size_bytes bigint NOT NULL,
    storage_kind text NOT NULL,
    object_ref text,
    inline_bytes bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_bundle_public_members_ck CHECK (role <> 'sensitive_reveal_map'::text AND bundle_path !~~ 'internal/%'::text AND media_type <> 'application/vnd.cartulary.reporting-reveal-map+json'::text),
    CONSTRAINT reporting_render_bundle_files_bundle_path_check CHECK (bundle_path <> ''::text AND bundle_path !~ '(^/|(^|/)\.\.?(/|$))'::text),
    CONSTRAINT reporting_render_bundle_files_check CHECK (storage_kind = 'database_inline'::text AND inline_bytes IS NOT NULL AND object_ref IS NULL OR storage_kind = 'object_store'::text AND inline_bytes IS NULL AND object_ref IS NOT NULL AND object_ref <> ''::text),
    CONSTRAINT reporting_render_bundle_files_file_sha256_check CHECK (file_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_render_bundle_files_media_type_check CHECK (media_type <> ''::text),
    CONSTRAINT reporting_render_bundle_files_pkey PRIMARY KEY (release_id, bundle_path),
    CONSTRAINT reporting_render_bundle_files_role_check CHECK (role <> ''::text),
    CONSTRAINT reporting_render_bundle_files_size_bytes_check CHECK (size_bytes >= 0),
    CONSTRAINT reporting_render_bundle_files_storage_kind_check CHECK (storage_kind = ANY (ARRAY['database_inline'::text, 'object_store'::text]))
);

REVOKE ALL ON TABLE public.reporting_render_bundle_files FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_render_bundle_files FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_render_bundle_files TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_render_bundle_files TO cartulary_recovery;

CREATE TABLE public.reporting_render_bundles (
    release_id uuid NOT NULL,
    bundle_manifest_sha256 text NOT NULL,
    bundle_manifest_json jsonb NOT NULL,
    primary_bundle_path text NOT NULL,
    primary_media_type text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_bundle_model_identity_v3_ck CHECK ((jsonb_typeof(bundle_manifest_json) = 'object'::text AND (bundle_manifest_json ->> 'schema_id'::text) = 'cartulary.render_bundle_manifest.v3'::text AND (bundle_manifest_json ->> 'manifest_version'::text) = '3'::text AND (bundle_manifest_json ->> 'admitted_export_model_sha256'::text) ~ '^[0-9a-f]{64}$'::text AND (bundle_manifest_json ->> 'export_model_sha256'::text) ~ '^[0-9a-f]{64}$'::text AND jsonb_typeof(bundle_manifest_json -> 'reference_packs'::text) = 'object'::text AND ((bundle_manifest_json -> 'reference_packs'::text) ->> 'pack_set_id'::text) ~ '^rpset_[0-9a-f]{64}$'::text AND ((bundle_manifest_json -> 'reference_packs'::text) ->> 'pack_set_sha256'::text) ~ '^[0-9a-f]{64}$'::text AND jsonb_typeof((bundle_manifest_json -> 'reference_packs'::text) -> 'provenance'::text) = 'array'::text AND ((bundle_manifest_json -> 'reference_packs'::text) ->> 'pack_set_id'::text) = ('rpset_'::text || ((bundle_manifest_json -> 'reference_packs'::text) ->> 'pack_set_sha256'::text)) AND (bundle_manifest_json -> 'reference_packs'::text) ?& ARRAY['pack_set_id'::text, 'pack_set_sha256'::text, 'provenance'::text]) IS TRUE),
    CONSTRAINT reporting_render_bundles_bundle_manifest_sha256_check CHECK (bundle_manifest_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_render_bundles_pkey PRIMARY KEY (release_id),
    CONSTRAINT reporting_render_bundles_primary_bundle_path_check CHECK (primary_bundle_path <> ''::text),
    CONSTRAINT reporting_render_bundles_primary_media_type_check CHECK (primary_media_type <> ''::text),
    CONSTRAINT reporting_structured_redaction_v1_ck CHECK (((bundle_manifest_json ->> 'renderer_version'::text) = '21'::text) IS TRUE)
);

REVOKE ALL ON TABLE public.reporting_render_bundles FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_render_bundles FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_render_bundles TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_render_bundles TO cartulary_recovery;

CREATE TABLE public.reporting_snapshots (
    snapshot_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    client_txn_id text NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    source_change_set_high_watermark text NOT NULL,
    source_boundary_json jsonb NOT NULL,
    derivation_version text NOT NULL,
    source_capture_sha256 text NOT NULL,
    create_job_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_snapshot_capture_v1_ck CHECK (source_capture_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT reporting_snapshots_pkey PRIMARY KEY (snapshot_id),
    CONSTRAINT reporting_snapshots_source_boundary_json_check CHECK (jsonb_typeof(source_boundary_json) = 'object'::text),
    CONSTRAINT reporting_snapshots_source_boundary_json_ck CHECK (jsonb_typeof(source_boundary_json) = 'object'::text),
    CONSTRAINT reporting_snapshots_incident_identity_key UNIQUE (incident_id, snapshot_id)
);

REVOKE ALL ON TABLE public.reporting_snapshots FROM PUBLIC;

REVOKE ALL ON TYPE public.reporting_snapshots FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reporting_snapshots TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_snapshots TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.reporting_reject_imported_artifact_update()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$ BEGIN
    RAISE EXCEPTION 'reporting imported artifacts are immutable' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.reporting_reject_imported_artifact_update() FROM PUBLIC;

CREATE INDEX reporting_imported_artifact_files_operation_id_fk_idx ON public.reporting_imported_artifact_files USING btree (operation_id);

CREATE INDEX reporting_job_payloads_actor_user_id_fk_idx ON public.reporting_job_payloads USING btree (actor_user_id);

CREATE INDEX reporting_job_payloads_incident_lookup_idx ON public.reporting_job_payloads USING btree (incident_id, created_at DESC, job_id);

CREATE INDEX reporting_job_payloads_job_id_scope_idx ON public.reporting_job_payloads (incident_id, job_id);

CREATE INDEX reporting_release_approvals_actor_user_id_fk_idx ON public.reporting_release_approvals USING btree (actor_user_id);

CREATE INDEX reporting_release_approvals_release_lookup_idx ON public.reporting_release_approvals USING btree (release_id, approval_role, created_at, approval_id);

CREATE INDEX reporting_releases_create_job_id_fk_idx ON public.reporting_releases USING btree (create_job_id);

CREATE INDEX reporting_releases_create_job_id_scope_idx ON public.reporting_releases (incident_id, create_job_id);

CREATE INDEX reporting_releases_created_by_lookup_idx ON public.reporting_releases USING btree (created_by_user_id, created_at DESC, release_id);

CREATE INDEX reporting_releases_incident_lookup_idx ON public.reporting_releases USING btree (incident_id, created_at DESC, release_id);

CREATE INDEX reporting_releases_snapshot_id_scope_idx ON public.reporting_releases (incident_id, snapshot_id);

CREATE INDEX reporting_releases_snapshot_lookup_idx ON public.reporting_releases USING btree (snapshot_id, created_at DESC, release_id);

CREATE INDEX reporting_render_bundle_files_release_idx ON public.reporting_render_bundle_files USING btree (release_id, role, bundle_path);

CREATE INDEX reporting_snapshots_create_job_id_fk_idx ON public.reporting_snapshots USING btree (create_job_id);

CREATE INDEX reporting_snapshots_create_job_id_scope_idx ON public.reporting_snapshots (incident_id, create_job_id);

CREATE INDEX reporting_snapshots_created_by_lookup_idx ON public.reporting_snapshots USING btree (created_by_user_id, created_at DESC, snapshot_id);

CREATE INDEX reporting_snapshots_incident_lookup_idx ON public.reporting_snapshots USING btree (incident_id, created_at DESC, snapshot_id);

ALTER TABLE public.reporting_composition_preview_output_files ADD CONSTRAINT reporting_composition_preview_output_files_attempt_fk FOREIGN KEY (preview_attempt_id) REFERENCES public.reporting_composition_preview_outputs(preview_attempt_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_composition_preview_outputs ADD CONSTRAINT reporting_composition_preview_outputs_job_fk FOREIGN KEY (render_attempt_id) REFERENCES public.jobs(job_id) ON DELETE RESTRICT ON UPDATE NO ACTION;

ALTER TABLE public.reporting_imported_artifact_files ADD CONSTRAINT reporting_imported_artifact_incident_fk FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE RESTRICT ON DELETE CASCADE;

ALTER TABLE public.reporting_imported_artifact_files ADD CONSTRAINT reporting_imported_artifact_operation_fk FOREIGN KEY (operation_id) REFERENCES public.jobs(job_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_job_payloads_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_job_payloads_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_job_payloads ADD CONSTRAINT reporting_job_payloads_job_id_fkey FOREIGN KEY (incident_id, job_id) REFERENCES public.jobs(incident_id, job_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_preview_reveal_maps ADD CONSTRAINT reporting_preview_reveal_maps_parent_fk FOREIGN KEY (preview_attempt_id) REFERENCES public.reporting_composition_preview_outputs(preview_attempt_id) ON UPDATE RESTRICT ON DELETE CASCADE;

ALTER TABLE public.reporting_release_approvals ADD CONSTRAINT reporting_release_approvals_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.reporting_release_approvals ADD CONSTRAINT reporting_release_approvals_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.reporting_releases(release_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_release_reveal_maps ADD CONSTRAINT reporting_release_reveal_maps_parent_fk FOREIGN KEY (release_id) REFERENCES public.reporting_render_bundles(release_id) ON UPDATE RESTRICT ON DELETE CASCADE;

ALTER TABLE public.reporting_releases ADD CONSTRAINT reporting_releases_create_job_id_fkey FOREIGN KEY (incident_id, create_job_id) REFERENCES public.jobs(incident_id, job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.reporting_releases ADD CONSTRAINT reporting_releases_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.reporting_releases ADD CONSTRAINT reporting_releases_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_releases ADD CONSTRAINT reporting_releases_snapshot_id_fkey FOREIGN KEY (incident_id, snapshot_id) REFERENCES public.reporting_snapshots(incident_id, snapshot_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_render_bundle_files ADD CONSTRAINT reporting_render_bundle_files_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.reporting_render_bundles(release_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_render_bundles ADD CONSTRAINT reporting_render_bundles_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.reporting_releases(release_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshots_create_job_id_fkey FOREIGN KEY (incident_id, create_job_id) REFERENCES public.jobs(incident_id, job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshots_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.reporting_snapshots ADD CONSTRAINT reporting_snapshots_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

CREATE TRIGGER reporting_imported_artifacts_immutable BEFORE UPDATE ON public.reporting_imported_artifact_files FOR EACH ROW EXECUTE FUNCTION public.reporting_reject_imported_artifact_update();

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER reporting_imported_artifacts_immutable ON public.reporting_imported_artifact_files;
ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_snapshots_incident_id_fkey;
ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_snapshots_created_by_user_id_fkey;
ALTER TABLE public.reporting_snapshots DROP CONSTRAINT reporting_snapshots_create_job_id_fkey;
ALTER TABLE public.reporting_render_bundles DROP CONSTRAINT reporting_render_bundles_release_id_fkey;
ALTER TABLE public.reporting_render_bundle_files DROP CONSTRAINT reporting_render_bundle_files_release_id_fkey;
ALTER TABLE public.reporting_releases DROP CONSTRAINT reporting_releases_snapshot_id_fkey;
ALTER TABLE public.reporting_releases DROP CONSTRAINT reporting_releases_incident_id_fkey;
ALTER TABLE public.reporting_releases DROP CONSTRAINT reporting_releases_created_by_user_id_fkey;
ALTER TABLE public.reporting_releases DROP CONSTRAINT reporting_releases_create_job_id_fkey;
ALTER TABLE public.reporting_release_reveal_maps DROP CONSTRAINT reporting_release_reveal_maps_parent_fk;
ALTER TABLE public.reporting_release_approvals DROP CONSTRAINT reporting_release_approvals_release_id_fkey;
ALTER TABLE public.reporting_release_approvals DROP CONSTRAINT reporting_release_approvals_actor_user_id_fkey;
ALTER TABLE public.reporting_preview_reveal_maps DROP CONSTRAINT reporting_preview_reveal_maps_parent_fk;
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_job_payloads_job_id_fkey;
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_job_payloads_incident_id_fkey;
ALTER TABLE public.reporting_job_payloads DROP CONSTRAINT reporting_job_payloads_actor_user_id_fkey;
ALTER TABLE public.reporting_imported_artifact_files DROP CONSTRAINT reporting_imported_artifact_operation_fk;
ALTER TABLE public.reporting_imported_artifact_files DROP CONSTRAINT reporting_imported_artifact_incident_fk;
ALTER TABLE public.reporting_composition_preview_outputs DROP CONSTRAINT reporting_composition_preview_outputs_job_fk;
ALTER TABLE public.reporting_composition_preview_output_files DROP CONSTRAINT reporting_composition_preview_output_files_attempt_fk;
DROP FUNCTION public.reporting_reject_imported_artifact_update();
DROP TABLE public.reporting_snapshots;
DROP TABLE public.reporting_render_bundles;
DROP TABLE public.reporting_render_bundle_files;
DROP TABLE public.reporting_releases;
DROP TABLE public.reporting_release_reveal_maps;
DROP TABLE public.reporting_release_approvals;
DROP TABLE public.reporting_preview_reveal_maps;
DROP TABLE public.reporting_job_payloads;
DROP TABLE public.reporting_imported_artifact_files;
DROP TABLE public.reporting_composition_preview_outputs;
DROP TABLE public.reporting_composition_preview_output_files;

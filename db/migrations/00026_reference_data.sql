-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: reference data.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.reference_pack_attempt_members (
    attempt_id uuid NOT NULL,
    ordinal bigint NOT NULL,
    pack_key text,
    pack_version text,
    verdict text NOT NULL,
    canonical_prepared bytea,
    failure_code text,
    check_id text,
    invalidated_content boolean DEFAULT false NOT NULL,
    canonical_validation_summary bytea,
    validation_summary_id text GENERATED ALWAYS AS (
CASE
    WHEN (canonical_validation_summary IS NULL) THEN NULL::text
    ELSE ('rpvs_'::text || encode(public.digest(canonical_validation_summary, 'sha256'::text), 'hex'::text))
END) STORED,
    CONSTRAINT rp_attempt_members_ck1 CHECK (ordinal > 0),
    CONSTRAINT rp_attempt_members_ck2 CHECK (verdict = ANY (ARRAY['succeeded'::text, 'content_rejected'::text])),
    CONSTRAINT rp_attempt_members_ck3 CHECK ((pack_key IS NULL) = (pack_version IS NULL)),
    CONSTRAINT rp_attempt_members_ck4 CHECK ((verdict = 'succeeded'::text) = (canonical_prepared IS NOT NULL)),
    CONSTRAINT rp_attempt_members_ck5 CHECK ((verdict = 'content_rejected'::text) = (failure_code IS NOT NULL AND check_id IS NOT NULL)),
    CONSTRAINT rp_attempt_members_ck6 CHECK (verdict <> 'succeeded'::text OR pack_key IS NOT NULL),
    CONSTRAINT rp_attempt_members_invalidation_ck CHECK (NOT invalidated_content OR verdict = 'content_rejected'::text AND pack_key IS NOT NULL),
    CONSTRAINT rp_attempt_members_pk1 PRIMARY KEY (attempt_id, ordinal),
    CONSTRAINT rp_attempt_members_summary_ck CHECK ((verdict = 'content_rejected'::text) = (canonical_validation_summary IS NOT NULL) AND (canonical_validation_summary IS NULL OR octet_length(canonical_validation_summary) >= 1 AND octet_length(canonical_validation_summary) <= 16777216))
);

REVOKE ALL ON TABLE public.reference_pack_attempt_members FROM PUBLIC;

GRANT UPDATE (invalidated_content) ON TABLE public.reference_pack_attempt_members TO cartulary_runtime;

REVOKE ALL ON TYPE public.reference_pack_attempt_members FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_attempt_members TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_attempt_members TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_attempt_members TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_attempt_members TO cartulary_recovery;

CREATE TABLE public.reference_pack_attempts (
    attempt_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    started_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone,
    outcome text,
    canonical_result bytea,
    CONSTRAINT rp_attempts_ck1 CHECK (outcome = ANY (ARRAY['succeeded'::text, 'content_rejected'::text, 'canceled'::text, 'timed_out'::text, 'stale_state'::text, 'execution_failed'::text, 'interrupted'::text])),
    CONSTRAINT rp_attempts_ck2 CHECK ((completed_at IS NULL) = (outcome IS NULL)),
    CONSTRAINT rp_attempts_ck3 CHECK ((outcome IS NULL) = (canonical_result IS NULL)),
    CONSTRAINT rp_attempts_pk1 PRIMARY KEY (attempt_id)
);

REVOKE ALL ON TABLE public.reference_pack_attempts FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_attempts FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_attempts TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_attempts TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_attempts TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_attempts TO cartulary_recovery;

CREATE TABLE public.reference_pack_candidates (
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    distribution_kind text NOT NULL,
    health text NOT NULL,
    last_failure_code text,
    administratively_disabled boolean DEFAULT false NOT NULL,
    removed boolean DEFAULT false NOT NULL,
    missing_reason text,
    current_envelope_id text,
    current_index_id uuid,
    admitted_at timestamp with time zone NOT NULL,
    admitted_by_user_id uuid,
    fallback_from_version text,
    CONSTRAINT rp_candidate_fallback_kind_ck CHECK (fallback_from_version IS NULL OR distribution_kind = 'packaged_builtin'::text AND fallback_from_version <> pack_version),
    CONSTRAINT rp_candidates_ck1 CHECK (distribution_kind = ANY (ARRAY['operator_imported'::text, 'packaged_builtin'::text])),
    CONSTRAINT rp_candidates_ck2 CHECK (health = ANY (ARRAY['staged'::text, 'verified_available'::text, 'failed'::text, 'missing'::text])),
    CONSTRAINT rp_candidates_ck3 CHECK (missing_reason = ANY (ARRAY['administrative_removal'::text, 'storage_loss'::text, 'staging_loss'::text])),
    CONSTRAINT rp_candidates_ck4 CHECK ((health = 'missing'::text) = (missing_reason IS NOT NULL)),
    CONSTRAINT rp_candidates_ck5 CHECK (NOT removed OR health = 'missing'::text AND missing_reason = 'administrative_removal'::text),
    CONSTRAINT rp_candidates_ck6 CHECK (distribution_kind <> 'packaged_builtin'::text OR NOT removed AND NOT administratively_disabled),
    CONSTRAINT rp_candidates_pk1 PRIMARY KEY (pack_key, pack_version)
);

REVOKE ALL ON TABLE public.reference_pack_candidates FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_candidates FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_candidates TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_candidates TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_candidates TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_candidates TO cartulary_recovery;

CREATE TABLE public.reference_pack_current_set (
    singleton boolean DEFAULT true NOT NULL,
    pack_set_id text,
    revision bigint NOT NULL,
    configuration_sha256 text NOT NULL,
    application_release_id text,
    profile_claimed boolean DEFAULT false NOT NULL,
    CONSTRAINT rp_current_set_ck1 CHECK (singleton),
    CONSTRAINT rp_current_set_ck2 CHECK (revision > 0),
    CONSTRAINT rp_current_set_ck3 CHECK (configuration_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_current_set_pk1 PRIMARY KEY (singleton)
);

REVOKE ALL ON TABLE public.reference_pack_current_set FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_current_set FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_current_set TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_current_set TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_current_set TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_current_set TO cartulary_recovery;

CREATE TABLE public.reference_pack_envelopes (
    envelope_id text NOT NULL,
    operation_id uuid NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    verified_at timestamp with time zone NOT NULL,
    trust_valid_until timestamp with time zone,
    container_sha256 text,
    container_ref text,
    repository_id text,
    root_version bigint,
    canonical_envelope bytea NOT NULL,
    CONSTRAINT rp_envelopes_ck1 CHECK (envelope_id ~ '^rpenv_[0-9a-f]{64}$'::text),
    CONSTRAINT rp_envelopes_ck2 CHECK (container_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_envelopes_ck3 CHECK ((container_sha256 IS NULL) = (container_ref IS NULL)),
    CONSTRAINT rp_envelopes_ck4 CHECK ((repository_id IS NULL) = (root_version IS NULL)),
    CONSTRAINT rp_envelopes_ck5 CHECK ((repository_id IS NULL) = (trust_valid_until IS NULL)),
    CONSTRAINT rp_envelopes_ck6 CHECK (trust_valid_until IS NULL OR verified_at < trust_valid_until),
    CONSTRAINT rp_envelopes_pk1 PRIMARY KEY (envelope_id),
    CONSTRAINT rp_envelopes_uq1 UNIQUE (envelope_id, pack_key, pack_version),
    CONSTRAINT rp_envelopes_uq2 UNIQUE (operation_id, pack_key, pack_version)
);

REVOKE ALL ON TABLE public.reference_pack_envelopes FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_envelopes FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_envelopes TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_envelopes TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_envelopes TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_envelopes TO cartulary_recovery;

CREATE TABLE public.reference_pack_events (
    attestation_id text NOT NULL,
    operation_id uuid NOT NULL,
    event_kind text NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    canonical_attestation bytea NOT NULL,
    CONSTRAINT rp_events_ck1 CHECK (attestation_id ~ '^rpa_[0-9a-f]{64}$'::text),
    CONSTRAINT rp_events_ck2 CHECK (event_kind = ANY (ARRAY['import_verification'::text, 'reverification'::text, 'refresh_verification'::text, 'activation'::text, 'rollback_activation'::text, 'safety_fallback'::text, 'profile_reconciliation'::text, 'dependency_invalidation'::text, 'disablement'::text, 'removal'::text, 'exact_reimport'::text, 'payload_invalidation'::text, 'trust_root_update'::text])),
    CONSTRAINT rp_events_pk1 PRIMARY KEY (operation_id, attestation_id),
    CONSTRAINT rp_events_uq1 UNIQUE (operation_id, event_kind, pack_key, pack_version)
);

REVOKE ALL ON TABLE public.reference_pack_events FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_events FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_events TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_events TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_events TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_events TO cartulary_recovery;

CREATE TABLE public.reference_pack_index_generations (
    index_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    entry_count bigint DEFAULT 0 NOT NULL,
    object_count bigint DEFAULT 0 NOT NULL,
    relationship_count bigint DEFAULT 0 NOT NULL,
    complete boolean DEFAULT false NOT NULL,
    CONSTRAINT rp_index_generations_ck1 CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_index_generations_ck2 CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_index_generations_ck3 CHECK (entry_count >= 0 AND entry_count <= 2000000),
    CONSTRAINT rp_index_generations_ck4 CHECK (object_count >= 0 AND object_count <= 2000000),
    CONSTRAINT rp_index_generations_ck5 CHECK (relationship_count >= 0 AND relationship_count <= 5000000),
    CONSTRAINT rp_index_generations_pk1 PRIMARY KEY (index_id),
    CONSTRAINT rp_index_generations_uq1 UNIQUE (index_id, pack_key, pack_version)
);

REVOKE ALL ON TABLE public.reference_pack_index_generations FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_index_generations FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_index_generations TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_index_generations TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_index_generations TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_index_generations TO cartulary_recovery;

CREATE TABLE public.reference_pack_indexes (
    index_id uuid NOT NULL,
    entry_kind text NOT NULL,
    entry_id text NOT NULL,
    canonical_item bytea NOT NULL,
    lookup_keys bytea NOT NULL,
    CONSTRAINT rp_indexes_ck1 CHECK (entry_kind = ANY (ARRAY['entry'::text, 'object'::text, 'relationship'::text])),
    CONSTRAINT rp_indexes_ck2 CHECK (octet_length(entry_id) >= 1 AND octet_length(entry_id) <= 512),
    CONSTRAINT rp_indexes_ck3 CHECK (octet_length(canonical_item) >= 1 AND octet_length(canonical_item) <= 1048576),
    CONSTRAINT rp_indexes_pk1 PRIMARY KEY (index_id, entry_kind, entry_id)
);

REVOKE ALL ON TABLE public.reference_pack_indexes FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_indexes FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_indexes TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_indexes TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_indexes TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_indexes TO cartulary_recovery;

CREATE TABLE public.reference_pack_key_state (
    pack_key text NOT NULL,
    revision bigint DEFAULT 1 NOT NULL,
    CONSTRAINT rp_key_state_ck1 CHECK (revision > 0),
    CONSTRAINT rp_key_state_pk1 PRIMARY KEY (pack_key)
);

REVOKE ALL ON TABLE public.reference_pack_key_state FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_key_state FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_key_state TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_key_state TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_key_state TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_key_state TO cartulary_recovery;

CREATE TABLE public.reference_pack_lookup_keys (
    index_id uuid NOT NULL,
    entry_kind text NOT NULL,
    entry_id text NOT NULL,
    lookup_kind text NOT NULL,
    lookup_value bytea NOT NULL,
    lookup_sha256 bytea GENERATED ALWAYS AS (public.digest(lookup_value, 'sha256'::text)) STORED NOT NULL,
    sort_key bytea NOT NULL,
    CONSTRAINT rp_lookup_keys_pk1 PRIMARY KEY (index_id, entry_kind, entry_id, lookup_kind, lookup_sha256)
);

REVOKE ALL ON TABLE public.reference_pack_lookup_keys FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_lookup_keys FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_lookup_keys TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_lookup_keys TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_lookup_keys TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_lookup_keys TO cartulary_recovery;

CREATE TABLE public.reference_pack_metadata_versions (
    repository_id text NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    role text NOT NULL,
    metadata_version bigint NOT NULL,
    canonical_bytes bytea NOT NULL,
    CONSTRAINT rp_metadata_versions_ck1 CHECK (role = ANY (ARRAY['timestamp'::text, 'snapshot'::text, 'targets'::text])),
    CONSTRAINT rp_metadata_versions_ck2 CHECK (metadata_version >= 1 AND metadata_version <= '9007199254740991'::bigint),
    CONSTRAINT rp_metadata_versions_ck3 CHECK (octet_length(canonical_bytes) >= 1 AND octet_length(canonical_bytes) <= 2097152),
    CONSTRAINT rp_metadata_versions_pk1 PRIMARY KEY (repository_id, pack_key, pack_version, role)
);

REVOKE ALL ON TABLE public.reference_pack_metadata_versions FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_metadata_versions FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_metadata_versions TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_metadata_versions TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_metadata_versions TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_metadata_versions TO cartulary_recovery;

CREATE TABLE public.reference_pack_object_refs (
    owner_kind text NOT NULL,
    owner_id text NOT NULL,
    logical_path text NOT NULL,
    object_id uuid NOT NULL,
    CONSTRAINT rp_object_refs_ck1 CHECK (owner_kind = ANY (ARRAY['version'::text, 'envelope'::text, 'operation'::text, 'backup'::text, 'release'::text])),
    CONSTRAINT rp_object_refs_pk1 PRIMARY KEY (owner_kind, owner_id, logical_path)
);

REVOKE ALL ON TABLE public.reference_pack_object_refs FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_object_refs FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_object_refs TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_object_refs TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_object_refs TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_object_refs TO cartulary_recovery;

CREATE TABLE public.reference_pack_objects (
    object_id uuid NOT NULL,
    sha256 text NOT NULL,
    storage_ref text NOT NULL,
    size_bytes bigint NOT NULL,
    generation bigint NOT NULL,
    available boolean DEFAULT true NOT NULL,
    CONSTRAINT rp_objects_ck1 CHECK (sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_objects_ck2 CHECK (size_bytes >= 0),
    CONSTRAINT rp_objects_ck3 CHECK (generation > 0),
    CONSTRAINT rp_objects_pk1 PRIMARY KEY (object_id),
    CONSTRAINT rp_objects_uq1 UNIQUE (storage_ref)
);

REVOKE ALL ON TABLE public.reference_pack_objects FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_objects FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_objects TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_objects TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_objects TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_objects TO cartulary_recovery;

CREATE TABLE public.reference_pack_operation_dependency_keys (
    operation_id uuid NOT NULL,
    pack_key text NOT NULL,
    admitted_revision bigint NOT NULL,
    CONSTRAINT rp_operation_dependency_keys_pk PRIMARY KEY (operation_id, pack_key),
    CONSTRAINT rp_operation_dependency_keys_revision_ck CHECK (admitted_revision > 0)
);

REVOKE ALL ON TABLE public.reference_pack_operation_dependency_keys FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_operation_dependency_keys FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_operation_dependency_keys TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_dependency_keys TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_operation_dependency_keys TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_operation_dependency_keys TO cartulary_recovery;

CREATE TABLE public.reference_pack_operation_keys (
    operation_id uuid NOT NULL,
    pack_key text NOT NULL,
    admitted_revision bigint NOT NULL,
    usage_revision bigint,
    CONSTRAINT rp_operation_keys_ck1 CHECK (admitted_revision > 0),
    CONSTRAINT rp_operation_keys_pk1 PRIMARY KEY (operation_id, pack_key)
);

REVOKE ALL ON TABLE public.reference_pack_operation_keys FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_operation_keys FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_operation_keys TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_keys TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_operation_keys TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_operation_keys TO cartulary_recovery;

CREATE TABLE public.reference_pack_operation_members (
    operation_id uuid NOT NULL,
    ordinal bigint NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    envelope_id text,
    CONSTRAINT rp_operation_members_ck1 CHECK (ordinal > 0),
    CONSTRAINT rp_operation_members_pk1 PRIMARY KEY (operation_id, ordinal),
    CONSTRAINT rp_operation_members_portable_binding_uq UNIQUE (operation_id, ordinal, pack_key, pack_version),
    CONSTRAINT rp_operation_members_uq1 UNIQUE (operation_id, pack_key, pack_version)
);

REVOKE ALL ON TABLE public.reference_pack_operation_members FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_operation_members FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_operation_members TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_members TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_operation_members TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_operation_members TO cartulary_recovery;

CREATE TABLE public.reference_pack_operation_repositories (
    operation_id uuid NOT NULL,
    repository_id text NOT NULL,
    admitted_revision bigint NOT NULL,
    root_version bigint NOT NULL,
    CONSTRAINT rp_operation_repositories_ck1 CHECK (admitted_revision > 0),
    CONSTRAINT rp_operation_repositories_pk1 PRIMARY KEY (operation_id, repository_id)
);

REVOKE ALL ON TABLE public.reference_pack_operation_repositories FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_operation_repositories FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_operation_repositories TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_repositories TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_operation_repositories TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_operation_repositories TO cartulary_recovery;

CREATE TABLE public.reference_pack_operations (
    operation_id uuid NOT NULL,
    job_id uuid,
    kind text NOT NULL,
    actor_kind text NOT NULL,
    actor_user_id uuid,
    admitted_at timestamp with time zone NOT NULL,
    terminal_at timestamp with time zone,
    frozen_input bytea NOT NULL,
    final_outcome bytea,
    CONSTRAINT rp_operations_ck1 CHECK (kind = ANY (ARRAY['import'::text, 'renewal'::text, 'reverify'::text, 'refresh'::text, 'activate'::text, 'disable'::text, 'remove'::text, 'integrity'::text, 'reconcile'::text, 'portable_retention'::text])),
    CONSTRAINT rp_operations_ck2 CHECK (actor_kind = ANY (ARRAY['user'::text, 'local_operator'::text, 'system'::text, 'application_release'::text])),
    CONSTRAINT rp_operations_ck3 CHECK ((terminal_at IS NULL) = (final_outcome IS NULL)),
    CONSTRAINT rp_operations_ck4 CHECK ((actor_kind = 'user'::text) = (actor_user_id IS NOT NULL)),
    CONSTRAINT rp_operations_pk1 PRIMARY KEY (operation_id),
    CONSTRAINT rp_operations_uq1 UNIQUE (job_id),
    CONSTRAINT rp_portable_content_input_ck CHECK (kind <> 'portable_retention'::text OR COALESCE(((convert_from(frozen_input, 'UTF8'::name)::jsonb -> 'content_manifest'::text) ->> 'schema_id'::text) = 'reference_pack_content.v1'::text AND jsonb_typeof((convert_from(frozen_input, 'UTF8'::name)::jsonb -> 'content_manifest'::text) -> 'containers'::text) = 'array'::text AND jsonb_typeof((convert_from(frozen_input, 'UTF8'::name)::jsonb -> 'content_manifest'::text) -> 'required_members'::text) = 'array'::text, false))
);

REVOKE ALL ON TABLE public.reference_pack_operations FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_operations FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_operations TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operations TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_operations TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_operations TO cartulary_recovery;

CREATE TABLE public.reference_pack_pins (
    owner_kind text NOT NULL,
    owner_id text NOT NULL,
    pack_set_id text NOT NULL,
    operation_id uuid NOT NULL,
    CONSTRAINT rp_pins_pk1 PRIMARY KEY (owner_kind, owner_id, pack_set_id)
);

REVOKE ALL ON TABLE public.reference_pack_pins FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_pins FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_pins TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_pins TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_pins TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_pins TO cartulary_recovery;

CREATE TABLE public.reference_pack_portable_catalogs (
    incident_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    catalog_sha256 text NOT NULL,
    canonical_references bytea NOT NULL,
    canonical_resolution bytea NOT NULL,
    CONSTRAINT rp_portable_catalogs_bytes_ck CHECK (encode(public.digest(canonical_references, 'sha256'::text), 'hex'::text) = catalog_sha256),
    CONSTRAINT rp_portable_catalogs_digest_ck CHECK (catalog_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_portable_catalogs_operation_uq UNIQUE (operation_id),
    CONSTRAINT rp_portable_catalogs_pk PRIMARY KEY (incident_id)
);

REVOKE ALL ON TABLE public.reference_pack_portable_catalogs FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_portable_catalogs FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_portable_catalogs TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_portable_catalogs TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_portable_catalogs TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_portable_catalogs TO cartulary_recovery;

CREATE TABLE public.reference_pack_portable_preparations (
    operation_id uuid NOT NULL,
    canonical_references bytea NOT NULL,
    canonical_context bytea NOT NULL,
    CONSTRAINT rp_portable_preparations_context_ck CHECK (octet_length(canonical_context) >= 1 AND octet_length(canonical_context) <= 4096),
    CONSTRAINT rp_portable_preparations_pk PRIMARY KEY (operation_id),
    CONSTRAINT rp_portable_preparations_references_ck CHECK (octet_length(canonical_references) >= 1 AND octet_length(canonical_references) <= 67108864)
);

REVOKE ALL ON TABLE public.reference_pack_portable_preparations FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_portable_preparations FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_portable_preparations TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_portable_preparations TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_portable_preparations TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_portable_preparations TO cartulary_recovery;

CREATE TABLE public.reference_pack_portable_selections (
    operation_id uuid NOT NULL,
    ordinal integer NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    envelope_id text,
    available boolean NOT NULL,
    reason_code text,
    verification_ordinal bigint,
    input_object_id uuid,
    expected_container_sha256 text,
    expected_container_bytes bigint,
    CONSTRAINT rp_portable_selections_availability_ck CHECK (available = (reason_code IS NULL) AND (NOT available OR envelope_id IS NOT NULL)),
    CONSTRAINT rp_portable_selections_digest_ck CHECK (expected_container_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_portable_selections_input_ck CHECK (verification_ordinal IS NULL AND input_object_id IS NULL AND expected_container_sha256 IS NULL AND expected_container_bytes IS NULL OR verification_ordinal IS NOT NULL AND input_object_id IS NOT NULL AND expected_container_sha256 IS NOT NULL AND expected_container_bytes IS NOT NULL AND NOT available),
    CONSTRAINT rp_portable_selections_ordinal_ck CHECK (ordinal >= 1 AND ordinal <= 16384),
    CONSTRAINT rp_portable_selections_pk PRIMARY KEY (operation_id, ordinal),
    CONSTRAINT rp_portable_selections_reason_ck CHECK (reason_code = ANY (ARRAY['not_retained'::text, 'removed'::text, 'not_usable'::text, 'content_unavailable'::text])),
    CONSTRAINT rp_portable_selections_size_ck CHECK (expected_container_bytes >= 1 AND expected_container_bytes <= '9007199254740991'::bigint),
    CONSTRAINT rp_portable_selections_verification_uq UNIQUE (operation_id, verification_ordinal),
    CONSTRAINT rp_portable_selections_version_uq UNIQUE (operation_id, pack_key, pack_version)
);

REVOKE ALL ON TABLE public.reference_pack_portable_selections FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_portable_selections FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_portable_selections TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_portable_selections TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_portable_selections TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_portable_selections TO cartulary_recovery;

CREATE TABLE public.reference_pack_registry_usage (
    pack_key text NOT NULL,
    revision bigint NOT NULL,
    CONSTRAINT rp_registry_usage_ck1 CHECK (revision > 0),
    CONSTRAINT rp_registry_usage_pk1 PRIMARY KEY (pack_key)
);

REVOKE ALL ON TABLE public.reference_pack_registry_usage FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_registry_usage FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_registry_usage TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_registry_usage TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_registry_usage TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_registry_usage TO cartulary_recovery;

CREATE TABLE public.reference_pack_release_bindings (
    application_release_id text NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    canonical_binding bytea NOT NULL,
    CONSTRAINT rp_release_bindings_pk1 PRIMARY KEY (application_release_id, pack_key)
);

REVOKE ALL ON TABLE public.reference_pack_release_bindings FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_release_bindings FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_release_bindings TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_release_bindings TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_release_bindings TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_release_bindings TO cartulary_recovery;

CREATE TABLE public.reference_pack_repositories (
    repository_id text NOT NULL,
    revision bigint NOT NULL,
    root_version bigint NOT NULL,
    CONSTRAINT rp_repositories_ck1 CHECK (revision > 0),
    CONSTRAINT rp_repositories_ck2 CHECK (root_version >= 1 AND root_version <= '9007199254740991'::bigint),
    CONSTRAINT rp_repositories_pk1 PRIMARY KEY (repository_id)
);

REVOKE ALL ON TABLE public.reference_pack_repositories FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_repositories FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_repositories TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_repositories TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_repositories TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_repositories TO cartulary_recovery;

CREATE TABLE public.reference_pack_roots (
    repository_id text NOT NULL,
    root_version bigint NOT NULL,
    canonical_bytes bytea NOT NULL,
    sha256 text NOT NULL,
    predecessor_version bigint,
    transition_evidence jsonb NOT NULL,
    CONSTRAINT rp_roots_ck1 CHECK (root_version >= 1 AND root_version <= '9007199254740991'::bigint),
    CONSTRAINT rp_roots_ck2 CHECK (octet_length(canonical_bytes) >= 1 AND octet_length(canonical_bytes) <= 2097152),
    CONSTRAINT rp_roots_ck3 CHECK (sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_roots_ck4 CHECK (jsonb_typeof(transition_evidence) = 'object'::text),
    CONSTRAINT rp_roots_ck5 CHECK (predecessor_version IS NULL OR predecessor_version = (root_version - 1)),
    CONSTRAINT rp_roots_pk1 PRIMARY KEY (repository_id, root_version)
);

REVOKE ALL ON TABLE public.reference_pack_roots FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_roots FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_roots TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_roots TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_roots TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_roots TO cartulary_recovery;

CREATE TABLE public.reference_pack_set_members (
    pack_set_id text NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    provenance_envelope_id text NOT NULL,
    CONSTRAINT rp_set_members_pk1 PRIMARY KEY (pack_set_id, pack_key)
);

REVOKE ALL ON TABLE public.reference_pack_set_members FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_set_members FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_set_members TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_set_members TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_set_members TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_set_members TO cartulary_recovery;

CREATE TABLE public.reference_pack_sets (
    pack_set_id text NOT NULL,
    pack_set_sha256 text NOT NULL,
    canonical_set bytea NOT NULL,
    canonical_provenance bytea NOT NULL,
    first_operation_id uuid NOT NULL,
    CONSTRAINT rp_sets_ck1 CHECK (pack_set_id ~ '^rpset_[0-9a-f]{64}$'::text),
    CONSTRAINT rp_sets_ck2 CHECK (pack_set_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_sets_ck3 CHECK (pack_set_id = ('rpset_'::text || pack_set_sha256)),
    CONSTRAINT rp_sets_pk1 PRIMARY KEY (pack_set_id),
    CONSTRAINT rp_sets_uq1 UNIQUE (pack_set_sha256)
);

REVOKE ALL ON TABLE public.reference_pack_sets FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_sets FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_sets TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_sets TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_sets TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_sets TO cartulary_recovery;

CREATE TABLE public.reference_pack_version_pins (
    owner_kind text NOT NULL,
    owner_id text NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    envelope_id text NOT NULL,
    operation_id uuid NOT NULL,
    CONSTRAINT rp_version_pins_pk PRIMARY KEY (owner_kind, owner_id, pack_key, pack_version)
);

REVOKE ALL ON TABLE public.reference_pack_version_pins FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_version_pins FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.reference_pack_version_pins TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_version_pins TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_version_pins TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_version_pins TO cartulary_recovery;

CREATE TABLE public.reference_pack_versions (
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    repository_id text,
    pack_release_sequence bigint NOT NULL,
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    manifest_bytes bytea NOT NULL,
    CONSTRAINT rp_versions_ck1 CHECK (pack_release_sequence >= 1 AND pack_release_sequence <= '9007199254740991'::bigint),
    CONSTRAINT rp_versions_ck2 CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_versions_ck3 CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT rp_versions_ck4 CHECK (octet_length(manifest_bytes) >= 1 AND octet_length(manifest_bytes) <= 1048576),
    CONSTRAINT rp_versions_pk1 PRIMARY KEY (pack_key, pack_version),
    CONSTRAINT rp_versions_uq1 UNIQUE NULLS NOT DISTINCT (repository_id, pack_key, pack_release_sequence),
    CONSTRAINT rp_versions_uq2 UNIQUE (pack_key, pack_version, manifest_sha256, payload_sha256)
);

REVOKE ALL ON TABLE public.reference_pack_versions FROM PUBLIC;

REVOKE ALL ON TYPE public.reference_pack_versions FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.reference_pack_versions TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_versions TO cartulary_recovery;

GRANT USAGE ON TYPE public.reference_pack_versions TO cartulary_runtime;

GRANT USAGE ON TYPE public.reference_pack_versions TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_attempt_guard()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF TG_OP = 'DELETE' OR OLD.completed_at IS NOT NULL THEN
        RAISE EXCEPTION 'immutable reference pack attempt' USING ERRCODE = '55000';
    END IF;
    IF ROW(NEW.attempt_id, NEW.operation_id, NEW.started_at) IS DISTINCT FROM ROW(OLD.attempt_id, OLD.operation_id, OLD.started_at) THEN
        RAISE EXCEPTION 'immutable reference pack execution identity' USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.reference_pack_attempt_guard() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_attempt_member_guard()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
DECLARE terminal timestamptz;
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION 'immutable reference pack preparation' USING ERRCODE='55000';
    END IF;
    IF OLD.invalidated_content OR NOT NEW.invalidated_content
       OR (to_jsonb(NEW)-'invalidated_content'-'validation_summary_id') IS DISTINCT FROM (to_jsonb(OLD)-'invalidated_content'-'validation_summary_id') THEN
        RAISE EXCEPTION 'immutable reference pack preparation' USING ERRCODE='55000';
    END IF;
    SELECT completed_at INTO STRICT terminal FROM public.reference_pack_attempts WHERE attempt_id=OLD.attempt_id FOR UPDATE;
    IF terminal IS NOT NULL THEN
        RAISE EXCEPTION 'immutable reference pack terminal effect' USING ERRCODE='55000';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.reference_pack_attempt_member_guard() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_immutable_row()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    RAISE EXCEPTION 'immutable reference pack history' USING ERRCODE = '55000';
END $$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.reference_pack_immutable_row() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_operation_guard()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF TG_OP = 'DELETE' OR OLD.terminal_at IS NOT NULL THEN
        RAISE EXCEPTION 'immutable reference pack operation' USING ERRCODE = '55000';
    END IF;
    IF ROW(NEW.operation_id, NEW.job_id, NEW.kind, NEW.actor_kind, NEW.actor_user_id, NEW.admitted_at, NEW.frozen_input)
       IS DISTINCT FROM ROW(OLD.operation_id, OLD.job_id, OLD.kind, OLD.actor_kind, OLD.actor_user_id, OLD.admitted_at, OLD.frozen_input) THEN
        RAISE EXCEPTION 'immutable reference pack admission' USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.reference_pack_operation_guard() FROM PUBLIC;

CREATE INDEX reference_pack_attempt_members_summary_idx ON public.reference_pack_attempt_members USING btree (validation_summary_id);

CREATE INDEX reference_pack_attempts_operation_idx ON public.reference_pack_attempts USING btree (operation_id, started_at);

CREATE INDEX reference_pack_lookup_query_idx ON public.reference_pack_lookup_keys USING btree (index_id, lookup_kind, lookup_sha256);

CREATE INDEX reference_pack_object_refs_object_idx ON public.reference_pack_object_refs USING btree (object_id);

CREATE INDEX reference_pack_objects_digest_idx ON public.reference_pack_objects USING btree (sha256);

CREATE INDEX reference_pack_operation_dependency_keys_key_idx ON public.reference_pack_operation_dependency_keys USING btree (pack_key);

CREATE INDEX reference_pack_operation_keys_pack_idx ON public.reference_pack_operation_keys USING btree (pack_key);

CREATE INDEX reference_pack_pins_set_idx ON public.reference_pack_pins USING btree (pack_set_id);

CREATE INDEX rp_candidate_fallback_version_idx ON public.reference_pack_candidates USING btree (pack_key, fallback_from_version);

CREATE INDEX rp_candidates_actor_idx ON public.reference_pack_candidates USING btree (admitted_by_user_id);

CREATE INDEX rp_candidates_envelope_idx ON public.reference_pack_candidates USING btree (current_envelope_id, pack_key, pack_version);

CREATE INDEX rp_candidates_index_idx ON public.reference_pack_candidates USING btree (current_index_id, pack_key, pack_version);

CREATE INDEX rp_current_set_set_idx ON public.reference_pack_current_set USING btree (pack_set_id);

CREATE INDEX rp_envelopes_root_idx ON public.reference_pack_envelopes USING btree (repository_id, root_version);

CREATE INDEX rp_envelopes_version_idx ON public.reference_pack_envelopes USING btree (pack_key, pack_version);

CREATE INDEX rp_events_version_idx ON public.reference_pack_events USING btree (pack_key, pack_version);

CREATE INDEX rp_index_generations_operation_idx ON public.reference_pack_index_generations USING btree (operation_id);

CREATE INDEX rp_metadata_versions_version_idx ON public.reference_pack_metadata_versions USING btree (pack_key, pack_version);

CREATE INDEX rp_operation_members_envelope_idx ON public.reference_pack_operation_members USING btree (envelope_id, pack_key, pack_version);

CREATE INDEX rp_operation_members_version_idx ON public.reference_pack_operation_members USING btree (pack_key, pack_version);

CREATE INDEX rp_operation_repositories_root_idx ON public.reference_pack_operation_repositories USING btree (repository_id, root_version);

CREATE INDEX rp_operations_actor_idx ON public.reference_pack_operations USING btree (actor_user_id);

CREATE INDEX rp_portable_selections_envelope_idx ON public.reference_pack_portable_selections USING btree (envelope_id, pack_key, pack_version);

CREATE INDEX rp_portable_selections_member_idx ON public.reference_pack_portable_selections USING btree (operation_id, verification_ordinal, pack_key, pack_version);

CREATE INDEX rp_portable_selections_object_idx ON public.reference_pack_portable_selections USING btree (input_object_id);

CREATE INDEX rp_release_bindings_version_idx ON public.reference_pack_release_bindings USING btree (pack_key, pack_version);

CREATE INDEX rp_repositories_root_idx ON public.reference_pack_repositories USING btree (repository_id, root_version);

CREATE INDEX rp_roots_predecessor_idx ON public.reference_pack_roots USING btree (repository_id, predecessor_version);

CREATE INDEX rp_set_members_envelope_idx ON public.reference_pack_set_members USING btree (provenance_envelope_id, pack_key, pack_version);

CREATE INDEX rp_set_members_version_idx ON public.reference_pack_set_members USING btree (pack_key, pack_version, manifest_sha256, payload_sha256);

CREATE INDEX rp_sets_operation_idx ON public.reference_pack_sets USING btree (first_operation_id);

CREATE INDEX rp_version_pins_envelope_idx ON public.reference_pack_version_pins USING btree (envelope_id, pack_key, pack_version);

CREATE INDEX rp_version_pins_operation_idx ON public.reference_pack_version_pins USING btree (operation_id);

CREATE INDEX rp_version_pins_version_idx ON public.reference_pack_version_pins USING btree (pack_key, pack_version);

ALTER TABLE public.reference_pack_attempt_members ADD CONSTRAINT rp_attempt_members_attempt_id_fk FOREIGN KEY (attempt_id) REFERENCES public.reference_pack_attempts(attempt_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_attempts ADD CONSTRAINT rp_attempts_operation_id_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT reference_pack_candidate_envelope_fk FOREIGN KEY (current_envelope_id, pack_key, pack_version) REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT rp_candidate_fallback_version_fk FOREIGN KEY (pack_key, fallback_from_version) REFERENCES public.reference_pack_candidates(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT rp_candidates_admitted_by_user_id_fk FOREIGN KEY (admitted_by_user_id) REFERENCES public.users(id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT rp_candidates_index_fk FOREIGN KEY (current_index_id, pack_key, pack_version) REFERENCES public.reference_pack_index_generations(index_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT rp_candidates_pack_key_fk FOREIGN KEY (pack_key) REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_current_set ADD CONSTRAINT rp_current_set_pack_set_id_fk FOREIGN KEY (pack_set_id) REFERENCES public.reference_pack_sets(pack_set_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_envelopes ADD CONSTRAINT rp_envelopes_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_versions(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_envelopes ADD CONSTRAINT rp_envelopes_fk2 FOREIGN KEY (repository_id, root_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_envelopes ADD CONSTRAINT rp_envelopes_operation_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_envelopes ADD CONSTRAINT rp_envelopes_repository_id_fk FOREIGN KEY (repository_id) REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_events ADD CONSTRAINT rp_events_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_candidates(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_events ADD CONSTRAINT rp_events_operation_id_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_index_generations ADD CONSTRAINT rp_index_generations_operation_id_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_indexes ADD CONSTRAINT rp_indexes_index_id_fk FOREIGN KEY (index_id) REFERENCES public.reference_pack_index_generations(index_id) ON UPDATE RESTRICT ON DELETE CASCADE;

ALTER TABLE public.reference_pack_lookup_keys ADD CONSTRAINT rp_lookup_keys_fk1 FOREIGN KEY (index_id, entry_kind, entry_id) REFERENCES public.reference_pack_indexes(index_id, entry_kind, entry_id) ON UPDATE RESTRICT ON DELETE CASCADE;

ALTER TABLE public.reference_pack_metadata_versions ADD CONSTRAINT rp_metadata_versions_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_versions(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_metadata_versions ADD CONSTRAINT rp_metadata_versions_repository_id_fk FOREIGN KEY (repository_id) REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_object_refs ADD CONSTRAINT rp_object_refs_object_id_fk FOREIGN KEY (object_id) REFERENCES public.reference_pack_objects(object_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_dependency_keys ADD CONSTRAINT rp_operation_dependency_keys_key_fk FOREIGN KEY (pack_key) REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_dependency_keys ADD CONSTRAINT rp_operation_dependency_keys_operation_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_keys ADD CONSTRAINT rp_operation_keys_operation_id_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_keys ADD CONSTRAINT rp_operation_keys_pack_key_fk FOREIGN KEY (pack_key) REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_members ADD CONSTRAINT rp_operation_members_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_candidates(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_members ADD CONSTRAINT rp_operation_members_fk2 FOREIGN KEY (envelope_id, pack_key, pack_version) REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_members ADD CONSTRAINT rp_operation_members_operation_id_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_repositories ADD CONSTRAINT rp_operation_repositories_fk1 FOREIGN KEY (repository_id, root_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operation_repositories ADD CONSTRAINT rp_operation_repositories_operation_id_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operations ADD CONSTRAINT rp_operations_actor_user_id_fk FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_operations ADD CONSTRAINT rp_operations_job_id_fk FOREIGN KEY (job_id) REFERENCES public.jobs(job_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_pins ADD CONSTRAINT rp_pins_pack_set_id_fk FOREIGN KEY (pack_set_id) REFERENCES public.reference_pack_sets(pack_set_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_portable_catalogs ADD CONSTRAINT rp_portable_catalogs_incident_fk FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_portable_catalogs ADD CONSTRAINT rp_portable_catalogs_operation_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_portable_preparations ADD CONSTRAINT rp_portable_preparations_operation_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_portable_selections ADD CONSTRAINT rp_portable_selections_envelope_fk FOREIGN KEY (envelope_id, pack_key, pack_version) REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_portable_selections ADD CONSTRAINT rp_portable_selections_member_fk FOREIGN KEY (operation_id, verification_ordinal, pack_key, pack_version) REFERENCES public.reference_pack_operation_members(operation_id, ordinal, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_portable_selections ADD CONSTRAINT rp_portable_selections_object_fk FOREIGN KEY (input_object_id) REFERENCES public.reference_pack_objects(object_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_portable_selections ADD CONSTRAINT rp_portable_selections_operation_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_portable_preparations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_registry_usage ADD CONSTRAINT rp_registry_usage_pack_key_fk FOREIGN KEY (pack_key) REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_release_bindings ADD CONSTRAINT rp_release_bindings_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_versions(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_repositories ADD CONSTRAINT reference_pack_current_root_fk FOREIGN KEY (repository_id, root_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE public.reference_pack_roots ADD CONSTRAINT rp_roots_fk1 FOREIGN KEY (repository_id, predecessor_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_roots ADD CONSTRAINT rp_roots_repository_id_fk FOREIGN KEY (repository_id) REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_set_members ADD CONSTRAINT rp_set_members_fk1 FOREIGN KEY (pack_key, pack_version, manifest_sha256, payload_sha256) REFERENCES public.reference_pack_versions(pack_key, pack_version, manifest_sha256, payload_sha256) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_set_members ADD CONSTRAINT rp_set_members_fk2 FOREIGN KEY (provenance_envelope_id, pack_key, pack_version) REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_set_members ADD CONSTRAINT rp_set_members_pack_set_id_fk FOREIGN KEY (pack_set_id) REFERENCES public.reference_pack_sets(pack_set_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_sets ADD CONSTRAINT rp_sets_first_operation_id_fk FOREIGN KEY (first_operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_version_pins ADD CONSTRAINT rp_version_pins_envelope_fk FOREIGN KEY (envelope_id, pack_key, pack_version) REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_version_pins ADD CONSTRAINT rp_version_pins_operation_fk FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_versions ADD CONSTRAINT rp_versions_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_candidates(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

ALTER TABLE public.reference_pack_versions ADD CONSTRAINT rp_versions_repository_id_fk FOREIGN KEY (repository_id) REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT;

CREATE TRIGGER reference_pack_attempt_members_guard BEFORE DELETE OR UPDATE ON public.reference_pack_attempt_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_attempt_member_guard();

CREATE TRIGGER reference_pack_attempts_guard BEFORE DELETE OR UPDATE ON public.reference_pack_attempts FOR EACH ROW EXECUTE FUNCTION public.reference_pack_attempt_guard();

CREATE TRIGGER reference_pack_envelopes_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_envelopes FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_events_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_events FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_operation_dependency_keys_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_operation_dependency_keys FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_operation_keys_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_operation_keys FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_operation_members_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_operation_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_operation_repositories_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_operation_repositories FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_operations_guard BEFORE DELETE OR UPDATE ON public.reference_pack_operations FOR EACH ROW EXECUTE FUNCTION public.reference_pack_operation_guard();

CREATE TRIGGER reference_pack_portable_catalogs_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_portable_catalogs FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_portable_preparations_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_portable_preparations FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_portable_selections_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_portable_selections FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_release_bindings_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_release_bindings FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_roots_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_roots FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_set_members_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_set_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_sets_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_sets FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_versions_immutable BEFORE DELETE OR UPDATE ON public.reference_pack_versions FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER reference_pack_versions_immutable ON public.reference_pack_versions;
DROP TRIGGER reference_pack_sets_immutable ON public.reference_pack_sets;
DROP TRIGGER reference_pack_set_members_immutable ON public.reference_pack_set_members;
DROP TRIGGER reference_pack_roots_immutable ON public.reference_pack_roots;
DROP TRIGGER reference_pack_release_bindings_immutable ON public.reference_pack_release_bindings;
DROP TRIGGER reference_pack_portable_selections_immutable ON public.reference_pack_portable_selections;
DROP TRIGGER reference_pack_portable_preparations_immutable ON public.reference_pack_portable_preparations;
DROP TRIGGER reference_pack_portable_catalogs_immutable ON public.reference_pack_portable_catalogs;
DROP TRIGGER reference_pack_operations_guard ON public.reference_pack_operations;
DROP TRIGGER reference_pack_operation_repositories_immutable ON public.reference_pack_operation_repositories;
DROP TRIGGER reference_pack_operation_members_immutable ON public.reference_pack_operation_members;
DROP TRIGGER reference_pack_operation_keys_immutable ON public.reference_pack_operation_keys;
DROP TRIGGER reference_pack_operation_dependency_keys_immutable ON public.reference_pack_operation_dependency_keys;
DROP TRIGGER reference_pack_events_immutable ON public.reference_pack_events;
DROP TRIGGER reference_pack_envelopes_immutable ON public.reference_pack_envelopes;
DROP TRIGGER reference_pack_attempts_guard ON public.reference_pack_attempts;
DROP TRIGGER reference_pack_attempt_members_guard ON public.reference_pack_attempt_members;
ALTER TABLE public.reference_pack_versions DROP CONSTRAINT rp_versions_repository_id_fk;
ALTER TABLE public.reference_pack_versions DROP CONSTRAINT rp_versions_fk1;
ALTER TABLE public.reference_pack_version_pins DROP CONSTRAINT rp_version_pins_operation_fk;
ALTER TABLE public.reference_pack_version_pins DROP CONSTRAINT rp_version_pins_envelope_fk;
ALTER TABLE public.reference_pack_sets DROP CONSTRAINT rp_sets_first_operation_id_fk;
ALTER TABLE public.reference_pack_set_members DROP CONSTRAINT rp_set_members_pack_set_id_fk;
ALTER TABLE public.reference_pack_set_members DROP CONSTRAINT rp_set_members_fk2;
ALTER TABLE public.reference_pack_set_members DROP CONSTRAINT rp_set_members_fk1;
ALTER TABLE public.reference_pack_roots DROP CONSTRAINT rp_roots_repository_id_fk;
ALTER TABLE public.reference_pack_roots DROP CONSTRAINT rp_roots_fk1;
ALTER TABLE public.reference_pack_repositories DROP CONSTRAINT reference_pack_current_root_fk;
ALTER TABLE public.reference_pack_release_bindings DROP CONSTRAINT rp_release_bindings_fk1;
ALTER TABLE public.reference_pack_registry_usage DROP CONSTRAINT rp_registry_usage_pack_key_fk;
ALTER TABLE public.reference_pack_portable_selections DROP CONSTRAINT rp_portable_selections_operation_fk;
ALTER TABLE public.reference_pack_portable_selections DROP CONSTRAINT rp_portable_selections_object_fk;
ALTER TABLE public.reference_pack_portable_selections DROP CONSTRAINT rp_portable_selections_member_fk;
ALTER TABLE public.reference_pack_portable_selections DROP CONSTRAINT rp_portable_selections_envelope_fk;
ALTER TABLE public.reference_pack_portable_preparations DROP CONSTRAINT rp_portable_preparations_operation_fk;
ALTER TABLE public.reference_pack_portable_catalogs DROP CONSTRAINT rp_portable_catalogs_operation_fk;
ALTER TABLE public.reference_pack_portable_catalogs DROP CONSTRAINT rp_portable_catalogs_incident_fk;
ALTER TABLE public.reference_pack_pins DROP CONSTRAINT rp_pins_pack_set_id_fk;
ALTER TABLE public.reference_pack_operations DROP CONSTRAINT rp_operations_job_id_fk;
ALTER TABLE public.reference_pack_operations DROP CONSTRAINT rp_operations_actor_user_id_fk;
ALTER TABLE public.reference_pack_operation_repositories DROP CONSTRAINT rp_operation_repositories_operation_id_fk;
ALTER TABLE public.reference_pack_operation_repositories DROP CONSTRAINT rp_operation_repositories_fk1;
ALTER TABLE public.reference_pack_operation_members DROP CONSTRAINT rp_operation_members_operation_id_fk;
ALTER TABLE public.reference_pack_operation_members DROP CONSTRAINT rp_operation_members_fk2;
ALTER TABLE public.reference_pack_operation_members DROP CONSTRAINT rp_operation_members_fk1;
ALTER TABLE public.reference_pack_operation_keys DROP CONSTRAINT rp_operation_keys_pack_key_fk;
ALTER TABLE public.reference_pack_operation_keys DROP CONSTRAINT rp_operation_keys_operation_id_fk;
ALTER TABLE public.reference_pack_operation_dependency_keys DROP CONSTRAINT rp_operation_dependency_keys_operation_fk;
ALTER TABLE public.reference_pack_operation_dependency_keys DROP CONSTRAINT rp_operation_dependency_keys_key_fk;
ALTER TABLE public.reference_pack_object_refs DROP CONSTRAINT rp_object_refs_object_id_fk;
ALTER TABLE public.reference_pack_metadata_versions DROP CONSTRAINT rp_metadata_versions_repository_id_fk;
ALTER TABLE public.reference_pack_metadata_versions DROP CONSTRAINT rp_metadata_versions_fk1;
ALTER TABLE public.reference_pack_lookup_keys DROP CONSTRAINT rp_lookup_keys_fk1;
ALTER TABLE public.reference_pack_indexes DROP CONSTRAINT rp_indexes_index_id_fk;
ALTER TABLE public.reference_pack_index_generations DROP CONSTRAINT rp_index_generations_operation_id_fk;
ALTER TABLE public.reference_pack_events DROP CONSTRAINT rp_events_operation_id_fk;
ALTER TABLE public.reference_pack_events DROP CONSTRAINT rp_events_fk1;
ALTER TABLE public.reference_pack_envelopes DROP CONSTRAINT rp_envelopes_repository_id_fk;
ALTER TABLE public.reference_pack_envelopes DROP CONSTRAINT rp_envelopes_operation_fk;
ALTER TABLE public.reference_pack_envelopes DROP CONSTRAINT rp_envelopes_fk2;
ALTER TABLE public.reference_pack_envelopes DROP CONSTRAINT rp_envelopes_fk1;
ALTER TABLE public.reference_pack_current_set DROP CONSTRAINT rp_current_set_pack_set_id_fk;
ALTER TABLE public.reference_pack_candidates DROP CONSTRAINT rp_candidates_pack_key_fk;
ALTER TABLE public.reference_pack_candidates DROP CONSTRAINT rp_candidates_index_fk;
ALTER TABLE public.reference_pack_candidates DROP CONSTRAINT rp_candidates_admitted_by_user_id_fk;
ALTER TABLE public.reference_pack_candidates DROP CONSTRAINT rp_candidate_fallback_version_fk;
ALTER TABLE public.reference_pack_candidates DROP CONSTRAINT reference_pack_candidate_envelope_fk;
ALTER TABLE public.reference_pack_attempts DROP CONSTRAINT rp_attempts_operation_id_fk;
ALTER TABLE public.reference_pack_attempt_members DROP CONSTRAINT rp_attempt_members_attempt_id_fk;
DROP FUNCTION public.reference_pack_operation_guard();
DROP FUNCTION public.reference_pack_immutable_row();
DROP FUNCTION public.reference_pack_attempt_member_guard();
DROP FUNCTION public.reference_pack_attempt_guard();
DROP TABLE public.reference_pack_versions;
DROP TABLE public.reference_pack_version_pins;
DROP TABLE public.reference_pack_sets;
DROP TABLE public.reference_pack_set_members;
DROP TABLE public.reference_pack_roots;
DROP TABLE public.reference_pack_repositories;
DROP TABLE public.reference_pack_release_bindings;
DROP TABLE public.reference_pack_registry_usage;
DROP TABLE public.reference_pack_portable_selections;
DROP TABLE public.reference_pack_portable_preparations;
DROP TABLE public.reference_pack_portable_catalogs;
DROP TABLE public.reference_pack_pins;
DROP TABLE public.reference_pack_operations;
DROP TABLE public.reference_pack_operation_repositories;
DROP TABLE public.reference_pack_operation_members;
DROP TABLE public.reference_pack_operation_keys;
DROP TABLE public.reference_pack_operation_dependency_keys;
DROP TABLE public.reference_pack_objects;
DROP TABLE public.reference_pack_object_refs;
DROP TABLE public.reference_pack_metadata_versions;
DROP TABLE public.reference_pack_lookup_keys;
DROP TABLE public.reference_pack_key_state;
DROP TABLE public.reference_pack_indexes;
DROP TABLE public.reference_pack_index_generations;
DROP TABLE public.reference_pack_events;
DROP TABLE public.reference_pack_envelopes;
DROP TABLE public.reference_pack_current_set;
DROP TABLE public.reference_pack_candidates;
DROP TABLE public.reference_pack_attempts;
DROP TABLE public.reference_pack_attempt_members;

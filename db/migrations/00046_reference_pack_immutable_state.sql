-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    packs bigint;
    pack_jobs bigint;
    retained_pack_jobs bigint;
    identities bigint;
    observations bigint;
    snapshots bigint;
    reporting_jobs bigint;
    portability_jobs bigint;
    bundles bigint;
BEGIN
    -- This is a pre-release cutover, never a trust or identity converter. Run
    -- before any DDL so an incompatible deployment remains wholly unchanged.
    LOCK TABLE public.reference_packs, public.reference_pack_job_payloads,
        public.jobs, public.indicators, public.indicator_observations,
        public.reporting_snapshots, public.incident_bundle_exports IN SHARE ROW EXCLUSIVE MODE;
    SELECT count(*) INTO packs FROM public.reference_packs;
    SELECT count(*) INTO pack_jobs FROM public.reference_pack_job_payloads;
    SELECT count(*) INTO retained_pack_jobs FROM public.jobs WHERE job_kind LIKE 'reference_pack.%';
    SELECT count(*) INTO identities FROM public.indicators;
    SELECT count(*) INTO observations FROM public.indicator_observations;
    SELECT count(*) INTO snapshots FROM public.reporting_snapshots;
    SELECT count(*) INTO reporting_jobs FROM public.jobs WHERE job_kind LIKE 'snapshot_reporting.%';
    SELECT count(*) INTO portability_jobs FROM public.jobs WHERE job_kind LIKE 'incident_portability.%';
    SELECT count(*) INTO bundles FROM public.incident_bundle_exports;
    IF portability_jobs > 0 OR bundles > 0 OR reporting_jobs > 0 OR packs > 0 OR pack_jobs > 0 OR retained_pack_jobs > 0 OR identities > 0 OR observations > 0 OR snapshots > 0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: packs=%, pack_jobs=%, retained_pack_jobs=%, indicator_identities=%, observations=%, snapshots=%, reporting_jobs=%, portability_jobs=%, incident_bundles=%; preserve a backup/export with the previous application, then use the explicit development reset procedure', packs, pack_jobs, retained_pack_jobs, identities, observations, snapshots, reporting_jobs, portability_jobs, bundles
            USING ERRCODE = '55000';
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE public.reference_pack_repositories (
    repository_id text CONSTRAINT rp_repositories_pk1 PRIMARY KEY,
    revision bigint NOT NULL CONSTRAINT rp_repositories_ck1 CHECK (revision > 0),
    root_version bigint NOT NULL CONSTRAINT rp_repositories_ck2 CHECK (root_version BETWEEN 1 AND 9007199254740991)
);

CREATE TABLE public.reference_pack_roots (
    repository_id text NOT NULL CONSTRAINT rp_roots_repository_id_fk REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    root_version bigint NOT NULL CONSTRAINT rp_roots_ck1 CHECK (root_version BETWEEN 1 AND 9007199254740991),
    canonical_bytes bytea NOT NULL CONSTRAINT rp_roots_ck2 CHECK (octet_length(canonical_bytes) BETWEEN 1 AND 2097152),
    sha256 text NOT NULL CONSTRAINT rp_roots_ck3 CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    predecessor_version bigint,
    transition_evidence jsonb NOT NULL CONSTRAINT rp_roots_ck4 CHECK (jsonb_typeof(transition_evidence) = 'object'),
    CONSTRAINT rp_roots_pk1 PRIMARY KEY (repository_id, root_version),
    CONSTRAINT rp_roots_fk1 FOREIGN KEY (repository_id, predecessor_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_roots_ck5 CHECK (predecessor_version IS NULL OR predecessor_version = root_version - 1)
);
ALTER TABLE public.reference_pack_repositories ADD CONSTRAINT reference_pack_current_root_fk
    FOREIGN KEY (repository_id, root_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE public.reference_pack_key_state (
    pack_key text CONSTRAINT rp_key_state_pk1 PRIMARY KEY,
    revision bigint NOT NULL DEFAULT 1 CONSTRAINT rp_key_state_ck1 CHECK (revision > 0)
);

CREATE TABLE public.reference_pack_candidates (
    pack_key text NOT NULL CONSTRAINT rp_candidates_pack_key_fk REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT,
    pack_version text NOT NULL,
    distribution_kind text NOT NULL CONSTRAINT rp_candidates_ck1 CHECK (distribution_kind IN ('operator_imported', 'packaged_builtin')),
    health text NOT NULL CONSTRAINT rp_candidates_ck2 CHECK (health IN ('staged', 'verified_available', 'failed', 'missing')),
    last_failure_code text,
    administratively_disabled boolean NOT NULL DEFAULT false,
    removed boolean NOT NULL DEFAULT false,
    missing_reason text CONSTRAINT rp_candidates_ck3 CHECK (missing_reason IN ('administrative_removal', 'storage_loss', 'staging_loss')),
    current_envelope_id text,
    current_index_id uuid,
    admitted_at timestamptz NOT NULL,
    admitted_by_user_id uuid CONSTRAINT rp_candidates_admitted_by_user_id_fk REFERENCES public.users(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_candidates_pk1 PRIMARY KEY (pack_key, pack_version),
    CONSTRAINT rp_candidates_ck4 CHECK ((health = 'missing') = (missing_reason IS NOT NULL)),
    CONSTRAINT rp_candidates_ck5 CHECK (NOT removed OR (health = 'missing' AND missing_reason = 'administrative_removal')),
    CONSTRAINT rp_candidates_ck6 CHECK (distribution_kind <> 'packaged_builtin' OR (NOT removed AND NOT administratively_disabled))
);

CREATE TABLE public.reference_pack_versions (
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    repository_id text CONSTRAINT rp_versions_repository_id_fk REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    pack_release_sequence bigint NOT NULL CONSTRAINT rp_versions_ck1 CHECK (pack_release_sequence BETWEEN 1 AND 9007199254740991),
    manifest_sha256 text NOT NULL CONSTRAINT rp_versions_ck2 CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    payload_sha256 text NOT NULL CONSTRAINT rp_versions_ck3 CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    manifest_bytes bytea NOT NULL CONSTRAINT rp_versions_ck4 CHECK (octet_length(manifest_bytes) BETWEEN 1 AND 1048576),
    CONSTRAINT rp_versions_pk1 PRIMARY KEY (pack_key, pack_version),
    CONSTRAINT rp_versions_uq1 UNIQUE NULLS NOT DISTINCT (repository_id, pack_key, pack_release_sequence),
    CONSTRAINT rp_versions_uq2 UNIQUE (pack_key, pack_version, manifest_sha256, payload_sha256),
    CONSTRAINT rp_versions_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_candidates(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE TABLE public.reference_pack_envelopes (
    envelope_id text CONSTRAINT rp_envelopes_pk1 PRIMARY KEY CONSTRAINT rp_envelopes_ck1 CHECK (envelope_id ~ '^rpenv_[0-9a-f]{64}$'),
    operation_id uuid NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    verified_at timestamptz NOT NULL,
    trust_valid_until timestamptz,
    container_sha256 text CONSTRAINT rp_envelopes_ck2 CHECK (container_sha256 ~ '^[0-9a-f]{64}$'),
    container_ref text,
    repository_id text CONSTRAINT rp_envelopes_repository_id_fk REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    root_version bigint,
    canonical_envelope bytea NOT NULL,
    CONSTRAINT rp_envelopes_uq1 UNIQUE (envelope_id, pack_key, pack_version),
    CONSTRAINT rp_envelopes_uq2 UNIQUE (operation_id, pack_key, pack_version),
    CONSTRAINT rp_envelopes_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_versions(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_envelopes_fk2 FOREIGN KEY (repository_id, root_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_envelopes_ck3 CHECK ((container_sha256 IS NULL) = (container_ref IS NULL)),
    CONSTRAINT rp_envelopes_ck4 CHECK ((repository_id IS NULL) = (root_version IS NULL)),
    CONSTRAINT rp_envelopes_ck5 CHECK ((repository_id IS NULL) = (trust_valid_until IS NULL)),
    CONSTRAINT rp_envelopes_ck6 CHECK (trust_valid_until IS NULL OR verified_at < trust_valid_until)
);
ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT reference_pack_candidate_envelope_fk
    FOREIGN KEY (current_envelope_id, pack_key, pack_version)
    REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE public.reference_pack_metadata_versions (
    repository_id text NOT NULL CONSTRAINT rp_metadata_versions_repository_id_fk REFERENCES public.reference_pack_repositories(repository_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    role text NOT NULL CONSTRAINT rp_metadata_versions_ck1 CHECK (role IN ('timestamp', 'snapshot', 'targets')),
    metadata_version bigint NOT NULL CONSTRAINT rp_metadata_versions_ck2 CHECK (metadata_version BETWEEN 1 AND 9007199254740991),
    canonical_bytes bytea NOT NULL CONSTRAINT rp_metadata_versions_ck3 CHECK (octet_length(canonical_bytes) BETWEEN 1 AND 2097152),
    CONSTRAINT rp_metadata_versions_pk1 PRIMARY KEY (repository_id, pack_key, pack_version, role),
    CONSTRAINT rp_metadata_versions_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_versions(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE TABLE public.reference_pack_operations (
    operation_id uuid CONSTRAINT rp_operations_pk1 PRIMARY KEY,
    job_id uuid CONSTRAINT rp_operations_uq1 UNIQUE CONSTRAINT rp_operations_job_id_fk REFERENCES public.jobs(job_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    kind text NOT NULL CONSTRAINT rp_operations_ck1 CHECK (kind IN ('import', 'renewal', 'reverify', 'refresh', 'activate', 'disable', 'remove', 'integrity', 'reconcile')),
    actor_kind text NOT NULL CONSTRAINT rp_operations_ck2 CHECK (actor_kind IN ('user', 'local_operator', 'system', 'application_release')),
    actor_user_id uuid CONSTRAINT rp_operations_actor_user_id_fk REFERENCES public.users(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    admitted_at timestamptz NOT NULL,
    terminal_at timestamptz,
    frozen_input bytea NOT NULL,
    final_outcome bytea,
    CONSTRAINT rp_operations_ck3 CHECK ((terminal_at IS NULL) = (final_outcome IS NULL)),
    CONSTRAINT rp_operations_ck4 CHECK ((actor_kind = 'user') = (actor_user_id IS NOT NULL))
);
CREATE TABLE public.reference_pack_operation_keys (
    operation_id uuid NOT NULL CONSTRAINT rp_operation_keys_operation_id_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    pack_key text NOT NULL CONSTRAINT rp_operation_keys_pack_key_fk REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT,
    admitted_revision bigint NOT NULL CONSTRAINT rp_operation_keys_ck1 CHECK (admitted_revision > 0),
    usage_revision bigint,
    CONSTRAINT rp_operation_keys_pk1 PRIMARY KEY (operation_id, pack_key)
);
ALTER TABLE public.reference_pack_envelopes ADD CONSTRAINT rp_envelopes_operation_fk
    FOREIGN KEY (operation_id) REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT;
CREATE INDEX reference_pack_operation_keys_pack_idx ON public.reference_pack_operation_keys(pack_key);

CREATE TABLE public.reference_pack_operation_members (
    operation_id uuid NOT NULL CONSTRAINT rp_operation_members_operation_id_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    ordinal bigint NOT NULL CONSTRAINT rp_operation_members_ck1 CHECK (ordinal > 0),
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    envelope_id text,
    CONSTRAINT rp_operation_members_pk1 PRIMARY KEY (operation_id, ordinal),
    CONSTRAINT rp_operation_members_uq1 UNIQUE (operation_id, pack_key, pack_version),
    CONSTRAINT rp_operation_members_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_candidates(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_operation_members_fk2 FOREIGN KEY (envelope_id, pack_key, pack_version) REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);
CREATE TABLE public.reference_pack_operation_repositories (
    operation_id uuid NOT NULL CONSTRAINT rp_operation_repositories_operation_id_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    repository_id text NOT NULL,
    admitted_revision bigint NOT NULL CONSTRAINT rp_operation_repositories_ck1 CHECK (admitted_revision > 0),
    root_version bigint NOT NULL,
    CONSTRAINT rp_operation_repositories_pk1 PRIMARY KEY (operation_id, repository_id),
    CONSTRAINT rp_operation_repositories_fk1 FOREIGN KEY (repository_id, root_version) REFERENCES public.reference_pack_roots(repository_id, root_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE TABLE public.reference_pack_attempts (
    attempt_id uuid CONSTRAINT rp_attempts_pk1 PRIMARY KEY,
    operation_id uuid NOT NULL CONSTRAINT rp_attempts_operation_id_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    outcome text CONSTRAINT rp_attempts_ck1 CHECK (outcome IN ('succeeded', 'content_rejected', 'canceled', 'timed_out', 'stale_state', 'execution_failed', 'interrupted')),
    canonical_result bytea,
    CONSTRAINT rp_attempts_ck2 CHECK ((completed_at IS NULL) = (outcome IS NULL)),
    CONSTRAINT rp_attempts_ck3 CHECK ((outcome IS NULL) = (canonical_result IS NULL))
);
CREATE INDEX reference_pack_attempts_operation_idx ON public.reference_pack_attempts(operation_id, started_at);

-- Preparation results are private to an execution attempt. A terminal
-- transaction consumes the complete ordered cohort; no partial verdict is
-- projected into candidate health or trust while preparation is running.
CREATE TABLE public.reference_pack_attempt_members (
    attempt_id uuid NOT NULL CONSTRAINT rp_attempt_members_attempt_id_fk REFERENCES public.reference_pack_attempts(attempt_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    ordinal bigint NOT NULL CONSTRAINT rp_attempt_members_ck1 CHECK (ordinal > 0),
    pack_key text,
    pack_version text,
    verdict text NOT NULL CONSTRAINT rp_attempt_members_ck2 CHECK (verdict IN ('succeeded', 'content_rejected')),
    canonical_prepared bytea,
    failure_code text,
    check_id text,
    CONSTRAINT rp_attempt_members_pk1 PRIMARY KEY (attempt_id, ordinal),
    CONSTRAINT rp_attempt_members_ck3 CHECK ((pack_key IS NULL) = (pack_version IS NULL)),
    CONSTRAINT rp_attempt_members_ck4 CHECK ((verdict = 'succeeded') = (canonical_prepared IS NOT NULL)),
    CONSTRAINT rp_attempt_members_ck5 CHECK ((verdict = 'content_rejected') = (failure_code IS NOT NULL AND check_id IS NOT NULL)),
    CONSTRAINT rp_attempt_members_ck6 CHECK (verdict <> 'succeeded' OR pack_key IS NOT NULL)
);

CREATE TABLE public.reference_pack_objects (
    object_id uuid CONSTRAINT rp_objects_pk1 PRIMARY KEY,
    sha256 text NOT NULL CONSTRAINT rp_objects_ck1 CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    storage_ref text NOT NULL CONSTRAINT rp_objects_uq1 UNIQUE,
    size_bytes bigint NOT NULL CONSTRAINT rp_objects_ck2 CHECK (size_bytes >= 0),
    generation bigint NOT NULL CONSTRAINT rp_objects_ck3 CHECK (generation > 0),
    available boolean NOT NULL DEFAULT true
);
CREATE TABLE public.reference_pack_object_refs (
    owner_kind text NOT NULL CONSTRAINT rp_object_refs_ck1 CHECK (owner_kind IN ('version', 'envelope', 'operation', 'backup', 'release')),
    owner_id text NOT NULL,
    logical_path text NOT NULL,
    object_id uuid NOT NULL CONSTRAINT rp_object_refs_object_id_fk REFERENCES public.reference_pack_objects(object_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_object_refs_pk1 PRIMARY KEY (owner_kind, owner_id, logical_path)
);
CREATE INDEX reference_pack_objects_digest_idx ON public.reference_pack_objects(sha256);
CREATE INDEX reference_pack_object_refs_object_idx ON public.reference_pack_object_refs(object_id);

CREATE TABLE public.reference_pack_index_generations (
    index_id uuid CONSTRAINT rp_index_generations_pk1 PRIMARY KEY,
    operation_id uuid NOT NULL CONSTRAINT rp_index_generations_operation_id_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    manifest_sha256 text NOT NULL CONSTRAINT rp_index_generations_ck1 CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    payload_sha256 text NOT NULL CONSTRAINT rp_index_generations_ck2 CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    entry_count bigint NOT NULL DEFAULT 0 CONSTRAINT rp_index_generations_ck3 CHECK (entry_count BETWEEN 0 AND 2000000),
    object_count bigint NOT NULL DEFAULT 0 CONSTRAINT rp_index_generations_ck4 CHECK (object_count BETWEEN 0 AND 2000000),
    relationship_count bigint NOT NULL DEFAULT 0 CONSTRAINT rp_index_generations_ck5 CHECK (relationship_count BETWEEN 0 AND 5000000),
    complete boolean NOT NULL DEFAULT false,
    CONSTRAINT rp_index_generations_uq1 UNIQUE (index_id, pack_key, pack_version)
);
ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT rp_candidates_index_fk
    FOREIGN KEY (current_index_id, pack_key, pack_version)
    REFERENCES public.reference_pack_index_generations(index_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;

CREATE TABLE public.reference_pack_indexes (
    index_id uuid NOT NULL CONSTRAINT rp_indexes_index_id_fk REFERENCES public.reference_pack_index_generations(index_id) ON UPDATE RESTRICT ON DELETE CASCADE,
    entry_kind text NOT NULL CONSTRAINT rp_indexes_ck1 CHECK (entry_kind IN ('entry', 'object', 'relationship')),
    entry_id text NOT NULL CONSTRAINT rp_indexes_ck2 CHECK (octet_length(entry_id) BETWEEN 1 AND 512),
    canonical_item bytea NOT NULL CONSTRAINT rp_indexes_ck3 CHECK (octet_length(canonical_item) BETWEEN 1 AND 1048576),
    lookup_keys bytea NOT NULL,
    CONSTRAINT rp_indexes_pk1 PRIMARY KEY (index_id, entry_kind, entry_id)
);
CREATE TABLE public.reference_pack_lookup_keys (
    index_id uuid NOT NULL,
    entry_kind text NOT NULL,
    entry_id text NOT NULL,
    lookup_kind text NOT NULL,
    lookup_value bytea NOT NULL,
    lookup_sha256 bytea GENERATED ALWAYS AS (public.digest(lookup_value, 'sha256')) STORED,
    sort_key bytea NOT NULL,
    CONSTRAINT rp_lookup_keys_pk1 PRIMARY KEY (index_id, entry_kind, entry_id, lookup_kind, lookup_sha256),
    CONSTRAINT rp_lookup_keys_fk1 FOREIGN KEY (index_id, entry_kind, entry_id)
        REFERENCES public.reference_pack_indexes(index_id, entry_kind, entry_id) ON UPDATE RESTRICT ON DELETE CASCADE
);
-- B-tree keys hash unbounded normalized lookup values to keep PostgreSQL's
-- index-tuple size independent of the admitted 8192-scalar lookup boundary.
CREATE INDEX reference_pack_lookup_query_idx ON public.reference_pack_lookup_keys(index_id, lookup_kind, lookup_sha256);

CREATE TABLE public.reference_pack_sets (
    pack_set_id text CONSTRAINT rp_sets_pk1 PRIMARY KEY CONSTRAINT rp_sets_ck1 CHECK (pack_set_id ~ '^rpset_[0-9a-f]{64}$'),
    pack_set_sha256 text NOT NULL CONSTRAINT rp_sets_uq1 UNIQUE CONSTRAINT rp_sets_ck2 CHECK (pack_set_sha256 ~ '^[0-9a-f]{64}$'),
    canonical_set bytea NOT NULL,
    canonical_provenance bytea NOT NULL,
    first_operation_id uuid NOT NULL CONSTRAINT rp_sets_first_operation_id_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_sets_ck3 CHECK (pack_set_id = 'rpset_' || pack_set_sha256)
);
CREATE TABLE public.reference_pack_set_members (
    pack_set_id text NOT NULL CONSTRAINT rp_set_members_pack_set_id_fk REFERENCES public.reference_pack_sets(pack_set_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    manifest_sha256 text NOT NULL,
    payload_sha256 text NOT NULL,
    provenance_envelope_id text NOT NULL,
    CONSTRAINT rp_set_members_pk1 PRIMARY KEY (pack_set_id, pack_key),
    CONSTRAINT rp_set_members_fk1 FOREIGN KEY (pack_key, pack_version, manifest_sha256, payload_sha256)
        REFERENCES public.reference_pack_versions(pack_key, pack_version, manifest_sha256, payload_sha256) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_set_members_fk2 FOREIGN KEY (provenance_envelope_id, pack_key, pack_version)
        REFERENCES public.reference_pack_envelopes(envelope_id, pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);
CREATE TABLE public.reference_pack_current_set (
    singleton boolean CONSTRAINT rp_current_set_pk1 PRIMARY KEY DEFAULT true CONSTRAINT rp_current_set_ck1 CHECK (singleton),
    pack_set_id text CONSTRAINT rp_current_set_pack_set_id_fk REFERENCES public.reference_pack_sets(pack_set_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    revision bigint NOT NULL CONSTRAINT rp_current_set_ck2 CHECK (revision > 0),
    configuration_sha256 text NOT NULL CONSTRAINT rp_current_set_ck3 CHECK (configuration_sha256 ~ '^[0-9a-f]{64}$'),
    application_release_id text,
    profile_claimed boolean NOT NULL DEFAULT false
);
CREATE TABLE public.reference_pack_pins (
    owner_kind text NOT NULL,
    owner_id text NOT NULL,
    pack_set_id text NOT NULL CONSTRAINT rp_pins_pack_set_id_fk REFERENCES public.reference_pack_sets(pack_set_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    operation_id uuid NOT NULL,
    CONSTRAINT rp_pins_pk1 PRIMARY KEY (owner_kind, owner_id, pack_set_id)
);
CREATE INDEX reference_pack_pins_set_idx ON public.reference_pack_pins(pack_set_id);
CREATE TABLE public.reference_pack_registry_usage (
    pack_key text CONSTRAINT rp_registry_usage_pk1 PRIMARY KEY CONSTRAINT rp_registry_usage_pack_key_fk REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT,
    revision bigint NOT NULL CONSTRAINT rp_registry_usage_ck1 CHECK (revision > 0)
);
CREATE TABLE public.reference_pack_events (
    attestation_id text NOT NULL CONSTRAINT rp_events_ck1 CHECK (attestation_id ~ '^rpa_[0-9a-f]{64}$'),
    operation_id uuid NOT NULL CONSTRAINT rp_events_operation_id_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    event_kind text NOT NULL CONSTRAINT rp_events_ck2 CHECK (event_kind IN ('import_verification','reverification','refresh_verification','activation','rollback_activation','safety_fallback','profile_reconciliation','dependency_invalidation','disablement','removal','exact_reimport','payload_invalidation','trust_root_update')),
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    canonical_attestation bytea NOT NULL,
    CONSTRAINT rp_events_pk1 PRIMARY KEY (operation_id, attestation_id),
    CONSTRAINT rp_events_uq1 UNIQUE (operation_id, event_kind, pack_key, pack_version),
    CONSTRAINT rp_events_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_candidates(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE TABLE public.reference_pack_release_bindings (
    application_release_id text NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    canonical_binding bytea NOT NULL,
    CONSTRAINT rp_release_bindings_pk1 PRIMARY KEY (application_release_id, pack_key),
    CONSTRAINT rp_release_bindings_fk1 FOREIGN KEY (pack_key, pack_version) REFERENCES public.reference_pack_versions(pack_key, pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_immutable_row() RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
    RAISE EXCEPTION 'immutable reference pack history' USING ERRCODE = '55000';
END $$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.reference_pack_immutable_row() FROM PUBLIC;
CREATE TRIGGER reference_pack_versions_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_versions FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_roots_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_roots FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_envelopes_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_envelopes FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_sets_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_sets FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_set_members_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_set_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_events_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_events FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_release_bindings_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_release_bindings FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

CREATE TRIGGER reference_pack_operation_keys_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_operation_keys FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_operation_members_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_operation_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_operation_repositories_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_operation_repositories FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_attempt_members_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_attempt_members FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

-- Frozen inputs and terminal outcomes cannot be edited to authorize a new
-- interpretation of already retained preparation or publication evidence.
-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_operation_guard() RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
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
CREATE TRIGGER reference_pack_operations_guard BEFORE UPDATE OR DELETE ON public.reference_pack_operations FOR EACH ROW EXECUTE FUNCTION public.reference_pack_operation_guard();

-- +goose StatementBegin
CREATE FUNCTION public.reference_pack_attempt_guard() RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
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
CREATE TRIGGER reference_pack_attempts_guard BEFORE UPDATE OR DELETE ON public.reference_pack_attempts FOR EACH ROW EXECUTE FUNCTION public.reference_pack_attempt_guard();

-- Grant only the operations used by the runtime; immutable history has no
-- runtime update/delete privilege. Recovery can load a complete snapshot.
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_repositories TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_repositories TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_roots TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_roots TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_key_state TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_key_state TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_candidates TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_candidates TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_versions TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_versions TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_envelopes TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_envelopes TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_metadata_versions TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_metadata_versions TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_operations TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operations TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_operation_keys TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_keys TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_operation_members TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_members TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_operation_repositories TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_repositories TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_attempts TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_attempts TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_attempt_members TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_attempt_members TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_objects TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_objects TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_object_refs TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_object_refs TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_index_generations TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_index_generations TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_indexes TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_indexes TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_lookup_keys TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_lookup_keys TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_sets TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_sets TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_set_members TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_set_members TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_current_set TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_current_set TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_pins TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_pins TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.reference_pack_registry_usage TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_registry_usage TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_events TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_events TO cartulary_recovery;
GRANT SELECT, INSERT ON TABLE public.reference_pack_release_bindings TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_release_bindings TO cartulary_recovery;

-- Foreign-key leading indexes bound parent mutation checks independently of
-- retained history size, including nullable and deferred references.
CREATE INDEX rp_roots_predecessor_idx ON public.reference_pack_roots(repository_id, predecessor_version);
CREATE INDEX rp_repositories_root_idx ON public.reference_pack_repositories(repository_id, root_version);
CREATE INDEX rp_candidates_actor_idx ON public.reference_pack_candidates(admitted_by_user_id);
CREATE INDEX rp_candidates_envelope_idx ON public.reference_pack_candidates(current_envelope_id, pack_key, pack_version);
CREATE INDEX rp_candidates_index_idx ON public.reference_pack_candidates(current_index_id, pack_key, pack_version);
CREATE INDEX rp_envelopes_version_idx ON public.reference_pack_envelopes(pack_key, pack_version);
CREATE INDEX rp_envelopes_root_idx ON public.reference_pack_envelopes(repository_id, root_version);
CREATE INDEX rp_metadata_versions_version_idx ON public.reference_pack_metadata_versions(pack_key, pack_version);
CREATE INDEX rp_operations_actor_idx ON public.reference_pack_operations(actor_user_id);
CREATE INDEX rp_operation_members_version_idx ON public.reference_pack_operation_members(pack_key, pack_version);
CREATE INDEX rp_operation_members_envelope_idx ON public.reference_pack_operation_members(envelope_id, pack_key, pack_version);
CREATE INDEX rp_operation_repositories_root_idx ON public.reference_pack_operation_repositories(repository_id, root_version);
CREATE INDEX rp_index_generations_operation_idx ON public.reference_pack_index_generations(operation_id);
CREATE INDEX rp_sets_operation_idx ON public.reference_pack_sets(first_operation_id);
CREATE INDEX rp_set_members_version_idx ON public.reference_pack_set_members(pack_key, pack_version, manifest_sha256, payload_sha256);
CREATE INDEX rp_set_members_envelope_idx ON public.reference_pack_set_members(provenance_envelope_id, pack_key, pack_version);
CREATE INDEX rp_current_set_set_idx ON public.reference_pack_current_set(pack_set_id);
CREATE INDEX rp_events_version_idx ON public.reference_pack_events(pack_key, pack_version);
CREATE INDEX rp_release_bindings_version_idx ON public.reference_pack_release_bindings(pack_key, pack_version);

-- Table composite types are schema objects too; restrict their use to the
-- admitted runtime and recovery purposes, as with the existing owner tables.
REVOKE USAGE ON TYPE
    public.reference_pack_repositories,
    public.reference_pack_roots,
    public.reference_pack_key_state,
    public.reference_pack_candidates,
    public.reference_pack_versions,
    public.reference_pack_envelopes,
    public.reference_pack_metadata_versions,
    public.reference_pack_operations,
    public.reference_pack_operation_keys,
    public.reference_pack_operation_members,
    public.reference_pack_operation_repositories,
    public.reference_pack_attempts,
    public.reference_pack_attempt_members,
    public.reference_pack_objects,
    public.reference_pack_object_refs,
    public.reference_pack_index_generations,
    public.reference_pack_indexes,
    public.reference_pack_lookup_keys,
    public.reference_pack_sets,
    public.reference_pack_set_members,
    public.reference_pack_current_set,
    public.reference_pack_pins,
    public.reference_pack_registry_usage,
    public.reference_pack_events,
    public.reference_pack_release_bindings
FROM PUBLIC;
GRANT USAGE ON TYPE
    public.reference_pack_repositories,
    public.reference_pack_roots,
    public.reference_pack_key_state,
    public.reference_pack_candidates,
    public.reference_pack_versions,
    public.reference_pack_envelopes,
    public.reference_pack_metadata_versions,
    public.reference_pack_operations,
    public.reference_pack_operation_keys,
    public.reference_pack_operation_members,
    public.reference_pack_operation_repositories,
    public.reference_pack_attempts,
    public.reference_pack_attempt_members,
    public.reference_pack_objects,
    public.reference_pack_object_refs,
    public.reference_pack_index_generations,
    public.reference_pack_indexes,
    public.reference_pack_lookup_keys,
    public.reference_pack_sets,
    public.reference_pack_set_members,
    public.reference_pack_current_set,
    public.reference_pack_pins,
    public.reference_pack_registry_usage,
    public.reference_pack_events,
    public.reference_pack_release_bindings
TO cartulary_runtime, cartulary_recovery;

-- +goose Down
-- No downgrade can invent the legacy mutable trust model or preserve the
-- semantics of retained immutable history. Restore a matching full backup.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'reference pack cutover cannot be downgraded; restore a complete backup with its matching application' USING ERRCODE = '55000'; END $$;
-- +goose StatementEnd

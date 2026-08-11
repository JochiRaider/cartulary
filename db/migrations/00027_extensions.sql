-- +goose Up
--
-- Name: extension coordination state; Type: TABLES; Schema: public; Owner: -
--

CREATE TABLE public.extension_state_metadata (
    profile_id text CONSTRAINT extension_state_metadata_pkey PRIMARY KEY,
    migration_lineage_id text NOT NULL,
    state_version integer NOT NULL,
    last_migration_id text,
    metadata_version integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT extension_state_metadata_profile_id_ck
        CHECK (profile_id ~ '^[a-z][a-z0-9_]{0,127}$'),
    CONSTRAINT extension_state_metadata_lineage_ck
        CHECK (migration_lineage_id ~ '^[a-z][a-z0-9_.]{0,159}$'),
    CONSTRAINT extension_state_metadata_state_version_ck
        CHECK (state_version BETWEEN 1 AND 2147483647),
    CONSTRAINT extension_state_metadata_metadata_version_ck
        CHECK (metadata_version BETWEEN 1 AND 2147483647),
    CONSTRAINT extension_state_metadata_last_migration_ck
        CHECK (last_migration_id IS NULL OR last_migration_id ~ '^[a-z][a-z0-9_.]{0,159}$'),
    CONSTRAINT extension_state_metadata_time_ck CHECK (updated_at >= created_at)
);

CREATE TABLE public.extension_migration_ledger (
    profile_id text NOT NULL,
    migration_lineage_id text NOT NULL,
    migration_id text NOT NULL,
    from_state_version integer NOT NULL,
    to_state_version integer NOT NULL,
    migration_definition_sha256 text NOT NULL,
    committed_at timestamp with time zone NOT NULL,
    resulting_state_version integer NOT NULL,
    CONSTRAINT extension_migration_ledger_pk
        PRIMARY KEY (profile_id, migration_lineage_id, from_state_version, to_state_version),
    CONSTRAINT extension_migration_ledger_migration_identity_uq
        UNIQUE (profile_id, migration_lineage_id, migration_id),
    CONSTRAINT extension_migration_ledger_profile_id_ck
        CHECK (profile_id ~ '^[a-z][a-z0-9_]{0,127}$'),
    CONSTRAINT extension_migration_ledger_lineage_ck
        CHECK (migration_lineage_id ~ '^[a-z][a-z0-9_.]{0,159}$'),
    CONSTRAINT extension_migration_ledger_migration_id_ck
        CHECK (migration_id ~ '^[a-z][a-z0-9_.]{0,159}$'),
    CONSTRAINT extension_migration_ledger_versions_ck
        CHECK (from_state_version BETWEEN 1 AND 2147483646
            AND to_state_version = from_state_version + 1
            AND resulting_state_version = to_state_version),
    CONSTRAINT extension_migration_ledger_digest_ck
        CHECK (migration_definition_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT extension_migration_ledger_metadata_fk
        FOREIGN KEY (profile_id) REFERENCES public.extension_state_metadata(profile_id) ON UPDATE NO ACTION ON DELETE NO ACTION
        DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE public.extension_staged_objects (
    staging_id text CONSTRAINT extension_staged_objects_pkey PRIMARY KEY,
    owner_profile_id text NOT NULL,
    storage_identity text NOT NULL
        CONSTRAINT extension_staged_objects_storage_identity_key UNIQUE,
    expected_byte_size bigint NOT NULL,
    expected_sha256 text NOT NULL,
    staged_at timestamp with time zone NOT NULL,
    staging_expires_at timestamp with time zone NOT NULL,
    ready_at timestamp with time zone,
    published_at timestamp with time zone,
    abandoned_at timestamp with time zone,
    state text NOT NULL,
    delete_state text NOT NULL,
    delete_attempt_count integer NOT NULL,
    next_delete_attempt_at timestamp with time zone,
    last_delete_error_code text,
    CONSTRAINT extension_staged_objects_id_ck
        CHECK (staging_id ~ '^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$'),
    CONSTRAINT extension_staged_objects_profile_id_ck
        CHECK (owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'),
    CONSTRAINT extension_staged_objects_storage_identity_ck
        CHECK (octet_length(storage_identity) BETWEEN 1 AND 512
            AND storage_identity !~ '[[:cntrl:]]'),
    CONSTRAINT extension_staged_objects_size_ck CHECK (expected_byte_size >= 0),
    CONSTRAINT extension_staged_objects_digest_ck CHECK (expected_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT extension_staged_objects_expiry_ck
        CHECK (staging_expires_at = staged_at + interval '24 hours'),
    CONSTRAINT extension_staged_objects_state_ck
        CHECK (state IN ('allocated', 'ready', 'published', 'abandoned')),
    CONSTRAINT extension_staged_objects_delete_state_ck
        CHECK (delete_state IN ('not_applicable', 'pending', 'deleted')),
    CONSTRAINT extension_staged_objects_attempt_count_ck
        CHECK (delete_attempt_count BETWEEN 0 AND 2147483647),
    CONSTRAINT extension_staged_objects_error_code_ck
        CHECK (last_delete_error_code IS NULL
            OR last_delete_error_code ~ '^[A-Za-z][A-Za-z0-9_.:-]{0,127}$'),
    CONSTRAINT extension_staged_objects_lifecycle_ck CHECK (
        (state = 'allocated' AND ready_at IS NULL AND published_at IS NULL
            AND abandoned_at IS NULL AND delete_state = 'not_applicable'
            AND next_delete_attempt_at IS NULL AND last_delete_error_code IS NULL)
        OR (state = 'ready' AND ready_at IS NOT NULL AND published_at IS NULL
            AND abandoned_at IS NULL AND delete_state = 'not_applicable'
            AND next_delete_attempt_at IS NULL AND last_delete_error_code IS NULL)
        OR (state = 'published' AND ready_at IS NOT NULL AND published_at IS NOT NULL
            AND abandoned_at IS NULL AND delete_state = 'not_applicable'
            AND next_delete_attempt_at IS NULL AND last_delete_error_code IS NULL)
        OR (state = 'abandoned' AND ready_at IS NULL AND published_at IS NULL
            AND abandoned_at IS NOT NULL
            AND ((delete_state = 'pending' AND next_delete_attempt_at IS NOT NULL)
                OR (delete_state = 'deleted' AND next_delete_attempt_at IS NULL
                    AND last_delete_error_code IS NULL)))
    )
);

CREATE INDEX extension_staged_objects_cleanup_idx
    ON public.extension_staged_objects (
        (CASE
            WHEN state IN ('allocated', 'ready') THEN staging_expires_at
            WHEN state = 'abandoned' AND delete_state = 'pending' THEN next_delete_attempt_at
            ELSE NULL
        END),
        staging_id
    )
    WHERE state IN ('allocated', 'ready') OR (state = 'abandoned' AND delete_state = 'pending');

CREATE TABLE public.extension_staged_object_references (
    staging_id text CONSTRAINT extension_staged_object_references_pkey PRIMARY KEY
        CONSTRAINT extension_staged_object_references_staging_id_fkey
        REFERENCES public.extension_staged_objects(staging_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    owner_resource_kind text NOT NULL,
    owner_resource_id text NOT NULL,
    expected_byte_size bigint NOT NULL,
    expected_sha256 text NOT NULL,
    committed_at timestamp with time zone NOT NULL,
    CONSTRAINT extension_staged_object_references_kind_ck
        CHECK (owner_resource_kind ~ '^[a-z][a-z0-9_]{0,127}$'),
    CONSTRAINT extension_staged_object_references_id_ck
        CHECK (octet_length(owner_resource_id) BETWEEN 1 AND 512
            AND owner_resource_id !~ '[[:cntrl:]]'),
    CONSTRAINT extension_staged_object_references_size_ck CHECK (expected_byte_size >= 0),
    CONSTRAINT extension_staged_object_references_digest_ck CHECK (expected_sha256 ~ '^[0-9a-f]{64}$')
);

CREATE TABLE public.extension_job_commit_proofs (
    job_id uuid CONSTRAINT extension_job_commit_proofs_pkey PRIMARY KEY
        CONSTRAINT extension_job_commit_proofs_job_id_fkey
        REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    owner_profile_id text NOT NULL,
    operation_kind text NOT NULL,
    final_commit_id text NOT NULL
        CONSTRAINT extension_job_commit_proofs_final_commit_id_key UNIQUE,
    idempotency_identity jsonb,
    normalized_request_sha256 text NOT NULL,
    terminal_result jsonb NOT NULL,
    terminal_result_sha256 text NOT NULL,
    resource_refs jsonb NOT NULL,
    audit_correlation_id text,
    committed_at timestamp with time zone NOT NULL,
    CONSTRAINT extension_job_commit_proofs_profile_ck
        CHECK (owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'),
    CONSTRAINT extension_job_commit_proofs_operation_ck
        CHECK (operation_kind ~ '^[a-z][a-z0-9_.]{0,159}$'),
    CONSTRAINT extension_job_commit_proofs_commit_id_ck
        CHECK (final_commit_id ~ '^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$'),
    CONSTRAINT extension_job_commit_proofs_request_digest_ck
        CHECK (normalized_request_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT extension_job_commit_proofs_result_digest_ck
        CHECK (terminal_result_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT extension_job_commit_proofs_result_ck
        CHECK (jsonb_typeof(terminal_result) = 'object'),
    CONSTRAINT extension_job_commit_proofs_refs_ck
        CHECK (jsonb_typeof(resource_refs) = 'array' AND jsonb_array_length(resource_refs) <= 1024),
    CONSTRAINT extension_job_commit_proofs_audit_ck
        CHECK (audit_correlation_id IS NULL
            OR audit_correlation_id ~ '^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$')
);

CREATE TABLE public.extension_job_cancellation_observations (
    cancellation_request_id text
        CONSTRAINT extension_job_cancellation_observations_pkey PRIMARY KEY,
    job_id uuid NOT NULL
        CONSTRAINT extension_job_cancellation_observations_job_id_key UNIQUE
        CONSTRAINT extension_job_cancellation_observations_job_id_fkey
        REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    observed_at timestamp with time zone NOT NULL,
    observed_before_final_commit boolean NOT NULL,
    CONSTRAINT extension_job_cancellation_observations_id_ck
        CHECK (cancellation_request_id ~ '^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$'),
    CONSTRAINT extension_job_cancellation_observations_precommit_ck
        CHECK (observed_before_final_commit)
);

-- +goose Down
DROP TABLE public.extension_job_cancellation_observations;
DROP TABLE public.extension_job_commit_proofs;
DROP TABLE public.extension_staged_object_references;
DROP TABLE public.extension_staged_objects;
DROP TABLE public.extension_migration_ledger;
DROP TABLE public.extension_state_metadata;

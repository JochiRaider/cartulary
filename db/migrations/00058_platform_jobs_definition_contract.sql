-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    invalid_progress_count bigint;
    unknown_mapping_count bigint;
    active_lease_count bigint;
    illegal_replay_count bigint;
    bounded_tokens text;
BEGIN
    SELECT count(*)
      INTO invalid_progress_count
      FROM jobs
     WHERE progress_completed < 0
        OR progress_total <= 0
        OR progress_completed > progress_total
        OR (status = 'succeeded' AND progress_total IS NOT NULL AND progress_completed <> progress_total);

    SELECT count(*)
      INTO unknown_mapping_count
      FROM jobs
     WHERE (status IN ('queued', 'running', 'cancel_requested')
            OR extension_owner_profile_id IS NOT NULL
            OR extension_job_kind IS NOT NULL)
       AND (
           (extension_owner_profile_id = 'import' AND extension_job_kind IN ('import.discovery_v1', 'import.apply_v1'))
           OR (extension_owner_profile_id = 'incident_portability' AND extension_job_kind IN ('incident_portability.export_v1', 'incident_portability.import_v1'))
           OR (extension_owner_profile_id = 'reference_pack' AND extension_job_kind IN ('reference_pack.import_v1', 'reference_pack.refresh_v1', 'reference_pack.reverify_v1'))
           OR (extension_owner_profile_id = 'snapshot_reporting' AND extension_job_kind IN ('snapshot_reporting.composition_preview_v1', 'snapshot_reporting.release_create_v1', 'snapshot_reporting.snapshot_create_v1'))
       ) IS NOT TRUE;

    SELECT count(*)
      INTO active_lease_count
      FROM jobs
     WHERE handler_lease_owner IS NOT NULL
       AND handler_lease_expires_at > now();

    WITH replay AS (
        SELECT source_identity,
               canonical_payload ->> 'status' AS status,
               lag(canonical_payload ->> 'status') OVER (
                   PARTITION BY source_identity
                   ORDER BY created_at, intent_key
               ) AS prior_status
          FROM collaboration_event_intents
         WHERE event_family = 'job_progress'
           AND created_at >= now() - interval '24 hours'
    )
    SELECT count(*)
      INTO illegal_replay_count
      FROM replay
     WHERE prior_status IS NOT NULL
       AND status IS DISTINCT FROM prior_status
       AND NOT (
           (prior_status = 'queued' AND status IN ('running', 'cancel_requested'))
           OR (prior_status = 'running' AND status IN ('cancel_requested', 'succeeded', 'failed'))
           OR (prior_status = 'cancel_requested' AND status IN ('succeeded', 'failed', 'canceled'))
       );

    SELECT COALESCE(string_agg(token, ',' ORDER BY token), 'none')
      INTO bounded_tokens
      FROM (
          SELECT DISTINCT
                 left(COALESCE(extension_job_kind, 'missing'), 63) || ':' ||
                 left(status, 31) AS token
            FROM jobs
           WHERE (status IN ('queued', 'running', 'cancel_requested')
                  OR extension_owner_profile_id IS NOT NULL
                  OR extension_job_kind IS NOT NULL)
             AND (
                 (extension_owner_profile_id = 'import' AND extension_job_kind IN ('import.discovery_v1', 'import.apply_v1'))
                 OR (extension_owner_profile_id = 'incident_portability' AND extension_job_kind IN ('incident_portability.export_v1', 'incident_portability.import_v1'))
                 OR (extension_owner_profile_id = 'reference_pack' AND extension_job_kind IN ('reference_pack.import_v1', 'reference_pack.refresh_v1', 'reference_pack.reverify_v1'))
                 OR (extension_owner_profile_id = 'snapshot_reporting' AND extension_job_kind IN ('snapshot_reporting.composition_preview_v1', 'snapshot_reporting.release_create_v1', 'snapshot_reporting.snapshot_create_v1'))
             ) IS NOT TRUE
           ORDER BY token
           LIMIT 10
      ) AS bounded;

    IF invalid_progress_count > 0 OR unknown_mapping_count > 0 OR
       active_lease_count > 0 OR illegal_replay_count > 0 THEN
        RAISE EXCEPTION USING MESSAGE = format(
            'jobs v2 compatibility preflight failed: invalid_progress_count=%s unknown_mapping_count=%s active_lease_count=%s illegal_replay_count=%s kind_status_tokens=%s; drain workers and use only the approved reset/reseed path for unsupported retained data',
            invalid_progress_count,
            unknown_mapping_count,
            active_lease_count,
            illegal_replay_count,
            bounded_tokens
        );
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE jobs RENAME COLUMN extension_job_kind TO job_kind;

ALTER TABLE jobs
    ADD COLUMN progress_unit_id text,
    DROP CONSTRAINT jobs_extension_ownership_ck,
    ADD CONSTRAINT jobs_extension_ownership_ck CHECK (
        (extension_owner_profile_id IS NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
        OR (extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'
            AND job_kind IS NOT NULL
            AND jsonb_typeof(extension_idempotency_identity) = 'object'
            AND octet_length(extension_idempotency_route_key) BETWEEN 1 AND 256
            AND octet_length(extension_idempotency_scope_key) BETWEEN 1 AND 512
            AND extension_normalized_request_sha256 ~ '^[0-9a-f]{64}$')
    );

UPDATE jobs
   SET progress_unit_id = CASE job_kind
       WHEN 'import.discovery_v1' THEN 'import.discovery.session.v1'
       WHEN 'import.apply_v1' THEN 'import.apply.import_unit.v1'
       WHEN 'incident_portability.export_v1' THEN 'incident_portability.export.request.v1'
       WHEN 'incident_portability.import_v1' THEN 'incident_portability.import.request.v1'
       WHEN 'reference_pack.import_v1' THEN 'reference_pack.import.request.v1'
       WHEN 'reference_pack.refresh_v1' THEN 'reference_pack.refresh.pack_key.v1'
       WHEN 'reference_pack.reverify_v1' THEN 'reference_pack.reverify.pack_version.v1'
       WHEN 'snapshot_reporting.composition_preview_v1' THEN 'snapshot_reporting.composition_preview.render_attempt.v1'
       WHEN 'snapshot_reporting.release_create_v1' THEN 'snapshot_reporting.release_create.render_attempt.v1'
       WHEN 'snapshot_reporting.snapshot_create_v1' THEN 'snapshot_reporting.snapshot_create.materialization.v1'
       ELSE NULL
   END
 WHERE job_kind IS NOT NULL;

ALTER TABLE jobs
    ADD CONSTRAINT jobs_definition_pair_ck CHECK (
        (job_kind IS NULL AND progress_unit_id IS NULL)
        OR (job_kind IS NOT NULL AND progress_unit_id IS NOT NULL)
    ),
    ADD CONSTRAINT jobs_nonterminal_definition_ck CHECK (
        status IN ('succeeded', 'failed', 'canceled')
        OR (job_kind IS NOT NULL AND progress_unit_id IS NOT NULL)
    ),
    ADD CONSTRAINT jobs_job_kind_nonempty_ck CHECK (
        job_kind IS NULL OR (job_kind <> '' AND length(job_kind) <= 191)
    ),
    ADD CONSTRAINT jobs_progress_unit_id_shape_ck CHECK (
        progress_unit_id IS NULL
        OR progress_unit_id ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*){2,}\.v[1-9][0-9]*$'
    ),
    ADD CONSTRAINT jobs_succeeded_progress_ck CHECK (
        status <> 'succeeded'
        OR progress_total IS NULL
        OR progress_completed = progress_total
    );

-- +goose Down
ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS jobs_extension_ownership_ck,
    DROP CONSTRAINT IF EXISTS jobs_succeeded_progress_ck,
    DROP CONSTRAINT IF EXISTS jobs_progress_unit_id_shape_ck,
    DROP CONSTRAINT IF EXISTS jobs_job_kind_nonempty_ck,
    DROP CONSTRAINT IF EXISTS jobs_nonterminal_definition_ck,
    DROP CONSTRAINT IF EXISTS jobs_definition_pair_ck,
    DROP COLUMN IF EXISTS progress_unit_id;

ALTER TABLE jobs RENAME COLUMN job_kind TO extension_job_kind;

ALTER TABLE jobs
    ADD CONSTRAINT jobs_extension_ownership_ck CHECK (
        (extension_owner_profile_id IS NULL
            AND extension_job_kind IS NULL
            AND extension_idempotency_identity IS NULL
            AND extension_idempotency_route_key IS NULL
            AND extension_idempotency_scope_key IS NULL
            AND extension_normalized_request_sha256 IS NULL)
        OR (extension_owner_profile_id ~ '^[a-z][a-z0-9_]{0,127}$'
            AND extension_job_kind ~ '^[a-z][a-z0-9_.]{0,159}$'
            AND jsonb_typeof(extension_idempotency_identity) = 'object'
            AND octet_length(extension_idempotency_route_key) BETWEEN 1 AND 256
            AND octet_length(extension_idempotency_scope_key) BETWEEN 1 AND 512
            AND extension_normalized_request_sha256 ~ '^[0-9a-f]{64}$')
    );

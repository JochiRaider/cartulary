-- name: CreateBackupSet :one
INSERT INTO backup_sets (
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until,
    verification_state,
    last_verified_restore_at,
    last_verification_basis_sha256;

-- name: GetBackupSetByID :one
SELECT
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until,
    verification_state,
    last_verified_restore_at,
    last_verification_basis_sha256
FROM backup_sets
WHERE backup_set_id = $1;

-- name: GetLatestSuccessfulRetainedBackupSet :one
SELECT
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until,
    verification_state,
    last_verified_restore_at,
    last_verification_basis_sha256
FROM backup_sets
WHERE retained_until >= $1
ORDER BY consistency_point_at DESC, backup_set_id ASC
LIMIT 1;

-- name: ListSuccessfulRetainedBackupSets :many
SELECT
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until,
    verification_state,
    last_verified_restore_at,
    last_verification_basis_sha256
FROM backup_sets
WHERE retained_until >= $1
ORDER BY created_at ASC, backup_set_id ASC;

-- name: ListBackupSetsDueForRestoreVerification :many
SELECT
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until,
    verification_state,
    last_verified_restore_at,
    last_verification_basis_sha256
FROM backup_sets
WHERE retained_until >= $1
  AND (
      last_verified_restore_at IS NULL
      OR last_verified_restore_at <= ($1::timestamptz - interval '7 days')
      OR last_verification_basis_sha256 IS DISTINCT FROM $2
  )
ORDER BY consistency_point_at ASC, backup_set_id ASC;

-- name: UpdateBackupSetVerificationState :one
UPDATE backup_sets
SET
    verification_state = $2,
    last_verified_restore_at = $3,
    last_verification_basis_sha256 = $4
WHERE backup_set_id = $1
RETURNING
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until,
    verification_state,
    last_verified_restore_at,
    last_verification_basis_sha256;

-- name: CreateRestoreVerificationRun :one
INSERT INTO restore_verification_runs (
    restore_verification_run_id,
    backup_set_id,
    started_at,
    completed_at,
    verification_state,
    verification_basis_sha256,
    failure_reason,
    failure_message,
    authoritative_rows_sha256,
    authoritative_row_count,
    change_sets_sha256,
    change_set_row_count,
    blob_hashes_sha256,
    blob_count
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING
    restore_verification_run_id,
    backup_set_id,
    started_at,
    completed_at,
    verification_state,
    verification_basis_sha256,
    failure_reason,
    failure_message,
    authoritative_rows_sha256,
    authoritative_row_count,
    change_sets_sha256,
    change_set_row_count,
    blob_hashes_sha256,
    blob_count;

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
    last_verified_restore_at;

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
    last_verified_restore_at
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
    last_verified_restore_at
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
    last_verified_restore_at
FROM backup_sets
WHERE retained_until >= $1
ORDER BY created_at ASC, backup_set_id ASC;

-- name: UpdateBackupSetVerificationState :one
UPDATE backup_sets
SET
    verification_state = $2,
    last_verified_restore_at = $3
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
    last_verified_restore_at;

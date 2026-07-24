-- name: GetReportingSnapshot :one
SELECT snapshot_id, incident_id, created_by_user_id, client_txn_id, snapshot_at,
       source_change_set_high_watermark, derivation_version, export_model_sha256,
       source_boundary_json, export_model_json, create_job_id, created_at
  FROM reporting_snapshots
 WHERE snapshot_id = $1;

-- name: GetReportingSnapshotByCreateJob :one
SELECT snapshot_id, incident_id, created_by_user_id, client_txn_id, snapshot_at,
       source_change_set_high_watermark, derivation_version, export_model_sha256,
       source_boundary_json, export_model_json, create_job_id, created_at
  FROM reporting_snapshots
 WHERE create_job_id = $1;

-- name: CreateReportingSnapshot :one
INSERT INTO reporting_snapshots (
    snapshot_id, incident_id, created_by_user_id, client_txn_id, snapshot_at,
    source_change_set_high_watermark, derivation_version, export_model_sha256,
    source_boundary_json, export_model_json, create_job_id, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $5)
RETURNING snapshot_id, incident_id, created_by_user_id, client_txn_id, snapshot_at,
          source_change_set_high_watermark, derivation_version, export_model_sha256,
          source_boundary_json, export_model_json, create_job_id, created_at;

-- name: GetReportingRelease :one
SELECT *
  FROM reporting_releases
 WHERE release_id = $1;

-- name: GetReportingReleaseForUpdate :one
SELECT *
  FROM reporting_releases
 WHERE release_id = $1
 FOR UPDATE;

-- name: GetReportingReleaseByCreateJob :one
SELECT *
  FROM reporting_releases
 WHERE create_job_id = $1;

-- name: CreateReportingRelease :one
INSERT INTO reporting_releases (
    release_id, incident_id, snapshot_id, created_by_user_id, client_txn_id, release_scope, release_state,
    snapshot_at, source_change_set_high_watermark, derivation_version, export_model_sha256,
    template_id, template_version, redaction_profile_id, redaction_profile_version, redaction_profile_sha256,
    output_kind, output_options, graph_projection_refs, composition_id, composition_version,
    composition_sha256, render_admitted_at,
    output_media_type, output_sha256, redaction_manifest_sha256, redaction_manifest_json,
    create_job_id, recipient_partition_refs, approved_at, created_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14, $15, $16,
    $17, $18, $19, $20, $21,
    $22, $23, $24, $25, $26,
    $27, $28, $29, $30, $31, $31
)
RETURNING *;

-- name: CreateRenderFailedReportingRelease :one
INSERT INTO reporting_releases (
    incident_id, snapshot_id, created_by_user_id, client_txn_id, release_scope, release_state,
    snapshot_at, source_change_set_high_watermark, derivation_version, export_model_sha256,
    template_id, template_version, redaction_profile_id, redaction_profile_version, redaction_profile_sha256,
    output_kind, output_options, graph_projection_refs, composition_id, composition_version,
    composition_sha256, render_admitted_at,
    output_media_type, output_sha256, redaction_manifest_sha256, redaction_manifest_json,
    create_job_id, render_failed_reason_code, recipient_partition_refs, created_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, 'render_failed',
    $6, $7, $8, $9,
    $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19,
    $20, $21, NULL, NULL, NULL, NULL,
    $22, $23, $24, $25, $25
)
RETURNING *;

-- name: InsertReportingReleaseApproval :exec
INSERT INTO reporting_release_approvals (
    release_id, actor_user_id, approval_role, reason, approval_tuple_json,
    redaction_profile_sha256, output_sha256, redaction_manifest_sha256, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: ReportingReleaseApprovalExists :one
SELECT EXISTS (
    SELECT 1
      FROM reporting_release_approvals
     WHERE release_id = $1
       AND actor_user_id = $2
       AND approval_role = $3
);

-- name: ListReportingReleaseApprovals :many
SELECT approval_role, actor_user_id
  FROM reporting_release_approvals
 WHERE release_id = $1
 ORDER BY approval_role ASC, created_at ASC, approval_id ASC;

-- name: InvalidateSupersededReportingReleases :exec
UPDATE reporting_releases
   SET release_state = 'invalidated',
       invalidated_at = COALESCE(invalidated_at, $15),
       invalidation_reason = COALESCE(invalidation_reason, 'superseded_by_new_render'),
       updated_at = $15
 WHERE snapshot_id = $1
   AND output_kind = $2
   AND release_scope = $3
   AND template_id = $4
   AND template_version = $5
   AND redaction_profile_id = $6
   AND redaction_profile_version = $7
   AND recipient_partition_refs = $8
   AND output_options = $9
   AND graph_projection_refs = $10
   AND composition_id IS NOT DISTINCT FROM $11
   AND composition_version IS NOT DISTINCT FROM $12
   AND composition_sha256 IS NOT DISTINCT FROM $13
   AND release_id <> $14
   AND release_state IN ('pending_approval', 'approved', 'published');

-- name: UpdateReportingReleaseState :one
UPDATE reporting_releases
   SET release_state = $2,
       approved_at = CASE WHEN $2 = 'approved' THEN COALESCE(approved_at, $3) ELSE approved_at END,
       published_at = CASE WHEN $2 = 'published' THEN COALESCE(published_at, $3) ELSE published_at END,
       updated_at = $3
 WHERE release_id = $1
RETURNING *;

-- name: InvalidateReportingRelease :one
UPDATE reporting_releases
   SET release_state = 'invalidated',
       invalidated_at = COALESCE(invalidated_at, $2),
       invalidation_reason = COALESCE($3, invalidation_reason, 'explicit_invalidation'),
       updated_at = $2
 WHERE release_id = $1
RETURNING *;

-- name: CreateReportingJobPayload :exec
INSERT INTO reporting_job_payloads (
    job_id, job_kind, incident_id, actor_user_id, request_json, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $6);

-- name: GetReportingJobPayload :one
SELECT job_id, job_kind, incident_id, actor_user_id, request_json, created_at, updated_at
  FROM reporting_job_payloads
 WHERE job_id = $1;

-- name: CreateSavedView :one
INSERT INTO saved_views (
    incident_id,
    view_schema_id,
    scope,
    display_name,
    query_json,
    layout_json,
    owner_user_id,
    created_at,
    updated_at,
    saved_view_version
)
VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $8, 1)
RETURNING
    saved_view_id,
    incident_id,
    view_schema_id,
    scope,
    display_name,
    query_json,
    layout_json,
    owner_user_id,
    created_at,
    updated_at,
    saved_view_version;

-- name: ListVisibleSavedViews :many
SELECT
    sv.saved_view_id,
    sv.incident_id,
    sv.view_schema_id,
    sv.scope,
    sv.display_name,
    sv.query_json,
    sv.layout_json,
    sv.owner_user_id,
    sv.created_at,
    sv.updated_at,
    sv.saved_view_version
FROM saved_views sv
JOIN incident_memberships m
  ON m.incident_id = sv.incident_id
 AND m.user_id = $2
WHERE sv.incident_id = $1
  AND (
      sv.scope IN ('shared', 'system')
      OR sv.owner_user_id = $2
      OR m.role = 'admin'
  )
  AND ($3::timestamptz IS NULL OR sv.updated_at <= $3)
  AND ($4::timestamptz IS NULL OR $5::uuid IS NULL OR sv.updated_at < $4 OR (sv.updated_at = $4 AND sv.saved_view_id > $5))
ORDER BY sv.updated_at DESC, sv.saved_view_id ASC
LIMIT $6;

-- name: GetVisibleSavedViewForUpdate :one
SELECT
    sv.saved_view_id,
    sv.incident_id,
    sv.view_schema_id,
    sv.scope,
    sv.display_name,
    sv.query_json,
    sv.layout_json,
    sv.owner_user_id,
    sv.created_at,
    sv.updated_at,
    sv.saved_view_version
FROM saved_views sv
JOIN incident_memberships m
  ON m.incident_id = sv.incident_id
 AND m.user_id = $3
WHERE sv.incident_id = $1
  AND sv.saved_view_id = $2
  AND (
      sv.scope IN ('shared', 'system')
      OR sv.owner_user_id = $3
      OR m.role = 'admin'
  )
FOR UPDATE OF sv;

-- name: UpdateSavedView :one
UPDATE saved_views
SET
    scope = $3,
    display_name = $4,
    query_json = $5::jsonb,
    layout_json = $6::jsonb,
    updated_at = $7,
    saved_view_version = saved_view_version + 1
WHERE incident_id = $1
  AND saved_view_id = $2
RETURNING
    saved_view_id,
    incident_id,
    view_schema_id,
    scope,
    display_name,
    query_json,
    layout_json,
    owner_user_id,
    created_at,
    updated_at,
    saved_view_version;

-- name: DeleteSavedView :exec
DELETE FROM saved_views
WHERE incident_id = $1
  AND saved_view_id = $2;

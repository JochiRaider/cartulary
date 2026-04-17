-- name: GetSessionMemberships :many
SELECT
    incident_id,
    role
FROM incident_memberships
WHERE user_id = $1
ORDER BY incident_id ASC;

-- name: ListVisibleIncidents :many
SELECT
    i.id AS incident_id,
    i.incident_key,
    i.title,
    i.description,
    i.status,
    i.severity,
    i.tlp,
    i.current_phase,
    i.primary_external_case_ref,
    i.created_by_user_id,
    i.created_at,
    i.updated_at,
    i.updated_by_user_id,
    i.incident_version,
    i.closed_at
FROM incidents i
JOIN incident_memberships m
  ON m.incident_id = i.id
WHERE m.user_id = $1
  AND i.updated_at <= $2
  AND ($3::timestamptz IS NULL OR $4::uuid IS NULL OR i.updated_at < $3 OR (i.updated_at = $3 AND i.id > $4))
ORDER BY i.updated_at DESC, i.id ASC
LIMIT $5;

-- name: GetVisibleIncidentByID :one
SELECT
    i.id AS incident_id,
    i.incident_key,
    i.title,
    i.description,
    i.status,
    i.severity,
    i.tlp,
    i.current_phase,
    i.primary_external_case_ref,
    i.created_by_user_id,
    i.created_at,
    i.updated_at,
    i.updated_by_user_id,
    i.incident_version,
    i.closed_at
FROM incidents i
JOIN incident_memberships m
  ON m.incident_id = i.id
WHERE i.id = $1
  AND m.user_id = $2;

-- name: GetIncidentMembershipForActor :one
SELECT
    m.incident_id,
    m.user_id,
    u.display_name,
    m.role,
    m.joined_at,
    m.added_by_user_id,
    m.updated_at,
    m.updated_by_user_id,
    m.membership_version
FROM incident_memberships m
JOIN users u
  ON u.id = m.user_id
WHERE m.incident_id = $1
  AND m.user_id = $2;

-- name: ListIncidentMemberships :many
SELECT
    m.incident_id,
    m.user_id,
    u.display_name,
    m.role,
    m.joined_at,
    m.added_by_user_id,
    m.updated_at,
    m.updated_by_user_id,
    m.membership_version
FROM incident_memberships m
JOIN users u
  ON u.id = m.user_id
WHERE m.incident_id = $1
  AND m.joined_at <= $2
  AND ($3::timestamptz IS NULL OR $4::uuid IS NULL OR m.joined_at > $3 OR (m.joined_at = $3 AND m.user_id > $4))
ORDER BY m.joined_at ASC, m.user_id ASC
LIMIT $5;

-- name: GetVisibleIncidentDefaultWorkbookPreferences :one
SELECT
    p.incident_id,
    p.default_sheet_ref,
    p.created_at,
    p.updated_at,
    p.updated_by_user_id
FROM incident_workbook_preferences p
JOIN incident_memberships m
  ON m.incident_id = p.incident_id
WHERE p.incident_id = $1
  AND m.user_id = $2;

-- name: GetVisibleUserWorkbookPreferences :one
SELECT
    p.incident_id,
    p.user_id,
    p.home_sheet_ref,
    p.created_at,
    p.updated_at
FROM user_workbook_preferences p
JOIN incident_memberships m
  ON m.incident_id = p.incident_id
WHERE p.incident_id = $1
  AND p.user_id = $2
  AND m.user_id = $2;

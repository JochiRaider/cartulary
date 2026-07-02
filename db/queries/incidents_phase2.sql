-- name: GetSessionMemberships :many
SELECT
    incident_id,
    role
FROM incident_memberships
WHERE user_id = $1
ORDER BY incident_id ASC;

-- name: ListVisibleIncidents :many
SELECT
    i.id,
    i.incident_key,
    i.incident_key_canonical,
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
  AND ($2::timestamptz IS NULL OR i.updated_at <= $2)
  AND ($3::timestamptz IS NULL OR $4::uuid IS NULL OR i.updated_at < $3 OR (i.updated_at = $3 AND i.id > $4))
  AND ($6::boolean = false OR i.status = $7)
ORDER BY i.updated_at DESC, i.id ASC
LIMIT $5;

-- name: GetVisibleIncidentByID :one
SELECT
    i.id,
    i.incident_key,
    i.incident_key_canonical,
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

-- name: GetIncidentForUpdate :one
SELECT
    id,
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    created_at,
    updated_at,
    updated_by_user_id,
    incident_version,
    closed_at
FROM incidents
WHERE id = $1
FOR UPDATE;

-- name: EnsureIncidentOpenForUpdate :one
SELECT status
FROM incidents
WHERE id = $1
FOR UPDATE;

-- name: CreateIncident :one
INSERT INTO incidents (
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    created_at,
    updated_at,
    updated_by_user_id,
    incident_version
)
VALUES ($1, $2, $3, $4, 'active', $5, $6, $7, $8, $9, $10, $10, $9, 1)
RETURNING
    id,
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    created_at,
    updated_at,
    updated_by_user_id,
    incident_version,
    closed_at;

-- name: UpdateIncidentMetadata :one
UPDATE incidents
SET
    description = $2,
    severity = $3,
    tlp = $4,
    current_phase = $5,
    primary_external_case_ref = $6,
    updated_at = $7,
    updated_by_user_id = $8,
    incident_version = incident_version + 1
WHERE id = $1
RETURNING
    id,
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    created_at,
    updated_at,
    updated_by_user_id,
    incident_version,
    closed_at;

-- name: UpdateIncidentLifecycle :one
UPDATE incidents
SET
    status = $2,
    closed_at = $3,
    updated_at = $4,
    updated_by_user_id = $5,
    incident_version = incident_version + 1
WHERE id = $1
RETURNING
    id,
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    created_at,
    updated_at,
    updated_by_user_id,
    incident_version,
    closed_at;

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

-- name: GetIncidentMembershipForUpdate :one
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
  AND m.user_id = $2
FOR UPDATE;

-- name: ListAllIncidentMemberships :many
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
ORDER BY m.joined_at ASC, m.user_id ASC;

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

-- name: CreateBootstrapIncidentMembership :one
INSERT INTO incident_memberships (
    incident_id,
    user_id,
    role,
    joined_at,
    added_by_user_id,
    updated_at,
    updated_by_user_id,
    membership_version
)
VALUES ($1, $2, $4, $3, $2, $3, $2, 1)
RETURNING
    incident_id,
    user_id,
    $5::text AS display_name,
    role,
    joined_at,
    added_by_user_id,
    updated_at,
    updated_by_user_id,
    membership_version;

-- name: CreateIncidentMembership :one
INSERT INTO incident_memberships (
    incident_id,
    user_id,
    role,
    joined_at,
    added_by_user_id,
    updated_at,
    updated_by_user_id,
    membership_version
)
VALUES ($1, $2, $3, $4, $5, $4, $5, 1)
RETURNING
    incident_id,
    user_id,
    $6::text AS display_name,
    role,
    joined_at,
    added_by_user_id,
    updated_at,
    updated_by_user_id,
    membership_version;

-- name: UpdateIncidentMembershipRole :one
UPDATE incident_memberships
SET
    role = $3,
    updated_at = $4,
    updated_by_user_id = $5,
    membership_version = membership_version + 1
WHERE incident_id = $1
  AND user_id = $2
RETURNING
    incident_id,
    user_id,
    $6::text AS display_name,
    role,
    joined_at,
    added_by_user_id,
    updated_at,
    updated_by_user_id,
    membership_version;

-- name: DeleteIncidentMembership :exec
DELETE FROM incident_memberships
WHERE incident_id = $1
  AND user_id = $2;

-- name: CountIncidentAdmins :one
SELECT COUNT(*)
FROM incident_memberships
WHERE incident_id = $1
  AND role = 'admin';

-- name: InsertIncidentAuditEvent :exec
INSERT INTO deployment_admin_audit_events (
    actor_user_id,
    target_user_id,
    incident_id,
    event_source,
    event_kind,
    reason_code,
    client_txn_id,
    request_id,
    before_json,
    after_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb);

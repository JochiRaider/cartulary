-- name: GetLocalUserByEmail :one
SELECT
    id,
    email::text AS email,
    display_name,
    password_hash,
    password_changed_at,
    mfa_required,
    is_active,
    is_deployment_admin,
    created_at,
    updated_at,
    updated_by_user_id,
    last_login_at,
    user_version,
    totp_enrolled_at,
    totp_secret_ciphertext,
    totp_secret_nonce
FROM users
WHERE email = $1;

-- name: GetSessionByFingerprint :one
SELECT
    id,
    user_id,
    token_fingerprint,
    authenticated_at,
    last_qualifying_activity_at,
    idle_expires_at,
    absolute_expires_at,
    session_expires_at,
    revoked_at,
    revoke_reason_code,
    created_at,
    updated_at
FROM user_sessions
WHERE token_fingerprint = $1;

-- name: ListActiveSessionsForUser :many
SELECT
    id,
    user_id,
    token_fingerprint,
    authenticated_at,
    last_qualifying_activity_at,
    idle_expires_at,
    absolute_expires_at,
    session_expires_at,
    revoked_at,
    revoke_reason_code,
    created_at,
    updated_at
FROM user_sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND session_expires_at > $2
ORDER BY last_qualifying_activity_at ASC, authenticated_at ASC, id ASC;

-- name: GetBootstrapTokenByFingerprint :one
SELECT
    id,
    user_id,
    token_fingerprint,
    issued_at,
    expires_at,
    consumed_at,
    superseded_at,
    created_at
FROM bootstrap_tokens
WHERE token_fingerprint = $1;

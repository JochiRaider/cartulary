-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS updated_by_user_id uuid REFERENCES users (id),
    ADD COLUMN IF NOT EXISTS totp_enrolled_at timestamptz,
    ADD COLUMN IF NOT EXISTS totp_secret_ciphertext bytea,
    ADD COLUMN IF NOT EXISTS totp_secret_nonce bytea;

ALTER TABLE deployment_admin_audit_events
    ADD COLUMN IF NOT EXISTS reason_code text,
    ADD COLUMN IF NOT EXISTS client_txn_id text,
    ADD COLUMN IF NOT EXISTS request_id text;

CREATE TABLE IF NOT EXISTS user_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_fingerprint bytea NOT NULL UNIQUE,
    authenticated_at timestamptz NOT NULL,
    last_qualifying_activity_at timestamptz NOT NULL,
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    session_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revoke_reason_code text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS user_sessions_user_lookup_idx
    ON user_sessions (user_id, revoked_at, session_expires_at, last_qualifying_activity_at);

CREATE TABLE IF NOT EXISTS bootstrap_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_fingerprint bytea NOT NULL UNIQUE,
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    superseded_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS bootstrap_tokens_user_lookup_idx
    ON bootstrap_tokens (user_id, consumed_at, superseded_at, expires_at);

CREATE TABLE IF NOT EXISTS pending_totp_enrollments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    auth_scope_kind text NOT NULL CHECK (auth_scope_kind IN ('session', 'bootstrap_token')),
    auth_scope_session_id uuid REFERENCES user_sessions (id) ON DELETE CASCADE,
    auth_scope_bootstrap_token_id uuid REFERENCES bootstrap_tokens (id) ON DELETE CASCADE,
    client_txn_id text NOT NULL,
    secret_ciphertext bytea NOT NULL,
    secret_nonce bytea NOT NULL,
    replaces_active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    CONSTRAINT pending_totp_enrollments_auth_scope_ck CHECK (
        (auth_scope_kind = 'session' AND auth_scope_session_id IS NOT NULL AND auth_scope_bootstrap_token_id IS NULL) OR
        (auth_scope_kind = 'bootstrap_token' AND auth_scope_bootstrap_token_id IS NOT NULL AND auth_scope_session_id IS NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS pending_totp_enrollments_one_current_per_user_idx
    ON pending_totp_enrollments (user_id)
    WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS pending_totp_enrollments_scope_lookup_idx
    ON pending_totp_enrollments (auth_scope_kind, auth_scope_session_id, auth_scope_bootstrap_token_id, client_txn_id);

CREATE TABLE IF NOT EXISTS route_idempotency (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    route_key text NOT NULL,
    scope_key text NOT NULL,
    client_txn_id text NOT NULL,
    actor_user_id uuid NOT NULL REFERENCES users (id),
    target_user_id uuid REFERENCES users (id),
    request_hash bytea NOT NULL,
    status_code integer NOT NULL,
    response_json jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS route_idempotency_route_actor_scope_client_txn_idx
    ON route_idempotency (route_key, actor_user_id, scope_key, client_txn_id);

CREATE INDEX IF NOT EXISTS route_idempotency_actor_lookup_idx
    ON route_idempotency (actor_user_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS route_idempotency_actor_lookup_idx;
DROP INDEX IF EXISTS route_idempotency_route_actor_scope_client_txn_idx;
DROP TABLE IF EXISTS route_idempotency;

DROP INDEX IF EXISTS pending_totp_enrollments_scope_lookup_idx;
DROP INDEX IF EXISTS pending_totp_enrollments_one_current_per_user_idx;
DROP TABLE IF EXISTS pending_totp_enrollments;

DROP INDEX IF EXISTS bootstrap_tokens_user_lookup_idx;
DROP TABLE IF EXISTS bootstrap_tokens;

DROP INDEX IF EXISTS user_sessions_user_lookup_idx;
DROP TABLE IF EXISTS user_sessions;

ALTER TABLE deployment_admin_audit_events
    DROP COLUMN IF EXISTS request_id,
    DROP COLUMN IF EXISTS client_txn_id,
    DROP COLUMN IF EXISTS reason_code;

ALTER TABLE users
    DROP COLUMN IF EXISTS totp_secret_nonce,
    DROP COLUMN IF EXISTS totp_secret_ciphertext,
    DROP COLUMN IF EXISTS totp_enrolled_at,
    DROP COLUMN IF EXISTS updated_by_user_id;

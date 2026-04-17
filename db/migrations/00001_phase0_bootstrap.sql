-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email citext NOT NULL UNIQUE,
    display_name text NOT NULL CHECK (char_length(display_name) <= 256),
    password_hash text NOT NULL,
    password_changed_at timestamptz NOT NULL DEFAULT now(),
    mfa_required boolean NOT NULL DEFAULT true,
    is_active boolean NOT NULL DEFAULT true,
    is_deployment_admin boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz,
    user_version bigint NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS deployment_bootstrap_state (
    slot text PRIMARY KEY CHECK (slot = 'first_deployment_admin'),
    bootstrap_schema_id text NOT NULL CHECK (bootstrap_schema_id = 'cartulary.bootstrap_admin.v1'),
    bootstrap_artifact_id uuid NOT NULL UNIQUE,
    artifact_sha256 bytea NOT NULL,
    created_user_id uuid NOT NULL REFERENCES users (id),
    consumed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS deployment_admin_audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid REFERENCES users (id),
    target_user_id uuid REFERENCES users (id),
    event_source text NOT NULL,
    event_kind text NOT NULL,
    before_json jsonb,
    after_json jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS deployment_admin_audit_events;
DROP TABLE IF EXISTS deployment_bootstrap_state;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS pgcrypto;

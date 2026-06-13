-- +goose Up
CREATE TABLE IF NOT EXISTS enterprise_auth_providers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_key text NOT NULL UNIQUE CHECK (provider_key ~ '^[a-z0-9][a-z0-9._-]{1,126}[a-z0-9]$'),
    provider_type text NOT NULL CHECK (provider_type IN ('oidc', 'saml')),
    display_name text NOT NULL CHECK (char_length(display_name) BETWEEN 1 AND 256),
    is_enabled boolean NOT NULL DEFAULT true,
    is_interactive boolean NOT NULL DEFAULT true,
    authorization_endpoint text,
    issuer text,
    audience text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS enterprise_auth_providers_discovery_idx
    ON enterprise_auth_providers (display_name ASC, provider_key ASC)
    WHERE is_enabled = true AND is_interactive = true;

CREATE TABLE IF NOT EXISTS enterprise_auth_transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id uuid NOT NULL REFERENCES enterprise_auth_providers (id) ON DELETE CASCADE,
    provider_key text NOT NULL,
    provider_type text NOT NULL CHECK (provider_type IN ('oidc', 'saml')),
    return_to text NOT NULL,
    state text,
    nonce text,
    pkce_verifier_hash bytea,
    relay_state text,
    browser_binding_hash bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    CONSTRAINT enterprise_auth_transactions_correlation_ck CHECK (
        (provider_type = 'oidc' AND state IS NOT NULL AND nonce IS NOT NULL AND relay_state IS NULL) OR
        (provider_type = 'saml' AND relay_state IS NOT NULL AND state IS NULL AND nonce IS NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS enterprise_auth_transactions_state_idx
    ON enterprise_auth_transactions (state)
    WHERE state IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS enterprise_auth_transactions_relay_state_idx
    ON enterprise_auth_transactions (relay_state)
    WHERE relay_state IS NOT NULL;

CREATE INDEX IF NOT EXISTS enterprise_auth_transactions_provider_expiry_idx
    ON enterprise_auth_transactions (provider_id, expires_at, consumed_at);

CREATE TABLE IF NOT EXISTS enterprise_auth_bindings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider_id uuid NOT NULL REFERENCES enterprise_auth_providers (id) ON DELETE RESTRICT,
    provider_key text NOT NULL,
    provider_type text NOT NULL CHECK (provider_type IN ('oidc', 'saml')),
    provider_subject text NOT NULL CHECK (provider_subject <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    last_auth_at timestamptz,
    retired_at timestamptz,
    retired_by_user_id uuid REFERENCES users (id),
    retire_reason text,
    replaced_by_auth_binding_id uuid REFERENCES enterprise_auth_bindings (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS enterprise_auth_bindings_active_provider_subject_idx
    ON enterprise_auth_bindings (provider_id, provider_subject)
    WHERE retired_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS enterprise_auth_bindings_active_user_provider_idx
    ON enterprise_auth_bindings (user_id, provider_id)
    WHERE retired_at IS NULL;

CREATE INDEX IF NOT EXISTS enterprise_auth_bindings_user_active_idx
    ON enterprise_auth_bindings (user_id, provider_type ASC, provider_key ASC, created_at ASC)
    WHERE retired_at IS NULL;

ALTER TABLE user_sessions
    ADD COLUMN IF NOT EXISTS provider_type text NOT NULL DEFAULT 'local' CHECK (provider_type IN ('local', 'oidc', 'saml')),
    ADD COLUMN IF NOT EXISTS auth_binding_id uuid REFERENCES enterprise_auth_bindings (id);

-- +goose Down
ALTER TABLE user_sessions
    DROP COLUMN IF EXISTS auth_binding_id,
    DROP COLUMN IF EXISTS provider_type;

DROP INDEX IF EXISTS enterprise_auth_bindings_user_active_idx;
DROP INDEX IF EXISTS enterprise_auth_bindings_active_user_provider_idx;
DROP INDEX IF EXISTS enterprise_auth_bindings_active_provider_subject_idx;
DROP TABLE IF EXISTS enterprise_auth_bindings;

DROP INDEX IF EXISTS enterprise_auth_transactions_provider_expiry_idx;
DROP INDEX IF EXISTS enterprise_auth_transactions_relay_state_idx;
DROP INDEX IF EXISTS enterprise_auth_transactions_state_idx;
DROP TABLE IF EXISTS enterprise_auth_transactions;

DROP INDEX IF EXISTS enterprise_auth_providers_discovery_idx;
DROP TABLE IF EXISTS enterprise_auth_providers;

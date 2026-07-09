-- +goose Up

ALTER TABLE public.enterprise_auth_providers
    ADD COLUMN token_endpoint text,
    ADD COLUMN jwks_uri text,
    ADD COLUMN client_id text,
    ADD COLUMN client_secret_ref_kind text,
    ADD COLUMN client_secret_ref_name text,
    ADD COLUMN additional_scopes jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN saml_idp_entity_id text,
    ADD COLUMN saml_sso_url text,
    ADD COLUMN saml_idp_signing_certificates jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN saml_sp_entity_id text,
    ADD COLUMN saml_subject_source jsonb;

ALTER TABLE public.enterprise_auth_transactions
    ADD COLUMN pkce_verifier_ciphertext bytea,
    ADD COLUMN pkce_verifier_nonce bytea,
    ADD COLUMN saml_request_id text;

-- +goose Down

ALTER TABLE public.enterprise_auth_transactions
    DROP COLUMN IF EXISTS saml_request_id,
    DROP COLUMN IF EXISTS pkce_verifier_nonce,
    DROP COLUMN IF EXISTS pkce_verifier_ciphertext;

ALTER TABLE public.enterprise_auth_providers
    DROP COLUMN IF EXISTS saml_subject_source,
    DROP COLUMN IF EXISTS saml_sp_entity_id,
    DROP COLUMN IF EXISTS saml_idp_signing_certificates,
    DROP COLUMN IF EXISTS saml_sso_url,
    DROP COLUMN IF EXISTS saml_idp_entity_id,
    DROP COLUMN IF EXISTS additional_scopes,
    DROP COLUMN IF EXISTS client_secret_ref_name,
    DROP COLUMN IF EXISTS client_secret_ref_kind,
    DROP COLUMN IF EXISTS client_id,
    DROP COLUMN IF EXISTS jwks_uri,
    DROP COLUMN IF EXISTS token_endpoint;

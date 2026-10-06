-- +goose Up
-- This release admits fresh deployments only. Never silently discard old secrets.
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.users WHERE totp_secret_ciphertext IS NOT NULL OR totp_secret_nonce IS NOT NULL)
       OR EXISTS (SELECT 1 FROM public.pending_totp_enrollments)
       OR EXISTS (SELECT 1 FROM public.enterprise_auth_transactions WHERE pkce_verifier_ciphertext IS NOT NULL OR pkce_verifier_nonce IS NOT NULL) THEN
        RAISE EXCEPTION 'incompatible retained authentication secrets';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.users
    DROP COLUMN totp_secret_ciphertext, DROP COLUMN totp_secret_nonce,
    ADD COLUMN totp_secret_envelope bytea,
    ADD CONSTRAINT users_totp_secret_envelope_check CHECK (
        (totp_enrolled_at IS NULL AND totp_secret_envelope IS NULL) OR
        (totp_enrolled_at IS NOT NULL AND totp_secret_envelope IS NOT NULL AND
         octet_length(totp_secret_envelope) = 93 AND get_byte(totp_secret_envelope, 0) = 1));
ALTER TABLE public.pending_totp_enrollments
    DROP COLUMN secret_ciphertext, DROP COLUMN secret_nonce,
    ADD COLUMN secret_envelope bytea NOT NULL,
    ADD CONSTRAINT pending_totp_secret_envelope_check CHECK (
        octet_length(secret_envelope) = 93 AND get_byte(secret_envelope, 0) = 1);
ALTER TABLE public.enterprise_auth_transactions
    DROP COLUMN pkce_verifier_ciphertext, DROP COLUMN pkce_verifier_nonce,
    ADD COLUMN pkce_verifier_envelope bytea,
    ADD CONSTRAINT enterprise_pkce_envelope_check CHECK (
        (provider_type = 'saml' AND pkce_verifier_envelope IS NULL) OR
        (provider_type = 'oidc' AND pkce_verifier_envelope IS NOT NULL AND
         octet_length(pkce_verifier_envelope) = 104 AND get_byte(pkce_verifier_envelope, 0) = 1));

-- +goose Down
-- A historical release is required to operate historical formats.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'authentication secret envelope migration is irreversible'; END $$;
-- +goose StatementEnd

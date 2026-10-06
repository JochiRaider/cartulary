-- +goose Up
-- The singleton is schema metadata. Table SELECT grants do not require PUBLIC
-- access to its implicit composite type.
REVOKE USAGE ON TYPE public.application_crypto_format FROM PUBLIC;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'cryptographic format metadata privileges cannot be weakened by downgrade' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd

-- +goose Up
ALTER TABLE enterprise_auth_transactions
    ADD COLUMN IF NOT EXISTS saml_completion_hash bytea,
    ADD COLUMN IF NOT EXISTS saml_subject text,
    ADD COLUMN IF NOT EXISTS saml_staged_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS enterprise_auth_transactions_saml_completion_hash_idx
    ON enterprise_auth_transactions (saml_completion_hash)
    WHERE saml_completion_hash IS NOT NULL;

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conname = 'enterprise_auth_transactions_saml_staging_ck'
    ) THEN
        ALTER TABLE enterprise_auth_transactions
            ADD CONSTRAINT enterprise_auth_transactions_saml_staging_ck CHECK (
                (
                    provider_type <> 'saml'
                    AND saml_completion_hash IS NULL
                    AND saml_subject IS NULL
                    AND saml_staged_at IS NULL
                )
                OR
                (
                    provider_type = 'saml'
                    AND (
                        (
                            saml_completion_hash IS NULL
                            AND saml_subject IS NULL
                            AND saml_staged_at IS NULL
                        )
                        OR
                        (
                            saml_completion_hash IS NOT NULL
                            AND saml_subject IS NOT NULL
                            AND saml_subject <> ''
                            AND saml_staged_at IS NOT NULL
                        )
                    )
                )
            );
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE enterprise_auth_transactions
    DROP CONSTRAINT IF EXISTS enterprise_auth_transactions_saml_staging_ck;

DROP INDEX IF EXISTS enterprise_auth_transactions_saml_completion_hash_idx;

ALTER TABLE enterprise_auth_transactions
    DROP COLUMN IF EXISTS saml_staged_at,
    DROP COLUMN IF EXISTS saml_subject,
    DROP COLUMN IF EXISTS saml_completion_hash;

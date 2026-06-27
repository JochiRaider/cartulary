-- +goose Up
ALTER TABLE hosts
    ADD COLUMN IF NOT EXISTS location text,
    ADD COLUMN IF NOT EXISTS os_platform text,
    ADD COLUMN IF NOT EXISTS business_owner text,
    ADD COLUMN IF NOT EXISTS criticality text,
    ADD COLUMN IF NOT EXISTS containment_status text;

ALTER TABLE identities
    ADD COLUMN IF NOT EXISTS privilege_level text,
    ADD COLUMN IF NOT EXISTS mfa_state text,
    ADD COLUMN IF NOT EXISTS reset_status text;

-- +goose Down
ALTER TABLE identities
    DROP COLUMN IF EXISTS reset_status,
    DROP COLUMN IF EXISTS mfa_state,
    DROP COLUMN IF EXISTS privilege_level;

ALTER TABLE hosts
    DROP COLUMN IF EXISTS containment_status,
    DROP COLUMN IF EXISTS criticality,
    DROP COLUMN IF EXISTS business_owner,
    DROP COLUMN IF EXISTS os_platform,
    DROP COLUMN IF EXISTS location;

-- +goose Up
ALTER TABLE reference_pack_job_payloads
    ADD COLUMN bundle_staging_path text;

-- +goose Down
ALTER TABLE reference_pack_job_payloads
    DROP COLUMN IF EXISTS bundle_staging_path;

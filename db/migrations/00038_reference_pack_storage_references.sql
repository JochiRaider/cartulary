-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM reference_packs LIMIT 1)
       OR EXISTS (SELECT 1 FROM reference_pack_job_payloads LIMIT 1) THEN
        RAISE EXCEPTION 'development database reset required; run CARTULARY_DESTRUCTIVE_CONFIRM=db-reset make db-reset and reseed development data';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE reference_packs
    RENAME COLUMN bundle_storage_path TO bundle_storage_ref;

ALTER TABLE reference_packs
    ADD CONSTRAINT reference_packs_bundle_storage_ref_relative_check CHECK (
        bundle_storage_ref <> ''
        AND bundle_storage_ref !~ '^/'
        AND strpos(bundle_storage_ref, E'\\') = 0
        AND bundle_storage_ref !~ '(^|/)\.\.?(/|$)'
        AND bundle_storage_ref !~ '//'
        AND bundle_storage_ref !~ '/$'
    );

ALTER TABLE reference_pack_job_payloads
    RENAME COLUMN bundle_staging_path TO bundle_staging_ref;

ALTER TABLE reference_pack_job_payloads
    ADD CONSTRAINT reference_pack_job_payloads_bundle_staging_ref_relative_check CHECK (
        bundle_staging_ref IS NULL
        OR (
            bundle_staging_ref <> ''
            AND bundle_staging_ref !~ '^/'
            AND strpos(bundle_staging_ref, E'\\') = 0
            AND bundle_staging_ref !~ '(^|/)\.\.?(/|$)'
            AND bundle_staging_ref !~ '//'
            AND bundle_staging_ref !~ '/$'
        )
    );

-- +goose Down
ALTER TABLE reference_pack_job_payloads
    DROP CONSTRAINT reference_pack_job_payloads_bundle_staging_ref_relative_check;

ALTER TABLE reference_pack_job_payloads
    RENAME COLUMN bundle_staging_ref TO bundle_staging_path;

ALTER TABLE reference_packs
    DROP CONSTRAINT reference_packs_bundle_storage_ref_relative_check;

ALTER TABLE reference_packs
    RENAME COLUMN bundle_storage_ref TO bundle_storage_path;

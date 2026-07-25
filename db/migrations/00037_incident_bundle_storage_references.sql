-- +goose Up

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.incident_bundle_exports LIMIT 1)
       OR EXISTS (SELECT 1 FROM public.incident_bundle_job_payloads LIMIT 1) THEN
        RAISE EXCEPTION
            'incident bundle storage-reference cutover: development database reset required; run CARTULARY_DESTRUCTIVE_CONFIRM=db-reset make db-reset and reseed development data';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.incident_bundle_exports
    RENAME COLUMN bundle_storage_path TO bundle_storage_ref;

ALTER TABLE public.incident_bundle_exports
    ADD CONSTRAINT incident_bundle_exports_storage_ref_ck CHECK (
        bundle_storage_ref <> ''
        AND bundle_storage_ref !~ '^/'
        AND strpos(bundle_storage_ref, E'\\') = 0
        AND bundle_storage_ref !~ '(^|/)\.\.?(/|$)'
        AND bundle_storage_ref !~ '//'
        AND bundle_storage_ref !~ '/$'
    );

ALTER TABLE public.incident_bundle_job_payloads
    RENAME COLUMN bundle_staging_path TO bundle_staging_ref;

ALTER TABLE public.incident_bundle_job_payloads
    ADD CONSTRAINT incident_bundle_job_payloads_staging_ref_ck CHECK (
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

ALTER TABLE public.incident_bundle_job_payloads
    DROP CONSTRAINT incident_bundle_job_payloads_staging_ref_ck;

ALTER TABLE public.incident_bundle_job_payloads
    RENAME COLUMN bundle_staging_ref TO bundle_staging_path;

ALTER TABLE public.incident_bundle_exports
    DROP CONSTRAINT incident_bundle_exports_storage_ref_ck;

ALTER TABLE public.incident_bundle_exports
    RENAME COLUMN bundle_storage_ref TO bundle_storage_path;

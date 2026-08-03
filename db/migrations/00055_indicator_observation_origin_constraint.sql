-- +goose Up

LOCK TABLE public.indicator_observations IN ACCESS EXCLUSIVE MODE;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.indicator_observations
         WHERE origin_kind NOT IN (
             'manual_entry',
             'clipboard_paste',
             'csv_import',
             'xlsx_import',
             'api_import',
             'extraction',
             'system'
         )
    ) THEN
        RAISE EXCEPTION 'indicator observation origin migration blocked by unclassified persisted provenance';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.indicator_observations
    ADD CONSTRAINT indicator_observations_origin_kind_ck
    CHECK (origin_kind IN (
        'manual_entry',
        'clipboard_paste',
        'csv_import',
        'xlsx_import',
        'api_import',
        'extraction',
        'system'
    ));

-- +goose Down

ALTER TABLE public.indicator_observations
    DROP CONSTRAINT IF EXISTS indicator_observations_origin_kind_ck;

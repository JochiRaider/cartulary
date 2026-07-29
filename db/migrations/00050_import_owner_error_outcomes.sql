-- +goose Up

-- Public decoders have never admitted use_null. Abort rather than silently
-- interpreting manually inserted or pre-contract retained state.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.import_units
         WHERE approved_mapping_json IS NOT NULL
           AND jsonb_path_exists(
               approved_mapping_json,
               '$.source_columns[*] ? (@.empty_value_policy == "use_null")'
           )
    ) THEN
        RAISE EXCEPTION
            'migration 00050 rejects retained import mappings with legacy use_null; repair to write_null and recompute the mapping fingerprint before retrying';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.import_unit_apply_outcomes
    ADD COLUMN error_retryable boolean DEFAULT false NOT NULL,
    ADD COLUMN error_details_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    DROP CONSTRAINT import_unit_apply_outcomes_error_shape_ck;

UPDATE public.import_unit_apply_outcomes
   SET error_details_json = jsonb_build_object('reason_code', reason_code)
 WHERE outcome_status IN ('failed', 'canceled');

ALTER TABLE public.import_unit_apply_outcomes
    ADD CONSTRAINT import_unit_apply_outcomes_error_details_ck
        CHECK (jsonb_typeof(error_details_json) = 'object'),
    ADD CONSTRAINT import_unit_apply_outcomes_error_shape_ck
        CHECK (
            (
                outcome_status = 'applied'
                AND error_code IS NULL
                AND reason_code IS NULL
                AND error_retryable = false
                AND error_details_json = '{}'::jsonb
            )
            OR
            (
                outcome_status IN ('failed', 'canceled')
                AND error_code IS NOT NULL
                AND reason_code IS NOT NULL
                AND error_details_json->>'reason_code' = reason_code
            )
        );

-- +goose Down

ALTER TABLE public.import_unit_apply_outcomes
    DROP CONSTRAINT import_unit_apply_outcomes_error_shape_ck,
    DROP CONSTRAINT import_unit_apply_outcomes_error_details_ck,
    DROP COLUMN error_details_json,
    DROP COLUMN error_retryable,
    ADD CONSTRAINT import_unit_apply_outcomes_error_shape_ck
        CHECK (
            (outcome_status = 'applied' AND error_code IS NULL AND reason_code IS NULL)
            OR
            (outcome_status IN ('failed', 'canceled') AND error_code IS NOT NULL)
        );

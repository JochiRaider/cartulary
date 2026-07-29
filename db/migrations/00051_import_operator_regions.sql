-- +goose Up

ALTER TABLE public.import_units
    ADD COLUMN base_import_unit_id uuid,
    ADD COLUMN operator_region_sequence integer,
    ADD COLUMN blocking_source_column_ordinals integer[] DEFAULT '{}'::integer[] NOT NULL,
    ADD CONSTRAINT import_units_base_import_unit_id_fkey
        FOREIGN KEY (base_import_unit_id)
        REFERENCES public.import_units(import_unit_id)
        ON DELETE CASCADE,
    ADD CONSTRAINT import_units_operator_region_shape_ck
        CHECK (
            (
                locator_kind = 'operator_region'
                AND base_import_unit_id IS NOT NULL
                AND operator_region_sequence IS NOT NULL
                AND operator_region_sequence > 0
            )
            OR
            (
                locator_kind <> 'operator_region'
                AND base_import_unit_id IS NULL
                AND operator_region_sequence IS NULL
            )
        ),
    ADD CONSTRAINT import_units_blocking_source_columns_ck
        CHECK (
            array_position(blocking_source_column_ordinals, 0) IS NULL
        );

CREATE UNIQUE INDEX import_units_operator_region_sequence_uq
    ON public.import_units (import_session_id, operator_region_sequence)
    WHERE locator_kind = 'operator_region';

CREATE UNIQUE INDEX import_units_operator_region_identity_uq
    ON public.import_units (base_import_unit_id, source_rect_a1)
    WHERE locator_kind = 'operator_region';

-- +goose Down

DROP INDEX IF EXISTS public.import_units_operator_region_identity_uq;
DROP INDEX IF EXISTS public.import_units_operator_region_sequence_uq;

ALTER TABLE public.import_units
    DROP CONSTRAINT import_units_blocking_source_columns_ck,
    DROP CONSTRAINT import_units_operator_region_shape_ck,
    DROP CONSTRAINT import_units_base_import_unit_id_fkey,
    DROP COLUMN blocking_source_column_ordinals,
    DROP COLUMN operator_region_sequence,
    DROP COLUMN base_import_unit_id;

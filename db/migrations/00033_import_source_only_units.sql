-- +goose Up
-- Fresh pre-production environment; CSV source units have no generic
-- interpretation until approval. Predecessor plan identities are unsupported.
ALTER TABLE public.import_units
    ALTER COLUMN source_rect_a1 DROP NOT NULL,
    ALTER COLUMN header_row_ref DROP NOT NULL,
    ALTER COLUMN data_start_row_ref DROP NOT NULL,
    ALTER COLUMN inferred_row_count DROP NOT NULL,
    ALTER COLUMN inferred_column_count DROP NOT NULL,
    ALTER COLUMN columns_json DROP NOT NULL,
    ALTER COLUMN source_rows_json DROP NOT NULL,
    ALTER COLUMN preview_rows_json DROP NOT NULL,
    ALTER COLUMN columns_json DROP DEFAULT,
    ALTER COLUMN source_rows_json DROP DEFAULT,
    ALTER COLUMN preview_rows_json DROP DEFAULT;
ALTER TABLE public.import_units ADD CONSTRAINT import_units_projection_shape_ck CHECK (
    (locator_kind='csv_file' AND num_nonnulls(source_rect_a1,header_row_ref,data_start_row_ref,
        inferred_row_count,inferred_column_count,columns_json,source_rows_json,preview_rows_json)=0)
    OR num_nonnulls(source_rect_a1,header_row_ref,data_start_row_ref,
        inferred_row_count,inferred_column_count,columns_json,source_rows_json,preview_rows_json)=8
);
ALTER TABLE public.import_apply_unit_plans ALTER COLUMN source_rect_a1 DROP NOT NULL;
ALTER TABLE public.import_apply_unit_plans ADD CONSTRAINT import_apply_unit_plans_rectangle_ck CHECK (
    (target_kind='view_schema' AND source_rect_a1 IS NOT NULL)
    OR (target_kind='network_flow_table' AND source_rect_a1 IS NULL)
);

-- +goose Down
-- Populated analytical plans or source-only units cannot be downgraded. Restore
-- the matching environment snapshot instead of synthesizing interpretation.
ALTER TABLE public.import_apply_unit_plans ALTER COLUMN source_rect_a1 SET NOT NULL;
ALTER TABLE public.import_apply_unit_plans DROP CONSTRAINT import_apply_unit_plans_rectangle_ck;
ALTER TABLE public.import_units
    ALTER COLUMN source_rect_a1 SET NOT NULL,
    ALTER COLUMN header_row_ref SET NOT NULL,
    ALTER COLUMN data_start_row_ref SET NOT NULL,
    ALTER COLUMN inferred_row_count SET NOT NULL,
    ALTER COLUMN inferred_column_count SET NOT NULL,
    ALTER COLUMN columns_json SET NOT NULL,
    ALTER COLUMN source_rows_json SET NOT NULL,
    ALTER COLUMN preview_rows_json SET NOT NULL,
    ALTER COLUMN columns_json SET DEFAULT '[]'::jsonb,
    ALTER COLUMN source_rows_json SET DEFAULT '[]'::jsonb,
    ALTER COLUMN preview_rows_json SET DEFAULT '[]'::jsonb;
ALTER TABLE public.import_units DROP CONSTRAINT import_units_projection_shape_ck;

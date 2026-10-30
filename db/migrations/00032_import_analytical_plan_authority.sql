-- +goose Up
-- Fresh pre-production environment; predecessor plans are not converted.
ALTER TABLE public.import_apply_unit_plans ALTER COLUMN source_rows_sha256 DROP NOT NULL;
ALTER TABLE public.import_apply_unit_plans DROP CONSTRAINT import_apply_unit_plans_digest_ck;
ALTER TABLE public.import_apply_unit_plans ADD CONSTRAINT import_apply_unit_plans_digest_ck
CHECK (source_content_sha256 ~ '^[0-9a-f]{64}$' AND mapping_fingerprint ~ '^[0-9a-f]{64}$'
AND approved_mapping_sha256 ~ '^[0-9a-f]{64}$' AND target_registry_sha256 ~ '^[0-9a-f]{64}$');
ALTER TABLE public.import_apply_unit_plans ADD CONSTRAINT import_apply_unit_plans_source_authority_ck
CHECK ((target_kind='view_schema' AND source_rows_sha256 IS NOT NULL AND source_rows_sha256 ~ '^[0-9a-f]{64}$')
OR (target_kind='network_flow_table' AND source_rows_sha256 IS NULL));

-- +goose Down
-- Analytical plans cannot be downgraded: restore the matching application and
-- database snapshot instead of synthesizing obsolete metadata.
ALTER TABLE public.import_apply_unit_plans ALTER COLUMN source_rows_sha256 SET NOT NULL;
ALTER TABLE public.import_apply_unit_plans DROP CONSTRAINT import_apply_unit_plans_source_authority_ck;
ALTER TABLE public.import_apply_unit_plans DROP CONSTRAINT import_apply_unit_plans_digest_ck;
ALTER TABLE public.import_apply_unit_plans ADD CONSTRAINT import_apply_unit_plans_digest_ck
CHECK (source_content_sha256 ~ '^[0-9a-f]{64}$' AND source_rows_sha256 ~ '^[0-9a-f]{64}$'
AND mapping_fingerprint ~ '^[0-9a-f]{64}$' AND approved_mapping_sha256 ~ '^[0-9a-f]{64}$'
AND target_registry_sha256 ~ '^[0-9a-f]{64}$');

-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS public.schema_migration_lineage (
    lineage_id text PRIMARY KEY,
    applied_at timestamp with time zone DEFAULT now() NOT NULL,
    description text NOT NULL
);

INSERT INTO public.schema_migration_lineage (lineage_id, description)
VALUES (
    'cartulary.prod_ddl_rebaseline.v1',
    'Clean production DDL rebaseline replacing historical phase-shaped migration line.'
)
ON CONFLICT (lineage_id) DO NOTHING;

-- +goose Down
DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS citext;
DROP TABLE IF EXISTS public.schema_migration_lineage;

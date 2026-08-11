-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    pgcrypto_version text;
    pgcrypto_schema text;
    citext_version text;
    citext_schema text;
BEGIN
    SELECT extension.extversion, namespace.nspname
    INTO pgcrypto_version, pgcrypto_schema
    FROM pg_catalog.pg_extension AS extension
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = extension.extnamespace
    WHERE extension.extname = 'pgcrypto';

    SELECT extension.extversion, namespace.nspname
    INTO citext_version, citext_schema
    FROM pg_catalog.pg_extension AS extension
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = extension.extnamespace
    WHERE extension.extname = 'citext';

    IF pgcrypto_version IS DISTINCT FROM '1.3'
       OR pgcrypto_schema IS DISTINCT FROM 'public'
       OR citext_version IS DISTINCT FROM '1.6'
       OR citext_schema IS DISTINCT FROM 'public' THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'schema_extension_prerequisite_invalid';
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE public.schema_migration_lineage (
    lineage_id text CONSTRAINT schema_migration_lineage_pkey PRIMARY KEY,
    applied_at timestamp with time zone DEFAULT now() NOT NULL,
    description text NOT NULL
);

INSERT INTO public.schema_migration_lineage (lineage_id, description)
VALUES (
    'cartulary.prod_ddl_rebaseline.v2',
    'Production DDL Rebaseline v2 final-state catalog.'
);

-- +goose Down
DROP TABLE public.schema_migration_lineage;

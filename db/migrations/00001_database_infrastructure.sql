-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: database infrastructure.
-- Administrator provisioning and final admission are separate lifecycle steps.

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

-- Future objects retain the same default-deny boundary as this baseline.
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public REVOKE ALL ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public REVOKE ALL ON SEQUENCES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner REVOKE USAGE ON TYPES FROM PUBLIC;
REVOKE USAGE ON TYPE public.goose_db_version FROM PUBLIC;

CREATE TABLE public.application_crypto_format (
    singleton boolean NOT NULL,
    format_id text NOT NULL,
    CONSTRAINT application_crypto_format_pkey PRIMARY KEY (singleton),
    CONSTRAINT application_crypto_format_singleton_check CHECK (singleton)
);

REVOKE ALL ON TABLE public.application_crypto_format FROM PUBLIC;

REVOKE ALL ON TYPE public.application_crypto_format FROM PUBLIC;

GRANT SELECT ON TABLE public.application_crypto_format TO cartulary_runtime;

GRANT SELECT ON TABLE public.application_crypto_format TO cartulary_recovery;

CREATE TABLE public.schema_migration_lineage (
    lineage_id text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL,
    description text NOT NULL,
    CONSTRAINT schema_migration_lineage_pkey PRIMARY KEY (lineage_id)
);

REVOKE ALL ON TABLE public.schema_migration_lineage FROM PUBLIC;

REVOKE ALL ON TYPE public.schema_migration_lineage FROM PUBLIC;

GRANT SELECT ON TABLE public.schema_migration_lineage TO cartulary_runtime;

GRANT SELECT ON TABLE public.schema_migration_lineage TO cartulary_recovery;

INSERT INTO public.schema_migration_lineage (lineage_id, description) VALUES ('cartulary.prod_ddl_rebaseline.v3', 'PostgreSQL 18.6 consolidated owner catalog.');

GRANT SELECT ON TABLE public.goose_db_version TO cartulary_runtime, cartulary_recovery;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TABLE public.schema_migration_lineage;
DROP TABLE public.application_crypto_format;

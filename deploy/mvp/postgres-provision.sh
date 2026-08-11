#!/bin/sh
set -eu

: "${CARTULARY_POSTGRES_MIGRATION_PASSWORD:?set CARTULARY_POSTGRES_MIGRATION_PASSWORD}"
: "${CARTULARY_POSTGRES_RUNTIME_PASSWORD:?set CARTULARY_POSTGRES_RUNTIME_PASSWORD}"
: "${CARTULARY_POSTGRES_RECOVERY_PASSWORD:?set CARTULARY_POSTGRES_RECOVERY_PASSWORD}"

target_database="${PGDATABASE:-${POSTGRES_DB:-cartulary}}"

psql \
  --username "${POSTGRES_USER:-cartulary}" \
  --dbname "$target_database" \
  --set ON_ERROR_STOP=on \
  --set migration_password="$CARTULARY_POSTGRES_MIGRATION_PASSWORD" \
  --set runtime_password="$CARTULARY_POSTGRES_RUNTIME_PASSWORD" \
  --set recovery_password="$CARTULARY_POSTGRES_RECOVERY_PASSWORD" <<'SQL'
SELECT format('CREATE ROLE cartulary_schema_owner NOLOGIN')
 WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_schema_owner')
\gexec
SELECT format('CREATE ROLE cartulary_runtime NOLOGIN')
 WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_runtime')
\gexec
SELECT format('CREATE ROLE cartulary_recovery NOLOGIN')
 WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_recovery')
\gexec
SELECT format('CREATE ROLE cartulary_migration_login LOGIN PASSWORD %L', :'migration_password')
 WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_migration_login')
\gexec
SELECT format('CREATE ROLE cartulary_runtime_login LOGIN PASSWORD %L', :'runtime_password')
 WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_runtime_login')
\gexec
SELECT format('CREATE ROLE cartulary_recovery_login LOGIN PASSWORD %L', :'recovery_password')
 WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_recovery_login')
\gexec

ALTER ROLE cartulary_schema_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
ALTER ROLE cartulary_runtime NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
ALTER ROLE cartulary_recovery NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SELECT format('ALTER ROLE cartulary_migration_login LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD %L', :'migration_password') \gexec
SELECT format('ALTER ROLE cartulary_runtime_login LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD %L', :'runtime_password') \gexec
SELECT format('ALTER ROLE cartulary_recovery_login LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD %L', :'recovery_password') \gexec

GRANT cartulary_schema_owner TO cartulary_migration_login WITH INHERIT FALSE, SET TRUE, ADMIN FALSE;
GRANT cartulary_runtime TO cartulary_runtime_login WITH INHERIT FALSE, SET TRUE, ADMIN FALSE;
GRANT cartulary_recovery TO cartulary_recovery_login WITH INHERIT FALSE, SET TRUE, ADMIN FALSE;
REVOKE cartulary_runtime, cartulary_recovery FROM cartulary_migration_login;
REVOKE cartulary_schema_owner, cartulary_recovery FROM cartulary_runtime_login;
REVOKE cartulary_schema_owner, cartulary_runtime FROM cartulary_recovery_login;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public VERSION '1.3';
CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public VERSION '1.6';
ALTER SCHEMA public OWNER TO cartulary_schema_owner;
REVOKE CREATE, USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO cartulary_schema_owner;
GRANT USAGE ON SCHEMA public TO cartulary_runtime, cartulary_recovery;
SELECT format('REVOKE CONNECT, TEMPORARY ON DATABASE %I FROM PUBLIC', current_database()) \gexec
SELECT format('GRANT CONNECT ON DATABASE %I TO cartulary_migration_login, cartulary_runtime_login, cartulary_recovery_login', current_database()) \gexec
GRANT SET ON PARAMETER session_replication_role TO cartulary_recovery;

SELECT format('REVOKE EXECUTE ON FUNCTION %s FROM PUBLIC', procedure.oid::pg_catalog.regprocedure)
  FROM pg_catalog.pg_proc AS procedure
  JOIN pg_catalog.pg_depend AS dependency
    ON dependency.classid = 'pg_catalog.pg_proc'::pg_catalog.regclass
   AND dependency.objid = procedure.oid
   AND dependency.deptype = 'e'
  JOIN pg_catalog.pg_extension AS extension ON extension.oid = dependency.refobjid
 WHERE extension.extname IN ('pgcrypto', 'citext')
 ORDER BY procedure.oid::pg_catalog.regprocedure::text
\gexec

SELECT format('REVOKE USAGE ON TYPE %s FROM PUBLIC', extension_type.oid::pg_catalog.regtype)
  FROM pg_catalog.pg_type AS extension_type
  JOIN pg_catalog.pg_depend AS dependency
    ON dependency.classid = 'pg_catalog.pg_type'::pg_catalog.regclass
   AND dependency.objid = extension_type.oid
   AND dependency.deptype = 'e'
  JOIN pg_catalog.pg_extension AS extension ON extension.oid = dependency.refobjid
 WHERE extension.extname IN ('pgcrypto', 'citext')
 ORDER BY extension_type.oid::pg_catalog.regtype::text
\gexec
SELECT format('GRANT USAGE ON TYPE %s TO cartulary_schema_owner, cartulary_runtime, cartulary_recovery', extension_type.oid::pg_catalog.regtype)
  FROM pg_catalog.pg_type AS extension_type
  JOIN pg_catalog.pg_depend AS dependency
    ON dependency.classid = 'pg_catalog.pg_type'::pg_catalog.regclass
   AND dependency.objid = extension_type.oid
   AND dependency.deptype = 'e'
  JOIN pg_catalog.pg_extension AS extension ON extension.oid = dependency.refobjid
 WHERE extension.extname IN ('pgcrypto', 'citext')
 ORDER BY extension_type.oid::pg_catalog.regtype::text
\gexec
SELECT format('GRANT EXECUTE ON FUNCTION %s TO cartulary_schema_owner, cartulary_runtime, cartulary_recovery', procedure.oid::pg_catalog.regprocedure)
  FROM pg_catalog.pg_proc AS procedure
  JOIN pg_catalog.pg_depend AS dependency
    ON dependency.classid = 'pg_catalog.pg_proc'::pg_catalog.regclass
   AND dependency.objid = procedure.oid
   AND dependency.deptype = 'e'
  JOIN pg_catalog.pg_extension AS extension ON extension.oid = dependency.refobjid
 WHERE extension.extname IN ('pgcrypto', 'citext')
 ORDER BY procedure.oid::pg_catalog.regprocedure::text
\gexec

ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public REVOKE ALL ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public REVOKE ALL ON SEQUENCES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner REVOKE USAGE ON TYPES FROM PUBLIC;
SQL

package pgtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
)

const clusterPrerequisitesSQL = `
DO $cartulary_roles$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_schema_owner') THEN
        CREATE ROLE cartulary_schema_owner NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_runtime') THEN
        CREATE ROLE cartulary_runtime NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_recovery') THEN
        CREATE ROLE cartulary_recovery NOLOGIN;
    END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_migration_login') THEN
		CREATE ROLE cartulary_migration_login LOGIN PASSWORD NULL;
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_runtime_login') THEN
		CREATE ROLE cartulary_runtime_login LOGIN PASSWORD NULL;
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = 'cartulary_recovery_login') THEN
		CREATE ROLE cartulary_recovery_login LOGIN PASSWORD NULL;
	END IF;
END
$cartulary_roles$;
ALTER ROLE cartulary_schema_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
ALTER ROLE cartulary_runtime NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
ALTER ROLE cartulary_recovery NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
ALTER ROLE cartulary_migration_login LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD NULL;
ALTER ROLE cartulary_runtime_login LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD NULL;
ALTER ROLE cartulary_recovery_login LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD NULL;
GRANT cartulary_schema_owner TO cartulary_migration_login WITH INHERIT FALSE, SET TRUE, ADMIN FALSE;
GRANT cartulary_runtime TO cartulary_runtime_login WITH INHERIT FALSE, SET TRUE, ADMIN FALSE;
GRANT cartulary_recovery TO cartulary_recovery_login WITH INHERIT FALSE, SET TRUE, ADMIN FALSE;
REVOKE cartulary_runtime, cartulary_recovery FROM cartulary_migration_login;
REVOKE cartulary_schema_owner, cartulary_recovery FROM cartulary_runtime_login;
REVOKE cartulary_schema_owner, cartulary_runtime FROM cartulary_recovery_login;
GRANT SET ON PARAMETER session_replication_role TO cartulary_recovery;`

const freshDatabasePrerequisitesSQL = `
CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public VERSION '1.3';
CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public VERSION '1.6';
ALTER SCHEMA public OWNER TO cartulary_schema_owner;
REVOKE CREATE, USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO cartulary_schema_owner;
GRANT USAGE ON SCHEMA public TO cartulary_runtime, cartulary_recovery;
DO $extension_acl$
DECLARE
    routine pg_catalog.regprocedure;
BEGIN
    FOR routine IN
        SELECT procedure.oid::pg_catalog.regprocedure
          FROM pg_catalog.pg_proc AS procedure
          JOIN pg_catalog.pg_depend AS dependency
            ON dependency.classid = 'pg_catalog.pg_proc'::pg_catalog.regclass
           AND dependency.objid = procedure.oid
           AND dependency.deptype = 'e'
          JOIN pg_catalog.pg_extension AS extension
            ON extension.oid = dependency.refobjid
         WHERE extension.extname IN ('pgcrypto', 'citext')
         ORDER BY procedure.oid::pg_catalog.regprocedure::text
    LOOP
        EXECUTE format('REVOKE EXECUTE ON FUNCTION %s FROM PUBLIC', routine);
        EXECUTE format('GRANT EXECUTE ON FUNCTION %s TO cartulary_schema_owner, cartulary_runtime, cartulary_recovery', routine);
    END LOOP;
END
$extension_acl$;
DO $extension_type_acl$
DECLARE
    extension_type text;
BEGIN
    FOR extension_type IN
        SELECT format('%I.%I', type_namespace.nspname, candidate.typname)
          FROM pg_catalog.pg_type AS candidate
          JOIN pg_catalog.pg_namespace AS type_namespace
            ON type_namespace.oid = candidate.typnamespace
          JOIN pg_catalog.pg_depend AS dependency
            ON dependency.classid = 'pg_catalog.pg_type'::pg_catalog.regclass
           AND dependency.objid = candidate.oid
           AND dependency.deptype = 'e'
         JOIN pg_catalog.pg_extension AS extension
            ON extension.oid = dependency.refobjid
         WHERE extension.extname IN ('pgcrypto', 'citext')
           AND candidate.typelem = 0
         ORDER BY type_namespace.nspname, candidate.typname
    LOOP
        EXECUTE format('REVOKE USAGE ON TYPE %s FROM PUBLIC', extension_type);
        EXECUTE format('GRANT USAGE ON TYPE %s TO cartulary_schema_owner, cartulary_runtime, cartulary_recovery', extension_type);
    END LOOP;
END
$extension_type_acl$;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public REVOKE ALL ON TABLES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner IN SCHEMA public REVOKE ALL ON SEQUENCES FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE cartulary_schema_owner REVOKE USAGE ON TYPES FROM PUBLIC;
`

const databaseGrantsSQL = `REVOKE CONNECT, TEMPORARY ON DATABASE %s FROM PUBLIC;
GRANT CONNECT ON DATABASE %s TO cartulary_migration_login, cartulary_runtime_login, cartulary_recovery_login;`

// ProvisioningDigest binds exactly the prerequisites executed by this binary.
// It is independent of repository paths and never reads documentation.
func ProvisioningDigest() string {
	sum := sha256.Sum256([]byte(strings.Join([]string{postgresImage, postgresInitdbArgs, postgresHBA, postgresDisablePasswordSQL, strings.Join(postgresServerCommand(), "\x00"), clusterPrerequisitesSQL, freshDatabasePrerequisitesSQL, databaseGrantsSQL}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func TemplateDigest(schemaDigest string) string {
	sum := sha256.Sum256([]byte(schemaDigest + "\x00" + ProvisioningDigest()))
	return hex.EncodeToString(sum[:])
}

// initializeCluster runs only in the physical owner's unpublished generation.
func initializeCluster(ctx context.Context, adminDSN string) (string, error) {
	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return "", err
	}
	defer db.Close()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, clusterPrerequisitesSQL); err != nil {
		return "", fmt.Errorf("initialize postgres prerequisites: %w", err)
	}
	var generation string
	if err = tx.QueryRowContext(ctx, `SELECT system_identifier::text FROM pg_catalog.pg_control_system()`).Scan(&generation); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return generation, nil
}

func inspectClusterGeneration(ctx context.Context, adminDSN string) (string, error) {
	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return "", err
	}
	defer db.Close()
	var generation string
	err = db.QueryRowContext(ctx, `SELECT system_identifier::text FROM pg_catalog.pg_control_system()`).Scan(&generation)
	return generation, err
}

func (h *Harness) ProvisioningGeneration() string { return h.generation }

// PrepareFreshDatabase establishes database-local prerequisites before migrations.
// The physical owner must already have initialized cluster prerequisites.
func PrepareFreshDatabase(ctx context.Context, adminDSN, name string) error {
	config, err := pgx.ParseConfig(adminDSN)
	if err != nil {
		return err
	}
	config.Database = name
	db := stdlib.OpenDB(*config)
	defer db.Close()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, freshDatabasePrerequisitesSQL); err != nil {
		return fmt.Errorf("prepare fresh database prerequisites: %w", err)
	}
	quoted := pgx.Identifier{name}.Sanitize()
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(databaseGrantsSQL, quoted, quoted)); err != nil {
		return err
	}
	return tx.Commit()
}

func finalizeClone(ctx context.Context, admin *sql.Conn, name string) error {
	quoted := pgx.Identifier{name}.Sanitize()
	_, err := admin.ExecContext(ctx, fmt.Sprintf(databaseGrantsSQL, quoted, quoted))
	return err
}

// CloneDatabase finalizes database-specific grants under the creation lock.
// created remains true on a later failure, granting the caller exact cleanup duty.
func CloneDatabase(ctx context.Context, adminDSN, name, template string) (bool, error) {
	if template == "" {
		return false, errors.New("template identity required")
	}
	return postgrescatalog.CloneDatabase(ctx, adminDSN, name, template, postgresCatalogAdmissionTimeout,
		func(admin *sql.Conn) error { return finalizeCloneFn(ctx, admin, name) })
}

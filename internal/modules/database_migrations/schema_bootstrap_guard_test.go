package database_migrations_test

import (
	"strings"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
)

func TestSchemaBootstrapMigrationGuard(t *testing.T) {
	data, err := dbmigrations.Files.ReadFile("00001_database_infrastructure.sql")
	if err != nil {
		t.Fatalf("read database infrastructure migration: %v", err)
	}

	sqlText := string(data)
	requiredIdempotentStatements := []string{
		"CREATE EXTENSION IF NOT EXISTS pgcrypto;",
		"CREATE EXTENSION IF NOT EXISTS citext;",
		"CREATE TABLE IF NOT EXISTS public.schema_migration_lineage (",
		"cartulary.prod_ddl_rebaseline.v1",
	}
	for _, statement := range requiredIdempotentStatements {
		if !strings.Contains(sqlText, statement) {
			t.Fatalf("database infrastructure migration must keep lineage-safe DDL %q", statement)
		}
	}

	nonIdempotentStatements := []string{
		"CREATE EXTENSION pgcrypto;",
		"CREATE EXTENSION citext;",
		"CREATE TABLE schema_migration_lineage (",
	}
	for _, statement := range nonIdempotentStatements {
		if strings.Contains(sqlText, statement) {
			t.Fatalf("database infrastructure migration must not use non-idempotent DDL %q", statement)
		}
	}
}

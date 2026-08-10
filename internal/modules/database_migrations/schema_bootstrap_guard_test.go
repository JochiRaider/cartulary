package database_migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSchemaBootstrapMigrationGuard(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve source test path")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "..", "..", "..", "db", "migrations", "00001_database_infrastructure.sql"))
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

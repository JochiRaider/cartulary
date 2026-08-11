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
	requiredStatements := []string{
		"WHERE extension.extname = 'pgcrypto'",
		"WHERE extension.extname = 'citext'",
		"pgcrypto_version IS DISTINCT FROM '1.3'",
		"pgcrypto_schema IS DISTINCT FROM 'public'",
		"citext_version IS DISTINCT FROM '1.6'",
		"citext_schema IS DISTINCT FROM 'public'",
		"MESSAGE = 'schema_extension_prerequisite_invalid'",
		"CREATE TABLE public.schema_migration_lineage (",
		"cartulary.prod_ddl_rebaseline.v2",
	}
	for _, statement := range requiredStatements {
		if !strings.Contains(sqlText, statement) {
			t.Fatalf("database infrastructure migration must keep prerequisite and lineage contract %q", statement)
		}
	}

	forbiddenStatements := []string{
		"CREATE EXTENSION",
		"IF NOT EXISTS",
		"cartulary.prod_ddl_rebaseline.v1",
	}
	for _, statement := range forbiddenStatements {
		if strings.Contains(sqlText, statement) {
			t.Fatalf("database infrastructure migration must not retain permissive or v1 DDL %q", statement)
		}
	}
}

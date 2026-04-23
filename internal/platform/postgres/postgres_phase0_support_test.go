package postgres_test

import (
	"strings"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
)

func TestSupportPhase0_SchemaBootstrapMigrationGuard(t *testing.T) {
	data, err := dbmigrations.Files.ReadFile("00001_phase0_bootstrap.sql")
	if err != nil {
		t.Fatalf("read phase 0 bootstrap migration: %v", err)
	}

	sqlText := string(data)
	requiredIdempotentStatements := []string{
		"CREATE EXTENSION IF NOT EXISTS pgcrypto;",
		"CREATE EXTENSION IF NOT EXISTS citext;",
		"CREATE TABLE IF NOT EXISTS users (",
		"CREATE TABLE IF NOT EXISTS deployment_bootstrap_state (",
		"CREATE TABLE IF NOT EXISTS deployment_admin_audit_events (",
	}
	for _, statement := range requiredIdempotentStatements {
		if !strings.Contains(sqlText, statement) {
			t.Fatalf("bootstrap migration must keep rerun-safe DDL %q", statement)
		}
	}

	nonIdempotentStatements := []string{
		"CREATE EXTENSION pgcrypto;",
		"CREATE EXTENSION citext;",
		"CREATE TABLE users (",
		"CREATE TABLE deployment_bootstrap_state (",
		"CREATE TABLE deployment_admin_audit_events (",
	}
	for _, statement := range nonIdempotentStatements {
		if strings.Contains(sqlText, statement) {
			t.Fatalf("bootstrap migration must not use non-idempotent DDL %q", statement)
		}
	}
}

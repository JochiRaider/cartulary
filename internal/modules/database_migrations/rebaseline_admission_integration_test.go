package database_migrations_test

import (
	"context"
	"database/sql"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	platformpostgres "github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestProductionDDLContaminationFailsBeforeMutation_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB, err := harness.NewDatabase(context.Background(), "ddl-v2-contaminated")
	if err != nil {
		t.Fatal(err)
	}
	db, err := pgtest.OpenPurposeDatabase(testDB.DSN, platformpostgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(), `CREATE TABLE public.s18_preexisting_cartulary_object (id bigint)`); err != nil {
		t.Fatal(err)
	}
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	err = database_migrations.Apply(context.Background(), db, source)
	report := requireMigrationRemediation(t, err)
	finding := report.Findings[0]
	if report.FromVersion != 0 || report.ToVersion != 29 ||
		finding.RawValue != nil || finding.RawValuePair.LineageTablePresent ||
		finding.RemediationHint != "Destroy and recreate this database, then apply the Production DDL Rebaseline v2 catalog from version 1." {
		t.Fatalf("unexpected contamination remediation: %#v", report)
	}
	var contaminantPresent, ledgerPresent, lineagePresent bool
	if err := db.QueryRowContext(context.Background(), `SELECT to_regclass('public.s18_preexisting_cartulary_object') IS NOT NULL, to_regclass('public.goose_db_version') IS NOT NULL, to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&contaminantPresent, &ledgerPresent, &lineagePresent); err != nil {
		t.Fatal(err)
	}
	if !contaminantPresent || ledgerPresent || lineagePresent {
		t.Fatalf("contamination admission mutated database: contaminant=%t ledger=%t lineage=%t", contaminantPresent, ledgerPresent, lineagePresent)
	}
}

func TestProductionDDLExtensionPrerequisiteMatrix_Integration(t *testing.T) {
	tests := []struct {
		name   string
		mutate string
	}{
		{name: "missing", mutate: `DROP EXTENSION citext`},
		{name: "wrong schema", mutate: `CREATE SCHEMA s18_wrong_extension_schema; ALTER EXTENSION citext SET SCHEMA s18_wrong_extension_schema`},
		{name: "wrong version", mutate: `UPDATE pg_catalog.pg_extension SET extversion = '0.0-test' WHERE extname = 'citext'`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			harness := pgtest.Start(t)
			testDB, err := harness.NewDatabase(context.Background(), "ddl-v2-extension-invalid")
			if err != nil {
				t.Fatal(err)
			}
			admin := openAdminDatabase(t, harness.AdminDSN(), testDB.Name)
			if _, err := admin.ExecContext(context.Background(), test.mutate); err != nil {
				t.Fatalf("establish invalid extension state: %v", err)
			}
			db, err := pgtest.OpenPurposeDatabase(testDB.DSN, platformpostgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			source, err := dbmigrations.Source()
			if err != nil {
				t.Fatal(err)
			}
			err = database_migrations.Apply(context.Background(), db, source)
			requireExternalMigrationFailureReason(t, err, "schema_extension_prerequisite_invalid")
			var ledgerPresent, lineagePresent bool
			if err := admin.QueryRowContext(context.Background(), `SELECT to_regclass('public.goose_db_version') IS NOT NULL, to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&ledgerPresent, &lineagePresent); err != nil {
				t.Fatal(err)
			}
			if ledgerPresent || lineagePresent {
				t.Fatalf("invalid prerequisite mutated database: ledger=%t lineage=%t", ledgerPresent, lineagePresent)
			}
			if err.Error() != "schema_extension_prerequisite_invalid" {
				t.Fatalf("unsafe extension diagnostic: %q", err.Error())
			}
		})
	}

	t.Run("correct", func(t *testing.T) {
		harness := pgtest.Start(t)
		testDB, err := harness.NewDatabase(context.Background(), "ddl-v2-extension-valid")
		if err != nil {
			t.Fatal(err)
		}
		db, err := pgtest.OpenPurposeDatabase(testDB.DSN, platformpostgres.PurposeMigration)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		source, err := dbmigrations.Source()
		if err != nil {
			t.Fatal(err)
		}
		if err := database_migrations.Apply(context.Background(), db, source); err != nil {
			t.Fatalf("valid extension prerequisites rejected: %v", err)
		}
		var head int
		if err := db.QueryRowContext(context.Background(), `SELECT max(version_id) FROM public.goose_db_version WHERE is_applied`).Scan(&head); err != nil || head != 29 {
			t.Fatalf("valid prerequisite head = %d: %v", head, err)
		}
	})
}

func TestProductionDDLRollbackThroughZeroResidue_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	database := harness.MigrationDatabaseT(t)
	if err := database.RollbackThrough(context.Background(), 0); err != nil {
		t.Fatalf("rollback through zero: %v", err)
	}
	db := database.SQL()
	if objects := queryManagedCatalogObjects(t, db); len(objects) != 0 {
		keys := make([]string, 0, len(objects))
		for key := range objects {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		t.Fatalf("Cartulary-authored residue remains: %v", keys)
	}
	rows, err := db.QueryContext(context.Background(), `
SELECT relation.relkind::text, relation.relname
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = 'public'
  AND relation.relname LIKE 'goose_db_version%'
ORDER BY relation.relkind, relation.relname
`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var gooseObjects []string
	for rows.Next() {
		var kind, name string
		if err := rows.Scan(&kind, &name); err != nil {
			t.Fatal(err)
		}
		gooseObjects = append(gooseObjects, kind+":"+name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	wantGooseObjects := []string{"S:goose_db_version_id_seq", "i:goose_db_version_pkey", "r:goose_db_version"}
	if !equalStrings(gooseObjects, wantGooseObjects) {
		t.Fatalf("Goose residue = %v, want %v", gooseObjects, wantGooseObjects)
	}
	var version int64
	var applied bool
	var ledgerRows int
	if err := db.QueryRowContext(context.Background(), `SELECT min(version_id), bool_and(is_applied), count(*) FROM public.goose_db_version`).Scan(&version, &applied, &ledgerRows); err != nil {
		t.Fatal(err)
	}
	if version != 0 || !applied || ledgerRows != 1 {
		t.Fatalf("Goose version-zero row = version %d applied %t rows %d", version, applied, ledgerRows)
	}
	var extensionCount, roleCount, extraSchemaCount int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_catalog.pg_extension AS extension JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = extension.extnamespace WHERE (extension.extname, extension.extversion, namespace.nspname) IN (('pgcrypto','1.3','public'),('citext','1.6','public'))`).Scan(&extensionCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname IN ('cartulary_schema_owner','cartulary_runtime','cartulary_recovery','cartulary_migration_login','cartulary_runtime_login','cartulary_recovery_login')`).Scan(&roleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_catalog.pg_namespace WHERE nspname NOT IN ('public','information_schema') AND nspname NOT LIKE 'pg\_%' ESCAPE E'\\'`).Scan(&extraSchemaCount); err != nil {
		t.Fatal(err)
	}
	if extensionCount != 2 || roleCount != 6 || extraSchemaCount != 0 {
		t.Fatalf("administrator residue = extensions %d roles %d extra schemas %d", extensionCount, roleCount, extraSchemaCount)
	}
}

func openAdminDatabase(t testing.TB, adminDSN string, databaseName string) *sql.DB {
	t.Helper()
	config, err := pgx.ParseConfig(adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	db := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

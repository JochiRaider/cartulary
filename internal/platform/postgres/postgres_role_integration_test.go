package postgres_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPostgresPurposeConnectionsEstablishExactIdentity_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "postgres-purpose-identity")
	ctx := context.Background()

	for _, test := range postgresPurposeFixtures() {
		t.Run(test.name, func(t *testing.T) {
			dsn, err := testDB.DSNForPurpose(test.purpose)
			if err != nil {
				t.Fatal(err)
			}
			settings := postgres.Settings{
				BindingKind:  "managed_service",
				DSN:          dsn,
				Purpose:      test.purpose,
				ExpectedRole: test.role,
			}

			db, err := postgres.OpenSQL(ctx, settings)
			if err != nil {
				t.Fatalf("open stdlib purpose connection: %v", err)
			}
			assertPostgresIdentity(t, db, test.login, test.role)
			if err := db.Close(); err != nil {
				t.Fatalf("close stdlib purpose connection: %v", err)
			}

			pool, err := postgres.Setup(ctx, settings)
			if err != nil {
				t.Fatalf("open pool purpose connection: %v", err)
			}
			first, err := pool.Acquire(ctx)
			if err != nil {
				pool.Close()
				t.Fatal(err)
			}
			second, err := pool.Acquire(ctx)
			if err != nil {
				first.Release()
				pool.Close()
				t.Fatal(err)
			}
			var firstPID int64
			var secondPID int64
			assertPGXPostgresIdentity(t, first.QueryRow(ctx, `SELECT session_user::text, current_user::text, pg_backend_pid()::bigint`), test.login, test.role, &firstPID)
			assertPGXPostgresIdentity(t, second.QueryRow(ctx, `SELECT session_user::text, current_user::text, pg_backend_pid()::bigint`), test.login, test.role, &secondPID)
			if firstPID == secondPID {
				t.Fatalf("pool did not establish distinct physical connections: %d", firstPID)
			}
			second.Release()
			first.Release()
			pool.Reset()
			var recycledPID int64
			assertPGXPostgresIdentity(t, pool.QueryRow(ctx, `SELECT session_user::text, current_user::text, pg_backend_pid()::bigint`), test.login, test.role, &recycledPID)
			pool.Close()
		})
	}
}

func TestPostgresEffectiveRoleMismatchClosesConnection_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "postgres-role-mismatch")
	migrationDSN, err := testDB.DSNForPurpose(postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	settings := postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          migrationDSN,
		Purpose:      postgres.PurposeRuntime,
		ExpectedRole: "cartulary_runtime",
	}

	db, err := postgres.OpenSQL(context.Background(), settings)
	if db != nil {
		defer db.Close()
	}
	var configurationErr *postgres.ConfigurationError
	if !errors.As(err, &configurationErr) || configurationErr.Reason() != postgres.ReasonEffectiveRoleMismatch {
		t.Fatalf("effective-role mismatch = %T %v", err, err)
	}
	if db != nil {
		t.Fatal("eager mismatch returned a usable database handle")
	}
}

func TestPostgresRolesOwnershipAndPrivileges_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "postgres-role-acl")
	ctx := context.Background()
	migration := openPurposeDB(t, testDB, postgres.PurposeMigration)
	runtime := openPurposeDB(t, testDB, postgres.PurposeRuntime)
	recovery := openPurposeDB(t, testDB, postgres.PurposeRecovery)

	assertExactRoleCatalog(t, migration)
	assertExactObjectOwnership(t, migration)
	assertPublicAndDefaultPrivileges(t, migration)
	assertManifestPrivilegeParity(t, migration, runtime, recovery)
	assertRuntimePrivilegeMatrix(t, runtime)
	assertRecoveryPrivilegeMatrix(t, recovery)
	assertFutureObjectDefaults(t, migration, runtime, recovery)

	if _, err := migration.ExecContext(ctx, `CREATE TABLE public.s18_migration_ddl_probe (id bigint PRIMARY KEY)`); err != nil {
		t.Fatalf("migration role cannot create owner object: %v", err)
	}
	if _, err := migration.ExecContext(ctx, `DROP TABLE public.s18_migration_ddl_probe`); err != nil {
		t.Fatalf("migration role cannot remove owner object: %v", err)
	}
}

type privilegeManifest struct {
	Entries []privilegeManifestEntry `json:"entries"`
}

type privilegeManifestEntry struct {
	ObjectKind          string `json:"object_kind"`
	QualifiedName       string `json:"qualified_name"`
	RuntimeAccessClass  string `json:"runtime_access_class"`
	RecoveryAccessClass string `json:"recovery_access_class"`
}

func assertManifestPrivilegeParity(t testing.TB, migration *sql.DB, runtimeDB *sql.DB, recoveryDB *sql.DB) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve manifest path")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "..", "..", "..", "tools", "schema_object_ownership_manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest privilegeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, entry := range manifest.Entries {
		switch entry.ObjectKind {
		case "table", "view", "migration_metadata":
			assertTableAccessClass(t, runtimeDB, entry.QualifiedName, entry.RuntimeAccessClass)
			assertTableAccessClass(t, recoveryDB, entry.QualifiedName, entry.RecoveryAccessClass)
		case "sequence":
			assertSequenceAccessClass(t, runtimeDB, entry.QualifiedName, entry.RuntimeAccessClass)
			assertSequenceAccessClass(t, recoveryDB, entry.QualifiedName, entry.RecoveryAccessClass)
		case "routine":
			assertRoutineAccessClass(t, migration, runtimeDB, entry.QualifiedName, entry.RuntimeAccessClass)
			assertRoutineAccessClass(t, migration, recoveryDB, entry.QualifiedName, entry.RecoveryAccessClass)
		case "schema":
			assertSchemaAccessClass(t, runtimeDB, entry.QualifiedName, entry.RuntimeAccessClass)
			assertSchemaAccessClass(t, recoveryDB, entry.QualifiedName, entry.RecoveryAccessClass)
		}
	}
	assertExtensionTypePrivileges(t, migration)
}

func assertTableAccessClass(t testing.TB, db *sql.DB, objectName string, class string) {
	t.Helper()
	want := map[string]bool{}
	switch class {
	case "table_read_write":
		want = truePrivileges("SELECT", "INSERT", "UPDATE", "DELETE")
	case "table_append_only":
		want = truePrivileges("SELECT", "INSERT")
	case "table_read_only", "view_read_only", "migration_ledger_read":
		want = truePrivileges("SELECT")
	case "table_restore":
		want = truePrivileges("SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE")
	case "table_rebuild":
		want = truePrivileges("SELECT", "TRUNCATE")
	case "table_no_access":
	case "not_applicable":
		return
	default:
		t.Fatalf("unknown table access class %q for %s", class, objectName)
	}
	for _, privilege := range []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"} {
		var got bool
		if err := db.QueryRowContext(context.Background(), `SELECT pg_catalog.has_table_privilege(current_user, $1, $2)`, objectName, privilege).Scan(&got); err != nil {
			t.Fatalf("inspect %s %s: %v", objectName, privilege, err)
		}
		if got != want[privilege] {
			t.Fatalf("%s privilege %s = %t, want %t for %s", objectName, privilege, got, want[privilege], class)
		}
	}
}

func assertSequenceAccessClass(t testing.TB, db *sql.DB, objectName string, class string) {
	t.Helper()
	want := map[string]bool{}
	switch class {
	case "sequence_use":
		want = truePrivileges("USAGE")
	case "sequence_restore":
		want = truePrivileges("USAGE", "SELECT", "UPDATE")
	case "sequence_no_access":
	default:
		t.Fatalf("unknown sequence access class %q for %s", class, objectName)
	}
	for _, privilege := range []string{"USAGE", "SELECT", "UPDATE"} {
		var got bool
		if err := db.QueryRowContext(context.Background(), `SELECT pg_catalog.has_sequence_privilege(current_user, $1, $2)`, objectName, privilege).Scan(&got); err != nil {
			t.Fatalf("inspect %s %s: %v", objectName, privilege, err)
		}
		if got != want[privilege] {
			t.Fatalf("%s privilege %s = %t, want %t for %s", objectName, privilege, got, want[privilege], class)
		}
	}
}

func assertRoutineAccessClass(t testing.TB, catalogDB *sql.DB, db *sql.DB, objectName string, class string) {
	t.Helper()
	name := strings.TrimPrefix(objectName, "public.")
	var oid uint32
	if err := catalogDB.QueryRowContext(context.Background(), `SELECT routine.oid::bigint FROM pg_catalog.pg_proc AS routine JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace WHERE namespace.nspname = 'public' AND routine.proname = $1`, name).Scan(&oid); err != nil {
		t.Fatalf("resolve routine %s: %v", objectName, err)
	}
	want := class == "routine_application" || class == "routine_recovery"
	if class != "routine_application" && class != "routine_recovery" && class != "routine_private" {
		t.Fatalf("unknown routine access class %q for %s", class, objectName)
	}
	var got bool
	if err := db.QueryRowContext(context.Background(), `SELECT pg_catalog.has_function_privilege(current_user, $1::oid, 'EXECUTE')`, oid).Scan(&got); err != nil {
		t.Fatalf("inspect routine %s: %v", objectName, err)
	}
	if got != want {
		t.Fatalf("%s EXECUTE = %t, want %t for %s", objectName, got, want, class)
	}
}

func assertSchemaAccessClass(t testing.TB, db *sql.DB, objectName string, class string) {
	t.Helper()
	if class != "schema_usage" {
		t.Fatalf("unknown schema access class %q", class)
	}
	for _, fact := range []struct {
		privilege string
		want      bool
	}{{"USAGE", true}, {"CREATE", false}} {
		var got bool
		if err := db.QueryRowContext(context.Background(), `SELECT pg_catalog.has_schema_privilege(current_user, $1, $2)`, objectName, fact.privilege).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != fact.want {
			t.Fatalf("schema %s %s = %t, want %t", objectName, fact.privilege, got, fact.want)
		}
	}
}

func assertExtensionTypePrivileges(t testing.TB, db *sql.DB) {
	t.Helper()
	var violations int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
FROM pg_catalog.pg_type AS candidate
JOIN pg_catalog.pg_depend AS dependency
  ON dependency.classid = 'pg_catalog.pg_type'::pg_catalog.regclass
 AND dependency.objid = candidate.oid AND dependency.deptype = 'e'
JOIN pg_catalog.pg_extension AS extension ON extension.oid = dependency.refobjid
WHERE extension.extname IN ('pgcrypto', 'citext') AND candidate.typelem = 0
  AND (
      EXISTS (
          SELECT 1
          FROM pg_catalog.aclexplode(COALESCE(candidate.typacl, pg_catalog.acldefault('T', candidate.typowner))) AS acl
          WHERE acl.grantee = 0 AND acl.privilege_type = 'USAGE'
      )
      OR NOT pg_catalog.has_type_privilege('cartulary_schema_owner', candidate.oid, 'USAGE')
      OR NOT pg_catalog.has_type_privilege('cartulary_runtime', candidate.oid, 'USAGE')
      OR NOT pg_catalog.has_type_privilege('cartulary_recovery', candidate.oid, 'USAGE')
  )
`).Scan(&violations); err != nil {
		t.Fatal(err)
	}
	if violations != 0 {
		t.Fatalf("extension type ACL violations = %d", violations)
	}
}

func truePrivileges(privileges ...string) map[string]bool {
	result := make(map[string]bool, len(privileges))
	for _, privilege := range privileges {
		result[privilege] = true
	}
	return result
}

type postgresPurposeFixture struct {
	name    string
	purpose postgres.Purpose
	login   string
	role    string
}

func postgresPurposeFixtures() []postgresPurposeFixture {
	return []postgresPurposeFixture{
		{name: "runtime", purpose: postgres.PurposeRuntime, login: "cartulary_runtime_login", role: "cartulary_runtime"},
		{name: "migration", purpose: postgres.PurposeMigration, login: "cartulary_migration_login", role: "cartulary_schema_owner"},
		{name: "recovery", purpose: postgres.PurposeRecovery, login: "cartulary_recovery_login", role: "cartulary_recovery"},
	}
}

func openPurposeDB(t testing.TB, testDB *pgtest.TestDatabase, purpose postgres.Purpose) *sql.DB {
	t.Helper()
	db, err := pgtest.OpenPurposeDatabase(testDB.DSN, purpose)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertPostgresIdentity(t testing.TB, db *sql.DB, wantSession string, wantCurrent string) {
	t.Helper()
	var sessionUser string
	var currentUser string
	if err := db.QueryRowContext(context.Background(), `SELECT session_user::text, current_user::text`).Scan(&sessionUser, &currentUser); err != nil {
		t.Fatal(err)
	}
	if sessionUser != wantSession || currentUser != wantCurrent {
		t.Fatalf("identity = %q/%q, want %q/%q", sessionUser, currentUser, wantSession, wantCurrent)
	}
}

type pgxIdentityRow interface {
	Scan(...any) error
}

func assertPGXPostgresIdentity(t testing.TB, row pgxIdentityRow, wantSession string, wantCurrent string, pid *int64) {
	t.Helper()
	var sessionUser string
	var currentUser string
	if err := row.Scan(&sessionUser, &currentUser, pid); err != nil {
		t.Fatal(err)
	}
	if sessionUser != wantSession || currentUser != wantCurrent {
		t.Fatalf("identity = %q/%q, want %q/%q", sessionUser, currentUser, wantSession, wantCurrent)
	}
}

func assertExactRoleCatalog(t testing.TB, db *sql.DB) {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), `
SELECT rolname, rolcanlogin, rolsuper, rolcreatedb, rolcreaterole, rolinherit, rolreplication, rolbypassrls
FROM pg_catalog.pg_roles
WHERE rolname = ANY($1::text[])
ORDER BY rolname
`, []string{
		"cartulary_migration_login", "cartulary_recovery", "cartulary_recovery_login",
		"cartulary_runtime", "cartulary_runtime_login", "cartulary_schema_owner",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var name string
		var canLogin, super, createDB, createRole, inherit, replication, bypassRLS bool
		if err := rows.Scan(&name, &canLogin, &super, &createDB, &createRole, &inherit, &replication, &bypassRLS); err != nil {
			t.Fatal(err)
		}
		wantLogin := strings.HasSuffix(name, "_login")
		if canLogin != wantLogin || super || createDB || createRole || inherit || replication || bypassRLS {
			t.Fatalf("unexpected role attributes for %s", name)
		}
		seen++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if seen != 6 {
		t.Fatalf("fixed role count = %d, want 6", seen)
	}

	var membershipCount int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS member ON member.oid = membership.member
JOIN pg_catalog.pg_roles AS target ON target.oid = membership.roleid
WHERE (member.rolname, target.rolname) IN (
    ('cartulary_migration_login', 'cartulary_schema_owner'),
    ('cartulary_runtime_login', 'cartulary_runtime'),
    ('cartulary_recovery_login', 'cartulary_recovery')
)
  AND NOT membership.admin_option
  AND NOT membership.inherit_option
  AND membership.set_option
`).Scan(&membershipCount); err != nil {
		t.Fatal(err)
	}
	if membershipCount != 3 {
		t.Fatalf("exact purpose membership count = %d, want 3", membershipCount)
	}
	var unexpectedMemberships int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS member ON member.oid = membership.member
JOIN pg_catalog.pg_roles AS target ON target.oid = membership.roleid
WHERE (
    member.rolname IN ('cartulary_migration_login','cartulary_runtime_login','cartulary_recovery_login')
    AND (member.rolname, target.rolname) NOT IN (
        ('cartulary_migration_login','cartulary_schema_owner'),
        ('cartulary_runtime_login','cartulary_runtime'),
        ('cartulary_recovery_login','cartulary_recovery')
    )
) OR member.rolname IN ('cartulary_schema_owner','cartulary_runtime','cartulary_recovery')
`).Scan(&unexpectedMemberships); err != nil {
		t.Fatal(err)
	}
	if unexpectedMemberships != 0 {
		t.Fatalf("unexpected purpose-role memberships = %d", unexpectedMemberships)
	}

	for _, pair := range [][2]string{
		{"cartulary_runtime", "cartulary_recovery"},
		{"cartulary_runtime", "cartulary_schema_owner"},
		{"cartulary_recovery", "cartulary_runtime"},
		{"cartulary_recovery", "cartulary_schema_owner"},
	} {
		var member bool
		if err := db.QueryRowContext(context.Background(), `SELECT pg_catalog.pg_has_role($1, $2, 'MEMBER')`, pair[0], pair[1]).Scan(&member); err != nil {
			t.Fatal(err)
		}
		if member {
			t.Fatalf("role %s can assume %s", pair[0], pair[1])
		}
	}
	assertDatabaseAccessBoundary(t, db)
}

func assertDatabaseAccessBoundary(t testing.TB, db *sql.DB) {
	t.Helper()
	var publicPrivileges int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
FROM pg_catalog.pg_database AS database,
LATERAL pg_catalog.aclexplode(COALESCE(database.datacl, pg_catalog.acldefault('d', database.datdba))) AS acl
WHERE database.datname = current_database() AND acl.grantee = 0
`).Scan(&publicPrivileges); err != nil {
		t.Fatal(err)
	}
	if publicPrivileges != 0 {
		t.Fatalf("PUBLIC database privileges = %d", publicPrivileges)
	}
	for _, role := range []string{"cartulary_migration_login", "cartulary_runtime_login", "cartulary_recovery_login"} {
		for _, fact := range []struct {
			privilege string
			want      bool
		}{{"CONNECT", true}, {"TEMPORARY", false}, {"CREATE", false}} {
			var got bool
			if err := db.QueryRowContext(context.Background(), `SELECT pg_catalog.has_database_privilege($1, current_database(), $2)`, role, fact.privilege).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != fact.want {
				t.Fatalf("database privilege %s/%s = %t, want %t", role, fact.privilege, got, fact.want)
			}
		}
	}
	for _, role := range []string{"cartulary_schema_owner", "cartulary_runtime", "cartulary_recovery"} {
		for _, privilege := range []string{"CONNECT", "TEMPORARY", "CREATE"} {
			var got bool
			if err := db.QueryRowContext(context.Background(), `SELECT pg_catalog.has_database_privilege($1, current_database(), $2)`, role, privilege).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got {
				t.Fatalf("fixed role %s has database privilege %s", role, privilege)
			}
		}
	}
}

func assertExactObjectOwnership(t testing.TB, db *sql.DB) {
	t.Helper()
	queries := []string{
		`SELECT count(*) FROM pg_catalog.pg_namespace WHERE nspname = 'public' AND pg_catalog.pg_get_userbyid(nspowner) <> 'cartulary_schema_owner'`,
		`SELECT count(*) FROM pg_catalog.pg_class AS relation JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace WHERE namespace.nspname = 'public' AND relation.relkind IN ('r','p','v','m','S','f') AND relation.relname NOT IN ('goose_db_version','goose_db_version_id_seq') AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_depend AS dependency WHERE dependency.classid = 'pg_catalog.pg_class'::pg_catalog.regclass AND dependency.objid = relation.oid AND dependency.deptype = 'e') AND pg_catalog.pg_get_userbyid(relation.relowner) <> 'cartulary_schema_owner'`,
		`SELECT count(*) FROM pg_catalog.pg_proc AS routine JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace WHERE namespace.nspname = 'public' AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_depend AS dependency WHERE dependency.classid = 'pg_catalog.pg_proc'::pg_catalog.regclass AND dependency.objid = routine.oid AND dependency.deptype = 'e') AND pg_catalog.pg_get_userbyid(routine.proowner) <> 'cartulary_schema_owner'`,
		`SELECT count(*) FROM pg_catalog.pg_class AS relation WHERE pg_catalog.pg_get_userbyid(relation.relowner) IN ('cartulary_runtime','cartulary_recovery')`,
		`SELECT count(*) FROM pg_catalog.pg_proc AS routine WHERE pg_catalog.pg_get_userbyid(routine.proowner) IN ('cartulary_runtime','cartulary_recovery')`,
	}
	for _, query := range queries {
		var count int
		if err := db.QueryRowContext(context.Background(), query).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("ownership query found %d violations: %s", count, query)
		}
	}
}

func assertPublicAndDefaultPrivileges(t testing.TB, db *sql.DB) {
	t.Helper()
	queries := []string{
		`SELECT count(*) FROM pg_catalog.pg_namespace AS namespace, LATERAL pg_catalog.aclexplode(COALESCE(namespace.nspacl, pg_catalog.acldefault('n', namespace.nspowner))) AS acl WHERE namespace.nspname = 'public' AND acl.grantee = 0`,
		`SELECT count(*) FROM pg_catalog.pg_class AS relation JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace, LATERAL pg_catalog.aclexplode(COALESCE(relation.relacl, pg_catalog.acldefault(CASE WHEN relation.relkind = 'S' THEN 'S'::"char" ELSE 'r'::"char" END, relation.relowner))) AS acl WHERE namespace.nspname = 'public' AND relation.relkind IN ('r','p','v','m','S','f') AND acl.grantee = 0`,
		`SELECT count(*) FROM pg_catalog.pg_proc AS routine JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace, LATERAL pg_catalog.aclexplode(COALESCE(routine.proacl, pg_catalog.acldefault('f', routine.proowner))) AS acl WHERE namespace.nspname = 'public' AND acl.grantee = 0`,
		`SELECT count(*) FROM pg_catalog.pg_type AS candidate JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = candidate.typnamespace, LATERAL pg_catalog.aclexplode(COALESCE(candidate.typacl, pg_catalog.acldefault('T', candidate.typowner))) AS acl WHERE namespace.nspname = 'public' AND candidate.typelem = 0 AND acl.grantee = 0`,
		`SELECT count(*) FROM pg_catalog.pg_default_acl AS defaults JOIN pg_catalog.pg_roles AS owner ON owner.oid = defaults.defaclrole LEFT JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = defaults.defaclnamespace, LATERAL pg_catalog.aclexplode(defaults.defaclacl) AS acl WHERE owner.rolname = 'cartulary_schema_owner' AND (defaults.defaclnamespace = 0 OR namespace.nspname = 'public') AND acl.grantee = 0`,
	}
	for _, query := range queries {
		var count int
		if err := db.QueryRowContext(context.Background(), query).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("PUBLIC/default privilege query found %d violations: %s", count, query)
		}
	}
}

func assertRuntimePrivilegeMatrix(t testing.TB, db *sql.DB) {
	t.Helper()
	assertPrivilegeFacts(t, db, []privilegeFact{
		{`SELECT has_table_privilege(current_user, 'public.incidents', 'SELECT,INSERT,UPDATE,DELETE')`, true},
		{`SELECT has_table_privilege(current_user, 'public.incidents', 'TRUNCATE')`, false},
		{`SELECT has_table_privilege(current_user, 'public.goose_db_version', 'SELECT')`, true},
		{`SELECT has_table_privilege(current_user, 'public.goose_db_version', 'INSERT,UPDATE,DELETE,TRUNCATE')`, false},
		{`SELECT has_table_privilege(current_user, 'public.backup_sets', 'SELECT,INSERT,UPDATE,DELETE,TRUNCATE')`, false},
		{`SELECT has_table_privilege(current_user, 'public.administrative_audit_projections', 'SELECT,INSERT')`, true},
		{`SELECT has_table_privilege(current_user, 'public.administrative_audit_projections', 'UPDATE,DELETE,TRUNCATE')`, false},
		{`SELECT has_sequence_privilege(current_user, 'public.record_revisions_revision_id_seq', 'USAGE')`, true},
		{`SELECT has_sequence_privilege(current_user, 'public.record_revisions_revision_id_seq', 'UPDATE')`, false},
		{`SELECT has_schema_privilege(current_user, 'public', 'USAGE')`, true},
		{`SELECT has_schema_privilege(current_user, 'public', 'CREATE')`, false},
		{`SELECT pg_catalog.pg_has_role(current_user, 'cartulary_recovery', 'MEMBER')`, false},
		{`SELECT pg_catalog.pg_has_role(current_user, 'cartulary_schema_owner', 'MEMBER')`, false},
	})
	assertOperationDenied(t, db, `TRUNCATE TABLE public.collaboration_event_intents`)
	assertOperationDenied(t, db, `SELECT pg_catalog.setval('public.record_revisions_revision_id_seq', 1, false)`)
	assertOperationDenied(t, db, `SET session_replication_role = replica`)
	assertOperationDenied(t, db, `CREATE TABLE public.s18_runtime_ddl_probe (id bigint)`)
	assertOperationDenied(t, db, `INSERT INTO public.goose_db_version (version_id, is_applied) VALUES (30, true)`)
}

func assertRecoveryPrivilegeMatrix(t testing.TB, db *sql.DB) {
	t.Helper()
	assertPrivilegeFacts(t, db, []privilegeFact{
		{`SELECT has_table_privilege(current_user, 'public.incidents', 'SELECT,INSERT,UPDATE,DELETE,TRUNCATE')`, true},
		{`SELECT has_table_privilege(current_user, 'public.goose_db_version', 'SELECT')`, true},
		{`SELECT has_table_privilege(current_user, 'public.goose_db_version', 'INSERT,UPDATE,DELETE,TRUNCATE')`, false},
		{`SELECT has_sequence_privilege(current_user, 'public.record_revisions_revision_id_seq', 'USAGE,SELECT,UPDATE')`, true},
		{`SELECT has_schema_privilege(current_user, 'public', 'CREATE')`, false},
		{`SELECT pg_catalog.pg_has_role(current_user, 'cartulary_runtime', 'MEMBER')`, false},
		{`SELECT pg_catalog.pg_has_role(current_user, 'cartulary_schema_owner', 'MEMBER')`, false},
	})
	if _, err := db.ExecContext(context.Background(), `TRUNCATE TABLE public.collaboration_event_intents`); err != nil {
		t.Fatalf("Recovery cannot truncate restore table: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `SELECT pg_catalog.setval('public.record_revisions_revision_id_seq', 1, false)`); err != nil {
		t.Fatalf("Recovery cannot restore sequence: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `SET session_replication_role = replica`); err != nil {
		t.Fatalf("Recovery cannot establish restore session: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `SET session_replication_role = origin`); err != nil {
		t.Fatalf("Recovery cannot restore origin session: %v", err)
	}
	assertOperationDenied(t, db, `CREATE TABLE public.s18_recovery_ddl_probe (id bigint)`)
	assertOperationDenied(t, db, `INSERT INTO public.goose_db_version (version_id, is_applied) VALUES (30, true)`)
}

func assertFutureObjectDefaults(t testing.TB, migration *sql.DB, runtime *sql.DB, recovery *sql.DB) {
	t.Helper()
	ctx := context.Background()
	statements := []string{
		`CREATE TABLE public.s18_future_table (id bigint PRIMARY KEY)`,
		`CREATE SEQUENCE public.s18_future_sequence`,
		`CREATE FUNCTION public.s18_future_function() RETURNS bigint LANGUAGE sql SET search_path = pg_catalog, public AS 'SELECT 1::bigint'`,
	}
	for _, statement := range statements {
		if _, err := migration.ExecContext(ctx, statement); err != nil {
			t.Fatalf("create future object: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = migration.ExecContext(context.Background(), `DROP FUNCTION public.s18_future_function()`)
		_, _ = migration.ExecContext(context.Background(), `DROP SEQUENCE public.s18_future_sequence`)
		_, _ = migration.ExecContext(context.Background(), `DROP TABLE public.s18_future_table`)
	})
	for name, db := range map[string]*sql.DB{"runtime": runtime, "recovery": recovery} {
		assertPrivilegeFacts(t, db, []privilegeFact{
			{`SELECT has_table_privilege(current_user, 'public.s18_future_table', 'SELECT,INSERT,UPDATE,DELETE,TRUNCATE')`, false},
			{`SELECT has_sequence_privilege(current_user, 'public.s18_future_sequence', 'USAGE,SELECT,UPDATE')`, false},
			{`SELECT has_function_privilege(current_user, 'public.s18_future_function()', 'EXECUTE')`, false},
		})
		t.Logf("%s future-object defaults are deny-by-default", name)
	}
}

type privilegeFact struct {
	query string
	want  bool
}

func assertPrivilegeFacts(t testing.TB, db *sql.DB, facts []privilegeFact) {
	t.Helper()
	for _, fact := range facts {
		var got bool
		if err := db.QueryRowContext(context.Background(), fact.query).Scan(&got); err != nil {
			t.Fatalf("privilege query failed: %v (%s)", err, fact.query)
		}
		if got != fact.want {
			t.Fatalf("privilege query = %t, want %t: %s", got, fact.want, fact.query)
		}
	}
}

func assertOperationDenied(t testing.TB, db *sql.DB, statement string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), statement); err == nil {
		t.Fatalf("prohibited operation succeeded: %s", statement)
	} else if strings.Contains(strings.ToLower(err.Error()), "password") || strings.Contains(strings.ToLower(err.Error()), "postgres://") {
		t.Fatalf("prohibited operation disclosed credentials: %v", err)
	}
}

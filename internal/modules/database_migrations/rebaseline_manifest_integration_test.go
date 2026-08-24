package database_migrations_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	platformpostgres "github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type schemaObjectManifest struct {
	SchemaID               string                      `json:"schema_id"`
	MigrationRoot          string                      `json:"migration_root"`
	SupportedPostgresMajor int                         `json:"supported_postgres_major"`
	ApplicationSchemas     []string                    `json:"application_schemas"`
	GooseLedger            string                      `json:"goose_ledger"`
	LineageRelation        string                      `json:"lineage_relation"`
	AllowedOwners          []string                    `json:"allowed_owners"`
	Entries                []schemaObjectManifestEntry `json:"entries"`
}

type schemaObjectManifestEntry struct {
	ObjectID               string   `json:"object_id"`
	ObjectKind             string   `json:"object_kind"`
	QualifiedName          string   `json:"qualified_name"`
	ManagementClass        string   `json:"management_class"`
	SourceOwner            string   `json:"source_owner"`
	MigrationVersion       *int     `json:"migration_version"`
	MigrationFile          *string  `json:"migration_file"`
	DependencyObjectIDs    []string `json:"dependency_object_ids"`
	RuntimeAccessClass     string   `json:"runtime_access_class"`
	RecoveryAccessClass    string   `json:"recovery_access_class"`
	RecoveryClassification string   `json:"recovery_classification"`
	ForeignKeyIndexStatus  string   `json:"foreign_key_index_status"`
	Approval               any      `json:"approval"`
}

type recoveryCatalogView struct {
	SchemaID string `json:"schema_id"`
	Tables   []struct {
		OwnerID         string `json:"owner_id"`
		TableName       string `json:"table_name"`
		BackupInclusion string `json:"backup_inclusion"`
	} `json:"tables"`
}

func TestProductionDDLObjectManifestContract(t *testing.T) {
	manifest := loadSchemaObjectManifest(t)
	migrationHistory := loadMigrationHistoryManifest(t)
	migrationFiles := make(map[int64]string, len(migrationHistory.Entries))
	for _, entry := range migrationHistory.Entries {
		migrationFiles[entry.Version] = entry.Filename
	}
	if manifest.SchemaID != "cartulary.schema_object_ownership_manifest.v2" ||
		manifest.MigrationRoot != "db/migrations" ||
		manifest.SupportedPostgresMajor != 16 ||
		manifest.GooseLedger != "public.goose_db_version" ||
		manifest.LineageRelation != "public.schema_migration_lineage" ||
		len(manifest.ApplicationSchemas) != 1 || manifest.ApplicationSchemas[0] != "public" {
		t.Fatalf("unexpected manifest identity: %#v", manifest)
	}
	if len(manifest.Entries) == 0 || !sort.StringsAreSorted(manifest.AllowedOwners) {
		t.Fatal("object manifest owner or entry catalog is empty or unstable")
	}
	owners := stringSet(manifest.AllowedOwners)
	entries := make(map[string]schemaObjectManifestEntry, len(manifest.Entries))
	identities := make(map[string]string, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		if entry.ObjectID == "" || entry.QualifiedName == "" || entry.SourceOwner == "" {
			t.Fatalf("incomplete manifest entry: %#v", entry)
		}
		if _, duplicate := entries[entry.ObjectID]; duplicate {
			t.Fatalf("duplicate object id %q", entry.ObjectID)
		}
		identity := entry.ObjectKind + "|" + entry.QualifiedName
		if prior, duplicate := identities[identity]; duplicate {
			t.Fatalf("duplicate object identity %q (%s, %s)", identity, prior, entry.ObjectID)
		}
		if !owners[entry.SourceOwner] {
			t.Fatalf("object %s has unknown owner %q", entry.ObjectID, entry.SourceOwner)
		}
		if !sort.StringsAreSorted(entry.DependencyObjectIDs) || hasDuplicateStrings(entry.DependencyObjectIDs) {
			t.Fatalf("object %s has unstable dependencies: %v", entry.ObjectID, entry.DependencyObjectIDs)
		}
		assertManifestAccessClass(t, entry)
		if entry.ManagementClass == "cartulary_authored" {
			if entry.MigrationVersion == nil || *entry.MigrationVersion < 1 || entry.MigrationFile == nil {
				t.Fatalf("authored object lacks migration allocation: %#v", entry)
			}
			wantMigrationFile, exists := migrationFiles[int64(*entry.MigrationVersion)]
			if !exists || *entry.MigrationFile != wantMigrationFile {
				t.Fatalf("object %s has unknown migration allocation: %d/%s", entry.ObjectID, *entry.MigrationVersion, *entry.MigrationFile)
			}
			if !strings.HasPrefix(*entry.MigrationFile, leftPadVersion(*entry.MigrationVersion)+"_") {
				t.Fatalf("object %s migration allocation is inconsistent: %d/%s", entry.ObjectID, *entry.MigrationVersion, *entry.MigrationFile)
			}
		} else if entry.MigrationVersion != nil || entry.MigrationFile != nil {
			t.Fatalf("externally managed object has authored allocation: %#v", entry)
		}
		entries[entry.ObjectID] = entry
		identities[identity] = entry.ObjectID
	}
	for _, entry := range manifest.Entries {
		for _, dependency := range entry.DependencyObjectIDs {
			if dependency == entry.ObjectID {
				t.Fatalf("object %s depends on itself", entry.ObjectID)
			}
			if _, ok := entries[dependency]; !ok {
				t.Fatalf("object %s has unknown dependency %s", entry.ObjectID, dependency)
			}
		}
	}
	assertRecoveryCardinality(t, manifest)
}

func TestProductionDDLCatalogManifestParity_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "ddl-v2-manifest-parity")
	db, err := pgtest.OpenPurposeDatabase(testDB.DSN, platformpostgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	manifest := loadSchemaObjectManifest(t)
	want := make(map[string]bool)
	for _, entry := range manifest.Entries {
		switch entry.ObjectKind {
		case "table", "view", "sequence", "routine", "trigger", "constraint", "index":
			want[entry.ObjectKind+"|"+entry.QualifiedName] = true
		}
	}
	got := queryManagedCatalogObjects(t, db)
	assertExactStringSet(t, got, want)
	assertCatalogStructuralFacts(t, db, manifest)
}

func assertManifestAccessClass(t testing.TB, entry schemaObjectManifestEntry) {
	t.Helper()
	allowed := map[string][2][]string{
		"schema":             {{"schema_usage"}, {"schema_usage"}},
		"extension":          {{"type_use"}, {"type_use"}},
		"migration_metadata": {{"migration_ledger_read"}, {"migration_ledger_read"}},
		"table":              {{"table_read_write", "table_append_only", "table_read_only", "table_no_access", "migration_ledger_read"}, {"table_restore", "table_rebuild", "table_read_only", "table_no_access", "migration_ledger_read"}},
		"view":               {{"view_read_only"}, {"view_read_only"}},
		"sequence":           {{"sequence_use", "sequence_no_access"}, {"sequence_restore", "sequence_no_access"}},
		"routine":            {{"routine_application", "routine_private"}, {"routine_recovery", "routine_private"}},
		"trigger":            {{"not_applicable"}, {"not_applicable"}},
		"constraint":         {{"not_applicable"}, {"not_applicable"}},
		"index":              {{"not_applicable"}, {"not_applicable"}},
	}
	classes, ok := allowed[entry.ObjectKind]
	if !ok || !containsString(classes[0], entry.RuntimeAccessClass) || !containsString(classes[1], entry.RecoveryAccessClass) {
		t.Fatalf("object %s has invalid access classes %q/%q", entry.ObjectID, entry.RuntimeAccessClass, entry.RecoveryAccessClass)
	}
	if entry.ForeignKeyIndexStatus == "intentionally_unindexed" && entry.Approval == nil {
		t.Fatalf("object %s lacks unindexed-FK approval", entry.ObjectID)
	}
	if entry.ForeignKeyIndexStatus != "not_applicable" && entry.ForeignKeyIndexStatus != "covered" && entry.ForeignKeyIndexStatus != "intentionally_unindexed" {
		t.Fatalf("object %s has invalid FK status %q", entry.ObjectID, entry.ForeignKeyIndexStatus)
	}
}

func assertRecoveryCardinality(t testing.TB, manifest schemaObjectManifest) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repositoryRoot(t), "contracts", "recovery", "fixtures", "recovery-state-catalog.v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var recovery recoveryCatalogView
	if err := json.Unmarshal(data, &recovery); err != nil {
		t.Fatal(err)
	}
	if recovery.SchemaID != "cartulary.recovery_state_catalog.v1" || len(recovery.Tables) != 115 {
		t.Fatalf("Recovery catalog identity/cardinality = %q/%d", recovery.SchemaID, len(recovery.Tables))
	}
	authoritative := 0
	revisionConflictFacts := 0
	manifestTables := make(map[string]schemaObjectManifestEntry)
	for _, entry := range manifest.Entries {
		if entry.ObjectKind == "table" {
			manifestTables[strings.TrimPrefix(entry.QualifiedName, "public.")] = entry
		}
	}
	for _, table := range recovery.Tables {
		if table.BackupInclusion == "authoritative_required" {
			authoritative++
		}
		entry, ok := manifestTables[table.TableName]
		if !ok {
			t.Fatalf("Recovery table %s is absent from object manifest", table.TableName)
		}
		if entry.SourceOwner != strings.TrimPrefix(table.OwnerID, "module.") || entry.RecoveryClassification != normalizedRecoveryClassification(table.BackupInclusion) {
			t.Fatalf("Recovery/object manifest mismatch for %s", table.TableName)
		}
		if table.TableName == "record_revision_conflict_facts" && table.OwnerID == "module.revisions" && table.BackupInclusion == "authoritative_required" {
			revisionConflictFacts++
		}
	}
	if authoritative != 84 || revisionConflictFacts != 1 || len(manifestTables) != 115 {
		t.Fatalf("Recovery facts = tables %d/%d, authoritative %d, revision conflict facts %d", len(recovery.Tables), len(manifestTables), authoritative, revisionConflictFacts)
	}
}

func queryManagedCatalogObjects(t testing.TB, db *sql.DB) map[string]bool {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), `
WITH extension_objects AS (
    SELECT dependency.classid, dependency.objid
    FROM pg_catalog.pg_depend AS dependency
    JOIN pg_catalog.pg_extension AS extension ON extension.oid = dependency.refobjid
    WHERE dependency.deptype = 'e' AND extension.extname IN ('pgcrypto', 'citext')
), managed AS (
    SELECT CASE relation.relkind WHEN 'r' THEN 'table' WHEN 'p' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'view' WHEN 'S' THEN 'sequence' END AS kind,
           'public.' || relation.relname AS qualified_name
    FROM pg_catalog.pg_class AS relation
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
    WHERE namespace.nspname = 'public' AND relation.relkind IN ('r','p','v','m','S')
      AND relation.relname NOT IN ('goose_db_version', 'goose_db_version_id_seq')
      AND NOT EXISTS (SELECT 1 FROM extension_objects WHERE classid = 'pg_catalog.pg_class'::pg_catalog.regclass AND objid = relation.oid)
    UNION ALL
    SELECT 'routine', 'public.' || routine.proname
    FROM pg_catalog.pg_proc AS routine
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace
    WHERE namespace.nspname = 'public'
      AND NOT EXISTS (SELECT 1 FROM extension_objects WHERE classid = 'pg_catalog.pg_proc'::pg_catalog.regclass AND objid = routine.oid)
    UNION ALL
    SELECT 'trigger', 'public.' || relation.relname || '.' || trigger.tgname
    FROM pg_catalog.pg_trigger AS trigger
    JOIN pg_catalog.pg_class AS relation ON relation.oid = trigger.tgrelid
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
    WHERE namespace.nspname = 'public' AND NOT trigger.tgisinternal
    UNION ALL
    SELECT 'constraint', 'public.' || relation.relname || '.' || constraint_row.conname
    FROM pg_catalog.pg_constraint AS constraint_row
    JOIN pg_catalog.pg_class AS relation ON relation.oid = constraint_row.conrelid
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
    WHERE namespace.nspname = 'public' AND relation.relname <> 'goose_db_version'
    UNION ALL
    SELECT 'index', 'public.' || index_relation.relname
    FROM pg_catalog.pg_index AS index_state
    JOIN pg_catalog.pg_class AS index_relation ON index_relation.oid = index_state.indexrelid
    JOIN pg_catalog.pg_class AS table_relation ON table_relation.oid = index_state.indrelid
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = index_relation.relnamespace
    WHERE namespace.nspname = 'public' AND table_relation.relname <> 'goose_db_version'
)
SELECT kind, qualified_name FROM managed ORDER BY kind, qualified_name
`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	result := make(map[string]bool)
	for rows.Next() {
		var kind, name string
		if err := rows.Scan(&kind, &name); err != nil {
			t.Fatal(err)
		}
		identity := kind + "|" + name
		if result[identity] {
			t.Fatalf("database contains duplicate manifest identity %s", identity)
		}
		result[identity] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertCatalogStructuralFacts(t testing.TB, db *sql.DB, manifest schemaObjectManifest) {
	t.Helper()
	var major int
	if err := db.QueryRowContext(context.Background(), `SELECT current_setting('server_version_num')::integer / 10000`).Scan(&major); err != nil || major != 16 {
		t.Fatalf("PostgreSQL major = %d, want 16: %v", major, err)
	}
	var invalidConstraints int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_catalog.pg_constraint AS constraint_row JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = constraint_row.connamespace WHERE namespace.nspname = 'public' AND NOT constraint_row.convalidated`).Scan(&invalidConstraints); err != nil || invalidConstraints != 0 {
		t.Fatalf("invalid constraint count = %d: %v", invalidConstraints, err)
	}
	var foreignKeys, uncovered int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*), count(*) FILTER (WHERE NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_index AS index_state
    WHERE index_state.indrelid = constraint_row.conrelid
      AND index_state.indisvalid AND index_state.indisready AND index_state.indpred IS NULL
      AND index_state.indnkeyatts >= cardinality(constraint_row.conkey)
      AND NOT EXISTS (
          SELECT 1 FROM generate_subscripts(constraint_row.conkey, 1) AS position
          WHERE (index_state.indkey::smallint[])[position - 1] <> constraint_row.conkey[position]
      )
))
FROM pg_catalog.pg_constraint AS constraint_row
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = constraint_row.connamespace
WHERE namespace.nspname = 'public' AND constraint_row.contype = 'f'
`).Scan(&foreignKeys, &uncovered); err != nil {
		t.Fatal(err)
	}
	coveredManifest := 0
	for _, entry := range manifest.Entries {
		if entry.ForeignKeyIndexStatus == "covered" {
			coveredManifest++
		}
	}
	if foreignKeys == 0 || uncovered != 0 || coveredManifest != foreignKeys {
		t.Fatalf("FK coverage = actual %d, manifest %d, uncovered %d (%v)", foreignKeys, coveredManifest, uncovered, queryUncoveredForeignKeys(t, db))
	}
	var unsafeRoutines int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace
WHERE namespace.nspname = 'public'
  AND NOT EXISTS (
      SELECT 1 FROM pg_catalog.pg_depend AS dependency
      WHERE dependency.classid = 'pg_catalog.pg_proc'::pg_catalog.regclass
        AND dependency.objid = routine.oid AND dependency.deptype = 'e'
  )
  AND (
      NOT ('search_path=pg_catalog, public' = ANY(COALESCE(routine.proconfig, ARRAY[]::text[])))
      OR EXISTS (
          SELECT 1
          FROM pg_catalog.aclexplode(COALESCE(routine.proacl, pg_catalog.acldefault('f', routine.proowner))) AS acl
          WHERE acl.grantee = 0 AND acl.privilege_type = 'EXECUTE'
      )
      OR (routine.prosecdef <> (routine.proname IN (
          'revisions_incident_bundle_sequence_begin_v1',
          'revisions_incident_bundle_sequence_finish_v1',
          'entities_refresh_active_identifier_claims_v1',
          'entities_release_active_identifier_claims_v1',
          'entities_sync_active_identifier_claims_v1',
          'entities_rebuild_active_identifier_claims_v1',
          'entities_active_identifier_claims_are_valid_v1',
          'parties_refresh_active_key_claims_v1',
          'parties_release_active_key_claims_v1',
          'parties_sync_active_key_claims_v1',
          'parties_rebuild_active_key_claims_v1',
          'parties_active_key_claims_are_valid_v1'
      )))
  )
`).Scan(&unsafeRoutines); err != nil || unsafeRoutines != 0 {
		t.Fatalf("unsafe routine count = %d: %v", unsafeRoutines, err)
	}
}

func queryUncoveredForeignKeys(t testing.TB, db *sql.DB) []string {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), `
SELECT relation.relname || '.' || constraint_row.conname
FROM pg_catalog.pg_constraint AS constraint_row
JOIN pg_catalog.pg_class AS relation ON relation.oid = constraint_row.conrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = constraint_row.connamespace
WHERE namespace.nspname = 'public' AND constraint_row.contype = 'f'
  AND NOT EXISTS (
      SELECT 1 FROM pg_catalog.pg_index AS index_state
      WHERE index_state.indrelid = constraint_row.conrelid
        AND index_state.indisvalid AND index_state.indisready AND index_state.indpred IS NULL
        AND index_state.indnkeyatts >= cardinality(constraint_row.conkey)
        AND NOT EXISTS (
            SELECT 1 FROM generate_subscripts(constraint_row.conkey, 1) AS position
            WHERE (index_state.indkey::smallint[])[position - 1] <> constraint_row.conkey[position]
        )
  )
ORDER BY relation.relname, constraint_row.conname
`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		result = append(result, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func loadSchemaObjectManifest(t testing.TB) schemaObjectManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repositoryRoot(t), "tools", "schema_object_ownership_manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest schemaObjectManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func assertExactStringSet(t testing.TB, got, want map[string]bool) {
	t.Helper()
	var missing, extra []string
	for identity := range want {
		if !got[identity] {
			missing = append(missing, identity)
		}
	}
	for identity := range got {
		if !want[identity] {
			extra = append(extra, identity)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) != 0 || len(extra) != 0 {
		t.Fatalf("manifest/catalog mismatch: missing=%v extra=%v", missing, extra)
	}
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func leftPadVersion(version int) string {
	return fmt.Sprintf("%05d", version)
}

func normalizedRecoveryClassification(value string) string {
	switch value {
	case "authoritative_required", "excluded_rebuildable", "excluded_security_state", "excluded_recovery_metadata":
		return value
	default:
		return "not_applicable"
	}
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	recoveryRegistrySchemaID = "cartulary.recovery_contract_registry.v1"
	recoveryCatalogSchemaID  = "cartulary.recovery_state_catalog.v1"
)

var (
	recoveryRegistryKeys = stringSet(
		"$schema",
		"schema_id",
		"canonicalization",
		"current_schema_ids",
		"historical_reader_schema_ids",
		"limits",
		"schemas",
		"canonical_fixtures",
	)
	recoverySchemaIDsByPath = map[string]string{
		"backup-artifact-envelope.v2.schema.json":            "cartulary.backup_artifact_envelope.v2",
		"backup-integrity-manifest.v3.schema.json":           "cartulary.backup_integrity_manifest.v3",
		"common.v1.schema.json":                              "cartulary.recovery_common.v1",
		"object-store-backup-manifest.v2.schema.json":        "cartulary.object_store_backup_manifest.v2",
		"object-store-backup-summary.v2.schema.json":         "cartulary.object_store_backup_summary.v2",
		"operator-recovery-audit-summary.v2.schema.json":     "cartulary.operator_recovery_audit_summary.v2",
		"operator-recovery-journal-payload.v2.schema.json":   "cartulary.operator_recovery_journal_payload.v2",
		"postgres-snapshot-artifact.v2.schema.json":          "cartulary.postgres_snapshot_artifact.v2",
		"postgres-snapshot-unit.v1.schema.json":              "cartulary.postgres_snapshot_unit.v1",
		"recovery-state-catalog.v1.schema.json":              recoveryCatalogSchemaID,
		"recovery-state-contribution.v1.schema.json":         "cartulary.recovery_state_contribution.v1",
		"restore-target-marker.v2.schema.json":               "cartulary.restore_target_marker.v2",
		"restore-verification.v2.schema.json":                "cartulary.restore_verification.v2",
		"restore-workbook-probe-registration.v1.schema.json": "cartulary.restore_workbook_probe_registration.v1",
	}
	recoveryFixtureIDsByPath = map[string]string{
		"fixtures/backup-artifact-envelope.v2.json":            "cartulary.backup_artifact_envelope.v2",
		"fixtures/backup-integrity-manifest.v3.json":           "cartulary.backup_integrity_manifest.v3",
		"fixtures/object-store-backup-manifest.v2.json":        "cartulary.object_store_backup_manifest.v2",
		"fixtures/object-store-backup-summary.v2.json":         "cartulary.object_store_backup_summary.v2",
		"fixtures/operator-recovery-audit-summary.v2.json":     "cartulary.operator_recovery_audit_summary.v2",
		"fixtures/operator-recovery-journal-payload.v2.json":   "cartulary.operator_recovery_journal_payload.v2",
		"fixtures/postgres-snapshot-artifact.v2.json":          "cartulary.postgres_snapshot_artifact.v2",
		"fixtures/postgres-snapshot-unit.v1.json":              "cartulary.postgres_snapshot_unit.v1",
		"fixtures/recovery-state-catalog.v1.json":              recoveryCatalogSchemaID,
		"fixtures/recovery-state-contribution.v1.json":         "cartulary.recovery_state_contribution.v1",
		"fixtures/restore-target-marker.v2.json":               "cartulary.restore_target_marker.v2",
		"fixtures/restore-verification.v2.json":                "cartulary.restore_verification.v2",
		"fixtures/restore-workbook-probe-registration.v1.json": "cartulary.restore_workbook_probe_registration.v1",
	}
	createTablePattern = regexp.MustCompile(`(?i)\bCREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+(?:public\.)?([a-z][a-z0-9_]*)`)
)

func validateRecoveryContractInput(relativePath string, value any) error {
	switch {
	case relativePath == "index.json":
		return validateRecoveryRegistry(value)
	case strings.HasSuffix(relativePath, ".schema.json"):
		expectedID, ok := recoverySchemaIDsByPath[relativePath]
		if !ok {
			return fmt.Errorf("unexpected recovery schema %s", relativePath)
		}
		object, err := asObject(value, "contracts/recovery/"+relativePath)
		if err != nil {
			return err
		}
		if err := requireDraftSchema(object, "contracts/recovery/"+relativePath); err != nil {
			return err
		}
		actualID, err := requiredString(object, "$id", "contracts/recovery/"+relativePath)
		if err != nil {
			return err
		}
		if actualID != expectedID {
			return fmt.Errorf("contracts/recovery/%s.$id must be %s", relativePath, expectedID)
		}
		return nil
	case strings.HasPrefix(relativePath, "fixtures/"):
		expectedID, ok := recoveryFixtureIDsByPath[relativePath]
		if !ok {
			return fmt.Errorf("unexpected recovery fixture %s", relativePath)
		}
		object, err := asObject(value, "contracts/recovery/"+relativePath)
		if err != nil {
			return err
		}
		actualID, err := requiredString(object, "schema_id", "contracts/recovery/"+relativePath)
		if err != nil {
			return err
		}
		if actualID != expectedID {
			return fmt.Errorf("contracts/recovery/%s.schema_id must be %s", relativePath, expectedID)
		}
		return nil
	default:
		return fmt.Errorf("unexpected recovery artifact %s", relativePath)
	}
}

func validateRecoveryRegistry(value any) error {
	object, err := asObject(value, "contracts/recovery/index.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, recoveryRegistryKeys, "contracts/recovery/index.json"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "contracts/recovery/index.json"); err != nil {
		return err
	}
	schemaID, err := requiredString(object, "schema_id", "contracts/recovery/index.json")
	if err != nil {
		return err
	}
	if schemaID != recoveryRegistrySchemaID {
		return fmt.Errorf("contracts/recovery/index.json.schema_id must be %s", recoveryRegistrySchemaID)
	}
	canonicalization, err := requiredString(object, "canonicalization", "contracts/recovery/index.json")
	if err != nil {
		return err
	}
	if canonicalization != "cartulary.recovery_canonical_json.v1" {
		return fmt.Errorf("contracts/recovery/index.json.canonicalization is unsupported")
	}
	for _, field := range []string{
		"current_schema_ids",
		"historical_reader_schema_ids",
		"schemas",
		"canonical_fixtures",
	} {
		values, err := stringArray(object[field], "contracts/recovery/index.json."+field, true)
		if err != nil {
			return err
		}
		if err := requireSortedUniqueStrings(values, "contracts/recovery/index.json."+field); err != nil {
			return err
		}
	}
	if _, err := asObject(object["limits"], "contracts/recovery/index.json.limits"); err != nil {
		return err
	}
	return nil
}

func requireSortedUniqueStrings(values []string, label string) error {
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return fmt.Errorf("%s must be unique and sorted by ascending UTF-8 bytes", label)
		}
	}
	return nil
}

func validateRecoveryContractFamily(root string) error {
	base := filepath.Join(root, "contracts", "recovery")
	registry, err := readRecoveryObject(base, "index.json")
	if err != nil {
		return err
	}
	if err := validateRecoveryRegistry(registry); err != nil {
		return err
	}
	if err := validateRecoveryRegistryPaths(base, registry); err != nil {
		return err
	}

	catalog, err := readRecoveryObject(base, "fixtures/recovery-state-catalog.v1.json")
	if err != nil {
		return err
	}
	requiredTables, err := validateRecoveryCatalog(root, catalog)
	if err != nil {
		return err
	}
	snapshot, err := readRecoveryObject(base, "fixtures/postgres-snapshot-artifact.v2.json")
	if err != nil {
		return err
	}
	return validateRecoverySnapshotFixture(snapshot, requiredTables)
}

func readRecoveryObject(base, relativePath string) (map[string]any, error) {
	raw, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(relativePath)))
	if err != nil {
		return nil, fmt.Errorf("read contracts/recovery/%s: %w", relativePath, err)
	}
	decoded, err := decodeContract(raw)
	if err != nil {
		return nil, fmt.Errorf("decode contracts/recovery/%s: %w", relativePath, err)
	}
	return asObject(decoded, "contracts/recovery/"+relativePath)
}

func validateRecoveryRegistryPaths(base string, registry map[string]any) error {
	expected := map[string]struct{}{"index.json": {}}
	for _, field := range []string{"schemas", "canonical_fixtures"} {
		paths, _ := stringArray(registry[field], field, true)
		for _, path := range paths {
			expected[path] = struct{}{}
			info, err := os.Stat(filepath.Join(base, filepath.FromSlash(path)))
			if err != nil {
				return fmt.Errorf("contracts/recovery/index.json references missing %s: %w", path, err)
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("contracts/recovery/index.json path %s is not a regular file", path)
			}
		}
	}
	actual := map[string]struct{}{}
	if err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}
		relative, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		actual[filepath.ToSlash(relative)] = struct{}{}
		return nil
	}); err != nil {
		return err
	}
	return compareStringSets(expected, actual, "recovery registry paths")
}

func validateRecoveryCatalog(root string, catalog map[string]any) ([]string, error) {
	schemaID, err := requiredString(catalog, "schema_id", "recovery catalog")
	if err != nil {
		return nil, err
	}
	if schemaID != recoveryCatalogSchemaID {
		return nil, fmt.Errorf("recovery catalog schema_id must be %s", recoveryCatalogSchemaID)
	}
	tables, err := objectArray(catalog["tables"], "recovery catalog tables")
	if err != nil {
		return nil, err
	}
	if len(tables) != 109 {
		return nil, fmt.Errorf("recovery catalog must classify exactly 109 authored tables, got %d", len(tables))
	}
	catalogNames := make([]string, 0, len(tables))
	requiredNames := []string{}
	for index, table := range tables {
		label := fmt.Sprintf("recovery catalog tables[%d]", index+1)
		name, err := requiredString(table, "table_name", label)
		if err != nil {
			return nil, err
		}
		catalogNames = append(catalogNames, name)
		inclusion, err := requiredString(table, "backup_inclusion", label)
		if err != nil {
			return nil, err
		}
		if inclusion == "authoritative_required" {
			requiredNames = append(requiredNames, name)
		}
	}
	if err := requireSortedUniqueStrings(catalogNames, "recovery catalog tables"); err != nil {
		return nil, err
	}
	if len(requiredNames) != 82 {
		return nil, fmt.Errorf("recovery catalog must contain exactly 82 authoritative_required tables, got %d", len(requiredNames))
	}
	authoredNames, err := authoredMigrationTableNames(root)
	if err != nil {
		return nil, err
	}
	if err := compareStringSlices(authoredNames, catalogNames, "authored migration tables and recovery catalog"); err != nil {
		return nil, err
	}
	objects, err := objectArray(catalog["object_families"], "recovery catalog object_families")
	if err != nil {
		return nil, err
	}
	wantObjectFamilies := []string{
		"evidence.blobs",
		"extensions.staged_objects",
		"imports.source_streams",
		"incident_bundles.files",
		"reference_packs.members",
		"reporting.render_preview_members",
	}
	gotObjectFamilies := make([]string, 0, len(objects))
	for index, object := range objects {
		familyID, err := requiredString(object, "object_family_id", fmt.Sprintf("object_families[%d]", index+1))
		if err != nil {
			return nil, err
		}
		gotObjectFamilies = append(gotObjectFamilies, familyID)
	}
	sort.Strings(gotObjectFamilies)
	if err := compareStringSlices(wantObjectFamilies, gotObjectFamilies, "recovery object families"); err != nil {
		return nil, err
	}
	return requiredNames, nil
}

func authoredMigrationTableNames(root string) ([]string, error) {
	migrationRoot := filepath.Join(root, "db", "migrations")
	entries, err := os.ReadDir(migrationRoot)
	if err != nil {
		return nil, fmt.Errorf("read migration root: %w", err)
	}
	names := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(migrationRoot, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		for _, match := range createTablePattern.FindAllSubmatch(raw, -1) {
			names[strings.ToLower(string(match[1]))] = struct{}{}
		}
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

func validateRecoverySnapshotFixture(snapshot map[string]any, requiredTables []string) error {
	units, err := objectArray(snapshot["units"], "postgres snapshot units")
	if err != nil {
		return err
	}
	if len(units) != 82 {
		return fmt.Errorf("postgres snapshot fixture must contain exactly 82 units, got %d", len(units))
	}
	names := make([]string, 0, len(units))
	for index, unit := range units {
		name, err := requiredString(unit, "table_name", fmt.Sprintf("postgres snapshot units[%d]", index+1))
		if err != nil {
			return err
		}
		names = append(names, name)
	}
	return compareStringSlices(requiredTables, names, "recovery catalog required tables and postgres snapshot units")
}

func compareStringSets(expected, actual map[string]struct{}, label string) error {
	expectedValues := make([]string, 0, len(expected))
	actualValues := make([]string, 0, len(actual))
	for value := range expected {
		expectedValues = append(expectedValues, value)
	}
	for value := range actual {
		actualValues = append(actualValues, value)
	}
	sort.Strings(expectedValues)
	sort.Strings(actualValues)
	return compareStringSlices(expectedValues, actualValues, label)
}

func compareStringSlices(expected, actual []string, label string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("%s differ: expected %d values, got %d", label, len(expected), len(actual))
	}
	for index := range expected {
		if expected[index] != actual[index] {
			return fmt.Errorf("%s differ at index %d: expected %s, got %s", label, index, expected[index], actual[index])
		}
	}
	return nil
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	databaseMigrationsIndexSchemaID    = "cartulary.database_migrations_contract_index.v2"
	migrationHistoryEvidenceSchemaID   = "cartulary.migration_history_evidence.v2"
	migrationHistoryEvidenceV1SchemaID = "cartulary.migration_history_evidence.v1"
)

var (
	migrationEvidenceSHA256Pattern  = regexp.MustCompile(`^[a-f0-9]{64}$`)
	migrationEvidenceFilename       = regexp.MustCompile(`^[0-9]{5}_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$`)
	migrationEvidenceLogicalID      = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	migrationEvidenceBindingKind    = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	migrationEvidenceServiceRef     = regexp.MustCompile(`^[A-Za-z0-9._:@+-]+$`)
	migrationEvidenceFindingReasons = stringSet(
		"manifest_schema_unsupported", "manifest_duplicate_version", "manifest_version_gap",
		"source_filename_invalid", "source_duplicate_version", "manifest_filename_mismatch",
		"manifest_hash_mismatch", "source_marker_missing", "future_phase_shaped_filename",
		"manifest_version_not_in_source", "source_version_not_in_manifest", "source_version_gap",
		"db_version_not_in_manifest", "db_applied_version_gap", "migration_metadata_missing",
		"protected_boundary_applied",
	)
)

func validateDatabaseMigrationsContractInput(relativePath string, value any) error {
	switch relativePath {
	case "index.json":
		return validateDatabaseMigrationsIndex(value)
	case "migration-history-evidence.v2.schema.json":
		return validateMigrationHistoryEvidenceSchema(value)
	case "fixtures/migration-history-evidence.v2.valid.json":
		return validateMigrationHistoryEvidenceV2(value)
	case "fixtures/migration-history-evidence.v1.rejected.json":
		object, err := asObject(value, "rejected v1 migration evidence fixture")
		if err != nil {
			return err
		}
		if object["schema_id"] != migrationHistoryEvidenceV1SchemaID {
			return fmt.Errorf("rejected v1 migration evidence fixture must declare v1")
		}
		if err := validateMigrationHistoryEvidenceV2(value); err == nil {
			return fmt.Errorf("rejected v1 migration evidence fixture was admitted as v2")
		}
		return nil
	case "fixtures/migration-history-evidence-negative.v2.json":
		return validateMigrationHistoryEvidenceNegativeFixtures(value)
	default:
		return fmt.Errorf("unexpected database-migrations artifact %s", relativePath)
	}
}

func validateDatabaseMigrationsIndex(value any) error {
	object, err := asObject(value, "contracts/database-migrations/index.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet(
		"$schema", "schema_id", "family_id", "contract_major", "owner_requirements",
		"current_schema_ids", "historical_reader_schema_ids", "contract_files", "fixtures",
		"compatibility_policy",
	), "database-migrations index"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "database-migrations index"); err != nil {
		return err
	}
	if object["schema_id"] != databaseMigrationsIndexSchemaID || object["family_id"] != "database-migrations" {
		return fmt.Errorf("database-migrations index identity is invalid")
	}
	major, ok := jsonInteger(object["contract_major"])
	if !ok || major != 2 {
		return fmt.Errorf("database-migrations index contract_major must be 2")
	}
	current, err := stringArray(object["current_schema_ids"], "database-migrations current_schema_ids", true)
	if err != nil || len(current) != 1 || current[0] != migrationHistoryEvidenceSchemaID {
		return fmt.Errorf("database-migrations index must declare only v2 as current")
	}
	historical, err := stringArray(object["historical_reader_schema_ids"], "database-migrations historical_reader_schema_ids", false)
	if err != nil || len(historical) != 0 {
		return fmt.Errorf("database-migrations index must declare no historical reader")
	}
	policy, err := asObject(object["compatibility_policy"], "database-migrations compatibility policy")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(policy, stringSet("v1_reader", "parser_aliases", "dual_output"), "database-migrations compatibility policy"); err != nil {
		return err
	}
	for _, key := range []string{"v1_reader", "parser_aliases", "dual_output"} {
		value, ok := policy[key].(bool)
		if !ok || value {
			return fmt.Errorf("database-migrations compatibility_policy.%s must be false", key)
		}
	}
	return nil
}

func validateMigrationHistoryEvidenceSchema(value any) error {
	object, err := asObject(value, "migration-history evidence v2 schema")
	if err != nil {
		return err
	}
	if err := requireDraftSchema(object, "migration-history evidence v2 schema"); err != nil {
		return err
	}
	if object["$id"] != migrationHistoryEvidenceSchemaID || object["additionalProperties"] != false {
		return fmt.Errorf("migration-history evidence schema must have the v2 id and close its root object")
	}
	return requireClosedObjectSchemas(object, "migration-history evidence v2 schema")
}

func requireClosedObjectSchemas(value any, label string) error {
	switch current := value.(type) {
	case map[string]any:
		if current["type"] == "object" && current["additionalProperties"] != false {
			return fmt.Errorf("%s contains an open object schema", label)
		}
		for key, child := range current {
			if err := requireClosedObjectSchemas(child, label+"."+key); err != nil {
				return err
			}
		}
	case []any:
		for index, child := range current {
			if err := requireClosedObjectSchemas(child, fmt.Sprintf("%s[%d]", label, index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateMigrationHistoryEvidenceV2(value any) error {
	object, err := asObject(value, "migration-history evidence v2")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet(
		"schema_id", "collected_at", "evidence_only", "rewrite_authorized", "database_binding",
		"manifest", "source_audit", "goose_ledger", "findings",
	), "migration-history evidence v2"); err != nil {
		return err
	}
	if len(object) != 9 || object["schema_id"] != migrationHistoryEvidenceSchemaID || object["evidence_only"] != true || object["rewrite_authorized"] != false {
		return fmt.Errorf("migration-history evidence v2 root identity or authorization fields are invalid")
	}
	if err := requireUTCTimestamp(object["collected_at"], "migration-history evidence collected_at"); err != nil {
		return err
	}
	if err := validateMigrationEvidenceDatabaseBinding(object["database_binding"]); err != nil {
		return err
	}
	if err := validateMigrationEvidenceManifest(object["manifest"]); err != nil {
		return err
	}
	audits, err := objectArray(object["source_audit"], "migration-history evidence source_audit")
	if err != nil {
		return err
	}
	for index, audit := range audits {
		if err := validateMigrationEvidenceSourceAudit(audit, fmt.Sprintf("source_audit[%d]", index)); err != nil {
			return err
		}
	}
	if err := validateMigrationEvidenceLedger(object["goose_ledger"]); err != nil {
		return err
	}
	findings, err := objectArray(object["findings"], "migration-history evidence findings")
	if err != nil {
		return err
	}
	for index, finding := range findings {
		if err := validateMigrationEvidenceFinding(finding, fmt.Sprintf("findings[%d]", index)); err != nil {
			return err
		}
	}
	return nil
}

func validateMigrationEvidenceDatabaseBinding(value any) error {
	object, err := asObject(value, "database_binding")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet("binding_kind", "service_ref"), "database_binding"); err != nil {
		return err
	}
	kind, err := requiredString(object, "binding_kind", "database_binding")
	if err != nil || !migrationEvidenceBindingKind.MatchString(kind) {
		return fmt.Errorf("database_binding.binding_kind is invalid")
	}
	if raw, present := object["service_ref"]; present {
		serviceRef, ok := raw.(string)
		if !ok || !migrationEvidenceServiceRef.MatchString(serviceRef) {
			return fmt.Errorf("database_binding.service_ref is invalid")
		}
	}
	return nil
}

func validateMigrationEvidenceManifest(value any) error {
	object, err := asObject(value, "manifest")
	if err != nil {
		return err
	}
	keys := stringSet("schema_id", "sha256", "migration_root", "immutable_through_version", "expected_min_version", "expected_max_version", "expected_version_count")
	if err := requireAllowedKeys(object, keys, "manifest"); err != nil || len(object) != len(keys) {
		return fmt.Errorf("manifest must contain only its seven logical members")
	}
	schemaID, err := requiredString(object, "schema_id", "manifest")
	if err != nil || !migrationEvidenceLogicalID.MatchString(schemaID) {
		return fmt.Errorf("manifest.schema_id is invalid")
	}
	sha, err := requiredString(object, "sha256", "manifest")
	if err != nil || !migrationEvidenceSHA256Pattern.MatchString(sha) {
		return fmt.Errorf("manifest.sha256 is invalid")
	}
	if object["migration_root"] != "db/migrations" {
		return fmt.Errorf("manifest.migration_root must retain its opaque canonical identity")
	}
	for _, key := range []string{"immutable_through_version", "expected_min_version", "expected_max_version", "expected_version_count"} {
		value, ok := jsonInteger(object[key])
		if !ok || value < 0 {
			return fmt.Errorf("manifest.%s must be a nonnegative integer", key)
		}
	}
	return nil
}

func validateMigrationEvidenceSourceAudit(object map[string]any, label string) error {
	allowed := stringSet("version", "filename", "sha256", "has_goose_up", "has_goose_down", "phase_shaped_name", "immutability_class", "manifest_filename", "manifest_sha256", "manifest_hash_matches")
	if err := requireAllowedKeys(object, allowed, label); err != nil {
		return err
	}
	for _, key := range []string{"version", "filename", "sha256", "has_goose_up", "has_goose_down", "phase_shaped_name", "immutability_class", "manifest_hash_matches"} {
		if _, present := object[key]; !present {
			return fmt.Errorf("%s.%s is required", label, key)
		}
	}
	version, ok := jsonInteger(object["version"])
	if !ok || version < 1 {
		return fmt.Errorf("%s.version must be positive", label)
	}
	filename, _ := object["filename"].(string)
	sha, _ := object["sha256"].(string)
	if !migrationEvidenceFilename.MatchString(filename) || !migrationEvidenceSHA256Pattern.MatchString(sha) {
		return fmt.Errorf("%s source identity is invalid", label)
	}
	for _, key := range []string{"has_goose_up", "has_goose_down", "phase_shaped_name", "manifest_hash_matches"} {
		if _, ok := object[key].(bool); !ok {
			return fmt.Errorf("%s.%s must be boolean", label, key)
		}
	}
	class, _ := object["immutability_class"].(string)
	if class != "protected" && class != "current" {
		return fmt.Errorf("%s.immutability_class is invalid", label)
	}
	manifestFilename, hasFilename := object["manifest_filename"]
	manifestSHA, hasSHA := object["manifest_sha256"]
	if hasFilename != hasSHA {
		return fmt.Errorf("%s manifest identity members must be paired", label)
	}
	if hasFilename {
		filename, filenameOK := manifestFilename.(string)
		sha, shaOK := manifestSHA.(string)
		if !filenameOK || !shaOK || !migrationEvidenceFilename.MatchString(filename) || !migrationEvidenceSHA256Pattern.MatchString(sha) {
			return fmt.Errorf("%s manifest identity is invalid", label)
		}
	}
	return nil
}

func validateMigrationEvidenceLedger(value any) error {
	object, err := asObject(value, "goose_ledger")
	if err != nil {
		return err
	}
	keys := stringSet("metadata_present", "row_count", "current_effective_applied_version", "latest_effective_states")
	if err := requireAllowedKeys(object, keys, "goose_ledger"); err != nil || len(object) != len(keys) {
		return fmt.Errorf("goose_ledger must contain exactly four members")
	}
	metadata, ok := object["metadata_present"].(bool)
	if !ok {
		return fmt.Errorf("goose_ledger.metadata_present must be boolean")
	}
	for _, key := range []string{"row_count", "current_effective_applied_version"} {
		value, ok := jsonInteger(object[key])
		if !ok || value < 0 {
			return fmt.Errorf("goose_ledger.%s must be nonnegative", key)
		}
	}
	if !metadata {
		if object["latest_effective_states"] != nil || object["row_count"].(json.Number).String() != "0" || object["current_effective_applied_version"].(json.Number).String() != "0" {
			return fmt.Errorf("missing goose metadata must retain null states and zero counts")
		}
		return nil
	}
	states, err := objectArray(object["latest_effective_states"], "goose_ledger.latest_effective_states")
	if err != nil {
		return err
	}
	for index, state := range states {
		label := fmt.Sprintf("goose_ledger.latest_effective_states[%d]", index)
		if err := requireAllowedKeys(state, stringSet("version", "is_applied", "tstamp"), label); err != nil || len(state) != 3 {
			return fmt.Errorf("%s must contain exactly three members", label)
		}
		if _, ok := jsonInteger(state["version"]); !ok {
			return fmt.Errorf("%s.version must be an integer", label)
		}
		if _, ok := state["is_applied"].(bool); !ok {
			return fmt.Errorf("%s.is_applied must be boolean", label)
		}
		if err := requireUTCTimestamp(state["tstamp"], label+".tstamp"); err != nil {
			return err
		}
	}
	return nil
}

func validateMigrationEvidenceFinding(object map[string]any, label string) error {
	if err := requireAllowedKeys(object, stringSet("severity", "reason_code", "version", "filename", "detail"), label); err != nil {
		return err
	}
	severity, _ := object["severity"].(string)
	if severity != "blocking" && severity != "warning" && severity != "info" {
		return fmt.Errorf("%s.severity is invalid", label)
	}
	reason, _ := object["reason_code"].(string)
	if _, ok := migrationEvidenceFindingReasons[reason]; !ok {
		return fmt.Errorf("%s.reason_code is invalid", label)
	}
	detail, err := requiredString(object, "detail", label)
	if err != nil || len(detail) > 1024 || strings.ContainsAny(detail, `/\`) {
		return fmt.Errorf("%s.detail is invalid or contains a locator", label)
	}
	if raw, present := object["version"]; present {
		version, ok := jsonInteger(raw)
		if !ok || version < 1 {
			return fmt.Errorf("%s.version must be positive", label)
		}
	}
	if raw, present := object["filename"]; present {
		filename, ok := raw.(string)
		if !ok || !migrationEvidenceFilename.MatchString(filename) {
			return fmt.Errorf("%s.filename is invalid", label)
		}
	}
	return nil
}

func requireUTCTimestamp(value any, label string) error {
	raw, ok := value.(string)
	if !ok || !strings.HasSuffix(raw, "Z") {
		return fmt.Errorf("%s must be an RFC3339 UTC timestamp", label)
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil || parsed.Location() != time.UTC {
		return fmt.Errorf("%s must be an RFC3339 UTC timestamp", label)
	}
	return nil
}

func validateMigrationHistoryEvidenceNegativeFixtures(value any) error {
	object, err := asObject(value, "migration-history evidence negative fixtures")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet("schema_id", "cases"), "migration-history evidence negative fixtures"); err != nil {
		return err
	}
	if object["schema_id"] != "cartulary.migration_history_evidence_negative_fixtures.v2" {
		return fmt.Errorf("migration-history evidence negative fixture identity is invalid")
	}
	cases, err := objectArray(object["cases"], "migration-history evidence negative cases")
	if err != nil || len(cases) != 11 {
		return fmt.Errorf("migration-history evidence negative fixtures must contain exactly 11 cases")
	}
	want := []string{
		"v1_schema_rejected", "manifest_path_rejected", "manifest_source_path_rejected",
		"manifest_repository_path_rejected", "manifest_embedded_path_rejected", "manifest_file_rejected",
		"manifest_uri_rejected", "manifest_path_hash_rejected", "unknown_top_level_rejected",
		"absolute_detail_rejected", "absolute_service_ref_rejected",
	}
	actual := make([]string, 0, len(cases))
	for _, current := range cases {
		caseID, err := requiredString(current, "case_id", "migration-history evidence negative case")
		if err != nil {
			return err
		}
		actual = append(actual, caseID)
	}
	if strings.Join(actual, "\x00") != strings.Join(want, "\x00") {
		return fmt.Errorf("migration-history evidence negative case set or order is invalid")
	}
	return nil
}

func validateDatabaseMigrationsContractFamily(root string) error {
	base := filepath.Join(root, "contracts", "database-migrations")
	expected := []string{
		"fixtures/migration-history-evidence-negative.v2.json",
		"fixtures/migration-history-evidence.v1.rejected.json",
		"fixtures/migration-history-evidence.v2.valid.json",
		"index.json",
		"migration-history-evidence.v2.schema.json",
	}
	actual := []string{}
	if err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		actual = append(actual, filepath.ToSlash(relative))
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(actual)
	if err := compareStringSlices(expected, actual, "database-migrations contract paths"); err != nil {
		return err
	}
	valid, err := readDatabaseMigrationsObject(base, "fixtures/migration-history-evidence.v2.valid.json")
	if err != nil {
		return err
	}
	negative, err := readDatabaseMigrationsObject(base, "fixtures/migration-history-evidence-negative.v2.json")
	if err != nil {
		return err
	}
	for _, current := range negative["cases"].([]any) {
		caseObject := current.(map[string]any)
		mutated, err := cloneMigrationEvidenceObject(valid)
		if err != nil {
			return err
		}
		applyMigrationEvidenceMutation(mutated, caseObject)
		if err := validateMigrationHistoryEvidenceV2(mutated); err == nil {
			return fmt.Errorf("negative migration evidence case %s was admitted", caseObject["case_id"])
		}
	}
	raw, err := os.ReadFile(filepath.Join(base, "fixtures", "migration-history-evidence.v2.valid.json"))
	if err != nil {
		return err
	}
	return requireJSONMemberOrder(raw, []string{
		"schema_id", "collected_at", "evidence_only", "rewrite_authorized", "database_binding",
		"manifest", "source_audit", "goose_ledger", "findings",
	})
}

func readDatabaseMigrationsObject(base string, relative string) (map[string]any, error) {
	raw, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(relative)))
	if err != nil {
		return nil, err
	}
	decoded, err := decodeContract(raw)
	if err != nil {
		return nil, err
	}
	return asObject(decoded, relative)
}

func cloneMigrationEvidenceObject(value map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoded, err := decodeContract(raw)
	if err != nil {
		return nil, err
	}
	return asObject(decoded, "cloned migration evidence")
}

func applyMigrationEvidenceMutation(value map[string]any, current map[string]any) {
	mutation, _ := current["mutation"].(string)
	switch mutation {
	case "schema_id_v1":
		value["schema_id"] = migrationHistoryEvidenceV1SchemaID
	case "add_manifest_member":
		value["manifest"].(map[string]any)[current["member"].(string)] = current["value"]
	case "add_top_level_member":
		value[current["member"].(string)] = current["value"]
	case "absolute_finding_detail":
		value["findings"].([]any)[0].(map[string]any)["detail"] = current["value"]
	case "absolute_service_ref":
		value["database_binding"].(map[string]any)["service_ref"] = current["value"]
	}
}

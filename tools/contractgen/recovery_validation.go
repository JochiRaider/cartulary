package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	recoveryRegistrySchemaID           = "cartulary.recovery_contract_registry.v1"
	recoveryCatalogSchemaID            = "cartulary.recovery_state_catalog.v1"
	recoveryGenerationRegistrySchemaID = "cartulary.recovery_generation_registry.v1"
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
		"backup-artifact-envelope.v2.schema.json":                        "cartulary.backup_artifact_envelope.v2",
		"backup-integrity-manifest.v3.schema.json":                       "cartulary.backup_integrity_manifest.v3",
		"common.v1.schema.json":                                          "cartulary.recovery_common.v1",
		"graph-projection-restore-implementation-binding.v2.schema.json": "cartulary.graph_projection_restore_implementation_binding.v2",
		"graph-projection-restore-implementation-binding.v3.schema.json": "cartulary.graph_projection_restore_implementation_binding.v3",
		"graph-projection-restore-rebuild-result.v2.schema.json":         "cartulary.graph_projection_restore_rebuild_result.v2",
		"graph-projection-restore-rebuild-result.v3.schema.json":         "cartulary.graph_projection_restore_rebuild_result.v3",
		"graph-projection-restore-source-registry.v2.schema.json":        "cartulary.graph_projection_restore_source_registry.v2",
		"graph-projection-restore-source-registry.v3.schema.json":        "cartulary.graph_projection_restore_source_registry.v3",
		"object-store-backup-manifest.v2.schema.json":                    "cartulary.object_store_backup_manifest.v2",
		"object-store-backup-summary.v2.schema.json":                     "cartulary.object_store_backup_summary.v2",
		"operator-recovery-audit-summary.v2.schema.json":                 "cartulary.operator_recovery_audit_summary.v2",
		"operator-recovery-journal-payload.v2.schema.json":               "cartulary.operator_recovery_journal_payload.v2",
		"operator-recovery-journal-payload.v3.schema.json":               "cartulary.operator_recovery_journal_payload.v3",
		"operator-recovery-journal-payload.v4.schema.json":               "cartulary.operator_recovery_journal_payload.v4",
		"postgres-snapshot-artifact.v2.schema.json":                      "cartulary.postgres_snapshot_artifact.v2",
		"postgres-snapshot-unit.v1.schema.json":                          "cartulary.postgres_snapshot_unit.v1",
		"recovery-generation-registry.v1.schema.json":                    recoveryGenerationRegistrySchemaID,
		"recovery-state-catalog.v1.schema.json":                          recoveryCatalogSchemaID,
		"recovery-state-contribution.v1.schema.json":                     "cartulary.recovery_state_contribution.v1",
		"restore-target-marker.v2.schema.json":                           "cartulary.restore_target_marker.v2",
		"restore-verification.v2.schema.json":                            "cartulary.restore_verification.v2",
		"restore-workbook-probe-registration.v1.schema.json":             "cartulary.restore_workbook_probe_registration.v1",
	}
	recoveryFixtureIDsByPath = map[string]string{
		"fixtures/backup-artifact-envelope.v2.json":                                               "cartulary.backup_artifact_envelope.v2",
		"fixtures/backup-integrity-manifest.v3.json":                                              "cartulary.backup_integrity_manifest.v3",
		"fixtures/graph-projection-restore-implementation-binding.v2.json":                        "cartulary.graph_projection_restore_implementation_binding.v2",
		"fixtures/graph-projection-restore-implementation-binding.v3.json":                        "cartulary.graph_projection_restore_implementation_binding.v3",
		"fixtures/graph-projection-restore-implementation-binding.v3.pre-workbook-ownership.json": "cartulary.graph_projection_restore_implementation_binding.v3",
		"fixtures/graph-projection-restore-rebuild-result.v2.json":                                "cartulary.graph_projection_restore_rebuild_result.v2",
		"fixtures/graph-projection-restore-rebuild-result.v3.json":                                "cartulary.graph_projection_restore_rebuild_result.v3",
		"fixtures/graph-projection-restore-source-registry.v2.json":                               "cartulary.graph_projection_restore_source_registry.v2",
		"fixtures/graph-projection-restore-source-registry.v3.json":                               "cartulary.graph_projection_restore_source_registry.v3",
		"fixtures/object-store-backup-manifest.v2.json":                                           "cartulary.object_store_backup_manifest.v2",
		"fixtures/object-store-backup-summary.v2.json":                                            "cartulary.object_store_backup_summary.v2",
		"fixtures/operator-recovery-audit-summary.v2.json":                                        "cartulary.operator_recovery_audit_summary.v2",
		"fixtures/operator-recovery-journal-payload.v2.json":                                      "cartulary.operator_recovery_journal_payload.v2",
		"fixtures/operator-recovery-journal-payload.v3.json":                                      "cartulary.operator_recovery_journal_payload.v3",
		"fixtures/operator-recovery-journal-payload.v4.json":                                      "cartulary.operator_recovery_journal_payload.v4",
		"fixtures/postgres-snapshot-artifact.v2.json":                                             "cartulary.postgres_snapshot_artifact.v2",
		"fixtures/postgres-snapshot-unit.v1.json":                                                 "cartulary.postgres_snapshot_unit.v1",
		"fixtures/recovery-generation-registry.v1.json":                                           recoveryGenerationRegistrySchemaID,
		"fixtures/recovery-state-catalog.v1.graph-v2.json":                                        recoveryCatalogSchemaID,
		"fixtures/recovery-state-catalog.v1.json":                                                 recoveryCatalogSchemaID,
		"fixtures/recovery-state-catalog.v1.pre-workbook-ownership.json":                          recoveryCatalogSchemaID,
		"fixtures/recovery-state-contribution.v1.json":                                            "cartulary.recovery_state_contribution.v1",
		"fixtures/restore-target-marker.v2.json":                                                  "cartulary.restore_target_marker.v2",
		"fixtures/restore-verification.v2.json":                                                   "cartulary.restore_verification.v2",
		"fixtures/restore-workbook-probe-registration.v1.json":                                    "cartulary.restore_workbook_probe_registration.v1",
	}
	createTablePattern = regexp.MustCompile(`(?i)\bCREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+(?:public\.)?([a-z][a-z0-9_]*)`)
	dropTablePattern   = regexp.MustCompile(`(?i)\bDROP\s+TABLE(?:\s+IF\s+EXISTS)?\s+(?:public\.)?([a-z][a-z0-9_]*)`)
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
	if err := validateGraphProjectionRestoreV2Contracts(base); err != nil {
		return err
	}
	if err := validateGraphProjectionRestoreV3Contracts(base); err != nil {
		return err
	}
	if err := validateRecoveryGenerationRegistry(base); err != nil {
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

func validateRecoveryGenerationRegistry(base string) error {
	registry, err := readRecoveryObject(base, "fixtures/recovery-generation-registry.v1.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(registry, stringSet("schema_id", "generations"), "Recovery generation registry"); err != nil {
		return err
	}
	if registry["schema_id"] != recoveryGenerationRegistrySchemaID {
		return fmt.Errorf("recovery generation registry has the wrong schema_id")
	}
	generations, err := objectArray(registry["generations"], "Recovery generation registry generations")
	if err != nil {
		return err
	}
	wantIDs := []string{
		"recovery.current.workbook_owned.graph_v3",
		"recovery.historical.incidents_owned.graph_v3",
		"recovery.historical.graph_v2",
	}
	wantCatalogPaths := []string{
		"fixtures/recovery-state-catalog.v1.json",
		"fixtures/recovery-state-catalog.v1.pre-workbook-ownership.json",
		"fixtures/recovery-state-catalog.v1.graph-v2.json",
	}
	wantBindingPaths := []string{
		"fixtures/graph-projection-restore-implementation-binding.v3.json",
		"fixtures/graph-projection-restore-implementation-binding.v3.pre-workbook-ownership.json",
		"fixtures/graph-projection-restore-implementation-binding.v2.json",
	}
	if len(generations) != len(wantIDs) {
		return fmt.Errorf("recovery generation registry must contain exactly three generations")
	}
	seenPairs := map[string]struct{}{}
	for index, generation := range generations {
		label := fmt.Sprintf("Recovery generation registry generations[%d]", index+1)
		if err := requireAllowedKeys(generation, stringSet("generation_id", "capture_current", "catalog", "codec_registry", "graph"), label); err != nil {
			return err
		}
		generationID, err := requiredString(generation, "generation_id", label)
		if err != nil {
			return err
		}
		captureCurrent, err := requiredBool(generation, "capture_current", label)
		if err != nil {
			return err
		}
		if generationID != wantIDs[index] || captureCurrent != (index == 0) {
			return fmt.Errorf("%s is not the exact generation at position %d", generationID, index+1)
		}
		catalog, err := asObject(generation["catalog"], label+".catalog")
		if err != nil {
			return err
		}
		catalogDigest, err := validateRecoveryGenerationCatalogBinding(base, catalog, wantCatalogPaths[index], label+".catalog")
		if err != nil {
			return err
		}
		codecs, err := asObject(generation["codec_registry"], label+".codec_registry")
		if err != nil {
			return err
		}
		codecDigest, err := validateRecoveryGenerationCodecBinding(codecs, label+".codec_registry")
		if err != nil {
			return err
		}
		pair := catalogDigest + "\x00" + codecDigest
		if _, duplicate := seenPairs[pair]; duplicate {
			return fmt.Errorf("recovery generation registry duplicates a catalog/codec lookup pair")
		}
		seenPairs[pair] = struct{}{}
		graph, err := asObject(generation["graph"], label+".graph")
		if err != nil {
			return err
		}
		if err := validateRecoveryGenerationGraphBinding(base, graph, wantBindingPaths[index], catalogDigest, label+".graph"); err != nil {
			return err
		}
	}
	return nil
}

func validateRecoveryGenerationCatalogBinding(
	base string,
	binding map[string]any,
	wantPath string,
	label string,
) (string, error) {
	keys := stringSet(
		"artifact_path", "schema_id", "canonical_sha256", "catalog_digest_sha256",
		"contribution_count", "authored_table_count", "required_table_count", "object_family_count",
	)
	if err := requireAllowedKeys(binding, keys, label); err != nil {
		return "", err
	}
	path, err := requiredString(binding, "artifact_path", label)
	if err != nil {
		return "", err
	}
	if path != wantPath || binding["schema_id"] != recoveryCatalogSchemaID {
		return "", fmt.Errorf("%s does not bind the required immutable catalog", label)
	}
	catalog, canonicalSHA256, err := readRecoveryCanonicalArtifact(base, path)
	if err != nil {
		return "", err
	}
	if binding["canonical_sha256"] != canonicalSHA256 || catalog["schema_id"] != recoveryCatalogSchemaID {
		return "", fmt.Errorf("%s catalog canonical bytes or schema mismatch", label)
	}
	catalogDigest, err := requiredString(catalog, "catalog_digest_sha256", "contracts/recovery/"+path)
	if err != nil {
		return "", err
	}
	if binding["catalog_digest_sha256"] != catalogDigest {
		return "", fmt.Errorf("%s catalog digest does not match the frozen document", label)
	}
	contributions, err := objectArray(catalog["contribution_digests"], label+" contribution digests")
	if err != nil {
		return "", err
	}
	tables, err := objectArray(catalog["tables"], label+" tables")
	if err != nil {
		return "", err
	}
	objects, err := objectArray(catalog["object_families"], label+" object families")
	if err != nil {
		return "", err
	}
	required := int64(0)
	for _, table := range tables {
		if table["backup_inclusion"] == "authoritative_required" {
			required++
		}
	}
	wantCounts := map[string]int64{
		"contribution_count":   int64(len(contributions)),
		"authored_table_count": int64(len(tables)),
		"required_table_count": required,
		"object_family_count":  int64(len(objects)),
	}
	for field, want := range wantCounts {
		got, err := requiredInt(binding, field, label)
		if err != nil {
			return "", err
		}
		if got != want {
			return "", fmt.Errorf("%s.%s = %d, want %d from frozen catalog", label, field, got, want)
		}
	}
	return catalogDigest, nil
}

func validateRecoveryGenerationCodecBinding(binding map[string]any, label string) (string, error) {
	if err := requireAllowedKeys(binding, stringSet("schema_ids", "digest_sha256"), label); err != nil {
		return "", err
	}
	schemaIDs, err := stringArray(binding["schema_ids"], label+".schema_ids", true)
	if err != nil {
		return "", err
	}
	if err := requireSortedUniqueStrings(schemaIDs, label+".schema_ids"); err != nil {
		return "", err
	}
	digest, err := requiredString(binding, "digest_sha256", label)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte("CARTULARY-RECOVERY-CODEC-REGISTRY-VNEXT\n" + strings.Join(schemaIDs, "\n") + "\n"))
	if digest != hex.EncodeToString(sum[:]) {
		return "", fmt.Errorf("%s digest does not match its sorted schema IDs", label)
	}
	return digest, nil
}

func validateRecoveryGenerationGraphBinding(
	base string,
	graph map[string]any,
	wantBindingPath string,
	catalogDigest string,
	label string,
) error {
	if err := requireAllowedKeys(graph, stringSet("algorithm_id", "source_registry", "implementation_binding"), label); err != nil {
		return err
	}
	algorithmID, err := requiredString(graph, "algorithm_id", label)
	if err != nil {
		return err
	}
	for field, wantPath := range map[string]string{
		"source_registry":        pathForGraphGeneration(wantBindingPath),
		"implementation_binding": wantBindingPath,
	} {
		binding, err := asObject(graph[field], label+"."+field)
		if err != nil {
			return err
		}
		if err := requireAllowedKeys(binding, stringSet("artifact_path", "schema_id", "canonical_sha256"), label+"."+field); err != nil {
			return err
		}
		path, err := requiredString(binding, "artifact_path", label+"."+field)
		if err != nil {
			return err
		}
		if path != wantPath {
			return fmt.Errorf("%s.%s does not bind the exact frozen artifact", label, field)
		}
		artifact, canonicalSHA256, err := readRecoveryCanonicalArtifact(base, path)
		if err != nil {
			return err
		}
		if binding["canonical_sha256"] != canonicalSHA256 || binding["schema_id"] != artifact["schema_id"] {
			return fmt.Errorf("%s.%s canonical bytes or schema mismatch", label, field)
		}
		if field == "implementation_binding" &&
			(artifact["algorithm_id"] != algorithmID || artifact["recovery_state_catalog_sha256"] != catalogDigest) {
			return fmt.Errorf("%s implementation binding does not match its algorithm and catalog", label)
		}
	}
	return nil
}

func pathForGraphGeneration(bindingPath string) string {
	if strings.Contains(bindingPath, ".v2.") {
		return "fixtures/graph-projection-restore-source-registry.v2.json"
	}
	return "fixtures/graph-projection-restore-source-registry.v3.json"
}

func readRecoveryCanonicalArtifact(base string, relativePath string) (map[string]any, string, error) {
	artifact, err := readRecoveryObject(base, relativePath)
	if err != nil {
		return nil, "", err
	}
	canonical, err := canonicalizeDecoded(artifact)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256([]byte(canonical))
	return artifact, hex.EncodeToString(sum[:]), nil
}

func validateGraphProjectionRestoreV2Contracts(base string) error {
	registry, err := readRecoveryObject(base, "fixtures/graph-projection-restore-source-registry.v2.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(registry, stringSet("schema_id", "entries"), "Graph Projection v2 restore source registry"); err != nil {
		return err
	}
	entries, err := objectArray(registry["entries"], "Graph Projection v2 restore source registry entries")
	if err != nil {
		return err
	}
	if len(entries) != 1 ||
		entries[0]["source_registration_id"] != "network_flow_activity.graph_views.v1" ||
		entries[0]["source_owner_id"] != "network_flow_activity" ||
		entries[0]["authoritative_family_id"] != "network_flow_activity.graph_views" ||
		entries[0]["projection_input_contract_id"] != "graph_projection.v2" ||
		entries[0]["projection_result_contract_id"] != "graph_projection_result.v2" ||
		entries[0]["status"] != "active" {
		return fmt.Errorf("current Graph Projection v2 restore registry must contain the exact Network Flow owner binding")
	}
	registryJSON, err := canonicalizeDecoded(registry)
	if err != nil {
		return err
	}
	registrySum := sha256.Sum256([]byte(registryJSON))
	registrySHA256 := hex.EncodeToString(registrySum[:])

	binding, err := readRecoveryObject(base, "fixtures/graph-projection-restore-implementation-binding.v2.json")
	if err != nil {
		return err
	}
	bindingKeys := stringSet(
		"schema_id", "algorithm_id", "binding_id", "graph_projection_contract_id",
		"recovery_state_catalog_sha256", "source_registry_sha256", "graph_table_ids",
		"graph_engine_algorithm_ids", "graph_engine_algorithm_digests", "database_schema_lineage",
		"database_schema_head", "packaged_subject_sha256", "build_provenance_sha256",
	)
	if err := requireAllowedKeys(binding, bindingKeys, "Graph Projection v2 restore implementation binding"); err != nil {
		return err
	}
	if len(binding) != len(bindingKeys) ||
		binding["schema_id"] != "cartulary.graph_projection_restore_implementation_binding.v2" ||
		binding["algorithm_id"] != "graphprojection.restore_rebuild.v2" ||
		binding["binding_id"] != "graphprojection.restore_rebuild.network_flow_graph_views.v2" ||
		binding["graph_projection_contract_id"] != "cartulary.graph_projection_nlspec.v2.0.0" ||
		binding["source_registry_sha256"] != registrySHA256 ||
		binding["database_schema_lineage"] != "cartulary.prod_ddl_rebaseline.v2" {
		return fmt.Errorf("current Graph Projection v2 restore implementation binding has drifted from its authored source set")
	}
	wantGraphTables := []string{
		"graph_projection_result_edges",
		"graph_projection_result_leases",
		"graph_projection_result_vertices",
		"graph_projection_results",
	}
	graphTables, err := stringArray(binding["graph_table_ids"], "Graph Projection v2 restore graph_table_ids", true)
	if err != nil {
		return err
	}
	if err := compareStringSlices(wantGraphTables, graphTables, "Graph Projection v2 restore graph tables"); err != nil {
		return err
	}
	algorithmIDs, err := stringArray(binding["graph_engine_algorithm_ids"], "Graph Projection v2 restore algorithm IDs", true)
	if err != nil {
		return err
	}
	if err := requireSortedUniqueStrings(algorithmIDs, "Graph Projection v2 restore algorithm IDs"); err != nil {
		return err
	}
	algorithmDigests, err := stringArray(binding["graph_engine_algorithm_digests"], "Graph Projection v2 restore algorithm digests", true)
	if err != nil {
		return err
	}
	if len(algorithmIDs) != len(algorithmDigests) {
		return fmt.Errorf("graph projection v2 restore algorithm identities and digests must align one-to-one")
	}

	result, err := readRecoveryObject(base, "fixtures/graph-projection-restore-rebuild-result.v2.json")
	if err != nil {
		return err
	}
	resultKeys := stringSet(
		"schema_id", "restore_operation_id", "target_generation_id", "status", "readiness_outcome",
		"algorithm_id", "implementation_binding_sha256", "source_registry_sha256", "cleared_table_ids",
		"rebuilt_views", "reconciled_nonterminal_job_count", "reconciled_lease_count", "postcondition_sha256",
		"warnings", "errors",
	)
	if err := requireAllowedKeys(result, resultKeys, "Graph Projection v2 restore result"); err != nil {
		return err
	}
	if len(result) != len(resultKeys) || result["status"] != "succeeded" || result["readiness_outcome"] != "ready" || result["source_registry_sha256"] != registrySHA256 {
		return fmt.Errorf("graph projection v2 restore result fixture must be the exact successful current result shape")
	}
	cleared, err := stringArray(result["cleared_table_ids"], "Graph Projection v2 restore cleared tables", true)
	if err != nil {
		return err
	}
	return compareStringSlices(wantGraphTables, cleared, "Graph Projection v2 restore cleared tables")
}

func validateGraphProjectionRestoreV3Contracts(base string) error {
	registry, err := readRecoveryObject(base, "fixtures/graph-projection-restore-source-registry.v3.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(registry, stringSet("schema_id", "entries"), "Graph Projection v3 restore source registry"); err != nil {
		return err
	}
	entries, err := objectArray(registry["entries"], "Graph Projection v3 restore source registry entries")
	if err != nil {
		return err
	}
	if len(entries) != 1 {
		return fmt.Errorf("current Graph Projection v3 restore registry must contain one Network Flow owner binding")
	}
	entry := entries[0]
	entryKeys := stringSet(
		"source_registration_id", "source_owner_id", "authoritative_family_id", "enumerator_binding_id",
		"validity_binding_id", "semantic_query_schema_ids", "projection_input_contract_id",
		"projection_result_contract_id", "status",
	)
	if err := requireAllowedKeys(entry, entryKeys, "Graph Projection v3 restore source registry entry"); err != nil {
		return err
	}
	semanticQuerySchemaIDs, err := stringArray(entry["semantic_query_schema_ids"], "Graph Projection v3 semantic query schemas", true)
	if err != nil {
		return err
	}
	if err := compareStringSlices([]string{
		"cartulary.network_flow.graph_semantic_query.v1",
		"cartulary.network_flow.graph_semantic_query.v2",
	}, semanticQuerySchemaIDs, "Graph Projection v3 semantic query schemas"); err != nil {
		return err
	}
	if entry["source_registration_id"] != "network_flow_activity.graph_views.v1" ||
		entry["source_owner_id"] != "network_flow_activity" ||
		entry["authoritative_family_id"] != "network_flow_activity.graph_views" ||
		entry["projection_input_contract_id"] != "graph_projection.v2" ||
		entry["projection_result_contract_id"] != "graph_projection_result.v2" ||
		entry["status"] != "active" {
		return fmt.Errorf("current Graph Projection v3 restore registry has drifted from its mixed-generation owner binding")
	}
	registryJSON, err := canonicalizeDecoded(registry)
	if err != nil {
		return err
	}
	registrySum := sha256.Sum256([]byte(registryJSON))
	registrySHA256 := hex.EncodeToString(registrySum[:])

	binding, err := readRecoveryObject(base, "fixtures/graph-projection-restore-implementation-binding.v3.json")
	if err != nil {
		return err
	}
	bindingKeys := stringSet(
		"schema_id", "algorithm_id", "binding_id", "graph_projection_contract_id", "semantic_query_schema_ids",
		"historical_dispatch_algorithm_ids", "recovery_state_catalog_sha256", "source_registry_sha256",
		"graph_table_ids", "graph_engine_algorithm_ids", "graph_engine_algorithm_digests", "database_schema_lineage",
		"database_schema_head", "packaged_subject_sha256", "build_provenance_sha256",
	)
	if err := requireAllowedKeys(binding, bindingKeys, "Graph Projection v3 restore implementation binding"); err != nil {
		return err
	}
	if len(binding) != len(bindingKeys) ||
		binding["schema_id"] != "cartulary.graph_projection_restore_implementation_binding.v3" ||
		binding["algorithm_id"] != "graphprojection.restore_rebuild.v3" ||
		binding["binding_id"] != "graphprojection.restore_rebuild.network_flow_graph_views.v3" ||
		binding["graph_projection_contract_id"] != "cartulary.graph_projection_nlspec.v2.1.0" ||
		binding["source_registry_sha256"] != registrySHA256 ||
		binding["database_schema_lineage"] != "cartulary.prod_ddl_rebaseline.v2" {
		return fmt.Errorf("current Graph Projection v3 restore implementation binding has drifted from its authored source set")
	}
	bindingSemanticSchemas, err := stringArray(binding["semantic_query_schema_ids"], "Graph Projection v3 binding semantic query schemas", true)
	if err != nil {
		return err
	}
	if err := compareStringSlices(semanticQuerySchemaIDs, bindingSemanticSchemas, "Graph Projection v3 registry/binding semantic query schemas"); err != nil {
		return err
	}
	historicalDispatch, err := stringArray(binding["historical_dispatch_algorithm_ids"], "Graph Projection v3 historical dispatch", true)
	if err != nil {
		return err
	}
	if err := compareStringSlices([]string{"graphprojection.restore_rebuild.v2"}, historicalDispatch, "Graph Projection v3 historical dispatch"); err != nil {
		return err
	}
	wantGraphTables := []string{
		"graph_projection_result_edges",
		"graph_projection_result_leases",
		"graph_projection_result_vertices",
		"graph_projection_results",
	}
	graphTables, err := stringArray(binding["graph_table_ids"], "Graph Projection v3 restore graph_table_ids", true)
	if err != nil {
		return err
	}
	if err := compareStringSlices(wantGraphTables, graphTables, "Graph Projection v3 restore graph tables"); err != nil {
		return err
	}
	algorithmIDs, err := stringArray(binding["graph_engine_algorithm_ids"], "Graph Projection v3 restore algorithm IDs", true)
	if err != nil {
		return err
	}
	if err := requireSortedUniqueStrings(algorithmIDs, "Graph Projection v3 restore algorithm IDs"); err != nil {
		return err
	}
	algorithmDigests, err := stringArray(binding["graph_engine_algorithm_digests"], "Graph Projection v3 restore algorithm digests", true)
	if err != nil {
		return err
	}
	if len(algorithmIDs) != len(algorithmDigests) {
		return fmt.Errorf("graph projection v3 restore algorithm identities and digests must align one-to-one")
	}
	bindingJSON, err := canonicalizeDecoded(binding)
	if err != nil {
		return err
	}
	bindingSum := sha256.Sum256([]byte(bindingJSON))
	bindingSHA256 := hex.EncodeToString(bindingSum[:])

	result, err := readRecoveryObject(base, "fixtures/graph-projection-restore-rebuild-result.v3.json")
	if err != nil {
		return err
	}
	resultKeys := stringSet(
		"schema_id", "restore_operation_id", "target_generation_id", "status", "readiness_outcome",
		"algorithm_id", "implementation_binding_sha256", "source_registry_sha256", "cleared_table_ids",
		"rebuilt_views", "reconciled_nonterminal_job_count", "reconciled_lease_count", "postcondition_sha256",
		"warnings", "errors",
	)
	if err := requireAllowedKeys(result, resultKeys, "Graph Projection v3 restore result"); err != nil {
		return err
	}
	if len(result) != len(resultKeys) || result["schema_id"] != "cartulary.graph_projection_restore_rebuild_result.v3" ||
		result["algorithm_id"] != "graphprojection.restore_rebuild.v3" || result["status"] != "succeeded" ||
		result["readiness_outcome"] != "ready" || result["source_registry_sha256"] != registrySHA256 ||
		result["implementation_binding_sha256"] != bindingSHA256 {
		return fmt.Errorf("graph projection v3 restore result fixture must bind the current source registry and implementation")
	}
	views, err := objectArray(result["rebuilt_views"], "Graph Projection v3 rebuilt views")
	if err != nil {
		return err
	}
	seenSemanticSchemas := map[string]struct{}{}
	for _, view := range views {
		seenSemanticSchemas[stringValue(view["semantic_query_schema_id"])] = struct{}{}
	}
	if len(seenSemanticSchemas) != 2 {
		return fmt.Errorf("graph projection v3 result fixture must cover both persisted semantic query generations")
	}
	cleared, err := stringArray(result["cleared_table_ids"], "Graph Projection v3 restore cleared tables", true)
	if err != nil {
		return err
	}
	return compareStringSlices(wantGraphTables, cleared, "Graph Projection v3 restore cleared tables")
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
	if len(tables) != 113 {
		return nil, fmt.Errorf("recovery catalog must classify exactly 113 authored tables, got %d", len(tables))
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
	if len(requiredNames) != 84 {
		return nil, fmt.Errorf("recovery catalog must contain exactly 84 authoritative_required tables, got %d", len(requiredNames))
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
		up := strings.SplitN(string(raw), "-- +goose Down", 2)[0]
		for _, match := range createTablePattern.FindAllStringSubmatch(up, -1) {
			names[strings.ToLower(match[1])] = struct{}{}
		}
		for _, match := range dropTablePattern.FindAllStringSubmatch(up, -1) {
			delete(names, strings.ToLower(match[1]))
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
	if len(units) != 84 {
		return fmt.Errorf("postgres snapshot fixture must contain exactly 84 units, got %d", len(units))
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

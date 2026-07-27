package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateExtensionDependencyDeclarationsRejectsOmittedAndNullArrays(t *testing.T) {
	valid := validExtensionDependencyDeclarationSet()
	dependencies := valid["dependencies"].([]any)

	for _, key := range []string{"imported_schema_ids", "imported_algorithm_ids", "imported_artifacts"} {
		t.Run("omitted "+key, func(t *testing.T) {
			row := cloneJSONMap(t, dependencies[0].(map[string]any))
			delete(row, key)
			copySet := cloneJSONMap(t, valid)
			copyRows := append([]any(nil), dependencies...)
			copyRows[0] = row
			copySet["dependencies"] = copyRows
			requireErrorContains(t, validateExtensionDependencyDeclarations(copySet), key+" must be a present non-null array")
		})
		t.Run("null "+key, func(t *testing.T) {
			row := cloneJSONMap(t, dependencies[0].(map[string]any))
			row[key] = nil
			copySet := cloneJSONMap(t, valid)
			copyRows := append([]any(nil), dependencies...)
			copyRows[0] = row
			copySet["dependencies"] = copyRows
			requireErrorContains(t, validateExtensionDependencyDeclarations(copySet), key+" must be a present non-null array")
		})
	}
}

func TestValidateExtensionOwnerFragmentRejectsCapabilities(t *testing.T) {
	fragment := map[string]any{
		"schema_id":         "cartulary.extension_owner_fragment.v3",
		"owner_fragment_id": "test.fragment.v1",
		"owner_id":          "test",
		"facts": []any{map[string]any{
			"fact_kind":     "capability",
			"profile_id":    "test",
			"capability_id": "test.capability",
		}},
	}
	requireErrorContains(t, validateExtensionOwnerFragment(fragment, "fragment.json"), "capability facts are prohibited")
}

func TestValidateExtensionConfigurationRequiresInertSchemaReference(t *testing.T) {
	contract := map[string]any{
		"schema_id":                    "cartulary.extension_profile_configuration_contract.v3",
		"configuration_contract_id":    "test.configuration.v1",
		"profile_id":                   "test",
		"configuration_contract_major": json.Number("1"),
		"namespace_schema_id":          "test.configuration.v1",
		"keys": []any{map[string]any{
			"key":                      "test.value",
			"inactive_policy":          "syntax_only",
			"inactive_value_schema_id": nil,
		}},
	}
	requireErrorContains(t, validateExtensionConfigurationContract(contract, "configuration.json"), "syntax_only requires a non-null inactive_value_schema_id")
}

func TestExtensionCanonicalDigestIncludesRequiredFinalLF(t *testing.T) {
	value := map[string]any{"schema_id": "test.v1"}
	digest, err := extensionCanonicalDigest(value)
	if err != nil {
		t.Fatalf("digest extension artifact: %v", err)
	}
	if digest != "a03b53978fe4447e8fd7539e7778e4764686c694690529ce76b9fe2e3b51f9ae" {
		t.Fatalf("unexpected extension canonical digest %s", digest)
	}
}

func TestValidateExtensionBindingSourcesRejectsUnsortedContributions(t *testing.T) {
	bindings := make([]any, 0, len(requiredExtensionProfiles))
	for _, profileID := range requiredExtensionProfiles {
		bindings = append(bindings, map[string]any{
			"profile_id":                   profileID,
			"contract_major":               json.Number("1"),
			"implementation_id":            "cartulary." + profileID + ".v1",
			"state_ownership_kind":         "core_managed",
			"implemented_contribution_ids": []any{},
			"implemented_job_kinds":        []any{},
			"implemented_worker_kinds":     []any{},
			"algorithm_ids":                []any{},
			"participant_implementations":  []any{},
		})
	}
	source := map[string]any{
		"schema_id": "cartulary.extension_implementation_binding_source_set.v1",
		"bindings":  bindings,
	}
	bindings[0].(map[string]any)["implemented_contribution_ids"] = []any{"z", "a"}
	requireErrorContains(t, validateExtensionBindingSources(source), "must be sorted and unique")
}

func TestValidateExtensionParticipantContractEnforcesAggregateCeiling(t *testing.T) {
	contract := map[string]any{
		"schema_id":                "cartulary.extension_participant_contract.v1",
		"participant_id":           "test.participant",
		"profile_id":               "test",
		"contribution_kind":        "incident_portability_participant",
		"context_schema_id":        "test.context.v1",
		"result_schema_id":         "test.result.v1",
		"maximum_input_bytes":      json.Number("67108865"),
		"maximum_result_bytes":     json.Number("67108864"),
		"preparation_side_effects": "forbidden",
		"mutation_protocol":        "shared_transaction_only",
		"staged_output_capability": "operation_scoped",
		"algorithm_ids":            []any{"test.prepare_v1"},
	}
	requireErrorContains(t, validateExtensionParticipantContract(contract, "participant.json"), "maximum_input_bytes must be in 1..67108864")
}

func TestValidateExtensionParticipantSpecializationRejectsWrongSharedResult(t *testing.T) {
	contract := map[string]any{
		"schema_id":                "cartulary.extension_participant_specialization.v3",
		"profile_id":               "snapshot_reporting",
		"participant_id":           "snapshot_reporting.render_export_v1",
		"participant_kind":         "snapshot_reporting",
		"shared_context_schema_id": "cartulary.extension_snapshot_reporting_participant_context.v1",
		"operations": []any{map[string]any{
			"operation_kind":        "emit",
			"result_schema_id":      "cartulary.obsolete_snapshot_reporting_result.v1",
			"algorithm_id":          "snapshot_reporting.render_export_v1",
			"output_schema_id":      "cartulary.reporting_export_model.v1",
			"ordering_algorithm_id": "materialize_reporting_export_model_v1",
			"state_family_ids":      []any{},
			"max_input_bytes":       json.Number("67108864"),
			"max_output_bytes":      json.Number("67108864"),
			"max_items":             json.Number("1048576"),
		}},
	}
	requireErrorContains(t, validateExtensionParticipantSpecialization(contract, "participant.json"), "result_schema_id does not match")
}

func TestDeriveExtensionArtifactsIsDeterministicAndPhaseFree(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	first, err := deriveExtensionArtifacts(root)
	if err != nil {
		t.Fatalf("derive extension artifacts: %v", err)
	}
	second, err := deriveExtensionArtifacts(root)
	if err != nil {
		t.Fatalf("derive extension artifacts again: %v", err)
	}
	if len(first) == 0 || len(first) != len(second) {
		t.Fatalf("unexpected derived artifact counts %d and %d", len(first), len(second))
	}
	seen := map[string]artifact{}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("derived artifact %d is not byte-stable", index)
		}
		if strings.Contains(first[index].Path, "phase2") || strings.Contains(first[index].JSON, "cartulary.extensions.phase2.v1") {
			t.Fatalf("derived artifact retains phase-shaped identity: %s", first[index].Path)
		}
		seen[first[index].Path] = first[index]
	}
	for _, requiredPath := range []string{
		generatedExtensionPrefix + "profile-registry.json",
		generatedExtensionPrefix + "registry-integrity.json",
		generatedExtensionPrefix + "validation-condition-registry.json",
		generatedExtensionPrefix + "implementation-bindings/network_flow_activity.json",
		generatedExtensionPrefix + "job-contracts/import/import.apply_v1.json",
		generatedExtensionPrefix + "job-contracts/import/import.discovery_v1.json",
		generatedExtensionPrefix + "job-contracts/incident_portability/incident_portability.export_v1.json",
		generatedExtensionPrefix + "job-contracts/incident_portability/incident_portability.import_v1.json",
		generatedExtensionPrefix + "job-contracts/reference_pack/reference_pack.import_v1.json",
		generatedExtensionPrefix + "job-contracts/reference_pack/reference_pack.refresh_v1.json",
		generatedExtensionPrefix + "job-contracts/reference_pack/reference_pack.reverify_v1.json",
		generatedExtensionPrefix + "job-contracts/snapshot_reporting/snapshot_reporting.composition_preview_v1.json",
		generatedExtensionPrefix + "job-contracts/snapshot_reporting/snapshot_reporting.release_create_v1.json",
		generatedExtensionPrefix + "job-contracts/snapshot_reporting/snapshot_reporting.snapshot_create_v1.json",
	} {
		if _, exists := seen[requiredPath]; !exists {
			t.Fatalf("missing generated artifact %s", requiredPath)
		}
	}
	registry := seen[generatedExtensionPrefix+"profile-registry.json"]
	if !strings.HasSuffix(registry.JSON, "\n") || strings.HasSuffix(registry.JSON, "\n\n") {
		t.Fatalf("registry does not use exactly one final LF")
	}

	var stateRegistry map[string]any
	if err := json.Unmarshal([]byte(seen[generatedExtensionPrefix+"state-registry.json"].JSON), &stateRegistry); err != nil {
		t.Fatalf("decode generated state registry: %v", err)
	}
	profiles, ok := stateRegistry["profiles"].([]any)
	if !ok || len(profiles) != 1 {
		t.Fatalf("generated state profiles = %#v", stateRegistry["profiles"])
	}
	networkFlow, ok := profiles[0].(map[string]any)
	if !ok ||
		networkFlow["profile_id"] != "network_flow_activity" ||
		networkFlow["initialization_kind"] != "empty" ||
		networkFlow["initialization_algorithm_id"] != nil ||
		networkFlow["initialization_algorithm_definition_sha256"] != nil ||
		networkFlow["initialization_definition_sha256"] == "" ||
		networkFlow["implementation_binding_sha256"] == "" {
		t.Fatalf("generated Network Flow state projection = %#v", profiles[0])
	}
}

func validExtensionDependencyDeclarationSet() map[string]any {
	rows := make([]any, 0, len(requiredExtensionDependencies))
	for _, dependencyID := range requiredExtensionDependencies {
		rows = append(rows, map[string]any{
			"dependency_id":          dependencyID,
			"imported_schema_ids":    []any{},
			"imported_algorithm_ids": []any{},
			"imported_artifacts":     []any{},
		})
	}
	return map[string]any{
		"schema_id":    "cartulary.extension_dependency_declaration_set.v3",
		"dependencies": rows,
	}
}

func cloneJSONMap(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal clone: %v", err)
	}
	var cloned map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(encoded)))
	decoder.UseNumber()
	if err := decoder.Decode(&cloned); err != nil {
		t.Fatalf("decode clone: %v", err)
	}
	return cloned
}

func TestValidateViewSchemaRejectsUnknownKeys(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["legacy_key"] = true

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "unknown key legacy_key")
}

func TestValidateViewSchemaRejectsMissingRequiredFields(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	delete(schema, "inline_create")

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "inline_create must be an object")
}

func TestValidateErrorRegistryRejectsDuplicateCodes(t *testing.T) {
	registry := validErrorRegistry()
	registry["errors"] = []any{
		validErrorEntry("cartulary.test.duplicate", "409"),
		validErrorEntry("cartulary.test.duplicate", "400"),
	}

	err := validateErrorRegistry(registry)
	requireErrorContains(t, err, "duplicate error code cartulary.test.duplicate")
}

func TestValidateErrorRegistryRejectsInvalidReasonRegistries(t *testing.T) {
	t.Run("unknown error code", func(t *testing.T) {
		registry := validErrorRegistry()
		registry["reason_registries"] = []any{
			validReasonRegistry("cartulary.test.missing", "reason"),
		}

		err := validateErrorRegistry(registry)
		requireErrorContains(t, err, "references unknown error code cartulary.test.missing")
	})

	t.Run("duplicate reason code", func(t *testing.T) {
		registry := validErrorRegistry()
		registry["reason_registries"] = []any{
			map[string]any{
				"error_code": "cartulary.test.conflict",
				"reason_codes": []any{
					validReasonCode("duplicate"),
					validReasonCode("duplicate"),
				},
			},
		}

		err := validateErrorRegistry(registry)
		requireErrorContains(t, err, "contains duplicate reason code duplicate")
	})
}

func TestValidateViewSchemaRejectsDuplicateFieldKeys(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["fields"] = []any{
		validViewField("name"),
		validViewField("name"),
	}

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "duplicate field_key name")
}

func TestValidateViewSchemaIndexRejectsStaleArtifactPaths(t *testing.T) {
	root := t.TempDir()
	writeJSONFile(t, filepath.Join(root, "contracts", "view-schemas", "index.json"), map[string]any{
		"$schema":     contractDraft202012Schema,
		"registry_id": "cartulary.view_schemas.v1",
		"note":        "test index",
		"view_schemas": []any{
			map[string]any{
				"view_schema_id":               "cartulary.view.test.v1",
				"title":                        "Test View",
				"surface_kind":                 "built_in_sheet",
				"surface_status":               "required_built_in_sheet",
				"source_record_types":          []any{"note"},
				"required_reference_pack_keys": []any{},
				"artifact_path":                "contracts/view-schemas/missing.json",
			},
		},
	})
	writeJSONFile(
		t,
		filepath.Join(root, "contracts", "view-schemas", "cartulary.view.test.v1.json"),
		validViewSchema("cartulary.view.test.v1"),
	)

	err := validateContractFamily(root, "view-schemas")
	requireErrorContains(t, err, "references missing artifact contracts/view-schemas/missing.json")
}

func TestLoadFamiliesUsesActiveRegistryEntries(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"contracts/openapi", "contracts/network-flow"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("create contract dir: %v", err)
		}
	}
	writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
		"$schema":     contractDraft202012Schema,
		"schema_id":   contractFamilyRegistrySchemaID,
		"registry_id": contractFamilyRegistryID,
		"note":        "test registry",
		"families": []any{
			validContractFamilyEntry("openapi", "contracts/openapi", "active", "OpenAPIArtifacts", "openAPIArtifacts", 0, nil),
			validContractFamilyEntry("network-flow", "contracts/network-flow", "planned", "NetworkFlowArtifacts", "networkFlowArtifacts", 1, []any{"NFA-GEN-002"}),
		},
	})

	families, err := loadFamilies(root)
	if err != nil {
		t.Fatalf("load families: %v", err)
	}
	if len(families) != 1 {
		t.Fatalf("expected one active family, got %d", len(families))
	}
	if families[0].Dir != "openapi" || families[0].GoName != "OpenAPIArtifacts" || families[0].TSName != "openAPIArtifacts" {
		t.Fatalf("unexpected active family: %#v", families[0])
	}
}

func TestLoadFamiliesRejectsUnsafeOrAmbiguousRegistry(t *testing.T) {
	t.Run("planned family requires activation dependency", func(t *testing.T) {
		root := t.TempDir()
		for _, dir := range []string{"contracts/openapi", "contracts/network-flow"} {
			if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
				t.Fatalf("create contract dir: %v", err)
			}
		}
		writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
			"$schema":     contractDraft202012Schema,
			"schema_id":   contractFamilyRegistrySchemaID,
			"registry_id": contractFamilyRegistryID,
			"note":        "test registry",
			"families": []any{
				validContractFamilyEntry("openapi", "contracts/openapi", "active", "OpenAPIArtifacts", "openAPIArtifacts", 0, nil),
				validContractFamilyEntry("network-flow", "contracts/network-flow", "planned", "NetworkFlowArtifacts", "networkFlowArtifacts", 1, nil),
			},
		})

		_, err := loadFamilies(root)
		requireErrorContains(t, err, "activation_dependency_ids must not be empty for planned families")
	})

	t.Run("duplicate generated name is rejected", func(t *testing.T) {
		root := t.TempDir()
		for _, dir := range []string{"contracts/openapi", "contracts/ws"} {
			if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
				t.Fatalf("create contract dir: %v", err)
			}
		}
		writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
			"$schema":     contractDraft202012Schema,
			"schema_id":   contractFamilyRegistrySchemaID,
			"registry_id": contractFamilyRegistryID,
			"note":        "test registry",
			"families": []any{
				validContractFamilyEntry("openapi", "contracts/openapi", "active", "OpenAPIArtifacts", "openAPIArtifacts", 0, nil),
				validContractFamilyEntry("ws", "contracts/ws", "active", "OpenAPIArtifacts", "wsArtifacts", 1, nil),
			},
		})

		_, err := loadFamilies(root)
		requireErrorContains(t, err, "go_name duplicates family openapi")
	})

	t.Run("escaped contract root is rejected", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "contracts", "openapi"), 0o755); err != nil {
			t.Fatalf("create contract dir: %v", err)
		}
		writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
			"$schema":     contractDraft202012Schema,
			"schema_id":   contractFamilyRegistrySchemaID,
			"registry_id": contractFamilyRegistryID,
			"note":        "test registry",
			"families": []any{
				validContractFamilyEntry("openapi", "contracts/../openapi", "active", "OpenAPIArtifacts", "openAPIArtifacts", 0, nil),
			},
		})

		_, err := loadFamilies(root)
		requireErrorContains(t, err, "contract_root escapes contracts/")
	})
}

func TestValidateViewSchemaRejectsInvalidFieldReferences(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["sort_fields"] = []any{"missing_field"}

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "sort_fields references unknown field missing_field")
}

func TestValidateViewSchemaRejectsMismatchedEnumMetadata(t *testing.T) {
	t.Run("enum values require enum read kind", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		field := validViewField("name")
		field["read_kind"] = "text"
		field["enum_values"] = []any{"open", "closed"}
		schema["fields"] = []any{field}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "enum_values requires read_kind=enum")
	})

	t.Run("enum read kind requires values", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		field := validViewField("name")
		field["read_kind"] = "enum"
		field["enum_values"] = nil
		schema["fields"] = []any{field}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "read_kind=enum requires enum_values")
	})
}

func TestValidateViewSchemaAcceptsCanonicalArtifactSourceFilter(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["base_projection"] = "artifact_grid_projection"
	schema["canonical_source_filter"] = map[string]any{
		"kind":  "artifact_type",
		"field": "artifact_type",
		"value": "test_artifact",
	}

	if err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json"); err != nil {
		t.Fatalf("expected canonical artifact source filter to validate: %v", err)
	}
}

func TestValidateViewSchemaRejectsInvalidCanonicalSourceFilter(t *testing.T) {
	tests := []struct {
		name       string
		filter     any
		wantDetail string
	}{
		{
			name:       "not object",
			filter:     "artifact_type=test",
			wantDetail: "canonical_source_filter must be an object",
		},
		{
			name: "unknown key",
			filter: map[string]any{
				"kind":   "artifact_type",
				"field":  "artifact_type",
				"value":  "test_artifact",
				"legacy": true,
			},
			wantDetail: "unknown key legacy",
		},
		{
			name: "unsupported kind",
			filter: map[string]any{
				"kind":  "record_type",
				"field": "artifact_type",
				"value": "test_artifact",
			},
			wantDetail: "kind must be one of artifact_type",
		},
		{
			name: "unsupported field",
			filter: map[string]any{
				"kind":  "artifact_type",
				"field": "kind",
				"value": "test_artifact",
			},
			wantDetail: "field must be one of artifact_type",
		},
		{
			name: "missing value",
			filter: map[string]any{
				"kind":  "artifact_type",
				"field": "artifact_type",
			},
			wantDetail: "value is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			schema := validViewSchema("cartulary.view.test.v1")
			schema["canonical_source_filter"] = tc.filter

			err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
			requireErrorContains(t, err, tc.wantDetail)
		})
	}
}

func TestValidateViewSchemaRejectsInvalidInspectorConfig(t *testing.T) {
	t.Run("missing config", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		delete(schema, "inspector_config")

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "inspector_config must be an object")
	})

	t.Run("mismatched view schema", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		config["view_schema_id"] = "cartulary.view.other.v1"

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "view_schema_id must match containing view_schema_id")
	})

	t.Run("unknown panel", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		config["panels"] = []any{map[string]any{"panel_id": "legacy", "label": "Legacy"}}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "panel_id must be one of")
	})

	t.Run("duplicate panel", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		config["panels"] = []any{
			map[string]any{"panel_id": "details", "label": "Details"},
			map[string]any{"panel_id": "details", "label": "Details Again"},
		}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "duplicate panel_id details")
	})

	t.Run("duplicate feature key", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		group := validInspectorFeatureGroup("details.read")
		config["feature_groups"] = []any{group, validInspectorFeatureGroup("details.read")}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "duplicate feature_group_key details.read")
	})

	t.Run("feature group bound", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		groups := make([]any, 0, 65)
		for index := 0; index < 65; index++ {
			groups = append(groups, validInspectorFeatureGroup("details.read_"+string(rune('a'+index%26))+"_"+string(rune('a'+index/26))))
		}
		config["feature_groups"] = groups

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "feature_groups must contain at most 64 entries")
	})

	t.Run("unknown route kind", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["route_binding"] = map[string]any{"kind": "legacy_route", "owner": "current_row_projection"}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "kind must be one of")
	})

	t.Run("unknown route owner", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["route_binding"] = map[string]any{"kind": "panel_read", "owner": "legacy_owner"}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "owner must be one of")
	})

	t.Run("unknown disabled condition", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["disabled_when"] = []any{"stale_legacy_state"}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "references unknown condition stale_legacy_state")
	})

	t.Run("unknown result behaviors", func(t *testing.T) {
		for field, want := range map[string]string{
			"success_result_behavior": "success_result_behavior must be one of",
			"failure_result_behavior": "failure_result_behavior must be one of",
		} {
			schema := validViewSchema("cartulary.view.test.v1")
			group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
			group[field] = "legacy_result"

			err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
			requireErrorContains(t, err, want)
		}
	})

	t.Run("invalid feature key grammar", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["feature_group_key"] = "Details Read"

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "feature_group_key must be ASCII lower snake or dotted key")
	})

	t.Run("current profile registry completeness", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.timeline.v2")
		config := schema["inspector_config"].(map[string]any)
		config["view_schema_id"] = "cartulary.view.timeline.v2"

		err := validateViewSchemaShape(schema, "cartulary.view.timeline.v2.json")
		requireErrorContains(t, err, "feature_groups must contain exactly 27 declared feature groups for cartulary.view.timeline.v2")

		groups := make([]any, 0, len(inspectorFeatureRegistry["cartulary.view.timeline.v2"]))
		for _, key := range inspectorFeatureRegistry["cartulary.view.timeline.v2"] {
			groups = append(groups, validInspectorFeatureGroup(key))
		}
		groups[0].(map[string]any)["feature_group_key"] = "details.future"
		config["feature_groups"] = groups

		err = validateViewSchemaShape(schema, "cartulary.view.timeline.v2.json")
		requireErrorContains(t, err, "missing required feature_group_key details.read for cartulary.view.timeline.v2")
	})
}

func validViewSchema(id string) map[string]any {
	return map[string]any{
		"$schema":                      contractDraft202012Schema,
		"view_schema_id":               id,
		"title":                        "Test View",
		"surface_kind":                 "built_in_sheet",
		"source_record_types":          []any{"note"},
		"technical_fields":             []any{"record_id"},
		"required_reference_pack_keys": []any{},
		"default_visible_fields":       []any{"name"},
		"default_hidden_fields":        []any{"record_id"},
		"default_sort": []any{
			map[string]any{
				"field_key": "record_id",
				"direction": "asc",
			},
		},
		"sort_fields":                 []any{"name"},
		"sort_null_order":             "last",
		"filter_fields":               []any{"name"},
		"synthetic_filter_predicates": []any{},
		"grouping_fields":             []any{},
		"create_capable":              true,
		"inline_create":               map[string]any{"permits_zero_field_create": false},
		"inspector_config":            validInspectorConfig(id),
		"fields":                      []any{validViewField("name")},
	}
}

func validInspectorConfig(id string) map[string]any {
	return map[string]any{
		"inspector_config_schema_id":   "cartulary.inspector_config.v1",
		"view_schema_id":               id,
		"default_open":                 false,
		"subject_binding":              map[string]any{"kind": "selected_record"},
		"no_row_state":                 "no_row_selected",
		"unsupported_feature_behavior": "omit_feature",
		"panels": []any{
			map[string]any{"panel_id": "details", "label": "Details"},
		},
		"feature_groups": []any{validInspectorFeatureGroup("details.read")},
	}
}

func validInspectorFeatureGroup(key string) map[string]any {
	return map[string]any{
		"feature_group_key":       key,
		"panel_id":                "details",
		"label":                   "Details panel",
		"minimum_incident_role":   nil,
		"mutates":                 false,
		"requires_confirmation":   false,
		"route_binding":           map[string]any{"kind": "panel_read", "owner": "current_row_projection"},
		"seed_bindings":           []any{},
		"disabled_when":           []any{"no_row_selected"},
		"success_result_behavior": "preserve_selected_row",
		"failure_result_behavior": "show_same_shell_error_preserve_selection",
	}
}

func validViewField(key string) map[string]any {
	return map[string]any{
		"field_key":                    key,
		"label":                        "Name",
		"default_hidden":               false,
		"sortable":                     true,
		"header_sort_field_key":        nil,
		"filter_ops":                   []any{"eq"},
		"groupable":                    true,
		"read_kind":                    "scalar",
		"write_kind":                   "direct_value",
		"grid_editable":                true,
		"conflict_resolution_class":    nil,
		"entity_binding_mode":          nil,
		"string_contract_id":           nil,
		"direct_scalar_contract_id":    nil,
		"direct_reference_contract_id": nil,
		"clearable":                    true,
		"writable":                     true,
		"read_model":                   "record",
		"write_target":                 key,
	}
}

func validErrorRegistry() map[string]any {
	return map[string]any{
		"$schema":     contractDraft202012Schema,
		"registry_id": "cartulary.errors.v1",
		"note":        "test registry",
		"errors": []any{
			validErrorEntry("cartulary.test.conflict", "409"),
		},
	}
}

func validReasonRegistry(errorCode string, reasonCodes ...string) map[string]any {
	entries := make([]any, 0, len(reasonCodes))
	for _, reasonCode := range reasonCodes {
		entries = append(entries, validReasonCode(reasonCode))
	}
	return map[string]any{
		"error_code":   errorCode,
		"reason_codes": entries,
	}
}

func validReasonCode(code string) map[string]any {
	return map[string]any{
		"code":    code,
		"summary": "test reason",
	}
}

func validContractFamilyEntry(familyID, root, status, goName, tsName string, order int, activationDependencyIDs []any) map[string]any {
	if activationDependencyIDs == nil {
		activationDependencyIDs = []any{}
	}
	return map[string]any{
		"family_id":                            familyID,
		"contract_root":                        root,
		"generation_status":                    status,
		"go_name":                              goName,
		"ts_name":                              tsName,
		"output_order":                         order,
		"generated_outputs":                    []any{"internal/gen/contracts/contracts_gen.go"},
		"typescript_runtime_artifact_prefixes": []any{root + "/"},
		"activation_dependency_ids":            activationDependencyIDs,
		"description":                          "test family",
	}
}

func validErrorEntry(code string, status string) map[string]any {
	return map[string]any{
		"code":        code,
		"http_status": json.Number(status),
		"summary":     "test error",
	}
}

func writeJSONFile(t *testing.T, file string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(file, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func requireErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q", substring)
	}
	if !strings.Contains(err.Error(), substring) {
		t.Fatalf("expected error containing %q, got %q", substring, err.Error())
	}
}

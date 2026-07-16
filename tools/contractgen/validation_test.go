package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
				"view_schema_id":      "cartulary.view.test.v1",
				"title":               "Test View",
				"surface_kind":        "built_in_sheet",
				"source_record_types": []any{"note"},
				"artifact_path":       "contracts/view-schemas/missing.json",
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
		"family_id":                 familyID,
		"contract_root":             root,
		"generation_status":         status,
		"go_name":                   goName,
		"ts_name":                   tsName,
		"output_order":              order,
		"owner_document":            "docs/spec/01_architecture_storage_and_view_contracts.md",
		"owner_sections":            []any{"3"},
		"generated_outputs":         []any{"internal/gen/contracts/contracts_gen.go"},
		"activation_dependency_ids": activationDependencyIDs,
		"description":               "test family",
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

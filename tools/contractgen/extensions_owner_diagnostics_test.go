package main

import (
	"errors"
	"strings"
	"testing"
)

func checkOwnerInputGenerationDiagnostics(t *testing.T, root string) {
	for _, conditionID := range []string{"extension_registry_collision", "extension_dependency_diagnostic_overflow", "extension_registry_limit"} {
		for _, mode := range []string{"wrong formatter", "wrong class"} {
			t.Run(conditionID+"/"+mode, func(t *testing.T) {
				fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
					if path != "contracts/extensions/validation/surfaces.json" {
						return
					}
					for _, declaration := range object["declarations"].([]any) {
						for _, surface := range declaration.(map[string]any)["procedural_surfaces"].([]any) {
							for _, raw := range surface.(map[string]any)["conditions"].([]any) {
								condition := raw.(map[string]any)
								if condition["condition_id"] != conditionID {
									continue
								}
								if mode == "wrong formatter" {
									condition["expected_formatter_id"] = "diagnostic_redacted_v1"
								} else {
									condition["condition_class"] = "invalid_scalar"
								}
							}
						}
					}
				})
				if _, err := collectArtifacts(fixture, "extensions"); err == nil {
					t.Fatal("producer admitted invalid diagnostic definition on a valid catalog")
				}
			})
		}

	}

	for _, mode := range []string{"missing", "unknown"} {
		t.Run(mode, func(t *testing.T) {
			fixture := extensionInputFixture(t, root, func(_ string, object map[string]any) {
				rows, ok := object["facts"].([]any)
				if !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
					return
				}
				if mode == "missing" {
					kept := []any{}
					for _, raw := range rows {
						fact := raw.(map[string]any)
						if fact["fact_kind"] == "egress" && (fact["profile_id"] == "import" || fact["profile_id"] == "reference_pack") {
							continue
						}
						kept = append(kept, fact)
					}
					object["facts"] = kept
				} else if object["owner_fragment_id"] == "core01.extension_profile_classifications_v1" {
					for _, raw := range rows {
						fact := raw.(map[string]any)
						if fact["fact_kind"] != "egress" {
							continue
						}
						for _, profile := range []string{"private_profile_alpha", "private_profile_beta", "private_profile_alpha"} {
							copy := cloneObject(fact)
							copy["profile_id"] = profile
							rows = append(rows, copy)
						}
						break
					}
					object["facts"] = rows
				}
			})
			_, err := collectArtifacts(fixture, "extensions")
			var failure *extensionRegistryError
			if !errors.As(err, &failure) || len(failure.findings) != 2 {
				t.Fatalf("generator did not collect owner diagnostics: %v", err)
			}
			seen := map[string]bool{}
			for _, finding := range failure.findings {
				details := finding["details"].(map[string]any)
				if finding["reason_code"] != "extension_registry_invalid" || finding["message"] != "The extension contract artifact set is invalid." {
					t.Fatal("incorrect owner diagnostic classification")
				}
				if mode == "missing" {
					if finding["path"] != "$.owner_fragments" || details["artifact_kind"] != "owner_input_registry" || details["expected"] != "exact:egress" || details["actual"] != "missing" {
						t.Fatalf("incorrect omission: %#v", finding)
					}
					seen[stringValue(details["safe_ref"])] = true
				} else if !strings.HasPrefix(stringValue(finding["path"]), "$.owner_fragments[") || !strings.HasSuffix(stringValue(finding["path"]), ".profile_id") || details["safe_ref"] != "cartulary:owner_fragment:core01.extension_profile_classifications_v1" || details["expected"] != "exact:recognized_profile_id" || details["actual"] != "redacted" {
					t.Fatalf("unsafe unknown diagnostic: %#v", finding)
				}
			}
			if mode == "missing" && (!seen["cartulary:profile:import"] || !seen["cartulary:profile:reference_pack"]) {
				t.Fatal("omission references did not identify both profiles")
			}
			if strings.Contains(err.Error(), "private_profile_") || strings.Contains(err.Error(), "sha256") {
				t.Fatal("private owner value escaped in build diagnostics")
			}
		})
	}
}

func checkGeneratorCollisionEntry(t *testing.T, root string) {
	for _, mode := range []string{"owner allocation", "namespace identity"} {
		t.Run(mode, func(t *testing.T) {
			fixture := extensionInputFixture(t, root, func(_ string, object map[string]any) {
				if mode == "namespace identity" && object["schema_id"] == "cartulary.extension_profile_configuration_contract.v4" {
					object["configuration_namespace"] = "private_marker"
				}
				rows, ok := object["facts"].([]any)
				if mode != "owner allocation" || !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" || object["owner_id"] != "core04" {
					return
				}
				fact := map[string]any{"fact_kind": "contribution", "profile_id": "import", "contribution": map[string]any{"kind": "extension_resource_kind", "contribution_id": "import.private_marker", "resource_kind": "private_marker"}}
				object["facts"] = append(rows, fact, cloneObject(fact))
			})
			_, err := collectArtifacts(fixture, "extensions")
			if err == nil {
				t.Fatal("invalid declaration admitted")
			}
			if strings.Contains(err.Error(), "private_marker") {
				t.Fatal("producer leaked an unadmitted declaration value")
			}
		})
	}

	fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
		if path != "contracts/extensions/build/implementation-bindings.json" {
			return
		}
		rows := object["bindings"].([]any)
		for i, raw := range rows {
			row := raw.(map[string]any)
			if row["profile_id"] != "import" {
				continue
			}
			duplicate := cloneObject(row)
			duplicate["implementation_id"] = "cartulary.import.alternate.v1"
			next := append([]any{}, rows[:i+1]...)
			next = append(next, duplicate)
			next = append(next, rows[i+1:]...)
			object["bindings"] = next
			break
		}
	})
	_, err := collectArtifacts(fixture, "extensions")
	var failure *extensionRegistryError
	if !errors.As(err, &failure) || len(failure.findings) != 1 {
		t.Fatalf("actual producer bypasses registered collision admission: %v", err)
	}
	details := failure.findings[0]["details"].(map[string]any)
	if details["collision_class"] != "implementation_binding_duplicate" || details["duplicate_count"] != 2 {
		t.Fatalf("incorrect binding collision: %#v", failure.findings)
	}
}

package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func checkGeneratorCollisionMatrix(t *testing.T, root string) {
	inputs, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ class, kind, contribution string }{
		{"profile_id", "recognized_profile", ""},
		{"claim_key", "claim_configuration", ""},
		{"public_schema_id", "public_schema", ""},
		{"worker_kind", "worker_kind", ""},
		{"job_kind", "job_kind", ""},
		{"workspace_identity", "workspace", ""},
		{"owner_fact_identity", "egress", ""},
		{"contribution_id", "contribution", "http_route_family"},
		{"extension_resource_kind", "contribution", "extension_resource_kind"},
		{"import_target", "contribution", "import_target"},
		{"transaction_participant", "contribution", "cross_owner_transaction_participant"},
	} {
		t.Run(tc.class, func(t *testing.T) {
			changed := false
			fixture := extensionInputFixture(t, root, func(_ string, object map[string]any) {
				if changed || object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
					return
				}
				if tc.kind == "public_schema" && object["owner_fragment_id"] == "core01.extension_discovery_v1" {
					fact := map[string]any{"fact_kind": "public_schema", "profile_id": "import", "public_schema_id": "cartulary.import.collision_fixture.v1"}
					object["facts"] = append(object["facts"].([]any), fact, cloneObject(fact), cloneObject(fact))
					changed = true
					return
				}
				for _, raw := range object["facts"].([]any) {
					fact := raw.(map[string]any)
					if fact["fact_kind"] != tc.kind {
						continue
					}
					if tc.contribution != "" && fact["contribution"].(map[string]any)["kind"] != tc.contribution {
						continue
					}
					object["facts"] = append(object["facts"].([]any), cloneObject(fact), cloneObject(fact))
					changed = true
					break
				}
			})
			if !changed {
				t.Fatal("missing declaration fixture")
			}
			findings := producerCollisionFindings(t, fixture)
			requireCollisionGroup(t, findings, tc.class, 3, nil)
			if tc.class != "owner_fact_identity" {
				for _, finding := range findings {
					if finding["details"].(map[string]any)["collision_class"] == "owner_fact_identity" {
						t.Fatal("narrow collision duplicated as fallback identity")
					}
				}
			}
		})
	}
	for _, tc := range []struct{ class, member, token string }{
		{"job_resource_ref_kind", "resource_ref_kind", "shared_resource"},
		{"deployment_admin_panel", "panel_key", "import.shared_panel"},
		{"authentication_entry", "entry_key", "import.shared_entry"},
		{"extension_resource_kind", "resource_kind", "shared_kind"},
	} {
		t.Run("grouped "+tc.class, func(t *testing.T) {
			fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
				if path != "contracts/extensions/fragments/core01.profile-classifications.json" {
					return
				}
				for i, profile := range []string{"import", "reference_pack", "import"} {
					contribution := map[string]any{"kind": tc.class, "contribution_id": fmt.Sprintf("%s.collision_%d", profile, i), tc.member: tc.token}
					object["facts"] = append(object["facts"].([]any), map[string]any{"fact_kind": "contribution", "profile_id": profile, "contribution": contribution})
				}
			})
			requireCollisionGroup(t, producerCollisionFindings(t, fixture), tc.class, 3, []string{"import", "reference_pack"})
		})
	}
	for _, mode := range []string{"owner_fragment_id", "configuration_namespace", "migration_lineage_id", "migration_id", "dependency_self", "dependency_duplicate", "dependency_cycle", "route_family_overlap", "base_route_capture"} {
		t.Run(mode, func(t *testing.T) {
			fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
				switch mode {
				case "owner_fragment_id":
					if path == "contracts/extensions/fragments/core01.profile-jobs.json" {
						object["owner_fragment_id"] = "core01.extension_profile_classifications_v1"
					}
				case "configuration_namespace":
					if path == "contracts/extensions/profiles/reference_pack/configuration.json" {
						for key := range object {
							delete(object, key)
						}
						for key, value := range cloneObject(inputs["profiles/import/configuration.json"]) {
							object[key] = value
						}
					}
				case "migration_lineage_id":
					if path == "contracts/extensions/fragments/network_flow_activity.profile.json" {
						for _, raw := range object["facts"].([]any) {
							fact := raw.(map[string]any)
							if fact["fact_kind"] == "state_ownership" {
								copy := cloneObject(fact)
								state := copy["state_ownership"].(map[string]any)
								version, err := jsonInt(state["current_state_version"], "current_state_version")
								if err != nil {
									t.Fatal(err)
								}
								state["current_state_version"] = version + 1
								object["facts"] = append(object["facts"].([]any), copy)
								break
							}
						}
					}
				case "migration_id":
					if path == "contracts/extensions/fragments/network_flow_activity.profile.json" {
						definition, ledger := migrationCollisionFacets(t)
						object["facts"] = append(object["facts"].([]any), definition, ledger, cloneObject(definition), cloneObject(ledger))
					}
				case "dependency_self", "dependency_duplicate", "dependency_cycle":
					if path != "contracts/extensions/fragments/core00.recognition.json" {
						return
					}
					edges := [][2]string{{"import", "import"}}
					if mode == "dependency_duplicate" {
						edges = [][2]string{{"import", "reference_pack"}, {"import", "reference_pack"}, {"import", "reference_pack"}}
					}
					if mode == "dependency_cycle" {
						edges = [][2]string{{"import", "reference_pack"}, {"reference_pack", "import"}, {"enterprise_authentication", "import"}}
					}
					for _, edge := range edges {
						object["facts"] = append(object["facts"].([]any), map[string]any{"fact_kind": "runtime_dependency", "profile_id": edge[0], "dependency_profile_id": edge[1], "required_contract_major": 1})
					}
				case "route_family_overlap":
					if path != "contracts/extensions/fragments/core01.discovery.json" {
						return
					}
					for _, route := range []string{"/api/v1/collision", "/api/v1/collision/{id}", "/api/v1/collision/{other}"} {
						object["facts"] = append(object["facts"].([]any), map[string]any{"fact_kind": "route_family", "profile_id": "import", "route_family": route})
					}
				case "base_route_capture":
					if path == "contracts/extensions/build/base-route-reservations.json" {
						object["reservations"] = append(object["reservations"].([]any), map[string]any{"reservation_id": "base.collision", "path_template": "/api/v1/import-sessions/{id}", "match_scope": "exact"})
					}
				}
			})
			findings := producerCollisionFindings(t, fixture)
			var count any = 2
			if mode == "dependency_duplicate" {
				count = 3
			}
			if mode == "dependency_self" || mode == "dependency_cycle" || mode == "route_family_overlap" || mode == "base_route_capture" {
				count = nil
			}
			var profiles []string
			if mode == "dependency_cycle" {
				profiles = []string{"import", "reference_pack"}
			}
			requireCollisionGroup(t, findings, mode, count, profiles)
			if mode == "migration_id" {
				requireCollisionGroup(t, findings, "migration_edge", 2, []string{"network_flow_activity"})
			}
			if mode == "route_family_overlap" && len(findings) != 3 {
				t.Fatalf("missing unordered route pairs: %d", len(findings))
			}
			if mode == "dependency_cycle" {
				details := findings[0]["details"].(map[string]any)
				if !reflect.DeepEqual(details["conflicting_tokens"], []string{"import->reference_pack", "reference_pack->import"}) {
					t.Fatalf("cycle includes a blocked nonmember: %#v", details)
				}
			}
		})
	}
	t.Run("matching migration facets describe one declaration", func(t *testing.T) {
		fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
			if path == "contracts/extensions/fragments/network_flow_activity.profile.json" {
				definition, ledger := migrationCollisionFacets(t)
				object["facts"] = append(object["facts"].([]any), definition, ledger)
			}
		})
		indexed, err := readExtensionInputCatalog(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateExtensionRegistryFindings(indexed); err != nil {
			t.Fatalf("paired facets falsely collide: %v", err)
		}
	})
	t.Run("collision pair overflow", func(t *testing.T) {
		fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
			if path != "contracts/extensions/fragments/core01.discovery.json" {
				return
			}
			for i := 0; i < 100; i++ {
				object["facts"] = append(object["facts"].([]any), map[string]any{"fact_kind": "route_family", "profile_id": "import", "route_family": fmt.Sprintf("/api/v1/collision/{id_%03d}", i)})
			}
		})
		findings := producerCollisionFindings(t, fixture)
		if len(findings) != 1 || findings[0]["reason_code"] != "extension_diagnostic_overflow" {
			t.Fatalf("collision overflow was truncated or misclassified: %#v", findings)
		}
		details := findings[0]["details"].(map[string]any)
		if details["phase"] != "registry_generation" || details["limit"] != 4096 || details["actual"] != 4097 {
			t.Fatalf("incorrect first-excess overflow: %#v", details)
		}
	})
}

func producerCollisionFindings(t *testing.T, root string) []map[string]any {
	t.Helper()
	_, err := collectArtifacts(root, "extensions")
	var failure *extensionRegistryError
	if !errors.As(err, &failure) {
		t.Fatalf("actual producer did not collect collisions: %v", err)
	}
	return failure.findings
}

func requireCollisionGroup(t *testing.T, findings []map[string]any, class string, count any, profiles []string) {
	t.Helper()
	found := 0
	for _, finding := range findings {
		details := finding["details"].(map[string]any)
		if details["collision_class"] != class {
			continue
		}
		found++
		if finding["reason_code"] != "extension_registry_conflict" || !reflect.DeepEqual(details["duplicate_count"], count) || (profiles != nil && !reflect.DeepEqual(details["conflicting_profile_ids"], profiles)) {
			t.Fatalf("incorrect grouped %s finding: %#v", class, finding)
		}
	}
	if found == 0 {
		t.Fatalf("missing %s collision: %#v", class, findings)
	}
}

func migrationCollisionFacets(t *testing.T) (map[string]any, map[string]any) {
	t.Helper()
	definition := map[string]any{"schema_id": "cartulary.extension_migration_definition.v1", "profile_id": "network_flow_activity", "migration_lineage_id": "network_flow_activity.state_v1", "migration_id": "network_flow_activity.advance_v1", "from_state_version": 5, "to_state_version": 6, "apply_algorithm_id": "network_flow_activity.advance_v1", "validation_algorithm_id": "network_flow_activity.validate_state_v6"}
	digest, err := extensionCanonicalDigest(definition)
	if err != nil {
		t.Fatal(err)
	}
	ledger := cloneObject(definition)
	delete(ledger, "apply_algorithm_id")
	delete(ledger, "validation_algorithm_id")
	ledger["schema_id"] = "cartulary.extension_migration_ledger_definition.v1"
	ledger["migration_definition_sha256"] = digest
	return map[string]any{"fact_kind": "migration_definition", "profile_id": "network_flow_activity", "migration_definition": definition}, map[string]any{"fact_kind": "migration_ledger_definition", "profile_id": "network_flow_activity", "migration_ledger_definition": ledger}
}

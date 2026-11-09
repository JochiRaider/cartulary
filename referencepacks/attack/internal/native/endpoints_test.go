package native

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func nativeTestObject(kind string, ordinal int) map[string]any {
	id := fmt.Sprintf("%s--00000000-0000-4000-8000-%012d", kind, ordinal)
	source := map[string]any{"id": id, "type": kind, "name": "Native"}
	if kind == "x-mitre-tactic" {
		source["x_mitre_shortname"] = "execution"
	}
	if kind == "x-mitre-collection" {
		source["x_mitre_version"] = "20.1"
		source["x_mitre_contents"] = []any{}
	}
	return map[string]any{"schema_id": "cartulary.reference_pack.framework.attack.object.v2", "object_id": id, "object_type": kind, "display_label": "Native", "description": nil, "aliases": []any{}, "deprecated": false, "replacement_object_id": nil, "external_refs": []any{}, "source": source, "source_domains": []any{"enterprise-attack"}, "source_refs": []any{map[string]any{"artifact_index": 0, "locator": "/objects/0"}}, "extensions": map[string]any{}}
}

func TestNativeATTACKTypedEmbeddedEndpoints_Unit(t *testing.T) {
	// Expectations are authored independently of embeddedRules.
	for _, tc := range []struct {
		source, target, field, member, predicate string
		membership                               bool
	}{
		{"x-mitre-matrix", "x-mitre-tactic", "tactic_refs", "", "has-tactic", false},
		{"attack-pattern", "x-mitre-tactic", "kill_chain_phases", "", "member-of-tactic", true},
		{"x-mitre-detection-strategy", "x-mitre-analytic", "x_mitre_analytic_refs", "", "has-analytic", false},
		{"x-mitre-analytic", "x-mitre-data-component", "x_mitre_log_source_references", "x_mitre_data_component_ref", "uses-data-component", false},
	} {
		t.Run(tc.predicate, func(t *testing.T) {
			for _, fault := range []string{"valid", "missing", "wrong target", "wrong source", "identity", "ambiguous tactic"} {
				if fault == "ambiguous tactic" && !tc.membership {
					continue
				}
				if fault == "wrong target" && tc.membership {
					continue
				}
				t.Run(fault, func(t *testing.T) {
					fromKind, toKind := tc.source, tc.target
					if fault == "wrong source" {
						fromKind = "tool"
					}
					if fault == "wrong target" {
						toKind = "tool"
					}
					from, to := nativeTestObject(fromKind, 1), nativeTestObject(toKind, 2)
					fromID, toID := from["object_id"].(string), to["object_id"].(string)
					var occurrence any = toID
					suffix := "/" + tc.field + "/0"
					if tc.member != "" {
						occurrence = map[string]any{tc.member: toID}
						suffix += "/" + tc.member
					}
					if tc.membership {
						occurrence = map[string]any{"kill_chain_name": "mitre-attack", "phase_name": "execution"}
					}
					from["source"].(map[string]any)[tc.field] = []any{occurrence, occurrence}
					from["source_refs"] = []any{map[string]any{"artifact_index": 0, "locator": "/objects/0"}, map[string]any{"artifact_index": 1, "locator": "/objects/7"}}
					to["source"].(map[string]any)["revoked"] = true
					objects := map[string]map[string]any{fromID: from, toID: to}
					if fault == "missing" {
						delete(objects, toID)
					}
					if fault == "identity" {
						to["source"].(map[string]any)["id"] = fromID
					}
					if fault == "ambiguous tactic" {
						other := nativeTestObject("x-mitre-tactic", 3)
						objects[other["object_id"].(string)] = other
					}
					edges, err := ATTACKEmbedded(objects)
					if fault != "valid" {
						var failure *Failure
						if !errors.As(err, &failure) || failure.Code != "content_semantic_invalid" {
							t.Fatalf("fault accepted: %v", err)
						}
						return
					}
					if err != nil || len(edges) != 2 {
						t.Fatalf("occurrences: %v %d", err, len(edges))
					}
					for i, edge := range edges {
						if edge["relationship_type"] != tc.predicate || edge["source_object_id"] != fromID || edge["target_object_id"] != toID || edge["source"] != nil || edge["description"] != nil {
							t.Fatal(edge)
						}
						refs := edge["source_refs"].([]map[string]any)
						want := strings.Replace(suffix, "/0", "/"+fmt.Sprint(i), 1)
						if len(refs) != 2 || refs[0]["locator"] != "/objects/0"+want || refs[1]["locator"] != "/objects/7"+want {
							t.Fatal(refs)
						}
					}
					if edges[0]["relationship_id"] == edges[1]["relationship_id"] {
						t.Fatal("occurrences coalesced")
					}
				})
			}
		})
	}
}

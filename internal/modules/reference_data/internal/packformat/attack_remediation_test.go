package packformat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
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

func TestNativeATTACKLimitSummary_Unit(t *testing.T) {
	if summary, err := CheckSummary(t.Context(), "content_schema", nil, func(context.Context, FindingSink) error {
		return errors.Join(errShapeLimit, io.ErrClosedPipe)
	}); summary != nil || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("resource abort concealed operational failure", summary, err)
	}
	for _, tc := range []struct {
		kind, field, limitID string
		maximum              int
		item                 any
	}{
		{"tool", "x_mitre_contributors", "max_profile_array", 64, "Contributor"},
		{"x-mitre-data-component", "x_mitre_log_sources", "max_attack_component_log_sources", 512, map[string]any{"name": "Log", "channel": "Channel"}},
		{"x-mitre-collection", "x_mitre_contents", "max_attack_collection_entries", 32768, map[string]any{"object_ref": "tool--00000000-0000-4000-8000-000000000001", "object_modified": "2026-11-05T00:00:00.000Z"}},
	} {
		t.Run(tc.field, func(t *testing.T) {
			row := nativeTestObject(tc.kind, 1)
			for _, count := range []int{tc.maximum, tc.maximum + 1} {
				items := make([]any, count)
				for i := range items {
					items[i] = tc.item
				}
				row["source"].(map[string]any)[tc.field] = items
				b, err := canonicaljson.Marshal(row)
				if err != nil {
					t.Fatal(err)
				}
				fixture := contentFixture{Key: "framework.attack", Members: map[string]string{"payload/objects.ndjson": string(b) + "\n", "payload/relationships.ndjson": ""}}
				summary, err := contentSchemaSummary(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), nil)
				if err != nil {
					t.Fatal(err)
				}
				if count == tc.maximum {
					if summary.Result != "succeeded" {
						t.Fatal(summary)
					}
					continue
				}
				assertNativeLimitIssue(t, summary, "$.objects[0].source."+tc.field, tc.limitID, fmt.Sprint(tc.maximum), fmt.Sprint(count))
				repeated, err := contentSchemaSummary(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), nil)
				if err != nil || !reflect.DeepEqual(summary, repeated) {
					t.Fatal("unstable diagnostic", err)
				}
				// Descendants are intentionally invalid. A proven resource excess must not
				// enumerate them or their hostile property names.
				for i := range items {
					items[i] = map[string]any{"SECRET_UNKNOWN_MEMBER": true}
				}
				b, _ = canonicaljson.Marshal(row)
				fixture.Members["payload/objects.ndjson"] = string(b) + "\n"
				stopped, err := contentSchemaSummary(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), nil)
				if err != nil || stopped.Total != 1 {
					t.Fatal("resource traversal continued", err, stopped)
				}
			}
		})
	}
}

func assertNativeLimitIssue(t *testing.T, summary *ValidationSummary, path, id, maximum, actual string) {
	t.Helper()
	if summary.Result != "failed" || summary.Total != 1 || len(summary.Issues) != 1 {
		t.Fatalf("summary: %#v", summary)
	}
	issue := summary.Issues[0]
	d := issue.Details
	if issue.Code != "content_schema_invalid" || issue.ReasonCode == nil || *issue.ReasonCode != "content_schema_invalid" || issue.Path != path || d.LimitID == nil || *d.LimitID != id || d.ExpectedToken == nil || *d.ExpectedToken != maximum || d.ActualToken == nil || *d.ActualToken != actual {
		b, _ := json.Marshal(issue)
		t.Fatal(string(b))
	}
}

func TestSchemaUnionDispatchAndLimitBindings_Unit(t *testing.T) {
	row := nativeTestObject("tool", 1)
	sourceShape := contentShapes["framework.attack/object"].properties["source"]
	for _, value := range []any{map[string]any{}, map[string]any{"type": "SECRET_TAG"}, map[string]any{"type": []any{"tool"}}} {
		findings := []Finding{}
		err := sourceShape.walk(value, "$.source", func(f Finding) error { findings = append(findings, f); return nil })
		if err != nil || len(findings) != 1 || findings[0].Path != "$.source" {
			t.Fatal("union selection", err, findings)
		}
	}
	relationship := contentShapes["framework.attack/relationship"].properties["source"]
	if !relationship.matches(nil) || relationship.matches(row["source"]) {
		t.Fatal("nullable typed source")
	}
	overlapping := compileShape(map[string]any{"oneOf": []any{map[string]any{"type": "string"}, map[string]any{"type": "string", "minLength": float64(2)}}}, map[string]bool{})
	if !overlapping.matches("x") || overlapping.matches("xx") {
		t.Fatal("oneOf exact cardinality")
	}
	numeric := compileShape(map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "string"}}}, map[string]bool{})
	if !numeric.matches(json.Number("2")) || numeric.matches(json.Number("2.5")) {
		t.Fatal("numeric kind dispatch")
	}
	numericOverlap := compileShape(map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "integer", "minimum": float64(2)}}}, map[string]bool{})
	if numericOverlap.matches(json.Number("2")) || !numericOverlap.matches(json.Number("1")) {
		t.Fatal("overlapping numeric union")
	}
	for _, schema := range []map[string]any{{"type": "array", "maxItems": float64(64), "x-limitId": "unknown"}, {"type": "array", "maxItems": float64(65), "x-limitId": "max_profile_array"}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid annotation accepted")
				}
			}()
			compileShape(schema, map[string]bool{})
		}()
	}
}

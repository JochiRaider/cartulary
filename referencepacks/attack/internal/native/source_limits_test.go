package native

import (
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"reflect"
	"testing"
)

func TestNativeATTACKSchemaLimits(t *testing.T) {
	for _, tc := range []struct {
		kind, field, id string
		maximum         int
		item            any
	}{
		{"tool", "x_mitre_contributors", "max_profile_array", 64, "Contributor"},
		{"x-mitre-data-component", "x_mitre_log_sources", "max_attack_component_log_sources", 512, map[string]any{"name": "Log", "channel": "Channel"}},
		{"x-mitre-collection", "x_mitre_contents", "max_attack_collection_entries", 32768, map[string]any{"object_ref": "tool--00000000-0000-4000-8000-000000000001", "object_modified": "2026-11-05T00:00:00.000Z"}},
	} {
		t.Run(tc.field, func(t *testing.T) {
			for _, count := range []int{tc.maximum, tc.maximum + 1} {
				row := nativeTestObject(tc.kind, 1)
				row["source_refs"].([]any)[0].(map[string]any)["artifact_index"] = json.Number("0")
				items := make([]any, count)
				for i := range items {
					items[i] = tc.item
				}
				row["source"].(map[string]any)[tc.field] = items
				rule := contentShapes[definition.Key+"/object"]
				collect := func() ([]Finding, error) {
					found := []Finding{}
					err := rule.Walk(row, "$.objects[0]", func(f Finding) error { found = append(found, f); return nil })
					return found, err
				}
				findings, err := collect()
				if count == tc.maximum {
					if err != nil || len(findings) != 0 {
						t.Fatal("equality", err, findings)
					}
					continue
				}
				if !schema.IsLimit(err) || len(findings) != 1 {
					t.Fatal("one over", err, findings)
				}
				f := findings[0]
				d := f.Details
				if f.Path != "$.objects[0].source."+tc.field || d.LimitID == nil || *d.LimitID != tc.id || d.ExpectedToken == nil || *d.ExpectedToken != fmt.Sprint(tc.maximum) || d.ActualToken == nil || *d.ActualToken != fmt.Sprint(count) {
					t.Fatal("guard identity", f)
				}
				again, err := collect()
				if !schema.IsLimit(err) || !reflect.DeepEqual(findings, again) {
					t.Fatal("unstable guard", err, again)
				}
				for i := range items {
					items[i] = map[string]any{"SECRET_UNKNOWN_MEMBER": true}
				}
				stopped, err := collect()
				if !schema.IsLimit(err) || len(stopped) != 1 || !reflect.DeepEqual(findings, stopped) {
					t.Fatal("descended beyond resource bound", err, stopped)
				}
			}
		})
	}
}

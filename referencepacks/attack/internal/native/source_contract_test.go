package native

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

// Current contract witnesses are independent of the historical proposal inputs.
func TestATTACKSourceContracts(t *testing.T) {
	const analytic = "x-mitre-analytic--11111111-1111-4111-8111-111111111111"
	const component = "x-mitre-data-component--22222222-2222-4222-8222-222222222222"
	source := map[string]any{
		"type": "x-mitre-analytic", "id": analytic, "name": "Native witness",
		"revoked": true, "x_mitre_deprecated": false,
		"external_references": []any{
			map[string]any{"source_name": "mitre-attack", "external_id": "AN1789", "url": "https://attack.mitre.org/example#AN1789"},
			map[string]any{"source_name": "citation", "url": "http://example.invalid/original#citation"},
		},
		"x_mitre_log_source_references": []any{map[string]any{"x_mitre_data_component_ref": component, "name": "source", "channel": "channel"}},
		"x_mitre_mutable_elements":      []any{map[string]any{"field": "Threshold", "description": "exact text"}},
	}
	clone := func(value any) map[string]any {
		t.Helper()
		b, err := canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		decoder := json.NewDecoder(bytes.NewReader(b))
		decoder.UseNumber()
		if err := decoder.Decode(&m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	native := compileTestReference("profiles/framework.attack.native_source.v2.schema.json#/$defs/x-mitre-analytic")
	for _, tc := range []struct {
		name   string
		change func(map[string]any)
		valid  bool
	}{
		{"native_optional_ids_original_urls_and_independent_flags", func(map[string]any) {}, true},
		{"closed_citation", func(m map[string]any) { m["external_references"].([]any)[0].(map[string]any)["unexpected"] = true }, false},
		{"lifecycle_boolean", func(m map[string]any) { m["revoked"] = "false" }, false},
		{"typed_component_reference", func(m map[string]any) {
			m["x_mitre_log_source_references"].([]any)[0].(map[string]any)["x_mitre_data_component_ref"] = 42
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := clone(source)
			tc.change(m)
			before, _ := canonicaljson.Marshal(m)
			if native.Matches(m) != tc.valid {
				t.Fatal("native admission changed")
			}
			after, _ := canonicaljson.Marshal(m)
			if !bytes.Equal(before, after) {
				t.Fatal("native values or optional presence mutated")
			}
		})
	}
	object := clone(map[string]any{
		"schema_id": "cartulary.reference_pack.framework.attack.object.v2", "object_id": analytic, "object_type": "x-mitre-analytic",
		"display_label": "Native witness", "description": nil, "aliases": []any{}, "deprecated": true, "replacement_object_id": nil,
		"external_refs": ATTACKExternalRefs(source), "source_refs": []any{map[string]any{"artifact_index": 0, "locator": "/objects/1"}},
		"extensions": map[string]any{}, "source_domains": []any{"enterprise-attack"}, "source": source,
	})
	for _, tc := range []struct {
		name   string
		change func(map[string]any)
		valid  bool
	}{
		{"object_native_and_canonical_projection", func(map[string]any) {}, true},
		{"native_source_required", func(m map[string]any) { delete(m, "source") }, false},
		{"extensions_closed", func(m map[string]any) { m["extensions"] = map[string]any{"opaque": true} }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := clone(object)
			tc.change(m)
			if contentShapes["framework.attack/object"].Matches(m) != tc.valid {
				t.Fatal("object admission changed")
			}
			if tc.valid {
				checks := &semanticChecks{}
				newAttackState().object(m, checks)
				if err := checks.finish("content_semantic_invalid"); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	const sroID = "relationship--44444444-4444-4444-8444-444444444444"
	relationship := clone(map[string]any{
		"schema_id": "cartulary.reference_pack.framework.attack.relationship.v2", "relationship_id": "rpr_1111111111111111111111111111111111111111111111111111111111111111",
		"relationship_type": "uses", "source_object_id": analytic, "target_object_id": component, "description": nil,
		"source_refs": []any{map[string]any{"artifact_index": 0, "locator": "/objects/1/x_mitre_log_source_references/0"}}, "extensions": map[string]any{}, "source_kind": "stix_sro",
		"source": map[string]any{"type": "relationship", "id": sroID, "relationship_type": "uses", "source_ref": analytic, "target_ref": component, "external_references": []any{map[string]any{"source_name": "citation", "url": "http://example.invalid/#citation"}}},
	})
	embedded := clone(relationship)
	embedded["source_kind"] = "embedded_reference"
	embedded["source"] = nil
	embedded["relationship_type"] = "uses-data-component"
	for _, tc := range []struct {
		name   string
		kind   string
		source any
		valid  bool
	}{
		{"sro_citations_and_upstream_identity", "stix_sro", relationship["source"], true},
		{"sro_original_required", "stix_sro", nil, false},
		{"embedded_has_no_fabricated_sro", "embedded_reference", nil, true},
		{"embedded_cannot_claim_sro", "embedded_reference", relationship["source"], false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := clone(relationship)
			m["source_kind"] = tc.kind
			m["source"] = tc.source
			if tc.kind == "embedded_reference" {
				m["relationship_type"] = "uses-data-component"
			}
			state := newAttackState()
			state.edges[m["relationship_id"].(string)] = embedded
			checks := &semanticChecks{}
			valid := contentShapes["framework.attack/relationship"].Matches(m)
			if valid {
				state.relationship(m, checks)
				valid = checks.finish("content_semantic_invalid") == nil
			}
			if valid != tc.valid {
				t.Fatal("relationship source-kind admission changed")
			}
			if tc.valid && tc.kind == "stix_sro" && !equalJSON(state.records[sroID], relationship["source"]) {
				t.Fatal("SRO identity or citations lost")
			}
		})
	}
}

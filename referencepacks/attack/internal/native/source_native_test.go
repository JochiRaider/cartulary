package native

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

func TestNativeATTACKProjectionAndLocalReferences_Unit(t *testing.T) {
	t.Run("native limits", testNativeATTACKLimits)
	source := map[string]any{"id": "tool--11111111-1111-4111-8111-111111111111", "type": "tool", "name": "  Cafe\u0301  ", "description": " \r\noriginal\rtext \r\n", "revoked": true, "x_mitre_deprecated": false, "external_references": []any{map[string]any{"source_name": "mitre-attack", "external_id": "S0001", "url": "https://example.invalid/x#fragment"}, map[string]any{"source_name": "citation", "url": "ftp://example.invalid/original"}}}
	before, _ := canonicaljson.Marshal(source)
	if ATTACKLabel(source) != "Café" || ATTACKDescription(source) != "original\ntext" || !ATTACKDeprecated(source) {
		t.Fatal("display or lifecycle projection")
	}
	refs := ATTACKExternalRefs(source)
	if len(refs) != 1 || refs[0]["external_id"] != "S0001" || refs[0]["url"] != nil {
		t.Fatal("external citation promotion", refs)
	}
	after, _ := canonicaljson.Marshal(source)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("native values mutated")
	}
	if ATTACKDescription(map[string]any{}) != nil || ATTACKDescription(map[string]any{"description": " \n"}) != nil {
		t.Fatal("empty description")
	}
}

func testNativeATTACKLimits(t *testing.T) {
	const id = "tool--00000000-0000-4000-8000-000000000001"
	check := func(kind, field string, item any, limit int) {
		t.Helper()
		s := compileTestReference("profiles/framework.attack.native_source.v2.schema.json#/$defs/" + kind)
		m := map[string]any{"type": kind, "id": kind + "--00000000-0000-4000-8000-000000000001"}
		if kind == "x-mitre-collection" {
			m["x_mitre_version"] = "20.1"
		}
		values := make([]any, limit)
		for i := range values {
			values[i] = item
		}
		m[field] = values
		if !s.Matches(m) {
			t.Fatal("native equality rejected", kind, field)
		}
		m[field] = append(values, item)
		if s.Matches(m) {
			t.Fatal("native excess admitted", kind, field)
		}
	}
	check("x-mitre-collection", "x_mitre_contents", map[string]any{"object_ref": id, "object_modified": "2026-11-05T00:00:00.000Z"}, 32768)
	check("x-mitre-data-component", "x_mitre_log_sources", map[string]any{"name": "Log", "channel": "Channel"}, 512)
	check("tool", "x_mitre_contributors", "Contributor", 64)
	// A native inventory's timestamp spelling is retained, including fractional
	// zeroes. Tune only that source string to isolate the inclusive line ceiling.
	contents := make([]any, 1000)
	for i := range contents {
		contents[i] = map[string]any{"object_ref": id, "object_modified": "2026-11-05T00:00:00." + strings.Repeat("0", 4000) + "Z"}
	}
	source := map[string]any{"type": "x-mitre-collection", "id": "x-mitre-collection--00000000-0000-4000-8000-000000000001", "name": "Collection", "x_mitre_version": "20.1", "x_mitre_contents": contents}
	row := map[string]any{"schema_id": "cartulary.reference_pack.framework.attack.object.v2", "object_id": source["id"], "object_type": source["type"], "display_label": "Collection", "description": nil, "aliases": []any{}, "deprecated": false, "replacement_object_id": nil, "external_refs": []any{}, "source_refs": []any{map[string]any{"artifact_index": 0, "locator": "/objects/0"}}, "source_domains": []any{"enterprise-attack"}, "extensions": map[string]any{}, "source": source}
	encoded, err := canonicaljson.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	delta := 4194304 - len(encoded) - 1
	for _, entry := range contents {
		if delta <= 0 {
			break
		}
		amount := min(delta, 4000)
		entry.(map[string]any)["object_modified"] = "2026-11-05T00:00:00." + strings.Repeat("0", 4000+amount) + "Z"
		delta -= amount
	}
	if delta != 0 {
		t.Fatal("line recipe cannot reach bound", delta)
	}
	encoded, err = canonicaljson.Marshal(row)
	if err != nil || len(encoded)+1 != 4194304 {
		t.Fatal("line recipe", len(encoded), err)
	}
	f := contentFixture{Key: "framework.attack", Members: map[string]string{"payload/objects.ndjson": string(encoded) + "\n", "payload/relationships.ndjson": ""}}
	if err := ValidateStructure(context.Background(), fixtureManifest(f), contentMemory(f.Members)); err != nil {
		t.Fatal("native line equality", err)
	}
	entry := contents[0].(map[string]any)
	value := entry["object_modified"].(string)
	entry["object_modified"] = strings.TrimSuffix(value, "Z") + "0Z"
	encoded, _ = canonicaljson.Marshal(row)
	f.Members["payload/objects.ndjson"] = string(encoded) + "\n"
	if err := ValidateStructure(context.Background(), fixtureManifest(f), contentMemory(f.Members)); err == nil {
		t.Fatal("native line excess admitted")
	}
}

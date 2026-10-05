package reference_data

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestIncidentBundleReferenceCatalogValidation_Unit(t *testing.T) {
	valid, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-references.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	refs, err := DecodeIncidentBundleReferences(valid)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := EncodeIncidentBundleReferences(refs.Sets, refs.Versions)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateIncidentBundleReferences(encoded); err != nil {
		t.Fatal(err)
	}
	empty, err := EncodeIncidentBundleReferences(nil, nil)
	if err != nil || string(empty) != "{\"schema_id\":\"reference_pack_refs.v1\",\"sets\":[],\"versions\":[]}\n" {
		t.Fatalf("empty catalog = %s, %v", empty, err)
	}
	for name, payload := range map[string]string{
		"retired array": "[]", "omitted fields": "{}", "null collections": `{"schema_id":"reference_pack_refs.v1","sets":null,"versions":[]}`,
		"duplicate key":     `{"schema_id":"reference_pack_refs.v1","sets":[],"sets":[],"versions":[]}`,
		"malformed Unicode": `{"schema_id":"\ud800","sets":[],"versions":[]}`, "trailing JSON": string(empty) + `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			err := ValidateIncidentBundleReferences([]byte(payload))
			if invariant, ok := IncidentBundleReferenceInvariant(err); !ok || invariant != IncidentBundleReferenceExactShapeInvariant {
				t.Fatalf("error=%v", err)
			}
		})
	}
	var root map[string]any
	if err := json.Unmarshal(valid, &root); err != nil {
		t.Fatal(err)
	}
	assertInvalid := func(t *testing.T, value any) {
		t.Helper()
		data, err := canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateIncidentBundleReferences(data); err == nil {
			t.Fatal("invalid reference graph admitted")
		}
	}
	// Enumerate every nested object boundary from an independently authored
	// fixture. No decoder may repair omitted or null signed identity members.
	var visit func(any, string)
	visit = func(value any, path string) {
		switch v := value.(type) {
		case map[string]any:
			for key, original := range v {
				for _, kind := range []string{"omitted", "null", "wrong type"} {
					t.Run(path+"/"+key+"/"+kind, func(t *testing.T) {
						if kind == "omitted" {
							delete(v, key)
						} else if kind == "null" {
							v[key] = nil
						} else {
							v[key] = true
						}
						assertInvalid(t, root)
						v[key] = original
					})
				}
			}
			t.Run(path+"/unknown", func(t *testing.T) {
				v["hostile-private-payload"] = true
				assertInvalid(t, root)
				delete(v, "hostile-private-payload")
			})
			for key, child := range v {
				visit(child, path+"/"+key)
			}
		case []any:
			if len(v) > 0 {
				visit(v[0], path+"/0")
			}
		}
	}
	visit(root, "catalog")
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"wrong set digest", func(m map[string]any) {
			m["sets"].([]any)[0].(map[string]any)["pack_set_sha256"] = strings.Repeat("0", 64)
		}},
		{"wrong set ID", func(m map[string]any) {
			m["sets"].([]any)[0].(map[string]any)["pack_set_id"] = "rpset_" + strings.Repeat("0", 64)
		}},
		{"duplicate set", func(m map[string]any) { a := m["sets"].([]any); m["sets"] = append(a, a[0]) }},
		{"missing version", func(m map[string]any) { m["versions"] = m["versions"].([]any)[1:] }},
		{"unreferenced versions", func(m map[string]any) { m["sets"] = []any{} }},
		{"duplicate version", func(m map[string]any) { a := m["versions"].([]any); m["versions"] = append(a, a[0]) }},
		{"unsorted versions", func(m map[string]any) { a := m["versions"].([]any); a[0], a[1] = a[1], a[0] }},
		{"unsorted members", func(m map[string]any) {
			a := m["sets"].([]any)[0].(map[string]any)["members"].([]any)
			a[0], a[1] = a[1], a[0]
		}},
		{"content substitution", func(m map[string]any) {
			m["versions"].([]any)[0].(map[string]any)["payload_sha256"] = strings.Repeat("0", 64)
		}},
		{"incorrect method", func(m map[string]any) {
			m["versions"].([]any)[0].(map[string]any)["verification_method"] = "packaged_release_manifest_v1"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var m map[string]any
			if err := json.Unmarshal(valid, &m); err != nil {
				t.Fatal(err)
			}
			tc.mutate(m)
			assertInvalid(t, m)
		})
	}
}

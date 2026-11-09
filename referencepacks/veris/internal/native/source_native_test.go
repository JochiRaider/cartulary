package native

import (
	"context"
	"encoding/json"
	"errors"
	verisinputs "github.com/JochiRaider/cartulary/referencepacks/veris/inputs"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

func verisCorpus(t testing.TB) (Manifest, contentMemory) {
	t.Helper()
	b, err := os.ReadFile("testdata/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m.ProfileID = definition.ID
	members := contentMemory{}
	for _, kind := range []string{"objects", "relationships"} {
		b, err = os.ReadFile("testdata/" + kind + ".ndjson")
		if err != nil {
			t.Fatal(err)
		}
		members["payload/"+kind+".ndjson"] = string(b)
	}
	return m, members
}

func TestVERISNativeCorpus_Unit(t *testing.T) {
	m, members := verisCorpus(t)
	if err := ValidateContent(context.Background(), m, members, nil); err != nil {
		t.Fatal(err)
	}
	rows := map[string]map[string]any{}
	counts := map[string]int{}
	root := ""
	for _, line := range strings.Split(strings.TrimSuffix(members["payload/objects.ndjson"], "\n"), "\n") {
		v, err := canonicaljson.DecodeStrict([]byte(line))
		if err != nil {
			t.Fatal(err)
		}
		row := v.(map[string]any)
		rows[row["object_id"].(string)] = row
		counts[row["object_type"].(string)]++
		if row["object_type"] == "schema-root" {
			root = row["object_id"].(string)
		}
	}
	if counts["schema-root"] != 1 || counts["schema-node"] != 260 || counts["enum-value"] != 2168 || counts["semantic-rule"] != 62 {
		t.Fatal("incomplete real corpus", counts)
	}
	var rebuild func(string) map[string]any
	rebuild = func(id string) map[string]any {
		s := rows[id]["source"].(map[string]any)
		result := map[string]any{}
		for k, v := range s["keywords"].(map[string]any) {
			switch k {
			case "properties":
				p := map[string]any{}
				for name, child := range v.(map[string]any) {
					p[name] = rebuild(child.(string))
				}
				result[k] = p
			case "items":
				result[k] = rebuild(v.(string))
			default:
				result[k] = v
			}
		}
		return result
	}
	b, err := verisinputs.Files.ReadFile("baseline.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	v, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		t.Fatal(err)
	}
	baseline := v.(map[string]any)
	if !equalJSON(rebuild(root), baseline["schema"]) {
		t.Fatal("schema reconstruction lost source meaning")
	}
	// Equal raw values in unrelated fields are distinct subjects.
	namibia := VERISIdentity("enum-value", "/properties/victim/properties/country/items", "NA")
	notApplicable := VERISIdentity("enum-value", "/properties/timeline/properties/compromise/properties/unit", "NA")
	if namibia == notApplicable || rows[namibia] == nil || rows[notApplicable] == nil {
		t.Fatal("NA contexts were conflated")
	}
	if rows[namibia]["source"].(map[string]any)["definition"] == rows[notApplicable]["source"].(map[string]any)["definition"] {
		t.Fatal("contextual definitions collapsed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(ValidateContent(ctx, m, members, nil), context.Canceled) {
		t.Fatal("cancellation ignored")
	}
}

func TestVERISNativeRejections_Unit(t *testing.T) {
	tests := []struct {
		name, kind, code string
		mutate           func([]map[string]any)
	}{
		{"predecessor", "objects", "content_schema_invalid", func(r []map[string]any) { r[0]["schema_id"] = "cartulary.reference_pack.framework.veris.object.v1" }},
		{"extensions", "objects", "content_schema_invalid", func(r []map[string]any) { r[0]["extensions"] = map[string]any{"execute": true} }},
		{"unsafe-display", "objects", "content_semantic_invalid", func(r []map[string]any) { r[0]["display_label"] = "<script>alert(1)</script>" }},
		{"context-identity", "objects", "content_semantic_invalid", func(r []map[string]any) { r[0]["source"].(map[string]any)["value"] = "Unknown" }},
		{"duplicate", "objects", "content_semantic_invalid", func(r []map[string]any) { r[1] = r[0] }},
		{"dangling-child", "objects", "content_semantic_invalid", func(r []map[string]any) {
			for _, o := range r {
				s := o["source"].(map[string]any)
				if s["kind"] == "schema-root" {
					s["keywords"].(map[string]any)["properties"].(map[string]any)["actor"] = "missing"
					return
				}
			}
		}},
		{"cycle", "objects", "content_semantic_invalid", func(r []map[string]any) {
			for _, o := range r {
				s := o["source"].(map[string]any)
				if s["kind"] == "schema-root" {
					s["keywords"].(map[string]any)["properties"].(map[string]any)["actor"] = o["object_id"]
					return
				}
			}
		}},
		{"unknown-keyword", "objects", "content_schema_invalid", func(r []map[string]any) {
			for _, o := range r {
				s := o["source"].(map[string]any)
				if s["kind"] == "schema-root" {
					s["keywords"].(map[string]any)["anyOf"] = []any{}
					return
				}
			}
		}},
		{"missing-required-property", "objects", "content_semantic_invalid", func(r []map[string]any) {
			for _, o := range r {
				s := o["source"].(map[string]any)
				if s["kind"] == "schema-root" {
					s["keywords"].(map[string]any)["required"] = []any{"missing"}
					return
				}
			}
		}},
		{"rule-target", "objects", "content_semantic_invalid", func(r []map[string]any) {
			for _, o := range r {
				s := o["source"].(map[string]any)
				if s["kind"] == "semantic-rule" {
					s["applies_to"] = []any{"/properties/missing"}
					return
				}
			}
		}},
		{"edge-provenance", "relationships", "content_semantic_invalid", func(r []map[string]any) {
			r[0]["source_refs"].([]any)[0].(map[string]any)["locator"] = "#wrong"
			id, _ := RelationshipID(VERISProfile, r[0]["relationship_type"].(string), r[0]["source_object_id"].(string), r[0]["target_object_id"].(string), r[0]["source_refs"])
			r[0]["relationship_id"] = id
		}},
		{"edge-direction", "relationships", "content_semantic_invalid", func(r []map[string]any) {
			r[0]["source_object_id"], r[0]["target_object_id"] = r[0]["target_object_id"], r[0]["source_object_id"]
			id, _ := RelationshipID(VERISProfile, r[0]["relationship_type"].(string), r[0]["source_object_id"].(string), r[0]["target_object_id"].(string), r[0]["source_refs"])
			r[0]["relationship_id"] = id
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, members := verisCorpus(t)
			path := "payload/" + tc.kind + ".ndjson"
			rows := []map[string]any{}
			for _, line := range strings.Split(strings.TrimSuffix(members[path], "\n"), "\n") {
				v, err := canonicaljson.DecodeStrict([]byte(line))
				if err != nil {
					t.Fatal(err)
				}
				rows = append(rows, v.(map[string]any))
			}
			tc.mutate(rows)
			var out strings.Builder
			for _, r := range rows {
				b, err := canonicaljson.Marshal(r)
				if err != nil {
					t.Fatal(err)
				}
				out.Write(b)
				out.WriteByte('\n')
			}
			members[path] = out.String()
			err := ValidateContent(context.Background(), m, members, nil)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Code != tc.code {
				t.Fatalf("source model accepted or wrong rejection: %v", err)
			}
		})
	}
}

func TestVERISNativeBounds_Unit(t *testing.T) {
	f := contentFixtureFor(t, "framework.veris")
	v, err := canonicaljson.DecodeStrict([]byte(strings.Split(f.Members["payload/objects.ndjson"], "\n")[0]))
	if err != nil {
		t.Fatal(err)
	}
	row := v.(map[string]any)
	kw := row["source"].(map[string]any)["keywords"].(map[string]any)
	for _, n := range []int{8192, 8193} {
		kw["description"] = strings.Repeat("😀", n)
		if contentShapes["framework.veris/object"].Matches(row) != (n == 8192) {
			t.Fatal("native scalar boundary", n)
		}
	}
	delete(kw, "description")
	for _, n := range []int{512, 513} {
		values := make([]any, n)
		for i := range values {
			values[i] = strings.Repeat("x", i+1)
		}
		kw["enum"] = values
		if contentShapes["framework.veris/object"].Matches(row) != (n == 512) {
			t.Fatal("enum boundary", n)
		}
	}
	// A flat row stays shallow JSON while its graph can exceed the structural limit.
	for _, depth := range []int{32, 33} {
		objects := []map[string]any{}
		edges := []map[string]any{}
		p := ""
		for i := 0; i <= depth; i++ {
			kind := "schema-node"
			if i == 0 {
				kind = "schema-root"
			}
			id := VERISIdentity(kind, p)
			kw := map[string]any{"type": "string"}
			next := p + "/properties/x"
			child := VERISIdentity("schema-node", next)
			if i < depth {
				kw = map[string]any{"type": "object", "properties": map[string]any{"x": child}}
				refs := []any{map[string]any{"artifact_index": 0, "locator": "#" + next}}
				eid, _ := RelationshipID(VERISProfile, "contains", id, child, refs)
				edges = append(edges, map[string]any{"schema_id": "cartulary.reference_pack.framework.veris.relationship.v2", "relationship_type": "contains", "relationship_id": eid, "source_object_id": id, "target_object_id": child, "source_refs": refs, "description": nil, "extensions": map[string]any{}})
			}
			s := map[string]any{"kind": kind, "schema_version": "fixture-v2", "pointer": p, "keywords": kw}
			label, desc := VERISDisplay(s)
			objects = append(objects, map[string]any{"schema_id": "cartulary.reference_pack.framework.veris.object.v2", "object_id": id, "object_type": kind, "source": s, "display_label": label, "description": desc, "aliases": []string{label}, "deprecated": false, "replacement_object_id": nil, "external_refs": []any{}, "source_refs": []any{map[string]any{"artifact_index": 0, "locator": "#" + p}}, "extensions": map[string]any{}})
			p = next
		}
		slices.SortFunc(objects, func(a, b map[string]any) int {
			return strings.Compare(a["object_id"].(string), b["object_id"].(string))
		})
		slices.SortFunc(edges, func(a, b map[string]any) int {
			return strings.Compare(a["source_object_id"].(string), b["source_object_id"].(string))
		})
		f.Members = map[string]string{}
		for kind, rows := range map[string][]map[string]any{"objects": objects, "relationships": edges} {
			var b strings.Builder
			for _, r := range rows {
				raw, e := canonicaljson.Marshal(r)
				if e != nil {
					t.Fatal(e)
				}
				b.Write(raw)
				b.WriteByte('\n')
			}
			f.Members["payload/"+kind+".ndjson"] = b.String()
		}
		err := ValidateContent(context.Background(), fixtureManifest(f), contentMemory(f.Members), nil)
		if (err == nil) != (depth == 32) {
			t.Fatal("structural depth boundary", depth, err)
		}
	}
}

package packformat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

type contentFixture struct {
	Key     string            `json:"pack_key"`
	Members map[string]string `json:"members"`
}

func contentFixtureFor(t *testing.T, key string) contentFixture {
	t.Helper()
	raw, err := os.ReadFile("../../../../../contracts/reference-pack-fixtures/fixtures/profiles/" + strings.ReplaceAll(key, ".", "_") + "_valid.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture contentFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Key != key {
		t.Fatal("profile fixture identity mismatch")
	}
	return fixture
}

type contentMemory map[string]string

func (m contentMemory) Open(_ context.Context, path string) (io.ReadCloser, error) {
	v, ok := m[path]
	if !ok {
		return nil, errors.New("missing fixture")
	}
	return io.NopCloser(strings.NewReader(v)), nil
}

type contentRows []ContentRow

func (rows *contentRows) Append(_ context.Context, row ContentRow) error {
	*rows = append(*rows, row)
	return nil
}
func fixtureManifest(f contentFixture) Manifest {
	m := Manifest{Key: f.Key, ProfileID: profiles[f.Key].ID, ProfileVersion: profiles[f.Key].Version, Artifacts: []SourceArtifact{{Ref: "fixture"}}}
	m.Summary.Kind = profiles[f.Key].Shape
	if m.Summary.Kind == "dataset" {
		m.Summary.Kind = "objects_relationships"
	}
	for path, value := range f.Members {
		n := int64(strings.Count(value, "\n"))
		switch path {
		case "payload/entries.ndjson":
			m.Summary.Entries = &n
		case "payload/objects.ndjson":
			m.Summary.Objects = &n
		case "payload/relationships.ndjson":
			m.Summary.Relationships = &n
		}
	}
	return m
}

func testCanonicalProfileContent(t *testing.T, f contentFixture) {
	var rows contentRows
	if err := ValidateContent(context.Background(), fixtureManifest(f), contentMemory(f.Members), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("empty profile fixture")
	}
	for _, check := range []string{"content_semantics", "type_registry"} {
		summary, err := CheckSummary(context.Background(), check, nil, func(ctx context.Context, emit FindingSink) error {
			if check == "type_registry" {
				return RegistryCompatibilityFindings(ctx, fixtureManifest(f), contentMemory(f.Members), emit)
			}
			return ContentSemanticsFindings(ctx, fixtureManifest(f), contentMemory(f.Members), emit)
		})
		if err != nil || summary.Result != "succeeded" {
			t.Fatalf("%s diagnostics changed valid profile admission: %#v %v", check, summary, err)
		}
	}
	for _, row := range rows {
		if row.ID == "" || len(row.ID) > 512 || len(row.Canonical) == 0 {
			t.Fatal("unaddressable index row")
		}
	}
}

func TestContentSchemaFailurePrecedence_Unit(t *testing.T) {
	t.Run("later schema failure precedes earlier semantic failure", func(t *testing.T) {
		for _, f := range []contentFixture{contentFixtureFor(t, "type_registry.host")} {
			value, err := canonicaljson.DecodeStrict([]byte(strings.Split(f.Members["payload/entries.ndjson"], "\n")[0]))
			if err != nil {
				t.Fatal(err)
			}
			value.(map[string]any)["aliases"] = []any{"same", "same"}
			row, err := canonicaljson.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			for _, body := range []string{string(row) + "\n{}\n", "{}\n" + string(row) + "\n"} {
				f.Members["payload/entries.ndjson"] = body
				var rows contentRows
				err := ValidateContent(context.Background(), fixtureManifest(f), contentMemory(f.Members), &rows)
				var failure *Failure
				if !errors.As(err, &failure) || failure.Code != "content_schema_invalid" || len(rows) != 0 {
					t.Fatalf("traversal selected the wrong check or wrote an index: %v, %d", err, len(rows))
				}
			}
		}
	})
}

func testProfileSchemaClosure(t *testing.T, f contentFixture) {
	for path, body := range f.Members {
		kind := "entry"
		if strings.Contains(path, "objects") {
			kind = "object"
		}
		if strings.Contains(path, "relationships") {
			kind = "relationship"
		}
		raw := []byte(strings.Split(body, "\n")[0])
		value, err := canonicaljson.DecodeStrict(raw)
		if err != nil {
			t.Fatal(err)
		}
		shape := contentShapes[f.Key+"/"+kind]
		m := value.(map[string]any)
		for name, original := range m {
			delete(m, name)
			if shape.matches(m) {
				t.Fatalf("%s accepts omission of %s", f.Key, name)
			}
			m[name] = nil
			nullable := slices.Contains([]string{"description", "replacement_entry_id", "replacement_object_id", "stix_mapping", "first_observed_at", "last_observed_at", "notes", "event_version", "level", "task", "opcode", "publisher", "source"}, name)
			if !nullable && shape.matches(m) {
				t.Fatalf("%s accepts null %s", f.Key, name)
			}
			m[name] = original
		}
		m["extensions"] = map[string]any{"org.example.field": "hostile-secret"}
		if shape.matches(m) {
			t.Fatalf("%s/%s permits an undeclared extension", f.Key, kind)
		}
		m["extensions"] = map[string]any{}
		m["unrecognized_member"] = true
		if shape.matches(m) {
			t.Fatalf("%s permits unknown member", f.Key)
		}
	}
}

func TestContentRejectsMalformedStreamsAndSemanticAmbiguity_Unit(t *testing.T) {
	t.Run("complete semantic findings", testContentSemanticFindings)
	t.Run("complete registry findings", testRegistrySemanticFindings)
	f := contentFixtureFor(t, "type_registry.host")
	original := f.Members["payload/entries.ndjson"]
	for name, body := range map[string]string{"no final newline": strings.TrimSuffix(original, "\n"), "blank line": original + "\n", "duplicate identity": original + original, "BOM": "\ufeff" + original, "noncanonical": " " + original, "oversize line": strings.Repeat(" ", 1048576) + "\n"} {
		t.Run(name, func(t *testing.T) {
			m := fixtureManifest(f)
			if err := ValidateContent(context.Background(), m, contentMemory{"payload/entries.ndjson": body}, nil); err == nil {
				t.Fatal("malformed content admitted")
			}
		})
	}
	mutate := func(change func(map[string]any)) string {
		v, err := canonicaljson.DecodeStrict([]byte(strings.TrimSuffix(original, "\n")))
		if err != nil {
			t.Fatal(err)
		}
		change(v.(map[string]any))
		b, err := canonicaljson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(b) + "\n"
	}
	for name, body := range map[string]string{
		"duplicate normalized alias": mutate(func(m map[string]any) { m["aliases"] = []any{"same", "same"} }),
		"uppercase registry alias":   mutate(func(m map[string]any) { m["aliases"] = []any{"UPPER"} }),
		"invalid source reference": mutate(func(m map[string]any) {
			m["source_refs"] = []any{map[string]any{"artifact_index": 1, "locator": "fixture"}}
		}),
		"deprecated unknown":  mutate(func(m map[string]any) { m["deprecated"] = true }),
		"extension namespace": mutate(func(m map[string]any) { m["extensions"] = map[string]any{"bad": true} }),
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateContent(context.Background(), fixtureManifest(f), contentMemory{"payload/entries.ndjson": body}, nil); err == nil {
				t.Fatal("ambiguous content admitted")
			}
		})
	}
}

func TestWindowsSIDIdentityBoundaries_Unit(t *testing.T) {
	sid := "S-1-281474976710655" + strings.Repeat("-4294967295", 15)
	if len(sid) != 184 || !ValidSID(sid) {
		t.Fatal("SID maximum not admitted")
	}
	for _, bad := range []string{sid + "-1", "S-1-5", "S-1-05-1", "S-1-5-4294967296", "S-1-281474976710656-1"} {
		if ValidSID(bad) {
			t.Fatal("invalid SID admitted")
		}
	}
}

func TestCISAKEVIdentityBoundaries_Unit(t *testing.T) {
	cve := "CVE-2026-" + strings.Repeat("1", 183)
	if len(cve) != 192 {
		t.Fatal("fixture length")
	}
	if _, err := NormalizeLookup("enrichment.cisa_kev", "cve_id", strings.ToLower(cve)); err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeLookup("enrichment.cisa_kev", "cve_id", cve+"1"); err == nil {
		t.Fatal("overlong CVE admitted")
	}
}

func testCanonicalProfileBytes(t *testing.T, f contentFixture) {
	var rows contentRows
	if err := ValidateContent(context.Background(), fixtureManifest(f), contentMemory(f.Members), &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		found := false
		for _, body := range f.Members {
			found = found || bytes.Contains([]byte(body), append(bytes.Clone(row.Canonical), '\n'))
		}
		if !found {
			t.Fatal("validation repaired signed content")
		}
	}
}

func TestHostRegistryProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "type_registry.host"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "type_registry.host"))
	testNestedContentAdmission(t, contentFixtureFor(t, "type_registry.host"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "type_registry.host"))
}

func TestEvidenceRegistryProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "type_registry.evidence"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "type_registry.evidence"))
	testNestedContentAdmission(t, contentFixtureFor(t, "type_registry.evidence"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "type_registry.evidence"))
}

func TestIndicatorRegistryProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "type_registry.indicator"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "type_registry.indicator"))
	testNestedContentAdmission(t, contentFixtureFor(t, "type_registry.indicator"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "type_registry.indicator"))
}

func TestD3FENDProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "framework.d3fend"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "framework.d3fend"))
	testNestedContentAdmission(t, contentFixtureFor(t, "framework.d3fend"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "framework.d3fend"))
	testFrameworkExternalLookupKeys(t, contentFixtureFor(t, "framework.d3fend"))
}

func TestTORProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.tor"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.tor"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.tor"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.tor"))
}

func TestCISAKEVProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.cisa_kev"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.cisa_kev"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.cisa_kev"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.cisa_kev"))
}

func TestMSPortalsProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.ms_portals"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.ms_portals"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.ms_portals"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.ms_portals"))
}

func TestLOLBASProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.lolbas"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.lolbas"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.lolbas"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.lolbas"))
	testCompleteContentLineBoundary(t)
}

func TestLOLDriversProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.loldrivers"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.loldrivers"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.loldrivers"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.loldrivers"))
}

func TestLOLESXiProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.lolesxi"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.lolesxi"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.lolesxi"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.lolesxi"))
}

func TestHijackLibsProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.hijacklibs"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.hijacklibs"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.hijacklibs"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.hijacklibs"))
}

func TestWindowsSIDsProfileContent_Unit(t *testing.T) {
	testCanonicalProfileContent(t, contentFixtureFor(t, "enrichment.windows_sids"))
	testProfileSchemaClosure(t, contentFixtureFor(t, "enrichment.windows_sids"))
	testNestedContentAdmission(t, contentFixtureFor(t, "enrichment.windows_sids"))
	testCanonicalProfileBytes(t, contentFixtureFor(t, "enrichment.windows_sids"))
}

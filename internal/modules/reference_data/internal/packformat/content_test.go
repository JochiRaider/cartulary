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

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type contentFixture struct {
	Key     string            `json:"pack_key"`
	Members map[string]string `json:"members"`
}

func contentFixtures(t *testing.T) []contentFixture {
	t.Helper()
	raw, err := os.ReadFile("../../../../../contracts/reference-pack-fixtures/fixtures/profiles.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []contentFixture `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture.Cases
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
	m := Manifest{Key: f.Key, ProfileID: profiles[f.Key].ID, Artifacts: []SourceArtifact{{Ref: "fixture"}}}
	m.Summary.Kind = profiles[f.Key].Shape
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

func TestAllDeclaredProfilesHaveCanonicalContentAndIndexes_Unit(t *testing.T) {
	t.Run("external references derive unique keys", testFrameworkExternalLookupKeys)
	t.Run("complete NDJSON and profile array boundaries", testCompleteContentLineBoundary)
	fixtures := contentFixtures(t)
	if len(fixtures) != 16 {
		t.Fatal("incomplete profile fixture catalog")
	}
	for _, f := range fixtures {
		t.Run(f.Key, func(t *testing.T) {
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
		})
	}
}

func TestContentSchemasRejectOmissionNullAndUnknownMembers_Unit(t *testing.T) {
	t.Run("nested admission", testNestedContentAdmission)
	t.Run("later schema failure precedes earlier semantic failure", func(t *testing.T) {
		for _, f := range contentFixtures(t) {
			if f.Key != "type_registry.host" {
				continue
			}
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
	for _, f := range contentFixtures(t) {
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
				nullable := slices.Contains([]string{"description", "replacement_entry_id", "replacement_object_id", "stix_mapping", "first_observed_at", "last_observed_at", "notes", "event_version", "level", "task", "opcode", "publisher"}, name)
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
}

func TestContentRejectsMalformedStreamsAndSemanticAmbiguity_Unit(t *testing.T) {
	t.Run("complete semantic findings", testContentSemanticFindings)
	t.Run("complete registry findings", testRegistrySemanticFindings)
	f := contentFixtures(t)[0]
	for _, candidate := range contentFixtures(t) {
		if candidate.Key == "type_registry.host" {
			f = candidate
			break
		}
	}
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

func TestEventAndSIDIdentitiesAndLookupBoundaries_Unit(t *testing.T) {
	provider := strings.Repeat("p", 250) + ":colon"
	version := int64(255)
	id := EventIdentity(provider, 65535, &version)
	if len(id) != 266 {
		t.Fatal("incorrect event bound")
	}
	got, event, v, ok := ParseEventIdentity(id)
	if !ok || got != provider || event != 65535 || v == nil || *v != 255 {
		t.Fatal("event identity did not round-trip")
	}
	nullID := EventIdentity(provider, 65535, nil)
	_, _, v, ok = ParseEventIdentity(nullID)
	if !ok || v != nil || nullID == id {
		t.Fatal("null-version identity collision")
	}
	for _, bad := range []string{provider + ":065535:*", provider + ":65536:*", provider + ":1:256", provider + ":1:", strings.ToUpper(id)} {
		if _, _, _, ok := ParseEventIdentity(bad); ok {
			t.Fatal("invalid event identity")
		}
	}
	sid := "S-1-281474976710655" + strings.Repeat("-4294967295", 15)
	if len(sid) != 184 || !ValidSID(sid) {
		t.Fatal("SID maximum not admitted")
	}
	for _, bad := range []string{sid + "-1", "S-1-5", "S-1-05-1", "S-1-5-4294967296", "S-1-281474976710656-1"} {
		if ValidSID(bad) {
			t.Fatal("invalid SID admitted")
		}
	}
	for _, input := range []string{`{"provider_name":"Example:Provider","event_id":65535}`, `{"provider_name":"Example:Provider","event_id":65535,"event_version":255}`} {
		value, _ := canonicaljson.DecodeStrict([]byte(input))
		query, err := NormalizeLookup("enrichment.windows_event_ids", "provider_event_id", value)
		if err != nil || !strings.HasPrefix(query.Value, "example:provider\x0065535") {
			t.Fatalf("event lookup: %#v %v", query, err)
		}
	}
	value, _ := canonicaljson.DecodeStrict([]byte(`{"provider_name":"Example","event_id":1,"event_version":null}`))
	if _, err := NormalizeLookup("enrichment.windows_event_ids", "provider_event_id", value); err == nil {
		t.Fatal("explicit null version admitted")
	}
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

func TestContentCanonicalMemberBytesRemainUnmodified_Unit(t *testing.T) {
	for _, f := range contentFixtures(t) {
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
}

package packformat

import (
	"context"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testFrameworkExternalLookupKeys(t *testing.T) {
	for _, fixture := range contentFixtures(t) {
		if !strings.HasPrefix(fixture.Key, "framework.") {
			continue
		}
		t.Run(fixture.Key, func(t *testing.T) {
			lines := strings.Split(strings.TrimSuffix(fixture.Members["payload/objects.ndjson"], "\n"), "\n")
			for i, line := range lines {
				value, err := canonicaljson.DecodeStrict([]byte(line))
				if err != nil {
					t.Fatal(err)
				}
				value.(map[string]any)["external_refs"] = []any{
					map[string]any{"source_name": "a", "external_id": "shared", "url": nil},
					map[string]any{"source_name": "a", "external_id": "shared", "url": "https://example.com/"},
					map[string]any{"source_name": "b", "external_id": "shared", "url": nil},
				}
				encoded, err := canonicaljson.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				lines[i] = string(encoded)
			}
			fixture.Members["payload/objects.ndjson"] = strings.Join(lines, "\n") + "\n"
			var rows contentRows
			if err := ValidateContent(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), &rows); err != nil {
				t.Fatal(err)
			}
			objects := 0
			for _, row := range rows {
				if row.Shape != "object" {
					continue
				}
				if string(row.Canonical) != lines[objects] {
					t.Fatal("canonical references changed")
				}
				objects++
				count := 0
				for _, key := range row.Keys {
					if key.Kind == "external_id" && key.Value == "shared" {
						count++
					}
				}
				if count != 1 {
					t.Errorf("%s has %d external lookup keys, want 1", row.ID, count)
				}
			}
			if objects != len(lines) {
				t.Fatal("objects coalesced")
			}
			// Exact source tuples remain invalid, even though index keys coalesce.
			value, _ := canonicaljson.DecodeStrict([]byte(lines[0]))
			object := value.(map[string]any)
			refs := object["external_refs"].([]any)
			object["external_refs"] = append([]any{refs[0]}, refs...)
			encoded, _ := canonicaljson.Marshal(object)
			lines[0] = string(encoded)
			fixture.Members["payload/objects.ndjson"] = strings.Join(lines, "\n") + "\n"
			if err := ValidateContent(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), &contentRows{}); err == nil {
				t.Fatal("duplicate reference admitted")
			}
		})
	}
}

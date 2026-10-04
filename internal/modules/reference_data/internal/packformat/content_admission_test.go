package packformat

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// The expectations here are independent of the compiled schema: every member
// is required, every object is closed, and only these named fields admit null.
// Mutations go through the same full content boundary that builds live indexes.
func testNestedContentAdmission(t *testing.T) {
	for _, fixture := range contentFixtures(t) {
		for _, path := range slices.Sorted(maps.Keys(fixture.Members)) {
			lines := strings.Split(strings.TrimSuffix(fixture.Members[path], "\n"), "\n")
			for lineIndex, line := range lines {
				value, err := canonicaljson.DecodeStrict([]byte(line))
				if err != nil {
					t.Fatal(err)
				}
				assertRejected := func(location, mutation string) {
					t.Helper()
					t.Run(fmt.Sprintf("%s/%s/%d%s/%s", fixture.Key, path, lineIndex, location, mutation), func(t *testing.T) {
						encoded, err := canonicaljson.Marshal(value)
						if err != nil {
							t.Fatal(err)
						}
						members := maps.Clone(fixture.Members)
						changed := slices.Clone(lines)
						changed[lineIndex] = string(encoded)
						members[path] = strings.Join(changed, "\n") + "\n"
						var rows contentRows
						err = ValidateContent(context.Background(), fixtureManifest(fixture), contentMemory(members), &rows)
						var failure *Failure
						if !errors.As(err, &failure) || failure.Code != "content_schema_invalid" || len(rows) != 0 {
							t.Fatalf("malformed content admitted or indexed: %v, rows=%d", err, len(rows))
						}
					})
				}
				var visit func(any, string)
				visit = func(node any, location string) {
					switch node := node.(type) {
					case map[string]any:
						for _, key := range slices.Sorted(maps.Keys(node)) {
							original := node[key]
							at := location + "/" + key
							delete(node, key)
							assertRejected(at, "omitted")
							node[key] = wrongContentType(original)
							assertRejected(at, "wrong_type")
							if !nullableContentMember(key) {
								node[key] = nil
								assertRejected(at, "null")
							}
							node[key] = original
							visit(original, at)
						}
						node["unknown_private_field"] = "value must never enter a diagnostic"
						assertRejected(location, "unknown_member")
						delete(node, "unknown_private_field")
					case []any:
						for index, original := range node {
							at := fmt.Sprintf("%s/%d", location, index)
							node[index] = nil
							assertRejected(at, "null_element")
							node[index] = wrongContentType(original)
							assertRejected(at, "wrong_element_type")
							node[index] = original
							visit(original, at)
						}
					}
				}
				visit(value, "")
			}
		}
	}
}

func wrongContentType(value any) any {
	if _, boolean := value.(bool); boolean {
		return "false"
	}
	return false
}

func nullableContentMember(key string) bool {
	switch key {
	case "description", "replacement_entry_id", "replacement_object_id", "stix_mapping", "first_observed_at", "last_observed_at", "notes", "event_version", "level", "task", "opcode", "publisher":
		return true
	default:
		return false
	}
}

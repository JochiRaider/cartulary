package packformat

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// This recipe expands a valid profile, including all required nullable members,
// into a complete canonical row. Expected byte/count limits are independently
// specified; no production limit is raised to accommodate the fixture.
func testCompleteContentLineBoundary(t *testing.T) {
	var fixture contentFixture
	for _, f := range contentFixtures(t) {
		if f.Key == "enrichment.lolbas" {
			fixture = f
			break
		}
	}
	value, err := canonicaljson.DecodeStrict([]byte(strings.TrimSuffix(fixture.Members["payload/entries.ndjson"], "\n")))
	if err != nil {
		t.Fatal(err)
	}
	row := value.(map[string]any)
	examples, references := make([]string, 64), make([]string, 64)
	for i := range examples {
		prefix := fmt.Sprintf("example %02d ", i)
		examples[i] = prefix + strings.Repeat("a", 8192-len(prefix))
		prefix = fmt.Sprintf("https://example.invalid/%02d/", i)
		references[i] = prefix + strings.Repeat("a", 8192-len(prefix))
	}
	row["usage_examples"], row["references"] = examples, references
	raw, err := canonicaljson.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	const lineLimit = 1048576
	excess := len(raw) + 1 - lineLimit
	if excess < 1 || excess >= 8000 {
		t.Fatal("unexpected boundary recipe size", excess)
	}
	references[63] = references[63][:len(references[63])-excess]
	raw, err = canonicaljson.Marshal(row)
	if err != nil || len(raw)+1 != lineLimit {
		t.Fatal("recipe does not reach exact line limit", len(raw), err)
	}
	fixture.Members["payload/entries.ndjson"] = string(raw) + "\n"
	var indexed contentRows
	if err := ValidateContent(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), &indexed); err != nil || len(indexed) != 1 {
		t.Fatal("complete at-limit content", err)
	}
	references[63] += "a"
	raw, err = canonicaljson.Marshal(row)
	if err != nil || len(raw)+1 != lineLimit+1 {
		t.Fatal("one-over recipe", err)
	}
	fixture.Members["payload/entries.ndjson"] = string(raw) + "\n"
	if err := ValidateContent(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), nil); err == nil {
		t.Fatal("one-over NDJSON line admitted")
	}
	// Keep the row small to isolate each collection's cardinality rejection.
	for _, field := range []string{"usage_examples", "references"} {
		row["usage_examples"] = []string{"example"}
		row["references"] = []string{"https://example.invalid/"}
		values := make([]string, 65)
		for i := range values {
			values[i] = fmt.Sprintf("https://example.invalid/%02d", i)
		}
		row[field] = values
		raw, err = canonicaljson.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		fixture.Members["payload/entries.ndjson"] = string(raw) + "\n"
		if err := ValidateContent(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), nil); err == nil {
			t.Fatal("one-over profile array admitted", field)
		}
	}
}

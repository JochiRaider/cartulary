package native

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"strings"
	"testing"
)

func TestWindowsEventSourceIdentity(t *testing.T) {
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
}

func TestWindowsDocumentationPresentation(t *testing.T) {
	t.Run("rich text has no trailing paragraphs", func(t *testing.T) {
		f := contentFixtureFor(t, "enrichment.windows_event_ids")
		lines := strings.Split(strings.TrimSuffix(f.Members["payload/entries.ndjson"], "\n"), "\n")
		var row map[string]any
		if e := json.Unmarshal([]byte(lines[0]), &row); e != nil {
			t.Fatal(e)
		}
		for _, text := range []string{"  %1\n\n  %2", "  %1\n\n  %2\n\n"} {
			row["documentation"] = []any{map[string]any{"kind": "message_template", "scope": "event_version", "source_refs": row["source_refs"], "content": map[string]any{"text": text}}}
			b, e := canonicaljson.Marshal(row)
			if e != nil {
				t.Fatal(e)
			}
			lines[0] = string(b)
			f.Members["payload/entries.ndjson"] = strings.Join(lines, "\n") + "\n"
			e = ValidateContent(t.Context(), fixtureManifest(f), contentMemory(f.Members), nil)
			if (e == nil) != !strings.HasSuffix(text, "\n") {
				t.Fatal("trailing paragraph admission", e)
			}
		}
	})
}

package testsupport

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestEntraCorpusExpectations_Unit(t *testing.T) {
	f := EntraRelease(t, "../../../..")
	manifest, err := packformat.DecodeManifest(f.Members["manifest.json"], true)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"historical missing", "historical changed", "summary changed", "null description", "wrong index", "source revision", "source digest", "source URL", "missing provenance", "duplicate provenance", "publisher", "SPN", "deprecated", "replacement", "extensions"} {
		t.Run(mode, func(t *testing.T) {
			m := manifest
			m.Artifacts = slices.Clone(manifest.Artifacts)
			lines := strings.Split(strings.TrimSuffix(string(f.Members["payload/entries.ndjson"]), "\n"), "\n")
			rows := make([]map[string]any, len(lines))
			historical, summary, empty := -1, -1, -1
			for i, line := range lines {
				if err := json.Unmarshal([]byte(line), &rows[i]); err != nil {
					t.Fatal(err)
				}
				text, hasText := rows[i]["description"].(string)
				if !hasText {
					empty = i
				} else if strings.HasPrefix(text, "Catalog note:") {
					historical = i
				} else {
					summary = i
				}
			}
			if historical < 0 || summary < 0 || empty < 0 {
				t.Fatal("missing description classes")
			}
			source := rows[historical]["source_refs"].([]any)[0].(map[string]any)
			index := int(source["artifact_index"].(float64))
			switch mode {
			case "historical missing":
				delete(rows[historical], "description")
			case "historical changed":
				rows[historical]["description"] = "Current application"
			case "summary changed":
				rows[summary]["description"] = "Unsupported editorial text"
			case "null description":
				rows[empty]["description"] = "Invented description"
			case "wrong index":
				source["artifact_index"] = (index + 1) % len(m.Artifacts)
			case "source revision":
				m.Artifacts[index].Version = "different-pinned-revision"
			case "source digest":
				m.Artifacts[index].SHA256 = strings.Repeat("a", 64)
			case "source URL":
				m.Artifacts[index].Ref += "?other-source"
			case "missing provenance":
				rows[historical]["source_refs"] = []any{}
			case "duplicate provenance":
				rows[historical]["source_refs"] = []any{source, source}
			case "publisher":
				rows[historical]["publisher"] = nil
			case "SPN":
				rows[historical]["service_principal_names"] = []string{"synthetic"}
			case "deprecated":
				rows[historical]["deprecated"] = true
			case "replacement":
				rows[historical]["replacement_entry_id"] = rows[summary]["entry_id"]
			case "extensions":
				rows[historical]["extensions"] = map[string]any{"extra": true}
			}
			for i, row := range rows {
				b, err := canonicaljson.Marshal(row)
				if err != nil {
					t.Fatal(err)
				}
				lines[i] = string(b)
			}
			if err := checkEntraBaseline(m, []byte(strings.Join(lines, "\n")+"\n")); err == nil {
				t.Fatal("corpus oracle admitted", mode)
			}
		})
	}
}

package native

import (
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"strings"
	"testing"
)

func TestEntraSourceModelBoundaries(t *testing.T) {
	f := contentFixtureFor(t, "enrichment.entra_app_ids")
	original := f.Members["payload/entries.ndjson"]
	for _, tc := range []struct {
		name  string
		spns  []string
		valid bool
	}{
		{"punctuation", []string{"api://Application/", "HTTPS://Example.invalid/Path/?A=1"}, true},
		{"duplicate ASCII case", []string{"HTTPS://Example.invalid/", "https://example.invalid/"}, false},
		{"ASCII only", []string{"https://éxample.invalid"}, false},
		{"1024 bytes", []string{strings.Repeat("a", 1024)}, true},
		{"1025 bytes", []string{strings.Repeat("a", 1025)}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var row map[string]any
			if err := json.Unmarshal([]byte(original), &row); err != nil {
				t.Fatal(err)
			}
			row["service_principal_names"] = tc.spns
			raw, err := canonicaljson.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			f.Members["payload/entries.ndjson"] = string(raw) + "\n"
			err = ValidateContent(t.Context(), fixtureManifest(f), contentMemory(f.Members), nil)
			if (err == nil) != tc.valid {
				t.Fatal("SPN admission", err)
			}
		})
	}
	for _, count := range []int{64, 65} {
		var row map[string]any
		if err := json.Unmarshal([]byte(original), &row); err != nil {
			t.Fatal(err)
		}
		spns := []string{}
		for i := 0; i < count; i++ {
			spns = append(spns, fmt.Sprintf("api://%02d/", i))
		}
		row["service_principal_names"] = spns
		raw, err := canonicaljson.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		f.Members["payload/entries.ndjson"] = string(raw) + "\n"
		err = ValidateContent(t.Context(), fixtureManifest(f), contentMemory(f.Members), nil)
		if (err == nil) != (count == 64) {
			t.Fatal("SPN count", count, err)
		}
	}
	f.Members["payload/entries.ndjson"] = strings.ReplaceAll(original, "entra_app_ids.entry.v2", "entra_app_ids.entry.v1")
	if err := ValidateContent(t.Context(), fixtureManifest(f), contentMemory(f.Members), nil); err == nil {
		t.Fatal("v1 entry admitted")
	}
}

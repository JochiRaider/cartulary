package producer

import (
	"bytes"
	"encoding/json"
	"errors"
	inputs "github.com/JochiRaider/cartulary/referencepacks/entra/inputs"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"testing"
)

func TestEntraGenericSPNCoverage(t *testing.T) {
	profile, err := inputs.Files.ReadFile("dataset-profile.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	entry := func(id, spn string) map[string]any {
		return map[string]any{"schema_id": "cartulary.reference_pack.enrichment.entra_app_ids.entry.v2", "entry_id": id, "app_id": id, "display_label": "Synthetic application", "description": nil, "aliases": []string{}, "deprecated": false, "replacement_entry_id": nil, "source_refs": []any{map[string]any{"artifact_index": 0, "locator": "/synthetic"}}, "extensions": map[string]any{}, "publisher": nil, "service_principal_names": []string{spn}}
	}
	first := entry("00000000-0000-0000-0000-000000000001", "HTTP/Example.test")
	second := entry("00000000-0000-0000-0000-000000000002", "http/another.test")
	project := func(records ...map[string]any) (packageSource, dataset.Result, error) {
		var payload []byte
		for _, record := range records {
			payload = append(payload, encoded(t, record)...)
			payload = append(payload, '\n')
		}
		count := int64(len(records))
		m := envelope.Manifest{Key: EntraKey, SourceProfileID: EntraSourceProfile, SourceProfileSHA256: Digest(profile), Artifacts: []envelope.SourceArtifact{{Ref: "urn:synthetic:spn"}}}
		m.Summary.Kind = "entries"
		m.Summary.Entries = &count
		members := packageSource{"payload/entries.ndjson": payload}
		result, err := projectDataset(t.Context(), &m, members)
		return members, result, err
	}
	members, result, err := project(first, second)
	if err != nil {
		t.Fatal(err)
	}
	key, err := dataset.Normalize(result.Descriptor, "service_principal_name", json.RawMessage(`"http/EXAMPLE.test"`))
	if err != nil || string(key) != `["http/example.test"]` {
		t.Fatal("ASCII normalization", string(key), err)
	}
	var coverage dataset.Coverage
	for _, lookup := range result.Descriptor.Lookups {
		if lookup.Name == "service_principal_name" {
			coverage = lookup.Coverage
		}
	}
	if coverage != (dataset.Coverage{Objects: 2, Postings: 2, DistinctKeys: 2}) {
		t.Fatal("actual SPN coverage", coverage)
	}
	rows := bytes.Split(bytes.TrimSuffix(members[dataset.ObjectsPath], []byte{'\n'}), []byte{'\n'})
	var object dataset.Object
	if err = json.Unmarshal(rows[0], &object); err != nil {
		t.Fatal(err)
	}
	if object.ID != first["entry_id"] || object.SourceRefs[0].Locator != "/synthetic" || !bytes.Equal(object.Data, encoded(t, first)) {
		t.Fatal("stable identity, provenance, or native data lost")
	}
	second["service_principal_names"] = []string{"http/example.test"}
	if _, _, err = project(first, second); err == nil {
		t.Fatal("shared normalized SPN accepted under current source policy")
	}
	first["service_principal_names"] = []string{"HTTP/Example.test", "http/example.test"}
	if _, _, err = project(first); err == nil {
		t.Fatal("within-entry normalized duplicate accepted")
	}
	first["service_principal_names"] = []string{}
	_, result, err = project(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, lookup := range result.Descriptor.Lookups {
		if lookup.Name == "service_principal_name" && lookup.Coverage != (dataset.Coverage{}) {
			t.Fatal("invented SPN coverage")
		}
	}
	if _, err = dataset.Normalize(result.Descriptor, "unknown", json.RawMessage(`"valid"`)); !errors.Is(err, dataset.ErrUnsupportedLookup) {
		t.Fatal("unsupported lookup outcome", err)
	}
	if _, err = dataset.Normalize(result.Descriptor, "service_principal_name", json.RawMessage(`"bad\u0000input"`)); !errors.Is(err, dataset.ErrInvalidInput) {
		t.Fatal("malformed lookup outcome", err)
	}
}

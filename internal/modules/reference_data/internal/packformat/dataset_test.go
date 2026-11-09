package packformat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

func TestGenericDatasetRuntime_Unit(t *testing.T) {
	ctx := context.Background()
	descriptor := `{"lookups":[{"components":[{"name":"value","normalization":"ascii_case_insensitive","type":"string"}],"coverage":{"distinct_keys":1,"objects":1,"postings":1},"input_type":"scalar","name":"label","unique":false}],"schema_id":"cartulary.reference_pack.dataset.descriptor.v1"}`
	object := `{"data":{"native":{"unrecognized_type":"inert"}},"deprecated":false,"description":null,"display_label":"Example","lookup_values":{"label":["EXAMPLE"]},"object_id":"source-stable:1","object_type":"new_example","replacement_object_id":null,"schema_id":"cartulary.reference_pack.dataset.object.v1","source_refs":[{"artifact_index":0,"locator":"/records/0"}]}`
	source := contentMemory{"payload/descriptor.json": descriptor, "payload/objects.ndjson": object + "\n", "payload/relationships.ndjson": ""}
	for _, key := range []string{"advisory.unregistered_fifth_dataset", "framework.attack", "framework.veris", "enrichment.windows_event_ids", "enrichment.entra_app_ids"} {
		t.Run(key, func(t *testing.T) {
			p, ok := ProfileForKey(key)
			if !ok {
				t.Fatal("generic profile registration required")
			}
			m := Manifest{Contract: ContractID, Key: key, Kind: p.Kind, ProfileID: p.ID, ProfileVersion: p.Version, Artifacts: []SourceArtifact{{Ref: "urn:example:source"}}}
			if p.Kind == "" {
				m.Kind = "enrichment"
			}
			m.Compatibility.Majors = []int{1}
			m.Compatibility.Capabilities = []string{}
			objects, relationships := int64(1), int64(0)
			m.Summary.Kind = "objects_relationships"
			m.Summary.Objects = &objects
			m.Summary.Relationships = &relationships
			if err := ValidateManifestContract(m); err != nil {
				t.Fatal(err)
			}
			var rows contentRows
			if err := ValidateContent(ctx, m, source, &rows); err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].ID != "source-stable:1" || string(rows[0].Canonical) != object || len(rows[0].Keys) != 1 || rows[0].Keys[0] != (LookupKey{Kind: "label", Value: `["example"]`, Order: "source-stable:1"}) {
				t.Fatal("generic item or index changed", rows)
			}
			if err := ValidateEntryIdentity(key, "source-stable:1"); err != nil {
				t.Fatal(err)
			}
			m.Contract = "cartulary.reference_pack_contract.v2"
			if ValidateManifestContract(m) == nil {
				t.Fatal("predecessor contract accepted")
			}
			m.Contract = ContractID
			var changed map[string]any
			if err := json.Unmarshal([]byte(object), &changed); err != nil {
				t.Fatal(err)
			}
			changed["display_label"] = " invalid "
			changed["description"] = " invalid "
			raw, err := canonicaljson.Marshal(changed)
			if err != nil {
				t.Fatal(err)
			}
			bad := contentMemory{"payload/descriptor.json": descriptor, "payload/objects.ndjson": string(raw) + "\n", "payload/relationships.ndjson": ""}
			summary, err := CheckSummary(ctx, "content_semantics", nil, func(ctx context.Context, emit FindingSink) error { return ContentSemanticsFindings(ctx, m, bad, emit) })
			if err != nil || summary.Total != 2 || summary.Result != "failed" {
				t.Fatal("generic diagnostics lost", summary, err)
			}
		})
	}
	for _, key := range []string{"type_registry.host", "enrichment.tor", "framework.d3fend"} {
		m := Manifest{Contract: ContractID, Key: key, Kind: "enrichment", ProfileID: "cartulary.reference_pack.dataset.v1", ProfileVersion: "1"}
		if ValidateManifestContract(m) == nil {
			t.Fatal("reserved key profile substitution accepted", key)
		}
	}
	if _, ok := ProfileForKey("type_registry.unregistered"); ok {
		t.Fatal("unknown registry granted authority")
	}
}

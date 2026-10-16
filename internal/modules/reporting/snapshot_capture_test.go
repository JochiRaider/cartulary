package reporting

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func TestSnapshotCaptureRetainsSourceOnce(t *testing.T) {
	incident := IncidentMetadataSnapshot{ID: "11111111-1111-4111-8111-111111111111", Title: "Capture", Status: "active", Version: 1}
	source := SnapshotSource{SourceFacts: exportprovider.EmptySourceFacts(), Fields: []ExportField{{SourceRecordID: "22222222-2222-4222-8222-222222222222", SourceRecordType: "host", Path: "/hosts/22222222-2222-4222-8222-222222222222", ContentClass: ContentClassSourceEvidence, SourceFamily: "host", Value: map[string]any{"display_name": "unique frozen source value", "nullable": nil}}}, Timeline: []exportprovider.TimelineSortFact{}}
	at := time.Date(2026, 10, 16, 7, 0, 0, 0, time.UTC)
	raw, digest, err := freezeSnapshotCapture(incident, fixtureSnapshotID, at, "boundary", reference_data.SetBinding{}, source)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(raw, []byte("unique frozen source value")) != 1 || bytes.Contains(raw, []byte(`"sections"`)) || bytes.Contains(raw, []byte(`"source_index"`)) {
		t.Fatal("capture duplicated source as presentation")
	}
	decoded, err := decodeSnapshotCapture(raw, digest)
	if err != nil {
		t.Fatal(err)
	}
	gotIncident, gotSource := decoded.source()
	if gotIncident.ID != incident.ID || len(gotSource.Fields) != 1 || len(gotSource.Timeline) != 0 || gotSource.Fields[0].Value.(map[string]any)["nullable"] != nil {
		t.Fatal("source shape changed")
	}
	source.Fields[0].Value.(map[string]any)["display_name"] = "changed caller state"
	if !bytes.Contains(raw, []byte("unique frozen source value")) {
		t.Fatal("capture did not freeze input bytes")
	}
	var original map[string]json.RawMessage
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]func(map[string]json.RawMessage){
		"unknown":        func(v map[string]json.RawMessage) { v["future"] = json.RawMessage(`true`) },
		"retired":        func(v map[string]json.RawMessage) { v["schema_id"] = json.RawMessage(`"retired"`) },
		"missing fields": func(v map[string]json.RawMessage) { delete(v, "fields") },
		"null fields":    func(v map[string]json.RawMessage) { v["fields"] = json.RawMessage(`null`) },
		"foreign incident": func(v map[string]json.RawMessage) {
			v["incident_id"] = json.RawMessage(`"33333333-3333-4333-8333-333333333333"`)
		},
	} {
		t.Run(name, func(t *testing.T) {
			value := map[string]json.RawMessage{}
			for k, v := range original {
				value[k] = v
			}
			edit(value)
			changed, err := canonicalJSON(value)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decodeSnapshotCapture(changed, reportingObjectDigest(snapshotCaptureSchemaID, changed)); err == nil {
				t.Fatal("malformed capture admitted after rehash")
			}
		})
	}
	if _, err := decodeSnapshotCapture(raw, strings.Repeat("0", 64)); err == nil {
		t.Fatal("altered capture digest admitted")
	}
	duplicate := decoded
	duplicate.Fields = append(append([]capturedSourceField{}, decoded.Fields...), decoded.Fields[0])
	if validateSnapshotCapture(duplicate) == nil {
		t.Fatal("duplicate source path admitted")
	}
	missingSort := decoded
	missingSort.Fields = append([]capturedSourceField{}, decoded.Fields...)
	missingSort.Fields[0].SourceFamily = "timeline_event"
	missingSort.Fields[0].SourceRecordType = "timeline_event"
	if validateSnapshotCapture(missingSort) == nil {
		t.Fatal("Timeline source without sort fact admitted")
	}
}

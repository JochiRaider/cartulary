package reporting

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func testSnapshotFrozenFieldAdmission(t *testing.T) {
	build := func() SnapshotModel {
		t.Helper()
		model, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID,
			time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource([]ExportField{{
				SourceRecordID: "event", SourceRecordType: "timeline_event", SourceFamily: "timeline_event",
				Path: "/opaque/source", ContentClass: ContentClassSourceEvidence,
				Value: map[string]any{"activity": "Frozen value", "capture_state": "rough", "nullable": nil},
			}}))
		if err != nil {
			t.Fatal(err)
		}
		return model
	}
	valid := build()
	if err := validateSnapshotModelIdentity(valid); err != nil {
		t.Fatal("valid source copies rejected", err)
	}
	for name, change := range map[string]func(*SnapshotModel){
		"section value differs":          func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].Value = "changed" },
		"record value differs":           func(m *SnapshotModel) { m.Records[0].Fields[0].Value = "changed" },
		"event value differs":            func(m *SnapshotModel) { m.TimelineEvents[0].Fields[0].Value = "changed" },
		"record source differs":          func(m *SnapshotModel) { m.Records[0].Fields[0].SourceRefs = []string{"/unknown"} },
		"record support differs":         func(m *SnapshotModel) { m.Records[0].Fields[0].SupportRefs = []string{"support"} },
		"record partition differs":       func(m *SnapshotModel) { m.Records[0].Fields[0].DisclosurePartitionRefs = []string{"party:other"} },
		"record field missing":           func(m *SnapshotModel) { m.Records[0].Fields = m.Records[0].Fields[:1] },
		"record reference missing":       func(m *SnapshotModel) { m.Records[0].SourceRefs = m.Records[0].SourceRefs[:1] },
		"record support union differs":   func(m *SnapshotModel) { m.Records[0].SupportRefs = []string{"support"} },
		"record partition union differs": func(m *SnapshotModel) { m.Records[0].DisclosurePartitionRefs = []string{"party:other"} },
		"duplicate record":               func(m *SnapshotModel) { m.Records = append(m.Records, m.Records[0]) },
		"duplicate section selector":     func(m *SnapshotModel) { m.Sections = append(m.Sections, m.Sections[0]) },
		"duplicate event":                func(m *SnapshotModel) { m.TimelineEvents = append(m.TimelineEvents, m.TimelineEvents[0]) },
		"missing event":                  func(m *SnapshotModel) { m.TimelineEvents = nil },
		"event identity differs":         func(m *SnapshotModel) { m.TimelineEvents[0].RecordID = "other" },
		"event partition differs":        func(m *SnapshotModel) { m.TimelineEvents[0].DisclosurePartitionRefs = []string{"party:other"} },
		"event support differs":          func(m *SnapshotModel) { m.TimelineEvents[0].SupportRefs = []string{"support"} },
		"event source list differs":      func(m *SnapshotModel) { m.TimelineEvents[0].SourceRefs = []string{"/unknown"} },
		"field schema differs":           func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].SchemaID = "retired" },
		"field ordinal gap":              func(m *SnapshotModel) { m.Records[0].Fields[0].FieldOrdinal = 2 },
		"field key invalid":              func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].FieldKey = "owner.value" },
		"duplicate block field key": func(m *SnapshotModel) {
			m.Sections[0].Blocks[0].Fields[1].FieldKey = m.Sections[0].Blocks[0].Fields[0].FieldKey
		},
		"duplicate record field key": func(m *SnapshotModel) { m.Records[0].Fields[1].FieldKey = m.Records[0].Fields[0].FieldKey },
		"duplicate event field key": func(m *SnapshotModel) {
			m.TimelineEvents[0].Fields[1].FieldKey = m.TimelineEvents[0].Fields[0].FieldKey
		},
		"raw digest retained": func(m *SnapshotModel) {
			digest := strings.Repeat("a", 64)
			m.Records[0].Fields[0].RawValueSHA256 = &digest
		},
		"null source for present value": func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].SourceValueState = "null" },
		"present source for null value": func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[2].SourceValueState = "present" },
		"retained dropped state":        func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].RedactedValueState = "dropped" },
		"missing source array":          func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].SourceRefs = nil },
		"missing support array":         func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].SupportRefs = nil },
		"missing partition array":       func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].DisclosurePartitionRefs = nil },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := build()
			change(&candidate)
			if err := validateSnapshotModelIdentity(candidate); err == nil {
				t.Fatal("inconsistent frozen field admitted")
			}
			if fields, err := candidate.RedactionFields(); err == nil || fields != nil {
				t.Fatal("redaction consumed inconsistent frozen fields")
			}
			// An independently recomputed digest does not make inconsistent
			// copies valid during import/recovery admission.
			raw, err := canonicalJSON(candidate)
			if err != nil {
				t.Fatal(err)
			}
			decoded, _, err := decodePortableObject[SnapshotModel](raw, reportingObjectDigest(SnapshotModelSchemaID, raw))
			if err == nil && validateSnapshotModelIdentity(decoded) == nil {
				t.Fatal("inconsistent field accepted with a matching hostile digest")
			}
		})
	}
	// JSONB decoders may use a different Go numeric type. Comparison follows
	// canonical value semantics and does not depend on the decoder's Go types.
	field := valid.Records[0].Fields[0]
	field.Value = json.Number("1")
	left, err := reportingFieldContent(field)
	if err != nil {
		t.Fatal(err)
	}
	field.Value = float64(1)
	right, err := reportingFieldContent(field)
	if err != nil || string(left) != string(right) {
		t.Fatal("equivalent numeric copies differ", err)
	}
}

package reporting

import (
	"testing"
	"time"
)

func testCoreRecordSourceSeparation(t *testing.T) {
	fields := []ExportField{
		{Path: "/owner/envelope", SourceFamily: "record_envelope", SourceRecordID: "host", SourceRecordType: "host", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "host", "row_version": 8}, DisclosurePartitionRefs: []string{"internal_only"}},
		{Path: "/owner/host", SourceFamily: "host", SourceRecordID: "host", SourceRecordType: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Host", "row_version": 2}, DisclosurePartitionRefs: []string{"public"}},
		{Path: "/owner/note", SourceFamily: "note", SourceRecordID: "note", SourceRecordType: "artifact", ContentClass: ContentClassWorkingMaterial, Value: map[string]any{"title": "Note"}},
		{Path: "/owner/tag", SourceFamily: "record_tag", SourceRecordID: "tag", SourceRecordType: "record_tag", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"tag_name": "review"}},
		{Path: "/owner/mention", SourceFamily: "entity_mention", SourceRecordID: "host", SourceRecordType: "entity_mention", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"raw_text": "Rough host", "entity_type": "host", "source_record_id": "note", "source_field_key": "note.body", "resolution_status": "unresolved", "resolved_record_id": nil}},
		{Path: "/owner/indicator", SourceFamily: "record_envelope", SourceRecordID: "indicator", SourceRecordType: "indicator", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "indicator"}},
	}
	build := func(source []ExportField) (SnapshotModel, error) {
		model, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource(source))
		return model, err
	}
	model, err := build(fields)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Records) != 3 || len(model.Subjects) != 2 {
		t.Fatal("auxiliary groups became Core records or lost their subjects")
	}
	byID := map[string]ReportingRecordSummary{}
	for _, record := range model.Records {
		byID[record.RecordID] = record
	}
	if byID["note"].RecordType != "artifact" || byID["note"].SourceRecordRef.SourceFamily != "note" || byID["indicator"].SourceRecordRef.SourceFamily != "record_envelope" {
		t.Fatal("Core type and source family were conflated")
	}
	host := byID["host"]
	if host.RecordType != "host" || len(host.Fields) != 2 || host.Fields[1].FieldKey != "row_version" || !sameReportingValue(host.Fields[1].Value, 2) {
		t.Fatal("envelope fields replaced or merged with domain fields")
	}
	groups, err := snapshotSourceGroups(model.SnapshotContent)
	if err != nil || len(groups) != len(fields) {
		t.Fatal("auxiliary provenance was lost", err)
	}
	if _, err := model.RedactionFields(); err != nil {
		t.Fatal("separate source groups cannot be redacted", err)
	}
	displays := map[string]string{}
	for _, subject := range model.Subjects {
		displays[subject.StableSubjectRef] = subject.CanonicalDisplayValueSHA256
	}
	if displays["subject:host:record:host"] == displays["subject:host:mention:host"] {
		t.Fatal("auxiliary ID collided with Core display binding")
	}
	for name, change := range map[string]func(*SnapshotModel){
		"missing Core record":                func(m *SnapshotModel) { m.Records = m.Records[:2] },
		"duplicate Core identity":            func(m *SnapshotModel) { m.Records = append(m.Records, m.Records[0]) },
		"artifact subtype used as Core type": func(m *SnapshotModel) { m.Records[0].RecordType = "note" },
		"auxiliary summary": func(m *SnapshotModel) {
			for _, group := range groups {
				if group.SourceRecordRef.SourceFamily == "record_tag" {
					m.Records = append(m.Records, group.coreSummary())
				}
			}
		},
		"changed display label": func(m *SnapshotModel) { value := "invented"; m.Records[0].DisplayName = &value },
	} {
		t.Run(name, func(t *testing.T) {
			candidate, err := build(fields)
			if err != nil {
				t.Fatal(err)
			}
			change(&candidate)
			if validateSnapshotModelIdentity(candidate) == nil {
				t.Fatal("inconsistent Core projection admitted")
			}
		})
	}
	for _, kind := range []string{"identity", "record_envelope", "post"} {
		candidate := append([]ExportField{}, fields...)
		candidate[0].SourceRecordType = kind
		candidate[0].Value = map[string]any{"record_type": kind, "row_version": 8}
		if _, err := build(candidate); err == nil {
			t.Fatal("conflicting or unowned envelope type admitted", kind)
		}
	}
}

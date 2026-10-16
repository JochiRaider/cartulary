package reporting

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestSnapshotRecordMaterialization(t *testing.T) {
	t.Run("record lifecycle", testSnapshotRecordLifecycle)
	t.Run("Core and auxiliary source separation", testCoreRecordSourceSeparation)
	t.Run("canonical source index", testSnapshotSourceIndex)
	t.Run("frozen disclosure sources", testFrozenDisclosureSources)
	t.Run("source classification and summary", testSnapshotContentClasses)
	t.Run("complete subject source binding", testSnapshotSubjectRecordBinding)
	t.Run("source container scalar projection", testSnapshotSourceContainerProjection)
	t.Run("frozen field admission", testSnapshotFrozenFieldAdmission)
	t.Run("timeline sort facts", testSnapshotTimelineSort)
	fields := []ExportField{
		{SourceRecordID: "event-1", SourceRecordType: "timeline_event", Path: "/timeline_metadata/event-1", SourceFamily: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"metadata": "metadata"}, DisclosurePartitionRefs: []string{"party:beta"}, SupportRefs: []string{"/timeline/event-2"}},
		{SourceRecordID: "event-2", SourceRecordType: "timeline_event", Path: "/timeline/event-2", SourceFamily: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "second"}, DisclosurePartitionRefs: []string{"public"}},
		{SourceRecordID: "event-1", SourceRecordType: "timeline_event", Path: "/timeline/event-1", SourceFamily: "timeline_event", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "first": "first"}, DisclosurePartitionRefs: []string{"party:alpha"}, SupportRefs: []string{"/timeline/event-1"}},
	}
	input, err := canonicalJSON(fields)
	if err != nil {
		t.Fatal(err)
	}
	build := func(input []ExportField) (SnapshotModel, string, error) {
		return buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource(input))
	}
	model, digest, err := build(fields)
	if err != nil {
		t.Fatal(err)
	}
	after, err := canonicalJSON(fields)
	if err != nil || !bytes.Equal(input, after) {
		t.Fatal("materialization mutated caller-owned source fields", err)
	}
	if len(model.Records) != 2 || len(model.TimelineEvents) != 2 {
		t.Fatalf("wanted one record and event per source identity, got %d/%d", len(model.Records), len(model.TimelineEvents))
	}
	for index, id := range []string{"event-1", "event-2"} {
		event, record := model.TimelineEvents[index], model.Records[index]
		if event.RecordID != id || record.RecordID != id || record.SourceRecordRef.SourceSnapshotID != fixtureSnapshotID {
			t.Fatal("event lost its canonical source binding")
		}
		if !reflect.DeepEqual(event.Fields, record.Fields) || !reflect.DeepEqual(event.SourceRefs, record.SourceRefs) || !reflect.DeepEqual(event.SupportRefs, record.SupportRefs) || !reflect.DeepEqual(event.DisclosurePartitionRefs, record.DisclosurePartitionRefs) {
			t.Fatal("event lost part of its source record")
		}
		for ordinal, field := range event.Fields {
			if field.FieldOrdinal != ordinal+1 {
				t.Fatal("event fields must have contiguous ordinals")
			}
		}
	}
	first := model.TimelineEvents[0]
	wantSupport, _, err := materializeSupportReferences([]string{"/timeline/event-1", "/timeline/event-2"}, "incident", fixtureSnapshotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Fields) != 3 || !reflect.DeepEqual(first.DisclosurePartitionRefs, []string{"party:alpha", "party:beta"}) || !reflect.DeepEqual(first.SupportRefs, wantSupport) {
		t.Fatal("multi-field record did not retain all fields, partitions and support")
	}
	reversed := []ExportField{fields[2], fields[1], fields[0]}
	other, otherDigest, err := build(reversed)
	if err != nil {
		t.Fatal(err)
	}
	oneBytes, err := canonicalJSON(model)
	if err != nil {
		t.Fatal(err)
	}
	otherBytes, err := canonicalJSON(other)
	if err != nil || digest != otherDigest || !bytes.Equal(oneBytes, otherBytes) {
		t.Fatal("source traversal order changed the frozen materialization", err)
	}
	renamed := append([]ExportField(nil), fields...)
	renamed[0].Path = "/owner-selector/arbitrary-display-path"
	renameModel, _, err := build(renamed)
	if err != nil || len(renameModel.Records) != 2 || renameModel.Records[0].RecordID != "event-1" || len(renameModel.Records[0].Fields) != 3 {
		t.Fatal("materialization inferred identity from selector spelling", err)
	}
	projected, err := renameModel.RedactionFields()
	if err != nil {
		t.Fatal(err)
	}
	foundRenamed := false
	for _, field := range projected {
		if field.Value != "metadata" {
			continue
		}
		foundRenamed = true
		if field.SourceFamily != "timeline_event" || field.SourceRecordID != "event-1" || field.SourceRecordType != "timeline_event" {
			t.Fatal("redaction reconstructed source identity from path spelling")
		}
	}
	if !foundRenamed {
		t.Fatal("renamed source binding disappeared")
	}
	// Distinct source selectors must not create ambiguous canonical field keys
	// in their shared record. Reject both input orders without dropping a value.
	collision := append([]ExportField(nil), fields...)
	collision[0].Value = map[string]any{"first": "competing"}
	for _, candidate := range [][]ExportField{collision, {collision[2], collision[1], collision[0]}} {
		model, digest, err := build(candidate)
		var invalid *renderValidationError
		if !errors.As(err, &invalid) || invalid.ReasonCode != "duplicate_stable_id" || digest != "" || len(model.Records) != 0 {
			t.Fatal("ambiguous field list produced model bytes", err)
		}
	}
	broken := renameModel
	broken.Records = nil
	if _, err := broken.RedactionFields(); err == nil {
		t.Fatal("orphaned structured field reused an inferred source family")
	}
	conflictingType := fields[0]
	conflictingType.SourceRecordType = "host"
	missingIdentity := fields[0]
	missingIdentity.SourceRecordID = ""
	for name, candidate := range map[string][]ExportField{
		"duplicate":              {fields[0], fields[0]},
		"missing owner identity": {missingIdentity},
		"conflicting owner type": {conflictingType, fields[2]},
		"empty":                  {{Path: "", SourceFamily: "timeline_event"}},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := build(candidate)
			var validation *renderValidationError
			if !errors.As(err, &validation) || validation.FailureCode != "export_model_invalid" {
				t.Fatal("invalid source path was admitted", err)
			}
		})
	}
	for name, change := range map[string]func(*SnapshotModel){
		"missing classification":   func(m *SnapshotModel) { m.Sections[0].Blocks[0].ContentClass = "" },
		"unknown classification":   func(m *SnapshotModel) { m.Sections[0].Blocks[0].ContentClass = "public" },
		"padded classification":    func(m *SnapshotModel) { m.Sections[0].Blocks[0].ContentClass = " source_evidence " },
		"missing source reference": func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].SourceRefs = nil },
		"empty source reference":   func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].SourceRefs = []string{""} },
	} {
		t.Run(name, func(t *testing.T) {
			candidate, _, err := build(fields)
			if err != nil {
				t.Fatal(err)
			}
			change(&candidate)
			projected, err := candidate.RedactionFields()
			var failure *renderValidationError
			if !errors.As(err, &failure) || projected != nil {
				t.Fatal("redaction repaired malformed structured source from path spelling")
			}
			wantCode, wantReason := "export_model_invalid", "invalid_scalar_value"
			if name == "missing classification" {
				wantCode, wantReason = "content_class_missing", "content_class_missing"
			}
			if name == "missing source reference" || name == "empty source reference" {
				wantReason = "dangling_source_ref"
			}
			if failure.FailureCode != wantCode || failure.ReasonCode != wantReason {
				t.Fatalf("failure = %v, want %s/%s", err, wantCode, wantReason)
			}
		})
	}
}

func testSnapshotSourceContainerProjection(t *testing.T) {
	source := ExportField{Path: "/opaque/source", SourceFamily: "host", SourceRecordID: "host-1", SourceRecordType: "host", ContentClass: ContentClassSourceEvidence,
		DisclosurePartitionRefs: []string{"public", "party:a", "internal_only", "blocked"}, SupportRefs: []string{"/hosts/host-1"},
		Value: map[string]any{"display_name": "Host A", "active": true, "aliases": []any{"one", "two"}, "nullable": nil},
	}
	build := func(fields []ExportField) (SnapshotModel, string, error) {
		return buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource(fields))
	}
	model, _, err := build([]ExportField{source})
	if err != nil {
		t.Fatal(err)
	}
	fields := model.Sections[0].Blocks[0].Fields
	for _, invalid := range [][]string{nil, {"party:a", "public"}, {"public", "public"}, {"public_summary"}, {"party:bad:nested"}} {
		changed := fields[0]
		changed.DisclosurePartitionRefs = invalid
		if validReportingField(changed, 1, false) {
			t.Fatal("invalid frozen partition metadata accepted", invalid)
		}
	}
	if len(fields) != 4 || len(model.Records[0].Fields) != 4 {
		t.Fatal("source object was retained as one field")
	}
	for i, key := range []string{"active", "aliases", "display_name", "nullable"} {
		field := fields[i]
		if field.FieldKey != key || field.FieldOrdinal != i+1 || field.DisplayLabel != nil || len(field.SourceRefs) != 1 || snapshotSourceIndex(model.SnapshotContent)[field.SourceRefs[0]].SourcePath == nil || !reflect.DeepEqual(field.SupportRefs, []string{model.SupportIndex[0].SupportRefID}) || !reflect.DeepEqual(field.DisclosurePartitionRefs, source.DisclosurePartitionRefs) {
			t.Fatal("projected field lost ordering, null label or owner metadata", field)
		}
	}
	projected, err := model.RedactionFields()
	if err != nil || len(projected) != 4 {
		t.Fatal("redaction did not consume exact scalar fields", err)
	}
	for _, field := range projected {
		if field.SourceRecordID != source.SourceRecordID || field.SourceFamily != source.SourceFamily {
			t.Fatal("scalar selector replaced source identity")
		}
	}
	repeated := source
	repeated.SupportRefs = []string{"/incident/status", "/hosts/host-1", "/incident/title", "/incident/status"}
	coalesced, _, err := build([]ExportField{repeated})
	if err != nil {
		t.Fatal(err)
	}
	if len(coalesced.SupportIndex) != 2 || !reflect.DeepEqual(coalesced.Records[0].Fields[0].SupportRefs, []string{coalesced.SupportIndex[0].SupportRefID, coalesced.SupportIndex[1].SupportRefID}) ||
		!reflect.DeepEqual(repeated.SupportRefs, []string{"/incident/status", "/hosts/host-1", "/incident/title", "/incident/status"}) {
		t.Fatal("Reporting did not coalesce references without changing owner input")
	}
	for _, invalid := range []any{map[string]any{}, map[string]any{"nested": map[string]any{"private": "value"}}, map[string]any{"list": []any{map[string]any{}}}, map[string]any{"unsafe/name": "value"}} {
		candidate := source
		candidate.Value = invalid
		model, digest, err := build([]ExportField{candidate})
		if err == nil || digest != "" || len(model.Sections) != 0 {
			t.Fatal("invalid container published a snapshot")
		}
	}
	alias := source
	alias.Path = source.Path + "/active"
	alias.Value = false
	if _, _, err := build([]ExportField{source, alias}); err == nil {
		t.Fatal("projected selector collision accepted")
	}
	model.Sections[0].Blocks[0].Fields[0].Value = map[string]any{"legacy": "row"}
	if _, err := model.RedactionFields(); err == nil {
		t.Fatal("retired object-valued field read as canonical content")
	}
	for _, change := range []func(*SnapshotModel){
		func(m *SnapshotModel) { m.Records = nil },
		func(m *SnapshotModel) { m.Sections[0].Blocks[0].Fields[0].Value = map[string]any{"legacy": "row"} },
		func(m *SnapshotModel) { m.Records[0].Fields[0].Value = []any{[]any{}} },
		func(m *SnapshotModel) {
			m.TimelineEvents = []ReportingTimelineEvent{{Fields: []ReportingField{{Value: map[string]any{}}}}}
		},
	} {
		candidate, _, err := build([]ExportField{source})
		if err != nil {
			t.Fatal(err)
		}
		change(&candidate)
		if err := validateSnapshotModelIdentity(candidate); err == nil {
			t.Fatal("retired value shape passed frozen model admission")
		}
	}
}

func testSnapshotSubjectRecordBinding(t *testing.T) {
	build := func(fields []ExportField) (SnapshotModel, string, error) {
		return buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource(fields))
	}
	fields := []ExportField{
		{SourceRecordID: "host-1", SourceRecordType: "host", Path: "/owner/a", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Host A"}, DisclosurePartitionRefs: []string{"party:z"}},
		{SourceRecordID: "host-1", SourceRecordType: "host", Path: "/owner/z", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"dns_name": "internal.example"}, DisclosurePartitionRefs: []string{"party:a"}},
	}
	model, digest, err := build(fields)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Records) != 1 || len(model.Subjects) != 1 {
		t.Fatal("one source record must produce one subject")
	}
	subject, record := model.Subjects[0], model.Records[0]
	testCanonicalSubjectAdmission(t, model)
	if subject.StableSubjectRef != "subject:host:record:host-1" || subject.SubjectFamily != "host" || subject.SourceRecordID == nil || *subject.SourceRecordID != record.RecordID || subject.SourceRecordType == nil || *subject.SourceRecordType != "host" || subject.EntityMentionID != nil {
		t.Fatal("subject binding was overwritten by a later source field")
	}
	if !reflect.DeepEqual(subject.DisclosurePartitionRefs, []string{"party:a", "party:z"}) || !reflect.DeepEqual(subject.DisclosurePartitionRefs, record.DisclosurePartitionRefs) {
		t.Fatalf("subject lost source disclosure labels: %v", subject.DisclosurePartitionRefs)
	}
	expectedDisplayDigest, err := subjectDisplayValueDigest(subject.StableSubjectRef, record.DisplayName)
	if err != nil || record.DisplayName == nil || *record.DisplayName != "Host A" || subject.CanonicalDisplayValueSHA256 != expectedDisplayDigest {
		t.Fatal("subject did not bind the source display value digest")
	}
	// Coalescing leaves spare capacity in a source reference slice. Aggregating
	// another field must not sort over that backing array and relabel the first.
	repeated := append([]ExportField(nil), fields...)
	repeated[0].DisclosurePartitionRefs = []string{"party:z", "party:z"}
	coalesced, _, err := build(repeated)
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range coalesced.Sections[0].Blocks {
		for _, field := range block.Fields {
			if field.FieldKey == "display_name" && !reflect.DeepEqual(field.DisclosurePartitionRefs, []string{"party:z"}) {
				t.Fatal("aggregate sorting relabeled the first source field")
			}
		}
	}
	other, otherDigest, err := build([]ExportField{fields[1], fields[0]})
	if err != nil || digest != otherDigest || !reflect.DeepEqual(other.Subjects, model.Subjects) {
		t.Fatal("source traversal order changed subject materialization", err)
	}
	// A Core envelope ID cannot name both a Host and an Identity.
	// Distinct records still carry independent subject families.
	identity := fields[0]
	identity.SourceFamily, identity.SourceRecordType, identity.Path = "identity", "identity", "/identity/a"
	if _, _, err := build(append(fields, identity)); err == nil {
		t.Fatal("one envelope acquired two Core record types")
	}
	identity.SourceRecordID = "identity-1"
	distinct, _, err := build(append(fields, identity))
	if err != nil || len(distinct.Subjects) != 2 || distinct.Subjects[0].SubjectFamily == distinct.Subjects[1].SubjectFamily {
		t.Fatal("subject identity discarded its source family", err)
	}
	// One mention record cannot acquire both Host and Identity subject identities.
	for _, family := range []string{"host", "identity"} {
		mentionFields := append([]ExportField(nil), fields...)
		for index := range mentionFields {
			mentionFields[index].SourceFamily = "entity_mention"
			mentionFields[index].SourceRecordType = "entity_mention"
			if index == 0 {
				mentionFields[index].Value = map[string]any{"entity_type": family, "source_record_id": "origin"}
			} else {
				mentionFields[index].Value = map[string]any{"source_field_key": "evidence.description", "resolution_status": "unresolved", "resolved_record_id": nil}
			}
		}
		mentionFields = append(mentionFields, ExportField{SourceRecordID: "origin", SourceRecordType: "evidence", SourceFamily: "evidence", Path: "/evidence/origin", ContentClass: ContentClassSourceEvidence, Value: "origin"})
		mention, _, err := build(mentionFields)
		if err != nil || len(mention.Subjects) != 1 || mention.Subjects[0].StableSubjectRef != "subject:"+family+":mention:host-1" || mention.Subjects[0].SourceRecordID != nil || mention.Subjects[0].SourceRecordType != nil || mention.Subjects[0].EntityMentionID == nil || *mention.Subjects[0].EntityMentionID != "host-1" || !reflect.DeepEqual(mention.Subjects[0].DisclosurePartitionRefs, []string{"party:a", "party:z"}) {
			t.Fatal("mention did not preserve its declared subject family and complete labels", err)
		}
		mentionFields[1].Value.(map[string]any)["entity_type"] = "party"
		if _, _, err := build(mentionFields); err == nil {
			t.Fatal("unsupported mention subject family admitted")
		}
		mentionFields[1].Value.(map[string]any)["entity_type"] = "identity"
		if family == "host" {
			_, _, err := build(mentionFields)
			var failure *renderValidationError
			if !errors.As(err, &failure) || failure.FailureCode != "token_manifest_invalid" || failure.ReasonCode != "token_subject_not_stable" {
				t.Fatal("one mention record produced conflicting subject families", err)
			}
		}
	}
}

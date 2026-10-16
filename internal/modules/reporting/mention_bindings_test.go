package reporting

import (
	"bytes"
	"testing"
	"time"
)

func testResolvedMentionBinding(t *testing.T) {
	value := map[string]any{
		"entity_type": "host", "source_record_id": "origin", "source_field_key": "evidence.description",
		"resolution_status": "resolved", "resolved_record_id": "host", "raw_text": "raw alias",
	}
	build := func(value map[string]any) (SnapshotModel, string, error) {
		return buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource([]ExportField{
			{SourceRecordID: "host", SourceRecordType: "host", SourceFamily: "host", Path: "/hosts/host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Canonical Host"}},
			{SourceRecordID: "origin", SourceRecordType: "evidence", SourceFamily: "evidence", Path: "/evidence/origin", ContentClass: ContentClassSourceEvidence, Value: "original source"},
			{SourceRecordID: "mention", SourceRecordType: "entity_mention", SourceFamily: "entity_mention", Path: "/entity_mentions/mention", ContentClass: ContentClassSourceEvidence, Value: value},
		}))
	}
	snapshot, _, err := build(value)
	if err != nil || len(snapshot.Subjects) != 1 || snapshot.Subjects[0].StableSubjectRef != "subject:host:record:host" {
		t.Fatal("resolved mention acquired an independent unresolved identity", err)
	}
	model, digest, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	profile, profileSHA, err := ResolveRedactionProfile(TokenizedRedactionProfileID, DefaultRedactionProfileVersion, nil, snapshot.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, snapshot.SourceFacts, profile, profileSHA, digest, ReleaseScopeInternalReview, nil)
	if err != nil || result.RevealMap == nil || len(result.RevealMap.Entries) != 1 || result.RevealMap.Entries[0].CanonicalDisplayValue != "Canonical Host" || bytes.Contains(result.RevealMapJSON, []byte("raw alias")) {
		t.Fatal("resolved mention did not use the canonical subject/display binding", err)
	}
	for name, mutate := range map[string]func(map[string]any){
		"missing status":       func(v map[string]any) { delete(v, "resolution_status") },
		"unknown status":       func(v map[string]any) { v["resolution_status"] = "guessed" },
		"missing target arm":   func(v map[string]any) { delete(v, "resolved_record_id") },
		"null resolved target": func(v map[string]any) { v["resolved_record_id"] = nil },
		"missing target":       func(v map[string]any) { v["resolved_record_id"] = "absent" },
		"wrong target family":  func(v map[string]any) { v["entity_type"] = "identity" },
		"missing origin":       func(v map[string]any) { v["source_record_id"] = "absent" },
		"missing origin key":   func(v map[string]any) { delete(v, "source_field_key") },
		"mixed resolution":     func(v map[string]any) { v["resolution_status"] = "unresolved" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := map[string]any{}
			for key, v := range value {
				candidate[key] = v
			}
			mutate(candidate)
			if model, digest, err := build(candidate); err == nil || model.SchemaID != "" || digest != "" {
				t.Fatal("invalid mention binding yielded a usable snapshot", err)
			}
		})
	}
	// Omitting a mandatory subject from a rehashed retained model is still an
	// admission defect, not a request to reconstruct the missing census.
	model.Subjects = []TokenizableSubject{}
	if validateSnapshotSubjects(model.SnapshotContent) == nil {
		t.Fatal("missing canonical subject census accepted")
	}
}

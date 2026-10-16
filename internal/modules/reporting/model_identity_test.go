package reporting

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

// Expected IDs were calculated independently with Python's JSON serializer and
// SHA-256 over literal ASCII tuples. No expected output calls the Go encoder.
func TestReportingAdmittedModelIdentityAndTime(t *testing.T) {
	t.Run("lifecycle isolation", testReportingRenderIsolation)
	at := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	snapshot, _, err := BuildSnapshotModel(IncidentMetadataSnapshot{ID: "00000000-0000-0000-0000-000000000401", Title: "Frozen", Status: "active", Version: 1}, fixtureSnapshotID, at, "boundary", fixtureSnapshotSource([]ExportField{
		{SourceRecordID: "host-1", SourceRecordType: "host", Path: "/hosts/host-1", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: "Host"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.SnapshotModelID != "snapm_1e2d1bd4fcc5d6805a91211681179518eac0ee48c0620342ec53eb08bc74b4be" {
		t.Fatal("snapshot identity", snapshot.SnapshotModelID)
	}
	before, _ := canonicalJSON(snapshot)
	if bytes.Contains(before, []byte(`"source_snapshot_id":""`)) || bytes.Contains(before, []byte(`"render_admitted_at"`)) {
		t.Fatal("snapshot contains a fabricated render or empty source binding")
	}
	model, digest, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	if model.ExportModelID != "expm_23212251cd29933f68e12a9121df04a304ac538b633f2b9a50cd53bff1b7b599" || model.ExportModelCreatedAt != "2026-05-23T12:00:00.000000Z" {
		t.Fatal("render identity", model.ExportModelID, model.ExportModelCreatedAt)
	}
	raw, _ := canonicalJSON(model)
	if digest != reportingObjectDigest(model.SchemaID, raw) {
		t.Fatal("render digest")
	}
	if err := validateRenderModelIdentity(model); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*ExportModel){
		func(m *ExportModel) { m.SchemaID = "cartulary.reporting_export_model.v2" },
		func(m *ExportModel) { m.SchemaID = "cartulary.reporting_export_model.v3" },
		func(m *ExportModel) { m.SchemaID = "cartulary.reporting_export_model.v4" },
		func(m *ExportModel) { m.SourceIndex = nil },
		func(m *ExportModel) { m.PresentationTexts = nil },
		func(m *ExportModel) { m.PresentationTexts = []ReportingPresentationText{{AuthoredTextID: "injected"}} },
	} {
		candidate := model
		change(&candidate)
		if validateRenderModelIdentity(candidate) == nil {
			t.Fatal("retired or prepopulated presentation model admitted")
		}
	}
	token, err := model.RenderIdentity.generatedID("token", "tok_", map[string]any{"stable_subject_ref": "subject:host:record:host-1"})
	if err != nil || token != "tok_cbe2964b46d497a6a6fb3b588f0eab8e50154db6a50457179e7d0f353b7953c1" {
		t.Fatal("token identity", token, err)
	}
	deckID, err := model.RenderIdentity.generatedID("deck", "deck_", map[string]any{"template_id": "cartulary.report.default", "template_version": "1", "export_model_id": model.ExportModelID})
	if err != nil || deckID != "deck_51313e2dfd31a3675d302c3e71e75266078071008034b91cba136e9e046c19d7" {
		t.Fatal("deck identity", deckID, err)
	}
	after, _ := canonicalJSON(snapshot)
	if !bytes.Equal(before, after) {
		t.Fatal("render binding rewrote its snapshot")
	}
	previewID := *model.ReleaseID
	preview := RenderIdentity{PreviewAttemptID: &previewID, RenderAdmittedAt: model.RenderAdmittedAt}
	draft, _, err := bindRenderModel(snapshot, preview, ReleaseScopeInternalDraft, nil)
	if err != nil || draft.ExportModelID == model.ExportModelID {
		t.Fatal("preview identity collision", err)
	}
	if _, _, err := bindRenderModel(snapshot, preview, ReleaseScopeExternal, nil); err == nil {
		t.Fatal("preview authorized external release")
	}
	for _, invalid := range []RenderIdentity{
		{RenderAdmittedAt: model.RenderAdmittedAt},
		{ReleaseID: model.ReleaseID, PreviewAttemptID: &previewID, RenderAdmittedAt: model.RenderAdmittedAt},
		{ReleaseID: model.ReleaseID, RenderAdmittedAt: "2026-05-23T12:00:00Z"},
		{ReleaseID: model.ReleaseID, RenderAdmittedAt: "2026-05-23T12:00:00.0000000Z"},
	} {
		if _, _, err := bindRenderModel(snapshot, invalid, ReleaseScopeInternalReview, nil); err == nil {
			t.Fatal("invalid identity admitted")
		}
	}
	if len(snapshot.Records) == 0 {
		t.Fatal("fixture has no record")
	}
	for _, mutate := range []func(*ReportingRecordSummary){
		func(r *ReportingRecordSummary) { r.SourceRecordRef.SourceRecordID = "different" },
		func(r *ReportingRecordSummary) { r.SourceRecordRef.SourceFamily = "incident_metadata" },
		func(r *ReportingRecordSummary) { r.RecordType = "" },
	} {
		candidate := snapshot
		candidate.Records = append([]ReportingRecordSummary(nil), snapshot.Records...)
		mutate(&candidate.Records[0])
		if _, _, err := bindRenderModel(candidate, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil); err == nil {
			t.Fatal("invalid or fabricated source record identity admitted")
		}
	}
	snapshot.Records[0].SourceRecordRef.SourceSnapshotID = ""
	if _, _, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil); err == nil {
		t.Fatal("empty nested source identity repaired")
	}
	// Closed identity members, including nullable counterparts, must be present.
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"release_id", "preview_attempt_id", "render_admitted_at", "export_model_created_at", "snapshot_model_id"} {
		original := object[key]
		for _, mutation := range []string{"omitted", "null", "wrong_type"} {
			if key == "preview_attempt_id" && mutation == "null" {
				continue
			}
			switch mutation {
			case "omitted":
				delete(object, key)
			case "null":
				object[key] = json.RawMessage(`null`)
			case "wrong_type":
				object[key] = json.RawMessage(`[]`)
			}
			bad, err := canonicalJSON(object)
			if err != nil {
				t.Fatal(err)
			}
			// Supply the digest of the hostile original bytes, not a stale valid hash.
			admitted, _, err := decodePortableObject[ExportModel](bad, reportingObjectDigest(ExportModelSchemaID, bad))
			if err == nil && validateRenderModelIdentity(admitted) == nil {
				t.Fatalf("%s %s admitted", key, mutation)
			}
			object[key] = original
		}
	}
	object["undeclared"] = json.RawMessage(`true`)
	bad, err := canonicalJSON(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := decodePortableObject[ExportModel](bad, reportingObjectDigest(ExportModelSchemaID, bad)); err == nil {
		t.Fatal("unknown member admitted")
	}

}

func TestReportingTimestampPrecisionIsFixedAtAdmission(t *testing.T) {
	for _, ns := range []int{0, 100000000, 123456000} {
		at := time.Date(2026, 5, 23, 12, 0, 0, ns, time.UTC)
		stamp, err := normalizedReportingTimestamp(at)
		if err != nil || len(stamp) != 27 {
			t.Fatal(stamp, err)
		}
		parsed, err := time.Parse(reportingTimestampLayout, stamp)
		if err != nil || !parsed.Equal(at) {
			t.Fatal("timestamp changed value")
		}
	}
	for _, at := range []time.Time{{}, time.Date(2026, 5, 23, 12, 0, 0, 1, time.UTC), time.Date(2026, 5, 23, 12, 0, 0, 123456789, time.UTC)} {
		if _, err := normalizedReportingTimestamp(at); err == nil {
			t.Fatal("invalid timestamp repaired")
		}
	}
}

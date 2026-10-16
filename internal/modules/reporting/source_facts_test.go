package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/links/partyassignment"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/google/uuid"
)

func testFrozenDisclosureSources(t *testing.T) {
	t.Run("private deletion facts", testDeletedSourceFacts)
	t.Run("subject assignment permissions", testSubjectAssignmentPermissions)
	t.Run("reference disclosure closure", testReferencedDisclosurePartitions)
	t.Run("resolved mention binding", testResolvedMentionBinding)
	incident := uuid.MustParse("00000000-0000-0000-0000-000000003001")
	host := uuid.MustParse("00000000-0000-0000-0000-000000003002")
	party := uuid.MustParse("00000000-0000-0000-0000-000000003003")
	fact, err := partyassignment.NewFact(partyassignment.State{AssignmentID: uuid.MustParse("00000000-0000-0000-0000-000000003004"), IncidentID: incident, SubjectRecordID: host, SubjectRecordType: "host", PartyID: party, AssignmentKind: "subject_party", AssignmentState: "active", Provenance: "manual", CreatedByUserID: uuid.MustParse("00000000-0000-0000-0000-000000003005"), CreatedAt: "2026-05-01T00:00:00.000000Z"})
	if err != nil {
		t.Fatal(err)
	}
	source := fixtureSnapshotSource([]ExportField{
		{SourceRecordID: host.String(), SourceRecordType: "host", SourceFamily: "record_envelope", Path: "/record_envelopes/" + host.String(), ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "host", "value": "Source envelope"}},
		{SourceRecordID: party.String(), SourceRecordType: "party", SourceFamily: "record_envelope", Path: "/record_envelopes/" + party.String(), ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "party", "value": "Source envelope"}},
		{SourceRecordID: host.String(), SourceRecordType: "host", SourceFamily: "host", Path: "/opaque/host", ContentClass: ContentClassSourceEvidence, Value: "Private host"},
		{SourceRecordID: party.String(), SourceRecordType: "party", SourceFamily: "party", Path: "/opaque/party", ContentClass: ContentClassSourceEvidence, Value: "Private party"},
	})
	source.SourceFacts.Assignments = append(source.SourceFacts.Assignments, fact)
	model, _, err := BuildSnapshotModel(IncidentMetadataSnapshot{ID: incident.String(), Title: "Source facts", Status: "active", Version: 1}, fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", source)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := canonicalJSON(model)
	if err != nil {
		t.Fatal(err)
	}
	captureBytes, captureDigest, err := freezeSnapshotCapture(IncidentMetadataSnapshot{ID: incident.String(), Title: "Source facts", Status: "active", Version: 1}, model.SnapshotID, model.SnapshotAt, "boundary", model.ReferencePacks, source)
	if err != nil {
		t.Fatal(err)
	}
	source.SourceFacts.Parties[0].PublicDirectoryEligible = true
	source.SourceFacts.Assignments[0].Provenance = "import"
	unchanged, _ := canonicalJSON(model)
	if !bytes.Equal(raw, unchanged) {
		t.Fatal("snapshot aliases mutable owner facts")
	}
	retained := SnapshotRecord{SnapshotID: uuid.MustParse(model.SnapshotID), IncidentID: incident, SnapshotAt: model.SnapshotAt, DerivationVersion: model.DerivationVersion, SourceChangeSetHighWatermark: model.SourceChangeSetHighWatermark, SourceCaptureJSON: captureBytes, SourceCaptureSHA256: captureDigest}
	render, digest, err := bindRenderModel(model, fixtureRenderIdentity(), ReleaseScopeInternalDraft, nil)
	if err != nil {
		t.Fatal(err)
	}
	renderBytes, _ := canonicalJSON(render)
	if bytes.Contains(renderBytes, []byte(`"source_facts"`)) || bytes.Contains(renderBytes, []byte(fact.RowVersion)) {
		t.Fatal("private disclosure inputs entered the export model")
	}
	factsDigest, _ := sourceFactsDigest(model.SourceFacts)
	input := renderInput{ExportModel: render, ExportModelSHA256: digest, SourceFacts: model.SourceFacts.Clone(), SourceFactsSHA256: factsDigest}
	if err := validateFrozenSourceFacts(input); err != nil {
		t.Fatal(err)
	}
	if err := validateRenderSnapshotBinding(input, retained); err != nil {
		t.Fatal(err)
	}
	input.SourceFacts.Parties[0].PublicDirectoryEligible = true
	if validateFrozenSourceFacts(input) == nil {
		t.Fatal("source mutation retained its admitted digest")
	}
	input.SourceFactsSHA256, _ = sourceFactsDigest(input.SourceFacts)
	if validateFrozenSourceFacts(input) != nil {
		t.Fatal("substitution fixture must be internally consistent")
	}
	if validateRenderSnapshotBinding(input, retained) == nil {
		t.Fatal("recomputed job digest authorized source substitution")
	}
	for name, change := range map[string]func(*exportprovider.SourceFacts){
		"absent source contract":    func(f *exportprovider.SourceFacts) { f.SchemaID = "" },
		"retired source contract":   func(f *exportprovider.SourceFacts) { f.SchemaID += ".retired" },
		"missing assignments array": func(f *exportprovider.SourceFacts) { f.Assignments = nil },
		"missing party":             func(f *exportprovider.SourceFacts) { f.Parties = []exportprovider.PartyFact{} },
		"duplicate party":           func(f *exportprovider.SourceFacts) { f.Parties = append(f.Parties, f.Parties[0]) },
		"invented segment":          func(f *exportprovider.SourceFacts) { f.Parties[0].PartitionSegment = "invented" },
		"duplicate assignment":      func(f *exportprovider.SourceFacts) { f.Assignments = append(f.Assignments, f.Assignments[0]) },
		"altered provenance":        func(f *exportprovider.SourceFacts) { f.Assignments[0].Provenance = "import" },
		"foreign incident": func(f *exportprovider.SourceFacts) {
			state := f.Assignments[0].State
			state.IncidentID = uuid.New()
			f.Assignments[0], _ = partyassignment.NewFact(state)
		},
		"foreign endpoint": func(f *exportprovider.SourceFacts) {
			state := f.Assignments[0].State
			state.PartyID = uuid.New()
			f.Assignments[0], _ = partyassignment.NewFact(state)
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := model
			candidate.SourceFacts = model.SourceFacts.Clone()
			change(&candidate.SourceFacts)
			if validateSnapshotModelIdentity(candidate) == nil {
				t.Fatal("invalid disclosure source admitted")
			}
			bad := input
			bad.SourceFacts = candidate.SourceFacts
			runtime := &limitObservingRuntime{maxBytes: -1}
			result, _, err := (&Renderer{runtime: runtime}).Render(context.Background(), bad)
			if err == nil || runtime.maxBytes != -1 || len(result.RenderBundle.Files) != 0 {
				t.Fatal("invalid source reached native renderer or emitted output")
			}
		})
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"source_facts", "schema_id"} {
		candidate := map[string]json.RawMessage{}
		for k, v := range object {
			candidate[k] = v
		}
		delete(candidate, key)
		encoded, _ := canonicalJSON(candidate)
		if _, _, err := decodePortableObject[SnapshotModel](encoded, reportingObjectDigest(model.SchemaID, encoded)); err == nil {
			t.Fatal("portable snapshot accepted omitted source shape", key)
		}
	}
}

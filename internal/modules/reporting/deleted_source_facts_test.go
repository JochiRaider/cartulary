package reporting

import (
	"bytes"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func testDeletedSourceFacts(t *testing.T) {
	source := fixtureSnapshotSource([]ExportField{{Path: "/incident/title", SourceFamily: "incident_metadata", ContentClass: ContentClassDerivedAnalytic, Value: "Kept"}})
	source.SourceFacts.DeletedRecords = []exportprovider.DeletedRecordFact{{RecordID: "deleted-record-identity", RecordType: "evidence"}}
	snapshot, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", source)
	if err != nil {
		t.Fatal(err)
	}
	model, _, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	profile, digest, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, snapshot.SourceFacts, profile, digest, "", ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(result.ModelJSON, []byte("deleted-record-identity")) || bytes.Contains(result.ModelJSON, []byte("deleted_records")) {
		t.Fatal("private deletion identity entered retained case content")
	}
	for name, mutate := range map[string]func(*exportprovider.SourceFacts){
		"retired schema":   func(f *exportprovider.SourceFacts) { f.SchemaID = "cartulary.reporting_source_facts.v1" },
		"missing roster":   func(f *exportprovider.SourceFacts) { f.DeletedRecords = nil },
		"duplicate":        func(f *exportprovider.SourceFacts) { f.DeletedRecords = append(f.DeletedRecords, f.DeletedRecords[0]) },
		"invalid type":     func(f *exportprovider.SourceFacts) { f.DeletedRecords[0].RecordType = "post" },
		"invalid identity": func(f *exportprovider.SourceFacts) { f.DeletedRecords[0].RecordID = " /private/record" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := snapshot.SourceFacts.Clone()
			mutate(&candidate)
			if candidate.Validate() == nil || validateSnapshotSourceFacts(candidate, snapshot.SnapshotContent) == nil {
				t.Fatal("unqualified deletion roster admitted")
			}
		})
	}
	active := fixtureStructuredRenderModel(t, []ExportField{{Path: "/hosts/active", SourceFamily: "host", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"display_name": "Active"}}})
	conflict := exportprovider.EmptySourceFacts()
	conflict.DeletedRecords = []exportprovider.DeletedRecordFact{{RecordID: "active", RecordType: "host"}}
	if validateSnapshotSourceFacts(conflict, active.SnapshotContent) == nil {
		t.Fatal("deleted identity also appeared in source content")
	}
	output, err := exportprovider.NewProviderOutput("links", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	output.SourceFacts = snapshot.SourceFacts.Clone()
	if output.Validate() == nil {
		t.Fatal("non-Records provider claimed deleted Core identity")
	}
}

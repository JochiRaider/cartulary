package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func testCompleteSnapshotValidation(t *testing.T) {
	t.Run("frozen section hierarchy", testSnapshotSectionStructure)
	t.Run("complete independent invariant sets", testSnapshotInvariantCollections)
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/incident/one", ContentClass: ContentClassSourceEvidence, Value: "private-one"},
		{Path: "/incident/two", ContentClass: ContentClassSourceEvidence, Value: "private-two"},
		{Path: "/incident/three", ContentClass: ContentClassSourceEvidence, Value: "private-three"},
	})
	model.Sections[0].SchemaID = "retired"
	model.Sections[0].Blocks[0].ContentClass = ""
	model.Sections[0].Blocks[1].Fields[0].SchemaID = "retired"
	model.Sections[0].Blocks[2].Fields[0].SupportRefs = []string{"missing-support"}
	model.Subjects = []TokenizableSubject{{StableSubjectRef: "foreign"}}

	payload, err := (&fakeReportingJobStore{}).ReleasePayloadForJob(context.Background(), jobs.Execution{})
	if err != nil {
		t.Fatal(err)
	}
	var first []byte
	for _, reverse := range []bool{false, true} {
		if reverse {
			slices.Reverse(model.Sections[0].Blocks)
			// This test varies diagnostic traversal without introducing new
			// structural identity/order defects into its five source faults.
			for i := range model.Sections[0].Blocks {
				model.Sections[0].Blocks[i].BlockOrdinal = i + 1
				model.Sections[0].Blocks[i].BlockID = fmt.Sprintf("sec-0001.b%04d", i+1)
			}
			// Re-author provenance with the fixture's changed addresses. A
			// frozen reader must reject such a change rather than repair it.
			owners := map[string]sourceOwner{}
			for _, ref := range model.SourceIndex {
				if ref.SourcePath != nil {
					owners[ref.SourceRefID] = sourceOwner{ref.SourceFamily, ref.SourceRecordID}
				}
			}
			if err := materializeSnapshotSourceIndex(&model.SnapshotContent, owners, nil); err != nil {
				t.Fatal(err)
			}
		}
		failure := renderFailureDetails(validateSnapshotContent(model.SnapshotContent))
		if failure.Stage != "export_model_materialization" || failure.issues == nil || failure.issues.count != 7 {
			t.Fatalf("independent model defects were not collected: %+v", failure)
		}
		raw, err := failedRenderSummary(payload.renderInput, failure)
		if err != nil || validateFailedRenderSummary(raw, payload.renderInput, failure.ReasonCode) != nil {
			t.Fatal("invalid retained model failure", err)
		}
		if bytes.Contains(raw, []byte("private-")) || bytes.Contains(raw, []byte("missing-support")) || bytes.Contains(raw, []byte("foreign")) {
			t.Fatal("source values or unadmitted identities escaped through diagnostics")
		}
		if reverse && !bytes.Equal(first, raw) {
			t.Fatal("independent validation traversal changed the retained result")
		}
		first = raw
		if fields, err := model.RedactionFields(); err == nil || fields != nil {
			t.Fatal("invalid model produced a partial render view")
		}
		if err := validateSnapshotModelIdentity(SnapshotModel{SourceFacts: fixtureSourceFactsForContent(model.SnapshotContent), SchemaID: SnapshotModelSchemaID, SnapshotModelID: model.SnapshotModelID, SnapshotContent: model.SnapshotContent}); err == nil {
			t.Fatal("snapshot admission bypassed content validation")
		}
	}
	// The real renderer path must reject the complete stage before any native
	// renderer receives content, even when multiple independent defects exist.
	// Supply a valid preceding template stage, rather than the worker mock's
	// deliberately minimal template, to isolate the source diagnostic census.
	payload.TemplateContract, _ = ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	payload.OutputKind = OutputKindSlidev
	payload.OutputOptions = fixtureOutputOptions(payload.OutputKind, payload.ReleaseScope)
	payload.ExportModel = model
	runtime := &limitObservingRuntime{maxBytes: -1}
	result, _, err := (&Renderer{runtime: runtime}).Render(context.Background(), payload.renderInput)
	if err == nil || runtime.maxBytes != -1 || result.OutputSHA256 != "" || len(result.RenderBundle.Files) != 0 {
		t.Fatal("invalid model reached renderer or published partial output", err)
	}
	if f := renderFailureDetails(err); f.issues == nil || f.issues.count != 7 {
		t.Fatal("render boundary discarded independent diagnostics")
	}

	t.Run("bounded complete field traversal", func(t *testing.T) {
		model := SnapshotContent{IncidentID: "incident", SourceIndex: []ReportingSourceRef{}}
		if err := materializeSnapshotSourceIndex(&model, nil, nil); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2100; i++ {
			model.Records = append(model.Records, ReportingRecordSummary{Fields: []ReportingField{{FieldKey: fmt.Sprint(i), Value: "private"}}})
		}
		failure := renderFailureDetails(validateSnapshotFieldValues(model))
		if failure.issues == nil || failure.issues.count != 2100 || len(failure.issues.retained) != 500 {
			t.Fatal("model validation stopped early or retained unbounded diagnostics")
		}
		raw, err := failedRenderSummary(payload.renderInput, failure)
		if err != nil || validateFailedRenderSummary(raw, payload.renderInput, failure.ReasonCode) != nil {
			t.Fatal("bounded model diagnostics did not round trip", err)
		}
		var summary reportingRenderValidationSummary
		if err := json.Unmarshal(raw, &summary); err != nil || summary.IssueCount != 2100 || summary.Issues[499].Severity != "warning" || strings.Contains(string(raw), "private") {
			t.Fatal("incorrect stage count, warning or disclosure", err)
		}
	})
}

func testSnapshotInvariantCollections(t *testing.T) {
	assertCount := func(err error, count int) {
		t.Helper()
		failure := renderFailureDetails(err)
		if failure.Stage != "export_model_materialization" || failure.issues == nil || failure.issues.count != count {
			t.Fatalf("expected %d independent defects, got %+v", count, failure)
		}
	}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/hosts/first", ContentClass: ContentClassSourceEvidence, Value: "original-one"},
		{Path: "/hosts/second", ContentClass: ContentClassSourceEvidence, Value: "original-two"},
	})
	model.Records[0].Fields[0].Value = "changed-one"
	model.Records[1].Fields[0].Value = "changed-two"
	assertCount(validateSnapshotFieldBindings(model.SnapshotContent), 2)
	// A duplicate source reference cannot choose one of two disagreeing
	// values for the remaining record comparisons.
	duplicate := model.Sections[0].Blocks[0]
	duplicate.Fields = append([]ReportingField{}, duplicate.Fields...)
	duplicate.Fields[0].Value = "third-value"
	model.Sections[0].Blocks = append(model.Sections[0].Blocks, duplicate)
	var first []byte
	for _, reverse := range []bool{false, true} {
		if reverse {
			slices.Reverse(model.Sections[0].Blocks)
		}
		err := validateSnapshotFieldBindings(model.SnapshotContent)
		assertCount(err, 2) // One ambiguous binding and the other record mismatch.
		failure := renderFailureDetails(err)
		raw, err := canonicalJSON(failure.issues.retained)
		if err != nil || (reverse && !bytes.Equal(raw, first)) {
			t.Fatal("duplicate source lookup depended on traversal order", err)
		}
		first = raw
	}
	model.Sections[0].Blocks[0].ContentClass = ""
	model.Sections[0].Blocks[1].ContentClass = ""
	model.Sections[0].Blocks[2].ContentClass = ""
	assertCount(validateSnapshotBlockClasses(model.SnapshotContent), 3)
	model.Sections[0].SupportRefs = []string{"missing-a", "missing-b"}
	assertCount(validateSnapshotSupportReferences(model.SnapshotContent), 2)
}

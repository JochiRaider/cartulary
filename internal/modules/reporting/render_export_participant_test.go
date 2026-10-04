package reporting

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRenderExportParticipantValidatesClosedOutputAndBounds(t *testing.T) {
	model := ExportModel{RenderIdentity: fixtureRenderIdentity(),
		SchemaID: ExportModelSchemaID,
		SnapshotContent: SnapshotContent{
			IncidentID:        "00000000-0000-0000-0000-000000000401",
			SnapshotID:        "00000000-0000-0000-0000-000000000402",
			DerivationVersion: DerivationVersion,
			Sections:          []ReportingSection{},
			Records:           []ReportingRecordSummary{},
			Relationships:     []ReportingRelationshipSummary{},
			TimelineEvents:    []ReportingTimelineEvent{},
			Subjects:          []TokenizableSubject{},
			Diagrams:          []ReportingDiagram{},
			Assets:            []ReportingAssetDeclaration{},
			SupportIndex:      []ReportingSupportRef{}},
	}
	model = bindFixtureRenderModel(model.SnapshotContent)
	modelJSON, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("canonical model: %v", err)
	}
	invocation := RenderExportInvocation{
		Context: RenderExportContext{
			SchemaID:                RenderExportContextSchemaID,
			Operation:               RenderExportOperationKind,
			ProfileID:               ProfileID,
			ContractMajor:           1,
			ClaimState:              "claimed",
			SnapshotRef:             "snapshot:" + model.SnapshotID,
			AuthorizationViewSHA256: strings.Repeat("a", 64),
			RedactionProfileSHA256:  strings.Repeat("b", 64),
			TimeoutSeconds:          30,
		},
		ImmutableModel:    model,
		ImmutableModelSHA: reportingObjectDigest(ExportModelSchemaID, modelJSON),
	}
	result, err := (BuiltInRenderExportParticipant{}).Emit(context.Background(), invocation)
	if err != nil {
		t.Fatalf("emit: %v", err)
	}
	admitted, digest, err := AdmitRenderExportResult(invocation, result)
	if err != nil {
		t.Fatalf("admit: %v", err)
	}
	if admitted.SnapshotID != model.SnapshotID || digest != invocation.ImmutableModelSHA {
		t.Fatalf("admitted output is not bound to immutable model: %#v %q", admitted, digest)
	}

	alteredModel := model
	alteredModel.SourceChangeSetHighWatermark = "other-source"
	altered := result
	altered.Output, _ = canonicalJSON(alteredModel)
	altered.OutputSHA256 = hashHex(altered.Output)
	altered.OutputByteSize = int64(len(altered.Output))
	altered.OutputRef = "snapshot:" + model.SnapshotID + "/reporting-export-model:" + altered.OutputSHA256
	if _, _, err := AdmitRenderExportResult(invocation, altered); !errors.Is(err, ErrRenderExportParticipant) {
		t.Fatal("participant rewrote frozen source", err)
	}
	malformed := result
	malformed.OutputSchema = "cartulary.reporting_export_model.v1"
	if _, _, err := AdmitRenderExportResult(invocation, malformed); !errors.Is(err, ErrRenderExportParticipant) {
		t.Fatalf("malformed schema err = %v", err)
	}
	oversized := result
	oversized.OutputByteSize = RenderExportMaxOutputBytes + 1
	if _, _, err := AdmitRenderExportResult(invocation, oversized); !errors.Is(err, ErrRenderExportParticipant) {
		t.Fatalf("oversized result err = %v", err)
	}
	tooManyItems := result
	tooManyItems.ItemCount = RenderExportMaxItems + 1
	if _, _, err := AdmitRenderExportResult(invocation, tooManyItems); !errors.Is(err, ErrRenderExportParticipant) {
		t.Fatalf("item bound err = %v", err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (BuiltInRenderExportParticipant{}).Emit(canceled, invocation); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled invocation err = %v", err)
	}
}

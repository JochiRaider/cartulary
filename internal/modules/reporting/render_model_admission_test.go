package reporting

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestFrozenRenderModelAdmission(t *testing.T) {
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
	if err := materializeSnapshotSourceIndex(&model.SnapshotContent, nil, nil); err != nil {
		t.Fatal(err)
	}
	model = bindFixtureRenderModel(model.SnapshotContent)
	if renderModelItemCount(model) != 1 {
		t.Fatal("source index omitted from model item budget")
	}
	modelJSON, err := canonicalJSON(model)
	if err != nil {
		t.Fatalf("canonical model: %v", err)
	}
	digest := reportingObjectDigest(ExportModelSchemaID, modelJSON)
	admitted, err := admitFrozenExportModel(context.Background(), model, digest)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := canonicalJSON(admitted)
	if reportingObjectDigest(ExportModelSchemaID, raw) != digest {
		t.Fatal("admission altered frozen identity")
	}
	for _, mutation := range []func(*ExportModel){
		func(m *ExportModel) { m.SourceChangeSetHighWatermark = "changed source" },
		func(m *ExportModel) { m.SchemaID = "cartulary.reporting_export_model.v1" },
		func(m *ExportModel) { m.DerivationVersion = "retired" },
	} {
		changed, err := copyReportingModel(model)
		if err != nil {
			t.Fatal(err)
		}
		mutation(&changed)
		if _, err := admitFrozenExportModel(context.Background(), changed, digest); err == nil {
			t.Fatal("changed frozen model admitted")
		}
	}
	t.Run("actual item ceiling", func(t *testing.T) {
		candidate := model
		candidate.Subjects = make([]TokenizableSubject, maxRenderModelItems)
		if renderModelItemCount(candidate) != maxRenderModelItems+1 {
			t.Fatal("source index omitted from total")
		}
		if _, err := admitFrozenExportModel(context.Background(), candidate, digest); err == nil || !strings.Contains(err.Error(), "item limit") {
			t.Fatal("oversize model item set admitted", err)
		}
	})
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := admitFrozenExportModel(canceled, model, digest); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost", err)
	}
}

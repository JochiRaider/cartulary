package projectioncontract

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	for _, source := range []SourceReader{nil, (*evidenceSourceStub)(nil)} {
		got, err := NewContribution(source)
		if err == nil || err.Error() != "evidence projection source is required" || got.Source() != nil {
			t.Fatalf("nil source admitted: %#v, %v", got, err)
		}
	}
	for _, source := range []SourceReader{&evidenceSourceStub{}, evidenceSourceStub{}} {
		got, err := NewContribution(source)
		if err != nil || got.Source() != source {
			t.Fatalf("valid source not retained: %#v, %v", got, err)
		}
	}
}

func TestEvidenceProjectionContractOwnsSemanticSurface(t *testing.T) {
	descriptor := Descriptor()
	wantDescriptor := providercontract.ProviderDescriptor{
		SchemaVersion: "projection_provider_descriptor.v3", Status: "active", ProviderID: "evidence",
		SourceOwnerModule: "evidence", ProjectionStorageOwnerModule: "projections",
		ViewSchemaIDs: []string{"cartulary.view.evidence.v1"}, SourceRecordTypes: []string{"evidence"},
		SourceAuthorityModules: []string{"evidence", "links", "records"}, ProjectionTableIDs: []string{"evidence_grid_projection"},
		Capabilities:   providercontract.ProviderCapabilities{Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true},
		RestoreRebuild: "required", FacadePackages: []string{"internal/modules/evidence/projectioncontract"}, RebuildAfter: []string{"artifact"},
		CharacterizationRefs: []string{"internal/modules/evidence/projection_collaboration_integration_test.go", "internal/modules/projections/internal/runtime/query_test.go"},
	}
	if !reflect.DeepEqual(descriptor, wantDescriptor) {
		t.Fatalf("unexpected Evidence descriptor: %#v", descriptor)
	}
	intent, err := SurfaceIntent()
	if err != nil {
		t.Fatalf("Evidence semantic intent: %v", err)
	}
	wantIntent := providercontract.SurfaceIntent{
		ViewSchemaID: "cartulary.view.evidence.v1",
		FieldKeys: []string{
			"evidence.blob_hash", "evidence.collector_party_id", "evidence.collector_party_text", "evidence.edited_at",
			"evidence.lifecycle_state", "evidence.linked_record_count", "evidence.owner_user_id", "evidence.received_at", "evidence.requested_at",
			"evidence.source_party_id", "evidence.source_party_text", "evidence.storage_ref", "evidence.title", "evidence.upload_state",
		},
	}
	if !reflect.DeepEqual(intent, wantIntent) {
		t.Fatalf("Evidence semantic intent drifted: %#v", intent)
	}
	contribution, err := NewContribution(evidenceSourceStub{})
	if err != nil {
		t.Fatal(err)
	}
	contract := contribution.ProjectionContribution()
	if contract.SourceOwnerModule() != "evidence" ||
		!reflect.DeepEqual(contract.Descriptors(), []providercontract.ProviderDescriptor{wantDescriptor}) ||
		!reflect.DeepEqual(contract.SurfaceIntents(), []providercontract.SurfaceIntent{wantIntent}) {
		t.Fatal("runtime Evidence contribution differs from its independent declaration expectation")
	}
}

func TestEvidenceProjectionIntentRejectsUnknownViewSchema(t *testing.T) {
	t.Parallel()
	if _, err := surfaceIntent("cartulary.view.unknown.v1"); err == nil {
		t.Fatal("unknown Evidence view schema produced a semantic intent")
	}
}

func TestEvidenceProjectionContributionDefensivelyCopiesFacts(t *testing.T) {
	contribution, err := NewContribution(evidenceSourceStub{})
	if err != nil {
		t.Fatalf("construct Evidence projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	descriptors[0].FacadePackages[0] = "mutated"
	intents := contribution.ProjectionContribution().SurfaceIntents()
	intents[0].FieldKeys[0] = "mutated"
	again := contribution.ProjectionContribution()
	if got := again.Descriptors()[0].FacadePackages[0]; got != "internal/modules/evidence/projectioncontract" {
		t.Fatalf("descriptor mutation escaped contribution: %q", got)
	}
	if got := again.SurfaceIntents()[0].FieldKeys[0]; got == "mutated" {
		t.Fatal("semantic intent mutation escaped contribution")
	}
	if contribution.Source() == nil {
		t.Fatal("runtime contribution has no typed source")
	}
}

type evidenceSourceStub struct{ SourceReader }

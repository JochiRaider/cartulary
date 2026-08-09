package workbookprojection

import "testing"

func TestEvidenceProjectionContractOwnsSemanticSurface(t *testing.T) {
	descriptor := Descriptor()
	if descriptor.ProviderID != "evidence" ||
		descriptor.SourceOwnerModule != "evidence" ||
		descriptor.ProjectionStorageOwnerModule != "projections" ||
		len(descriptor.ViewSchemaIDs) != 1 ||
		descriptor.ViewSchemaIDs[0] != evidenceViewSchemaID ||
		len(descriptor.FacadePackages) != 1 ||
		descriptor.FacadePackages[0] != "internal/modules/evidence/workbookprojection" {
		t.Fatalf("unexpected Evidence descriptor: %#v", descriptor)
	}
	intent := SurfaceIntent()
	if intent.ViewSchemaID != evidenceViewSchemaID || len(intent.FieldKeys) == 0 {
		t.Fatalf("incomplete Evidence semantic intent: %#v", intent)
	}
}

func TestEvidenceProjectionContributionDefensivelyCopiesFacts(t *testing.T) {
	contribution, err := NewRuntimeContribution(evidenceSourceStub{})
	if err != nil {
		t.Fatalf("construct Evidence projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	descriptors[0].FacadePackages[0] = "mutated"
	intents := contribution.ProjectionContribution().SurfaceIntents()
	intents[0].FieldKeys[0] = "mutated"
	again := contribution.ProjectionContribution()
	if got := again.Descriptors()[0].FacadePackages[0]; got != "internal/modules/evidence/workbookprojection" {
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

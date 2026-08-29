package projectioncontract

import "testing"

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Evidence projection contribution unexpectedly constructed")
	}
}

func TestEvidenceProjectionContractOwnsSemanticSurface(t *testing.T) {
	descriptor := Descriptor()
	if descriptor.ProviderID != "evidence" ||
		descriptor.SourceOwnerModule != "evidence" ||
		descriptor.ProjectionStorageOwnerModule != "projections" ||
		len(descriptor.ViewSchemaIDs) != 1 ||
		descriptor.ViewSchemaIDs[0] != evidenceViewSchemaID ||
		len(descriptor.FacadePackages) != 1 ||
		descriptor.FacadePackages[0] != "internal/modules/evidence/projectioncontract" {
		t.Fatalf("unexpected Evidence descriptor: %#v", descriptor)
	}
	intent, err := SurfaceIntent()
	if err != nil {
		t.Fatalf("Evidence semantic intent: %v", err)
	}
	if intent.ViewSchemaID != evidenceViewSchemaID || len(intent.FieldKeys) == 0 {
		t.Fatalf("incomplete Evidence semantic intent: %#v", intent)
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

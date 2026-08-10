package workbookprojection

import "testing"

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Artifact projection contribution unexpectedly constructed")
	}
}

func TestArtifactProjectionContractOwnsEightSemanticSurfaces(t *testing.T) {
	contribution, err := NewContribution(artifactSourceStub{})
	if err != nil {
		t.Fatalf("construct Artifact projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	if len(descriptors) != 1 {
		t.Fatalf("Artifact descriptors = %d, want 1", len(descriptors))
	}
	descriptor := descriptors[0]
	if descriptor.ProviderID != "artifact" ||
		descriptor.SourceOwnerModule != "artifacts" ||
		descriptor.ProjectionStorageOwnerModule != "projections" ||
		len(descriptor.ViewSchemaIDs) != 8 ||
		len(descriptor.FacadePackages) != 1 ||
		descriptor.FacadePackages[0] != "internal/modules/artifacts/workbookprojection" {
		t.Fatalf("unexpected Artifact descriptor: %#v", descriptor)
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 8 {
		t.Fatalf("Artifact semantic intents = %d, want 8", len(intents))
	}
	seen := make(map[string]struct{}, len(intents))
	for _, intent := range intents {
		if len(intent.FieldKeys) == 0 || intent.CanonicalSourceFilter == nil ||
			intent.CanonicalSourceFilter.Kind != "artifact_type" ||
			intent.CanonicalSourceFilter.Value == "" {
			t.Fatalf("incomplete Artifact semantic intent: %#v", intent)
		}
		if _, duplicate := seen[intent.ViewSchemaID]; duplicate {
			t.Fatalf("duplicate Artifact semantic intent %q", intent.ViewSchemaID)
		}
		seen[intent.ViewSchemaID] = struct{}{}
	}
}

func TestArtifactProjectionIntentRejectsUnknownViewSchema(t *testing.T) {
	t.Parallel()
	if _, err := surfaceIntent("cartulary.view.unknown.v1"); err == nil {
		t.Fatal("unknown Artifact view schema produced a semantic intent")
	}
}

func TestArtifactProjectionContributionDefensivelyCopiesFacts(t *testing.T) {
	contribution, err := NewContribution(artifactSourceStub{})
	if err != nil {
		t.Fatalf("construct Artifact projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	descriptors[0].FacadePackages[0] = "mutated"
	intents := contribution.ProjectionContribution().SurfaceIntents()
	intents[0].FieldKeys[0] = "mutated"
	again := contribution.ProjectionContribution()
	if got := again.Descriptors()[0].FacadePackages[0]; got != "internal/modules/artifacts/workbookprojection" {
		t.Fatalf("descriptor mutation escaped contribution: %q", got)
	}
	if got := again.SurfaceIntents()[0].FieldKeys[0]; got == "mutated" {
		t.Fatal("semantic intent mutation escaped contribution")
	}
	if contribution.Source() == nil {
		t.Fatal("runtime contribution has no typed source")
	}
}

type artifactSourceStub struct{ SourceReader }

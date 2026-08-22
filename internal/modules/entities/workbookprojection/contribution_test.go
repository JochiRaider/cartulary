package workbookprojection

import (
	"reflect"
	"testing"
)

type contributionSource struct{ SourceReader }

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Entities projection contribution unexpectedly constructed")
	}
}

func TestRuntimeContributionOwnsTypedEntityDescriptorFacts(t *testing.T) {
	contribution, err := NewContribution(&contributionSource{})
	if err != nil {
		t.Fatalf("construct Entities projection contribution: %v", err)
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	if len(descriptors) != 2 {
		t.Fatalf("descriptor count = %d, want 2", len(descriptors))
	}
	for index, want := range Descriptors() {
		got := descriptors[index]
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("descriptor %d drifted:\ngot  %#v\nwant %#v", index, got, want)
		}
		if !reflect.DeepEqual(got.FacadePackages, []string{"internal/modules/entities/workbookprojection"}) {
			t.Fatalf("provider %q facade packages = %#v", got.ProviderID, got.FacadePackages)
		}
		if got.ProviderID == "host" && !got.Capabilities.Query {
			t.Fatal("Host provider is not query-capable")
		}
		if got.ProviderID == "identity" && !got.Capabilities.Query {
			t.Fatal("Identity provider is not query-capable")
		}
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 2 ||
		!reflect.DeepEqual(intents[0], HostSurfaceIntent()) ||
		!reflect.DeepEqual(intents[1], IdentitySurfaceIntent()) {
		t.Fatalf("Entities surface intents drifted: %#v", intents)
	}
}

package projectioncontract

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
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
		if !reflect.DeepEqual(got.FacadePackages, []string{"internal/modules/entities/projectioncontract"}) {
			t.Fatalf("provider %q facade packages = %#v", got.ProviderID, got.FacadePackages)
		}
		if !got.Capabilities.Query {
			t.Fatalf("provider %q is not query-capable", got.ProviderID)
		}
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 2 ||
		!reflect.DeepEqual(intents[0], hostSurfaceIntent()) ||
		!reflect.DeepEqual(intents[1], identitySurfaceIntent()) {
		t.Fatalf("Entities surface intents drifted: %#v", intents)
	}

	assertInterfaceMethods(t, (*projectionports.MutationRows)(nil), []string{
		"DeleteHostTx",
		"DeleteIdentityTx",
		"RefreshHostTx",
		"RefreshIdentityTx",
	})
	assertInterfaceMethods(t, (*projectionports.QueryReader)(nil), []string{
		"SelectHostQueryProjections",
		"SelectIdentityQueryProjections",
	})
	assertInterfaceMethods(t, (*projectionports.ReportingReader)(nil), []string{
		"CollectHostDerivedFactsTx",
		"CollectIdentityDerivedFactsTx",
	})
}

func assertInterfaceMethods(t *testing.T, pointer any, want []string) {
	t.Helper()
	interfaceType := reflect.TypeOf(pointer).Elem()
	got := make([]string, interfaceType.NumMethod())
	for index := range interfaceType.NumMethod() {
		got[index] = interfaceType.Method(index).Name
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s methods = %v, want %v", interfaceType, got, want)
	}
}

package workbookprojection

import (
	"reflect"
	"testing"
)

func TestRuntimeContributionOwnsCompleteIndicatorProjectionContract(t *testing.T) {
	t.Parallel()
	contribution, err := NewRuntimeContribution(indicatorSourceStub{})
	if err != nil {
		t.Fatalf("construct Indicator projection contribution: %v", err)
	}
	if contribution.Source() == nil {
		t.Fatal("Indicator projection contribution has no source")
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	if len(descriptors) != 1 || !reflect.DeepEqual(descriptors[0], Descriptor()) {
		t.Fatalf("Indicator descriptors = %#v", descriptors)
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 1 || !reflect.DeepEqual(intents[0], SurfaceIntent()) {
		t.Fatalf("Indicator semantic intents = %#v", intents)
	}
}

type indicatorSourceStub struct {
	SourceReader
}

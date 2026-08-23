package workbookprojection

import (
	"reflect"
	"testing"
)

func TestNewContributionRequiresSource(t *testing.T) {
	t.Parallel()
	if _, err := NewContribution(nil); err == nil {
		t.Fatal("source-less Indicator projection contribution unexpectedly constructed")
	}
}

func TestRuntimeContributionOwnsCompleteIndicatorProjectionContract(t *testing.T) {
	t.Parallel()
	contribution, err := NewContribution(indicatorSourceStub{})
	if err != nil {
		t.Fatalf("construct Indicator projection contribution: %v", err)
	}
	if contribution.Source() == nil {
		t.Fatal("Indicator projection contribution has no source")
	}
	descriptors := contribution.ProjectionContribution().Descriptors()
	if len(descriptors) != 1 || !reflect.DeepEqual(descriptors[0], descriptor()) {
		t.Fatalf("Indicator descriptors = %#v", descriptors)
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 1 || !reflect.DeepEqual(intents[0], surfaceIntent()) {
		t.Fatalf("Indicator semantic intents = %#v", intents)
	}
}

type indicatorSourceStub struct {
	SourceReader
}

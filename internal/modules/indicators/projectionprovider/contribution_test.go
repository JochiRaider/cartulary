package projectionprovider

import (
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestContributionOwnsCompleteIndicatorSurfaceIntent(t *testing.T) {
	t.Parallel()
	contribution, err := NewContribution()
	if err != nil {
		t.Fatalf("construct Indicator projection contribution: %v", err)
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 1 || intents[0].ViewSchemaID != indicators.ViewSchemaID {
		t.Fatalf("Indicator projection intents = %#v", intents)
	}
	schema, ok := viewschema.Lookup(indicators.ViewSchemaID)
	if !ok {
		t.Fatalf("missing Indicator view schema %s", indicators.ViewSchemaID)
	}
	got := append([]string(nil), intents[0].FieldKeys...)
	slices.Sort(got)
	want := make([]string, 0, len(schema.Fields()))
	for key := range schema.Fields() {
		want = append(want, key)
	}
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Indicator projection fields drifted from schema registry:\ngot  %#v\nwant %#v", got, want)
	}
}

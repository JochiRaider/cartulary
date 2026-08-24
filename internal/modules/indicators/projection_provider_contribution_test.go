package indicators

import (
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestContributionOwnsCompleteIndicatorSurfaceIntent(t *testing.T) {
	t.Parallel()
	contribution, err := NewProjectionContribution()
	if err != nil {
		t.Fatalf("construct Indicator projection contribution: %v", err)
	}
	intents := contribution.ProjectionContribution().SurfaceIntents()
	if len(intents) != 1 || intents[0].ViewSchemaID != ViewSchemaID {
		t.Fatalf("Indicator projection intents = %#v", intents)
	}
	schema, ok := viewschema.Lookup(ViewSchemaID)
	if !ok {
		t.Fatalf("missing Indicator view schema %s", ViewSchemaID)
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

func TestIndicatorImportContributionOwnsStableIdentity(t *testing.T) {
	t.Parallel()
	contribution, err := NewImportContribution(&Application{})
	if err != nil {
		t.Fatalf("construct Indicator Import contribution: %v", err)
	}
	binding := contribution.ImportOwnerCreateBinding()
	if binding.TargetViewSchemaID != ViewSchemaID || binding.FacadeID != "indicators.import_create" {
		t.Fatalf("Indicator Import binding = %#v", binding)
	}
	if contribution, err := NewImportContribution(nil); err == nil || contribution != nil {
		t.Fatalf("nil application contribution = %#v, error=%v", contribution, err)
	}
}

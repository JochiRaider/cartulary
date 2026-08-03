package indicators

import (
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestProjectionContributionOwnsCompleteIndicatorQuerySurface(t *testing.T) {
	t.Parallel()
	contribution := NewProjectionContribution()
	if contribution.Source() == nil {
		t.Fatal("Indicator projection contribution has no source")
	}
	surfaces := contribution.QuerySurfaces()
	if len(surfaces) != 1 || surfaces[0].ViewSchemaID != ViewSchemaID {
		t.Fatalf("Indicator projection surfaces = %#v", surfaces)
	}
	schema, ok := viewschema.Lookup(ViewSchemaID)
	if !ok {
		t.Fatalf("missing Indicator view schema %s", ViewSchemaID)
	}
	got := make([]string, 0, len(surfaces[0].Fields))
	for _, field := range surfaces[0].Fields {
		got = append(got, field.Key)
	}
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

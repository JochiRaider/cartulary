package reporting

import (
	"reflect"
	"sort"
	"testing"
)

func TestSupportedOutputKindsAreClosed_Unit(t *testing.T) {
	got := supportedOutputKinds()
	sort.Strings(got)
	want := []string{OutputKindMermaid, OutputKindSlidev}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("supported output kinds = %v, want %v", got, want)
	}
}

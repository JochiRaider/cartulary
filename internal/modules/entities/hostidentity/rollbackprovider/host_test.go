package rollbackprovider

import (
	"reflect"
	"testing"
)

func TestHostSourceForRollbackValue(t *testing.T) {
	t.Parallel()
	value := map[string]any{"source": map[string]any{
		"display_name": "Host A",
		"hostname":     nil,
		"host_state":   "canonical",
	}}
	got, ok := hostSourceForRollbackValue(value)
	if !ok {
		t.Fatal("hostSourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"display_name": "Host A", "hostname": nil, "host_state": "canonical"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("host source = %#v, want %#v", got, want)
	}
	if _, ok := hostSourceForRollbackValue(map[string]any{"record_id": "legacy", "display_name": "Legacy"}); ok {
		t.Fatal("schema-less direct row was accepted")
	}
}

package rollbackprovider

import (
	"reflect"
	"testing"
)

func TestHostSourceForRollbackValue(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"host.display_name": map[string]any{"value": "Host A"},
		"host.hostname":     map[string]any{"value": nil},
		"party.notes":       map[string]any{"value": "unrelated"},
	}}
	got, ok := hostSourceForRollbackValue(value)
	if !ok {
		t.Fatal("hostSourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"display_name": "Host A", "hostname": nil, "host_state": "canonical"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("host source = %#v, want %#v", got, want)
	}
}

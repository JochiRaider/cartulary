package rollback

import (
	"reflect"
	"testing"
)

func TestSourceForRollbackValueExcludesDerivedChildEffects(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"indicator.display_value":     map[string]any{"value": "192.0.2.10"},
		"indicator.hash_algorithm":    map[string]any{"value": nil},
		"indicator.observation_count": map[string]any{"value": 3},
	}}
	got, ok := sourceForRollbackValue(value)
	if !ok {
		t.Fatal("sourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"display_value": "192.0.2.10", "hash_algorithm": nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("indicator source = %#v, want %#v", got, want)
	}
}

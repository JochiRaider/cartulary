package rollbackprovider

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

func TestBuildDedupeKeyIsStable(t *testing.T) {
	t.Parallel()
	state := rowState{indicatorType: "ipv4_addr", valueKind: "atomic", displayValue: "192.0.2.10"}
	if got, want := buildDedupeKey(state), "1144f8a3e5afba287cacdf6c895c89cc67c4769522a28f1cadd8f7e2959f5ca2"; got != want {
		t.Fatalf("dedupe key = %q, want %q", got, want)
	}
}

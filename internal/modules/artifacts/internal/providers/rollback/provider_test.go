package rollback

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/sourcecontract"
)

func TestSourceForRollbackValueMapsAllArtifactVariantsWithoutCollections(t *testing.T) {
	t.Parallel()
	value := map[string]any{"source": map[string]any{
		"body": "note", "summary": "brief", "next_checks": nil, "active_risks_summary": "risk",
		"closure_state": "closed", "kind": "hypothesis", "platform": "KQL", "case_sensitive": true,
	}}
	got, ok := sourcecontract.ExtractRollbackSource(value)
	if !ok {
		t.Fatal("ExtractRollbackSource returned ok=false")
	}
	want := map[string]any{
		"body": "note", "summary": "brief", "next_checks": nil, "active_risks_summary": "risk",
		"closure_state": "closed", "kind": "hypothesis", "platform": "KQL", "case_sensitive": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("artifact source = %#v, want %#v", got, want)
	}
	if _, ok := sourcecontract.ExtractRollbackSource(map[string]any{"cells": map[string]any{"note.body": map[string]any{"value": "legacy"}}}); ok {
		t.Fatal("schema-less projection row was accepted")
	}
}

func TestValidSourceRejectsSubtypeInvariantViolations(t *testing.T) {
	t.Parallel()
	for _, source := range []map[string]any{
		{"comm_type": "invalid"},
		{"confidence_score": float64(101)},
		{"match_mode": "glob"},
	} {
		if sourcecontract.ValidRollbackSource(source) {
			t.Fatalf("ValidRollbackSource(%#v) = true", source)
		}
	}
}

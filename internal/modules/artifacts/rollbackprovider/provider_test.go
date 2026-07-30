package rollbackprovider

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/sourcecontract"
)

func TestSourceForRollbackValueMapsAllArtifactVariantsWithoutCollections(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"note.body":                          map[string]any{"value": "note"},
		"comm_log.summary":                   map[string]any{"value": "brief"},
		"handoff.next_checks":                map[string]any{"value": nil},
		"status_review.active_risks_summary": map[string]any{"value": "risk"},
		"lesson.closure_state":               map[string]any{"value": "closed"},
		"finding.kind":                       map[string]any{"value": "hypothesis"},
		"investigative_query.platform":       map[string]any{"value": "KQL"},
		"forensic_keyword.case_sensitive":    map[string]any{"value": true},
		"handoff.open_risk_refs":             map[string]any{"value": []any{"ignored"}},
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

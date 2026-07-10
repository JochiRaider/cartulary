package rollbackprovider

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestSourceForRollbackValueExcludesDerivedConfidenceBand(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"assessment.confidence_score": map[string]any{"value": nil},
		"assessment.rationale":        map[string]any{"value": "Retained"},
		"assessment.confidence_band":  map[string]any{"value": "unset"},
	}}
	got, ok := sourceForRollbackValue(value)
	if !ok {
		t.Fatal("sourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"confidence_score": nil, "rationale": "Retained"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("assessment source = %#v, want %#v", got, want)
	}
}

func TestProviderRejectsInvalidOwnerInvariants(t *testing.T) {
	t.Parallel()
	provider := NewProvider()
	for _, value := range []map[string]any{
		{"source": map[string]any{"subject_type": "artifact"}},
		{"source": map[string]any{"subject_record_id": nil}},
		{"source": map[string]any{"confidence_score": float64(101)}},
	} {
		if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("ValidateRollbackValue error = %v, want target-not-reversible", err)
		}
	}
}

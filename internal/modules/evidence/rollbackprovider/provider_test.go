package rollbackprovider

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestSourceForRollbackValuePreservesPresenceAndAssociations(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"evidence.title":               map[string]any{"value": "Memory image"},
		"evidence.source_party_id":     map[string]any{"value": nil},
		"evidence.linked_record_count": map[string]any{"value": 9},
	}}
	got, ok := sourceForRollbackValue(value)
	if !ok {
		t.Fatal("sourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"title": "Memory image", "source_party_id": nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("evidence source = %#v, want %#v", got, want)
	}
	if _, present := got["object_blob_id"]; present {
		t.Fatal("absent blob identity unexpectedly became present")
	}
}

func TestProviderRejectsInvalidLifecycleAndReferences(t *testing.T) {
	t.Parallel()
	provider := NewProvider()
	for _, value := range []map[string]any{
		{"source": map[string]any{"lifecycle_state": "invalid"}},
		{"source": map[string]any{"collector_party_id": "not-a-uuid"}},
	} {
		if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("ValidateRollbackValue error = %v, want target-not-reversible", err)
		}
	}
}

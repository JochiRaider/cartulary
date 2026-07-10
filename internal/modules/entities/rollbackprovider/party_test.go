package rollbackprovider

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestPartySourceForRollbackValuePreservesPresence(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"party.display_name":  map[string]any{"value": "Incident Lead"},
		"party.primary_email": map[string]any{"value": nil},
		"host.hostname":       map[string]any{"value": "unrelated"},
	}}
	got, ok := partySourceForRollbackValue(value)
	if !ok {
		t.Fatal("partySourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"display_name": "Incident Lead", "primary_email": nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("party source = %#v, want %#v", got, want)
	}
	if _, present := got["notes"]; present {
		t.Fatal("absent notes unexpectedly became present")
	}
}

func TestPartyProviderRejectsInvalidRequiredValues(t *testing.T) {
	t.Parallel()
	provider := NewPartyProvider()
	for _, value := range []map[string]any{
		{"source": map[string]any{"display_name": nil}},
		{"source": map[string]any{"party_kind": "invalid"}},
	} {
		if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("ValidateRollbackValue error = %v, want target-not-reversible", err)
		}
	}
}

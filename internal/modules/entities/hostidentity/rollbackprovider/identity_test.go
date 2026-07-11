package rollbackprovider

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestIdentitySourceForRollbackValuePreservesNullableIdentifiers(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"identity.display_name":    map[string]any{"value": "Operator"},
		"identity.aad_object_id":   map[string]any{"value": nil},
		"identity.privilege_level": map[string]any{"value": "admin"},
		"host.hostname":            map[string]any{"value": "unrelated"},
	}}
	got, ok := identitySourceForRollbackValue(value)
	if !ok {
		t.Fatal("identitySourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"display_name": "Operator", "aad_object_id": nil, "privilege_level": "admin"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("identity source = %#v, want %#v", got, want)
	}
	if _, present := got["identity_state"]; present {
		t.Fatal("absent identity state unexpectedly defaulted")
	}
}

func TestIdentityProviderRejectsMalformedStateAndMergeID(t *testing.T) {
	t.Parallel()
	provider := NewIdentityProvider()
	for _, value := range []map[string]any{
		{"source": map[string]any{"identity_state": "invalid"}},
		{"source": map[string]any{"merged_into_record_id": "not-a-uuid"}},
	} {
		if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("ValidateRollbackValue error = %v, want target-not-reversible", err)
		}
	}
}

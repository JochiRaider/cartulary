package rollback

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestPartySourceForRollbackValuePreservesPresence(t *testing.T) {
	t.Parallel()
	value := map[string]any{"source": map[string]any{
		"display_name":  "Incident Lead",
		"primary_email": nil,
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
	if _, ok := partySourceForRollbackValue(map[string]any{"record_id": "legacy", "display_name": "Legacy"}); ok {
		t.Fatal("schema-less direct row was accepted")
	}
}

func TestPartyProviderRejectsInvalidRequiredValues(t *testing.T) {
	t.Parallel()
	provider := NewProvider()
	for _, value := range []map[string]any{
		{"source": map[string]any{"display_name": nil}},
		{"source": map[string]any{"party_kind": "invalid"}},
	} {
		if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("ValidateRollbackValue error = %v, want target-not-reversible", err)
		}
	}
}

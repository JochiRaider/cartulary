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

func TestPartyRollbackEveryPresentFieldUsesGeneratedPolicy(t *testing.T) {
	t.Parallel()
	valid := map[string]any{
		"display_name":      "Incident Lead",
		"party_kind":        "person",
		"organization_name": "Example Org",
		"role_title":        "Lead",
		"primary_email":     "lead@example.test",
		"timezone_name":     "UTC",
		"external_ref":      "Directory/Lead",
		"notes":             "first\nsecond",
	}
	provider := NewProvider()
	if err := provider.ValidateRollbackValue(map[string]any{"source": valid}); err != nil {
		t.Fatalf("valid full Party rollback value: %v", err)
	}
	for field := range valid {
		t.Run(field+"_wrong_type", func(t *testing.T) {
			if err := provider.ValidateRollbackValue(map[string]any{
				"source": map[string]any{field: 42},
			}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
				t.Fatalf("wrong-type %s error = %v", field, err)
			}
		})
	}
	for _, optional := range []string{
		"organization_name", "role_title", "primary_email", "timezone_name", "external_ref", "notes",
	} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": map[string]any{optional: nil}}); err != nil {
			t.Fatalf("nullable %s rollback value: %v", optional, err)
		}
	}
	for _, invalid := range []map[string]any{
		{"display_name": " Incident Lead "},
		{"party_kind": "invalid"},
		{"primary_email": "not an email"},
		{"timezone_name": "Etc/Not_A_Zone"},
		{"notes": "bad\x00control"},
	} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": invalid}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("invalid policy value %#v error = %v", invalid, err)
		}
	}
}

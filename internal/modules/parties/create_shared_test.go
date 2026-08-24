package parties

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

func TestPartyCreateAdmissionDefaultsAreShared_Unit(t *testing.T) {
	displayName := "Acme"
	partyKind := "organization"
	values, admissionErr := admitCreateValues(map[string]createValueInput{
		"party.display_name": {present: true, text: &displayName},
		"party.party_kind":   {present: true, text: &partyKind},
	})
	if admissionErr != nil {
		t.Fatalf("admit shared Party create values: %#v", admissionErr)
	}
	if len(values) != 8 {
		t.Fatalf("admitted Party value count = %d, want 8", len(values))
	}
	for _, optional := range []string{
		"party.organization_name", "party.role_title", "party.primary_email",
		"party.timezone_name", "party.external_ref", "party.notes",
	} {
		if _, present := values[optional].StoredValue(); present {
			t.Fatalf("omitted optional field %s did not default to null", optional)
		}
	}
	if _, admissionErr := admitCreateValues(map[string]createValueInput{
		"party.party_kind": {present: true, text: &partyKind},
	}); admissionErr == nil || admissionErr.field != "party.display_name" || admissionErr.reasonCode != "missing_required_field" {
		t.Fatalf("missing required field admission = %#v", admissionErr)
	}
}

func TestPartyImportCreateScalarAdmissionMatrix_Unit(t *testing.T) {
	displayName := "Acme"
	partyKind := "organization"
	values, err := valuesFromImport(map[string]ownerfacade.ImportScalarValue{
		"party.display_name": {Kind: "text", Text: &displayName},
		"party.party_kind":   {Kind: "text", Text: &partyKind},
		"party.notes":        {Kind: "null"},
	})
	if err != nil {
		t.Fatalf("admit text/null Party import values: %v", err)
	}
	if len(values) != 8 {
		t.Fatalf("Party import default count = %d, want 8", len(values))
	}
	if _, present := values["party.notes"].StoredValue(); present {
		t.Fatal("explicit Party import null was not preserved")
	}

	timestamp := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	id := uuid.New()
	number := int64(1)
	boolean := true
	for name, malformed := range map[string]ownerfacade.ImportScalarValue{
		"timestamp":         {Kind: "timestamp", Timestamp: &timestamp},
		"uuid":              {Kind: "uuid", UUID: &id},
		"number":            {Kind: "number", Number: &number},
		"bool":              {Kind: "bool", Bool: &boolean},
		"collection":        {Kind: "collection", CollectionToken: &ownerfacade.ImportCollectionToken{}},
		"unknown":           {Kind: "future_scalar"},
		"text_without_text": {Kind: "text"},
		"null_with_text":    {Kind: "null", Text: &displayName},
		"mixed_variant":     {Kind: "text", Text: &displayName, Bool: &boolean},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := valuesFromImport(map[string]ownerfacade.ImportScalarValue{
				"party.display_name": malformed,
				"party.party_kind":   {Kind: "text", Text: &partyKind},
			})
			detail, ok := ownerfacade.ImportOwnerCreateErrorDetail(err)
			if !ok {
				t.Fatalf("malformed Party import scalar error = %v", err)
			}
			safe := detail["safe_details"].(map[string]any)
			if safe["reason_code"] != "invalid_text" || safe["field"] != "party.display_name" {
				t.Fatalf("malformed Party import scalar detail = %#v", detail)
			}
		})
	}
}

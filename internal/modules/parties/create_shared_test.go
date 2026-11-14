package parties

import (
	"encoding/json"
	"maps"
	"strings"
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
	t.Run("historical stored scalar policy", func(t *testing.T) {
		validate := NewRevisionContribution().Records[0].ValidateSnapshotSource
		record := map[string]any{"record_id": "11111111-1111-4111-8111-111111111111", "incident_id": "22222222-2222-4222-8222-222222222222"}
		base := map[string]any{
			"record_id": record["record_id"], "incident_id": record["incident_id"],
			"display_name": "Café Team", "party_kind": "team", "organization_name": "Acme", "role_title": "Responder",
			"primary_email": "Case.Sensitive@Example.test", "timezone_name": "Etc/UTC", "external_ref": "case/123", "notes": "Line one\n\tLine two",
		}
		for _, kind := range []string{"person", "team", "organization", "distribution_list", "other"} {
			value := maps.Clone(base)
			value["party_kind"] = kind
			before, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if err := validate(record, value); err != nil {
				t.Fatalf("valid %s snapshot: %v", kind, err)
			}
			after, err := json.Marshal(value)
			if err != nil || string(after) != string(before) {
				t.Fatalf("valid stored values changed: %s / %s (%v)", before, after, err)
			}
		}
		for _, test := range []struct{ name, field, value string }{
			{"empty required", "display_name", ""},
			{"display spaces", "display_name", "Café  Team"},
			{"display NFC", "display_name", "Cafe\u0301 Team"},
			{"unknown kind", "party_kind", "Team"},
			{"organization trim", "organization_name", " Acme "},
			{"role control", "role_title", "Role\x00"},
			{"email syntax", "primary_email", "invalid"},
			{"timezone registry", "timezone_name", "utc"},
			{"external empty", "external_ref", ""},
			{"notes newline", "notes", "Line one\r\nLine two"},
			{"name bound", "display_name", strings.Repeat("a", 4096)},
		} {
			t.Run(test.name, func(t *testing.T) {
				value := maps.Clone(base)
				value[test.field] = test.value
				before, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if err := validate(record, value); err == nil {
					t.Fatal("invalid retained scalar accepted")
				}
				after, err := json.Marshal(value)
				if err != nil || string(after) != string(before) {
					t.Fatal("invalid retained scalar was repaired")
				}
			})
		}
		optionalNulls := maps.Clone(base)
		for _, field := range []string{"organization_name", "role_title", "primary_email", "timezone_name", "external_ref", "notes"} {
			optionalNulls[field] = nil
		}
		if err := validate(record, optionalNulls); err != nil {
			t.Fatalf("explicit optional nulls rejected: %v", err)
		}
	})
	displayName := "Acme"
	partyKind := "organization"
	values, err := valuesFromImport(map[string]ownerfacade.ImportScalarValue{
		"party.display_name": ownerfacade.NewTextImportScalar(displayName),
		"party.party_kind":   ownerfacade.NewTextImportScalar(partyKind),
		"party.notes":        ownerfacade.NewNullImportScalar(),
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
	for name, malformed := range map[string]ownerfacade.ImportScalarValue{
		"zero":       {},
		"timestamp":  ownerfacade.NewTimestampImportScalar(timestamp),
		"uuid":       ownerfacade.NewUUIDImportScalar(id),
		"number":     ownerfacade.NewNumberImportScalar(1),
		"bool":       ownerfacade.NewBoolImportScalar(true),
		"collection": ownerfacade.NewCollectionTokenImportScalar(ownerfacade.ImportCollectionToken{}),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := valuesFromImport(map[string]ownerfacade.ImportScalarValue{
				"party.display_name": malformed,
				"party.party_kind":   ownerfacade.NewTextImportScalar(partyKind),
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

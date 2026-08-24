package parties_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/parties"
)

func TestPartyMutationAdmissionPublicIdentity_Unit(t *testing.T) {
	create, apiErr := parties.AdmitCreateJSON(strings.NewReader(
		`{"client_txn_id":"txn-create","party.party_kind":"organization","party.display_name":" Acme "}`,
	))
	if apiErr != nil {
		t.Fatalf("admit Party create request: %#v", apiErr)
	}
	if create.ClientTransactionID() != "txn-create" {
		t.Fatalf("unexpected create identity")
	}

	patch, apiErr := parties.AdmitPatchJSON(strings.NewReader(
		`{"view_schema_id":"cartulary.view.parties.v1","base_row_version":4,"client_txn_id":"txn-patch","changes":[{"field_key":"party.display_name","value":" Acme "}]}`,
	))
	if apiErr != nil {
		t.Fatalf("admit Party patch request: %#v", apiErr)
	}
	if patch.AdmittedBaseRowVersion() != 4 || patch.ClientTransactionID() != "txn-patch" {
		t.Fatalf("unexpected patch identity")
	}

	claims := parties.ConflictClaims{
		RecordID:     uuid.MustParse("11111111-2222-3333-4444-555555555555"),
		ViewSchemaID: parties.ViewSchemaID, FieldKey: "party.display_name", CurrentRowVersion: 7,
	}
	conflict, apiErr := parties.AdmitConflictResolveJSON(
		strings.NewReader(`{"conflict_token":"opaque","resolution_kind":"merged_value","client_txn_id":"txn-conflict","resolved_value":" Merged "}`),
		"opaque",
		claims,
	)
	if apiErr != nil {
		t.Fatalf("admit Party conflict request: %#v", apiErr)
	}
	if conflict.ClientTransactionID() != "txn-conflict" {
		t.Fatalf("unexpected conflict identity")
	}
}

func TestPartyMutationFieldBoundariesAndClears_Unit(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		fieldKey  string
		max       int
		character string
	}{
		{name: "display", fieldKey: "party.display_name", max: 256, character: "d"},
		{name: "organization", fieldKey: "party.organization_name", max: 256, character: "o"},
		{name: "role", fieldKey: "party.role_title", max: 256, character: "r"},
		{name: "email", fieldKey: "party.primary_email", max: 320, character: "e"},
		{name: "external", fieldKey: "party.external_ref", max: 1024, character: "x"},
		{name: "notes", fieldKey: "party.notes", max: 16384, character: "n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			maxValue := strings.Repeat(testCase.character, testCase.max)
			if testCase.fieldKey == "party.primary_email" {
				maxValue = strings.Repeat("e", testCase.max-2) + "@x"
			}
			assertPartyPatchAdmission(t, testCase.fieldKey, maxValue, true)
			assertPartyPatchAdmission(t, testCase.fieldKey, maxValue+testCase.character, false)
		})
	}
	for _, optional := range []string{
		"party.organization_name", "party.role_title", "party.primary_email",
		"party.timezone_name", "party.external_ref", "party.notes",
	} {
		assertPartyPatchAdmission(t, optional, nil, true)
		assertPartyPatchAdmission(t, optional, " \t ", true)
	}
	assertPartyPatchAdmission(t, "party.display_name", nil, false)
	assertPartyPatchAdmission(t, "party.party_kind", "Person", false)
	assertPartyPatchAdmission(t, "party.party_kind", " person ", false)
	assertPartyPatchAdmission(t, "party.primary_email", "a@@b", false)
	assertPartyPatchAdmission(t, "party.primary_email", "a b@example.test", false)
	assertPartyPatchAdmission(t, "party.timezone_name", "America/New_York", true)
	assertPartyPatchAdmission(t, "party.timezone_name", "US/Eastern", true)
	assertPartyPatchAdmission(t, "party.timezone_name", "america/new_york", false)
	assertPartyPatchAdmission(t, "party.timezone_name", "EST", false)
	assertPartyPatchAdmission(t, "party.timezone_name", "+05:00", false)
	assertPartyPatchAdmission(t, "party.timezone_name", "../America/New_York", false)
	assertPartyPatchAdmission(t, "party.notes", "allowed\tline\nnext", true)
	assertPartyPatchAdmission(t, "party.notes", "forbidden\x00control", false)
}

func TestPartyMutationRejectsUnknownAndDuplicateMembers_Unit(t *testing.T) {
	for name, reject := range map[string]func() bool{
		"unknown create field": func() bool {
			_, apiErr := parties.AdmitCreateJSON(strings.NewReader(`{"client_txn_id":"txn","party.display_name":"Acme","party.party_kind":"organization","legacy":true}`))
			return apiErr != nil
		},
		"wrong patch surface": func() bool {
			_, apiErr := parties.AdmitPatchJSON(strings.NewReader(`{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"party.display_name","value":"Acme"}]}`))
			return apiErr != nil
		},
		"duplicate patch field": func() bool {
			_, apiErr := parties.AdmitPatchJSON(strings.NewReader(`{"view_schema_id":"cartulary.view.parties.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"party.display_name","value":"A"},{"field_key":"party.display_name","value":"B"}]}`))
			return apiErr != nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			if !reject() {
				t.Fatal("expected malformed Party mutation rejection")
			}
		})
	}
}

func assertPartyPatchAdmission(t *testing.T, fieldKey string, value any, accepted bool) {
	t.Helper()
	payload := fmt.Sprintf(
		`{"view_schema_id":"cartulary.view.parties.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":%q,"value":%s}]}`,
		fieldKey,
		jsonLiteral(value),
	)
	_, apiErr := parties.AdmitPatchJSON(strings.NewReader(payload))
	if (apiErr == nil) != accepted {
		t.Fatalf("admission for %s value %#v accepted=%t, error=%#v", fieldKey, value, accepted, apiErr)
	}
}

func jsonLiteral(value any) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%q", value)
}

package parties_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/parties"
)

func TestPartyMutationAdmissionAndReplayHashing_Unit(t *testing.T) {
	create, apiErr := parties.AdmitCreateJSON(strings.NewReader(
		`{"client_txn_id":"txn-create","party.party_kind":"organization","party.display_name":" Acme "}`,
	))
	if apiErr != nil {
		t.Fatalf("admit Party create request: %#v", apiErr)
	}
	if create.AdmittedViewSchemaID() != parties.ViewSchemaID || create.ClientTransactionID() != "txn-create" {
		t.Fatalf("unexpected create identity")
	}
	if got := hex.EncodeToString(create.RequestHash()); got != "62ec84f72b5d719aa98c540998dc2c967b22cb6947ef928db85ce171d5ee86a1" {
		t.Fatalf("create request hash = %s", got)
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
	if got := hex.EncodeToString(patch.RequestHash()); got != "b8838d0c04b72734af771f7e38abca9e94f4f8b4b6ac1e8235204ad9e814139e" {
		t.Fatalf("patch request hash = %s", got)
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
	if conflict.ResolutionKind() != "merged_value" || conflict.ClientTransactionID() != "txn-conflict" {
		t.Fatalf("unexpected conflict identity")
	}
	if got := hex.EncodeToString(conflict.RequestHash()); got != "2b424f63a70372b4f5dcd4209e9948a4c9000a9d8e20c426148461554798b885" {
		t.Fatalf("conflict request hash = %s", got)
	}
}

func TestPartyMutationHashNormalizationEquivalence_Unit(t *testing.T) {
	createHash := func(t *testing.T, email, externalRef, notes string) []byte {
		t.Helper()
		admission, apiErr := parties.AdmitCreateJSON(strings.NewReader(fmt.Sprintf(
			`{"client_txn_id":"ignored","party.display_name":"José","party.party_kind":"person","party.primary_email":%q,"party.external_ref":%q,"party.notes":%q}`,
			email,
			externalRef,
			notes,
		)))
		if apiErr != nil {
			t.Fatalf("admit hash-equivalence create: %#v", apiErr)
		}
		return admission.RequestHash()
	}
	left := createHash(t, "Analyst@Example.COM", "Directory/AbC", "first\r\nsecond")
	right := createHash(t, "analyst@example.com", "Directory/AbC", "first\nsecond")
	if !bytes.Equal(left, right) {
		t.Fatalf("equivalent NFC/email/line-ending requests hash differently: %x != %x", left, right)
	}
	caseVariant := createHash(t, "analyst@example.com", "Directory/abc", "first\nsecond")
	if bytes.Equal(left, caseVariant) {
		t.Fatalf("external-reference case variants hash equally: %x", left)
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

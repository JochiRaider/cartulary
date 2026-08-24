package parties

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPartyMutationAdmissionAndReplayHashing_Unit(t *testing.T) {
	create, admissionErr := AdmitCreateJSON(strings.NewReader(
		`{"client_txn_id":"txn-create","party.party_kind":"organization","party.display_name":" Acme "}`,
	))
	if admissionErr != nil {
		t.Fatalf("admit Party create request: %#v", admissionErr)
	}
	if got := hex.EncodeToString(create.requestHash[:]); got != "62ec84f72b5d719aa98c540998dc2c967b22cb6947ef928db85ce171d5ee86a1" {
		t.Fatalf("create request hash = %s", got)
	}

	patch, admissionErr := AdmitPatchJSON(strings.NewReader(
		`{"view_schema_id":"cartulary.view.parties.v1","base_row_version":4,"client_txn_id":"txn-patch","changes":[{"field_key":"party.display_name","value":" Acme "}]}`,
	))
	if admissionErr != nil {
		t.Fatalf("admit Party patch request: %#v", admissionErr)
	}
	if got := hex.EncodeToString(patch.requestHash[:]); got != "b8838d0c04b72734af771f7e38abca9e94f4f8b4b6ac1e8235204ad9e814139e" {
		t.Fatalf("patch request hash = %s", got)
	}

	conflict, admissionErr := AdmitConflictResolveJSON(
		strings.NewReader(`{"conflict_token":"opaque","resolution_kind":"merged_value","client_txn_id":"txn-conflict","resolved_value":" Merged "}`),
		"opaque",
		ConflictClaims{
			RecordID:          uuid.MustParse("11111111-2222-3333-4444-555555555555"),
			ViewSchemaID:      ViewSchemaID,
			FieldKey:          "party.display_name",
			CurrentRowVersion: 7,
		},
	)
	if admissionErr != nil {
		t.Fatalf("admit Party conflict request: %#v", admissionErr)
	}
	if conflict.resolutionKind != "merged_value" {
		t.Fatalf("conflict resolution kind = %q", conflict.resolutionKind)
	}
	if got := hex.EncodeToString(conflict.requestHash[:]); got != "2b424f63a70372b4f5dcd4209e9948a4c9000a9d8e20c426148461554798b885" {
		t.Fatalf("conflict request hash = %s", got)
	}
}

func TestPartyMutationHashNormalizationEquivalence_Unit(t *testing.T) {
	createHash := func(t *testing.T, email, externalRef, notes string) []byte {
		t.Helper()
		admission, admissionErr := AdmitCreateJSON(strings.NewReader(fmt.Sprintf(
			`{"client_txn_id":"ignored","party.display_name":"José","party.party_kind":"person","party.primary_email":%q,"party.external_ref":%q,"party.notes":%q}`,
			email,
			externalRef,
			notes,
		)))
		if admissionErr != nil {
			t.Fatalf("admit hash-equivalence create: %#v", admissionErr)
		}
		return admission.requestHash[:]
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

func TestPartyPatchRejectsActionPayloadAsUnknownField_Unit(t *testing.T) {
	_, admissionErr := AdmitPatchJSON(strings.NewReader(
		`{"view_schema_id":"cartulary.view.parties.v1","base_row_version":1,"client_txn_id":"txn-action","changes":[{"field_key":"party.notes","action_payload":{}}]}`,
	))
	if admissionErr == nil || admissionErr.Field != "changes" || admissionErr.ReasonCode != "unknown_field" {
		t.Fatalf("Party action_payload admission = %#v", admissionErr)
	}
}

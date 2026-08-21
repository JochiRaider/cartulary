package parties_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/parties"
)

func TestPartyMutationAdmissionAndReplayHashing_Unit(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		request, apiErr := parties.DecodeCreateRequest(strings.NewReader(
			`{"client_txn_id":"txn-create","party.party_kind":"organization","party.display_name":" Acme "}`,
		))
		if apiErr != nil {
			t.Fatalf("decode Party create request: %#v", apiErr)
		}
		if request.ViewSchemaID != parties.ViewSchemaID || request.ClientTxnID != "txn-create" {
			t.Fatalf("unexpected create identity: %#v", request)
		}
		if got := *request.Values["party.display_name"].Text; got != "Acme" {
			t.Fatalf("normalized display name = %q, want Acme", got)
		}
		if got := hex.EncodeToString(parties.CreateRequestHash(request)); got != "fe9207b00d24a2580278277efe883ef49962cb0011407c3dee054ef3961c138b" {
			t.Fatalf("create request hash = %s", got)
		}
	})

	t.Run("patch", func(t *testing.T) {
		request, apiErr := parties.DecodePatchRequest(strings.NewReader(
			`{"view_schema_id":"cartulary.view.parties.v1","base_row_version":4,"client_txn_id":"txn-patch","changes":[{"field_key":"party.display_name","value":" Acme "}]}`,
		))
		if apiErr != nil {
			t.Fatalf("decode Party patch request: %#v", apiErr)
		}
		if len(request.Changes) != 1 || request.Changes[0].CanonicalValue != "Acme" {
			t.Fatalf("unexpected canonical patch: %#v", request)
		}
		if got := hex.EncodeToString(parties.PatchRequestHash(request)); got != "374c89ae3fd19215cabebd9c166b743fe391cee74c06de94a9fe79d35c01d791" {
			t.Fatalf("patch request hash = %s", got)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		recordID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
		claims := parties.ConflictClaims{
			RecordID: recordID, ViewSchemaID: parties.ViewSchemaID,
			FieldKey: "party.display_name", CurrentRowVersion: 7,
		}
		request, apiErr := parties.DecodeConflictResolveRequest(
			strings.NewReader(`{"conflict_token":"opaque","resolution_kind":"merged_value","client_txn_id":"txn-conflict","resolved_value":" Merged "}`),
			"opaque",
			claims,
		)
		if apiErr != nil {
			t.Fatalf("decode Party conflict request: %#v", apiErr)
		}
		if request.Patch == nil || request.Patch.BaseRowVersion != 7 || request.CanonicalValue != "Merged" {
			t.Fatalf("unexpected admitted conflict: %#v", request)
		}
		if got := hex.EncodeToString(parties.ConflictResolveRequestHash(claims, request)); got != "fdfee5e346e6b8db016f01b00eb0f80cc6fdea77c82ba099d91753baac2f6bfa" {
			t.Fatalf("conflict request hash = %s", got)
		}
	})

	for name, testCase := range map[string]struct {
		body   string
		decode func(string) bool
	}{
		"unknown create field": {
			body: `{"client_txn_id":"txn","party.display_name":"Acme","party.party_kind":"organization","legacy":true}`,
			decode: func(body string) bool {
				_, apiErr := parties.DecodeCreateRequest(strings.NewReader(body))
				return apiErr != nil
			},
		},
		"invalid kind": {
			body: `{"client_txn_id":"txn","party.display_name":"Acme","party.party_kind":"legacy"}`,
			decode: func(body string) bool {
				_, apiErr := parties.DecodeCreateRequest(strings.NewReader(body))
				return apiErr != nil
			},
		},
		"wrong patch surface": {
			body: `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"party.display_name","value":"Acme"}]}`,
			decode: func(body string) bool {
				_, apiErr := parties.DecodePatchRequest(strings.NewReader(body))
				return apiErr != nil
			},
		},
		"duplicate patch field": {
			body: `{"view_schema_id":"cartulary.view.parties.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"party.display_name","value":"A"},{"field_key":"party.display_name","value":"B"}]}`,
			decode: func(body string) bool {
				_, apiErr := parties.DecodePatchRequest(strings.NewReader(body))
				return apiErr != nil
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if !testCase.decode(testCase.body) {
				t.Fatalf("expected malformed Party mutation to be rejected: %s", testCase.body)
			}
		})
	}
}

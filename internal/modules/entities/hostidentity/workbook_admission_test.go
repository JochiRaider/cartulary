package hostidentity

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCreateRequestAdmissionAndHashCompatibility(t *testing.T) {
	left, apiErr := DecodeCreateRequest(HostsViewSchemaID, strings.NewReader(`{
		"client_txn_id":"txn-host",
		"host.display_name":" Gateway ",
		"host.hostname":"gateway"
	}`))
	if apiErr != nil {
		t.Fatalf("decode Host create: %#v", apiErr)
	}
	right, apiErr := DecodeCreateRequest(HostsViewSchemaID, strings.NewReader(`{
		"host.hostname":"gateway",
		"host.display_name":"Gateway",
		"client_txn_id":"txn-host"
	}`))
	if apiErr != nil {
		t.Fatalf("decode reordered Host create: %#v", apiErr)
	}
	leftHash := CreateRequestHash(HostsViewSchemaID, left)
	if !bytes.Equal(leftHash, CreateRequestHash(HostsViewSchemaID, right)) {
		t.Fatal("member order changed Entity create replay hash")
	}
	want, err := hex.DecodeString("a2152bd8e7195652b82d4667706f01db5dd24c81be4f82963d4ff06621aea9ce")
	if err != nil {
		t.Fatalf("decode expected create hash: %v", err)
	}
	if !bytes.Equal(leftHash, want) {
		t.Fatalf("create hash = %x, want %x", leftHash, want)
	}
}

func TestWorkbookConflictResolveAdmissionAndHashCompatibility(t *testing.T) {
	claims := WorkbookConflictClaims{
		RecordID:     uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		ViewSchemaID: HostsViewSchemaID, FieldKey: "host.display_name", CurrentRowVersion: 3,
	}
	request, apiErr := DecodeWorkbookConflictResolveRequest(strings.NewReader(`{
		"conflict_token":"opaque",
		"resolution_kind":"use_unsaved",
		"client_txn_id":"txn-conflict",
		"resolved_value":" Gateway "
	}`), "opaque", claims)
	if apiErr != nil {
		t.Fatalf("decode Entity conflict: %#v", apiErr)
	}
	if request.Patch == nil || len(request.Patch.Changes) != 1 ||
		request.Patch.Changes[0].Value == nil || *request.Patch.Changes[0].Value != "Gateway" {
		t.Fatalf("unexpected conflict patch: %#v", request.Patch)
	}
	want, err := hex.DecodeString("7c4c5d36e0fd45670c9c83cf2fa7dc0f5f2fb485b09cc4b517fdd8b17b484f58")
	if err != nil {
		t.Fatalf("decode expected conflict hash: %v", err)
	}
	if got := WorkbookConflictResolveRequestHash(claims, request); !bytes.Equal(got, want) {
		t.Fatalf("conflict hash = %x, want %x", got, want)
	}

	aliasClaims := claims
	aliasClaims.FieldKey = "host.aliases"
	aliasRequest, apiErr := DecodeWorkbookConflictResolveRequest(strings.NewReader(`{
		"conflict_token":"opaque-alias",
		"resolution_kind":"merged_value",
		"client_txn_id":"txn-alias-conflict",
		"resolved_value":{"kind":"collection_actions_v1","actions":[
			{"op":"remove_alias","item_ref":"entity_alias:00000000-0000-4000-8000-000000000002"},
			{"op":"add_alias","alias_text":"  Cafe\u0301 Gateway  "}
		]}
	}`), "opaque-alias", aliasClaims)
	if apiErr != nil {
		t.Fatalf("decode Entity alias conflict: %#v", apiErr)
	}
	if aliasRequest.Patch == nil || len(aliasRequest.Patch.Changes[0].CollectionActions) != 2 ||
		aliasRequest.Patch.Changes[0].CollectionActions[1].NormalizedText != "Café Gateway" {
		t.Fatalf("unexpected alias conflict patch: %#v", aliasRequest.Patch)
	}
}

func TestWorkbookConflictResolveAdmissionRejectsInvalidEntityValues(t *testing.T) {
	claims := WorkbookConflictClaims{
		RecordID:     uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		ViewSchemaID: HostsViewSchemaID, FieldKey: "host.aliases", CurrentRowVersion: 3,
	}
	tests := []struct {
		name   string
		claims WorkbookConflictClaims
		body   string
		field  string
	}{
		{
			name: "invalid alias reference", claims: claims, field: "host.aliases",
			body: `{"conflict_token":"opaque","resolution_kind":"use_unsaved","client_txn_id":"txn","resolved_value":{"kind":"collection_actions_v1","actions":[{"op":"remove_alias","item_ref":"entity_alias:not-a-uuid"}]}}`,
		},
		{
			name: "non-owner field", claims: func() WorkbookConflictClaims {
				value := claims
				value.FieldKey = "host.linked_event_count"
				return value
			}(), field: "field_key",
			body: `{"conflict_token":"opaque","resolution_kind":"use_unsaved","client_txn_id":"txn","resolved_value":1}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, apiErr := DecodeWorkbookConflictResolveRequest(strings.NewReader(test.body), "opaque", test.claims)
			if apiErr == nil || apiErr.Code != "invalid_mutation_payload" || apiErr.Details["field"] != test.field {
				t.Fatalf("API error = %#v", apiErr)
			}
		})
	}
}

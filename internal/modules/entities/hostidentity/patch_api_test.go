package hostidentity

import (
	"strings"
	"testing"
)

func TestPatchRequestDecodeSortsAndHashesEntityChanges(t *testing.T) {
	request, apiErr := DecodePatchRequest(strings.NewReader(`{
		"view_schema_id":"cartulary.view.hosts.v1",
		"base_row_version":3,
		"client_txn_id":"txn-host-patch",
		"changes":[
			{"field_key":"host.location","value":null},
			{"field_key":"host.display_name","value":" Gateway One "}
		]
	}`))
	if apiErr != nil {
		t.Fatalf("decode patch request: %#v", apiErr)
	}
	if request.ViewSchemaID != HostsViewSchemaID || request.BaseRowVersion != 3 || request.ClientTxnID != "txn-host-patch" {
		t.Fatalf("unexpected request identity: %#v", request)
	}
	if len(request.Changes) != 2 || request.Changes[0].FieldKey != "host.display_name" || request.Changes[1].FieldKey != "host.location" {
		t.Fatalf("changes were not sorted by field key: %#v", request.Changes)
	}
	if request.Changes[0].Value == nil || *request.Changes[0].Value != "Gateway One" {
		t.Fatalf("display name was not normalized: %#v", request.Changes[0])
	}
	if request.Changes[1].Value != nil {
		t.Fatalf("nullable location did not decode as nil: %#v", request.Changes[1])
	}
	if hash := PatchRequestHash(request); len(hash) != 32 {
		t.Fatalf("unexpected hash length: %d", len(hash))
	}
}

func TestPatchRequestDecodeRejectsUnsupportedEntityChanges(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		field      string
		reasonCode string
	}{
		{
			name: "empty collection action",
			body: `{
				"view_schema_id":"cartulary.view.hosts.v1",
				"base_row_version":3,
				"client_txn_id":"txn-host-patch",
				"changes":[{"field_key":"host.aliases","action_payload":{"kind":"collection_actions_v1","actions":[]}}]
			}`,
			field:      "host.aliases",
			reasonCode: "invalid_value",
		},
		{
			name: "duplicate field",
			body: `{
				"view_schema_id":"cartulary.view.hosts.v1",
				"base_row_version":3,
				"client_txn_id":"txn-host-patch",
				"changes":[
					{"field_key":"host.hostname","value":"one"},
					{"field_key":"host.hostname","value":"two"}
				]
			}`,
			field:      "changes",
			reasonCode: "duplicate_field_key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := DecodePatchRequest(strings.NewReader(tc.body))
			if apiErr == nil {
				t.Fatalf("expected api error")
			}
			if apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("unexpected code: %s", apiErr.Code)
			}
			if apiErr.Details["field"] != tc.field || apiErr.Details["reason_code"] != tc.reasonCode {
				t.Fatalf("unexpected details: %#v", apiErr.Details)
			}
		})
	}
}

func TestPatchRequestDecodeAcceptsOrderedAliasActions(t *testing.T) {
	request, apiErr := DecodePatchRequest(strings.NewReader(`{
		"view_schema_id":"cartulary.view.hosts.v1",
		"base_row_version":3,
		"client_txn_id":"txn-host-alias-patch",
		"changes":[{
			"field_key":"host.aliases",
			"action_payload":{"kind":"collection_actions_v1","actions":[
				{"op":"remove_alias","item_ref":"entity_alias:00000000-0000-0000-0000-000000000001"},
				{"op":"add_alias","alias_text":"  Cafe\u0301 Gateway  "}
			]}
		}]
	}`))
	if apiErr != nil {
		t.Fatalf("decode alias patch: %#v", apiErr)
	}
	actions := request.Changes[0].CollectionActions
	if len(actions) != 2 || actions[0].Op != "remove_alias" || actions[1].NormalizedText != "Café Gateway" {
		t.Fatalf("unexpected ordered alias actions: %#v", actions)
	}
	if len(PatchRequestHash(request)) != 32 {
		t.Fatal("alias patch hash must be sha256")
	}
}

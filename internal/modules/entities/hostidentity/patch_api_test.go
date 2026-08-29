package hostidentity

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
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
	if request.ViewSchemaID != entitycontract.HostsViewSchemaID || request.BaseRowVersion != 3 || request.ClientTxnID != "txn-host-patch" {
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

	t.Run("Host and Identity accept exactly eight ordinary patch fields", func(t *testing.T) {
		cases := []struct {
			viewSchemaID string
			fields       []string
		}{
			{
				viewSchemaID: entitycontract.HostsViewSchemaID,
				fields: []string{
					"host.aliases", "host.business_owner", "host.containment_status", "host.criticality",
					"host.display_name", "host.hostname", "host.location", "host.os_platform",
				},
			},
			{
				viewSchemaID: entitycontract.IdentitiesViewSchemaID,
				fields: []string{
					"identity.aliases", "identity.display_name", "identity.email", "identity.mfa_state",
					"identity.privilege_level", "identity.reset_status", "identity.sam_account_name", "identity.upn",
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.viewSchemaID, func(t *testing.T) {
				changes := make([]map[string]any, 0, len(tc.fields))
				for _, fieldKey := range tc.fields {
					change := map[string]any{"field_key": fieldKey, "value": "sentinel-" + fieldKey}
					if strings.HasSuffix(fieldKey, ".aliases") {
						change = map[string]any{
							"field_key": fieldKey,
							"action_payload": map[string]any{
								"kind":    "collection_actions_v1",
								"actions": []map[string]any{{"op": "add_alias", "alias_text": "Sentinel Alias"}},
							},
						}
					}
					changes = append(changes, change)
				}
				payload := map[string]any{
					"view_schema_id":   tc.viewSchemaID,
					"base_row_version": 3,
					"client_txn_id":    "txn-exact-patch-partition",
					"changes":          changes,
				}
				data, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("marshal exact patch partition: %v", err)
				}
				decoded, apiErr := DecodePatchRequest(strings.NewReader(string(data)))
				if apiErr != nil || len(decoded.Changes) != 8 {
					t.Fatalf("decode exact patch partition: request=%#v error=%#v", decoded, apiErr)
				}
				gotKeys := make([]string, 0, len(decoded.Changes))
				for _, change := range decoded.Changes {
					gotKeys = append(gotKeys, change.FieldKey)
				}
				if !slices.Equal(gotKeys, tc.fields) {
					t.Fatalf("accepted patch partition mismatch: got %#v want %#v", gotKeys, tc.fields)
				}

				reversed := append([]map[string]any(nil), changes...)
				slices.Reverse(reversed)
				payload["changes"] = reversed
				reversedData, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("marshal reversed patch partition: %v", err)
				}
				reordered, apiErr := DecodePatchRequest(strings.NewReader(string(reversedData)))
				if apiErr != nil || !slices.Equal(PatchRequestHash(decoded), PatchRequestHash(reordered)) {
					t.Fatalf("patch hash must be independent of submitted field order: request=%#v error=%#v", reordered, apiErr)
				}
			})
		}
	})
}

func TestPatchRequestDecodeRejectsUnsupportedEntityChanges(t *testing.T) {
	nonpatchable := map[string][]string{
		entitycontract.HostsViewSchemaID: {
			"host.aad_device_id", "host.edited_at", "host.evidence_count", "host.fqdn",
			"host.host_state", "host.linked_event_count", "host.reusable_identifiers",
		},
		entitycontract.IdentitiesViewSchemaID: {
			"identity.aad_object_id", "identity.edited_at", "identity.evidence_count", "identity.identity_state",
			"identity.linked_event_count", "identity.reusable_identifiers", "identity.sid",
		},
	}
	for viewSchemaID, fieldKeys := range nonpatchable {
		if len(fieldKeys) != 7 {
			t.Fatalf("%s nonpatchable field count = %d, want 7", viewSchemaID, len(fieldKeys))
		}
		for _, fieldKey := range fieldKeys {
			t.Run(fieldKey, func(t *testing.T) {
				payload := map[string]any{
					"view_schema_id":   viewSchemaID,
					"base_row_version": 3,
					"client_txn_id":    "txn-reject-" + fieldKey,
					"changes": []map[string]any{{
						"field_key": fieldKey,
						"value":     map[string]any{"malformed": true},
					}},
				}
				data, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("marshal nonpatchable field: %v", err)
				}
				_, failure := DecodePatchRequest(strings.NewReader(string(data)))
				failureField, _ := failure.Field()
				if failure == nil || failureField != "field_key" || string(failure.ReasonCode()) != "unsupported_field_key" {
					t.Fatalf("nonpatchable field must reject before value interpretation, got %#v", failure)
				}
			})
		}
	}

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
			_, failure := DecodePatchRequest(strings.NewReader(tc.body))
			if failure == nil {
				t.Fatalf("expected admission failure")
			}
			failureField, _ := failure.Field()
			if failureField != tc.field || string(failure.ReasonCode()) != tc.reasonCode {
				t.Fatalf("unexpected failure: %#v", failure)
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

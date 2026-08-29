package hostidentity

import (
	"bytes"
	"encoding/hex"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestCreateRequestAdmissionAndHashCompatibility(t *testing.T) {
	left, apiErr := DecodeCreateRequest(entitycontract.HostsViewSchemaID, strings.NewReader(`{
		"client_txn_id":"txn-host",
		"host.display_name":" Gateway ",
		"host.hostname":"gateway"
	}`))
	if apiErr != nil {
		t.Fatalf("decode Host create: %#v", apiErr)
	}
	right, apiErr := DecodeCreateRequest(entitycontract.HostsViewSchemaID, strings.NewReader(`{
		"host.hostname":"gateway",
		"host.display_name":"Gateway",
		"client_txn_id":"txn-host"
	}`))
	if apiErr != nil {
		t.Fatalf("decode reordered Host create: %#v", apiErr)
	}
	leftHash := CreateRequestHash(entitycontract.HostsViewSchemaID, left)
	if !bytes.Equal(leftHash, CreateRequestHash(entitycontract.HostsViewSchemaID, right)) {
		t.Fatal("member order changed Entity create replay hash")
	}
	want, err := hex.DecodeString("a2152bd8e7195652b82d4667706f01db5dd24c81be4f82963d4ff06621aea9ce")
	if err != nil {
		t.Fatalf("decode expected create hash: %v", err)
	}
	if !bytes.Equal(leftHash, want) {
		t.Fatalf("create hash = %x, want %x", leftHash, want)
	}

	t.Run("create-only identifiers remain admitted on Host and Identity create", func(t *testing.T) {
		host, apiErr := DecodeCreateRequest(entitycontract.HostsViewSchemaID, strings.NewReader(`{
			"client_txn_id":"txn-host-create-only",
			"host.aad_device_id":"AAD-DEVICE-CREATE",
			"host.fqdn":"host.example.test"
		}`))
		if apiErr != nil || host.Values["host.aad_device_id"] != "AAD-DEVICE-CREATE" || host.Values["host.fqdn"] != "host.example.test" {
			t.Fatalf("Host create-only admission mismatch: request=%#v error=%#v", host, apiErr)
		}
		identity, apiErr := DecodeCreateRequest(entitycontract.IdentitiesViewSchemaID, strings.NewReader(`{
			"client_txn_id":"txn-identity-create-only",
			"identity.aad_object_id":"AAD-OBJECT-CREATE",
			"identity.sid":"S-1-5-21-100"
		}`))
		if apiErr != nil || identity.Values["identity.aad_object_id"] != "AAD-OBJECT-CREATE" || identity.Values["identity.sid"] != "S-1-5-21-100" {
			t.Fatalf("Identity create-only admission mismatch: request=%#v error=%#v", identity, apiErr)
		}
	})

	t.Run("field partitions and materialized row shapes are exact", func(t *testing.T) {
		cases := []struct {
			viewSchemaID string
			createOnly   []string
			patchable    []string
			buildRow     func() map[string]any
		}{
			{
				viewSchemaID: entitycontract.HostsViewSchemaID,
				createOnly:   []string{"host.aad_device_id", "host.fqdn"},
				patchable: []string{
					"host.aliases", "host.business_owner", "host.containment_status", "host.criticality",
					"host.display_name", "host.hostname", "host.location", "host.os_platform",
				},
				buildRow: func() map[string]any {
					return buildHostRow(HostRecord{
						RecordID: uuid.MustParse("10000000-0000-4000-8000-000000000001"), RowVersion: 7,
						DisplayName: "Host Sentinel", AADDeviceID: testStringPointer("AAD-SENTINEL"), FQDN: testStringPointer("sentinel.example.test"),
						HostState: "canonical", LinkedEventCount: 2, EvidenceCount: 3, Location: nil,
						UpdatedAt: time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC),
					})
				},
			},
			{
				viewSchemaID: entitycontract.IdentitiesViewSchemaID,
				createOnly:   []string{"identity.aad_object_id", "identity.sid"},
				patchable: []string{
					"identity.aliases", "identity.display_name", "identity.email", "identity.mfa_state",
					"identity.privilege_level", "identity.reset_status", "identity.sam_account_name", "identity.upn",
				},
				buildRow: func() map[string]any {
					return buildIdentityRow(IdentityRecord{
						RecordID: uuid.MustParse("10000000-0000-4000-8000-000000000002"), RowVersion: 11,
						DisplayName: "Identity Sentinel", AADObjectID: testStringPointer("AAD-OBJECT-SENTINEL"), SID: testStringPointer("S-1-5-21-200"),
						IdentityState: "canonical", LinkedEventCount: 4, EvidenceCount: 5, PrivilegeLevel: nil,
						UpdatedAt: time.Date(2026, time.August, 21, 13, 0, 0, 0, time.UTC),
					})
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.viewSchemaID, func(t *testing.T) {
				schema, ok := viewschema.Lookup(tc.viewSchemaID)
				if !ok {
					t.Fatalf("missing view schema %s", tc.viewSchemaID)
				}
				fields := schema.Fields()
				if len(fields) != 15 {
					t.Fatalf("%s field count = %d, want 15", tc.viewSchemaID, len(fields))
				}
				var patchable, nonpatchable, createOnly []string
				for key, field := range fields {
					if field.Writable {
						patchable = append(patchable, key)
					} else {
						nonpatchable = append(nonpatchable, key)
					}
					if field.CreateWritable {
						createOnly = append(createOnly, key)
						if field.Writable || field.GridEditable || field.WriteKind != "direct_value" {
							t.Fatalf("create-only field %s has mismatched flags: %#v", key, field)
						}
					}
				}
				slices.Sort(patchable)
				slices.Sort(nonpatchable)
				slices.Sort(createOnly)
				if !slices.Equal(patchable, tc.patchable) || len(nonpatchable) != 7 || !slices.Equal(createOnly, tc.createOnly) {
					t.Fatalf("%s partition mismatch: patchable=%#v nonpatchable=%#v create_only=%#v", tc.viewSchemaID, patchable, nonpatchable, createOnly)
				}

				row := tc.buildRow()
				cells := row["cells"].(map[string]any)
				if len(cells) != 15 || len(row) != 4 || row["record_id"] == nil || row["row_version"] == nil {
					t.Fatalf("%s row must have exactly 15 cells and root technical members, got %#v", tc.viewSchemaID, row)
				}
				for key := range fields {
					if _, ok := cells[key]; !ok {
						t.Fatalf("%s materializer omitted %s", tc.viewSchemaID, key)
					}
				}
				for _, key := range []string{"record_id", "row_version"} {
					if _, leaked := cells[key]; leaked {
						t.Fatalf("technical member %s must remain outside cells", key)
					}
				}
				prefix := "host"
				if tc.viewSchemaID == entitycontract.IdentitiesViewSchemaID {
					prefix = "identity"
				}
				for _, key := range []string{prefix + ".aliases", prefix + ".reusable_identifiers"} {
					collection := cells[key].(map[string]any)["value"].(map[string]any)
					if collection["kind"] != "collection_value_v1" || collection["ordered"] != false || len(collection["items"].([]map[string]any)) != 0 {
						t.Fatalf("%s collection shape mismatch: %#v", key, collection)
					}
				}
			})
		}
	})
}

func testStringPointer(value string) *string { return &value }

func TestWorkbookConflictResolveAdmissionAndHashCompatibility(t *testing.T) {
	claims := WorkbookConflictClaims{
		RecordID:     uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		ViewSchemaID: entitycontract.HostsViewSchemaID, FieldKey: "host.display_name", CurrentRowVersion: 3,
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
		ViewSchemaID: entitycontract.HostsViewSchemaID, FieldKey: "host.aliases", CurrentRowVersion: 3,
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
	for _, createOnly := range []struct {
		viewSchemaID string
		fieldKey     string
	}{
		{viewSchemaID: entitycontract.HostsViewSchemaID, fieldKey: "host.aad_device_id"},
		{viewSchemaID: entitycontract.HostsViewSchemaID, fieldKey: "host.fqdn"},
		{viewSchemaID: entitycontract.IdentitiesViewSchemaID, fieldKey: "identity.aad_object_id"},
		{viewSchemaID: entitycontract.IdentitiesViewSchemaID, fieldKey: "identity.sid"},
	} {
		createOnlyClaims := claims
		createOnlyClaims.ViewSchemaID = createOnly.viewSchemaID
		createOnlyClaims.FieldKey = createOnly.fieldKey
		tests = append(tests, struct {
			name   string
			claims WorkbookConflictClaims
			body   string
			field  string
		}{
			name:   "create-only " + createOnly.fieldKey,
			claims: createOnlyClaims,
			field:  "field_key",
			body:   `{"conflict_token":"opaque","resolution_kind":"use_unsaved","client_txn_id":"txn","resolved_value":{"malformed":true}}`,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, failure := DecodeWorkbookConflictResolveRequest(strings.NewReader(test.body), "opaque", test.claims)
			failureField, _ := failure.Field()
			if failure == nil || failureField != test.field {
				t.Fatalf("admission failure = %#v", failure)
			}
		})
	}
}

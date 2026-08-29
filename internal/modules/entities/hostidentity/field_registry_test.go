package hostidentity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestEntityFieldRegistryMatchesOwnerProjection_Unit(t *testing.T) {
	hostAliasID := uuid.MustParse("21000000-0000-4000-8000-000000000001")
	hostIdentifierID := uuid.MustParse("21000000-0000-4000-8000-000000000002")
	identityAliasID := uuid.MustParse("22000000-0000-4000-8000-000000000001")
	identityIdentifierID := uuid.MustParse("22000000-0000-4000-8000-000000000002")
	cases := []struct {
		viewSchemaID string
		createOnly   []string
		row          map[string]any
		rowHash      string
	}{
		{
			viewSchemaID: entitycontract.HostsViewSchemaID,
			createOnly:   []string{"host.aad_device_id", "host.fqdn"},
			row: buildHostRow(HostRecord{
				RecordID:         uuid.MustParse("21000000-0000-4000-8000-000000000000"),
				RowVersion:       17,
				DisplayName:      "Registry Host",
				AADDeviceID:      testStringPointer("AAD-REGISTRY"),
				FQDN:             testStringPointer("registry.example.test"),
				Hostname:         testStringPointer("registry-host"),
				HostState:        "canonical",
				LinkedEventCount: 4,
				EvidenceCount:    5,
				Location:         nil,
				OSPlatform:       testStringPointer("linux"),
				BusinessOwner:    testStringPointer("operations"),
				Criticality:      testStringPointer("high"),
				ContainmentStatus: testStringPointer(
					"contained",
				),
				SuggestionOnlyAliases: []AliasValue{{EntityAliasID: hostAliasID, AliasText: "Gateway"}},
				ReusableIdentifiers: []ReusableIdentifier{{
					EntityPreservedIdentifierID: hostIdentifierID,
					IdentifierClass:             "hostname",
					RawValue:                    "legacy-host",
					NormalizedValue:             "legacy-host",
				}},
				UpdatedAt: time.Date(2026, time.August, 22, 4, 0, 0, 123, time.UTC),
			}),
			rowHash: "5d233b139a0a5ca826540c2c0909f7ee79cfa65610684bbd8157eb1893452fb4",
		},
		{
			viewSchemaID: entitycontract.IdentitiesViewSchemaID,
			createOnly:   []string{"identity.aad_object_id", "identity.sid"},
			row: buildIdentityRow(IdentityRecord{
				RecordID:              uuid.MustParse("22000000-0000-4000-8000-000000000000"),
				RowVersion:            23,
				DisplayName:           "Registry Identity",
				AADObjectID:           testStringPointer("OBJECT-REGISTRY"),
				SID:                   testStringPointer("S-1-5-21-220"),
				UPN:                   testStringPointer("registry@example.test"),
				Email:                 nil,
				SamAccountName:        testStringPointer("registry"),
				IdentityState:         "canonical",
				LinkedEventCount:      6,
				EvidenceCount:         7,
				PrivilegeLevel:        testStringPointer("administrator"),
				MFAState:              testStringPointer("enforced"),
				ResetStatus:           testStringPointer("not_required"),
				SuggestionOnlyAliases: []AliasValue{{EntityAliasID: identityAliasID, AliasText: "Operator"}},
				ReusableIdentifiers: []ReusableIdentifier{{
					EntityPreservedIdentifierID: identityIdentifierID,
					IdentifierClass:             "email",
					RawValue:                    "old@example.test",
					NormalizedValue:             "old@example.test",
				}},
				UpdatedAt: time.Date(2026, time.August, 22, 5, 0, 0, 456, time.UTC),
			}),
			rowHash: "4a3db4e82f4867c2d06ced152f17fc7511bfb6b2c1ad360eecfc7597ac6bec0b",
		},
	}

	for _, tc := range cases {
		t.Run(tc.viewSchemaID, func(t *testing.T) {
			schema, ok := viewschema.Lookup(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing owner projection %s", tc.viewSchemaID)
			}
			resource, ok := viewschema.LookupPublicResource(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing public owner projection %s", tc.viewSchemaID)
			}
			descriptors := entityFields.descriptors(tc.viewSchemaID)
			if len(descriptors) != 15 || len(schema.Fields()) != 15 || len(resource.Fields) != 15 {
				t.Fatalf("%s closure counts descriptors=%d owner=%d public=%d, want 15 each", tc.viewSchemaID, len(descriptors), len(schema.Fields()), len(resource.Fields))
			}

			ownerFields := schema.Fields()
			var patchable, nonpatchable, createOnly []string
			patchStrategies := map[entityFieldPatchStrategy]int{}
			for index, descriptor := range descriptors {
				owner, ok := ownerFields[descriptor.fieldKey]
				if !ok {
					t.Fatalf("descriptor %s is absent from owner projection", descriptor.fieldKey)
				}
				if resource.Fields[index].FieldKey != descriptor.fieldKey {
					t.Fatalf("descriptor order %d = %s, want %s", index, descriptor.fieldKey, resource.Fields[index].FieldKey)
				}
				if !reflect.DeepEqual(descriptor.owner, owner) {
					t.Fatalf("descriptor %s owner metadata drifted: got %#v, want %#v", descriptor.fieldKey, descriptor.owner, owner)
				}
				if descriptor.patch != entityPatchStrategyForOwner(owner) || descriptor.clearable() != owner.Clearable || descriptor.supportsGrouping() != owner.Groupable || descriptor.participatesInCreate() != (owner.Writable || owner.CreateWritable) {
					t.Fatalf("descriptor %s implementation metadata does not match owner %#v", descriptor.fieldKey, owner)
				}
				lookedUp, ok := entityFields.lookup(tc.viewSchemaID, descriptor.fieldKey)
				if !ok || !reflect.DeepEqual(lookedUp.owner, owner) {
					t.Fatalf("registry lookup failed for (%s, %s)", tc.viewSchemaID, descriptor.fieldKey)
				}
				patchStrategies[descriptor.patch]++
				if owner.Writable {
					patchable = append(patchable, descriptor.fieldKey)
				} else {
					nonpatchable = append(nonpatchable, descriptor.fieldKey)
				}
				if owner.CreateWritable {
					createOnly = append(createOnly, descriptor.fieldKey)
				}
			}
			slices.Sort(patchable)
			slices.Sort(nonpatchable)
			slices.Sort(createOnly)
			if len(patchable) != 8 || len(nonpatchable) != 7 || !slices.Equal(createOnly, tc.createOnly) {
				t.Fatalf("%s partition patchable=%#v nonpatchable=%#v create_only=%#v", tc.viewSchemaID, patchable, nonpatchable, createOnly)
			}
			if patchStrategies[entityFieldPatchDirect] != 7 || patchStrategies[entityFieldPatchCollection] != 1 || patchStrategies[entityFieldPatchNone] != 7 {
				t.Fatalf("%s patch strategies = %#v, want direct=7 collection=1 none=7", tc.viewSchemaID, patchStrategies)
			}

			cells := tc.row["cells"].(map[string]any)
			groupValues := tc.row["group_values"].(map[string]any)
			if len(cells) != 15 || len(groupValues) != len(schema.GroupingFields()) {
				t.Fatalf("%s row closure cells=%d groups=%d", tc.viewSchemaID, len(cells), len(groupValues))
			}
			for key, field := range ownerFields {
				cell, exists := cells[key]
				if !exists {
					t.Fatalf("row omitted owner field %s", key)
				}
				_, grouped := groupValues[key]
				if grouped != field.Groupable {
					t.Fatalf("field %s grouping support=%t, want %t", key, grouped, field.Groupable)
				}
				if grouped && !reflect.DeepEqual(groupValues[key], cell.(map[string]any)["value"]) {
					t.Fatalf("field %s grouping value diverged from its cell", key)
				}
			}
			payload, err := json.Marshal(tc.row)
			if err != nil {
				t.Fatalf("marshal stable row: %v", err)
			}
			sum := sha256.Sum256(payload)
			if got := hex.EncodeToString(sum[:]); got != tc.rowHash {
				t.Fatalf("%s row hash = %s, want %s", tc.viewSchemaID, got, tc.rowHash)
			}
		})
	}

	t.Run("construction rejects duplicate omission unknown and metadata mismatch", func(t *testing.T) {
		base := entityFieldDescriptors()
		tests := []struct {
			name        string
			descriptors []entityFieldDescriptor
			wantError   string
		}{
			{
				name:        "duplicate",
				descriptors: append(append([]entityFieldDescriptor(nil), base...), base[0]),
				wantError:   "duplicate descriptor cartulary.view.hosts.v1/host.display_name",
			},
			{
				name:        "omission",
				descriptors: append([]entityFieldDescriptor(nil), base[:len(base)-1]...),
				wantError:   "cartulary.view.identities.v1 descriptor count 14 does not match owner field count 15",
			},
			{
				name: "unknown",
				descriptors: func() []entityFieldDescriptor {
					result := append([]entityFieldDescriptor(nil), base...)
					result[0].fieldKey = "host.unknown"
					return result
				}(),
				wantError: "descriptor 0 has unknown field cartulary.view.hosts.v1/host.unknown",
			},
			{
				name: "metadata mismatch",
				descriptors: func() []entityFieldDescriptor {
					result := append([]entityFieldDescriptor(nil), base...)
					result[0].patch = entityFieldPatchNone
					return result
				}(),
				wantError: `descriptor cartulary.view.hosts.v1/host.display_name: patch strategy "none" does not match owner strategy "direct"`,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := newEntityFieldRegistry(tc.descriptors)
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("constructor error = %v, want containing %q", err, tc.wantError)
				}
			})
		}
	})
}

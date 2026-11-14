package hostidentity

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func exerciseEntityCollectionIdentityOrder(t *testing.T) {
	const low = "10000000-0000-4000-8000-000000000001"
	const high = "20000000-0000-4000-8000-000000000001"
	aliases := []AliasValue{{EntityAliasID: uuid.MustParse(high), AliasText: "Alpha"}, {EntityAliasID: uuid.MustParse(low), AliasText: "Zulu"}}
	refs := []PartyReference{{PartyID: uuid.MustParse(high), DisplayText: "Alpha"}, {PartyID: uuid.MustParse(low), DisplayText: "Zulu"}}
	for _, family := range []string{"host", "identity"} {
		t.Run(family, func(t *testing.T) {
			class := "hostname"
			if family == "identity" {
				class = "email"
			}
			identifiers := []ReusableIdentifier{
				{EntityPreservedIdentifierID: uuid.MustParse(high), IdentifierClass: class, RawValue: "Alpha", NormalizedValue: "alpha"},
				{EntityPreservedIdentifierID: uuid.MustParse(low), IdentifierClass: class, RawValue: "Zulu", NormalizedValue: "zulu"},
			}
			assignments := map[string][]PartyReference{}
			for _, role := range []string{"subject", "responsible", "observer"} {
				assignments[family+"."+role+"_party_ids"] = refs
			}
			var row map[string]any
			if family == "host" {
				row = buildHostRow(HostRecord{SuggestionOnlyAliases: aliases, ReusableIdentifiers: identifiers, PartyAssignments: assignments})
			} else {
				row = buildIdentityRow(IdentityRecord{SuggestionOnlyAliases: aliases, ReusableIdentifiers: identifiers, PartyAssignments: assignments})
			}
			expected := map[string][]map[string]any{
				family + ".aliases": {
					{"item_ref": "entity_alias:" + low, "item_kind": "alias", "alias_text": "Zulu", "display_text": "Zulu"},
					{"item_ref": "entity_alias:" + high, "item_kind": "alias", "alias_text": "Alpha", "display_text": "Alpha"},
				},
			}
			label := "Hostname"
			if family == "identity" {
				label = "Email"
			}
			expected[family+".reusable_identifiers"] = []map[string]any{
				{"item_ref": "entity_preserved_identifier:" + low, "item_kind": "reusable_identifier", "identifier_class": class, "raw_value": "Zulu", "normalized_value": "zulu", "display_text": label + ": Zulu"},
				{"item_ref": "entity_preserved_identifier:" + high, "item_kind": "reusable_identifier", "identifier_class": class, "raw_value": "Alpha", "normalized_value": "alpha", "display_text": label + ": Alpha"},
			}
			for _, role := range []string{"subject", "responsible", "observer"} {
				expected[family+"."+role+"_party_ids"] = []map[string]any{
					{"item_ref": "party_ref:" + low, "item_kind": "party_ref", "party_id": low, "display_text": "Zulu"},
					{"item_ref": "party_ref:" + high, "item_kind": "party_ref", "party_id": high, "display_text": "Alpha"},
				}
			}
			for field, items := range expected {
				cell := row["cells"].(map[string]any)[field].(map[string]any)
				want := map[string]any{"kind": "collection_value_v1", "ordered": false, "items": items}
				if !reflect.DeepEqual(cell["value"], want) {
					t.Fatalf("%s did not retain exact typed values in identity order: %#v", field, cell["value"])
				}
			}
			if aliases[0].EntityAliasID.String() != high || refs[0].PartyID.String() != high || identifiers[0].EntityPreservedIdentifierID.String() != high {
				t.Fatal("serialization reordered owner input")
			}
		})
	}
	// The shared constructor preserves authoritative order when explicitly ordered
	// and emits [] for empty values, independently of the entity field registries.
	items := []map[string]any{{"item_ref": "z"}, {"item_ref": "a"}}
	ordered := collectionValue(true, items)
	if !reflect.DeepEqual(ordered, map[string]any{"kind": "collection_value_v1", "ordered": true, "items": []map[string]any{{"item_ref": "z"}, {"item_ref": "a"}}}) {
		t.Fatal("ordered collection was reordered")
	}
	unordered := collectionValue(false, items)
	if !reflect.DeepEqual(unordered["items"], []map[string]any{{"item_ref": "a"}, {"item_ref": "z"}}) || items[0]["item_ref"] != "z" {
		t.Fatal("unordered output changed caller order or failed identity sort")
	}
	if !reflect.DeepEqual(collectionValue(false, nil)["items"], []map[string]any{}) {
		t.Fatal("empty collection is not an array")
	}
}

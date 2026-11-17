package collectionvalue

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func item(ref, label string) map[string]any {
	return map[string]any{"item_ref": ref, "item_kind": "tag", "display_text": label}
}

func refs(items []map[string]any) []string {
	out := make([]string, 0, len(items))
	for _, entry := range items {
		ref, _ := entry["item_ref"].(string)
		out = append(out, ref)
	}
	return out
}

// REQ-01-067: unordered items order by folded label, then raw label, then
// item_ref, bytewise; ordered items keep their authoritative order.
func TestItemsServesUnorderedCollectionsInLabelOrder(t *testing.T) {
	input := []map[string]any{
		item("r:9", "zulu"),
		item("r:1", "alpha"),   // fold tie with "Alpha": raw bytes decide ("A" < "a")
		item("r:8", "Alpha"),   // ...so "Alpha" precedes "alpha" despite its larger ref
		item("r:3", "alpha"),   // exact tie with r:1: item_ref decides
		item("r:2", "café"),    // diacritics stay significant: "cafe" < "café"
		item("r:7", "cafe"),    // ...so "cafe" precedes "café" despite its larger ref
		item("r:5", "Straße"),  // full case folding makes "Straße" a fold tie with
		item("r:4", "strasse"), // "strasse"; raw "S" < "s" puts "Straße" first
		{"item_ref": "r:6"},    // absent label sorts first as ""
		item("r:0", "Ωmega"),   // non-Latin folds to "ωmega", after Latin bytes
	}
	snapshot := refs(input)
	got := refs(Items(false, input))
	want := []string{"r:6", "r:8", "r:1", "r:3", "r:7", "r:2", "r:5", "r:4", "r:9", "r:0"}
	if !slices.Equal(got, want) {
		t.Fatalf("unordered order\n got %v\nwant %v", got, want)
	}
	if !slices.Equal(refs(input), snapshot) {
		t.Fatalf("input was reordered: %v", refs(input))
	}
	if got := refs(Items(true, input)); !slices.Equal(got, snapshot) {
		t.Fatalf("ordered collection must keep authoritative order: %v", got)
	}
}

// The same item set serializes identically whatever its storage order, and
// every stored form decodes to that order.
func TestNewIsStorageOrderIndependentAcrossDecodedForms(t *testing.T) {
	a, b, c := item("r:c", "Gamma"), item("r:a", "beta"), item("r:b", "Alpha")
	forward := New(false, []map[string]any{a, b, c})
	reverse := New(false, []map[string]any{c, b, a})
	if !reflect.DeepEqual(forward, reverse) {
		t.Fatalf("serialization depends on storage order:\n%#v\n%#v", forward, reverse)
	}
	want := []string{"r:b", "r:a", "r:c"}
	if got := refs(forward["items"].([]map[string]any)); !slices.Equal(got, want) {
		t.Fatalf("order got %v want %v", got, want)
	}
	for name, stored := range map[string]any{
		"json string": `[{"item_ref":"r:c","display_text":"Gamma"},{"item_ref":"r:a","display_text":"beta"},{"item_ref":"r:b","display_text":"Alpha"}]`,
		"json bytes":  []byte(`[{"item_ref":"r:c","display_text":"Gamma"},{"item_ref":"r:a","display_text":"beta"},{"item_ref":"r:b","display_text":"Alpha"}]`),
		"decoded any": []any{map[string]any(c), map[string]any(a), map[string]any(b)},
		"maps":        []map[string]any{c, a, b},
	} {
		items, ok := DecodeItems(stored)
		if !ok {
			t.Fatalf("%s: not decoded", name)
		}
		if got := refs(Items(false, items)); !slices.Equal(got, want) {
			t.Fatalf("%s: order got %v want %v", name, got, want)
		}
	}
	for name, invalid := range map[string]any{"number": 7, "object": map[string]any{}, "mixed": []any{map[string]any{}, "x"}, "bad json": "[{"} {
		if _, ok := DecodeItems(invalid); ok {
			t.Fatalf("%s must not decode", name)
		}
	}
}

func TestNewEmptyAndNilItemsSerializeAsEmptyArray(t *testing.T) {
	for _, items := range [][]map[string]any{nil, {}} {
		value := New(false, items)
		got, ok := value["items"].([]map[string]any)
		if !ok || got == nil || len(got) != 0 {
			t.Fatalf("empty items must be a non-nil empty slice: %#v", value["items"])
		}
		if value["kind"] != Kind || value["ordered"] != false {
			t.Fatalf("shape changed: %#v", value)
		}
	}
}

// Conflict values decoded from stored facts are re-served in the same order as
// fresh rows; non-collection values pass through untouched.
func TestCanonicalizeReservesStoredCollectionsOnly(t *testing.T) {
	stored := map[string]any{"kind": Kind, "ordered": false, "items": []any{
		map[string]any{"item_ref": "r:2", "display_text": "b"},
		map[string]any{"item_ref": "r:1", "display_text": "a"},
	}}
	got := Canonicalize(stored).(map[string]any)
	if order := refs(got["items"].([]map[string]any)); !slices.Equal(order, []string{"r:1", "r:2"}) {
		t.Fatalf("stored collection not re-served: %v", order)
	}
	if _, still := stored["items"].([]any); !still {
		t.Fatal("Canonicalize must not mutate its input")
	}
	ordered := map[string]any{"kind": Kind, "ordered": true, "items": []any{
		map[string]any{"item_ref": "r:2", "display_text": "b"},
		map[string]any{"item_ref": "r:1", "display_text": "a"},
	}}
	if order := refs(Canonicalize(ordered).(map[string]any)["items"].([]map[string]any)); !slices.Equal(order, []string{"r:2", "r:1"}) {
		t.Fatalf("ordered collection reordered: %v", order)
	}
	for _, value := range []any{nil, "text", 3, map[string]any{"kind": "other"}, map[string]any{"kind": Kind, "items": 5}} {
		if got := Canonicalize(value); !reflect.DeepEqual(got, value) {
			t.Fatalf("non-collection changed: %#v -> %#v", value, got)
		}
	}
}

func TestWithLinkedRecordLabelsRelabelsWithoutMutatingInput(t *testing.T) {
	known := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	party := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	absent := uuid.MustParse("00000000-0000-4000-8000-000000000003")
	input := map[string]any{"kind": Kind, "ordered": false, "items": []any{
		map[string]any{"item_ref": "record_ref:" + absent.String(), "item_kind": "record_ref", "display_text": "task_request:" + absent.String(), "linked_record_id": absent.String()},
		map[string]any{"item_ref": "record_ref:" + known.String(), "item_kind": "record_ref", "display_text": "decision:" + known.String(), "linked_record_id": known.String()},
		map[string]any{"item_ref": "party_ref:" + party.String(), "item_kind": "party_ref", "display_text": "party:" + party.String(), "party_id": party.String()},
		map[string]any{"item_ref": "record_tag:x", "item_kind": "tag", "display_text": "Keep"},
	}}
	before := input["items"].([]any)[0].(map[string]any)["display_text"]
	if got := LinkedRecordIDs(input, input, "not a collection"); !reflect.DeepEqual(got, []uuid.UUID{absent, known, party}) {
		t.Fatalf("LinkedRecordIDs = %v", got)
	}
	got := WithLinkedRecordLabels(input, map[uuid.UUID]string{known: "Contain host", party: "Blue team"}).(map[string]any)
	labels := []string{}
	for _, item := range got["items"].([]map[string]any) {
		labels = append(labels, item["display_text"].(string))
	}
	// Re-served in REQ-01-067 label order; the absent record reads the
	// REQ-01-333a unavailable label, and non-linked items keep their label.
	if want := []string{"Blue team", "Contain host", "Keep", UnavailableRecordLabel}; !reflect.DeepEqual(labels, want) {
		t.Fatalf("labels = %v, want %v", labels, want)
	}
	if after := input["items"].([]any)[0].(map[string]any)["display_text"]; after != before {
		t.Fatalf("input item mutated: %v", after)
	}
	if WithLinkedRecordLabels("plain", nil) != "plain" {
		t.Fatal("a non-collection value must pass through")
	}
}

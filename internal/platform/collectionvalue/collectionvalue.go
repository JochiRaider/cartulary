// Package collectionvalue owns the collection_value_v1 wire value. Every server
// emission of a collection value is built or re-served here, so ordered
// collections keep their authoritative order and unordered collections are
// served in the REQ-01-067 label order everywhere.
package collectionvalue

import (
	"cmp"
	"encoding/json"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

// Kind is the collection_value_v1 discriminator.
const Kind = "collection_value_v1"

// New returns a collection_value_v1 value whose items are a fresh slice in
// served order. Nil and empty item slices both serialize as an empty array.
func New(ordered bool, items []map[string]any) map[string]any {
	return map[string]any{"kind": Kind, "ordered": ordered, "items": Items(ordered, items)}
}

// Items returns a fresh, non-nil slice in served order. Ordered collections keep
// their authoritative order. Unordered collections sort stably by the label sort
// key of display_text, then the raw display_text, then item_ref, each compared
// bytewise (Unicode scalar order). An absent label or ref sorts as "". The input
// slice is never reordered.
func Items(ordered bool, items []map[string]any) []map[string]any {
	out := make([]map[string]any, len(items))
	copy(out, items)
	if ordered || len(out) < 2 {
		return out
	}
	type keyed struct {
		fold, label, ref string
		item             map[string]any
	}
	sorted := make([]keyed, len(out))
	for index, item := range out {
		label, _ := item["display_text"].(string)
		ref, _ := item["item_ref"].(string)
		sorted[index] = keyed{fold: fieldnorm.TextSortKey(label), label: label, ref: ref, item: item}
	}
	slices.SortStableFunc(sorted, func(left, right keyed) int {
		return cmp.Or(
			strings.Compare(left.fold, right.fold),
			strings.Compare(left.label, right.label),
			strings.Compare(left.ref, right.ref),
		)
	})
	for index, entry := range sorted {
		out[index] = entry.item
	}
	return out
}

// DecodeItems accepts the stored item forms: JSON text or bytes holding an array
// of objects, a decoded []any of objects, or []map[string]any. It reports false
// for any other shape.
func DecodeItems(value any) ([]map[string]any, bool) {
	var data []byte
	switch typed := value.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	case []map[string]any:
		return typed, true
	case []any:
		items := make([]map[string]any, 0, len(typed))
		for _, entry := range typed {
			item, ok := entry.(map[string]any)
			if !ok {
				return nil, false
			}
			items = append(items, item)
		}
		return items, true
	default:
		return nil, false
	}
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, false
	}
	return items, true
}

// Canonicalize re-serves an existing collection_value_v1 object, for example
// one decoded from stored conflict facts, in served order. Any other value is
// returned unchanged.
func Canonicalize(value any) any {
	object, ok := value.(map[string]any)
	if !ok || object["kind"] != Kind {
		return value
	}
	ordered, _ := object["ordered"].(bool)
	items, ok := DecodeItems(object["items"])
	if !ok {
		return value
	}
	canonical := make(map[string]any, len(object))
	for key, member := range object {
		canonical[key] = member
	}
	canonical["items"] = Items(ordered, items)
	return canonical
}

// UnavailableRecordLabel is the REQ-01-333a label of a conflict preview item
// whose linked record is not live in the referring record's incident. It
// discloses neither the record's type nor its state.
const UnavailableRecordLabel = "Unavailable record"

// linkedRecordKey returns the member naming the linked record of a record_ref or
// party_ref item; a Party is a record, so both carry REQ-01-333a labels.
func linkedRecordKey(item map[string]any) (string, bool) {
	switch item["item_kind"] {
	case "record_ref":
		return "linked_record_id", true
	case "party_ref":
		return "party_id", true
	default:
		return "", false
	}
}

// LinkedRecordIDs returns, in first-seen order and without duplicates, the
// records named by record_ref and party_ref items of the given collection
// values. Other values and items are ignored.
func LinkedRecordIDs(values ...any) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	ids := []uuid.UUID{}
	for _, value := range values {
		object, ok := value.(map[string]any)
		if !ok || object["kind"] != Kind {
			continue
		}
		items, ok := DecodeItems(object["items"])
		if !ok {
			continue
		}
		for _, item := range items {
			key, ok := linkedRecordKey(item)
			if !ok {
				continue
			}
			text, _ := item[key].(string)
			id, err := uuid.Parse(text)
			if err != nil || seen[id] {
				continue
			}
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// WithLinkedRecordLabels returns value re-served with each record_ref and
// party_ref item labelled from labels, or UnavailableRecordLabel when its record
// is absent. Items are copied, never mutated. Any other value is returned
// unchanged.
func WithLinkedRecordLabels(value any, labels map[uuid.UUID]string) any {
	object, ok := value.(map[string]any)
	if !ok || object["kind"] != Kind {
		return value
	}
	items, ok := DecodeItems(object["items"])
	if !ok {
		return value
	}
	relabelled := make([]map[string]any, len(items))
	for index, item := range items {
		relabelled[index] = item
		key, ok := linkedRecordKey(item)
		if !ok {
			continue
		}
		label := UnavailableRecordLabel
		text, _ := item[key].(string)
		if id, err := uuid.Parse(text); err == nil {
			if known, found := labels[id]; found {
				label = known
			}
		}
		copied := make(map[string]any, len(item))
		for member, memberValue := range item {
			copied[member] = memberValue
		}
		copied["display_text"] = label
		relabelled[index] = copied
	}
	ordered, _ := object["ordered"].(bool)
	served := make(map[string]any, len(object))
	for member, memberValue := range object {
		served[member] = memberValue
	}
	served["items"] = Items(ordered, relabelled)
	return served
}

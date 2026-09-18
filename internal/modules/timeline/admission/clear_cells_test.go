package admission

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
)

func TestClearCellsAdmissionAndIdentity_Unit(t *testing.T) {
	base := map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "clear-identity", "kind": "clear_cells_v1",
		"field_keys": []string{"timeline.raw_activity_text", "timeline.analyst_text"},
		"targets":    []map[string]any{{"record_id": "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "base_row_version": 2}},
	}
	decode := func(body map[string]any) (BulkMutationRequest, bool) {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, failure := DecodeBulkMutationRequest(strings.NewReader(string(encoded)), timeline.TimelineViewSchemaID)
		return request, failure == nil
	}
	request, ok := decode(base)
	if !ok || len(request.FieldKeys) != 2 {
		t.Fatalf("valid clear rejected: %#v", request)
	}
	hash := BulkMutationRequestHash(request)
	base["targets"].([]map[string]any)[0]["record_id"] = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	canonical, ok := decode(base)
	if !ok || !bytes.Equal(hash, BulkMutationRequestHash(canonical)) {
		t.Fatal("UUID spelling changed identity")
	}
	base["field_keys"] = []string{"timeline.analyst_text", "timeline.raw_activity_text"}
	reordered, ok := decode(base)
	if !ok || bytes.Equal(hash, BulkMutationRequestHash(reordered)) {
		t.Fatal("field order must participate in identity")
	}
	for _, fields := range []any{nil, []string{}, []string{"timeline.raw_activity_text", "timeline.raw_activity_text"}, []string{"timeline.tags"}, []string{"timeline.capture_state"}, []string{"timeline.missing"}, []string{"timeline.raw_activity_text", "timeline.evidence_count"}, make([]string, 11)} {
		base["field_keys"] = fields
		if _, ok := decode(base); ok {
			t.Fatalf("admitted invalid fields %#v", fields)
		}
	}
	base["field_keys"] = mutationpolicy.DirectWritableFieldKeys()
	if _, ok := decode(base); !ok {
		t.Fatal("all ten operational fields must be clearable")
	}
	for _, key := range []string{"value", "field_key", "tag_name", "clipboard_text", "selector"} {
		base[key] = nil
		if _, ok := decode(base); ok {
			t.Fatalf("admitted forbidden %s", key)
		}
		delete(base, key)
	}
	target := base["targets"].([]map[string]any)[0]
	base["targets"] = []map[string]any{target, target}
	if _, ok := decode(base); ok {
		t.Fatal("duplicate record admitted")
	}
	base["targets"] = make([]map[string]any, 501)
	if _, ok := decode(base); ok {
		t.Fatal("oversize clear admitted")
	}
	base["targets"] = []map[string]any{target}
	delete(base, "field_keys")
	if _, ok := decode(base); ok {
		t.Fatal("missing fields admitted")
	}
}

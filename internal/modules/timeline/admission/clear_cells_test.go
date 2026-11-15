package admission

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

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
	base["field_keys"] = mutationpolicy.DirectWritableFieldKeys()
	if _, ok := decode(base); !ok {
		t.Fatal("all ten operational fields must be clearable")
	}
	target := base["targets"].([]map[string]any)[0]
	delete(base, "field_keys")
	// REQ-01-672: each rejection is invalid_mutation_payload at its own member.
	reject := func(body map[string]any, field, reason string) {
		t.Helper()
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		_, failure := DecodeBulkMutationRequest(strings.NewReader(string(encoded)), timeline.TimelineViewSchemaID)
		if failure == nil || failure.Code != "invalid_mutation_payload" || len(failure.Details) != 2 || failure.Details["field"] != field || failure.Details["reason_code"] != reason {
			t.Fatalf("rejection = %#v, want %s/%s", failure, field, reason)
		}
	}
	reject(base, "field_keys", "missing_required_field")
	for _, fields := range []any{nil, []string{}, []string{"timeline.raw_activity_text", "timeline.raw_activity_text"}, []string{"timeline.tags"}, []string{"timeline.capture_state"}, []string{"timeline.missing"}, []string{"timeline.raw_activity_text", "timeline.evidence_count"}, make([]string, 11)} {
		base["field_keys"] = fields
		reject(base, "field_keys", "invalid_value")
	}
	base["field_keys"] = []string{"timeline.raw_activity_text"}
	for key, reason := range map[string]string{"value": "forbidden_field", "field_key": "forbidden_field", "tag_name": "forbidden_field", "clipboard_text": "unknown_field", "selector": "unknown_field"} {
		base[key] = nil
		reject(base, key, reason)
		delete(base, key)
	}
	base["targets"] = []map[string]any{target, target}
	reject(base, "targets", "duplicate_target")
	base["targets"] = []map[string]any{target, {"kind": "create"}}
	reject(base, "targets", "invalid_value")
	distinct := make([]map[string]any, mutationpolicy.MaxOwnerBatchTargets+1)
	for index := range distinct {
		distinct[index] = map[string]any{"record_id": uuid.NewSHA1(uuid.NameSpaceOID, []byte(strconv.Itoa(index))).String(), "base_row_version": 1}
	}
	base["targets"] = distinct[:mutationpolicy.MaxOwnerBatchTargets]
	if request, ok := decode(base); !ok || len(request.Targets) != mutationpolicy.MaxOwnerBatchTargets {
		t.Fatal("500 distinct clear targets must be admitted")
	}
	base["targets"] = distinct
	reject(base, "targets", "invalid_value")
}

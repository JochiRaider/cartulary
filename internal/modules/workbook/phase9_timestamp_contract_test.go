package workbook

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
)

func TestPhase9_TimestampInstantPatchDecoder_U_9_10(t *testing.T) {
	utc := mustDecodeTimestampPatch(t, "2026-04-24T12:00:00Z")
	offset := mustDecodeTimestampPatch(t, "2026-04-24T08:00:00-04:00")
	if !utc.Changes[0].Value.Timestamp.Equal(*offset.Changes[0].Value.Timestamp) {
		t.Fatalf("offset-equivalent timestamps were not equal: %s vs %s", utc.Changes[0].Value.Timestamp, offset.Changes[0].Value.Timestamp)
	}
	if got := utc.Changes[0].Value.Timestamp.UTC().Format(time.RFC3339Nano); got != "2026-04-24T12:00:00Z" {
		t.Fatalf("unexpected UTC normalization: %s", got)
	}
	if !bytes.Equal(PatchRequestHash(utc), PatchRequestHash(offset)) {
		t.Fatalf("offset-equivalent timestamps must produce the same idempotency hash")
	}

	clearableNull := mustDecodePatch(t, TaskRequestsViewSchemaID, "task.due_at", nil)
	if clearableNull.Changes[0].Value.Kind != "null" {
		t.Fatalf("clearable timestamp null decoded as %#v", clearableNull.Changes[0].Value)
	}
	expectDecodePatchRejected(t, CommLogViewSchemaID, "comm_log.timestamp_utc", nil)

	for _, value := range []any{
		"2026-04-24T12:00:00",
		"2026-04-24",
		"",
		" 2026-04-24T12:00:00Z",
		"2026-04-24T12:00:00Z ",
		float64(42),
		true,
		[]any{},
		map[string]any{},
	} {
		expectDecodePatchRejected(t, TaskRequestsViewSchemaID, "task.due_at", value)
	}

	query, err := viewquery.Decode(strings.NewReader(`{"filters":[{"field_key":"task.due_at","op":"eq","arg":{"value":"2026-04-24T08:00:00-04:00"}}]}`), TaskRequestsViewSchemaID)
	if err != nil {
		t.Fatalf("decode timestamp query operand: %#v", err)
	}
	if got := query.Meta.Filters[0].Arg["value"]; got != "2026-04-24T12:00:00Z" {
		t.Fatalf("unexpected query timestamp normalization: %#v", got)
	}
	if _, err := viewquery.Decode(strings.NewReader(`{"filters":[{"field_key":"task.due_at","op":"eq","arg":{"value":" 2026-04-24T12:00:00Z"}}]}`), TaskRequestsViewSchemaID); err == nil {
		t.Fatalf("expected padded query timestamp to fail closed")
	}
}

func mustDecodeTimestampPatch(t testing.TB, value string) PatchRequest {
	t.Helper()
	request := mustDecodePatch(t, TaskRequestsViewSchemaID, "task.due_at", value)
	change := request.Changes[0]
	if change.Value == nil || change.Value.Kind != "timestamp" || change.Value.Timestamp == nil {
		t.Fatalf("expected timestamp value change, got %#v", change.Value)
	}
	return request
}

func mustDecodePatch(t testing.TB, viewSchemaID string, fieldKey string, value any) PatchRequest {
	t.Helper()
	payload := patchPayload(t, viewSchemaID, fieldKey, value)
	request, apiErr := DecodePatchRequest(strings.NewReader(payload))
	if apiErr != nil {
		t.Fatalf("decode patch unexpectedly failed for %s=%#v: %#v", fieldKey, value, apiErr)
	}
	return request
}

func expectDecodePatchRejected(t testing.TB, viewSchemaID string, fieldKey string, value any) {
	t.Helper()
	payload := patchPayload(t, viewSchemaID, fieldKey, value)
	if _, apiErr := DecodePatchRequest(strings.NewReader(payload)); apiErr == nil {
		t.Fatalf("expected timestamp patch %s=%#v to fail closed", fieldKey, value)
	}
}

func patchPayload(t testing.TB, viewSchemaID string, fieldKey string, value any) string {
	t.Helper()
	payload := map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-u-9-10",
		"changes": []map[string]any{
			{
				"field_key": fieldKey,
				"value":     value,
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal patch payload: %v", err)
	}
	return string(data)
}

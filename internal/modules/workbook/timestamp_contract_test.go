package workbook

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
)

func TestTimestampInstantPatchDecoder_Unit(t *testing.T) {
	utc := mustDecodeTimestampPatch(t, "2026-04-24T12:00:00Z")
	offset := mustDecodeTimestampPatch(t, "2026-04-24T08:00:00-04:00")
	if !utc.Changes[0].Value.Timestamp.Equal(*offset.Changes[0].Value.Timestamp) {
		t.Fatalf("offset-equivalent timestamps were not equal: %s vs %s", utc.Changes[0].Value.Timestamp, offset.Changes[0].Value.Timestamp)
	}
	if got := utc.Changes[0].Value.Timestamp.UTC().Format(time.RFC3339Nano); got != "2026-04-24T12:00:00Z" {
		t.Fatalf("unexpected UTC normalization: %s", got)
	}
	if !bytes.Equal(tasksdecisions.PatchRequestHash(utc), tasksdecisions.PatchRequestHash(offset)) {
		t.Fatalf("offset-equivalent timestamps must produce the same idempotency hash")
	}

	clearableTask := mustDecodePatch(t, tasksdecisions.TaskRequestsViewSchemaID, "task.due_at", nil)
	clearableValue := clearableTask.Changes[0].Value
	if clearableValue == nil || clearableValue.Text != nil || clearableValue.Timestamp != nil ||
		clearableValue.UUID != nil || clearableValue.Number != nil || clearableValue.Bool != nil {
		t.Fatalf("clearable timestamp task.due_at null decoded as %#v", clearableTask.Changes[0].Value)
	}
	for _, tc := range []struct {
		viewSchemaID string
		fieldKey     string
	}{
		{artifacts.CommLogViewSchemaID, "comm_log.next_report_at"},
		{artifacts.HandoffViewSchemaID, "handoff.acknowledged_at"},
		{artifacts.StatusReviewViewSchemaID, "status_review.next_report_at"},
	} {
		clearableNull := mustDecodeArtifactPatch(t, tc.viewSchemaID, tc.fieldKey, nil)
		value := clearableNull.Changes[0].Value
		if value == nil || value.Text != nil || value.Timestamp != nil || value.UUID != nil || value.Number != nil || value.Bool != nil {
			t.Fatalf("clearable timestamp %s null decoded as %#v", tc.fieldKey, clearableNull.Changes[0].Value)
		}
	}
	for _, tc := range []struct {
		viewSchemaID string
		fieldKey     string
	}{
		{artifacts.CommLogViewSchemaID, "comm_log.timestamp_utc"},
		{artifacts.HandoffViewSchemaID, "handoff.timestamp_utc"},
		{artifacts.StatusReviewViewSchemaID, "status_review.timestamp_utc"},
		{artifacts.LessonViewSchemaID, "lesson.timestamp_utc"},
	} {
		expectArtifactDecodePatchRejected(t, tc.viewSchemaID, tc.fieldKey, nil)
	}

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
		expectDecodePatchRejected(t, tasksdecisions.TaskRequestsViewSchemaID, "task.due_at", value)
	}

	query, err := viewquery.Decode(strings.NewReader(`{"filters":[{"field_key":"task.due_at","op":"eq","arg":{"value":"2026-04-24T08:00:00-04:00"}}]}`), tasksdecisions.TaskRequestsViewSchemaID)
	if err != nil {
		t.Fatalf("decode timestamp query operand: %#v", err)
	}
	if got := query.Meta.Filters[0].Arg["value"]; got != "2026-04-24T12:00:00Z" {
		t.Fatalf("unexpected query timestamp normalization: %#v", got)
	}
	if _, err := viewquery.Decode(strings.NewReader(`{"filters":[{"field_key":"task.due_at","op":"eq","arg":{"value":" 2026-04-24T12:00:00Z"}}]}`), tasksdecisions.TaskRequestsViewSchemaID); err == nil {
		t.Fatalf("expected padded query timestamp to fail closed")
	}
}

func mustDecodeTimestampPatch(t testing.TB, value string) tasksdecisions.PatchRequest {
	t.Helper()
	request := mustDecodePatch(t, tasksdecisions.TaskRequestsViewSchemaID, "task.due_at", value)
	change := request.Changes[0]
	if change.Value == nil || change.Value.Timestamp == nil {
		t.Fatalf("expected timestamp value change, got %#v", change.Value)
	}
	return request
}

func mustDecodePatch(t testing.TB, viewSchemaID string, fieldKey string, value any) tasksdecisions.PatchRequest {
	t.Helper()
	payload := patchPayload(t, viewSchemaID, fieldKey, value)
	request, apiErr := tasksdecisions.DecodePatchRequest(strings.NewReader(payload))
	if apiErr != nil {
		t.Fatalf("decode patch unexpectedly failed for %s=%#v: %#v", fieldKey, value, apiErr)
	}
	return request
}

func expectDecodePatchRejected(t testing.TB, viewSchemaID string, fieldKey string, value any) {
	t.Helper()
	payload := patchPayload(t, viewSchemaID, fieldKey, value)
	if _, apiErr := tasksdecisions.DecodePatchRequest(strings.NewReader(payload)); apiErr == nil {
		t.Fatalf("expected timestamp patch %s=%#v to fail closed", fieldKey, value)
	}
}

func mustDecodeArtifactPatch(t testing.TB, viewSchemaID string, fieldKey string, value any) artifacts.PatchRequest {
	t.Helper()
	request, apiErr := artifacts.DecodePatchRequest(strings.NewReader(patchPayload(t, viewSchemaID, fieldKey, value)))
	if apiErr != nil {
		t.Fatalf("decode Artifact patch unexpectedly failed for %s=%#v: %#v", fieldKey, value, apiErr)
	}
	return request
}

func expectArtifactDecodePatchRejected(t testing.TB, viewSchemaID string, fieldKey string, value any) {
	t.Helper()
	if _, apiErr := artifacts.DecodePatchRequest(strings.NewReader(patchPayload(t, viewSchemaID, fieldKey, value))); apiErr == nil {
		t.Fatalf("expected Artifact timestamp patch %s=%#v to fail closed", fieldKey, value)
	}
}

func patchPayload(t testing.TB, viewSchemaID string, fieldKey string, value any) string {
	t.Helper()
	payload := map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook_interaction-u-9-10",
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

package workbook

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSupportPhase6Unit_ConflictHelperDetectsOverlappingFields(t *testing.T) {
	changed := map[string]workbookPatchChangedField{
		"note.title": {ServerUpdatedBy: uuid.New(), ServerUpdatedAt: time.Now().UTC()},
	}
	requested := []PatchChange{
		{FieldKey: "note.body", Value: &ValueChange{Kind: "text", Text: ptr("client body")}},
	}
	if _, _, ok := overlappingWorkbookPatchChange(requested, changed); ok {
		t.Fatal("different-field stale write must be eligible for auto-rebase")
	}
	requested = append(requested, PatchChange{FieldKey: "note.title", Value: &ValueChange{Kind: "text", Text: ptr("client title")}})
	if change, _, ok := overlappingWorkbookPatchChange(requested, changed); !ok || change.FieldKey != "note.title" {
		t.Fatalf("same-field stale write must be detected by field_key, got %q ok=%v", change.FieldKey, ok)
	}
}

func TestSupportPhase6Unit_ConflictHelperBuildsOpaqueConflictToken(t *testing.T) {
	recordID := uuid.New()
	actorID := uuid.New()
	requestHash := hashRequestPayload(map[string]any{"client": "phase6-u-6-02"})
	conflict, err := buildWorkbookSameFieldConflict(
		recordID,
		NotesViewSchemaID,
		1,
		2,
		requestHash,
		workbookPatchConflictWindow{BaseRow: phase6Row("note.title", "Base"), ChangedFields: nil},
		PatchChange{FieldKey: "note.title", Value: &ValueChange{Kind: "text", Text: ptr("Client")}},
		workbookPatchChangedField{ServerUpdatedBy: actorID, ServerUpdatedAt: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)},
		phase6Row("note.title", "Server"),
	)
	if err != nil {
		t.Fatalf("build conflict: %v", err)
	}
	payload := conflict.Conflict
	if payload["record_id"] != recordID.String() ||
		payload["field_key"] != "note.title" ||
		payload["conflict_resolution_class"] != "text_compare_merge" ||
		payload["base_row_version"] != int64(1) ||
		payload["current_row_version"] != int64(2) ||
		payload["base_value"] != "Base" ||
		payload["server_value"] != "Server" ||
		payload["client_value"] != "Client" ||
		payload["server_updated_by"] != actorID.String() ||
		payload["server_updated_at"] == "" {
		t.Fatalf("unexpected conflict payload: %#v", payload)
	}
	token, ok := payload["conflict_token"].(string)
	if !ok || token == "" || strings.Contains(token, "Client") || strings.Contains(token, "Server") || strings.Contains(token, "Base") {
		t.Fatalf("conflict token must be opaque and content-free, got %q", token)
	}
	claims, ok := parseWorkbookConflictToken(token)
	if !ok ||
		claims.RecordID != recordID.String() ||
		claims.ViewSchemaID != NotesViewSchemaID ||
		claims.FieldKey != "note.title" ||
		claims.CurrentRowVersion != 2 ||
		claims.RequestHash != base64.RawURLEncoding.EncodeToString(requestHash) {
		t.Fatalf("unexpected conflict token claims: %#v ok=%v", claims, ok)
	}
}

func TestSupportPhase6Unit_ConflictHelperSuggestsOnlyCleanTextMerges(t *testing.T) {
	suggested, ok := suggestedTextMergeValue("one\r\ntwo\nthree", "one\r\nTWO\nthree", "one\ntwo\nTHREE")
	if !ok || suggested != "one\nTWO\nTHREE" {
		t.Fatalf("clean line merge got %q ok=%v", suggested, ok)
	}
	if suggested, ok := suggestedTextMergeValue("one\ntwo", "one\nserver", "one\nclient"); ok {
		t.Fatalf("overlapping line merge must not suggest a value, got %q", suggested)
	}
	conflict, err := buildWorkbookSameFieldConflict(
		uuid.New(),
		NotesViewSchemaID,
		1,
		2,
		hashRequestPayload(map[string]any{"client": "phase6-u-6-03"}),
		workbookPatchConflictWindow{BaseRow: phase6Row("note.body", "a\nb\nc"), ChangedFields: nil},
		PatchChange{FieldKey: "note.body", Value: &ValueChange{Kind: "text", Text: ptr("a\nb\nclient")}},
		workbookPatchChangedField{ServerUpdatedBy: uuid.New(), ServerUpdatedAt: time.Now().UTC()},
		phase6Row("note.body", "server\nb\nc"),
	)
	if err != nil {
		t.Fatalf("build text conflict: %v", err)
	}
	if conflict.Conflict["suggested_merged_value"] != "server\nb\nclient" {
		t.Fatalf("expected optional clean suggestion without commit semantics, got %#v", conflict.Conflict)
	}
}

func TestSupportPhase6Unit_ConflictHelperBuildsCollectionConflictValues(t *testing.T) {
	recordID := uuid.New()
	base := map[string]any{
		"kind":    "collection_value_v1",
		"ordered": false,
		"items":   []any{},
	}
	conflict, err := buildWorkbookSameFieldConflict(
		recordID,
		NotesViewSchemaID,
		1,
		2,
		hashRequestPayload(map[string]any{"client": "phase6-u-6-04"}),
		workbookPatchConflictWindow{BaseRow: phase6Row("note.tags", base), ChangedFields: nil},
		PatchChange{FieldKey: "note.tags", Collection: &CollectionActionPayload{Actions: []CollectionAction{{Op: "add_token", NormalizedText: "client-tag"}}}},
		workbookPatchChangedField{ServerUpdatedBy: uuid.New(), ServerUpdatedAt: time.Now().UTC()},
		phase6Row("note.tags", map[string]any{
			"kind":    "collection_value_v1",
			"ordered": false,
			"items": []any{map[string]any{
				"item_ref":     "tag:server-tag",
				"item_kind":    "tag",
				"display_text": "server-tag",
				"raw_text":     "server-tag",
			}},
		}),
	)
	if err != nil {
		t.Fatalf("build collection conflict: %v", err)
	}
	for _, key := range []string{"base_value", "server_value", "client_value"} {
		value, ok := conflict.Conflict[key].(map[string]any)
		if !ok || value["kind"] != "collection_value_v1" {
			t.Fatalf("%s must use collection_value_v1, got %#v", key, conflict.Conflict[key])
		}
	}
	request, apiErr := DecodeConflictResolveRequest(strings.NewReader(`{
		"conflict_token":"token",
		"resolution_kind":"merged_value",
		"client_txn_id":"txn-resolve",
		"resolved_value":{"kind":"collection_actions_v1","actions":[{"op":"add_token","raw_text":"final-tag"}]}
	}`), "token", workbookConflictTokenClaims{ViewSchemaID: NotesViewSchemaID, FieldKey: "note.tags"})
	if apiErr != nil {
		t.Fatalf("decode collection resolver request: %#v", apiErr)
	}
	if request.ResolvedChange == nil || request.ResolvedChange.Collection == nil || request.ResolvedChange.Collection.Actions[0].NormalizedText != "final-tag" {
		t.Fatalf("resolver must accept collection_actions_v1 for collection_review merged values: %#v", request)
	}
}

func phase6Row(fieldKey string, value any) map[string]any {
	return map[string]any{"cells": map[string]any{fieldKey: map[string]any{"value": value}}}
}

func ptr(value string) *string {
	return &value
}

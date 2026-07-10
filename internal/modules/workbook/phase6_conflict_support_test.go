package workbook

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/workbook/conflicts"
)

func TestSupportPhase6Unit_ConflictHelperDetectsOverlappingFields(t *testing.T) {
	changed := map[string]conflicts.PatchChangedField{
		"note.title": {ServerUpdatedBy: uuid.New(), ServerUpdatedAt: time.Now().UTC()},
	}
	requested := []conflicts.PatchChange{
		{FieldKey: "note.body", Value: "client body"},
	}
	if _, _, ok := conflicts.OverlappingPatchChange(requested, changed); ok {
		t.Fatal("different-field stale write must be eligible for auto-rebase")
	}
	requested = append(requested, conflicts.PatchChange{FieldKey: "note.title", Value: "client title"})
	if change, _, ok := conflicts.OverlappingPatchChange(requested, changed); !ok || change.FieldKey != "note.title" {
		t.Fatalf("same-field stale write must be detected by field_key, got %q ok=%v", change.FieldKey, ok)
	}
}

func TestSupportPhase6Unit_ConflictHelperBuildsOpaqueConflictToken(t *testing.T) {
	recordID := uuid.New()
	actorID := uuid.New()
	requestHash := hashRequestPayload(map[string]any{"client": "phase6-u-6-02"})
	conflict, err := conflicts.BuildSameFieldConflict(conflicts.SameFieldConflictParams{
		RouteKey:          workbookConflictResolveRouteKey,
		RecordID:          recordID,
		ViewSchemaID:      NotesViewSchemaID,
		BaseRowVersion:    1,
		CurrentRowVersion: 2,
		RequestHash:       requestHash,
		Window:            conflicts.PatchConflictWindow{BaseRow: phase6Row("note.title", "Base"), ChangedFields: nil},
		Change:            conflicts.PatchChange{FieldKey: "note.title", Value: "Client"},
		Changed:           conflicts.PatchChangedField{ServerUpdatedBy: actorID, ServerUpdatedAt: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)},
		CurrentRow:        phase6Row("note.title", "Server"),
		Codec:             defaultWorkbookConflictTokenCodec,
	})
	if err != nil {
		t.Fatalf("build conflict: %v", err)
	}
	payload := conflict
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
	suggested, ok := conflicts.SuggestedTextMergeValue("one\r\ntwo\nthree", "one\r\nTWO\nthree", "one\ntwo\nTHREE")
	if !ok || suggested != "one\nTWO\nTHREE" {
		t.Fatalf("clean line merge got %q ok=%v", suggested, ok)
	}
	if suggested, ok := conflicts.SuggestedTextMergeValue("one\ntwo", "one\nserver", "one\nclient"); ok {
		t.Fatalf("overlapping line merge must not suggest a value, got %q", suggested)
	}
	conflict, err := conflicts.BuildSameFieldConflict(conflicts.SameFieldConflictParams{
		RouteKey:          workbookConflictResolveRouteKey,
		RecordID:          uuid.New(),
		ViewSchemaID:      NotesViewSchemaID,
		BaseRowVersion:    1,
		CurrentRowVersion: 2,
		RequestHash:       hashRequestPayload(map[string]any{"client": "phase6-u-6-03"}),
		Window:            conflicts.PatchConflictWindow{BaseRow: phase6Row("note.body", "a\nb\nc"), ChangedFields: nil},
		Change:            conflicts.PatchChange{FieldKey: "note.body", Value: "a\nb\nclient"},
		Changed:           conflicts.PatchChangedField{ServerUpdatedBy: uuid.New(), ServerUpdatedAt: time.Now().UTC()},
		CurrentRow:        phase6Row("note.body", "server\nb\nc"),
		Codec:             defaultWorkbookConflictTokenCodec,
	})
	if err != nil {
		t.Fatalf("build text conflict: %v", err)
	}
	if conflict["suggested_merged_value"] != "server\nb\nclient" {
		t.Fatalf("expected optional clean suggestion without commit semantics, got %#v", conflict)
	}
}

func TestSupportPhase6Unit_ConflictHelperBuildsCollectionConflictValues(t *testing.T) {
	recordID := uuid.New()
	serverTagID := uuid.NewString()
	base := map[string]any{
		"kind":    "collection_value_v1",
		"ordered": false,
		"items":   []any{},
	}
	conflict, err := conflicts.BuildSameFieldConflict(conflicts.SameFieldConflictParams{
		RouteKey:          workbookConflictResolveRouteKey,
		RecordID:          recordID,
		ViewSchemaID:      NotesViewSchemaID,
		BaseRowVersion:    1,
		CurrentRowVersion: 2,
		RequestHash:       hashRequestPayload(map[string]any{"client": "phase6-u-6-04"}),
		Window:            conflicts.PatchConflictWindow{BaseRow: phase6Row("note.tags", base), ChangedFields: nil},
		Change: conflicts.PatchChange{
			FieldKey:   "note.tags",
			Collection: &conflicts.CollectionActionPayload{Actions: []conflicts.CollectionAction{{Op: "add_tag", RawText: "client-tag", NormalizedText: "client-tag"}}},
		},
		Changed: conflicts.PatchChangedField{ServerUpdatedBy: uuid.New(), ServerUpdatedAt: time.Now().UTC()},
		CurrentRow: phase6Row("note.tags", map[string]any{
			"kind":    "collection_value_v1",
			"ordered": false,
			"items": []any{map[string]any{
				"item_ref":     "record_tag:" + recordID.String() + ":" + serverTagID,
				"item_kind":    "tag",
				"display_text": "server-tag",
				"tag_id":       serverTagID,
			}},
		}),
		Codec: defaultWorkbookConflictTokenCodec,
	})
	if err != nil {
		t.Fatalf("build collection conflict: %v", err)
	}
	for _, key := range []string{"base_value", "server_value", "client_value"} {
		value, ok := conflict[key].(map[string]any)
		if !ok || value["kind"] != "collection_value_v1" {
			t.Fatalf("%s must use collection_value_v1, got %#v", key, conflict[key])
		}
	}
	request, apiErr := DecodeConflictResolveRequest(strings.NewReader(`{
		"conflict_token":"token",
		"resolution_kind":"merged_value",
		"client_txn_id":"txn-resolve",
		"resolved_value":{"kind":"collection_actions_v1","actions":[{"op":"add_tag","tag_name":"final-tag"}]}
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

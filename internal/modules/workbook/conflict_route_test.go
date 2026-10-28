package workbook_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestGridWriteConcurrencyRoute_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "collaboration-u-6-01-grid-write-concurrency", "IR-COLLABORATION-workbook-storage")
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-01-create", "Base title", "Base body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))

	routeLeft := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-01-route-left-create", "Route left", "Left base body")
	routeLeftID := appsupport.MustUUID(t, routeLeft["record_id"].(string))
	routeRight := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-01-route-right-create", "Route right", "Right base body")
	routeRightID := appsupport.MustUUID(t, routeRight["record_id"].(string))
	routeBoundPatchBody := map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-route-bound-body",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "Path-selected body"}},
	}
	routeLeftPatch := requireWorkbookPatch(t, harness, login, routeLeftID, routeBoundPatchBody)
	requireCellValue(t, routeLeftPatch["row"].(map[string]any), "note.body", "Path-selected body")
	routeRightAfterLeft := RequireQueriedRow(t, harness, login, incidentID, artifacts.NotesViewSchemaID, routeRightID)
	requireCellValue(t, routeRightAfterLeft, "note.body", "Right base body")
	routeRightPatch := requireWorkbookPatch(t, harness, login, routeRightID, routeBoundPatchBody)
	requireCellValue(t, routeRightPatch["row"].(map[string]any), "note.body", "Path-selected body")
	routeLeftAfterRight := RequireQueriedRow(t, harness, login, incidentID, artifacts.NotesViewSchemaID, routeLeftID)
	requireCellValue(t, routeLeftAfterRight, "note.body", "Path-selected body")

	beforeMissingBaseState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	missingBase := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id": artifacts.NotesViewSchemaID,
		"client_txn_id":  "txn-collaboration-u-6-01-missing-base",
		"changes":        []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	httptestx.RequireErrorEnvelope(t, missingBase, http.StatusBadRequest, "invalid_mutation_payload")
	beforeFullRowPatchState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	requireWorkbookStateUnchanged(t, beforeMissingBaseState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	fullRowPatch := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"record_id":        recordID.String(),
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-record-id-in-body",
		"note.body":        "Client body",
	})
	httptestx.RequireErrorEnvelope(t, fullRowPatch, http.StatusBadRequest, "invalid_mutation_payload")
	requireWorkbookStateUnchanged(t, beforeFullRowPatchState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))

	titlePatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-server-title",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})
	requireCellValue(t, titlePatch["row"].(map[string]any), "note.title", "Server title")

	bodyPatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-client-body",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	bodyRow := bodyPatch["row"].(map[string]any)
	requireCellValue(t, bodyRow, "note.title", "Server title")
	requireCellValue(t, bodyRow, "note.body", "Client body")
	if got := int64(bodyRow["row_version"].(float64)); got != 3 {
		t.Fatalf("different-field stale edit row_version = %d want 3", got)
	}
	RequireMutationChangedFields(t, harness, bodyPatch["change_set_id"].(string), []string{"note.body"})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	beforeSameFieldState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	sameField := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-client-title",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Client title"}},
	})
	body := httptestx.RequireErrorEnvelope(t, sameField, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeSameFieldState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["field_key"] != "note.title" || conflict["client_value"] != "Client title" || conflict["server_value"] != "Server title" {
		t.Fatalf("unexpected same-field conflict payload: %#v", conflict)
	}
	afterConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("same-field stale edit wrote durable side effects: before=%+v after=%+v", beforeConflict, afterConflict)
	}
}

func TestSameFieldConflictHTTP_Integration(t *testing.T) {
	harness, login, actorID, incidentID := ConflictFixture(t, "collaboration-u-6-02-same-field-http", "IR-COLLABORATION-workbook-storage")
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-02-create", "Base title", "Base body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-02-server",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})

	beforeRespState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-02-client",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Client title"}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeRespState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	errorObject := body["error"].(map[string]any)
	if errorObject["status"] != float64(http.StatusConflict) {
		t.Fatalf("same-field conflict error.status = %#v want 409", errorObject["status"])
	}
	conflict := errorObject["conflict"].(map[string]any)
	RequireNoLegacyConflictAliases(t, conflict)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "note.title" ||
		conflict["conflict_resolution_class"] != "text_compare_merge" ||
		conflict["base_value"] != "Base title" ||
		conflict["server_value"] != "Server title" ||
		conflict["client_value"] != "Client title" ||
		conflict["server_updated_by"] != actorID.String() {
		t.Fatalf("unexpected same-field conflict identity/value fields: %#v", conflict)
	}
	RequireConflictVersion(t, conflict, "base_row_version", 1)
	RequireConflictVersion(t, conflict, "current_row_version", 2)
	if _, err := time.Parse(time.RFC3339Nano, conflict["server_updated_at"].(string)); err != nil {
		t.Fatalf("server_updated_at was not RFC3339Nano: %v conflict=%#v", err, conflict)
	}
}

func TestTextCompareMergeDurability_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "collaboration-u-6-03-text-merge-durability", "IR-COLLABORATION-workbook-storage")
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-03-clean-create", "Merge note", "one\ntwo\nthree")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-clean-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nTWO\nthree"}},
	})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	beforeRespState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-clean-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\ntwo\nTHREE"}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeRespState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["suggested_merged_value"] != "one\nTWO\nTHREE" {
		t.Fatalf("expected clean text suggestion without accepting patch, got %#v", conflict)
	}
	afterConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("clean text conflict wrote durable side effects: before=%+v after=%+v", beforeConflict, afterConflict)
	}
	current := RequireQueriedRow(t, harness, login, incidentID, artifacts.NotesViewSchemaID, recordID)
	requireCellValue(t, current, "note.body", "one\nTWO\nthree")
	if got := int64(current["row_version"].(float64)); got != 2 {
		t.Fatalf("rejected clean text conflict changed row_version = %d want 2", got)
	}

	resolved := ResolveConflict(t, harness, login, recordID, conflict["conflict_token"].(string), map[string]any{
		"conflict_token":  conflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-collaboration-u-6-03-clean-resolve",
		"resolved_value":  "one\nTWO\nTHREE",
	})
	requireCellValue(t, resolved["row"].(map[string]any), "note.body", "one\nTWO\nTHREE")
	afterResolve := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if afterResolve.ChangeSets != beforeConflict.ChangeSets+1 || afterResolve.RecordRevisions != beforeConflict.RecordRevisions+1 {
		t.Fatalf("explicit text resolution should be the next durable revision: before=%+v after=%+v", beforeConflict, afterResolve)
	}

	overlap := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-03-overlap-create", "Overlap note", "one\ntwo")
	overlapID := appsupport.MustUUID(t, overlap["record_id"].(string))
	requireWorkbookPatch(t, harness, login, overlapID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-overlap-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nserver"}},
	})
	beforeOverlapRespState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	overlapResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", overlapID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-overlap-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nclient"}},
	})
	overlapBody := httptestx.RequireErrorEnvelope(t, overlapResp, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeOverlapRespState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	overlapConflict := overlapBody["error"].(map[string]any)["conflict"].(map[string]any)
	if _, ok := overlapConflict["suggested_merged_value"]; ok {
		t.Fatalf("overlapping text conflict must omit suggested_merged_value: %#v", overlapConflict)
	}

	renderedLookingBase := "# Title\r\n<div data-x=\"base\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-1\">HOST</entity-chip>"
	renderedLookingServer := "# Title\n<div data-x=\"server\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-1\">HOST</entity-chip>"
	renderedLookingClient := "# Title\n<div data-x=\"base\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-2\">HOST</entity-chip>"
	renderedLookingSuggestion := "# Title\n<div data-x=\"server\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-2\">HOST</entity-chip>"
	rendered := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-03-rendered-create", "Rendered-looking note", renderedLookingBase)
	renderedID := appsupport.MustUUID(t, rendered["record_id"].(string))
	requireWorkbookPatch(t, harness, login, renderedID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-rendered-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": renderedLookingServer}},
	})
	beforeRenderedConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, renderedID)
	beforeRenderedRespState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	renderedResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", renderedID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-rendered-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": renderedLookingClient}},
	})
	renderedBody := httptestx.RequireErrorEnvelope(t, renderedResp, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeRenderedRespState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	renderedConflict := renderedBody["error"].(map[string]any)["conflict"].(map[string]any)
	if renderedConflict["base_value"] != "# Title\n<div data-x=\"base\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-1\">HOST</entity-chip>" ||
		renderedConflict["server_value"] != renderedLookingServer ||
		renderedConflict["client_value"] != renderedLookingClient ||
		renderedConflict["suggested_merged_value"] != renderedLookingSuggestion {
		t.Fatalf("rendered-looking text must remain raw plain text in conflict payload: %#v", renderedConflict)
	}
	afterRenderedConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, renderedID)
	if beforeRenderedConflict != afterRenderedConflict {
		t.Fatalf("rendered-looking clean text conflict wrote durable side effects: before=%+v after=%+v", beforeRenderedConflict, afterRenderedConflict)
	}
	renderedCurrent := RequireQueriedRow(t, harness, login, incidentID, artifacts.NotesViewSchemaID, renderedID)
	requireCellValue(t, renderedCurrent, "note.body", renderedLookingServer)
}

func TestConflictResolveDurability_Integration(t *testing.T) {
	harness, login, actorID, incidentID := ConflictFixture(t, "collaboration-u-6-06-conflict-resolve-durability", "IR-COLLABORATION-workbook-storage")

	keepRecordID, keepConflict := CreateNoteTitleConflict(t, harness, login, incidentID, "keep", "Keep saved", "Keep local")
	beforeKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID)
	keepBody := map[string]any{
		"conflict_token":  keepConflict["conflict_token"].(string),
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-collaboration-u-6-06-keep-resolve",
	}
	beforeKeepdataState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	keepData := ResolveConflict(t, harness, login, keepRecordID, keepConflict["conflict_token"].(string), keepBody)
	requireWorkbookKeepSaved(t, beforeKeepdataState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	RequireNoChangeSetID(t, keepData)
	requireCellValue(t, keepData["row"].(map[string]any), "note.title", "Keep saved")
	afterKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID)
	RequireSideEffectDelta(t, "keep_saved", beforeKeep, afterKeep, workbookConflictSideEffects{
		RouteIdempotency: 1,
	})
	beforeKeepReplayState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	keepReplayData := ResolveConflict(t, harness, login, keepRecordID, keepConflict["conflict_token"].(string), keepBody)
	requireWorkbookStateUnchanged(t, beforeKeepReplayState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	if !reflect.DeepEqual(keepReplayData, keepData) {
		t.Fatalf("keep_saved exact replay changed the original success:\nfirst  %#v\nreplay %#v", keepData, keepReplayData)
	}
	RequireSideEffectDelta(t, "keep_saved replay", beforeKeep, snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID), workbookConflictSideEffects{
		RouteIdempotency: 1,
	})

	useRecordID, useConflict := CreateNoteTitleConflict(t, harness, login, incidentID, "use", "Use saved", "Use local")
	beforeUse := snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID)
	useBody := map[string]any{
		"conflict_token":  useConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-collaboration-u-6-06-use-resolve",
		"resolved_value":  "Use local",
	}
	useData := ResolveConflict(t, harness, login, useRecordID, useConflict["conflict_token"].(string), useBody)
	requireCellValue(t, useData["row"].(map[string]any), "note.title", "Use local")
	requireWorkbookNoteFields(t, useData["row"].(map[string]any))
	requireCellValue(t, RequireQueriedRow(t, harness, login, incidentID, artifacts.NotesViewSchemaID, useRecordID), "note.title", "Use local")
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, useData["change_set_id"].(string))
	workbookscenariotest.RequireChangeSetAttribution(t, harness.DB, useData["change_set_id"].(string), actorID.String(), "workbook.records.conflicts.resolve", "txn-collaboration-u-6-06-use-resolve")
	afterUse := snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID)
	RequireSideEffectDelta(t, "use_unsaved", beforeUse, afterUse, workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
	beforeUseReplayState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	useReplayData := ResolveConflict(t, harness, login, useRecordID, useConflict["conflict_token"].(string), useBody)
	requireWorkbookStateUnchanged(t, beforeUseReplayState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	if !reflect.DeepEqual(useReplayData, useData) {
		t.Fatalf("use_unsaved exact replay changed the original success:\nfirst  %#v\nreplay %#v", useData, useReplayData)
	}
	RequireSideEffectDelta(t, "use_unsaved replay", beforeUse, snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID), workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})

	mergedRecordID, mergedConflict := CreateNoteTitleConflict(t, harness, login, incidentID, "merged", "Merged saved", "Merged local")
	if mergedConflict["field_key"] != "note.title" {
		t.Fatal("conflict returned wrong field")
	}
	beforeInvalidKind := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	invalid := ResolveConflictRaw(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), map[string]any{
		"conflict_token": mergedConflict["conflict_token"], "resolution_kind": "unsupported_resolution", "client_txn_id": "txn-invalid-resolution-kind",
	})
	invalidBody := httptestx.RequireErrorEnvelope(t, invalid, http.StatusBadRequest, "invalid_mutation_payload")
	if httptestx.RequireErrorDetails(t, invalidBody)["field"] != "resolution_kind" {
		t.Fatal("invalid resolution must identify resolution_kind")
	}
	requireWorkbookStateUnchanged(t, beforeInvalidKind, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	beforeMerged := snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID)
	mergedBody := map[string]any{
		"conflict_token":  mergedConflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-collaboration-u-6-06-merged-resolve",
		"resolved_value":  "  Merged final  ",
	}
	mergedData := ResolveConflict(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), mergedBody)
	requireCellValue(t, mergedData["row"].(map[string]any), "note.title", "Merged final")
	requireWorkbookNoteFields(t, mergedData["row"].(map[string]any))
	requireCellValue(t, RequireQueriedRow(t, harness, login, incidentID, artifacts.NotesViewSchemaID, mergedRecordID), "note.title", "Merged final")
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, mergedData["change_set_id"].(string))
	workbookscenariotest.RequireChangeSetAttribution(t, harness.DB, mergedData["change_set_id"].(string), actorID.String(), "workbook.records.conflicts.resolve", "txn-collaboration-u-6-06-merged-resolve")
	afterMerged := snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID)
	RequireSideEffectDelta(t, "merged_value", beforeMerged, afterMerged, workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
	beforeMergedReplayState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	mergedReplayData := ResolveConflict(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), mergedBody)
	requireWorkbookStateUnchanged(t, beforeMergedReplayState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	if !reflect.DeepEqual(mergedReplayData, mergedData) {
		t.Fatalf("merged_value exact replay changed the original success:\nfirst  %#v\nreplay %#v", mergedData, mergedReplayData)
	}
	RequireSideEffectDelta(t, "merged_value replay", beforeMerged, snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID), workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
	beforeDivergence := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	mergedBody["resolved_value"] = "Divergent resolution"
	divergent := ResolveConflictRaw(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), mergedBody)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
	requireWorkbookStateUnchanged(t, beforeDivergence, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
}

func TestSourceOwnerConflictContributions_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "workbook-source-owner-conflict-contributions", "IR-WORKBOOK-SOURCE-CONFLICTS")

	party := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-source-conflict-party-create",
		"party.display_name": "Party base",
		"party.party_kind":   "organization",
	})["row"].(map[string]any)
	partyID := appsupport.MustUUID(t, party["record_id"].(string))
	partyConflict := createAtomicConflict(
		t,
		harness,
		login,
		incidentID,
		partyID,
		"cartulary.view.parties.v1",
		"party.display_name",
		"Party saved",
		"Party local",
		"party",
	)
	partyResolved := ResolveConflict(t, harness, login, partyID, partyConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  partyConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-workbook-source-conflict-party-resolve",
		"resolved_value":  "Party local",
	})
	requireCellValue(t, partyResolved["row"].(map[string]any), "party.display_name", "Party local")
	partyKeep := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-source-conflict-party-keep-create",
		"party.display_name": "Party keep base",
		"party.party_kind":   "organization",
	})["row"].(map[string]any)
	partyKeepID := appsupport.MustUUID(t, partyKeep["record_id"].(string))
	partyKeepConflict := createAtomicConflict(
		t,
		harness,
		login,
		incidentID,
		partyKeepID,
		"cartulary.view.parties.v1",
		"party.display_name",
		"Party saved again",
		"Party local again",
		"party-keep",
	)
	partyKeepBody := map[string]any{
		"conflict_token":  partyKeepConflict["conflict_token"].(string),
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-workbook-source-conflict-party-keep",
	}
	beforePartykeptState := snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.parties.v1")
	partyKept := ResolveConflict(t, harness, login, partyKeepID, partyKeepConflict["conflict_token"].(string), partyKeepBody)
	requireWorkbookKeepSaved(t, beforePartykeptState, snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.parties.v1"))
	RequireNoChangeSetID(t, partyKept)
	requireCellValue(t, partyKept["row"].(map[string]any), "party.display_name", "Party saved again")
	beforePartykeepReplayState := snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.parties.v1")
	partyKeepReplay := ResolveConflict(t, harness, login, partyKeepID, partyKeepConflict["conflict_token"].(string), partyKeepBody)
	requireWorkbookStateUnchanged(t, beforePartykeepReplayState, snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.parties.v1"))
	if !reflect.DeepEqual(partyKeepReplay, partyKept) {
		t.Fatalf("source-owned Party keep_saved replay changed the original success:\nfirst  %#v\nreplay %#v", partyKept, partyKeepReplay)
	}

	evidence := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-workbook-source-conflict-evidence-create",
		"evidence.title": "Evidence base",
	})["row"].(map[string]any)
	evidenceID := appsupport.MustUUID(t, evidence["record_id"].(string))
	evidenceConflict := createAtomicConflict(
		t,
		harness,
		login,
		incidentID,
		evidenceID,
		"cartulary.view.evidence.v1",
		"evidence.title",
		"Evidence saved",
		"Evidence local",
		"evidence",
	)
	beforeKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, evidenceID)
	keepBody := map[string]any{
		"conflict_token":  evidenceConflict["conflict_token"].(string),
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-workbook-source-conflict-evidence-keep",
	}
	beforeEvidencekeptState := snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.evidence.v1")
	evidenceKept := ResolveConflict(t, harness, login, evidenceID, evidenceConflict["conflict_token"].(string), keepBody)
	requireWorkbookKeepSaved(t, beforeEvidencekeptState, snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.evidence.v1"))
	RequireNoChangeSetID(t, evidenceKept)
	requireCellValue(t, evidenceKept["row"].(map[string]any), "evidence.title", "Evidence saved")
	RequireSideEffectDelta(
		t,
		"source-owned evidence keep_saved",
		beforeKeep,
		snapshotWorkbookConflictSideEffects(t, harness, incidentID, evidenceID),
		workbookConflictSideEffects{RouteIdempotency: 1},
	)
	beforeEvidenceReplayState := snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.evidence.v1")
	evidenceReplay := ResolveConflict(t, harness, login, evidenceID, evidenceConflict["conflict_token"].(string), keepBody)
	requireWorkbookStateUnchanged(t, beforeEvidenceReplayState, snapshotWorkbookState(t, harness, login, incidentID, "cartulary.view.evidence.v1"))
	if !reflect.DeepEqual(evidenceReplay, evidenceKept) {
		t.Fatalf("source-owned evidence keep_saved replay changed the original success:\nfirst  %#v\nreplay %#v", evidenceKept, evidenceReplay)
	}
	RequireSideEffectDelta(
		t,
		"source-owned evidence keep_saved replay",
		beforeKeep,
		snapshotWorkbookConflictSideEffects(t, harness, incidentID, evidenceID),
		workbookConflictSideEffects{RouteIdempotency: 1},
	)
}

func TestConflictResolutionRevalidatesStaleToken_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "collaboration-u-6-07-stale-conflict-resolution", "IR-COLLABORATION-STALE-CONFLICT")
	recordID, conflict := CreateNoteTitleConflict(t, harness, login, incidentID, "stale", "Saved at token issue", "Unsaved at token issue")
	oldToken := conflict["conflict_token"].(string)

	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 2,
		"client_txn_id":    "txn-collaboration-u-6-07-server-advanced",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Saved after token issue",
		}},
	})
	beforeRejectedResolution := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)

	beforeResponseState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	response := ResolveConflictRaw(t, harness, login, recordID, oldToken, map[string]any{
		"conflict_token":  oldToken,
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-collaboration-u-6-07-stale-resolution",
		"resolved_value":  "Unsaved at token issue",
	})
	body := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeResponseState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	fresh := body["error"].(map[string]any)["conflict"].(map[string]any)
	if fresh["conflict_token"] == oldToken {
		t.Fatalf("stale resolution returned the consumed token instead of a fresh conflict: %#v", fresh)
	}
	RequireConflictVersion(t, fresh, "base_row_version", 2)
	RequireConflictVersion(t, fresh, "current_row_version", 3)
	if got := fresh["server_value"]; got != "Saved after token issue" {
		t.Fatalf("fresh conflict did not revalidate current source value: %#v", fresh)
	}
	afterRejectedResolution := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if afterRejectedResolution != beforeRejectedResolution {
		t.Fatalf("stale conflict resolution wrote durable effects: before=%+v after=%+v", beforeRejectedResolution, afterRejectedResolution)
	}
}

func createAtomicConflict(
	t testing.TB,
	harness *appsupport.ServerHarness,
	login authcookietest.AuthCookies,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	viewSchemaID string,
	fieldKey string,
	savedValue any,
	localValue any,
	txnSuffix string,
) map[string]any {
	t.Helper()
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-source-conflict-" + txnSuffix + "-server",
		"changes":          []map[string]any{{"field_key": fieldKey, "value": savedValue}},
	})
	beforeResponseState := snapshotWorkbookState(t, harness, login, incidentID, viewSchemaID)
	response := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-source-conflict-" + txnSuffix + "-client",
		"changes":          []map[string]any{{"field_key": fieldKey, "value": localValue}},
	})
	body := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeResponseState, snapshotWorkbookState(t, harness, login, incidentID, viewSchemaID))
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != fieldKey ||
		conflict["server_value"] != savedValue ||
		conflict["client_value"] != localValue ||
		conflict["conflict_token"] == "" {
		t.Fatalf("unexpected source-owner conflict payload for %s: %#v", fieldKey, conflict)
	}
	return conflict
}

func ConflictFixture(t testing.TB, name string, incidentKey string) (*appsupport.ServerHarness, authcookietest.AuthCookies, uuid.UUID, uuid.UUID) {
	t.Helper()
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: name, TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-" + incidentKey,
		"incident_key":  incidentKey,
		"title":         name,
	})
	return harness, login, actorID, appsupport.MustUUID(t, incident["incident_id"].(string))
}

func CreateNote(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, clientTxnID string, title string, body string) map[string]any {
	t.Helper()
	data := requireWorkbookCreate(t, harness, login, incidentID, artifacts.NotesViewSchemaID, map[string]any{
		"client_txn_id": clientTxnID,
		"note.title":    title,
		"note.body":     body,
	})
	return data["row"].(map[string]any)
}

func CreateNoteTitleConflict(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, suffix string, savedValue string, localValue string) (uuid.UUID, map[string]any) {
	t.Helper()
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-06-"+suffix+"-create", suffix+" base", suffix+" body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-06-" + suffix + "-server",
		"changes":          []map[string]any{{"field_key": "note.title", "value": savedValue}},
	})
	beforeRespState := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-06-" + suffix + "-client",
		"changes":          []map[string]any{{"field_key": "note.title", "value": localValue}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeRespState, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "note.title" ||
		conflict["server_value"] != savedValue ||
		conflict["client_value"] != localValue {
		t.Fatalf("unexpected workbook-storage conflict payload: %#v", conflict)
	}
	return recordID, conflict
}

func RequireNoChangeSetID(t testing.TB, data map[string]any) {
	t.Helper()
	if _, ok := data["change_set_id"]; ok {
		t.Fatalf("response must not carry change_set_id: %#v", data)
	}
}

func RequireSideEffectDelta(t testing.TB, label string, before workbookConflictSideEffects, after workbookConflictSideEffects, want workbookConflictSideEffects) {
	t.Helper()
	got := workbookConflictSideEffects{
		ChangeSets:        after.ChangeSets - before.ChangeSets,
		MutationRows:      after.MutationRows - before.MutationRows,
		RecordRevisions:   after.RecordRevisions - before.RecordRevisions,
		RouteIdempotency:  after.RouteIdempotency - before.RouteIdempotency,
		ActiveRecordLinks: after.ActiveRecordLinks - before.ActiveRecordLinks,
		ActiveRiskRefs:    after.ActiveRiskRefs - before.ActiveRiskRefs,
	}
	if got != want {
		t.Fatalf("%s side-effect delta = %+v want %+v (before=%+v after=%+v)", label, got, want, before, after)
	}
}

func RequireMutationChangedFields(t testing.TB, harness *appsupport.ServerHarness, changeSetID string, want []string) {
	t.Helper()
	rows, err := harness.DB.QueryContext(context.Background(), `
SELECT fact.field_key
  FROM record_revision_conflict_facts AS fact
  JOIN record_revisions AS revision
    ON revision.revision_id = fact.revision_id
 WHERE revision.change_set_id = $1
 ORDER BY fact.field_key
`, changeSetID)
	if err != nil {
		t.Fatalf("load change-set conflict facts %s: %v", changeSetID, err)
	}
	defer rows.Close()
	got := make([]string, 0, len(want))
	for rows.Next() {
		var fieldKey string
		if err := rows.Scan(&fieldKey); err != nil {
			t.Fatalf("scan change-set conflict fact %s: %v", changeSetID, err)
		}
		if fieldKey == "note.updated_at" {
			continue
		}
		got = append(got, fieldKey)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate change-set conflict facts %s: %v", changeSetID, err)
	}
	if !StringSlicesEqual(got, want) {
		t.Fatalf("change-set conflict fact fields = %#v want %#v", got, want)
	}
}

func StringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func RequireNoLegacyConflictAliases(t testing.TB, conflict map[string]any) {
	t.Helper()
	for _, key := range []string{"current_field_value", "conflict_resolution"} {
		if _, ok := conflict[key]; ok {
			t.Fatalf("same-field conflict preserved legacy alias %q: %#v", key, conflict)
		}
	}
}

func RequireConflictVersion(t testing.TB, conflict map[string]any, key string, want int64) {
	t.Helper()
	got, ok := conflict[key].(float64)
	if !ok || int64(got) != want {
		t.Fatalf("unexpected %s: got %#v want %d conflict=%#v", key, conflict[key], want, conflict)
	}
}

func RequireQueriedRow(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		httptestx.WithCookies(login.Session),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	rows := body["data"].(map[string]any)["rows"].([]any)
	for _, row := range rows {
		rowMap := row.(map[string]any)
		if rowMap["record_id"] == recordID.String() {
			return rowMap
		}
	}
	t.Fatalf("query did not return record %s in %s rows=%#v", recordID, viewSchemaID, rows)
	return nil
}

func ResolveConflictRaw(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, recordID uuid.UUID, conflictToken string, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/conflicts/"+conflictToken+"/resolve",
		body,
		httptestx.WithCookies(login.Session, login.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value),
	)
}

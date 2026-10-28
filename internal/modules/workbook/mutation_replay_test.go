package workbook_test

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
	"github.com/google/uuid"
)

func TestWorkbookCreateReplay_Integration(t *testing.T) {
	h, login, _, incidentID := ConflictFixture(t, "workbook-create-replay", "IR-WORKBOOK-CREATE-REPLAY")
	body := map[string]any{"client_txn_id": "txn-create-replay", "note.title": "  Created note  ", "note.body": "Created body"}
	response := doWorkbookJSON(t, h, login, http.MethodPost, incidentID, artifacts.NotesViewSchemaID, uuid.Nil, body)
	created := httptestx.RequireSuccessEnvelope(t, response, http.StatusCreated)["data"].(map[string]any)
	row := created["row"].(map[string]any)
	recordID := appsupport.MustUUID(t, row["record_id"].(string))
	requireCellValue(t, row, "note.title", "Created note")
	requireCellValue(t, row, "note.body", "Created body")
	requireWorkbookNoteFields(t, row)
	// Replay returns the committed result even after the live row advances.
	requireWorkbookPatch(t, h, login, recordID, map[string]any{
		"view_schema_id": artifacts.NotesViewSchemaID, "base_row_version": 1, "client_txn_id": "txn-create-later-write",
		"changes": []map[string]any{{"field_key": "note.body", "value": "Later body"}},
	})
	before := snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID)
	response = doWorkbookJSON(t, h, login, http.MethodPost, incidentID, artifacts.NotesViewSchemaID, uuid.Nil, body)
	replay := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(created, replay) {
		t.Fatal("create replay changed the original committed result")
	}
	requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID))
	body["note.title"] = "Divergent title"
	before = snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID)
	response = doWorkbookJSON(t, h, login, http.MethodPost, incidentID, artifacts.NotesViewSchemaID, uuid.Nil, body)
	httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "client_txn_conflict")
	requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID))
}

func TestWorkbookPatchReplay_Integration(t *testing.T) {
	h, login, _, incidentID := ConflictFixture(t, "workbook-patch-replay", "IR-WORKBOOK-PATCH-REPLAY")
	row := CreateNote(t, h, login, incidentID, "txn-patch-note", "Initial title", "Initial body")
	recordID := appsupport.MustUUID(t, row["record_id"].(string))
	body := map[string]any{
		"view_schema_id": artifacts.NotesViewSchemaID, "base_row_version": 1, "client_txn_id": "txn-patch-replay",
		"changes": []map[string]any{{"field_key": "note.body", "value": "  Patched body  "}},
	}
	patched := requireWorkbookPatch(t, h, login, recordID, body)
	requireCellValue(t, patched["row"].(map[string]any), "note.body", "Patched body")
	requireWorkbookNoteFields(t, patched["row"].(map[string]any))
	RequireMutationChangedFields(t, h, patched["change_set_id"].(string), []string{"note.body"})
	requireWorkbookPatch(t, h, login, recordID, map[string]any{
		"view_schema_id": artifacts.NotesViewSchemaID, "base_row_version": 2, "client_txn_id": "txn-patch-later-write",
		"changes": []map[string]any{{"field_key": "note.title", "value": "Later title"}},
	})
	before := snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID)
	replay := requireWorkbookPatch(t, h, login, recordID, body)
	if !reflect.DeepEqual(patched, replay) {
		t.Fatal("patch replay changed the original committed result")
	}
	requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID))
	body["changes"] = []map[string]any{{"field_key": "note.body", "value": "Divergent body"}}
	before = snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID)
	response := doWorkbookJSON(t, h, login, http.MethodPatch, uuid.Nil, "", recordID, body)
	httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "client_txn_conflict")
	requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incidentID, artifacts.NotesViewSchemaID))
}

func requireWorkbookNoteFields(t testing.TB, row map[string]any) {
	t.Helper()
	var actual []string
	for key := range row["cells"].(map[string]any) {
		actual = append(actual, key)
	}
	if len(actual) == 0 {
		t.Fatal("note row has no observed cells")
	}
	sort.Strings(actual)
	contractassert.RequireFieldKeyConformance(t, actual, workbookscenariotest.AllowedFieldKeys(t, "workbook-note-response", artifacts.NotesViewSchemaID))
}

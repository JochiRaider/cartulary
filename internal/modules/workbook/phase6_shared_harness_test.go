package workbook_test

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
	"github.com/JochiRaider/cartulary/internal/testutil/phase6test"
)

func TestSupportPhase6SharedHarness_WorkbookRouteInventoryCoverage(t *testing.T) {
	inventory := phase6test.Phase6WorkbookRouteInventory()
	phase6test.RequireSharedHarnessInventory(t, inventory)

	required := phase6test.RequiredHarnessIDs(inventory)
	for _, harness := range []phase6test.SharedHarnessID{
		phase6test.HarnessEnvelopeConsistency,
		phase6test.HarnessAuthorizationRederived,
		phase6test.HarnessDivergentReplay,
		phase6test.HarnessClosedVocabulary,
		phase6test.HarnessWritableStringNormalize,
		phase6test.HarnessFieldKeyConformance,
		phase6test.HarnessProjectionRebuild,
		phase6test.HarnessWebSocketLifecycle,
		phase6test.HarnessTopologyAuditSource,
	} {
		if !slices.Contains(required, harness) {
			t.Fatalf("workbook route inventory must require %s, got %v", harness, required)
		}
	}
}

func TestSupportPhase6SharedHarness_WorkbookRouteConformance(t *testing.T) {
	harness, login, actorID, incidentID := phase6ConflictFixture(t, "phase6-support-shared-workbook-routes", "IR-PHASE6-SUPPORT-WORKBOOK")
	allowedNoteFields := phase4test.AllowedFieldKeys(t, "phase6-support-shared-workbook-routes", phase6NotesViewSchemaID)

	createTxnID := "txn-phase6-support-create"
	createResp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, phase6NotesViewSchemaID, uuid.Nil, map[string]any{
		"client_txn_id": createTxnID,
		"note.title":    "  Shared harness note  ",
		"note.body":     "Shared body",
	})
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	createRow := createData["row"].(map[string]any)
	recordID := phase4test.MustUUID(t, createRow["record_id"].(string))
	httptestx.RequireWritableStringNormalization(t, cellStringValue(t, createRow, "note.title"), "Shared harness note")
	httptestx.RequireFieldKeyConformance(t, []string{"note.body", "note.title"}, allowedNoteFields)

	stableBeforeCreateReplay := phase4test.SnapshotReplayCounts(t, harness.DB, incidentID.String(), recordID.String())
	createReplayResp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, phase6NotesViewSchemaID, uuid.Nil, map[string]any{
		"client_txn_id": createTxnID,
		"note.title":    "  Shared harness note  ",
		"note.body":     "Shared body",
	})
	httptestx.RequireSuccessEnvelope(t, createReplayResp, http.StatusOK)
	stableAfterCreateReplay := phase4test.SnapshotReplayCounts(t, harness.DB, incidentID.String(), recordID.String())
	httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
		FirstStatus:     http.StatusCreated,
		ReplayStatus:    http.StatusOK,
		DivergentStatus: http.StatusConflict,
		DivergentCode:   "client_txn_conflict",
		StableBefore:    stableBeforeCreateReplay,
		StableAfter:     stableAfterCreateReplay,
	})
	createDivergentResp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, phase6NotesViewSchemaID, uuid.Nil, map[string]any{
		"client_txn_id": createTxnID,
		"note.title":    "Divergent shared harness note",
		"note.body":     "Shared body",
	})
	createDivergentBody := httptestx.RequireErrorEnvelope(t, createDivergentResp, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, createDivergentResp.StatusCode, createDivergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

	patchTxnID := "txn-phase6-support-patch"
	patchResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    patchTxnID,
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "  Patched shared body  ",
		}},
	})
	patchData := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	patchRow := patchData["row"].(map[string]any)
	httptestx.RequireWritableStringNormalization(t, cellStringValue(t, patchRow, "note.body"), "Patched shared body")
	httptestx.RequireFieldKeyConformance(t, []string{"note.body"}, allowedNoteFields)
	phase6RequireMutationChangedFields(t, harness, patchData["change_set_id"].(string), []string{"note.body"})

	stableBeforePatchReplay := phase4test.SnapshotReplayCounts(t, harness.DB, incidentID.String(), recordID.String())
	patchReplayResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    patchTxnID,
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "  Patched shared body  ",
		}},
	})
	httptestx.RequireSuccessEnvelope(t, patchReplayResp, http.StatusOK)
	stableAfterPatchReplay := phase4test.SnapshotReplayCounts(t, harness.DB, incidentID.String(), recordID.String())
	if stableBeforePatchReplay != stableAfterPatchReplay {
		t.Fatalf("patch replay changed durable counts: before=%+v after=%+v", stableBeforePatchReplay, stableAfterPatchReplay)
	}
	patchDivergentResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    patchTxnID,
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "Divergent shared body",
		}},
	})
	patchDivergentBody := httptestx.RequireErrorEnvelope(t, patchDivergentResp, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, patchDivergentResp.StatusCode, patchDivergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

	phase6RequireWorkbookAuthorizationRederived(t, harness, login, actorID, incidentID)
	phase6RequireConflictResolveSharedHarness(t, harness, login, incidentID, allowedNoteFields)
}

func phase6RequireWorkbookAuthorizationRederived(t testing.TB, harness *phase4test.ServerHarness, adminLogin phase4test.LoginResult, adminUserID uuid.UUID, incidentID uuid.UUID) {
	t.Helper()

	editor := phase4test.SeedLocalUserFlags(t, harness.DB, "phase6-shared-editor@example.test", "Phase 6 Shared Editor", "Phase6SharedEditor1!", false, false, true)
	phase4test.SeedIncidentMembership(t, harness.DB, incidentID, editor.ID, editor.DisplayName, "editor", adminUserID)
	editorLogin := phase6LoginLocalUserNoMFA(t, harness, editor.Email, "Phase6SharedEditor1!")
	authRow := phase6CreateNote(t, harness, adminLogin, incidentID, "txn-phase6-support-auth-create", "Authorization row", "Authorization body")
	authRecordID := phase4test.MustUUID(t, authRow["record_id"].(string))

	beforeResp := doWorkbookJSON(t, harness, editorLogin, http.MethodPatch, uuid.Nil, "", authRecordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-support-auth-before",
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "Editor write before demotion",
		}},
	})
	httptestx.RequireSuccessEnvelope(t, beforeResp, http.StatusOK)

	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       membership_version = membership_version + 1,
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, editor.ID, adminUserID); err != nil {
		t.Fatalf("demote shared harness editor: %v", err)
	}

	afterRow := phase6CreateNote(t, harness, adminLogin, incidentID, "txn-phase6-support-auth-after-create", "Authorization after row", "Authorization body")
	afterRecordID := phase4test.MustUUID(t, afterRow["record_id"].(string))
	afterResp := doWorkbookJSON(t, harness, editorLogin, http.MethodPatch, uuid.Nil, "", afterRecordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-support-auth-after",
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "Viewer write after demotion",
		}},
	})
	afterBody := httptestx.RequireErrorEnvelope(t, afterResp, http.StatusForbidden, "authorization_denied")
	httptestx.RequireAuthorizationReDerived(t, httptestx.AuthorizationOutcome{Status: beforeResp.StatusCode}, httptestx.AuthorizationOutcome{
		Status: afterResp.StatusCode,
		Code:   afterBody["error"].(map[string]any)["code"].(string),
	})
}

func phase6RequireConflictResolveSharedHarness(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, allowedFieldKeys []string) {
	t.Helper()

	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-support-resolve-create", "Resolve base", "Resolve body")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-support-resolve-server",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Resolve server",
		}},
	})
	conflictResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-support-resolve-client",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Resolve client",
		}},
	})
	conflictBody := httptestx.RequireErrorEnvelope(t, conflictResp, http.StatusConflict, "same_field_conflict")
	conflict := conflictBody["error"].(map[string]any)["conflict"].(map[string]any)
	httptestx.RequireFieldKeyConformance(t, []string{conflict["field_key"].(string)}, allowedFieldKeys)
	conflictToken := conflict["conflict_token"].(string)

	invalidResp := phase6ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "unsupported_resolution",
		"client_txn_id":   "txn-phase6-support-resolve-invalid-kind",
	})
	invalidBody := httptestx.RequireErrorEnvelope(t, invalidResp, http.StatusBadRequest, "invalid_mutation_payload")
	httptestx.RequireClosedVocabularyRejected(t, invalidBody["error"].(map[string]any)["code"].(string), httptestx.RequireErrorDetails(t, invalidBody), "resolution_kind", "")

	resolveTxnID := "txn-phase6-support-resolve"
	resolveResp := phase6ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "merged_value",
		"client_txn_id":   resolveTxnID,
		"resolved_value":  "Resolve merged",
	})
	resolveData := httptestx.RequireSuccessEnvelope(t, resolveResp, http.StatusOK)["data"].(map[string]any)
	phase4test.RequireChangeSetAttribution(t, harness.DB, resolveData["change_set_id"].(string), "", "workbook.records.conflicts.resolve", resolveTxnID)
	httptestx.RequireWritableStringNormalization(t, cellStringValue(t, resolveData["row"].(map[string]any), "note.title"), "Resolve merged")

	stableBeforeResolveReplay := phase4test.SnapshotReplayCounts(t, harness.DB, incidentID.String(), recordID.String())
	resolveReplayResp := phase6ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "merged_value",
		"client_txn_id":   resolveTxnID,
		"resolved_value":  "Resolve merged",
	})
	httptestx.RequireSuccessEnvelope(t, resolveReplayResp, http.StatusOK)
	stableAfterResolveReplay := phase4test.SnapshotReplayCounts(t, harness.DB, incidentID.String(), recordID.String())
	if stableBeforeResolveReplay != stableAfterResolveReplay {
		t.Fatalf("resolve replay changed durable counts: before=%+v after=%+v", stableBeforeResolveReplay, stableAfterResolveReplay)
	}

	resolveDivergentResp := phase6ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "merged_value",
		"client_txn_id":   resolveTxnID,
		"resolved_value":  "Resolve divergent",
	})
	resolveDivergentBody := httptestx.RequireErrorEnvelope(t, resolveDivergentResp, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, resolveDivergentResp.StatusCode, resolveDivergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")
}

func phase6ResolveConflictRaw(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, recordID uuid.UUID, conflictToken string, body map[string]any) *http.Response {
	t.Helper()
	return phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/conflicts/"+conflictToken+"/resolve",
		body,
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func cellStringValue(t testing.TB, row map[string]any, fieldKey string) string {
	t.Helper()
	cells := row["cells"].(map[string]any)
	value, ok := cells[fieldKey].(map[string]any)["value"].(string)
	if !ok {
		t.Fatalf("expected %s string cell value, got %#v", fieldKey, cells[fieldKey])
	}
	return value
}

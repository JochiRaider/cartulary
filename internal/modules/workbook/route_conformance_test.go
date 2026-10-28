package workbook_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestWorkbookStartupConstructionFailure(t *testing.T) {
	constructionErr := errors.New("startup construction failed")
	registrar := workbook.RegisterRoutes(workbook.RouteDependencies{
		StartupStoreFactory: func(httpapi.DependencySet) (*workbookstartup.Store, error) {
			return nil, constructionErr
		},
	})
	err := registrar(http.NewServeMux(), httpapi.DependencySet{})
	if !errors.Is(err, constructionErr) {
		t.Fatalf("route composition did not propagate startup construction error: %v", err)
	}
}

func TestWorkbookRouteConformance(t *testing.T) {
	harness, login, actorID, incidentID := ConflictFixture(t, "collaboration-support-shared-workbook-routes", "IR-COLLABORATION-SUPPORT-WORKBOOK")
	allowedNoteFields := workbookscenariotest.AllowedFieldKeys(t, "collaboration-support-shared-workbook-routes", artifacts.NotesViewSchemaID)

	createTxnID := "txn-collaboration-support-create"
	createResp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, artifacts.NotesViewSchemaID, uuid.Nil, map[string]any{
		"client_txn_id": createTxnID,
		"note.title":    "  Shared harness note  ",
		"note.body":     "Shared body",
	})
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	createRow := createData["row"].(map[string]any)
	recordID := appsupport.MustUUID(t, createRow["record_id"].(string))
	contractassert.RequireWritableStringNormalization(t, cellStringValue(t, createRow, "note.title"), "Shared harness note")
	contractassert.RequireFieldKeyConformance(t, []string{"note.body", "note.title"}, allowedNoteFields)

	stableBeforeCreateReplay := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	sourceBeforeCreateReplay := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID)
	createReplayResp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, artifacts.NotesViewSchemaID, uuid.Nil, map[string]any{
		"client_txn_id": createTxnID,
		"note.title":    "  Shared harness note  ",
		"note.body":     "Shared body",
	})
	createReplayData := httptestx.RequireSuccessEnvelope(t, createReplayResp, http.StatusOK)["data"].(map[string]any)
	if createReplayData["change_set_id"] != createData["change_set_id"] || !reflect.DeepEqual(createReplayData["row"], createData["row"]) {
		t.Fatalf("create replay changed stored result")
	}
	stableAfterCreateReplay := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID); !reflect.DeepEqual(sourceBeforeCreateReplay, got) {
		t.Fatal("create replay or divergence changed source")
	}
	if stableBeforeCreateReplay != stableAfterCreateReplay {
		t.Fatalf("create replay changed mutation effects: %+v -> %+v", stableBeforeCreateReplay, stableAfterCreateReplay)
	}
	createDivergentResp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, artifacts.NotesViewSchemaID, uuid.Nil, map[string]any{
		"client_txn_id": createTxnID,
		"note.title":    "Divergent shared harness note",
		"note.body":     "Shared body",
	})
	httptestx.RequireErrorEnvelope(t, createDivergentResp, http.StatusConflict, "client_txn_conflict")
	if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID); !reflect.DeepEqual(sourceBeforeCreateReplay, got) {
		t.Fatal("create replay or divergence changed source")
	}
	if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != stableBeforeCreateReplay {
		t.Fatalf("create divergence changed effects: %+v -> %+v", stableBeforeCreateReplay, got)
	}

	patchTxnID := "txn-collaboration-support-patch"
	patchResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    patchTxnID,
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "  Patched shared body  ",
		}},
	})
	patchData := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	patchRow := patchData["row"].(map[string]any)
	contractassert.RequireWritableStringNormalization(t, cellStringValue(t, patchRow, "note.body"), "Patched shared body")
	contractassert.RequireFieldKeyConformance(t, []string{"note.body"}, allowedNoteFields)
	RequireMutationChangedFields(t, harness, patchData["change_set_id"].(string), []string{"note.body"})

	stableBeforePatchReplay := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	sourceBeforePatchReplay := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID)
	patchReplayResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    patchTxnID,
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "  Patched shared body  ",
		}},
	})
	patchReplayData := httptestx.RequireSuccessEnvelope(t, patchReplayResp, http.StatusOK)["data"].(map[string]any)
	if patchReplayData["change_set_id"] != patchData["change_set_id"] || !reflect.DeepEqual(patchReplayData["row"], patchData["row"]) {
		t.Fatalf("patch replay changed stored result")
	}
	stableAfterPatchReplay := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID); !reflect.DeepEqual(sourceBeforePatchReplay, got) {
		t.Fatal("patch replay or divergence changed source")
	}
	if stableBeforePatchReplay != stableAfterPatchReplay {
		t.Fatalf("patch replay changed mutation effects: before=%+v after=%+v", stableBeforePatchReplay, stableAfterPatchReplay)
	}
	patchDivergentResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    patchTxnID,
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "Divergent shared body",
		}},
	})
	httptestx.RequireErrorEnvelope(t, patchDivergentResp, http.StatusConflict, "client_txn_conflict")
	if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID); !reflect.DeepEqual(sourceBeforePatchReplay, got) {
		t.Fatal("patch replay or divergence changed source")
	}
	if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != stableBeforePatchReplay {
		t.Fatalf("patch divergence changed effects: %+v -> %+v", stableBeforePatchReplay, got)
	}

	RequireWorkbookAuthorizationRederived(t, harness, login, actorID, incidentID)
	RequireConflictResolveSharedHarness(t, harness, login, incidentID, allowedNoteFields)
}

func RequireWorkbookAuthorizationRederived(t testing.TB, harness *appsupport.ServerHarness, adminLogin appsupport.LoginResult, adminUserID uuid.UUID, incidentID uuid.UUID) {
	t.Helper()

	editor := authflowtest.SeedLocalUserRecord(t, harness.DB, "collaboration-shared-editor@example.test", "Collaboration Shared Editor", "CollaborationSharedEditor1!", false, false, true)
	incidentstoretest.SeedMembership(t, harness.DB, incidentID, editor.ID, editor.DisplayName, "editor", adminUserID)
	editorLogin := LoginLocalUserNoMFA(t, harness, editor.Email, "CollaborationSharedEditor1!")
	authRow := CreateNote(t, harness, adminLogin, incidentID, "txn-collaboration-support-auth-create", "Authorization row", "Authorization body")
	authRecordID := appsupport.MustUUID(t, authRow["record_id"].(string))

	beforeResp := doWorkbookJSON(t, harness, editorLogin, http.MethodPatch, uuid.Nil, "", authRecordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-support-auth-before",
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

	afterRow := CreateNote(t, harness, adminLogin, incidentID, "txn-collaboration-support-auth-after-create", "Authorization after row", "Authorization body")
	afterRecordID := appsupport.MustUUID(t, afterRow["record_id"].(string))
	beforeDeniedEffects := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	sourceBeforeDenial := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), afterRecordID)
	beforeDeniedRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), artifacts.NotesViewSchemaID, adminLogin), afterRecordID.String())
	afterResp := doWorkbookJSON(t, harness, editorLogin, http.MethodPatch, uuid.Nil, "", afterRecordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-support-auth-after",
		"changes": []map[string]any{{
			"field_key": "note.body",
			"value":     "Viewer write after demotion",
		}},
	})
	httptestx.RequireErrorEnvelope(t, afterResp, http.StatusForbidden, "authorization_denied")
	if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), afterRecordID); !reflect.DeepEqual(sourceBeforeDenial, got) {
		t.Fatal("denied note patch changed source")
	}
	if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != beforeDeniedEffects {
		t.Fatalf("denial changed mutation effects: %+v -> %+v", beforeDeniedEffects, got)
	}
	afterDeniedRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), artifacts.NotesViewSchemaID, adminLogin), afterRecordID.String())
	if !reflect.DeepEqual(beforeDeniedRow, afterDeniedRow) {
		t.Fatal("denial changed note source projection")
	}

}

func RequireConflictResolveSharedHarness(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, allowedFieldKeys []string) {
	t.Helper()

	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-support-resolve-create", "Resolve base", "Resolve body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-support-resolve-server",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Resolve server",
		}},
	})
	conflictResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-support-resolve-client",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Resolve client",
		}},
	})
	conflictBody := httptestx.RequireErrorEnvelope(t, conflictResp, http.StatusConflict, "same_field_conflict")
	conflict := conflictBody["error"].(map[string]any)["conflict"].(map[string]any)
	contractassert.RequireFieldKeyConformance(t, []string{conflict["field_key"].(string)}, allowedFieldKeys)
	conflictToken := conflict["conflict_token"].(string)

	invalidResp := ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "unsupported_resolution",
		"client_txn_id":   "txn-collaboration-support-resolve-invalid-kind",
	})
	invalidBody := httptestx.RequireErrorEnvelope(t, invalidResp, http.StatusBadRequest, "invalid_mutation_payload")
	contractassert.RequireClosedVocabularyRejected(t, invalidBody["error"].(map[string]any)["code"].(string), httptestx.RequireErrorDetails(t, invalidBody), "resolution_kind", "")

	resolveTxnID := "txn-collaboration-support-resolve"
	resolveResp := ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "merged_value",
		"client_txn_id":   resolveTxnID,
		"resolved_value":  "Resolve merged",
	})
	resolveData := httptestx.RequireSuccessEnvelope(t, resolveResp, http.StatusOK)["data"].(map[string]any)
	workbookscenariotest.RequireChangeSetAttribution(t, harness.DB, resolveData["change_set_id"].(string), "", "workbook.records.conflicts.resolve", resolveTxnID)
	contractassert.RequireWritableStringNormalization(t, cellStringValue(t, resolveData["row"].(map[string]any), "note.title"), "Resolve merged")

	stableBeforeResolveReplay := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	sourceBeforeResolveReplay := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID)
	resolveReplayResp := ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "merged_value",
		"client_txn_id":   resolveTxnID,
		"resolved_value":  "Resolve merged",
	})
	resolveReplayData := httptestx.RequireSuccessEnvelope(t, resolveReplayResp, http.StatusOK)["data"].(map[string]any)
	if resolveReplayData["change_set_id"] != resolveData["change_set_id"] || !reflect.DeepEqual(resolveReplayData["row"], resolveData["row"]) {
		t.Fatalf("resolve replay changed stored result")
	}
	stableAfterResolveReplay := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID); !reflect.DeepEqual(sourceBeforeResolveReplay, got) {
		t.Fatal("resolve replay or divergence changed source")
	}
	if stableBeforeResolveReplay != stableAfterResolveReplay {
		t.Fatalf("resolve replay changed mutation effects: before=%+v after=%+v", stableBeforeResolveReplay, stableAfterResolveReplay)
	}

	resolveDivergentResp := ResolveConflictRaw(t, harness, login, recordID, conflictToken, map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "merged_value",
		"client_txn_id":   resolveTxnID,
		"resolved_value":  "Resolve divergent",
	})
	httptestx.RequireErrorEnvelope(t, resolveDivergentResp, http.StatusConflict, "client_txn_conflict")
	if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), recordID); !reflect.DeepEqual(sourceBeforeResolveReplay, got) {
		t.Fatal("resolve replay or divergence changed source")
	}
	if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != stableBeforeResolveReplay {
		t.Fatalf("resolve divergence changed effects: %+v -> %+v", stableBeforeResolveReplay, got)
	}

}

func ResolveConflictRaw(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, conflictToken string, body map[string]any) *http.Response {
	t.Helper()
	return appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/conflicts/"+conflictToken+"/resolve",
		body,
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
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

package workbook_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbookRouteGuardsFailBeforeMutation(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase9-workbook-route-guards")
	adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-incident",
		"incident_key":  "IR-PHASE9-ROUTE-GUARDS",
		"title":         "Workbook inspector workbook route guards",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.notes.v1/rows"
	socket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), "cartulary.view.notes.v1", adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")

	beforeAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	body := map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-auth",
		"note.title":    "guarded note",
	}

	unauthenticated := workbookscenariotest.DoJSON(t, http.MethodPost, createURL, body)
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
	)
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	invalidCSRF := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		workbookscenariotest.WithHeader(authn.CSRFHeaderName, "wrong-csrf-token"),
	)
	httptestx.RequireErrorEnvelope(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

	afterAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if afterAuth != beforeAuth {
		t.Fatalf("auth/csrf failures mutated workbook state: before=%+v after=%+v", beforeAuth, afterAuth)
	}
	workbookscenariotest.ExpectNoSocketMessage(t, socket)

	noteData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1", map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-note",
		"note.title":    "closed incident guard note",
	})
	recordID := workbookscenariotest.MustUUID(t, noteData["row"].(map[string]any)["record_id"].(string))
	requireConflictResolveSecurityPrecedence(t, harness, adminLogin, adminUserID, incidentID)
	closedSocket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), "cartulary.view.notes.v1", adminLogin.SessionCookie.Value)
	defer closedSocket.Close(1000, "test_complete")

	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incidents
   SET status = 'closed',
       closed_at = $1
 WHERE id = $2
`, time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC), incidentID); err != nil {
		t.Fatalf("close incident: %v", err)
	}

	beforeClosed := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	closedCreate := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.notes.v1", uuid.Nil, map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-closed-create",
		"note.title":    "blocked after close",
	})
	httptestx.RequireErrorEnvelope(t, closedCreate, http.StatusConflict, "incident_closed")

	closedPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-workbook-route-guards-closed-patch",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "blocked patch after close",
		}},
	})
	httptestx.RequireErrorEnvelope(t, closedPatch, http.StatusConflict, "incident_closed")

	afterClosed := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if afterClosed != beforeClosed {
		t.Fatalf("closed incident failures mutated workbook state: before=%+v after=%+v", beforeClosed, afterClosed)
	}
	workbookscenariotest.ExpectNoSocketMessage(t, closedSocket)
}

func requireConflictResolveSecurityPrecedence(t testing.TB, harness *workbookscenariotest.ServerHarness, adminLogin workbookscenariotest.LoginResult, adminUserID uuid.UUID, incidentID uuid.UUID) {
	t.Helper()

	note := CreateNote(t, harness, adminLogin, incidentID, "txn-phase9-conflict-guard-create", "Conflict guard", "Conflict guard body")
	recordID := workbookscenariotest.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, adminLogin, recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-conflict-guard-server",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Conflict guard server",
		}},
	})
	conflictResponse := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-conflict-guard-client",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Conflict guard client",
		}},
	})
	conflictBody := httptestx.RequireErrorEnvelope(t, conflictResponse, http.StatusConflict, "same_field_conflict")
	conflictToken := conflictBody["error"].(map[string]any)["conflict"].(map[string]any)["conflict_token"].(string)
	validBody := map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-phase9-conflict-guard-resolution",
	}
	resolveURL := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String() + "/conflicts/" + conflictToken + "/resolve"
	invalidTokenURL := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String() + "/conflicts/not-a-token/resolve"

	viewer := workbookscenariotest.SeedLocalUserFlags(t, harness.DB, "phase9-conflict-viewer@example.test", "Workbook inspector Conflict Viewer", "Phase9ConflictViewer1!", false, false, true)
	workbookscenariotest.SeedIncidentMembership(t, harness.DB, incidentID, viewer.ID, viewer.DisplayName, "viewer", adminUserID)
	viewerLogin := LoginLocalUserNoMFA(t, harness, viewer.Email, "Phase9ConflictViewer1!")
	nonMember := workbookscenariotest.SeedLocalUserFlags(t, harness.DB, "phase9-conflict-nonmember@example.test", "Workbook inspector Conflict Nonmember", "Phase9ConflictNonmember1!", false, false, true)
	nonMemberLogin := LoginLocalUserNoMFA(t, harness, nonMember.Email, "Phase9ConflictNonmember1!")

	otherNote := CreateNote(t, harness, adminLogin, incidentID, "txn-phase9-conflict-guard-other-record", "Other record", "Other body")
	otherRecordID := workbookscenariotest.MustUUID(t, otherNote["record_id"].(string))
	wrongRecordURL := harness.Server.HTTP.URL + "/api/v1/records/" + otherRecordID.String() + "/conflicts/" + conflictToken + "/resolve"
	otherIncident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-conflict-guard-other-incident",
		"incident_key":  "IR-PHASE9-CONFLICT-OTHER",
		"title":         "Workbook inspector conflict other incident",
	})
	otherIncidentID := workbookscenariotest.MustUUID(t, otherIncident["incident_id"].(string))
	otherIncidentNote := CreateNote(t, harness, adminLogin, otherIncidentID, "txn-phase9-conflict-guard-cross-incident-record", "Cross incident record", "Cross incident body")
	otherIncidentRecordID := workbookscenariotest.MustUUID(t, otherIncidentNote["record_id"].(string))
	crossIncidentURL := harness.Server.HTTP.URL + "/api/v1/records/" + otherIncidentRecordID.String() + "/conflicts/" + conflictToken + "/resolve"
	missingRecordURL := harness.Server.HTTP.URL + "/api/v1/records/00000000-0000-4000-8000-000000009999/conflicts/not-a-token/resolve"

	socket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), NotesViewSchemaID, adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")
	before := snapshotWorkbookRouteGuardState(t, harness, incidentID)

	unauthenticated := workbookscenariotest.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/not-a-uuid/conflicts/not-a-token/resolve", map[string]any{})
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := workbookscenariotest.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie))
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	missingRecord := workbookscenariotest.DoJSON(t, http.MethodPost, missingRecordURL, map[string]any{}, workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, missingRecord, http.StatusNotFound, "incident_not_found")

	nonMemberResponse := workbookscenariotest.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, workbookscenariotest.WithCookies(nonMemberLogin.SessionCookie, nonMemberLogin.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, nonMemberLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, nonMemberResponse, http.StatusNotFound, "incident_not_found")

	viewerResponse := workbookscenariotest.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, workbookscenariotest.WithCookies(viewerLogin.SessionCookie, viewerLogin.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, viewerResponse, http.StatusForbidden, "authorization_denied")

	invalidToken := workbookscenariotest.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	invalidTokenBody := httptestx.RequireErrorEnvelope(t, invalidToken, http.StatusBadRequest, "invalid_mutation_payload")
	if httptestx.RequireErrorDetails(t, invalidTokenBody)["field"] != "conflict_token" {
		t.Fatalf("authorized invalid token must identify conflict_token: %#v", invalidTokenBody)
	}

	wrongRecord := workbookscenariotest.DoJSON(t, http.MethodPost, wrongRecordURL, validBody, workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, wrongRecord, http.StatusBadRequest, "invalid_mutation_payload")

	crossIncident := workbookscenariotest.DoJSON(t, http.MethodPost, crossIncidentURL, validBody, workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, crossIncident, http.StatusBadRequest, "invalid_mutation_payload")

	invalidBody := workbookscenariotest.DoJSON(t, http.MethodPost, resolveURL, map[string]any{
		"conflict_token":  "different-token",
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-phase9-conflict-guard-invalid-body",
	}, workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, invalidBody, http.StatusBadRequest, "invalid_mutation_payload")

	after := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if after != before {
		t.Fatalf("rejected conflict precedence matrix mutated workbook state: before=%+v after=%+v", before, after)
	}
	workbookscenariotest.ExpectNoSocketMessage(t, socket)
}

type workbookRouteGuardState struct {
	Records         int
	Artifacts       int
	ChangeSets      int
	Mutations       int
	RecordRevisions int
}

func snapshotWorkbookRouteGuardState(t testing.TB, harness *workbookscenariotest.ServerHarness, incidentID uuid.UUID) workbookRouteGuardState {
	t.Helper()
	return workbookRouteGuardState{
		Records: workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM records
 WHERE incident_id = $1
`, incidentID),
		Artifacts: workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM artifacts
 WHERE incident_id = $1
`, incidentID),
		ChangeSets: workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_sets
 WHERE incident_id = $1
`, incidentID),
		Mutations: workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		RecordRevisions: workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_revisions rr
  JOIN records r ON r.record_id = rr.record_id
 WHERE r.incident_id = $1
`, incidentID),
	}
}

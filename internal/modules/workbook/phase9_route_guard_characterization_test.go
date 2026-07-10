package workbook_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestSupportPhase9_WorkbookRouteGuardsFailBeforeMutation(t *testing.T) {
	harness := phase4test.StartServer(t, "phase9-workbook-route-guards")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-incident",
		"incident_key":  "IR-PHASE9-ROUTE-GUARDS",
		"title":         "Phase 9 workbook route guards",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.notes.v1/rows"
	socket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), "cartulary.view.notes.v1", adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")

	beforeAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	body := map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-auth",
		"note.title":    "guarded note",
	}

	unauthenticated := phase4test.DoJSON(t, http.MethodPost, createURL, body)
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := phase4test.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
	)
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	invalidCSRF := phase4test.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, "wrong-csrf-token"),
	)
	httptestx.RequireErrorEnvelope(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

	afterAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if afterAuth != beforeAuth {
		t.Fatalf("auth/csrf failures mutated workbook state: before=%+v after=%+v", beforeAuth, afterAuth)
	}
	phase4test.ExpectNoSocketMessage(t, socket)

	noteData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1", map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-note",
		"note.title":    "closed incident guard note",
	})
	recordID := phase4test.MustUUID(t, noteData["row"].(map[string]any)["record_id"].(string))
	requireConflictResolveSecurityPrecedence(t, harness, adminLogin, adminUserID, incidentID)
	closedSocket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), "cartulary.view.notes.v1", adminLogin.SessionCookie.Value)
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
	phase4test.ExpectNoSocketMessage(t, closedSocket)
}

func requireConflictResolveSecurityPrecedence(t testing.TB, harness *phase4test.ServerHarness, adminLogin phase4test.LoginResult, adminUserID uuid.UUID, incidentID uuid.UUID) {
	t.Helper()

	note := phase6CreateNote(t, harness, adminLogin, incidentID, "txn-phase9-conflict-guard-create", "Conflict guard", "Conflict guard body")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, adminLogin, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-conflict-guard-server",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Conflict guard server",
		}},
	})
	conflictResponse := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
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

	viewer := phase4test.SeedLocalUserFlags(t, harness.DB, "phase9-conflict-viewer@example.test", "Phase 9 Conflict Viewer", "Phase9ConflictViewer1!", false, false, true)
	phase4test.SeedIncidentMembership(t, harness.DB, incidentID, viewer.ID, viewer.DisplayName, "viewer", adminUserID)
	viewerLogin := phase6LoginLocalUserNoMFA(t, harness, viewer.Email, "Phase9ConflictViewer1!")
	nonMember := phase4test.SeedLocalUserFlags(t, harness.DB, "phase9-conflict-nonmember@example.test", "Phase 9 Conflict Nonmember", "Phase9ConflictNonmember1!", false, false, true)
	nonMemberLogin := phase6LoginLocalUserNoMFA(t, harness, nonMember.Email, "Phase9ConflictNonmember1!")

	otherNote := phase6CreateNote(t, harness, adminLogin, incidentID, "txn-phase9-conflict-guard-other-record", "Other record", "Other body")
	otherRecordID := phase4test.MustUUID(t, otherNote["record_id"].(string))
	wrongRecordURL := harness.Server.HTTP.URL + "/api/v1/records/" + otherRecordID.String() + "/conflicts/" + conflictToken + "/resolve"
	otherIncident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-conflict-guard-other-incident",
		"incident_key":  "IR-PHASE9-CONFLICT-OTHER",
		"title":         "Phase 9 conflict other incident",
	})
	otherIncidentID := phase4test.MustUUID(t, otherIncident["incident_id"].(string))
	otherIncidentNote := phase6CreateNote(t, harness, adminLogin, otherIncidentID, "txn-phase9-conflict-guard-cross-incident-record", "Cross incident record", "Cross incident body")
	otherIncidentRecordID := phase4test.MustUUID(t, otherIncidentNote["record_id"].(string))
	crossIncidentURL := harness.Server.HTTP.URL + "/api/v1/records/" + otherIncidentRecordID.String() + "/conflicts/" + conflictToken + "/resolve"
	missingRecordURL := harness.Server.HTTP.URL + "/api/v1/records/00000000-0000-4000-8000-000000009999/conflicts/not-a-token/resolve"

	socket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), phase6NotesViewSchemaID, adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")
	before := snapshotWorkbookRouteGuardState(t, harness, incidentID)

	unauthenticated := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/not-a-uuid/conflicts/not-a-token/resolve", map[string]any{})
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := phase4test.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie))
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	missingRecord := phase4test.DoJSON(t, http.MethodPost, missingRecordURL, map[string]any{}, phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, missingRecord, http.StatusNotFound, "incident_not_found")

	nonMemberResponse := phase4test.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, phase4test.WithCookies(nonMemberLogin.SessionCookie, nonMemberLogin.CSRFCookie), phase4test.WithHeader(authn.CSRFHeaderName, nonMemberLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, nonMemberResponse, http.StatusNotFound, "incident_not_found")

	viewerResponse := phase4test.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, phase4test.WithCookies(viewerLogin.SessionCookie, viewerLogin.CSRFCookie), phase4test.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, viewerResponse, http.StatusForbidden, "authorization_denied")

	invalidToken := phase4test.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	invalidTokenBody := httptestx.RequireErrorEnvelope(t, invalidToken, http.StatusBadRequest, "invalid_mutation_payload")
	if httptestx.RequireErrorDetails(t, invalidTokenBody)["field"] != "conflict_token" {
		t.Fatalf("authorized invalid token must identify conflict_token: %#v", invalidTokenBody)
	}

	wrongRecord := phase4test.DoJSON(t, http.MethodPost, wrongRecordURL, validBody, phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, wrongRecord, http.StatusBadRequest, "invalid_mutation_payload")

	crossIncident := phase4test.DoJSON(t, http.MethodPost, crossIncidentURL, validBody, phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, crossIncident, http.StatusBadRequest, "invalid_mutation_payload")

	invalidBody := phase4test.DoJSON(t, http.MethodPost, resolveURL, map[string]any{
		"conflict_token":  "different-token",
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-phase9-conflict-guard-invalid-body",
	}, phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, invalidBody, http.StatusBadRequest, "invalid_mutation_payload")

	after := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if after != before {
		t.Fatalf("rejected conflict precedence matrix mutated workbook state: before=%+v after=%+v", before, after)
	}
	phase4test.ExpectNoSocketMessage(t, socket)
}

type workbookRouteGuardState struct {
	Records         int
	Artifacts       int
	ChangeSets      int
	Mutations       int
	RecordRevisions int
}

func snapshotWorkbookRouteGuardState(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID) workbookRouteGuardState {
	t.Helper()
	return workbookRouteGuardState{
		Records: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM records
 WHERE incident_id = $1
`, incidentID),
		Artifacts: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM artifacts
 WHERE incident_id = $1
`, incidentID),
		ChangeSets: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_sets
 WHERE incident_id = $1
`, incidentID),
		Mutations: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		RecordRevisions: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_revisions rr
  JOIN records r ON r.record_id = rr.record_id
 WHERE r.incident_id = $1
`, incidentID),
	}
}

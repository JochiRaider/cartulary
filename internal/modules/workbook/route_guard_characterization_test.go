package workbook_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbookRouteAuthentication_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "workbook-authentication", "IR-WORKBOOK-AUTHENTICATION")
	note := CreateNote(t, harness, login, incidentID, "txn-guard-note", "Guarded note", "Guarded body")
	incidentPath := "/api/v1/incidents/" + incidentID.String()
	recordPath := "/api/v1/records/" + note["record_id"].(string)
	cases := []struct{ name, method, path string }{
		{"applyWorkbookBulkMutation", "POST", incidentPath + "/views/cartulary.view.notes.v1/bulk-mutations"},
		{"createRecordLinkedNote", "POST", recordPath + "/linked-notes"},
		{"createViewRow", "POST", incidentPath + "/views/cartulary.view.notes.v1/rows"},
		{"getCurrentUserWorkbookPreferences", "GET", incidentPath + "/workbook-preferences/me"},
		{"getIncidentDefaultWorkbookPreferences", "GET", incidentPath + "/workbook-preferences/default"},
		{"getIncidentWorkbookStartup", "GET", incidentPath + "/workbook-startup"},
		{"listNoteAssociations", "GET", recordPath + "/note-associations"},
		{"mutateNoteAssociations", "POST", recordPath + "/note-associations"},
		{"patchRecord", "PATCH", recordPath},
		{"pasteWorkbookClipboard", "POST", incidentPath + "/views/cartulary.view.notes.v1/clipboard-paste"},
		{"putCurrentUserWorkbookPreferences", "PUT", incidentPath + "/workbook-preferences/me"},
		{"putIncidentDefaultWorkbookPreferences", "PUT", incidentPath + "/workbook-preferences/default"},
		{"queryWorkbookView", "POST", incidentPath + "/views/cartulary.view.notes.v1/query"},
		{"locateWorkbookViewRecord", "POST", incidentPath + "/views/cartulary.view.notes.v1/locate"},
		{"resolveRecordSameFieldConflict", "POST", recordPath + "/conflicts/not-a-token/resolve"},
		{"supersedeRecord", "POST", recordPath + "/supersede"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
			response := httptestx.DoJSON(t, http.DefaultClient, tc.method, harness.Server.HTTP.URL+tc.path, map[string]any{"client_txn_id": "txn-denied-guard", "note.title": "must not persist"})
			httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, "session_required")
			requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
		})
	}
}

func TestWorkbookRouteMutationCSRF_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "workbook-mutationcsrf", "IR-WORKBOOK-MUTATIONCSRF")
	note := CreateNote(t, harness, login, incidentID, "txn-guard-note", "Guarded note", "Guarded body")
	incidentPath := "/api/v1/incidents/" + incidentID.String()
	recordPath := "/api/v1/records/" + note["record_id"].(string)
	cases := []struct{ name, method, path string }{
		{"applyWorkbookBulkMutation", "POST", incidentPath + "/views/cartulary.view.notes.v1/bulk-mutations"},
		{"createRecordLinkedNote", "POST", recordPath + "/linked-notes"},
		{"createViewRow", "POST", incidentPath + "/views/cartulary.view.notes.v1/rows"},
		{"mutateNoteAssociations", "POST", recordPath + "/note-associations"},
		{"patchRecord", "PATCH", recordPath},
		{"pasteWorkbookClipboard", "POST", incidentPath + "/views/cartulary.view.notes.v1/clipboard-paste"},
		{"putCurrentUserWorkbookPreferences", "PUT", incidentPath + "/workbook-preferences/me"},
		{"putIncidentDefaultWorkbookPreferences", "PUT", incidentPath + "/workbook-preferences/default"},
		{"resolveRecordSameFieldConflict", "POST", recordPath + "/conflicts/not-a-token/resolve"},
		{"supersedeRecord", "POST", recordPath + "/supersede"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, token := range []string{"", "invalid-csrf-token"} {
				name := "missing"
				if token != "" {
					name = "invalid"
				}
				t.Run(name, func(t *testing.T) {
					options := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF)}
					if token != "" {
						options = append(options, httptestx.WithHeader(authn.CSRFHeaderName, token))
					}
					before := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
					response := httptestx.DoJSON(t, http.DefaultClient, tc.method, harness.Server.HTTP.URL+tc.path, map[string]any{"client_txn_id": "txn-denied-guard", "note.title": "must not persist"}, options...)
					httptestx.RequireErrorEnvelope(t, response, http.StatusForbidden, "csrf_verification_failed")
					requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
				})
			}
		})
	}
}

func TestWorkbookClosedIncidentWrites_Integration(t *testing.T) {
	harness, adminLogin, _, incidentID := ConflictFixture(t, "workbook-closed-writes", "IR-WORKBOOK-CLOSED")
	note := CreateNote(t, harness, adminLogin, incidentID, "txn-closed-note", "Closed note", "Closed body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incidents
   SET status = 'closed',
       closed_at = $1
 WHERE id = $2
`, time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC), incidentID); err != nil {
		t.Fatalf("close incident: %v", err)
	}

	beforeClosedCreateState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	closedCreate := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.notes.v1", uuid.Nil, map[string]any{
		"client_txn_id": "txn-workbook_interaction-workbook-route-guards-closed-create",
		"note.title":    "blocked after close",
	})
	httptestx.RequireErrorEnvelope(t, closedCreate, http.StatusConflict, "incident_closed")
	requireWorkbookStateUnchanged(t, beforeClosedCreateState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

	beforeClosedPatchState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	closedPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook_interaction-workbook-route-guards-closed-patch",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "blocked patch after close",
		}},
	})
	httptestx.RequireErrorEnvelope(t, closedPatch, http.StatusConflict, "incident_closed")
	requireWorkbookStateUnchanged(t, beforeClosedPatchState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

}

func TestWorkbookConflictSecurityPrecedence_Integration(t *testing.T) {
	harness, adminLogin, adminUserID, incidentID := ConflictFixture(t, "workbook-conflict-guards", "IR-WORKBOOK-CONFLICT-GUARDS")

	note := CreateNote(t, harness, adminLogin, incidentID, "txn-workbook_interaction-conflict-guard-create", "Conflict guard", "Conflict guard body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, adminLogin, recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook_interaction-conflict-guard-server",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Conflict guard server",
		}},
	})
	beforeConflictResponseState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	conflictResponse := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook_interaction-conflict-guard-client",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Conflict guard client",
		}},
	})
	conflictBody := httptestx.RequireErrorEnvelope(t, conflictResponse, http.StatusConflict, "same_field_conflict")
	requireWorkbookStateUnchanged(t, beforeConflictResponseState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))
	conflictToken := conflictBody["error"].(map[string]any)["conflict"].(map[string]any)["conflict_token"].(string)
	validBody := map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-workbook_interaction-conflict-guard-resolution",
	}
	resolveURL := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String() + "/conflicts/" + conflictToken + "/resolve"
	invalidTokenURL := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String() + "/conflicts/not-a-token/resolve"

	viewer := authflowtest.SeedLocalUserRecord(t, harness.DB, "workbook_interaction-conflict-viewer@example.test", "Workbook inspector Conflict Viewer", "WorkbookInteractionConflictViewer1!", false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, viewer.ID, "viewer", adminUserID)
	viewerLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, viewer.Email, "WorkbookInteractionConflictViewer1!", nil)
	nonMember := authflowtest.SeedLocalUserRecord(t, harness.DB, "workbook_interaction-conflict-nonmember@example.test", "Workbook inspector Conflict Nonmember", "WorkbookInteractionConflictNonmember1!", false, false, true)
	nonMemberLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, nonMember.Email, "WorkbookInteractionConflictNonmember1!", nil)

	otherNote := CreateNote(t, harness, adminLogin, incidentID, "txn-workbook_interaction-conflict-guard-other-record", "Other record", "Other body")
	otherRecordID := appsupport.MustUUID(t, otherNote["record_id"].(string))
	wrongRecordURL := harness.Server.HTTP.URL + "/api/v1/records/" + otherRecordID.String() + "/conflicts/" + conflictToken + "/resolve"
	otherIncident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook_interaction-conflict-guard-other-incident",
		"incident_key":  "IR-WORKBOOK-INTERACTION-CONFLICT-OTHER",
		"title":         "Workbook inspector conflict other incident",
	})
	otherIncidentID := appsupport.MustUUID(t, otherIncident["incident_id"].(string))
	otherIncidentNote := CreateNote(t, harness, adminLogin, otherIncidentID, "txn-workbook_interaction-conflict-guard-cross-incident-record", "Cross incident record", "Cross incident body")
	otherIncidentRecordID := appsupport.MustUUID(t, otherIncidentNote["record_id"].(string))
	crossIncidentURL := harness.Server.HTTP.URL + "/api/v1/records/" + otherIncidentRecordID.String() + "/conflicts/" + conflictToken + "/resolve"
	missingRecordURL := harness.Server.HTTP.URL + "/api/v1/records/00000000-0000-4000-8000-000000009999/conflicts/not-a-token/resolve"

	beforeUnauthenticatedState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	unauthenticated := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/not-a-uuid/conflicts/not-a-token/resolve", map[string]any{})
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")
	requireWorkbookStateUnchanged(t, beforeUnauthenticatedState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

	beforeMissingCSRFState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	missingCSRF := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, invalidTokenURL, map[string]any{}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF))
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")
	requireWorkbookStateUnchanged(t, beforeMissingCSRFState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

	beforeMissingRecordState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	missingRecord := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, missingRecordURL, map[string]any{}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, missingRecord, http.StatusNotFound, "incident_not_found")
	requireWorkbookStateUnchanged(t, beforeMissingRecordState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

	beforeNonMemberResponseState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	nonMemberResponse := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, invalidTokenURL, map[string]any{}, httptestx.WithCookies(nonMemberLogin.Session, nonMemberLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, nonMemberLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, nonMemberResponse, http.StatusNotFound, "incident_not_found")
	requireWorkbookStateUnchanged(t, beforeNonMemberResponseState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

	beforeViewerResponseState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	viewerResponse := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, invalidTokenURL, map[string]any{}, httptestx.WithCookies(viewerLogin.Session, viewerLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, viewerResponse, http.StatusForbidden, "authorization_denied")
	requireWorkbookStateUnchanged(t, beforeViewerResponseState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

	beforeInvalidTokenState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	invalidToken := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, invalidTokenURL, map[string]any{}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	invalidTokenBody := httptestx.RequireErrorEnvelope(t, invalidToken, http.StatusBadRequest, "invalid_mutation_payload")
	requireWorkbookStateUnchanged(t, beforeInvalidTokenState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))
	if httptestx.RequireErrorDetails(t, invalidTokenBody)["field"] != "conflict_token" {
		t.Fatalf("authorized invalid token must identify conflict_token: %#v", invalidTokenBody)
	}

	beforeWrongRecordState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	wrongRecord := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, wrongRecordURL, validBody, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	wrongRecordBody := httptestx.RequireErrorEnvelope(t, wrongRecord, http.StatusBadRequest, "invalid_mutation_payload")
	if httptestx.RequireErrorDetails(t, wrongRecordBody)["field"] != "conflict_token" {
		t.Fatal("token binding failure must identify conflict_token")
	}
	requireWorkbookStateUnchanged(t, beforeWrongRecordState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

	beforeCrossIncidentState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	beforeCrossIncidentStateOther := snapshotWorkbookState(t, harness, adminLogin, otherIncidentID, artifacts.NotesViewSchemaID)
	crossIncident := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, crossIncidentURL, validBody, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	crossIncidentBody := httptestx.RequireErrorEnvelope(t, crossIncident, http.StatusBadRequest, "invalid_mutation_payload")
	if httptestx.RequireErrorDetails(t, crossIncidentBody)["field"] != "conflict_token" {
		t.Fatal("token binding failure must identify conflict_token")
	}
	requireWorkbookStateUnchanged(t, beforeCrossIncidentState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))
	requireWorkbookStateUnchanged(t, beforeCrossIncidentStateOther, snapshotWorkbookState(t, harness, adminLogin, otherIncidentID, artifacts.NotesViewSchemaID))

	beforeInvalidBodyState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	invalidBody := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, resolveURL, map[string]any{
		"conflict_token":  "different-token",
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-workbook_interaction-conflict-guard-invalid-body",
	}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	invalidBodyBody := httptestx.RequireErrorEnvelope(t, invalidBody, http.StatusBadRequest, "invalid_mutation_payload")
	if httptestx.RequireErrorDetails(t, invalidBodyBody)["field"] != "conflict_token" {
		t.Fatal("token binding failure must identify conflict_token")
	}
	requireWorkbookStateUnchanged(t, beforeInvalidBodyState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))

}

func TestWorkbookCurrentMembership_Integration(t *testing.T) {
	harness, adminLogin, adminUserID, incidentID := ConflictFixture(t, "workbook-current-membership", "IR-WORKBOOK-CURRENT-MEMBERSHIP")

	editor := authflowtest.SeedLocalUserRecord(t, harness.DB, "collaboration-shared-editor@example.test", "Collaboration Shared Editor", "CollaborationSharedEditor1!", false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, editor.ID, "editor", adminUserID)
	editorLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, editor.Email, "CollaborationSharedEditor1!", nil)
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

	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, editor.ID, "viewer", adminUserID)

	afterRow := CreateNote(t, harness, adminLogin, incidentID, "txn-collaboration-support-auth-after-create", "Authorization after row", "Authorization body")
	afterRecordID := appsupport.MustUUID(t, afterRow["record_id"].(string))
	beforeAfterRespState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
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
	requireWorkbookStateUnchanged(t, beforeAfterRespState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))
}

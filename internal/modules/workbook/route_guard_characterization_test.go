package workbook_test

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
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
	// History and a soft-deleted record exist before the close, so rollback and
	// restore below reach the lifecycle check rather than a precondition.
	historyNote := CreateNote(t, harness, adminLogin, incidentID, "txn-closed-history-note", "History note", "History body")
	historyID := appsupport.MustUUID(t, historyNote["record_id"].(string))
	requireWorkbookPatch(t, harness, adminLogin, historyID, map[string]any{
		"view_schema_id":   artifacts.NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-closed-history-patch",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "History note revised"}},
	})
	deletedNote := CreateNote(t, harness, adminLogin, incidentID, "txn-closed-deleted-note", "Deleted note", "Deleted body")
	deletedID := appsupport.MustUUID(t, deletedNote["record_id"].(string))
	deleted := httptestx.RequireSuccessEnvelope(t, doWorkbookJSON(t, harness, adminLogin, http.MethodDelete, uuid.Nil, "", deletedID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-closed-delete-before-close",
		"reason":           "Prepare restore",
	}), http.StatusOK)["data"].(map[string]any)
	tombstoneVersion := int64(deleted["row_version"].(float64))
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incidents
   SET status = 'closed',
       closed_at = now()
 WHERE id = $1
`, incidentID); err != nil {
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

	// REQ-01-590: delete, restore, rollback and clipboard paste reject a closed
	// incident with the declared 409 incident_closed and commit nothing.
	recordsURL := harness.Server.HTTP.URL + "/api/v1/records/"
	csrf := []func(*http.Request){httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value)}
	for _, tc := range []struct {
		name, method, url string
		body              map[string]any
	}{
		{"delete", http.MethodDelete, recordsURL + recordID.String(), map[string]any{
			"base_row_version": 1, "client_txn_id": "txn-closed-delete", "reason": "blocked after close",
		}},
		{"restore", http.MethodPost, recordsURL + deletedID.String() + "/restore", map[string]any{
			"base_row_version": tombstoneVersion, "client_txn_id": "txn-closed-restore",
		}},
		{"rollback", http.MethodPost, recordsURL + historyID.String() + "/rollback", map[string]any{
			"base_row_version": 2, "client_txn_id": "txn-closed-rollback",
			"target": map[string]any{"kind": "row_restore", "restore_to_revision_no": 1},
		}},
		{"clipboard paste", http.MethodPost, harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + timeline.TimelineViewSchemaID + "/clipboard-paste", map[string]any{
			"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "txn-closed-paste",
			"clipboard_text": "blocked after close", "format": "csv", "header_mode": "none",
			"start_field_key": "timeline.raw_activity_text", "columns": []string{"timeline.raw_activity_text"},
			"targets": []map[string]any{{"kind": "create"}},
		}},
	} {
		before := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
		response := httptestx.DoJSON(t, http.DefaultClient, tc.method, tc.url, tc.body, csrf...)
		if response.StatusCode != http.StatusConflict {
			t.Fatalf("closed %s: status %d body=%#v", tc.name, response.StatusCode, httptestx.ReadJSONBody(t, response))
		}
		httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "incident_closed")
		requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))
	}
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

	// REQ-01-704 / AC-612: supersede decides visibility and role before the
	// deleted-state classification. A non-member's supersede of a soft-deleted
	// record matches its supersede of a missing record and commits nothing; the
	// demoted viewer receives the route's authorization failure; only a
	// reviewer or admin reaches record_deleted_use_restore.
	deletedNote := CreateNote(t, harness, adminLogin, incidentID, "txn-current-membership-supersede-deleted-create", "Deleted supersede target", "Deleted body")
	deletedID := appsupport.MustUUID(t, deletedNote["record_id"].(string))
	httptestx.RequireSuccessEnvelope(t, doWorkbookJSON(t, harness, adminLogin, http.MethodDelete, uuid.Nil, "", deletedID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-current-membership-supersede-deleted-delete",
		"reason":           "Prepare concealed supersede",
	}), http.StatusOK)
	nonMember := authflowtest.SeedLocalUserRecord(t, harness.DB, "workbook-current-membership-nonmember@example.test", "Workbook Current Membership Nonmember", "WorkbookCurrentMembershipNonmember1!", false, false, true)
	nonMemberLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, nonMember.Email, "WorkbookCurrentMembershipNonmember1!", nil)
	supersede := func(login authcookietest.AuthCookies, recordID string, clientTxnID string) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede", map[string]any{
			"base_row_version":      2,
			"client_txn_id":         clientTxnID,
			"reason":                "Concealed supersede",
			"replacement_record_id": afterRecordID.String(),
		}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	beforeSupersedeState := snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID)
	deletedSupersede := supersede(nonMemberLogin, deletedID.String(), "txn-current-membership-supersede-nonmember-deleted")
	missingSupersede := supersede(nonMemberLogin, uuid.NewString(), "txn-current-membership-supersede-nonmember-missing")
	bodies := make([]map[string]any, 0, 2)
	for _, response := range []*http.Response{deletedSupersede, missingSupersede} {
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "incident_not_found")
		delete(body["error"].(map[string]any), "request_id")
		bodies = append(bodies, body)
	}
	if !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("deleted-record supersede response %v differs from missing-record response %v", bodies[0], bodies[1])
	}
	viewerSupersede := httptestx.RequireErrorEnvelope(t, supersede(editorLogin, deletedID.String(), "txn-current-membership-supersede-viewer-deleted"), http.StatusForbidden, "authorization_denied")
	httptestx.RequireErrorDetail(t, viewerSupersede, "required_role", "reviewer|admin")
	httptestx.RequireErrorEnvelope(t, supersede(adminLogin, deletedID.String(), "txn-current-membership-supersede-admin-deleted"), http.StatusConflict, "record_deleted_use_restore")
	requireWorkbookStateUnchanged(t, beforeSupersedeState, snapshotWorkbookState(t, harness, adminLogin, incidentID, artifacts.NotesViewSchemaID))
}

// TestWorkbookRouteAdmissionOrder_Integration covers Core 01 REQ-01-707 and
// AC-617 for the sixteen workbook routes: route resolution rejects undeclared
// methods and HEAD with registered envelopes; authentication and CSRF precede
// path parameters; a malformed or encoded-slash incident_id, record_id,
// note_record_id, view_schema_id or conflict_token receives the declared
// incident_not_found, identical to a missing resource's, and never reaches
// another operation, while a well-formed unknown view schema or conflict token
// keeps its route-handling 400 (REQ-01-707: selectors declare syntax only);
// and visibility precedes the role gate, the query and the request body
// (REQ-01-704, REQ-01-074, REQ-01-186), with no committed effect.
func TestWorkbookRouteAdmissionOrder_Integration(t *testing.T) {
	harness, admin, adminID, incidentID := ConflictFixture(t, "workbook-admission-order", "IR-WORKBOOK-ADMISSION-ORDER")
	base := harness.Server.HTTP.URL
	note := CreateNote(t, harness, admin, incidentID, "txn-workbook-admission-order-note", "Admission order", "Admission order body")
	recordID := note["record_id"].(string)
	viewerPassword := "WorkbookAdmissionViewer1!"
	viewer := authflowtest.SeedLocalUserRecord(t, harness.DB, "workbook-admission-viewer@example.test", "Workbook Admission Viewer", viewerPassword, false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, viewer.ID, "viewer", adminID)
	viewerLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, base, viewer.Email, viewerPassword, nil)
	outsiderPassword := "WorkbookAdmissionOutsider1!"
	outsider := authflowtest.SeedLocalUserRecord(t, harness.DB, "workbook-admission-outsider@example.test", "Workbook Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, outsiderPassword, nil)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireAlike := func(t testing.TB, existing, missing *http.Response) {
		t.Helper()
		bodies := make([]map[string]any, 0, 2)
		for _, response := range []*http.Response{existing, missing} {
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "incident_not_found")
			delete(body["error"].(map[string]any), "request_id")
			bodies = append(bodies, body)
		}
		if !reflect.DeepEqual(bodies[0], bodies[1]) {
			t.Fatalf("existing-resource response %v differs from missing-resource response %v", bodies[0], bodies[1])
		}
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow ...string) {
		t.Helper()
		if got := response.Header.Get("Allow"); got != strings.Join(allow, ", ") {
			t.Fatalf("Allow = %q, want %q", got, strings.Join(allow, ", "))
		}
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
		want := make([]any, 0, len(allow))
		for _, method := range allow {
			want = append(want, method)
		}
		if details := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(details, map[string]any{"allowed_methods": want}) {
			t.Fatalf("details = %#v, want allowed_methods %v", details, allow)
		}
	}
	const view = artifacts.NotesViewSchemaID
	const token = "well-formed-unknown-conflict-token"
	incidentPath := "/api/v1/incidents/" + incidentID.String()
	missingIncident := "/api/v1/incidents/" + uuid.NewString()
	recordPath := "/api/v1/records/" + recordID
	missingRecord := "/api/v1/records/" + uuid.NewString()
	type route struct {
		name, method, suffix string
		// csrf: the operation is state-changing; gated: the route's role gate
		// excludes a viewer.
		csrf, gated bool
	}
	incidentRoutes := []route{
		{name: "applyWorkbookBulkMutation", method: http.MethodPost, suffix: "/views/" + view + "/bulk-mutations", csrf: true, gated: true},
		{name: "pasteWorkbookClipboard", method: http.MethodPost, suffix: "/views/" + view + "/clipboard-paste", csrf: true, gated: true},
		{name: "locateWorkbookViewRecord", method: http.MethodPost, suffix: "/views/" + view + "/locate"},
		{name: "queryWorkbookView", method: http.MethodPost, suffix: "/views/" + view + "/query"},
		{name: "createViewRow", method: http.MethodPost, suffix: "/views/" + view + "/rows", csrf: true, gated: true},
		{name: "getIncidentDefaultWorkbookPreferences", method: http.MethodGet, suffix: "/workbook-preferences/default"},
		{name: "putIncidentDefaultWorkbookPreferences", method: http.MethodPut, suffix: "/workbook-preferences/default", csrf: true, gated: true},
		{name: "getCurrentUserWorkbookPreferences", method: http.MethodGet, suffix: "/workbook-preferences/me"},
		{name: "putCurrentUserWorkbookPreferences", method: http.MethodPut, suffix: "/workbook-preferences/me", csrf: true},
		{name: "getIncidentWorkbookStartup", method: http.MethodGet, suffix: "/workbook-startup"},
	}
	recordRoutes := []route{
		{name: "listNoteAssociations", method: http.MethodGet, suffix: "/note-associations"},
		{name: "mutateNoteAssociations", method: http.MethodPost, suffix: "/note-associations", csrf: true, gated: true},
		{name: "patchRecord", method: http.MethodPatch, csrf: true, gated: true},
		{name: "resolveRecordSameFieldConflict", method: http.MethodPost, suffix: "/conflicts/" + token + "/resolve", csrf: true, gated: true},
		{name: "createRecordLinkedNote", method: http.MethodPost, suffix: "/linked-notes", csrf: true, gated: true},
		{name: "supersedeRecord", method: http.MethodPost, suffix: "/supersede", csrf: true, gated: true},
	}
	// invalid is a request body no route accepts, or none for a read.
	invalid := func(r route) any {
		if r.method == http.MethodGet {
			return nil
		}
		return []any{1}
	}
	viewRoutes := incidentRoutes[:5]
	httptestx.RequireSuccessEnvelope(t, send(http.MethodGet, "/api/v1/incidents/"+strings.ToUpper(incidentID.String())+"/workbook-preferences/me", admin, nil), http.StatusOK)
	before := snapshotWorkbookState(t, harness, admin, incidentID, view)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		paths := map[string]route{}
		for _, r := range incidentRoutes {
			paths["/api/v1/incidents/not-an-incident-id"+r.suffix] = r
		}
		for _, r := range viewRoutes {
			paths[incidentPath+strings.Replace(r.suffix, view, "NOT_A_SCHEMA", 1)] = r
		}
		for _, r := range recordRoutes {
			paths["/api/v1/records/not-a-record-id"+r.suffix] = r
		}
		paths[recordPath+"/conflicts/bad!token/resolve"] = recordRoutes[3]
		for path, r := range paths {
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, base+path, invalid(r)), http.StatusUnauthorized, "session_required")
			if r.csrf {
				withoutCSRF := httptestx.DoJSON(t, client, r.method, base+path, invalid(r), httptestx.WithCookies(admin.Session))
				httptestx.RequireErrorEnvelope(t, withoutCSRF, http.StatusForbidden, "csrf_verification_failed")
			}
		}
	})
	t.Run("a malformed identifier is a missing resource", func(t *testing.T) {
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(incidentID.String(), "-", ""), "%7B" + incidentID.String() + "%7D"} {
			for _, r := range incidentRoutes {
				requireAlike(t, send(r.method, "/api/v1/incidents/"+malformed+r.suffix, admin, invalid(r)), send(r.method, missingIncident+r.suffix, admin, invalid(r)))
			}
		}
		for _, malformed := range []string{"NOT_A_SCHEMA", "cartulary.view.notes", "cartulary.view.notes.v0", strings.Repeat("a", 126) + ".v1"} {
			for _, r := range viewRoutes {
				requireAlike(t, send(r.method, incidentPath+strings.Replace(r.suffix, view, malformed, 1), admin, invalid(r)), send(r.method, missingIncident+r.suffix, admin, invalid(r)))
			}
		}
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(recordID, "-", ""), "%7B" + recordID + "%7D"} {
			for _, r := range recordRoutes {
				requireAlike(t, send(r.method, "/api/v1/records/"+malformed+r.suffix, admin, invalid(r)), send(r.method, missingRecord+r.suffix, admin, invalid(r)))
			}
		}
		for _, malformed := range []string{"bad!token", "bad%20token", strings.Repeat("t", 4097)} {
			requireAlike(t, send(http.MethodPost, recordPath+"/conflicts/"+malformed+"/resolve", admin, invalid(recordRoutes[3])), send(http.MethodPost, missingRecord+"/conflicts/"+token+"/resolve", admin, invalid(recordRoutes[3])))
		}
	})
	t.Run("a well-formed unknown selector keeps its route-handling response", func(t *testing.T) {
		const unknownView = "cartulary.view.unknown_surface.v1"
		for _, r := range viewRoutes {
			path := incidentPath + strings.Replace(r.suffix, view, unknownView, 1)
			requireAlike(t, send(r.method, path, outsiderLogin, map[string]any{}), send(r.method, missingIncident+strings.Replace(r.suffix, view, unknownView, 1), outsiderLogin, map[string]any{}))
			code := "invalid_mutation_payload"
			if !r.csrf {
				code = "invalid_view_query"
			}
			body := httptestx.RequireErrorEnvelope(t, send(r.method, path, admin, map[string]any{}), http.StatusBadRequest, code)
			if httptestx.RequireErrorDetails(t, body)["field"] != "view_schema_id" {
				t.Fatalf("%s: unknown view schema details = %#v", r.name, body)
			}
		}
		httptestx.RequireErrorEnvelope(t, send(http.MethodPost, recordPath+"/conflicts/"+token+"/resolve", viewerLogin, map[string]any{}), http.StatusForbidden, "authorization_denied")
		body := httptestx.RequireErrorEnvelope(t, send(http.MethodPost, recordPath+"/conflicts/"+token+"/resolve", admin, map[string]any{}), http.StatusBadRequest, "invalid_mutation_payload")
		if httptestx.RequireErrorDetails(t, body)["field"] != "conflict_token" {
			t.Fatalf("unknown conflict token details = %#v", body)
		}
	})
	t.Run("an encoded slash stays inside the identifier", func(t *testing.T) {
		requireAlike(t, send(http.MethodPost, incidentPath+"/views/"+view+"%2Fx/query", admin, map[string]any{}), send(http.MethodPost, missingIncident+"/views/"+view+"/query", admin, map[string]any{}))
		requireAlike(t, send(http.MethodPost, incidentPath+"%2Fx/views/"+view+"/rows", admin, map[string]any{}), send(http.MethodPost, missingIncident+"/views/"+view+"/rows", admin, map[string]any{}))
		requireAlike(t, send(http.MethodPatch, recordPath+"%2Fsupersede", admin, map[string]any{}), send(http.MethodPatch, missingRecord, admin, map[string]any{}))
		requireAlike(t, send(http.MethodPost, recordPath+"/conflicts/"+token+"%2Fx/resolve", admin, map[string]any{}), send(http.MethodPost, missingRecord+"/conflicts/"+token+"/resolve", admin, map[string]any{}))
		requireAlike(t, send(http.MethodGet, incidentPath+"%2Fworkbook-startup", admin, nil), send(http.MethodGet, missingIncident, admin, nil))
		// POST on the record template: the encoded slash never selects supersede.
		requireMethodNotAllowed(t, send(http.MethodPost, recordPath+"%2Fsupersede", admin, map[string]any{}), http.MethodDelete, http.MethodPatch)
		httptestx.RequireErrorEnvelope(t, send(http.MethodPut, incidentPath+"/workbook-preferences%2Fme", admin, map[string]any{}), http.StatusNotFound, "route_not_found")
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		for path, allow := range map[string]string{
			incidentPath + "/workbook-preferences/default": "GET, PUT",
			incidentPath + "/workbook-preferences/me":      "GET, PUT",
			incidentPath + "/workbook-startup":             "GET",
			recordPath + "/note-associations":              "GET, POST",
		} {
			head := httptestx.DoJSON(t, client, http.MethodHead, base+path, nil, httptestx.WithCookies(admin.Session))
			body, err := io.ReadAll(head.Body)
			_ = head.Body.Close()
			if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != allow || len(body) != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q err=%v, want 405 Allow %s without a body", path, head.StatusCode, head.Header.Get("Allow"), body, err, allow)
			}
		}
		requireMethodNotAllowed(t, send(http.MethodDelete, incidentPath+"/workbook-preferences/default", admin, nil), http.MethodGet, http.MethodPut)
		requireMethodNotAllowed(t, send(http.MethodGet, incidentPath+"/views/"+view+"/query", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(http.MethodPut, incidentPath+"/views/"+view+"/rows", admin, map[string]any{}), http.MethodPost)
		requireMethodNotAllowed(t, send(http.MethodGet, recordPath, admin, nil), http.MethodDelete, http.MethodPatch)
		requireMethodNotAllowed(t, send(http.MethodGet, recordPath+"/conflicts/"+token+"/resolve", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(http.MethodPut, recordPath+"/supersede", admin, map[string]any{}), http.MethodPost)
		requireMethodNotAllowed(t, send(http.MethodPut, recordPath+"/note-associations", admin, map[string]any{}), http.MethodGet, http.MethodPost)
		httptestx.RequireErrorEnvelope(t, send(http.MethodPost, incidentPath+"/views/"+view+"/unknown", admin, map[string]any{}), http.StatusNotFound, "route_not_found")
	})
	t.Run("every resolved workbook startup response is no-store", func(t *testing.T) {
		// Core 01 REQ-01-151.1: authentication, path-parameter, visibility and
		// query failures carry no-store like the success; other reads do not.
		requireCacheControl := func(t testing.TB, response *http.Response, want string) {
			t.Helper()
			defer response.Body.Close()
			if got := response.Header.Get("Cache-Control"); got != want {
				t.Fatalf("%s %s = %d Cache-Control %q, want %q", response.Request.Method, response.Request.URL.Path, response.StatusCode, got, want)
			}
		}
		requireCacheControl(t, httptestx.DoJSON(t, client, http.MethodGet, base+incidentPath+"/workbook-startup", nil), "no-store")
		requireCacheControl(t, send(http.MethodGet, "/api/v1/incidents/not-an-incident-id/workbook-startup", admin, nil), "no-store")
		requireCacheControl(t, send(http.MethodGet, missingIncident+"/workbook-startup", admin, nil), "no-store")
		requireCacheControl(t, send(http.MethodGet, incidentPath+"/workbook-startup", outsiderLogin, nil), "no-store")
		requireCacheControl(t, send(http.MethodGet, incidentPath+"/workbook-startup?view_schema_id=x", viewerLogin, nil), "no-store")
		requireCacheControl(t, send(http.MethodGet, incidentPath+"/workbook-startup", viewerLogin, nil), "no-store")
		requireCacheControl(t, send(http.MethodGet, incidentPath+"/workbook-preferences/me", viewerLogin, nil), "")
	})
	t.Run("visibility precedes the role gate and route handling", func(t *testing.T) {
		for _, r := range incidentRoutes {
			requireAlike(t, send(r.method, incidentPath+r.suffix+"?limit=1", outsiderLogin, invalid(r)), send(r.method, missingIncident+r.suffix+"?limit=1", outsiderLogin, invalid(r)))
		}
		for _, r := range recordRoutes {
			requireAlike(t, send(r.method, recordPath+r.suffix+"?limit=1", outsiderLogin, invalid(r)), send(r.method, missingRecord+r.suffix+"?limit=1", outsiderLogin, invalid(r)))
		}
		// A viewer passes visibility: a gated route refuses the role before
		// reading the query or body, and an ungated one classifies them.
		codes := map[string]string{
			"locateWorkbookViewRecord":              "invalid_view_query",
			"queryWorkbookView":                     "invalid_view_query",
			"getIncidentDefaultWorkbookPreferences": "invalid_pagination_request",
			"getCurrentUserWorkbookPreferences":     "invalid_pagination_request",
			"putCurrentUserWorkbookPreferences":     "invalid_mutation_payload",
			"getIncidentWorkbookStartup":            "invalid_startup_request",
			"listNoteAssociations":                  "invalid_list_query",
		}
		for prefix, routes := range map[string][]route{incidentPath: incidentRoutes, recordPath: recordRoutes} {
			for _, r := range routes {
				response := send(r.method, prefix+r.suffix+"?limit=1", viewerLogin, invalid(r))
				if r.gated {
					httptestx.RequireErrorEnvelope(t, response, http.StatusForbidden, "authorization_denied")
					continue
				}
				httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, codes[r.name])
			}
		}
	})
	requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, harness, admin, incidentID, view))
}

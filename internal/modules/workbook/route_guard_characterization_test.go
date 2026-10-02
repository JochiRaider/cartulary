package workbook_test

import (
	"context"
	"encoding/json"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbookRouteGuardsFailBeforeMutation(t *testing.T) {
	var activeRoutes *httpapi.RouteRegistry
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{
		Prefix:        "workbook_interaction-workbook-route-guards",
		TestRouteMode: httptestx.TestRouteModeDisabled,
		AdditionalRoutes: []httpapi.RouteRegistrar{func(_ *http.ServeMux, deps httpapi.DependencySet) error {
			activeRoutes = deps.PublicRoutes
			return nil
		}},
	})
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook_interaction-workbook-route-guards-incident",
		"incident_key":  "IR-WORKBOOK-INTERACTION-ROUTE-GUARDS",
		"title":         "Workbook inspector workbook route guards",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.notes.v1/rows"

	beforeAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	body := map[string]any{
		"client_txn_id": "txn-workbook_interaction-workbook-route-guards-auth",
		"note.title":    "guarded note",
	}

	unauthenticated := appsupport.DoJSON(t, http.MethodPost, createURL, body)
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := appsupport.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
	)
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	invalidCSRF := appsupport.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, "wrong-csrf-token"),
	)
	httptestx.RequireErrorEnvelope(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

	afterAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if afterAuth != beforeAuth {
		t.Fatalf("auth/csrf failures mutated workbook state: before=%+v after=%+v", beforeAuth, afterAuth)
	}

	noteData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1", map[string]any{
		"client_txn_id": "txn-workbook_interaction-workbook-route-guards-note",
		"note.title":    "closed incident guard note",
	})
	recordID := appsupport.MustUUID(t, noteData["row"].(map[string]any)["record_id"].(string))
	requireWorkbookOperationGuardMatrix(t, harness, adminLogin, incidentID, recordID, activeRoutes)
	requireConflictResolveSecurityPrecedence(t, harness, adminLogin, adminUserID, incidentID)

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
		"client_txn_id": "txn-workbook_interaction-workbook-route-guards-closed-create",
		"note.title":    "blocked after close",
	})
	httptestx.RequireErrorEnvelope(t, closedCreate, http.StatusConflict, "incident_closed")

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

	afterClosed := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if afterClosed != beforeClosed {
		t.Fatalf("closed incident failures mutated workbook state: before=%+v after=%+v", beforeClosed, afterClosed)
	}
}

func requireWorkbookOperationGuardMatrix(t *testing.T, harness *appsupport.ServerHarness, adminLogin appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, activeRoutes *httpapi.RouteRegistry) {
	t.Helper()

	// Independent guard fixture: never derive expected identities from registration.
	operations := []workbookGuardOperation{
		{"applyWorkbookBulkMutation", "POST", "/api/v1/incidents/{incident_id}/views/{view_schema_id}/bulk-mutations", true},
		{"createRecordLinkedNote", "POST", "/api/v1/records/{record_id}/linked-notes", true},
		{"createViewRow", "POST", "/api/v1/incidents/{incident_id}/views/{view_schema_id}/rows", true},
		{"getCurrentUserWorkbookPreferences", "GET", "/api/v1/incidents/{incident_id}/workbook-preferences/me", false},
		{"getIncidentDefaultWorkbookPreferences", "GET", "/api/v1/incidents/{incident_id}/workbook-preferences/default", false},
		{"getIncidentWorkbookStartup", "GET", "/api/v1/incidents/{incident_id}/workbook-startup", false},
		{"listNoteAssociations", "GET", "/api/v1/records/{note_record_id}/note-associations", false},
		{"mutateNoteAssociations", "POST", "/api/v1/records/{note_record_id}/note-associations", true},
		{"patchRecord", "PATCH", "/api/v1/records/{record_id}", true},
		{"pasteWorkbookClipboard", "POST", "/api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste", true},
		{"putCurrentUserWorkbookPreferences", "PUT", "/api/v1/incidents/{incident_id}/workbook-preferences/me", true},
		{"putIncidentDefaultWorkbookPreferences", "PUT", "/api/v1/incidents/{incident_id}/workbook-preferences/default", true},
		{"queryWorkbookView", "POST", "/api/v1/incidents/{incident_id}/views/{view_schema_id}/query", false},
		{"resolveRecordSameFieldConflict", "POST", "/api/v1/records/{record_id}/conflicts/{conflict_token}/resolve", true},
		{"supersedeRecord", "POST", "/api/v1/records/{record_id}/supersede", true},
	}
	requireWorkbookOperationParity(t, operations, activeRoutes)
	instantiate := strings.NewReplacer("{incident_id}", incidentID.String(), "{record_id}", recordID.String(), "{note_record_id}", recordID.String(), "{view_schema_id}", "cartulary.view.notes.v1", "{conflict_token}", "not-a-token")

	before := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	for _, operation := range operations {
		t.Run(operation.name+"/authentication", func(t *testing.T) {
			response := appsupport.DoJSON(t, operation.method, harness.Server.HTTP.URL+instantiate.Replace(operation.path), map[string]any{})
			httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, "session_required")
		})
		if operation.requiresCSRF {
			t.Run(operation.name+"/csrf", func(t *testing.T) {
				response := appsupport.DoJSON(
					t,
					operation.method,
					harness.Server.HTTP.URL+instantiate.Replace(operation.path),
					map[string]any{},
					appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				)
				httptestx.RequireErrorEnvelope(t, response, http.StatusForbidden, "csrf_verification_failed")
			})
		}
	}
	after := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if after != before {
		t.Fatalf("operation guard matrix mutated workbook state: before=%+v after=%+v", before, after)
	}
}

func requireConflictResolveSecurityPrecedence(t testing.TB, harness *appsupport.ServerHarness, adminLogin appsupport.LoginResult, adminUserID uuid.UUID, incidentID uuid.UUID) {
	t.Helper()

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
	conflictToken := conflictBody["error"].(map[string]any)["conflict"].(map[string]any)["conflict_token"].(string)
	validBody := map[string]any{
		"conflict_token":  conflictToken,
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-workbook_interaction-conflict-guard-resolution",
	}
	resolveURL := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String() + "/conflicts/" + conflictToken + "/resolve"
	invalidTokenURL := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String() + "/conflicts/not-a-token/resolve"

	viewer := authflowtest.SeedLocalUserRecord(t, harness.DB, "workbook_interaction-conflict-viewer@example.test", "Workbook inspector Conflict Viewer", "WorkbookInteractionConflictViewer1!", false, false, true)
	incidentstoretest.SeedMembership(t, harness.DB, incidentID, viewer.ID, viewer.DisplayName, "viewer", adminUserID)
	viewerLogin := LoginLocalUserNoMFA(t, harness, viewer.Email, "WorkbookInteractionConflictViewer1!")
	nonMember := authflowtest.SeedLocalUserRecord(t, harness.DB, "workbook_interaction-conflict-nonmember@example.test", "Workbook inspector Conflict Nonmember", "WorkbookInteractionConflictNonmember1!", false, false, true)
	nonMemberLogin := LoginLocalUserNoMFA(t, harness, nonMember.Email, "WorkbookInteractionConflictNonmember1!")

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

	before := snapshotWorkbookRouteGuardState(t, harness, incidentID)

	unauthenticated := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/not-a-uuid/conflicts/not-a-token/resolve", map[string]any{})
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := appsupport.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie))
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	missingRecord := appsupport.DoJSON(t, http.MethodPost, missingRecordURL, map[string]any{}, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, missingRecord, http.StatusNotFound, "incident_not_found")

	nonMemberResponse := appsupport.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, appsupport.WithCookies(nonMemberLogin.SessionCookie, nonMemberLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, nonMemberLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, nonMemberResponse, http.StatusNotFound, "incident_not_found")

	viewerResponse := appsupport.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, appsupport.WithCookies(viewerLogin.SessionCookie, viewerLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, viewerResponse, http.StatusForbidden, "authorization_denied")

	invalidToken := appsupport.DoJSON(t, http.MethodPost, invalidTokenURL, map[string]any{}, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	invalidTokenBody := httptestx.RequireErrorEnvelope(t, invalidToken, http.StatusBadRequest, "invalid_mutation_payload")
	if httptestx.RequireErrorDetails(t, invalidTokenBody)["field"] != "conflict_token" {
		t.Fatalf("authorized invalid token must identify conflict_token: %#v", invalidTokenBody)
	}

	wrongRecord := appsupport.DoJSON(t, http.MethodPost, wrongRecordURL, validBody, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, wrongRecord, http.StatusBadRequest, "invalid_mutation_payload")

	crossIncident := appsupport.DoJSON(t, http.MethodPost, crossIncidentURL, validBody, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, crossIncident, http.StatusBadRequest, "invalid_mutation_payload")

	invalidBody := appsupport.DoJSON(t, http.MethodPost, resolveURL, map[string]any{
		"conflict_token":  "different-token",
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-workbook_interaction-conflict-guard-invalid-body",
	}, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, invalidBody, http.StatusBadRequest, "invalid_mutation_payload")

	after := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if after != before {
		t.Fatalf("rejected conflict precedence matrix mutated workbook state: before=%+v after=%+v", before, after)
	}
}

type workbookRouteGuardState struct {
	Records         int
	Artifacts       int
	ChangeSets      int
	Mutations       int
	RecordRevisions int
	Collaboration   int
}

func snapshotWorkbookRouteGuardState(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID) workbookRouteGuardState {
	t.Helper()
	return workbookRouteGuardState{
		Records: appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM records
 WHERE incident_id = $1
`, incidentID),
		Artifacts: appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM artifacts
 WHERE incident_id = $1
`, incidentID),
		ChangeSets: appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_sets
 WHERE incident_id = $1
`, incidentID),
		Mutations: appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		RecordRevisions: appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_revisions rr
  JOIN records r ON r.record_id = rr.record_id
 WHERE r.incident_id = $1
`, incidentID),
		Collaboration: collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}),
	}
}

// The active registry is captured from the running application through its
// existing additional-registrar seam; no production registry or endpoint is added.
type workbookGuardOperation struct {
	name, method, path string
	requiresCSRF       bool
}

func requireWorkbookOperationParity(t *testing.T, fixtures []workbookGuardOperation, registry *httpapi.RouteRegistry) {
	t.Helper()
	if registry == nil {
		t.Fatal("missing active route registry")
	}
	expected := map[string]workbookGuardOperation{}
	for _, fixture := range fixtures {
		if _, exists := expected[fixture.name]; exists {
			t.Fatal("duplicate guard fixture")
		}
		expected[fixture.name] = fixture
	}
	actual := map[string]workbookGuardOperation{}
	for _, op := range registry.Snapshot() {
		if op.OwnerID != "module.workbook" {
			continue
		}
		if _, exists := actual[op.OperationID]; exists {
			t.Fatal("duplicate active operation")
		}
		// StateChanging is a generated HTTP-method annotation, not the
		// authentication policy (POST queryWorkbookView is a read).
		cookieCSRF := false
		for _, requirement := range op.Security {
			if slices.Contains(requirement, "sessionCookie") {
				if slices.Contains(requirement, "csrfCookie") != slices.Contains(requirement, "csrfHeader") {
					t.Errorf("incomplete cookie CSRF declaration for %s", op.OperationID)
				}
				cookieCSRF = slices.Contains(requirement, "csrfCookie")
			}
		}
		actual[op.OperationID] = workbookGuardOperation{op.OperationID, op.Method, op.PathTemplate, cookieCSRF}
	}
	for name, fixture := range expected {
		if got, exists := actual[name]; !exists || got != fixture {
			t.Errorf("active binding %s: got=%v want=%v", name, got, fixture)
		}
	}
	for name := range actual {
		if _, exists := expected[name]; !exists {
			t.Errorf("unguarded active operation %s", name)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
	root, err := suiteservices.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "contracts/openapi-source/owners/module.workbook/openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	var authored struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &authored); err != nil {
		t.Fatal(err)
	}
	identities := map[string]string{}
	for path, methods := range authored.Paths {
		for method, raw := range methods {
			if !slices.Contains([]string{"get", "post", "put", "patch", "delete", "head", "options", "trace"}, method) {
				continue
			}
			var op struct {
				OperationID string `json:"operationId"`
			}
			if err := json.Unmarshal(raw, &op); err != nil {
				t.Fatal(err)
			}
			if op.OperationID == "" {
				continue
			}
			if _, exists := identities[op.OperationID]; exists {
				t.Fatal("duplicate authored operation")
			}
			identities[op.OperationID] = strings.ToUpper(method) + " " + path
		}
	}
	fixtureIdentities := map[string]string{}
	for _, fixture := range fixtures {
		fixtureIdentities[fixture.name] = fixture.method + " " + fixture.path
	}
	if !reflect.DeepEqual(identities, fixtureIdentities) {
		t.Fatalf("authored OpenAPI differs from guard fixtures: expected=%v authored=%v", fixtureIdentities, identities)
	}
}

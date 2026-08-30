package imports_test

import (
	"database/sql"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func TestImportsRouteCSRFSecurityMatrix_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "imports-csrf-security-matrix")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-imports-csrf-matrix-incident",
		"incident_key":  "IR-IMPORTS-CSRF-MATRIX",
		"title":         "Imports CSRF matrix",
	})
	incidentID := incident["incident_id"].(string)
	sessionID, unitID := startCSVImportSession(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		incidentID,
		"txn-imports-csrf-matrix-session",
		"time,summary\n2026-08-30T12:00:00Z,alpha\n",
		"csrf-matrix.csv",
	)

	viewerPassword := "ImportsCSRFViewer1!"
	viewer := flowtest.SeedLocalUserRecord(
		t,
		harness.DB,
		"imports-csrf-viewer@example.test",
		"Imports CSRF Viewer",
		viewerPassword,
		false,
		false,
		true,
	)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-imports-csrf-viewer-membership",
		"user_id":       viewer.ID.String(),
		"role":          "viewer",
	})
	viewerSession, viewerCSRF := flowtest.LoginLocalUser(
		t,
		harness.Server.HTTP.URL,
		viewer.Email,
		viewerPassword,
		nil,
	)
	viewerLogin := flowtest.LoginResult{SessionCookie: viewerSession, CSRFCookie: viewerCSRF}

	baseURL := harness.Server.HTTP.URL
	basePath := "/api/v1/import-sessions/" + sessionID
	unitPath := basePath + "/units/" + unitID
	malformedBody := map[string]any{"unknown": true}
	mutations := []struct {
		name   string
		method string
		path   string
	}{
		{name: "create", method: http.MethodPost, path: "/api/v1/import-sessions"},
		{name: "mapping", method: http.MethodPut, path: unitPath + "/mapping"},
		{name: "select", method: http.MethodPost, path: unitPath + "/select"},
		{name: "skip", method: http.MethodPost, path: unitPath + "/skip"},
		{name: "regions", method: http.MethodPost, path: unitPath + "/regions"},
		{name: "apply", method: http.MethodPost, path: basePath + "/apply"},
	}

	before := importsSecurityState(t, harness.DB)
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			unauthenticated := httptestx.DoJSON(t, mutation.method, baseURL+mutation.path, malformedBody)
			httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

			missingCSRF := httptestx.DoJSON(
				t,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithCookies(adminLogin.SessionCookie),
			)
			httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

			invalidCSRF := httptestx.DoJSON(
				t,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				httptestx.WithHeader(authn.CSRFHeaderName, "invalid-csrf-proof"),
			)
			httptestx.RequireErrorEnvelope(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

			bearer := httptestx.DoJSON(
				t,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithHeader("Authorization", "Bearer "+adminLogin.SessionCookie.Value),
			)
			httptestx.RequireErrorEnvelope(t, bearer, http.StatusBadRequest, "invalid_import_request")

			validCookie := httptestx.DoJSON(
				t,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
			)
			httptestx.RequireErrorEnvelope(t, validCookie, http.StatusBadRequest, "invalid_import_request")
		})
	}

	readOnly := []struct {
		name string
		path string
	}{
		{name: "session", path: basePath},
		{name: "units", path: basePath + "/units"},
		{name: "unit", path: unitPath},
		{name: "preview", path: unitPath + "/preview"},
	}
	for _, read := range readOnly {
		t.Run("read-only-"+read.name, func(t *testing.T) {
			response := httptestx.DoJSON(
				t,
				http.MethodGet,
				baseURL+read.path,
				nil,
				httptestx.WithCookies(adminLogin.SessionCookie),
			)
			httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
		})
	}

	mappingPreview := httptestx.DoJSON(
		t,
		http.MethodPost,
		baseURL+unitPath+"/mapping-preview",
		malformedBody,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, mappingPreview, http.StatusBadRequest, "invalid_import_request")

	hidden := httptestx.DoJSON(
		t,
		http.MethodPut,
		baseURL+"/api/v1/import-sessions/"+uuid.NewString()+"/units/"+uuid.NewString()+"/mapping",
		malformedBody,
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, hidden, http.StatusNotFound, "import_unit_not_found")

	viewerDenied := httptestx.DoJSON(
		t,
		http.MethodPut,
		baseURL+unitPath+"/mapping",
		malformedBody,
		httptestx.WithCookies(viewerLogin.SessionCookie, viewerLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, viewerDenied, http.StatusForbidden, "authorization_denied")

	after := importsSecurityState(t, harness.DB)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("rejected Imports security matrix changed durable or publication state: before=%v after=%v", before, after)
	}
}

func importsSecurityState(t testing.TB, db *sql.DB) map[string]int {
	t.Helper()
	queries := map[string]string{
		"sessions":             `SELECT COUNT(*) FROM import_sessions`,
		"units":                `SELECT COUNT(*) FROM import_units`,
		"source_streams":       `SELECT COUNT(*) FROM import_source_streams`,
		"apply_plans":          `SELECT COUNT(*) FROM import_apply_unit_plans`,
		"unit_outcomes":        `SELECT COUNT(*) FROM import_unit_apply_outcomes`,
		"apply_journal":        `SELECT COUNT(*) FROM import_apply_journal`,
		"jobs":                 `SELECT COUNT(*) FROM jobs`,
		"route_idempotency":    `SELECT COUNT(*) FROM route_idempotency WHERE route_key LIKE 'imports.%'`,
		"change_sets":          `SELECT COUNT(*) FROM change_sets`,
		"change_set_mutations": `SELECT COUNT(*) FROM change_set_mutations`,
		"record_revisions":     `SELECT COUNT(*) FROM record_revisions`,
	}
	state := make(map[string]int, len(queries)+1)
	for name, query := range queries {
		state[name] = dbassert.CountSQL(t, db, query)
	}
	state["collaboration_intents"] = collaborationsupport.CountIntents(
		t,
		db,
		collaborationsupport.IntentSelector{},
	)
	return state
}

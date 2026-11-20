package imports_test

import (
	"database/sql"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func TestImportsRouteCSRFSecurityMatrix_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "imports-csrf-security-matrix", TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
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
	viewerLogin := flowtest.LoginLocalUser(
		t, http.DefaultClient,
		harness.Server.HTTP.URL,
		viewer.Email,
		viewerPassword,
		nil,
	)

	nonMemberPassword := "ImportsNonMember1!"
	nonMember := flowtest.SeedLocalUserRecord(
		t,
		harness.DB,
		"imports-concealment-nonmember@example.test",
		"Imports Concealment Nonmember",
		nonMemberPassword,
		false,
		false,
		true,
	)
	nonMemberLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, nonMember.Email, nonMemberPassword, nil)

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
			unauthenticated := httptestx.DoJSON(t, http.DefaultClient, mutation.method, baseURL+mutation.path, malformedBody)
			httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

			missingCSRF := httptestx.DoJSON(
				t, http.DefaultClient,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithCookies(adminLogin.Session),
			)
			httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

			invalidCSRF := httptestx.DoJSON(
				t, http.DefaultClient,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
				httptestx.WithHeader(authn.CSRFHeaderName, "invalid-csrf-proof"),
			)
			httptestx.RequireErrorEnvelope(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

			bearer := httptestx.DoJSON(
				t, http.DefaultClient,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithHeader("Authorization", "Bearer "+adminLogin.Session.Value),
			)
			httptestx.RequireErrorEnvelope(t, bearer, http.StatusBadRequest, "invalid_import_request")

			validCookie := httptestx.DoJSON(
				t, http.DefaultClient,
				mutation.method,
				baseURL+mutation.path,
				malformedBody,
				httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
				httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
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
				t, http.DefaultClient,
				http.MethodGet,
				baseURL+read.path,
				nil,
				httptestx.WithCookies(adminLogin.Session),
			)
			httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
		})
	}

	mappingPreview := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		baseURL+unitPath+"/mapping-preview",
		malformedBody,
		httptestx.WithCookies(adminLogin.Session),
	)
	httptestx.RequireErrorEnvelope(t, mappingPreview, http.StatusBadRequest, "invalid_import_request")

	// REQ-01-704 / AC-612: every import-session member route decides the
	// visibility of the session's incident immediately after authentication.
	// A non-member receives the same import_session_not_found for an existing
	// session as for a missing one, before any unit or body is classified.
	missingSessionPath := "/api/v1/import-sessions/" + uuid.NewString()
	memberRoutes := []struct {
		name, method, suffix string
		body                 any
	}{
		{name: "session", method: http.MethodGet},
		{name: "units", method: http.MethodGet, suffix: "/units"},
		{name: "unit", method: http.MethodGet, suffix: "/units/" + unitID},
		{name: "preview", method: http.MethodGet, suffix: "/units/" + unitID + "/preview"},
		{name: "mapping-preview", method: http.MethodPost, suffix: "/units/" + unitID + "/mapping-preview", body: malformedBody},
		{name: "source-inspection", method: http.MethodPost, suffix: "/units/" + unitID + "/source-inspection", body: malformedBody},
		{name: "mapping", method: http.MethodPut, suffix: "/units/" + unitID + "/mapping", body: malformedBody},
		{name: "select", method: http.MethodPost, suffix: "/units/" + unitID + "/select", body: malformedBody},
		{name: "skip", method: http.MethodPost, suffix: "/units/" + unitID + "/skip", body: malformedBody},
		{name: "regions", method: http.MethodPost, suffix: "/units/" + unitID + "/regions", body: malformedBody},
		{name: "apply", method: http.MethodPost, suffix: "/apply", body: malformedBody},
	}
	nonMemberRequest := func(t *testing.T, method, path string, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(
			t, http.DefaultClient, method, baseURL+path, body,
			httptestx.WithCookies(nonMemberLogin.Session, nonMemberLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, nonMemberLogin.CSRF.Value),
		)
	}
	for _, route := range memberRoutes {
		t.Run("concealed-"+route.name, func(t *testing.T) {
			existing := nonMemberRequest(t, route.method, basePath+route.suffix, route.body)
			missing := nonMemberRequest(t, route.method, missingSessionPath+route.suffix, route.body)
			requireImportResponsesAlike(t, existing, missing, http.StatusNotFound, "import_session_not_found")
		})
	}

	// A caller who can see the session receives import_unit_not_found for a
	// unit the session does not contain.
	missingUnitPath := basePath + "/units/" + uuid.NewString()
	regionBody := map[string]any{
		"client_txn_id": "txn-imports-concealment-missing-unit-region",
		"source_rect":   map[string]any{"start_row": 1, "start_column": 1, "end_row": 1, "end_column": 1},
	}
	for _, route := range []struct {
		name, method, suffix string
		body                 any
	}{
		{name: "unit", method: http.MethodGet},
		{name: "preview", method: http.MethodGet, suffix: "/preview"},
		{name: "mapping-preview", method: http.MethodPost, suffix: "/mapping-preview", body: malformedBody},
		{name: "source-inspection", method: http.MethodPost, suffix: "/source-inspection", body: malformedBody},
		{name: "mapping", method: http.MethodPut, suffix: "/mapping", body: malformedBody},
		{name: "select", method: http.MethodPost, suffix: "/select", body: malformedBody},
		{name: "skip", method: http.MethodPost, suffix: "/skip", body: malformedBody},
		{name: "regions", method: http.MethodPost, suffix: "/regions", body: regionBody},
	} {
		t.Run("missing-unit-"+route.name, func(t *testing.T) {
			response := httptestx.DoJSON(
				t, http.DefaultClient, route.method, baseURL+missingUnitPath+route.suffix, route.body,
				httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
				httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
			)
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "import_unit_not_found")
			if details := httptestx.RequireErrorDetails(t, body); len(details) != 0 {
				t.Fatalf("import_unit_not_found details = %v, want none", details)
			}
		})
	}

	hidden := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		baseURL+"/api/v1/import-sessions/"+uuid.NewString()+"/units/"+uuid.NewString()+"/mapping",
		malformedBody,
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, hidden, http.StatusNotFound, "import_session_not_found")

	viewerDenied := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPut,
		baseURL+unitPath+"/mapping",
		malformedBody,
		httptestx.WithCookies(viewerLogin.Session, viewerLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, viewerDenied, http.StatusForbidden, "authorization_denied")

	after := importsSecurityState(t, harness.DB)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("rejected Imports security matrix changed durable or publication state: before=%v after=%v", before, after)
	}
}

// TestImportRouteAdmissionOrder_Integration covers Core 01 REQ-01-707 and
// AC-617 for the Import routes, including every import-unit route kind:
// route resolution rejects undeclared methods and HEAD with registered
// envelopes; authentication and CSRF precede path parameters; a malformed or
// encoded-slash identifier receives its declared not-found response, identical
// to a missing resource's, and never reaches another operation; and
// visibility precedes the role gate, the query check and the request body
// (REQ-01-704).
func TestImportRouteAdmissionOrder_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "imports-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-imports-admission-order-incident",
		"incident_key":  "IR-IMPORTS-ADMISSION-ORDER",
		"title":         "Imports route admission order",
	})["incident_id"].(string)
	sessionID, unitID := startCSVImportSession(t, base, admin, incidentID, "txn-imports-admission-order-session", "time,summary\n2026-08-30T12:00:00Z,alpha\n", "admission-order.csv")
	outsiderPassword := "ImportsAdmissionOutsider1!"
	outsider := flowtest.SeedLocalUserRecord(t, harness.DB, "imports-admission-outsider@example.test", "Imports Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, outsiderPassword, nil)
	viewerPassword := "ImportsAdmissionViewer1!"
	viewer := flowtest.SeedLocalUserRecord(t, harness.DB, "imports-admission-viewer@example.test", "Imports Admission Viewer", viewerPassword, false, false, true)
	scenariotest.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-imports-admission-order-viewer", "user_id": viewer.ID.String(), "role": "viewer",
	})
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, viewer.Email, viewerPassword, nil)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	// Each helper takes the calling (sub)test's t, so a failure inside t.Run
	// fails that subtest from its own goroutine.
	send := func(t testing.TB, method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow string) {
		t.Helper()
		if response.Header.Get("Allow") != allow {
			t.Fatalf("Allow = %q, want %q", response.Header.Get("Allow"), allow)
		}
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
		if details := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(details, map[string]any{"allowed_methods": []any{allow}}) {
			t.Fatalf("details = %#v, want allowed_methods [%s]", details, allow)
		}
	}
	sessionPath := "/api/v1/import-sessions/" + sessionID
	missingSessionPath := "/api/v1/import-sessions/" + uuid.NewString()
	malformedSessionPath := "/api/v1/import-sessions/not-a-session-id"
	unitPath := sessionPath + "/units/" + unitID
	missingUnitPath := sessionPath + "/units/" + uuid.NewString()
	sourcesPath := "/api/v1/incidents/" + incidentID + "/committed-sources"
	missingSourcesPath := "/api/v1/incidents/" + uuid.NewString() + "/committed-sources"
	// Each unit route kind with a body its handler accepts up to the unit
	// lookup, so a missing unit is classified as such.
	unitRoutes := []struct {
		name, method, suffix string
		body                 any
	}{
		{name: "unit", method: http.MethodGet},
		{name: "preview", method: http.MethodGet, suffix: "/preview"},
		{name: "mapping-preview", method: http.MethodPost, suffix: "/mapping-preview", body: map[string]any{}},
		{name: "source-inspection", method: http.MethodPost, suffix: "/source-inspection", body: map[string]any{}},
		{name: "mapping", method: http.MethodPut, suffix: "/mapping", body: map[string]any{}},
		{name: "select", method: http.MethodPost, suffix: "/select", body: map[string]any{}},
		{name: "skip", method: http.MethodPost, suffix: "/skip", body: map[string]any{}},
		{name: "regions", method: http.MethodPost, suffix: "/regions", body: map[string]any{
			"client_txn_id": "txn-imports-admission-order-region",
			"source_rect":   map[string]any{"start_row": 1, "start_column": 1, "end_row": 1, "end_column": 1},
		}},
	}
	sessionRoutes := []struct {
		name, method, suffix string
		body                 any
	}{
		{name: "session", method: http.MethodGet},
		{name: "units", method: http.MethodGet, suffix: "/units"},
		{name: "apply", method: http.MethodPost, suffix: "/apply", body: map[string]any{}},
	}
	httptestx.RequireSuccessEnvelope(t, send(t, http.MethodGet, "/api/v1/import-sessions/"+strings.ToUpper(sessionID)+"/units/"+strings.ToUpper(unitID), admin, nil), http.StatusOK)
	before := importsSecurityState(t, harness.DB)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		for _, route := range append(sessionRoutes, unitRoutes...) {
			path := malformedSessionPath + "/units/not-a-unit-id" + route.suffix
			if route.name == "session" || route.name == "units" || route.name == "apply" {
				path = malformedSessionPath + route.suffix
			}
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, route.method, base+path, route.body), http.StatusUnauthorized, "session_required")
			if route.method != http.MethodGet && route.name != "mapping-preview" && route.name != "source-inspection" {
				withoutCSRF := httptestx.DoJSON(t, client, route.method, base+path, route.body, httptestx.WithCookies(admin.Session))
				httptestx.RequireErrorEnvelope(t, withoutCSRF, http.StatusForbidden, "csrf_verification_failed")
			}
		}
		for _, path := range []string{"/api/v1/incidents/not-an-incident-id/committed-sources", sourcesPath + "/not-a-source-id", sourcesPath + "/not-a-source-id/content"} {
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, http.MethodGet, base+path, nil), http.StatusUnauthorized, "session_required")
		}
	})
	t.Run("a malformed identifier is a missing resource", func(t *testing.T) {
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(unitID, "-", ""), "%7B" + unitID + "%7D"} {
			for _, route := range sessionRoutes {
				requireImportResponsesAlike(t,
					send(t, route.method, "/api/v1/import-sessions/"+malformed+route.suffix, admin, route.body),
					send(t, route.method, missingSessionPath+route.suffix, admin, route.body), http.StatusNotFound, "import_session_not_found")
			}
			for _, route := range unitRoutes {
				requireImportResponsesAlike(t,
					send(t, route.method, "/api/v1/import-sessions/"+malformed+"/units/"+unitID+route.suffix, admin, route.body),
					send(t, route.method, missingSessionPath+"/units/"+unitID+route.suffix, admin, route.body), http.StatusNotFound, "import_session_not_found")
				requireImportResponsesAlike(t,
					send(t, route.method, sessionPath+"/units/"+malformed+route.suffix, admin, route.body),
					send(t, route.method, missingUnitPath+route.suffix, admin, route.body), http.StatusNotFound, "import_unit_not_found")
			}
			requireImportResponsesAlike(t, send(t, http.MethodGet, "/api/v1/incidents/"+malformed+"/committed-sources", admin, nil), send(t, http.MethodGet, missingSourcesPath, admin, nil), http.StatusNotFound, "incident_not_found")
			for _, suffix := range []string{"", "/content"} {
				requireImportResponsesAlike(t, send(t, http.MethodGet, "/api/v1/incidents/"+malformed+"/committed-sources/"+uuid.NewString()+suffix, admin, nil), send(t, http.MethodGet, missingSourcesPath+"/"+uuid.NewString()+suffix, admin, nil), http.StatusNotFound, "incident_not_found")
				requireImportResponsesAlike(t, send(t, http.MethodGet, sourcesPath+"/"+malformed+suffix, admin, nil), send(t, http.MethodGet, sourcesPath+"/"+uuid.NewString()+suffix, admin, nil), http.StatusNotFound, "committed_source_not_found")
			}
		}
	})
	t.Run("an encoded slash stays inside the identifier", func(t *testing.T) {
		requireImportResponsesAlike(t, send(t, http.MethodGet, sessionPath+"%2Funits", admin, nil), send(t, http.MethodGet, missingSessionPath, admin, nil), http.StatusNotFound, "import_session_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodGet, unitPath+"%2Fpreview", admin, nil), send(t, http.MethodGet, missingUnitPath, admin, nil), http.StatusNotFound, "import_unit_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodPost, unitPath+"%2Fx/select", admin, map[string]any{}), send(t, http.MethodPost, missingUnitPath+"/select", admin, map[string]any{}), http.StatusNotFound, "import_unit_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodPost, sessionPath+"%2Fx/apply", admin, map[string]any{}), send(t, http.MethodPost, missingSessionPath+"/apply", admin, map[string]any{}), http.StatusNotFound, "import_session_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodGet, sourcesPath+"/"+uuid.NewString()+"%2Fcontent", admin, nil), send(t, http.MethodGet, sourcesPath+"/"+uuid.NewString(), admin, nil), http.StatusNotFound, "committed_source_not_found")
		// POST on the session read template: the encoded slash never selects apply.
		requireMethodNotAllowed(t, send(t, http.MethodPost, sessionPath+"%2Fapply", admin, map[string]any{}), http.MethodGet)
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		for _, path := range []string{sessionPath, sessionPath + "/units", unitPath, unitPath + "/preview", sourcesPath, sourcesPath + "/" + uuid.NewString(), sourcesPath + "/" + uuid.NewString() + "/content"} {
			head := httptestx.DoJSON(t, client, http.MethodHead, base+path, nil, httptestx.WithCookies(admin.Session))
			body, err := io.ReadAll(head.Body)
			_ = head.Body.Close()
			if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != http.MethodGet || len(body) != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q err=%v, want 405 Allow GET without a body", path, head.StatusCode, head.Header.Get("Allow"), body, err)
			}
		}
		requireMethodNotAllowed(t, send(t, http.MethodDelete, sessionPath, admin, nil), http.MethodGet)
		requireMethodNotAllowed(t, send(t, http.MethodGet, sessionPath+"/apply", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodPost, unitPath+"/mapping", admin, map[string]any{}), http.MethodPut)
		requireMethodNotAllowed(t, send(t, http.MethodGet, unitPath+"/regions", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodGet, "/api/v1/import-sessions", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodPut, sourcesPath, admin, map[string]any{}), http.MethodGet)
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, unitPath+"/unknown", admin, nil), http.StatusNotFound, "route_not_found")
	})
	t.Run("visibility precedes the role gate and route handling", func(t *testing.T) {
		requireImportResponsesAlike(t, send(t, http.MethodGet, sessionPath+"?limit=1", outsiderLogin, nil), send(t, http.MethodGet, missingSessionPath+"?limit=1", outsiderLogin, nil), http.StatusNotFound, "import_session_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodPost, sessionPath+"/apply?limit=1", outsiderLogin, []any{1}), send(t, http.MethodPost, missingSessionPath+"/apply?limit=1", outsiderLogin, []any{1}), http.StatusNotFound, "import_session_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodPost, unitPath+"/select?limit=1", outsiderLogin, []any{1}), send(t, http.MethodPost, missingSessionPath+"/units/"+unitID+"/select?limit=1", outsiderLogin, []any{1}), http.StatusNotFound, "import_session_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodGet, sourcesPath+"?limit=0", outsiderLogin, nil), send(t, http.MethodGet, missingSourcesPath+"?limit=0", outsiderLogin, nil), http.StatusNotFound, "incident_not_found")
		requireImportResponsesAlike(t, send(t, http.MethodGet, sourcesPath+"/"+uuid.NewString()+"?limit=1", outsiderLogin, nil), send(t, http.MethodGet, missingSourcesPath+"/"+uuid.NewString()+"?limit=1", outsiderLogin, nil), http.StatusNotFound, "incident_not_found")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, unitPath+"/select?limit=1", viewerLogin, []any{1}), http.StatusForbidden, "authorization_denied")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, sourcesPath+"/"+uuid.NewString()+"?limit=1", viewerLogin, nil), http.StatusForbidden, "authorization_denied")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, sourcesPath+"/"+uuid.NewString()+"/content?limit=1", admin, nil), http.StatusBadRequest, "invalid_pagination_request")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, unitPath+"/skip?limit=1", admin, []any{1}), http.StatusBadRequest, "invalid_pagination_request")
	})
	t.Run("the unit list admits only singleton limit and cursor_token", func(t *testing.T) {
		// Core 01 Table 17.2-A and REQ-01-583: listImportUnits declares no
		// list-query members, so its query is validated in this order after
		// the session's visibility.
		unitsPath := sessionPath + "/units"
		for _, invalid := range []struct{ query, code, reason string }{
			{"?%zz", "invalid_list_query", "malformed_query"},
			{"?limit=1&limit=2", "invalid_list_query", "duplicate_query_member"},
			{"?unknown=1&unknown=2&page=2", "invalid_list_query", "duplicate_query_member"},
			{"?page=2&unknown=1", "invalid_pagination_request", "invalid_limit"},
			{"?unknown=1", "invalid_list_query", "unknown_query_member"},
			{"?limit=", "invalid_pagination_request", "invalid_limit"},
			{"?limit=0", "invalid_pagination_request", "invalid_limit"},
			{"?cursor_token=", "invalid_pagination_request", "invalid_cursor_token"},
			{"?cursor_token=not-a-cursor", "invalid_pagination_request", "invalid_cursor_token"},
		} {
			body := httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, unitsPath+invalid.query, admin, nil), http.StatusBadRequest, invalid.code)
			if details := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(details, map[string]any{"reason_code": invalid.reason}) {
				t.Fatalf("%s details = %#v, want reason_code %s", invalid.query, details, invalid.reason)
			}
		}
		requireImportResponsesAlike(t, send(t, http.MethodGet, unitsPath+"?unknown=1", outsiderLogin, nil), send(t, http.MethodGet, missingSessionPath+"/units?unknown=1", outsiderLogin, nil), http.StatusNotFound, "import_session_not_found")
		httptestx.RequireSuccessEnvelope(t, send(t, http.MethodGet, unitsPath+"?limit=1", admin, nil), http.StatusOK)
	})
	if after := importsSecurityState(t, harness.DB); !reflect.DeepEqual(after, before) {
		t.Fatalf("rejected Imports admission requests changed durable or publication state: before=%v after=%v", before, after)
	}
}

// requireImportResponsesAlike requires the response for an existing resource
// to match the response for a missing one in everything but its request
// identity: status, code, message, details and retryability (REQ-01-704).
func requireImportResponsesAlike(t testing.TB, existing, missing *http.Response, status int, code string) {
	t.Helper()
	bodies := make([]map[string]any, 0, 2)
	for _, response := range []*http.Response{existing, missing} {
		body := httptestx.RequireErrorEnvelope(t, response, status, code)
		delete(body["error"].(map[string]any), "request_id")
		bodies = append(bodies, body)
	}
	if !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("existing-resource response %v differs from missing-resource response %v", bodies[0], bodies[1])
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

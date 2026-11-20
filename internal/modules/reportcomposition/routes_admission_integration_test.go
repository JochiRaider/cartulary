package reportcomposition_test

// Report composition route admission order (Core 01 REQ-01-707, Core 04
// AC-617).
import (
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestReportCompositionRouteAdmissionOrder_Integration drives the nine
// composition operations through central admission. Authentication and CSRF
// precede path parameters; a malformed incident_id, composition_id or
// composition_version receives the parameter's declared not-found response,
// which for a composition carries its validation_code and is identical to the
// response for a missing resource; an encoded slash stays inside its segment;
// undeclared methods and HEAD are enveloped 405s. Visibility precedes the
// role gate, and both precede query and body validation. No rejected request
// commits an effect.
func TestReportCompositionRouteAdmissionOrder_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "composition-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-composition-admission-order-incident",
		"incident_key":  "IR-COMPOSITION-ADMISSION-ORDER",
		"title":         "Composition route admission order",
	})["incident_id"].(string)
	viewerPassword := "CompositionAdmissionViewer1!"
	viewer := flowtest.SeedLocalUserRecord(t, harness.DB, "composition-admission-viewer@example.test", "Composition Admission Viewer", viewerPassword, false, false, true)
	scenariotest.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-composition-admission-order-viewer", "user_id": viewer.ID.String(), "role": "viewer",
	})
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, viewer.Email, viewerPassword, nil)
	outsiderPassword := "CompositionAdmissionOutsider1!"
	outsider := flowtest.SeedLocalUserRecord(t, harness.DB, "composition-admission-outsider@example.test", "Composition Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, outsiderPassword, nil)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(t testing.TB, method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireError := func(t testing.TB, response *http.Response, status int, code string, details map[string]any) map[string]any {
		t.Helper()
		body := httptestx.RequireErrorEnvelope(t, response, status, code)
		if got := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(got, details) {
			t.Fatalf("%s details = %#v, want %#v", code, got, details)
		}
		delete(body["error"].(map[string]any), "request_id")
		return body
	}
	requireAlike := func(t testing.TB, existing, missing *http.Response, status int, code string, details map[string]any) {
		t.Helper()
		if first, second := requireError(t, existing, status, code, details), requireError(t, missing, status, code, details); !reflect.DeepEqual(first, second) {
			t.Fatalf("existing-resource response %v differs from missing-resource response %v", first, second)
		}
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow ...string) {
		t.Helper()
		if got := response.Header.Get("Allow"); got != strings.Join(allow, ", ") {
			t.Fatalf("Allow = %q, want %q", got, strings.Join(allow, ", "))
		}
		want := make([]any, 0, len(allow))
		for _, method := range allow {
			want = append(want, method)
		}
		requireError(t, response, http.StatusMethodNotAllowed, "method_not_allowed", map[string]any{"allowed_methods": want})
	}

	collection := "/api/v1/incidents/" + incidentID + "/report-compositions"
	created := httptestx.RequireSuccessData(t, send(t, http.MethodPost, collection, admin, map[string]any{
		"client_txn_id": "txn-composition-admission-order-create", "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
	}), http.StatusCreated)
	compositionID := created["composition_id"].(string)
	httptestx.RequireSuccessData(t, send(t, http.MethodPost, collection+"/"+compositionID+"/versions", admin, map[string]any{
		"client_txn_id": "txn-composition-admission-order-freeze", "base_draft_version": 1,
	}), http.StatusCreated)
	httptestx.RequireSuccessData(t, send(t, http.MethodGet, "/api/v1/incidents/"+strings.ToUpper(incidentID)+"/report-compositions/"+strings.ToUpper(compositionID)+"/versions/v1", viewerLogin, nil), http.StatusOK)

	incidentNotFound := map[string]any{}
	compositionNotFound := map[string]any{"validation_code": "composition_not_found"}
	versionNotFound := map[string]any{"validation_code": "composition_version_not_found"}
	draftVersion := map[string]any{"client_txn_id": "txn-composition-admission-order-missing", "base_draft_version": 1}
	preview := map[string]any{
		"client_txn_id": "txn-composition-admission-order-preview", "source_kind": "draft", "snapshot_id": uuid.NewString(),
		"derivation_version": reporting.DerivationVersion, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
		"redaction_profile_id": "cartulary.redaction.internal", "redaction_profile_version": "5", "redaction_profile_sha256": strings.Repeat("a", 64),
		"render_environment_profile_id": "cartulary.reporting.local.v1", "output_kind": reporting.OutputKindSlidev, "output_options": map[string]any{},
		"recipient_partition_refs": []any{}, "graph_projection_refs": []any{},
	}
	// Each composition-addressed operation with a body its handler accepts up
	// to the composition lookup, so a missing composition is classified as
	// such.
	type route struct {
		name, method, suffix string
		csrf                 bool
		body                 any
		notFound             map[string]any
	}
	compositionRoutes := []route{
		{name: "getReportComposition", method: http.MethodGet, notFound: compositionNotFound},
		{name: "updateReportCompositionDraft", method: http.MethodPatch, csrf: true, body: draftVersion, notFound: compositionNotFound},
		{name: "retireReportComposition", method: http.MethodDelete, csrf: true, body: draftVersion, notFound: compositionNotFound},
		{name: "freezeReportCompositionVersion", method: http.MethodPost, suffix: "/versions", csrf: true, body: draftVersion, notFound: compositionNotFound},
		{name: "getReportCompositionVersion", method: http.MethodGet, suffix: "/versions/v1", notFound: versionNotFound},
		{name: "validateReportComposition", method: http.MethodPost, suffix: "/validate", body: map[string]any{}, notFound: compositionNotFound},
		{name: "previewReportComposition", method: http.MethodPost, suffix: "/preview", csrf: true, body: preview, notFound: compositionNotFound},
	}
	collectionRoutes := []route{
		{name: "listReportCompositions", method: http.MethodGet},
		{name: "createReportComposition", method: http.MethodPost, csrf: true, body: map[string]any{}},
	}
	effects := func(t testing.TB) string {
		t.Helper()
		var state string
		if err := harness.DB.QueryRow(`SELECT jsonb_build_object(
 'resources',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY composition_id),'[]') FROM report_compositions r),
 'versions',(SELECT count(*) FROM report_composition_versions),
 'previews',(SELECT count(*) FROM report_composition_preview_attempts),
 'jobs',(SELECT count(*) FROM jobs),
 'receipts',(SELECT count(*) FROM route_idempotency)
 )::text`).Scan(&state); err != nil {
			t.Fatalf("read composition state: %v", err)
		}
		return state + " collaboration_intents=" + strconv.Itoa(collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{}))
	}
	before := effects(t)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		malformedIncident := "/api/v1/incidents/not-an-incident-id/report-compositions"
		paths := func(r route) []string {
			if r.suffix == "" && (r.name == "listReportCompositions" || r.name == "createReportComposition") {
				return []string{malformedIncident}
			}
			return []string{malformedIncident + "/" + compositionID + r.suffix, collection + "/not-a-composition-id" + r.suffix}
		}
		for _, r := range append(append([]route{}, collectionRoutes...), compositionRoutes...) {
			for _, path := range paths(r) {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, base+path, []any{1}), http.StatusUnauthorized, "session_required")
				if r.csrf {
					withoutCSRF := httptestx.DoJSON(t, client, r.method, base+path, []any{1}, httptestx.WithCookies(admin.Session))
					httptestx.RequireErrorEnvelope(t, withoutCSRF, http.StatusForbidden, "csrf_verification_failed")
				}
			}
		}
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, http.MethodGet, base+collection+"/"+compositionID+"/versions/not-a-version", nil), http.StatusUnauthorized, "session_required")
	})
	t.Run("a malformed identifier is a missing resource", func(t *testing.T) {
		missingIncident := "/api/v1/incidents/" + uuid.NewString() + "/report-compositions"
		missingComposition := collection + "/" + uuid.NewString()
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(compositionID, "-", ""), "%7B" + compositionID + "%7D"} {
			malformedIncident := "/api/v1/incidents/" + malformed + "/report-compositions"
			for _, r := range collectionRoutes {
				requireAlike(t, send(t, r.method, malformedIncident, admin, r.body), send(t, r.method, missingIncident, admin, r.body), http.StatusNotFound, "incident_not_found", incidentNotFound)
			}
			for _, r := range compositionRoutes {
				requireAlike(t, send(t, r.method, malformedIncident+"/"+compositionID+r.suffix, admin, r.body), send(t, r.method, missingIncident+"/"+compositionID+r.suffix, admin, r.body), http.StatusNotFound, "incident_not_found", incidentNotFound)
				requireAlike(t, send(t, r.method, collection+"/"+malformed+r.suffix, admin, r.body), send(t, r.method, missingComposition+r.suffix, admin, r.body), http.StatusNotFound, "not_found", r.notFound)
			}
		}
		// A malformed version, and one past the representable range, are the
		// missing version.
		for _, malformed := range []string{"1", "v0", "V1", "vx", "v01", "v99999999999999999999"} {
			requireAlike(t, send(t, http.MethodGet, collection+"/"+compositionID+"/versions/"+malformed, admin, nil), send(t, http.MethodGet, collection+"/"+compositionID+"/versions/v99", admin, nil), http.StatusNotFound, "not_found", versionNotFound)
		}
	})
	t.Run("an encoded slash stays inside the identifier", func(t *testing.T) {
		missingComposition := collection + "/" + uuid.NewString()
		requireAlike(t, send(t, http.MethodGet, collection+"/"+compositionID+"%2Fversions", admin, nil), send(t, http.MethodGet, missingComposition, admin, nil), http.StatusNotFound, "not_found", compositionNotFound)
		requireAlike(t, send(t, http.MethodPost, collection+"/"+compositionID+"%2Fx/validate", admin, map[string]any{}), send(t, http.MethodPost, missingComposition+"/validate", admin, map[string]any{}), http.StatusNotFound, "not_found", compositionNotFound)
		requireAlike(t, send(t, http.MethodGet, "/api/v1/incidents/"+incidentID+"%2Fx/report-compositions", admin, nil), send(t, http.MethodGet, "/api/v1/incidents/"+uuid.NewString()+"/report-compositions", admin, nil), http.StatusNotFound, "incident_not_found", incidentNotFound)
		requireAlike(t, send(t, http.MethodGet, collection+"/"+compositionID+"/versions/v1%2Fx", admin, nil), send(t, http.MethodGet, collection+"/"+compositionID+"/versions/v99", admin, nil), http.StatusNotFound, "not_found", versionNotFound)
		// POST on the resource template: the encoded slash never selects freeze.
		requireMethodNotAllowed(t, send(t, http.MethodPost, collection+"/"+compositionID+"%2Fversions", admin, draftVersion), http.MethodDelete, http.MethodGet, http.MethodPatch)
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		for path, allow := range map[string]string{collection: "GET, POST", collection + "/" + compositionID: "DELETE, GET, PATCH", collection + "/" + compositionID + "/versions/v1": "GET"} {
			head := httptestx.DoJSON(t, client, http.MethodHead, base+path, nil, httptestx.WithCookies(admin.Session))
			body, err := io.ReadAll(head.Body)
			_ = head.Body.Close()
			if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != allow || len(body) != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q err=%v, want 405 Allow %s without a body", path, head.StatusCode, head.Header.Get("Allow"), body, err, allow)
			}
		}
		requireMethodNotAllowed(t, send(t, http.MethodPut, collection, admin, map[string]any{}), http.MethodGet, http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodGet, collection+"/"+compositionID+"/versions", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodGet, collection+"/"+compositionID+"/validate", admin, nil), http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodGet, collection+"/"+compositionID+"/preview", admin, nil), http.MethodPost)
		requireError(t, send(t, http.MethodGet, collection+"/"+compositionID+"/unknown", admin, nil), http.StatusNotFound, "route_not_found", map[string]any{})
	})
	t.Run("visibility precedes the role gate and query and body validation", func(t *testing.T) {
		missingIncident := "/api/v1/incidents/" + uuid.NewString() + "/report-compositions"
		for _, r := range collectionRoutes {
			requireAlike(t, send(t, r.method, collection+"?unknown=1", outsiderLogin, []any{1}), send(t, r.method, missingIncident+"?unknown=1", outsiderLogin, []any{1}), http.StatusNotFound, "incident_not_found", incidentNotFound)
		}
		for _, r := range compositionRoutes {
			body := any([]any{1})
			if r.method == http.MethodGet {
				body = nil
			}
			requireAlike(t, send(t, r.method, collection+"/"+compositionID+r.suffix+"?limit=1", outsiderLogin, body), send(t, r.method, missingIncident+"/"+compositionID+r.suffix+"?limit=1", outsiderLogin, body), http.StatusNotFound, "incident_not_found", incidentNotFound)
		}
		// A visible viewer is denied the editor routes before the body is
		// validated.
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, collection, viewerLogin, []any{1}), http.StatusForbidden, "authorization_denied")
		for _, r := range compositionRoutes {
			if r.csrf && r.name != "previewReportComposition" {
				httptestx.RequireErrorEnvelope(t, send(t, r.method, collection+"/"+compositionID+r.suffix, viewerLogin, []any{1}), http.StatusForbidden, "authorization_denied")
			}
		}
		// A permitted caller then meets the query and body validation.
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, collection+"?unknown=1", viewerLogin, nil), http.StatusBadRequest, "invalid_request")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, collection+"/"+compositionID+"?limit=1", viewerLogin, nil), http.StatusBadRequest, "invalid_pagination_request")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, collection+"/"+compositionID+"/versions/v1?limit=1", viewerLogin, nil), http.StatusBadRequest, "invalid_pagination_request")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, collection+"/"+compositionID+"/validate", viewerLogin, []any{1}), http.StatusBadRequest, "invalid_request")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, collection, admin, []any{1}), http.StatusBadRequest, "invalid_request")
	})
	if after := effects(t); after != before {
		t.Fatalf("rejected composition admission requests committed an effect:\nbefore=%s\nafter=%s", before, after)
	}
}

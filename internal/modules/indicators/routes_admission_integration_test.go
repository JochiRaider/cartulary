package indicators_test

// Indicator route admission order (Core 01 REQ-01-707, Core 04 AC-617).
import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// TestIndicatorRouteAdmissionOrder_Integration drives the eight Indicator
// operations through central admission. Authentication and CSRF precede path
// parameters; a malformed identifier receives the parameter's declared
// not-found response, identical to a missing resource; an encoded slash stays
// inside its segment; undeclared methods and HEAD are enveloped 405s.
// Visibility precedes the role gate, and both precede query and body
// validation. No rejected request commits an effect.
func TestIndicatorRouteAdmissionOrder_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "indicator-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := appsupport.MustUUID(t, appsupport.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-indicator-admission-order-incident",
		"incident_key":  "IR-IND-ADMISSION-ORDER",
		"title":         "Indicator route admission order",
	})["incident_id"].(string))
	viewerPassword := "IndicatorAdmissionViewer1!"
	viewer := flowtest.SeedLocalUserRecord(t, harness.DB, "indicator-admission-viewer@example.test", "Indicator Admission Viewer", viewerPassword, false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, viewer.ID, "viewer", adminID)
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, viewer.Email, viewerPassword, nil)
	outsiderPassword := "IndicatorAdmissionOutsider1!"
	outsider := flowtest.SeedLocalUserRecord(t, harness.DB, "indicator-admission-outsider@example.test", "Indicator Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, outsiderPassword, nil)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(t testing.TB, method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireAlike := func(t testing.TB, existing, missing *http.Response, code string) {
		t.Helper()
		bodies := make([]map[string]any, 0, 2)
		for _, response := range []*http.Response{existing, missing} {
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, code)
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

	// An Indicator, a source record and one captured observation.
	indicatorRow := httptestx.RequireSuccessData(t, send(t, http.MethodPost, "/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IndicatorsViewSchemaID+"/rows", admin, map[string]any{
		"client_txn_id": "txn-indicator-admission-order-indicator", "indicator.indicator_type": "domain_name",
		"indicator.value_kind": "atomic", "indicator.display_value": "admission-order.example.test",
	}), http.StatusCreated)["row"].(map[string]any)
	indicatorID := indicatorRow["record_id"].(string)
	const sourceText = "admission-source.example"
	sourceID := uuid.New()
	timelinetest.SeedTimelineRecordWithSourceText(t, harness.DB, incidentID, adminID, sourceID, sourceText)
	capture := httptestx.RequireSuccessData(t, send(t, http.MethodPost, "/api/v1/records/"+sourceID.String()+"/indicator-observations", admin, map[string]any{
		"client_txn_id": "txn-indicator-admission-order-capture", "base_row_version": 1, "source_field_key": timelinetest.FieldSourceText,
		"span_start_byte": 0, "span_end_byte": len(sourceText), "resolved_indicator_record_id": indicatorID,
	}), http.StatusCreated)
	observationID := capture["observation"].(map[string]any)["observation_id"].(string)
	httptestx.RequireSuccessEnvelope(t, send(t, http.MethodGet, "/api/v1/indicators/"+strings.ToUpper(indicatorID)+"/observations", viewerLogin, nil), http.StatusOK)

	type route struct {
		name, method string
		at           func(value string) string
		value, code  string
		csrf         bool
	}
	sourcePath := func(value string) string { return "/api/v1/records/" + value + "/indicator-observations" }
	intervalsPath := func(value string) string { return "/api/v1/indicators/" + value + "/state-intervals" }
	actionPath := func(action string) func(string) string {
		return func(value string) string { return "/api/v1/indicator-observations/" + value + "/" + action }
	}
	routes := []route{
		{name: "listSourceRecordIndicatorObservations", method: http.MethodGet, at: sourcePath, value: sourceID.String(), code: "indicator_source_record_not_found"},
		{name: "createManualIndicatorObservation", method: http.MethodPost, at: sourcePath, value: sourceID.String(), code: "indicator_source_record_not_found", csrf: true},
		{name: "listIndicatorObservations", method: http.MethodGet, at: func(value string) string { return "/api/v1/indicators/" + value + "/observations" }, value: indicatorID, code: "indicator_not_found"},
		{name: "listIndicatorStateIntervals", method: http.MethodGet, at: intervalsPath, value: indicatorID, code: "indicator_not_found"},
		{name: "appendIndicatorStateInterval", method: http.MethodPost, at: intervalsPath, value: indicatorID, code: "indicator_not_found", csrf: true},
		{name: "resolveIndicatorObservation", method: http.MethodPost, at: actionPath("resolve"), value: observationID, code: "indicator_observation_not_found", csrf: true},
		{name: "dismissIndicatorObservation", method: http.MethodPost, at: actionPath("dismiss"), value: observationID, code: "indicator_observation_not_found", csrf: true},
		{name: "restoreIndicatorObservation", method: http.MethodPost, at: actionPath("restore"), value: observationID, code: "indicator_observation_not_found", csrf: true},
	}
	// invalid is a request body no route accepts, or none for a read.
	invalid := func(r route) any {
		if r.method == http.MethodGet {
			return nil
		}
		return []any{1}
	}
	type effectState struct {
		observations string
		mutations    workbookscenariotest.MutationEffects
	}
	effects := func(t testing.TB) effectState {
		t.Helper()
		var observations string
		if err := harness.DB.QueryRow(`SELECT coalesce(jsonb_agg(to_jsonb(o) ORDER BY to_jsonb(o)::text),'[]')::text FROM indicator_observations o`).Scan(&observations); err != nil {
			t.Fatalf("read observations: %v", err)
		}
		return effectState{observations: observations, mutations: workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())}
	}
	before := effects(t)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		for _, r := range routes {
			path := base + r.at("not-a-valid-id")
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, path, invalid(r)), http.StatusUnauthorized, "session_required")
			if r.csrf {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, r.method, path, invalid(r), httptestx.WithCookies(admin.Session)), http.StatusForbidden, "csrf_verification_failed")
			}
		}
	})
	t.Run("a malformed identifier is a missing resource", func(t *testing.T) {
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(indicatorID, "-", ""), "%7B" + indicatorID + "%7D"} {
			for _, r := range routes {
				requireAlike(t, send(t, r.method, r.at(malformed), admin, invalid(r)), send(t, r.method, r.at(uuid.NewString()), admin, invalid(r)), r.code)
			}
		}
	})
	t.Run("an encoded slash stays inside the identifier", func(t *testing.T) {
		for _, r := range routes {
			requireAlike(t, send(t, r.method, r.at(r.value+"%2Fx"), admin, invalid(r)), send(t, r.method, r.at(uuid.NewString()), admin, invalid(r)), r.code)
		}
		// No Indicator operation is the parent of these encoded paths.
		requireMethodNotAllowed(t, send(t, http.MethodGet, "/api/v1/records/"+sourceID.String()+"%2Findicator-observations", admin, nil), http.MethodDelete, http.MethodPatch)
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, "/api/v1/indicators/"+indicatorID+"%2Fobservations", admin, nil), http.StatusNotFound, "route_not_found")
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodPost, "/api/v1/indicator-observations/"+observationID+"%2Fdismiss", admin, map[string]any{}), http.StatusNotFound, "route_not_found")
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		for path, allow := range map[string]string{
			sourcePath(sourceID.String()):                         "GET, POST",
			"/api/v1/indicators/" + indicatorID + "/observations": "GET",
			intervalsPath(indicatorID):                            "GET, POST",
		} {
			head := httptestx.DoJSON(t, client, http.MethodHead, base+path, nil, httptestx.WithCookies(admin.Session))
			body, err := io.ReadAll(head.Body)
			_ = head.Body.Close()
			if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != allow || len(body) != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q err=%v, want 405 Allow %s without a body", path, head.StatusCode, head.Header.Get("Allow"), body, err, allow)
			}
		}
		requireMethodNotAllowed(t, send(t, http.MethodPut, intervalsPath(indicatorID), admin, map[string]any{}), http.MethodGet, http.MethodPost)
		requireMethodNotAllowed(t, send(t, http.MethodPost, "/api/v1/indicators/"+indicatorID+"/observations", admin, map[string]any{}), http.MethodGet)
		for _, action := range []string{"resolve", "dismiss", "restore"} {
			requireMethodNotAllowed(t, send(t, http.MethodGet, actionPath(action)(observationID), admin, nil), http.MethodPost)
		}
		httptestx.RequireErrorEnvelope(t, send(t, http.MethodGet, "/api/v1/indicator-observations/"+observationID, admin, nil), http.StatusNotFound, "route_not_found")
	})
	t.Run("visibility precedes the role gate and query and body validation", func(t *testing.T) {
		for _, r := range routes {
			requireAlike(t, send(t, r.method, r.at(r.value)+"?limit=0", outsiderLogin, invalid(r)), send(t, r.method, r.at(uuid.NewString())+"?limit=0", outsiderLogin, invalid(r)), r.code)
		}
		for _, r := range routes {
			if r.method == http.MethodGet {
				// A visible viewer meets the list query validation.
				httptestx.RequireErrorEnvelope(t, send(t, r.method, r.at(r.value)+"?limit=0", viewerLogin, nil), http.StatusBadRequest, "invalid_pagination_request")
				continue
			}
			// A visible viewer is denied every mutation before the body is
			// validated; a permitted caller then meets the body validation.
			httptestx.RequireErrorEnvelope(t, send(t, r.method, r.at(r.value), viewerLogin, invalid(r)), http.StatusForbidden, "authorization_denied")
			httptestx.RequireErrorEnvelope(t, send(t, r.method, r.at(r.value), admin, invalid(r)), http.StatusBadRequest, "invalid_mutation_payload")
		}
	})
	if after := effects(t); after != before {
		t.Fatalf("rejected Indicator admission requests committed an effect:\nbefore=%+v\nafter=%+v", before, after)
	}
}

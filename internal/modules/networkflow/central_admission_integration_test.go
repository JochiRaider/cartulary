package networkflow_test

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

// TestNetworkFlowRoutesAreAdmittedCentrally_Integration covers Core 01
// REQ-01-707 and AC-617 for the Network Flow family bound through the public
// route registry: route resolution precedes authentication, so an undeclared
// method, HEAD included, receives the enveloped 405 method_not_allowed with
// Allow, and an unknown or non-canonical descendant receives route_not_found;
// the incident identifier is admitted in the uuid form; an encoded slash stays
// inside the table identifier; and a session the central authenticator finds
// expired is published to its sockets (REQ-01-277). Table 21-C's own order is
// covered by the major-7 admission matrix.
func TestNetworkFlowRoutesAreAdmittedCentrally_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "central-admission")
	base := harness.Server.HTTP.URL
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "nf-central-admission-incident", "incident_key": "IR-NF-CENTRAL", "title": "Network Flow central admission"})
	incidentID := incident["incident_id"].(string)
	root := base + "/api/v1/incidents/" + incidentID + "/network-flow"
	table, graph := "nft_"+strings.Repeat("a", 32), "nfgv_"+strings.Repeat("b", 32)
	cookies := httptestx.WithCookies(login.Session, login.CSRF)

	t.Run("an undeclared method is rejected before authentication", func(t *testing.T) {
		for _, tc := range []struct {
			method, path string
			allowed      []string
		}{
			{http.MethodPut, "/tables/" + table, []string{"DELETE", "GET", "PATCH"}},
			{http.MethodPost, "/tables", []string{"GET"}},
			{http.MethodOptions, "/graph-views", []string{"GET", "POST"}},
			{http.MethodGet, "/graph-views/" + graph + "/refresh", []string{"POST"}},
			{http.MethodPatch, "/source-profiles", []string{"GET"}},
		} {
			for _, authenticated := range []bool{false, true} {
				var options []func(*http.Request)
				if authenticated {
					options = append(options, cookies)
				}
				response := httptestx.DoJSON(t, http.DefaultClient, tc.method, root+tc.path, nil, options...)
				if got := response.Header.Get("Allow"); got != strings.Join(tc.allowed, ", ") {
					t.Fatalf("%s %s Allow = %q, want %q", tc.method, tc.path, got, strings.Join(tc.allowed, ", "))
				}
				envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
				methods := envelope["error"].(map[string]any)["details"].(map[string]any)["allowed_methods"].([]any)
				if len(methods) != len(tc.allowed) {
					t.Fatalf("%s %s allowed_methods = %v", tc.method, tc.path, methods)
				}
				for index, method := range tc.allowed {
					if methods[index] != method {
						t.Fatalf("%s %s allowed_methods = %v, want %v", tc.method, tc.path, methods, tc.allowed)
					}
				}
			}
		}
	})
	t.Run("HEAD receives a headers-only 405", func(t *testing.T) {
		for _, tc := range []struct{ path, allow string }{
			{"/tables", "GET"},
			{"/source-profiles", "GET"},
			{"/tables/" + table, "DELETE, GET, PATCH"},
			{"/graph-views", "GET, POST"},
			{"/graph-views/" + graph, "DELETE, GET, PATCH"},
			{"/graph-views/" + graph + "/result", "GET"},
		} {
			request, err := http.NewRequest(http.MethodHead, root+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			cookies(request)
			response := httptestx.Do(t, http.DefaultClient, request)
			body, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusMethodNotAllowed || response.Header.Get("Allow") != tc.allow || len(body) != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q", tc.path, response.StatusCode, response.Header.Get("Allow"), body)
			}
		}
	})
	t.Run("an unknown or non-canonical descendant matches no operation", func(t *testing.T) {
		for _, path := range []string{root, root + "/", root + "/tables/" + table + "/unknown", root + "/graphs", root + "//tables", root + "/tables/../tables"} {
			for _, authenticated := range []bool{false, true} {
				var options []func(*http.Request)
				if authenticated {
					options = append(options, cookies)
				}
				response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, path, nil, options...)
				if location := response.Header.Get("Location"); location != "" {
					t.Fatalf("%s redirected to %s", path, location)
				}
				httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "route_not_found")
			}
		}
	})
	t.Run("the incident identifier is admitted in the uuid form", func(t *testing.T) {
		for _, form := range []string{strings.ReplaceAll(incidentID, "-", ""), "%7B" + incidentID + "%7D", "urn:uuid:" + incidentID, incidentID + "%2Fx"} {
			response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, base+"/api/v1/incidents/"+form+"/network-flow/tables", nil, cookies)
			envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "incident_not_found")
			if details := envelope["error"].(map[string]any)["details"].(map[string]any); len(details) != 0 {
				t.Fatalf("%s concealed details = %v", form, details)
			}
		}
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, base+"/api/v1/incidents/"+strings.ToUpper(incidentID)+"/network-flow/tables", nil, cookies)
		httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	})
	t.Run("an encoded slash stays inside the table identifier", func(t *testing.T) {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, root+"/tables/"+table+"%2Fquery", nil, cookies)
		envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "network_flow_table_not_found")
		details := envelope["error"].(map[string]any)["details"].(map[string]any)
		if details["network_flow_table_id"] != nil || details["table_status"] != nil || details["reason_code"] != "not_found" || details["retry_action"] != "refresh_resource" {
			t.Fatalf("malformed table details = %v", details)
		}
		// NF NLSpec Table 21-B: a path-addressed table's not-found carries
		// network_flow_table_id null, so a well-formed missing identifier gets
		// the malformed identifier's response (Core 01 REQ-01-707 step 3).
		requireNetworkFlowResponsesAlike(t,
			httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, root+"/tables/"+table+"%2Fquery", nil, cookies),
			httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, root+"/tables/"+table, nil, cookies),
			"network_flow_table_not_found")
		deleteBody := map[string]any{"client_txn_id": "nf-central-admission-delete", "base_table_version": 1}
		csrf := httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)
		requireNetworkFlowResponsesAlike(t,
			httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, root+"/tables/bad", deleteBody, cookies, csrf),
			httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, root+"/tables/"+table, deleteBody, cookies, csrf),
			"network_flow_table_not_found")
		// A saved graph's not-found carries its Table 21-B1 retry action, on
		// both the malformed and the well-formed missing identifier.
		malformedGraph := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, root+"/graph-views/bad", nil, cookies)
		missingGraph := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, root+"/graph-views/"+graph, nil, cookies)
		graphDetails := requireNetworkFlowResponsesAlike(t, malformedGraph, missingGraph, "network_flow_graph_view_not_found")
		if !reflect.DeepEqual(graphDetails, map[string]any{"field": "graph_view_id", "reason_code": "not_found", "retry_action": "refresh_resource"}) {
			t.Fatalf("graph view not-found details = %v", graphDetails)
		}
		// The encoded slash never selects the refresh operation.
		response = httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, root+"/graph-views/"+graph+"%2Frefresh", map[string]any{}, cookies)
		if got := response.Header.Get("Allow"); got != "DELETE, GET, PATCH" {
			t.Fatalf("encoded refresh Allow = %q", got)
		}
		httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
	})
	t.Run("an expired session is revoked and published to its sockets", func(t *testing.T) {
		const password = "NetworkFlowCentralExpiry1!"
		expiring := flowtest.SeedLocalUserRecord(t, harness.DB, "nf-central-expiring@example.test", "NF central expiring", password, false, false, true)
		scenariotest.CreateMembershipForUser(t, harness.Server, login, incidentID, expiring.ID.String(), expiring.Email, "viewer")
		expiringLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, expiring.Email, password, nil)
		socket := incidentwstest.ConnectAndHello(t, base, incidentID, incidentwstest.ConnectOptions{SessionToken: expiringLogin.Session.Value})
		defer socket.Close(wstest.StatusNormalClosure, "integration_cleanup")
		// Expire the session in storage only: the socket still holds the
		// expiry it saw at connect time, so only the central authenticator's
		// publication can revoke it now.
		if _, err := harness.DB.Exec(`UPDATE user_sessions SET session_expires_at=authenticated_at-interval '1 second', idle_expires_at=authenticated_at-interval '1 second' WHERE user_id=$1 AND revoked_at IS NULL`, expiring.ID); err != nil {
			t.Fatal(err)
		}
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, root+"/tables", nil, httptestx.WithCookies(expiringLogin.Session))
		httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, "session_required")
		incidentwstest.ExpectSessionRevoked(t, socket, "session_expired")
		var reason string
		if err := harness.DB.QueryRow(`SELECT revoke_reason_code FROM user_sessions WHERE user_id=$1`, expiring.ID).Scan(&reason); err != nil || reason != "session_expired" {
			t.Fatalf("revoke_reason_code = %q, %v", reason, err)
		}
	})
}

// requireNetworkFlowResponsesAlike requires two 404 error responses with code
// to be identical in everything but their request identity, and returns their
// details.
func requireNetworkFlowResponsesAlike(t testing.TB, got, want *http.Response, code string) map[string]any {
	t.Helper()
	bodies := make([]map[string]any, 0, 2)
	for _, response := range []*http.Response{got, want} {
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, code)
		delete(body["error"].(map[string]any), "request_id")
		bodies = append(bodies, body)
	}
	if !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("response %v differs from the missing-resource response %v", bodies[0], bodies[1])
	}
	return bodies[0]["error"].(map[string]any)["details"].(map[string]any)
}

package networkflow_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/networkflowroutes"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func assertMajor7AdmissionMatrix(t *testing.T, base string, db postgres.DB, incident, actor uuid.UUID, session, csrf *http.Cookie) {
	ctx := context.Background()
	options := []func(*http.Request){httptestx.WithCookies(session, csrf), httptestx.WithHeader(authn.CSRFHeaderName, csrf.Value)}
	routes := networkflowroutes.All()
	if len(routes) != 19 {
		t.Fatal("update admission matrix for new operations")
	}
	pathFor := func(path, incidentID, tableID, graphID string) string {
		return base + strings.NewReplacer("{incident_id}", incidentID, "{network_flow_table_id}", tableID, "{graph_view_id}", graphID).Replace(path)
	}
	table, graph := "nft_"+strings.Repeat("a", 32), "nfgv_"+strings.Repeat("b", 32)
	for _, route := range routes {
		t.Run(route.RouteID, func(t *testing.T) {
			mutation := strings.HasSuffix(route.RouteID, ".create") || strings.HasSuffix(route.RouteID, ".patch") || strings.HasSuffix(route.RouteID, ".delete") || strings.HasSuffix(route.RouteID, ".refresh")
			bad := pathFor(route.Path, "malformed", "malformed", "malformed") + "?unexpected=1"
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, bad, map[string]any{}), 401, "session_required")
			if mutation {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, bad, map[string]any{}, httptestx.WithCookies(session)), 403, "csrf_verification_failed")
			}
			for _, id := range []string{"malformed", uuid.NewString()} {
				response := httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, pathFor(route.Path, id, "malformed", "malformed")+"?unexpected=1", map[string]any{}, options...), 404, "incident_not_found")
				if len(response["error"].(map[string]any)["details"].(map[string]any)) != 0 {
					t.Fatal("concealed incident details")
				}
			}
			valid := pathFor(route.Path, incident.String(), table, graph)
			envelope := httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, valid+"?unexpected=1", map[string]any{}, options...), 400, "network_flow_invalid_request")
			if envelope["error"].(map[string]any)["details"].(map[string]any)["field"] != "query" {
				t.Fatal("query framing precedence")
			}
			if route.Method == http.MethodGet {
				envelope = httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, valid, map[string]any{}, options...), 400, "network_flow_invalid_request")
				if envelope["error"].(map[string]any)["details"].(map[string]any)["field"] != "body" {
					t.Fatal("bodyless framing precedence")
				}
			} else {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, valid, map[string]any{}, options...), 400, "network_flow_invalid_request")
			}
			if strings.Contains(route.Path, "{network_flow_table_id}") {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, pathFor(route.Path, incident.String(), "bad", graph)+"?unexpected=1", map[string]any{}, options...), 404, "network_flow_table_not_found")
			}
			if strings.Contains(route.Path, "{graph_view_id}") {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, pathFor(route.Path, incident.String(), table, "bad")+"?unexpected=1", map[string]any{}, options...), 404, "network_flow_graph_view_not_found")
			}
			if mutation {
				if _, err := db.Exec(ctx, "UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2", incident, actor); err != nil {
					t.Fatal(err)
				}
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, pathFor(route.Path, incident.String(), "bad", "bad"), map[string]any{}, options...), 403, "authorization_denied")
				if _, err := db.Exec(ctx, "UPDATE incident_memberships SET role='admin' WHERE incident_id=$1 AND user_id=$2", incident, actor); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := db.Exec(ctx, "UPDATE incidents SET status='closed',closed_at=now() WHERE id=$1", incident); err != nil {
				t.Fatal(err)
			}
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, route.Method, pathFor(route.Path, incident.String(), "bad", "bad"), map[string]any{}, options...), 409, "incident_closed")
			if _, err := db.Exec(ctx, "UPDATE incidents SET status='active',closed_at=NULL WHERE id=$1", incident); err != nil {
				t.Fatal(err)
			}
		})
	}
}

package routetest

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func CreateRow(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		body,
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

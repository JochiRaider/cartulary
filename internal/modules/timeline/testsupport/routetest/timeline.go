package routetest

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func CreateRow(t testing.TB, server *httptestx.Server, actor authcookietest.AuthCookies, incidentID string, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		body,
		httptestx.WithCookies(actor.Session, actor.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRF.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

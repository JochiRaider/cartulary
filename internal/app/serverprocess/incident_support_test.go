package serverprocess

import (
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func connectSessionSocket(t testing.TB, server *processtest.Server, login authcookietest.AuthCookies, tag string) *incidentwstest.Client {
	t.Helper()
	incidentID := createSocketIncident(t, server, login, tag)
	return incidentwstest.ConnectAndHello(t, server.BaseURL, incidentID, incidentwstest.ConnectOptions{HTTPClient: server.Client, SessionToken: login.Session.Value})
}

func createSocketIncident(t testing.TB, server *processtest.Server, login authcookietest.AuthCookies, tag string) string {
	t.Helper()
	resp := doJSON(t, server, http.MethodPost, "/api/v1/incidents", map[string]any{
		"client_txn_id": "txn-" + tag,
		"incident_key":  "IR-" + tag,
		"title":         "Process socket " + tag,
	}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	return data["incident_id"].(string)
}

package serverprocess

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func connectSessionSocket(t testing.TB, server *processtest.Server, login flowtest.LoginResult, tag string) *flowtest.SessionSocketClient {
	t.Helper()
	incidentID := createSocketIncident(t, server, login, tag)
	return flowtest.ConnectSessionSocket(t, server.BaseURL, incidentID, login.SessionCookie.Value)
}

func createSocketIncident(t testing.TB, server *processtest.Server, login flowtest.LoginResult, tag string) string {
	t.Helper()
	resp := doJSON(t, server, http.MethodPost, "/api/v1/incidents", map[string]any{
		"client_txn_id": "txn-" + tag,
		"incident_key":  "IR-" + tag,
		"title":         "Process socket " + tag,
	}, withCookies(login.SessionCookie, login.CSRFCookie), withHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	return data["incident_id"].(string)
}

package scenariotest

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func CreateIncident(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func PatchIncident(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		body,
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func CreateMembership(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		body,
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func PatchMembership(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, userID string, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		body,
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func DeleteMembership(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, userID string, body map[string]any) {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodDelete,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+userID,
		body,
		httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	httptestx.RequireStatus(t, resp, http.StatusNoContent)
}

func CreateMembershipForUser(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, userID string, email string, role string) {
	t.Helper()

	created := CreateMembership(t, server, actor, incidentID, map[string]any{
		"client_txn_id": "txn-membership-create-" + userID,
		"email":         email,
		"role":          role,
	})
	if created["user_id"] != userID {
		t.Fatalf("unexpected membership create payload: %#v", created)
	}
}

func UpdateMembershipRole(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, userID string, baseVersion int, role string) {
	t.Helper()

	PatchMembership(t, server, actor, incidentID, userID, map[string]any{
		"base_membership_version": baseVersion,
		"role":                    role,
	})
}

func DeleteMembershipVersion(t testing.TB, server *httptestx.Server, actor flowtest.LoginResult, incidentID string, userID string, baseVersion int64) {
	t.Helper()

	DeleteMembership(t, server, actor, incidentID, userID, map[string]any{
		"base_membership_version": baseVersion,
	})
}

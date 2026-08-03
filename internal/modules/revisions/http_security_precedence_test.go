package revisions_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestRevisionsMutationSecurityPrecedence_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "revisions-security-precedence")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-REVISIONS-PRECEDENCE")
	seedHostProjection(t, harness.DB, incidentID, recordID)
	validBody := `{"base_row_version":1,"client_txn_id":"txn-security-precedence"}`
	malformedBody := `{"base_row_version":`

	unauthenticated := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/not-a-uuid", malformedBody, "text/plain", nil, "")
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	csrfInvalid := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/not-a-uuid", malformedBody, "text/plain", &login, "incorrect-csrf-token")
	httptestx.RequireErrorEnvelope(t, csrfInvalid, http.StatusForbidden, "csrf_verification_failed")

	malformedPath := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/not-a-uuid", malformedBody, "text/plain", &login, login.CSRFCookie.Value)
	httptestx.RequireErrorEnvelope(t, malformedPath, http.StatusNotFound, "incident_not_found")

	hiddenUser := flowtest.SeedLocalUserRecord(t, harness.DB, "revisions-precedence-hidden@example.test", "Revisions Hidden", "HiddenPass1!", false, false, true)
	hiddenLogin := loginLocalUser(t, harness, hiddenUser.Email, "HiddenPass1!")
	hidden := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), malformedBody, "text/plain", &hiddenLogin, hiddenLogin.CSRFCookie.Value)
	httptestx.RequireErrorEnvelope(t, hidden, http.StatusNotFound, "incident_not_found")

	setMembershipRole(t, harness.DB, incidentID, actorID, "viewer")
	wrongRole := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), malformedBody, "text/plain", &login, login.CSRFCookie.Value)
	httptestx.RequireErrorEnvelope(t, wrongRole, http.StatusForbidden, "authorization_denied")

	setMembershipRole(t, harness.DB, incidentID, actorID, "editor")
	invalidMedia := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), validBody, "text/plain", &login, login.CSRFCookie.Value)
	mediaError := httptestx.RequireErrorEnvelope(t, invalidMedia, http.StatusBadRequest, "invalid_mutation_payload")
	if reason := mediaError["error"].(map[string]any)["details"].(map[string]any)["reason_code"]; reason != "invalid_content_type" {
		t.Fatalf("invalid media reason = %#v", reason)
	}

	malformed := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), malformedBody, "application/json", &login, login.CSRFCookie.Value)
	httptestx.RequireErrorEnvelope(t, malformed, http.StatusBadRequest, "invalid_mutation_payload")

	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND row_version = 1 AND deleted_at IS NULL`, recordID); got != 1 {
		t.Fatalf("precedence rejections changed record state, count=%d", got)
	}

	valid := rawRevisionsMutation(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), validBody, "application/json; charset=utf-8", &login, login.CSRFCookie.Value)
	httptestx.RequireSuccessEnvelope(t, valid, http.StatusOK)
}

func rawRevisionsMutation(
	t testing.TB,
	method string,
	url string,
	body string,
	contentType string,
	login *appsupport.LoginResult,
	csrfHeader string,
) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("create raw Revisions request: %v", err)
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if login != nil {
		request.AddCookie(login.SessionCookie)
		request.AddCookie(login.CSRFCookie)
	}
	if csrfHeader != "" {
		request.Header.Set(authn.CSRFHeaderName, csrfHeader)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("perform raw Revisions request: %v", err)
	}
	t.Cleanup(func() {
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	})
	return response
}

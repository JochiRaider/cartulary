package serverprocess

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

const (
	authenticationBootstrapAdminEmail    = "bootstrap-admin@example.test"
	authenticationBootstrapAdminPassword = "BootstrapPass1!"
)

func loginLocalUser(t testing.TB, server *processtest.Server, username string, password string) flowtest.LoginResult {
	t.Helper()
	sessionCookie, csrfCookie := flowtest.LoginLocalUser(t, server.BaseURL, username, password, nil)
	return flowtest.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
}

func loginLocalUserWithSecondFactor(t testing.TB, server *processtest.Server, username string, password string, code string) flowtest.LoginResult {
	t.Helper()
	return flowtest.LoginLocalUserWithSecondFactor(t, server.BaseURL, username, password, code)
}

func requireBootstrapLogin(t testing.TB, server *processtest.Server, username string, password string) string {
	t.Helper()
	return flowtest.RequireBootstrapLogin(t, server.BaseURL, username, password)
}

func createUser(t testing.TB, server *processtest.Server, adminSession *http.Cookie, adminCSRF *http.Cookie, body map[string]any) map[string]any {
	t.Helper()
	resp := doJSON(t, server, http.MethodPost, "/api/v1/users", body, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func provisionBootstrapAdmin(t testing.TB, server *processtest.Server) (flowtest.LoginResult, string) {
	t.Helper()
	bootstrapToken := requireBootstrapLogin(t, server, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword)
	begin := flowtest.BeginTOTPEnrollment(t, server.BaseURL, bootstrapToken, map[string]any{"client_txn_id": "txn-bootstrap-admin-begin"})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	flowtest.CompleteInitialEnrollment(t, server.BaseURL, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	return loginLocalUserWithSecondFactor(t, server, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, flowtest.GenerateTOTPCode(t, secretBase32)), secretBase32
}

func provisionTOTPUser(t testing.TB, server *processtest.Server, adminSession *http.Cookie, adminCSRF *http.Cookie, email string, displayName string, password string) (map[string]any, string) {
	t.Helper()
	user := createUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "create-" + email,
		"auth_kind":        "local",
		"email":            email,
		"display_name":     displayName,
		"initial_password": password,
	})
	bootstrapToken := requireBootstrapLogin(t, server, email, password)
	begin := flowtest.BeginTOTPEnrollment(t, server.BaseURL, bootstrapToken, map[string]any{"client_txn_id": "begin-" + email})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	flowtest.CompleteInitialEnrollment(t, server.BaseURL, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "complete-"+email)
	return user, secretBase32
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return flowtest.WithCookies(cookies...)
}

func withHeader(key string, value string) func(*http.Request) {
	return flowtest.WithHeader(key, value)
}

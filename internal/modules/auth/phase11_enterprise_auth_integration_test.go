package auth_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase1test"
)

func TestPhase11EnterpriseAuthProviderOIDC_I_11_ENTERPRISE_AUTH_01(t *testing.T) {
	claimEnterpriseAuthenticationForTest(t)
	runtime := phase1test.StartRuntime(t)
	server, db := startPhase1Server(t, runtime, "phase11-enterprise-oidc")
	defer db.Close()

	userID := seedLocalUser(t, db, "oidc.enterprise@example.test", "OIDC Enterprise", "EnterpriseOIDC1!", false)
	providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "corp-oidc",
		Type:        "oidc",
		Name:        "Corporate OIDC",
		Enabled:     true,
		Interactive: true,
	})
	seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "disabled-oidc",
		Type:        "oidc",
		Name:        "Disabled OIDC",
		Enabled:     false,
		Interactive: true,
	})
	seedEnterpriseBinding(t, db, userID, providerID, "Subject-Exact-001", userID)

	listResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/providers", nil)
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)
	providers := listBody["data"].(map[string]any)["providers"].([]any)
	if len(providers) != 1 {
		t.Fatalf("expected only one enabled interactive provider, got %#v", providers)
	}
	if got := providers[0].(map[string]any); got["provider_key"] != "corp-oidc" || got["provider_type"] != "oidc" {
		t.Fatalf("unexpected provider discovery row: %#v", got)
	}

	badBegin := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/providers/corp-oidc/begin", map[string]any{
		"client_txn_id": "forbidden",
	})
	httptestx.RequireErrorEnvelope(t, badBegin, http.StatusBadRequest, "invalid_enterprise_auth_request")

	beginResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/providers/corp-oidc/begin", map[string]any{
		"return_to": "/?incident_id=abc",
	})
	beginBody := httptestx.RequireSuccessEnvelope(t, beginResp, http.StatusOK)
	beginData := beginBody["data"].(map[string]any)
	if beginData["provider_key"] != "corp-oidc" || beginData["provider_type"] != "oidc" {
		t.Fatalf("unexpected begin response: %#v", beginData)
	}
	if _, ok := beginData["auth_transaction_id"]; ok {
		t.Fatalf("begin response must not expose public transaction ids: %#v", beginData)
	}
	enterpriseCookie := requireEnterpriseAuthCookie(t, beginResp.Cookies())
	redirectURL, err := url.Parse(beginData["redirect_url"].(string))
	if err != nil {
		t.Fatalf("parse redirect_url: %v", err)
	}
	state := redirectURL.Query().Get("state")
	nonce := redirectURL.Query().Get("nonce")
	if state == "" || nonce == "" {
		t.Fatalf("begin redirect missing OIDC state/nonce: %s", redirectURL.String())
	}

	callbackResp := doNoRedirect(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/oidc/corp-oidc/callback?state="+url.QueryEscape(state)+"&nonce="+url.QueryEscape(nonce)+"&code=valid-code&subject="+url.QueryEscape("Subject-Exact-001"), nil, withCookies(enterpriseCookie))
	if callbackResp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected OIDC callback 303, got %d body=%#v", callbackResp.StatusCode, httptestx.ReadJSONBody(t, callbackResp))
	}
	if location := callbackResp.Header.Get("Location"); location != "/?incident_id=abc" {
		t.Fatalf("unexpected callback location: got %q", location)
	}
	authCookies := httptestx.RequireAuthCookies(t, callbackResp.Cookies())

	sessionResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(authCookies.Session))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)
	sessionData := sessionBody["data"].(map[string]any)
	if sessionData["user_id"] != userID || sessionData["provider_type"] != "oidc" {
		t.Fatalf("enterprise callback did not converge to same session resource: %#v", sessionData)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM users WHERE email = 'oidc.enterprise@example.test'`); got != 1 {
		t.Fatalf("OIDC callback must not JIT-create users, got %d rows", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE last_auth_at IS NOT NULL`); got != 1 {
		t.Fatalf("expected last_auth_at on resolved active binding, got %d", got)
	}

	replayResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/oidc/corp-oidc/callback?state="+url.QueryEscape(state)+"&nonce="+url.QueryEscape(nonce)+"&code=valid-code&subject="+url.QueryEscape("Subject-Exact-001"), nil, withCookies(enterpriseCookie))
	httptestx.RequireErrorEnvelope(t, replayResp, http.StatusConflict, "enterprise_auth_transaction_rejected")

	noLinkBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	noLinkURL, _ := url.Parse(noLinkBegin.redirectURL)
	noLinkResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/oidc/corp-oidc/callback?state="+url.QueryEscape(noLinkURL.Query().Get("state"))+"&nonce="+url.QueryEscape(noLinkURL.Query().Get("nonce"))+"&code=valid-code&subject="+url.QueryEscape("missing-subject"), nil, withCookies(noLinkBegin.cookie))
	httptestx.RequireErrorEnvelope(t, noLinkResp, http.StatusConflict, "provider_identity_rejected")
}

func TestPhase11EnterpriseSAMLACS_I_11_ENTERPRISE_AUTH_02(t *testing.T) {
	claimEnterpriseAuthenticationForTest(t)
	runtime := phase1test.StartRuntime(t)
	server, db := startPhase1Server(t, runtime, "phase11-enterprise-saml")
	defer db.Close()

	userID := seedLocalUser(t, db, "saml.enterprise@example.test", "SAML Enterprise", "EnterpriseSAML1!", false)
	providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "corp-saml",
		Type:        "saml",
		Name:        "Corporate SAML",
		Enabled:     true,
		Interactive: true,
		Issuer:      "https://idp.example.test",
		Audience:    "cartulary-sp",
	})
	seedEnterpriseBinding(t, db, userID, providerID, "saml-subject-001", userID)

	begin := beginEnterpriseAuth(t, server, "corp-saml")
	redirectURL, err := url.Parse(begin.redirectURL)
	if err != nil {
		t.Fatalf("parse SAML redirect_url: %v", err)
	}
	relayState := redirectURL.Query().Get("RelayState")
	if relayState == "" {
		t.Fatalf("SAML begin redirect missing RelayState: %s", redirectURL.String())
	}
	assertion := fakeSAMLResponse(t, map[string]any{
		"subject":         "saml-subject-001",
		"issuer":          "https://idp.example.test",
		"audience":        "cartulary-sp",
		"signature_valid": true,
		"expires_at":      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	})
	acsResp := doFormNoRedirect(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
		"RelayState":   []string{relayState},
		"SAMLResponse": []string{assertion},
	})
	if acsResp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected SAML ACS 303, got %d body=%#v", acsResp.StatusCode, httptestx.ReadJSONBody(t, acsResp))
	}
	if hasCookie(acsResp.Cookies(), "cartulary_session") {
		t.Fatalf("SAML ACS must not issue the session directly: %#v", acsResp.Cookies())
	}
	completionLocation := acsResp.Header.Get("Location")
	if !strings.HasPrefix(completionLocation, "/api/v1/auth/saml/corp-saml/acs/complete?completion=") {
		t.Fatalf("unexpected SAML completion redirect: %q", completionLocation)
	}

	noCookieCompletion := doNoRedirect(t, http.MethodGet, server.HTTP.URL+completionLocation, nil)
	httptestx.RequireErrorEnvelope(t, noCookieCompletion, http.StatusConflict, "enterprise_auth_transaction_rejected")

	wrongCompletion := doNoRedirect(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs/complete?completion=wrong", nil, withCookies(begin.cookie))
	httptestx.RequireErrorEnvelope(t, wrongCompletion, http.StatusConflict, "enterprise_auth_transaction_rejected")

	completionResp := doNoRedirect(t, http.MethodGet, server.HTTP.URL+completionLocation, nil, withCookies(begin.cookie))
	if completionResp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected SAML completion 303, got %d body=%#v", completionResp.StatusCode, httptestx.ReadJSONBody(t, completionResp))
	}
	authCookies := httptestx.RequireAuthCookies(t, completionResp.Cookies())
	sessionResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(authCookies.Session))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if sessionData["provider_type"] != "saml" || sessionData["user_id"] != userID {
		t.Fatalf("unexpected SAML session data: %#v", sessionData)
	}

	replayResp := doNoRedirect(t, http.MethodGet, server.HTTP.URL+completionLocation, nil, withCookies(begin.cookie))
	httptestx.RequireErrorEnvelope(t, replayResp, http.StatusConflict, "enterprise_auth_transaction_rejected")

	failBegin := beginEnterpriseAuth(t, server, "corp-saml")
	failURL, _ := url.Parse(failBegin.redirectURL)
	badAssertion := fakeSAMLResponse(t, map[string]any{
		"subject":         "saml-subject-001",
		"issuer":          "https://idp.example.test",
		"audience":        "cartulary-sp",
		"signature_valid": false,
		"expires_at":      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	})
	badACS := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
		"RelayState":   []string{failURL.Query().Get("RelayState")},
		"SAMLResponse": []string{badAssertion},
	})
	httptestx.RequireErrorEnvelope(t, badACS, http.StatusConflict, "provider_response_rejected")

	idpInitiated := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
		"SAMLResponse": []string{assertion},
	})
	httptestx.RequireErrorEnvelope(t, idpInitiated, http.StatusConflict, "enterprise_auth_transaction_rejected")
}

func TestPhase11EnterpriseAuthBindingLifecycle_I_11_ENTERPRISE_AUTH_03(t *testing.T) {
	claimEnterpriseAuthenticationForTest(t)
	runtime := phase1test.StartRuntime(t)
	server, db := startPhase1Server(t, runtime, "phase11-enterprise-bindings")
	defer db.Close()

	adminID := seedLocalUserFlags(t, db, "enterprise.admin@example.test", "Enterprise Admin", "EnterpriseAdmin1!", false, true, true)
	targetID := seedLocalUser(t, db, "binding.target@example.test", "Binding Target", "BindingTarget1!", false)
	otherID := seedLocalUser(t, db, "binding.other@example.test", "Binding Other", "BindingOther1!", false)
	_ = otherID
	providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "corp-bind",
		Type:        "oidc",
		Name:        "Corporate Bind",
		Enabled:     false,
		Interactive: false,
	})
	adminSession, adminCSRF := loginLocalUser(t, server, "enterprise.admin@example.test", "EnterpriseAdmin1!", nil)

	createResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-create-binding",
		"provider_key":      "corp-bind",
		"provider_subject":  "BindingSubject",
		"reason":            "",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	bindings := createData["auth_bindings"].([]any)
	if len(bindings) != 2 {
		t.Fatalf("safe user must include local plus enterprise binding, got %#v", bindings)
	}
	localBinding := bindings[0].(map[string]any)
	enterpriseBinding := bindings[1].(map[string]any)
	if localBinding["provider_type"] != "local" || localBinding["username"] != "binding.target@example.test" {
		t.Fatalf("unexpected local binding summary: %#v", localBinding)
	}
	authBindingID := enterpriseBinding["auth_binding_id"].(string)
	if enterpriseBinding["provider_subject"] != "BindingSubject" || enterpriseBinding["provider_type"] != "oidc" {
		t.Fatalf("unexpected enterprise binding summary: %#v", enterpriseBinding)
	}

	replayResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": 1,
		"client_txn_id":     "txn-create-binding",
		"provider_key":      "corp-bind",
		"provider_subject":  "BindingSubject",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)

	targetSession, _ := loginLocalUser(t, server, "binding.target@example.test", "BindingTarget1!", nil)
	activeBeforeRotate := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text = $1 AND revoked_at IS NULL`, targetID)
	if activeBeforeRotate == 0 {
		t.Fatal("expected active target session before rotate")
	}
	rotateResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings/"+authBindingID+"/rotate", map[string]any{
		"base_user_version":    queryUserVersion(t, db, targetID),
		"client_txn_id":        "txn-rotate-binding",
		"new_provider_subject": "BindingSubjectRotated",
		"reason":               "admin rotation",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	rotateData := httptestx.RequireSuccessEnvelope(t, rotateResp, http.StatusOK)["data"].(map[string]any)
	rotatedBindings := rotateData["auth_bindings"].([]any)
	if got := rotatedBindings[1].(map[string]any)["provider_subject"]; got != "BindingSubjectRotated" {
		t.Fatalf("rotate did not expose replacement subject: %#v", rotatedBindings)
	}
	revokedAfterRotate := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text = $1 AND revoked_at IS NOT NULL`, targetID)
	if revokedAfterRotate < activeBeforeRotate {
		t.Fatalf("rotate must revoke target sessions, revoked=%d before=%d", revokedAfterRotate, activeBeforeRotate)
	}
	sessionAfterRotate := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(targetSession))
	httptestx.RequireErrorEnvelope(t, sessionAfterRotate, http.StatusUnauthorized, "session_required")

	newAuthBindingID := rotatedBindings[1].(map[string]any)["auth_binding_id"].(string)
	_, _ = loginLocalUser(t, server, "binding.target@example.test", "BindingTarget1!", nil)
	retireResp := doJSON(t, http.MethodDelete, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings/"+newAuthBindingID, map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-retire-binding",
		"reason":            "retire provider",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	retireData := httptestx.RequireSuccessEnvelope(t, retireResp, http.StatusOK)["data"].(map[string]any)
	if got := len(retireData["auth_bindings"].([]any)); got != 1 {
		t.Fatalf("retire must remove enterprise binding from safe user, got %#v", retireData["auth_bindings"])
	}

	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE user_id::text = $1 AND retired_at IS NULL`, targetID); got != 0 {
		t.Fatalf("expected no active enterprise bindings after retire, got %d", got)
	}
	for _, source := range []string{"users.auth_bindings.create", "users.auth_bindings.rotate", "users.auth_bindings.retire"} {
		if got := queryCount(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE actor_user_id::text = $1 AND target_user_id::text = $2 AND event_source = $3`, adminID, targetID, source); got != 1 {
			t.Fatalf("expected one audit event for %s, got %d", source, got)
		}
	}

	seedEnterpriseBinding(t, db, otherID, providerID, "OtherSubject", adminID)
	conflictResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-create-conflict",
		"provider_key":      "corp-bind",
		"provider_subject":  "OtherSubject",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, conflictResp, http.StatusConflict, "auth_binding_conflict")
}

type enterpriseProviderSeed struct {
	Key         string
	Type        string
	Name        string
	Enabled     bool
	Interactive bool
	Issuer      string
	Audience    string
}

type enterpriseBeginResult struct {
	cookie      *http.Cookie
	redirectURL string
}

func claimEnterpriseAuthenticationForTest(t testing.TB) {
	t.Helper()
	profiles := httpapi.CurrentExtensionProfiles()
	for index := range profiles {
		if profiles[index].ProfileID == "enterprise_authentication" {
			profiles[index].Claimed = true
		}
	}
	restore := httpapi.SetCurrentExtensionProfilesForTesting(profiles)
	t.Cleanup(restore)
}

func seedEnterpriseProvider(t testing.TB, db *sql.DB, seed enterpriseProviderSeed) string {
	t.Helper()
	var providerID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO enterprise_auth_providers (provider_key, provider_type, display_name, is_enabled, is_interactive, issuer, audience)
VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''))
RETURNING id::text
`, seed.Key, seed.Type, seed.Name, seed.Enabled, seed.Interactive, seed.Issuer, seed.Audience).Scan(&providerID); err != nil {
		t.Fatalf("seed enterprise provider: %v", err)
	}
	return providerID
}

func seedEnterpriseBinding(t testing.TB, db *sql.DB, userID string, providerID string, subject string, actorUserID string) string {
	t.Helper()
	var bindingID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO enterprise_auth_bindings (user_id, provider_id, provider_key, provider_type, provider_subject, created_by_user_id)
SELECT $1::uuid, p.id, p.provider_key, p.provider_type, $3, $4::uuid
  FROM enterprise_auth_providers p
 WHERE p.id::text = $2
RETURNING id::text
`, userID, providerID, subject, actorUserID).Scan(&bindingID); err != nil {
		t.Fatalf("seed enterprise binding: %v", err)
	}
	return bindingID
}

func beginEnterpriseAuth(t testing.TB, server *httptestx.Server, providerKey string) enterpriseBeginResult {
	t.Helper()
	resp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/providers/"+providerKey+"/begin", map[string]any{})
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	return enterpriseBeginResult{
		cookie:      requireEnterpriseAuthCookie(t, resp.Cookies()),
		redirectURL: body["data"].(map[string]any)["redirect_url"].(string),
	}
}

func requireEnterpriseAuthCookie(t testing.TB, cookies []*http.Cookie) *http.Cookie {
	t.Helper()
	cookie := requireCookie(t, cookies, "cartulary_enterprise_auth_txn")
	if !cookie.HttpOnly {
		t.Fatalf("enterprise auth cookie must be HttpOnly: %#v", cookie)
	}
	if !cookie.Secure {
		t.Fatalf("enterprise auth cookie must be Secure: %#v", cookie)
	}
	if cookie.Path != "/api/v1/auth" {
		t.Fatalf("enterprise auth cookie must be scoped to /api/v1/auth, got %q", cookie.Path)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("enterprise auth cookie must use SameSite=Lax, got %v", cookie.SameSite)
	}
	return cookie
}

func requireCookie(t testing.TB, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("missing cookie %s in %#v", name, cookies)
	return nil
}

func hasCookie(cookies []*http.Cookie, name string) bool {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return true
		}
	}
	return false
}

func queryUserVersion(t testing.TB, db *sql.DB, userID string) int64 {
	t.Helper()
	var version int64
	if err := db.QueryRowContext(context.Background(), `SELECT user_version FROM users WHERE id::text = $1`, userID).Scan(&version); err != nil {
		t.Fatalf("query user version: %v", err)
	}
	return version
}

func fakeSAMLResponse(t testing.TB, payload map[string]any) string {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal fake SAML response: %v", err)
	}
	return string(data)
}

func doNoRedirect(t testing.TB, method string, rawURL string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()
	req := httptestx.NewJSONRequest(t, method, rawURL, body)
	for _, option := range options {
		option(req)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return httptestx.Do(t, client, req)
}

func doForm(t testing.TB, rawURL string, values url.Values, options ...func(*http.Request)) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatalf("new form request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func doFormNoRedirect(t testing.TB, rawURL string, values url.Values, options ...func(*http.Request)) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatalf("new form request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, option := range options {
		option(req)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return httptestx.Do(t, client, req)
}

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
	localSession, localCSRF := loginLocalUser(t, server, "oidc.enterprise@example.test", "EnterpriseOIDC1!", nil)
	providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "corp-oidc",
		Type:        "oidc",
		Name:        "Corporate OIDC",
		Enabled:     true,
		Interactive: true,
	})
	seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "alpha-saml",
		Type:        "saml",
		Name:        "Alpha SAML",
		Enabled:     true,
		Interactive: true,
	})
	seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "corp-oidc-b",
		Type:        "oidc",
		Name:        "Corporate OIDC",
		Enabled:     true,
		Interactive: true,
	})
	otherProviderID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "other-oidc",
		Type:        "oidc",
		Name:        "Other OIDC",
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
	seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "hidden-oidc",
		Type:        "oidc",
		Name:        "Hidden OIDC",
		Enabled:     true,
		Interactive: false,
	})
	seedEnterpriseBinding(t, db, userID, providerID, "Subject-Exact-001", userID)

	listResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/providers", nil)
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)
	providers := listBody["data"].(map[string]any)["providers"].([]any)
	if len(providers) != 4 {
		t.Fatalf("expected only enabled interactive providers, got %#v", providers)
	}
	wantProviderKeys := []string{"alpha-saml", "corp-oidc", "corp-oidc-b", "other-oidc"}
	for index, raw := range providers {
		got := raw.(map[string]any)
		requireExactMapKeys(t, got, "display_name", "provider_key", "provider_type")
		if got["provider_key"] != wantProviderKeys[index] {
			t.Fatalf("provider discovery order mismatch at %d: got %#v want %q", index, got, wantProviderKeys[index])
		}
	}
	requireJSONDoesNotContain(t, listBody, "client_secret", "issuer", "audience", "claim_map", "authorization_endpoint", "private_key")

	pagedList := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/providers?limit=1", nil)
	requireErrorReason(t, pagedList, http.StatusBadRequest, "invalid_pagination_request", "pagination_not_supported")

	badBeginObject := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/providers/corp-oidc/begin", "not-object")
	requireErrorReason(t, badBeginObject, http.StatusBadRequest, "invalid_enterprise_auth_request", "request_not_object")

	badBeginField := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/providers/corp-oidc/begin", map[string]any{
		"client_txn_id": "forbidden",
	})
	requireErrorReason(t, badBeginField, http.StatusBadRequest, "invalid_enterprise_auth_request", "unknown_field")

	for _, returnTo := range []any{"", "https://evil.example.test/callback", "//evil.example.test/callback", "/safe#fragment", "relative/path"} {
		beginResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/providers/corp-oidc/begin", map[string]any{"return_to": returnTo})
		requireErrorReason(t, beginResp, http.StatusBadRequest, "invalid_enterprise_auth_request", "return_to_not_allowed")
	}

	defaultBegin := beginEnterpriseAuthWithBody(t, server, "corp-oidc", map[string]any{})
	nullBegin := beginEnterpriseAuthWithBody(t, server, "corp-oidc", map[string]any{"return_to": nil})
	defaultRedirect := parseEnterpriseRedirect(t, defaultBegin.redirectURL)
	nullRedirect := parseEnterpriseRedirect(t, nullBegin.redirectURL)
	if defaultRedirect.Query().Get("state") == nullRedirect.Query().Get("state") || defaultBegin.cookie.Value == nullBegin.cookie.Value {
		t.Fatalf("provider begin must mint fresh non-idempotent transactions")
	}
	nullCallbackResp := doNoRedirect(
		t,
		http.MethodGet,
		oidcCallbackURL(server, "corp-oidc", nullRedirect.Query().Get("state"), nullRedirect.Query().Get("nonce"), "valid-code", "Subject-Exact-001"),
		nil,
		withCookies(nullBegin.cookie),
	)
	if nullCallbackResp.StatusCode != http.StatusSeeOther || nullCallbackResp.Header.Get("Location") != "/" {
		t.Fatalf("null return_to must normalize to root, got status=%d location=%q", nullCallbackResp.StatusCode, nullCallbackResp.Header.Get("Location"))
	}

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
	redirectURL := parseEnterpriseRedirect(t, beginData["redirect_url"].(string))
	state := redirectURL.Query().Get("state")
	nonce := redirectURL.Query().Get("nonce")
	if state == "" || nonce == "" {
		t.Fatalf("begin redirect missing OIDC state/nonce: %s", redirectURL.String())
	}

	callbackResp := doNoRedirect(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", state, nonce, "valid-code", "Subject-Exact-001"), nil, withCookies(enterpriseCookie))
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
	if got := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text = $1 AND provider_type = 'oidc' AND auth_binding_id IS NOT NULL AND revoked_at IS NULL`, userID); got == 0 {
		t.Fatalf("OIDC callback session must preserve binding attribution in user_sessions")
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM users WHERE email = 'oidc.enterprise@example.test'`); got != 1 {
		t.Fatalf("OIDC callback must not JIT-create users, got %d rows", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE user_id::text = $1`, userID); got != 0 {
		t.Fatalf("OIDC callback must not provision incident memberships, got %d rows", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE last_auth_at IS NOT NULL`); got != 1 {
		t.Fatalf("expected last_auth_at on resolved active binding, got %d", got)
	}

	incident := createIncidentResource(t, server, localSession, localCSRF, map[string]any{
		"client_txn_id": "txn-enterprise-audit-lineage-create",
		"incident_key":  "IR-EAUTH-LINEAGE",
		"title":         "Enterprise Auth Lineage",
	})
	incidentID := incident["incident_id"].(string)
	lineagePatch := doJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "amber",
		},
		withCookies(authCookies.Session, authCookies.CSRF),
		withHeader(authn.CSRFHeaderName, authCookies.CSRF.Value),
	)
	lineageData := httptestx.RequireSuccessEnvelope(t, lineagePatch, http.StatusOK)["data"].(map[string]any)
	if lineageData["updated_by_user_id"] != userID {
		t.Fatalf("enterprise login must preserve existing incident attribution user_id, got %#v", lineageData)
	}

	replayResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", state, nonce, "valid-code", "Subject-Exact-001"), nil, withCookies(enterpriseCookie))
	requireErrorReason(t, replayResp, http.StatusConflict, "enterprise_auth_transaction_rejected", "already_used")
	requireNoSessionCookie(t, replayResp)

	missingStateBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	missingStateURL := parseEnterpriseRedirect(t, missingStateBegin.redirectURL)
	missingStateResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", "", missingStateURL.Query().Get("nonce"), "valid-code", "Subject-Exact-001"), nil, withCookies(missingStateBegin.cookie))
	requireErrorReason(t, missingStateResp, http.StatusConflict, "provider_response_rejected", "missing_required_field")
	requireNoSessionCookie(t, missingStateResp)

	badCodeBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	badCodeURL := parseEnterpriseRedirect(t, badCodeBegin.redirectURL)
	badCodeResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", badCodeURL.Query().Get("state"), badCodeURL.Query().Get("nonce"), "invalid-code", "Subject-Exact-001"), nil, withCookies(badCodeBegin.cookie))
	requireErrorReason(t, badCodeResp, http.StatusConflict, "provider_response_rejected", "code_exchange_failed")
	requireNoSessionCookie(t, badCodeResp)

	badStateBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	badStateURL := parseEnterpriseRedirect(t, badStateBegin.redirectURL)
	badStateResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", "wrong-state", badStateURL.Query().Get("nonce"), "valid-code", "Subject-Exact-001"), nil, withCookies(badStateBegin.cookie))
	requireErrorReason(t, badStateResp, http.StatusConflict, "provider_response_rejected", "state_mismatch")
	requireNoSessionCookie(t, badStateResp)

	badNonceBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	badNonceURL := parseEnterpriseRedirect(t, badNonceBegin.redirectURL)
	badNonceResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", badNonceURL.Query().Get("state"), "wrong-nonce", "valid-code", "Subject-Exact-001"), nil, withCookies(badNonceBegin.cookie))
	requireErrorReason(t, badNonceResp, http.StatusConflict, "provider_response_rejected", "nonce_mismatch")
	requireNoSessionCookie(t, badNonceResp)

	mismatchBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	mismatchURL := parseEnterpriseRedirect(t, mismatchBegin.redirectURL)
	mismatchResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "other-oidc", mismatchURL.Query().Get("state"), mismatchURL.Query().Get("nonce"), "valid-code", "Subject-Exact-001"), nil, withCookies(mismatchBegin.cookie))
	requireErrorReason(t, mismatchResp, http.StatusConflict, "enterprise_auth_transaction_rejected", "provider_mismatch")
	requireNoSessionCookie(t, mismatchResp)

	expiredBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	expiredURL := parseEnterpriseRedirect(t, expiredBegin.redirectURL)
	httptestx.AdvanceClock(t, server, authn.EnterpriseAuthTransactionTTL+time.Second)
	expiredResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", expiredURL.Query().Get("state"), expiredURL.Query().Get("nonce"), "valid-code", "Subject-Exact-001"), nil, withCookies(expiredBegin.cookie))
	requireErrorReason(t, expiredResp, http.StatusConflict, "enterprise_auth_transaction_rejected", "expired")
	requireNoSessionCookie(t, expiredResp)

	noLinkBegin := beginEnterpriseAuth(t, server, "corp-oidc")
	noLinkURL := parseEnterpriseRedirect(t, noLinkBegin.redirectURL)
	beforeUsers := queryCount(t, db, `SELECT COUNT(*) FROM users`)
	beforeBindings := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings`)
	beforeMemberships := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships`)
	noLinkResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-oidc", noLinkURL.Query().Get("state"), noLinkURL.Query().Get("nonce"), "valid-code", "missing-subject"), nil, withCookies(noLinkBegin.cookie))
	requireErrorReason(t, noLinkResp, http.StatusConflict, "provider_identity_rejected", "no_linked_user")
	requireNoSessionCookie(t, noLinkResp)
	if got := queryCount(t, db, `SELECT COUNT(*) FROM users`); got != beforeUsers {
		t.Fatalf("unknown OIDC subject must not create users, before=%d after=%d", beforeUsers, got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings`); got != beforeBindings {
		t.Fatalf("unknown OIDC subject must not create auth bindings, before=%d after=%d", beforeBindings, got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships`); got != beforeMemberships {
		t.Fatalf("OIDC claims must not create incident memberships, before=%d after=%d", beforeMemberships, got)
	}

	inactiveUserID := seedLocalUserFlags(t, db, "oidc.inactive@example.test", "Inactive OIDC", "InactiveOIDC1!", false, false, false)
	seedEnterpriseBinding(t, db, inactiveUserID, otherProviderID, "inactive-subject", userID)
	inactiveBegin := beginEnterpriseAuth(t, server, "other-oidc")
	inactiveURL := parseEnterpriseRedirect(t, inactiveBegin.redirectURL)
	inactiveResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "other-oidc", inactiveURL.Query().Get("state"), inactiveURL.Query().Get("nonce"), "valid-code", "inactive-subject"), nil, withCookies(inactiveBegin.cookie))
	requireErrorReason(t, inactiveResp, http.StatusConflict, "provider_identity_rejected", "inactive_user")
	requireNoSessionCookie(t, inactiveResp)
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
	otherProviderID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "other-saml",
		Type:        "saml",
		Name:        "Other SAML",
		Enabled:     true,
		Interactive: true,
		Issuer:      "https://idp.example.test",
		Audience:    "cartulary-sp",
	})
	_ = otherProviderID
	seedEnterpriseBinding(t, db, userID, providerID, "saml-subject-001", userID)

	begin := beginEnterpriseAuth(t, server, "corp-saml")
	redirectURL := parseEnterpriseRedirect(t, begin.redirectURL)
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
	requireErrorReason(t, noCookieCompletion, http.StatusConflict, "enterprise_auth_transaction_rejected", "browser_binding_mismatch")
	requireNoSessionCookie(t, noCookieCompletion)

	wrongCompletion := doNoRedirect(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs/complete?completion=wrong", nil, withCookies(begin.cookie))
	requireErrorReason(t, wrongCompletion, http.StatusConflict, "enterprise_auth_transaction_rejected", "completion_mismatch")
	requireNoSessionCookie(t, wrongCompletion)

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
	if got := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text = $1 AND provider_type = 'saml' AND auth_binding_id IS NOT NULL AND revoked_at IS NULL`, userID); got == 0 {
		t.Fatalf("SAML callback session must preserve binding attribution in user_sessions")
	}

	replayResp := doNoRedirect(t, http.MethodGet, server.HTTP.URL+completionLocation, nil, withCookies(begin.cookie))
	requireErrorReason(t, replayResp, http.StatusConflict, "enterprise_auth_transaction_rejected", "already_used")
	requireNoSessionCookie(t, replayResp)

	failBegin := beginEnterpriseAuth(t, server, "corp-saml")
	failURL := parseEnterpriseRedirect(t, failBegin.redirectURL)
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
	requireErrorReason(t, badACS, http.StatusConflict, "provider_response_rejected", "signature_invalid")
	requireNoSessionCookie(t, badACS)

	idpInitiated := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
		"SAMLResponse": []string{assertion},
	})
	requireErrorReason(t, idpInitiated, http.StatusConflict, "enterprise_auth_transaction_rejected", "not_found")
	requireNoSessionCookie(t, idpInitiated)

	for _, tc := range []struct {
		name      string
		assertion map[string]any
		code      string
		reason    string
	}{
		{
			name: "issuer",
			assertion: map[string]any{
				"subject":         "saml-subject-001",
				"issuer":          "https://wrong-idp.example.test",
				"audience":        "cartulary-sp",
				"signature_valid": true,
				"expires_at":      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
			},
			code:   "provider_response_rejected",
			reason: "issuer_mismatch",
		},
		{
			name: "audience",
			assertion: map[string]any{
				"subject":         "saml-subject-001",
				"issuer":          "https://idp.example.test",
				"audience":        "wrong-sp",
				"signature_valid": true,
				"expires_at":      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
			},
			code:   "provider_response_rejected",
			reason: "audience_mismatch",
		},
		{
			name: "expired_assertion",
			assertion: map[string]any{
				"subject":         "saml-subject-001",
				"issuer":          "https://idp.example.test",
				"audience":        "cartulary-sp",
				"signature_valid": true,
				"expires_at":      time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
			},
			code:   "provider_response_rejected",
			reason: "assertion_expired",
		},
		{
			name: "missing_subject",
			assertion: map[string]any{
				"issuer":          "https://idp.example.test",
				"audience":        "cartulary-sp",
				"signature_valid": true,
				"expires_at":      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
			},
			code:   "provider_identity_rejected",
			reason: "subject_missing",
		},
	} {
		testBegin := beginEnterpriseAuth(t, server, "corp-saml")
		testURL := parseEnterpriseRedirect(t, testBegin.redirectURL)
		resp := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
			"RelayState":   []string{testURL.Query().Get("RelayState")},
			"SAMLResponse": []string{fakeSAMLResponse(t, tc.assertion)},
		})
		requireErrorReason(t, resp, http.StatusConflict, tc.code, tc.reason)
		requireNoSessionCookie(t, resp)
	}

	missingResponseBegin := beginEnterpriseAuth(t, server, "corp-saml")
	missingResponseURL := parseEnterpriseRedirect(t, missingResponseBegin.redirectURL)
	missingResponse := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
		"RelayState": []string{missingResponseURL.Query().Get("RelayState")},
	})
	requireErrorReason(t, missingResponse, http.StatusConflict, "provider_response_rejected", "missing_required_field")
	requireNoSessionCookie(t, missingResponse)

	relayMismatchBegin := beginEnterpriseAuth(t, server, "corp-saml")
	relayMismatchResp := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
		"RelayState":   []string{"wrong-relay-state"},
		"SAMLResponse": []string{assertion},
	})
	_ = relayMismatchBegin
	requireErrorReason(t, relayMismatchResp, http.StatusConflict, "provider_response_rejected", "relay_state_mismatch")
	requireNoSessionCookie(t, relayMismatchResp)

	providerMismatchBegin := beginEnterpriseAuth(t, server, "corp-saml")
	providerMismatchURL := parseEnterpriseRedirect(t, providerMismatchBegin.redirectURL)
	providerMismatchResp := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/other-saml/acs", url.Values{
		"RelayState":   []string{providerMismatchURL.Query().Get("RelayState")},
		"SAMLResponse": []string{assertion},
	})
	requireErrorReason(t, providerMismatchResp, http.StatusConflict, "enterprise_auth_transaction_rejected", "provider_mismatch")
	requireNoSessionCookie(t, providerMismatchResp)

	expiredBegin := beginEnterpriseAuth(t, server, "corp-saml")
	expiredURL := parseEnterpriseRedirect(t, expiredBegin.redirectURL)
	httptestx.AdvanceClock(t, server, authn.EnterpriseAuthTransactionTTL+time.Second)
	expiredACS := doForm(t, server.HTTP.URL+"/api/v1/auth/saml/corp-saml/acs", url.Values{
		"RelayState":   []string{expiredURL.Query().Get("RelayState")},
		"SAMLResponse": []string{assertion},
	})
	requireErrorReason(t, expiredACS, http.StatusConflict, "enterprise_auth_transaction_rejected", "expired")
	requireNoSessionCookie(t, expiredACS)
}

func TestPhase11EnterpriseAuthBindingLifecycle_I_11_ENTERPRISE_AUTH_03(t *testing.T) {
	claimEnterpriseAuthenticationForTest(t)
	runtime := phase1test.StartRuntime(t)
	server, db := startPhase1Server(t, runtime, "phase11-enterprise-bindings")
	defer db.Close()

	adminID := seedLocalUserFlags(t, db, "enterprise.admin@example.test", "Enterprise Admin", "EnterpriseAdmin1!", false, true, true)
	targetID := seedLocalUser(t, db, "binding.target@example.test", "Binding Target", "BindingTarget1!", false)
	otherID := seedLocalUser(t, db, "binding.other@example.test", "Binding Other", "BindingOther1!", false)
	caseVariantID := seedLocalUser(t, db, "binding.case@example.test", "Binding Case", "BindingCase1!", false)
	spaceVariantID := seedLocalUser(t, db, "binding.space@example.test", "Binding Space", "BindingSpace1!", false)
	conflictTargetID := seedLocalUser(t, db, "binding.conflict@example.test", "Binding Conflict", "BindingConflict1!", false)
	incidentAdminID := seedLocalUserFlags(t, db, "enterprise.incident.admin@example.test", "Enterprise Incident Admin", "EnterpriseIncidentAdmin1!", false, false, true)
	providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{
		Key:         "corp-bind",
		Type:        "oidc",
		Name:        "Corporate Bind",
		Enabled:     false,
		Interactive: false,
	})
	adminSession, adminCSRF := loginLocalUser(t, server, "enterprise.admin@example.test", "EnterpriseAdmin1!", nil)
	incidentAdminSession, incidentAdminCSRF := loginLocalUser(t, server, "enterprise.incident.admin@example.test", "EnterpriseIncidentAdmin1!", nil)

	incident := createIncidentResource(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id": "txn-enterprise-binding-scope-incident",
		"incident_key":  "IR-EAUTH-BIND",
		"title":         "Enterprise Binding Scope",
	})
	createIncidentMembership(t, server, incident["incident_id"].(string), adminSession, adminCSRF, map[string]any{
		"client_txn_id": "txn-enterprise-binding-scope-membership",
		"email":         "enterprise.incident.admin@example.test",
		"role":          "admin",
	})
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE user_id::text = $1 AND role = 'admin'`, incidentAdminID); got != 1 {
		t.Fatalf("expected incident admin setup, got %d memberships", got)
	}

	incidentAdminDenied := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-incident-admin-denied",
		"provider_key":      "corp-bind",
		"provider_subject":  "DeniedSubject",
	}, withCookies(incidentAdminSession, incidentAdminCSRF), withHeader(authn.CSRFHeaderName, incidentAdminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, incidentAdminDenied, http.StatusUnauthorized, "session_required")

	targetNonAdminSession, targetNonAdminCSRF := loginLocalUser(t, server, "binding.target@example.test", "BindingTarget1!", nil)
	nonAdminDenied := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-non-admin-denied",
		"provider_key":      "corp-bind",
		"provider_subject":  "DeniedSubject",
	}, withCookies(targetNonAdminSession, targetNonAdminCSRF), withHeader(authn.CSRFHeaderName, targetNonAdminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, nonAdminDenied, http.StatusUnauthorized, "session_required")

	for _, tc := range []struct {
		name   string
		body   any
		reason string
	}{
		{name: "request_not_object", body: "not-object", reason: "request_not_object"},
		{name: "unknown_field", body: map[string]any{
			"base_user_version": queryUserVersion(t, db, targetID),
			"client_txn_id":     "txn-create-unknown-field",
			"provider_key":      "corp-bind",
			"provider_subject":  "BindingSubject",
			"extra":             true,
		}, reason: "unknown_field"},
		{name: "missing_required_field", body: map[string]any{
			"base_user_version": queryUserVersion(t, db, targetID),
			"client_txn_id":     "txn-create-missing-subject",
			"provider_key":      "corp-bind",
		}, reason: "missing_required_field"},
		{name: "field_not_nullable", body: map[string]any{
			"base_user_version": queryUserVersion(t, db, targetID),
			"client_txn_id":     "txn-create-null-subject",
			"provider_key":      "corp-bind",
			"provider_subject":  nil,
		}, reason: "field_not_nullable"},
	} {
		resp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", tc.body, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
		requireErrorReason(t, resp, http.StatusBadRequest, "invalid_mutation_payload", tc.reason)
	}

	unknownProvider := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-create-unknown-provider",
		"provider_key":      "missing-provider",
		"provider_subject":  "BindingSubject",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, unknownProvider, http.StatusNotFound, "auth_provider_not_found")

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
	requireExactMapKeys(t, createData, "auth_bindings", "created_at", "display_name", "email", "is_active", "is_deployment_admin", "last_login_at", "mfa_required", "updated_at", "updated_by_user_id", "user_id", "user_version")
	localBinding := bindings[0].(map[string]any)
	enterpriseBinding := bindings[1].(map[string]any)
	requireExactMapKeys(t, localBinding, "created_at", "provider_key", "provider_type", "username")
	requireExactMapKeys(t, enterpriseBinding, "auth_binding_id", "created_at", "last_auth_at", "provider_key", "provider_subject", "provider_type")
	if localBinding["provider_type"] != "local" || localBinding["username"] != "binding.target@example.test" {
		t.Fatalf("unexpected local binding summary: %#v", localBinding)
	}
	authBindingID := enterpriseBinding["auth_binding_id"].(string)
	if enterpriseBinding["provider_subject"] != "BindingSubject" || enterpriseBinding["provider_type"] != "oidc" {
		t.Fatalf("unexpected enterprise binding summary: %#v", enterpriseBinding)
	}
	if enterpriseBinding["last_auth_at"] != nil {
		t.Fatalf("new binding must not synthesize last_auth_at: %#v", enterpriseBinding)
	}
	requireJSONDoesNotContain(t, createData, "password_hash", "client_secret", "SAMLResponse", "id_token", "access_token")

	replayResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": 1,
		"client_txn_id":     "txn-create-binding",
		"provider_key":      "corp-bind",
		"provider_subject":  "BindingSubject",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)

	divergentReplay := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": 1,
		"client_txn_id":     "txn-create-binding",
		"provider_key":      "corp-bind",
		"provider_subject":  "BindingSubjectDivergent",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, divergentReplay, http.StatusConflict, "client_txn_conflict")

	alreadyLinked := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-create-provider-already-linked",
		"provider_key":      "corp-bind",
		"provider_subject":  "AnotherBindingSubject",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	requireErrorReason(t, alreadyLinked, http.StatusConflict, "auth_binding_conflict", "provider_already_linked_for_user")

	seedEnterpriseBinding(t, db, otherID, providerID, "OtherSubject", adminID)
	conflictResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+conflictTargetID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, conflictTargetID),
		"client_txn_id":     "txn-create-conflict",
		"provider_key":      "corp-bind",
		"provider_subject":  "OtherSubject",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	requireErrorReason(t, conflictResp, http.StatusConflict, "auth_binding_conflict", "provider_subject_in_use")

	caseVariantResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+caseVariantID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, caseVariantID),
		"client_txn_id":     "txn-create-case-variant",
		"provider_key":      "corp-bind",
		"provider_subject":  "othersubject",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, caseVariantResp, http.StatusCreated)

	spaceVariantResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+spaceVariantID+"/auth-bindings", map[string]any{
		"base_user_version": queryUserVersion(t, db, spaceVariantID),
		"client_txn_id":     "txn-create-space-variant",
		"provider_key":      "corp-bind",
		"provider_subject":  "OtherSubject ",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, spaceVariantResp, http.StatusCreated)

	initialTargetVersion := queryUserVersion(t, db, targetID)
	sameSubjectResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings/"+authBindingID+"/rotate", map[string]any{
		"base_user_version":    initialTargetVersion,
		"client_txn_id":        "txn-rotate-same-subject",
		"new_provider_subject": "BindingSubject",
		"reason":               "same subject no-op",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	sameSubjectData := httptestx.RequireSuccessEnvelope(t, sameSubjectResp, http.StatusOK)["data"].(map[string]any)
	if sameSubjectData["user_version"] != float64(initialTargetVersion) || queryUserVersion(t, db, targetID) != initialTargetVersion {
		t.Fatalf("same-subject rotate must not advance user version: response=%#v", sameSubjectData)
	}
	sameBindings := sameSubjectData["auth_bindings"].([]any)
	if sameBindings[1].(map[string]any)["auth_binding_id"] != authBindingID {
		t.Fatalf("same-subject rotate must not re-key the binding: %#v", sameBindings)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE user_id::text = $1 AND provider_subject = 'BindingSubject' AND retired_at IS NULL`, targetID); got != 1 {
		t.Fatalf("same-subject rotate must preserve one active binding row, got %d", got)
	}

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
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE id::text = $1 AND retired_at IS NOT NULL AND replaced_by_auth_binding_id::text = $2`, authBindingID, newAuthBindingID); got != 1 {
		t.Fatalf("rotate must retire old row and link replacement, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE id::text = $1 AND provider_subject = 'BindingSubjectRotated' AND retired_at IS NULL`, newAuthBindingID); got != 1 {
		t.Fatalf("rotate must create one active replacement row, got %d", got)
	}

	if _, err := db.ExecContext(context.Background(), `UPDATE enterprise_auth_providers SET is_enabled = true, is_interactive = true WHERE id::text = $1`, providerID); err != nil {
		t.Fatalf("enable binding provider for callback validation: %v", err)
	}
	oldSubjectBegin := beginEnterpriseAuth(t, server, "corp-bind")
	oldSubjectURL := parseEnterpriseRedirect(t, oldSubjectBegin.redirectURL)
	oldSubjectResp := doJSON(t, http.MethodGet, oidcCallbackURL(server, "corp-bind", oldSubjectURL.Query().Get("state"), oldSubjectURL.Query().Get("nonce"), "valid-code", "BindingSubject"), nil, withCookies(oldSubjectBegin.cookie))
	requireErrorReason(t, oldSubjectResp, http.StatusConflict, "provider_identity_rejected", "no_linked_user")
	requireNoSessionCookie(t, oldSubjectResp)

	newSubjectBegin := beginEnterpriseAuth(t, server, "corp-bind")
	newSubjectURL := parseEnterpriseRedirect(t, newSubjectBegin.redirectURL)
	newSubjectResp := doNoRedirect(t, http.MethodGet, oidcCallbackURL(server, "corp-bind", newSubjectURL.Query().Get("state"), newSubjectURL.Query().Get("nonce"), "valid-code", "BindingSubjectRotated"), nil, withCookies(newSubjectBegin.cookie))
	if newSubjectResp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected rotated subject callback success, got %d body=%#v", newSubjectResp.StatusCode, httptestx.ReadJSONBody(t, newSubjectResp))
	}
	newSubjectCookies := httptestx.RequireAuthCookies(t, newSubjectResp.Cookies())
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE id::text = $1 AND last_auth_at IS NULL`, authBindingID); got != 1 {
		t.Fatalf("old retired binding must not receive last_auth_at, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE id::text = $1 AND last_auth_at IS NOT NULL`, newAuthBindingID); got != 1 {
		t.Fatalf("replacement binding must receive last_auth_at after callback, got %d", got)
	}

	targetPasswordHashBeforeRetire := queryUserPasswordHash(t, db, targetID)
	targetMembershipsBeforeRetire := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE user_id::text = $1`, targetID)
	retireBaseVersion := queryUserVersion(t, db, targetID)
	retireResp := doJSON(t, http.MethodDelete, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings/"+newAuthBindingID, map[string]any{
		"base_user_version": retireBaseVersion,
		"client_txn_id":     "txn-retire-binding",
		"reason":            "retire provider",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	retireData := httptestx.RequireSuccessEnvelope(t, retireResp, http.StatusOK)["data"].(map[string]any)
	if got := len(retireData["auth_bindings"].([]any)); got != 1 {
		t.Fatalf("retire must remove enterprise binding from safe user, got %#v", retireData["auth_bindings"])
	}
	requireExactMapKeys(t, retireData["auth_bindings"].([]any)[0].(map[string]any), "created_at", "provider_key", "provider_type", "username")

	if got := queryCount(t, db, `SELECT COUNT(*) FROM enterprise_auth_bindings WHERE user_id::text = $1 AND retired_at IS NULL`, targetID); got != 0 {
		t.Fatalf("expected no active enterprise bindings after retire, got %d", got)
	}
	sessionAfterRetire := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(newSubjectCookies.Session))
	httptestx.RequireErrorEnvelope(t, sessionAfterRetire, http.StatusUnauthorized, "session_required")

	retireReplay := doJSON(t, http.MethodDelete, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings/"+newAuthBindingID, map[string]any{
		"base_user_version": retireBaseVersion,
		"client_txn_id":     "txn-retire-binding",
		"reason":            "retire provider",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, retireReplay, http.StatusOK)

	retireAlreadyInactive := doJSON(t, http.MethodDelete, server.HTTP.URL+"/api/v1/users/"+targetID+"/auth-bindings/"+newAuthBindingID, map[string]any{
		"base_user_version": queryUserVersion(t, db, targetID),
		"client_txn_id":     "txn-retire-binding-inactive",
		"reason":            "retire already inactive",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	requireErrorReason(t, retireAlreadyInactive, http.StatusConflict, "auth_binding_conflict", "binding_not_active")

	if got := queryUserPasswordHash(t, db, targetID); got != targetPasswordHashBeforeRetire {
		t.Fatalf("retire must not mutate local credential hash")
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE user_id::text = $1`, targetID); got != targetMembershipsBeforeRetire {
		t.Fatalf("retire must not mutate incident memberships, before=%d after=%d", targetMembershipsBeforeRetire, got)
	}
	postRetireLocalSession, postRetireLocalCSRF := loginLocalUser(t, server, "binding.target@example.test", "BindingTarget1!", nil)
	postRetireSession := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(postRetireLocalSession, postRetireLocalCSRF))
	postRetireSessionData := httptestx.RequireSuccessEnvelope(t, postRetireSession, http.StatusOK)["data"].(map[string]any)
	if postRetireSessionData["provider_type"] != "local" || postRetireSessionData["user_id"] != targetID {
		t.Fatalf("local login must remain available after retire: %#v", postRetireSessionData)
	}

	secretKeys := enterpriseForbiddenSecretKeys()
	createIDem := lookupRouteIdempotency(t, db, "users.auth_bindings.create", adminID, targetID, "txn-create-binding")
	if createIDem.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected create idempotency status: %#v", createIDem)
	}
	httptestx.RequireSecretSafePayload(t, createIDem.Response, secretKeys)
	rotateIDem := lookupRouteIdempotency(t, db, "users.auth_bindings.rotate", adminID, authBindingID, "txn-rotate-binding")
	if rotateIDem.StatusCode != http.StatusOK {
		t.Fatalf("unexpected rotate idempotency status: %#v", rotateIDem)
	}
	httptestx.RequireSecretSafePayload(t, rotateIDem.Response, secretKeys)
	retireIDem := lookupRouteIdempotency(t, db, "users.auth_bindings.retire", adminID, newAuthBindingID, "txn-retire-binding")
	if retireIDem.StatusCode != http.StatusOK {
		t.Fatalf("unexpected retire idempotency status: %#v", retireIDem)
	}
	httptestx.RequireSecretSafePayload(t, retireIDem.Response, secretKeys)

	events := lookupUserAuditEvents(t, db, targetID)
	createEvent := requireAuditEventBySource(t, events, "users.auth_bindings.create")
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: createEvent.ActorUserID,
		Source:      createEvent.EventSource,
		ClientTxnID: createEvent.ClientTxnID,
		RequestID:   createEvent.RequestID,
		CreatedAt:   createEvent.CreatedAt,
	}, adminID, "users.auth_bindings.create", "txn-create-binding")
	if createEvent.After["provider_key"] != "corp-bind" || createEvent.After["provider_subject"] != "BindingSubject" || createEvent.After["auth_binding_id"] != authBindingID {
		t.Fatalf("unexpected create audit payload: %#v", createEvent.After)
	}
	httptestx.RequireSecretSafePayload(t, createEvent.After, secretKeys)

	rotateEvent := requireAuditEventBySource(t, events, "users.auth_bindings.rotate")
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: rotateEvent.ActorUserID,
		Source:      rotateEvent.EventSource,
		ClientTxnID: rotateEvent.ClientTxnID,
		RequestID:   rotateEvent.RequestID,
		CreatedAt:   rotateEvent.CreatedAt,
	}, adminID, "users.auth_bindings.rotate", "txn-rotate-binding")
	if rotateEvent.Before["auth_binding_id"] != authBindingID || rotateEvent.Before["provider_subject"] != "BindingSubject" ||
		rotateEvent.After["auth_binding_id"] != newAuthBindingID || rotateEvent.After["provider_subject"] != "BindingSubjectRotated" {
		t.Fatalf("unexpected rotate audit payload: before=%#v after=%#v", rotateEvent.Before, rotateEvent.After)
	}
	httptestx.RequireSecretSafePayload(t, rotateEvent.Before, secretKeys)
	httptestx.RequireSecretSafePayload(t, rotateEvent.After, secretKeys)

	retireEvent := requireAuditEventBySource(t, events, "users.auth_bindings.retire")
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: retireEvent.ActorUserID,
		Source:      retireEvent.EventSource,
		ClientTxnID: retireEvent.ClientTxnID,
		RequestID:   retireEvent.RequestID,
		CreatedAt:   retireEvent.CreatedAt,
	}, adminID, "users.auth_bindings.retire", "txn-retire-binding")
	if retireEvent.Before["auth_binding_id"] != newAuthBindingID || retireEvent.Before["provider_subject"] != "BindingSubjectRotated" {
		t.Fatalf("unexpected retire audit before payload: %#v", retireEvent.Before)
	}
	httptestx.RequireSecretSafePayload(t, retireEvent.Before, secretKeys)
	httptestx.RequireSecretSafePayload(t, retireEvent.After, secretKeys)

	for _, source := range []string{"users.auth_bindings.create", "users.auth_bindings.rotate", "users.auth_bindings.retire"} {
		if got := queryCount(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE actor_user_id::text = $1 AND target_user_id::text = $2 AND event_source = $3`, adminID, targetID, source); got != 1 {
			t.Fatalf("expected one audit event for %s, got %d", source, got)
		}
	}
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
	return beginEnterpriseAuthWithBody(t, server, providerKey, map[string]any{})
}

func beginEnterpriseAuthWithBody(t testing.TB, server *httptestx.Server, providerKey string, body map[string]any) enterpriseBeginResult {
	t.Helper()
	resp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/providers/"+providerKey+"/begin", body)
	responseBody := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	return enterpriseBeginResult{
		cookie:      requireEnterpriseAuthCookie(t, resp.Cookies()),
		redirectURL: responseBody["data"].(map[string]any)["redirect_url"].(string),
	}
}

func parseEnterpriseRedirect(t testing.TB, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse enterprise redirect_url %q: %v", rawURL, err)
	}
	return parsed
}

func oidcCallbackURL(server *httptestx.Server, providerKey string, state string, nonce string, code string, subject string) string {
	query := url.Values{}
	if state != "" {
		query.Set("state", state)
	}
	if nonce != "" {
		query.Set("nonce", nonce)
	}
	if code != "" {
		query.Set("code", code)
	}
	if subject != "" {
		query.Set("subject", subject)
	}
	return server.HTTP.URL + "/api/v1/auth/oidc/" + url.PathEscape(providerKey) + "/callback?" + query.Encode()
}

func requireErrorReason(t testing.TB, resp *http.Response, status int, code string, reason string) {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, resp, status, code)
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != reason {
		t.Fatalf("unexpected reason_code: got %#v want %q in %#v", details["reason_code"], reason, details)
	}
}

func requireNoSessionCookie(t testing.TB, resp *http.Response) {
	t.Helper()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "cartulary_session" || cookie.Name == "cartulary_csrf" {
			t.Fatalf("failure response must not issue auth cookies: %#v", resp.Cookies())
		}
	}
}

func requireExactMapKeys(t testing.TB, got map[string]any, keys ...string) {
	t.Helper()
	if len(got) != len(keys) {
		t.Fatalf("unexpected key set: got %#v want %v", got, keys)
	}
	for _, key := range keys {
		if _, ok := got[key]; !ok {
			t.Fatalf("missing key %q in %#v", key, got)
		}
	}
}

func requireJSONDoesNotContain(t testing.TB, value any, forbidden ...string) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload for forbidden scan: %v", err)
	}
	for _, needle := range forbidden {
		if strings.Contains(string(payload), needle) {
			t.Fatalf("payload contains forbidden marker %q: %s", needle, payload)
		}
	}
}

func enterpriseForbiddenSecretKeys() []string {
	keys := append([]string{}, forbiddenSecretKeys()...)
	keys = append(keys,
		"client_secret",
		"private_key",
		"assertion",
		"SAMLResponse",
		"id_token",
		"access_token",
		"refresh_token",
		"code_verifier",
		"pkce_verifier",
	)
	return keys
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

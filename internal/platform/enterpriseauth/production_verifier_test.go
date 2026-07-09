package enterpriseauth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/crewjam/saml"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestProductionOIDCVerifierAuthCodePKCEInterop(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate OIDC signing key: %v", err)
	}
	const (
		clientID     = "cartulary-client"
		clientSecret = "oidc-client-secret"
		code         = "provider-auth-code"
		state        = "oidc-state"
		nonce        = "oidc-nonce"
		pkceVerifier = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdef"
		subject      = "oidc-subject-123"
	)

	var issuer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]any{
				"issuer":                 issuer,
				"authorization_endpoint": issuer + "/authorize",
				"token_endpoint":         issuer + "/token",
				"jwks_uri":               issuer + "/jwks",
				"id_token_signing_alg_values_supported": []string{
					"RS256",
				},
			})
		case "/jwks":
			writeJSON(t, w, map[string]any{"keys": []map[string]any{rsaJWK(key, "test-key")}})
		case "/token":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse token request form: %v", err)
			}
			if got := r.Form.Get("grant_type"); got != "authorization_code" {
				t.Fatalf("unexpected grant_type: %q", got)
			}
			if got := r.Form.Get("code"); got != code {
				t.Fatalf("unexpected code: %q", got)
			}
			if got := r.Form.Get("code_verifier"); got != pkceVerifier {
				t.Fatalf("unexpected code_verifier: %q", got)
			}
			if username, password, ok := r.BasicAuth(); ok {
				if username != clientID || password != clientSecret {
					t.Fatalf("unexpected basic auth: %q/%q", username, password)
				}
			} else if r.Form.Get("client_id") != clientID || r.Form.Get("client_secret") != clientSecret {
				t.Fatalf("missing client credentials in token request")
			}
			idToken := signOIDCIDToken(t, key, "test-key", map[string]any{
				"iss":   issuer,
				"sub":   subject,
				"aud":   clientID,
				"exp":   now.Add(time.Hour).Unix(),
				"iat":   now.Add(-time.Minute).Unix(),
				"nonce": nonce,
			})
			writeJSON(t, w, map[string]any{
				"access_token": "access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
				"id_token":     idToken,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	issuer = server.URL

	secretKind := "env"
	secretName := "corp-oidc-secret"
	provider := authn.EnterpriseAuthProviderRecord{
		ProviderKey:           "corp-oidc",
		ProviderType:          "oidc",
		AuthorizationEndpoint: strPtr(issuer + "/authorize"),
		Issuer:                &issuer,
		TokenEndpoint:         strPtr(issuer + "/token"),
		JWKSURI:               strPtr(issuer + "/jwks"),
		ClientID:              strPtr(clientID),
		ClientSecretRefKind:   &secretKind,
		ClientSecretRefName:   &secretName,
		AdditionalScopes:      []string{"email"},
	}
	transaction := authn.EnterpriseAuthTransactionRecord{
		ProviderKey:  provider.ProviderKey,
		ProviderType: provider.ProviderType,
		State:        strPtr(state),
		Nonce:        strPtr(nonce),
	}

	begin, err := BuildBeginRedirect(provider, "https://cartulary.example.test", transaction, pkceVerifier)
	if err != nil {
		t.Fatalf("build OIDC begin redirect: %v", err)
	}
	authURL, err := url.Parse(begin.URL)
	if err != nil {
		t.Fatalf("parse OIDC begin redirect: %v", err)
	}
	if got := authURL.Query().Get("code_challenge"); got != pkceChallenge(pkceVerifier) {
		t.Fatalf("unexpected code_challenge: got %q want %q", got, pkceChallenge(pkceVerifier))
	}
	if got := authURL.Query().Get("code_challenge_method"); got != "S256" {
		t.Fatalf("unexpected code_challenge_method: %q", got)
	}
	if got := authURL.Query().Get("nonce"); got != nonce {
		t.Fatalf("unexpected nonce: %q", got)
	}

	result, apiErr := (ProductionOIDCVerifier{}).VerifyCallback(context.Background(), OIDCCallbackVerificationRequest{
		Provider:     provider,
		Transaction:  transaction,
		Values:       url.Values{"code": []string{code}, "state": []string{state}},
		PKCEVerifier: pkceVerifier,
		PublicOrigin: "https://cartulary.example.test",
		Env:          map[string]string{"CARTULARY_SECRET_CORP_OIDC_SECRET": clientSecret},
		Now:          now,
	})
	if apiErr != nil {
		t.Fatalf("verify OIDC callback: %#v", apiErr)
	}
	if result.ProviderSubject != subject || result.State != state || result.Nonce != nonce {
		t.Fatalf("unexpected OIDC verification result: %#v", result)
	}
}

func TestProductionOIDCVerifierNegativeReasonCodes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mutate     func(*oidcVerifierFixture)
		wantCode   string
		wantReason string
	}{
		{
			name: "missing code",
			mutate: func(f *oidcVerifierFixture) {
				f.request.Values.Del("code")
			},
			wantCode:   "provider_response_rejected",
			wantReason: "missing_required_field",
		},
		{
			name: "state mismatch",
			mutate: func(f *oidcVerifierFixture) {
				f.request.Values.Set("state", "wrong-state")
			},
			wantCode:   "provider_response_rejected",
			wantReason: "state_mismatch",
		},
		{
			name: "nonce missing from transaction",
			mutate: func(f *oidcVerifierFixture) {
				f.request.Transaction.Nonce = nil
			},
			wantCode:   "provider_response_rejected",
			wantReason: "nonce_mismatch",
		},
		{
			name: "pkce verifier missing",
			mutate: func(f *oidcVerifierFixture) {
				f.request.PKCEVerifier = ""
			},
			wantCode:   "provider_response_rejected",
			wantReason: "code_exchange_failed",
		},
		{
			name: "client secret missing",
			mutate: func(f *oidcVerifierFixture) {
				f.request.Env = map[string]string{}
			},
			wantCode:   "provider_response_rejected",
			wantReason: "code_exchange_failed",
		},
		{
			name: "code exchange rejected",
			mutate: func(f *oidcVerifierFixture) {
				f.tokenStatus = http.StatusBadRequest
			},
			wantCode:   "provider_response_rejected",
			wantReason: "code_exchange_failed",
		},
		{
			name: "id token missing",
			mutate: func(f *oidcVerifierFixture) {
				f.includeIDToken = false
			},
			wantCode:   "provider_response_rejected",
			wantReason: "missing_required_field",
		},
		{
			name: "discovery issuer mismatch",
			mutate: func(f *oidcVerifierFixture) {
				f.discoveryIssuer = f.issuer + "/other"
			},
			wantCode:   "provider_response_rejected",
			wantReason: "issuer_mismatch",
		},
		{
			name: "id token issuer mismatch",
			mutate: func(f *oidcVerifierFixture) {
				f.tokenClaims["iss"] = "https://issuer.example.invalid"
			},
			wantCode:   "provider_response_rejected",
			wantReason: "issuer_mismatch",
		},
		{
			name: "audience mismatch",
			mutate: func(f *oidcVerifierFixture) {
				f.tokenClaims["aud"] = "other-client"
			},
			wantCode:   "provider_response_rejected",
			wantReason: "audience_mismatch",
		},
		{
			name: "expired id token",
			mutate: func(f *oidcVerifierFixture) {
				f.tokenClaims["exp"] = f.now.Add(-time.Hour).Unix()
			},
			wantCode:   "provider_response_rejected",
			wantReason: "assertion_expired",
		},
		{
			name: "signature invalid",
			mutate: func(f *oidcVerifierFixture) {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					t.Fatalf("generate alternate OIDC signing key: %v", err)
				}
				f.signingKey = key
			},
			wantCode:   "provider_response_rejected",
			wantReason: "signature_invalid",
		},
		{
			name: "subject missing",
			mutate: func(f *oidcVerifierFixture) {
				f.tokenClaims["sub"] = ""
			},
			wantCode:   "provider_identity_rejected",
			wantReason: "subject_missing",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newOIDCVerifierFixture(t)
			defer fixture.close()
			tc.mutate(fixture)

			_, apiErr := (ProductionOIDCVerifier{}).VerifyCallback(context.Background(), fixture.request)
			requireEnterpriseAuthAPIError(t, apiErr, tc.wantCode, tc.wantReason)
		})
	}
}

func TestProductionSAMLVerifierSPInitiatedInterop(t *testing.T) {
	key, cert := testSAMLKeyPair(t)
	publicOrigin := "https://cartulary.example.test"
	relayState := "saml-relay-state"
	subject := "saml-subject-123"
	idpMetadataURL := url.URL{Scheme: "https", Host: "idp.example.test", Path: "/metadata"}
	idpSSOURL := url.URL{Scheme: "https", Host: "idp.example.test", Path: "/sso"}
	idpEntityID := idpMetadataURL.String()
	spEntityID := SAMLMetadataURL(publicOrigin, "corp-saml")
	provider := authn.EnterpriseAuthProviderRecord{
		ProviderKey:               "corp-saml",
		ProviderType:              "saml",
		SAMLIDPEntityID:           &idpEntityID,
		SAMLSSOURL:                strPtr(idpSSOURL.String()),
		SAMLIDPSigningCertificate: []string{base64.StdEncoding.EncodeToString(cert.Raw)},
		SAMLSPHostEntityID:        &spEntityID,
		SAMLSubjectSource:         &authn.EnterpriseAuthSAMLSubjectSource{Kind: "name_id"},
	}
	transaction := authn.EnterpriseAuthTransactionRecord{
		ProviderKey:  provider.ProviderKey,
		ProviderType: provider.ProviderType,
		RelayState:   &relayState,
	}
	begin, err := BuildBeginRedirect(provider, publicOrigin, transaction, "")
	if err != nil {
		t.Fatalf("build SAML begin redirect: %v", err)
	}
	if begin.SAMLRequestID == nil || *begin.SAMLRequestID == "" {
		t.Fatalf("expected SAML request ID")
	}
	transaction.SAMLRequestID = begin.SAMLRequestID

	sp, err := SAMLServiceProvider(provider, publicOrigin)
	if err != nil {
		t.Fatalf("build SAML SP: %v", err)
	}
	idp := saml.IdentityProvider{
		Key:                     key,
		Certificate:             cert,
		MetadataURL:             idpMetadataURL,
		SSOURL:                  idpSSOURL,
		ServiceProviderProvider: staticSAMLServiceProvider{metadata: sp.Metadata()},
		SessionProvider: staticSAMLSessionProvider{session: &saml.Session{
			ID:           "session-1",
			CreateTime:   time.Now().UTC().Add(-time.Minute),
			ExpireTime:   time.Now().UTC().Add(time.Hour),
			NameID:       subject,
			NameIDFormat: "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent",
		}},
	}

	redirectURL, err := url.Parse(begin.URL)
	if err != nil {
		t.Fatalf("parse SAML redirect: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, redirectURL.String(), nil)
	response := httptest.NewRecorder()
	idp.ServeSSO(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("SAML IdP SSO status: got %d body %s", response.Code, response.Body.String())
	}
	values := url.Values{
		"SAMLResponse": []string{hiddenInputValue(t, response.Body.String(), "SAMLResponse")},
		"RelayState":   []string{hiddenInputValue(t, response.Body.String(), "RelayState")},
	}

	result, apiErr := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), SAMLACSVerificationRequest{
		Provider:     provider,
		Transaction:  transaction,
		Values:       values,
		PublicOrigin: publicOrigin,
		Now:          time.Now().UTC(),
	})
	if apiErr != nil {
		t.Fatalf("verify SAML ACS: %#v", apiErr)
	}
	if result.ProviderSubject != subject {
		t.Fatalf("unexpected SAML subject: got %q want %q", result.ProviderSubject, subject)
	}
}

func TestProductionSAMLVerifierNegativeReasonCodes(t *testing.T) {
	provider, transaction := samlVerifierFixture(t)
	publicOrigin := "https://cartulary.example.test"

	for _, tc := range []struct {
		name       string
		request    SAMLACSVerificationRequest
		wantReason string
	}{
		{
			name: "missing SAMLResponse",
			request: SAMLACSVerificationRequest{
				Provider:     provider,
				Transaction:  transaction,
				Values:       url.Values{"RelayState": []string{*transaction.RelayState}},
				PublicOrigin: publicOrigin,
			},
			wantReason: "missing_required_field",
		},
		{
			name: "relay state mismatch",
			request: SAMLACSVerificationRequest{
				Provider:     provider,
				Transaction:  transaction,
				Values:       url.Values{"SAMLResponse": []string{"invalid"}, "RelayState": []string{"wrong-relay-state"}},
				PublicOrigin: publicOrigin,
			},
			wantReason: "relay_state_mismatch",
		},
		{
			name: "invalid assertion payload",
			request: SAMLACSVerificationRequest{
				Provider:     provider,
				Transaction:  transaction,
				Values:       url.Values{"SAMLResponse": []string{"invalid"}, "RelayState": []string{*transaction.RelayState}},
				PublicOrigin: publicOrigin,
			},
			wantReason: "signature_invalid",
		},
		{
			name: "provider config unavailable at verification",
			request: SAMLACSVerificationRequest{
				Provider:     authn.EnterpriseAuthProviderRecord{ProviderKey: "corp-saml", ProviderType: "saml"},
				Transaction:  transaction,
				Values:       url.Values{"SAMLResponse": []string{"invalid"}, "RelayState": []string{*transaction.RelayState}},
				PublicOrigin: publicOrigin,
			},
			wantReason: "signature_invalid",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), tc.request)
			requireEnterpriseAuthAPIError(t, apiErr, "provider_response_rejected", tc.wantReason)
		})
	}
}

func TestProductionSAMLVerifierClassifiesPrivateErrors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		err        error
		wantReason string
	}{
		{
			name:       "request correlation mismatch",
			err:        &saml.InvalidResponseError{PrivateErr: errors.New("`InResponseTo` does not match any of the possible request IDs")},
			wantReason: "relay_state_mismatch",
		},
		{
			name:       "issuer mismatch",
			err:        &saml.InvalidResponseError{PrivateErr: errors.New("response Issuer does not match the IDP metadata")},
			wantReason: "issuer_mismatch",
		},
		{
			name:       "audience mismatch",
			err:        &saml.InvalidResponseError{PrivateErr: errors.New("assertion Conditions AudienceRestriction does not contain service provider")},
			wantReason: "audience_mismatch",
		},
		{
			name:       "assertion expired",
			err:        &saml.InvalidResponseError{PrivateErr: errors.New("assertion Conditions is expired")},
			wantReason: "assertion_expired",
		},
		{
			name:       "signature invalid fallback",
			err:        &saml.InvalidResponseError{PrivateErr: errors.New("cannot validate signature on Response")},
			wantReason: "signature_invalid",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifySAMLVerificationError(tc.err); got != tc.wantReason {
				t.Fatalf("unexpected reason: got %q want %q", got, tc.wantReason)
			}
		})
	}
}

func TestProductionVerifierReasonRegistriesRemainClosed(t *testing.T) {
	for _, reasonCode := range []string{
		"missing_required_field",
		"state_mismatch",
		"relay_state_mismatch",
		"nonce_mismatch",
		"code_exchange_failed",
		"issuer_mismatch",
		"audience_mismatch",
		"signature_invalid",
		"assertion_expired",
	} {
		if got := normalizeProviderResponseReason(reasonCode); got != reasonCode {
			t.Fatalf("provider response reason %q normalized to %q", reasonCode, got)
		}
	}
	if got := normalizeProviderResponseReason("provider_config_invalid"); got != "signature_invalid" {
		t.Fatalf("unsupported provider response reason escaped: got %q", got)
	}

	for _, reasonCode := range []string{"subject_missing", "no_linked_user", "ambiguous_link", "inactive_user"} {
		if got := normalizeProviderIdentityReason(reasonCode); got != reasonCode {
			t.Fatalf("provider identity reason %q normalized to %q", reasonCode, got)
		}
	}
	if got := normalizeProviderIdentityReason("unsupported"); got != "no_linked_user" {
		t.Fatalf("unsupported provider identity reason escaped: got %q", got)
	}
}

type staticSAMLServiceProvider struct {
	metadata *saml.EntityDescriptor
}

func (p staticSAMLServiceProvider) GetServiceProvider(*http.Request, string) (*saml.EntityDescriptor, error) {
	return p.metadata, nil
}

type staticSAMLSessionProvider struct {
	session *saml.Session
}

func (p staticSAMLSessionProvider) GetSession(http.ResponseWriter, *http.Request, *saml.IdpAuthnRequest) *saml.Session {
	return p.session
}

func writeJSON(t testing.TB, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write JSON: %v", err)
	}
}

type oidcVerifierFixture struct {
	now             time.Time
	issuer          string
	clientID        string
	clientSecret    string
	code            string
	state           string
	nonce           string
	pkceVerifier    string
	subject         string
	discoveryIssuer string
	tokenStatus     int
	includeIDToken  bool
	tokenClaims     map[string]any
	key             *rsa.PrivateKey
	signingKey      *rsa.PrivateKey
	request         OIDCCallbackVerificationRequest
	server          *httptest.Server
}

func newOIDCVerifierFixture(t testing.TB) *oidcVerifierFixture {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate OIDC signing key: %v", err)
	}
	fixture := &oidcVerifierFixture{
		now:            time.Now().UTC().Truncate(time.Second),
		clientID:       "cartulary-client",
		clientSecret:   "oidc-client-secret",
		code:           "provider-auth-code",
		state:          "oidc-state",
		nonce:          "oidc-nonce",
		pkceVerifier:   "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdef",
		subject:        "oidc-subject-123",
		tokenStatus:    http.StatusOK,
		includeIDToken: true,
		key:            key,
		signingKey:     key,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]any{
				"issuer":                 fixture.discoveryIssuer,
				"authorization_endpoint": fixture.issuer + "/authorize",
				"token_endpoint":         fixture.issuer + "/token",
				"jwks_uri":               fixture.issuer + "/jwks",
				"id_token_signing_alg_values_supported": []string{
					"RS256",
				},
			})
		case "/jwks":
			writeJSON(t, w, map[string]any{"keys": []map[string]any{rsaJWK(fixture.key, "test-key")}})
		case "/token":
			if fixture.tokenStatus != http.StatusOK {
				w.WriteHeader(fixture.tokenStatus)
				_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
				return
			}
			payload := map[string]any{
				"access_token": "access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			}
			if fixture.includeIDToken {
				payload["id_token"] = signOIDCIDToken(t, fixture.signingKey, "test-key", fixture.tokenClaims)
			}
			writeJSON(t, w, payload)
		default:
			http.NotFound(w, r)
		}
	}))
	fixture.server = server
	fixture.issuer = server.URL
	fixture.discoveryIssuer = fixture.issuer
	fixture.tokenClaims = map[string]any{
		"iss":   fixture.issuer,
		"sub":   fixture.subject,
		"aud":   fixture.clientID,
		"exp":   fixture.now.Add(time.Hour).Unix(),
		"iat":   fixture.now.Add(-time.Minute).Unix(),
		"nonce": fixture.nonce,
	}
	secretKind := "env"
	secretName := "corp-oidc-secret"
	provider := authn.EnterpriseAuthProviderRecord{
		ProviderKey:           "corp-oidc",
		ProviderType:          "oidc",
		AuthorizationEndpoint: strPtr(fixture.issuer + "/authorize"),
		Issuer:                &fixture.issuer,
		TokenEndpoint:         strPtr(fixture.issuer + "/token"),
		JWKSURI:               strPtr(fixture.issuer + "/jwks"),
		ClientID:              strPtr(fixture.clientID),
		ClientSecretRefKind:   &secretKind,
		ClientSecretRefName:   &secretName,
	}
	fixture.request = OIDCCallbackVerificationRequest{
		Provider: provider,
		Transaction: authn.EnterpriseAuthTransactionRecord{
			ProviderKey:  provider.ProviderKey,
			ProviderType: provider.ProviderType,
			State:        strPtr(fixture.state),
			Nonce:        strPtr(fixture.nonce),
		},
		Values:       url.Values{"code": []string{fixture.code}, "state": []string{fixture.state}},
		PKCEVerifier: fixture.pkceVerifier,
		PublicOrigin: "https://cartulary.example.test",
		Env:          map[string]string{"CARTULARY_SECRET_CORP_OIDC_SECRET": fixture.clientSecret},
		Now:          fixture.now,
	}
	return fixture
}

func (f *oidcVerifierFixture) close() {
	if f.server != nil {
		f.server.Close()
	}
}

func samlVerifierFixture(t testing.TB) (authn.EnterpriseAuthProviderRecord, authn.EnterpriseAuthTransactionRecord) {
	t.Helper()
	_, cert := testSAMLKeyPair(t)
	publicOrigin := "https://cartulary.example.test"
	relayState := "saml-relay-state"
	idpMetadataURL := url.URL{Scheme: "https", Host: "idp.example.test", Path: "/metadata"}
	idpSSOURL := url.URL{Scheme: "https", Host: "idp.example.test", Path: "/sso"}
	idpEntityID := idpMetadataURL.String()
	spEntityID := SAMLMetadataURL(publicOrigin, "corp-saml")
	provider := authn.EnterpriseAuthProviderRecord{
		ProviderKey:               "corp-saml",
		ProviderType:              "saml",
		SAMLIDPEntityID:           &idpEntityID,
		SAMLSSOURL:                strPtr(idpSSOURL.String()),
		SAMLIDPSigningCertificate: []string{base64.StdEncoding.EncodeToString(cert.Raw)},
		SAMLSPHostEntityID:        &spEntityID,
		SAMLSubjectSource:         &authn.EnterpriseAuthSAMLSubjectSource{Kind: "name_id"},
	}
	return provider, authn.EnterpriseAuthTransactionRecord{
		ProviderKey:   provider.ProviderKey,
		ProviderType:  provider.ProviderType,
		RelayState:    &relayState,
		SAMLRequestID: strPtr("saml-request-id"),
	}
}

func requireEnterpriseAuthAPIError(t testing.TB, err *httpapi.APIError, wantCode string, wantReason string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected API error %s/%s", wantCode, wantReason)
		return
	}
	if err.Code != wantCode {
		t.Fatalf("unexpected error code: got %q want %q in %#v", err.Code, wantCode, err)
	}
	if got := err.Details["reason_code"]; got != wantReason {
		t.Fatalf("unexpected reason_code: got %v want %q in %#v", got, wantReason, err)
	}
}

func rsaJWK(key *rsa.PrivateKey, keyID string) map[string]any {
	exponent := big.NewInt(int64(key.PublicKey.E)).Bytes()
	return map[string]any{
		"kty": "RSA",
		"use": "sig",
		"kid": keyID,
		"alg": "RS256",
		"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(exponent),
	}
}

func signOIDCIDToken(t testing.TB, key *rsa.PrivateKey, keyID string, claims map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]any{"alg": "RS256", "kid": keyID, "typ": "JWT"})
	if err != nil {
		t.Fatalf("marshal JWT header: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal JWT payload: %v", err)
	}
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func pkceChallenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func testSAMLKeyPair(t testing.TB) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate SAML key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create SAML certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse SAML certificate: %v", err)
	}
	return key, cert
}

func hiddenInputValue(t testing.TB, body string, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`<input type="hidden" name="` + regexp.QuoteMeta(name) + `" value="([^"]*)" />`)
	matches := pattern.FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("missing hidden input %q in %s", name, body)
	}
	return html.UnescapeString(matches[1])
}

func strPtr(value string) *string {
	return &value
}

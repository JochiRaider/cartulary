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

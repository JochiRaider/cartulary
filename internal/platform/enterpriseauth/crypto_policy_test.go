package enterpriseauth

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"

	jose "github.com/go-jose/go-jose/v4"

	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

func TestEnterpriseOIDCCryptographicPolicy(t *testing.T) {
	for _, algorithm := range []jose.SignatureAlgorithm{jose.RS256, jose.PS256, jose.ES256} {
		t.Run(string(algorithm), func(t *testing.T) {
			fixture := newOIDCVerifierFixture(t)
			defer fixture.close()
			var key any = fixture.key
			var public any = &fixture.key.PublicKey
			if algorithm == jose.ES256 {
				ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				if err != nil {
					t.Fatal(err)
				}
				key = ec
				public = &ec.PublicKey
			}
			setOIDCSigningFixture(t, fixture, algorithm, key, public)
			verify := ProductionOIDCVerifier{rootCertificatePath: fixture.rootPath}
			result, err := verify.VerifyCallback(context.Background(), fixture.request)
			if err != nil || result.ProviderSubject != fixture.subject {
				t.Fatalf("allowed algorithm failed: %#v", err)
			}
			// A subsequent callback fetches current provider keys, including replacement.
			replacement, e := rsa.GenerateKey(rand.Reader, 2048)
			if e != nil {
				t.Fatal(e)
			}
			setOIDCSigningFixture(t, fixture, jose.PS256, replacement, &replacement.PublicKey)
			if _, err := verify.VerifyCallback(context.Background(), fixture.request); err != nil {
				t.Fatalf("key rotation: %#v", err)
			}
		})
	}
	for _, name := range []string{"small RSA", "odd bit length", "even modulus", "small exponent", "wrong curve", "wrong algorithm", "symmetric key", "private key", "encryption use", "certificate JWK", "embedded JWK", "unsupported token algorithm", "discovery endpoint substitution", "plaintext endpoint", "untrusted TLS", "redirect"} {
		t.Run(name, func(t *testing.T) {
			fixture := newOIDCVerifierFixture(t)
			defer fixture.close()
			key := rsaJWK(fixture.key, "test-key")
			verify := ProductionOIDCVerifier{rootCertificatePath: fixture.rootPath}
			wantReason := "signature_invalid"
			switch name {
			case "small RSA":
				key["n"] = base64.RawURLEncoding.EncodeToString(new(big.Int).Lsh(big.NewInt(1), 1023).Bytes())
			case "odd bit length":
				modulus := new(big.Int).Lsh(big.NewInt(1), 2048)
				modulus.SetBit(modulus, 0, 1)
				key["n"] = base64.RawURLEncoding.EncodeToString(modulus.Bytes())
			case "even modulus":
				modulus := new(big.Int).Set(fixture.key.N)
				modulus.SetBit(modulus, 0, 0)
				key["n"] = base64.RawURLEncoding.EncodeToString(modulus.Bytes())
			case "small exponent":
				key["e"] = "Aw"
			case "wrong curve":
				ec, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
				if err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(jose.JSONWebKey{Key: &ec.PublicKey, Algorithm: "ES256", KeyID: "test-key", Use: "sig"})
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(raw, &key); err != nil {
					t.Fatal(err)
				}
			case "wrong algorithm":
				key["alg"] = "RS512"
			case "symmetric key":
				key = map[string]any{"kty": "oct", "k": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "alg": "HS256"}
			case "private key":
				key["d"] = "AA"
			case "encryption use":
				key["use"] = "enc"
			case "certificate JWK":
				_, certificate := testSAMLKeyPair(t)
				key["x5c"] = []string{base64.StdEncoding.EncodeToString(certificate.Raw)}
			case "embedded JWK":
				_, certificate := testSAMLKeyPair(t)
				key["x5c"] = []string{base64.StdEncoding.EncodeToString(certificate.Raw)}
				header, _ := json.Marshal(map[string]any{"alg": "RS256", "jwk": key})
				fixture.idTokenOverride = base64.RawURLEncoding.EncodeToString(header) + ".e30.AA"
			case "unsupported token algorithm":
				fixture.idTokenOverride = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`)) + ".e30.AA"
			case "discovery endpoint substitution":
				fixture.discoveryJWKSURI = "https://unconfigured.example.test/jwks"
			case "plaintext endpoint":
				fixture.request.Provider.TokenEndpoint = strPtr("http://127.0.0.1/token")
			case "untrusted TLS":
				verify.rootCertificatePath = ""
				wantReason = "code_exchange_failed"
			case "redirect":
				fixture.tokenRedirect = fixture.issuer + "/redirected"
				wantReason = "code_exchange_failed"
			}
			var err error
			fixture.jwksOverride, err = json.Marshal(map[string]any{"keys": []any{key}})
			if err != nil {
				t.Fatal(err)
			}
			_, apiErr := verify.VerifyCallback(context.Background(), fixture.request)
			requireEnterpriseAuthAPIError(t, apiErr, "provider_response_rejected", wantReason)
		})
	}
}

func setOIDCSigningFixture(t *testing.T, f *oidcVerifierFixture, algorithm jose.SignatureAlgorithm, key, public any) {
	t.Helper()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: algorithm, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "test-key"))
	if err != nil {
		t.Fatal(err)
	}
	claims, err := json.Marshal(f.tokenClaims)
	if err != nil {
		t.Fatal(err)
	}
	object, err := signer.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	f.idTokenOverride, err = object.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	f.jwksOverride, err = json.Marshal(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: public, KeyID: "test-key", Use: "sig", Algorithm: string(algorithm)}}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnterpriseSAMLAlgorithmAdmission(t *testing.T) {
	request := signedSAMLFixture(t)
	raw, err := base64.StdEncoding.DecodeString(request.Values.Get("SAMLResponse"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		transform func(string) string
	}{
		{"SHA1 signature", func(s string) string {
			return strings.ReplaceAll(s, "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256", "http://www.w3.org/2000/09/xmldsig#rsa-sha1")
		}},
		{"SHA1 digest", func(s string) string {
			return strings.ReplaceAll(s, "http://www.w3.org/2001/04/xmlenc#sha256", "http://www.w3.org/2000/09/xmldsig#sha1")
		}},
		{"encrypted assertion", func(string) string {
			return `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"><saml:EncryptedAssertion/></samlp:Response>`
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := tc.transform(string(raw))
			if changed == string(raw) {
				t.Fatal("fixture mutation had no effect")
			}
			request.Values.Set("SAMLResponse", base64.StdEncoding.EncodeToString([]byte(changed)))
			_, apiErr := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), request)
			requireEnterpriseAuthAPIError(t, apiErr, "provider_response_rejected", "signature_invalid")
			if _, err := admitSAMLResponse(request.Values.Get("SAMLResponse")); err == nil {
				t.Fatal("excluded algorithm reached protocol verifier")
			}
		})
	}
	for _, name := range []string{"weak RSA", "EC", "expired", "future", "encryption only"} {
		t.Run(name, func(t *testing.T) {
			provider, _ := samlVerifierFixture(t)
			_, certificate := testSAMLKeyPair(t)
			public := certificate.PublicKey
			switch name {
			case "weak RSA":
				public = &rsa.PublicKey{N: new(big.Int).Lsh(big.NewInt(1), 1023), E: 65537}
			case "EC":
				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				if err != nil {
					t.Fatal(err)
				}
				public = &key.PublicKey
			case "expired":
				certificate.NotAfter = certificate.NotBefore
			case "future":
				certificate.NotBefore = certificate.NotAfter
			case "encryption only":
				certificate.KeyUsage = x509.KeyUsageKeyEncipherment
			}
			authority, err := tlstest.NewAuthority()
			if err != nil {
				t.Fatal(err)
			}
			certificate.SignatureAlgorithm = x509.UnknownSignatureAlgorithm
			der, err := authority.Sign(certificate, public)
			if err != nil {
				t.Fatal(err)
			}
			provider.SAMLIDPSigningCertificate = []string{base64.StdEncoding.EncodeToString(der)}
			if _, err := SAMLServiceProvider(provider, request.PublicOrigin); err == nil {
				t.Fatal("unsupported signing certificate admitted")
			}
		})
	}
}

func TestEnterpriseProviderInputBounds(t *testing.T) {
	if admitOIDCToken(strings.Repeat("a", providerResponseLimit+1)) == nil {
		t.Fatal("oversized JWT admitted")
	}
	if admitOIDCKeySet([]byte(`{"keys":[],"keys":[]}`)) == nil {
		t.Fatal("duplicate JWK set field admitted")
	}
	if _, err := admitSAMLResponse(base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", providerResponseLimit+1)))); err == nil {
		t.Fatal("oversized SAML admitted")
	}
	server, root := newProviderHTTPSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", providerResponseLimit+1)))
	}))
	client, closeClient, err := providerHTTPClient(server.URL+"/jwks", root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeClient()
	if response, err := client.Get(server.URL + "/jwks"); err == nil {
		response.Body.Close()
		t.Fatal("oversized provider response admitted")
	}
}

func TestEnterpriseSAMLClaimAndTrustAdmission(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		change       func(*SAMLACSVerificationRequest)
	}{
		{"issuer", "issuer_mismatch", func(r *SAMLACSVerificationRequest) {
			r.Provider.SAMLIDPEntityID = strPtr("https://different.example.test/issuer")
		}},
		{"audience", "audience_mismatch", func(r *SAMLACSVerificationRequest) {
			r.Provider.SAMLSPHostEntityID = strPtr("https://different.example.test/sp")
		}},
		{"request", "relay_state_mismatch", func(r *SAMLACSVerificationRequest) { r.Transaction.SAMLRequestID = strPtr("unrelated-request") }},
		{"trust", "signature_invalid", func(r *SAMLACSVerificationRequest) {
			_, cert := testSAMLKeyPair(t)
			r.Provider.SAMLIDPSigningCertificate = []string{base64.StdEncoding.EncodeToString(cert.Raw)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := signedSAMLFixture(t)
			tc.change(&r)
			_, err := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), r)
			requireEnterpriseAuthAPIError(t, err, "provider_response_rejected", tc.reason)
		})
	}
}

func TestEnterpriseOIDCCertificateReplacement(t *testing.T) {
	fixture := newOIDCVerifierFixture(t)
	defer fixture.close()
	verify := ProductionOIDCVerifier{rootCertificatePath: fixture.rootPath}
	if _, err := verify.VerifyCallback(context.Background(), fixture.request); err != nil {
		t.Fatal(err)
	}
	address := fixture.server.Listener.Addr().String()
	handler := fixture.server.Config.Handler
	fixture.server.Close()
	replacement, root := newProviderHTTPSServer(t, handler, address)
	fixture.server = replacement
	if replacement.URL != fixture.issuer {
		t.Fatal("certificate replacement changed provider origin")
	}
	_, apiErr := verify.VerifyCallback(context.Background(), fixture.request)
	requireEnterpriseAuthAPIError(t, apiErr, "provider_response_rejected", "code_exchange_failed")
	verify.rootCertificatePath = root
	if _, err := verify.VerifyCallback(context.Background(), fixture.request); err != nil {
		t.Fatalf("restarted trust binding: %#v", err)
	}
	// A name mismatch is checked by TLS, with a trusted root and reachable peer.
	parsed, _ := url.Parse(fixture.issuer)
	parsed.Host = "localhost:" + parsed.Port()
	fixture.request.Provider.TokenEndpoint = strPtr(parsed.String() + "/token")
	_, apiErr = verify.VerifyCallback(context.Background(), fixture.request)
	requireEnterpriseAuthAPIError(t, apiErr, "provider_response_rejected", "code_exchange_failed")
}

func TestEnterpriseOIDCRejectsNonJOSEPSSSalt(t *testing.T) {
	fixture := newOIDCVerifierFixture(t)
	defer fixture.close()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"PS256","kid":"test-key"}`))
	claims, err := json.Marshal(fixture.tokenClaims)
	if err != nil {
		t.Fatal(err)
	}
	payload := header + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(payload))
	// A 16-byte salt is an approved RSA service but violates PS256's exact
	// hash-length salt. The library's default auto-detection used to accept it.
	signature, err := rsa.SignPSS(rand.Reader, fixture.key, crypto.SHA256, digest[:], &rsa.PSSOptions{SaltLength: 16})
	if err != nil {
		t.Fatal(err)
	}
	fixture.idTokenOverride = payload + "." + base64.RawURLEncoding.EncodeToString(signature)
	jwk := rsaJWK(fixture.key, "test-key")
	jwk["alg"] = "PS256"
	fixture.jwksOverride, err = json.Marshal(map[string]any{"keys": []any{jwk}})
	if err != nil {
		t.Fatal(err)
	}
	_, apiErr := (ProductionOIDCVerifier{rootCertificatePath: fixture.rootPath}).VerifyCallback(context.Background(), fixture.request)
	requireEnterpriseAuthAPIError(t, apiErr, "provider_response_rejected", "signature_invalid")
}

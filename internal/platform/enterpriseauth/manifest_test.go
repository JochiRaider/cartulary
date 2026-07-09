package enterpriseauth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

func TestProviderManifestValidMixedProviders(t *testing.T) {
	dir := t.TempDir()
	certPath := writeTestCertificate(t, dir)
	manifestPath := filepath.Join(dir, "enterprise-auth-providers.json")
	writeFile(t, manifestPath, `{
  "provider_manifest_schema_id": "cartulary.enterprise_auth_providers.v1",
  "providers": [
    {
      "provider_key": "corp-saml",
      "provider_type": "saml",
      "display_name": "Corporate SAML",
      "idp_entity_id": "https://idp.example.test/entity",
      "sso_url": "https://idp.example.test/sso",
      "idp_signing_certificate_paths": [`+quoteJSON(certPath)+`],
      "sp_entity_id": "https://cartulary.example.test/saml/sp",
      "subject_source": { "kind": "attribute", "attribute_name": "uid" }
    },
    {
      "provider_key": "corp-oidc",
      "provider_type": "oidc",
      "display_name": "Corporate OIDC",
      "issuer": "https://oidc.example.test",
      "authorization_endpoint": "https://oidc.example.test/auth",
      "token_endpoint": "https://oidc.example.test/token",
      "jwks_uri": "https://oidc.example.test/jwks",
      "client_id": "cartulary-client",
      "client_secret_ref": { "kind": "env", "name": "corp-oidc-secret" },
      "additional_scopes": ["profile"]
    }
  ]
}`)

	definitions, err := LoadProviderManifest(providerManifestConfig(manifestPath), map[string]string{
		"CARTULARY_SECRET_CORP_OIDC_SECRET": "client-secret",
	})
	if err != nil {
		t.Fatalf("load valid provider manifest: %v", err)
	}
	if len(definitions) != 2 {
		t.Fatalf("expected 2 provider definitions, got %#v", definitions)
	}
	if definitions[0].ProviderKey != "corp-oidc" || definitions[0].ClientSecretRefName == nil || *definitions[0].ClientSecretRefName != "corp-oidc-secret" {
		t.Fatalf("unexpected OIDC provider definition: %#v", definitions[0])
	}
	if definitions[1].ProviderKey != "corp-saml" || definitions[1].SAMLSubjectSource == nil || definitions[1].SAMLSubjectSource.AttributeName != "uid" {
		t.Fatalf("unexpected SAML provider definition: %#v", definitions[1])
	}
}

func TestProviderManifestRejectsDuplicateMembers(t *testing.T) {
	manifestPath := writeManifest(t, `{
  "provider_manifest_schema_id": "cartulary.enterprise_auth_providers.v1",
  "provider_manifest_schema_id": "cartulary.enterprise_auth_providers.v1",
  "providers": []
}`)
	_, err := LoadProviderManifest(providerManifestConfig(manifestPath), nil)
	requireProviderDiagnostic(t, err, providerManifestRootPath, "provider_manifest_schema_invalid")
}

func TestProviderManifestRejectsMissingSecret(t *testing.T) {
	manifestPath := writeManifest(t, `{
  "provider_manifest_schema_id": "cartulary.enterprise_auth_providers.v1",
  "providers": [{
    "provider_key": "corp-oidc",
    "provider_type": "oidc",
    "display_name": "Corporate OIDC",
    "issuer": "https://oidc.example.test",
    "authorization_endpoint": "https://oidc.example.test/auth",
    "token_endpoint": "https://oidc.example.test/token",
    "jwks_uri": "https://oidc.example.test/jwks",
    "client_id": "cartulary-client",
    "client_secret_ref": { "kind": "env", "name": "corp-oidc-secret" }
  }]
}`)
	_, err := LoadProviderManifest(providerManifestConfig(manifestPath), nil)
	requireProviderDiagnostic(t, err, providerPath(0, "client_secret_ref"), "provider_manifest_secret_missing")
}

func TestProviderManifestRejectsInvalidSigningCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "bad-cert.pem")
	writeFile(t, certPath, "not a certificate")
	manifestPath := filepath.Join(dir, "enterprise-auth-providers.json")
	writeFile(t, manifestPath, `{
  "provider_manifest_schema_id": "cartulary.enterprise_auth_providers.v1",
  "providers": [{
    "provider_key": "corp-saml",
    "provider_type": "saml",
    "display_name": "Corporate SAML",
    "idp_entity_id": "https://idp.example.test/entity",
    "sso_url": "https://idp.example.test/sso",
    "idp_signing_certificate_paths": [`+quoteJSON(certPath)+`],
    "sp_entity_id": "https://cartulary.example.test/saml/sp",
    "subject_source": { "kind": "name_id" }
  }]
}`)
	_, err := LoadProviderManifest(providerManifestConfig(manifestPath), nil)
	requireProviderDiagnostic(t, err, providerPath(0, "idp_signing_certificate_paths")+"[0]", "provider_manifest_referenced_file_invalid")
}

func providerManifestConfig(path string) config.Config {
	return config.Config{
		EnterpriseAuthentication: config.EnterpriseAuthenticationConfig{
			Claimed:              true,
			ProviderManifestPath: path,
		},
	}
}

func writeManifest(t testing.TB, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "enterprise-auth-providers.json")
	writeFile(t, path, content)
	return path
}

func writeFile(t testing.TB, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeTestCertificate(t testing.TB, dir string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create test certificate: %v", err)
	}
	path := filepath.Join(dir, "idp-signing.pem")
	data := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test certificate: %v", err)
	}
	return path
}

func quoteJSON(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func requireProviderDiagnostic(t testing.TB, err error, wantPath string, wantReason string) {
	t.Helper()
	diagnostics, ok := config.DiagnosticsFromError(err)
	if !ok {
		t.Fatalf("expected diagnostics error, got %v", err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Path == wantPath && diagnostic.ReasonCode == wantReason {
			return
		}
	}
	t.Fatalf("missing diagnostic path=%q reason=%q in %#v", wantPath, wantReason, diagnostics)
}

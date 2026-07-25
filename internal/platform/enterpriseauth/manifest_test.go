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
	}, testDocumentReader())
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
	_, err := LoadProviderManifest(providerManifestConfig(manifestPath), nil, testDocumentReader())
	requireProviderDiagnostic(t, err, providerManifestRootPath, "provider_manifest_schema_invalid")
}

func TestProviderManifestRejectsMalformedJSON(t *testing.T) {
	manifestPath := writeManifest(t, `{"provider_manifest_schema_id":`)
	_, err := LoadProviderManifest(providerManifestConfig(manifestPath), nil, testDocumentReader())
	requireProviderDiagnostic(t, err, providerManifestPathKey, "provider_manifest_parse_error")
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
	_, err := LoadProviderManifest(providerManifestConfig(manifestPath), nil, testDocumentReader())
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
	_, err := LoadProviderManifest(providerManifestConfig(manifestPath), nil, testDocumentReader())
	requireProviderDiagnostic(t, err, providerPath(0, "idp_signing_certificate_paths")+"[0]", "provider_manifest_referenced_file_invalid")
}

func TestProviderManifestDocumentPort_Unit(t *testing.T) {
	t.Run("unclaimed configuration performs no document read", func(t *testing.T) {
		var reads int
		definitions, err := LoadProviderManifest(Configuration{}, nil, DocumentReadFunc(func(string, int64) ([]byte, DocumentReadFailure) {
			reads++
			return nil, DocumentUnavailable
		}))
		if err != nil || definitions != nil || reads != 0 {
			t.Fatalf("unclaimed load = definitions %#v, reads %d, error %v", definitions, reads, err)
		}
	})

	for name, testCase := range map[string]struct {
		failure DocumentReadFailure
		reason  string
	}{
		"unavailable": {failure: DocumentUnavailable, reason: "provider_manifest_not_readable"},
		"unsafe":      {failure: DocumentUnsafe, reason: "provider_manifest_not_regular_file"},
		"too large":   {failure: DocumentTooLarge, reason: "provider_manifest_schema_invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := LoadProviderManifest(providerManifestConfig("/operator/private/providers.json"), nil, DocumentReadFunc(func(path string, maximumBytes int64) ([]byte, DocumentReadFailure) {
				if path != "/operator/private/providers.json" || maximumBytes != ProviderManifestMaximumSize {
					t.Fatalf("read request = %q / %d", path, maximumBytes)
				}
				return nil, testCase.failure
			}))
			requireProviderDiagnostic(t, err, providerManifestPathKey, testCase.reason)
			if finding, ok := ConfigurationFindingFromError(err); !ok || finding.Message == "" ||
				contains(finding.Message, "/operator/private") {
				t.Fatalf("unsafe document diagnostic = %#v / %v", finding, err)
			}
		})
	}
}

func providerManifestConfig(path string) Configuration {
	return Configuration{
		Claimed:              true,
		ProviderManifestPath: path,
	}
}

func testDocumentReader() DocumentReader {
	return DocumentReadFunc(func(path string, maximumBytes int64) ([]byte, DocumentReadFailure) {
		info, err := os.Lstat(path)
		if err != nil {
			return nil, DocumentUnavailable
		}
		if !info.Mode().IsRegular() {
			return nil, DocumentUnsafe
		}
		if info.Size() > maximumBytes {
			return nil, DocumentTooLarge
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, DocumentUnavailable
		}
		return content, ""
	})
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
	finding, ok := ConfigurationFindingFromError(err)
	if !ok {
		t.Fatalf("expected configuration error, got %v", err)
	}
	if finding.Path != wantPath || finding.ReasonCode != wantReason {
		t.Fatalf("finding = %#v, want path=%q reason=%q", finding, wantPath, wantReason)
	}
}

func contains(value string, substring string) bool {
	for index := 0; index+len(substring) <= len(value); index++ {
		if value[index:index+len(substring)] == substring {
			return true
		}
	}
	return false
}

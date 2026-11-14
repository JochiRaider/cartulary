package enterpriseauth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// These are admission fixtures, not trusted assertions or verified JWTs. The
// signed protocol tests independently qualify downstream signature/claim checks.
func samlAdmissionDocument(content string) string {
	return `<p:Response xmlns:p="urn:oasis:names:tc:SAML:2.0:protocol"><a:Assertion xmlns:a="urn:oasis:names:tc:SAML:2.0:assertion">` + content + `</a:Assertion></p:Response>`
}

func assertEnterpriseProviderBoundaries(t *testing.T) {
	t.Helper()
	// Authored limits are independent of production constants.
	const limit = 1048576
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	jwt := header + "." + strings.Repeat("a", limit-len(header)-4) + ".AA"
	if len(jwt) != limit || admitOIDCToken(jwt) != nil {
		t.Fatal("JWT admission rejected the exact byte limit")
	}
	if admitOIDCToken(jwt+"A") == nil {
		t.Fatal("JWT admission accepted one byte over the limit")
	}

	raw := samlAdmissionDocument(strings.Repeat("x", limit-len(samlAdmissionDocument(""))))
	if len(raw) != limit {
		t.Fatal("incorrect SAML byte-bound fixture")
	}
	if admitted, err := admitSAMLResponse(base64.StdEncoding.EncodeToString([]byte(raw))); err != nil || string(admitted) != raw {
		t.Fatal("SAML admission rejected or changed exact-limit bytes")
	}
	if _, err := admitSAMLResponse(base64.StdEncoding.EncodeToString([]byte(raw + "\n"))); err == nil {
		t.Fatal("SAML admission accepted one decoded byte over the limit")
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	encodedKeys, err := json.Marshal([]any{rsaJWK(key, "bounded-key")})
	if err != nil {
		t.Fatal(err)
	}
	keySet := `{"keys":` + string(encodedKeys) + `}`
	keySet += strings.Repeat(" ", limit-len(keySet))
	if admitOIDCKeySet([]byte(keySet)) != nil {
		t.Fatal("valid JWKS rejected at the exact byte limit")
	}
	if admitOIDCKeySet([]byte(keySet+" ")) == nil {
		t.Fatal("JWKS accepted one byte over the limit")
	}
	duplicate := `{"keys":` + string(encodedKeys) + `,"keys":` + string(encodedKeys) + `}`
	if admitOIDCKeySet([]byte(duplicate)) == nil {
		t.Fatal("duplicate valid JWKS fields admitted")
	}

	server, root := newProviderHTTPSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size := limit
		if r.URL.Path == "/oversized" {
			size++
		}
		_, _ = w.Write([]byte(strings.Repeat("x", size)))
	}))
	client, closeClient, err := providerHTTPClient(server.URL+"/boundary", root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeClient()
	response, err := client.Get(server.URL + "/boundary")
	if err != nil {
		t.Fatal("provider HTTP response rejected at the byte limit", err)
	}
	body, readErr := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if readErr != nil || len(body) != limit || string(body) != strings.Repeat("x", limit) {
		t.Fatal("provider HTTP boundary bytes truncated or changed")
	}
	if response, err := client.Get(server.URL + "/oversized"); err == nil {
		_ = response.Body.Close()
		t.Fatal("provider HTTP accepted one byte over the limit")
	}
}

func assertSAMLStructuralBounds(t *testing.T) {
	t.Helper()
	for _, tc := range []struct {
		name, raw string
		accept    bool
	}{
		{"depth64", samlAdmissionDocument(strings.Repeat("<x>", 62) + strings.Repeat("</x>", 62)), true},
		{"depth65", samlAdmissionDocument(strings.Repeat("<x>", 63) + strings.Repeat("</x>", 63)), false},
		// Response/Assertion contribute four tokens; every empty x adds two.
		{"tokens50000", samlAdmissionDocument(strings.Repeat("<x/>", 24998)), true},
		{"tokens50001", samlAdmissionDocument(strings.Repeat("<x/>", 24998)) + "\n", false},
		{"single root", samlAdmissionDocument(""), true},
		{"second root", samlAdmissionDocument("") + "<x/>", false},
		{"DTD", "<!DOCTYPE Response>" + samlAdmissionDocument(""), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := admitSAMLXML([]byte(tc.raw)); (err == nil) != tc.accept {
				t.Fatalf("XML admission accepted=%v, want %v", err == nil, tc.accept)
			}
		})
	}
}

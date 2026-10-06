package enterpriseauth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	jose "github.com/go-jose/go-jose/v4"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

const providerResponseLimit = 1 << 20

var errProviderPolicy = errors.New("enterprise provider cryptographic policy rejected")

// Keep protocol parsing, signature dispatch and claim checks in go-oidc/go-jose.
// Admission precedes their JWK decoder: certificate-bearing JWKs in the pinned
// dependency execute SHA-1 even when no SHA-1 thumbprint is supplied.
func admitOIDCKeySet(raw []byte) error {
	if len(raw) > providerResponseLimit || rejectDuplicateObjectMembers(raw) != nil {
		return errProviderPolicy
	}
	var set struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if json.Unmarshal(raw, &set) != nil || len(set.Keys) == 0 {
		return errProviderPolicy
	}
	for _, encoded := range set.Keys {
		var fields map[string]json.RawMessage
		if json.Unmarshal(encoded, &fields) != nil || fields == nil {
			return errProviderPolicy
		}
		for _, excluded := range []string{"d", "p", "q", "dp", "dq", "qi", "oth", "k", "x5c", "x5t", "x5t#S256", "x5u"} {
			if _, exists := fields[excluded]; exists {
				return errProviderPolicy
			}
		}
		var key jose.JSONWebKey
		if json.Unmarshal(encoded, &key) != nil || !key.Valid() || !key.IsPublic() || (key.Use != "" && key.Use != "sig") {
			return errProviderPolicy
		}
		if ops, exists := fields["key_ops"]; exists {
			var operations []string
			if json.Unmarshal(ops, &operations) != nil || len(operations) != 1 || operations[0] != "verify" {
				return errProviderPolicy
			}
		}
		switch public := key.Key.(type) {
		case *rsa.PublicKey:
			if cryptography.ValidateRSAPublicKey(public) != nil || (key.Algorithm != "" && key.Algorithm != "RS256" && key.Algorithm != "PS256") {
				return errProviderPolicy
			}
		case *ecdsa.PublicKey:
			if public.Curve != elliptic.P256() || (key.Algorithm != "" && key.Algorithm != "ES256") {
				return errProviderPolicy
			}
		default:
			return errProviderPolicy
		}
	}
	return nil
}

func admitOIDCToken(raw string) error {
	if len(raw) == 0 || len(raw) > providerResponseLimit || strings.Count(raw, ".") != 2 {
		return errProviderPolicy
	}
	header, _, _ := strings.Cut(raw, ".")
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(header)
	if err != nil || rejectDuplicateObjectMembers(decoded) != nil {
		return errProviderPolicy
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(decoded, &fields) != nil || fields == nil {
		return errProviderPolicy
	}
	for _, excluded := range []string{"jwk", "jku", "x5c", "x5t", "x5t#S256", "x5u"} {
		if _, exists := fields[excluded]; exists {
			return errProviderPolicy
		}
	}
	var algorithm string
	if json.Unmarshal(fields["alg"], &algorithm) != nil {
		return errProviderPolicy
	}
	switch algorithm {
	case "RS256", "PS256", "ES256":
		return nil
	}
	return errProviderPolicy
}

type providerTransport struct {
	transports map[string]*http.Transport
}

func (transport providerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if !validHTTPSProviderURL(request.URL.String()) {
		return nil, errProviderPolicy
	}
	qualified := transport.transports[request.URL.Host]
	if qualified == nil {
		return nil, errProviderPolicy
	}
	response, err := qualified.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, providerResponseLimit+1))
	if err != nil || len(body) > providerResponseLimit {
		return nil, errProviderPolicy
	}
	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	return response, nil
}

func providerHTTPClient(jwksURI, rootCertificatePath string, endpoints ...string) (*http.Client, func(), error) {
	transports := make(map[string]*http.Transport)
	closeTransports := func() {
		for _, transport := range transports {
			transport.CloseIdleConnections()
		}
	}
	for _, endpoint := range append([]string{jwksURI}, endpoints...) {
		if !validHTTPSProviderURL(endpoint) {
			closeTransports()
			return nil, nil, errProviderPolicy
		}
		parsed, err := url.Parse(endpoint)
		if err != nil {
			closeTransports()
			return nil, nil, errProviderPolicy
		}
		if _, exists := transports[parsed.Host]; exists {
			continue
		}
		tlsConfig, err := cryptography.TLSClient(cryptography.TLSClientOptions{ServerName: parsed.Hostname(), RootCertificatePath: rootCertificatePath})
		if err != nil {
			closeTransports()
			return nil, nil, err
		}
		transports[parsed.Host] = &http.Transport{TLSClientConfig: tlsConfig, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 10 * time.Second}
	}
	client := &http.Client{
		Transport: providerTransport{transports: transports}, Timeout: 15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return errProviderPolicy },
	}
	return client, closeTransports, nil
}

// Every callback reads the current configured key set. Protocol parsing and
// signature-input construction stay with go-jose, claims with go-oidc.
// The RSA hook supplies the fixed PS256 salt length its default lacks.
type oidcPolicyKeySet struct {
	client   *http.Client
	endpoint string
}

func (keys oidcPolicyKeySet) VerifySignature(ctx context.Context, token string) ([]byte, error) {
	if err := admitOIDCToken(token); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, keys.endpoint, nil)
	if err != nil {
		return nil, errProviderPolicy
	}
	response, err := keys.client.Do(request)
	if err != nil {
		return nil, errProviderPolicy
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errProviderPolicy
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, providerResponseLimit+1))
	if err != nil || admitOIDCKeySet(raw) != nil {
		return nil, errProviderPolicy
	}
	var set jose.JSONWebKeySet
	if json.Unmarshal(raw, &set) != nil {
		return nil, errProviderPolicy
	}
	signed, err := jose.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256, jose.PS256, jose.ES256})
	if err != nil || len(signed.Signatures) != 1 {
		return nil, errProviderPolicy
	}
	header := signed.Signatures[0].Header
	for _, key := range set.Keys {
		if (header.KeyID != "" && key.KeyID != header.KeyID) || (key.Algorithm != "" && key.Algorithm != header.Algorithm) {
			continue
		}
		verifier := key.Key
		if rsaKey, ok := key.Key.(*rsa.PublicKey); ok {
			verifier = oidcRSAVerifier{key: rsaKey}
		}
		if payload, err := signed.Verify(verifier); err == nil {
			return payload, nil
		}
	}
	return nil, errProviderPolicy
}

type oidcRSAVerifier struct{ key *rsa.PublicKey }

func (verifier oidcRSAVerifier) VerifyPayload(payload, signature []byte, algorithm jose.SignatureAlgorithm) error {
	switch algorithm {
	case jose.RS256:
		return cryptography.VerifyRSAPKCS1SHA256(verifier.key, payload, signature)
	case jose.PS256:
		return cryptography.VerifyRSAPSSSHA256(verifier.key, payload, signature)
	default:
		return errProviderPolicy
	}
}

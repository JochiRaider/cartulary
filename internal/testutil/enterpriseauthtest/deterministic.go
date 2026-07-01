package enterpriseauthtest

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type DeterministicOIDCVerifier struct{}

func (DeterministicOIDCVerifier) VerifyCallback(_ authn.EnterpriseAuthProviderRecord, query url.Values) (enterpriseauth.OIDCCallback, *httpapi.APIError) {
	state := strings.TrimSpace(query.Get("state"))
	code := strings.TrimSpace(query.Get("code"))
	nonce := strings.TrimSpace(query.Get("nonce"))
	subject := query.Get("subject")
	if state == "" || code == "" || nonce == "" {
		return enterpriseauth.OIDCCallback{}, providerResponseRejected("missing_required_field")
	}
	if code != "valid-code" {
		return enterpriseauth.OIDCCallback{}, providerResponseRejected("code_exchange_failed")
	}
	if subject == "" {
		return enterpriseauth.OIDCCallback{}, providerIdentityRejected("subject_missing")
	}
	return enterpriseauth.OIDCCallback{State: state, Nonce: nonce, ProviderSubject: subject}, nil
}

type DeterministicSAMLVerifier struct{}

type deterministicSAMLAssertion struct {
	Subject        string `json:"subject"`
	Issuer         string `json:"issuer"`
	Audience       string `json:"audience"`
	SignatureValid *bool  `json:"signature_valid"`
	ExpiresAt      string `json:"expires_at"`
}

func (DeterministicSAMLVerifier) VerifyACS(provider authn.EnterpriseAuthProviderRecord, form url.Values, now time.Time) (enterpriseauth.SAMLAssertionResult, *httpapi.APIError) {
	assertion, apiErr := decodeDeterministicSAMLAssertion(form.Get("SAMLResponse"))
	if apiErr != nil {
		return enterpriseauth.SAMLAssertionResult{}, apiErr
	}
	if provider.Issuer != nil && assertion.Issuer != *provider.Issuer {
		return enterpriseauth.SAMLAssertionResult{}, providerResponseRejected("issuer_mismatch")
	}
	if provider.Audience != nil && assertion.Audience != *provider.Audience {
		return enterpriseauth.SAMLAssertionResult{}, providerResponseRejected("audience_mismatch")
	}
	if assertion.SignatureValid == nil || !*assertion.SignatureValid {
		return enterpriseauth.SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	if assertion.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, assertion.ExpiresAt)
		if err != nil || !expiresAt.After(now.UTC()) {
			return enterpriseauth.SAMLAssertionResult{}, providerResponseRejected("assertion_expired")
		}
	}
	if assertion.Subject == "" {
		return enterpriseauth.SAMLAssertionResult{}, providerIdentityRejected("subject_missing")
	}
	return enterpriseauth.SAMLAssertionResult{ProviderSubject: assertion.Subject}, nil
}

func decodeDeterministicSAMLAssertion(value string) (deterministicSAMLAssertion, *httpapi.APIError) {
	if strings.TrimSpace(value) == "" {
		return deterministicSAMLAssertion{}, providerResponseRejected("missing_required_field")
	}
	payload := []byte(value)
	var assertion deterministicSAMLAssertion
	if err := json.Unmarshal(payload, &assertion); err != nil {
		decoded, decodeErr := base64.StdEncoding.DecodeString(value)
		if decodeErr != nil || json.Unmarshal(decoded, &assertion) != nil {
			return deterministicSAMLAssertion{}, providerResponseRejected("missing_required_field")
		}
	}
	return assertion, nil
}

func providerResponseRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "provider_response_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func providerIdentityRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "provider_identity_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

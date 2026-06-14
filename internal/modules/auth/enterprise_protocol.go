package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type enterpriseOIDCCallback struct {
	State           string
	Nonce           string
	ProviderSubject string
}

type enterpriseSAMLAssertionResult struct {
	ProviderSubject string
}

type enterpriseOIDCVerifier interface {
	VerifyCallback(authn.EnterpriseAuthProviderRecord, url.Values) (enterpriseOIDCCallback, *APIError)
}

type enterpriseSAMLVerifier interface {
	VerifyACS(authn.EnterpriseAuthProviderRecord, url.Values, time.Time) (enterpriseSAMLAssertionResult, *APIError)
}

type deterministicEnterpriseOIDCVerifier struct{}

func (deterministicEnterpriseOIDCVerifier) VerifyCallback(_ authn.EnterpriseAuthProviderRecord, query url.Values) (enterpriseOIDCCallback, *APIError) {
	state := strings.TrimSpace(query.Get("state"))
	code := strings.TrimSpace(query.Get("code"))
	nonce := strings.TrimSpace(query.Get("nonce"))
	subject := query.Get("subject")
	if state == "" || code == "" || nonce == "" {
		return enterpriseOIDCCallback{}, providerResponseRejected("missing_required_field")
	}
	if code != "valid-code" {
		return enterpriseOIDCCallback{}, providerResponseRejected("code_exchange_failed")
	}
	if subject == "" {
		return enterpriseOIDCCallback{}, providerIdentityRejected("subject_missing")
	}
	return enterpriseOIDCCallback{State: state, Nonce: nonce, ProviderSubject: subject}, nil
}

type deterministicEnterpriseSAMLVerifier struct{}

type deterministicSAMLAssertion struct {
	Subject        string `json:"subject"`
	Issuer         string `json:"issuer"`
	Audience       string `json:"audience"`
	SignatureValid *bool  `json:"signature_valid"`
	ExpiresAt      string `json:"expires_at"`
}

func (deterministicEnterpriseSAMLVerifier) VerifyACS(provider authn.EnterpriseAuthProviderRecord, form url.Values, now time.Time) (enterpriseSAMLAssertionResult, *APIError) {
	assertion, apiErr := decodeDeterministicSAMLAssertion(form.Get("SAMLResponse"))
	if apiErr != nil {
		return enterpriseSAMLAssertionResult{}, apiErr
	}
	if provider.Issuer != nil && assertion.Issuer != *provider.Issuer {
		return enterpriseSAMLAssertionResult{}, providerResponseRejected("issuer_mismatch")
	}
	if provider.Audience != nil && assertion.Audience != *provider.Audience {
		return enterpriseSAMLAssertionResult{}, providerResponseRejected("audience_mismatch")
	}
	if assertion.SignatureValid == nil || !*assertion.SignatureValid {
		return enterpriseSAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	if assertion.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, assertion.ExpiresAt)
		if err != nil || !expiresAt.After(now.UTC()) {
			return enterpriseSAMLAssertionResult{}, providerResponseRejected("assertion_expired")
		}
	}
	if assertion.Subject == "" {
		return enterpriseSAMLAssertionResult{}, providerIdentityRejected("subject_missing")
	}
	return enterpriseSAMLAssertionResult{ProviderSubject: assertion.Subject}, nil
}

func decodeDeterministicSAMLAssertion(value string) (deterministicSAMLAssertion, *APIError) {
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

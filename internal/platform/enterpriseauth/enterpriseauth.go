package enterpriseauth

import (
	"net/http"
	"net/url"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type OIDCCallback struct {
	State           string
	Nonce           string
	ProviderSubject string
}

type SAMLAssertionResult struct {
	ProviderSubject string
}

type OIDCVerifier interface {
	VerifyCallback(authn.EnterpriseAuthProviderRecord, url.Values) (OIDCCallback, *httpapi.APIError)
}

type SAMLVerifier interface {
	VerifyACS(authn.EnterpriseAuthProviderRecord, url.Values, time.Time) (SAMLAssertionResult, *httpapi.APIError)
}

type UnconfiguredOIDCVerifier struct{}

func (UnconfiguredOIDCVerifier) VerifyCallback(authn.EnterpriseAuthProviderRecord, url.Values) (OIDCCallback, *httpapi.APIError) {
	return OIDCCallback{}, verifierUnavailable()
}

type UnconfiguredSAMLVerifier struct{}

func (UnconfiguredSAMLVerifier) VerifyACS(authn.EnterpriseAuthProviderRecord, url.Values, time.Time) (SAMLAssertionResult, *httpapi.APIError) {
	return SAMLAssertionResult{}, verifierUnavailable()
}

func verifierUnavailable() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "provider_response_rejected",
		Message: "enterprise auth verifier is not configured",
		Details: map[string]any{"reason_code": "verifier_unavailable"},
	}
}

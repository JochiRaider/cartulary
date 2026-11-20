package auth

import (
	"net/url"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
)

func normalizeEnterpriseReturnTo(value string) (string, bool) {
	if strings.TrimSpace(value) == "" || strings.HasPrefix(value, "//") {
		return "", false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Fragment != "" {
		return "", false
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	if !strings.HasPrefix(parsed.Path, "/") {
		return "", false
	}
	return parsed.RequestURI(), true
}

func deterministicEnterpriseBeginRedirect(provider authn.EnterpriseAuthProviderRecord, _ string, transaction authn.EnterpriseAuthTransactionRecord, _ string) (enterpriseauth.BeginRedirect, error) {
	return enterpriseauth.BeginRedirect{
		URL:           enterpriseRedirectURL(provider, transaction),
		SAMLRequestID: transaction.SAMLRequestID,
	}, nil
}

func enterpriseRedirectURL(provider authn.EnterpriseAuthProviderRecord, transaction authn.EnterpriseAuthTransactionRecord) string {
	base := ""
	if provider.AuthorizationEndpoint != nil {
		base = *provider.AuthorizationEndpoint
	}
	if base == "" {
		switch provider.ProviderType {
		case "oidc":
			base = "/api/v1/auth/oidc/" + url.PathEscape(provider.ProviderKey) + "/callback"
		default:
			base = "/api/v1/auth/saml/" + url.PathEscape(provider.ProviderKey) + "/acs"
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	query := parsed.Query()
	if transaction.State != nil {
		query.Set("state", *transaction.State)
	}
	if transaction.Nonce != nil {
		query.Set("nonce", *transaction.Nonce)
	}
	if transaction.RelayState != nil {
		query.Set("RelayState", *transaction.RelayState)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func enterpriseSAMLCompletionURL(providerKey string, completionToken string) string {
	query := url.Values{}
	query.Set("completion", completionToken)
	return "/api/v1/auth/saml/" + url.PathEscape(providerKey) + "/acs/complete?" + query.Encode()
}

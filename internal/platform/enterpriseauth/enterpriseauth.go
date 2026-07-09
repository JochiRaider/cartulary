package enterpriseauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/crewjam/saml"
	"golang.org/x/oauth2"

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

type OIDCCallbackVerificationRequest struct {
	Provider     authn.EnterpriseAuthProviderRecord
	Transaction  authn.EnterpriseAuthTransactionRecord
	Values       url.Values
	PKCEVerifier string
	PublicOrigin string
	Env          map[string]string
	Now          time.Time
}

type SAMLACSVerificationRequest struct {
	Provider     authn.EnterpriseAuthProviderRecord
	Transaction  authn.EnterpriseAuthTransactionRecord
	Values       url.Values
	PublicOrigin string
	Now          time.Time
}

type BeginRedirect struct {
	URL           string
	SAMLRequestID *string
}

type OIDCVerifier interface {
	VerifyCallback(context.Context, OIDCCallbackVerificationRequest) (OIDCCallback, *httpapi.APIError)
}

type SAMLVerifier interface {
	VerifyACS(context.Context, SAMLACSVerificationRequest) (SAMLAssertionResult, *httpapi.APIError)
}

type UnconfiguredOIDCVerifier struct{}

func (UnconfiguredOIDCVerifier) VerifyCallback(context.Context, OIDCCallbackVerificationRequest) (OIDCCallback, *httpapi.APIError) {
	return OIDCCallback{}, verifierUnavailable()
}

type UnconfiguredSAMLVerifier struct{}

func (UnconfiguredSAMLVerifier) VerifyACS(context.Context, SAMLACSVerificationRequest) (SAMLAssertionResult, *httpapi.APIError) {
	return SAMLAssertionResult{}, verifierUnavailable()
}

type ProductionOIDCVerifier struct{}

func (ProductionOIDCVerifier) VerifyCallback(ctx context.Context, request OIDCCallbackVerificationRequest) (OIDCCallback, *httpapi.APIError) {
	code := strings.TrimSpace(request.Values.Get("code"))
	state := strings.TrimSpace(request.Values.Get("state"))
	if code == "" || state == "" {
		return OIDCCallback{}, providerResponseRejected("missing_required_field")
	}
	if request.Transaction.State == nil || state != *request.Transaction.State {
		return OIDCCallback{}, providerResponseRejected("state_mismatch")
	}
	if request.Transaction.Nonce == nil || *request.Transaction.Nonce == "" {
		return OIDCCallback{}, providerResponseRejected("nonce_mismatch")
	}
	if request.PKCEVerifier == "" {
		return OIDCCallback{}, providerResponseRejected("code_exchange_failed")
	}
	clientID, apiErr := requiredProviderValue(request.Provider.ClientID, "audience_mismatch")
	if apiErr != nil {
		return OIDCCallback{}, apiErr
	}
	clientSecret, apiErr := resolveOIDCClientSecret(request.Provider, request.Env)
	if apiErr != nil {
		return OIDCCallback{}, apiErr
	}
	tokenEndpoint, apiErr := requiredProviderValue(request.Provider.TokenEndpoint, "code_exchange_failed")
	if apiErr != nil {
		return OIDCCallback{}, apiErr
	}
	authorizationEndpoint, apiErr := requiredProviderValue(request.Provider.AuthorizationEndpoint, "code_exchange_failed")
	if apiErr != nil {
		return OIDCCallback{}, apiErr
	}
	issuer, apiErr := requiredProviderValue(request.Provider.Issuer, "issuer_mismatch")
	if apiErr != nil {
		return OIDCCallback{}, apiErr
	}

	oauthConfig := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authorizationEndpoint,
			TokenURL: tokenEndpoint,
		},
		RedirectURL: OIDCRedirectURL(request.PublicOrigin, request.Provider.ProviderKey),
		Scopes:      append([]string{"openid"}, request.Provider.AdditionalScopes...),
	}
	token, err := oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(request.PKCEVerifier))
	if err != nil {
		return OIDCCallback{}, providerResponseRejected("code_exchange_failed")
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return OIDCCallback{}, providerResponseRejected("missing_required_field")
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return OIDCCallback{}, providerResponseRejected(classifyOIDCDiscoveryError(err))
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
		Now: func() time.Time {
			if request.Now.IsZero() {
				return time.Now().UTC()
			}
			return request.Now.UTC()
		},
	})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return OIDCCallback{}, providerResponseRejected(classifyOIDCVerificationError(err))
	}
	if idToken.Nonce != *request.Transaction.Nonce {
		return OIDCCallback{}, providerResponseRejected("nonce_mismatch")
	}
	if idToken.Subject == "" {
		return OIDCCallback{}, providerIdentityRejected("subject_missing")
	}
	return OIDCCallback{State: state, Nonce: idToken.Nonce, ProviderSubject: idToken.Subject}, nil
}

type ProductionSAMLVerifier struct{}

func (ProductionSAMLVerifier) VerifyACS(_ context.Context, request SAMLACSVerificationRequest) (SAMLAssertionResult, *httpapi.APIError) {
	if strings.TrimSpace(request.Values.Get("SAMLResponse")) == "" {
		return SAMLAssertionResult{}, providerResponseRejected("missing_required_field")
	}
	relayState := strings.TrimSpace(request.Values.Get("RelayState"))
	if request.Transaction.RelayState == nil || relayState != *request.Transaction.RelayState {
		return SAMLAssertionResult{}, providerResponseRejected("relay_state_mismatch")
	}
	sp, err := SAMLServiceProvider(request.Provider, request.PublicOrigin)
	if err != nil {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	httpRequest, err := http.NewRequest(http.MethodPost, SAMLACSURL(request.PublicOrigin, request.Provider.ProviderKey), strings.NewReader(request.Values.Encode()))
	if err != nil {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := httpRequest.ParseForm(); err != nil {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	requestIDs := make([]string, 0, 1)
	if request.Transaction.SAMLRequestID != nil && *request.Transaction.SAMLRequestID != "" {
		requestIDs = append(requestIDs, *request.Transaction.SAMLRequestID)
	}
	assertion, err := sp.ParseResponse(httpRequest, requestIDs)
	if err != nil {
		return SAMLAssertionResult{}, providerResponseRejected(classifySAMLVerificationError(err))
	}
	subject := assertionSubject(assertion, request.Provider.SAMLSubjectSource)
	if subject == "" {
		return SAMLAssertionResult{}, providerIdentityRejected("subject_missing")
	}
	return SAMLAssertionResult{ProviderSubject: subject}, nil
}

func OIDCRedirectURL(publicOrigin string, providerKey string) string {
	return strings.TrimRight(publicOrigin, "/") + "/api/v1/auth/oidc/" + url.PathEscape(providerKey) + "/callback"
}

func SAMLACSURL(publicOrigin string, providerKey string) string {
	return strings.TrimRight(publicOrigin, "/") + "/api/v1/auth/saml/" + url.PathEscape(providerKey) + "/acs"
}

func SAMLMetadataURL(publicOrigin string, providerKey string) string {
	return strings.TrimRight(publicOrigin, "/") + "/api/v1/auth/saml/" + url.PathEscape(providerKey) + "/metadata"
}

func BuildBeginRedirect(provider authn.EnterpriseAuthProviderRecord, publicOrigin string, transaction authn.EnterpriseAuthTransactionRecord, pkceVerifier string) (BeginRedirect, error) {
	switch provider.ProviderType {
	case "oidc":
		if transaction.State == nil || transaction.Nonce == nil {
			return BeginRedirect{}, errors.New("enterprise auth OIDC transaction missing state or nonce")
		}
		clientID, err := requiredString(provider.ClientID, "missing OIDC client_id")
		if err != nil {
			return BeginRedirect{}, err
		}
		authURL, err := requiredString(provider.AuthorizationEndpoint, "missing OIDC authorization endpoint")
		if err != nil {
			return BeginRedirect{}, err
		}
		tokenURL, err := requiredString(provider.TokenEndpoint, "missing OIDC token endpoint")
		if err != nil {
			return BeginRedirect{}, err
		}
		config := oauth2.Config{
			ClientID: clientID,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
			RedirectURL: OIDCRedirectURL(publicOrigin, provider.ProviderKey),
			Scopes:      append([]string{"openid"}, provider.AdditionalScopes...),
		}
		return BeginRedirect{
			URL: config.AuthCodeURL(
				*transaction.State,
				oauth2.S256ChallengeOption(pkceVerifier),
				oauth2.SetAuthURLParam("nonce", *transaction.Nonce),
			),
		}, nil
	case "saml":
		if transaction.RelayState == nil {
			return BeginRedirect{}, errors.New("enterprise auth SAML transaction missing relay state")
		}
		sp, err := SAMLServiceProvider(provider, publicOrigin)
		if err != nil {
			return BeginRedirect{}, err
		}
		ssoURL, err := requiredString(provider.SAMLSSOURL, "missing SAML SSO URL")
		if err != nil {
			return BeginRedirect{}, err
		}
		authnRequest, err := sp.MakeAuthenticationRequest(ssoURL, saml.HTTPRedirectBinding, saml.HTTPPostBinding)
		if err != nil {
			return BeginRedirect{}, err
		}
		redirectURL, err := authnRequest.Redirect(*transaction.RelayState, &sp)
		if err != nil {
			return BeginRedirect{}, err
		}
		requestID := authnRequest.ID
		return BeginRedirect{URL: redirectURL.String(), SAMLRequestID: &requestID}, nil
	default:
		return BeginRedirect{}, fmt.Errorf("unsupported enterprise auth provider type %q", provider.ProviderType)
	}
}

func SAMLServiceProvider(provider authn.EnterpriseAuthProviderRecord, publicOrigin string) (saml.ServiceProvider, error) {
	entityID, err := requiredString(provider.SAMLSPHostEntityID, "missing SAML SP entity ID")
	if err != nil {
		return saml.ServiceProvider{}, err
	}
	idpEntityID, err := requiredString(provider.SAMLIDPEntityID, "missing SAML IdP entity ID")
	if err != nil {
		return saml.ServiceProvider{}, err
	}
	ssoURL, err := requiredString(provider.SAMLSSOURL, "missing SAML SSO URL")
	if err != nil {
		return saml.ServiceProvider{}, err
	}
	metadataURL, err := url.Parse(SAMLMetadataURL(publicOrigin, provider.ProviderKey))
	if err != nil {
		return saml.ServiceProvider{}, err
	}
	acsURL, err := url.Parse(SAMLACSURL(publicOrigin, provider.ProviderKey))
	if err != nil {
		return saml.ServiceProvider{}, err
	}
	keyDescriptors := make([]saml.KeyDescriptor, 0, len(provider.SAMLIDPSigningCertificate))
	for _, certificate := range provider.SAMLIDPSigningCertificate {
		keyDescriptors = append(keyDescriptors, saml.KeyDescriptor{
			Use: "signing",
			KeyInfo: saml.KeyInfo{
				X509Data: saml.X509Data{
					X509Certificates: []saml.X509Certificate{{Data: certificate}},
				},
			},
		})
	}
	return saml.ServiceProvider{
		EntityID:    entityID,
		MetadataURL: *metadataURL,
		AcsURL:      *acsURL,
		IDPMetadata: &saml.EntityDescriptor{
			EntityID: idpEntityID,
			IDPSSODescriptors: []saml.IDPSSODescriptor{{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
						KeyDescriptors:             keyDescriptors,
					},
				},
				SingleSignOnServices: []saml.Endpoint{{
					Binding:  saml.HTTPRedirectBinding,
					Location: ssoURL,
				}, {
					Binding:  saml.HTTPPostBinding,
					Location: ssoURL,
				}},
			}},
		},
	}, nil
}

func verifierUnavailable() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "provider_response_rejected",
		Message: "enterprise auth verifier is not configured",
		Details: map[string]any{"reason_code": "signature_invalid"},
	}
}

func requiredProviderValue(value *string, reasonCode string) (string, *httpapi.APIError) {
	if value == nil || *value == "" {
		return "", providerResponseRejected(reasonCode)
	}
	return *value, nil
}

func requiredString(value *string, message string) (string, error) {
	if value == nil || *value == "" {
		return "", errors.New(message)
	}
	return *value, nil
}

func resolveOIDCClientSecret(provider authn.EnterpriseAuthProviderRecord, env map[string]string) (string, *httpapi.APIError) {
	if provider.ClientSecretRefKind == nil || *provider.ClientSecretRefKind != "env" || provider.ClientSecretRefName == nil || *provider.ClientSecretRefName == "" {
		return "", providerResponseRejected("code_exchange_failed")
	}
	value, ok := lookupEnv(env, secretRefEnvName(*provider.ClientSecretRefName))
	if !ok || value == "" {
		return "", providerResponseRejected("code_exchange_failed")
	}
	return value, nil
}

func assertionSubject(assertion *saml.Assertion, source *authn.EnterpriseAuthSAMLSubjectSource) string {
	if assertion == nil {
		return ""
	}
	if source == nil || source.Kind == "name_id" {
		if assertion.Subject == nil || assertion.Subject.NameID == nil {
			return ""
		}
		return strings.TrimSpace(assertion.Subject.NameID.Value)
	}
	if source.Kind != "attribute" || source.AttributeName == "" {
		return ""
	}
	for _, statement := range assertion.AttributeStatements {
		for _, attribute := range statement.Attributes {
			if attribute.Name != source.AttributeName {
				continue
			}
			for _, value := range attribute.Values {
				if strings.TrimSpace(value.Value) != "" {
					return strings.TrimSpace(value.Value)
				}
				if value.NameID != nil && strings.TrimSpace(value.NameID.Value) != "" {
					return strings.TrimSpace(value.NameID.Value)
				}
			}
		}
	}
	return ""
}

func providerResponseRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "provider_response_rejected", Details: map[string]any{"reason_code": normalizeProviderResponseReason(reasonCode)}}
}

func providerIdentityRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "provider_identity_rejected", Details: map[string]any{"reason_code": normalizeProviderIdentityReason(reasonCode)}}
}

var providerResponseReasonRegistry = map[string]struct{}{
	"missing_required_field": {},
	"state_mismatch":         {},
	"relay_state_mismatch":   {},
	"nonce_mismatch":         {},
	"code_exchange_failed":   {},
	"issuer_mismatch":        {},
	"audience_mismatch":      {},
	"signature_invalid":      {},
	"assertion_expired":      {},
}

var providerIdentityReasonRegistry = map[string]struct{}{
	"subject_missing": {},
	"no_linked_user":  {},
	"ambiguous_link":  {},
	"inactive_user":   {},
}

func normalizeProviderResponseReason(reasonCode string) string {
	if _, ok := providerResponseReasonRegistry[reasonCode]; ok {
		return reasonCode
	}
	return "signature_invalid"
}

func normalizeProviderIdentityReason(reasonCode string) string {
	if _, ok := providerIdentityReasonRegistry[reasonCode]; ok {
		return reasonCode
	}
	return "no_linked_user"
}

func classifyOIDCDiscoveryError(err error) string {
	var issuerMismatch *oidc.IssuerMismatchError
	if errors.As(err, &issuerMismatch) {
		return "issuer_mismatch"
	}
	return "signature_invalid"
}

func classifyOIDCVerificationError(err error) string {
	var expired *oidc.TokenExpiredError
	if errors.As(err, &expired) {
		return "assertion_expired"
	}
	message := err.Error()
	switch {
	case strings.Contains(message, "issued by a different provider"):
		return "issuer_mismatch"
	case strings.Contains(message, "expected audience"), strings.Contains(message, "clientID must be provided"):
		return "audience_mismatch"
	default:
		return "signature_invalid"
	}
}

func classifySAMLVerificationError(err error) string {
	var invalidResponse *saml.InvalidResponseError
	if errors.As(err, &invalidResponse) && invalidResponse.PrivateErr != nil {
		err = invalidResponse.PrivateErr
	}
	message := err.Error()
	switch {
	case strings.Contains(message, "InResponseTo"), strings.Contains(message, "possible request IDs"):
		return "relay_state_mismatch"
	case strings.Contains(message, "Issuer does not match"), strings.Contains(message, "issuer is not"):
		return "issuer_mismatch"
	case strings.Contains(message, "AudienceRestriction"), strings.Contains(message, "audience restriction"),
		strings.Contains(message, "Recipient is not"), strings.Contains(message, "Destination"):
		return "audience_mismatch"
	case strings.Contains(message, "expired"), strings.Contains(message, "not yet valid"):
		return "assertion_expired"
	default:
		return "signature_invalid"
	}
}

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}
	return os.LookupEnv(key)
}

func secretRefEnvName(name string) string {
	return "CARTULARY_SECRET_" + normalizedSecretRefSuffix(name)
}

func normalizedSecretRefSuffix(name string) string {
	var builder strings.Builder
	previousUnderscore := false
	for _, r := range name {
		var next rune
		switch {
		case r >= 'a' && r <= 'z':
			next = r - ('a' - 'A')
		case r >= 'A' && r <= 'Z':
			next = r
		case r >= '0' && r <= '9':
			next = r
		default:
			next = '_'
		}
		if next == '_' {
			if builder.Len() == 0 || previousUnderscore {
				previousUnderscore = true
				continue
			}
			previousUnderscore = true
			builder.WriteRune(next)
			continue
		}
		previousUnderscore = false
		builder.WriteRune(next)
	}
	return strings.Trim(builder.String(), "_")
}

package enterpriseauth

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type ProductionSAMLVerifier struct{}

func (ProductionSAMLVerifier) VerifyACS(ctx context.Context, request SAMLACSVerificationRequest) (SAMLAssertionResult, *httpapi.APIError) {
	encoded := request.Values.Get("SAMLResponse")
	if strings.TrimSpace(encoded) == "" {
		return SAMLAssertionResult{}, providerResponseRejected("missing_required_field")
	}
	if request.Transaction.RelayState == nil || request.Values.Get("RelayState") != *request.Transaction.RelayState || request.Transaction.SAMLRequestID == nil || *request.Transaction.SAMLRequestID == "" {
		return SAMLAssertionResult{}, providerResponseRejected("relay_state_mismatch")
	}
	now := request.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	sp, err := samlServiceProvider(request.Provider, request.PublicOrigin, now)
	if err != nil || ctx.Err() != nil {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	if _, err := admitSAMLResponse(encoded); err != nil {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	response, err := sp.ValidateEncodedResponse(encoded)
	if err != nil {
		return SAMLAssertionResult{}, providerResponseRejected(classifySAMLVerificationError(err))
	}
	if ctx.Err() != nil || len(response.Assertions) != 1 || response.ID == "" {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	assertion := &response.Assertions[0]
	if !response.SignatureValidated && !assertion.SignatureValidated {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	if response.InResponseTo != *request.Transaction.SAMLRequestID || assertion.Subject == nil || assertion.Subject.SubjectConfirmation == nil || assertion.Subject.SubjectConfirmation.SubjectConfirmationData == nil {
		return SAMLAssertionResult{}, providerResponseRejected("relay_state_mismatch")
	}
	confirmation := assertion.Subject.SubjectConfirmation.SubjectConfirmationData
	if confirmation.InResponseTo != *request.Transaction.SAMLRequestID {
		return SAMLAssertionResult{}, providerResponseRejected("relay_state_mismatch")
	}
	if response.Destination != sp.AssertionConsumerServiceURL || confirmation.Recipient != sp.AssertionConsumerServiceURL {
		return SAMLAssertionResult{}, providerResponseRejected("audience_mismatch")
	}
	warnings, err := sp.VerifyAssertionConditions(assertion)
	if err != nil {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	if warnings.NotInAudience || len(assertion.Conditions.AudienceRestrictions) == 0 {
		return SAMLAssertionResult{}, providerResponseRejected("audience_mismatch")
	}
	if warnings.InvalidTime || !samlBeforeExpiry(now, assertion.Conditions.NotOnOrAfter) || !samlBeforeExpiry(now, confirmation.NotOnOrAfter) || response.IssueInstant.IsZero() || response.IssueInstant.After(now) || assertion.IssueInstant.IsZero() || assertion.IssueInstant.After(now) {
		return SAMLAssertionResult{}, providerResponseRejected("assertion_expired")
	}
	if assertion.ID == "" || assertion.Version != "2.0" || warnings.ProxyRestriction != nil {
		return SAMLAssertionResult{}, providerResponseRejected("signature_invalid")
	}
	subject := assertionSubject(assertion, request.Provider.SAMLSubjectSource)
	if subject == "" {
		return SAMLAssertionResult{}, providerIdentityRejected("subject_missing")
	}
	return SAMLAssertionResult{ProviderSubject: subject}, nil
}

func samlBeforeExpiry(now time.Time, value string) bool {
	expiry, err := time.Parse(time.RFC3339Nano, value)
	return err == nil && now.Before(expiry)
}

func samlServiceProvider(provider authn.EnterpriseAuthProviderRecord, publicOrigin string, now time.Time) (*saml2.SAMLServiceProvider, error) {
	entity, err := requiredString(provider.SAMLSPHostEntityID, "missing SAML SP entity ID")
	if err != nil {
		return nil, err
	}
	issuer, err := requiredString(provider.SAMLIDPEntityID, "missing SAML IdP entity ID")
	if err != nil {
		return nil, err
	}
	sso, err := requiredString(provider.SAMLSSOURL, "missing SAML SSO URL")
	if err != nil || !validHTTPSProviderURL(sso) || len(provider.SAMLIDPSigningCertificate) == 0 {
		return nil, errProviderPolicy
	}
	certs := make([]*x509.Certificate, 0, len(provider.SAMLIDPSigningCertificate))
	for _, value := range provider.SAMLIDPSigningCertificate {
		raw, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, errProviderPolicy
		}
		cert, err := x509.ParseCertificate(raw)
		if err != nil || admitSAMLSigningCertificate(cert) != nil || now.Before(cert.NotBefore) || !now.Before(cert.NotAfter) {
			return nil, errProviderPolicy
		}
		certs = append(certs, cert)
	}
	return &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL: sso, IdentityProviderSSOBinding: saml2.BindingHttpRedirect,
		IdentityProviderIssuer: issuer, ServiceProviderIssuer: entity, AudienceURI: entity,
		AssertionConsumerServiceURL: SAMLACSURL(publicOrigin, provider.ProviderKey),
		IDPCertificateStore:         &dsig.MemoryX509CertificateStore{Roots: certs},
		Clock:                       dsig.NewFakeClockAt(now), AllowMissingAttributes: true,
		MaximumDecompressedBodySize: providerResponseLimit, MaximumXMLTokens: 50000,
	}, nil
}

func buildSAMLRedirect(provider authn.EnterpriseAuthProviderRecord, origin string, transaction authn.EnterpriseAuthTransactionRecord) (BeginRedirect, error) {
	sp, err := samlServiceProvider(provider, origin, time.Now().UTC())
	if err != nil {
		return BeginRedirect{}, err
	}
	doc, err := sp.BuildAuthRequestDocumentNoSig()
	if err != nil {
		return BeginRedirect{}, err
	}
	if transaction.SAMLRequestID != nil && *transaction.SAMLRequestID != "" {
		doc.Root().CreateAttr("ID", *transaction.SAMLRequestID)
	}
	id := doc.Root().SelectAttrValue("ID", "")
	redirect, err := sp.BuildAuthURLRedirect(*transaction.RelayState, doc)
	if err != nil {
		return BeginRedirect{}, err
	}
	return BeginRedirect{URL: redirect, SAMLRequestID: &id}, nil
}

func assertionSubject(assertion *types.Assertion, source *authn.EnterpriseAuthSAMLSubjectSource) string {
	if source == nil || source.Kind == "name_id" {
		if assertion.Subject.NameID == nil {
			return ""
		}
		return strings.TrimSpace(assertion.Subject.NameID.Value)
	}
	if source.Kind != "attribute" || assertion.AttributeStatement == nil {
		return ""
	}
	var result string
	for _, attribute := range assertion.AttributeStatement.Attributes {
		if attribute.Name != source.AttributeName {
			continue
		}
		for _, value := range attribute.Values {
			candidate := strings.TrimSpace(value.Value)
			if value.NameID != nil {
				candidate = strings.TrimSpace(value.NameID.Value)
			}
			if candidate == "" {
				continue
			}
			if result != "" {
				return ""
			}
			result = candidate
		}
	}
	return result
}

func classifySAMLVerificationError(err error) string {
	var invalid saml2.ErrInvalidValue
	if errors.As(err, &invalid) {
		switch invalid.Key {
		case "Issuer":
			return "issuer_mismatch"
		case "Destination", "Recipient":
			return "audience_mismatch"
		case "InResponseTo":
			return "relay_state_mismatch"
		case "NotOnOrAfter", "NotBefore":
			return "assertion_expired"
		}
	}
	return "signature_invalid"
}

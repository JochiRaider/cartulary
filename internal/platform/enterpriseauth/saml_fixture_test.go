package enterpriseauth

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

func signedSAMLFixture(t *testing.T) SAMLACSVerificationRequest { return signedSAMLFixtureWith(t, nil) }

// An independently authored IdP message: neither the production SP serializer
// nor its SAML response types construct the assertion under test.
func signedSAMLFixtureWith(t *testing.T, mutate func(*etree.Element)) SAMLACSVerificationRequest {
	return signedSAMLFixtureScope(t, mutate, false)
}
func signedSAMLFixtureScope(t *testing.T, mutate func(*etree.Element), assertionOnly bool) SAMLACSVerificationRequest {
	t.Helper()
	key, cert := testSAMLKeyPair(t)
	origin, relay := "https://cartulary.example.test", "saml-relay-state"
	provider := authn.EnterpriseAuthProviderRecord{
		ProviderKey: "corp-saml", ProviderType: "saml",
		SAMLIDPEntityID: strPtr("https://idp.example.test/metadata"), SAMLSSOURL: strPtr("https://idp.example.test/sso"),
		SAMLIDPSigningCertificate: []string{base64.StdEncoding.EncodeToString(cert.Raw)},
		SAMLSPHostEntityID:        strPtr(SAMLMetadataURL(origin, "corp-saml")), SAMLSubjectSource: &authn.EnterpriseAuthSAMLSubjectSource{Kind: "name_id"},
	}
	transaction := authn.EnterpriseAuthTransactionRecord{ProviderKey: provider.ProviderKey, ProviderType: provider.ProviderType, RelayState: &relay}
	begin, err := BuildBeginRedirect(provider, origin, transaction, "")
	if err != nil {
		t.Fatal(err)
	}
	transaction.SAMLRequestID = begin.SAMLRequestID
	now := time.Now().UTC().Truncate(time.Second)
	stamp := func(v time.Time) string { return v.Format(time.RFC3339) }
	acs := SAMLACSURL(origin, provider.ProviderKey)
	doc := etree.NewDocument()
	err = doc.ReadFromString(fmt.Sprintf(`<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="_response" Version="2.0" IssueInstant="%s" Destination="%s" InResponseTo="%s"><saml:Issuer>%s</saml:Issuer><samlp:Status><samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></samlp:Status><saml:Assertion ID="_assertion" Version="2.0" IssueInstant="%s"><saml:Issuer>%s</saml:Issuer><saml:Subject><saml:NameID>saml-subject-123</saml:NameID><saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer"><saml:SubjectConfirmationData Recipient="%s" InResponseTo="%s" NotOnOrAfter="%s"/></saml:SubjectConfirmation></saml:Subject><saml:Conditions NotBefore="%s" NotOnOrAfter="%s"><saml:AudienceRestriction><saml:Audience>%s</saml:Audience></saml:AudienceRestriction></saml:Conditions></saml:Assertion></samlp:Response>`, stamp(now), acs, *begin.SAMLRequestID, *provider.SAMLIDPEntityID, stamp(now), *provider.SAMLIDPEntityID, acs, *begin.SAMLRequestID, stamp(now.Add(time.Minute)), stamp(now.Add(-time.Minute)), stamp(now.Add(time.Minute)), *provider.SAMLSPHostEntityID))
	if err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(doc.Root())
	}
	signer := dsig.NewDefaultSigningContext(dsig.TLSCertKeyStore(tls.Certificate{PrivateKey: key, Certificate: [][]byte{cert.Raw}}))
	signer.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	if err := signer.SetSignatureMethod(dsig.RSASHA256SignatureMethod); err != nil {
		t.Fatal(err)
	}
	toSign := doc.Root()
	if assertionOnly {
		toSign = doc.Root().FindElement("saml:Assertion")
		toSign.CreateAttr("xmlns:saml", samlAssertionNamespace)
	}
	signed, err := signer.SignEnveloped(toSign)
	if err != nil {
		t.Fatal(err)
	}
	if assertionOnly {
		doc.Root().RemoveChild(toSign)
		doc.Root().AddChild(signed)
	} else {
		doc.SetRoot(signed)
	}
	raw, err := doc.WriteToBytes()
	if err != nil {
		t.Fatal(err)
	}
	return SAMLACSVerificationRequest{Provider: provider, Transaction: transaction, PublicOrigin: origin, Now: now, Values: url.Values{"SAMLResponse": {base64.StdEncoding.EncodeToString(raw)}, "RelayState": {relay}}}
}

func TestEnterpriseSAMLStructuralAdmission(t *testing.T) {
	assertSAMLStructuralBounds(t)
	t.Run("signed assertion alone", func(t *testing.T) {
		r := signedSAMLFixtureScope(t, nil, true)
		if _, err := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), r); err != nil {
			sp, buildErr := samlServiceProvider(r.Provider, r.PublicOrigin, r.Now)
			if buildErr != nil {
				t.Fatal(buildErr)
			}
			_, validationErr := sp.ValidateEncodedResponse(r.Values.Get("SAMLResponse"))
			t.Fatal("admitted assertion-only signature rejected", err, validationErr)
		}
	})
	for _, attack := range []string{"unsigned", "signed value changed", "wrapped assertion", "foreign signature"} {
		t.Run(attack, func(t *testing.T) {
			r := signedSAMLFixtureScope(t, nil, attack == "wrapped assertion")
			raw, _ := base64.StdEncoding.DecodeString(r.Values.Get("SAMLResponse"))
			doc := etree.NewDocument()
			if err := doc.ReadFromBytes(raw); err != nil {
				t.Fatal(err)
			}
			root := doc.Root()
			switch attack {
			case "unsigned":
				signature := root.FindElement(".//ds:Signature")
				if signature == nil {
					t.Fatal("fixture signature absent")
				}
				signature.Parent().RemoveChild(signature)
			case "signed value changed":
				root.FindElement(".//saml:NameID").SetText("different-account")
			case "wrapped assertion":
				signed := root.FindElement("saml:Assertion")
				root.RemoveChild(signed)
				extension := root.CreateElement("samlp:Extensions")
				extension.AddChild(signed)
				untrusted := signed.Copy()
				untrusted.CreateAttr("ID", "_unsigned")
				signature := untrusted.FindElement(".//ds:Signature")
				if signature == nil {
					t.Fatal("fixture signature absent")
				}
				signature.Parent().RemoveChild(signature)
				untrusted.FindElement(".//saml:NameID").SetText("different-account")
				root.AddChild(untrusted)
			case "foreign signature":
				signature := root.FindElement(".//ds:Signature")
				if signature == nil {
					t.Fatal("fixture signature absent")
				}
				signature.Space = "foreign"
				signature.CreateAttr("xmlns:foreign", "urn:foreign")
			}
			raw, err := doc.WriteToBytes()
			if err != nil {
				t.Fatal(err)
			}
			r.Values.Set("SAMLResponse", base64.StdEncoding.EncodeToString(raw))
			if _, err := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), r); err == nil {
				t.Fatal("signature coverage attack accepted")
			}
		})
	}

	for _, name := range []string{"Issuer", "Subject", "Conditions", "NameID", "AudienceRestriction", "SubjectConfirmation", "SubjectConfirmationData"} {
		t.Run("missing "+name, func(t *testing.T) {
			r := signedSAMLFixtureWith(t, func(root *etree.Element) {
				e := root.FindElement(".//saml:Assertion//saml:" + name)
				if e == nil {
					t.Fatal("missing fixture element")
				}
				e.Parent().RemoveChild(e)
			})
			if _, err := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), r); err == nil {
				t.Fatal("missing required element accepted")
			}
		})
	}
	for _, name := range []string{"foreign NameID", "multiple assertions", "duplicate ID", "duplicate Subject", "missing destination", "expiry boundary", "wrong confirmation request", "qualified time attribute", "bearer NotBefore"} {
		t.Run(name, func(t *testing.T) {
			r := signedSAMLFixtureWith(t, func(root *etree.Element) {
				a := root.FindElement("saml:Assertion")
				switch name {
				case "qualified time attribute":
					e := a.FindElement("saml:Conditions")
					expiry := e.SelectAttrValue("NotOnOrAfter", "")
					e.RemoveAttr("NotOnOrAfter")
					e.CreateAttr("xmlns:foreign", "urn:foreign")
					e.CreateAttr("foreign:NotOnOrAfter", expiry)
				case "bearer NotBefore":
					a.FindElement("saml:Subject/saml:SubjectConfirmation/saml:SubjectConfirmationData").CreateAttr("NotBefore", root.SelectAttrValue("IssueInstant", ""))
				case "foreign NameID":
					e := a.FindElement("saml:Subject/saml:NameID")
					e.Space = "foreign"
					e.CreateAttr("xmlns:foreign", "urn:foreign")
				case "multiple assertions":
					e := a.Copy()
					e.CreateAttr("ID", "_second")
					root.AddChild(e)
				case "duplicate ID":
					a.CreateAttr("ID", "_response")
				case "duplicate Subject":
					a.AddChild(a.FindElement("saml:Subject").Copy())
				case "missing destination":
					root.RemoveAttr("Destination")
				case "expiry boundary":
					a.FindElement("saml:Conditions").CreateAttr("NotOnOrAfter", root.SelectAttrValue("IssueInstant", ""))
				case "wrong confirmation request":
					a.FindElement("saml:Subject/saml:SubjectConfirmation/saml:SubjectConfirmationData").CreateAttr("InResponseTo", "_other")
				}
			})
			if _, err := (ProductionSAMLVerifier{}).VerifyACS(context.Background(), r); err == nil {
				t.Fatal("invalid signed response accepted")
			}
		})
	}

}

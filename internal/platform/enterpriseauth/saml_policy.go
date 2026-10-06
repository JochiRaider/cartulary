package enterpriseauth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"time"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

func admitSAMLSigningCertificate(certificate *x509.Certificate) error {
	if certificate == nil {
		return errProviderPolicy
	}
	key, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok || cryptography.ValidateRSAPublicKey(key) != nil || time.Now().Before(certificate.NotBefore) || !time.Now().Before(certificate.NotAfter) || (certificate.KeyUsage != 0 && certificate.KeyUsage&x509.KeyUsageDigitalSignature == 0) {
		return errProviderPolicy
	}
	return nil
}

// This is admission only. The unchanged bytes still go through crewjam's
// round-trip validator, signature canonicalization, trust and assertion checks.
func admitSAMLResponse(encoded string) ([]byte, error) {
	if len(encoded) == 0 || len(encoded) > base64.StdEncoding.EncodedLen(providerResponseLimit) {
		return nil, errProviderPolicy
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) > providerResponseLimit {
		return nil, errProviderPolicy
	}
	document := etree.NewDocument()
	if document.ReadFromBytes(raw) != nil || document.Root() == nil {
		return nil, errProviderPolicy
	}
	if err := admitSAMLAlgorithms(document.Root()); err != nil {
		return nil, err
	}
	return raw, nil
}

func admitSAMLAlgorithms(root *etree.Element) error {
	// Iterative traversal avoids a second recursive stack on untrusted XML.
	pending := []*etree.Element{root}
	for len(pending) > 0 {
		element := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		switch element.Tag {
		case "EncryptedAssertion", "EncryptedData", "EncryptedKey":
			return errProviderPolicy
		case "SignatureMethod":
			if element.SelectAttrValue("Algorithm", "") != dsig.RSASHA256SignatureMethod {
				return errProviderPolicy
			}
		case "DigestMethod":
			if element.SelectAttrValue("Algorithm", "") != "http://www.w3.org/2001/04/xmlenc#sha256" {
				return errProviderPolicy
			}
		case "X509Certificate":
			encoded := strings.Join(strings.Fields(element.Text()), "")
			raw, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return errProviderPolicy
			}
			certificate, err := x509.ParseCertificate(raw)
			if err != nil || admitSAMLSigningCertificate(certificate) != nil {
				return errProviderPolicy
			}
		}
		pending = append(pending, element.ChildElements()...)
	}
	return nil
}

type samlSignaturePolicy struct{}

func (samlSignaturePolicy) VerifySignature(validation *dsig.ValidationContext, element *etree.Element) error {
	if err := admitSAMLAlgorithms(element); err != nil {
		return err
	}
	certificates, err := validation.CertificateStore.Certificates()
	if err != nil {
		return errProviderPolicy
	}
	for _, certificate := range certificates {
		if err := admitSAMLSigningCertificate(certificate); err != nil {
			return err
		}
	}
	_, err = validation.Validate(element)
	return err
}

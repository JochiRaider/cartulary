package enterpriseauth

import (
	"bytes"
	"encoding/xml"
	"io"
)

const samlAssertionNamespace = "urn:oasis:names:tc:SAML:2.0:assertion"
const samlProtocolNamespace = "urn:oasis:names:tc:SAML:2.0:protocol"

// Bound parser work before constructing any DOM. Reject ambiguous singleton
// elements and identities before the signature library chooses an assertion.
// This admission pass grants no trust; the original bytes are verified later.
func admitSAMLXML(raw []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	ids := map[string]bool{}
	type frame struct {
		name     xml.Name
		children map[xml.Name]int
	}
	stack := []frame{}
	roots, assertions := 0, 0
	for tokens := 0; ; tokens++ {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil || tokens >= 50000 {
			return errProviderPolicy
		}
		switch value := token.(type) {
		case xml.Directive:
			return errProviderPolicy
		case xml.CharData:
			if len(stack) == 0 && len(bytes.TrimSpace(value)) != 0 {
				return errProviderPolicy
			}
		case xml.StartElement:
			if len(stack) >= 64 {
				return errProviderPolicy
			}
			if len(stack) == 0 {
				roots++
				if roots != 1 || value.Name != (xml.Name{Space: samlProtocolNamespace, Local: "Response"}) {
					return errProviderPolicy
				}
			} else {
				parent := &stack[len(stack)-1]
				parent.children[value.Name]++
				switch value.Name.Local {
				case "Issuer", "Subject", "Conditions", "NameID", "SubjectConfirmation", "SubjectConfirmationData", "Status", "StatusCode", "AttributeStatement", "AuthnStatement", "Signature":
					if parent.children[value.Name] > 1 {
						return errProviderPolicy
					}
				}
			}
			switch value.Name.Local {
			case "Assertion":
				assertions++
				if assertions != 1 || len(stack) != 1 || value.Name.Space != samlAssertionNamespace {
					return errProviderPolicy
				}
			case "Issuer", "Subject", "Conditions", "NameID", "SubjectConfirmation", "SubjectConfirmationData", "AudienceRestriction", "Audience", "AttributeStatement", "Attribute", "AttributeValue", "AuthnStatement":
				if value.Name.Space != samlAssertionNamespace {
					return errProviderPolicy
				}
			case "Signature", "SignedInfo", "SignatureValue", "SignatureMethod", "DigestMethod", "DigestValue", "Reference", "Transforms", "Transform", "CanonicalizationMethod", "KeyInfo", "X509Data", "X509Certificate":
				if value.Name.Space != "http://www.w3.org/2000/09/xmldsig#" {
					return errProviderPolicy
				}
			case "Response", "Status", "StatusCode":
				if value.Name.Space != samlProtocolNamespace {
					return errProviderPolicy
				}
			}
			attributes := map[xml.Name]bool{}
			for _, attribute := range value.Attr {
				if attributes[attribute.Name] {
					return errProviderPolicy
				}
				attributes[attribute.Name] = true
				switch attribute.Name.Local {
				case "Version", "IssueInstant", "Destination", "InResponseTo", "Recipient", "Method", "NotBefore", "NotOnOrAfter", "Value", "Name", "Format", "Algorithm", "URI":
					if attribute.Name.Space != "" {
						return errProviderPolicy
					}
				}
				if value.Name.Local == "SubjectConfirmationData" && attribute.Name.Local == "NotBefore" {
					return errProviderPolicy
				}

				if attribute.Name.Local == "ID" || attribute.Name.Local == "Id" || attribute.Name.Local == "id" {
					if attribute.Name.Space != "" || attribute.Name.Local != "ID" || attribute.Value == "" || ids[attribute.Value] {
						return errProviderPolicy
					}
					ids[attribute.Value] = true
				}
			}
			stack = append(stack, frame{name: value.Name, children: map[xml.Name]int{}})
		case xml.EndElement:
			if len(stack) == 0 {
				return errProviderPolicy
			}
			stack = stack[:len(stack)-1]
		}
	}
	if roots != 1 || assertions != 1 || len(stack) != 0 {
		return errProviderPolicy
	}
	return nil
}

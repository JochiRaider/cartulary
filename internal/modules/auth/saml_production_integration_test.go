package auth_test

import (
	"bytes"
	"compress/flate"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

// Exercise the public protocol and session owner with the real verifier. The
// IdP response is independently authored XML, not a production SAML serializer.
func TestEnterpriseSAMLProductionACS_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	const origin = "https://cartulary.example.test"
	const issuer = "https://idp.example.test/metadata"
	const entity = origin + "/api/v1/auth/saml/signed-local/metadata"
	const acs = origin + "/api/v1/auth/saml/signed-local/acs"
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: "signed-saml", TestRouteMode: httptestx.TestRouteModeDisabled,
		Env: map[string]string{
			"CARTULARY__APPLICATION__PUBLIC_ORIGIN":                        origin,
			"CARTULARY__ENTERPRISE_AUTHENTICATION__CLAIMED":                "true",
			"CARTULARY__ENTERPRISE_AUTHENTICATION__PROVIDER_MANIFEST_PATH": fixtures.Path("enterprise-auth", "empty.json"),
		},
	})
	server, db := harness.Server, harness.DB
	userID := flowtest.SeedActor(t, db, flowtest.ActorAttributes{Email: "signed-saml@example.test", DisplayName: "Signed SAML"}).ID.String()
	providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{Key: "signed-local", Type: "saml", Name: "Local signed fixture", Enabled: true, Interactive: true})
	seedEnterpriseBinding(t, db, userID, providerID, "signed-subject", userID)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	certificate := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certificates, err := json.Marshal([]string{base64.StdEncoding.EncodeToString(der)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `UPDATE enterprise_auth_providers SET saml_idp_entity_id=$2,saml_sso_url='https://idp.example.test/sso',saml_idp_signing_certificates=$3,saml_sp_entity_id=$4,saml_subject_source='{"kind":"name_id"}' WHERE id::text=$1`, providerID, issuer, certificates, entity); err != nil {
		t.Fatal(err)
	}
	signer := dsig.NewDefaultSigningContext(dsig.TLSCertKeyStore(tls.Certificate{PrivateKey: key, Certificate: [][]byte{der}}))
	signer.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	if err := signer.SetSignatureMethod(dsig.RSASHA256SignatureMethod); err != nil {
		t.Fatal(err)
	}
	for _, tamper := range []bool{true, false} {
		begin := beginEnterpriseAuth(t, server, "signed-local")
		redirect := parseEnterpriseRedirect(t, begin.redirectURL)
		compressed, err := base64.StdEncoding.DecodeString(redirect.Query().Get("SAMLRequest"))
		if err != nil {
			t.Fatal(err)
		}
		reader := flate.NewReader(bytes.NewReader(compressed))
		var request struct {
			ID     string `xml:"ID,attr"`
			ACS    string `xml:"AssertionConsumerServiceURL,attr"`
			Issuer string `xml:"Issuer"`
		}
		err = xml.NewDecoder(reader).Decode(&request)
		_ = reader.Close()
		if err != nil || request.ID == "" || request.ACS != acs || request.Issuer != entity {
			t.Fatal("invalid SP request", err)
		}
		stamp := time.Now().UTC().Truncate(time.Second)
		doc := etree.NewDocument()
		err = doc.ReadFromString(fmt.Sprintf(`<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="_response" Version="2.0" IssueInstant="%s" Destination="%s" InResponseTo="%s"><saml:Issuer>%s</saml:Issuer><samlp:Status><samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></samlp:Status><saml:Assertion ID="_assertion" Version="2.0" IssueInstant="%s"><saml:Issuer>%s</saml:Issuer><saml:Subject><saml:NameID>signed-subject</saml:NameID><saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer"><saml:SubjectConfirmationData Recipient="%s" InResponseTo="%s" NotOnOrAfter="%s"/></saml:SubjectConfirmation></saml:Subject><saml:Conditions NotBefore="%s" NotOnOrAfter="%s"><saml:AudienceRestriction><saml:Audience>%s</saml:Audience></saml:AudienceRestriction></saml:Conditions></saml:Assertion></samlp:Response>`, stamp.Format(time.RFC3339), acs, request.ID, issuer, stamp.Format(time.RFC3339), issuer, acs, request.ID, stamp.Add(time.Minute).Format(time.RFC3339), stamp.Add(-time.Minute).Format(time.RFC3339), stamp.Add(time.Minute).Format(time.RFC3339), entity))
		if err != nil {
			t.Fatal(err)
		}
		signed, err := signer.SignEnveloped(doc.Root())
		if err != nil {
			t.Fatal(err)
		}
		doc.SetRoot(signed)
		if tamper {
			doc.Root().FindElement(".//saml:NameID").SetText("other-subject")
		}
		raw, err := doc.WriteToBytes()
		if err != nil {
			t.Fatal(err)
		}
		values := url.Values{"RelayState": {redirect.Query().Get("RelayState")}, "SAMLResponse": {base64.StdEncoding.EncodeToString(raw)}}
		response := doFormNoRedirect(t, server.HTTP.URL+"/api/v1/auth/saml/signed-local/acs", values)
		requireNoSessionCookie(t, response)
		if tamper {
			requireErrorReason(t, response, http.StatusConflict, "provider_response_rejected", "signature_invalid")
			if count := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, userID); count != 0 {
				t.Fatal("rejected signature issued a session")
			}
			continue
		}
		if response.StatusCode != http.StatusSeeOther {
			t.Fatalf("signed ACS status=%d", response.StatusCode)
		}
		location := response.Header.Get("Location")
		_ = response.Body.Close()
		if !strings.HasPrefix(location, "/api/v1/auth/saml/signed-local/acs/complete?completion=") {
			t.Fatal("unexpected completion route")
		}
		completion := doNoRedirect(t, http.MethodGet, server.HTTP.URL+location, nil, httptestx.WithCookies(begin.cookie))
		if completion.StatusCode != http.StatusSeeOther {
			t.Fatalf("completion status=%d", completion.StatusCode)
		}
		cookies := authcookietest.RequireAuthCookies(t, completion.Cookies())
		_ = completion.Body.Close()
		data := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(cookies.Session)), http.StatusOK)["data"].(map[string]any)
		if data["user_id"] != userID || data["provider_type"] != "saml" {
			t.Fatal("signed subject/session binding changed")
		}
		replay := doNoRedirect(t, http.MethodGet, server.HTTP.URL+location, nil, httptestx.WithCookies(begin.cookie))
		requireErrorReason(t, replay, http.StatusConflict, "enterprise_auth_transaction_rejected", "already_used")
		requireNoSessionCookie(t, replay)
		acsReplay := doFormNoRedirect(t, server.HTTP.URL+"/api/v1/auth/saml/signed-local/acs", values)
		if acsReplay.StatusCode < 400 {
			t.Fatal("replayed assertion accepted")
		}
		requireNoSessionCookie(t, acsReplay)
		_ = acsReplay.Body.Close()
		if count := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, userID); count != 1 {
			t.Fatal("replay changed session multiplicity")
		}
	}
}

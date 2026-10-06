package cryptography

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

func TestTLSCertificateAdmissionBeforeVerification(t *testing.T) {
	ca, server, client := tlsFixture(t)
	if _, err := TLSClient(TLSClientOptions{RootCertificatePath: writeTLSFixture(t, t.TempDir(), "root.pem", ca.CertificatePEM)}); err == nil {
		t.Fatal("unbound destination admitted")
	}
	issuerKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"odd bits", "small exponent", "even modulus"} {
		t.Run(name, func(t *testing.T) {
			public := &rsa.PublicKey{N: new(big.Int).Set(issuerKey.N), E: issuerKey.E}
			switch name {
			case "odd bits":
				public.N = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 2049), big.NewInt(1))
			case "small exponent":
				public.E = 3
			case "even modulus":
				public.N.SetBit(public.N, 0, 0)
			}
			intermediateTemplate := &x509.Certificate{SerialNumber: big.NewInt(51), Subject: pkix.Name{CommonName: "excluded fixture intermediate"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
			intermediateDER, err := ca.Sign(intermediateTemplate, public)
			if err != nil {
				t.Fatal(err)
			}
			intermediate, err := x509.ParseCertificate(intermediateDER)
			if err != nil {
				t.Fatal(err)
			}
			if validTLSCertificate(intermediate, time.Now()) {
				t.Fatal("excluded issuer reaches chain signature dispatch")
			}
			// Sign with an admitted fixture key, retaining the excluded key in the
			// transmitted intermediate. No excluded private operation is performed.
			parent := *intermediate
			parent.PublicKey = &issuerKey.PublicKey
			leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			leafTemplate := &x509.Certificate{SerialNumber: big.NewInt(52), DNSNames: []string{"localhost"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
			leafDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, &parent, &leafKey.PublicKey, issuerKey)
			if err != nil {
				t.Fatal(err)
			}
			external := &tls.Config{MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13, CurvePreferences: []tls.CurveID{tls.CurveP256}, Certificates: []tls.Certificate{{Certificate: [][]byte{leafDER, intermediateDER, ca.Certificate.Raw}, PrivateKey: leafKey}}}
			if err := tlsHandshake(t, external, client); err == nil {
				t.Fatal("excluded certificate chain admitted")
			}
			path := writeTLSFixture(t, t.TempDir(), "unsupported-root.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: intermediateDER}))
			if _, err := TLSClient(TLSClientOptions{ServerName: "localhost", RootCertificatePath: path}); err == nil {
				t.Fatal("excluded trust anchor admitted")
			}
		})
	}
	// Standard-library EKU validation remains mandatory after parameter admission.
	identity, err := ca.Issue("wrong usage", []string{"localhost"}, x509.ExtKeyUsageClientAuth)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := tls.X509KeyPair(identity.CertificatePEM, identity.PrivateKeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	external := server.Clone()
	external.Certificates = []tls.Certificate{pair}
	if err := tlsHandshake(t, external, client); err == nil {
		t.Fatal("client-only certificate accepted as server")
	}
	// Mutation of returned metadata cannot widen the immutable verification pool.
	replacement, replacementServer, _ := tlsFixture(t)
	client.RootCAs.AddCert(replacement.Certificate)
	if err := tlsHandshake(t, replacementServer, client); err == nil {
		t.Fatal("mutable metadata widened trust")
	}
}

func TestTLSSystemTrustBundleAdmission(t *testing.T) {
	ca, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	expiredTemplate := *ca.Certificate
	expiredTemplate.NotAfter = time.Now().Add(-time.Minute)
	expiredDER, err := ca.Sign(&expiredTemplate, ca.Certificate.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	expiredPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: expiredDER})
	bundle := append(append([]byte{}, expiredPEM...), ca.CertificatePEM...)
	path := writeTLSFixture(t, t.TempDir(), "bundle.pem", bundle)
	expectedRoots := x509.NewCertPool()
	expectedRoots.AddCert(ca.Certificate)
	roots, err := loadTLSRoots(path, systemTLSRootMaximumBytes, true)
	if err != nil || !roots.Equal(expectedRoots) {
		t.Fatalf("system policy filtering: %v", err)
	}
	if _, err := loadTLSRoots(path, systemTLSRootMaximumBytes, false); err == nil {
		t.Fatal("explicit CA binding silently discarded invalid anchor")
	}
	path = writeTLSFixture(t, t.TempDir(), "expired.pem", expiredPEM)
	if _, err := loadTLSRoots(path, systemTLSRootMaximumBytes, true); err == nil {
		t.Fatal("empty qualified trust accepted")
	}
	path = writeTLSFixture(t, t.TempDir(), "malformed.pem", append(bundle, []byte("garbage")...))
	if _, err := loadTLSRoots(path, systemTLSRootMaximumBytes, true); err == nil {
		t.Fatal("malformed system bundle accepted")
	}
}

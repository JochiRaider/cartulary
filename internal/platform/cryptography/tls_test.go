package cryptography

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

func TestTLSVerifiedTransport(t *testing.T) {
	ca, server, client := tlsFixture(t)
	if err := tlsHandshake(t, server, client); err != nil {
		t.Fatal(err)
	}
	t.Run("remote identity policy", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			key      func() (crypto.Signer, error)
			accepted bool
		}{
			{"P256", func() (crypto.Signer, error) { return ecdsa.GenerateKey(elliptic.P256(), rand.Reader) }, true},
			{"P384", func() (crypto.Signer, error) { return ecdsa.GenerateKey(elliptic.P384(), rand.Reader) }, true},
			{"RSA2048", func() (crypto.Signer, error) { return rsa.GenerateKey(rand.Reader, 2048) }, true},
			{"P521", func() (crypto.Signer, error) { return ecdsa.GenerateKey(elliptic.P521(), rand.Reader) }, false},
			{"Ed25519", func() (crypto.Signer, error) { _, key, err := ed25519.GenerateKey(rand.Reader); return key, err }, false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				key, err := tc.key()
				if err != nil {
					t.Fatal(err)
				}
				external := tlsExternalIdentity(t, ca, key, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
				if err := tlsHandshake(t, external, client); (err == nil) != tc.accepted {
					t.Fatalf("accepted=%t: %v", tc.accepted, err)
				}
			})
		}
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			name        string
			from, until time.Time
		}{
			{"expired", time.Now().Add(-2 * time.Hour), time.Now().Add(-time.Hour)},
			{"future", time.Now().Add(time.Hour), time.Now().Add(2 * time.Hour)},
		} {
			t.Run(tc.name, func(t *testing.T) {
				if err := tlsHandshake(t, tlsExternalIdentity(t, ca, key, tc.from, tc.until), client); err == nil {
					t.Fatal("invalid peer validity accepted")
				}
			})
		}
	})
	t.Run("mutual certificate", func(t *testing.T) {
		identity, err := ca.Issue("cartulary_runtime_login", nil, x509.ExtKeyUsageClientAuth)
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		clientWithCertificate, err := TLSClient(TLSClientOptions{
			ServerName: "localhost", RootCertificatePath: writeTLSFixture(t, dir, "root.pem", ca.CertificatePEM),
			ClientCertificatePath: writeTLSFixture(t, dir, "client.pem", identity.CertificatePEM),
			ClientPrivateKeyPath:  writeTLSFixture(t, dir, "client.key", identity.PrivateKeyPEM),
		})
		if err != nil {
			t.Fatal(err)
		}
		mutual := server.Clone()
		mutual.ClientAuth = tls.RequireAndVerifyClientCert
		mutual.ClientCAs = clientWithCertificate.RootCAs.Clone()
		if err := tlsHandshake(t, mutual, clientWithCertificate); err != nil {
			t.Fatal(err)
		}
		if err := tlsHandshake(t, mutual, client); err == nil {
			t.Fatal("missing client identity accepted")
		}
	})
	for _, tc := range []struct {
		name   string
		change func(*tls.Config, *tls.Config)
	}{
		{"wrong name", func(_ *tls.Config, c *tls.Config) { c.ServerName = "elsewhere.invalid" }},
		{"client TLS12", func(_ *tls.Config, c *tls.Config) { c.MinVersion = tls.VersionTLS12; c.MaxVersion = tls.VersionTLS12 }},
		{"server TLS12", func(s *tls.Config, _ *tls.Config) { s.MinVersion = tls.VersionTLS12; s.MaxVersion = tls.VersionTLS12 }},
		{"no admitted key agreement", func(_ *tls.Config, c *tls.Config) { c.CurvePreferences = []tls.CurveID{tls.X25519} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, c := server.Clone(), client.Clone()
			tc.change(s, c)
			if err := tlsHandshake(t, s, c); err == nil {
				t.Fatal("excluded transport accepted")
			}
		})
	}
	t.Run("replacement needs current trust and restart", func(t *testing.T) {
		_, replacedServer, replacedClient := tlsFixture(t)
		if err := tlsHandshake(t, replacedServer, client); err == nil {
			t.Fatal("replacement accepted under old isolated trust")
		}
		if err := tlsHandshake(t, replacedServer, replacedClient); err != nil {
			t.Fatal(err)
		}
	})
	if !server.SessionTicketsDisabled || !client.SessionTicketsDisabled || client.ClientSessionCache != nil {
		t.Fatal("certificate replacement can reuse session state")
	}
}

func TestTLSCertificateAndFileAdmission(t *testing.T) {
	ca, _, _ := tlsFixture(t)
	identity, err := ca.Issue("localhost", []string{"localhost"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath := writeTLSFixture(t, dir, "server.pem", identity.CertificatePEM)
	keyPath := writeTLSFixture(t, dir, "server.key", identity.PrivateKeyPEM)
	if _, err := TLSServer("wrong.invalid", certPath, keyPath); err == nil {
		t.Fatal("wrong origin admitted")
	}
	other, err := ca.Issue("localhost", []string{"localhost"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	otherKey := writeTLSFixture(t, dir, "other.key", other.PrivateKeyPEM)
	if _, err := TLSServer("localhost", certPath, otherKey); err == nil {
		t.Fatal("mismatched key admitted")
	}
	if _, err := TLSClient(TLSClientOptions{ServerName: "localhost", ClientCertificatePath: certPath, ClientPrivateKeyPath: keyPath}); err == nil {
		t.Fatal("server-only client identity admitted")
	}
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"empty", nil}, {"oversized", []byte(strings.Repeat("x", TLSPEMMaximumBytes+1))},
		{"trailing garbage", append(append([]byte{}, identity.CertificatePEM...), []byte("garbage")...)},
		{"leading garbage", append([]byte("garbage"), identity.CertificatePEM...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTLSFixture(t, t.TempDir(), "invalid.pem", tc.data)
			if _, err := TLSServer("localhost", path, keyPath); err == nil || err.Error() != ErrTLSConfiguration.Error() {
				t.Fatalf("certificate admission: %v", err)
			}
		})
	}
	link := filepath.Join(dir, "link.pem")
	if err := os.Symlink(certPath, link); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{link, dir, "relative.pem", filepath.Join(dir, "missing"), dir + "/../" + filepath.Base(dir) + "/server.pem"} {
		if _, err := TLSServer("localhost", path, keyPath); err == nil {
			t.Fatalf("unsafe file admitted: %s", path)
		}
	}
	block, _ := pem.Decode(identity.CertificatePEM)
	base, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*x509.Certificate)
	}{
		{"expired", func(c *x509.Certificate) { c.NotAfter = time.Now().Add(-time.Minute) }},
		{"future", func(c *x509.Certificate) { c.NotBefore = time.Now().Add(time.Hour) }},
		{"weak RSA", func(c *x509.Certificate) {
			c.PublicKey = &rsa.PublicKey{N: new(big.Int).Lsh(big.NewInt(1), 1023), E: 65537}
		}},
		{"unsupported curve", func(c *x509.Certificate) {
			key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			c.PublicKey = key.Public()
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cert := *base
			tc.change(&cert)
			raw, err := ca.Sign(&cert, cert.PublicKey)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parseTLSCertificates(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw})); err == nil {
				t.Fatal("unsupported certificate admitted")
			}
		})
	}
	for _, algorithm := range []x509.SignatureAlgorithm{x509.SHA1WithRSA, x509.ECDSAWithSHA1, x509.PureEd25519} {
		cert := *base
		cert.SignatureAlgorithm = algorithm
		if validTLSCertificate(&cert, time.Now()) {
			t.Fatal("unsupported signature admitted")
		}
	}
}

func tlsFixture(t *testing.T) (*tlstest.Authority, *tls.Config, *tls.Config) {
	t.Helper()
	ca, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	identity, err := ca.Issue("localhost", []string{"localhost", "127.0.0.1"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	server, err := TLSServer("localhost", writeTLSFixture(t, dir, "server.pem", identity.CertificatePEM), writeTLSFixture(t, dir, "server.key", identity.PrivateKeyPEM))
	if err != nil {
		t.Fatal(err)
	}
	client, err := TLSClient(TLSClientOptions{ServerName: "localhost", RootCertificatePath: writeTLSFixture(t, dir, "root.pem", ca.CertificatePEM)})
	if err != nil {
		t.Fatal(err)
	}
	return ca, server, client
}

// This external-service fixture deliberately permits identities that Cartulary
// rejects. Positive evidence always retains normal client chain/name checks.
func tlsExternalIdentity(t *testing.T, ca *tlstest.Authority, key crypto.Signer, from, until time.Time) *tls.Config {
	t.Helper()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), DNSNames: []string{"localhost"},
		NotBefore: from, NotAfter: until,
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	raw, err := ca.Sign(template, key.Public())
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Config{
		MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		Certificates:     []tls.Certificate{{Certificate: [][]byte{raw, ca.Certificate.Raw}, PrivateKey: key}},
	}
}

func writeTLSFixture(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path, err := tlstest.WriteFile(dir, name, data)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func tlsHandshake(t *testing.T, server, client *tls.Config) error {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	finished := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			finished <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		secured := tls.Server(conn, server)
		if err = secured.Handshake(); err == nil {
			_, err = secured.Write([]byte{1})
		}
		finished <- err
	}()
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", listener.Addr().String(), client)
	if err == nil {
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		var payload [1]byte
		_, err = io.ReadFull(conn, payload[:])
		_ = conn.Close()
	}
	serverErr := <-finished
	if err != nil {
		return err
	}
	return serverErr
}

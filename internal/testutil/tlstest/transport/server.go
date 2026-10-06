package transport

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

// ServerIdentity owns one temporary authority and server identity. Clients
// explicitly trust that authority; host trust stores and defaults are untouched.
type ServerIdentity struct {
	CertificatePath     string
	PrivateKeyPath      string
	RootCertificatePath string
	TLS                 *tls.Config
}

func NewServer(t testing.TB, hosts ...string) ServerIdentity {
	t.Helper()
	if len(hosts) == 0 {
		t.Fatal("server identity requires a host")
	}
	authority, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	identity, err := authority.Issue("isolated application", hosts, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	write := func(name string, data []byte) string {
		path, err := tlstest.WriteFile(directory, name, data)
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	result := ServerIdentity{
		CertificatePath:     write("server.crt", identity.CertificatePEM),
		PrivateKeyPath:      write("server.key", identity.PrivateKeyPEM),
		RootCertificatePath: write("root.crt", authority.CertificatePEM),
	}
	result.TLS, err = cryptography.TLSServer(hosts[0], result.CertificatePath, result.PrivateKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func (identity ServerIdentity) Client(t testing.TB, host string) *http.Client {
	t.Helper()
	configuration, err := cryptography.TLSClient(cryptography.TLSClientOptions{ServerName: host, RootCertificatePath: identity.RootCertificatePath})
	if err != nil {
		t.Fatal(err)
	}
	transport := &http.Transport{TLSClientConfig: configuration, TLSHandshakeTimeout: 5 * time.Second}
	t.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Transport: transport, Timeout: 10 * time.Second}
}

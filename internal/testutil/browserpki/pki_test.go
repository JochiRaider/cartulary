package browserpki

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserPKIRequiresFreshPrivateState(t *testing.T) {
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o700); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(parent, "tls")
	if err := Provision(directory); err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil || info.Mode().Perm() != 0o600 {
			t.Fatal("fixture material is not private")
		}
		return data
	}
	root := read("ca.pem")
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(root) {
		t.Fatal("invalid fixture root")
	}
	for _, role := range []string{"server", "frontend"} {
		block, _ := pem.Decode(read(role + ".crt"))
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		for _, host := range []string{"127.0.0.1", "localhost", "::1"} {
			if _, err := certificate.Verify(x509.VerifyOptions{Roots: roots, DNSName: host}); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := certificate.Verify(x509.VerifyOptions{Roots: roots, DNSName: "other.invalid"}); err == nil {
			t.Fatal("fixture certificate covers an unrelated host")
		}
	}
	if bytes.Equal(read("server.key"), read("frontend.key")) {
		t.Fatal("frontend and server share a private key")
	}
	if err := Provision(directory); err == nil || !bytes.Equal(root, read("ca.pem")) {
		t.Fatal("retry replaced existing fixture identity")
	}
	if err := os.Chmod(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Provision(filepath.Join(parent, "unsafe")); err == nil {
		t.Fatal("accepted a nonprivate parent")
	}
}

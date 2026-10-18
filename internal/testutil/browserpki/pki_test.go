package browserpki

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
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
	t.Run("owner_scope_and_leaf_isolation", func(t *testing.T) {
		parent := t.TempDir()
		if err := os.Chmod(parent, 0700); err != nil {
			t.Fatal(err)
		}
		object, err := tlstest.NewAuthority()
		if err != nil {
			t.Fatal(err)
		}
		owner := &Owner{}
		for _, name := range []string{"first", "second", "foreign"} {
			scope := "functional"
			if name == "foreign" {
				scope = "measurement"
			}
			if err := owner.Provision(filepath.Join(parent, name), scope, object.CertificatePEM); err != nil {
				t.Fatal(err)
			}
		}
		read := func(name, file string) []byte {
			t.Helper()
			b, err := os.ReadFile(filepath.Join(parent, name, file))
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		if !bytes.Equal(read("first", "trust.pem"), read("second", "trust.pem")) {
			t.Fatal("compatible stacks lost stable trust")
		}
		if bytes.Equal(read("first", "server.key"), read("second", "server.key")) {
			t.Fatal("stack keys reused")
		}
		roots := x509.NewCertPool()
		roots.AppendCertsFromPEM(read("first", "trust.pem"))
		for _, pair := range [][2]string{{"first", "wrong-name.crt"}, {"first", "untrusted.crt"}, {"foreign", "server.crt"}} {
			block, _ := pem.Decode(read(pair[0], pair[1]))
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := cert.Verify(x509.VerifyOptions{Roots: roots, DNSName: "localhost"}); err == nil {
				t.Fatal("foreign authority or hostname accepted")
			}
		}
		if err := owner.Provision(filepath.Join(parent, "changed"), "functional", nil); err == nil {
			t.Fatal("mutable trust accepted")
		}
		if _, err := os.Stat(filepath.Join(parent, "changed")); !os.IsNotExist(err) {
			t.Fatal("failed issuance published files")
		}
	})
}

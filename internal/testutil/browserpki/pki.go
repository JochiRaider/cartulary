// Package browserpki provisions fresh identities for disposable browser fixtures.
package browserpki

import (
	"crypto/x509"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
	"os"
	"path/filepath"
)

func Provision(directory string) (retErr error) {
	if !filepath.IsAbs(directory) || filepath.Clean(directory) != directory {
		return errors.New("absolute normalized fixture directory required")
	}
	parent, err := os.Lstat(filepath.Dir(directory))
	if err != nil || !parent.IsDir() || parent.Mode().Perm() != 0o700 {
		return errors.New("private fixture parent required")
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, os.RemoveAll(directory))
		}
	}()
	authority, err := tlstest.NewAuthority()
	if err != nil {
		return err
	}
	if _, err := tlstest.WriteFile(directory, "ca.pem", authority.CertificatePEM); err != nil {
		return err
	}
	untrusted, err := tlstest.NewAuthority()
	if err != nil {
		return err
	}
	for _, role := range []string{"server", "frontend", "wrong-name", "untrusted"} {
		issuer := authority
		hosts := []string{"127.0.0.1", "localhost", "::1"}
		if role == "wrong-name" {
			hosts = []string{"unrelated.invalid"}
		}
		if role == "untrusted" {
			issuer = untrusted
		}
		identity, err := issuer.Issue(role, hosts, x509.ExtKeyUsageServerAuth)
		if err != nil {
			return err
		}
		for name, data := range map[string][]byte{role + ".crt": identity.CertificatePEM, role + ".key": identity.PrivateKeyPEM} {
			if _, err := tlstest.WriteFile(directory, name, data); err != nil {
				return err
			}
		}
	}
	return nil
}

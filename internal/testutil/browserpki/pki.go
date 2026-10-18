// Package browserpki provisions fresh identities for disposable browser fixtures.
package browserpki

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
	"os"
	"path/filepath"
	"sync"
)

// Owner keeps signing keys only in the native owner. Public bundles are fixed
// for each scope; an interrupted owner cannot resume issuance from disk.
type Owner struct {
	mu     sync.Mutex
	scopes map[string]*trustScope
}
type trustScope struct {
	authority  *tlstest.Authority
	objectRoot []byte
}

func (o *Owner) Provision(directory, scope string, objectRoot []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if scope == "" {
		return errors.New("trust scope required")
	}
	if o.scopes == nil {
		o.scopes = map[string]*trustScope{}
	}
	entry := o.scopes[scope]
	if entry == nil {
		authority, err := tlstest.NewAuthority()
		if err != nil {
			return err
		}
		entry = &trustScope{authority, bytes.Clone(objectRoot)}
		o.scopes[scope] = entry
	} else if !bytes.Equal(entry.objectRoot, objectRoot) {
		return errors.New("trust scope object-store root changed")
	}
	return provision(directory, entry.authority, entry.objectRoot)
}

// Provision owns a dedicated authority for standalone negative-TLS fixtures.
func Provision(directory string) error { return (&Owner{}).Provision(directory, "dedicated", nil) }

func provision(directory string, authority *tlstest.Authority, objectRoot []byte) (retErr error) {
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
	if _, err := tlstest.WriteFile(directory, "ca.pem", authority.CertificatePEM); err != nil {
		return err
	}
	if len(objectRoot) > 0 {
		block, rest := pem.Decode(objectRoot)
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 || len(bytes.TrimSpace(rest)) != 0 {
			return errors.New("invalid object-store fixture authority")
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil || !certificate.IsCA || certificate.CheckSignatureFrom(certificate) != nil {
			return errors.New("invalid object-store fixture root")
		}
		bundle := append(bytes.Clone(authority.CertificatePEM), objectRoot...)
		if _, err := tlstest.WriteFile(directory, "trust.pem", bundle); err != nil {
			return err
		}
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

// Package packagepki issues private, short-lived identities only for disposable
// package qualification. Production certificates are provisioned by operators.
package packagepki

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

func Provision(directory string) (retErr error) {
	if !filepath.IsAbs(directory) || filepath.Clean(directory) != directory {
		return errors.New("absolute normalized fixture directory required")
	}
	parent, err := os.Lstat(filepath.Dir(directory))
	if err != nil || !parent.IsDir() || parent.Mode().Perm() != 0700 {
		return errors.New("private fixture parent required")
	}
	if err := os.Mkdir(directory, 0700); err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, os.RemoveAll(directory))
		}
	}()
	ca, err := tlstest.NewAuthority()
	if err != nil {
		return err
	}
	write := func(name string, data []byte) error {
		path, err := tlstest.WriteFile(directory, name, data)
		if err != nil {
			return err
		}
		// The host parent is private. Each container receives only its own files,
		// allowing the unprivileged application UID to read a read-only bind mount.
		return os.Chmod(path, 0644)
	}
	if err := write("ca.pem", ca.CertificatePEM); err != nil {
		return err
	}
	for _, name := range []string{"postgres", "seaweed", "application", "migration", "runtime", "recovery", "restore-migration", "restore-recovery"} {
		usage := x509.ExtKeyUsageClientAuth
		var hosts []string
		switch name {
		case "postgres":
			usage = x509.ExtKeyUsageServerAuth
			hosts = []string{"postgres", "localhost", "127.0.0.1", "::1"}
		case "seaweed":
			usage = x509.ExtKeyUsageServerAuth
			hosts = []string{"seaweedfs-s3", "localhost", "127.0.0.1", "::1"}
		case "application":
			usage = x509.ExtKeyUsageServerAuth
			hosts = []string{"app", "localhost", "127.0.0.1", "::1"}
		}
		identity, err := ca.Issue(name, hosts, usage)
		if err != nil {
			return err
		}
		if err := write(name+".crt", identity.CertificatePEM); err != nil {
			return err
		}
		if err := write(name+".key", identity.PrivateKeyPEM); err != nil {
			return err
		}
		// Preissue independent replacement leaves; the authority key never leaves
		// memory. Only the qualification runner can see these unmounted files.
		replacement, err := ca.Issue(name, hosts, usage)
		if err != nil {
			return err
		}
		if err := write(name+".replacement.crt", replacement.CertificatePEM); err != nil {
			return err
		}
		if err := write(name+".replacement.key", replacement.PrivateKeyPEM); err != nil {
			return err
		}
	}
	return nil
}

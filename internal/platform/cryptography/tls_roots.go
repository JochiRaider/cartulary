package cryptography

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/securefile"
)

const (
	systemTLSRootBundle       = "/etc/ssl/certs/ca-certificates.crt"
	systemTLSRootMaximumBytes = 16 << 20
)

// The supported Ubuntu/WSL2 and application-image profile uses a PEM system
// bundle. Capture it once per process so trust replacement requires restart.
// SSL_CERT_FILE is the ordinary process-local system-bundle binding, useful for
// isolated harness processes. Directory search and alternate fallback stores
// would bypass parameter admission and are not part of this profile.
var systemTLSRoots = sync.OnceValues(func() (*x509.CertPool, error) {
	if os.Getenv("SSL_CERT_DIR") != "" {
		return nil, ErrTLSConfiguration
	}
	path := os.Getenv("SSL_CERT_FILE")
	if path == "" {
		path = systemTLSRootBundle
	}
	return loadTLSRoots(path, systemTLSRootMaximumBytes, true)
})

func loadTLSRoots(path string, maximum int64, filter bool) (*x509.CertPool, error) {
	document, err := securefile.Read(path, maximum)
	if err != nil || document.Size() == 0 {
		return nil, ErrTLSConfiguration
	}
	pool := x509.NewCertPool()
	count := 0
	for remaining := bytes.TrimSpace(document.Bytes()); len(remaining) > 0; {
		if !bytes.HasPrefix(remaining, []byte("-----BEGIN CERTIFICATE-----")) {
			return nil, ErrTLSConfiguration
		}
		block, rest := pem.Decode(remaining)
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return nil, ErrTLSConfiguration
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, ErrTLSConfiguration
		}
		if certificate.IsCA && certificate.KeyUsage&x509.KeyUsageCertSign != 0 && validTLSCertificate(certificate, time.Now()) {
			pool.AddCert(certificate)
			count++
		} else if !filter {
			return nil, ErrTLSConfiguration
		}
		remaining = bytes.TrimSpace(rest)
	}
	if count == 0 {
		return nil, ErrTLSConfiguration
	}
	return pool, nil
}

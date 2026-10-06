package s3test

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func prepareObjectStoreTLS(req testcontainers.ContainerRequest) (_ testcontainers.ContainerRequest, _ string, retErr error) {
	parent, _, err := suiteservices.ResolveSuiteRuntimeDir(nil)
	if err != nil {
		return req, "", err
	}
	directory, err := os.MkdirTemp(parent, "object-store-tls-")
	if err != nil {
		return req, "", err
	}
	success := false
	defer func() {
		if !success {
			retErr = errors.Join(retErr, os.RemoveAll(directory))
		}
	}()
	ca, err := tlstest.NewAuthority()
	if err != nil {
		return req, "", err
	}
	if _, err = tlstest.WriteFile(directory, "root.pem", ca.CertificatePEM); err != nil {
		return req, "", err
	}
	identity, err := ca.Issue("object-store", []string{"localhost", "127.0.0.1", "::1", "object-store", "host.docker.internal", "gateway.docker.internal"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		return req, "", err
	}
	for _, file := range []struct {
		name string
		data []byte
	}{{"server.pem", identity.CertificatePEM}, {"server.key", identity.PrivateKeyPEM}} {
		path, err := tlstest.WriteFile(directory, file.name, file.data)
		if err != nil {
			return req, "", err
		}
		req.Files = append(req.Files, testcontainers.ContainerFile{HostFilePath: path, ContainerFilePath: "/etc/cartulary-s3-tls/" + file.name, FileMode: 0o600})
	}
	// A separate key serves the existing CORS proxy; the CA signing key stays
	// in memory and only the SeaweedFS identity enters the container.
	proxy, err := ca.Issue("object-store-cors-proxy", []string{"127.0.0.1", "localhost"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		return req, "", err
	}
	if _, err := tlstest.WriteFile(directory, "proxy.pem", proxy.CertificatePEM); err != nil {
		return req, "", err
	}
	if _, err := tlstest.WriteFile(directory, "proxy.key", proxy.PrivateKeyPEM); err != nil {
		return req, "", err
	}
	// SeaweedFS with port.https=0 and a keypair serves TLS on the main S3
	// listener. A separate positive HTTPS port would leave plaintext open.
	req.Cmd = append(append([]string(nil), req.Cmd...), "-s3.port.https=0", "-s3.cert.file=/etc/cartulary-s3-tls/server.pem", "-s3.key.file=/etc/cartulary-s3-tls/server.key")
	// The image entrypoint drops to its seaweed account. Give only that
	// account access to the private identity before retaining the normal entrypoint.
	req.Entrypoint = []string{"/bin/sh", "-ec", "chown seaweed:seaweed /etc/cartulary-s3-tls/server.pem /etc/cartulary-s3-tls/server.key; exec /entrypoint.sh \"$@\"", "--"}
	success = true
	return req, directory, nil
}

func (h *Harness) RootCertificatePath() string {
	if h.RootCertificateFile != "" {
		return h.RootCertificateFile
	}
	if h.tlsDirectory != "" {
		return filepath.Join(h.tlsDirectory, "root.pem")
	}
	return ""
}

// ProxyTLSFiles returns owned fixture identities only; attached harnesses do
// not allocate or infer private key material.
func (h *Harness) ProxyTLSFiles() (certificate, privateKey string) {
	if h.tlsDirectory == "" {
		return "", ""
	}
	return filepath.Join(h.tlsDirectory, "proxy.pem"), filepath.Join(h.tlsDirectory, "proxy.key")
}

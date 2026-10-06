package cryptography

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/securefile"
)

const TLSPEMMaximumBytes = 262144

var (
	ErrTLSConfiguration = errors.New("TLS configuration rejected")
	ErrTLSPeer          = errors.New("TLS peer rejected")
)

// TLSClientOptions contains bindings, not protocol-policy overrides. Empty
// RootCertificatePath selects the admitted system bundle. ServerName must be
// the explicit destination identity, including for IP endpoints. Client
// certificate and key must be supplied together.
type TLSClientOptions struct {
	ServerName            string
	RootCertificatePath   string
	ClientCertificatePath string
	ClientPrivateKeyPath  string
}

func TLSClient(options TLSClientOptions) (*tls.Config, error) {
	if !validTLSServerName(options.ServerName) {
		return nil, ErrTLSConfiguration
	}
	config := tlsPolicy()
	config.ServerName = options.ServerName
	var roots *x509.CertPool
	var err error
	if options.RootCertificatePath != "" {
		roots, err = loadTLSRoots(options.RootCertificatePath, TLSPEMMaximumBytes, false)
	} else {
		roots, err = systemTLSRoots()
	}
	if err != nil {
		return nil, err
	}
	// The verification closure owns its immutable roots. RootCAs is a copy for
	// transport metadata; callers cannot widen the admitted verification pool.
	config.RootCAs = roots.Clone()
	if options.ClientCertificatePath != "" || options.ClientPrivateKeyPath != "" {
		certificate, err := loadTLSCertificate(options.ClientCertificatePath, options.ClientPrivateKeyPath, x509.ExtKeyUsageClientAuth)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{certificate}
	}
	// Go calls its normal chain verifier before VerifyConnection. Disable that
	// duplicate pass so excluded certificate parameters never reach signature
	// dispatch; the mandatory callback below performs full standard-library
	// chain/name/EKU/time verification after admission. This is not configurable.
	config.InsecureSkipVerify = true // #nosec G402 -- complete x509 verification in the mandatory callback.
	config.VerifyConnection = func(state tls.ConnectionState) error {
		return verifyTLSPeer(state, roots, options.ServerName)
	}
	return config, nil
}

// TLSServer validates immutable certificate/key bytes and their public origin
// name before a listener is opened. Replacement requires a new configuration
// and an owner-controlled restart; no file is reopened during a handshake.
func TLSServer(serverName, certificatePath, privateKeyPath string) (*tls.Config, error) {
	if !validTLSServerName(serverName) {
		return nil, ErrTLSConfiguration
	}
	certificate, err := loadTLSCertificate(certificatePath, privateKeyPath, x509.ExtKeyUsageServerAuth)
	if err != nil || certificate.Leaf.VerifyHostname(serverName) != nil {
		return nil, ErrTLSConfiguration
	}
	config := tlsPolicy()
	config.Certificates = []tls.Certificate{certificate}
	config.VerifyConnection = func(state tls.ConnectionState) error {
		if err := verifyTLSProtocol(state); err != nil {
			return err
		}
		// A server must stop presenting an expired identity even if its
		// operator has not yet restarted it with the replacement.
		for _, raw := range certificate.Certificate {
			cert, err := x509.ParseCertificate(raw)
			if err != nil || !validTLSCertificate(cert, time.Now()) {
				return ErrTLSPeer
			}
		}
		return nil
	}
	return config, nil
}

func tlsPolicy() *tls.Config {
	return &tls.Config{
		MinVersion:             tls.VersionTLS13,
		MaxVersion:             tls.VersionTLS13,
		CurvePreferences:       []tls.CurveID{tls.CurveP256, tls.CurveP384},
		SessionTicketsDisabled: true,
		// TLS 1.3 suites are controlled by the admitted Go FIPS module,
		// not Config.CipherSuites (which applies only to older TLS).
	}
}

func verifyTLSProtocol(state tls.ConnectionState) error {
	if state.Version != tls.VersionTLS13 || (state.CipherSuite != tls.TLS_AES_128_GCM_SHA256 && state.CipherSuite != tls.TLS_AES_256_GCM_SHA384) {
		return ErrTLSPeer
	}
	return nil
}

func verifyTLSPeer(state tls.ConnectionState, roots *x509.CertPool, serverName string) error {
	if verifyTLSProtocol(state) != nil || len(state.PeerCertificates) == 0 || roots == nil || !validTLSServerName(serverName) {
		return ErrTLSPeer
	}
	expectedSNI := serverName
	if net.ParseIP(serverName) != nil {
		expectedSNI = ""
	}
	if state.ServerName != expectedSNI {
		return ErrTLSPeer
	}
	leaf := state.PeerCertificates[0]
	if leaf.IsCA || leaf.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return ErrTLSPeer
	}
	now := time.Now()
	intermediates := x509.NewCertPool()
	for index, certificate := range state.PeerCertificates {
		if !validTLSCertificate(certificate, now) {
			return ErrTLSPeer
		}
		if index > 0 {
			intermediates.AddCert(certificate)
		}
	}
	// Only admitted peer keys and admitted immutable trust anchors reach X.509
	// signature checks. Verify retains all standard path, name, validity,
	// constraint and server-authentication EKU rules.
	_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{
		Roots: roots, Intermediates: intermediates, DNSName: serverName,
		CurrentTime: now, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return ErrTLSPeer
	}
	return nil
}

func loadTLSCertificate(certificatePath, keyPath string, usage x509.ExtKeyUsage) (tls.Certificate, error) {
	certificatePEM, err := readTLSPEM(certificatePath)
	if err != nil {
		return tls.Certificate{}, err
	}
	certificates, err := parseTLSCertificates(certificatePEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM, err := readTLSPEM(keyPath)
	if err != nil {
		return tls.Certificate{}, err
	}
	block, rest := pem.Decode(bytes.TrimSpace(keyPEM))
	if block == nil || len(block.Headers) != 0 || len(bytes.TrimSpace(rest)) != 0 || !bytes.HasPrefix(bytes.TrimSpace(keyPEM), []byte("-----BEGIN ")) {
		return tls.Certificate{}, ErrTLSConfiguration
	}
	switch block.Type {
	case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
	default:
		return tls.Certificate{}, ErrTLSConfiguration
	}
	certificate, err := tls.X509KeyPair(certificatePEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, ErrTLSConfiguration
	}
	leaf := certificates[0]
	if leaf.IsCA || leaf.KeyUsage&x509.KeyUsageDigitalSignature == 0 || !tlsCertificateUsage(leaf, usage) {
		return tls.Certificate{}, ErrTLSConfiguration
	}
	for i := 0; i+1 < len(certificates); i++ {
		if certificates[i].CheckSignatureFrom(certificates[i+1]) != nil {
			return tls.Certificate{}, ErrTLSConfiguration
		}
	}
	certificate.Leaf = leaf
	return certificate, nil
}

func tlsCertificateUsage(certificate *x509.Certificate, required x509.ExtKeyUsage) bool {
	if len(certificate.ExtKeyUsage) == 0 && len(certificate.UnknownExtKeyUsage) == 0 {
		return true
	}
	for _, usage := range certificate.ExtKeyUsage {
		if usage == required || usage == x509.ExtKeyUsageAny {
			return true
		}
	}
	return false
}

func readTLSPEM(path string) ([]byte, error) {
	document, err := securefile.Read(path, TLSPEMMaximumBytes)
	if err != nil || document.Size() == 0 {
		return nil, ErrTLSConfiguration
	}
	return document.Bytes(), nil
}

func parseTLSCertificates(data []byte) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate
	for remaining := bytes.TrimSpace(data); len(remaining) != 0; {
		if !bytes.HasPrefix(remaining, []byte("-----BEGIN CERTIFICATE-----")) {
			return nil, ErrTLSConfiguration
		}
		block, rest := pem.Decode(remaining)
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return nil, ErrTLSConfiguration
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil || !validTLSCertificate(certificate, time.Now()) {
			return nil, ErrTLSConfiguration
		}
		certificates = append(certificates, certificate)
		remaining = bytes.TrimSpace(rest)
	}
	if len(certificates) == 0 {
		return nil, ErrTLSConfiguration
	}
	return certificates, nil
}

func validTLSCertificate(certificate *x509.Certificate, now time.Time) bool {
	if certificate == nil || now.Before(certificate.NotBefore) || !now.Before(certificate.NotAfter) {
		return false
	}
	switch key := certificate.PublicKey.(type) {
	case *rsa.PublicKey:
		if ValidateRSAPublicKey(key) != nil {
			return false
		}
	case *ecdsa.PublicKey:
		if key.Curve != elliptic.P256() && key.Curve != elliptic.P384() {
			return false
		}
	default:
		return false
	}
	switch certificate.SignatureAlgorithm {
	case x509.SHA256WithRSA, x509.SHA384WithRSA, x509.SHA512WithRSA,
		x509.SHA256WithRSAPSS, x509.SHA384WithRSAPSS, x509.SHA512WithRSAPSS,
		x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512:
		return true
	default:
		return false
	}
}

func validTLSServerName(name string) bool {
	if name == "" || strings.TrimSpace(name) != name || strings.ContainsAny(name, "/\\@?#\x00 \t\r\n") {
		return false
	}
	return !strings.Contains(name, ":") || net.ParseIP(name) != nil
}

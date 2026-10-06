// Package tlstest issues short-lived, isolated fixture identities. It never
// installs a trust root in a host trust store or supplies a production identity.
package tlstest

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Authority struct {
	Certificate    *x509.Certificate
	CertificatePEM []byte
	key            *ecdsa.PrivateKey
}

type Identity struct {
	CertificatePEM []byte
	PrivateKeyPEM  []byte
}

func NewAuthority() (*Authority, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	template := &x509.Certificate{
		SerialNumber: serial, Subject: pkix.Name{CommonName: "Cartulary isolated fixture CA"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour),
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:     true, BasicConstraintsValid: true,
	}
	raw, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
	if err != nil {
		return nil, err
	}
	certificate, err := x509.ParseCertificate(raw)
	if err != nil {
		return nil, err
	}
	return &Authority{Certificate: certificate, CertificatePEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw}), key: key}, nil
}

func (authority *Authority) Issue(name string, hosts []string, usage x509.ExtKeyUsage) (Identity, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return Identity{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return Identity{}, err
	}
	template := &x509.Certificate{
		SerialNumber: serial, Subject: pkix.Name{CommonName: name},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(12 * time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage},
		BasicConstraintsValid: true,
	}
	for _, host := range hosts {
		if address := net.ParseIP(host); address != nil {
			template.IPAddresses = append(template.IPAddresses, address)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}
	raw, err := authority.Sign(template, key.Public())
	if err != nil {
		return Identity{}, err
	}
	private, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return Identity{}, err
	}
	return Identity{
		CertificatePEM: append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw}), authority.CertificatePEM...),
		PrivateKeyPEM:  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: private}),
	}, nil
}

// Sign supports adversarial certificate fixtures without performing excluded
// private-key operations (for example, a too-small RSA public key).
func (authority *Authority) Sign(template *x509.Certificate, public crypto.PublicKey) ([]byte, error) {
	return x509.CreateCertificate(rand.Reader, template, authority.Certificate, public, authority.key)
}

func WriteFile(directory, name string, data []byte) (string, error) {
	if name == "" || name != filepath.Base(name) || name == "." || name == ".." {
		return "", os.ErrInvalid
	}
	path := filepath.Join(directory, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	return path, closeErr
}

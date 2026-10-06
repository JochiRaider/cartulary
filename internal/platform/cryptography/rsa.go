package cryptography

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
)

var ErrRSAVerification = errors.New("RSA verification rejected")

// ValidateRSAPublicKey admits the selected module's public-key parameters
// before a protocol library can dispatch a signature operation.
func ValidateRSAPublicKey(key *rsa.PublicKey) error {
	if key == nil || key.N == nil || key.N.Sign() <= 0 || key.N.Bit(0) == 0 || key.N.BitLen() < 2048 || key.N.BitLen()%2 != 0 || key.E < 65537 || key.E > 1<<31-1 || key.E%2 == 0 {
		return ErrRSAVerification
	}
	return nil
}

func VerifyRSAPKCS1SHA256(key *rsa.PublicKey, payload, signature []byte) error {
	if err := ValidateRSAPublicKey(key); err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature) != nil {
		return ErrRSAVerification
	}
	return nil
}

// PS256 fixes the salt length to the SHA-256 output length; automatic salt
// detection would admit non-JOSE parameters and unapproved module services.
func VerifyRSAPSSSHA256(key *rsa.PublicKey, payload, signature []byte) error {
	if err := ValidateRSAPublicKey(key); err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	if rsa.VerifyPSS(key, crypto.SHA256, digest[:], signature, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}) != nil {
		return ErrRSAVerification
	}
	return nil
}

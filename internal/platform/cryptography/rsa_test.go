package cryptography

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"math/big"
	"testing"
)

func TestRSAParameterAndSignatureAdmission(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("independently signed RSA service admission payload")
	digest := sha256.Sum256(payload)
	pkcs, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	pss, err := rsa.SignPSS(rand.Reader, key, crypto.SHA256, digest[:], &rsa.PSSOptions{SaltLength: 32})
	if err != nil {
		t.Fatal(err)
	}
	if VerifyRSAPKCS1SHA256(&key.PublicKey, payload, pkcs) != nil || VerifyRSAPSSSHA256(&key.PublicKey, payload, pss) != nil {
		t.Fatal("valid standard-library signatures rejected")
	}
	if VerifyRSAPKCS1SHA256(&key.PublicKey, []byte("different"), pkcs) == nil || VerifyRSAPSSSHA256(&key.PublicKey, []byte("different"), pss) == nil {
		t.Fatal("substituted payload admitted")
	}
	if VerifyRSAPKCS1SHA256(&key.PublicKey, payload, pss) == nil || VerifyRSAPSSSHA256(&key.PublicKey, payload, pkcs) == nil {
		t.Fatal("signature method substitution admitted")
	}
	shortSalt, err := rsa.SignPSS(rand.Reader, key, crypto.SHA256, digest[:], &rsa.PSSOptions{SaltLength: 16})
	if err != nil {
		t.Fatal(err)
	}
	if VerifyRSAPSSSHA256(&key.PublicKey, payload, shortSalt) == nil {
		t.Fatal("automatic PSS salt accepted")
	}
	for _, name := range []string{"nil", "missing modulus", "negative modulus", "small", "odd bits", "even modulus", "small exponent", "even exponent", "large exponent"} {
		t.Run(name, func(t *testing.T) {
			public := &rsa.PublicKey{N: new(big.Int).Set(key.N), E: key.E}
			switch name {
			case "nil":
				public = nil
			case "missing modulus":
				public.N = nil
			case "negative modulus":
				public.N.Neg(public.N)
			case "small":
				public.N = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 1024), big.NewInt(1))
			case "odd bits":
				public.N = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 2049), big.NewInt(1))
			case "even modulus":
				public.N.SetBit(public.N, 0, 0)
			case "small exponent":
				public.E = 3
			case "even exponent":
				public.E = 65538
			case "large exponent":
				public.E = 1 << 31
			}
			if ValidateRSAPublicKey(public) == nil || VerifyRSAPKCS1SHA256(public, payload, pkcs) == nil || VerifyRSAPSSSHA256(public, payload, pss) == nil {
				t.Fatal("excluded RSA parameters admitted")
			}
		})
	}
}

package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// Seal creates version || fresh salt || module-generated nonce || ciphertext || tag.
// Neither callers nor tests can inject production entropy or choose a nonce.
func (key MasterKey) Seal(purpose string, binding []string, plaintext, aad []byte) ([]byte, error) {
	if len(plaintext) > MaxPlaintextBytes {
		return nil, ErrInput
	}
	info, err := encodeContext("cartulary.message.v1", purpose, binding, len(aad))
	if err != nil {
		return nil, err
	}
	if !key.admitted {
		return nil, ErrKey
	}
	header := make([]byte, 1+SaltBytes, len(plaintext)+EnvelopeOverhead)
	header[0] = EnvelopeVersion
	rand.Read(header[1:])
	aead, err := key.aead(header[1:], info)
	if err != nil {
		return nil, err
	}
	return aead.Seal(header, nil, plaintext, authenticatedContext(info, header, aad)), nil
}

func (key MasterKey) Open(purpose string, binding []string, envelope, aad []byte) ([]byte, error) {
	if len(envelope) < EnvelopeOverhead || len(envelope) > MaxPlaintextBytes+EnvelopeOverhead || envelope[0] != EnvelopeVersion {
		return nil, ErrOpen
	}
	info, err := encodeContext("cartulary.message.v1", purpose, binding, len(aad))
	if err != nil {
		return nil, ErrOpen
	}
	header := envelope[:1+SaltBytes]
	aead, err := key.aead(header[1:], info)
	if err != nil {
		return nil, ErrOpen
	}
	plaintext, err := aead.Open(nil, nil, envelope[len(header):], authenticatedContext(info, header, aad))
	if err != nil {
		return nil, ErrOpen
	}
	return plaintext, nil
}

func (key MasterKey) aead(salt, info []byte) (cipher.AEAD, error) {
	derived, err := key.derive(salt, info)
	if err != nil {
		return nil, err
	}
	defer clear(derived[:])
	block, err := aes.NewCipher(derived[:])
	if err != nil {
		return nil, ErrKey
	}
	return cipher.NewGCMWithRandomNonce(block)
}

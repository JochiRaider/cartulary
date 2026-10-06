package cryptography

import (
	"crypto/cipher"
	"crypto/rand"
	"sync"
)

// ArtifactSealer is a fresh, non-resumable writing capability. Copies of the
// interface share one private usage counter. There is no constructor from salt.
type ArtifactSealer interface {
	Salt() [SaltBytes]byte
	Seal(plaintext, aad []byte) ([]byte, error)
}

// ArtifactOpener cannot seal. Recovery owns chunk ordering, finality, schemas,
// immutable input admission and publication; this capability owns only mechanics.
type ArtifactOpener interface {
	Open(ciphertext, aad []byte) ([]byte, error)
}

type artifactCipher struct {
	aead         cipher.AEAD
	info         []byte
	salt         [SaltBytes]byte
	contextBytes int
}

type artifactSealer struct {
	mu    sync.Mutex
	used  uint32
	crypt artifactCipher
}

type artifactOpener struct{ crypt artifactCipher }

func (key MasterKey) NewArtifactSealer(purpose string, binding ...string) (ArtifactSealer, error) {
	var salt [SaltBytes]byte
	rand.Read(salt[:])
	crypt, err := key.artifactCipher(purpose, binding, salt)
	if err != nil {
		return nil, err
	}
	return &artifactSealer{crypt: crypt}, nil
}

func (key MasterKey) ArtifactOpener(purpose string, binding []string, salt [SaltBytes]byte) (ArtifactOpener, error) {
	crypt, err := key.artifactCipher(purpose, binding, salt)
	if err != nil {
		return nil, ErrOpen
	}
	return &artifactOpener{crypt: crypt}, nil
}

func (key MasterKey) artifactCipher(purpose string, binding []string, salt [SaltBytes]byte) (artifactCipher, error) {
	info, err := encodeContext("cartulary.artifact.v1", purpose, binding, 0)
	if err != nil {
		return artifactCipher{}, err
	}
	aead, err := key.aead(salt[:], info)
	if err != nil {
		return artifactCipher{}, err
	}
	contextBytes := len(purpose)
	for _, component := range binding {
		contextBytes += len(component)
	}
	return artifactCipher{aead: aead, info: info, salt: salt, contextBytes: contextBytes}, nil
}

func (sealer *artifactSealer) Salt() [SaltBytes]byte { return sealer.crypt.salt }

func (sealer *artifactSealer) Seal(plaintext, aad []byte) ([]byte, error) {
	if len(plaintext) > MaxPlaintextBytes || len(aad) > MaxContextBytes-sealer.crypt.contextBytes {
		return nil, ErrInput
	}
	sealer.mu.Lock()
	defer sealer.mu.Unlock()
	if sealer.used >= MaxArtifactSeals {
		return nil, ErrArtifactLimit
	}
	sealer.used++
	return sealer.crypt.aead.Seal(nil, nil, plaintext, authenticatedContext(sealer.crypt.info, sealer.crypt.salt[:], aad)), nil
}

func (opener *artifactOpener) Open(ciphertext, aad []byte) ([]byte, error) {
	if len(ciphertext) < opener.crypt.aead.Overhead() || len(ciphertext) > MaxPlaintextBytes+opener.crypt.aead.Overhead() || len(aad) > MaxContextBytes-opener.crypt.contextBytes {
		return nil, ErrOpen
	}
	plaintext, err := opener.crypt.aead.Open(nil, nil, ciphertext, authenticatedContext(opener.crypt.info, opener.crypt.salt[:], aad))
	if err != nil {
		return nil, ErrOpen
	}
	return plaintext, nil
}

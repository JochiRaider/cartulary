package cryptography

import (
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

const (
	MasterKeyBytes         = 32
	SaltBytes              = 32
	MaxPlaintextBytes      = 16 << 20
	MaxContextBytes        = 64 << 10
	MaxArtifactSeals       = 1 << 20
	EnvelopeVersion   byte = 1
	EnvelopeOverhead       = 1 + SaltBytes + 12 + 16
)

var (
	ErrKey           = errors.New("cryptographic key rejected")
	ErrInput         = errors.New("cryptographic input rejected")
	ErrOpen          = errors.New("sealed value rejected")
	ErrArtifactLimit = errors.New("artifact key use limit reached")
)

// MasterKey is immutable and cannot be used until exact-length admission.
// Rotation and key identifiers belong to the owner that holds this capability.
type MasterKey struct {
	material [MasterKeyBytes]byte
	admitted bool
}

func AdmitKey(material []byte) (MasterKey, error) {
	if len(material) != MasterKeyBytes {
		return MasterKey{}, ErrKey
	}
	var key MasterKey
	copy(key.material[:], material)
	key.admitted = true
	return key, nil
}

// Derive provides deterministic purpose keys for owner-local HMAC protocols.
// Sealing uses different constructions and fresh salts instead of these keys.
func (key MasterKey) Derive(purpose string, binding ...string) ([MasterKeyBytes]byte, error) {
	info, err := encodeContext("cartulary.derivation.v1", purpose, binding, 0)
	if err != nil {
		return [MasterKeyBytes]byte{}, err
	}
	return key.derive(nil, info)
}

func (key MasterKey) derive(salt, info []byte) ([MasterKeyBytes]byte, error) {
	if !key.admitted {
		return [MasterKeyBytes]byte{}, ErrKey
	}
	derived, err := hkdf.Key(sha256.New, key.material[:], salt, string(info), MasterKeyBytes)
	if err != nil {
		return [MasterKeyBytes]byte{}, ErrKey
	}
	var result [MasterKeyBytes]byte
	copy(result[:], derived)
	clear(derived)
	return result, nil
}

// Each field has a uint32 big-endian length. A separate component count makes
// empty and omitted bindings distinct. Bound both content and framing work.
func encodeContext(construction, purpose string, binding []string, aadBytes int) ([]byte, error) {
	if purpose == "" || len(purpose) > MaxContextBytes || aadBytes < 0 || aadBytes > MaxContextBytes || len(binding) > MaxContextBytes/4 {
		return nil, ErrInput
	}
	remaining := MaxContextBytes - len(purpose) - aadBytes
	if remaining < 0 {
		return nil, ErrInput
	}
	for _, component := range binding {
		if len(component) > remaining {
			return nil, ErrInput
		}
		remaining -= len(component)
	}
	info := appendField(nil, []byte(construction))
	info = appendField(info, []byte(purpose))
	info = binary.BigEndian.AppendUint32(info, uint32(len(binding)))
	for _, component := range binding {
		info = appendField(info, []byte(component))
	}
	return info, nil
}

func appendField(dst, field []byte) []byte {
	dst = binary.BigEndian.AppendUint32(dst, uint32(len(field)))
	return append(dst, field...)
}

func authenticatedContext(info, header, aad []byte) []byte {
	data := appendField(nil, info)
	data = appendField(data, header)
	return appendField(data, aad)
}

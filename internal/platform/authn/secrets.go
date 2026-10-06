package authn

import (
	"errors"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

// SecretBinding names an Auth-owned record and its owning user or provider.
// Pending enrollment and active TOTP records deliberately use different purposes.
type SecretBinding struct {
	Purpose   SecretPurpose
	RecordID  uuid.UUID
	SubjectID uuid.UUID
}

type SecretPurpose string

const (
	ActiveTOTPSecret     SecretPurpose = "totp.active"
	PendingTOTPSecret    SecretPurpose = "totp.pending"
	EnterprisePKCESecret SecretPurpose = "enterprise.pkce"
)
const secretEnvelopePurpose = "cartulary.auth.secret.v1"

var ErrSecret = errors.New("authentication secret rejected")

func (binding SecretBinding) context() ([]string, bool) {
	if binding.RecordID == uuid.Nil || binding.SubjectID == uuid.Nil {
		return nil, false
	}
	switch binding.Purpose {
	case ActiveTOTPSecret:
		if binding.RecordID != binding.SubjectID {
			return nil, false
		}
	case PendingTOTPSecret, EnterprisePKCESecret:
	default:
		return nil, false
	}
	return []string{string(binding.Purpose), binding.RecordID.String(), binding.SubjectID.String()}, true
}

func SealSecret(keys MasterKeys, binding SecretBinding, plaintext []byte) ([]byte, error) {
	context, ok := binding.context()
	if !ok || len(plaintext) != binding.plaintextBytes() {
		return nil, ErrSecret
	}
	envelope, err := keys.master.Seal(secretEnvelopePurpose, context, plaintext, nil)
	if err != nil {
		return nil, ErrSecret
	}
	return envelope, nil
}

func OpenSecret(keys MasterKeys, binding SecretBinding, envelope []byte) ([]byte, error) {
	context, ok := binding.context()
	if !ok || len(envelope) != binding.plaintextBytes()+cryptography.EnvelopeOverhead {
		return nil, ErrSecret
	}
	plaintext, err := keys.master.Open(secretEnvelopePurpose, context, envelope, nil)
	if err != nil || len(plaintext) != binding.plaintextBytes() {
		return nil, ErrSecret
	}
	return plaintext, nil
}

func (binding SecretBinding) plaintextBytes() int {
	if binding.Purpose == EnterprisePKCESecret {
		return 43
	}
	return 32
}

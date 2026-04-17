package authn

import (
	"golang.org/x/crypto/argon2"

	"github.com/pquerna/otp/totp"
)

// DerivePasswordHash is a bootstrap-only placeholder for future auth wiring.
func DerivePasswordHash(password []byte, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}

// ValidateTOTP is a bootstrap-only placeholder for future MFA wiring.
func ValidateTOTP(passcode string, secret string) bool {
	return totp.Validate(passcode, secret)
}

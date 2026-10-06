package authn

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
)

const (
	AuthMasterKeyEnv           = "CARTULARY_AUTH_MASTER_KEY"
	SessionCookieName          = "cartulary_session"
	CSRFCookieName             = "cartulary_csrf"
	CSRFHeaderName             = "X-CSRF-Token"
	SessionIdleTTL             = 30 * time.Minute
	SessionAbsoluteTTL         = 12 * time.Hour
	BootstrapTokenTTL          = 10 * time.Minute
	PendingTOTPEnrollmentTTL   = 10 * time.Minute
	ConcurrencyLimitReasonCode = "concurrency_limit"
	minPasswordScalars         = 12
	maxPasswordScalars         = 1024
	maxEmailScalars            = 320
	displayNameMaxScalars      = 256
	reasonNoteMaxScalars       = 4096
)

var base32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

type SessionTiming struct {
	AuthenticatedAt          time.Time
	LastQualifyingActivityAt time.Time
	IdleExpiresAt            time.Time
	AbsoluteExpiresAt        time.Time
	SessionExpiresAt         time.Time
}

type SessionSummary struct {
	SessionID                uuid.UUID
	LastQualifyingActivityAt time.Time
	AuthenticatedAt          time.Time
}

type RevocationScope int

const (
	RevokeCurrentSessionOnly RevocationScope = iota + 1
	RevokeAllUserSessions
)

type RevocationAction int

const (
	RevocationActionLogout RevocationAction = iota + 1
	RevocationActionPasswordChange
	RevocationActionTOTPReplacement
	RevocationActionAdminPasswordReset
	RevocationActionAdminTOTPReset
	RevocationActionAccountDisablement
	RevocationActionExplicitRevokeAll
)

type MasterKeys struct {
	tokenFingerprintKey   [32]byte
	csrfKey               [32]byte
	master                cryptography.MasterKey
	requestFingerprintKey [32]byte
}

type BootstrapTokenStatus int

const (
	BootstrapTokenValid BootstrapTokenStatus = iota + 1
	BootstrapTokenExpired
	BootstrapTokenConsumed
	BootstrapTokenSuperseded
)

type PendingEnrollmentStatus int

const (
	PendingEnrollmentValid PendingEnrollmentStatus = iota + 1
	PendingEnrollmentNotFound
	PendingEnrollmentExpired
	PendingEnrollmentConsumed
)

func ValidatePasswordProvision(value string) (string, error) {
	if value == "" {
		return "", errors.New("password must not be empty")
	}
	if len(value) > maxPasswordScalars*utf8.UTFMax {
		return "", fmt.Errorf("password length must be between %d and %d Unicode scalar values", minPasswordScalars, maxPasswordScalars)
	}

	scalars := 0
	allWhitespace := true
	for _, r := range value {
		scalars++
		if unicode.Is(unicode.Cc, r) {
			return "", errors.New("password must not contain control characters")
		}
		if !unicode.IsSpace(r) {
			allWhitespace = false
		}
	}
	if !utf8.ValidString(value) {
		return "", errors.New("password must be valid utf-8")
	}
	if allWhitespace {
		return "", errors.New("password must not be all whitespace")
	}
	if scalars < minPasswordScalars || scalars > maxPasswordScalars {
		return "", fmt.Errorf("password length must be between %d and %d Unicode scalar values", minPasswordScalars, maxPasswordScalars)
	}
	return value, nil
}

func NormalizeEmailAddress(raw string) (string, string, bool) {
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	if normalized == "" || utf8.RuneCountInString(normalized) > maxEmailScalars {
		return "", "", false
	}
	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return "", "", false
		}
		if unicode.IsSpace(r) {
			return "", "", false
		}
	}
	parts := strings.Split(normalized, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return normalized, strings.ToLower(normalized), true
}

func NormalizeDisplayNameLine(raw string) (string, bool) {
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	if normalized == "" || utf8.RuneCountInString(normalized) > displayNameMaxScalars {
		return "", false
	}
	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return "", false
		}
	}
	return normalized, true
}

func NormalizeReasonNote(raw *string) *string {
	if raw == nil {
		return nil
	}
	normalized := norm.NFC.String(*raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	if normalized == "" || utf8.RuneCountInString(normalized) > reasonNoteMaxScalars {
		return nil
	}
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r):
			return nil
		}
	}
	return &normalized
}

func NewSessionTiming(now time.Time) SessionTiming {
	timing := SessionTiming{
		AuthenticatedAt:          now.UTC(),
		LastQualifyingActivityAt: now.UTC(),
	}
	timing.IdleExpiresAt = timing.LastQualifyingActivityAt.Add(SessionIdleTTL)
	timing.AbsoluteExpiresAt = timing.AuthenticatedAt.Add(SessionAbsoluteTTL)
	timing.SessionExpiresAt = earlierTime(timing.IdleExpiresAt, timing.AbsoluteExpiresAt)
	return timing
}

func (s SessionTiming) Slide(now time.Time) SessionTiming {
	now = now.UTC()
	last := s.LastQualifyingActivityAt.UTC()
	if now.Before(last) {
		now = last
	}
	s.LastQualifyingActivityAt = now
	s.IdleExpiresAt = s.LastQualifyingActivityAt.Add(SessionIdleTTL)
	s.SessionExpiresAt = earlierTime(s.IdleExpiresAt, s.AbsoluteExpiresAt)
	return s
}

func SelectSessionForConcurrencyLimit(active []SessionSummary, currentSessionID uuid.UUID) (SessionSummary, bool) {
	candidates := make([]SessionSummary, 0, len(active))
	for _, current := range active {
		if current.SessionID == currentSessionID {
			continue
		}
		candidates = append(candidates, current)
	}
	if len(candidates) == 0 {
		return SessionSummary{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		if !candidates[i].LastQualifyingActivityAt.Equal(candidates[j].LastQualifyingActivityAt) {
			return candidates[i].LastQualifyingActivityAt.Before(candidates[j].LastQualifyingActivityAt)
		}
		if !candidates[i].AuthenticatedAt.Equal(candidates[j].AuthenticatedAt) {
			return candidates[i].AuthenticatedAt.Before(candidates[j].AuthenticatedAt)
		}
		return candidates[i].SessionID.String() < candidates[j].SessionID.String()
	})
	return candidates[0], true
}

func RevocationScopeForAction(action RevocationAction) RevocationScope {
	switch action {
	case RevocationActionLogout:
		return RevokeCurrentSessionOnly
	default:
		return RevokeAllUserSessions
	}
}

func GenerateOpaqueToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate opaque token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func masterKeyMaterial(env map[string]string) ([]byte, error) {
	raw, ok := lookupEnv(env, AuthMasterKeyEnv)
	if !ok || raw == "" {
		return nil, fmt.Errorf("%s is required", AuthMasterKeyEnv)
	}
	decoded, err := base64.RawStdEncoding.Strict().DecodeString(raw)
	if err != nil {
		decoded, err = base64.StdEncoding.Strict().DecodeString(raw)
	}
	if err != nil || len(decoded) != cryptography.MasterKeyBytes || (base64.RawStdEncoding.EncodeToString(decoded) != raw && base64.StdEncoding.EncodeToString(decoded) != raw) {
		return nil, fmt.Errorf("%s must encode exactly 32 bytes", AuthMasterKeyEnv)
	}
	return decoded, nil
}

func LoadMasterKeys(env map[string]string) (MasterKeys, error) {
	material, err := masterKeyMaterial(env)
	if err != nil {
		return MasterKeys{}, err
	}
	defer clear(material)
	master, err := cryptography.AdmitKey(material)
	if err != nil {
		return MasterKeys{}, err
	}
	keys := MasterKeys{master: master}
	for purpose, destination := range map[string]*[32]byte{
		"token-fingerprint":   &keys.tokenFingerprintKey,
		"csrf":                &keys.csrfKey,
		"request-fingerprint": &keys.requestFingerprintKey,
	} {
		*destination, err = master.Derive("cartulary.auth.key.v1", purpose)
		if err != nil {
			return MasterKeys{}, err
		}
	}
	return keys, nil
}

func RegisterMasterSecretPurpose(registry *secretpurpose.Registry, env map[string]string) error {
	material, err := masterKeyMaterial(env)
	if err != nil {
		return errors.New("register authentication master secret: invalid material")
	}
	defer clear(material)
	if err := registry.Register(AuthMasterKeyEnv, "authentication_master", material); err != nil {
		return fmt.Errorf("register authentication master secret purpose: %w", err)
	}
	return nil
}

func FingerprintToken(keys MasterKeys, token string) []byte {
	return hmacSHA256(keys.tokenFingerprintKey[:], token)
}

func FingerprintRequestValue(keys MasterKeys, value string) []byte {
	return hmacSHA256(keys.requestFingerprintKey[:], value)
}

func DerivePurposeKey(keys MasterKeys, purpose string) ([32]byte, error) {
	if purpose == "" {
		return [32]byte{}, cryptography.ErrInput
	}
	return keys.master.Derive("cartulary.auth.consumer-key.v1", purpose)
}

func CSRFTokenForSessionToken(keys MasterKeys, sessionToken string) string {
	return base64.RawURLEncoding.EncodeToString(hmacSHA256(keys.csrfKey[:], sessionToken))
}

func GenerateTOTPSecret() ([]byte, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, "", fmt.Errorf("generate totp secret: %w", err)
	}
	return raw, EncodeSecretBase32(raw), nil
}

func EncodeSecretBase32(secret []byte) string {
	return strings.ToUpper(base32NoPadding.EncodeToString(secret))
}

func ValidateTOTPCode(secretBase32 string, passcode string, now time.Time) bool {
	secret, err := base32NoPadding.DecodeString(secretBase32)
	if err != nil || len(secret) != 32 || EncodeSecretBase32(secret) != secretBase32 || len(passcode) != 6 {
		return false
	}
	ok, err := totp.ValidateCustom(passcode, secretBase32, now, totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA256,
	})
	return err == nil && ok
}

func ValidateTOTP(passcode string, secret string) bool {
	return ValidateTOTPCode(secret, passcode, time.Now())
}

func BootstrapStatus(record BootstrapTokenRecord, now time.Time) BootstrapTokenStatus {
	switch {
	case record.ConsumedAt != nil:
		return BootstrapTokenConsumed
	case record.SupersededAt != nil:
		return BootstrapTokenSuperseded
	case !record.ExpiresAt.After(now.UTC()):
		return BootstrapTokenExpired
	default:
		return BootstrapTokenValid
	}
}

func BootstrapReasonCode(record BootstrapTokenRecord, now time.Time) string {
	switch BootstrapStatus(record, now) {
	case BootstrapTokenExpired:
		return "expired"
	case BootstrapTokenConsumed:
		return "consumed"
	case BootstrapTokenSuperseded:
		return "superseded"
	default:
		return ""
	}
}

func PendingEnrollmentStatusAt(record *PendingTOTPEnrollmentRecord, now time.Time) PendingEnrollmentStatus {
	if record == nil {
		return PendingEnrollmentNotFound
	}
	switch {
	case record.ConsumedAt != nil:
		return PendingEnrollmentConsumed
	case !record.ExpiresAt.After(now.UTC()):
		return PendingEnrollmentExpired
	default:
		return PendingEnrollmentValid
	}
}

func PendingEnrollmentReasonCode(record *PendingTOTPEnrollmentRecord, now time.Time) string {
	switch PendingEnrollmentStatusAt(record, now) {
	case PendingEnrollmentNotFound:
		return "not_found"
	case PendingEnrollmentExpired:
		return "expired"
	case PendingEnrollmentConsumed:
		return "consumed"
	default:
		return ""
	}
}

func earlierTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}
	return os.LookupEnv(key)
}

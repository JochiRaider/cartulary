package evidence

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	objectUploadTokenPrefix  = "upl_"
	objectUploadTokenPurpose = "evidence-object-upload-v2"
	objectUploadTokenVersion = 2
)

var errInvalidObjectUploadToken = errors.New("invalid object upload token")

type objectUploadTokenClaims struct {
	Version                int       `json:"v"`
	LeaseID                uuid.UUID `json:"lease_id"`
	ObjectBlobID           uuid.UUID `json:"object_blob_id"`
	IncidentID             uuid.UUID `json:"incident_id"`
	IssuingUserID          uuid.UUID `json:"issuing_user_id"`
	IssuingSessionID       uuid.UUID `json:"issuing_session_id"`
	StorageKey             string    `json:"storage_key"`
	ByteSize               int64     `json:"byte_size"`
	ExpectedSHA256Hex      string    `json:"expected_sha256_hex,omitempty"`
	RequiredMethod         string    `json:"required_method"`
	RequiredHeadersSHA256  string    `json:"required_headers_sha256"`
	AcceptedContractSHA256 string    `json:"accepted_contract_sha256"`
	IssuedAtUnixNano       int64     `json:"issued_at_unix_nano"`
	ExpiresAtUnixNano      int64     `json:"expires_at_unix_nano"`
}

func encodeObjectUploadToken(keys authn.MasterKeys, claims objectUploadTokenClaims) (string, error) {
	if claims.Version == 0 {
		claims.Version = objectUploadTokenVersion
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal object upload token claims: %w", err)
	}
	payloadSegment := base64.RawURLEncoding.EncodeToString(payload)
	signature := signObjectUploadToken(keys, payloadSegment)
	return objectUploadTokenPrefix + payloadSegment + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func decodeObjectUploadToken(keys authn.MasterKeys, token string) (objectUploadTokenClaims, error) {
	token = strings.TrimSpace(token)
	if !strings.HasPrefix(token, objectUploadTokenPrefix) {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	token = strings.TrimPrefix(token, objectUploadTokenPrefix)
	payloadSegment, signatureSegment, ok := strings.Cut(token, ".")
	if !ok || payloadSegment == "" || signatureSegment == "" {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(signatureSegment)
	if err != nil {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	if !hmac.Equal(signature, signObjectUploadToken(keys, payloadSegment)) {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadSegment)
	if err != nil {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	var claims objectUploadTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	if claims.Version != objectUploadTokenVersion ||
		claims.LeaseID == uuid.Nil || claims.ObjectBlobID == uuid.Nil ||
		claims.IncidentID == uuid.Nil || claims.IssuingUserID == uuid.Nil ||
		claims.IssuingSessionID == uuid.Nil || claims.StorageKey == "" ||
		claims.ByteSize < 0 || claims.RequiredMethod != "PUT" ||
		len(claims.RequiredHeadersSHA256) != 64 || len(claims.AcceptedContractSHA256) != 64 ||
		claims.IssuedAtUnixNano <= 0 || claims.ExpiresAtUnixNano <= claims.IssuedAtUnixNano {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	return claims, nil
}

func objectUploadTokenDigest(token string) []byte {
	digest := sha256.Sum256([]byte(token))
	return append([]byte(nil), digest[:]...)
}

func objectUploadBindingDigest(value any) ([]byte, string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, "", fmt.Errorf("marshal object upload binding: %w", err)
	}
	digest := sha256.Sum256(payload)
	return append([]byte(nil), digest[:]...), fmt.Sprintf("%x", digest[:]), nil
}

func signObjectUploadToken(keys authn.MasterKeys, payloadSegment string) []byte {
	key := authn.DerivePurposeKey(keys, objectUploadTokenPurpose)
	mac := hmac.New(sha256.New, key[:])
	_, _ = mac.Write([]byte(payloadSegment))
	return mac.Sum(nil)
}

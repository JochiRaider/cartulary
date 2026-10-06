package evidence

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	objectUploadTokenPrefix       = "upl_"
	objectUploadTokenPurpose      = "evidence-object-upload-v3"
	objectUploadTokenVersion      = 3
	objectUploadTokenMaximumBytes = 4096
)

var errInvalidObjectUploadToken = errors.New("invalid object upload token")

type objectUploadTokenClaims struct {
	Version                int       `json:"v"`
	LeaseID                uuid.UUID `json:"lease_id"`
	ObjectBlobID           uuid.UUID `json:"object_blob_id"`
	IncidentID             uuid.UUID `json:"incident_id"`
	IssuingUserID          uuid.UUID `json:"issuing_user_id"`
	IssuingSessionID       uuid.UUID `json:"issuing_session_id"`
	StorageKeyBinding      string    `json:"storage_key_binding"`
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
	signature, err := signObjectUploadToken(keys, payloadSegment)
	if err != nil {
		return "", errInvalidObjectUploadToken
	}
	token := objectUploadTokenPrefix + payloadSegment + "." + base64.RawURLEncoding.EncodeToString(signature)
	if len(token) > objectUploadTokenMaximumBytes {
		return "", errInvalidObjectUploadToken
	}
	return token, nil
}

func decodeObjectUploadToken(keys authn.MasterKeys, token string) (objectUploadTokenClaims, error) {
	if len(token) > objectUploadTokenMaximumBytes || !strings.HasPrefix(token, objectUploadTokenPrefix) {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	token = strings.TrimPrefix(token, objectUploadTokenPrefix)
	payloadSegment, signatureSegment, ok := strings.Cut(token, ".")
	if !ok || payloadSegment == "" || signatureSegment == "" {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	signature, err := base64.RawURLEncoding.Strict().DecodeString(signatureSegment)
	if err != nil || len(signature) != sha256.Size || base64.RawURLEncoding.EncodeToString(signature) != signatureSegment {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	expected, err := signObjectUploadToken(keys, payloadSegment)
	if err != nil || !hmac.Equal(signature, expected) {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	payload, err := base64.RawURLEncoding.Strict().DecodeString(payloadSegment)
	if err != nil || base64.RawURLEncoding.EncodeToString(payload) != payloadSegment {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	var claims objectUploadTokenClaims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	if trailing, err := decoder.Token(); err != io.EOF || trailing != nil {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	binding, err := base64.RawURLEncoding.Strict().DecodeString(claims.StorageKeyBinding)
	if err != nil || len(binding) != sha256.Size || base64.RawURLEncoding.EncodeToString(binding) != claims.StorageKeyBinding {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	if claims.Version != objectUploadTokenVersion ||
		claims.LeaseID == uuid.Nil || claims.ObjectBlobID == uuid.Nil ||
		claims.IncidentID == uuid.Nil || claims.IssuingUserID == uuid.Nil ||
		claims.IssuingSessionID == uuid.Nil || claims.StorageKeyBinding == "" ||
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

func signObjectUploadToken(keys authn.MasterKeys, payloadSegment string) ([]byte, error) {
	key, err := authn.DerivePurposeKey(keys, objectUploadTokenPurpose)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, key[:])
	_, _ = mac.Write([]byte(payloadSegment))
	return mac.Sum(nil), nil
}

func objectUploadStorageBinding(keys authn.MasterKeys, storageKey string) (string, error) {
	key, err := authn.DerivePurposeKey(keys, "evidence-object-storage-binding-v1")
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key[:])
	_, _ = mac.Write([]byte(storageKey))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

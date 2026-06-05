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
	objectUploadTokenPurpose = "evidence-object-upload-v1"
)

var errInvalidObjectUploadToken = errors.New("invalid object upload token")

type objectUploadTokenClaims struct {
	Version           int       `json:"v"`
	ObjectBlobID      uuid.UUID `json:"object_blob_id"`
	IncidentID        uuid.UUID `json:"incident_id"`
	StorageKey        string    `json:"storage_key"`
	ByteSize          int64     `json:"byte_size"`
	ExpiresAtUnixNano int64     `json:"expires_at_unix_nano"`
}

func encodeObjectUploadToken(keys authn.MasterKeys, claims objectUploadTokenClaims) (string, error) {
	if claims.Version == 0 {
		claims.Version = 1
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
	if claims.Version != 1 || claims.ObjectBlobID == uuid.Nil || claims.IncidentID == uuid.Nil || claims.StorageKey == "" || claims.ByteSize < 0 || claims.ExpiresAtUnixNano <= 0 {
		return objectUploadTokenClaims{}, errInvalidObjectUploadToken
	}
	return claims, nil
}

func signObjectUploadToken(keys authn.MasterKeys, payloadSegment string) []byte {
	key := authn.DerivePurposeKey(keys, objectUploadTokenPurpose)
	mac := hmac.New(sha256.New, key[:])
	_, _ = mac.Write([]byte(payloadSegment))
	return mac.Sum(nil)
}

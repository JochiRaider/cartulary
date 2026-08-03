package conflicts

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ConflictTokenVersion    = 3
	ConflictTokenTTL        = 30 * time.Minute
	ConflictTokenClockSkew  = 60 * time.Second
	ConflictTokenMaximumLen = 4096
	conflictTokenPrefix     = "cft3."
	conflictTokenPurpose    = "cartulary.conflict-token.v3"
)

var ErrConflictTokenUnavailable = errors.New("conflict token unavailable")

type ConflictTokenClaims struct {
	Version                 int64     `json:"version"`
	RouteKey                string    `json:"route_key"`
	RecordID                string    `json:"record_id"`
	ViewSchemaID            string    `json:"view_schema_id"`
	FieldKey                string    `json:"field_key"`
	ConflictResolutionClass string    `json:"conflict_resolution_class"`
	BaseRowVersion          int64     `json:"base_row_version"`
	CurrentRowVersion       int64     `json:"current_row_version"`
	RequestHash             string    `json:"request_hash"`
	IssuedAt                time.Time `json:"issued_at"`
	ExpiresAt               time.Time `json:"expires_at"`
}

type ConflictTokenBinding struct {
	RouteKey                string
	RecordID                string
	ViewSchemaID            string
	FieldKey                string
	ConflictResolutionClass string
	BaseRowVersion          int64
	CurrentRowVersion       int64
	RequestHash             string
}

func (claims ConflictTokenClaims) ValidFor(binding ConflictTokenBinding) bool {
	return claims.RouteKey == binding.RouteKey &&
		claims.RecordID == binding.RecordID &&
		claims.ViewSchemaID == binding.ViewSchemaID &&
		claims.FieldKey == binding.FieldKey &&
		claims.ConflictResolutionClass == binding.ConflictResolutionClass &&
		claims.BaseRowVersion == binding.BaseRowVersion &&
		claims.CurrentRowVersion == binding.CurrentRowVersion &&
		claims.RequestHash == binding.RequestHash
}

type conflictTokenCipher struct {
	aead          cipher.AEAD
	state         string
	deactivatedAt *time.Time
	retireAt      *time.Time
}

type ConflictTokenCodec struct {
	activeKeyID string
	keys        map[string]conflictTokenCipher
	now         func() time.Time
	entropy     io.Reader
}

type CodecOption func(*ConflictTokenCodec)

func WithClock(now func() time.Time) CodecOption {
	return func(codec *ConflictTokenCodec) {
		codec.now = now
	}
}

func WithEntropySource(entropy io.Reader) CodecOption {
	return func(codec *ConflictTokenCodec) {
		codec.entropy = entropy
	}
}

func NewConflictTokenCodec(ring *ConflictTokenKeyRing, options ...CodecOption) (ConflictTokenCodec, error) {
	if ring == nil || ring.activeKeyID == "" || len(ring.keys) == 0 {
		return ConflictTokenCodec{}, ErrConflictTokenUnavailable
	}
	codec := ConflictTokenCodec{
		activeKeyID: ring.activeKeyID,
		keys:        make(map[string]conflictTokenCipher, len(ring.keys)),
		now:         func() time.Time { return time.Now().UTC() },
		entropy:     rand.Reader,
	}
	for _, option := range options {
		if option != nil {
			option(&codec)
		}
	}
	if codec.now == nil || codec.entropy == nil {
		return ConflictTokenCodec{}, ErrConflictTokenUnavailable
	}
	for keyID, material := range ring.keys {
		block, err := aes.NewCipher(material.key)
		if err != nil {
			return ConflictTokenCodec{}, ErrConflictTokenUnavailable
		}
		aead, err := cipher.NewGCM(block)
		if err != nil || aead.NonceSize() != 12 {
			return ConflictTokenCodec{}, ErrConflictTokenUnavailable
		}
		codec.keys[keyID] = conflictTokenCipher{
			aead:          aead,
			state:         material.state,
			deactivatedAt: cloneTime(material.deactivatedAt),
			retireAt:      cloneTime(material.retireAt),
		}
	}
	return codec, nil
}

func RequestHashTokenValue(requestHash []byte) string {
	return base64.RawURLEncoding.EncodeToString(requestHash)
}

func (c ConflictTokenCodec) Issue(claims ConflictTokenClaims) (string, error) {
	key, ok := c.keys[c.activeKeyID]
	if !ok || key.state != conflictTokenKeyStateActive || c.now == nil || c.entropy == nil {
		return "", ErrConflictTokenUnavailable
	}
	now := c.now().UTC()
	claims.Version = ConflictTokenVersion
	claims.IssuedAt = now
	claims.ExpiresAt = now.Add(ConflictTokenTTL)
	if !validConflictTokenClaims(claims) {
		return "", ErrConflictTokenUnavailable
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", ErrConflictTokenUnavailable
	}
	nonce := make([]byte, key.aead.NonceSize())
	if _, err := io.ReadFull(c.entropy, nonce); err != nil {
		return "", ErrConflictTokenUnavailable
	}
	sealed := key.aead.Seal(nonce, nonce, payload, conflictTokenAAD(c.activeKeyID))
	token := conflictTokenPrefix + c.activeKeyID + "." + base64.RawURLEncoding.EncodeToString(sealed)
	if len(token) > ConflictTokenMaximumLen {
		return "", ErrConflictTokenUnavailable
	}
	return token, nil
}

func (c ConflictTokenCodec) Parse(token string) (ConflictTokenClaims, bool) {
	if len(token) == 0 || len(token) > ConflictTokenMaximumLen || !strings.HasPrefix(token, conflictTokenPrefix) || c.now == nil {
		return ConflictTokenClaims{}, false
	}
	remainder := strings.TrimPrefix(token, conflictTokenPrefix)
	separator := strings.LastIndexByte(remainder, '.')
	if separator <= 0 || separator == len(remainder)-1 {
		return ConflictTokenClaims{}, false
	}
	keyID := remainder[:separator]
	if !conflictTokenKeyIDPattern.MatchString(keyID) {
		return ConflictTokenClaims{}, false
	}
	key, ok := c.keys[keyID]
	if !ok {
		return ConflictTokenClaims{}, false
	}
	now := c.now().UTC()
	if key.retireAt != nil && !now.Before(*key.retireAt) {
		return ConflictTokenClaims{}, false
	}
	sealed, err := base64.RawURLEncoding.DecodeString(remainder[separator+1:])
	if err != nil || len(sealed) <= key.aead.NonceSize() {
		return ConflictTokenClaims{}, false
	}
	payload, err := key.aead.Open(nil, sealed[:key.aead.NonceSize()], sealed[key.aead.NonceSize():], conflictTokenAAD(keyID))
	if err != nil {
		return ConflictTokenClaims{}, false
	}
	var claims ConflictTokenClaims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return ConflictTokenClaims{}, false
	}
	if token, err := decoder.Token(); err != io.EOF || token != nil {
		return ConflictTokenClaims{}, false
	}
	if !validConflictTokenClaims(claims) || claims.IssuedAt.After(now.Add(ConflictTokenClockSkew)) || !now.Before(claims.ExpiresAt) {
		return ConflictTokenClaims{}, false
	}
	if key.state == conflictTokenKeyStateDecryptOnly && (key.deactivatedAt == nil || !claims.IssuedAt.Before(*key.deactivatedAt)) {
		return ConflictTokenClaims{}, false
	}
	return claims, true
}

func conflictTokenAAD(keyID string) []byte {
	return []byte(conflictTokenPurpose + "\x00" + keyID)
}

func validConflictTokenClaims(claims ConflictTokenClaims) bool {
	recordID, err := uuid.Parse(claims.RecordID)
	requestHash, hashErr := base64.RawURLEncoding.DecodeString(claims.RequestHash)
	return err == nil && recordID != uuid.Nil && hashErr == nil && len(requestHash) > 0 && len(requestHash) <= 128 &&
		claims.Version == ConflictTokenVersion &&
		claims.RouteKey != "" && len(claims.RouteKey) <= 256 &&
		claims.ViewSchemaID != "" && len(claims.ViewSchemaID) <= 256 &&
		claims.FieldKey != "" && len(claims.FieldKey) <= 256 &&
		claims.ConflictResolutionClass != "" && len(claims.ConflictResolutionClass) <= 128 &&
		claims.BaseRowVersion >= 1 &&
		claims.CurrentRowVersion >= claims.BaseRowVersion &&
		!claims.IssuedAt.IsZero() && claims.IssuedAt.Location() == time.UTC &&
		!claims.ExpiresAt.IsZero() && claims.ExpiresAt.Location() == time.UTC &&
		claims.ExpiresAt.Equal(claims.IssuedAt.Add(ConflictTokenTTL))
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

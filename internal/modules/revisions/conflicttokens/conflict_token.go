package conflicttokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const ConflictTokenVersion = 2

type ConflictTokenClaims struct {
	Version                 int64  `json:"cartulary_conflict_token_v"`
	RouteKey                string `json:"route_key"`
	RecordID                string `json:"record_id"`
	ViewSchemaID            string `json:"view_schema_id"`
	FieldKey                string `json:"field_key"`
	ConflictResolutionClass string `json:"conflict_resolution_class"`
	BaseRowVersion          int64  `json:"base_row_version"`
	CurrentRowVersion       int64  `json:"current_row_version"`
	RequestHash             string `json:"request_hash"`
	Signature               string `json:"sig"`
}

type ConflictTokenCodec struct {
	key [32]byte
}

func NewConflictTokenCodec(keys authn.MasterKeys) ConflictTokenCodec {
	return ConflictTokenCodec{key: authn.DerivePurposeKey(keys, "record-conflict-token-v2")}
}

func NewConflictTokenCodecForTesting(scope string) ConflictTokenCodec {
	return ConflictTokenCodec{key: sha256.Sum256([]byte("cartulary-test-conflict-token-v2:" + scope))}
}

func RequestHashTokenValue(requestHash []byte) string {
	return base64.RawURLEncoding.EncodeToString(requestHash)
}

func (c ConflictTokenCodec) Issue(claims ConflictTokenClaims) string {
	claims.Version = ConflictTokenVersion
	claims.Signature = ""
	claims.Signature = c.signature(claims)
	data, _ := json.Marshal(claims)
	return base64.RawURLEncoding.EncodeToString(data)
}

func (c ConflictTokenCodec) Parse(token string) (ConflictTokenClaims, bool) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return ConflictTokenClaims{}, false
	}
	var claims ConflictTokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return ConflictTokenClaims{}, false
	}
	if !validConflictTokenClaims(claims) {
		return ConflictTokenClaims{}, false
	}
	signature := claims.Signature
	claims.Signature = ""
	if !hmac.Equal([]byte(signature), []byte(c.signature(claims))) {
		return ConflictTokenClaims{}, false
	}
	claims.Signature = signature
	return claims, true
}

func (c ConflictTokenCodec) signature(claims ConflictTokenClaims) string {
	claims.Signature = ""
	payload, _ := json.Marshal(struct {
		Version                 int64  `json:"cartulary_conflict_token_v"`
		RouteKey                string `json:"route_key"`
		RecordID                string `json:"record_id"`
		ViewSchemaID            string `json:"view_schema_id"`
		FieldKey                string `json:"field_key"`
		ConflictResolutionClass string `json:"conflict_resolution_class"`
		BaseRowVersion          int64  `json:"base_row_version"`
		CurrentRowVersion       int64  `json:"current_row_version"`
		RequestHash             string `json:"request_hash"`
	}{
		Version:                 claims.Version,
		RouteKey:                claims.RouteKey,
		RecordID:                claims.RecordID,
		ViewSchemaID:            claims.ViewSchemaID,
		FieldKey:                claims.FieldKey,
		ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion:          claims.BaseRowVersion,
		CurrentRowVersion:       claims.CurrentRowVersion,
		RequestHash:             claims.RequestHash,
	})
	mac := hmac.New(sha256.New, c.key[:])
	_, _ = mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validConflictTokenClaims(claims ConflictTokenClaims) bool {
	recordID, err := uuid.Parse(claims.RecordID)
	if err != nil || recordID == uuid.Nil {
		return false
	}
	return claims.Version == ConflictTokenVersion &&
		claims.Signature != "" &&
		claims.RouteKey != "" &&
		claims.ViewSchemaID != "" &&
		claims.FieldKey != "" &&
		claims.ConflictResolutionClass != "" &&
		claims.BaseRowVersion >= 1 &&
		claims.CurrentRowVersion >= claims.BaseRowVersion &&
		claims.RequestHash != ""
}

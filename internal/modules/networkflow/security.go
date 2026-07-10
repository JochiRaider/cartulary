package networkflow

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	cursorVersion = "cartulary.network_flow.cursor.v1"
	cursorTTL     = 15 * time.Minute
)

var safeKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type CursorCodec struct {
	keyID string
	aead  cipher.AEAD
	now   func() time.Time
}

type CursorBinding struct {
	Route       string
	ActorUserID string
	IncidentID  string
	Scope       map[string]string
	QueryHash   string
	QueryEcho   json.RawMessage
	Limit       int
}

type CursorPayload struct {
	Version     string            `json:"version"`
	Route       string            `json:"route"`
	ActorUserID string            `json:"actor_user_id"`
	IncidentID  string            `json:"incident_id"`
	Scope       map[string]string `json:"scope"`
	QueryHash   string            `json:"query_hash"`
	QueryEcho   json.RawMessage   `json:"query_echo"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
	IssuedAt    time.Time         `json:"issued_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
}

func NewCursorCodec(keyID string, key []byte, now func() time.Time) (*CursorCodec, error) {
	if !safeKeyIDPattern.MatchString(keyID) {
		return nil, fmt.Errorf("network flow cursor key id invalid")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("network flow cursor key invalid: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("network flow cursor aead invalid: %w", err)
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &CursorCodec{keyID: keyID, aead: aead, now: now}, nil
}

func (c *CursorCodec) Encode(binding CursorBinding, offset int) (string, error) {
	if c == nil {
		return "", ErrInvalidCursor
	}
	issuedAt := c.now().UTC()
	payload := CursorPayload{
		Version:     cursorVersion,
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		IncidentID:  binding.IncidentID,
		Scope:       cloneStringMap(binding.Scope),
		QueryHash:   binding.QueryHash,
		QueryEcho:   append(json.RawMessage(nil), binding.QueryEcho...),
		Limit:       binding.Limit,
		Offset:      offset,
		IssuedAt:    issuedAt,
		ExpiresAt:   issuedAt.Add(cursorTTL),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, encoded, []byte(cursorVersion+"."+c.keyID))
	return "nfc1." + c.keyID + "." + base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (c *CursorCodec) Decode(token string) (CursorPayload, string) {
	if c == nil || token == "" {
		return CursorPayload{}, "malformed"
	}
	if len(token) > 4096 {
		return CursorPayload{}, "too_long"
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nfc1" || parts[1] != c.keyID {
		return CursorPayload{}, "malformed"
	}
	sealed, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return CursorPayload{}, "malformed"
	}
	nonceSize := c.aead.NonceSize()
	if len(sealed) <= nonceSize {
		return CursorPayload{}, "malformed"
	}
	payload, err := c.aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], []byte(cursorVersion+"."+c.keyID))
	if err != nil {
		return CursorPayload{}, "malformed"
	}
	var decoded CursorPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return CursorPayload{}, "malformed"
	}
	if decoded.Version != cursorVersion || decoded.Route == "" || decoded.ActorUserID == "" || decoded.IncidentID == "" || decoded.Limit < 1 || decoded.Offset < 0 || decoded.QueryHash == "" || len(decoded.QueryEcho) == 0 || !json.Valid(decoded.QueryEcho) {
		return CursorPayload{}, "malformed"
	}
	current := c.now().UTC()
	if decoded.IssuedAt.After(current) || !current.Before(decoded.ExpiresAt) {
		return CursorPayload{}, "expired"
	}
	return decoded, ""
}

func (c CursorPayload) Validate(binding CursorBinding) string {
	if c.Route != binding.Route {
		return "route_mismatch"
	}
	if c.ActorUserID != binding.ActorUserID {
		return "actor_mismatch"
	}
	if c.IncidentID != binding.IncidentID || !equalStringMap(c.Scope, binding.Scope) || c.QueryHash != binding.QueryHash || c.Limit != binding.Limit {
		return "semantic_query_mismatch"
	}
	return ""
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func equalStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, left := range a {
		if right, ok := b[key]; !ok || right != left {
			return false
		}
	}
	return true
}

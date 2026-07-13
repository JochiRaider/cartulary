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
	"sync"
	"time"
)

const (
	cursorVersion = "cartulary.network_flow.cursor.v2"
	cursorTTL     = 15 * time.Minute
)

var safeKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type cursorCipher struct {
	aead          cipher.AEAD
	state         string
	deactivatedAt *time.Time
	retireAt      *time.Time
}

type CursorCodec struct {
	mu          sync.Mutex
	activeKeyID string
	keys        map[string]cursorCipher
	now         func() time.Time
}

type CursorProtector interface {
	Encode(CursorBinding, string, any) (string, error)
	Decode(string) (CursorPayload, string)
}

type CursorBinding struct {
	Route       string
	ActorUserID string
	SessionID   string
	IncidentID  string
	Scope       map[string]string
	QueryHash   string
	QueryEcho   json.RawMessage
	Limit       int
}

type CursorPayload struct {
	Version      string            `json:"version"`
	PositionKind string            `json:"position_kind"`
	Position     json.RawMessage   `json:"position"`
	Route        string            `json:"route"`
	ActorUserID  string            `json:"actor_user_id"`
	SessionID    string            `json:"session_id,omitempty"`
	IncidentID   string            `json:"incident_id"`
	Scope        map[string]string `json:"scope"`
	QueryHash    string            `json:"query_hash"`
	QueryEcho    json.RawMessage   `json:"query_echo"`
	Limit        int               `json:"limit"`
	IssuedAt     time.Time         `json:"issued_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
}

func newCursorCodec(rings *KeyRings, now func() time.Time) (*CursorCodec, error) {
	if rings == nil || rings.cursorActiveID == "" {
		return nil, fmt.Errorf("network flow cursor key ring unavailable")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	codec := &CursorCodec{activeKeyID: rings.cursorActiveID, keys: make(map[string]cursorCipher), now: now}
	for keyID, material := range rings.cursorKeys {
		block, err := aes.NewCipher(material.key)
		if err != nil {
			return nil, fmt.Errorf("network flow cursor key invalid: %w", err)
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("network flow cursor aead invalid: %w", err)
		}
		codec.keys[keyID] = cursorCipher{aead: aead, state: material.state, deactivatedAt: material.deactivatedAt, retireAt: material.retireAt}
	}
	return codec, nil
}

func (c *CursorCodec) Encode(binding CursorBinding, positionKind string, position any) (string, error) {
	if c == nil || positionKind == "" || position == nil {
		return "", ErrInvalidCursor
	}
	current := c.now().UTC()
	c.mu.Lock()
	c.purgeRetiredLocked(current)
	key, ok := c.keys[c.activeKeyID]
	c.mu.Unlock()
	if !ok || key.state != "active" {
		return "", ErrInvalidCursor
	}
	positionJSON, err := json.Marshal(position)
	if err != nil {
		return "", err
	}
	issuedAt := current
	payload := CursorPayload{
		Version:      cursorVersion,
		PositionKind: positionKind,
		Position:     positionJSON,
		Route:        binding.Route,
		ActorUserID:  binding.ActorUserID,
		SessionID:    binding.SessionID,
		IncidentID:   binding.IncidentID,
		Scope:        cloneStringMap(binding.Scope),
		QueryHash:    binding.QueryHash,
		QueryEcho:    append(json.RawMessage(nil), binding.QueryEcho...),
		Limit:        binding.Limit,
		IssuedAt:     issuedAt,
		ExpiresAt:    issuedAt.Add(cursorTTL),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, key.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	aad := []byte("nfc2." + c.activeKeyID)
	sealed := key.aead.Seal(nonce, nonce, encoded, aad)
	return "nfc2." + c.activeKeyID + "." + base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (c *CursorCodec) Decode(token string) (CursorPayload, string) {
	if c == nil || token == "" {
		return CursorPayload{}, "malformed"
	}
	if len(token) > 4096 {
		return CursorPayload{}, "too_long"
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nfc2" || !safeKeyIDPattern.MatchString(parts[1]) {
		return CursorPayload{}, "malformed"
	}
	current := c.now().UTC()
	c.mu.Lock()
	c.purgeRetiredLocked(current)
	key, ok := c.keys[parts[1]]
	c.mu.Unlock()
	if !ok {
		return CursorPayload{}, "malformed"
	}
	sealed, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return CursorPayload{}, "malformed"
	}
	nonceSize := key.aead.NonceSize()
	if len(sealed) <= nonceSize {
		return CursorPayload{}, "malformed"
	}
	payload, err := key.aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], []byte("nfc2."+parts[1]))
	if err != nil {
		return CursorPayload{}, "malformed"
	}
	var decoded CursorPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return CursorPayload{}, "malformed"
	}
	if decoded.Version != cursorVersion || decoded.PositionKind == "" || len(decoded.Position) == 0 || !json.Valid(decoded.Position) || decoded.Route == "" || decoded.ActorUserID == "" || decoded.IncidentID == "" || decoded.Limit < 1 || decoded.QueryHash == "" || len(decoded.QueryEcho) == 0 || !json.Valid(decoded.QueryEcho) {
		return CursorPayload{}, "malformed"
	}
	if decoded.IssuedAt.After(current) || !current.Before(decoded.ExpiresAt) || !decoded.ExpiresAt.Equal(decoded.IssuedAt.Add(cursorTTL)) {
		return CursorPayload{}, "expired"
	}
	if key.state == "decrypt_only" && (key.deactivatedAt == nil || !decoded.IssuedAt.Before(*key.deactivatedAt)) {
		return CursorPayload{}, "malformed"
	}
	return decoded, ""
}

func (c *CursorCodec) purgeRetiredLocked(now time.Time) {
	for keyID, key := range c.keys {
		if key.retireAt != nil && !now.Before(*key.retireAt) {
			delete(c.keys, keyID)
		}
	}
}

func (c CursorPayload) Validate(binding CursorBinding) string {
	if c.Route != binding.Route {
		return "route_mismatch"
	}
	if c.ActorUserID != binding.ActorUserID || c.SessionID != binding.SessionID {
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

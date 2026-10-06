package networkflow

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

const (
	cursorVersion = "cartulary.network_flow.cursor.v3"
	cursorTTL     = 15 * time.Minute
)

var safeKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type cursorCipher struct {
	key           cryptography.MasterKey
	state         string
	deactivatedAt *time.Time
	retireAt      *time.Time
}

type cursorCodec struct {
	mu          sync.Mutex
	activeKeyID string
	keys        map[string]cursorCipher
	now         func() time.Time
}

type cursorProtector interface {
	Encode(cursorBinding, string, any) (string, error)
	Decode(string) (cursorPayload, string)
}

type cursorBinding struct {
	Route       string
	ActorUserID string
	SessionID   string
	IncidentID  string
	Scope       map[string]string
	QueryHash   string
	QueryEcho   json.RawMessage
	Limit       int
}

type cursorPayload struct {
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

func newCursorCodec(rings *KeyRings, now func() time.Time) (*cursorCodec, error) {
	if rings == nil || rings.cursorActiveID == "" {
		return nil, fmt.Errorf("network flow cursor key ring unavailable")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	codec := &cursorCodec{activeKeyID: rings.cursorActiveID, keys: make(map[string]cursorCipher), now: now}
	for keyID, material := range rings.cursorKeys {
		key, err := cryptography.AdmitKey(material.key)
		if err != nil {
			return nil, fmt.Errorf("network flow cursor key invalid: %w", err)
		}
		codec.keys[keyID] = cursorCipher{key: key, state: material.state, deactivatedAt: material.deactivatedAt, retireAt: material.retireAt}
	}
	return codec, nil
}

func (c *cursorCodec) Encode(binding cursorBinding, positionKind string, position any) (string, error) {
	if c == nil || positionKind == "" || position == nil {
		return "", errInvalidCursor
	}
	current := c.now().UTC()
	c.mu.Lock()
	c.purgeRetiredLocked(current)
	key, ok := c.keys[c.activeKeyID]
	c.mu.Unlock()
	if !ok || key.state != "active" {
		return "", errInvalidCursor
	}
	positionJSON, err := json.Marshal(position)
	if err != nil {
		return "", err
	}
	issuedAt := current
	payload := cursorPayload{
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
	sealed, err := key.key.Seal(cursorVersion, []string{c.activeKeyID}, encoded, []byte("nfc3."))
	if err != nil {
		return "", errInvalidCursor
	}
	token := "nfc3." + c.activeKeyID + "." + base64.RawURLEncoding.EncodeToString(sealed)
	if len(token) > 4096 {
		return "", errInvalidCursor
	}
	return token, nil
}

func (c *cursorCodec) Decode(token string) (cursorPayload, string) {
	if c == nil || token == "" {
		return cursorPayload{}, "malformed"
	}
	if len(token) > 4096 {
		return cursorPayload{}, "too_long"
	}
	if !strings.HasPrefix(token, "nfc3.") {
		return cursorPayload{}, "malformed"
	}
	remainder := strings.TrimPrefix(token, "nfc3.")
	separator := strings.LastIndexByte(remainder, '.')
	if separator <= 0 || separator == len(remainder)-1 {
		return cursorPayload{}, "malformed"
	}
	keyID, encoded := remainder[:separator], remainder[separator+1:]
	if !safeKeyIDPattern.MatchString(keyID) {
		return cursorPayload{}, "malformed"
	}
	current := c.now().UTC()
	c.mu.Lock()
	c.purgeRetiredLocked(current)
	key, ok := c.keys[keyID]
	c.mu.Unlock()
	if !ok {
		return cursorPayload{}, "malformed"
	}
	sealed, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.RawURLEncoding.EncodeToString(sealed) != encoded {
		return cursorPayload{}, "malformed"
	}
	payload, err := key.key.Open(cursorVersion, []string{keyID}, sealed, []byte("nfc3."))
	if err != nil {
		return cursorPayload{}, "malformed"
	}
	var decoded cursorPayload
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return cursorPayload{}, "malformed"
	}
	if trailing, err := decoder.Token(); err != io.EOF || trailing != nil {
		return cursorPayload{}, "malformed"
	}
	if decoded.Version != cursorVersion || decoded.PositionKind == "" || len(decoded.Position) == 0 || !json.Valid(decoded.Position) || decoded.Route == "" || decoded.ActorUserID == "" || decoded.IncidentID == "" || decoded.Limit < 1 || decoded.QueryHash == "" || len(decoded.QueryEcho) == 0 || !json.Valid(decoded.QueryEcho) {
		return cursorPayload{}, "malformed"
	}
	if decoded.IssuedAt.After(current) || !current.Before(decoded.ExpiresAt) || !decoded.ExpiresAt.Equal(decoded.IssuedAt.Add(cursorTTL)) {
		return cursorPayload{}, "expired"
	}
	if key.state == "decrypt_only" && (key.deactivatedAt == nil || !decoded.IssuedAt.Before(*key.deactivatedAt)) {
		return cursorPayload{}, "malformed"
	}
	return decoded, ""
}

func (c *cursorCodec) purgeRetiredLocked(now time.Time) {
	for keyID, key := range c.keys {
		if key.retireAt != nil && !now.Before(*key.retireAt) {
			delete(c.keys, keyID)
		}
	}
}

func (c cursorPayload) Validate(binding cursorBinding) string {
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

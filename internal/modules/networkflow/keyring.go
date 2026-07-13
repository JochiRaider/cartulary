package networkflow

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
)

const (
	KeyRingsOverrideKey     = "networkflow.key_rings.v1"
	keyRingManifestSchemaID = "cartulary.network_flow_key_rings.v1"
	keyRingManifestMaxBytes = 65536
	keyRingMaxKeys          = 8
	keyRingConfigPath       = "network_flow_activity.key_ring_manifest_path"
)

var keyRingTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?Z$`)

type KeyRings struct {
	mu             sync.Mutex
	cursorActiveID string
	cursorKeys     map[string]cursorKeyMaterial
	safeActiveID   string
	safeKeys       map[string]safeDigestKeyMaterial
}

type cursorKeyMaterial struct {
	key           []byte
	state         string
	deactivatedAt *time.Time
	retireAt      *time.Time
}

type safeDigestKeyMaterial struct {
	key           []byte
	state         string
	deactivatedAt *time.Time
	retainUntil   *time.Time
}

type SafeDigester interface {
	Digest(valueClass string, canonicalValue string) (string, string, error)
}

type keyRingSafeDigester struct {
	rings *KeyRings
	now   func() time.Time
}

type keyRingSecretRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type keyRingManifest struct {
	SchemaID          string                    `json:"schema_id"`
	CursorKeyRing     cursorKeyRingManifest     `json:"cursor_key_ring"`
	SafeDigestKeyRing safeDigestKeyRingManifest `json:"safe_digest_key_ring"`
}

type cursorKeyRingManifest struct {
	Algorithm string              `json:"algorithm"`
	Keys      []cursorKeyManifest `json:"keys"`
}

type cursorKeyManifest struct {
	KeyID         string           `json:"cursor_key_id"`
	State         string           `json:"state"`
	SecretRef     keyRingSecretRef `json:"secret_ref"`
	DeactivatedAt string           `json:"deactivated_at,omitempty"`
	RetireAt      string           `json:"retire_at,omitempty"`
}

type safeDigestKeyRingManifest struct {
	Algorithm string                  `json:"algorithm"`
	Keys      []safeDigestKeyManifest `json:"keys"`
}

type safeDigestKeyManifest struct {
	KeyID         string           `json:"safe_digest_key_id"`
	State         string           `json:"state"`
	SecretRef     keyRingSecretRef `json:"secret_ref"`
	DeactivatedAt string           `json:"deactivated_at,omitempty"`
	RetainUntil   string           `json:"retain_until,omitempty"`
}

func LoadKeyRings(cfg config.Config, env map[string]string, now time.Time) (*KeyRings, error) {
	registry := secretpurpose.NewRegistry()
	if err := authn.RegisterMasterSecretPurpose(registry, env); err != nil {
		return nil, err
	}
	return LoadKeyRingsWithRegistry(cfg, env, now, registry)
}

func LoadKeyRingsWithRegistry(cfg config.Config, env map[string]string, now time.Time, registry *secretpurpose.Registry) (*KeyRings, error) {
	if !cfg.NetworkFlowActivity.Claimed {
		return nil, nil
	}
	path := cfg.NetworkFlowActivity.KeyRingManifestPath
	info, err := os.Stat(path)
	if err != nil {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_missing", safeKeyRingFileError("stat Network Flow key-ring manifest", err))
	}
	if !info.Mode().IsRegular() {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_invalid", "Network Flow key-ring manifest path must reference one regular file")
	}
	if info.Size() > keyRingManifestMaxBytes {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_invalid", "Network Flow key-ring manifest exceeds 65536 bytes")
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- validated operator-configured absolute path.
	if err != nil {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_missing", safeKeyRingFileError("read Network Flow key-ring manifest", err))
	}
	return parseKeyRings(raw, env, now, registry)
}

func ParseKeyRings(raw []byte, env map[string]string, now time.Time) (*KeyRings, error) {
	registry := secretpurpose.NewRegistry()
	if err := authn.RegisterMasterSecretPurpose(registry, env); err != nil {
		return nil, err
	}
	return parseKeyRings(raw, env, now, registry)
}

func parseKeyRings(raw []byte, env map[string]string, now time.Time, registry *secretpurpose.Registry) (*KeyRings, error) {
	if len(raw) == 0 || len(raw) > keyRingManifestMaxBytes || !utf8.Valid(raw) {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_invalid", "Network Flow key-ring manifest must be non-empty UTF-8 JSON no larger than 65536 bytes")
	}
	if err := validateClosedKeyRingJSON(raw); err != nil {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_invalid", err.Error())
	}
	var manifest keyRingManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_invalid", "parse Network Flow key-ring manifest: "+err.Error())
	}
	if manifest.SchemaID != keyRingManifestSchemaID {
		return nil, keyRingConfigError(keyRingConfigPath+".schema_id", "network_flow_cursor_key_invalid", "schema_id must equal cartulary.network_flow_key_rings.v1")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	rings := &KeyRings{
		cursorKeys: make(map[string]cursorKeyMaterial),
		safeKeys:   make(map[string]safeDigestKeyMaterial),
	}
	if registry == nil {
		return nil, keyRingConfigError(keyRingConfigPath, "network_flow_cursor_key_invalid", "startup secret-purpose registry is unavailable")
	}
	if err := rings.loadCursorRing(manifest.CursorKeyRing, env, now, registry); err != nil {
		return nil, err
	}
	if err := rings.loadSafeDigestRing(manifest.SafeDigestKeyRing, env, now, registry); err != nil {
		return nil, err
	}
	return rings, nil
}

func (r *KeyRings) loadCursorRing(ring cursorKeyRingManifest, env map[string]string, now time.Time, registry *secretpurpose.Registry) error {
	if ring.Algorithm != "aes_256_gcm_v1" || len(ring.Keys) < 1 || len(ring.Keys) > keyRingMaxKeys {
		return keyRingConfigError(keyRingConfigPath+".cursor_key_ring", "network_flow_cursor_key_invalid", "cursor key ring must use aes_256_gcm_v1 and contain 1..8 keys")
	}
	active := 0
	for index, entry := range ring.Keys {
		path := fmt.Sprintf("%s.cursor_key_ring.keys[%d]", keyRingConfigPath, index)
		if !safeKeyIDPattern.MatchString(entry.KeyID) {
			return keyRingConfigError(path+".cursor_key_id", "network_flow_cursor_key_invalid", "cursor_key_id is invalid")
		}
		if _, exists := r.cursorKeys[entry.KeyID]; exists {
			return keyRingConfigError(path+".cursor_key_id", "network_flow_cursor_key_id_conflict", "cursor_key_id values must be unique")
		}
		key, err := resolveKeyRingSecret(entry.SecretRef, env, path+".secret_ref", "network_flow_cursor_key_secret_missing")
		if err != nil {
			return err
		}
		if err := registry.Register(entry.SecretRef.Name, "network_flow_cursor", key); err != nil {
			return keyRingConfigError(path+".secret_ref", "network_flow_cursor_key_invalid", "resolved secret reference or key material is already registered for another purpose")
		}
		stored := cursorKeyMaterial{key: key, state: entry.State}
		switch entry.State {
		case "active":
			active++
			if entry.DeactivatedAt != "" || entry.RetireAt != "" {
				return keyRingConfigError(path, "network_flow_cursor_rotation_invalid", "active cursor key must omit deactivated_at and retire_at")
			}
			r.cursorActiveID = entry.KeyID
		case "decrypt_only":
			deactivatedAt, err := parseKeyRingTimestamp(entry.DeactivatedAt)
			if err != nil {
				return keyRingConfigError(path+".deactivated_at", "network_flow_cursor_rotation_invalid", err.Error())
			}
			retireAt, err := parseKeyRingTimestamp(entry.RetireAt)
			if err != nil || !retireAt.Equal(deactivatedAt.Add(cursorTTL)) || !now.Before(retireAt) {
				return keyRingConfigError(path+".retire_at", "network_flow_cursor_rotation_invalid", "retire_at must equal deactivated_at plus 15 minutes and remain in the future")
			}
			stored.deactivatedAt = &deactivatedAt
			stored.retireAt = &retireAt
		default:
			return keyRingConfigError(path+".state", "network_flow_cursor_rotation_invalid", "cursor key state must be active or decrypt_only")
		}
		r.cursorKeys[entry.KeyID] = stored
	}
	if active != 1 {
		return keyRingConfigError(keyRingConfigPath+".cursor_key_ring.keys", "network_flow_cursor_rotation_invalid", "cursor key ring must contain exactly one active key")
	}
	return nil
}

func (r *KeyRings) loadSafeDigestRing(ring safeDigestKeyRingManifest, env map[string]string, now time.Time, registry *secretpurpose.Registry) error {
	if ring.Algorithm != "hmac_sha256_v1" || len(ring.Keys) < 1 || len(ring.Keys) > keyRingMaxKeys {
		return keyRingConfigError(keyRingConfigPath+".safe_digest_key_ring", "network_flow_safe_digest_key_invalid", "safe-digest key ring must use hmac_sha256_v1 and contain 1..8 keys")
	}
	active := 0
	for index, entry := range ring.Keys {
		path := fmt.Sprintf("%s.safe_digest_key_ring.keys[%d]", keyRingConfigPath, index)
		if !safeKeyIDPattern.MatchString(entry.KeyID) {
			return keyRingConfigError(path+".safe_digest_key_id", "network_flow_safe_digest_key_invalid", "safe_digest_key_id is invalid")
		}
		if _, exists := r.safeKeys[entry.KeyID]; exists {
			return keyRingConfigError(path+".safe_digest_key_id", "network_flow_safe_digest_key_id_conflict", "safe_digest_key_id values must be unique")
		}
		key, err := resolveKeyRingSecret(entry.SecretRef, env, path+".secret_ref", "network_flow_safe_digest_key_secret_missing")
		if err != nil {
			return err
		}
		if err := registry.Register(entry.SecretRef.Name, "network_flow_safe_digest", key); err != nil {
			return keyRingConfigError(path+".secret_ref", "network_flow_safe_digest_key_purpose_conflict", "resolved secret reference or key material is already registered for another purpose")
		}
		stored := safeDigestKeyMaterial{key: key, state: entry.State}
		switch entry.State {
		case "active":
			active++
			if entry.DeactivatedAt != "" || entry.RetainUntil != "" {
				return keyRingConfigError(path, "network_flow_safe_digest_rotation_invalid", "active safe-digest key must omit deactivated_at and retain_until")
			}
			r.safeActiveID = entry.KeyID
		case "inactive":
			deactivatedAt, err := parseKeyRingTimestamp(entry.DeactivatedAt)
			if err != nil {
				return keyRingConfigError(path+".deactivated_at", "network_flow_safe_digest_rotation_invalid", err.Error())
			}
			retainUntil, err := parseKeyRingTimestamp(entry.RetainUntil)
			if err != nil || !retainUntil.After(deactivatedAt) || !now.Before(retainUntil) {
				return keyRingConfigError(path+".retain_until", "network_flow_safe_digest_rotation_invalid", "retain_until must be later than deactivated_at and remain in the future")
			}
			stored.deactivatedAt = &deactivatedAt
			stored.retainUntil = &retainUntil
		default:
			return keyRingConfigError(path+".state", "network_flow_safe_digest_rotation_invalid", "safe-digest key state must be active or inactive")
		}
		r.safeKeys[entry.KeyID] = stored
	}
	if active != 1 {
		return keyRingConfigError(keyRingConfigPath+".safe_digest_key_ring.keys", "network_flow_safe_digest_rotation_invalid", "safe-digest key ring must contain exactly one active key")
	}
	return nil
}

func newSafeDigester(rings *KeyRings, now func() time.Time) (SafeDigester, error) {
	if rings == nil || rings.safeActiveID == "" {
		return nil, errors.New("network flow safe-digest key ring unavailable")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	digester := &keyRingSafeDigester{rings: rings, now: now}
	if _, _, err := digester.Digest("startup_validation", "startup_validation"); err != nil {
		return nil, err
	}
	return digester, nil
}

func (d *keyRingSafeDigester) Digest(valueClass string, canonicalValue string) (string, string, error) {
	if d == nil || d.rings == nil || valueClass == "" {
		return "", "", errors.New("network flow safe digester unavailable")
	}
	d.rings.mu.Lock()
	defer d.rings.mu.Unlock()
	d.rings.purgeExpiredLocked(d.now().UTC())
	entry, ok := d.rings.safeKeys[d.rings.safeActiveID]
	if !ok || entry.state != "active" || len(entry.key) != 32 {
		return "", "", errors.New("network flow active safe-digest key unavailable")
	}
	digest, keyID := SafeDigest(d.rings.safeActiveID, entry.key, valueClass, canonicalValue)
	return digest, keyID, nil
}

func (r *KeyRings) purgeExpiredLocked(now time.Time) {
	for keyID, material := range r.cursorKeys {
		if material.retireAt != nil && !now.Before(*material.retireAt) {
			clear(material.key)
			delete(r.cursorKeys, keyID)
		}
	}
	for keyID, material := range r.safeKeys {
		if material.retainUntil != nil && !now.Before(*material.retainUntil) {
			clear(material.key)
			delete(r.safeKeys, keyID)
		}
	}
}

func resolveKeyRingSecret(ref keyRingSecretRef, env map[string]string, path string, reason string) ([]byte, error) {
	if ref.Kind != "env" || !safeKeyIDPattern.MatchString(ref.Name) {
		return nil, keyRingConfigError(path, reason, "secret_ref must contain kind env and a valid name")
	}
	name := "CARTULARY_SECRET_" + normalizedKeyRingSecretSuffix(ref.Name)
	value, ok := env[name]
	if env == nil {
		value, ok = os.LookupEnv(name)
	}
	if !ok || value == "" || strings.Contains(value, "=") {
		return nil, keyRingConfigError(path, reason, "secret_ref could not be resolved to unpadded base64url key material")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return nil, keyRingConfigError(path, reason, "resolved key material must be unpadded base64url encoding of exactly 32 bytes")
	}
	return append([]byte(nil), decoded...), nil
}

func parseKeyRingTimestamp(value string) (time.Time, error) {
	if !keyRingTimestampPattern.MatchString(value) {
		return time.Time{}, errors.New("timestamp must use canonical UTC timestamp_utc_v1 syntax")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC {
		return time.Time{}, errors.New("timestamp is invalid")
	}
	return parsed.UTC(), nil
}

func validateClosedKeyRingJSON(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := scanClosedKeyRingJSONValue(decoder); err != nil {
		return err
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("network flow key-ring manifest contains trailing JSON token %v", token)
	}
	return nil
}

func scanClosedKeyRingJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("explicit null is invalid in Network Flow key-ring manifests")
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			if !ok {
				return errors.New("network flow key-ring object member is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate Network Flow key-ring member %q", key)
			}
			seen[key] = struct{}{}
			if err := scanClosedKeyRingJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanClosedKeyRingJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return nil
	}
}

func normalizedKeyRingSecretSuffix(name string) string {
	var builder strings.Builder
	underscore := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r - ('a' - 'A'))
			underscore = false
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			underscore = false
		default:
			if builder.Len() > 0 && !underscore {
				builder.WriteByte('_')
				underscore = true
			}
		}
	}
	return strings.Trim(builder.String(), "_")
}

func keyRingConfigError(path string, reason string, message string) error {
	return config.NewDiagnosticsError(config.Diagnostic{Path: path, ReasonCode: reason, Message: message})
}

func safeKeyRingFileError(action string, err error) string {
	if errors.Is(err, fs.ErrPermission) {
		return action + ": " + fs.ErrPermission.Error()
	}
	return action + ": " + err.Error()
}

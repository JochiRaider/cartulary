package conflicts

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
)

const (
	conflictTokenKeyRingSchemaID       = "cartulary.revisions_conflict_token_key_ring.v1"
	conflictTokenKeyRingAlgorithm      = "aes_256_gcm_v1"
	conflictTokenKeyRingMaximumKeys    = 8
	conflictTokenKeyStateActive        = "active"
	conflictTokenKeyStateDecryptOnly   = "decrypt_only"
	conflictTokenSecretPurpose         = "revisions_conflict_token"
	ConflictTokenFixtureSecretRefName  = "revisions-conflict-token-fixture-v1"
	ConflictTokenFixtureRuntimeMarker  = "harness-owned"
	ConflictTokenFixtureRuntimeEnvName = "CARTULARY_REVISIONS_CONFLICT_TOKEN_FIXTURE_RUNTIME_MARKER"
)

var (
	conflictTokenKeyIDPattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	conflictTokenTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?Z$`)
	fixtureConflictTokenKeySHA256 = func() [32]byte {
		fixtureKey := sha256.Sum256([]byte("cartulary-revisions-conflict-fixture"))
		return sha256.Sum256(fixtureKey[:])
	}()
)

type ConflictTokenKeyRing struct {
	activeKeyID string
	keys        map[string]conflictTokenKeyMaterial
}

type conflictTokenKeyMaterial struct {
	key           []byte
	state         string
	deactivatedAt *time.Time
	retireAt      *time.Time
}

type conflictTokenSecretRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type conflictTokenKeyRingManifest struct {
	SchemaID  string                     `json:"schema_id"`
	Algorithm string                     `json:"algorithm"`
	Keys      []conflictTokenKeyManifest `json:"keys"`
}

type conflictTokenKeyManifest struct {
	KeyID         string                 `json:"conflict_token_key_id"`
	State         string                 `json:"state"`
	SecretRef     conflictTokenSecretRef `json:"secret_ref"`
	DeactivatedAt string                 `json:"deactivated_at,omitempty"`
	RetireAt      string                 `json:"retire_at,omitempty"`
}

type KeyRingParseOptions struct {
	AllowFixtureKey bool
}

func ParseConflictTokenKeyRing(raw []byte, env map[string]string, now time.Time, options KeyRingParseOptions) (*ConflictTokenKeyRing, error) {
	registry := secretpurpose.NewRegistry()
	return ParseConflictTokenKeyRingWithRegistry(raw, env, now, registry, options)
}

func ParseConflictTokenKeyRingWithRegistry(raw []byte, env map[string]string, now time.Time, registry *secretpurpose.Registry, options KeyRingParseOptions) (*ConflictTokenKeyRing, error) {
	if env == nil {
		return nil, conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_secret_missing", "explicit Revisions conflict-token environment is required")
	}
	if len(raw) == 0 || len(raw) > int(ConflictTokenKeyRingManifestMaximumSize) || !utf8.Valid(raw) {
		return nil, conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_manifest_invalid", "Revisions conflict-token key-ring manifest must be non-empty UTF-8 JSON no larger than 65536 bytes")
	}
	if err := validateClosedConflictTokenKeyRingJSON(raw); err != nil {
		return nil, conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_manifest_invalid", err.Error())
	}
	var manifest conflictTokenKeyRingManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_manifest_invalid", "parse Revisions conflict-token key-ring manifest")
	}
	if manifest.SchemaID != conflictTokenKeyRingSchemaID || manifest.Algorithm != conflictTokenKeyRingAlgorithm || len(manifest.Keys) < 1 || len(manifest.Keys) > conflictTokenKeyRingMaximumKeys {
		return nil, conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_manifest_invalid", "manifest must use the supported schema and algorithm and contain 1..8 keys")
	}
	if registry == nil {
		return nil, conflictTokenConfigError(conflictTokenKeyRingConfigPath, "revisions_conflict_token_key_purpose_conflict", "startup secret-purpose registry is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	ring := &ConflictTokenKeyRing{keys: make(map[string]conflictTokenKeyMaterial, len(manifest.Keys))}
	seenRefs := make(map[string]struct{}, len(manifest.Keys))
	seenMaterials := make(map[[32]byte]struct{}, len(manifest.Keys))
	active := 0
	for index, entry := range manifest.Keys {
		path := fmt.Sprintf("%s.keys[%d]", conflictTokenKeyRingConfigPath, index)
		if !conflictTokenKeyIDPattern.MatchString(entry.KeyID) {
			return nil, conflictTokenConfigError(path+".conflict_token_key_id", "revisions_conflict_token_key_invalid", "conflict_token_key_id is invalid")
		}
		if _, exists := ring.keys[entry.KeyID]; exists {
			return nil, conflictTokenConfigError(path+".conflict_token_key_id", "revisions_conflict_token_key_id_conflict", "conflict_token_key_id values must be unique")
		}
		if _, exists := seenRefs[entry.SecretRef.Name]; exists {
			return nil, conflictTokenConfigError(path+".secret_ref", "revisions_conflict_token_key_purpose_conflict", "secret references must be unique within the key ring")
		}
		key, err := resolveConflictTokenSecret(entry.SecretRef, env, path+".secret_ref")
		if err != nil {
			return nil, err
		}
		fingerprint := sha256.Sum256(key)
		if _, exists := seenMaterials[fingerprint]; exists {
			return nil, conflictTokenConfigError(path+".secret_ref", "revisions_conflict_token_key_purpose_conflict", "key material must be unique within the key ring")
		}
		if fingerprint == fixtureConflictTokenKeySHA256 && !options.AllowFixtureKey {
			return nil, conflictTokenConfigError(path+".secret_ref", "revisions_conflict_token_fixture_key_forbidden", "fixture conflict-token key material is forbidden outside a harness-owned runtime")
		}
		if err := registry.Register(entry.SecretRef.Name, conflictTokenSecretPurpose, key); err != nil {
			return nil, conflictTokenConfigError(path+".secret_ref", "revisions_conflict_token_key_purpose_conflict", "secret reference or key material is registered for another startup purpose")
		}
		material := conflictTokenKeyMaterial{key: key, state: entry.State}
		switch entry.State {
		case conflictTokenKeyStateActive:
			active++
			if entry.DeactivatedAt != "" || entry.RetireAt != "" {
				return nil, conflictTokenConfigError(path, "revisions_conflict_token_rotation_invalid", "active key must omit deactivated_at and retire_at")
			}
			ring.activeKeyID = entry.KeyID
		case conflictTokenKeyStateDecryptOnly:
			deactivatedAt, err := parseConflictTokenTimestamp(entry.DeactivatedAt)
			if err != nil {
				return nil, conflictTokenConfigError(path+".deactivated_at", "revisions_conflict_token_rotation_invalid", err.Error())
			}
			retireAt, err := parseConflictTokenTimestamp(entry.RetireAt)
			if err != nil || retireAt.Before(deactivatedAt.Add(conflictTokenTTL+conflictTokenClockSkew)) || !now.Before(retireAt) {
				return nil, conflictTokenConfigError(path+".retire_at", "revisions_conflict_token_rotation_invalid", "retire_at must be at least 31 minutes after deactivated_at and later than startup time")
			}
			material.deactivatedAt = &deactivatedAt
			material.retireAt = &retireAt
		default:
			return nil, conflictTokenConfigError(path+".state", "revisions_conflict_token_rotation_invalid", "key state must be active or decrypt_only")
		}
		ring.keys[entry.KeyID] = material
		seenRefs[entry.SecretRef.Name] = struct{}{}
		seenMaterials[fingerprint] = struct{}{}
	}
	if active != 1 {
		return nil, conflictTokenConfigError(conflictTokenKeyRingConfigPath+".keys", "revisions_conflict_token_rotation_invalid", "key ring must contain exactly one active key")
	}
	return ring, nil
}

func resolveConflictTokenSecret(ref conflictTokenSecretRef, env map[string]string, path string) ([]byte, error) {
	if ref.Kind != "env" || !conflictTokenKeyIDPattern.MatchString(ref.Name) {
		return nil, conflictTokenConfigError(path, "revisions_conflict_token_secret_missing", "secret_ref must contain kind env and a valid name")
	}
	name := "CARTULARY_SECRET_" + normalizedConflictTokenSecretSuffix(ref.Name)
	value, ok := env[name]
	if !ok || value == "" || strings.Contains(value, "=") {
		return nil, conflictTokenConfigError(path, "revisions_conflict_token_secret_missing", "secret_ref could not be resolved to unpadded base64url key material")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return nil, conflictTokenConfigError(path, "revisions_conflict_token_key_invalid", "resolved key material must be unpadded base64url encoding of exactly 32 bytes")
	}
	return append([]byte(nil), decoded...), nil
}

func parseConflictTokenTimestamp(value string) (time.Time, error) {
	if !conflictTokenTimestampPattern.MatchString(value) {
		return time.Time{}, errors.New("timestamp must use canonical UTC timestamp_utc_v1 syntax")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC {
		return time.Time{}, errors.New("timestamp is invalid")
	}
	return parsed.UTC(), nil
}

func validateClosedConflictTokenKeyRingJSON(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := scanClosedConflictTokenJSONValue(decoder); err != nil {
		return err
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("revisions conflict-token key-ring manifest contains trailing JSON token %v", token)
	}
	return nil
}

func scanClosedConflictTokenJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("explicit null is invalid in Revisions conflict-token key-ring manifests")
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			member, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := member.(string)
			if !ok {
				return errors.New("key-ring object member is not a string")
			}
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate Revisions conflict-token key-ring member %q", name)
			}
			seen[name] = struct{}{}
			if err := scanClosedConflictTokenJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanClosedConflictTokenJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return nil
	}
}

func normalizedConflictTokenSecretSuffix(name string) string {
	var builder strings.Builder
	underscore := false
	for _, value := range name {
		switch {
		case value >= 'a' && value <= 'z':
			builder.WriteRune(value - ('a' - 'A'))
			underscore = false
		case value >= 'A' && value <= 'Z', value >= '0' && value <= '9':
			builder.WriteRune(value)
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

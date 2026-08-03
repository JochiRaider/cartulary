package conflicts_test

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
)

func TestConflictTokenV3SealsClaimsAndRejectsInvalidTokens(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 123000000, time.UTC)
	codec := testCodec(t, now, "active.key", "active-key", "active", "", "", nil)
	claims := validClaims()
	token, err := codec.Issue(claims)
	if err != nil {
		t.Fatalf("issue conflict token: %v", err)
	}
	if !strings.HasPrefix(token, "cft3.active.key.") || len(token) > conflicts.ConflictTokenMaximumLen {
		t.Fatalf("unexpected v3 wire token: %q", token)
	}
	for _, plaintext := range []string{claims.RouteKey, claims.RecordID, claims.ViewSchemaID, claims.FieldKey, claims.RequestHash} {
		if strings.Contains(token, plaintext) {
			t.Fatalf("sealed token disclosed claim %q", plaintext)
		}
	}
	parsed, ok := codec.Parse(token)
	if !ok || parsed.Version != conflicts.ConflictTokenVersion || !parsed.IssuedAt.Equal(now) || !parsed.ExpiresAt.Equal(now.Add(conflicts.ConflictTokenTTL)) {
		t.Fatalf("unexpected parsed claims: ok=%v claims=%#v", ok, parsed)
	}
	binding := bindingFor(parsed)
	if !parsed.ValidFor(binding) {
		t.Fatal("issued token did not satisfy its exact binding")
	}
	for name, mutate := range map[string]func(conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding{
		"route": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.RouteKey += ".other"
			return value
		},
		"record": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.RecordID = uuid.NewString()
			return value
		},
		"view": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.ViewSchemaID += ".other"
			return value
		},
		"field": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.FieldKey += ".other"
			return value
		},
		"class": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.ConflictResolutionClass = "atomic_replace"
			return value
		},
		"base_version": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.BaseRowVersion++
			return value
		},
		"current_version": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.CurrentRowVersion++
			return value
		},
		"request": func(value conflicts.ConflictTokenBinding) conflicts.ConflictTokenBinding {
			value.RequestHash += "x"
			return value
		},
	} {
		t.Run("wrong_binding_"+name, func(t *testing.T) {
			if parsed.ValidFor(mutate(binding)) {
				t.Fatalf("token accepted wrong %s binding", name)
			}
		})
	}
	second, err := codec.Issue(claims)
	if err != nil || second == token {
		t.Fatalf("fresh nonce not reflected in second token: err=%v", err)
	}
	tampered := token[:len(token)-1] + alternateTokenByte(token[len(token)-1])
	for name, candidate := range map[string]string{
		"tampered":    tampered,
		"truncated":   token[:len(token)-1],
		"v2":          base64.RawURLEncoding.EncodeToString([]byte(`{"cartulary_conflict_token_v":2}`)),
		"too_long":    strings.Repeat("x", conflicts.ConflictTokenMaximumLen+1),
		"unknown_key": strings.Replace(token, "active.key", "unknown", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := codec.Parse(candidate); ok {
				t.Fatalf("invalid %s token parsed", name)
			}
		})
	}
}

func TestConflictTokenV3ExpiryRotationAndClockSkew(t *testing.T) {
	issuedAt := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	oldCodec := testCodec(t, issuedAt, "old", "old-secret", "active", "", "", nil)
	token, err := oldCodec.Issue(validClaims())
	if err != nil {
		t.Fatalf("issue old token: %v", err)
	}
	deactivatedAt := issuedAt.Add(time.Minute)
	retireAt := deactivatedAt.Add(31 * time.Minute)
	clock := issuedAt.Add(2 * time.Minute)
	rotated := testCodecWithManifest(t, &clock, []manifestKey{
		{id: "new", ref: "new-secret", state: "active"},
		{id: "old", ref: "old-secret", state: "decrypt_only", deactivatedAt: deactivatedAt.Format(time.RFC3339), retireAt: retireAt.Format(time.RFC3339)},
	})
	if _, ok := rotated.Parse(token); !ok {
		t.Fatal("decrypt-only key rejected token issued before deactivation")
	}
	clock = issuedAt.Add(conflicts.ConflictTokenTTL)
	if _, ok := rotated.Parse(token); ok {
		t.Fatal("expired token parsed")
	}
	clock = retireAt
	if _, ok := rotated.Parse(token); ok {
		t.Fatal("retired key parsed")
	}

	issuerClock := issuedAt
	futureCodec := testCodecWithManifest(t, &issuerClock, []manifestKey{{id: "future", ref: "future-secret", state: "active"}})
	futureToken, err := futureCodec.Issue(validClaims())
	if err != nil {
		t.Fatalf("issue future token: %v", err)
	}
	verifyClock := issuedAt.Add(-conflicts.ConflictTokenClockSkew)
	verifier := testCodecWithManifest(t, &verifyClock, []manifestKey{{id: "future", ref: "future-secret", state: "active"}})
	if _, ok := verifier.Parse(futureToken); !ok {
		t.Fatal("maximum positive clock skew was not admitted")
	}
	verifyClock = verifyClock.Add(-time.Nanosecond)
	if _, ok := verifier.Parse(futureToken); ok {
		t.Fatal("excessive positive clock skew was admitted")
	}
}

func TestConflictTokenV3PropagatesEntropyFailure(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	ring := testRing(t, now, []manifestKey{{id: "active", ref: "active", state: "active"}}, nil, conflicts.KeyRingParseOptions{})
	codec, err := conflicts.NewConflictTokenCodec(ring, conflicts.WithClock(func() time.Time { return now }), conflicts.WithEntropySource(errorReader{}))
	if err != nil {
		t.Fatalf("construct codec: %v", err)
	}
	if _, err := codec.Issue(validClaims()); !errors.Is(err, conflicts.ErrConflictTokenUnavailable) {
		t.Fatalf("entropy failure = %v, want closed unavailable error", err)
	}
}

func TestConflictTokenKeyRingValidationAndSecretIsolation(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	valid := []manifestKey{{id: "active", ref: "active", state: "active"}}
	if _, err := conflicts.ParseConflictTokenKeyRing([]byte(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v1","algorithm":"aes_256_gcm_v1","keys":[]}`), nil, now, conflicts.KeyRingParseOptions{}); err == nil {
		t.Fatal("empty key ring was admitted")
	}
	if _, err := conflicts.ParseConflictTokenKeyRing([]byte(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v1","algorithm":"aes_256_gcm_v1","keys":[],"unknown":true}`), nil, now, conflicts.KeyRingParseOptions{}); err == nil {
		t.Fatal("unknown manifest member was admitted")
	}
	if _, err := conflicts.ParseConflictTokenKeyRing([]byte(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v1","schema_id":"cartulary.revisions_conflict_token_key_ring.v1","algorithm":"aes_256_gcm_v1","keys":[]}`), nil, now, conflicts.KeyRingParseOptions{}); err == nil {
		t.Fatal("duplicate manifest member was admitted")
	}
	for name, keys := range map[string][]manifestKey{
		"duplicate_id":    {{id: "same", ref: "one", state: "active"}, {id: "same", ref: "two", state: "decrypt_only", deactivatedAt: now.Add(-time.Minute).Format(time.RFC3339), retireAt: now.Add(30 * time.Minute).Format(time.RFC3339)}},
		"duplicate_ref":   {{id: "one", ref: "same", state: "active"}, {id: "two", ref: "same", state: "decrypt_only", deactivatedAt: now.Add(-time.Minute).Format(time.RFC3339), retireAt: now.Add(30 * time.Minute).Format(time.RFC3339)}},
		"multiple_active": {{id: "one", ref: "one", state: "active"}, {id: "two", ref: "two", state: "active"}},
		"expired_old":     {{id: "one", ref: "one", state: "active"}, {id: "old", ref: "old", state: "decrypt_only", deactivatedAt: now.Add(-time.Hour).Format(time.RFC3339), retireAt: now.Add(-time.Minute).Format(time.RFC3339)}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := testRingError(now, keys, nil, conflicts.KeyRingParseOptions{}); err == nil {
				t.Fatalf("invalid %s ring was admitted", name)
			}
		})
	}
	registry := secretpurpose.NewRegistry()
	activeKey := keyFor("active")
	if err := registry.Register("authentication", "authentication_master", activeKey); err != nil {
		t.Fatalf("register competing purpose: %v", err)
	}
	if _, err := testRingError(now, valid, registry, conflicts.KeyRingParseOptions{}); err == nil || !strings.Contains(err.Error(), "purpose_conflict") {
		t.Fatalf("cross-purpose material reuse error = %v", err)
	}
	fixtureKey := sha256.Sum256([]byte("cartulary-revisions-conflict-fixture"))
	fixtureManifest := []manifestKey{{id: "fixture", ref: conflicts.ConflictTokenFixtureSecretRefName, state: "active", key: fixtureKey[:]}}
	if _, err := testRingError(now, fixtureManifest, nil, conflicts.KeyRingParseOptions{}); err == nil || !strings.Contains(err.Error(), "fixture_key_forbidden") {
		t.Fatalf("fixture key outside harness error = %v", err)
	}
	if _, err := testRingError(now, fixtureManifest, nil, conflicts.KeyRingParseOptions{AllowFixtureKey: true}); err != nil {
		t.Fatalf("harness-owned fixture key rejected: %v", err)
	}
}

func TestConflictTokenConfigurationRequiresSecureManifestPath(t *testing.T) {
	for _, configuration := range []conflicts.Configuration{
		{},
		{ConflictTokenKeyRingManifestPath: "relative/key-ring.json"},
		{ConflictTokenKeyRingManifestPath: "/etc/cartulary/../key-ring.json"},
	} {
		if _, findings := conflicts.NormalizeAndValidateConfiguration(configuration); len(findings) == 0 {
			t.Fatalf("invalid configuration admitted: %#v", configuration)
		}
	}
	normalized, findings := conflicts.NormalizeAndValidateConfiguration(conflicts.Configuration{ConflictTokenKeyRingManifestPath: "/etc/cartulary/key-ring.json"})
	if len(findings) != 0 || normalized.ConflictTokenKeyRingManifestPath != "/etc/cartulary/key-ring.json" {
		t.Fatalf("valid configuration rejected: normalized=%#v findings=%#v", normalized, findings)
	}
}

type manifestKey struct {
	id            string
	ref           string
	state         string
	deactivatedAt string
	retireAt      string
	key           []byte
}

func testCodec(t testing.TB, now time.Time, id string, ref string, state string, deactivatedAt string, retireAt string, registry *secretpurpose.Registry) conflicts.ConflictTokenCodec {
	t.Helper()
	clock := now
	return testCodecWithManifest(t, &clock, []manifestKey{{id: id, ref: ref, state: state, deactivatedAt: deactivatedAt, retireAt: retireAt}})
}

func testCodecWithManifest(t testing.TB, now *time.Time, keys []manifestKey) conflicts.ConflictTokenCodec {
	t.Helper()
	ring := testRing(t, now.UTC(), keys, nil, conflicts.KeyRingParseOptions{})
	codec, err := conflicts.NewConflictTokenCodec(ring, conflicts.WithClock(func() time.Time { return now.UTC() }))
	if err != nil {
		t.Fatalf("construct conflict token codec: %v", err)
	}
	return codec
}

func testRing(t testing.TB, now time.Time, keys []manifestKey, registry *secretpurpose.Registry, options conflicts.KeyRingParseOptions) *conflicts.ConflictTokenKeyRing {
	t.Helper()
	ring, err := testRingError(now, keys, registry, options)
	if err != nil {
		t.Fatalf("parse test key ring: %v", err)
	}
	return ring
}

func testRingError(now time.Time, keys []manifestKey, registry *secretpurpose.Registry, options conflicts.KeyRingParseOptions) (*conflicts.ConflictTokenKeyRing, error) {
	entries := make([]string, 0, len(keys))
	env := make(map[string]string, len(keys))
	for _, entry := range keys {
		rotation := ""
		if entry.deactivatedAt != "" {
			rotation += fmt.Sprintf(`,"deactivated_at":%q`, entry.deactivatedAt)
		}
		if entry.retireAt != "" {
			rotation += fmt.Sprintf(`,"retire_at":%q`, entry.retireAt)
		}
		entries = append(entries, fmt.Sprintf(`{"conflict_token_key_id":%q,"state":%q,"secret_ref":{"kind":"env","name":%q}%s}`, entry.id, entry.state, entry.ref, rotation))
		key := entry.key
		if len(key) == 0 {
			key = keyFor(entry.ref)
		}
		env["CARTULARY_SECRET_"+secretSuffix(entry.ref)] = base64.RawURLEncoding.EncodeToString(key)
	}
	manifest := []byte(fmt.Sprintf(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v1","algorithm":"aes_256_gcm_v1","keys":[%s]}`, strings.Join(entries, ",")))
	if registry == nil {
		return conflicts.ParseConflictTokenKeyRing(manifest, env, now, options)
	}
	return conflicts.ParseConflictTokenKeyRingWithRegistry(manifest, env, now, registry, options)
}

func validClaims() conflicts.ConflictTokenClaims {
	return conflicts.ConflictTokenClaims{
		RouteKey:                "workbook.records.conflicts.resolve",
		RecordID:                uuid.NewString(),
		ViewSchemaID:            "cartulary.view.notes.v1",
		FieldKey:                "note.title",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             conflicts.RequestHashTokenValue([]byte("request")),
	}
}

func bindingFor(claims conflicts.ConflictTokenClaims) conflicts.ConflictTokenBinding {
	return conflicts.ConflictTokenBinding{
		RouteKey: claims.RouteKey, RecordID: claims.RecordID, ViewSchemaID: claims.ViewSchemaID,
		FieldKey: claims.FieldKey, ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion: claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion, RequestHash: claims.RequestHash,
	}
}

func keyFor(value string) []byte {
	key := sha256.Sum256([]byte("revisions-conflict-token-test:" + value))
	return key[:]
}

func secretSuffix(value string) string {
	return strings.ToUpper(strings.NewReplacer("-", "_", ".", "_", "_", "_").Replace(value))
}

func alternateTokenByte(value byte) string {
	if value == 'A' {
		return "B"
	}
	return "A"
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("entropy source failed") }

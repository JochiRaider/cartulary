package networkflow

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	testCursorKeyText = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"
	testSafeKeyText   = "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"
)

func TestNetworkFlowKeyRingsAndCursorRotation(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	manifest := `{
  "schema_id":"cartulary.network_flow_key_rings.v1",
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[
    {"cursor_key_id":"cursor-v2","state":"active","secret_ref":{"kind":"env","name":"cursor-active"}},
    {"cursor_key_id":"cursor-v1","state":"decrypt_only","secret_ref":{"kind":"env","name":"cursor-old"},"deactivated_at":"2026-07-13T11:55:00Z","retire_at":"2026-07-13T12:10:00Z"}
  ]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[
    {"safe_digest_key_id":"safe-v2","state":"active","secret_ref":{"kind":"env","name":"safe-active"}}
  ]}
}`
	rings, err := ParseKeyRings([]byte(manifest), map[string]string{
		"CARTULARY_SECRET_CURSOR_ACTIVE": testCursorKeyText,
		"CARTULARY_SECRET_CURSOR_OLD":    "AwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwM",
		"CARTULARY_SECRET_SAFE_ACTIVE":   testSafeKeyText,
	}, now)
	if err != nil {
		t.Fatalf("parse key rings: %v", err)
	}
	clock := now
	codec, err := newCursorCodec(rings, func() time.Time { return clock })
	if err != nil {
		t.Fatalf("create cursor codec: %v", err)
	}
	queryEcho := json.RawMessage(`{"filters":[],"sort":[]}`)
	token, err := codec.Encode(CursorBinding{
		Route: "nf.rows.query", ActorUserID: "actor", SessionID: "session", IncidentID: "incident",
		Scope: map[string]string{"table_ids": "nft_a"}, QueryHash: "hash", QueryEcho: queryEcho, Limit: 25,
	}, "row_keyset_v1", map[string]any{"network_flow_row_id": "nfr_a"})
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}
	if !strings.HasPrefix(token, "nfc2.cursor-v2.") {
		t.Fatalf("unexpected cursor envelope: %q", token)
	}
	payload, reason := codec.Decode(token)
	if reason != "" || payload.PositionKind != "row_keyset_v1" {
		t.Fatalf("decode cursor: reason=%q payload=%#v", reason, payload)
	}
	if reason := payload.Validate(CursorBinding{
		Route: "nf.rows.query", ActorUserID: "different-actor", SessionID: "session", IncidentID: "incident",
		Scope: map[string]string{"table_ids": "nft_a"}, QueryHash: "hash", QueryEcho: queryEcho, Limit: 25,
	}); reason != "actor_mismatch" {
		t.Fatalf("actor mismatch reason = %q", reason)
	}
	parts := strings.Split(token, ".")
	if _, reason := codec.Decode("nfc2.unknown." + parts[2]); reason != "malformed" {
		t.Fatalf("unknown key reason = %q", reason)
	}
	sealed, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode test cursor bytes: %v", err)
	}
	sealed[len(sealed)-1] ^= 1
	if _, reason := codec.Decode(parts[0] + "." + parts[1] + "." + base64.RawURLEncoding.EncodeToString(sealed)); reason != "malformed" {
		t.Fatalf("tampered cursor reason = %q", reason)
	}

	oldMaterial := rings.cursorKeys["cursor-v1"]
	oldMaterial.state = "active"
	oldMaterial.deactivatedAt = nil
	oldMaterial.retireAt = nil
	oldClock := now.Add(-6 * time.Minute)
	oldCodec, err := newCursorCodec(&KeyRings{
		cursorActiveID: "cursor-v1",
		cursorKeys:     map[string]cursorKeyMaterial{"cursor-v1": oldMaterial},
	}, func() time.Time { return oldClock })
	if err != nil {
		t.Fatalf("create prior cursor codec: %v", err)
	}
	oldToken, err := oldCodec.Encode(CursorBinding{
		Route: "nf.rows.query", ActorUserID: "actor", SessionID: "session", IncidentID: "incident",
		Scope: map[string]string{"table_ids": "nft_a"}, QueryHash: "hash", QueryEcho: queryEcho, Limit: 25,
	}, "row_keyset_v1", map[string]any{"network_flow_row_id": "nfr_old"})
	if err != nil {
		t.Fatalf("encode prior cursor: %v", err)
	}
	if _, reason := codec.Decode(oldToken); reason != "" {
		t.Fatalf("decrypt-only cursor before retirement reason = %q", reason)
	}
	clock = now.Add(10 * time.Minute)
	if _, reason := codec.Decode(oldToken); reason != "malformed" {
		t.Fatalf("retired cursor reason = %q", reason)
	}
	if _, exists := codec.keys["cursor-v1"]; exists {
		t.Fatal("retired cursor key remained in live codec")
	}
	clock = now
	clock = now.Add(cursorTTL)
	if _, reason := codec.Decode(token); reason != "expired" {
		t.Fatalf("expected equality with expiry to be expired, got %q", reason)
	}
}

func TestNetworkFlowSafeDigestRingPurgesInactiveEpoch(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	manifest := `{
  "schema_id":"cartulary.network_flow_key_rings.v1",
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[
    {"cursor_key_id":"cursor-v2","state":"active","secret_ref":{"kind":"env","name":"cursor-active"}}
  ]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[
    {"safe_digest_key_id":"safe-v2","state":"active","secret_ref":{"kind":"env","name":"safe-active"}},
    {"safe_digest_key_id":"safe-v1","state":"inactive","secret_ref":{"kind":"env","name":"safe-old"},"deactivated_at":"2026-07-13T11:50:00Z","retain_until":"2026-07-13T12:10:00Z"}
  ]}
}`
	rings, err := ParseKeyRings([]byte(manifest), map[string]string{
		"CARTULARY_SECRET_CURSOR_ACTIVE": testCursorKeyText,
		"CARTULARY_SECRET_SAFE_ACTIVE":   testSafeKeyText,
		"CARTULARY_SECRET_SAFE_OLD":      "BAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ",
	}, now)
	if err != nil {
		t.Fatalf("parse key rings: %v", err)
	}
	clock := now
	digester, err := newSafeDigester(rings, func() time.Time { return clock })
	if err != nil {
		t.Fatalf("create safe digester: %v", err)
	}
	digest, keyID, err := digester.Digest("source_filename", "flows.csv")
	if err != nil || digest == "" || keyID != "safe-v2" {
		t.Fatalf("active safe digest = %q, %q, %v", digest, keyID, err)
	}
	clock = now.Add(10 * time.Minute)
	if _, _, err := digester.Digest("source_filename", "flows.csv"); err != nil {
		t.Fatalf("active safe digest after old retention: %v", err)
	}
	if _, exists := rings.safeKeys["safe-v1"]; exists {
		t.Fatal("expired inactive safe-digest key remained in live provider")
	}
}

func TestNetworkFlowKeyRingValidationRejectsPurposeReuseAndNull(t *testing.T) {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	t.Run("purpose reuse", func(t *testing.T) {
		manifest := `{
  "schema_id":"cartulary.network_flow_key_rings.v1",
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[{"cursor_key_id":"cursor","state":"active","secret_ref":{"kind":"env","name":"shared"}}]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[{"safe_digest_key_id":"safe","state":"active","secret_ref":{"kind":"env","name":"shared"}}]}
}`
		if _, err := ParseKeyRings([]byte(manifest), map[string]string{"CARTULARY_SECRET_SHARED": testCursorKeyText}, now); err == nil || !strings.Contains(err.Error(), "purpose_conflict") {
			t.Fatalf("expected purpose conflict, got %v", err)
		}
	})
	t.Run("authentication master material reuse", func(t *testing.T) {
		manifest := `{
  "schema_id":"cartulary.network_flow_key_rings.v1",
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[{"cursor_key_id":"cursor","state":"active","secret_ref":{"kind":"env","name":"cursor"}}]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[{"safe_digest_key_id":"safe","state":"active","secret_ref":{"kind":"env","name":"safe"}}]}
}`
		env := map[string]string{
			authn.AuthMasterKeyEnv:    base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)),
			"CARTULARY_SECRET_CURSOR": testCursorKeyText,
			"CARTULARY_SECRET_SAFE":   testSafeKeyText,
		}
		if _, err := ParseKeyRings([]byte(manifest), env, now); err == nil || !strings.Contains(err.Error(), "network_flow_cursor_key_invalid") {
			t.Fatalf("expected authentication-purpose conflict, got %v", err)
		}
	})
	t.Run("explicit null", func(t *testing.T) {
		manifest := `{"schema_id":"cartulary.network_flow_key_rings.v1","cursor_key_ring":null,"safe_digest_key_ring":{}}`
		if _, err := ParseKeyRings([]byte(manifest), nil, now); err == nil || !strings.Contains(err.Error(), "explicit null") {
			t.Fatalf("expected explicit-null rejection, got %v", err)
		}
	})
}

package harnessredact

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStringUsesSharedHarnessRules(t *testing.T) {
	input := strings.Join([]string{
		"postgres://cartulary:supersecret@127.0.0.1:5432/postgres?sslmode=disable",
		"password=supersecret",
		"Authorization: Bearer abc.def",
		"-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
	}, "\n")

	got := String(input)
	for _, secret := range []string{"supersecret", "abc.def", "BEGIN PRIVATE KEY", "END PRIVATE KEY"} {
		if strings.Contains(got, secret) {
			t.Fatalf("redacted output still contains %q: %s", secret, got)
		}
	}
	if !strings.Contains(got, "[REDACTED") {
		t.Fatalf("redacted output must contain replacement marker: %s", got)
	}
}

func TestValueRedactsStructuredKeysAndCLIArgs(t *testing.T) {
	input := map[string]any{
		"Authorization":  "Bearer nested.secret",
		"Cookie":         "session=nested-cookie",
		"args":           []any{"--token", "route-secret", "--dsn=postgres://cartulary:secret@127.0.0.1:5432/postgres"},
		"session_target": "not-redacted-token-substring",
		"service_sessions": []any{
			map[string]any{
				"target":            "service-timing-suite",
				"session_target":    "browser-stage-token-name",
				"cleanup_status":    "pass",
				"setup_duration_ms": float64(12),
				"healthy":           true,
				"absent":            nil,
				"session_token":     "nested-session-token",
			},
		},
		"nested": map[string]any{
			"CARTULARY_S3TEST_SECRET_ACCESS_KEY": "object-store-secret",
		},
	}

	got := Value(input, "").(map[string]any)
	sessions, ok := got["service_sessions"].([]any)
	if !ok || len(sessions) != 1 {
		t.Fatalf("service_sessions must remain an array after redaction: %#v", got["service_sessions"])
	}
	session := sessions[0].(map[string]any)
	if session["target"] != "service-timing-suite" || session["cleanup_status"] != "pass" {
		t.Fatalf("service session structural fields were redacted: %#v", session)
	}
	if session["session_target"] != "browser-stage-token-name" || session["setup_duration_ms"] != float64(12) || session["healthy"] != true || session["absent"] != nil {
		t.Fatalf("service session shape or scalar values were not preserved: %#v", session)
	}
	if got["session_target"] != "not-redacted-token-substring" {
		t.Fatalf("substring-only structural key was redacted: %#v", got)
	}
	raw := strings.Join([]string{
		got["Authorization"].(string),
		got["Cookie"].(string),
		strings.Join([]string{got["args"].([]any)[0].(string), got["args"].([]any)[1].(string), got["args"].([]any)[2].(string)}, " "),
		got["nested"].(map[string]any)["CARTULARY_S3TEST_SECRET_ACCESS_KEY"].(string),
		session["session_token"].(string),
	}, "\n")
	for _, secret := range []string{"nested.secret", "nested-cookie", "route-secret", "object-store-secret", "secret@127.0.0.1", "nested-session-token"} {
		if strings.Contains(raw, secret) {
			t.Fatalf("structured redaction leaked %q: %#v", secret, got)
		}
	}
}

func TestStructuredStringDecodesJSONBeforeRedaction(t *testing.T) {
	got := StructuredString(`{"Authorization":"Bearer nested.secret","service_sessions":[{"target":"service-timing-suite","cleanup_status":"pass","setup_duration_ms":12}]}`)
	for _, secret := range []string{"nested.secret"} {
		if strings.Contains(got, secret) {
			t.Fatalf("structured string redaction leaked %q: %s", secret, got)
		}
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("decode structured redaction output: %v", err)
	}
	if decoded["Authorization"] != "[REDACTED]" {
		t.Fatalf("authorization was not redacted: %s", got)
	}
	sessions, ok := decoded["service_sessions"].([]any)
	if !ok || len(sessions) != 1 {
		t.Fatalf("service_sessions shape was not preserved: %s", got)
	}
	session := sessions[0].(map[string]any)
	if session["target"] != "service-timing-suite" || session["cleanup_status"] != "pass" || session["setup_duration_ms"] != float64(12) {
		t.Fatalf("service session structural fields were not preserved: %#v", session)
	}

	raw := StructuredString("Authorization: Bearer raw.secret")
	if strings.Contains(raw, "raw.secret") {
		t.Fatalf("raw fallback redaction leaked secret: %s", raw)
	}
}

func TestStringRedactsClosedRawFamilies(t *testing.T) {
	input := strings.Join([]string{
		"https://user:secret@example.test/path",
		"postgres://cartulary:supersecret@127.0.0.1:5432/postgres password=supersecret",
		"Authorization: Bearer abc.def.ghi",
		"X-Cartulary-Test-Route-Token: route-secret",
		"minio_secret_access_key=minio-secret",
		"-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
	}, "\n")
	got := String(input)
	for _, secret := range []string{"secret@example", "supersecret", "abc.def.ghi", "route-secret", "minio-secret", "BEGIN PRIVATE KEY"} {
		if strings.Contains(got, secret) {
			t.Fatalf("redacted output still contains %q: %s", secret, got)
		}
	}
}

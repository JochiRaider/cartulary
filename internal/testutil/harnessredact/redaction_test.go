package harnessredact

import (
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
		"Authorization": "Bearer nested.secret",
		"Cookie":        "session=nested-cookie",
		"args":          []any{"--token", "route-secret", "--dsn=postgres://cartulary:secret@127.0.0.1:5432/postgres"},
		"service_sessions": []any{
			map[string]any{
				"target":            "service-timing-suite",
				"cleanup_status":    "pass",
				"setup_duration_ms": float64(12),
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

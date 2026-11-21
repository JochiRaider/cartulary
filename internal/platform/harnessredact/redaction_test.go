package harnessredact

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestStringUsesSharedHarnessRules(t *testing.T) {
	t.Run("linear_on_adversarial_inputs", testLinearRedaction)

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
	if !strings.Contains(got, Marker) || strings.Contains(got, "REDACTED") {
		t.Fatalf("redacted output must contain only the replacement marker: %s", got)
	}
}

// TH-HARNESS-REQ-617: redaction is linear in its input. Adversarial runs of
// 4 MiB finish far inside a bound that a quadratic redactor cannot meet.
func testLinearRedaction(t *testing.T) {
	const size = 4 << 20
	fill := func(unit string) string { return strings.Repeat(unit, size/len(unit)+1)[:size] }
	for name, run := range map[string]string{
		"word_without_assignment":  fill("Password_Token_Session_9"),
		"unquoted_terms":           fill("passwordsecrettokencookiedsn"),
		"scheme_without_separator": fill("ab0+.-"),
		"token_without_dot":        fill("eyJ-_aZ9"),
	} {
		for context, input := range map[string]string{
			"raw":                run,
			"unterminated_quote": `"` + run,
			"json_value":         `{"message":"` + run + `","next":"x"}`,
			"json_key":           `{"` + run + `":"value"}`,
		} {
			started := time.Now()
			_ = String(input)
			_ = StructuredString(input)
			if elapsed := time.Since(started); elapsed >= 5*time.Second {
				t.Errorf("%s/%s took %s for a 4 MiB input", context, name, elapsed)
			}
		}
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

	// TH-HARNESS-REQ-604 key canonicalization, with independent expectations.
	for key, want := range map[string]string{
		"dbPassword":                   "DB_PASSWORD",
		"clientSecret":                 "CLIENT_SECRET",
		"api2Key":                      "API2_KEY",
		"HTTPServer":                   "HTTPSERVER",
		"X-Cartulary-Test-Route-Token": "X_CARTULARY_TEST_ROUTE_TOKEN",
		"__session..target__":          "SESSION_TARGET",
		"pässword":                     "P_SSWORD",
		"paßword":                      "PA_WORD",
		"ſecret":                       "ECRET",
		"tökenSecret":                  "T_KEN_SECRET",
		"--":                           "",
	} {
		if got := canonicalStructuredKey(key); got != want {
			t.Errorf("canonicalStructuredKey(%q) = %q, want %q", key, got, want)
		}
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
	if decoded["Authorization"] != Marker || !strings.Contains(got, `"Authorization":"`+Marker+`"`) {
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
	for _, tc := range []struct{ input, want string }{
		{"https://PrivateFixtureAlpha@example.test/path", "https://<redacted>@example.test/path"},
		{"https://PrivateUser:PrivatePassword@example.test:443/path?q=1", "https://<redacted>@example.test:443/path?q=1"},
		{"https://PrivateUser:@example.test", "https://<redacted>@example.test"},
		{"https://Private%3AToken@example.test", "https://<redacted>@example.test"},
		{"postgresql://PrivateUser:PrivatePassword@localhost/db", "postgresql://<redacted>@localhost/db"},
		{"https://example.test/path@item?q=a@b#c@d", "https://example.test/path@item?q=a@b#c@d"},
		{"https://example.test?q=a@b", "https://example.test?q=a@b"},
		{"https://example.test#c@d", "https://example.test#c@d"},
	} {
		for _, got := range []string{String(tc.input), StructuredString(tc.input), Value(tc.input, "").(string)} {
			if got != tc.want || String(got) != tc.want {
				t.Fatal("URL authority redaction changed noncredential components or retained userinfo")
			}
		}
	}
	for _, tc := range []struct{ name, input, secret string }{
		{"URL userinfo", "https://user:PrivateFixtureAlpha@example.test/path", "PrivateFixtureAlpha"},
		{"basic header", "Authorization: Basic Zml4dHVyZTpwYXNzd29yZA==", "Zml4dHVyZTpwYXNzd29yZA=="},
		{"JWT", "eyJhbGciOiJub25lIn0.eyJzdWIiOiJmaXh0dXJlIn0.c2ln", "eyJhbGciOiJub25lIn0"},
		{"DSN", "host=local password=PrivateFixtureBeta dbname=test", "PrivateFixtureBeta"},
		{"CLI", "--client-secret=PrivateFixtureGamma", "PrivateFixtureGamma"},
		{"object store", "AWS_SECRET_ACCESS_KEY=PrivateFixtureDelta", "PrivateFixtureDelta"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, got := range []string{String(tc.input), StructuredString(tc.input)} {
				if strings.Contains(got, tc.secret) || !strings.Contains(got, Marker) || strings.Contains(got, "REDACTED") {
					t.Fatal("required raw credential family was not redacted")
				}
			}
		})
	}
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

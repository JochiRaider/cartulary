//go:build linux

package harnesssession

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/harnessredact"
)

func TestSessionScan(t *testing.T) {
	t.Run("sealed_inputs_preserve_permissions_and_secret_enforcement", func(t *testing.T) {
		for _, candidate := range []struct{ name, contents string }{{"safe", "safe fixture bytes"}, {"syntax", "TOKEN=synthetic-private-value"}, {"registered", "registered-private-sentinel"}} {
			t.Run(candidate.name, func(t *testing.T) {
				contents := candidate.contents
				root := t.TempDir()
				if err := os.Chmod(root, 0700); err != nil {
					t.Fatal(err)
				}
				inputs := filepath.Join(root, "inputs")
				if err := os.Mkdir(inputs, 0700); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(inputs, 0700) })
				file := filepath.Join(inputs, "input.bin")
				if err := os.WriteFile(file, []byte(contents), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(file, 0400); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(inputs, 0500); err != nil {
					t.Fatal(err)
				}
				result, err := ScanRetained(root, []string{"registered-private-sentinel"})
				if contents == "safe fixture bytes" {
					if err != nil || result["status"] != "pass" {
						t.Fatal(result, err)
					}
					info, err := os.Stat(file)
					if err != nil || info.Mode().Perm() != 0400 {
						t.Fatal("sealed file permission changed", info, err)
					}
					actual, err := os.ReadFile(file)
					if err != nil || string(actual) != contents {
						t.Fatal("sealed bytes changed", err)
					}
				} else {
					if err == nil {
						t.Fatal("sealed secret escaped scan")
					}
					if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
						t.Fatal("unsafe bytes retained", err)
					}
				}
				info, err := os.Stat(inputs)
				if err != nil || info.Mode().Perm() != 0500 {
					t.Fatal("sealed directory permission changed", info, err)
				}
			})
		}
	})
	t.Run("private_candidate_is_safe_retained_input", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Chmod(root, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "fixture.png"), []byte("old-golden"), 0600); err != nil {
			t.Fatal(err)
		}
		if result, err := ScanRetained(root, nil); err != nil || result["status"] != "pass" {
			t.Fatal(result, err)
		}
	})
	t.Run("fragmentation_does_not_change_acceptance", func(t *testing.T) {
		for _, c := range []struct {
			text string
			safe bool
		}{
			{`{"tokens":"ordinary","secret":"<redacted>suffix"}`, true},
			{`{"tokens_token":"leak"}`, false},
			{`{"tokenS":"ordinary"}`, true},
			{`{"secret":"<redact"}`, false},
			{`{"secret":"<REDACTED>"}`, false},
			{`{"secret":""}`, true},
			{`{"secret":"unfinished`, true},
			{`TOKEN=<redacted>suffix`, true},
			{`TOKEN=<redact`, false},
			{`TOKENS=ordinary`, true},
			{`postgresql://user:<redacted>@host`, true},
			{`postgresql://user:<redact@host`, false},
			{`{"中文password":"秘密"}`, false},
			{"\xffsafe 中文", true},
			{strings.Repeat("x", 65530) + "private-sentinel-12345", false},
		} {
			for _, size := range []int{1, 2, 7, 257, 65536} {
				r := &fragmentReader{Reader: strings.NewReader(c.text), size: size}
				if err := scanContent(r, []string{"private-sentinel-12345"}); (err == nil) != c.safe {
					t.Errorf("size %d input %q: %v", size, c.text, err)
				}
			}
		}
		errRead := errors.New("reader failed")
		if err := scanContent(errorReader{errRead}, nil); !errors.Is(err, errRead) {
			t.Fatal("read failure lost", err)
		}
	})
	t.Run("long_unregistered_syntax", func(t *testing.T) {
		for _, input := range []string{
			`{"password":"` + strings.Repeat("x", 150000) + `"}`,
			`{"password` + strings.Repeat("x", 150000) + `":"value"}`,
			`{"password"` + strings.Repeat(" ", 150000) + `:"value"}`,
			`postgres://` + strings.Repeat("u", 150000) + `:value@host`,
		} {
			if err := scanContent(strings.NewReader(input), nil); err == nil {
				t.Error("unregistered secret escaped streaming scan")
			}
		}
	})
	t.Run("scan_independent_expectations", func(t *testing.T) {
		cases := []struct {
			name, text string
			want       bool
		}{{"plain", `{"cpu_tokens":"2","message":"<redacted>"}`, true}, {"redacted", `{"secret":"<redacted>"} PASSWORD=<redacted> postgres://u:<redacted>@h`, true}, {"token_substring", `{"tokens_token":"leak"}`, false}, {"json", `{"my_password":"leak"}`, false}, {"assignment", "TOKEN=leak", false}, {"url", "postgresql://user:leak@host", false}, {"registered", strings.Repeat("x", 65530) + "private-sentinel-12345", false}, {"unicode", "abc 中文\n", true}}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				a := t.TempDir()
				for _, p := range []string{a} {
					_ = os.Chmod(p, 0700)
					if err := os.WriteFile(filepath.Join(p, "data.json"), []byte(c.text), 0600); err != nil {
						t.Fatal(err)
					}
				}
				ours, err := ScanRetained(a, []string{"private-sentinel-12345"})
				if (err == nil) != c.want {
					t.Fatalf("native scan: %v", err)
				}
				if c.want && (ours["policy_digest"] != "sha256:eaa74efae69ea73a6e0f18823a94ccc70a19b242e2cb2880e446611c3084ee11" || ours["scanned_files"] != 1 || ours["status"] != "pass") {
					t.Fatal("scan result differs from frozen independent expectations", ours)
				}
			})
		}
	})
	// TH-HARNESS-AC-160: every manifest rule and every structured secret-key
	// class, in every context the retained-secret scan checks, yields exactly
	// its independently specified output, carries no secret, uses only the
	// single marker and passes the scan. The script redactor's matrix
	// (test-foundation-characterization.mjs) states the same expectations.
	t.Run("redaction_matrix_matches_expectations_and_passes_retained_scan", func(t *testing.T) {
		cases, failed := redactionMatrix(), []string{}
		// Retained logs keep a bounded tail; summarize every failed case last.
		defer func() {
			if len(failed) != 0 {
				t.Errorf("%d of %d redaction matrix cases failed: %s", len(failed), len(cases), strings.Join(failed, " "))
			}
		}()
		for _, c := range cases {
			passed := t.Run(c.name, func(t *testing.T) {
				got := harnessredact.String(c.input)
				if c.structured {
					got = harnessredact.StructuredString(c.input)
				}
				if got != c.want {
					t.Errorf("redacted output differs:\n got %q\nwant %q", got, c.want)
				}
				if c.secret != "" && strings.Contains(got, c.secret) {
					t.Errorf("redacted output retains its secret: %q", got)
				}
				if strings.Contains(strings.ToUpper(got), "REDACTED]") || strings.Contains(got, "REDACTED:") {
					t.Errorf("redacted output uses a marker other than %s: %q", harnessredact.Marker, got)
				}
				secrets := []string{}
				if c.secret != "" {
					secrets = append(secrets, c.secret)
				}
				if err := scanContent(strings.NewReader(got), secrets); err != nil {
					t.Errorf("redacted output fails the retained-secret scan: %q: %v", got, err)
				}
			})
			if !passed {
				failed = append(failed, c.name)
			}
		}
	})
	// The scan exempts only the single marker, never a bracketed or typed form.
	t.Run("scan_rejects_other_marker_forms", func(t *testing.T) {
		for _, text := range []string{
			`{"secret":"[REDACTED]"}`,
			`{"secret":"[REDACTED:token]"}`,
			`SECRET=[REDACTED:password]`,
			`postgres://u:[REDACTED:dsn]@h`,
			`{"password":"REDACTED"}`,
		} {
			if err := scanContent(strings.NewReader(text), nil); err == nil {
				t.Errorf("scan exempted another marker form: %q", text)
			}
		}
	})
	t.Run("scan_rejects_unsafe_entries", func(t *testing.T) {
		for _, kind := range []string{"symlink", "hardlink", "fifo", "public_file", "forbidden_name"} {
			t.Run(kind, func(t *testing.T) {
				root := t.TempDir()
				_ = os.Chmod(root, 0700)
				file := filepath.Join(root, "unsafe")
				switch kind {
				case "symlink":
					if err := os.Symlink("/dev/null", file); err != nil {
						t.Fatal(err)
					}
				case "hardlink":
					source := filepath.Join(t.TempDir(), "source")
					_ = os.WriteFile(source, []byte("safe"), 0600)
					if err := os.Link(source, file); err != nil {
						t.Fatal(err)
					}
				case "fifo":
					if err := syscall.Mkfifo(file, 0600); err != nil {
						t.Fatal(err)
					}
				case "public_file":
					if err := os.WriteFile(file, []byte("safe"), 0600); err != nil {
						t.Fatal(err)
					}
					// Creation modes are filtered by the caller's umask.
					// Establish the unsafe fixture explicitly under secure callers.
					if err := os.Chmod(file, 0644); err != nil {
						t.Fatal(err)
					}
				case "forbidden_name":
					_ = os.WriteFile(filepath.Join(root, "suite-lease.json"), []byte("safe"), 0600)
				}
				if _, err := ScanRetained(root, nil); err == nil {
					t.Fatal("unsafe entry accepted")
				}
			})
		}
	})
}

type fragmentReader struct {
	io.Reader
	size int
}

func (r *fragmentReader) Read(p []byte) (int, error) { return r.Reader.Read(p[:min(len(p), r.size)]) }

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

type redactionCase struct {
	name       string
	structured bool
	input      string
	want       string
	secret     string
}

// redactionMatrix is the AC-160 expectation table. Expected outputs are
// written by hand from TH-HARNESS-REQ-604/617; no redactor computes them.
func redactionMatrix() []redactionCase {
	const secret = "fixture-secret-1234"
	cases := []redactionCase{
		// Each manifest value rule in raw text.
		{name: "rule/bearer_token", input: "token type Bearer " + secret + " end", want: "token type Bearer <redacted> end", secret: secret},
		{name: "rule/basic_auth", input: "scheme Basic Zml4dHVyZS1zZWNyZXQtMTIzNA== end", want: "scheme Basic <redacted> end", secret: "Zml4dHVyZS1zZWNyZXQtMTIzNA=="},
		{name: "rule/authorization_header", input: "Authorization: Digest " + secret, want: "Authorization: <redacted>", secret: secret},
		{name: "rule/authorization_bearer_header", input: "Authorization: Bearer " + secret, want: "Authorization: <redacted>", secret: secret},
		{name: "rule/cookie_header", input: "Set-Cookie: sid=" + secret + "; Path=/", want: "Set-Cookie: <redacted>", secret: secret},
		{name: "rule/route_token_header", input: "X-Cartulary-Test-Route-Token: " + secret, want: "X-Cartulary-Test-Route-Token: <redacted>", secret: secret},
		{name: "rule/jwt", input: "jwt eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.fixturesignature1234 end", want: "jwt <redacted> end", secret: "fixturesignature1234"},
		{name: "rule/generic_url_userinfo", input: "https://user:" + secret + "@example.test/path", want: "https://<redacted>@example.test/path", secret: secret},
		{name: "rule/postgres_url_userinfo_query", input: "postgres://u:fixture?secret-1234@h/db", want: "postgres://u:<redacted>@h/db", secret: "fixture?secret-1234"},
		{name: "rule/postgres_url_userinfo_fragment", input: "postgres://u:fixture#secret-1234@h/db", want: "postgres://u:<redacted>@h/db", secret: "fixture#secret-1234"},
		{name: "rule/cli_secret_argument_equals", input: "--client-secret=" + secret, want: "--client-secret=<redacted>", secret: secret},
		{name: "rule/cli_secret_argument_space", input: "--token " + secret, want: "--token <redacted>", secret: secret},
		{name: "rule/secret_assignment", input: "DB_PASSWORD=" + secret, want: "DB_PASSWORD=<redacted>", secret: secret},
		{name: "rule/json_like_secret_assignment", input: `{"pwd":"` + secret + `"}`, want: `{"pwd":"<redacted>"}`, secret: secret},
		{name: "rule/private_key_block", input: "-----BEGIN PRIVATE KEY-----\n" + secret + "\n-----END PRIVATE KEY-----", want: "<redacted>", secret: secret},
		{name: "rule/json_member_closure", input: `{"database_dsn":"` + secret + `"}`, want: `{"database_dsn":"<redacted>"}`, secret: secret},
		{name: "rule/json_member_closure_url", input: `{"database_dsn":"postgres://u:` + secret + `@h"}`, want: `{"database_dsn":"<redacted>"}`, secret: secret},

		// Uppercase assignments with and without a name prefix.
		{name: "assignment/PASSWORD", input: "PASSWORD=" + secret, want: "PASSWORD=<redacted>", secret: secret},
		{name: "assignment/ACCESS_TOKEN", input: "ACCESS_TOKEN=" + secret, want: "ACCESS_TOKEN=<redacted>", secret: secret},
		{name: "assignment/CARTULARY_S3_ACCESS_KEY", input: "CARTULARY_S3_ACCESS_KEY=" + secret, want: "CARTULARY_S3_ACCESS_KEY=<redacted>", secret: secret},
		{name: "assignment/AWS_ACCESS_KEY_ID", input: "AWS_ACCESS_KEY_ID=" + secret, want: "AWS_ACCESS_KEY_ID=<redacted>", secret: secret},
		{name: "assignment/SESSION_SECRET", input: "SESSION_SECRET=" + secret, want: "SESSION_SECRET=<redacted>", secret: secret},
		{name: "assignment/TLS_PRIVATE_KEY", input: "TLS_PRIVATE_KEY=" + secret, want: "TLS_PRIVATE_KEY=<redacted>", secret: secret},
		{name: "assignment/COOKIE", input: "COOKIE=" + secret, want: "COOKIE=<redacted>", secret: secret},
		{name: "assignment/CARTULARY_DSN", input: "CARTULARY_DSN=postgres://u:" + secret + "@h/db", want: "CARTULARY_DSN=<redacted>", secret: secret},
		{name: "assignment/export_prefix", input: "export DB_PASSWORD=" + secret, want: "export DB_PASSWORD=<redacted>", secret: secret},
		{name: "assignment/lowercase_dsn_keyword", input: "host=db password=" + secret + " dbname=cartulary", want: "host=db password=<redacted> dbname=cartulary", secret: secret},
		{name: "assignment/spaced", input: "client_secret = " + secret, want: "client_secret=<redacted>", secret: secret},

		// SESSION is a raw-assignment credential name only.
		{name: "assignment/SESSION", input: "SESSION=" + secret, want: "SESSION=<redacted>", secret: secret},
		{name: "assignment/APP_SESSION", input: "APP_SESSION=" + secret, want: "APP_SESSION=<redacted>", secret: secret},
		{name: "assignment/export_session", input: "export CARTULARY_SESSION=" + secret, want: "export CARTULARY_SESSION=<redacted>", secret: secret},
		{name: "assignment/session_query", input: "https://example.test/cb?session=" + secret, want: "https://example.test/cb?session=<redacted>", secret: secret},
		// Explicit ASCII classes: a non-breaking space is part of the value.
		{name: "nbsp/assignment", input: "DB_PASSWORD=fixture\u00a0secret-1234", want: "DB_PASSWORD=<redacted>", secret: "secret-1234"},
		{name: "nbsp/cli_secret_argument_equals", input: "--token=fixture\u00a0secret-1234", want: "--token=<redacted>", secret: "secret-1234"},

		// URL userinfo and DSN passwords.
		{name: "url/postgres", input: "postgres://cartulary:" + secret + "@127.0.0.1:5432/db", want: "postgres://<redacted>@127.0.0.1:5432/db", secret: secret},
		{name: "url/postgresql_query", input: "postgresql://u:" + secret + "@h/db?sslmode=disable", want: "postgresql://<redacted>@h/db?sslmode=disable", secret: secret},
		{name: "url/postgres_query_at", input: "postgres://u:" + secret + "@h:5432?user=a@b", want: "postgres://<redacted>@h:<redacted>@b", secret: secret},

		// Values that are not secrets are preserved.
		{name: "preserve/SESSION_ID", input: "SESSION_ID=abc123", want: "SESSION_ID=abc123"},
		{name: "preserve/raw_session_member", input: `{"session":"browser-session-name"}`, want: `{"session":"browser-session-name"}`},
		{name: "preserve/TOKENS", input: "TOKENS=ordinary", want: "TOKENS=ordinary"},
		{name: "preserve/max_tokens", input: "max_tokens=5", want: "max_tokens=5"},
		{name: "preserve/url_at_outside_authority", input: "https://example.test/path@item?q=a@b#c@d", want: "https://example.test/path@item?q=a@b#c@d"},
		{name: "preserve/json_tokens", input: `{"tokens":"ordinary"}`, want: `{"tokens":"ordinary"}`},
	}

	// A JSON member under every structured secret-key class and every scan
	// term, including camelCase and prefixed keys, is replaced whole.
	for _, key := range []string{
		"password", "Password", "PASS", "pwd", "secret", "SECRET", "client_secret", "session_secret", "secret_ref",
		"dsn", "DSN", "database_dsn", "recovery_dsn", "dsn_ref", "private_key", "private_key_path", "api_key",
		"access_key", "secret_key", "jwt", "bearer", "Authorization", "Cookie", "Set-Cookie", "X-Cartulary-Test-Route-Token",
		"token", "session_token", "cookies", "tokenizer", "passwordless", "CARTULARY_S3TEST_ACCESS_KEY_ID",
		"dbPassword", "clientSecret", "secretKey", "accessToken", "refreshToken", "sessionCookie",
	} {
		cases = append(cases, redactionCase{name: "member/" + key, structured: true, input: `{"` + key + `":"` + secret + `"}`, want: `{"` + key + `":"<redacted>"}` + "\n", secret: secret})
	}
	cases = append(cases,
		// Rule output that would keep a prefix under a secret-named key.
		redactionCase{name: "member/dsn_url", structured: true, input: `{"dsn":"postgres://u:` + secret + `@h/db"}`, want: `{"dsn":"<redacted>"}` + "\n", secret: secret},
		redactionCase{name: "member/database_dsn_url", structured: true, input: `{"database_dsn":"postgres://u:` + secret + `@h/db"}`, want: `{"database_dsn":"<redacted>"}` + "\n", secret: secret},
		redactionCase{name: "member/cookies_assignment", structured: true, input: `{"cookies":"session=` + secret + `"}`, want: `{"cookies":"<redacted>"}` + "\n", secret: secret},
		redactionCase{name: "member/session_secret_bearer", structured: true, input: `{"session_secret":"Bearer ` + secret + `"}`, want: `{"session_secret":"<redacted>"}` + "\n", secret: secret},
		redactionCase{name: "member/nested", structured: true, input: `{"nested":{"SECRET_KEY":"` + secret + `","api_key":"` + secret + `"}}`, want: `{"nested":{"SECRET_KEY":"<redacted>","api_key":"<redacted>"}}` + "\n", secret: secret},
		// A rule applied inside a JSON member whose key is not secret-named.
		redactionCase{name: "member/message_url", structured: true, input: `{"message":"postgres://u:` + secret + `@h/db"}`, want: `{"message":"postgres://<redacted>@h/db"}` + "\n", secret: secret},
		redactionCase{name: "member/message_assignment", structured: true, input: `{"message":"run DB_PASSWORD=` + secret + ` now"}`, want: `{"message":"run DB_PASSWORD=<redacted> now"}` + "\n", secret: secret},
		// A whole string that is a secret-named assignment, and CLI arguments.
		redactionCase{name: "member/whole_string_assignment", structured: true, input: `{"env":["DB_PASSWORD=fixture secret 1234","TOKEN_FILE=/run/` + secret + `","PLAIN=value"]}`, want: `{"env":["DB_PASSWORD=<redacted>","TOKEN_FILE=<redacted>","PLAIN=value"]}` + "\n", secret: secret},
		redactionCase{name: "member/cli_arguments", structured: true, input: `{"args":["--token","` + secret + `","--dsn=postgres://u:` + secret + `@h/db"]}`, want: `{"args":["--token","<redacted>","--dsn=<redacted>"]}` + "\n", secret: secret},
		// SESSION is not a structured key token.
		redactionCase{name: "member/message_session", structured: true, input: `{"message":"SESSION=` + secret + `"}`, want: `{"message":"SESSION=<redacted>"}` + "\n", secret: secret},
		redactionCase{name: "member/env_session", structured: true, input: `{"env":["SESSION=` + secret + ` tail"]}`, want: `{"env":["SESSION=<redacted> tail"]}` + "\n", secret: secret},
		redactionCase{name: "preserve/session_member", structured: true, input: `{"session":"browser-session-name"}`, want: `{"session":"browser-session-name"}` + "\n"},
		// Canonicalization uppercases only ASCII letters; the scan closure still
		// follows the scan's case folding.
		redactionCase{name: "member/long_s_secret", structured: true, input: `{"ſecret":"` + secret + `"}`, want: `{"ſecret":"<redacted>"}` + "\n", secret: secret},
		redactionCase{name: "member/non_ascii_prefix_secret", structured: true, input: `{"tökenSecret":"` + secret + `"}`, want: `{"tökenSecret":"<redacted>"}` + "\n", secret: secret},
		redactionCase{name: "preserve/sharp_s_key", structured: true, input: `{"paßword":"fixture-plain"}`, want: `{"paßword":"fixture-plain"}` + "\n"},
		// Structural fields, containers and non-string scalars are preserved.
		redactionCase{name: "preserve/session_target", structured: true, input: `{"session_target":"browser-stage-token-name"}`, want: `{"session_target":"browser-stage-token-name"}` + "\n"},
		redactionCase{name: "preserve/cpu_tokens", structured: true, input: `{"cpu_tokens":"2"}`, want: `{"cpu_tokens":"2"}` + "\n"},
		redactionCase{name: "preserve/closure_container", structured: true, input: `{"tokenizer":{"kind":"bpe"}}`, want: `{"tokenizer":{"kind":"bpe"}}` + "\n"},
		redactionCase{name: "preserve/closure_scalar", structured: true, input: `{"cookies":3,"tokenizer_count":2}`, want: `{"cookies":3,"tokenizer_count":2}` + "\n"},
		redactionCase{name: "preserve/service_sessions", structured: true, input: `{"service_sessions":[{"cleanup_status":"pass","setup_duration_ms":12}]}`, want: `{"service_sessions":[{"cleanup_status":"pass","setup_duration_ms":12}]}` + "\n"},
	)
	return cases
}

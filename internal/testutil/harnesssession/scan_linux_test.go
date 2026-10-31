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

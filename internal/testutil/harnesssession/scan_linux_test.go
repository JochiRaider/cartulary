//go:build linux

package harnesssession

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestSessionScan(t *testing.T) {
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
					_ = os.WriteFile(file, []byte("safe"), 0644)
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

package renderprocess

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestRenderInstallationAdmission(t *testing.T) {
	for _, mutation := range []string{"none", "changed_bytes", "extra_file", "missing_file", "link", "hardlink", "human_material", "manifest_digest", "package_digest", "font_digest", "duplicate_path", "unknown_field", "duplicate_field", "root_escape", "package_version", "version_label"} {
		t.Run(mutation, func(t *testing.T) {
			directory := t.TempDir()
			staging := filepath.Join(directory, "staging")
			files := []InstallationFile{}
			for _, name := range []string{"bin/cartulary-render-sandbox", "bin/bwrap", "lib64/ld-linux-x86-64.so.2", "lib/x86_64-linux-gnu/libc.so.6", "runtime/node/bin/node", "runtime/renderer/src/stage-main.ts", "runtime/renderer/src/browser-profile.json", "runtime/fonts/fonts.conf", "runtime/fonts/fixture.ttf", "runtime/workspace-lock.yaml", "runtime/browser/chrome-headless-shell"} {
				if err := os.MkdirAll(filepath.Join(staging, filepath.Dir(name)), 0o700); err != nil {
					t.Fatal(err)
				}
				content := []byte(name)
				if name == "runtime/renderer/src/browser-profile.json" {
					content = []byte(`{"schema_id":"cartulary.reporting_browser_profile.v1","browser_launch_args":["--headless","--no-sandbox","--remote-debugging-pipe","--user-data-dir=/work/browser-profile"]}`)
				}
				if err := os.WriteFile(filepath.Join(staging, name), content, 0o600); err != nil {
					t.Fatal(err)
				}
				files = append(files, InstallationFile{Path: name, SHA256: installationHash(content)})
			}
			slices.SortFunc(files, func(a, b InstallationFile) int { return strings.Compare(a.Path, b.Path) })
			setDigest := func(prefix string) string {
				entries := []InstallationFile{}
				for _, file := range files {
					if strings.HasPrefix(file.Path, prefix) {
						entries = append(entries, InstallationFile{Path: strings.TrimPrefix(file.Path, prefix), SHA256: file.SHA256})
					}
				}
				data, err := canonicaljson.Marshal(entries)
				if err != nil {
					t.Fatal(err)
				}
				return installationHash(data)
			}
			manifest := InstallationManifest{SchemaID: "cartulary.reporting_runtime_installation.v1", Platform: "linux/amd64", NodeVersion: "1.0.0", PackageManager: "pnpm@1.0.0", LockfileSHA256: installationHash([]byte("runtime/workspace-lock.yaml")), PackageStoreDigest: setDigest("runtime/renderer/"), FontManifestSHA256: setDigest("runtime/fonts/"), ChromiumExecutablePath: "runtime/browser/chrome-headless-shell", ChromiumExecutableSHA256: installationHash([]byte("runtime/browser/chrome-headless-shell")), ChromiumVersion: "synthetic-test-browser", BubblewrapVersion: "1.0.0", Files: files, Packages: map[string]string{}}
			for _, name := range []string{"@slidev/cli", "@slidev/theme-default", "markdown-it", "mermaid", "pdf-lib", "playwright-chromium"} {
				manifest.Packages[name] = "1.0.0"
			}
			changedFile := filepath.Join(staging, "runtime/renderer/src/stage-main.ts")
			switch mutation {
			case "changed_bytes":
				if err := os.WriteFile(changedFile, []byte("altered"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "extra_file":
				if err := os.WriteFile(filepath.Join(staging, "undeclared"), []byte("extra"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "missing_file":
				if err := os.Remove(changedFile); err != nil {
					t.Fatal(err)
				}
			case "link":
				if err := os.Remove(changedFile); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(directory, "outside"), changedFile); err != nil {
					t.Fatal(err)
				}
			case "hardlink":
				if err := os.Link(changedFile, filepath.Join(directory, "alias")); err != nil {
					t.Fatal(err)
				}
			case "human_material":
				// The name must reject before following even a nonexistent target.
				if err := os.Symlink(filepath.Join(directory, "absent"), filepath.Join(staging, "README.md")); err != nil {
					t.Fatal(err)
				}
			case "package_version":
				manifest.Packages["mermaid"] = "floating"
			case "version_label":
				manifest.ChromiumVersion = "invalid\nlabel"
			case "package_digest":
				manifest.PackageStoreDigest = strings.Repeat("0", 64)
			case "font_digest":
				manifest.FontManifestSHA256 = strings.Repeat("0", 64)
			case "duplicate_path":
				manifest.Files = append(manifest.Files, manifest.Files[0])
			}
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if mutation == "unknown_field" {
				data = append([]byte(`{"unknown":true,`), data[1:]...)
			}
			if mutation == "duplicate_field" {
				data = append([]byte(`{"platform":"linux/amd64",`), data[1:]...)
			}
			digest := installationHash(data)
			rootName := digest + ".root"
			if err := os.Rename(staging, filepath.Join(directory, rootName)); err != nil {
				t.Fatal(err)
			}
			if mutation == "manifest_digest" {
				data = append(data, '\n')
			}
			if err := os.WriteFile(filepath.Join(directory, digest+".json"), data, 0o600); err != nil {
				t.Fatal(err)
			}
			if mutation == "root_escape" {
				rootName = "../outside"
			}
			receipt, err := json.Marshal(map[string]string{"schema_id": "cartulary.reporting_runtime_receipt.v1", "manifest_sha256": digest, "root": rootName, "manifest": digest + ".json"})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "installation.json"), receipt, 0o600); err != nil {
				t.Fatal(err)
			}
			installation, err := OpenInstallation(directory)
			if mutation == "none" {
				if err != nil || installation.Digest() != digest {
					t.Fatalf("valid installation: %v", err)
				}
				copy := installation.Manifest()
				copy.Files[0].Path = "altered"
				copy.Packages["mermaid"] = "altered"
				if installation.Verify() != nil {
					t.Fatal("caller acquired writable admission state")
				}
			} else if !errors.Is(err, ErrInstallation) || installation != nil {
				t.Fatalf("unsafe installation admitted: %v", err)
			}
		})
	}
}

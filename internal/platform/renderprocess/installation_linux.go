package renderprocess

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"golang.org/x/text/unicode/norm"
)

var ErrInstallation = errors.New("render runtime installation invalid")
var installationVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

func boundedInstallationLabel(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character > 0x7e {
			return false
		}
	}
	return true
}

func validInstallationVersion(value string) bool {
	return boundedInstallationLabel(value) && installationVersion.MatchString(value)
}

type InstallationFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// InstallationManifest describes actual packaged bytes. It is an operational
// projection, not a specification/adoption record or a release qualification.
type InstallationManifest struct {
	SchemaID                 string             `json:"schema_id"`
	Platform                 string             `json:"platform"`
	NodeVersion              string             `json:"node_version"`
	PackageManager           string             `json:"package_manager"`
	LockfileSHA256           string             `json:"lockfile_sha256"`
	PackageStoreDigest       string             `json:"package_store_digest"`
	FontManifestSHA256       string             `json:"font_manifest_sha256"`
	ChromiumExecutablePath   string             `json:"chromium_executable_path"`
	ChromiumExecutableSHA256 string             `json:"chromium_executable_sha256"`
	ChromiumVersion          string             `json:"chromium_version"`
	BubblewrapVersion        string             `json:"bubblewrap_version"`
	Packages                 map[string]string  `json:"packages"`
	Files                    []InstallationFile `json:"files"`
}

type Installation struct {
	root     string
	digest   string
	manifest InstallationManifest
}

func decodeInstallationJSON(data []byte, value any) error {
	if !utf8.Valid(data) || strictjson.ValidateObject(data) != nil {
		return ErrInstallation
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(value) != nil {
		return ErrInstallation
	}
	return nil
}

func validInstallationDigest(value string) bool {
	data, err := hex.DecodeString(value)
	return err == nil && len(data) == sha256.Size && strings.ToLower(value) == value
}

func installationHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// OpenInstallation never infers missing files or discovers host dependencies.
// Receipt paths are derived from the validated content identity before opening.
func OpenInstallation(directory string) (*Installation, error) {
	if !filepath.IsAbs(directory) || filepath.Clean(directory) != directory || directory == "/" {
		return nil, ErrInstallation
	}
	for _, segment := range strings.Split(directory, string(filepath.Separator)) {
		if forbiddenInstallationName(segment) {
			return nil, ErrInstallation
		}
	}
	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil || resolved != directory {
		return nil, ErrInstallation
	}
	data, err := readOutput(directory, Output{Path: "installation.json", MaxBytes: 4096})
	if err != nil {
		return nil, ErrInstallation
	}
	var receipt struct {
		SchemaID       string `json:"schema_id"`
		ManifestSHA256 string `json:"manifest_sha256"`
		Root           string `json:"root"`
		Manifest       string `json:"manifest"`
	}
	if decodeInstallationJSON(data, &receipt) != nil || receipt.SchemaID != "cartulary.reporting_runtime_receipt.v1" || !validInstallationDigest(receipt.ManifestSHA256) || receipt.Root != receipt.ManifestSHA256+".root" || receipt.Manifest != receipt.ManifestSHA256+".json" {
		return nil, ErrInstallation
	}
	data, err = readOutput(directory, Output{Path: receipt.Manifest, MaxBytes: 32 << 20})
	if err != nil || installationHash(data) != receipt.ManifestSHA256 {
		return nil, ErrInstallation
	}
	var manifest InstallationManifest
	if decodeInstallationJSON(data, &manifest) != nil ||
		manifest.SchemaID != "cartulary.reporting_runtime_installation.v1" || manifest.Platform != "linux/amd64" ||
		!validInstallationVersion(manifest.NodeVersion) || !validInstallationVersion(manifest.BubblewrapVersion) ||
		!boundedInstallationLabel(manifest.ChromiumVersion) || !boundedInstallationLabel(manifest.PackageManager) ||
		!strings.HasPrefix(manifest.PackageManager, "pnpm@") || !validInstallationVersion(strings.TrimPrefix(manifest.PackageManager, "pnpm@")) {
		return nil, ErrInstallation
	}
	if !validInstallationDigest(manifest.LockfileSHA256) || !validInstallationDigest(manifest.PackageStoreDigest) || !validInstallationDigest(manifest.FontManifestSHA256) || !validInstallationDigest(manifest.ChromiumExecutableSHA256) {
		return nil, ErrInstallation
	}
	if manifest.ChromiumExecutablePath != "runtime/browser/chrome-headless-shell" {
		return nil, ErrInstallation
	}
	if len(manifest.Packages) != 6 {
		return nil, ErrInstallation
	}
	for _, name := range []string{"@slidev/cli", "@slidev/theme-default", "markdown-it", "mermaid", "pdf-lib", "playwright-chromium"} {
		if !validInstallationVersion(manifest.Packages[name]) {
			return nil, ErrInstallation
		}
	}
	installation := &Installation{root: filepath.Join(directory, receipt.Root), digest: receipt.ManifestSHA256, manifest: manifest}
	if err := installation.Verify(); err != nil {
		return nil, err
	}
	if _, err := installation.BrowserLaunchArgs(); err != nil {
		return nil, err
	}
	return installation, nil
}

// BrowserLaunchArgs reads the installed profile whose bytes are bound by the
// complete file manifest. The driver must run with automatic arguments disabled.
func (i *Installation) BrowserLaunchArgs() ([]string, error) {
	data, err := readOutput(i.root, Output{Path: "runtime/renderer/src/browser-profile.json", MaxBytes: 32 << 10})
	if err != nil {
		return nil, ErrInstallation
	}
	var profile struct {
		SchemaID  string   `json:"schema_id"`
		Arguments []string `json:"browser_launch_args"`
	}
	if decodeInstallationJSON(data, &profile) != nil || profile.SchemaID != "cartulary.reporting_browser_profile.v1" || len(profile.Arguments) == 0 || len(profile.Arguments) > 128 {
		return nil, ErrInstallation
	}
	seen := map[string]bool{}
	for _, argument := range profile.Arguments {
		if len(argument) > 4096 || !strings.HasPrefix(argument, "--") || strings.ContainsAny(argument, "\x00\r\n") || seen[argument] {
			return nil, ErrInstallation
		}
		seen[argument] = true
	}
	for _, required := range []string{"--user-data-dir=/work/browser-profile", "--remote-debugging-pipe", "--headless", "--no-sandbox"} {
		if !seen[required] {
			return nil, ErrInstallation
		}
	}
	return profile.Arguments, nil
}

func forbiddenInstallationName(name string) bool {
	lower := strings.ToLower(name)
	return lower == "docs" || lower == "readme" || strings.HasPrefix(lower, "readme.") || strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}

// Verify closes the complete set, including unexpected files. It rejects human
// material by name before descending into a directory or opening a file.
func (i *Installation) Verify() error {
	if i == nil || len(i.manifest.Files) == 0 || len(i.manifest.Files) > 100000 {
		return ErrInstallation
	}
	resolved, err := filepath.EvalSymlinks(i.root)
	if err != nil || resolved != i.root {
		return ErrInstallation
	}
	expected := make(map[string]string, len(i.manifest.Files))
	previous := ""
	for _, entry := range i.manifest.Files {
		if !safeRelativePath(entry.Path) || !utf8.ValidString(entry.Path) || !norm.NFC.IsNormalString(entry.Path) || entry.Path <= previous || !validInstallationDigest(entry.SHA256) {
			return ErrInstallation
		}
		for _, segment := range strings.Split(entry.Path, "/") {
			if forbiddenInstallationName(segment) {
				return ErrInstallation
			}
		}
		previous = entry.Path
		expected[entry.Path] = entry.SHA256
	}
	for _, required := range []string{"bin/cartulary-render-sandbox", "bin/bwrap", "lib64/ld-linux-x86-64.so.2", "lib/x86_64-linux-gnu/libc.so.6", "runtime/node/bin/node", "runtime/renderer/src/stage-main.ts", "runtime/renderer/src/browser-profile.json", "runtime/fonts/fonts.conf", "runtime/workspace-lock.yaml", i.manifest.ChromiumExecutablePath} {
		if expected[required] == "" {
			return ErrInstallation
		}
	}
	if expected["runtime/workspace-lock.yaml"] != i.manifest.LockfileSHA256 || expected[i.manifest.ChromiumExecutablePath] != i.manifest.ChromiumExecutableSHA256 {
		return ErrInstallation
	}
	for prefix, want := range map[string]string{"runtime/renderer/": i.manifest.PackageStoreDigest, "runtime/fonts/": i.manifest.FontManifestSHA256} {
		entries := []InstallationFile{}
		for _, entry := range i.manifest.Files {
			if strings.HasPrefix(entry.Path, prefix) {
				entries = append(entries, InstallationFile{Path: strings.TrimPrefix(entry.Path, prefix), SHA256: entry.SHA256})
			}
		}
		data, err := canonicaljson.Marshal(entries)
		if err != nil || len(entries) == 0 || installationHash(data) != want {
			return ErrInstallation
		}
	}
	opened, err := os.OpenRoot(i.root)
	if err != nil {
		return ErrInstallation
	}
	defer opened.Close()
	count := 0
	err = fs.WalkDir(opened.FS(), ".", func(relative string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || forbiddenInstallationName(entry.Name()) {
			return ErrInstallation
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() || expected[relative] == "" {
			return ErrInstallation
		}
		file, err := opened.OpenFile(relative, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
		if err != nil {
			return ErrInstallation
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<30 {
			return ErrInstallation
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok || stat.Nlink != 1 {
			return ErrInstallation
		}
		digest := sha256.New()
		if _, err := io.Copy(digest, io.LimitReader(file, (1<<30)+1)); err != nil || hex.EncodeToString(digest.Sum(nil)) != expected[relative] {
			return ErrInstallation
		}
		count++
		return nil
	})
	if err != nil || count != len(expected) {
		return ErrInstallation
	}
	return nil
}

func (i *Installation) Digest() string { return i.digest }

func (i *Installation) Manifest() InstallationManifest {
	// Callers receive a copy, never writable ownership of admission state.
	data, _ := json.Marshal(i.manifest)
	var result InstallationManifest
	_ = json.Unmarshal(data, &result)
	return result
}

func (i *Installation) Runtime(temporaryRoot string, loopbackPort uint16) Runtime {
	return Runtime{
		// The launcher executes before the private root exists. Use its installed
		// ELF loader and library closure so a different host libc cannot select
		// another implementation. No inherited loader environment is admitted.
		SandboxCommand: []string{filepath.Join(i.root, "lib64/ld-linux-x86-64.so.2"), "--inhibit-cache", "--inhibit-rpath", "", "--library-path", filepath.Join(i.root, "lib/x86_64-linux-gnu"), filepath.Join(i.root, "bin/bwrap")}, Root: i.root, TemporaryRoot: temporaryRoot,
		Command:  []string{"/runtime/node/bin/node", "/runtime/renderer/src/stage-main.ts"},
		Launcher: []string{"/bin/cartulary-render-sandbox"}, LoopbackPort: loopbackPort,
		Environment: map[string]string{"LANG": "C.UTF-8", "LC_ALL": "C.UTF-8", "TZ": "UTC", "HOME": "/work/home", "TMPDIR": "/work/tmp", "PATH": "/runtime/node/bin", "FONTCONFIG_FILE": "/runtime/fonts/fonts.conf", "UV_USE_IO_URING": "0", "CARTULARY_RENDER_CHROMIUM": "/" + i.manifest.ChromiumExecutablePath},
	}
}

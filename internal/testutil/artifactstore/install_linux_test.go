//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func writeExecutable(t *testing.T, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

// nodeArchive builds a release archive whose node reports version and whose
// corepack enables a pnpm reporting pnpmVersion, and returns its bytes.
func nodeArchive(t *testing.T, version, pnpmVersion string) []byte {
	t.Helper()
	tree := filepath.Join(t.TempDir(), "node-v"+version)
	writeExecutable(t, filepath.Join(tree, "bin", "node"), "#!/bin/sh\n[ \"$1\" = --version ] && echo v"+version+"\n")
	writeExecutable(t, filepath.Join(tree, "bin", "corepack"), `#!/bin/sh
case "$1" in
  enable) printf '#!/bin/sh\necho `+pnpmVersion+`\n' >"$3/pnpm" && chmod 755 "$3/pnpm" ;;
  prepare) [ -n "$COREPACK_HOME" ] && mkdir -p "$COREPACK_HOME" ;;
  *) exit 3 ;;
esac
`)
	archive := filepath.Join(t.TempDir(), "node.tar.xz")
	if output, err := exec.Command("tar", "-cJf", archive, "-C", filepath.Dir(tree), filepath.Base(tree)).CombinedOutput(); err != nil {
		t.Fatalf("tar: %v: %s", err, output)
	}
	content, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

// releaseServer serves content at /v<version>/<archive>, after failing the
// first failures requests, and counts every request.
func releaseServer(t *testing.T, version, archive string, content []byte, failures int32) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var served atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if served.Add(1) <= failures {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		if r.URL.Path != "/v"+version+"/"+archive {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(content)
	}))
	t.Cleanup(server.Close)
	return server, &served
}

func digestOf(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func staged(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			names = append(names, entry.Name())
		}
	}
	return names
}

// The Node runtime installs from its verified archive with pnpm prepared,
// swaps into place, and is reused while its receipt records the current key.
func TestNodeRuntimeInstallsFromItsArchive(t *testing.T) {
	downloadRetryDelay = 0
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := NodeArchive("9.9.9")
	if err != nil {
		t.Fatal(err)
	}
	content := nodeArchive(t, "9.9.9", "8.8.8")
	// The first request fails, so the download is retried.
	server, served := releaseServer(t, "9.9.9", archive, content, 1)
	scratch := t.TempDir()
	runtime := NodeRuntime{Root: root, Directory: filepath.Join(scratch, "node-runtime"), Version: "9.9.9", PNPMVersion: "8.8.8",
		Archive: archive, SHA256: digestOf(content), Archives: filepath.Join(scratch, "archives"), Releases: server.URL, Env: os.Environ()}
	lock := runtime.Directory + ".lock"
	install := func(r NodeRuntime, controls Controls, installedOnly bool) (bool, error) {
		t.Helper()
		return Install(r, lock, controls, installedOnly)
	}
	if _, err := install(runtime, Controls{}, true); err == nil || !strings.Contains(err.Error(), "installed-only") {
		t.Fatalf("the installed-only policy installed: %v", err)
	}
	if reused, err := install(runtime, Controls{}, false); err != nil || reused || served.Load() != 2 {
		t.Fatalf("first install reused %v, %v, requests %d", reused, err, served.Load())
	}
	if output, err := exec.Command(filepath.Join(runtime.Directory, "bin", "node"), "--version").Output(); err != nil || string(output) != "v9.9.9\n" {
		t.Fatalf("installed node %q, %v", output, err)
	}
	if names := staged(t, scratch); len(names) != 0 {
		t.Fatalf("staging remained: %v", names)
	}
	if reused, err := install(runtime, Controls{}, true); err != nil || !reused || served.Load() != 2 {
		t.Fatalf("a current runtime was not reused: %v, %v", reused, err)
	}
	// A forced install reinstalls from the kept archive.
	if reused, err := install(runtime, Controls{Force: true}, false); err != nil || reused || served.Load() != 2 {
		t.Fatalf("forced install reused %v, %v, requests %d", reused, err, served.Load())
	}
	// Concurrent installs serialize: one installs, the other reuses.
	if err := os.Remove(filepath.Join(runtime.Directory, NodeRuntimeReceipt)); err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	var reusedCount atomic.Int32
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			reused, err := Install(runtime, lock, Controls{}, false)
			if err != nil {
				t.Error(err)
			}
			if reused {
				reusedCount.Add(1)
			}
		}()
	}
	group.Wait()
	if reusedCount.Load() != 1 {
		t.Fatalf("concurrent installs reused %d times", reusedCount.Load())
	}
	// Another pnpm pin changes the key, so the receipt no longer matches.
	repinned := runtime
	repinned.PNPMVersion = "8.8.9"
	if current, err := repinned.Current(); err != nil || current {
		t.Fatalf("a repinned runtime counted as current: %v", err)
	}
	// The installed pnpm reports 8.8.8, so the install fails and keeps the
	// previous runtime.
	if _, err := install(repinned, Controls{}, false); err == nil || !strings.Contains(err.Error(), "8.8.9") {
		t.Fatalf("a mismatched pnpm was installed: %v", err)
	}
	if current, err := runtime.Current(); err != nil || !current {
		t.Fatalf("a failed install replaced the runtime: %v", err)
	}
	if names := staged(t, scratch); len(names) != 0 {
		t.Fatalf("a failed install left %v", names)
	}
}

// An archive without the committed digest is removed and never extracted.
func TestNodeArchiveIsVerified(t *testing.T) {
	downloadRetryDelay = 0
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := NodeArchive("9.9.9")
	if err != nil {
		t.Fatal(err)
	}
	server, _ := releaseServer(t, "9.9.9", archive, nodeArchive(t, "9.9.9", "8.8.8"), 0)
	scratch := t.TempDir()
	runtime := NodeRuntime{Root: root, Directory: filepath.Join(scratch, "node-runtime"), Version: "9.9.9", PNPMVersion: "8.8.8",
		Archive: archive, SHA256: strings.Repeat("0", 64), Archives: filepath.Join(scratch, "archives"), Releases: server.URL, Env: os.Environ()}
	if _, err := Install(runtime, runtime.Directory+".lock", Controls{}, false); err == nil || !strings.Contains(err.Error(), "not the committed") {
		t.Fatalf("an archive with another digest was accepted: %v", err)
	}
	if entries, _ := os.ReadDir(runtime.Archives); len(entries) != 0 {
		t.Fatalf("a refused archive remained: %v", entries)
	}
	if _, err := os.Stat(runtime.Directory); !os.IsNotExist(err) {
		t.Fatalf("a refused archive was extracted: %v", err)
	}
	// A server that keeps failing is retried, then reported.
	failing, served := releaseServer(t, "9.9.9", archive, nil, 100)
	runtime.Releases = failing.URL
	if _, err := Install(runtime, runtime.Directory+".lock", Controls{}, false); err == nil || served.Load() != downloadAttempts {
		t.Fatalf("download failures: %v after %d requests", err, served.Load())
	}
}

// frontendFakes writes a pnpm and a node standing in for the receipt's
// owner. The node reports the installation current while the receipt exists
// and publishes it only after an install ran; the pnpm refuses to install
// while the receipt exists.
func frontendFakes(t *testing.T) (FrontendInstall, string) {
	t.Helper()
	scratch := t.TempDir()
	receipt := filepath.Join(scratch, "proof", "installed.v1.json")
	marker := filepath.Join(scratch, "installs")
	node, pnpm := filepath.Join(scratch, "bin", "node"), filepath.Join(scratch, "bin", "pnpm")
	writeExecutable(t, node, `#!/bin/sh
[ "$1" = `+readinessOwner+` ] || exit 90
case "$2" in
  frontend-installation) [ -f "$FAKE_RECEIPT" ] ;;
  publish) [ -s "$FAKE_MARKER" ] && mkdir -p "$(dirname "$FAKE_RECEIPT")" && : >"$FAKE_RECEIPT" ;;
  *) exit 91 ;;
esac
`)
	writeExecutable(t, pnpm, `#!/bin/sh
case "$*" in
  "config get store-dir") echo "${FAKE_STORE_DIR:-.pnpm-store}" ;;
  "config get confirmModulesPurge") echo false ;;
  "install --frozen-lockfile --reporter=append-only --loglevel=warn")
    [ "$CI" = true ] || exit 92
    [ ! -f "$FAKE_RECEIPT" ] || exit 93
    if [ -n "$FAKE_OUTDATED" ]; then echo " ERR_PNPM_OUTDATED_LOCKFILE  Cannot install" >&2; exit 1; fi
    echo x >>"$FAKE_MARKER" ;;
  *) exit 94 ;;
esac
`)
	install := FrontendInstall{Root: scratch, Receipt: receipt, Node: node, PNPM: pnpm, NodeRuntime: filepath.Join(scratch, "runtime"),
		Env: append(os.Environ(), "FAKE_RECEIPT="+receipt, "FAKE_MARKER="+marker)}
	return install, marker
}

func installCount(t *testing.T, marker string) int {
	t.Helper()
	data, err := os.ReadFile(marker)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(data), "x")
}

// The frontend installation checks the workspace's pnpm settings, removes
// its receipt, installs with a frozen lockfile and has the receipt's owner
// publish the receipt; a current receipt is reused.
func TestFrontendInstall(t *testing.T) {
	install, marker := frontendFakes(t)
	lock := filepath.Join(install.Root, "frontend-install.lock")
	if reused, err := Install(install, lock, Controls{}, false); err != nil || reused || installCount(t, marker) != 1 {
		t.Fatalf("first install reused %v, %v, installs %d", reused, err, installCount(t, marker))
	}
	if reused, err := Install(install, lock, Controls{}, false); err != nil || !reused || installCount(t, marker) != 1 {
		t.Fatalf("a current installation was not reused: %v, %v", reused, err)
	}
	// Disabling reuse reinstalls; the pnpm fake fails if the receipt was
	// still present when it ran.
	if reused, err := Install(install, lock, Controls{Disable: true}, false); err != nil || reused || installCount(t, marker) != 2 {
		t.Fatalf("a disabled cache reused %v, %v", reused, err)
	}
	outdated := install
	outdated.Env = append(outdated.Env, "FAKE_OUTDATED=1")
	if _, err := Install(outdated, lock, Controls{Force: true}, false); err == nil || !strings.Contains(err.Error(), "make frontend-lockfile-update") {
		t.Fatalf("an outdated lockfile: %v", err)
	}
	if _, err := os.Stat(install.Receipt); !os.IsNotExist(err) {
		t.Fatalf("a failed install kept its receipt: %v", err)
	}
	misconfigured := install
	misconfigured.Env = append(misconfigured.Env, "FAKE_STORE_DIR=/elsewhere")
	if _, err := Install(misconfigured, lock, Controls{}, false); err == nil || !strings.Contains(err.Error(), "store-dir") || installCount(t, marker) != 2 {
		t.Fatalf("another pnpm store was used: %v", err)
	}
	if _, err := Install(install, lock, Controls{}, true); err == nil {
		t.Fatal("the installed-only policy accepted a missing receipt")
	}
}

// Playwright browsers are reused only while the receipt records the current
// key and the browser readiness proof passes, so a removed renderer image
// installs again.
func TestPlaywrightReuseNeedsItsProof(t *testing.T) {
	root := t.TempDir()
	for _, name := range playwrightInputs {
		writeFile(t, root, name, name)
	}
	count := filepath.Join(root, "installs")
	writeExecutable(t, filepath.Join(root, filepath.FromSlash(playwrightInstaller)), "#!/bin/sh\n[ -n \"$NODE_BIN\" ] && [ -n \"$PNPM\" ] && echo x >>"+count+"\n")
	proof := filepath.Join(root, "proof-fails")
	node := filepath.Join(root, "bin", "node")
	writeExecutable(t, node, "#!/bin/sh\n[ \"$1\" = "+readinessOwner+" ] && [ \"$2\" = browser ] && [ ! -f "+proof+" ]\n")
	browsers := Playwright{Root: root, Receipt: filepath.Join(root, "playwright", "installation.json"), Node: node, PNPM: "/bin/true",
		NodeRuntime: filepath.Join(root, "runtime"), Env: os.Environ()}
	lock := filepath.Join(root, "playwright.lock")
	installs := func() int { return installCount(t, count) }
	if reused, err := Install(browsers, lock, Controls{}, false); err != nil || reused || installs() != 1 {
		t.Fatalf("first install reused %v, %v", reused, err)
	}
	if reused, err := Install(browsers, lock, Controls{}, false); err != nil || !reused || installs() != 1 {
		t.Fatalf("a proven installation was not reused: %v, %v", reused, err)
	}
	if err := os.WriteFile(proof, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(browsers, lock, Controls{}, true); err == nil {
		t.Fatal("the installed-only policy accepted a failed proof")
	}
	if reused, err := Install(browsers, lock, Controls{}, false); err != nil || reused || installs() != 2 {
		t.Fatalf("a failed proof was reused: %v, %v", reused, err)
	}
	if err := os.Remove(proof); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "tools/frontend_visual_renderer_profile.json", "another renderer")
	if reused, err := Install(browsers, lock, Controls{}, false); err != nil || reused || installs() != 3 {
		t.Fatalf("a changed renderer profile was reused: %v, %v", reused, err)
	}
}

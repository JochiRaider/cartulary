//go:build linux

package artifactstore

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func goForTest(t *testing.T) string {
	t.Helper()
	if goExecutable := os.Getenv("GO"); goExecutable != "" {
		return goExecutable
	}
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		t.Fatal("the go command is unavailable:", err)
	}
	return goExecutable
}

// sealedStore is a store whose read-only roots the test removes.
func sealedStore(t *testing.T) Store {
	t.Helper()
	root := t.TempDir()
	t.Cleanup(func() { _ = removeTree(root) })
	return Store{Root: filepath.Join(root, "store")}
}

// receiptNode is a node standing in for the cryptographic receipt writer: it
// accepts the pinned module and writes a receipt naming the binary it was
// given, after checking that the binary exists.
func receiptNode(t *testing.T, directory string) string {
	t.Helper()
	node := filepath.Join(directory, "node")
	writeExecutable(t, node, `#!/bin/sh
[ "$1" = --version ] && { echo v0.0.0-fixture; exit 0; }
[ "$1" = tools/harness/readiness/cryptographic-build-cli.mjs ] || exit 90
case "$2" in
  verify) exit 0 ;;
  receipt) [ -x "$3" ] && printf '{"binary":"%s"}\n' "$3" >"$4" ;;
  *) exit 91 ;;
esac
`)
	return node
}

// receiptModule writes a module with a command, the pins and the receipt
// writer's files a Go binary's key reads.
func receiptModule(t *testing.T) string {
	t.Helper()
	module := t.TempDir()
	writeFile(t, module, "go.mod", "module example.test/launcher\n\ngo 1.22\n")
	writeFile(t, module, "go.sum", "")
	writeFile(t, module, "tools/toolchain_pins.json", `{"go_toolchain":"go1.22","cryptographic_module":{"selector":"fixture"}}`)
	for _, name := range append([]string{producerImplementation}, receiptWriter...) {
		writeFile(t, module, name, "producer and receipt writer\n")
	}
	writeFile(t, module, "cmd/launcher/main.go", "package main\n\nfunc main() {}\n")
	return module
}

// A Go binary builds into staging without VCS stamping, then its receipt is
// written for the staged binary, so a binary the receipt rejects never
// reaches a root.
func TestGoBinaryBuildsAndReceipts(t *testing.T) {
	module := receiptModule(t)
	binary := GoBinary{Module: module, Package: "./cmd/launcher", Name: "launcher", Go: goForTest(t), Node: receiptNode(t, t.TempDir()),
		Env: append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")}
	store := sealedStore(t)
	result, err := store.Produce(binary, map[string]string{}, Controls{})
	if err != nil {
		t.Fatal(err)
	}
	root := store.RootPath(result.Root)
	receipt, err := os.ReadFile(filepath.Join(root, "launcher.crypto.json"))
	if err != nil || !strings.Contains(string(receipt), "launcher") {
		t.Fatalf("receipt %q, %v", receipt, err)
	}
	if info, err := os.Stat(filepath.Join(root, "launcher")); err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("the launcher is not an executable: %v", err)
	}
	if output, err := exec.Command(goForTest(t), "version", "-m", filepath.Join(root, "launcher")).Output(); err != nil || strings.Contains(string(output), "vcs.revision") {
		t.Fatalf("the build carries VCS stamping: %s, %v", output, err)
	}
}

// The report renderer builds its launcher through the store and hands the
// packaging script the launcher and the launcher's root.
func TestReportRendererGetsItsLauncherFromTheStore(t *testing.T) {
	module := receiptModule(t)
	node := receiptNode(t, t.TempDir())
	seen := filepath.Join(t.TempDir(), "seen")
	// The node fixture runs "packaging" by recording what it was given.
	writeExecutable(t, node, `#!/bin/sh
case "$1 $2" in
  "tools/harness/readiness/cryptographic-build-cli.mjs verify") exit 0 ;;
  "tools/harness/readiness/cryptographic-build-cli.mjs receipt") printf '{}\n' >"$4"; exit 0 ;;
esac
[ "$1" = --version ] && { echo v0.0.0-fixture; exit 0; }
[ "$1" = tools/report-renderer/build-runtime.mjs ] || exit 91
[ -x "$RENDER_SANDBOX_BIN" ] || exit 92
printf '%s\n' "$RENDER_SANDBOX_ROOT" >`+seen+`
`)
	store := sealedStore(t)
	renderer := ReportRenderer{Root: module, Directory: filepath.Join(t.TempDir(), "report-renderer"), Store: store, Node: node, PNPM: "/bin/true",
		NodeRuntime: t.TempDir(), Env: os.Environ(),
		Sandbox: GoBinary{Module: module, Package: "./cmd/launcher", Name: "launcher", Go: goForTest(t), Node: node, Env: append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")}}
	if err := renderer.Produce(); err != nil {
		t.Fatal(err)
	}
	root, err := os.ReadFile(seen)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(store.RootPath(strings.TrimSpace(string(root))), "launcher")); err != nil {
		t.Fatalf("the packaging script was not given the launcher's store root: %v", err)
	}
}

// The release inventory runs its generator with both outputs beside each
// other in the SBOM's directory and the scanners its requirements publish.
func TestReleaseInventoryGenerates(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(t.TempDir(), "release-artifacts")
	node := filepath.Join(root, "node")
	writeExecutable(t, node, `#!/bin/sh
[ "$1" = ./tools/release-evidence/generate-sbom-license-evidence.mjs ] || exit 91
[ "$CYCLONEDX_GOMOD_BIN" = /tools/cyclonedx ] && [ "$SYFT_BIN" = /tools/syft ] || exit 92
[ "$(dirname "$SBOM_ARTIFACT")" = "$RELEASE_ARTIFACT_DIR" ] && [ "$(dirname "$LICENSE_REPORT_ARTIFACT")" = "$RELEASE_ARTIFACT_DIR" ] || exit 93
printf sbom >"$SBOM_ARTIFACT" && printf licenses >"$LICENSE_REPORT_ARTIFACT"
`)
	inventory := ReleaseInventory{Root: root, SBOM: filepath.Join(directory, "sbom.cyclonedx.json"), Node: node, PNPM: "/bin/true",
		NodeRuntime: t.TempDir(), Go: "/bin/true", CycloneDX: "/tools/cyclonedx", Syft: "/tools/syft", Env: os.Environ()}
	if err := inventory.Generate(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"sbom.cyclonedx.json": "sbom", LicenseReportName: "licenses"} {
		if data, err := os.ReadFile(filepath.Join(directory, name)); err != nil || string(data) != want {
			t.Fatalf("%s holds %q, %v", name, data, err)
		}
	}
}

// Test service images are reused only while the receipt records the binary
// that names them and every image is present in the Docker store.
func TestTestServiceImagesReuseNeedsEveryImage(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, testServiceImagesImplementation, "producer\n")
	bin := filepath.Join(root, "bin")
	pulls := filepath.Join(root, "pulls")
	absent := filepath.Join(root, "absent")
	binary := filepath.Join(bin, "cartulary-test-services")
	writeExecutable(t, binary, `#!/bin/sh
case "$1" in
  images) printf 'postgres@sha256:1\nseaweedfs@sha256:2\n' ;;
  warm-images) echo x >>`+pulls+` ;;
  *) exit 91 ;;
esac
`)
	writeExecutable(t, filepath.Join(bin, "docker"), `#!/bin/sh
[ "$1 $2" = "image inspect" ] || exit 91
grep -qx "$3" `+absent+` 2>/dev/null && exit 1
exit 0
`)
	images := TestServiceImages{Root: root, Receipt: filepath.Join(root, "test-service-images", "installation.json"), Binary: binary,
		Env: append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))}
	lock := filepath.Join(root, "test-service-images.lock")
	count := func() int { return installCount(t, pulls) }
	if reused, err := Install(images, lock, Controls{}, false); err != nil || reused || count() != 1 {
		t.Fatalf("first warmup reused %v, %v", reused, err)
	}
	if reused, err := Install(images, lock, Controls{}, false); err != nil || !reused || count() != 1 {
		t.Fatalf("present images were not reused: %v, %v", reused, err)
	}
	if err := os.WriteFile(absent, []byte("seaweedfs@sha256:2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if reused, err := Install(images, lock, Controls{}, false); err != nil || reused || count() != 2 {
		t.Fatalf("a missing image was reused: %v, %v", reused, err)
	}
	if err := os.Remove(absent); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, binary, "#!/bin/sh\n[ \"$1\" = images ] && echo other@sha256:3\n[ \"$1\" = warm-images ] && echo x >>"+pulls+"\nexit 0\n")
	if reused, err := Install(images, lock, Controls{}, false); err != nil || reused || count() != 3 {
		t.Fatalf("another test-services binary was reused: %v, %v", reused, err)
	}
}

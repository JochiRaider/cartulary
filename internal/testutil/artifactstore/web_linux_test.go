//go:build linux

package artifactstore

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A web bundle's key covers every repository file under its inputs, its mode
// and the VITE_ settings it sees; documentation and files outside its inputs
// never enter it.
func TestWebBundleKey(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	writeFile(t, root, webBundleImplementation, "producer\n")
	writeFile(t, root, "apps/web/src/main.ts", "export const one = 1;\n")
	writeFile(t, root, "apps/web/index.html", "<html></html>\n")
	writeFile(t, root, "packages/ui/src/index.ts", "export {};\n")
	writeFile(t, root, "pnpm-lock.yaml", "lockfileVersion: 9\n")
	writeFile(t, root, "node", "#!/bin/sh\necho v0.0.0-fixture\n")
	if err := os.Chmod(filepath.Join(root, "node"), 0o755); err != nil {
		t.Fatal(err)
	}
	bundle := WebBundle{Root: root, Inputs: []string{"apps/web", "packages/ui", "pnpm-lock.yaml"}, Mode: "production",
		Node: filepath.Join(root, "node"), PNPM: "pnpm", NodeRuntime: root, PNPMVersion: "1.0.0", Env: []string{"PATH=/usr/bin:/bin"}}
	key := func(b WebBundle) string {
		t.Helper()
		k, err := b.Key()
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	base := key(bundle)
	if key(bundle) != base {
		t.Fatal("an unchanged bundle changed its key")
	}
	for name, change := range map[string]func(){
		"documentation":         func() { writeFile(t, root, "apps/web/README.md", "notes\n") },
		"a file outside inputs": func() { writeFile(t, root, "cmd/server/main.go", "package main\n") },
		"an ignored file": func() {
			writeFile(t, root, ".gitignore", "dist/\n")
			writeFile(t, root, "apps/web/dist/index.html", "built\n")
		},
	} {
		change()
		if name == "an ignored file" {
			base = key(bundle) // the .gitignore itself is outside the inputs
		}
		if key(bundle) != base {
			t.Fatalf("%s changed the key", name)
		}
	}
	for name, change := range map[string]func(){
		"a source edit":       func() { writeFile(t, root, "apps/web/src/main.ts", "export const one = 2;\n") },
		"a new package file":  func() { writeFile(t, root, "packages/ui/src/extra.ts", "export {};\n") },
		"the lockfile":        func() { writeFile(t, root, "pnpm-lock.yaml", "lockfileVersion: 10\n") },
		"the producer source": func() { writeFile(t, root, webBundleImplementation, "changed\n") },
	} {
		previous := key(bundle)
		change()
		if key(bundle) == previous {
			t.Fatalf("%s did not change the key", name)
		}
	}
	current := key(bundle)
	measurement := bundle
	measurement.Mode = "measurement"
	vite := bundle
	vite.Env = append(vite.Env, "VITE_FEATURE=on")
	for name, other := range map[string]WebBundle{"the mode": measurement, "a VITE_ setting": vite} {
		if key(other) == current {
			t.Fatalf("%s kept the key", name)
		}
	}
	// Reporting never changes the bundle: the arguments carry hidden source
	// maps whatever the output mode.
	if args := bundle.buildArgs(); !slices.Equal(args[len(args)-4:len(args)-2], []string{"--sourcemap", "hidden"}) {
		t.Fatalf("build arguments %v", args)
	}
}

// Embedded assets are keyed by the bundle they package, the client support
// source and the embedder's import closure; a linked bundle entry is refused.
func TestEmbeddedAssetsKey(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.test/embed\n\ngo 1.22\n")
	writeFile(t, root, "go.sum", "")
	writeFile(t, root, embeddedAssetsImplementation, "producer\n")
	writeFile(t, root, "tools/embedwebassets/main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, root, "contracts/client-support.json", "{}\n")
	bundle := t.TempDir()
	writeFile(t, bundle, "index.html", "<html></html>\n")
	writeFile(t, bundle, "assets/app.js", "one\n")
	goExecutable := os.Getenv("GO")
	if goExecutable == "" {
		var err error
		if goExecutable, err = exec.LookPath("go"); err != nil {
			t.Fatal("the go command is unavailable:", err)
		}
	}
	assets := EmbeddedAssets{Root: root, Source: bundle, ClientSupport: "contracts/client-support.json", Go: goExecutable, Env: append(os.Environ(), "GOWORK=off")}
	key := func() string {
		t.Helper()
		k, err := assets.Key()
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	for name, change := range map[string]func(){
		"a bundle file":      func() { writeFile(t, bundle, "assets/app.js", "two\n") },
		"a new bundle file":  func() { writeFile(t, bundle, "assets/app.js.map", "{}\n") },
		"the client support": func() { writeFile(t, root, "contracts/client-support.json", "{\"a\":1}\n") },
		"the embedder": func() {
			writeFile(t, root, "tools/embedwebassets/main.go", "package main\n\nfunc main() { println() }\n")
		},
	} {
		previous := key()
		change()
		if key() == previous {
			t.Fatalf("%s did not change the key", name)
		}
	}
	if err := os.Symlink("index.html", filepath.Join(bundle, "linked.html")); err != nil {
		t.Fatal(err)
	}
	if _, err := assets.Key(); err == nil {
		t.Fatal("a linked bundle entry was accepted")
	}
}

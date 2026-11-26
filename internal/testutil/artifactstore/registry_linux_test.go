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
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

const repositoryRoot = "../../.."

func testWorkspace(t *testing.T, env ...string) *Workspace {
	t.Helper()
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	base := []string{"NODE_BIN=/bin/true", "GO=/bin/true", "NODE_RUNTIME_DIR=/node", "PNPM=/node/bin/pnpm",
		"GO_CACHE_DIR=" + filepath.Join(scratch, "build"), "GO_MOD_CACHE_DIR=" + filepath.Join(scratch, "mod"), "GO_TMP_DIR=" + filepath.Join(scratch, "tmp")}
	w, err := OpenWorkspace(root, append(base, env...))
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// Every registered producer plans: its kind, binding and default path come
// from the registry, the toolchain pins and the workspace layout. A binding
// set to another path names an executable the caller provides.
func TestRegisteredProducersPlan(t *testing.T) {
	w := testWorkspace(t)
	registry, err := ReadRegistry(w.Root)
	if err != nil {
		t.Fatal(err)
	}
	if len(registry.Producers) == 0 {
		t.Fatal("the registry declares no producer")
	}
	for _, d := range registry.Producers {
		switch d.Kind {
		case "web_bundle":
			// A bundle publishes no view; its root becomes the run's
			// frontend artifact.
			if _, views, err := w.Plan(d); err != nil || len(views) != 0 {
				t.Fatalf("%s: views %v, %v", d.ID, views, err)
			}
			continue
		case "embedded_assets":
			// Packaging reads the run's frontend artifact, so it plans only
			// inside a graph.
			if _, _, err := w.Plan(d); err == nil || !strings.Contains(err.Error(), "graph") {
				t.Fatalf("%s planned outside a graph: %v", d.ID, err)
			}
			if name, err := w.DefaultName(d); err != nil || name != EmbeddedAssetNames[0] {
				t.Fatalf("%s default %q, %v", d.ID, name, err)
			}
			continue
		}
		producer, views, err := w.Plan(d)
		if err != nil {
			t.Fatalf("%s: %v", d.ID, err)
		}
		directory, err := w.Directory(d.Directory)
		if err != nil {
			t.Fatal(err)
		}
		name, err := w.DefaultName(d)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, view := range views {
			found = found || view == filepath.Join(directory, name)
		}
		if !found || producer.ID() == "" {
			t.Fatalf("%s publishes %v, not its default %s", d.ID, views, filepath.Join(directory, name))
		}
	}
	sqlc, _ := registry.Lookup("sqlc")
	if name, err := w.DefaultName(sqlc); err != nil || !strings.HasPrefix(name, "sqlc-v") {
		t.Fatalf("sqlc's default name %q, %v", name, err)
	}
	server, _ := registry.Lookup("server")
	if _, views, err := w.Plan(server); err != nil || views["server.crypto.json"] != views["server"]+".crypto.json" {
		t.Fatalf("the server's receipt is not beside it: %v, %v", views, err)
	}
	view, err := w.View(server)
	if err != nil {
		t.Fatal(err)
	}
	if _, provided, err := testWorkspace(t, "SERVER_BIN="+view).Provided(server); err != nil || provided {
		t.Fatalf("the default path counted as provided: %v", err)
	}
	link := filepath.Join(t.TempDir(), "checkout")
	if err := os.Symlink(w.Root, link); err != nil {
		t.Fatal(err)
	}
	if _, provided, err := testWorkspace(t, "SERVER_BIN="+filepath.Join(link, "build", "bin", "server")).Provided(server); err != nil || provided {
		t.Fatalf("the default path through a linked checkout counted as provided: %v", err)
	}
	own := filepath.Join(t.TempDir(), "server")
	if err := os.WriteFile(own, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if path, provided, err := testWorkspace(t, "SERVER_BIN="+own).Provided(server); err != nil || !provided || path != own {
		t.Fatalf("a caller's executable was not provided: %q %v %v", path, provided, err)
	}
	for _, binding := range []string{"SERVER_BIN=relative/server", "SERVER_BIN=/does-not-exist/server"} {
		if _, _, err := testWorkspace(t, binding).Provided(server); err == nil {
			t.Fatalf("%s was accepted", binding)
		}
	}
}

// A declaration carries exactly its kind's fields.
func TestDeclarationsAreValidatedByKind(t *testing.T) {
	valid := map[string]Declaration{
		"go_binary":       {ID: "b", Kind: "go_binary", Binding: "B_BIN", Directory: "binaries", Package: "./cmd/b", Name: "b"},
		"go_tool":         {ID: "t", Kind: "go_tool", Binding: "T_BIN", Directory: "toolbin", Tool: "sqlc"},
		"shellcheck":      {ID: "s", Kind: "shellcheck", Binding: "S_BIN", Directory: "toolbin", Archives: map[string]string{"a": "b"}},
		"web_bundle":      {ID: "w", Kind: "web_bundle", Target: "build-web", Mode: "production", Inputs: []string{"apps/web"}},
		"embedded_assets": {ID: "e", Kind: "embedded_assets", Binding: "E_ZIP", Directory: "embedded_assets", Source: "build-web", ClientSupport: "c.json"},
	}
	for kind, d := range valid {
		if err := d.validate(); err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
	}
	invalid := map[string]Declaration{
		"an unknown kind":           {ID: "x", Kind: "script", Binding: "X", Directory: "toolbin"},
		"a binary with a tool":      {ID: "b", Kind: "go_binary", Binding: "B_BIN", Directory: "binaries", Package: "./cmd/b", Name: "b", Tool: "sqlc"},
		"a binary outside ./":       {ID: "b", Kind: "go_binary", Binding: "B_BIN", Directory: "binaries", Package: "cmd/b", Name: "b"},
		"a tool with a package":     {ID: "t", Kind: "go_tool", Binding: "T_BIN", Directory: "toolbin", Tool: "sqlc", Package: "./x"},
		"shellcheck without digest": {ID: "s", Kind: "shellcheck", Binding: "S_BIN", Directory: "toolbin"},
		"a lowercase binding":       {ID: "t", Kind: "go_tool", Binding: "t_bin", Directory: "toolbin", Tool: "sqlc"},
		"a target without a label":  {ID: "t", Kind: "go_tool", Binding: "T_BIN", Directory: "toolbin", Tool: "sqlc", Target: "x"},
		"a bundle with a view":      {ID: "w", Kind: "web_bundle", Binding: "W", Directory: "web_dist", Target: "build-web", Mode: "production", Inputs: []string{"apps/web"}},
		"a bundle in another mode":  {ID: "w", Kind: "web_bundle", Target: "build-web", Mode: "development", Inputs: []string{"apps/web"}},
		"a bundle without inputs":   {ID: "w", Kind: "web_bundle", Target: "build-web", Mode: "production"},
		"assets without a source":   {ID: "e", Kind: "embedded_assets", Binding: "E_ZIP", Directory: "embedded_assets", ClientSupport: "c.json"},
		"a tool with inputs":        {ID: "t", Kind: "go_tool", Binding: "T_BIN", Directory: "toolbin", Tool: "sqlc", Inputs: []string{"x"}},
	}
	for name, d := range invalid {
		if d.validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

// A Go tool's key is its pinned module and version, its module directory's
// go.mod and go.sum, and the effective Go settings.
func TestGoToolKey(t *testing.T) {
	root := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(goToolImplementation, "producer\n")
	write("analysis/go.mod", "module example.test/analysis\n\ngo 1.22\n")
	write("analysis/go.sum", "")
	goExecutable := os.Getenv("GO")
	if goExecutable == "" {
		var err error
		if goExecutable, err = exec.LookPath("go"); err != nil {
			t.Fatal("the go command is unavailable:", err)
		}
	}
	tool := GoTool{Root: root, ModuleDir: "analysis", Module: "example.test/tool/cmd/tool@v1.0.0", Go: goExecutable, Env: append(os.Environ(), "GOWORK=off")}
	if tool.Name() != "tool" || tool.Version() != "v1.0.0" {
		t.Fatalf("name %q version %q", tool.Name(), tool.Version())
	}
	// A tool with a module directory builds from its locked graph, unchanged.
	if want := []string{"-C", "analysis", "install", "-mod=readonly", "example.test/tool/cmd/tool"}; !slices.Equal(tool.installArgs(), want) {
		t.Fatalf("install arguments %v, want %v", tool.installArgs(), want)
	}
	if want := []string{"install", "example.test/tool/cmd/tool@v1.0.0"}; !slices.Equal((GoTool{Module: tool.Module}).installArgs(), want) {
		t.Fatalf("pinned install arguments %v", (GoTool{Module: tool.Module}).installArgs())
	}
	key := func(t2 GoTool) string {
		t.Helper()
		k, err := t2.Key()
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	base := key(tool)
	if key(tool) != base {
		t.Fatal("an unchanged tool changed its key")
	}
	other := tool
	other.Module = "example.test/tool/cmd/tool@v1.0.1"
	if key(other) == base {
		t.Fatal("another version kept the key")
	}
	write("analysis/go.sum", "example.test/dependency v1.0.0 h1:x=\n")
	if key(tool) == base {
		t.Fatal("the module directory's go.sum did not change the key")
	}
	if _, err := (GoTool{Root: root, Module: "unpinned", Go: goExecutable}).Key(); err == nil {
		t.Fatal("an unpinned module was accepted")
	}
}

// ShellCheck is extracted only from an archive with the committed digest: a
// kept archive is reused, an altered one is fetched again, and a download
// with another digest is never kept.
func TestShellCheckArchiveIsVerified(t *testing.T) {
	scratch := t.TempDir()
	tree := filepath.Join(scratch, "tree", "shellcheck-v9.9.9")
	if err := os.MkdirAll(tree, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree, "shellcheck"), []byte("#!/bin/sh\necho fixture\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	archiveName := "shellcheck-v9.9.9.linux.fixture.tar.xz"
	built := filepath.Join(scratch, archiveName)
	if output, err := exec.Command("tar", "-cJf", built, "-C", filepath.Join(scratch, "tree"), "shellcheck-v9.9.9").CombinedOutput(); err != nil {
		t.Fatalf("tar: %v: %s", err, output)
	}
	content, err := os.ReadFile(built)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	var served atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served.Add(1)
		if r.URL.Path != "/v9.9.9/"+archiveName {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()
	archives := filepath.Join(scratch, "archives")
	producer := ShellCheck{Version: "9.9.9", Archive: archiveName, SHA256: hex.EncodeToString(sum[:]), Archives: archives, Releases: server.URL}
	build := func() error {
		t.Helper()
		staging := t.TempDir()
		if err := producer.Build(staging); err != nil {
			return err
		}
		if data, err := os.ReadFile(filepath.Join(staging, "shellcheck")); err != nil || !strings.Contains(string(data), "fixture") {
			t.Fatalf("staged shellcheck %q, %v", data, err)
		}
		if entries, _ := os.ReadDir(staging); len(entries) != 1 {
			t.Fatalf("staging holds %v", entries)
		}
		return nil
	}
	if err := build(); err != nil || served.Load() != 1 {
		t.Fatalf("first build %v, downloads %d", err, served.Load())
	}
	if err := build(); err != nil || served.Load() != 1 {
		t.Fatalf("a kept archive was downloaded again (%d), %v", served.Load(), err)
	}
	if err := os.WriteFile(filepath.Join(archives, archiveName), []byte("altered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := build(); err != nil || served.Load() != 2 {
		t.Fatalf("an altered archive was not replaced (%d), %v", served.Load(), err)
	}
	producer.SHA256 = strings.Repeat("0", 64)
	if err := os.Remove(filepath.Join(archives, archiveName)); err != nil {
		t.Fatal(err)
	}
	if err := producer.Build(t.TempDir()); err == nil {
		t.Fatal("an archive with another digest was extracted")
	}
	if entries, _ := os.ReadDir(archives); len(entries) != 0 {
		t.Fatalf("a refused download remained: %v", entries)
	}
}

// The generated Make include defaults each producer binding to the path the
// producer itself publishes at, so Make and a bare `produce` agree.
func TestGeneratedMakeDefaultsMatchProducers(t *testing.T) {
	w := testWorkspace(t)
	registry, err := ReadRegistry(w.Root)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(w.Root, "tools/task_surface.runtime.generated.mk"))
	if err != nil {
		t.Fatal(err)
	}
	defaults := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		if binding, value, ok := strings.Cut(line, " ?= $(CURDIR)/"); ok {
			defaults[binding] = filepath.Join(w.Root, filepath.FromSlash(value))
		}
	}
	for _, d := range registry.Producers {
		if d.Binding == "" {
			continue
		}
		view, err := w.View(d)
		if err != nil {
			t.Fatal(err)
		}
		if defaults[d.Binding] != view {
			t.Fatalf("Make defaults %s to %q; its producer publishes at %q", d.Binding, defaults[d.Binding], view)
		}
	}
}

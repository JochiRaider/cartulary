package harnessplan

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A row that runs a built binary depends on what the binary is built from:
// its Go import closure without tests and, where that closure embeds a
// generated directory, the inputs the directory is generated from (P6-6).
func TestRuntimeBinaryInputs(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                       "module example.com/m\n\ngo 1.27\n",
		"go.sum":                       "",
		"cmd/tool/main.go":             "package main\n\nimport _ \"example.com/m/web\"\n",
		"cmd/tool/main_test.go":        "package main\n\nimport _ \"example.com/m/testonly\"\n",
		"web/web.go":                   "package web\n",
		"web/fallback/index.html":      "<html></html>\n",
		"testonly/testonly.go":         "package testonly\n",
		"cmd/plain/main.go":            "package main\n",
		"ui/app.ts":                    "export {};\n",
		"ui/client.json":               "{}\n",
		"elsewhere/unrelated.go":       "package elsewhere\n",
		"tools/harness_producers.json": `{"producers":[{"id":"tool","kind":"go_binary","binding":"TOOL_BIN","package":"./cmd/tool"},{"id":"plain","kind":"go_binary","binding":"PLAIN_BIN","package":"./cmd/plain"},{"id":"bundle","kind":"web_bundle","target":"build-bundle","inputs":["ui/app.ts"]},{"id":"assets","kind":"embedded_assets","directory":"embedded","source":"build-bundle","client_support":"ui/client.json"}]}`,
		"tools/workspace_layout.json":  `{"directories":[{"id":"embedded","path":"web/generated"}]}`,
	}
	var source []SourceInput
	for relative, content := range files {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256([]byte(content))
		source = append(source, SourceInput{Path: relative, Kind: "file", Mode: "0644", ByteDigest: "sha256:" + hex.EncodeToString(sum[:])})
	}
	slices.SortFunc(source, func(a, b SourceInput) int { return strings.Compare(a.Path, b.Path) })
	in := &Inputs{Topology: Topology{RuntimeBinaries: []RuntimeBinary{{ID: "tool", OutputMakeVariable: "TOOL_BIN"}, {ID: "plain", OutputMakeVariable: "PLAIN_BIN"}, {ID: "stray", OutputMakeVariable: "STRAY_BIN"}}}}
	planner := func() *cachePlanner {
		return &cachePlanner{root: root, in: in, source: source, imports: map[string][]string{}, entries: map[int][]byte{}, literals: map[int][]string{}, scans: loadSourceScans(root, false)}
	}
	closure := func(ids ...string) ([]string, []string, error) {
		p := planner()
		set := map[int]bool{}
		packages, err := p.binaryInputs(ids, set)
		var paths []string
		for index := range set {
			paths = append(paths, source[index].Path)
		}
		slices.Sort(paths)
		return packages, paths, err
	}
	packages, paths, err := closure("tool")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"cmd/tool", "web"}; !slices.Equal(packages, want) {
		t.Fatalf("packages %v, want %v", packages, want)
	}
	// The binary's packages without tests, their module files, the embedded
	// assets' declared inputs and the declarations that say so.
	if want := []string{"cmd/tool/main.go", "go.mod", "go.sum", "tools/harness_producers.json", "tools/workspace_layout.json", "ui/app.ts", "ui/client.json", "web/fallback/index.html", "web/web.go"}; !slices.Equal(paths, want) {
		t.Fatalf("closure %v\nwant    %v", paths, want)
	}
	// A binary that embeds nothing brings only its own packages.
	if _, paths, err := closure("plain"); err != nil || slices.Contains(paths, "ui/app.ts") || !slices.Contains(paths, "cmd/plain/main.go") {
		t.Fatalf("plain binary closure %v, %v", paths, err)
	}
	// An unknown binary, or one with no Go binary producer, cannot be keyed.
	for _, id := range []string{"absent", "stray"} {
		if _, _, err := closure(id); err == nil {
			t.Fatalf("binary %s was keyed", id)
		}
	}
}

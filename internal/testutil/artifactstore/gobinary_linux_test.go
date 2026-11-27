//go:build linux

package artifactstore

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// A Go binary's key is its import closure: every non-test source and embedded
// file of every package it imports, its own declaration and the files that
// shape its receipt (G47).
func TestGoBinaryKeyCoversItsImportClosure(t *testing.T) {
	module := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(module, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.test/closure\n\ngo 1.22\n")
	write("go.sum", "")
	write("tools/toolchain_pins.json", `{"go_toolchain":"go1.22","cryptographic_module":{"selector":"fixture"}}`)
	for _, name := range append([]string{producerImplementation}, receiptWriter...) {
		write(name, "producer and receipt writer\n")
	}
	write("app/main.go", "package main\n\nimport \"example.test/closure/data\"\n\nfunc main() { _ = data.Schema }\n")
	write("data/data.go", "package data\n\nimport _ \"embed\"\n\n//go:embed schema.sql\nvar Schema string\n")
	write("data/schema.sql", "create table one();\n")
	write("data/data_test.go", "package data\n")
	node := filepath.Join(module, "node")
	write("node", "#!/bin/sh\necho v0.0.0-fixture\n")
	if err := os.Chmod(node, 0o755); err != nil {
		t.Fatal(err)
	}
	goExecutable := os.Getenv("GO")
	if goExecutable == "" {
		var err error
		if goExecutable, err = exec.LookPath("go"); err != nil {
			t.Fatal("the go command is unavailable:", err)
		}
	}
	binary := GoBinary{Module: module, Package: "./app", Name: "app", Go: goExecutable, Node: node, Env: append(os.Environ(), "GOWORK=off")}
	key := func(b GoBinary) string {
		t.Helper()
		k, err := b.Key()
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	closure, err := binary.closure()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"app/main.go", "data/data.go", "data/schema.sql"}; !slices.Equal(closure, want) {
		t.Fatalf("closure %v, want %v", closure, want)
	}
	base := key(binary)
	if key(binary) != base {
		t.Fatal("an unchanged closure changed its key")
	}
	write("data/data_test.go", "package data\n\nfunc helper() {}\n")
	if key(binary) != base {
		t.Fatal("a test file entered the key")
	}
	changed := map[string]func(){
		"an embedded file":   func() { write("data/schema.sql", "create table two();\n") },
		"a new source file":  func() { write("data/extra.go", "package data\n\nconst Extra = 1\n") },
		"the receipt writer": func() { write(receiptWriter[0], "changed writer\n") },
		"the producer":       func() { write(producerImplementation, "changed producer\n") },
		"the pinned module": func() {
			write("tools/toolchain_pins.json", `{"go_toolchain":"go1.22","cryptographic_module":{"selector":"other"}}`)
		},
		"the module requirement": func() { write("go.sum", "example.test/dependency v1.0.0 h1:x=\n") },
	}
	for _, name := range []string{"an embedded file", "a new source file", "the receipt writer", "the producer", "the pinned module", "the module requirement"} {
		previous := key(binary)
		changed[name]()
		if key(binary) == previous {
			t.Fatalf("%s did not change the key", name)
		}
	}
	current := key(binary)
	for name, declaration := range map[string]GoBinary{
		"tags": func() GoBinary { b := binary; b.Tags = "fixture"; return b }(),
		"name": func() GoBinary { b := binary; b.Name = "other"; return b }(),
	} {
		if key(declaration) == current {
			t.Fatalf("a different %s kept the key", name)
		}
	}
}

// The key memo (P6-5) answers a key without running Go or Node while
// everything those commands read is unchanged, and the key is always the one
// the commands would give.
func TestGoBinaryKeyMemo(t *testing.T) {
	defer func(margin time.Duration) { memoSettleMargin = margin }(memoSettleMargin)
	memoSettleMargin = -time.Hour // every file counts as settled
	module, memo := t.TempDir(), t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(module, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.test/memo\n\ngo 1.22\n")
	write("go.sum", "")
	write("tools/toolchain_pins.json", `{"go_toolchain":"go1.22","cryptographic_module":{"selector":"fixture"}}`)
	for _, name := range append([]string{producerImplementation}, receiptWriter...) {
		write(name, "producer and receipt writer\n")
	}
	write("app/main.go", "package main\n\nimport \"example.test/memo/data\"\n\nfunc main() { _ = data.Schema }\n")
	write("data/data.go", "package data\n\nimport \"embed\"\n\n//go:embed schema\nvar Schema embed.FS\n")
	write("data/schema/one.sql", "create table one();\n")
	realGo := os.Getenv("GO")
	if realGo == "" {
		var err error
		if realGo, err = exec.LookPath("go"); err != nil {
			t.Fatal("the go command is unavailable:", err)
		}
	}
	// Go and Node record each call, so a memo answer is visible.
	calls := filepath.Join(t.TempDir(), "calls")
	tools := t.TempDir()
	for name, body := range map[string]string{
		"go":   "#!/bin/sh\necho go >>" + calls + "\nexec " + realGo + " \"$@\"\n",
		"node": "#!/bin/sh\necho node >>" + calls + "\necho v0.0.0-fixture\n",
	} {
		if err := os.WriteFile(filepath.Join(tools, name), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	binary := GoBinary{Module: module, Package: "./app", Name: "app", Go: filepath.Join(tools, "go"), Node: filepath.Join(tools, "node"), Env: append(os.Environ(), "GOWORK=off")}
	ran := func() int {
		data, _ := os.ReadFile(calls)
		_ = os.Remove(calls)
		return strings.Count(string(data), "\n")
	}
	check := func(label string, b GoBinary, memoized bool) string {
		t.Helper()
		got, err := b.MemoKey(memo)
		if err != nil {
			t.Fatal(err)
		}
		if commands := ran(); memoized != (commands == 0) {
			t.Fatalf("%s: memoized %v, but %d commands ran", label, memoized, commands)
		}
		want, err := b.Key()
		if err != nil {
			t.Fatal(err)
		}
		ran()
		if got != want {
			t.Fatalf("%s: the memo key %s differs from the key %s", label, got, want)
		}
		return got
	}
	base := check("the first key", binary, false)
	if check("an unchanged closure", binary, true) != base {
		t.Fatal("the memo changed the key")
	}
	write("data/data_test.go", "package data\n")
	check("a test file", binary, true)
	for _, change := range []struct {
		label string
		apply func()
	}{
		{"a source edit", func() {
			write("data/data.go", "package data\n\nimport \"embed\"\n\n//go:embed schema\nvar Schema embed.FS\n\nconst Edited = 1\n")
		}},
		{"a file a build constraint excludes", func() { write("data/other.go", "//go:build ignore\n\npackage data\n") }},
		{"a new file in an embedded directory", func() { write("data/schema/two.sql", "create table two();\n") }},
		{"a new embedded subdirectory", func() { write("data/schema/more/three.sql", "create table three();\n") }},
		{"the module requirement", func() { write("go.sum", "example.test/dependency v1.0.0 h1:x=\n") }},
		{"a workspace file", func() { write("go.work", "go 1.22\n\nuse .\n") }},
	} {
		change.apply()
		check(change.label, binary, false)
		check(change.label+", again", binary, true)
	}
	other := binary
	other.Env = append(slices.Clone(binary.Env), "GOFLAGS=-tags=other")
	check("another build environment", other, false)
	check("the first environment", binary, true)
}

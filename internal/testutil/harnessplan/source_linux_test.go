//go:build linux

package harnessplan

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestSourceSnapshot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatal("git is required to enumerate repository inputs")
	}
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-c", "user.name=planner", "-c", "user.email=planner@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	write := func(relative, content string, mode os.FileMode) {
		t.Helper()
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	run("init", "-q")
	write(executableInputPolicyPath, `{"schema_id":"cartulary.executable_input_policy.v1","restricted_roots":["notes"],"documentation_only_sources":[],"machine_evidence_roots":[]}`, 0o644)
	write("main.go", "package main\n", 0o644)
	write("tools/run.sh", "#!/bin/sh\n", 0o755)
	write("notes/guide.go", "package notes\n", 0o644)
	write("guide.md", "# guide\n", 0o644)
	write("pkg/README", "read me\n", 0o644)
	write("pkg/Readme.txt", "read me\n", 0o644)
	write(".gitignore", "ignored.go\n", 0o644)
	write("ignored.go", "package ignored\n", 0o644)
	write("removed.go", "package removed\n", 0o644)
	if err := os.Symlink("main.go", filepath.Join(root, "link.go")); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-q", "-m", "fixture")
	if err := os.Remove(filepath.Join(root, "removed.go")); err != nil {
		t.Fatal(err)
	}
	write("untracked.go", "package untracked\n", 0o644)

	t.Run("enumerates executable source and its state", func(t *testing.T) {
		snapshot, err := CaptureSource(root)
		if err != nil {
			t.Fatal(err)
		}
		// Git does not ignore the graph cache here, so no memo is kept: it
		// would become a source input itself.
		if _, err := os.Stat(filepath.Join(root, GraphCacheDirectory)); snapshot.Memoized || !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("memoized %v, graph cache %v", snapshot.Memoized, err)
		}
		var paths []string
		kinds := map[string]SourceInput{}
		for _, input := range snapshot.Inputs {
			paths = append(paths, input.Path)
			kinds[input.Path] = input
		}
		want := []string{".gitignore", "link.go", "main.go", "tools/executable_input_policy.json", "tools/run.sh", "untracked.go"}
		if !slices.Equal(paths, want) {
			t.Fatalf("paths = %v, want %v", paths, want)
		}
		if kinds["link.go"].Kind != "symlink" || kinds["tools/run.sh"].Mode != "0755" || kinds["main.go"].Mode != "0644" {
			t.Fatalf("kinds and modes = %+v", kinds)
		}
		if snapshot.State != "dirty" || len(snapshot.Commit) != 40 {
			t.Fatalf("state %s, commit %q", snapshot.State, snapshot.Commit)
		}
		again, err := CaptureSource(root)
		if err != nil || !slices.Equal(again.Inputs, snapshot.Inputs) {
			t.Fatal("snapshot is not deterministic")
		}
	})

	t.Run("a committed worktree is clean", func(t *testing.T) {
		run("add", "-A")
		run("commit", "-q", "-m", "settle")
		snapshot, err := CaptureSource(root)
		if err != nil || snapshot.State != "clean" {
			t.Fatalf("state %s, %v", snapshot.State, err)
		}
		write("guide.md", "# changed guide\n", 0o644)
		if snapshot, err := CaptureSource(root); err != nil || snapshot.State != "clean" {
			t.Fatalf("documentation changed the executable state: %s, %v", snapshot.State, err)
		}
	})

	t.Run("a settled file keeps its digest until its identity changes", func(t *testing.T) {
		defer func(margin time.Duration) { settleMargin = margin }(settleMargin)
		settleMargin = -time.Hour // every file counts as settled
		write(".gitignore", "ignored.go\n.cache/\n", 0o644)
		write("memo.txt", "one\n", 0o644)
		digestOf := func(snapshot SourceSnapshot) string {
			for _, input := range snapshot.Inputs {
				if input.Path == "memo.txt" {
					return input.ByteDigest
				}
			}
			t.Fatal("memo.txt is not in the snapshot")
			return ""
		}
		first, err := CaptureSource(root)
		if err != nil || !first.Memoized {
			t.Fatalf("memoized %v, %v", first.Memoized, err)
		}
		// A recorded digest is used while the file's identity is unchanged.
		memo := filepath.Join(root, GraphCacheDirectory, digestMemoName)
		data, err := os.ReadFile(memo)
		if err != nil {
			t.Fatal(err)
		}
		forged := "sha256:" + strings.Repeat("0", 64)
		if err := os.WriteFile(memo, []byte(strings.Replace(string(data), digestOf(first), forged, 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		if second, err := CaptureSource(root); err != nil || digestOf(second) != forged {
			t.Fatalf("the memo was not consulted: %v", err)
		}
		// A change with the same size and modification time still moves the
		// change time, so the file is hashed again.
		info, err := os.Stat(filepath.Join(root, "memo.txt"))
		if err != nil {
			t.Fatal(err)
		}
		write("memo.txt", "two\n", 0o644)
		if err := os.Chtimes(filepath.Join(root, "memo.txt"), info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
		third, err := CaptureSource(root)
		sum := sha256.Sum256([]byte("two\n"))
		if err != nil || digestOf(third) != "sha256:"+hex.EncodeToString(sum[:]) {
			t.Fatalf("a changed file kept its recorded digest: %v", err)
		}
	})

	t.Run("a nested repository fails the snapshot", func(t *testing.T) {
		nested := filepath.Join(root, "vendor/other")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		command := exec.Command("git", "init", "-q")
		command.Dir = nested
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, output)
		}
		if _, err := CaptureSource(root); !errors.Is(err, ErrPlan) {
			t.Fatalf("CaptureSource = %v, want ErrPlan", err)
		}
	})
}

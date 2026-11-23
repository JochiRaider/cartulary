//go:build linux

package harnessruntime

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
)

func TestRuntimeGoTemporaryWork(t *testing.T) {
	base := filepath.Join(t.TempDir(), "go-tmp")
	writeOwner := func(t *testing.T, name string, owner harnessadmission.Proof) string {
		t.Helper()
		directory := filepath.Join(base, name)
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(owner)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, goTempOwnerFile), data, 0600); err != nil {
			t.Fatal(err)
		}
		return directory
	}
	exists := func(path string) bool { _, err := os.Lstat(path); return err == nil }

	t.Run("creates an owned invocation directory", func(t *testing.T) {
		path, err := newGoTemp(base)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Dir(path) != base || !strings.HasPrefix(filepath.Base(path), goTempPrefix) {
			t.Fatalf("invocation directory %s is not directly beneath %s", path, base)
		}
		owner, ok := readGoTempOwner(path)
		self, err := harnessadmission.Identity(os.Getpid())
		if err != nil || !ok || owner != self {
			t.Fatalf("owner=%+v ok=%v self=%+v err=%v", owner, ok, self, err)
		}
		if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0700 {
			t.Fatalf("invocation directory mode: %v %v", info, err)
		}
		if err := removeGoTemp(path); err != nil || exists(path) {
			t.Fatalf("removal: %v", err)
		}
	})

	t.Run("selects the memory-backed root only when eligible", func(t *testing.T) {
		machine := filepath.Join(t.TempDir(), "machine-go-tmp")
		uid := os.Geteuid()
		memory := func(available uint64) memoryStat {
			return func(string) (bool, uint64, error) { return true, available, nil }
		}
		fresh := func(t *testing.T) (string, string) {
			parent := t.TempDir()
			return parent, filepath.Join(parent, goTempMemoryName+strconv.Itoa(uid))
		}
		t.Run("eligible root is created private", func(t *testing.T) {
			parent, root := fresh(t)
			if got := selectGoTempBase(machine, parent, uid, memory(goTempMemoryMinimum)); got != root {
				t.Fatalf("base = %s, want %s", got, root)
			}
			if info, err := os.Lstat(root); err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
				t.Fatalf("root: %v %v", info, err)
			}
			path, err := newGoTemp(root)
			if err != nil || filepath.Dir(path) != root {
				t.Fatalf("invocation %s beneath %s: %v", path, root, err)
			}
		})
		for _, tc := range []struct {
			name    string
			stat    memoryStat
			prepare func(t *testing.T, root string)
			uid     int
		}{
			{name: "disk-backed parent", stat: func(string) (bool, uint64, error) { return false, goTempMemoryMinimum, nil }},
			{name: "capacity below the minimum", stat: memory(goTempMemoryMinimum - 1)},
			{name: "parent unavailable", stat: func(string) (bool, uint64, error) { return false, 0, os.ErrNotExist }},
			{name: "root grants group or other access", stat: memory(goTempMemoryMinimum), prepare: func(t *testing.T, root string) {
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(root, 0750); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "root is a symbolic link", stat: memory(goTempMemoryMinimum), prepare: func(t *testing.T, root string) {
				if err := os.Symlink(t.TempDir(), root); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "root is a file", stat: memory(goTempMemoryMinimum), prepare: func(t *testing.T, root string) {
				if err := os.WriteFile(root, nil, 0600); err != nil {
					t.Fatal(err)
				}
			}},
			// The root carries the invoking user's ID, so a directory another
			// user created under that name is foreign.
			{name: "root owned by another user", stat: memory(goTempMemoryMinimum), uid: uid + 1, prepare: func(t *testing.T, root string) {
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
			}},
		} {
			t.Run(tc.name+" falls back to GO_TMP_DIR", func(t *testing.T) {
				caller := uid
				if tc.uid != 0 {
					caller = tc.uid
				}
				parent := t.TempDir()
				root := filepath.Join(parent, goTempMemoryName+strconv.Itoa(caller))
				if tc.prepare != nil {
					tc.prepare(t, root)
				}
				if got := selectGoTempBase(machine, parent, caller, tc.stat); got != machine {
					t.Fatalf("base = %s, want fallback %s", got, machine)
				}
			})
		}
	})

	t.Run("reclaims only proven-dead invocation directories", func(t *testing.T) {
		child := exec.Command("sleep", "30")
		if err := child.Start(); err != nil {
			t.Fatal(err)
		}
		dead, err := harnessadmission.Identity(child.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		_ = child.Process.Kill()
		_ = child.Wait()
		self, err := harnessadmission.Identity(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		deadPublished := writeOwner(t, goTempPrefix+"dead", dead)
		deadStaging := writeOwner(t, goTempStaging+"dead", dead)
		live := writeOwner(t, goTempPrefix+"live", self)
		unproven := filepath.Join(base, goTempPrefix+"unproven")
		unrelated := filepath.Join(base, "go-build123")
		for _, directory := range []string{unproven, unrelated} {
			if err := os.MkdirAll(directory, 0700); err != nil {
				t.Fatal(err)
			}
		}
		// A dead owner's tree may hold read-only directories left by a test.
		sealed := filepath.Join(deadPublished, "TestSealed1", "001")
		if err := os.MkdirAll(sealed, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(sealed, 0500); err != nil {
			t.Fatal(err)
		}
		path, err := newGoTemp(base)
		if err != nil {
			t.Fatal(err)
		}
		for _, removed := range []string{deadPublished, deadStaging} {
			if exists(removed) {
				t.Fatalf("dead-owner directory survived: %s", removed)
			}
		}
		for _, kept := range []string{live, unproven, unrelated, path} {
			if !exists(kept) {
				t.Fatalf("reclamation removed %s", kept)
			}
		}
	})
}

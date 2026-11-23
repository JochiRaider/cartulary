//go:build linux

package harnessruntime

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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

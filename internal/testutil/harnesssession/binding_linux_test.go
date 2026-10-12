//go:build linux

package harnesssession

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"os"
	"path/filepath"
	"testing"
)

func testRuntimeBindings(t *testing.T) {
	t.Run("repository_retained_and_private_output_binding", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		caller, err := harnessadmission.Identity(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		for _, binding := range [][2]string{{t.TempDir(), run}, {repo, t.TempDir()}, {repo, filepath.Join(r.Root, "undeclared")}} {
			if _, err := BorrowRuntime(binding[0], binding[1], r.Root, r.RunID, r.LeaseID); err == nil {
				t.Fatal("foreign borrow accepted", binding)
			}
			if _, err := OwnRuntime(binding[0], binding[1], r.Root, r.RunID, r.LeaseID, caller, true); err == nil {
				t.Fatal("foreign recovery accepted", binding)
			}
		}
		output := filepath.Join(r.Root, "lifecycle", r.RunID)
		if err := r.AuthorizeOutputRoot(output); err != nil {
			t.Fatal(err)
		}
		b, err := BorrowRuntime(repo, output, r.Root, r.RunID, r.LeaseID)
		if err != nil {
			t.Fatal(err)
		}
		defer b.ReleaseHandles()
		if b.RunRoot != run || b.Repository != repo {
			t.Fatal("output scope replaced retained authority")
		}
		if err := b.AuthorizeOutputRoot(filepath.Join(r.Root, "caller-selected")); err == nil {
			t.Fatal("borrower extended authority")
		}
		if err := r.AuthorizeOutputRoot(t.TempDir()); err == nil {
			t.Fatal("foreign private output authorized")
		}
		alias := filepath.Join(t.TempDir(), "repository")
		if err := os.Symlink(repo, alias); err != nil {
			t.Fatal(err)
		}
		if _, err := NewRuntime(alias, run, "fixture", t.TempDir()); err == nil {
			t.Fatal("repository alias admitted")
		}
	})
}

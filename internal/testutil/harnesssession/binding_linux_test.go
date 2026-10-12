//go:build linux

package harnesssession

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func testRuntimeBindings(t *testing.T) {
	t.Run("fresh_recovery_output_rejects_forgery_and_concurrent_replay", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		a, err := r.AllocateRecoveryOutput()
		if err != nil {
			t.Fatal(err)
		}
		b, err := r.AllocateRecoveryOutput()
		if err != nil || a == b {
			t.Fatal("reused recovery output", err)
		}
		for _, attempt := range [][5]string{{repo, run, r.Root, r.RunID, r.LeaseID}, {repo, t.TempDir(), r.Root, r.RunID, r.LeaseID}, {t.TempDir(), a, r.Root, r.RunID, r.LeaseID}, {repo, a, r.Root, "foreign", r.LeaseID}, {repo, a, r.Root, r.RunID, "00000000-0000-0000-0000-000000000000"}} {
			if err := ClaimRecoveryOutput(attempt[0], attempt[1], attempt[2], attempt[3], attempt[4]); err == nil {
				t.Fatal("forged output accepted")
			}
		}
		var workers sync.WaitGroup
		results := make(chan error, 2)
		for range 2 {
			workers.Go(func() { results <- ClaimRecoveryOutput(repo, a, r.Root, r.RunID, r.LeaseID) })
		}
		workers.Wait()
		close(results)
		success := 0
		for err := range results {
			if err == nil {
				success++
			}
		}
		if success != 1 {
			t.Fatal("output claim not exclusive", success)
		}
		if err := ClaimRecoveryOutput(repo, a, r.Root, r.RunID, r.LeaseID); err == nil {
			t.Fatal("prior output reused")
		}
		if err := ClaimRecoveryOutput(repo, b, r.Root, r.RunID, r.LeaseID); err != nil {
			t.Fatal("fresh output blocked", err)
		}
		if err := os.RemoveAll(a); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(a, 0700); err != nil {
			t.Fatal(err)
		}
		if err := ClaimRecoveryOutput(repo, a, r.Root, r.RunID, r.LeaseID); err == nil {
			t.Fatal("detail removal lost consumed-output authority")
		}
	})
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

//go:build linux

package harnesssession

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func testRecoveryLocators(t *testing.T) {
	t.Run("exact_lookup_retains_authority_across_failed_close_and_is_read_only", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		scratch := filepath.Dir(filepath.Dir(r.Root))
		proof, err := r.RecoveryPath("pending-suite.json")
		if err != nil {
			t.Fatal(err)
		}
		if err := r.RecordDependentResource("managed_suite", proof, "pending", []string{}); err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); !errors.Is(err, ErrUnresolved) {
			t.Fatal("pending acquisition was not preserved", err)
		}
		file := filepath.Join(scratch, "recovery-index", locatorName(repo, run, r.RunID))
		before, err := harnesslocal.ReadFile(file, 32768, false)
		if err != nil {
			t.Fatal(err)
		}
		for range 2 {
			v, err := ResolveRecovery(repo, run, r.RunID, scratch)
			if err != nil || v.Root != r.Root || v.LeaseID != r.LeaseID || v.State != "active" {
				t.Fatal("exact authority lost", err)
			}
		}
		after, err := harnesslocal.ReadFile(file, 32768, false)
		if err != nil || string(before) != string(after) {
			t.Fatal("lookup changed receipt", err)
		}
		if _, err := ResolveRecovery(repo, run, "another-run", scratch); err == nil {
			t.Fatal("foreign run resolved")
		}
		if _, err := NewRuntime(repo, run, r.RunID, scratch); err == nil {
			t.Fatal("reused run replaced original locator")
		}
		if err := r.RecordResource("managed_suite", proof, "released"); err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			v, err := ResolveRecovery(repo, run, r.RunID, scratch)
			if err != nil || v.State != "settled" {
				t.Fatal("settled lookup is not idempotent", err)
			}
		}
	})
	t.Run("locator_does_not_replace_native_binding_or_admit_links", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		scratch := filepath.Dir(filepath.Dir(r.Root))
		directory := filepath.Join(scratch, "recovery-index")
		file := filepath.Join(directory, locatorName(repo, run, r.RunID))
		v, err := readLocator(directory, repo, run, r.RunID)
		if err != nil {
			t.Fatal(err)
		}
		original := v
		v.LeaseID = identifier()
		if err := privateJSON(file, recoveryLocatorSchema, v, true); err != nil {
			t.Fatal(err)
		}
		if _, err := ResolveRecovery(repo, run, r.RunID, scratch); err == nil {
			t.Fatal("mismatched lease admitted")
		}
		if err := privateJSON(file, recoveryLocatorSchema, original, true); err != nil {
			t.Fatal(err)
		}
		moved := file + ".original"
		if err := os.Rename(file, moved); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(moved, file); err != nil {
			t.Fatal(err)
		}
		if _, err := ResolveRecovery(repo, run, r.RunID, scratch); err == nil {
			t.Fatal("symlink locator admitted")
		}
		if err := os.Remove(file); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(moved, file); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("absence_does_not_invent_historical_cleanup_authority", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		scratch := filepath.Dir(filepath.Dir(r.Root))
		if err := os.Remove(filepath.Join(scratch, "recovery-index", locatorName(repo, run, r.RunID))); err != nil {
			t.Fatal(err)
		}
		if _, err := ResolveRecovery(repo, run, r.RunID, scratch); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("historical run received speculative authority", err)
		}
	})
}

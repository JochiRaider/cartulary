//go:build linux

package harnesssession

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func runtimeFixture(t *testing.T) (*Runtime, string, string) {
	t.Helper()
	root, run, scratch := t.TempDir(), t.TempDir(), t.TempDir()
	runtime, err := NewRuntime(root, run, "fixture", scratch)
	if err != nil {
		t.Fatal(err)
	}
	if harnesslocal.CheckDirectory(runtime.Directory, runtime.Directory.Name()) != nil {
		t.Fatal("runtime descriptor retained staging pathname")
	}
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Error(err)
		}
	})
	return runtime, root, run
}
func TestSessionRuntimeProof(t *testing.T) {
	testLifecycleGaps(t)
	testCorruptJournalDetail(t)
	testCleanupAttempts(t)
	testRuntimeBindings(t)
	t.Run("failed_registration_and_duplicate_journal_preserve_ownership", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		target, _ := r.RecoveryPath("browser-stack-leases", "fixture", "owner.json")
		if err := r.RecordResource("browser_stack", target, "pending"); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(filepath.Join(r.Root, "recovery"))
		if err != nil || len(entries) != 1 {
			t.Fatal(entries, err)
		}
		file := filepath.Join(r.Root, "recovery", entries[0].Name())
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.Chmod(file, 0644); err != nil {
			t.Fatal(err)
		}
		if err = r.RecordResource("browser_stack", target, "acquired"); err == nil {
			t.Fatal("unsafe journal replacement accepted")
		}
		if err = os.Chmod(file, 0600); err != nil {
			t.Fatal(err)
		}
		resources, err := r.Resources(64)
		if err != nil || len(resources) != 1 || resources[0].State != "pending" {
			t.Fatal("pending ownership lost", resources, err)
		}
		duplicate := filepath.Join(r.Root, "recovery", "duplicate.json")
		if err = os.WriteFile(duplicate, data, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = r.Resources(64); err == nil {
			t.Fatal("duplicate journal accepted")
		}
		if err = r.Close(); err == nil {
			t.Fatal("unresolved root removed")
		}
		if err = os.Remove(duplicate); err != nil {
			t.Fatal(err)
		}
		if err = r.RecordResource("browser_stack", target, "released"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("preservation_and_janitor", testFixturePreservation)
	t.Run("borrow_and_recovery_closure", func(t *testing.T) {
		r, root, run := runtimeFixture(t)
		borrowed, err := BorrowRuntime(root, run, r.Root, r.RunID, r.LeaseID)
		if err != nil {
			t.Fatal(err)
		}
		if err := borrowed.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(r.Root); err != nil {
			t.Fatal("borrower removed owner root", err)
		}
		if _, err := BorrowRuntime(root, run, r.Root, r.RunID, "forged"); err == nil {
			t.Fatal("forged lease accepted")
		}
		if _, err := r.RecoveryPath(".."); err == nil {
			t.Fatal("escaping path accepted")
		}
		proof, _ := r.RecoveryPath("service", "suite-lease.json")
		if err := r.RecordResource("managed_suite", proof, "pending"); err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err == nil {
			t.Fatal("unresolved ownership removed")
		}
		resources, err := r.Resources(4096)
		if err != nil || len(resources) != 1 {
			t.Fatal(resources, err)
		}
		if err := r.RecordResource("managed_suite", proof, "released"); err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Lstat(r.Root); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("owned root was not removed", err)
		}
	})
	t.Run("replacement_is_preserved", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		original := r.Root + "-original"
		if err := os.Rename(r.Root, original); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(r.Root, 0700); err != nil {
			t.Fatal(err)
		}
		sentinel := filepath.Join(r.Root, "sentinel")
		if err := os.WriteFile(sentinel, []byte("foreign"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err == nil {
			t.Fatal("replacement accepted")
		}
		if data, err := os.ReadFile(sentinel); err != nil || string(data) != "foreign" {
			t.Fatal("replacement removed", err)
		}
		if err := os.Remove(sentinel); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(r.Root); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(original, r.Root); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("retained_path_and_duplicate_owner_rejected", func(t *testing.T) {
		r, root, run := runtimeFixture(t)
		if _, err := NewRuntime(root, run, "fixture", filepath.Join(root, "scratch")); err == nil {
			t.Fatal("runtime created inside source")
		}
		path := filepath.Join(r.Root, "runtime-owner.json")
		original, err := harnesslocal.ReadFile(path, 4096, false)
		if err != nil {
			t.Fatal(err)
		}
		var owner map[string]any
		if err := json.Unmarshal(original, &owner); err != nil {
			t.Fatal(err)
		}
		forged := append([]byte(`{"run_id":"wrong",`), original[1:]...)
		if err := harnesslocal.ReplaceFile(path, forged); err != nil {
			t.Fatal(err)
		}
		if _, err := BorrowRuntime(root, run, r.Root, r.RunID, r.LeaseID); err == nil {
			t.Fatal("duplicate owner accepted")
		}
		if err := harnesslocal.ReplaceFile(path, original); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("exact_resource_proofs", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		for kind, target := range map[string]any{
			"diagnostic_scope": map[string]any{"boot": "boot", "start": "1", "token": strings.Repeat("a", 64), "uid": os.Getuid()},
			"browser_renderer": map[string]any{"daemon_id": "engine", "image_id": "sha256:" + strings.Repeat("b", 64), "name": "cartulary-visual-12-aaaaaaaaaaaa", "token": strings.Repeat("a", 32)},
			"browser_process":  map[string]any{"boot": "boot", "start": "1", "pid": 12, "group": true, "pgid": 12, "sid": 12},
		} {
			proof := target.(map[string]any)
			if err := r.RecordResource(kind, proof, "pending"); err != nil {
				t.Fatal(kind, err)
			}
			if err := r.RecordResource(kind, proof, "released"); err != nil {
				t.Fatal(err)
			}
			proof["extra"] = true
			if err := r.RecordResource(kind, proof, "pending"); err == nil {
				t.Fatal("open proof accepted", kind)
			}
			delete(proof, "extra")
			if kind == "diagnostic_scope" {
				proof["uid"] = os.Getuid() + 1
			} else if kind == "browser_renderer" {
				proof["token"] = strings.Repeat("c", 32)
			} else {
				proof["pid"] = 1
			}
			if err := r.RecordResource(kind, proof, "pending"); err == nil {
				t.Fatal("forged proof accepted", kind)
			}
		}
		resources, err := r.Resources(4096)
		if err != nil || len(resources) != 0 {
			t.Fatal("invalid proof wrote journal", resources, err)
		}
	})
	t.Run("resource_bounds_and_registered_secrets", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		a, _ := r.RecoveryPath("one")
		b, _ := r.RecoveryPath("two")
		for _, path := range []string{a, b} {
			if err := r.RecordResource("managed_suite", path, "pending"); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := r.Resources(1); err == nil {
			t.Fatal("resource bound ignored")
		}
		for _, path := range []string{a, b} {
			if err := r.RecordResource("managed_suite", path, "released"); err != nil {
				t.Fatal(err)
			}
		}
		if err := r.RegisterEnvironment(map[string]string{"PASSWORD_TOKENS": "secret-value", "tokens_token": "another-secret", "CPU_TOKENS": "999999999", "SAFE": "nonsecret"}); err != nil {
			t.Fatal(err)
		}
		values, err := r.ForbiddenValues()
		if err != nil {
			t.Fatal(err)
		}
		if len(values) != 2 || values[0] != "another-secret" || values[1] != "secret-value" {
			t.Fatal(values)
		}
	})
}

//go:build linux

package harnessfixtures

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
func TestFixtureRuntimeProof(t *testing.T) {
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
		if _, err := r.PrivatePath(".."); err == nil {
			t.Fatal("escaping path accepted")
		}
		proof, _ := r.PrivatePath("service", "suite-lease.json")
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
	t.Run("resource_bounds_and_registered_secrets", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		a, _ := r.PrivatePath("one")
		b, _ := r.PrivatePath("two")
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
		r.RegisterEnvironment(map[string]string{"PASSWORD": "secret-value", "CPU_TOKENS": "999999999", "SAFE": "nonsecret"})
		values := r.ForbiddenValues()
		if len(values) != 1 || values[0] != "secret-value" {
			t.Fatal(values)
		}
	})
}

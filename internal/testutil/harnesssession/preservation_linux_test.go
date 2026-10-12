//go:build linux

package harnesssession

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func testFixturePreservation(t *testing.T) {
	t.Run("purge_only_unprotected_detail", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		protected := filepath.Join(r.Root, "suite", "suite-lease.json")
		dir, err := harnesslocal.PrivateDirectory(filepath.Dir(protected))
		if err != nil {
			t.Fatal(err)
		}
		dir.Close()
		if err = os.WriteFile(protected, []byte("private proof"), 0600); err != nil {
			t.Fatal(err)
		}
		if err = r.RecordResource("managed_suite", protected, "acquired"); err != nil {
			t.Fatal(err)
		}
		detail := filepath.Join(r.Root, "unrelated")
		dir, err = harnesslocal.PrivateDirectory(detail)
		if err != nil {
			t.Fatal(err)
		}
		dir.Close()
		if err = os.WriteFile(filepath.Join(detail, "private.log"), []byte("private detail"), 0600); err != nil {
			t.Fatal(err)
		}
		if err = r.PreserveRecovery(); err != nil {
			t.Fatal(err)
		}
		if _, err = os.Stat(protected); err != nil {
			t.Fatal("lost owner proof", err)
		}
		if _, err = os.Stat(detail); !os.IsNotExist(err) {
			t.Fatal("private detail retained", err)
		}
		if err = r.RecordResource("managed_suite", protected, "released"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("live_and_recovery_roots_survive_age", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		base := filepath.Dir(r.Root)
		if removed, err := CleanupStaleRoots(base, time.Now().Add(48*time.Hour)); err != nil || removed != 0 {
			t.Fatal("live owner pruned", removed, err)
		}
		// A dead predecessor with recovery proof is retained. Its current test
		// owner still closes it explicitly after the exact recovery is resolved.
		proof := map[string]any{"boot": "retired-boot", "pid": os.Getpid(), "start": "1"}
		data, _ := json.Marshal(proof)
		if err := harnesslocal.ReplaceFile(filepath.Join(r.Root, "runtime-process.json"), data); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(r.Root, "suite", "lease.json")
		if err := r.RecordResource("managed_suite", file, "pending"); err != nil {
			t.Fatal(err)
		}
		if removed, err := CleanupStaleRoots(base, time.Now().Add(48*time.Hour)); err != nil || removed != 0 {
			t.Fatal("unresolved predecessor pruned", removed, err)
		}
		if err := r.RecordResource("managed_suite", file, "released"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("dead_empty_owned_root_removed", func(t *testing.T) {
		root, run, scratch := t.TempDir(), t.TempDir(), t.TempDir()
		r, err := NewRuntime(root, run, "fixture", scratch)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Directory.Close()
		defer r.parent.Close()
		data, _ := json.Marshal(map[string]any{"boot": "retired-boot", "pid": os.Getpid(), "start": "1"})
		if err = harnesslocal.ReplaceFile(filepath.Join(r.Root, "runtime-process.json"), data); err != nil {
			t.Fatal(err)
		}
		if removed, err := CleanupStaleRoots(filepath.Dir(r.Root), time.Now().Add(48*time.Hour)); err != nil || removed != 1 {
			t.Fatal(removed, err)
		}
		if _, err := os.Stat(r.Root); !os.IsNotExist(err) {
			t.Fatal("dead empty root remains", err)
		}
	})
}

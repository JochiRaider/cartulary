//go:build linux

package harnesssession

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
)

func testFixturePreservation(t *testing.T) {
	testFixturePreservationBaseline(t)
	testSiblingDetailRemoval(t)
	testRecoveryDependencyValidation(t)
	testScratchPlacement(t)
	testStaleSettledRoots(t)
}

func testScratchPlacement(t *testing.T) {
	for _, mode := range []string{"machine_cache", "xdg", "home", "override", "empty", "relative"} {
		t.Run("scratch_"+mode, func(t *testing.T) {
			root, run, home := t.TempDir(), t.TempDir(), t.TempDir()
			for _, name := range []string{"CARTULARY_HARNESS_SCRATCH_ROOT", "CARTULARY_MACHINE_CACHE_DIR", "XDG_CACHE_HOME"} {
				t.Setenv(name, "")
				if err := os.Unsetenv(name); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("HOME", home)
			t.Setenv("TMPDIR", filepath.Join(home, "small-tmpfs"))
			want := filepath.Join(home, ".cache", "cartulary", "harness-scratch")
			switch mode {
			case "machine_cache":
				t.Setenv("CARTULARY_MACHINE_CACHE_DIR", filepath.Join(home, "cache"))
				want = filepath.Join(home, "cache", "harness-scratch")
			case "xdg":
				t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "xdg"))
				want = filepath.Join(home, "xdg", "cartulary", "harness-scratch")
			case "home":
				t.Setenv("XDG_CACHE_HOME", "relative")
			case "override":
				want = filepath.Join(home, "override")
				t.Setenv("CARTULARY_HARNESS_SCRATCH_ROOT", want)
			case "empty":
				t.Setenv("CARTULARY_HARNESS_SCRATCH_ROOT", "")
			case "relative":
				t.Setenv("CARTULARY_HARNESS_SCRATCH_ROOT", "relative")
			}
			r, err := NewRuntime(root, run, "placement", "")
			if mode == "empty" || mode == "relative" {
				if err == nil {
					r.Close()
					t.Fatal("invalid scratch accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			if filepath.Dir(filepath.Dir(r.Root)) != want {
				t.Fatalf("scratch %s, want %s", r.Root, want)
			}
		})
	}
}

func testStaleSettledRoots(t *testing.T) {
	newRoot := func(t *testing.T, scratch, id string) *Runtime {
		t.Helper()
		r, err := NewRuntime(t.TempDir(), t.TempDir(), id, scratch)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := r.ReleaseHandles(); err != nil {
				t.Error(err)
			}
		})
		return r
	}
	dead := func(t *testing.T, r *Runtime) {
		t.Helper()
		data, _ := json.Marshal(map[string]any{"boot": "retired-boot", "pid": os.Getpid(), "start": "1"})
		if err := harnesslocal.ReplaceFile(filepath.Join(r.Root, "runtime-process.json"), data); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("dead_released_journal_reclaimed", func(t *testing.T) {
		r := newRoot(t, t.TempDir(), "settled")
		file := filepath.Join(r.Root, "proof.json")
		if err := r.RecordDependentResource("managed_suite", file, "pending", []string{}); err != nil {
			t.Fatal(err)
		}
		if err := r.RecordResource("managed_suite", file, "released"); err != nil {
			t.Fatal(err)
		}
		dead(t, r)
		if removed, err := CleanupStaleRoots(filepath.Dir(r.Root), time.Now().Add(48*time.Hour)); err != nil || removed != 1 {
			t.Fatal(removed, err)
		}
		if _, err := os.Stat(r.Root); !os.IsNotExist(err) {
			t.Fatal("settled root retained", err)
		}
	})
	t.Run("busy_root_preserved", func(t *testing.T) {
		r := newRoot(t, t.TempDir(), "busy")
		dead(t, r)
		if err := unix.Flock(int(r.Directory.Fd()), unix.LOCK_EX); err != nil {
			t.Fatal(err)
		}
		defer unix.Flock(int(r.Directory.Fd()), unix.LOCK_UN)
		if removed, err := CleanupStaleRoots(filepath.Dir(r.Root), time.Now().Add(48*time.Hour)); err != nil || removed != 0 {
			t.Fatal(removed, err)
		}
		if _, err := os.Stat(r.Root); err != nil {
			t.Fatal("busy root removed", err)
		}
	})
	t.Run("active_recovery_preserved", func(t *testing.T) {
		r := newRoot(t, t.TempDir(), "recovering")
		attempt, err := r.BeginRecoveryAttempt()
		if err != nil {
			t.Fatal(err)
		}
		dead(t, r)
		if removed, err := CleanupStaleRoots(filepath.Dir(r.Root), time.Now().Add(48*time.Hour)); err != nil || removed != 0 {
			t.Fatal(removed, err)
		}
		authority, err := r.readDetail(attempt)
		if err != nil || authority.State != "open" {
			t.Fatal("live recovery modified", authority.State, err)
		}
	})
	t.Run("resource_free_scan_recovery_preserved", func(t *testing.T) {
		r := newRoot(t, t.TempDir(), "pending-scan")
		if err := r.RegisterSecret("test-pending-scan-authority"); err != nil {
			t.Fatal(err)
		}
		if err := r.PreserveRecovery(); err != nil {
			t.Fatal(err)
		}
		resources, err := r.Resources(1)
		if err != nil || len(resources) != 0 {
			t.Fatal("fixture must have no physical resources", err)
		}
		dead(t, r)
		if removed, err := CleanupStaleRoots(filepath.Dir(r.Root), time.Now().Add(48*time.Hour)); err != nil || removed != 0 {
			t.Fatal(removed, err)
		}
		secrets, err := os.ReadDir(filepath.Join(r.Root, "registered-secrets"))
		if err != nil || len(secrets) != 1 {
			t.Fatal("lost pending retained-scan authority", err)
		}
	})
	t.Run("retained_roots_do_not_starve_cleanup", func(t *testing.T) {
		scratch := t.TempDir()
		for i := range 65 {
			newRoot(t, scratch, fmt.Sprintf("a-live-%02d", i))
		}
		r := newRoot(t, scratch, "z-dead")
		dead(t, r)
		if removed, err := CleanupStaleRoots(filepath.Dir(r.Root), time.Now().Add(48*time.Hour)); err != nil || removed != 1 {
			t.Fatal(removed, err)
		}
	})
	t.Run("removal_budget_is_bounded", func(t *testing.T) {
		scratch := t.TempDir()
		for i := range 65 {
			dead(t, newRoot(t, scratch, fmt.Sprintf("dead-%02d", i)))
		}
		if removed, err := CleanupStaleRoots(filepath.Join(scratch, "suite-runtime"), time.Now().Add(48*time.Hour)); err != nil || removed != 64 {
			t.Fatal(removed, err)
		}
		if removed, err := CleanupStaleRoots(filepath.Join(scratch, "suite-runtime"), time.Now().Add(48*time.Hour)); err != nil || removed != 1 {
			t.Fatal(removed, err)
		}
	})
}

// Independent correction fixture, activated by the exact-dependency slice.
func testSiblingDetailRemoval(t *testing.T) {
	t.Run("purge_sibling_detail", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		proof := filepath.Join(r.Root, "suite", "lease.json")
		dir, err := harnesslocal.PrivateDirectory(filepath.Dir(proof))
		if err != nil {
			t.Fatal(err)
		}
		dir.Close()
		if err := os.WriteFile(proof, []byte("private proof"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := r.RecordDependentResource("managed_suite", proof, "acquired", []string{}); err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("managed_suite", proof, "released")
		detail := filepath.Join(filepath.Dir(proof), "credentials.log")
		if err := os.WriteFile(detail, []byte("unnecessary private detail"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := r.PreserveRecovery(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(proof); err != nil {
			t.Fatal("lost proof", err)
		}
		if _, err := os.Stat(detail); !os.IsNotExist(err) {
			t.Error("sibling detail survived", err)
		}
	})
}

func testFixturePreservationBaseline(t *testing.T) {
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
		if err = r.RecordDependentResource("managed_suite", protected, "acquired", []string{}); err != nil {
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
		if err := r.RecordDependentResource("managed_suite", file, "pending", []string{}); err != nil {
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

func testRecoveryDependencyValidation(t *testing.T) {
	for _, variant := range []string{"valid", "missing", "symlink", "foreign", "directory"} {
		t.Run("dependency_"+variant, func(t *testing.T) {
			r, _, _ := runtimeFixture(t)
			proof := filepath.Join(r.Root, "proof.json")
			cert := filepath.Join(r.Root, "root.pem")
			detail := filepath.Join(r.Root, "credentials.log")
			for _, file := range []string{proof, cert, detail} {
				if err := os.WriteFile(file, []byte("private"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if variant == "foreign" {
				cert = filepath.Join(t.TempDir(), "foreign")
			}
			if err := r.RecordDependentResource("object_store_proxy", proof, "acquired", []string{}, cert); variant == "foreign" {
				if err == nil {
					t.Fatal("foreign dependency admitted")
				}
				return
			} else if err != nil {
				t.Fatal(err)
			}
			defer r.RecordResource("object_store_proxy", proof, "released")
			if variant == "missing" || variant == "symlink" || variant == "directory" {
				if err := os.Remove(cert); err != nil {
					t.Fatal(err)
				}
			}
			if variant == "symlink" {
				if err := os.Symlink(proof, cert); err != nil {
					t.Fatal(err)
				}
			}
			if variant == "directory" {
				if err := os.Mkdir(cert, 0700); err != nil {
					t.Fatal(err)
				}
			}
			err := r.PreserveRecovery()
			if (err == nil) != (variant == "valid") {
				t.Fatal("dependency validation", variant, err)
			}
			if variant != "valid" && DescribeCleanupFailure(err).Stage != "recovery_proof" {
				t.Fatal("required-proof failure lost its physical boundary", DescribeCleanupFailure(err))
			}
			if _, err := os.Stat(detail); !os.IsNotExist(err) {
				t.Fatal("independent purge skipped", err)
			}
			if _, err := os.Stat(proof); err != nil {
				t.Fatal("authority removed", err)
			}
			if variant == "valid" {
				if _, err := os.Stat(cert); err != nil {
					t.Fatal("certificate removed", err)
				}
			}
		})
	}
}

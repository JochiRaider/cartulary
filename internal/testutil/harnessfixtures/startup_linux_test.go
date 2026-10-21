//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/leaseproof"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/startupproof"
)

func testStartupJournal(t *testing.T) {
	for _, mode := range []string{"prepared", "registered", "acquiring", "acquired", "settled", "parallel", "foreign", "malformed", "publication_failure"} {
		t.Run("startup_journal_"+mode, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Chmod(root, 0700); err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(root, startupproof.Filename)
			binding := startupproof.Binding{RuntimeRoot: root, RuntimeLease: "lease", RunRoot: root, RunID: "run", SuiteID: "suite", Services: []string{"object_store", "postgres"}, Dependencies: []string{}}
			if err := startupproof.Create(file, binding); err != nil {
				t.Fatal(err)
			}
			if err := startupproof.Create(file, binding); err == nil {
				t.Fatal("replaced existing intent")
			}
			if mode == "foreign" {
				binding.RuntimeLease = "foreign"
				if _, err := startupproof.Open(file, binding); err == nil {
					t.Fatal("foreign binding accepted")
				}
				return
			}
			if mode == "malformed" {
				if err := os.WriteFile(file, []byte(`{"schema_id":"forged"}`), 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := startupproof.Read(file); err == nil {
					t.Fatal("malformed journal accepted")
				}
				return
			}
			j, err := startupproof.Open(file, binding)
			if err != nil {
				t.Fatal(err)
			}
			if mode != "prepared" {
				if err := j.Register("exact helper proof"); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "publication_failure" {
				if err := os.Remove(file); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("/invalid", file); err != nil {
					t.Fatal(err)
				}
				if err := j.Begin("postgres", map[string]string{}, nil); err == nil {
					t.Fatal("unsafe publication accepted")
				}
				return
			}
			if mode == "parallel" {
				var wg sync.WaitGroup
				for _, service := range binding.Services {
					wg.Go(func() {
						if err := j.Begin(service, map[string]string{"service": service}, nil); err != nil {
							t.Error(err)
							return
						}
						if err := j.Acquired(service, service+"-container"); err != nil {
							t.Error(err)
						}
					})
				}
				wg.Wait()
			} else if mode != "prepared" && mode != "registered" {
				if err := j.Begin("postgres", map[string]string{}, nil); err != nil {
					t.Fatal(err)
				}
				if mode != "acquiring" {
					if err := j.Acquired("postgres", "container"); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "settled" {
					if err := j.Settled("postgres"); err != nil {
						t.Fatal(err)
					}
					if err := j.Finish(false); err != nil {
						t.Fatal(err)
					}
				}
			}
			d, err := startupproof.Read(file)
			if err != nil {
				t.Fatal(err)
			}
			want := mode == "prepared" || mode == "registered" || mode == "settled"
			if d.SettledWithoutLease() != want {
				t.Fatal("unsafe settlement decision", mode, d)
			}
			if mode == "parallel" && len(d.Attempts) != 2 {
				t.Fatal("sibling authority overwritten")
			}
			if mode == "acquiring" {
				if err := j.Settled("postgres"); err == nil {
					t.Fatal("unknown creation treated as absent")
				}
				if err := j.Begin("postgres", map[string]string{}, nil); err == nil {
					t.Fatal("unresolved attempt retried")
				}
			}
		})
	}
}

func testStartupCommandCancellation(t *testing.T) {
	t.Run("inherited_control_close_interrupts_read", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer w.Close()
		fd, err := syscall.Dup(int(r.Fd()))
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		control, err := startupproof.OpenControl(fd)
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan struct{})
		go func() { defer close(done); var b [1]byte; _, _ = control.Read(b[:]) }()
		control.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			w.Close()
			t.Fatal("inherited read survived close")
		}
	})
	runtime, root, _ := runtimeFixture(t)
	script := filepath.Join(root, "startup-helper")
	// FD 3 is the broker's first inherited capability. Reading the admission byte
	// precedes the readiness marker; EOF must allow settlement before process reap.
	body := "#!/bin/sh\ndd bs=1 count=1 <&3 >/dev/null 2>&1\necho admitted\ncat <&3 >/dev/null\nsleep 0.1\necho settled\n"
	if err := os.WriteFile(script, []byte(body), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
	defer cancel()
	result, err := commandRunner(root, runtime, []string{})(ctx, script, []string{"start-suite"}, map[string]string{"PATH": "/usr/bin:/bin"})
	if !errors.Is(err, context.DeadlineExceeded) || string(result.Stdout) != "admitted\nsettled\n" || result.Exit != 0 {
		t.Fatal("startup settlement was killed", result, err)
	}
	resources, err := runtime.Resources(64)
	if err != nil || len(resources) != 0 {
		t.Fatal("helper proof leaked", resources, err)
	}
}

func testPendingStartupRecovery(t *testing.T) {
	for _, mode := range []string{"prepared", "registered", "acquiring", "acquired", "settled", "legacy", "foreign"} {
		t.Run("pending_start_"+mode, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			directory, err := runtime.RecoveryPath("service-owners", "pending-start")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(directory, 0700); err != nil {
				t.Fatal(err)
			}
			lease := filepath.Join(directory, "suite-lease.json")
			journalFile := filepath.Join(directory, startupproof.Filename)
			env := map[string]string{suiteservices.SuiteIDEnv: "fixture-suite"}
			envFile := filepath.Join(directory, "suite-environment.json")
			writeFixtureJSON(t, envFile, env)
			binding := startupproof.Binding{RuntimeRoot: runtime.Root, RuntimeLease: runtime.LeaseID, RunRoot: run, RunID: runtime.RunID, SuiteID: env[suiteservices.SuiteIDEnv], Services: []string{"postgres"}, Dependencies: []string{}}
			files := []string{envFile}
			if mode != "legacy" {
				if mode == "foreign" {
					binding.RuntimeLease = "foreign"
				}
				if err := startupproof.Create(journalFile, binding); err != nil {
					t.Fatal(err)
				}
				files = append(files, journalFile)
				j, err := startupproof.Open(journalFile, binding)
				if err != nil {
					t.Fatal(err)
				}
				if mode != "prepared" && mode != "foreign" {
					proof, err := harnesssession.CaptureProcess(os.Getpid())
					if err != nil {
						t.Fatal(err)
					}
					data, _ := json.Marshal(proof)
					if err := j.Register(string(data)); err != nil {
						t.Fatal(err)
					}
					if mode != "registered" {
						labels := map[string]string{leaseproof.LabelManaged: leaseproof.ManagedValue, leaseproof.LabelRunID: runtime.RunID, leaseproof.LabelSuiteID: binding.SuiteID, leaseproof.LabelService: "postgres", leaseproof.LabelTarget: "test"}
						if err := j.Begin("postgres", labels, nil); err != nil {
							t.Fatal(err)
						}
						if mode != "acquiring" {
							if err := j.Acquired("postgres", "exact-container"); err != nil {
								t.Fatal(err)
							}
						}
						if mode == "settled" {
							if err := j.Settled("postgres"); err != nil {
								t.Fatal(err)
							}
							if err := j.Finish(false); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
			}
			if err := runtime.RecordDependentResource("managed_suite", lease, "pending", []string{}, files...); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("managed_suite", lease, "released")
			p := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, RecoveryOutput: true}
			if mode == "acquired" {
				data, err := p.startupCleanupLease(lease, env)
				if err != nil {
					t.Fatal(err)
				}
				var proof leaseproof.Lease
				if json.Unmarshal(data, &proof) != nil || len(proof.Resources) != 1 || proof.Resources[0].ContainerID != "exact-container" {
					t.Fatal("lost exact resource", string(data))
				}
				return
			}
			err = p.settlePendingSuite(lease, env)
			want := mode == "prepared" || mode == "registered" || mode == "settled"
			if (err == nil) != want {
				t.Fatal("incorrect pending settlement", mode, err)
			}
			if mode == "acquiring" {
				if err := runtime.PreserveRecovery(); err != nil {
					t.Fatal("pending journal cannot survive missing lease", err)
				}
				if _, err := startupproof.Read(journalFile); err != nil {
					t.Fatal("uncertain intent discarded", err)
				}
			}
		})
	}
}

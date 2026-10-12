//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func writeFixtureJSON(t *testing.T, name string, v any) {
	t.Helper()
	directory, err := harnesslocal.PrivateDirectory(filepath.Dir(name))
	if err != nil {
		t.Fatal(err)
	}
	directory.Close()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := harnesslocal.PublishNew(name, data); err != nil {
		t.Fatal(err)
	}
}
func serviceScope(run, target, suite string) map[string]any {
	startup := map[string]any{"attempt_count": 1, "retry_count": 0, "slowest_attempt_duration_ms": 1, "final_attempt": 1, "final_retryable": false, "final_retry_blocked_by_context": false}
	return map[string]any{"schema_id": "cartulary.test_services.scope.v2", "run_id": run, "target": target, "suite_id": suite, "artifact_dir": "_shared/test-services/" + suite, "readiness_generation": "sha256:" + strings.Repeat("a", 64),
		"wrapper": map[string]int{"owned_count": 1, "pass_through_count": 0}, "preflight": map[string]any{"docker_ok": true, "reaper_ready": true, "stale_containers_scanned": 0, "stale_containers_removed": 0, "stale_containers_deferred": 0, "ryuk_disabled_for_suite_startup": false}, "failures": map[string]any{}, "cleanup": map[string]any{},
		"postgres":     map[string]any{"started": true, "startup": startup, "attached_harness_count": 0, "created_database_count": 0, "migrated_database_count": 0, "template_clone_count": 0},
		"object_store": map[string]any{"started": true, "secure": true, "startup": startup, "attached_harness_count": 0, "bucket_create_count": 0, "bucket_cleanup_count": 0},
		"browser_e2e":  map[string]int{"retired_fixture_count": 0, "cleaned_fixture_count": 0, "reclaimed_fixture_count": 0}, "fixture": map[string]int{"total_count": 0, "total_duration_ms": 0, "strategy_aggregate_count": 0}, "started_services": map[string]any{}}
}
func TestFixtureSuiteOwner(t *testing.T) {
	t.Run("borrowed_suite_proof_and_detach_never_terminate", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		detached := 0
		source := map[string]any{"schema_id": "cartulary.test_services.lease.v1", "lease_id": "foreign-lease", "suite_id": "persistent-suite", "run_id": "persistent-run", "result_root": "/private/persistent", "run_root": "/private/persistent/persistent-run", "target": "test-services-start", "mode": "owned", "ownership_mode": "owned", "owner_pid": os.Getpid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano), "resources": []any{}, "proof_labels": map[string]any{}, "proof_prefixes": map[string]any{}, "cleanup_state": "not_started"}
		p := ProductionOptions{Root: root, RunRoot: run, Target: "browser-e2e", Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node"}}
		p.Run = func(_ context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
			if len(args) != 2 {
				t.Fatalf("unexpected physical operation %v", args)
			}
			if args[1] == "detach" {
				detached++
				return CommandResult{Started: true, Stdout: []byte(`{}`)}, nil
			}
			if args[1] != "attach" {
				t.Fatal(args)
			}
			writeFixtureJSON(t, filepath.Join(runtime.Root, "borrowed-suite.json"), map[string]any{"private": "binding"})
			data, _ := json.Marshal(map[string]any{"environment": map[string]string{"CARTULARY_TEST_SUITE_ID": "run-suite"}, "container_proof": map[string]any{"lease": source, "ownership": "borrowed"}})
			return CommandResult{Started: true, Stdout: data}, nil
		}
		suite, err := p.AttachSuite(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		lease, err := readObject(suite.LeaseFile, "cartulary.test_services.lease.v1", 1048576)
		if err != nil {
			t.Fatal(err)
		}
		if lease["mode"] != "attach" || lease["ownership_mode"] != "attach" || lease["run_id"] != runtime.RunID || lease["suite_id"] != "run-suite" || source["mode"] != "owned" {
			t.Fatal("borrower proof lost identity or mutated owner", lease)
		}
		if err := suite.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := suite.Close(context.Background()); err != nil || detached != 1 {
			t.Fatal(err, detached)
		}
		resources, err := runtime.Resources(64)
		if err != nil || len(resources) != 0 {
			t.Fatal(resources, err)
		}
	})

	t.Run("concurrent_single_acquisition_and_close", func(t *testing.T) {
		var acquired, closed atomic.Int32
		owner := &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
			acquired.Add(1)
			return &Suite{Close: func(context.Context) error { closed.Add(1); return nil }}, nil
		}}
		var wg sync.WaitGroup
		for range 12 {
			wg.Go(func() {
				if _, err := owner.Ensure(context.Background()); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		for range 12 {
			wg.Go(func() {
				if err := owner.Close(context.Background()); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		if acquired.Load() != 1 || closed.Load() != 1 {
			t.Fatal(acquired.Load(), closed.Load())
		}
		if _, err := owner.Ensure(context.Background()); err == nil {
			t.Fatal("acquire after close")
		}
	})
	for _, variant := range []string{"ready", "forged_scope", "duplicate_environment", "startup_failure", "missing_result"} {
		t.Run(variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			stops := 0
			options := ProductionOptions{Root: root, RunRoot: run, Target: "test", Runtime: runtime, Environment: map[string]string{"TEST_SERVICES_BIN": "/fixed/testservices"}}
			options.Run = func(_ context.Context, executable string, args []string, environment map[string]string) (CommandResult, error) {
				if executable != "/fixed/testservices" {
					t.Fatal(executable)
				}
				if args[0] == "terminate-suite" {
					stops++
					if !strings.Contains(args[2], ".cleanup-") {
						t.Fatal("original proof passed to destructive helper")
					}
					if _, err := harnesslocal.ReadFile(args[2], 1048576, false); err != nil {
						t.Fatal(err)
					}
					return CommandResult{Started: true}, nil
				}
				suite := "work-graph-test"
				scope := serviceScope(runtime.RunID, "test", suite)
				status := map[string]any{"schema_id": "cartulary.test_services.start_result.v1", "status": "ready", "run_id": runtime.RunID, "target": "test", "suite_id": suite, "service_scope_ref": "_shared/test-services/" + suite + "/service-scope.json", "failure_class": nil, "failure_reason": nil}
				writeFixtureJSON(t, args[4], map[string]string{"private": "lease"})
				if variant == "missing_result" {
					return CommandResult{Started: true, Exit: 1}, errors.New("worker death")
				}
				if variant == "forged_scope" {
					scope["run_id"] = "foreign"
				}
				if variant == "startup_failure" {
					status["status"] = "failed"
					status["failure_class"] = "infra"
					status["failure_reason"] = "service_start_error"
					scope["failure"] = map[string]any{"failure_class": "infra", "failure_reason": "service_start_error", "attempts_started": 1, "max_attempts": 1, "retryable": false, "retry_blocked_by_context": false}
				}
				writeFixtureJSON(t, filepath.Join(run, status["service_scope_ref"].(string)), scope)
				writeFixtureJSON(t, args[6], status)
				if variant == "duplicate_environment" {
					if err := harnesslocal.PublishNew(args[2], []byte(`{"SSL_CERT_FILE":"a","SSL_CERT_FILE":"b"}`)); err != nil {
						t.Fatal(err)
					}
				} else {
					writeFixtureJSON(t, args[2], map[string]string{"CARTULARY_PGTEST_ADMIN_DSN": "private-password", "IGNORE_THIS": "foreign"})
				}
				if variant == "startup_failure" {
					return CommandResult{Started: true, Exit: 1}, errors.New("startup")
				}
				return CommandResult{Started: true}, nil
			}
			suite, err := options.StartSuite(context.Background())
			if variant == "ready" {
				if err != nil {
					t.Fatal(err)
				}
				if suite.Environment["IGNORE_THIS"] != "" {
					t.Fatal("unadmitted binding")
				}
				if values, err := runtime.ForbiddenValues(); err != nil || len(values) != 1 {
					t.Fatal("secret not registered")
				}
				if err := suite.Close(context.Background()); err != nil {
					t.Fatal(err)
				}
				_ = suite.Close(context.Background())
			} else {
				if err == nil {
					t.Fatal("invalid publication accepted")
				}
				var failure *Failure
				if !errors.As(err, &failure) {
					t.Fatal(err)
				}
				expected := "artifact"
				if variant == "startup_failure" {
					expected = "infra"
				}
				if failure.Class != expected {
					t.Fatal(err)
				}
			}
			if stops != 1 {
				t.Fatal("physical cleanup count", stops)
			}
			resources, err := runtime.Resources(64)
			if err != nil || len(resources) != 0 {
				t.Fatal(resources, err)
			}
		})
	}
}
func TestFixtureCommandLifecycle(t *testing.T) {
	runtime, root, _ := runtimeFixture(t)
	run := commandRunner(root, runtime)
	for _, mode := range []string{"exit", "cancel", "overflow"} {
		t.Run(mode, func(t *testing.T) {
			script := filepath.Join(root, "helper")
			body := "#!/bin/sh\nexit 7\n"
			if mode == "cancel" {
				body = "#!/bin/sh\nsleep 30 &\nwait\n"
			}
			if mode == "overflow" {
				body = "#!/bin/sh\nhead -c 1048577 /dev/zero\n"
			}
			if err := os.WriteFile(script, []byte(body), 0700); err != nil {
				t.Fatal(err)
			}
			ctx, done := context.WithTimeout(context.Background(), time.Second)
			defer done()
			result, err := run(ctx, script, nil, map[string]string{"PATH": "/usr/bin:/bin"})
			if err == nil || !result.Started {
				t.Fatal(result, err)
			}
			if mode == "exit" && result.Exit != 7 {
				t.Fatal(result.Exit)
			}
			if mode == "cancel" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
			if mode == "overflow" && len(result.Stdout) != 1048576 {
				t.Fatal("unbounded output")
			}
			resources, err := runtime.Resources(64)
			if err != nil || len(resources) != 0 {
				t.Fatal(resources, err)
			}
		})
	}
}

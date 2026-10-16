//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func fixtureReference(t *testing.T, runtime *harnesssession.Runtime, kind string, target any) string {
	t.Helper()
	reference, err := runtime.ResourceReference(kind, target)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}
func TestFixtureRecovery(t *testing.T) {
	testRecoveryCommand(t)
	testCleanupDependencies(t)
	for _, mode := range []string{"ordinary", "recovery"} {
		t.Run("diagnostic_dependency_blocks_suite_"+mode, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			suite := filepath.Join(runtime.Root, "test-services", "suite-lease.json")
			writeFixtureJSON(t, suite, map[string]string{"proof": "fixture"})
			writeFixtureJSON(t, filepath.Join(filepath.Dir(suite), "suite-environment.json"), map[string]string{})
			if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("managed_suite", suite, "released")
			scope := map[string]any{"boot": "unresolved-boot", "start": "1", "token": strings.Repeat("d", 64), "uid": os.Getuid()}
			if err := runtime.RecordDependentResource("diagnostic_scope", scope, "acquired", []string{fixtureReference(t, runtime, "managed_suite", suite)}); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("diagnostic_scope", scope, "released")
			stops := 0
			p := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"TEST_SERVICES_BIN": "services"}, Run: func(context.Context, string, []string, map[string]string) (CommandResult, error) {
				stops++
				return CommandResult{}, nil
			}}
			var err error
			if mode == "recovery" {
				err = p.Recover(context.Background())
			} else {
				production := NewProduction(p, Options{})
				production.Suite.suite = &Suite{LeaseFile: suite, Close: func(context.Context) error { stops++; return nil }}
				err = production.Close(context.Background())
			}
			if err == nil || stops != 0 {
				t.Fatal("diagnostic consumer did not block its suite", mode, err, stops)
			}
		})
	}
	t.Run("failed_consumer_blocks_only_its_declared_suite", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		a, b := filepath.Join(runtime.Root, "suite-a", "suite-lease.json"), filepath.Join(runtime.Root, "suite-b", "suite-lease.json")
		browser := filepath.Join(runtime.Root, "browser", "owner.json")
		for _, suite := range []string{a, b} {
			writeFixtureJSON(t, suite, map[string]string{"proof": "synthetic"})
			writeFixtureJSON(t, filepath.Join(filepath.Dir(suite), "suite-environment.json"), map[string]string{})
			if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("managed_suite", suite, "released")
		}
		if err := runtime.RecordDependentResource("browser_stack", browser, "pending", []string{fixtureReference(t, runtime, "managed_suite", a)}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("browser_stack", browser, "released")
		inspections, stops := 0, 0
		outcomes := []CleanupOutcome{}
		p := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node", "TEST_SERVICES_BIN": "services"}, ReportCleanup: func(v CleanupOutcome) { outcomes = append(outcomes, v) }}
		p.Run = func(_ context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
			if args[0] == "terminate-suite" {
				stops++
				if !strings.HasPrefix(args[2], b+".cleanup-") {
					t.Fatal("dependent suite stopped")
				}
				return CommandResult{}, nil
			}
			inspections++
			return CommandResult{}, errors.New("inspection failed")
		}
		if _, err := runtime.BeginRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		defer runtime.FinishRecoveryAttempt()
		for range 2 {
			if err := p.Recover(context.Background()); !errors.Is(err, ErrCleanupBlocked) {
				t.Fatal("missing blocked outcome", err)
			}
		}
		if inspections != 1 || stops != 1 {
			t.Fatal("duplicate or starved release", inspections, stops)
		}
		counts := map[string]int{}
		for _, outcome := range outcomes {
			counts[outcome.Outcome]++
		}
		if counts["failed"] != 2 || counts["blocked"] != 2 || counts["completed"] != 1 {
			t.Fatal("lost cleanup outcome", outcomes)
		}
		if err := runtime.FinishRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		if _, err := runtime.BeginRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		if err := p.Recover(context.Background()); !errors.Is(err, ErrCleanupBlocked) || inspections != 2 || stops != 1 {
			t.Fatal("fresh attempt did not retry unresolved consumer", err, inspections, stops)
		}
	})
	t.Run("stalled_dependency_inspection_is_bounded_once", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		browser := filepath.Join(runtime.Root, "browser-stack-leases", "a", "acquisition", "owner.json")
		suite := filepath.Join(runtime.Root, "suite", "suite-lease.json")
		if err := runtime.RecordDependentResource("managed_suite", suite, "pending", []string{}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("managed_suite", suite, "released")
		if err := runtime.RecordDependentResource("browser_stack", browser, "pending", []string{fixtureReference(t, runtime, "managed_suite", suite)}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("browser_stack", browser, "released")
		if _, err := runtime.BeginRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		defer runtime.FinishRecoveryAttempt()
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			p := ProductionOptions{Root: root, RunRoot: run, Target: "review", Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node"}}
			p.Run = func(ctx context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
				calls++
				if args[1] != "inspect" {
					t.Error("unexpected cleanup")
				}
				<-ctx.Done()
				return CommandResult{}, ctx.Err()
			}
			before := time.Now()
			if err := p.Recover(context.Background()); err == nil {
				t.Fatal("stalled inspection passed")
			}
			if time.Since(before) != 120*time.Second {
				t.Fatal("inspection did not use resource deadline")
			}
			if err := p.Recover(context.Background()); err == nil || calls != 1 {
				t.Fatal("duplicate resource budget", calls, err)
			}
		})
	})
	for _, variant := range []string{"success", "missing_dependency", "unknown_dependency", "failed_browser", "borrowed", "borrowed_browser_failure"} {
		t.Run(variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			suite := filepath.Join(runtime.Root, "test-services", "suite-lease.json")
			browser := filepath.Join(runtime.Root, "browser-stack-leases", "a", "acquisition", "owner.json")
			writeFixtureJSON(t, suite, map[string]string{"proof": "synthetic"})
			writeFixtureJSON(t, filepath.Join(filepath.Dir(suite), "suite-environment.json"), map[string]string{})
			writeFixtureJSON(t, filepath.Join(runtime.Root, "borrowed-environment.json"), map[string]string{})
			writeFixtureJSON(t, browser, map[string]string{"proof": "synthetic"})
			if variant != "missing_dependency" && !strings.HasPrefix(variant, "borrowed") {
				if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
					t.Fatal(err)
				}
			}
			if strings.HasPrefix(variant, "borrowed") {
				file := filepath.Join(runtime.Root, "borrowed-suite.json")
				writeFixtureJSON(t, file, map[string]string{"proof": "synthetic"})
				if err := runtime.RecordDependentResource("borrowed_suite", file, "acquired", []string{}); err != nil {
					t.Fatal(err)
				}
			}
			if variant != "borrowed" {
				kind, dependency := "managed_suite", suite
				if strings.HasPrefix(variant, "borrowed") {
					kind, dependency = "borrowed_suite", filepath.Join(runtime.Root, "borrowed-suite.json")
				}
				err := runtime.RecordDependentResource("browser_stack", browser, "pending", []string{fixtureReference(t, runtime, kind, dependency)})
				if variant == "missing_dependency" {
					if err == nil {
						t.Fatal("unresolved dependency admitted")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() {
				resources, err := runtime.Resources(64)
				if err != nil {
					t.Error(err)
					return
				}
				for _, r := range resources {
					var target any
					json.Unmarshal(r.Target, &target)
					runtime.RecordResource(r.Kind, target, "released")
				}
			})
			calls := []string{}
			p := ProductionOptions{Root: root, RunRoot: run, Target: "review", Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node", "TEST_SERVICES_BIN": "services"}}
			p.Run = func(_ context.Context, executable string, args []string, _ map[string]string) (CommandResult, error) {
				if args[0] == "terminate-suite" {
					calls = append(calls, "suite")
					if _, err := harnesslocal.ReadFile(args[2], 1048576, false); err != nil {
						t.Fatal(err)
					}
					return CommandResult{Started: true}, nil
				}
				op := args[1]
				calls = append(calls, op)
				if op == "inspect" {
					dependency := suite
					if variant == "borrowed_browser_failure" {
						dependency = filepath.Join(runtime.Root, "borrowed-service-lease.json")
					}
					if variant == "unknown_dependency" {
						dependency = filepath.Join(runtime.Root, "foreign", "suite-lease.json")
					}
					data, _ := json.Marshal(map[string]string{"suite_lease": dependency})
					return CommandResult{Started: true, Stdout: data}, nil
				}
				if op == "recover" && (variant == "failed_browser" || variant == "borrowed_browser_failure") {
					return CommandResult{Started: true, Exit: 12}, errors.New("physical stop failed")
				}
				return CommandResult{Started: true, Stdout: []byte(`{}`)}, nil
			}
			err := p.Recover(context.Background())
			joined := strings.Join(calls, ",")
			resources, readErr := runtime.Resources(64)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if variant == "success" && len(resources) != 0 {
				t.Fatal("successful recovery left resource authority", resources)
			}
			if variant == "failed_browser" && len(resources) != 2 {
				t.Fatal("failed browser recovery lost exact browser and suite authority", resources)
			}
			if variant == "success" {
				if err != nil || joined != "inspect,recover,suite" {
					t.Fatal(calls, err)
				}
			} else if variant == "borrowed" {
				if err != nil || joined != "detach" {
					t.Fatal(calls, err)
				}
			} else {
				if err == nil || strings.Contains(joined, "suite") || strings.Contains(joined, "detach") {
					t.Fatal("dependency failure authorized suite stop", calls, err)
				}
			}
		})
	}
}

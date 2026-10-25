//go:build linux

package harnessruntime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// Exercise continuation through the real graph engine and fixture broker. The
// injected physical adapter counts allocations; it contains no scheduling policy.
func TestRuntimeBrowserFailureContinuation(t *testing.T) {
	t.Run("retained_reset_failure_preserves_lifecycle_classification", func(t *testing.T) {
		root := t.TempDir()
		run := filepath.Join(root, "run")
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		const attempt = `{"schema_id":"cartulary.browser_reset_attempt.v2","reset_id":"before-successor","status":"fail","attempt":1,"duration_ms":5,"clock":"linux_boottime_10ms","clock_identity":"clock:test","resolution_ms":10,"availability":"available","runtime_profile_id":"default","stages":[{"stage":"allocation_validated","status":"pass"},{"stage":"backend_stopped","status":"fail"},{"stage":"database_reset","status":"skipped"},{"stage":"object_store_reset","status":"skipped"},{"stage":"browser_state_cleared","status":"skipped"},{"stage":"replacement_backend_ready","status":"skipped"},{"stage":"generation_published","status":"skipped"}],"backend_generation_before":1,"backend_generation_after":1,"failure_stage":"backend_stop_or_preflight","database_diagnostic_ref":null,"database_stage":null,"database_sqlstate":null,"persistent_state_reset":false,"browser_state_cleared":false,"backend_ready":false,"backend_generation_ref":null,"tainted":true,"failure_class":"harness","failure_reason":"fixture_error"}`
		if err := os.WriteFile(filepath.Join(run, "reset.attempt.json"), []byte(attempt), 0600); err != nil {
			t.Fatal(err)
		}
		u := engineUnit("reset", 1)
		u.Kind = "lifecycle"
		u.Evidence = []string{"reset.attempt.json"}
		u.Command = command{Executable: "/bin/sh", Args: []string{"-c", "exit 1"}}
		p := processRunner{Root: root, Environment: map[string]string{"CARTULARY_TEST_RESULTS_DIR": root, "CARTULARY_TEST_RUN_ID": "run"}}
		r := p.Execute(context.Background(), executionRequest{Unit: u})
		if r.Failure == nil || r.Failure.Class != "harness" || r.Failure.Reason != "fixture_error" || r.ExitCode != 1 || failureExitCode(r.Failure, r.Signal) != 3 {
			t.Fatalf("%+v", r)
		}
	})
	for _, failing := range []string{"group", "preparation"} {
		t.Run(failing, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "run"), 0700); err != nil {
				t.Fatal(err)
			}
			runtime, err := harnesssession.NewRuntime(root, filepath.Join(root, "run"), "browser", t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer runtime.Close()
			allocated, released := 0, 0
			var active atomic.Int64
			broker := harnessfixtures.New(harnessfixtures.Options{Providers: map[string]harnessfixtures.Provider{
				"browser_stack": &harnessfixtures.ProviderFuncs{AcquireFunc: func(context.Context, harnessfixtures.Request) (*harnessfixtures.Allocation, error) {
					if active.Load() != 1 {
						t.Error("allocation escaped consuming admission")
					}
					allocated++
					id := fmt.Sprint(allocated)
					return &harnessfixtures.Allocation{Ownership: "owned", ResourceIDs: []string{"browser:" + id}, Environment: map[string]string{"ALLOCATION": id}, Release: func(context.Context, bool) error { released++; return nil },
						Prepare: func(_ context.Context, request harnessfixtures.Request) error {
							if active.Load() != 1 {
								t.Error("preparation escaped consuming admission")
							}
							if failing == "preparation" && request.UnitID == "second" {
								return &harnessfixtures.Failure{Class: "harness", Reason: "fixture_error"}
							}
							return nil
						}}, nil
				}},
			}})
			production := harnessfixtures.NewProduction(harnessfixtures.ProductionOptions{Runtime: runtime}, harnessfixtures.Options{})
			production.Broker = broker
			fixtures := &nativeFixtures{production: production}
			group, second, next := engineUnit("group", 1), engineUnit("second", 1, "group"), engineUnit("next", 1, "second")
			for _, u := range []*unit{&group, &second, &next} {
				u.Command.Environment = map[string]string{"CARTULARY_BROWSER_PREPARATION_LABEL": u.ID}
				u.Fixture = "browser_stack"
				u.PostgresIsolation = "exclusive"
				u.Services = []string{"object_store", "postgres"}
				u.Affinity = "lane"
				u.FailurePolicy.BlockDescendants = false
			}
			requirements := []harnessfixtures.WorkRequirement{workRequirement(group), workRequirement(second), workRequirement(next)}
			if err := production.RegisterConsumers(requirements); err != nil {
				t.Fatal(err)
			}
			var executions []string
			result := boundedGraph(t, enginePlan(group, second, next), enginePorts{ConsumerTerminal: production.ConsumerTerminal, Fixtures: fixtures,
				Admit: func(_ context.Context, u unit) (*hostLease, error) {
					if active.Add(1) != 1 {
						t.Error("overlapping admission")
					}
					return &hostLease{Release: func(context.Context) error {
						if active.Add(-1) != 0 {
							t.Error("duplicate admission release")
						}
						return nil
					}}, nil
				}, Execute: func(_ context.Context, r executionRequest) unitResult {
					if active.Load() != 1 {
						t.Error("execution escaped consuming admission")
					}
					executions = append(executions, r.Unit.ID)
					if r.Unit.ID == "next" && r.Environment["ALLOCATION"] != "2" {
						t.Error("later work did not receive a fresh allocation")
					}
					if r.Unit.ID == failing {
						if failing == "group" {
							return unitResult{Status: "failed", Failure: failed("product", "test_assertion_failure")}
						}
						return unitResult{Status: "failed", Failure: failed("harness", "fixture_error")}
					}
					return unitResult{Status: "passed"}
				}})
			expected := []string{"group", "second", "next"}
			if failing == "preparation" {
				expected = []string{"group", "next"}
			}
			if active.Load() != 0 || !slices.Equal(executions, expected) || result.Status != "failed" || result.States["next"] != "passed" || allocated != 2 || released != 2 || broker.Unresolved() {
				t.Fatalf("executions=%v allocated=%d released=%d result=%+v", executions, allocated, released, result)
			}
		})
	}
}

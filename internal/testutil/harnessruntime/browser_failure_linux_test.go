//go:build linux

package harnessruntime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
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
	for _, failing := range []string{"group", "reset"} {
		t.Run(failing, func(t *testing.T) {
			root := t.TempDir()
			runtime, err := harnessfixtures.NewRuntime(root, filepath.Join(root, "run"), "browser", t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer runtime.Close()
			allocated, released := 0, 0
			broker := harnessfixtures.New(harnessfixtures.Options{Providers: map[string]harnessfixtures.Provider{
				"browser_stack": &harnessfixtures.ProviderFuncs{AcquireFunc: func(context.Context, harnessfixtures.Request) (*harnessfixtures.Allocation, error) {
					allocated++
					id := fmt.Sprint(allocated)
					return &harnessfixtures.Allocation{Ownership: "owned", ResourceIDs: []string{"browser:" + id}, Environment: map[string]string{"ALLOCATION": id}, Release: func(context.Context, bool) error { released++; return nil }}, nil
				}},
			}})
			fixtures := &nativeFixtures{production: &harnessfixtures.Production{Broker: broker, Options: harnessfixtures.ProductionOptions{Runtime: runtime}}}
			group, reset, next := engineUnit("group", 1), engineUnit("reset", 1, "group"), engineUnit("next", 1, "reset")
			reset.Kind = "lifecycle"
			for _, u := range []*unit{&group, &reset, &next} {
				u.Fixture = "browser_stack"
				u.Affinity = "lane"
				u.FailurePolicy.BlockDescendants = false
			}
			next.Command.Environment["CARTULARY_BROWSER_RELEASE_AFFINITY"] = "1"
			var executions []string
			result := boundedGraph(t, enginePlan(group, reset, next), enginePorts{Fixtures: fixtures, Execute: func(_ context.Context, r executionRequest) unitResult {
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
			if !slices.Equal(executions, []string{"group", "reset", "next"}) || result.Status != "failed" || result.States["next"] != "passed" || allocated != 2 || released != 2 || broker.Unresolved() {
				t.Fatalf("executions=%v allocated=%d released=%d result=%+v", executions, allocated, released, result)
			}
		})
	}
}

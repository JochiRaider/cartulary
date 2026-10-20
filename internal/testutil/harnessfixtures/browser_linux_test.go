//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func TestFixtureBrowserAdapter(t *testing.T) {
	testRendererOwnership(t)
	for _, uncertain := range []bool{false, true} {
		t.Run(fmt.Sprintf("failed_service_generation_uncertain_%t", uncertain), func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			p := NewProduction(ProductionOptions{Runtime: runtime, Root: root, RunRoot: run}, Options{})
			request := Request{PostgresIsolation: "exclusive", Services: []string{"postgres"}, Capability: "browser_stack", UnitID: "first", Affinity: "lane", RuntimeProfile: "default"}
			primary, settlement := errors.New("startup failed"), errors.New("cleanup uncertain")
			starts, releases := 0, 0
			owner := &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
				starts++
				if uncertain {
					return &Suite{Close: func(context.Context) error { return settlement }}, primary
				}
				return nil, primary
			}, ReleaseResidency: func(context.Context) error { releases++; return nil }}
			p.clusters[clusterKey(request)] = owner
			if _, err := p.Broker.Acquire(t.Context(), request); !errors.Is(err, primary) || errors.Is(err, settlement) != uncertain {
				t.Fatal("startup or settlement failure lost", err)
			}
			if starts != 1 || (p.clusters[clusterKey(request)] != nil) != uncertain || releases != map[bool]int{false: 1, true: 0}[uncertain] {
				t.Fatal("failed generation settlement or residency lost", starts, releases)
			}
			request.UnitID = "successor"
			if !uncertain {
				// A later cohort selects its own generation; the failed request
				// is never replayed. No physical browser is needed for this fault.
				p.clusters[clusterKey(request)] = &SuiteOwner{Acquire: func(context.Context) (*Suite, error) { starts++; return nil, primary }}
			}
			if _, err := p.Broker.Acquire(t.Context(), request); err == nil {
				t.Fatal("failed acquisition succeeded")
			}
			if starts != map[bool]int{false: 2, true: 1}[uncertain] {
				t.Fatal("failed generation retried or successor poisoned", starts)
			}
		})
	}
	for _, cleanupFails := range []bool{false, true} {
		t.Run(fmt.Sprintf("partial_browser_ownership_cleanup_failure_%t", cleanupFails), func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			suiteFile := filepath.Join(runtime.Root, "synthetic-suite.json")
			writeFixtureJSON(t, suiteFile, map[string]any{"fixture": "owned-suite"})
			if err := runtime.RecordDependentResource("managed_suite", suiteFile, "acquired", []string{}); err != nil {
				t.Fatal(err)
			}
			ref := fixtureReference(t, runtime, "managed_suite", suiteFile)
			// Only synthetic resources exist in this fault fixture. Clear their
			// journal entries after assertions so an expected failure is bounded.
			t.Cleanup(func() {
				resources, err := runtime.Resources(4096)
				if err != nil {
					t.Error(err)
				}
				for _, resource := range resources {
					if err := runtime.RecordResource(resource.Kind, resource.Target, "released"); err != nil {
						t.Error(err)
					}
				}
			})
			primary := errors.New("preparation failed")
			secondary := errors.New("settlement uncertain")
			prepares, settles, serviceCloses := 0, 0, 0
			options := ProductionOptions{Runtime: runtime, Root: root, RunRoot: run, Environment: map[string]string{"NODE_BIN": "synthetic-node"}}
			options.residency = &residency{allocations: map[string]*residentRecord{}}
			options.Run = func(_ context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
				switch args[1] {
				case "prepare":
					prepares++
					return CommandResult{Started: true, Exit: 10}, primary
				case "settle":
					settles++
					if cleanupFails {
						return CommandResult{Started: true, Exit: 12}, secondary
					}
					return CommandResult{Started: true, Stdout: []byte(`{}`)}, nil
				default:
					t.Fatal("unexpected physical operation", args[1])
					return CommandResult{}, primary
				}
			}
			provider := &browserProvider{options: options, allocations: map[string]int{}, clones: map[string]int{},
				services: func(context.Context, Request) (*Suite, error) {
					return &Suite{LeaseFile: suiteFile, ResourceRef: ref, ResourceRefs: []string{ref}, Environment: map[string]string{}}, nil
				},
				releaseServices: func(context.Context, Request) error {
					serviceCloses++
					return runtime.RecordResource("managed_suite", suiteFile, "released")
				}}
			broker := New(Options{Providers: map[string]Provider{"browser_stack": provider}})
			_, err := broker.Acquire(t.Context(), testRequest())
			if !errors.Is(err, primary) || prepares != 1 || settles != 1 {
				t.Fatal("partial acquisition lost primary failure or repeated cleanup", err, prepares, settles)
			}
			if cleanupFails {
				if !errors.Is(err, secondary) || !broker.Unresolved() || serviceCloses != 0 {
					t.Fatal("failed partial ownership escaped quarantine", err, broker.Unresolved(), serviceCloses)
				}
				if _, err := broker.Acquire(t.Context(), testRequest()); !errors.Is(err, ErrCleanupBlocked) || prepares != 1 {
					t.Fatal("partial generation was reused", err, prepares)
				}
			} else if broker.Unresolved() || serviceCloses != 1 {
				t.Fatal("settled partial acquisition retained its exclusive service", broker.Unresolved(), serviceCloses)
			}
			_ = broker.Close(t.Context())
			if settles != 1 {
				t.Fatal("terminal closure retried product cleanup", settles)
			}
		})
	}
	node := os.Getenv("NODE_BIN")
	if node == "" {
		t.Fatal("Make-owned Node binding required")
	}
	cwd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(cwd, "../../.."))
	t.Run("closed_cleanup_detail_survives_browser_failure_without_reclassification", func(t *testing.T) {
		primary := errors.New("private transport sentinel")
		for index, data := range []string{
			`{"schema_id":"cartulary.harness_cleanup_diagnostic.v4","stage":"process_authority","cause":"unresolved","errno":0}`,
			`{"schema_id":"cartulary.harness_cleanup_diagnostic.v4","stage":"process_authority","cause":"unresolved","errno":0,"pid":123}`,
		} {
			p := ProductionOptions{Environment: map[string]string{"NODE_BIN": node}, Run: func(context.Context, string, []string, map[string]string) (CommandResult, error) {
				return CommandResult{Exit: 12, Stdout: []byte(`{"failure_class":"harness","failure_reason":"cleanup_error"}`), Stderr: []byte(data)}, primary
			}}
			_, err := p.browserOperation(context.Background(), "settle", nil, nil)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Class != "harness" || failure.Reason != "cleanup_error" || !errors.Is(err, primary) {
				t.Fatal("failure classification or native cause changed", err)
			}
			var diagnostic *harnesssession.CleanupDiagnostic
			got := errors.As(err, &diagnostic)
			want := index == 0
			if got != want {
				t.Fatal("invalid diagnostic crossed the boundary", got, want)
			}
			if got && harnesssession.DescribeCleanupFailure(err).Stage != "process_authority" {
				t.Fatal("native cause was discarded")
			}
		}
	})
	t.Run("closed_cli_rejects_unknown_operation", func(t *testing.T) {
		cmd := exec.Command(node, filepath.Join(root, "tools/harness/browser/fixture-lifecycle.mjs"), "unknown")
		cmd.Stderr = os.Stderr
		stdout, err := cmd.Output()
		if err == nil {
			t.Fatal("unknown operation accepted")
		}
		var failure map[string]any
		if json.Unmarshal(stdout, &failure) != nil || len(failure) != 2 || failure["failure_class"] == nil {
			t.Fatalf("missing closed failure: %s", stdout)
		}
	})
	t.Run("private_browser_preparation_joins_native_runtime", func(t *testing.T) {
		run := filepath.Join(t.TempDir(), "native-browser-prepare")
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		runtime, err := harnesssession.NewRuntime(root, run, "native-browser-prepare", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer runtime.Close()
		suite := filepath.Join(runtime.Root, "test-services", "suite-lease.json")
		writeFixtureJSON(t, suite, map[string]any{"schema_id": "cartulary.test_services.lease.v3", "postgres_provisioning": nil, "run_id": runtime.RunID, "suite_id": "fixture-suite", "lease_id": runtime.LeaseID, "result_root": filepath.Dir(run), "run_root": filepath.Join(filepath.Dir(run), runtime.RunID), "target": "browser-e2e", "mode": "owned", "ownership_mode": "owned", "owner_pid": os.Getpid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano), "resources": []any{}, "proof_labels": map[string]any{}, "proof_prefixes": map[string]any{}, "cleanup_state": "not_started"})
		sessionRoot, _ := runtime.RecoveryPath("browser-stack-leases", "attempt")
		directory, err := harnesslocal.PrivateDirectory(sessionRoot)
		if err != nil {
			t.Fatal(err)
		}
		directory.Close()
		if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("managed_suite", suite, "released")
		browserFile := filepath.Join(sessionRoot, "acquisition", "owner.json")
		if err := runtime.RecordDependentResource("browser_stack", browserFile, "pending", []string{fixtureReference(t, runtime, "managed_suite", suite)}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("browser_stack", browserFile, "released")
		p := ProductionOptions{Root: root, RunRoot: run, Target: "browser-e2e", Runtime: runtime, Environment: map[string]string{"NODE_BIN": node, "PATH": os.Getenv("PATH"), "CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT": sessionRoot}}
		native := p.command()
		p.Run = func(ctx context.Context, command string, _ []string, env map[string]string) (CommandResult, error) {
			// Call the physical module directly so a failed synthetic proof exposes
			// its stack in test diagnostics instead of the private CLI's closed error.
			module, _ := json.Marshal("file://" + filepath.Join(root, "tools/harness/browser/fixture-lifecycle.mjs"))
			args := []string{"--input-type=module", "--eval", "import {fixtureLifecycle} from " + string(module) + "; console.log(JSON.stringify(await fixtureLifecycle('prepare',process.argv.slice(1))));", suite}
			result, err := native(ctx, command, args, env)
			if err != nil {
				t.Logf("operation %s failed: %v; stderr=%s", args[1], err, result.Stderr)
			}
			return result, err
		}
		value, err := p.browserOperation(context.Background(), "prepare", []string{suite}, p.environment())
		if err != nil {
			t.Fatal(err)
		}
		if value["file"] != filepath.Join(sessionRoot, "acquisition", "owner.json") || len(value["launch_id"].(string)) != 36 {
			t.Fatal(value)
		}
	})
}

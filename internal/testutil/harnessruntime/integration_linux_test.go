//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRuntimeSelectedGraphExecution(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	// Keep real nested workload evidence beneath this Make-owned invocation,
	// including failures. Synthetic fault fixtures use separate temporary roots.
	results := os.Getenv("CARTULARY_TEST_RESULTS_DIR")
	runID := os.Getenv("CARTULARY_TEST_RUN_ID")
	if results == "" || runID == "" {
		t.Fatal("Make must supply retained invocation identity")
	}
	if !filepath.IsAbs(results) {
		results = filepath.Join(root, results)
	}
	results = filepath.Join(results, runID, "native-engine")
	id := fmt.Sprintf("selected-graph-%d", os.Getpid())
	environment := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			environment[key] = value
		}
	}
	for _, key := range []string{"CARTULARY_TEST_SERVICES_MODE", "CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER", "CARTULARY_TEST_SERVICES_SESSION_FILE", "MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "JSON", "OWNER", "ROWS", "SERVICE_BACKED_ONLY", "CARTULARY_HARNESS_SUITE_RUNTIME_ROOT", "CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID", "CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID", "CARTULARY_HARNESS_EXECUTION_ENABLED", launchContextKey, failureContextKey} {
		delete(environment, key)
	}
	environment["HARNESS_DIAGNOSTICS"] = "off"
	environment["CARTULARY_HARNESS_CACHE_MODE"] = "off"
	environment["CARTULARY_TEST_RESULTS_DIR"] = results
	environment["CARTULARY_TEST_RUN_ID"] = id
	values := []string{}
	for k, v := range environment {
		values = append(values, k+"="+v)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	plan, err := Prepare(ctx, root, []string{"--selection", "owner", "--target", "test-slice", "--owner", "harness.evidence_accounting", "--rows", "harness.evidence_accounting.behavior.local_migration_activities"}, values)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range plan.facts.Graph.Units {
		if u.Fixture != "none" {
			t.Fatal("fixture-free integration selection acquired a new fixture requirement")
		}
	}
	run := filepath.Join(results, id)
	publication, err := newGraphPublication(plan, run)
	if err != nil {
		t.Fatal(err)
	}
	defer publication.Abort()
	// The existing secure runtime owner is test composition until S-029/S-030
	// migrate shared fixture ownership and graph finalization. No public dispatch
	// or native fallback is introduced by this integration test.
	node := environment["NODE_BIN"]
	if !filepath.IsAbs(node) {
		node = filepath.Join(root, node)
	}
	setup := exec.CommandContext(ctx, node, "--input-type=module", "-e", `import {createSuiteRuntime} from './tools/harness/runtime/suite-runtime.mjs'; const r=createSuiteRuntime({repoRoot:process.cwd(),runRoot:process.argv[1],runID:process.argv[2]});console.log(JSON.stringify({root:r.root,leaseID:r.leaseID,runID:r.runID}));`, run, id)
	setup.Dir = root
	bytes, err := setup.Output()
	if err != nil {
		t.Fatal(err)
	}
	var runtimeOwner struct{ Root, LeaseID, RunID string }
	if err := json.Unmarshal(bytes, &runtimeOwner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup := exec.Command(node, "--input-type=module", "-e", `import {closeSuiteRuntimeRoot} from './tools/harness/runtime/suite-runtime.mjs';closeSuiteRuntimeRoot({root:process.argv[1],expectedLeaseID:process.argv[2]});`, runtimeOwner.Root, runtimeOwner.LeaseID)
		cleanup.Dir = root
		if output, err := cleanup.CombinedOutput(); err != nil {
			t.Errorf("private runtime cleanup: %v %s", err, output)
		}
	}()
	environment["CARTULARY_HARNESS_SUITE_RUNTIME_ROOT"] = runtimeOwner.Root
	environment["CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID"] = runtimeOwner.LeaseID
	environment["CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID"] = runtimeOwner.RunID
	maps.Copy(environment, plan.facts.Bindings)
	for k, v := range map[string]string{"CARTULARY_HARNESS_GRAPH_CHILD": "1", "CARTULARY_HARNESS_IDENTITY_PREPARED": "1", "CARTULARY_HARNESS_SKIP_PREREQUISITES": "1", "CARTULARY_SUPPRESS_CHILD_SUCCESS": "1"} {
		environment[k] = v
	}
	private, err := harnesslocal.OpenDirectory(runtimeOwner.Root, true)
	if err != nil {
		t.Fatal(err)
	}
	defer private.Close()
	processes := processRunner{Root: root, Environment: environment, FailureScope: func(identity map[string]any) (failureScope, error) { return createFailureScope(private, identity) }}
	cache := newNativeCache(plan, root, run)
	defer cache.Close()
	result, err := executeGraph(ctx, plan, enginePorts{Execute: processes.Execute, Cache: cache, Event: publication.Event, Terminal: publication.Terminal})
	if err != nil {
		t.Fatalf("engine failed: %v; retained %s", err, run)
	}
	summary, err := publication.Finish(result)
	if err != nil {
		t.Fatalf("publication: %v; retained %s", err, run)
	}
	if summary["status"] != "pass" {
		t.Fatalf("selected graph failed: %v; retained %s", result.Failure, run)
	}
	if err := checkRegularEvidence(filepath.Join(run, "rows/harness.evidence_accounting.behavior.local_migration_activities.json")); err != nil {
		t.Fatal(err)
	}
	t.Logf("native selected graph retained at %s", run)
}

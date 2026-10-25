//go:build linux

package boundary

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

// This separate Make-owned route exercises real service/browser boundaries.
// It is excluded from the fixture-free native unit route and never collects
// qualification observations, pads a workload, or changes sampling policy.
func TestNativeGraphServiceObservations(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	results, outer := os.Getenv("CARTULARY_TEST_RESULTS_DIR"), os.Getenv("CARTULARY_TEST_RUN_ID")
	if results == "" || outer == "" {
		t.Fatal("Make-owned retained identity required")
	}
	if !filepath.IsAbs(results) {
		results = filepath.Join(root, results)
	}
	results = filepath.Join(results, outer, "native-boundary")
	// These small harness probes exercise real checkpoint and TLS observations.
	// Product migration correctness remains in its independent owner selection.
	for _, selection := range []struct{ id, owner, row string }{{"postgres", "harness.browser", "harness.browser.integration.postgres_transaction_isolation"}, {"https", "harness.browser", "harness.browser.boundary_support.https_trust"}} {
		t.Run(selection.id, func(t *testing.T) {
			env := map[string]string{}
			for _, entry := range os.Environ() {
				k, v, ok := strings.Cut(entry, "=")
				if !ok {
					continue
				}
				if strings.HasPrefix(k, "CARTULARY_TEST_SERVICES_") && k != "CARTULARY_TEST_SERVICES_BIN" || strings.HasPrefix(k, "CARTULARY_PGTEST_") || strings.HasPrefix(k, "CARTULARY_S3TEST_") || strings.HasPrefix(k, "CARTULARY_HARNESS_SUITE_RUNTIME_") {
					continue
				}
				env[k] = v
			}
			for _, k := range []string{"CARTULARY_HARNESS_LAUNCH_CONTEXT", "CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT", "CARTULARY_HARNESS_INVOCATION_ROOT", "MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "JSON", "OWNER", "ROWS", "SERVICE_BACKED_ONLY", "SCOPE", "SUITE"} {
				delete(env, k)
			}
			env["CARTULARY_TEST_RESULTS_DIR"] = results
			probeID := "native-observation-" + selection.id
			env["CARTULARY_TEST_RUN_ID"] = probeID
			env["HARNESS_DIAGNOSTICS"] = "basic"
			env["OWNER"] = selection.owner
			env["ROWS"] = selection.row
			env["SCOPE"] = "services"
			env["CARTULARY_MAKE_INPUT_SOURCES"] = "OWNER=cli ROWS=cli SCOPE=cli HARNESS_DIAGNOSTICS=cli CARTULARY_HARNESS_CACHE_MODE=cli"
			env["CARTULARY_HARNESS_CACHE_MODE"] = "off"
			input := []string{}
			for k, v := range env {
				input = append(input, k+"="+v)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			binary := env["CARTULARY_HARNESS_RUNTIME_BIN"]
			if binary == "" {
				binary = filepath.Join(root, ".cache/cartulary/toolbin/cartulary-harness")
			}
			run := filepath.Join(results, probeID)
			runInstalledGraph(t, ctx, root, binary, []string{"--selection", "command", "--target", "test-slice"}, input, run)
			node := env["NODE_BIN"]
			if !filepath.IsAbs(node) {
				node = filepath.Join(root, node)
			}
			read := exec.CommandContext(ctx, node, "--input-type=module", "-e", `import {loadRetainedObservability} from './tools/harness/observability/observability.mjs';import {readFileSync} from 'node:fs';import {validateLiveSnapshot} from './tools/harness/contract/live-observation.mjs';const r=await loadRetainedObservability(process.argv[1]);validateLiveSnapshot(JSON.parse(readFileSync(process.argv[1]+'/diagnostics/live-status.json')),r.run.manifest);if(!r.execution||!r.envelope||!r.resourceIndex)throw Error('missing native observation');if(r.execution.host_admission.status!=='complete'||!r.execution.host_admission.records.some(x=>x.kind==='decision'&&x.decision.granted))throw Error('missing native admission coverage');if(process.argv[2]==='postgres'&&(!r.postgresCheckpoints?.records.length||r.postgresCheckpoints.status!=='observed'))throw Error('missing bounded checkpoint completions');if(process.argv[2]==='https'&&!r.resourceIndex.containers.identities.length)throw Error('no real owned-container observation');`, run, selection.id)
			read.Dir = root
			if out, err := read.CombinedOutput(); err != nil {
				t.Fatalf("independent retained reader: %v %s retained=%s", err, out, run)
			}
			t.Logf("native service graph and retained validation passed: %s", run)
		})
	}
}

// Exercise the prepared executable, including its source-identity check, under
// the enclosing admission and accounting lifetime. This is test composition;
// execution and settlement remain owned by the native graph command.
func runInstalledGraph(t *testing.T, ctx context.Context, root, binary string, args, environment []string, run string) {
	t.Helper()
	child := exec.CommandContext(ctx, binary, append([]string{"graph"}, args...)...)
	child.Dir, child.Env = root, environment
	child.Cancel = func() error { return child.Process.Signal(syscall.SIGTERM) }
	child.WaitDelay = 2 * time.Minute
	closeDescriptors, err := harnessadmission.InheritWorkspace(child)
	if err != nil {
		t.Fatal(err)
	}
	defer closeDescriptors()
	if out, err := child.CombinedOutput(); err != nil {
		t.Fatalf("installed native graph: %v %.4096s retained=%s", err, out, run)
	}
	data, err := harnesslocal.ReadFile(filepath.Join(run, "run-summary.json"), 16*1024*1024, false)
	var summary struct {
		Status string `json:"status"`
	}
	if err != nil || json.Unmarshal(data, &summary) != nil || summary.Status != "pass" {
		t.Fatalf("installed native graph lacks a passing summary: %v retained=%s", err, run)
	}
}

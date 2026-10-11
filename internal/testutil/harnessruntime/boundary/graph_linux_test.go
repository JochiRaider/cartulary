//go:build linux

package boundary

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessruntime"
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
	for _, selection := range []struct{ id, owner, row string }{{"postgres", "module.database_migrations", "module.database_migrations.integration.schema_bootstrap"}, {"https", "harness.browser", "harness.browser.boundary_support.https_trust"}} {
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
			for _, k := range []string{"CARTULARY_HARNESS_LAUNCH_CONTEXT", "CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT", "CARTULARY_HARNESS_INVOCATION_ROOT", "MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "JSON", "OWNER", "ROWS", "SERVICE_BACKED_ONLY"} {
				delete(env, k)
			}
			env["CARTULARY_TEST_RESULTS_DIR"] = results
			env["CARTULARY_TEST_RUN_ID"] = selection.id
			env["HARNESS_DIAGNOSTICS"] = "basic"
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
			summary, code, err := harnessruntime.Run(ctx, root, binary, []string{"--selection", "owner", "--target", "service-backed-test-slice", "--owner", selection.owner, "--rows", selection.row, "--service-backed-only"}, input)
			run := filepath.Join(results, selection.id)
			if err != nil || code != 0 || summary["status"] != "pass" {
				t.Fatalf("native service graph failed: exit=%d error=%v retained=%s", code, err, run)
			}
			node := env["NODE_BIN"]
			if !filepath.IsAbs(node) {
				node = filepath.Join(root, node)
			}
			read := exec.CommandContext(ctx, node, "--input-type=module", "-e", `import {loadRetainedObservability} from './tools/harness/observability/observability.mjs';import {readFileSync} from 'node:fs';import {validateLiveSnapshot} from './tools/harness/contract/live-observation.mjs';const r=await loadRetainedObservability(process.argv[1]);validateLiveSnapshot(JSON.parse(readFileSync(process.argv[1]+'/diagnostics/live-status.json')),r.run.manifest);if(!r.execution||!r.envelope||!r.resourceIndex)throw Error('missing native observation');if(process.argv[2]==='https'&&!r.resourceIndex.containers.identities.length)throw Error('no real owned-container observation');`, run, selection.id)
			read.Dir = root
			if out, err := read.CombinedOutput(); err != nil {
				t.Fatalf("independent retained reader: %v %s retained=%s", err, out, run)
			}
			t.Logf("native service graph and retained validation passed: %s", run)
		})
	}
}

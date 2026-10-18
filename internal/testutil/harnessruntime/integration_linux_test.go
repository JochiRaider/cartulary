//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	binary := filepath.Join(root, ".cache/cartulary/toolbin/cartulary-harness")
	private := t.TempDir()
	if err := os.Chmod(private, 0700); err != nil {
		t.Fatal(err)
	}
	proof, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	ns, err := os.Readlink("/proc/self/ns/time")
	if err != nil {
		t.Fatal(err)
	}
	_, start, err := bootReading()
	if err != nil {
		t.Fatal(err)
	}
	owner := map[string]any{"version": 1, "invocation_id": uuid(), "pid": os.Getpid(), "start_ticks": proof.Start, "start_ms": start, "boot_id": proof.Boot, "namespace": ns}
	b, _ := json.Marshal(owner)
	if err := harnesslocal.PublishNew(filepath.Join(private, "owner.json"), b); err != nil {
		t.Fatal(err)
	}
	// A nested integration probe must fit the enclosing unit's reservation.
	// Go row compilation intentionally scales to host capacity; use the existing
	// one-token shell row for graph/receipt proof, with Go covered independently.
	rowID := "harness.evidence_accounting.behavior.current_epoch_evidence"
	for _, mode := range []string{"off", "basic", "basic-second"} {
		if !t.Run(mode, func(t *testing.T) {
			diagnosticMode := mode
			if diagnosticMode != "off" {
				diagnosticMode = "basic"
			}
			environment["HARNESS_DIAGNOSTICS"] = diagnosticMode
			environment["CARTULARY_TEST_RUN_ID"] = id + "-" + mode
			environment["CARTULARY_HARNESS_INVOCATION_ROOT"] = private
			input := []string{}
			for k, v := range environment {
				input = append(input, k+"="+v)
			}
			summary, code, err := Run(ctx, root, binary, []string{"--selection", "owner", "--target", "test-slice", "--owner", "harness.evidence_accounting", "--rows", rowID}, input)
			run := filepath.Join(results, id+"-"+mode)
			if err != nil || code != 0 || summary == nil || summary.Summary["status"] != "pass" {
				t.Fatalf("native graph: code=%d err=%v; retained %s", code, err, run)
			}
			if err := checkRegularEvidence(filepath.Join(run, "rows/"+rowID+".json")); err != nil {
				t.Fatal(err)
			}
			node := environment["NODE_BIN"]
			if !filepath.IsAbs(node) {
				node = filepath.Join(root, node)
			}
			reader := exec.CommandContext(ctx, node, "--input-type=module", "-e", `import {readFileSync} from 'node:fs';import {validateLiveSnapshot} from './tools/harness/contract/live-observation.mjs';import {loadRetainedObservability} from './tools/harness/observability/observability.mjs';const r=await loadRetainedObservability(process.argv[1]);validateLiveSnapshot(JSON.parse(readFileSync(process.argv[1]+'/diagnostics/live-status.json')),r.run.manifest);if(process.argv[2]==='basic'&&(!r.execution||!r.envelope||!r.resourceIndex))throw Error('missing native observations');`, run, diagnosticMode)
			reader.Dir = root
			if out, err := reader.CombinedOutput(); err != nil {
				t.Fatalf("independent Node retained reader: %v %s; retained %s", err, out, run)
			}
			t.Logf("native graph and independent reader passed: %s", run)
		}) {
			return // Keep the first failed graph as the primary test evidence.
		}
	}
	_, end, err := bootReading()
	if err != nil {
		t.Fatal(err)
	}
	b, _ = json.Marshal(map[string]any{"end_ms": end, "exit_code": 0, "signal": 0, "boot_id": proof.Boot, "namespace": ns})
	if err := harnesslocal.PublishNew(filepath.Join(private, "completion.json"), b); err != nil {
		t.Fatal(err)
	}
	if err := harnessdiag.PublishTerminalReceipts(private); err != nil {
		t.Fatal("native wrapper receipt", err)
	}
	for _, mode := range []string{"basic", "basic-second"} {
		run := filepath.Join(results, id+"-"+mode)
		node := environment["NODE_BIN"]
		if !filepath.IsAbs(node) {
			node = filepath.Join(root, node)
		}
		reader := exec.CommandContext(ctx, node, "--input-type=module", "-e", `import {loadRetainedObservability} from './tools/harness/observability/observability.mjs';const r=await loadRetainedObservability(process.argv[1]);if(r.terminalReceipt?.outcome!=='passed'||r.terminalReceipt.graph_memberships.length!==2)throw Error('incorrect multiple-graph receipt');`, run)
		reader.Dir = root
		if out, err := reader.CombinedOutput(); err != nil {
			t.Fatalf("receipt reader: %v %s", err, out)
		}
	}

}

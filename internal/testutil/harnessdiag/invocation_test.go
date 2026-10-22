//go:build linux

package harnessdiag

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInvocationAccountingPublication(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := OpenInvocation(root); err != nil {
		t.Fatal(err)
	}
	id, err := InvocationIdentity(root)
	if err != nil {
		t.Fatal(err)
	}
	source, graphDigest := digest([]byte("source")), digest([]byte("graph"))
	roots := []string{filepath.Join(t.TempDir(), "one"), filepath.Join(t.TempDir(), "two")}
	write := func(file string, v any) {
		t.Helper()
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(file, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	members := []any{}
	for _, run := range roots {
		if err = os.MkdirAll(filepath.Join(run, "diagnostics"), 0700); err != nil {
			t.Fatal(err)
		}
		registration := RegisterGraph(root, run, object{"run_id": filepath.Base(run), "source_digest": source, "graph_digest": graphDigest})
		if registration == nil {
			t.Fatal("registration unavailable")
		}
		registration.ScanStarted()
		registration.ScanFinished(true)
		registration.Complete()
		members = append(members, object{"run_digest": digest([]byte(filepath.Base(run))), "graph_digest": graphDigest, "envelope_digest": digest([]byte("envelope")), "scan_digest": digest([]byte("scan")), "scan_outcome": "passed", "scan_interval": object{"start_ms": 0, "end_ms": 1}})
		write(filepath.Join(run, "cleanup-results.json"), object{"schema_id": "cartulary.harness_cleanup_results.v2", "run_id": filepath.Base(run), "results": []any{}})
	}
	receipt := object{"schema_id": "cartulary.harness_terminal_receipt.v1", "invocation_digest": id, "source_digest": source, "outer_duration_ms": 10, "child_exit_code": 7, "outcome": "failed", "graph_memberships": members, "publication_start_ms": 11}
	for _, run := range roots {
		write(filepath.Join(run, "diagnostics/terminal-receipt.json"), receipt)
	}
	metrics := map[string]any{"duration_ms": 10.0, "cpu_usage_us": 5000.0, "memory_peak_bytes": 8192.0, "initial_memory_peak_bytes": 4096.0, "controller_setup_ms": 1.0, "controller_publication_prefix_ms": 1.0, "child_exit_code": 7.0, "signal": 0.0, "scope": "invocation_cgroup", "memory_semantics": "peak_cgroup_charge", "settlement": "empty"}
	full := errors.New("publication storage unavailable")
	calls := 0
	if err = publishInvocationAccounting(root, metrics, func(_ string, _ []byte) error { calls++; return full }); !errors.Is(err, full) || calls != 1 {
		t.Fatalf("publication failure was hidden or later graph was written: calls=%d err=%v", calls, err)
	}
	// All graph validation must close before publishing any member.
	cleanupFile := filepath.Join(roots[1], "cleanup-results.json")
	write(cleanupFile, object{"schema_id": "cartulary.harness_cleanup_results.v2", "run_id": "foreign", "results": []any{}})
	calls = 0
	if err = publishInvocationAccounting(root, metrics, func(_ string, _ []byte) error { calls++; return nil }); err == nil || calls != 0 {
		t.Fatal("foreign cleanup partially published accounting", err, calls)
	}
	write(cleanupFile, object{"schema_id": "cartulary.harness_cleanup_results.v2", "run_id": "two", "results": []any{}})
	if err = PublishInvocationAccounting(root, metrics); err != nil {
		t.Fatal(err)
	}
	for _, run := range roots {
		value, _, err := boundObject(filepath.Join(run, "diagnostics/invocation-accounting.json"), "cartulary.harness_invocation_accounting.v1", 65536)
		if err != nil || value["metrics"].(object)["child_exit_code"] != float64(7) || len(value["graph_memberships"].([]any)) != 2 {
			t.Fatal("lost primary exit or membership", err)
		}
	}
	if err = PublishInvocationAccounting(root, metrics); err == nil {
		t.Fatal("overwrote closed accounting")
	}
	if _, err = RegisterRequiredGraph(root, roots[0], object{"run_id": "one", "source_digest": source, "graph_digest": graphDigest}); err == nil {
		t.Fatal("duplicate required graph was admitted")
	}
	owner, err := wrapperOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = wrapperGraphs(root, owner); err == nil {
		t.Fatal("failed registration did not poison closure")
	}
}

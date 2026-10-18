package harnessruntime

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func planFixture(t *testing.T) map[string]any {
	t.Helper()
	d := "sha256:" + strings.Repeat("a", 64)
	policyBytes, err := os.ReadFile("../../../tools/harness_instrumentation_policy.json")
	if err != nil {
		t.Fatal(err)
	}
	var policy any
	if err := json.Unmarshal(policyBytes, &policy); err != nil {
		t.Fatal(err)
	}
	policyDigest, err := semanticDigest(policy)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := map[string]any{"schema_id": "cartulary.harness_capability_snapshot.v1", "cpu_tokens": 2, "memory_bytes": 104857600, "process_slots": 4, "io_tokens": 2, "postgres_lanes": 1, "object_store_lanes": 1, "writable_volume": true, "port_lanes": 1, "services": map[string]any{}, "sources": map[string]any{}}
	hash, _ := semanticDigest(snapshot)
	snapshot["snapshot_digest"] = hash
	u := map[string]any{"unit_id": "unit:a", "owner_id": "harness.command_surface", "kind": "runner", "command": map[string]any{"executable": "true", "args": []any{}, "environment": map[string]any{}}, "needs": []any{}, "resource_claims": map[string]any{"cpu": 1, "process": 1}, "shared_locks": []any{"host_activity"}, "exclusive_locks": []any{}, "fixture_lease": "none", "service_dependencies": []any{}, "cache_policy": "none", "timeout_ms": 1000, "current_run_evidence_outputs": []any{"unit-results/unit-a.json"}, "row_bindings": []any{}, "reusable_artifact_outputs": []any{}, "failure_policy": map[string]any{"block_descendants": true, "continue_independent": true, "aggregate_effect": "required"}, "estimated_work_ms": 1}
	sourceInputs := []any{map[string]any{"path": "fixture.go", "kind": "file", "mode": "0644", "byte_digest": d}}
	u["renderer_requirement"] = "none"
	u["physical_allocations"] = map[string]any{}
	sourceDigest, _ := semanticDigest(map[string]any{"schema_id": "cartulary.source_snapshot_digest.v2", "entries": sourceInputs})
	u["semantic_digest"], _ = semanticDigest(u)
	units := []any{u}
	graphDigest, _ := semanticDigest(map[string]any{"units": units})
	return map[string]any{"schema_id": planSchema, "host_profile": map[string]any{"platform": "linux", "architecture": "amd64", "kernel_release": "fixture", "cpu_models": []any{"fixture"}}, "instrumentation_policy_bytes": string(policyBytes), "source_inputs": sourceInputs, "cache_inputs": map[string]any{}, "graph": map[string]any{"schema_id": "cartulary.harness_work_graph.v7", "selector": "fixture", "units": units, "graph_digest": graphDigest}, "manifest": map[string]any{"schema_id": "cartulary.harness_run_manifest.v3", "run_id": "fixture", "command_id": "cartulary.harness.command.test_slice.v1", "target": "test-slice", "declared_inputs": map[string]any{}, "source_commit": strings.Repeat("a", 40), "source_state": "clean", "source_digest": sourceDigest, "toolchain_digest": d, "system_digest": hash, "capability_snapshot": snapshot, "go_worker_budget": 2, "graph_digest": graphDigest, "cache_mode": "off", "started_at": "2026-10-11T00:00:00Z", "instrumentation": map[string]any{"mode": "off", "policy_digest": policyDigest}}, "projections": map[string]any{"test-slice": map[string]any{"command_id": "cartulary.harness.command.test_slice.v1", "unit_ids": []any{"unit:a"}}}, "capacities": map[string]any{"cpu": 2, "memory_mb": 100, "process": 4, "io": 2, "postgres": 1, "object_store": 1, "port_lane": 1, "browser_stack": 1, "service_stack": 1, "volume": 1}, "aging_quantum_ms": 1000, "bindings": map[string]any{"NODE_BIN": "/private/node"}, "input_environment": map[string]any{"clear": []any{"CARTULARY_MAKE_INPUT_SOURCES"}, "forward": map[string]any{}}, "helper_digest": d, "cache": []any{map[string]any{"unit_id": "unit:a", "eligible": false, "reason": "mode_off", "profile_id": nil, "policy": nil, "input_digest": nil, "dependency_digest": nil, "dependency_strategy": nil, "vulnerability_database_revision": nil}}, "preparation": map[string]any{"clock": nil, "intervals": []any{}}}
}
func planBytes(t *testing.T, value map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func rehashPlan(t *testing.T, p map[string]any) {
	t.Helper()
	graph := p["graph"].(map[string]any)
	for _, raw := range graph["units"].([]any) {
		u := raw.(map[string]any)
		delete(u, "semantic_digest")
		u["semantic_digest"], _ = semanticDigest(u)
	}
	hash, _ := semanticDigest(map[string]any{"units": graph["units"]})
	graph["graph_digest"] = hash
	p["manifest"].(map[string]any)["graph_digest"] = hash
}
func TestRuntimeUnitInputEnvironment(t *testing.T) {
	fixture := planFixture(t)
	fixture["input_environment"] = map[string]any{
		"clear":   []any{"CARTULARY_MAKE_INPUT_SOURCES", "FUTURE_SELECTOR", "RUNTIME_VALUE", "SUITE"},
		"forward": map[string]any{"RUNTIME_VALUE": "admitted"},
	}
	plan, err := decodePlan(planBytes(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	parent := map[string]string{"CARTULARY_MAKE_INPUT_SOURCES": "SUITE=cli", "SUITE": "runtime", "FUTURE_SELECTOR": "selected", "RUNTIME_VALUE": "unvalidated", "PATH": "/owned/path"}
	child := plan.unitEnvironment(parent)
	for _, name := range []string{"CARTULARY_MAKE_INPUT_SOURCES", "SUITE", "FUTURE_SELECTOR"} {
		if _, exists := child[name]; exists {
			t.Fatalf("public control %s escaped", name)
		}
	}
	if child["RUNTIME_VALUE"] != "admitted" || child["PATH"] != "/owned/path" || parent["SUITE"] != "runtime" || parent["RUNTIME_VALUE"] != "unvalidated" {
		t.Fatal("runtime forwarding or parent preservation failed")
	}
}
func TestRuntimePlanClosure(t *testing.T) {
	valid := planBytes(t, planFixture(t))
	if p, err := decodePlan(valid); err != nil || p.UnitCount() != 1 {
		t.Fatalf("valid fixture: %v", err)
	}
	for name, mutate := range map[string]func(map[string]any){
		"missing_worker_budget":              func(p map[string]any) { delete(p["manifest"].(map[string]any), "go_worker_budget") },
		"excess_worker_budget":               func(p map[string]any) { p["manifest"].(map[string]any)["go_worker_budget"] = 3 },
		"zero_worker_budget":                 func(p map[string]any) { p["manifest"].(map[string]any)["go_worker_budget"] = 0 },
		"fractional_worker_budget":           func(p map[string]any) { p["manifest"].(map[string]any)["go_worker_budget"] = 1.5 },
		"unsafe_worker_budget":               func(p map[string]any) { p["manifest"].(map[string]any)["go_worker_budget"] = 9007199254740992 },
		"previous_plan_version":              func(p map[string]any) { p["schema_id"] = "cartulary.harness_runtime_plan.v4" },
		"missing_input_policy":               func(p map[string]any) { delete(p, "input_environment") },
		"input_policy_without_source_marker": func(p map[string]any) { p["input_environment"].(map[string]any)["clear"] = []any{"OWNER"} },
		"input_policy_duplicate": func(p map[string]any) {
			p["input_environment"].(map[string]any)["clear"] = []any{"CARTULARY_MAKE_INPUT_SOURCES", "CARTULARY_MAKE_INPUT_SOURCES"}
		},
		"input_policy_unknown_forward": func(p map[string]any) {
			p["input_environment"].(map[string]any)["forward"] = map[string]any{"UNDECLARED": "value"}
		},
		"input_policy_source_marker_forward": func(p map[string]any) {
			p["input_environment"].(map[string]any)["forward"] = map[string]any{"CARTULARY_MAKE_INPUT_SOURCES": "SUITE=cli"}
		},
		"source_preimage_off": func(p map[string]any) { p["source_inputs"].([]any)[0].(map[string]any)["mode"] = "0755" },
		"retired_version":     func(p map[string]any) { p["schema_id"] = "cartulary.harness_runtime_plan.v1" },
		"policy_forgery":      func(p map[string]any) { p["instrumentation_policy_bytes"] = "{}" },
		"policy_digest": func(p map[string]any) {
			p["manifest"].(map[string]any)["instrumentation"].(map[string]any)["policy_digest"] = "sha256:" + strings.Repeat("f", 64)
		},
		"extra_field": func(p map[string]any) { p["command"] = "sh" },
		"foreign_manifest": func(p map[string]any) {
			p["manifest"].(map[string]any)["graph_digest"] = "sha256:" + strings.Repeat("b", 64)
		},
		"foreign_projection": func(p map[string]any) {
			p["projections"].(map[string]any)["test-slice"].(map[string]any)["unit_ids"] = []any{"foreign"}
		},
		"capacity_forgery": func(p map[string]any) { p["capacities"].(map[string]any)["cpu"] = 3 },
		"cache_binding":    func(p map[string]any) { p["cache"].([]any)[0].(map[string]any)["unit_id"] = "foreign" },
	} {
		t.Run(name, func(t *testing.T) {
			p := planFixture(t)
			mutate(p)
			if _, err := decodePlan(planBytes(t, p)); err == nil {
				t.Fatal("accepted invalid plan")
			}
		})
	}
	for name, mutate := range map[string]func(map[string]any){
		"missing_dependency": func(u map[string]any) { u["needs"] = []any{"missing"} },
		"cycle":              func(u map[string]any) { u["needs"] = []any{"unit:a"} },
		"unsafe_output":      func(u map[string]any) { u["current_run_evidence_outputs"] = []any{"../outside"} },
		"conflicting_locks":  func(u map[string]any) { u["exclusive_locks"] = []any{"host_activity"} },
		"exhausted_capacity": func(u map[string]any) { u["resource_claims"] = map[string]any{"cpu": 3} },
		"unbound_row_output": func(u map[string]any) {
			u["current_run_evidence_outputs"] = []any{"rows/harness.command_surface.unit.example.json", "unit-results/unit-a.json"}
		},
		"row_without_output": func(u map[string]any) {
			u["row_bindings"] = []any{map[string]any{"row_id": "harness.command_surface.unit.example", "runner": "go"}}
		},
		"unknown_runner": func(u map[string]any) {
			u["row_bindings"] = []any{map[string]any{"row_id": "harness.command_surface.unit.example", "runner": "unknown"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := planFixture(t)
			mutate(p["graph"].(map[string]any)["units"].([]any)[0].(map[string]any))
			rehashPlan(t, p)
			if _, err := decodePlan(planBytes(t, p)); err == nil {
				t.Fatal("accepted invalid graph")
			}
		})
	}
	for _, bytes := range [][]byte{[]byte(strings.Replace(string(valid), `"schema_id":"cartulary.harness_runtime_plan.v5"`, `"schema_id":"cartulary.harness_runtime_plan.v5","schema_id":"cartulary.harness_runtime_plan.v5"`, 1)), append(valid, []byte("{}")...), valid[:len(valid)-1], []byte(strings.Repeat(" ", maximumPlanBytes+1))} {
		if _, err := decodePlan(bytes); err == nil {
			t.Fatal("accepted malformed publication")
		}
	}
}

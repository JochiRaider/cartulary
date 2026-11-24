package harnessruntime

import (
	"encoding/json"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
)

// fixtureUnit is a planner unit that runs `true` with no fixture.
func fixtureUnit(id string, needs ...string) *harnessplan.Unit {
	return &harnessplan.Unit{ID: id, OwnerID: "harness.command_surface", Kind: "runner", Command: harnessplan.Command{Executable: "true"},
		Needs: needs, Claims: map[string]int64{"cpu": 1, "process": 1}, SharedLocks: []string{"host_activity"}, FixtureLease: "none",
		PostgresIsolation: "none", CachePolicy: "none", TimeoutMS: 1000, FailurePolicy: harnessplan.FailurePolicy{BlockDescendants: true}, EstimatedWorkMS: 1}
}

// fixturePlan is a test-slice plan over units (one unit:a by default), shaped
// as the planner hands it to the runtime.
func fixturePlan(t *testing.T, units ...*harnessplan.Unit) *harnessplan.Plan {
	t.Helper()
	if len(units) == 0 {
		units = []*harnessplan.Unit{fixtureUnit("unit:a")}
	}
	graph, err := harnessplan.NewGraph(units...)
	if err != nil {
		t.Fatal(err)
	}
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
	d := "sha256:" + strings.Repeat("a", 64)
	snapshot := harnessplan.CapabilitySnapshot{SchemaID: "cartulary.harness_capability_snapshot.v1", CPUTokens: 2, MemoryBytes: 104857600, ProcessSlots: 4, IOTokens: 2,
		PostgresLanes: 1, PostgresHeavyLanes: 1, ObjectStoreLanes: 1, WritableVolume: true, PortLanes: 1, Services: map[string]bool{}, Sources: map[string]string{}}
	snapshot.SnapshotDigest, _ = semanticDigest(snapshot)
	sourceInputs := []harnessplan.SourceInput{{Path: "fixture.go", Kind: "file", Mode: "0644", ByteDigest: d}}
	sourceDigest, _ := semanticDigest(map[string]any{"schema_id": "cartulary.source_snapshot_digest.v2", "entries": sourceInputs})
	const commandID = "cartulary.harness.command.test_slice.v1"
	ids := slices.Sorted(maps.Keys(graph.Units))
	reason := "mode_off"
	cache := make([]harnessplan.CacheFact, len(ids))
	for i, id := range ids {
		cache[i] = harnessplan.CacheFact{UnitID: id, Reason: &reason}
	}
	return &harnessplan.Plan{
		SchemaID: harnessplan.PlanSchema, Graph: graph,
		Manifest: harnessplan.Manifest{SchemaID: "cartulary.harness_run_manifest.v3", RunID: "fixture", CommandID: commandID, Target: "test-slice",
			DeclaredInputs: map[string]string{}, SourceCommit: strings.Repeat("a", 40), SourceState: "clean", SourceDigest: sourceDigest, ToolchainDigest: d,
			SystemDigest: snapshot.SnapshotDigest, CapabilitySnapshot: snapshot, GoWorkerBudget: 2, GraphDigest: graph.Digest, CacheMode: "off",
			StartedAt: "2026-10-11T00:00:00Z", Instrumentation: harnessplan.Instrumentation{Mode: "off", PolicyDigest: policyDigest}},
		InputEnvironment:           harnessplan.InputEnvironment{Clear: []string{"CARTULARY_MAKE_INPUT_SOURCES"}, Forward: map[string]string{}},
		BrowserLaneLimit:           4,
		HostProfile:                harnessplan.HostProfile{Platform: "linux", Architecture: "amd64", KernelRelease: "fixture", CPUModels: []string{"fixture"}},
		InstrumentationPolicyBytes: string(policyBytes),
		Projections:                map[string]harnessplan.Projection{"test-slice": {CommandID: commandID, UnitIDs: ids}},
		Capacities: map[string]int64{"cpu": 2, "memory_mb": 100, "process": 4, "io": 2, "postgres": 1, "postgres_heavy": 1, "object_store": 1,
			"port_lane": 1, "browser_stack": 1, "service_stack": 1, "volume": 1},
		AgingQuantumMS: 1000, Bindings: map[string]string{"NODE_BIN": "/private/node"},
		Cache: cache, SourceInputs: sourceInputs, CacheInputs: map[string][]int{},
	}
}

// planFixture adopts fixturePlan(units...) as the runtime executes it.
func planFixture(t *testing.T, units ...*harnessplan.Unit) *Plan {
	t.Helper()
	return adoptFixture(t, fixturePlan(t, units...))
}

func adoptFixture(t *testing.T, source *harnessplan.Plan) *Plan {
	t.Helper()
	plan, err := adoptPlan(source)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestRuntimeUnitInputEnvironment(t *testing.T) {
	source := fixturePlan(t)
	source.InputEnvironment = harnessplan.InputEnvironment{
		Clear:   []string{"CARTULARY_MAKE_INPUT_SOURCES", "FUTURE_SELECTOR", "RUNTIME_VALUE", "SUITE"},
		Forward: map[string]string{"RUNTIME_VALUE": "admitted"},
	}
	plan := adoptFixture(t, source)
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

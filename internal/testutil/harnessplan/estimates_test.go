package harnessplan

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// retainedRun writes a minimal passing retained run: its summary, plan,
// unit events (runner intervals, one cache hit) and row results.
func retainedRun(t *testing.T, units []map[string]any, events []map[string]any, rows map[string]int64) string {
	t.Helper()
	dir := t.TempDir()
	write := func(name string, value any) {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("run-summary.json", map[string]any{"run_id": "fixture-run", "status": "pass"})
	write("plan.json", map[string]any{"schema_id": "cartulary.harness_runtime_plan.v10", "graph": map[string]any{"units": units},
		"manifest": map[string]any{"run_id": "fixture-run", "target": "check", "source_digest": "sha256:" + strings.Repeat("a", 64)}})
	var stream []byte
	for _, event := range events {
		line, _ := json.Marshal(event)
		stream = append(append(stream, line...), '\n')
	}
	if err := os.WriteFile(filepath.Join(dir, "unit-events.ndjson"), stream, 0o600); err != nil {
		t.Fatal(err)
	}
	for id, duration := range rows {
		write("rows/"+id+".json", map[string]any{"row_id": id, "terminal_state": "passed", "duration_ms": duration})
	}
	return dir
}

func runner(unit string, interval string, start, end int64) []map[string]any {
	return []map[string]any{
		{"event": "phase_started", "unit_id": unit, "monotonic_ms": start, "interval_id": interval, "phase": "runner"},
		{"event": "phase_finished", "unit_id": unit, "monotonic_ms": end, "interval_id": interval, "phase": "runner"},
	}
}

func TestObservedEstimates(t *testing.T) {
	in, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	in.Estimates = Estimates{SchemaID: EstimatesSchema, Rows: map[string][]int64{}, Operations: map[string][]int64{}, Generations: map[string][]int64{}}
	var goRows []Row
	for _, row := range in.Catalog.Rows {
		if row.Runner == RunnerGo && row.EvidenceClass == "unit" && row.EstimatedWorkMS == 0 {
			goRows = append(goRows, row)
		}
	}
	if len(goRows) < 4 {
		t.Fatal("fixture needs four unestimated Go unit rows")
	}
	a, b := goRows[0], goRows[1]
	go_ := in.Owner.InvocationPolicies[RunnerGo]
	overhead := go_.StartupMS + go_.CleanupMS
	var operation string
	for id := range in.Owner.PolicyUnits {
		operation = id
		break
	}

	t.Run("an observed median replaces the authored prior", func(t *testing.T) {
		if got := in.rowEstimate(a); got != in.Owner.EvidenceEstimatesMS[a.EvidenceClass] {
			t.Fatalf("prior %d", got)
		}
		in.Estimates.Rows[a.RowID] = []int64{900, 100, 300}
		in.Estimates.Operations[operation] = []int64{4000, 6000}
		defer func() { delete(in.Estimates.Rows, a.RowID); delete(in.Estimates.Operations, operation) }()
		if got := in.rowEstimate(a); got != 300 {
			t.Fatalf("median of three: %d", got)
		}
		if got := in.operationEstimate(operation); got != 5000 {
			t.Fatalf("median of two: %d", got)
		}
	})

	t.Run("a run folds row shares and operation time, and skips cached units", func(t *testing.T) {
		units := []map[string]any{
			{"unit_id": "go:pair", "kind": "runner", "row_bindings": []map[string]string{{"row_id": a.RowID, "runner": "go"}, {"row_id": b.RowID, "runner": "go"}}},
			{"unit_id": "go:cached", "kind": "runner", "row_bindings": []map[string]string{{"row_id": goRows[2].RowID, "runner": "go"}}},
			{"unit_id": "operation:" + operation, "kind": "policy", "row_bindings": []map[string]string{}},
			{"unit_id": "operation:retired.operation", "kind": "policy", "row_bindings": []map[string]string{}},
		}
		events := slices.Concat(
			runner("go:pair", "phase:1", 1000, 1000+overhead+4000),
			runner("operation:"+operation, "phase:2", 0, 7000),
			[]map[string]any{{"event": "cache_hit", "unit_id": "go:cached", "monotonic_ms": 5}},
		)
		run := retainedRun(t, units, events, map[string]int64{a.RowID: 1000, b.RowID: 3000, goRows[2].RowID: 50})
		folded, report, err := in.foldRun(run)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(folded.Rows[a.RowID], []int64{1000}) || !slices.Equal(folded.Rows[b.RowID], []int64{3000}) ||
			folded.Rows[goRows[2].RowID] != nil || !slices.Equal(folded.Operations[operation], []int64{7000}) {
			t.Fatalf("folded %v %v", folded.Rows, folded.Operations)
		}
		if report.Rows != 2 || report.Operations != 1 || len(folded.Runs) != 1 || folded.Runs[0].RunID != "fixture-run" {
			t.Fatalf("report %+v runs %v", report, folded.Runs)
		}
		if len(in.Estimates.Rows) != 0 {
			t.Fatal("folding changed the loaded estimates")
		}
		in.Estimates = folded
		defer func() {
			in.Estimates = Estimates{SchemaID: EstimatesSchema, Rows: map[string][]int64{}, Operations: map[string][]int64{}}
		}()
		if _, _, err := in.foldRun(run); !errors.Is(err, ErrRefresh) {
			t.Fatalf("a run folded twice: %v", err)
		}
	})

	t.Run("runner time beyond the allowance is shared by reported duration; history is bounded", func(t *testing.T) {
		estimates := Estimates{SchemaID: EstimatesSchema, Rows: map[string][]int64{a.RowID: {1, 2, 3}}, Operations: map[string][]int64{}}
		saved := in.Estimates
		in.Estimates = estimates
		defer func() { in.Estimates = saved }()
		units := []map[string]any{{"unit_id": "go:pair", "kind": "runner", "row_bindings": []map[string]string{{"row_id": a.RowID, "runner": "go"}, {"row_id": b.RowID, "runner": "go"}}}}
		run := retainedRun(t, units, runner("go:pair", "phase:1", 0, overhead+8000), map[string]int64{a.RowID: 1000, b.RowID: 3000})
		folded, _, err := in.foldRun(run)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(folded.Rows[a.RowID], []int64{2, 3, 2000}) || !slices.Equal(folded.Rows[b.RowID], []int64{6000}) {
			t.Fatalf("shares %v %v", folded.Rows[a.RowID], folded.Rows[b.RowID])
		}
	})

	t.Run("a failed run and an estimate over the packing bound are refused", func(t *testing.T) {
		units := []map[string]any{{"unit_id": "go:one", "kind": "runner", "row_bindings": []map[string]string{{"row_id": a.RowID, "runner": "go"}}}}
		run := retainedRun(t, units, runner("go:one", "phase:1", 0, in.Owner.GoUnitTargetMS+overhead), map[string]int64{a.RowID: in.Owner.GoUnitTargetMS})
		if _, _, err := in.foldRun(run); !errors.Is(err, ErrRefresh) || !strings.Contains(err.Error(), "packing bound") {
			t.Fatalf("an oversized row was folded: %v", err)
		}
		failed := retainedRun(t, units, runner("go:one", "phase:1", 0, 10), map[string]int64{a.RowID: 1})
		if err := os.WriteFile(filepath.Join(failed, "run-summary.json"), []byte(`{"run_id":"fixture-run","status":"fail"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := in.foldRun(failed); !errors.Is(err, ErrRefresh) {
			t.Fatalf("a failed run was folded: %v", err)
		}
	})

	t.Run("a unit generation claims three halves of its base and its rows' shares", func(t *testing.T) {
		saved := in.Estimates
		defer func() { in.Estimates = saved }()
		topology := in.Topology.PhysicalResourceClaims["postgres"]
		in.Estimates = Estimates{SchemaID: EstimatesSchema, Rows: map[string][]int64{}, Operations: map[string][]int64{},
			Generations: map[string][]int64{a.RowID: {300, 600, 400}, b.RowID: {100}, goRows[2].RowID: {10_000}}, GenerationBase: []int64{150, 160}}
		sized := in.generationClaims([]string{a.RowID, b.RowID})
		if sized["memory_mb"] != 1536 {
			t.Fatalf("three halves of 160 + 600 + 100 MiB rounded to the step: %d", sized["memory_mb"])
		}
		for resource, quantity := range topology {
			if resource != "memory_mb" && sized[resource] != quantity {
				t.Fatalf("sizing changed %s: %d", resource, sized[resource])
			}
		}
		if got := in.generationClaims([]string{b.RowID})["memory_mb"]; got != generationFloorMB {
			t.Fatalf("a small generation is held at the floor: %d", got)
		}
		if got := in.generationClaims([]string{goRows[2].RowID})["memory_mb"]; got != topology["memory_mb"] {
			t.Fatalf("a large generation is held at the topology claim: %d", got)
		}
		var unobserved string
		for _, row := range in.Catalog.Rows {
			if in.Estimates.Generations[row.RowID] == nil {
				unobserved = row.RowID
				break
			}
		}
		for _, rows := range [][]string{nil, {a.RowID, unobserved}} {
			if got := in.generationClaims(rows); !maps.Equal(got, topology) {
				t.Fatalf("rows %v without every share take the topology claim: %v", rows, got)
			}
		}
		in.Estimates.GenerationBase = nil
		if got := in.generationClaims([]string{a.RowID}); !maps.Equal(got, topology) {
			t.Fatalf("without a base a generation takes the topology claim: %v", got)
		}
		if topology["memory_mb"] != 4352 || in.generationBudget() != 2901 {
			t.Fatalf("sizing changed the loaded topology, or its budget is %d", in.generationBudget())
		}
	})

	t.Run("a run shares each unit generation's memory above its base by duration", func(t *testing.T) {
		units := []map[string]any{
			{"unit_id": "go:pair", "kind": "runner", "row_bindings": []map[string]string{{"row_id": a.RowID, "runner": "go"}, {"row_id": b.RowID, "runner": "go"}}},
			{"unit_id": "go:small", "kind": "runner", "row_bindings": []map[string]string{{"row_id": goRows[2].RowID, "runner": "go"}}},
			{"unit_id": "go:cached", "kind": "runner", "row_bindings": []map[string]string{{"row_id": goRows[3].RowID, "runner": "go"}}},
		}
		events := slices.Concat(runner("go:pair", "phase:1", 0, overhead+4000), runner("go:small", "phase:2", 0, overhead+100),
			[]map[string]any{{"event": "cache_hit", "unit_id": "go:cached", "monotonic_ms": 5}})
		run := retainedRun(t, units, events, map[string]int64{a.RowID: 1000, b.RowID: 3000, goRows[2].RowID: 50, goRows[3].RowID: 10})
		footprint := func(suite, ref string, peak int64) {
			dir := filepath.Join(run, "_shared", "test-services", suite)
			data, _ := json.Marshal(map[string]any{"schema_id": GenerationFootprintSchema, "allocation_ref": ref, "data_capacity_bytes": 1 << 30, "memory_peak_bytes": peak})
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "postgres-footprint.json"), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		footprint("one", "postgres:unit:go:pair", 950<<20)
		footprint("two", "postgres:unit:go:small", 150<<20)
		footprint("three", "postgres:unit:go:cached", 100<<20)
		footprint("four", "postgres:shared", 2000<<20)
		folded, report, err := in.foldRun(run)
		if err != nil {
			t.Fatal(err)
		}
		// The base is the smallest uncached unit generation; the pair's
		// 800 MiB above it divides one to three by the rows' durations.
		if !slices.Equal(folded.GenerationBase, []int64{150}) || !slices.Equal(folded.Generations[a.RowID], []int64{200}) ||
			!slices.Equal(folded.Generations[b.RowID], []int64{600}) || !slices.Equal(folded.Generations[goRows[2].RowID], []int64{1}) ||
			folded.Generations[goRows[3].RowID] != nil || report.Generations != 3 {
			t.Fatalf("base %v generations %v report %+v", folded.GenerationBase, folded.Generations, report)
		}
		footprint("five", "", 1)
		if _, _, err := in.foldRun(run); !errors.Is(err, ErrRefresh) || !strings.Contains(err.Error(), "footprint") {
			t.Fatalf("a malformed footprint was folded: %v", err)
		}
	})

	t.Run("the written document is canonical, sorted and strict", func(t *testing.T) {
		data, err := canonicalEstimates(Estimates{SchemaID: EstimatesSchema, Runs: []FoldedRun{{RunID: "r", Target: "check", SourceDigest: "sha256:" + strings.Repeat("b", 64)}},
			Rows: map[string][]int64{"z.row": {2}, "a.row": {1, 3}}, Operations: map[string][]int64{}, Generations: map[string][]int64{"g.row": {768}}, GenerationBase: []int64{150}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := canonicaljson.DecodeStrict(data); err != nil {
			t.Fatal(err)
		}
		if strings.Index(string(data), `"a.row"`) > strings.Index(string(data), `"z.row"`) || !strings.Contains(string(data), "\n    \"a.row\": [1,3],\n") ||
			!strings.Contains(string(data), "\"generations\": {\n    \"g.row\": [768]\n  },\n  \"generation_base\": [150]\n}\n") {
			t.Fatalf("document %s", data)
		}
	})
}

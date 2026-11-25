package harnessplan

import (
	"encoding/json"
	"errors"
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
	in.Estimates = Estimates{SchemaID: EstimatesSchema, Rows: map[string][]int64{}, Operations: map[string][]int64{}}
	var goRows []Row
	for _, row := range in.Catalog.Rows {
		if row.Runner == RunnerGo && row.EvidenceClass == "unit" && row.EstimatedWorkMS == 0 {
			goRows = append(goRows, row)
		}
	}
	if len(goRows) < 3 {
		t.Fatal("fixture needs three unestimated Go unit rows")
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

	t.Run("the written document is canonical, sorted and strict", func(t *testing.T) {
		data, err := canonicalEstimates(Estimates{SchemaID: EstimatesSchema, Runs: []FoldedRun{{RunID: "r", Target: "check", SourceDigest: "sha256:" + strings.Repeat("b", 64)}},
			Rows: map[string][]int64{"z.row": {2}, "a.row": {1, 3}}, Operations: map[string][]int64{}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := canonicaljson.DecodeStrict(data); err != nil {
			t.Fatal(err)
		}
		if strings.Index(string(data), `"a.row"`) > strings.Index(string(data), `"z.row"`) || !strings.Contains(string(data), "\n    \"a.row\": [1,3],\n") {
			t.Fatalf("document %s", data)
		}
	})
}

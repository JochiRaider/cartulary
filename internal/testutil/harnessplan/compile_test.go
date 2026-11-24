package harnessplan

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestPlannerCompilation(t *testing.T) {
	inputs, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	options := Options{CPUTokens: 19, GoVersion: "1.27.4"}
	rowsWhere := func(keep func(Row) bool) []string {
		var ids []string
		for _, row := range inputs.Catalog.Rows {
			if keep(row) {
				ids = append(ids, row.RowID)
			}
		}
		return ids
	}
	plannable := func(row Row) bool {
		return row.Runner == RunnerShell || row.Runner == RunnerGo && row.FixtureProfileID == "" && row.EvidenceClass != "measurement"
	}
	bound := func(g *Graph) map[string]string {
		rows := map[string]string{}
		for _, unit := range g.Units {
			for _, binding := range unit.RowBindings {
				rows[binding.RowID] = unit.ID
			}
		}
		return rows
	}

	t.Run("every plannable row compiles and is bound exactly once", func(t *testing.T) {
		rows := rowsWhere(plannable)
		graph, err := Compile(inputs, Selection{Rows: rows}, options)
		if err != nil {
			t.Fatal(err)
		}
		if got := bound(graph); len(got) != len(rows) {
			t.Fatalf("bound %d rows, selected %d", len(got), len(rows))
		}
		again, err := Compile(inputs, Selection{Rows: rows}, options)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(unitIDs(graph), unitIDs(again)) {
			t.Fatal("compilation is not deterministic")
		}
	})

	t.Run("an operation target compiles its dependency closure", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Target: "harness.go_vulncheck"}, options)
		if err != nil {
			t.Fatal(err)
		}
		root, ok := graph.Units["operation:harness.go_vulncheck"]
		if !ok || len(root.Needs) == 0 {
			t.Fatalf("operation unit missing or without needs: %+v", root)
		}
		if root.EstimatedWorkMS != inputs.Owner.PolicyUnits["harness.go_vulncheck"].EstimatedWorkMS {
			t.Fatal("operation estimate does not come from its policy unit")
		}
	})

	t.Run("services name their residents and scopes", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(func(row Row) bool {
			return row.Runner == RunnerGo && row.FixtureProfileID == "" && row.EvidenceClass != "measurement" && row.PostgresIsolation != "none"
		})}, options)
		if err != nil {
			t.Fatal(err)
		}
		for _, unit := range graph.Units {
			if unit.Kind != "runner" || !strings.HasPrefix(unit.ID, "go:") {
				continue
			}
			keys := map[string]string{}
			for _, resident := range unit.Residents {
				keys[resident.Key] = resident.Scope
			}
			switch unit.PostgresIsolation {
			case "exclusive":
				if keys["postgres:unit:"+unit.ID] != ScopeUnit {
					t.Fatalf("%s lacks its own generation: %v", unit.ID, keys)
				}
			case "shared":
				if keys["postgres:shared"] != ScopeGraph {
					t.Fatalf("%s lacks the shared generation: %v", unit.ID, keys)
				}
			}
			if slices.Contains(unit.ServiceDependencies, "object_store") != (keys["object_store:shared"] == ScopeGraph) {
				t.Fatalf("%s object-store resident disagrees with its services: %v", unit.ID, keys)
			}
			if !slices.Contains(unit.Needs, "operation:harness.test_service_images") {
				t.Fatalf("%s does not wait for the service images", unit.ID)
			}
		}
	})

	t.Run("Go units are balanced under their size target", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(func(row Row) bool {
			return row.Runner == RunnerGo && row.FixtureProfileID == "" && row.EvidenceClass != "measurement"
		})}, options)
		if err != nil {
			t.Fatal(err)
		}
		overhead := inputs.Owner.InvocationPolicies[RunnerGo].StartupMS + inputs.Owner.InvocationPolicies[RunnerGo].CleanupMS
		for _, unit := range graph.Units {
			if !strings.HasPrefix(unit.ID, "go:") || len(unit.RowBindings) < 2 {
				continue
			}
			work := unit.EstimatedWorkMS - overhead
			target := inputs.Owner.GoUnitTargetMS
			if unit.Claims["postgres_heavy"] > 0 {
				target = inputs.Owner.GoHeavyUnitTargetMS
			}
			// Longest-first balancing keeps a multi-row unit within the
			// target plus at most one row's estimate.
			largest := int64(0)
			for _, binding := range unit.RowBindings {
				for _, row := range inputs.Catalog.Rows {
					if row.RowID == binding.RowID {
						largest = max(largest, max(row.EstimatedWorkMS, inputs.Owner.EvidenceEstimatesMS[row.EvidenceClass]))
					}
				}
			}
			if work > target+largest {
				t.Errorf("%s holds %d ms of work against a %d ms target", unit.ID, work, target)
			}
		}
	})

	t.Run("a shell row binds to its operation's unit and policy", func(t *testing.T) {
		rows := rowsWhere(func(row Row) bool { return row.Runner == RunnerShell })
		graph, err := Compile(inputs, Selection{Rows: rows[:1]}, options)
		if err != nil {
			t.Fatal(err)
		}
		var row Row
		for _, candidate := range inputs.Catalog.Rows {
			if candidate.RowID == rows[0] {
				row = candidate
			}
		}
		unit, ok := graph.Units["operation:"+row.Selector.OperationID]
		if !ok || len(unit.RowBindings) == 0 || unit.EstimatedWorkMS != inputs.Owner.PolicyUnits[row.Selector.OperationID].EstimatedWorkMS {
			t.Fatalf("shell row %s did not bind to its operation under the operation's policy: %+v", row.RowID, unit)
		}
	})

	t.Run("unsupported and invalid selections fail", func(t *testing.T) {
		vitest := rowsWhere(func(row Row) bool { return row.Runner == RunnerVitest })
		if _, err := Compile(inputs, Selection{Rows: vitest[:1]}, options); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("vitest row: %v", err)
		}
		for name, selection := range map[string]Selection{
			"unknown row":     {Rows: []string{"module.missing.unit.x"}},
			"unknown target":  {Target: "harness.missing"},
			"two kinds":       {Target: "harness.go_vulncheck", Rows: []string{"x"}},
			"empty selection": {},
		} {
			if _, err := Compile(inputs, selection, options); !errors.Is(err, ErrSelection) {
				t.Errorf("%s: %v", name, err)
			}
		}
	})
}

func unitIDs(g *Graph) []string {
	var ids []string
	for _, unit := range g.Sorted() {
		ids = append(ids, unit.ID)
	}
	return ids
}

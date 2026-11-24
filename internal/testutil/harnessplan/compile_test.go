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
		return row.Runner == RunnerShell || row.Runner == RunnerVitest || row.Runner == RunnerGo && row.FixtureProfileID == "" && row.EvidenceClass != "measurement"
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
		graph, err := Compile(inputs, Selection{Targets: []string{"harness.go_vulncheck"}}, options)
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
			if unit.FixtureLease == "none" {
				if len(keys) != 0 {
					t.Fatalf("%s has no fixture lease but holds residents %v", unit.ID, keys)
				}
				continue
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

	t.Run("units are canonical and identified by their own contract", func(t *testing.T) {
		rows := rowsWhere(plannable)
		graph, err := Compile(inputs, Selection{Rows: rows}, options)
		if err != nil {
			t.Fatal(err)
		}
		for _, unit := range graph.Units {
			if unit.SemanticDigest != semanticDigest(unit, "semantic_digest") {
				t.Fatalf("%s semantic digest does not cover the unit", unit.ID)
			}
			if !slices.Contains(unit.EvidenceOutputs, unit.ResultPath()) || !slices.IsSorted(unit.EvidenceOutputs) {
				t.Fatalf("%s evidence outputs %v lack its result or are unsorted", unit.ID, unit.EvidenceOutputs)
			}
			if !unit.FailurePolicy.BlockDescendants {
				t.Fatalf("%s does not block its descendants", unit.ID)
			}
		}
		// A row whose contract is shared with no other selected row compiles
		// to the same unit alone as in the full selection.
		for _, unit := range graph.Units {
			if !strings.HasPrefix(unit.ID, "go:") || len(unit.RowBindings) != 1 {
				continue
			}
			alone, err := Compile(inputs, Selection{Rows: []string{unit.RowBindings[0].RowID}}, options)
			if err != nil {
				t.Fatal(err)
			}
			if other, ok := alone.Units[unit.ID]; !ok || other.SemanticDigest != unit.SemanticDigest {
				t.Fatalf("%s changes identity when compiled alone", unit.ID)
			}
			break
		}
	})

	t.Run("policy evidence outputs come from the policy unit", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Targets: []string{"harness.go_vulncheck"}}, options)
		if err != nil {
			t.Fatal(err)
		}
		unit := graph.Units["operation:harness.go_vulncheck"]
		for _, output := range inputs.Owner.PolicyUnits["harness.go_vulncheck"].EvidenceOutputs {
			if !slices.Contains(unit.EvidenceOutputs, output) {
				t.Fatalf("operation unit lacks its policy's evidence output %s: %v", output, unit.EvidenceOutputs)
			}
		}
	})

	t.Run("a row reached twice is bound once", func(t *testing.T) {
		// The suite union selects every backend unit row; selecting one of
		// them directly as well must not plan it twice.
		suite := rowsWhere(func(row Row) bool {
			return row.Runner == RunnerGo && goSuite(row) == "unit" && row.FixtureProfileID == "" && row.EvidenceClass != "measurement"
		})
		graph, err := Compile(inputs, Selection{Rows: suite[:1], Targets: []string{"harness.suite_backend_unit"}}, options)
		if err != nil {
			t.Fatal(err)
		}
		if got := bound(graph); got[suite[0]] == "" {
			t.Fatalf("%s is not bound", suite[0])
		}
	})

	t.Run("forwarded inputs reach only the selected operations", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Targets: []string{"harness.go_vulncheck"}, Forwarded: map[string]string{"GOVULNCHECK_DB": "/db"}}, options)
		if err != nil {
			t.Fatal(err)
		}
		for id, unit := range graph.Units {
			if got := unit.Command.Environment["GOVULNCHECK_DB"]; (id == "operation:harness.go_vulncheck") != (got == "/db") {
				t.Fatalf("%s environment GOVULNCHECK_DB = %q", id, got)
			}
		}
	})

	t.Run("an attached session holds the shared allocations", func(t *testing.T) {
		attached := options
		attached.AttachedServices = true
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(func(row Row) bool {
			return row.Runner == RunnerGo && row.FixtureProfileID == "" && row.EvidenceClass != "measurement" && row.PostgresIsolation != "none" && row.FixtureCapability != "none"
		})}, attached)
		if err != nil {
			t.Fatal(err)
		}
		for _, unit := range graph.Units {
			for _, resident := range unit.Residents {
				if want := map[bool]string{true: ScopeUnit, false: ScopeAttached}[strings.HasPrefix(resident.Key, "postgres:unit:")]; resident.Scope != want {
					t.Fatalf("%s resident %s has scope %s, want %s", unit.ID, resident.Key, resident.Scope, want)
				}
			}
		}
	})

	t.Run("runner units carry their rows on the command line", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(plannable)}, options)
		if err != nil {
			t.Fatal(err)
		}
		for _, unit := range graph.Units {
			if !strings.HasPrefix(unit.ID, "go:") && !strings.HasPrefix(unit.ID, "vitest:") {
				continue
			}
			var rows []string
			for _, binding := range unit.RowBindings {
				rows = append(rows, binding.RowID)
			}
			if len(unit.Command.Args) != 3 || unit.Command.Args[1] != "--row-ids" || unit.Command.Args[2] != strings.Join(rows, ",") {
				t.Fatalf("%s command %v does not name its rows", unit.ID, unit.Command.Args)
			}
		}
	})

	t.Run("runner units are balanced under their size target", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(plannable)}, options)
		if err != nil {
			t.Fatal(err)
		}
		for _, unit := range graph.Units {
			runner, _, _ := strings.Cut(unit.ID, ":")
			if (runner != RunnerGo && runner != RunnerVitest) || len(unit.RowBindings) < 2 {
				continue
			}
			policy := inputs.Owner.InvocationPolicies[runner]
			work := unit.EstimatedWorkMS - policy.StartupMS - policy.CleanupMS
			target := policy.UnitTargetMS
			if runner == RunnerGo {
				target = inputs.Owner.GoUnitTargetMS
				if unit.Claims["postgres_heavy"] > 0 {
					target = inputs.Owner.GoHeavyUnitTargetMS
				}
			}
			// Longest-first balancing keeps a multi-item unit within the
			// target plus at most one item's estimate; a Vitest item is a
			// whole file.
			items := map[string]int64{}
			largest := int64(0)
			for _, binding := range unit.RowBindings {
				row := inputs.row(binding.RowID)
				key := row.RowID
				if runner == RunnerVitest {
					key = row.Selector.File
				}
				items[key] += max(row.EstimatedWorkMS, inputs.Owner.EvidenceEstimatesMS[row.EvidenceClass])
				largest = max(largest, items[key])
			}
			if work > target+largest {
				t.Errorf("%s holds %d ms of work against a %d ms target", unit.ID, work, target)
			}
		}
	})

	t.Run("a Vitest file's rows of one contract share a unit with the runner's preparation", func(t *testing.T) {
		rows := rowsWhere(func(row Row) bool { return row.Runner == RunnerVitest })
		graph, err := Compile(inputs, Selection{Rows: rows}, options)
		if err != nil {
			t.Fatal(err)
		}
		fileUnit := map[string]string{}
		policy := inputs.Owner.InvocationPolicies[RunnerVitest]
		for _, unit := range graph.Units {
			if !strings.HasPrefix(unit.ID, "vitest:") {
				continue
			}
			if unit.Command.Environment["VITEST_MAX_WORKERS"] != "1" || unit.Command.Environment["CARTULARY_TEST_TARGET"] != "frontend-unit" {
				t.Fatalf("%s environment %v", unit.ID, unit.Command.Environment)
			}
			if !slices.Contains(unit.Needs, "operation:harness.frontend_install") || unit.Claims["memory_mb"] < policy.CoordinatorClaims["memory_mb"] {
				t.Fatalf("%s needs %v claims %v", unit.ID, unit.Needs, unit.Claims)
			}
			projects := map[string]bool{}
			for _, binding := range unit.RowBindings {
				row := inputs.row(binding.RowID)
				projects[row.Selector.ProjectID] = true
				// The unit ID's second segment names its execution contract;
				// a file's rows of one contract share an invocation.
				key := strings.Split(unit.ID, ":")[1] + " " + row.Selector.File
				if other, ok := fileUnit[key]; ok && other != unit.ID {
					t.Fatalf("file %s runs in %s and %s", row.Selector.File, other, unit.ID)
				}
				fileUnit[key] = unit.ID
				if row.Selector.ProjectID == "renderer-node" && row.EvidenceClass == "integration" && !slices.Contains(unit.Needs, "operation:harness.playwright_install") {
					t.Fatalf("%s runs renderer integration rows without the browser install", unit.ID)
				}
			}
			if len(projects) != 1 {
				t.Fatalf("%s mixes Vitest projects %v", unit.ID, projects)
			}
		}
		if len(fileUnit) == 0 {
			t.Fatal("no Vitest units compiled")
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
		browser := rowsWhere(func(row Row) bool { return row.Runner == RunnerPlaywright })
		if _, err := Compile(inputs, Selection{Rows: browser[:1]}, options); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("playwright row: %v", err)
		}
		for name, selection := range map[string]Selection{
			"unknown row":     {Rows: []string{"module.missing.unit.x"}},
			"unknown target":  {Targets: []string{"harness.missing"}},
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

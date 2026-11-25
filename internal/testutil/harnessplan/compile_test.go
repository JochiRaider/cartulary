package harnessplan

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	"github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
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
	every := func(Row) bool { return true }
	bound := func(g *Graph) map[string]string {
		rows := map[string]string{}
		for _, unit := range g.Units {
			for _, binding := range unit.RowBindings {
				rows[binding.RowID] = unit.ID
			}
		}
		return rows
	}

	t.Run("a browser row builds exactly the frontend artifact its stage serves", func(t *testing.T) {
		for id, producer := range map[string]string{
			"module.auth.browser.browser_sign_in_exposes_the_ordinary_session_sur_2ec8df6229":      "build_web",
			"web.design.accessibility.account_menu_accessible_keyboard_and_focus_state_3c31f28451": "build_web",
			"web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde":        "build_web",
			"module.networkflow.measurement.saved_graph_dom_ceiling":                               "build_web_measurement",
		} {
			graph, err := Compile(inputs, Selection{Rows: []string{id}}, options)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := graph.Units["operation:harness."+producer]; !ok {
				t.Fatalf("%s does not build %s", id, producer)
			}
			// Measurement still builds the backend's embedded production assets;
			// production work never acquires the measurement producer.
			if _, ok := graph.Units["operation:harness.build_web_measurement"]; ok && producer == "build_web" {
				t.Fatalf("production row %s builds measurement output", id)
			}
		}
	})

	t.Run("a race row runs under the race detector and a shuffle seed reaches Go units", func(t *testing.T) {
		goRows := rowsWhere(func(row Row) bool { return row.Runner == RunnerGo && row.Status == "active" })
		var race, plain string
		for _, id := range goRows {
			if inputs.Owner.RowExecutionContexts[id].GoRace {
				race = id
			} else if plain == "" {
				plain = id
			}
		}
		if race == "" || plain == "" {
			t.Fatal("the catalog needs a race and an ordinary Go row")
		}
		graph, err := Compile(inputs, Selection{Rows: []string{race, plain}}, options)
		if err != nil {
			t.Fatal(err)
		}
		units := bound(graph)
		raced, ordinary := graph.Units[units[race]].Command, graph.Units[units[plain]].Command
		if !slices.Contains(raced.Args, "-race") || raced.Environment["CGO_ENABLED"] != "1" || slices.Contains(ordinary.Args, "-race") || ordinary.Environment["CGO_ENABLED"] != "0" {
			t.Fatalf("race commands %v / %v", raced, ordinary)
		}
		seeded := options
		seeded.GoShuffleSeed = "17"
		shuffled, err := Compile(inputs, Selection{Rows: []string{plain}}, seeded)
		if err != nil {
			t.Fatal(err)
		}
		unseeded, err := Compile(inputs, Selection{Rows: []string{plain}}, options)
		if err != nil {
			t.Fatal(err)
		}
		unit := shuffled.Units[bound(shuffled)[plain]]
		if !slices.Contains(unit.Command.Args, "-shuffle=17") || shuffled.Digest == unseeded.Digest {
			t.Fatal("the shuffle seed does not reach the Go unit's identity")
		}
		for _, invalid := range []string{"0", "on", "017", "2147483648"} {
			seeded.GoShuffleSeed = invalid
			if _, err := Compile(inputs, Selection{Rows: []string{plain}}, seeded); !errors.Is(err, ErrSelection) {
				t.Fatalf("shuffle seed %q: %v", invalid, err)
			}
		}
	})

	t.Run("every row compiles and is bound exactly once", func(t *testing.T) {
		rows := rowsWhere(every)
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
		if root.EstimatedWorkMS != inputs.operationEstimate("harness.go_vulncheck") {
			t.Fatal("operation estimate is not its planning estimate")
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
		rows := rowsWhere(every)
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
			// A browser group is finalizable: its summaries and target
			// finalizer still run when it fails.
			if group := strings.HasPrefix(unit.ID, "browser_group:"); unit.FailurePolicy.BlockDescendants == group {
				t.Fatalf("%s block_descendants = %v", unit.ID, unit.FailurePolicy.BlockDescendants)
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
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(every)}, options)
		if err != nil {
			t.Fatal(err)
		}
		modules := 0
		for _, unit := range graph.Units {
			var rows, tests []string
			for _, binding := range unit.RowBindings {
				rows = append(rows, binding.RowID)
				tests = append(tests, binding.Selectors...)
				if (binding.Runner == RunnerGo) != (len(binding.Selectors) != 0) {
					t.Fatalf("%s binding %+v", unit.ID, binding)
				}
			}
			switch {
			case strings.HasPrefix(unit.ID, "vitest:"):
				if len(unit.Command.Args) != 3 || unit.Command.Args[1] != "--row-ids" || unit.Command.Args[2] != strings.Join(rows, ",") {
					t.Fatalf("%s command %v does not name its rows", unit.ID, unit.Command.Args)
				}
			case strings.HasPrefix(unit.ID, "go:"):
				// One package, exactly its rows' tests, under the fixed build context.
				args := unit.Command.Args
				pkg := inputs.Catalog.Rows[inputs.rowIndex[rows[0]]].Selector.Package
				tests = dedupe(tests)
				for i := range tests {
					tests[i] = regexp.QuoteMeta(tests[i])
				}
				tail := []string{"-json", "-count=1"}
				if inputs.Owner.RowExecutionContexts[rows[0]].GoRace {
					tail = append(tail, "-race")
				}
				tail = append(tail, fmt.Sprintf("-p=%d", min(unit.Claims["cpu"], 16)), "-run", "^(?:"+strings.Join(tests, "|")+")$")
				head := []string{"test"}
				if strings.HasPrefix(pkg, "./referencepacks/") {
					modules++
					directory, within, _ := inputs.goModule(pkg)
					head = []string{"-C", directory, "test", "-mod=readonly"}
					pkg = within
				}
				want := append(append(head, tail...), pkg)
				if unit.Command.Executable != "go" || !slices.Equal(args, want) || unit.Command.Environment["GOWORK"] != "off" || unit.Command.Environment["GOFLAGS"] != "" {
					t.Fatalf("%s command %v\nwant %v", unit.ID, unit.Command, want)
				}
			}
		}
		if modules == 0 {
			t.Fatal("the catalog needs a Go row of an independent module")
		}
		if _, _, err := inputs.goModule("./referencepacks/undeclared/data"); !errors.Is(err, ErrSelection) {
			t.Fatalf("an unregistered module package: %v", err)
		}
	})

	t.Run("runner units are balanced under their size target", func(t *testing.T) {
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(every)}, options)
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
				items[key] += inputs.rowEstimate(row)
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
		if !ok || len(unit.RowBindings) == 0 || unit.EstimatedWorkMS != inputs.operationEstimate(row.Selector.OperationID) {
			t.Fatalf("shell row %s did not bind to its operation under the operation's policy: %+v", row.RowID, unit)
		}
	})

	t.Run("a Go measurement row is planned alone", func(t *testing.T) {
		measurement := rowsWhere(func(row Row) bool { return row.Runner == RunnerGo && row.EvidenceClass == "measurement" })
		if len(measurement) == 0 {
			t.Fatal("no Go measurement rows")
		}
		graph, err := Compile(inputs, Selection{Rows: rowsWhere(func(row Row) bool { return row.Runner == RunnerGo })}, options)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range measurement {
			unit := graph.Units[bound(graph)[id]]
			if len(unit.RowBindings) != 1 {
				t.Errorf("measurement row %s shares unit %s", id, unit.ID)
			}
			// Its resource profile's runner timeout, not the default, covers it.
			if unit.TimeoutMS < unit.EstimatedWorkMS {
				t.Errorf("measurement unit %s times out at %d ms before its %d ms estimate", unit.ID, unit.TimeoutMS, unit.EstimatedWorkMS)
			}
		}
	})

	t.Run("measurement groups clone one snapshot per fixture profile", func(t *testing.T) {
		populated := rowsWhere(func(row Row) bool { return row.FixtureProfileID != "" })
		graph, err := Compile(inputs, Selection{Rows: populated}, options)
		if err != nil {
			t.Fatal(err)
		}
		builders := map[string]*Unit{}
		for _, unit := range graph.Units {
			if unit.Kind == "fixture_builder" {
				builders[unit.FixtureProfileID] = unit
			}
		}
		profiles := map[string]bool{}
		for _, id := range populated {
			profiles[inputs.row(id).FixtureProfileID] = true
		}
		if len(builders) != len(profiles) {
			t.Fatalf("%d builders for %d fixture profiles", len(builders), len(profiles))
		}
		for profileID, builder := range builders {
			profile, _ := performancefixtureprofile.Lookup(profileID)
			key, err := performancefixture.SnapshotKey(profile, inputs.MigrationDigest)
			if err != nil {
				t.Fatal(err)
			}
			cluster := builder.Command.Environment[clusterKeyEnvironment]
			if builder.ID != "fixture_snapshot:default:"+profileID+":"+key || builder.SnapshotKey != key ||
				len(builder.Residents) != 1 || builder.Residents[0].Key != cluster || builder.Residents[0].Scope != ScopeGraph {
				t.Fatalf("builder %s is not keyed and resident by its snapshot: %+v", builder.ID, builder.Residents)
			}
			for _, need := range builder.Needs {
				if strings.Contains(need, "playwright") || strings.Contains(need, "frontend") {
					t.Fatalf("builder %s waits for browser preparation %s", builder.ID, need)
				}
			}
		}
		target := graph.Units["browser_target_summary:browser-e2e-measurement"]
		if target == nil || !slices.Contains(target.EvidenceOutputs, "browser-e2e-measurement/frontend-measurement-aggregate.json") {
			t.Fatal("the measurement target finalizer does not aggregate the summaries")
		}
		for _, id := range populated {
			group := graph.Units[bound(graph)[id]]
			builder := builders[inputs.row(id).FixtureProfileID]
			environment := group.Command.Environment
			if len(group.RowBindings) != 1 || !slices.Contains(group.Needs, builder.ID) || group.SnapshotKey != builder.SnapshotKey ||
				environment["CARTULARY_FIXTURE_ROW_ID"] != id || environment["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"] != builder.ID ||
				environment[clusterKeyEnvironment] != builder.Command.Environment[clusterKeyEnvironment] || !slices.ContainsFunc(group.Residents, func(r Resident) bool {
				return r.Key == environment[clusterKeyEnvironment]
			}) {
				t.Fatalf("group %s does not clone its row's snapshot alone", group.ID)
			}
			summary := graph.Units["browser_measurement_summary:measurement:"+strings.TrimPrefix(group.ID, "browser_group:measurement:")]
			if summary == nil || !slices.Equal(summary.Needs, []string{group.ID}) || !slices.Contains(target.Needs, summary.ID) {
				t.Fatalf("group %s has no summary the target finalizer waits for", group.ID)
			}
		}
	})

	t.Run("invalid selections fail", func(t *testing.T) {
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

package harnessplan

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCacheFacts(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                            "module example.com/m\n\ngo 1.27\n",
		"a/a.go":                            "package a\n\nimport (\n\t\"fmt\"\n\t\"example.com/m/b\"\n)\n\nvar _ = fmt.Sprint(b.B)\n",
		"a/a_test.go":                       "//go:build never\n\npackage a\n\nimport _ \"example.com/m/t\"\n",
		"a/testdata/input.txt":              "data\n",
		"b/b.go":                            "package b\n\nconst B = 1\n",
		"b/launch.go":                       "package b\n\nconst (\n\tlaunched = \"tools/launch/entry.mjs\"\n\tabsent   = \"tools/launch/absent.mjs\"\n\tdir      = \"tools/launch\"\n\tread     = \"configs2/read.json\"\n)\n",
		"tools/launch/entry.mjs":            "import { d } from \"./dep.mjs\";\n",
		"tools/launch/dep.mjs":              "export const d = 1;\n",
		"tools/launch/unused.mjs":           "export {};\n",
		"configs2/read.json":                "{}\n",
		"t/t.go":                            "package t\n",
		"c/c.go":                            "package c\n",
		"r/r.go":                            "package r\n\nimport _ \"example.com/m/s\"\n",
		"r/r_test.go":                       "package r\n\nimport _ \"example.com/m/c\"\n",
		"s/s.go":                            "package s\n",
		"configs/app.json":                  "{}\n",
		"tools/harness_helpers.json":        "{}\n",
		"tools/other.json":                  "{}\n",
		"tools/test_families/owner.json":    "{}\n",
		"contracts/probe-owner/case.json":   "{}\n",
		"fixtures/case/input.json":          "{}\n",
		"fixtures/other/input.json":         "{}\n",
		"nested/go.mod":                     "module example.com/m/nested\n",
		"nested/go.sum":                     "\n",
		"nested/p/p.go":                     "package p\n",
		"pnpm-workspace.yaml":               "packages:\n  - apps/web\n  - packages/*\nstoreDir: .pnpm-store\n",
		"apps/web/package.json":             `{"name":"@m/web","dependencies":{"@m/lib":"workspace:*","react":"19"}}`,
		"apps/web/src/a.test.ts":            "test('a', () => {})\n",
		"packages/lib/package.json":         `{"name":"@m/lib","devDependencies":{"@m/base":"workspace:*"}}`,
		"packages/lib/src/lib.test.ts":      "test('lib', () => {})\n",
		"packages/base/package.json":        `{"name":"@m/base"}`,
		"packages/base/src/base.ts":         "export const base = 1\n",
		"packages/other/package.json":       `{"name":"@m/other"}`,
		"packages/other/src/other.ts":       "export const other = 1\n",
		"scripts/stray.test.ts":             "test('stray', () => {})\n",
		"tools/runner/cli.mjs":              "import { a } from \"./lib/a.mjs\";\nexport { b } from \"../runner/b\";\nimport \"node:fs\";\nimport x from \"some-package\";\nconst r = requireFrom(\"./gen/index.cjs\");\nawait import(\"./late.mjs\");\n",
		"tools/runner/lib/a.mjs":            "export const a = 1;\n",
		"tools/runner/b.mjs":                "export const b = 2;\n",
		"tools/runner/gen/index.cjs":        "module.exports = require(\"./one.cjs\");\n",
		"tools/runner/gen/one.cjs":          "module.exports = 1;\n",
		"tools/runner/late.mjs":             "export {};\n",
		"tools/runner/unused.mjs":           "export {};\n",
		"tools/testservices/main.go":        "package main\n",
		"tools/harness_work_estimates.json": "{}\n",
	}
	write := func(relative, content string) {
		t.Helper()
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := func() []SourceInput {
		t.Helper()
		var inputs []SourceInput
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == ".cache" {
					return filepath.SkipDir // machine state, as Git ignores it
				}
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relative, _ := filepath.Rel(root, path)
			sum := sha256.Sum256(data)
			inputs = append(inputs, SourceInput{Path: filepath.ToSlash(relative), Kind: "file", Mode: "0644", ByteDigest: "sha256:" + hex.EncodeToString(sum[:])})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		slices.SortFunc(inputs, func(a, b SourceInput) int {
			if a.Path < b.Path {
				return -1
			}
			return 1
		})
		return inputs
	}
	for relative, content := range files {
		write(relative, content)
	}

	unitAware := CacheProfile{ProfileID: "test_rows", Policy: "content_addressed", DependencyStrategy: "unit_aware",
		InputRoots: []string{"a", "b", "c", "go.mod", "t"}, RunnerTypes: []string{RunnerGo},
		UnitInputs: &UnitInputs{Paths: []string{"configs", "contracts", "go.mod"}, Prefixes: []string{"tools/harness"},
			RunnerPackages: []string{"r"}, ServicePaths: []string{"tools/testservices"}}}
	frontend := CacheProfile{ProfileID: "vitest_rows", Policy: "content_addressed", DependencyStrategy: "unit_aware",
		InputRoots: []string{"apps", "packages", "scripts"}, RunnerTypes: []string{RunnerVitest},
		UnitInputs: &UnitInputs{Paths: []string{"pnpm-workspace.yaml"}, Prefixes: []string{}, RunnerEntries: []string{"tools/runner/cli.mjs"}}}
	broad := CacheProfile{ProfileID: "build", Policy: "content_addressed", DependencyStrategy: "broad", InputRoots: []string{"b", "go.mod"}, OperationIDs: []string{"harness.build"}}
	scan := CacheProfile{ProfileID: "scan", Policy: "content_addressed", DependencyStrategy: "broad", InputRoots: []string{"go.mod"}, OperationIDs: []string{"harness.scan"}, RequiresVulnerabilityDatabaseRevision: true}
	in := &Inputs{
		Cache: CacheRegistry{Profiles: []CacheProfile{unitAware, frontend, broad, scan}},
		Catalog: Catalog{
			Owners: []CatalogOwner{{OwnerID: "module.probe_owner", ManifestPath: "tools/test_families/owner.json"}},
			Rows: []Row{{RowID: "module.probe_owner.unit.a", OwnerID: "module.probe_owner", Runner: RunnerGo,
				Selector: Selector{Package: "./a", Tests: []string{"TestA"}}, FixtureInputRoots: []string{"fixtures/case"}},
				{RowID: "module.probe_owner.unit.nested", OwnerID: "module.probe_owner", Runner: RunnerGo,
					Selector: Selector{Package: "./nested/p", Tests: []string{"TestP"}}},
				{RowID: "web.probe.unit.web", OwnerID: "web.probe", Runner: RunnerVitest, Selector: Selector{File: "apps/web/src/a.test.ts", ProjectID: "browser-unit", Titles: []string{"a"}}},
				{RowID: "web.probe.unit.lib", OwnerID: "web.probe", Runner: RunnerVitest, Selector: Selector{File: "packages/lib/src/lib.test.ts", ProjectID: "node-unit", Titles: []string{"lib"}}},
				{RowID: "web.probe.unit.stray", OwnerID: "web.probe", Runner: RunnerVitest, Selector: Selector{File: "scripts/stray.test.ts", ProjectID: "node-unit", Titles: []string{"stray"}}}},
		},
		rowIndex: map[string]int{"module.probe_owner.unit.a": 0, "module.probe_owner.unit.nested": 1, "web.probe.unit.web": 2, "web.probe.unit.lib": 3, "web.probe.unit.stray": 4},
	}
	goUnit := &Unit{ID: "go:a", CachePolicy: "content_addressed", RowBindings: []RowBinding{{RowID: "module.probe_owner.unit.a", Runner: RunnerGo}},
		EvidenceOutputs: []string{"rows/module.probe_owner.unit.a.json", "unit-results/go-a.json"}, SemanticDigest: "sha256:" + digest("go:a")}
	artifact := []ArtifactOutput{{ArtifactType: "file", RelativePath: "build/out", DestinationClass: "repository_artifact", Mode: "0644", ProducerIdentity: "operation:harness.build"}}
	buildUnit := &Unit{ID: "operation:harness.build", CachePolicy: "content_addressed", ArtifactOutputs: artifact, EvidenceOutputs: []string{"unit-results/operation-harness.build.json"}}
	facts := func(units ...*Unit) ([]CacheFact, map[string][]int, error) {
		graph := &Graph{Units: map[string]*Unit{}}
		for _, unit := range units {
			graph.Units[unit.ID] = unit
		}
		return CacheFacts(root, in, graph, snapshot(), CacheContext{ToolchainDigest: "sha256:" + digest("toolchain"), Platform: "linux/amd64"}, false)
	}
	paths := func(source []SourceInput, indices []int) []string {
		var result []string
		for _, index := range indices {
			result = append(result, source[index].Path)
		}
		return result
	}

	t.Run("a Go unit's closure is its packages, their local imports and declared inputs", func(t *testing.T) {
		facts, inputs, err := facts(goUnit)
		if err != nil {
			t.Fatal(err)
		}
		fact := facts[0]
		if !fact.Eligible || *fact.DependencyStrategy != strategyGoPackages || *fact.ProfileID != "test_rows" {
			t.Fatalf("fact = %+v", fact)
		}
		got := paths(snapshot(), inputs[*fact.DependencyDigest])
		// Files the Go code names by a literal path join the closure, and a
		// named Node module brings its module closure; a named directory or
		// an absent file does not.
		want := []string{"a/a.go", "a/a_test.go", "a/testdata/input.txt", "b/b.go", "b/launch.go", "configs/app.json", "configs2/read.json", "contracts/probe-owner/case.json",
			"fixtures/case/input.json", "go.mod", "r/r.go", "s/s.go", "t/t.go", "tools/harness_helpers.json", "tools/launch/dep.mjs", "tools/launch/entry.mjs", "tools/test_families/owner.json"}
		if !slices.Equal(got, want) {
			t.Fatalf("closure = %v\nwant      %v", got, want)
		}
	})

	t.Run("a package in a nested module closes over that module's manifests", func(t *testing.T) {
		nested := &Unit{ID: "go:nested", CachePolicy: "content_addressed", RowBindings: []RowBinding{{RowID: "module.probe_owner.unit.nested", Runner: RunnerGo}},
			EvidenceOutputs: []string{"rows/module.probe_owner.unit.nested.json", "unit-results/go-nested.json"}, SemanticDigest: "sha256:" + digest("go:nested")}
		facts, inputs, err := facts(nested)
		if err != nil || *facts[0].DependencyStrategy != strategyGoPackages {
			t.Fatalf("fact = %+v, %v", facts, err)
		}
		got := paths(snapshot(), inputs[*facts[0].DependencyDigest])
		for _, member := range []string{"nested/go.mod", "nested/go.sum", "nested/p/p.go"} {
			if !slices.Contains(got, member) {
				t.Fatalf("closure %v lacks %s", got, member)
			}
		}
	})

	t.Run("a Vitest unit closes over its workspaces and their declared dependencies", func(t *testing.T) {
		vitest := func(row string) *Unit {
			return &Unit{ID: "vitest:" + row, CachePolicy: "content_addressed", RowBindings: []RowBinding{{RowID: row, Runner: RunnerVitest}},
				EvidenceOutputs: []string{"rows/" + row + ".json"}, SemanticDigest: "sha256:" + digest(row)}
		}
		for row, want := range map[string][]string{
			"web.probe.unit.web": {"apps/web/package.json", "apps/web/src/a.test.ts", "packages/base/package.json", "packages/base/src/base.ts",
				"packages/lib/package.json", "packages/lib/src/lib.test.ts", "pnpm-workspace.yaml", "tools/runner/b.mjs", "tools/runner/cli.mjs", "tools/runner/gen/index.cjs", "tools/runner/gen/one.cjs", "tools/runner/late.mjs", "tools/runner/lib/a.mjs"},
			"web.probe.unit.lib": {"packages/base/package.json", "packages/base/src/base.ts", "packages/lib/package.json", "packages/lib/src/lib.test.ts", "pnpm-workspace.yaml",
				"tools/runner/b.mjs", "tools/runner/cli.mjs", "tools/runner/gen/index.cjs", "tools/runner/gen/one.cjs", "tools/runner/late.mjs", "tools/runner/lib/a.mjs"},
		} {
			facts, inputs, err := facts(vitest(row))
			if err != nil || *facts[0].DependencyStrategy != strategyWorkspaces || *facts[0].ProfileID != "vitest_rows" {
				t.Fatalf("%s fact = %+v, %v", row, facts, err)
			}
			if got := paths(snapshot(), inputs[*facts[0].DependencyDigest]); !slices.Equal(got, want) {
				t.Fatalf("%s closure = %v\nwant %v", row, got, want)
			}
		}
		facts, _, err := facts(vitest("web.probe.unit.stray"))
		if err != nil || *facts[0].DependencyStrategy != strategyBroadFallback {
			t.Fatalf("a file outside every workspace: %+v, %v", facts, err)
		}
	})

	t.Run("a closure digest is the semantic digest of its document", func(t *testing.T) {
		source := snapshot()
		planner := &cachePlanner{source: source, entries: map[int][]byte{}}
		set := map[int]bool{0: true, 2: true}
		metadata := map[string]any{"packages": []string{"a"}}
		got := planner.digest(strategyGoPackages, metadata, set)
		want := semanticDigest(map[string]any{"strategy": strategyGoPackages, "metadata": metadata, "entries": []SourceInput{source[0], source[2]}}, "")
		if got.digest != want {
			t.Fatalf("digest %s, want %s", got.digest, want)
		}
	})

	t.Run("only closure changes change the keys", func(t *testing.T) {
		before, _, err := facts(goUnit)
		if err != nil {
			t.Fatal(err)
		}
		write("c/c.go", "package c\n\nconst C = 2\n")
		write("tools/other.json", "{\"changed\":true}\n")
		unrelated, _, err := facts(goUnit)
		if err != nil {
			t.Fatal(err)
		}
		write("b/b.go", "package b\n\nconst B = 2\n")
		changed, _, err := facts(goUnit)
		if err != nil {
			t.Fatal(err)
		}
		if *unrelated[0].InputDigest != *before[0].InputDigest || *changed[0].InputDigest == *before[0].InputDigest {
			t.Fatal("cache keys do not follow the closure")
		}
	})

	t.Run("scheduling fields never change a key", func(t *testing.T) {
		key := func(unit Unit) string {
			t.Helper()
			result, _, err := facts(&unit)
			if err != nil {
				t.Fatal(err)
			}
			return *result[0].InputDigest
		}
		base := *goUnit
		base.Command = Command{Executable: "go", Args: []string{"test", "-json", "-p=1", "-run", "^(?:TestA)$", "./a"}, Environment: map[string]string{"GOMAXPROCS": "1"}}
		scheduled := base
		scheduled.EstimatedWorkMS, scheduled.TimeoutMS = 90000, 600000
		scheduled.Claims = map[string]int64{"cpu": 4, "postgres": 1}
		scheduled.Residents = []Resident{{Key: "cluster"}}
		scheduled.SharedLocks, scheduled.ExclusiveLocks = []string{"host_activity"}, []string{"exclusive"}
		scheduled.SemanticDigest = "sha256:" + digest("another schedule")
		scheduled.Command = Command{Executable: "go", Args: []string{"test", "-json", "-p=4", "-run", "^(?:TestA)$", "./a"}, Environment: map[string]string{"GOMAXPROCS": "4"}}
		if key(scheduled) != key(base) {
			t.Fatal("a scheduling field changed a cache key")
		}
		selected := base
		selected.Command = Command{Executable: "go", Args: []string{"test", "-json", "-p=1", "-run", "^(?:TestB)$", "./a"}, Environment: map[string]string{"GOMAXPROCS": "1"}}
		if key(selected) == key(base) {
			t.Fatal("another test selection kept its cache key")
		}
	})

	t.Run("incomplete dependency knowledge selects the broad closure", func(t *testing.T) {
		write("b/b.go", "package b\n\nimport _ \"example.com/m/missing\"\n")
		defer write("b/b.go", "package b\n\nconst B = 1\n")
		facts, inputs, err := facts(goUnit)
		if err != nil {
			t.Fatal(err)
		}
		if *facts[0].DependencyStrategy != strategyBroadFallback || !slices.Contains(paths(snapshot(), inputs[*facts[0].DependencyDigest]), "c/c.go") {
			t.Fatalf("fact = %+v", facts[0])
		}
	})

	t.Run("a service-backed unit adds its service paths and migration template", func(t *testing.T) {
		in.MigrationDigest = "sha256:" + digest("migrations-1")
		in.Catalog.Rows[0].ServiceDependencies = []string{"postgres"}
		defer func() { in.Catalog.Rows[0].ServiceDependencies = nil }()
		first, inputs, err := facts(goUnit)
		if err != nil || !first[0].Eligible {
			t.Fatalf("fact = %+v, %v", first, err)
		}
		if got := paths(snapshot(), inputs[*first[0].DependencyDigest]); !slices.Contains(got, "tools/testservices/main.go") {
			t.Fatalf("closure %v lacks the service helper", got)
		}
		in.MigrationDigest = "sha256:" + digest("migrations-2")
		second, _, err := facts(goUnit)
		if err != nil || *second[0].InputDigest == *first[0].InputDigest {
			t.Fatal("a migration change kept a service-backed key", err)
		}
	})

	t.Run("a computed runner import outside the declared externals selects the broad closure", func(t *testing.T) {
		row := "web.probe.unit.lib"
		vitestUnit := &Unit{ID: "vitest:" + row, CachePolicy: "content_addressed", RowBindings: []RowBinding{{RowID: row, Runner: RunnerVitest}},
			EvidenceOutputs: []string{"rows/" + row + ".json"}, SemanticDigest: "sha256:" + digest(row)}
		write("tools/runner/lib/a.mjs", "export const a = await import(name);\n")
		defer write("tools/runner/lib/a.mjs", "export const a = 1;\n")
		computed, _, err := facts(vitestUnit, goUnit)
		if err != nil || *computed[1].DependencyStrategy != strategyBroadFallback || *computed[0].DependencyStrategy != strategyGoPackages {
			t.Fatalf("facts = %+v, %v", computed, err)
		}
		frontend.UnitInputs.RunnerExternalImports = []string{"tools/runner/lib/a.mjs"}
		in.Cache.Profiles[1] = frontend
		defer func() { frontend.UnitInputs.RunnerExternalImports = nil; in.Cache.Profiles[1] = frontend }()
		declared, _, err := facts(vitestUnit)
		if err != nil || *declared[0].DependencyStrategy != strategyWorkspaces {
			t.Fatalf("declared external: %+v, %v", declared, err)
		}
	})

	t.Run("a Go runner package change changes Go keys only", func(t *testing.T) {
		row := "web.probe.unit.lib"
		vitestUnit := &Unit{ID: "vitest:" + row, CachePolicy: "content_addressed", RowBindings: []RowBinding{{RowID: row, Runner: RunnerVitest}},
			EvidenceOutputs: []string{"rows/" + row + ".json"}, SemanticDigest: "sha256:" + digest(row)}
		before, _, err := facts(goUnit, vitestUnit)
		if err != nil {
			t.Fatal(err)
		}
		write("s/s.go", "package s\n\nconst S = 1\n")
		defer write("s/s.go", "package s\n")
		after, _, err := facts(goUnit, vitestUnit)
		if err != nil || *after[0].InputDigest == *before[0].InputDigest || *after[1].InputDigest != *before[1].InputDigest {
			t.Fatalf("before %+v after %+v: %v", before, after, err)
		}
	})

	t.Run("file scans are kept by content digest", func(t *testing.T) {
		row := "web.probe.unit.lib"
		vitestUnit := &Unit{ID: "vitest:" + row, CachePolicy: "content_addressed", RowBindings: []RowBinding{{RowID: row, Runner: RunnerVitest}},
			EvidenceOutputs: []string{"rows/" + row + ".json"}, SemanticDigest: "sha256:" + digest(row)}
		memoized := func() ([]CacheFact, map[string][]int) {
			t.Helper()
			result, inputs, err := CacheFacts(root, in, &Graph{Units: map[string]*Unit{vitestUnit.ID: vitestUnit}}, snapshot(),
				CacheContext{ToolchainDigest: "sha256:" + digest("toolchain"), Platform: "linux/amd64"}, true)
			if err != nil {
				t.Fatal(err)
			}
			return result, inputs
		}
		plain, _, err := facts(vitestUnit)
		if err != nil {
			t.Fatal(err)
		}
		if first, _ := memoized(); *first[0].InputDigest != *plain[0].InputDigest {
			t.Fatal("a memoized scan changed a key")
		}
		// The memo holds the runner entry's scan; a recorded scan is used for
		// the same bytes.
		source := snapshot()
		index := slices.IndexFunc(source, func(input SourceInput) bool { return input.Path == "tools/runner/cli.mjs" })
		scans := loadSourceScans(root, true)
		if _, ok := scans.Modules[source[index].ByteDigest]; !ok {
			t.Fatal("the runner entry's scan was not kept")
		}
		scans.Modules[source[index].ByteDigest], scans.changed = moduleScan{Specifiers: []string{"./lib/a.mjs"}}, true
		scans.save(root, source)
		reused, inputs := memoized()
		if got := paths(snapshot(), inputs[*reused[0].DependencyDigest]); slices.Contains(got, "tools/runner/b.mjs") {
			t.Fatalf("the recorded scan was not used: %v", got)
		}
		// Other bytes are scanned again, and another scanner discards the memo.
		write("tools/runner/cli.mjs", files["tools/runner/cli.mjs"]+"// changed\n")
		defer write("tools/runner/cli.mjs", files["tools/runner/cli.mjs"])
		rescanned, inputs := memoized()
		if got := paths(snapshot(), inputs[*rescanned[0].DependencyDigest]); !slices.Contains(got, "tools/runner/b.mjs") {
			t.Fatalf("changed bytes reused a scan: %v", got)
		}
		scans = loadSourceScans(root, true)
		scans.Scanner, scans.changed = "another", true
		scans.save(root, snapshot())
		if len(loadSourceScans(root, true).Modules) != 0 {
			t.Fatal("a memo from another scanner was used")
		}
	})

	t.Run("eligibility follows policy, outputs and freshness", func(t *testing.T) {
		policyNone := &Unit{ID: "operation:harness.none", CachePolicy: "none"}
		noOutputs := &Unit{ID: "operation:harness.build", CachePolicy: "content_addressed", EvidenceOutputs: []string{"unit-results/x.json"}}
		freshOutputs := &Unit{ID: "operation:harness.build", CachePolicy: "content_addressed", ArtifactOutputs: artifact, EvidenceOutputs: []string{"report/summary.json"}}
		scanUnit := &Unit{ID: "operation:harness.scan", CachePolicy: "content_addressed", ArtifactOutputs: []ArtifactOutput{{ProducerIdentity: "operation:harness.scan"}}}
		for want, unit := range map[string]*Unit{"policy_none": policyNone, "reusable_outputs_incomplete": noOutputs, "fresh_outputs_incomplete": freshOutputs, "freshness_unknown": scanUnit} {
			facts, _, err := facts(unit)
			if err != nil || facts[0].Eligible || *facts[0].Reason != want || facts[0].InputDigest != nil {
				t.Errorf("%s: %+v, %v", want, facts, err)
			}
		}
		built, inputs, err := facts(buildUnit)
		if err != nil || !built[0].Eligible || *built[0].DependencyStrategy != strategyBroad ||
			!slices.Equal(paths(snapshot(), inputs[*built[0].DependencyDigest]), []string{"b/b.go", "b/launch.go", "go.mod"}) {
			t.Fatalf("broad operation fact = %+v, %v", built, err)
		}
		if _, _, err := facts(&Unit{ID: "operation:harness.unbound", CachePolicy: "same_run", ArtifactOutputs: artifact}); !errors.Is(err, ErrInput) {
			t.Fatalf("a cacheable unit without a profile: %v", err)
		}
	})
}

func TestRowCachePolicy(t *testing.T) {
	for _, tc := range []struct {
		name string
		row  Row
		want string
	}{
		{"a fixture-free Go row", Row{Runner: RunnerGo, EvidenceClass: "unit", FixtureCapability: "none"}, "content_addressed"},
		{"a Vitest row", Row{Runner: RunnerVitest, EvidenceClass: "unit", FixtureCapability: "none"}, "content_addressed"},
		{"a Go row on its own generation", Row{Runner: RunnerGo, EvidenceClass: "integration", FixtureCapability: "postgres_dedicated", ServiceDependencies: []string{"object_store", "postgres"}}, "content_addressed"},
		{"a Go row in a shared-generation transaction", Row{Runner: RunnerGo, EvidenceClass: "integration", FixtureCapability: "postgres_transaction", ServiceDependencies: []string{"postgres"}}, "content_addressed"},
		{"a Go row on its own object-store namespace", Row{Runner: RunnerGo, EvidenceClass: "integration", FixtureCapability: "object_store_namespace", ServiceDependencies: []string{"object_store"}}, "content_addressed"},
		{"a migration row", Row{Runner: RunnerGo, EvidenceClass: "integration", FixtureCapability: "postgres_migration", ServiceDependencies: []string{"postgres"}}, "none"},
		{"a process-exclusive row", Row{Runner: RunnerGo, EvidenceClass: "integration", FixtureCapability: "postgres_dedicated", ServiceDependencies: []string{"postgres"}, ProcessIsolation: "exclusive"}, "none"},
		{"a security row", Row{Runner: RunnerGo, EvidenceClass: "security", FixtureCapability: "postgres_dedicated", ServiceDependencies: []string{"postgres"}}, "none"},
		{"a measurement row", Row{Runner: RunnerGo, EvidenceClass: "measurement", FixtureCapability: "none"}, "none"},
		{"a Playwright row", Row{Runner: RunnerPlaywright, EvidenceClass: "browser", FixtureCapability: "browser_stack"}, "none"},
		{"a service-backed Vitest row", Row{Runner: RunnerVitest, EvidenceClass: "integration", FixtureCapability: "postgres_dedicated", ServiceDependencies: []string{"postgres"}}, "none"},
	} {
		if got := rowCachePolicy(tc.row); got != tc.want {
			t.Errorf("%s: %s, want %s", tc.name, got, tc.want)
		}
	}
}

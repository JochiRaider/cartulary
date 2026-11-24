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
		"go.mod":                          "module example.com/m\n\ngo 1.27\n",
		"a/a.go":                          "package a\n\nimport (\n\t\"fmt\"\n\t\"example.com/m/b\"\n)\n\nvar _ = fmt.Sprint(b.B)\n",
		"a/a_test.go":                     "//go:build never\n\npackage a\n\nimport _ \"example.com/m/t\"\n",
		"a/testdata/input.txt":            "data\n",
		"b/b.go":                          "package b\n\nconst B = 1\n",
		"t/t.go":                          "package t\n",
		"c/c.go":                          "package c\n",
		"configs/app.json":                "{}\n",
		"tools/harness_helpers.json":      "{}\n",
		"tools/other.json":                "{}\n",
		"tools/test_families/owner.json":  "{}\n",
		"contracts/probe-owner/case.json": "{}\n",
		"fixtures/case/input.json":        "{}\n",
		"fixtures/other/input.json":       "{}\n",
		"nested/go.mod":                   "module example.com/m/nested\n",
		"nested/go.sum":                   "\n",
		"nested/p/p.go":                   "package p\n",
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
			if err != nil || d.IsDir() {
				return err
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
		UnitInputs: &UnitInputs{Paths: []string{"configs", "contracts", "go.mod"}, Prefixes: []string{"tools/harness"}}}
	broad := CacheProfile{ProfileID: "build", Policy: "content_addressed", DependencyStrategy: "broad", InputRoots: []string{"b", "go.mod"}, OperationIDs: []string{"harness.build"}}
	scan := CacheProfile{ProfileID: "scan", Policy: "content_addressed", DependencyStrategy: "broad", InputRoots: []string{"go.mod"}, OperationIDs: []string{"harness.scan"}, RequiresVulnerabilityDatabaseRevision: true}
	in := &Inputs{
		Cache: CacheRegistry{Profiles: []CacheProfile{unitAware, broad, scan}},
		Catalog: Catalog{
			Owners: []CatalogOwner{{OwnerID: "module.probe_owner", ManifestPath: "tools/test_families/owner.json"}},
			Rows: []Row{{RowID: "module.probe_owner.unit.a", OwnerID: "module.probe_owner", Runner: RunnerGo,
				Selector: Selector{Package: "./a", Tests: []string{"TestA"}}, FixtureInputRoots: []string{"fixtures/case"}},
				{RowID: "module.probe_owner.unit.nested", OwnerID: "module.probe_owner", Runner: RunnerGo,
					Selector: Selector{Package: "./nested/p", Tests: []string{"TestP"}}}},
		},
		rowIndex: map[string]int{"module.probe_owner.unit.a": 0, "module.probe_owner.unit.nested": 1},
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
		return CacheFacts(root, in, graph, snapshot(), CacheContext{ToolchainDigest: "sha256:" + digest("toolchain"), Platform: "linux/amd64"})
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
		want := []string{"a/a.go", "a/a_test.go", "a/testdata/input.txt", "b/b.go", "configs/app.json", "contracts/probe-owner/case.json",
			"fixtures/case/input.json", "go.mod", "t/t.go", "tools/harness_helpers.json", "tools/test_families/owner.json"}
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
			!slices.Equal(paths(snapshot(), inputs[*built[0].DependencyDigest]), []string{"b/b.go", "go.mod"}) {
			t.Fatalf("broad operation fact = %+v, %v", built, err)
		}
		if _, _, err := facts(&Unit{ID: "operation:harness.unbound", CachePolicy: "same_run", ArtifactOutputs: artifact}); !errors.Is(err, ErrInput) {
			t.Fatalf("a cacheable unit without a profile: %v", err)
		}
	})
}

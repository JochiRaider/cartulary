package packformat

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDependencyVerificationGraphAndPrecedence_Unit(t *testing.T) {
	ctx := context.Background()
	edge := func(version string) Dependency {
		return Dependency{"enrichment.lolbas", version, strings.Repeat("a", 64)}
	}
	root := Manifest{Key: "enrichment.lolbas", Version: "root"}
	graph := map[string]Manifest{}
	resolve := func(_ context.Context, d Dependency) (Manifest, bool, error) {
		m, ok := graph[d.Version]
		return m, ok, nil
	}
	check := func(want string, count int) *ValidationSummary {
		t.Helper()
		err := VerifyDependencies(ctx, root, resolve, nil)
		if want == "" {
			if err != nil {
				t.Fatal(err)
			}
			return nil
		}
		var failure *Failure
		if !errors.As(err, &failure) || failure.CheckID != want || failure.Summary == nil || failure.Summary.Total != count {
			t.Fatalf("check=%s count=%d: %#v, %v", want, count, failure, err)
		}
		return failure.Summary
	}
	root.Dependencies = []Dependency{edge("absent-a"), edge("absent-b")}
	root.Conflicts = []Conflict{{Key: root.Key}}
	check("dependencies", 2)
	root.Dependencies = []Dependency{edge("a")}
	graph["a"] = Manifest{Key: root.Key, Version: "a", Dependencies: []Dependency{edge("b"), edge("c")}}
	graph["b"] = Manifest{Key: root.Key, Version: "b", Dependencies: []Dependency{edge("a")}}
	graph["c"] = Manifest{Key: root.Key, Version: "c", Dependencies: []Dependency{edge("b")}}
	first := check("dependency_cycle", 3)
	graph["a"] = Manifest{Key: root.Key, Version: "a", Dependencies: []Dependency{edge("c"), edge("b")}}
	if other := check("dependency_cycle", 3); !reflect.DeepEqual(first, other) {
		t.Fatal("cycle diagnostics depend on traversal order")
	}
	graph["b"] = Manifest{Key: root.Key, Version: "b"}
	graph["c"] = Manifest{Key: root.Key, Version: "c"}
	check("conflicts", 1)
	otherVersion := "unrelated"
	root.Conflicts = []Conflict{{Key: root.Key, Version: &otherVersion}}
	check("", 0)
	root.Conflicts = []Conflict{{Key: "type_registry.host"}, {Key: "type_registry.evidence"}}
	check("conflicts", 2)
	root.Conflicts = nil
	graph["a"] = Manifest{Key: root.Key, Version: "a", Dependencies: []Dependency{edge("missing-transitive")}}
	check("dependencies", 1)
	if err := VerifyDependencies(ctx, root, nil, nil); err == nil {
		t.Fatal("missing resolver skipped dependency checks")
	}
	operational := errors.New("injected read failure")
	if err := VerifyDependencies(ctx, root, func(context.Context, Dependency) (Manifest, bool, error) { return Manifest{}, false, operational }, nil); !errors.Is(err, operational) {
		t.Fatal("operational error became a content verdict", err)
	}
}

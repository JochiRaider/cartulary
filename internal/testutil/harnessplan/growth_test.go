package harnessplan

import (
	"bytes"
	"encoding/json"
	"testing"
)

// TestPlanGrowth bounds the serialized size of every unit of the ci graph
// under a fixed planning profile, so graph growth stays structural: each
// unit has a contract allowance, plus linear allowances for its bound rows,
// their selected Go tests and its dependency edges. Two compiles of the same
// inputs are byte-identical.
func TestPlanGrowth(t *testing.T) {
	inputs, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	selection, err := inputs.aggregateSelection("ci")
	if err != nil {
		t.Fatal(err)
	}
	options := Options{CPUTokens: 4, GoVersion: "1.27.4", BrowserLanes: 4}
	const (
		baseBytes          = 4096
		bytesPerRowBinding = 384
		// A Go unit names each selected test twice: in its row's binding
		// and in its command's exact -run pattern.
		bytesPerSelectedTest = 96
		bytesPerDependency   = 192
		// A finalizer carries its group, target and unit mappings beside
		// its edges, each naming a digest-identified cohort.
		finalizerBytesPerDependency = 448
	)
	compile := func() []byte {
		graph, err := Compile(inputs, selection, options)
		if err != nil {
			t.Fatal(err)
		}
		document, err := json.Marshal(graph)
		if err != nil {
			t.Fatal(err)
		}
		for _, unit := range graph.Sorted() {
			size, err := json.Marshal(unit)
			if err != nil {
				t.Fatal(err)
			}
			perDependency := bytesPerDependency
			if unit.Kind == "finalizer" {
				perDependency = finalizerBytesPerDependency
			}
			tests := 0
			for _, binding := range unit.RowBindings {
				tests += len(binding.Selectors)
			}
			if budget := baseBytes + len(unit.RowBindings)*bytesPerRowBinding + tests*bytesPerSelectedTest + len(unit.Needs)*perDependency; len(size) > budget {
				t.Errorf("%s unit %s uses %d bytes; its structural budget is %d", unit.Kind, unit.ID, len(size), budget)
			}
		}
		return document
	}
	if !bytes.Equal(compile(), compile()) {
		t.Fatal("two compiles of the same inputs are not byte-identical")
	}
	for _, aggregate := range []string{"test-fast", "check"} {
		selection, err := inputs.aggregateSelection(aggregate)
		if err != nil {
			t.Fatal(err)
		}
		first, err := Compile(inputs, selection, options)
		if err != nil {
			t.Fatal(err)
		}
		second, err := Compile(inputs, selection, options)
		if err != nil {
			t.Fatal(err)
		}
		if first.Digest != second.Digest {
			t.Fatalf("two compiles of %s differ", aggregate)
		}
	}
}

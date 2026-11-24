package harnessplan

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

func TestPlanAssembly(t *testing.T) {
	inputs, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := Capture(Host{CPUs: 24, MemoryBytes: 32 << 30, ProcessLimit: 127192, WritableRoot: true}, inputs.Registry, nil)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := Compile(inputs, Selection{Target: "harness.go_vulncheck"}, Options{CPUTokens: snapshot.CPUTokens, GoVersion: "1.27.4"})
	if err != nil {
		t.Fatal(err)
	}
	text := func(value string) *string { return &value }
	invocation := func() Invocation {
		var cache []CacheFact
		for _, unit := range graph.Sorted() {
			cache = append(cache, CacheFact{UnitID: unit.ID, Reason: text("mode_off")})
		}
		return Invocation{
			RunID: "plan-test", CommandID: "cartulary.harness.command.go_vulncheck.v1", Target: "go-vulncheck",
			DeclaredInputs:   map[string]string{"BROWSER_LANES": "4"},
			InputEnvironment: InputEnvironment{Clear: []string{"BROWSER_LANES", makeInputSources}, Forward: map[string]string{}},
			BrowserLaneLimit: 4, Snapshot: snapshot, GoWorkerBudget: 8,
			SourceInputs: []SourceInput{
				{Path: "go.mod", Kind: "file", Mode: "0644", ByteDigest: "sha256:" + digest("go.mod")},
				{Path: "internal/a.go", Kind: "file", Mode: "0644", ByteDigest: "sha256:" + digest("a")},
			},
			SourceCommit: "0123456789abcdef0123456789abcdef01234567", SourceState: "clean",
			ToolchainDigest: "sha256:" + digest("toolchain"), CacheMode: "off",
			InstrumentationMode: "off", InstrumentationPolicy: []byte(`{"schema_id":"policy"}`),
			Host:     HostProfile{Platform: "linux", Architecture: "amd64", KernelRelease: "6.6.0", CPUModels: []string{"test cpu"}},
			Bindings: map[string]string{"CARTULARY_TEST_SERVICES_MODE": "owned"},
			Cache:    cache, StartedAt: time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC),
		}
	}

	t.Run("a plan satisfies its schema and binds the graph", func(t *testing.T) {
		plan, err := Assemble(inputs, graph, invocation())
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(plan)
		if err != nil {
			t.Fatal(err)
		}
		document, err := harnessschema.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		if err := harnessschema.Validate(PlanSchema, document); err != nil {
			t.Fatalf("plan does not satisfy %s: %v", PlanSchema, err)
		}
		if plan.Manifest.GraphDigest != graph.Digest || plan.Manifest.SystemDigest != snapshot.SnapshotDigest {
			t.Fatal("manifest does not bind the graph and snapshot")
		}
		again, err := Assemble(inputs, graph, invocation())
		if err != nil || again.Manifest.SourceDigest != plan.Manifest.SourceDigest || again.Manifest.Instrumentation != plan.Manifest.Instrumentation {
			t.Fatal("plan digests are not deterministic")
		}
	})

	t.Run("invalid invocations are rejected", func(t *testing.T) {
		for name, mutate := range map[string]func(*Invocation){
			"worker budget above CPU tokens":  func(i *Invocation) { i.GoWorkerBudget = i.Snapshot.CPUTokens + 1 },
			"lane limit disagrees with input": func(i *Invocation) { i.DeclaredInputs["BROWSER_LANES"] = "2" },
			"forwarded input not cleared":     func(i *Invocation) { i.InputEnvironment.Forward["OTHER"] = "x" },
			"Make input sources forwarded":    func(i *Invocation) { i.InputEnvironment.Forward[makeInputSources] = "x" },
			"unknown cache mode":              func(i *Invocation) { i.CacheMode = "warm" },
			"cache facts out of order":        func(i *Invocation) { i.Cache[0], i.Cache[len(i.Cache)-1] = i.Cache[len(i.Cache)-1], i.Cache[0] },
			"cache fact missing":              func(i *Invocation) { i.Cache = i.Cache[1:] },
			"incomplete eligible fact":        func(i *Invocation) { i.Cache[0] = CacheFact{UnitID: i.Cache[0].UnitID, Eligible: true} },
			"cache inputs with cache off":     func(i *Invocation) { i.CacheInputs = map[string][]int{"sha256:" + digest("x"): {0}} },
			"unsorted source inputs":          func(i *Invocation) { i.SourceInputs[0], i.SourceInputs[1] = i.SourceInputs[1], i.SourceInputs[0] },
			"escaping source path":            func(i *Invocation) { i.SourceInputs[1].Path = "../outside.go" },
			"invalid instrumentation policy":  func(i *Invocation) { i.InstrumentationPolicy = []byte("{") },
			"host too small for the graph": func(i *Invocation) {
				small, err := Capture(Host{CPUs: 2, MemoryBytes: 64 << 20, WritableRoot: true}, inputs.Registry, nil)
				if err != nil {
					t.Fatal(err)
				}
				i.Snapshot, i.GoWorkerBudget = small, 1
			},
		} {
			candidate := invocation()
			mutate(&candidate)
			_, err := Assemble(inputs, graph, candidate)
			if !errors.Is(err, ErrPlan) {
				t.Errorf("%s: Assemble = %v, want ErrPlan", name, err)
			}
			t.Logf("%s: %v", name, err)
		}
	})
}

package reporting

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
	"github.com/google/uuid"
	"testing"
	"time"
)

func testFrozenGraphSources(t *testing.T) {
	ref := testSourceProjectionRef("gv-frozen", "gpres_frozen")
	result := reportingGraphSelectionFixture()
	result.Binding = ref.binding()
	result.PublishedAt = time.Now().UTC()
	input := []resolvedGraphResult{{Ref: ref, Result: result, LabelCandidates: graphLabelCandidatesForResult(result)}}
	frozen, err := freezeGraphSources(input)
	if err != nil || validateFrozenGraphSources(frozen, []sourceProjectionRef{ref}) != nil {
		t.Fatal("owner graph contribution did not freeze", err)
	}
	digest, err := graphSourcesDigest(frozen)
	if err != nil {
		t.Fatal(err)
	}
	refs, _ := json.Marshal([]sourceProjectionRef{ref})
	payload := renderInput{GraphSources: frozen, GraphProjectionRefs: refs}
	if err := validateAdmittedGraphSources(payload, digest); err != nil {
		t.Fatal(err)
	}
	if !frozen.Results[0].Result.PublishedAt.IsZero() || input[0].Result.PublishedAt.IsZero() {
		t.Fatal("operational publication time entered admission or caller was changed")
	}
	input[0].LabelCandidates.VertexLabelCandidates[0].Endpoint.StringValue = "198.51.100.1"
	input[0].Result.Vertices[0].VertexID = "substituted"
	if frozen.Results[0].Result.Vertices[0].VertexID != "v1" || frozen.Results[0].LabelCandidates.VertexLabelCandidates[0].Endpoint.StringValue == "198.51.100.1" {
		t.Fatal("frozen source aliases caller memory")
	}
	t.Run("admission owns inputs after source becomes unavailable", func(t *testing.T) {
		contribution := frozen.Results[0]
		provider := &fakeGraphSourceProvider{ownerID: ref.SourceOwnerID, result: graphsourcecontract.Result{Projection: contribution.Result, LabelCandidates: contribution.LabelCandidates}}
		registry, err := NewGraphSourceRegistry(provider)
		if err != nil {
			t.Fatal(err)
		}
		admitted, err := registry.FreezeTx(t.Context(), borrowedReportingTx{}, uuid.New(), []sourceProjectionRef{ref})
		if err != nil {
			t.Fatal(err)
		}
		admittedSHA, err := graphSourcesDigest(admitted)
		if err != nil {
			t.Fatal(err)
		}
		provider.unavailable = true
		provider.result = graphsourcecontract.Result{}
		calls := provider.calls
		if err := validateAdmittedGraphSources(renderInput{GraphSources: admitted, GraphProjectionRefs: refs}, admittedSHA); err != nil {
			t.Fatal(err)
		}
		if provider.calls != calls {
			t.Fatal("retained input invoked unavailable owner")
		}
		if _, err := registry.FreezeTx(t.Context(), borrowedReportingTx{}, uuid.New(), []sourceProjectionRef{ref}); err == nil {
			t.Fatal("new admission accepted unavailable source")
		}
		if provider.calls != calls {
			t.Fatal("new admission invoked unavailable owner")
		}
	})
	for name, mutate := range map[string]func(*frozenGraphSources){
		"retired schema":   func(f *frozenGraphSources) { f.SchemaID = "retired" },
		"missing result":   func(f *frozenGraphSources) { f.Results = []resolvedGraphResult{} },
		"duplicate result": func(f *frozenGraphSources) { f.Results = append(f.Results, f.Results[0]) },
		"binding":          func(f *frozenGraphSources) { f.Results[0].Result.Binding.GraphViewID = "other" },
		"publication time": func(f *frozenGraphSources) { f.Results[0].Result.PublishedAt = time.Now() },
		"label roster":     func(f *frozenGraphSources) { f.Results[0].LabelCandidates.VertexLabelCandidates = nil },
		"unadmitted value": func(f *frozenGraphSources) {
			f.Results[0].LabelCandidates.VertexLabelCandidates[0].Endpoint.StringValue = "198.51.100.2"
		},
		"unadmitted topology": func(f *frozenGraphSources) { f.Results[0].Result.Edges[0].DstVertexID = "v3" },
	} {
		t.Run(name, func(t *testing.T) {
			changed, err := freezeGraphSources(frozen.Results)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&changed)
			candidate := payload
			candidate.GraphSources = changed
			if validateAdmittedGraphSources(candidate, digest) == nil {
				t.Fatal("changed graph input bypassed independent admission digest")
			}
		})
	}
	if validateAdmittedGraphSources(payload, "") == nil {
		t.Fatal("missing independent digest admitted")
	}
}

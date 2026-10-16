package reporting

import (
	"context"
	"encoding/json"
	"testing"
)

func assertFrozenMermaidSourceBinding(t *testing.T, payload renderInput) {
	t.Helper()
	ref := testSourceProjectionRef("gv-binding", "gpres_binding")
	result := reportingGraphSelectionFixture()
	result.Binding = ref.binding()
	var err error
	payload.GraphSources, err = freezeGraphSources([]resolvedGraphResult{{Ref: ref, Result: result, LabelCandidates: graphLabelCandidatesForResult(result)}})
	if err != nil {
		t.Fatal(err)
	}
	payload.GraphProjectionRefs, _ = json.Marshal([]sourceProjectionRef{ref})
	payload.OutputKind = OutputKindMermaid
	payload.OutputOptions = fixtureOutputOptions(OutputKindMermaid, payload.ReleaseScope)
	raw, _ := json.Marshal(map[string]any{
		"schema_id": "cartulary.report_composition.v1", "deck_ops": []any{}, "authored_texts": []any{},
		"diagram_decls": []any{map[string]any{
			"decl_id": "retained-graph", "replaces_decl_id": nil, "diagram_kind": "flowchart", "diagram_source_kind": "graph",
			"source_graph_view_id": ref.GraphViewID, "layout_mode": "auto", "layout": nil, "label_overrides": []any{},
			"selection_rule": diagramSelectionRule{SchemaID: "diagram_selection_rule.v1", RuleID: "all_with_bounds", VertexRefs: []string{}, EdgeRefs: []string{}, SeedVertexRefs: []string{}, EdgeKindFilter: []string{}, TimelineRecordIDs: []string{}, OverflowPolicy: "fail"},
		}},
	})
	payload.CompositionJSON = sealedReportingCompositionFixture(t, raw, payload.IncidentID, payload.TemplateContract)
	rendered, _, err := fixtureRenderer().Render(context.Background(), payload)
	if err != nil || validateRenderCandidateBinding(payload, rendered) != nil {
		t.Fatal("graph presentation did not bind to frozen input", err)
	}
	for name, mutate := range map[string]func(*RenderedRelease){
		"changed diagram": func(r *RenderedRelease) {
			for i := range r.RenderBundle.Files {
				f := &r.RenderBundle.Files[i]
				if f.Role == renderBundleRoleDiagramModel {
					var diagram reportingDiagramModel
					if err := json.Unmarshal(f.Bytes, &diagram); err != nil {
						t.Fatal(err)
					}
					diagram.Nodes[0].Label = "Substituted"
					f.Bytes, _ = canonicalJSON(diagram)
				}
			}
		},
		"changed source": func(r *RenderedRelease) {
			for i := range r.RenderBundle.Files {
				f := &r.RenderBundle.Files[i]
				if f.Role == renderBundleRoleSourceMermaid {
					f.Bytes = []byte("flowchart LR\n n0001[\"Substituted\"]\n")
				}
			}
		},
		"missing source": func(r *RenderedRelease) {
			for i, f := range r.RenderBundle.Files {
				if f.Role == renderBundleRoleDiagramModel {
					r.RenderBundle.Files = append(r.RenderBundle.Files[:i], r.RenderBundle.Files[i+1:]...)
					return
				}
			}
		},
		"unrequested SVG": func(r *RenderedRelease) {
			r.RenderBundle.Files = append(r.RenderBundle.Files, newBundleFile("diagrams/unrequested.svg", renderBundleRoleRenderedSVG, "image/svg+xml", []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`), true))
		},
		"missing requested SVG": func(r *RenderedRelease) {
			for i, f := range r.RenderBundle.Files {
				if f.Role == renderBundleRoleRenderedSVG {
					r.RenderBundle.Files = append(r.RenderBundle.Files[:i], r.RenderBundle.Files[i+1:]...)
					rebindFixtureBrowserUse(t, r, false)
					return
				}
			}
		},
	} {
		t.Run("Mermaid "+name, func(t *testing.T) {
			candidate := rendered
			candidate.RenderBundle.Files = append([]RenderBundleFile{}, rendered.RenderBundle.Files...)
			mutate(&candidate)
			rehashPresentationFixture(t, &candidate)
			if err := validateShareableBundle(candidate.RenderBundle); err != nil {
				t.Fatal("fixture must be internally consistent", err)
			}
			if validateRenderCandidateBinding(payload, candidate) == nil {
				t.Fatal("coherently rehashed diagram bypassed frozen source")
			}
		})
	}
}

package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func testGraphProfileRedaction(t *testing.T) {
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/notes/n", SourceFamily: "note", SourceRecordID: "n", SourceRecordType: "artifact", ContentClass: ContentClassWorkingMaterial, DisclosurePartitionRefs: []string{"public"}, Value: "Narrative"}})
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	ref := testSourceProjectionRef("gv-render", "gpres_"+strings.Repeat("d", 64))
	result := reportingGraphSelectionFixture()
	result.Binding = ref.binding()
	resolved := resolvedGraphResult{Ref: ref, Result: result, LabelCandidates: graphLabelCandidatesForResult(result)}
	rule := diagramSelectionRule{SchemaID: "diagram_selection_rule.v1", RuleID: "explicit_refs", VertexRefs: []string{"v1", "v2"}, EdgeRefs: []string{"e12"}, SeedVertexRefs: []string{}, EdgeKindFilter: []string{}, TimelineRecordIDs: []string{}, OverflowPolicy: "fail"}
	contract.Sections[0].DiagramDecls = []TemplateDiagramDeclaration{{SchemaID: "template_diagram_decl.v1", DeclID: "graph", DiagramKind: "flowchart", DiagramSourceKind: "graph", SourceGraphViewID: &ref.GraphViewID, SelectionRule: rule, Required: true}}
	refs, _ := json.Marshal([]sourceProjectionRef{ref})
	encoded, _ := canonicalJSON(model)
	request := CreateReleaseRequest{ReleaseScope: ReleaseScopeInternalReview, OutputKind: OutputKindSlidev, OutputOptions: fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalReview), RedactionProfileID: InternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion, GraphProjectionRefs: refs}
	input := renderCandidateFixture(request, contract, model, reportingObjectDigest(ExportModelSchemaID, encoded), resolved)
	endpoint := "$.diagrams['graph'].nodes['v1'].label_components['endpoint_value']"
	protocol := "$.diagrams['graph'].edges['e12'].label_components['protocol']"
	port := "$.diagrams['graph'].edges['e12'].label_components['destination_port']"
	masked, stubbed := "hidden endpoint", "protocol withheld"
	input.RedactionProfile.Rules = []RedactionRule{
		{RuleID: "endpoint_mask", Path: &endpoint, Action: RedactionActionSpec{Type: ActionMask, ReplacementText: &masked}},
		{RuleID: "protocol_stub", Path: &protocol, Action: RedactionActionSpec{Type: ActionStub, StubText: &stubbed}},
		{RuleID: "port_drop", Path: &port, Action: RedactionActionSpec{Type: ActionDrop}},
	}
	input.RedactionProfileSHA256, _ = ValidateRedactionProfile(input.RedactionProfile)
	before, _ := canonicalJSON(input)
	redacted, err := redactRenderInput(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if len(redacted.Model.Diagrams) != 1 {
		t.Fatal("missing graph")
	}
	diagram := redacted.Model.Diagrams[0]
	if diagram.Nodes[0].Label != masked || diagram.Nodes[1].Label != "192.0.2.2" || diagram.Edges[0].Label != "protocol="+stubbed {
		t.Fatal("graph labels bypassed selected actions")
	}
	if diagram.Nodes[0].NodeID != "v1" || diagram.Edges[0].EdgeID != "e12" || diagram.Edges[0].From != "v1" || diagram.Edges[0].To != "v2" {
		t.Fatal("canonical topology used renderer-local ordinals")
	}
	for _, prohibited := range []string{"192.0.2.1", "192.0.2.3", "port=443"} {
		if bytes.Contains(redacted.ModelJSON, []byte(prohibited)) {
			t.Fatal("unretained graph value entered canonical model")
		}
	}
	decisions := map[string]RedactionManifestEntry{}
	for _, entry := range redacted.Manifest.Entries {
		decisions[entry.Path] = entry
	}
	if decisions[endpoint].RuleID != "endpoint_mask" || decisions[port].Outcome != "dropped" || decisions[protocol].Outcome != "stubbed" {
		t.Fatal("missing graph rule evidence")
	}
	for _, prohibited := range []string{"192.0.2.1", masked, stubbed, "network-flow-source-object", "/graph/vertices/"} {
		if bytes.Contains(redacted.ManifestJSON, []byte(prohibited)) {
			t.Fatal("graph evidence leaked source or replacement content")
		}
	}
	prepared, err := prepareDiagrams(contract, redacted.Model, OutputKindSlidev, mustFixtureRenderLimits(t, contract))
	if err != nil || len(prepared.Diagrams) != 1 {
		t.Fatal("canonical graph consumer failed", err)
	}
	if !reflect.DeepEqual(prepared.Diagrams[0].Model.Nodes, diagram.Nodes) {
		t.Fatal("renderer reselected raw graph labels")
	}
	after, _ := canonicalJSON(input)
	if !bytes.Equal(before, after) {
		t.Fatal("redaction mutated frozen source")
	}
	replay, err := redactRenderInput(context.Background(), input)
	if err != nil || compareRenderMaterialization(redacted, replay) != nil {
		t.Fatal("graph replay differs", err)
	}
	t.Run("manual geometry and overrides resolve stable refs", func(t *testing.T) {
		graph := diagram
		graph.Nodes = append([]reportingDiagramNode{}, diagram.Nodes...)
		graph.LayoutMode = "manual"
		graph.CompositionLayout = &compositionDiagramLayout{SchemaID: "composition_diagram_layout.v1", CoordinateSpace: diagramCoordinateSpace{Unit: "css_px", Origin: "top_left", Width: 400, Height: 300},
			NodePositions: []diagramNodePosition{{TargetRef: "v1", X: 10, Y: 20, Width: 80, Height: 40}, {TargetRef: "v2", X: 200, Y: 100, Width: 60, Height: 40}}, EdgeRoutes: []diagramEdgeRoute{}}
		override := diagramLabelOverride{Label: "Reviewed endpoint"}
		override.Target.Kind, override.Target.Ref = "vertex", "v1"
		if err := applyDiagramLabelOverrides(&graph, []diagramLabelOverride{override}); err != nil {
			t.Fatal(err)
		}
		wire, err := serializeManualDiagramSVG(graph)
		if err != nil || !bytes.Contains(wire, []byte(override.Label)) || bytes.Contains(wire, []byte("192.0.2.1")) {
			t.Fatal("manual render lost retained graph binding", err)
		}
		override.Target.Ref = "n0001"
		if err := applyDiagramLabelOverrides(&graph, []diagramLabelOverride{override}); err == nil {
			t.Fatal("renderer-local override target admitted")
		}
	})
	t.Run("truncate uses the selected scalar rule", func(t *testing.T) {
		candidate := input
		count := 4
		candidate.RedactionProfile.Rules = []RedactionRule{{RuleID: "truncate_endpoint", Path: &endpoint, Action: RedactionActionSpec{Type: ActionTruncate, MaxChars: &count}}}
		candidate.RedactionProfileSHA256, _ = ValidateRedactionProfile(candidate.RedactionProfile)
		got, err := redactRenderInput(context.Background(), candidate)
		if err != nil || got.Model.Diagrams[0].Nodes[0].Label != "192."+DefaultTruncateMark {
			t.Fatal("graph truncation bypassed rule", err)
		}
	})

	t.Run("drop components keeps topology with ordinal fallback", func(t *testing.T) {
		candidate := input
		class := ContentClassDerivedAnalytic
		candidate.RedactionProfile.Rules = []RedactionRule{{RuleID: "drop_graph_values", ContentClass: &class, Action: RedactionActionSpec{Type: ActionDrop}}}
		candidate.RedactionProfileSHA256, _ = ValidateRedactionProfile(candidate.RedactionProfile)
		got, err := redactRenderInput(context.Background(), candidate)
		if err != nil {
			t.Fatal(err)
		}
		graph := got.Model.Diagrams[0]
		if graph.Nodes[0].Label != "Vertex 0001" || graph.Edges[0].Label != "Edge 0001" || !reflect.DeepEqual(graph.IncludedVertexRefs, diagram.IncludedVertexRefs) {
			t.Fatal("removed components restored source labels or changed topology")
		}
	})
	t.Run("exact path rejects unselected graph components", func(t *testing.T) {
		candidate := input
		path := "$.diagrams['graph'].nodes['v3'].label_components['endpoint_value']"
		candidate.RedactionProfile.Rules = []RedactionRule{{RuleID: "unselected", Path: &path, Action: RedactionActionSpec{Type: ActionDrop}}}
		candidate.RedactionProfileSHA256, _ = ValidateRedactionProfile(candidate.RedactionProfile)
		if _, err := redactRenderInput(context.Background(), candidate); err == nil {
			t.Fatal("unbound path ignored")
		}
	})
	t.Run("graph token subject is not inferred", func(t *testing.T) {
		candidate := input
		candidate.RedactionProfile.Rules = []RedactionRule{{RuleID: "token", Path: &endpoint, Action: RedactionActionSpec{Type: ActionMask, TokenBacked: true}}}
		candidate.RedactionProfileSHA256, _ = ValidateRedactionProfile(candidate.RedactionProfile)
		_, err := redactRenderInput(context.Background(), candidate)
		var invalid *renderValidationError
		if !errors.As(err, &invalid) || invalid.ReasonCode != "token_subject_not_unique" {
			t.Fatal("unmapped graph subject tokenized", err)
		}
	})
	t.Run("profile partition restriction still applies", func(t *testing.T) {
		candidate := input
		candidate.RedactionProfile.AllowedDisclosurePartitionRefs = []string{"public"}
		candidate.RedactionProfileSHA256, _ = ValidateRedactionProfile(candidate.RedactionProfile)
		_, err := redactRenderInput(context.Background(), candidate)
		var invalid *renderValidationError
		if !errors.As(err, &invalid) || invalid.ReasonCode != "disclosure_partition_unresolved" {
			t.Fatal("mask declassified graph source", err)
		}
	})
}

func mustFixtureRenderLimits(t *testing.T, contract TemplateContract) renderLimits {
	t.Helper()
	limits, err := resolveRenderLimits(contract.DeclaredLimits)
	if err != nil {
		t.Fatal(err)
	}
	return limits
}

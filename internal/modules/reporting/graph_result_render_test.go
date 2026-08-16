package reporting

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func TestGraphResultSelectionRulesPreserveExactProjectionOrder_Unit(t *testing.T) {
	result := reportingGraphSelectionFixture()
	depth := 2
	cases := []struct {
		name         string
		rule         diagramSelectionRule
		wantVertices string
		wantEdges    string
	}{
		{
			name:         "explicit",
			rule:         diagramSelectionRule{RuleID: "explicit_refs", VertexRefs: []string{"v2", "v1"}, EdgeRefs: []string{"e12"}, OverflowPolicy: "fail"},
			wantVertices: "v1,v2", wantEdges: "e12",
		},
		{
			name:         "neighborhood",
			rule:         diagramSelectionRule{RuleID: "neighborhood", SeedVertexRefs: []string{"v2"}, Depth: &depth, OverflowPolicy: "fail"},
			wantVertices: "v1,v2,v3,v4", wantEdges: "e12,e23,e34",
		},
		{
			name:         "all",
			rule:         diagramSelectionRule{RuleID: "all_with_bounds", OverflowPolicy: "fail"},
			wantVertices: "v1,v2,v3,v4", wantEdges: "e12,e23,e34",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vertices, edges, overflow, err := selectGraphDiagramObjects(result, tc.rule)
			if err != nil || overflow != nil {
				t.Fatalf("select graph result: overflow=%#v err=%v", overflow, err)
			}
			if got := joinVertexIDs(vertices); got != tc.wantVertices {
				t.Fatalf("vertices = %q want %q", got, tc.wantVertices)
			}
			if got := joinEdgeIDs(edges); got != tc.wantEdges {
				t.Fatalf("edges = %q want %q", got, tc.wantEdges)
			}
		})
	}

	_, _, _, err := selectGraphDiagramObjects(result, diagramSelectionRule{RuleID: "explicit_refs", VertexRefs: []string{"v1", "v1"}, OverflowPolicy: "fail"})
	var renderErr *renderValidationError
	if !errors.As(err, &renderErr) || renderErr.ReasonCode != "diagram_selection_duplicate_ref" {
		t.Fatalf("duplicate selection error = %#v", err)
	}
	_, _, _, err = selectGraphDiagramObjects(result, diagramSelectionRule{RuleID: "explicit_refs", VertexRefs: []string{"missing"}, OverflowPolicy: "fail"})
	if !errors.As(err, &renderErr) || renderErr.ReasonCode != "diagram_selection_missing_ref" {
		t.Fatalf("missing selection error = %#v", err)
	}
}

func TestGraphResultSelectionBoundsAndActualTopologyRendering_Unit(t *testing.T) {
	bounded := graphprojection.CompletedResultV2{}
	for index := 1; index <= 81; index++ {
		bounded.Vertices = append(bounded.Vertices, graphprojection.ResultVertexV2{VertexID: fmt.Sprintf("v%03d", index), SortKey: fmt.Sprintf("%03d", index)})
	}
	_, _, _, err := selectGraphDiagramObjects(bounded, diagramSelectionRule{RuleID: "all_with_bounds", OverflowPolicy: "fail"})
	var renderErr *renderValidationError
	if !errors.As(err, &renderErr) || renderErr.ReasonCode != "diagram_hard_limit_exceeded" {
		t.Fatalf("hard-limit error = %#v", err)
	}
	vertices, edges, overflow, err := selectGraphDiagramObjects(bounded, diagramSelectionRule{RuleID: "all_with_bounds", OverflowPolicy: "summarize"})
	if err != nil || len(vertices) != 80 || len(edges) != 0 || overflow == nil || overflow.OmittedVertexCount != 1 || overflow.FirstOmittedRef == nil || *overflow.FirstOmittedRef != "v081" {
		t.Fatalf("bounded summary = vertices=%d edges=%d overflow=%#v err=%v", len(vertices), len(edges), overflow, err)
	}

	ref := testSourceProjectionRef("gv-render", "gpres_"+strings.Repeat("d", 64))
	result := reportingGraphSelectionFixture()
	result.Binding = ref.binding()
	diagram, err := graphDiagramFromResult(compositionDiagramDecl{
		DeclID: "diagram-graph", DiagramSourceKind: "graph",
		SelectionRule: diagramSelectionRule{SchemaID: "diagram_selection_rule.v1", RuleID: "explicit_refs", VertexRefs: []string{"v1", "v2"}, EdgeRefs: []string{"e12"}, OverflowPolicy: "fail"},
	}, ref, result, ReleaseScopeInternalReview)
	if err != nil {
		t.Fatalf("derive exact graph diagram: %v", err)
	}
	source, err := serializeMermaidSource(diagram)
	if err != nil {
		t.Fatalf("serialize exact graph topology: %v", err)
	}
	if len(diagram.Nodes) != 2 || len(diagram.Edges) != 1 || !strings.Contains(string(source), "n0001 --> n0002") || strings.Contains(string(source), "gv-render") {
		t.Fatalf("rendered graph topology is not exact or disclosed identity: diagram=%#v source=%s", diagram, source)
	}
	_, err = graphDiagramFromResult(compositionDiagramDecl{
		DeclID: "diagram-external", DiagramSourceKind: "graph",
		SelectionRule: diagramSelectionRule{SchemaID: "diagram_selection_rule.v1", RuleID: "all_with_bounds", OverflowPolicy: "fail"},
	}, ref, result, ReleaseScopeExternal)
	if !errors.As(err, &renderErr) || renderErr.ReasonCode != "redaction_action_unresolved" {
		t.Fatalf("unclassified external graph values must fail closed, got %#v", err)
	}
}

func TestDiagramSelectionRuleSchemaIsClosed_Unit(t *testing.T) {
	valid := `{"schema_id":"diagram_selection_rule.v1","rule_id":"all_with_bounds","vertex_refs":[],"edge_refs":[],"seed_vertex_refs":[],"depth":null,"edge_kind_filter":[],"timeline_record_ids":[],"overflow_policy":"fail"}`
	var rule diagramSelectionRule
	if err := json.Unmarshal([]byte(valid), &rule); err != nil {
		t.Fatalf("decode closed selection rule: %v", err)
	}
	for _, invalid := range []string{
		strings.Replace(valid, `,"overflow_policy"`, `,"unknown":true,"overflow_policy"`, 1),
		strings.Replace(valid, `"vertex_refs":[],`, "", 1),
		strings.Replace(valid, `"depth":null`, `"depth":1`, 1),
		strings.Replace(valid, `"overflow_policy":"fail"`, `"overflow_policy":"truncate"`, 1),
	} {
		if err := json.Unmarshal([]byte(invalid), &rule); err == nil {
			t.Fatalf("invalid selection rule was accepted: %s", invalid)
		}
	}
}

func reportingGraphSelectionFixture() graphprojection.CompletedResultV2 {
	return graphprojection.CompletedResultV2{
		Vertices: []graphprojection.ResultVertexV2{{VertexID: "v1"}, {VertexID: "v2"}, {VertexID: "v3"}, {VertexID: "v4"}},
		Edges: []graphprojection.ResultEdgeV2{
			{EdgeID: "e12", EdgeKind: "flow", SrcVertexID: "v1", DstVertexID: "v2"},
			{EdgeID: "e23", EdgeKind: "dns", SrcVertexID: "v2", DstVertexID: "v3"},
			{EdgeID: "e34", EdgeKind: "flow", SrcVertexID: "v3", DstVertexID: "v4"},
		},
	}
}

func joinVertexIDs(vertices []graphprojection.ResultVertexV2) string {
	ids := make([]string, 0, len(vertices))
	for _, vertex := range vertices {
		ids = append(ids, vertex.VertexID)
	}
	return strings.Join(ids, ",")
}

func joinEdgeIDs(edges []graphprojection.ResultEdgeV2) string {
	ids := make([]string, 0, len(edges))
	for _, edge := range edges {
		ids = append(ids, edge.EdgeID)
	}
	return strings.Join(ids, ",")
}

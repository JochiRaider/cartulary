package reporting

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestMermaidSourceCanonicalGrammar(t *testing.T) {
	data, err := os.ReadFile("../../../contracts/reporting/fixtures/mermaid-source.v3.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Diagram     reportingDiagramModel `json:"diagram"`
			SourceLines []string              `json:"source_lines"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, vector := range fixture.Cases {
		got, err := serializeMermaidSource(vector.Diagram)
		want := strings.Join(vector.SourceLines, "\n")
		if err != nil || string(got) != want {
			t.Fatalf("%s bytes = %q, error = %v; want %q", vector.Diagram.DiagramKind, got, err, want)
		}
	}
	normalized, err := normalizeDiagramLabel("Cafe\u0301\t\n  A", false)
	if err != nil || normalized != "Café A" {
		t.Fatalf("normalization: %q %v", normalized, err)
	}

}

func TestMermaidSourceRejectsUnresolvedAndUnsafeInput(t *testing.T) {
	base := func() reportingDiagramModel {
		return reportingDiagramModel{DiagramKind: "flowchart", Nodes: []reportingDiagramNode{{NodeID: "a", Label: "A"}, {NodeID: "b", Label: "B"}}, Edges: []reportingDiagramEdge{{EdgeID: "e", From: "a", To: "b", Label: "next"}}}
	}
	for _, mutate := range []func(*reportingDiagramModel){
		func(d *reportingDiagramModel) { d.DiagramKind = "pie" },
		func(d *reportingDiagramModel) { d.Nodes[1].NodeID = "a" },
		func(d *reportingDiagramModel) { d.Nodes[0].NodeID = "" },
		func(d *reportingDiagramModel) { d.Edges[0].To = "unselected" },
		func(d *reportingDiagramModel) { d.Edges = append(d.Edges, d.Edges[0]) },
		func(d *reportingDiagramModel) { d.Nodes[0].Label = "<script>" },
		func(d *reportingDiagramModel) { d.Nodes[0].Label = "```mermaid" },
		func(d *reportingDiagramModel) { d.Nodes[0].Label = "bad\x85" },
		func(d *reportingDiagramModel) { d.Edges[0].Label = "bad\u0085" },
		func(d *reportingDiagramModel) { d.Edges[0].Label = "bad\x00" },
		func(d *reportingDiagramModel) { d.Nodes[0].Label = strings.Repeat("字", 121) },
		func(d *reportingDiagramModel) { d.Nodes = make([]reportingDiagramNode, 81) },
		func(d *reportingDiagramModel) { d.DiagramKind = "sequence"; d.Nodes = make([]reportingDiagramNode, 25) },
		func(d *reportingDiagramModel) { d.Edges = make([]reportingDiagramEdge, 161) },
	} {
		diagram := base()
		mutate(&diagram)
		data, err := serializeMermaidSource(diagram)
		var failure *renderValidationError
		if len(data) != 0 || !errors.As(err, &failure) || failure.FailureCode != "mermaid_invalid" {
			t.Fatalf("unsafe source must return no bytes and a bounded failure: %q %v", data, err)
		}
	}
	label, err := normalizeDiagramLabel(strings.Repeat("字", 121), true)
	if err != nil || label != strings.Repeat("字", 119)+"…" {
		t.Fatalf("scalar truncation = %q %v", label, err)
	}
	if _, err := normalizeDiagramLabel(strings.Repeat("字", 241), true); err == nil {
		t.Fatal("hard label bound bypassed by truncation")
	}
}

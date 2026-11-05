package reporting

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func testDiagramDirection(t *testing.T) {
	for _, tc := range []struct {
		direction string
		heads     int
		connector string
	}{{"forward", 1, "-->"}, {"reverse", 1, "-->"}, {"undirected", 0, "---"}, {"bidirectional", 2, "<-->"}} {
		t.Run("direction_"+tc.direction, func(t *testing.T) {
			diagram := manualDiagramFixture()
			diagram.Edges[0].Direction = tc.direction
			svg, err := serializeManualDiagramSVG(diagram)
			if err != nil || bytes.Count(svg, []byte("<polygon")) != tc.heads {
				t.Fatalf("SVG topology: %s %v", svg, err)
			}
			source, err := serializeMermaidSource(diagram)
			if err != nil || !strings.Contains(string(source), tc.connector) {
				t.Fatalf("Mermaid topology: %s %v", source, err)
			}
			if tc.direction == "reverse" && (!strings.Contains(string(source), `n0002 -->|"next"| n0001`) || !bytes.Contains(svg, []byte(`points="90 40 96 43 96 37"`))) {
				t.Fatalf("reverse edge arrow is attached to the wrong endpoint: %s / %s", source, svg)
			}
			diagram.DiagramKind = "sequence"
			source, err = serializeMermaidSource(diagram)
			if tc.direction == "undirected" || tc.direction == "bidirectional" {
				var failure *renderValidationError
				if len(source) != 0 || !errors.As(err, &failure) || failure.ReasonCode != "diagram_direction_unsupported" {
					t.Fatalf("ambiguous sequence admitted: %s %v", source, err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

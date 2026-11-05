package reporting

import (
	"fmt"
	"strings"
)

func diagramEdgeConnector(direction string) (string, bool) {
	switch direction {
	case "forward", "reverse":
		return "-->", true
	case "undirected":
		return "---", true
	case "bidirectional":
		return "<-->", true
	default:
		return "", false
	}
}

// Passive SVG uses explicit geometry, never URL markers or external resources.
func writeDiagramArrow(out *strings.Builder, tip, previous diagramPoint) {
	dx, dy := previous.X-tip.X, previous.Y-tip.Y
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	ux, uy := 0, 0
	if abs(dx) >= abs(dy) {
		ux = 1
		if dx < 0 {
			ux = -1
		}
	} else {
		uy = 1
		if dy < 0 {
			uy = -1
		}
	}
	base := diagramPoint{tip.X + 6*ux, tip.Y + 6*uy}
	fmt.Fprintf(out, "  <polygon points=\"%d %d %d %d %d %d\" fill=\"#334155\"/>\n", tip.X, tip.Y, base.X-3*uy, base.Y+3*ux, base.X+3*uy, base.Y-3*ux)
}

func reportingProjectionDirection(direction string) string {
	if direction == "directed" {
		return "forward"
	}
	return direction
}

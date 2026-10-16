package reporting

import (
	"fmt"
	"html"
	"strings"
)

// The companion owns admission of the closed wire shape. Reporting retains its
// typed projection and resolves opaque refs only against the selected diagram.
type compositionDiagramLayout struct {
	SchemaID        string                 `json:"schema_id"`
	CoordinateSpace diagramCoordinateSpace `json:"coordinate_space"`
	NodePositions   []diagramNodePosition  `json:"node_positions"`
	EdgeRoutes      []diagramEdgeRoute     `json:"edge_routes"`
}
type diagramCoordinateSpace struct {
	Unit   string `json:"unit"`
	Origin string `json:"origin"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
type diagramPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type diagramNodePosition struct {
	TargetRef string `json:"target_ref"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}
type diagramEdgeRoute struct {
	TargetRef string         `json:"target_ref"`
	RouteKind string         `json:"route_kind"`
	Waypoints []diagramPoint `json:"waypoints"`
}

func resolveDiagramLayout(d reportingDiagramModel) (map[string]diagramNodePosition, map[string][]diagramPoint, error) {
	invalid := func(reason string) (map[string]diagramNodePosition, map[string][]diagramPoint, error) {
		return nil, nil, compositionFailure(reason)
	}
	l := d.CompositionLayout
	if d.LayoutMode != "manual" || d.DiagramKind != "flowchart" || l == nil || l.SchemaID != "composition_diagram_layout.v1" || l.NodePositions == nil || l.EdgeRoutes == nil {
		return invalid("diagram_layout_invalid")
	}
	s := l.CoordinateSpace
	if s.Unit != "css_px" || s.Origin != "top_left" || s.Width < 1 || s.Width > 10000 || s.Height < 1 || s.Height > 10000 || len(d.Nodes) == 0 || len(d.Nodes) != len(d.IncludedVertexRefs) || len(d.Edges) != len(d.IncludedEdgeRefs) {
		return invalid("diagram_layout_invalid")
	}
	vertices, edges := map[string]string{}, map[string]string{}
	for i, ref := range d.IncludedVertexRefs {
		if ref == "" || vertices[ref] != "" || d.Nodes[i].NodeID == "" {
			return invalid("diagram_layout_invalid")
		}
		vertices[ref] = d.Nodes[i].NodeID
	}
	for i, ref := range d.IncludedEdgeRefs {
		if ref == "" || edges[ref] != "" || d.Edges[i].EdgeID == "" {
			return invalid("diagram_layout_invalid")
		}
		edges[ref] = d.Edges[i].EdgeID
	}
	positions, routes := map[string]diagramNodePosition{}, map[string][]diagramPoint{}
	for i, p := range l.NodePositions {
		id := vertices[p.TargetRef]
		if id == "" {
			return invalid("diagram_layout_unknown_target")
		}
		if _, exists := positions[id]; exists {
			return invalid("diagram_layout_duplicate_target")
		}
		if i > 0 && l.NodePositions[i-1].TargetRef >= p.TargetRef || p.X < 0 || p.Y < 0 || p.Width <= 0 || p.Height <= 0 || p.Width > s.Width || p.Height > s.Height || p.X > s.Width-p.Width || p.Y > s.Height-p.Height {
			return invalid("diagram_layout_invalid")
		}
		positions[id] = p
	}
	if len(positions) != len(d.Nodes) {
		return invalid("diagram_layout_missing_node_position")
	}
	for i, r := range l.EdgeRoutes {
		id := edges[r.TargetRef]
		if id == "" {
			return invalid("diagram_layout_unknown_target")
		}
		if _, exists := routes[id]; exists {
			return invalid("diagram_layout_duplicate_target")
		}
		if i > 0 && l.EdgeRoutes[i-1].TargetRef >= r.TargetRef || r.RouteKind != "polyline" || r.Waypoints == nil || len(r.Waypoints) > 32 {
			return invalid("diagram_layout_invalid")
		}
		for _, p := range r.Waypoints {
			if p.X < 0 || p.Y < 0 || p.X > s.Width || p.Y > s.Height {
				return invalid("diagram_layout_invalid")
			}
		}
		routes[id] = append([]diagramPoint{}, r.Waypoints...)
	}
	return positions, routes, nil
}

func diagramNodeCenter(p diagramNodePosition) diagramPoint {
	return diagramPoint{p.X + p.Width/2, p.Y + p.Height/2}
}
func diagramEdgeEndpoint(p diagramNodePosition, toward diagramPoint) (diagramPoint, error) {
	c := diagramNodeCenter(p)
	dx, dy := toward.X-c.X, toward.Y-c.Y
	if dx == 0 && dy == 0 {
		return diagramPoint{}, compositionFailure("diagram_layout_invalid")
	}
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	if abs(dx) >= abs(dy) {
		c.X = p.X
		if dx > 0 {
			c.X += p.Width
		}
	} else {
		c.Y = p.Y
		if dy > 0 {
			c.Y += p.Height
		}
	}
	return c, nil
}

func serializeManualDiagramSVG(d reportingDiagramModel) ([]byte, error) {
	positions, routes, err := resolveDiagramLayout(d)
	if err != nil {
		return nil, err
	}
	var out strings.Builder
	s := d.CompositionLayout.CoordinateSpace
	fmt.Fprintf(&out, "<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 %d %d\">\n", s.Width, s.Height)
	seen := map[string]bool{}
	for _, e := range d.Edges {
		from, okFrom := positions[e.From]
		to, okTo := positions[e.To]
		if !okFrom || !okTo || seen[e.EdgeID] {
			return nil, compositionFailure("diagram_layout_invalid")
		}
		seen[e.EdgeID] = true
		points := append([]diagramPoint{diagramNodeCenter(from)}, routes[e.EdgeID]...)
		points = append(points, diagramNodeCenter(to))
		start, err := diagramEdgeEndpoint(from, points[1])
		if err != nil {
			return nil, err
		}
		end, err := diagramEdgeEndpoint(to, points[len(points)-2])
		if err != nil {
			return nil, err
		}
		points[0], points[len(points)-1] = start, end
		out.WriteString("  <path d=\"")
		for i, p := range points {
			command := " L"
			if i == 0 {
				command = "M"
			}
			fmt.Fprintf(&out, "%s%d %d", command, p.X, p.Y)
		}
		out.WriteString("\" fill=\"none\" stroke=\"#334155\" stroke-width=\"2\"/>\n")
		// A marker-free arrow preserves the selected edge's direction without
		// fragment IDs, URL paints, or renderer-dependent marker geometry.
		center := diagramNodeCenter(to)
		toward := diagramNodeCenter(from)
		if route := routes[e.EdgeID]; len(route) > 0 {
			toward = route[len(route)-1]
		}
		dx, dy := toward.X-center.X, toward.Y-center.Y
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
		base := diagramPoint{end.X + 6*ux, end.Y + 6*uy}
		fmt.Fprintf(&out, "  <polygon points=\"%d %d %d %d %d %d\" fill=\"#334155\"/>\n", end.X, end.Y, base.X-3*uy, base.Y+3*ux, base.X+3*uy, base.Y-3*ux)
		label, err := normalizeDiagramLabel(e.Label, false)
		if err != nil {
			return nil, err
		}
		if label != "" {
			start, end := points[0], points[len(points)-1]
			fmt.Fprintf(&out, "  <text x=\"%d\" y=\"%d\" font-family=\"sans-serif\" font-size=\"12\" fill=\"#0f172a\">%s</text>\n", (start.X+end.X)/2, (start.Y+end.Y)/2, html.EscapeString(label))
		}
	}
	seen = map[string]bool{}
	for _, n := range d.Nodes {
		if seen[n.NodeID] {
			return nil, compositionFailure("diagram_layout_invalid")
		}
		seen[n.NodeID] = true
		p := positions[n.NodeID]
		label, err := normalizeDiagramLabel(n.Label, false)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&out, "  <rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"#ffffff\" stroke=\"#334155\" stroke-width=\"1\"/>\n", p.X, p.Y, p.Width, p.Height)
		c := diagramNodeCenter(p)
		fmt.Fprintf(&out, "  <text x=\"%d\" y=\"%d\" font-family=\"sans-serif\" font-size=\"14\" text-anchor=\"middle\" fill=\"#0f172a\">%s</text>\n", c.X, c.Y, html.EscapeString(label))
	}
	out.WriteString("</svg>\n")
	data := []byte(out.String())
	if err := validateRenderedSVG(data); err != nil {
		return nil, err
	}
	return data, nil
}

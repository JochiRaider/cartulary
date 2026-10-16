package reporting

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Source identifiers are assigned only from retained selection order. Opaque
// source refs never become Mermaid syntax or persist as a renderer ID mapping.
func serializeMermaidSource(diagram reportingDiagramModel) ([]byte, error) {
	issues := renderIssueCollector{stage: "mermaid_source"}
	invalid := func() {
		issues.add(newRenderValidationError("mermaid_invalid", "invalid_mermaid_construct"))
	}
	prefix, header, vertexLimit := "n", "flowchart TD", 80
	switch diagram.DiagramKind {
	case "flowchart":
	case "sequence":
		prefix, header, vertexLimit = "p", "sequenceDiagram", 24
	default:
		invalid()
	}
	if len(diagram.Nodes) == 0 {
		issues.add(newRenderValidationError("mermaid_invalid", "no_diagrams_selected"))
	}
	if len(diagram.Nodes) > vertexLimit || len(diagram.Edges) > 160 {
		issues.add(newRenderValidationError("mermaid_invalid", "diagram_hard_limit_exceeded"))
	}
	var out strings.Builder
	out.WriteString(header + "\n")
	ids := make(map[string]string, len(diagram.Nodes))
	for index, node := range diagram.Nodes {
		if node.NodeID == "" || ids[node.NodeID] != "" {
			invalid()
		}
		id := fmt.Sprintf("%s%04d", prefix, index+1)
		ids[node.NodeID] = id
		label, err := normalizeDiagramLabel(node.Label, false)
		if err != nil {
			issues.add(err)
			continue
		}
		if diagram.DiagramKind == "sequence" {
			fmt.Fprintf(&out, "  participant %s as %s\n", id, escapeMermaidLabel(label))
		} else {
			fmt.Fprintf(&out, "  %s[\"%s\"]\n", id, escapeMermaidLabel(label))
		}
	}
	edges := make(map[string]bool, len(diagram.Edges))
	for _, edge := range diagram.Edges {
		from, to := ids[edge.From], ids[edge.To]
		if from == "" || to == "" || edge.EdgeID == "" || edges[edge.EdgeID] {
			invalid()
		}
		edges[edge.EdgeID] = true
		label, err := normalizeDiagramLabel(edge.Label, false)
		if err != nil {
			issues.add(err)
			continue
		}
		if diagram.DiagramKind == "sequence" {
			fmt.Fprintf(&out, "  %s->>%s: %s\n", from, to, escapeMermaidLabel(label))
		} else if label == "" {
			fmt.Fprintf(&out, "  %s --> %s\n", from, to)
		} else {
			fmt.Fprintf(&out, "  %s -->|\"%s\"| %s\n", from, escapeMermaidLabel(label), to)
		}
	}
	if err := issues.finish(); err != nil {
		return nil, err
	}
	return []byte(out.String()), nil
}

func normalizeDiagramLabel(raw string, truncate bool) (string, error) {
	if !utf8.ValidString(raw) {
		return "", newRenderValidationError("mermaid_invalid", "invalid_mermaid_construct")
	}
	var out strings.Builder
	space := false
	for _, ch := range norm.NFC.String(raw) {
		if (ch < 0x20 && ch != '\n' && ch != '\r' && ch != '\t') || (ch >= 0x7f && ch <= 0x9f) || ch == '<' || ch == '>' {
			return "", newRenderValidationError("mermaid_invalid", "invalid_mermaid_construct")
		}
		if unicode.IsSpace(ch) {
			if !space {
				out.WriteByte(' ')
			}
			space = true
			continue
		}
		space = false
		out.WriteRune(ch)
	}
	label := out.String()
	if strings.Contains(label, "```") {
		return "", newRenderValidationError("mermaid_invalid", "invalid_mermaid_construct")
	}
	runes := []rune(label)
	if len(runes) > 240 || (len(runes) > 120 && !truncate) {
		return "", newRenderValidationError("mermaid_invalid", "label_length_exceeded")
	}
	if len(runes) > 120 {
		label = string(runes[:119]) + "…"
	}
	return label, nil
}

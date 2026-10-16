package reporting

type preparedMermaidDiagram struct {
	ID        string
	Model     reportingDiagramModel
	ModelJSON []byte
	Source    []byte
	SVG       []byte
}

type preparedMermaid struct {
	Diagrams []preparedMermaidDiagram
	Files    []RenderBundleFile
}

// Derivation and source validation are pure and finish for every diagram before
// any renderer starts. Recovery uses the same computation with retained inputs.
func prepareDiagrams(contract TemplateContract, model RedactedExportModel, kind string, limits renderLimits) (preparedMermaid, error) {
	model, err := readAssembledRenderModel(contract, model)
	if err != nil {
		return preparedMermaid{}, err
	}
	diagrams := model.Diagrams
	if len(diagrams) == 0 && kind == OutputKindMermaid {
		return preparedMermaid{}, newRenderValidationError("mermaid_invalid", "no_diagrams_selected")
	}
	if err := limits.check("diagrams.count", int64(len(diagrams))); err != nil {
		return preparedMermaid{}, err
	}
	prepared := preparedMermaid{Diagrams: []preparedMermaidDiagram{}, Files: []RenderBundleFile{}}
	issues := renderIssueCollector{stage: "mermaid_source"}
	for _, diagram := range diagrams {
		var source, svg []byte
		if diagram.LayoutMode == "manual" {
			svg, err = serializeManualDiagramSVG(diagram)
		} else {
			source, err = serializeMermaidSource(diagram)
		}
		issues.add(err)
		model, err := canonicalJSON(diagram)
		if err != nil {
			return preparedMermaid{}, err
		}
		prepared.Diagrams = append(prepared.Diagrams, preparedMermaidDiagram{ID: diagram.DiagramID, Model: diagram, ModelJSON: model, Source: source, SVG: svg})
		prepared.Files = append(prepared.Files, newBundleFile("intermediate/diagrams/"+diagram.DiagramID+".json", renderBundleRoleDiagramModel, "application/vnd.cartulary.reporting-diagram+json", model, true))
		if diagram.LayoutMode == "manual" {
			prepared.Files = append(prepared.Files, newBundleFile("diagrams/"+diagram.DiagramID+".svg", renderBundleRoleRenderedSVG, "image/svg+xml", svg, true))
		} else {
			prepared.Files = append(prepared.Files, newBundleFile("diagrams/"+diagram.DiagramID+".mmd", renderBundleRoleSourceMermaid, "text/vnd.cartulary.mermaid; charset=utf-8", source, true))
		}
	}
	if err := issues.finish(); err != nil {
		return preparedMermaid{}, err
	}
	return prepared, nil
}

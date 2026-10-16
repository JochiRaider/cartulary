package reporting

import (
	"encoding/json"
	"sort"
)

// Pure preparation is shared by production rendering and retained-output
// admission. It reads only frozen inputs and never starts a browser or service.
type preparedSlidev struct {
	Deck     reportingDeckModel
	DeckJSON []byte
	Source   []byte
	Files    []RenderBundleFile
	Diagrams []preparedMermaidDiagram
	Assets   []renderStageAsset
}

func prepareSlidev(contract TemplateContract, model RedactedExportModel, options renderOutputOptions, composition json.RawMessage, limits renderLimits) (preparedSlidev, error) {
	model, err := readAssembledRenderModel(contract, model)
	if err != nil {
		return preparedSlidev{}, err
	}
	diagrams, err := prepareDiagrams(contract, model, OutputKindSlidev, limits)
	if err != nil {
		return preparedSlidev{}, err
	}
	assets := []renderStageAsset{}
	for _, d := range diagrams.Diagrams {
		if d.Model.LayoutMode == "manual" {
			assets = append(assets, renderStageAsset{Path: "diagrams/" + d.ID + ".svg", MediaType: "image/svg+xml", Bytes: d.SVG, SHA256: hashHex(d.SVG)})
		}
	}
	deck, err := deriveDeckModel(contract, model, options, composition)
	if err != nil {
		return preparedSlidev{}, err
	}
	files, err := retainedTemplateAssetFiles(contract, model.assembly.templateFiles, model.Assets)
	if err != nil {
		return preparedSlidev{}, err
	}
	for _, file := range files {
		assets = append(assets, renderStageAsset{Path: file.Path, MediaType: file.MediaType, Bytes: file.Bytes, SHA256: file.SHA256})
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	diagrams.Files = append(diagrams.Files, files...)
	issues := renderIssueCollector{stage: "slidev_source"}
	issues.add(validateDeckLimits(limits, deck))
	source, err := serializeSlidevMarkdown(deck, contract, diagrams.Diagrams...)
	issues.add(err)
	if err := issues.finish(); err != nil {
		return preparedSlidev{}, err
	}
	encoded, err := canonicalJSON(deck)
	if err != nil {
		return preparedSlidev{}, err
	}
	return preparedSlidev{Deck: deck, DeckJSON: encoded, Source: source, Diagrams: diagrams.Diagrams, Assets: assets, Files: append(diagrams.Files, []RenderBundleFile{
		newBundleFile("intermediate/deck-model.json", renderBundleRoleDeckModel, "application/vnd.cartulary.reporting-deck-model+json", encoded, true),
		newBundleFile("slides.md", renderBundleRoleSourceSlidev, "text/markdown; charset=utf-8", source, true),
	}...)}, nil
}

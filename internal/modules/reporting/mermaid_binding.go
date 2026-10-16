package reporting

import (
	"bytes"
	"errors"
)

func validateMermaidCandidateBinding(payload renderInput, model RedactedExportModel, bundle RenderBundle) error {
	invalid := errors.New("reporting: retained presentation differs from admission")
	options, err := parseRenderOutputOptions(payload.OutputOptions, payload.OutputKind, payload.ReleaseScope)
	if err != nil {
		return invalid
	}
	limits, err := resolveRenderLimits(payload.TemplateContract.DeclaredLimits)
	if err != nil {
		return invalid
	}
	refs, err := decodeSourceProjectionRefs(payload.GraphProjectionRefs)
	if err != nil || validateFrozenGraphSources(payload.GraphSources, refs) != nil {
		return invalid
	}
	prepared, err := prepareDiagrams(payload.TemplateContract, model, OutputKindMermaid, limits)
	if err != nil {
		return invalid
	}
	expected := make(map[string]RenderBundleFile, len(prepared.Files))
	for _, file := range prepared.Files {
		expected[file.Path] = file
	}
	rendered := map[string]bool{}
	if options.SVG && !options.SourceOnly {
		for _, diagram := range prepared.Diagrams {
			rendered["diagrams/"+diagram.ID+".svg"] = true
		}
	}
	for _, file := range bundle.Files {
		switch file.Role {
		case renderBundleRoleDiagramModel, renderBundleRoleSourceMermaid:
			want, exists := expected[file.Path]
			if !exists || want.RenderBundleManifestFile != file.RenderBundleManifestFile || !bytes.Equal(want.Bytes, file.Bytes) {
				return invalid
			}
			delete(expected, file.Path)
		case renderBundleRoleRenderedSVG:
			if !rendered[file.Path] || file.MediaType != "image/svg+xml" || !file.RequiredForRelease {
				return invalid
			}
			delete(rendered, file.Path)
		case renderBundleRoleDeckModel, renderBundleRoleSourceSlidev, renderBundleRoleRenderedPDF:
			return invalid
		}
	}
	if len(expected) != 0 || len(rendered) != 0 || bundle.PrimaryPath != "diagrams/"+prepared.Diagrams[0].ID+".mmd" || bundle.PrimaryMedia != "text/vnd.cartulary.mermaid; charset=utf-8" {
		return invalid
	}
	return nil
}

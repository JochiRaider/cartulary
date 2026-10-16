package reporting

import (
	"bytes"
	"errors"
)

// Content hashes establish internal consistency, not fidelity to the frozen
// source. Recompute deterministic presentation artifacts without a renderer.
func validateSlidevCandidateBinding(payload renderInput, retained RedactedExportModel, bundle RenderBundle) error {
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
	prepared, err := prepareSlidev(payload.TemplateContract, retained, options, payload.CompositionJSON, limits)
	if err != nil {
		return invalid
	}
	expected := make(map[string]RenderBundleFile, len(prepared.Files))
	for _, file := range prepared.Files {
		expected[file.Path] = file
	}
	pdfCount := 0
	rendered := map[string]bool{}
	if options.RenderedDiagrams && !options.SourceOnly {
		for _, d := range prepared.Diagrams {
			if d.Model.LayoutMode == "auto" {
				rendered["diagrams/"+d.ID+".svg"] = true
			}
		}
	}
	for _, file := range bundle.Files {
		switch file.Role {
		case renderBundleRoleDeckModel, renderBundleRoleSourceSlidev, renderBundleRoleDiagramModel, renderBundleRoleSourceMermaid, renderBundleRoleLocalAsset, renderBundleRoleLocalTheme:
			want, exists := expected[file.Path]
			if !exists || want.RenderBundleManifestFile != file.RenderBundleManifestFile || !bytes.Equal(want.Bytes, file.Bytes) {
				return invalid
			}
			delete(expected, file.Path)
		case renderBundleRoleRenderedPDF:
			pdfCount++
			if !options.PDF || options.SourceOnly || file.Path != "deck.pdf" || file.MediaType != "application/pdf" || !file.RequiredForRelease {
				return invalid
			}
		case renderBundleRoleRenderedSVG:
			if want, exists := expected[file.Path]; exists {
				if want.RenderBundleManifestFile != file.RenderBundleManifestFile || !bytes.Equal(want.Bytes, file.Bytes) {
					return invalid
				}
				delete(expected, file.Path)
			} else {
				if !rendered[file.Path] || file.MediaType != "image/svg+xml" || !file.RequiredForRelease {
					return invalid
				}
				delete(rendered, file.Path)
			}
		}
	}
	wantPDF := 0
	if options.PDF && !options.SourceOnly {
		wantPDF = 1
	}
	if len(expected) != 0 || len(rendered) != 0 || pdfCount != wantPDF || bundle.PrimaryPath != "slides.md" || bundle.PrimaryMedia != "text/markdown; charset=utf-8" {
		return invalid
	}
	return nil
}

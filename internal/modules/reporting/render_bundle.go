package reporting

import (
	"fmt"
	"strings"
)

const (
	RenderBundleManifestSchemaID = "cartulary.render_bundle_manifest.v1"
	RenderBundleManifestVersion  = "1"

	renderBundleRendererID      = "cartulary.reporting.renderer.bundle_manifest_v1"
	renderBundleRendererVersion = "1"
	renderBundleStorageInline   = "database_inline"
)

type RenderBundle struct {
	Manifest       RenderBundleManifest
	ManifestJSON   []byte
	ManifestSHA256 string
	Files          []RenderBundleFile
	PrimaryPath    string
	PrimaryMedia   string
}

type RenderBundleManifest struct {
	SchemaID                  string                     `json:"schema_id"`
	ManifestVersion           string                     `json:"manifest_version"`
	RendererID                string                     `json:"renderer_id"`
	RendererVersion           string                     `json:"renderer_version"`
	TemplateID                string                     `json:"template_id"`
	TemplateVersion           string                     `json:"template_version"`
	OutputKind                string                     `json:"output_kind"`
	ReleaseScope              string                     `json:"release_scope"`
	PrimaryPath               string                     `json:"primary_path"`
	PrimaryMediaType          string                     `json:"primary_media_type"`
	RedactionManifestSHA256   string                     `json:"redaction_manifest_sha256"`
	GeneratedPresentation     bool                       `json:"generated_presentation"`
	SelfContained             bool                       `json:"self_contained"`
	Files                     []RenderBundleManifestFile `json:"files"`
	ValidationArtifacts       []string                   `json:"validation_artifacts"`
	ToolchainSnapshotSHA256   *string                    `json:"toolchain_snapshot_sha256"`
	SandboxObservationSHA256  *string                    `json:"sandbox_observation_sha256"`
	ExternalDeterminismSHA256 *string                    `json:"external_determinism_sha256"`
}

type RenderBundleManifestFile struct {
	Path      string `json:"path"`
	Role      string `json:"role"`
	MediaType string `json:"media_type"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
}

type RenderBundleFile struct {
	RenderBundleManifestFile
	StorageKind string
	Bytes       []byte
}

func buildRenderBundle(contract TemplateContract, kind string, releaseScope string, output []byte, mediaType string, redactionManifestSHA256 string) (RenderBundle, error) {
	primaryPath, role, err := primaryBundleFile(kind)
	if err != nil {
		return RenderBundle{}, err
	}
	file := RenderBundleFile{
		RenderBundleManifestFile: RenderBundleManifestFile{
			Path:      primaryPath,
			Role:      role,
			MediaType: mediaType,
			SHA256:    hashHex(output),
			SizeBytes: int64(len(output)),
		},
		StorageKind: renderBundleStorageInline,
		Bytes:       append([]byte(nil), output...),
	}
	manifest := RenderBundleManifest{
		SchemaID:                RenderBundleManifestSchemaID,
		ManifestVersion:         RenderBundleManifestVersion,
		RendererID:              renderBundleRendererID,
		RendererVersion:         renderBundleRendererVersion,
		TemplateID:              contract.TemplateID,
		TemplateVersion:         contract.TemplateVersion,
		OutputKind:              kind,
		ReleaseScope:            releaseScope,
		PrimaryPath:             primaryPath,
		PrimaryMediaType:        mediaType,
		RedactionManifestSHA256: redactionManifestSHA256,
		GeneratedPresentation:   true,
		SelfContained:           true,
		Files:                   []RenderBundleManifestFile{file.RenderBundleManifestFile},
		ValidationArtifacts:     []string{},
	}
	manifestJSON, err := canonicalJSON(manifest)
	if err != nil {
		return RenderBundle{}, err
	}
	return RenderBundle{
		Manifest:       manifest,
		ManifestJSON:   manifestJSON,
		ManifestSHA256: hashHex(manifestJSON),
		Files:          []RenderBundleFile{file},
		PrimaryPath:    primaryPath,
		PrimaryMedia:   mediaType,
	}, nil
}

func renderReportBundle(contract TemplateContract, kind string, model RedactedExportModel, redactionManifestSHA256 string, releaseScope string) (RenderBundle, error) {
	if err := validateTemplateContract(contract, kind, model, releaseScope); err != nil {
		return RenderBundle{}, err
	}
	output, mediaType, err := renderPrimaryBundleFile(kind, model)
	if err != nil {
		return RenderBundle{}, err
	}
	if err := ValidateSelfContainedOutput(kind, output); err != nil {
		return RenderBundle{}, err
	}
	return buildRenderBundle(contract, kind, releaseScope, output, mediaType, redactionManifestSHA256)
}

func renderPrimaryBundleFile(kind string, model RedactedExportModel) ([]byte, string, error) {
	switch kind {
	case OutputKindSlidev:
		return renderSlidevSource(model), "text/markdown; charset=utf-8", nil
	case OutputKindMermaid:
		return renderMermaidSource(model), "text/vnd.mermaid; charset=utf-8", nil
	default:
		return nil, "", fmt.Errorf("unsupported output kind %q", kind)
	}
}

func renderSlidevSource(model RedactedExportModel) []byte {
	var b strings.Builder
	b.WriteString("# Incident Report\n\n")
	for _, field := range model.Fields {
		b.WriteString("- `")
		b.WriteString(field.Path)
		b.WriteString("`: ")
		b.WriteString(formatFieldValue(field.Value))
		b.WriteString("\n")
	}
	return []byte(b.String())
}

func renderMermaidSource(model RedactedExportModel) []byte {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	b.WriteString("  snapshot[Snapshot]\n")
	for i, field := range model.Fields {
		nodeID := fmt.Sprintf("field_%d", i)
		b.WriteString("  snapshot --> ")
		b.WriteString(nodeID)
		b.WriteString("[\"")
		b.WriteString(escapeMermaidLabel(field.Path))
		b.WriteString("\"]\n")
	}
	if len(model.Fields) == 0 {
		b.WriteString("  snapshot --> report[Report]\n")
	}
	return []byte(b.String())
}

func escapeMermaidLabel(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", " ", "\r", " ")
	return replacer.Replace(value)
}

func primaryBundleFile(kind string) (string, string, error) {
	switch kind {
	case OutputKindSlidev:
		return "slides.md", "slidev_source", nil
	case OutputKindMermaid:
		return "diagrams/default.mmd", "mermaid_source", nil
	default:
		return "", "", fmt.Errorf("unsupported output kind %q", kind)
	}
}

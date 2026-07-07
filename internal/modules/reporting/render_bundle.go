package reporting

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	RenderBundleManifestSchemaID = "cartulary.render_bundle_manifest.v1"
	RenderBundleManifestVersion  = "1"

	renderBundleRendererID      = "cartulary.reporting.renderer.bundle_manifest_v1"
	renderBundleRendererVersion = "1"
	renderBundleStorageInline   = "database_inline"

	renderBundleRoleValidationSummary    = "validation_summary"
	renderBundleRoleToolchainSnapshot    = "toolchain_snapshot"
	renderBundleRoleSandboxObservation   = "sandbox_observation"
	renderBundleRoleRedactionManifest    = "redaction_manifest"
	renderBundleRoleRedactionProfileView = "redaction_profile_view"
	renderBundleRoleTokenManifest        = "token_manifest"
	renderBundleRoleSensitiveRevealMap   = "sensitive_reveal_map"
	renderBundleRoleSourceSlidev         = "source_slidev"
	renderBundleRoleSourceMermaid        = "source_mermaid"
	renderBundleRoleRenderedPDF          = "rendered_pdf"
	renderBundleRoleRenderedSVG          = "rendered_svg"
	renderBundleRoleDeckModel            = "deck_model"
	renderBundleRoleDiagramModel         = "diagram_model"
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
	TemplateManifestSHA256    string                     `json:"template_manifest_sha256"`
	OutputKind                string                     `json:"output_kind"`
	ReleaseScope              string                     `json:"release_scope"`
	PrimaryPath               string                     `json:"primary_path"`
	PrimaryMediaType          string                     `json:"primary_media_type"`
	RedactionManifestSHA256   string                     `json:"redaction_manifest_sha256"`
	TokenManifestSHA256       *string                    `json:"token_manifest_sha256,omitempty"`
	ValidationSummarySHA256   string                     `json:"validation_summary_sha256"`
	GeneratedPresentation     bool                       `json:"generated_presentation"`
	SelfContained             bool                       `json:"self_contained"`
	Files                     []RenderBundleManifestFile `json:"files"`
	ValidationArtifacts       []string                   `json:"validation_artifacts"`
	ToolchainSnapshotSHA256   *string                    `json:"toolchain_snapshot_sha256"`
	SandboxObservationSHA256  *string                    `json:"sandbox_observation_sha256"`
	ExternalDeterminismSHA256 *string                    `json:"external_determinism_sha256"`
}

type RenderBundleManifestFile struct {
	Path               string `json:"path"`
	Role               string `json:"role"`
	MediaType          string `json:"media_type"`
	SHA256             string `json:"sha256"`
	SizeBytes          int64  `json:"size_bytes"`
	RequiredForRelease bool   `json:"required_for_release"`
}

type RenderBundleFile struct {
	RenderBundleManifestFile
	StorageKind string
	Bytes       []byte
}

type RedactionBundleArtifacts struct {
	RedactionManifestJSON []byte
	ProfileViewJSON       []byte
	TokenManifestJSON     []byte
	TokenManifestSHA256   string
	RevealMapJSON         []byte
}

type renderOutputOptions struct {
	SchemaID         string `json:"schema_id"`
	SourceOnly       bool   `json:"source_only"`
	PDF              bool   `json:"pdf"`
	SVG              bool   `json:"svg"`
	PNG              bool   `json:"png"`
	PPTX             bool   `json:"pptx"`
	RenderedDiagrams bool   `json:"rendered_diagrams"`
}

type renderBundleMetadata struct {
	TemplateManifestSHA256    string
	ToolchainSnapshotSHA256   string
	SandboxObservationSHA256  string
	ValidationSummarySHA256   string
	TokenManifestSHA256       string
	ExternalDeterminismSHA256 string
}

type renderPipelineResult struct {
	Files        []RenderBundleFile
	PrimaryPath  string
	PrimaryMedia string
	Metadata     renderBundleMetadata
}

type renderValidationError struct {
	FailureCode string
	ReasonCode  string
}

func (e *renderValidationError) Error() string {
	return e.FailureCode + ": " + e.ReasonCode
}

func newRenderValidationError(failureCode string, reasonCode string) error {
	return &renderValidationError{FailureCode: failureCode, ReasonCode: reasonCode}
}

type reportingToolchainSnapshot struct {
	SchemaID               string   `json:"schema_id"`
	RendererID             string   `json:"renderer_id"`
	RendererVersion        string   `json:"renderer_version"`
	SlidevVersion          *string  `json:"slidev_version"`
	MermaidVersion         *string  `json:"mermaid_version"`
	ChromiumExecutablePath *string  `json:"chromium_executable_path"`
	Locale                 string   `json:"locale"`
	Timezone               string   `json:"timezone"`
	ViewportCSSPX          string   `json:"viewport_css_px"`
	DeviceScaleFactor      int      `json:"device_scale_factor"`
	ColorScheme            string   `json:"color_scheme"`
	BrowserLaunchArgs      []string `json:"browser_launch_args"`
	EnvAllowlist           []string `json:"env_allowlist"`
	NetworkPolicyID        string   `json:"network_policy_id"`
}

type reportingSandboxObservation struct {
	SchemaID                string   `json:"schema_id"`
	PolicyID                string   `json:"policy_id"`
	NetworkPolicyID         string   `json:"network_policy_id"`
	NetworkAccessObserved   bool     `json:"network_access_observed"`
	RemoteAssetRefsObserved []string `json:"remote_asset_refs_observed"`
	ProcessIsolation        string   `json:"process_isolation"`
	FilesystemMode          string   `json:"filesystem_mode"`
}

type reportingRenderValidationSummary struct {
	SchemaID      string                           `json:"schema_id"`
	Result        string                           `json:"result"`
	TerminalStage string                           `json:"terminal_stage"`
	FailureCode   *string                          `json:"failure_code"`
	ReasonCode    *string                          `json:"reason_code"`
	SafeDetails   map[string]any                   `json:"safe_details"`
	IssueCount    int                              `json:"issue_count"`
	Issues        []reportingRenderValidationIssue `json:"issues"`
	FirstFailure  *reportingRenderValidationIssue  `json:"first_failure"`
}

type reportingRenderValidationIssue struct {
	Stage           string         `json:"stage"`
	Severity        string         `json:"severity"`
	ExportModelPath *string        `json:"export_model_path"`
	BundlePath      *string        `json:"bundle_path"`
	FailureCode     *string        `json:"failure_code"`
	ReasonCode      *string        `json:"reason_code"`
	SafeDetails     map[string]any `json:"safe_details"`
}

type reportingDeckModel struct {
	SchemaID            string               `json:"schema_id"`
	DerivationAlgorithm string               `json:"derivation_algorithm"`
	Title               string               `json:"title"`
	Slides              []reportingDeckSlide `json:"slides"`
	SlideCount          int                  `json:"slide_count"`
	ClickStepCount      int                  `json:"click_step_count"`
	ExpectedPageCount   int                  `json:"expected_export_page_count"`
	CompositionApplied  bool                 `json:"composition_applied"`
	CompositionSummary  map[string]any       `json:"composition_summary"`
}

type reportingDeckSlide struct {
	SlideID      string               `json:"slide_id"`
	Ordinal      int                  `json:"ordinal"`
	Layout       string               `json:"layout"`
	Title        string               `json:"title"`
	Blocks       []reportingDeckBlock `json:"blocks"`
	SpeakerNotes []string             `json:"speaker_notes"`
}

type reportingDeckBlock struct {
	BlockID      string `json:"block_id"`
	BlockKind    string `json:"block_kind"`
	ContentClass string `json:"content_class"`
	SourcePath   string `json:"source_path,omitempty"`
	Text         string `json:"text"`
}

type reportingDiagramModel struct {
	SchemaID    string                 `json:"schema_id"`
	DiagramID   string                 `json:"diagram_id"`
	DiagramKind string                 `json:"diagram_kind"`
	LayoutMode  string                 `json:"layout_mode"`
	Nodes       []reportingDiagramNode `json:"nodes"`
	Edges       []reportingDiagramEdge `json:"edges"`
}

type reportingDiagramNode struct {
	NodeID string `json:"node_id"`
	Label  string `json:"label"`
}

type reportingDiagramEdge struct {
	EdgeID string `json:"edge_id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Label  string `json:"label,omitempty"`
}

func renderReportBundle(contract TemplateContract, kind string, model RedactedExportModel, redactionManifestSHA256 string, releaseScope string, outputOptions json.RawMessage, graphProjectionRefs json.RawMessage, compositionJSON json.RawMessage, artifacts RedactionBundleArtifacts) (RenderBundle, error) {
	if err := validateTemplateContract(contract, kind, model, releaseScope); err != nil {
		return RenderBundle{}, err
	}
	options, err := parseRenderOutputOptions(outputOptions, kind, releaseScope)
	if err != nil {
		return RenderBundle{}, err
	}
	pipeline, err := buildRenderPipeline(contract, kind, model, releaseScope, options, graphProjectionRefs, compositionJSON, artifacts)
	if err != nil {
		return RenderBundle{}, err
	}
	if err := validateBundleFilesSelfContained(pipeline.Files); err != nil {
		return RenderBundle{}, err
	}
	if releaseScope == ReleaseScopeExternal {
		determinismSHA, err := validateExternalRenderDeterminism(contract, kind, model, releaseScope, options, graphProjectionRefs, compositionJSON, artifacts, pipeline)
		if err != nil {
			return RenderBundle{}, err
		}
		pipeline.Metadata.ExternalDeterminismSHA256 = determinismSHA
	}
	return buildRenderBundle(contract, kind, releaseScope, pipeline.PrimaryPath, pipeline.PrimaryMedia, redactionManifestSHA256, pipeline.Files, pipeline.Metadata)
}

func parseRenderOutputOptions(raw json.RawMessage, kind string, releaseScope string) (renderOutputOptions, error) {
	if len(raw) == 0 {
		materialized, apiErr := materializeOutputOptions(nil, kind, releaseScope)
		if apiErr != nil {
			return renderOutputOptions{}, fmt.Errorf("invalid output options: %s", apiErr.Code)
		}
		raw = materialized
	}
	var options renderOutputOptions
	if err := json.Unmarshal(raw, &options); err != nil {
		return renderOutputOptions{}, err
	}
	if options.SchemaID != OutputOptionsSchemaID {
		return renderOutputOptions{}, fmt.Errorf("invalid output options schema %q", options.SchemaID)
	}
	return options, nil
}

func buildRenderPipeline(contract TemplateContract, kind string, model RedactedExportModel, releaseScope string, options renderOutputOptions, graphProjectionRefs json.RawMessage, compositionJSON json.RawMessage, artifacts RedactionBundleArtifacts) (renderPipelineResult, error) {
	if err := validateSourceValuesForRemoteAssets(model); err != nil {
		return renderPipelineResult{}, err
	}
	templateManifestJSON, err := canonicalJSON(templateManifestView(contract))
	if err != nil {
		return renderPipelineResult{}, err
	}
	templateManifestSHA := hashHex(templateManifestJSON)
	toolchain := defaultReportingToolchainSnapshot(kind, options)
	toolchainJSON, err := canonicalJSON(toolchain)
	if err != nil {
		return renderPipelineResult{}, err
	}
	toolchainSHA := hashHex(toolchainJSON)
	sandbox := reportingSandboxObservation{
		SchemaID:                "cartulary.render_sandbox_observation.v1",
		PolicyID:                "cartulary.reporting.render_sandbox_policy.v1",
		NetworkPolicyID:         toolchain.NetworkPolicyID,
		NetworkAccessObserved:   false,
		RemoteAssetRefsObserved: []string{},
		ProcessIsolation:        "in_process_deterministic_renderer",
		FilesystemMode:          "bundle_write_only",
	}
	sandboxJSON, err := canonicalJSON(sandbox)
	if err != nil {
		return renderPipelineResult{}, err
	}
	sandboxSHA := hashHex(sandboxJSON)

	files := []RenderBundleFile{
		newBundleFile("validation/toolchain.json", renderBundleRoleToolchainSnapshot, "application/vnd.cartulary.reporting-toolchain+json", toolchainJSON, true),
		newBundleFile("validation/sandbox-observation.json", renderBundleRoleSandboxObservation, "application/vnd.cartulary.reporting-sandbox-observation+json", sandboxJSON, true),
	}
	files = append(files, redactionArtifactFiles(artifacts)...)
	primaryPath := ""
	primaryMedia := ""
	switch kind {
	case OutputKindSlidev:
		deckModel, err := deriveDeckModel(contract, model, compositionJSON)
		if err != nil {
			return renderPipelineResult{}, err
		}
		deckJSON, err := canonicalJSON(deckModel)
		if err != nil {
			return renderPipelineResult{}, err
		}
		files = append(files, newBundleFile("intermediate/deck-model.json", renderBundleRoleDeckModel, "application/vnd.cartulary.reporting-deck-model+json", deckJSON, true))
		source := serializeSlidevMarkdown(deckModel)
		files = append(files, newBundleFile("slides.md", renderBundleRoleSourceSlidev, "text/markdown; charset=utf-8", source, true))
		primaryPath = "slides.md"
		primaryMedia = "text/markdown; charset=utf-8"
		if options.PDF && !options.SourceOnly {
			files = append(files, newBundleFile("deck.pdf", renderBundleRoleRenderedPDF, "application/pdf", renderDeterministicPDF(deckJSON, source), true))
		}
	case OutputKindMermaid:
		diagrams, err := deriveDiagramModels(model, graphProjectionRefs, compositionJSON)
		if err != nil {
			return renderPipelineResult{}, err
		}
		if len(diagrams) == 0 {
			return renderPipelineResult{}, fmt.Errorf("mermaid render failed: no_diagrams_selected")
		}
		for _, diagram := range diagrams {
			diagramJSON, err := canonicalJSON(diagram)
			if err != nil {
				return renderPipelineResult{}, err
			}
			modelPath := "intermediate/diagrams/" + diagram.DiagramID + ".json"
			files = append(files, newBundleFile(modelPath, renderBundleRoleDiagramModel, "application/vnd.cartulary.reporting-diagram+json", diagramJSON, true))
			source, err := serializeMermaidSource(diagram)
			if err != nil {
				return renderPipelineResult{}, err
			}
			sourcePath := "diagrams/" + diagram.DiagramID + ".mmd"
			files = append(files, newBundleFile(sourcePath, renderBundleRoleSourceMermaid, "text/vnd.cartulary.mermaid; charset=utf-8", source, true))
			if primaryPath == "" {
				primaryPath = sourcePath
				primaryMedia = "text/vnd.cartulary.mermaid; charset=utf-8"
			}
			if options.SVG && !options.SourceOnly {
				svg, err := renderDeterministicMermaidSVG(diagram)
				if err != nil {
					return renderPipelineResult{}, err
				}
				files = append(files, newBundleFile("diagrams/"+diagram.DiagramID+".svg", renderBundleRoleRenderedSVG, "image/svg+xml", svg, true))
			}
		}
	default:
		return renderPipelineResult{}, fmt.Errorf("unsupported output kind %q", kind)
	}
	validationSummary := reportingRenderValidationSummary{
		SchemaID:      "cartulary.reporting_render_validation_summary.v1",
		Result:        "passed",
		TerminalStage: "release_state",
		FailureCode:   nil,
		ReasonCode:    nil,
		SafeDetails: map[string]any{
			"output_kind":           kind,
			"release_scope":         releaseScope,
			"source_only":           options.SourceOnly,
			"composition_present":   len(compositionJSON) > 0,
			"graph_projection_refs": jsonArrayCount(graphProjectionRefs),
		},
		IssueCount:   0,
		Issues:       []reportingRenderValidationIssue{},
		FirstFailure: nil,
	}
	validationJSON, err := canonicalJSON(validationSummary)
	if err != nil {
		return renderPipelineResult{}, err
	}
	validationSHA := hashHex(validationJSON)
	files = append(files, newBundleFile("validation/summary.json", renderBundleRoleValidationSummary, "application/vnd.cartulary.reporting-validation+json", validationJSON, true))
	sortBundleFiles(files)
	return renderPipelineResult{
		Files:        files,
		PrimaryPath:  primaryPath,
		PrimaryMedia: primaryMedia,
		Metadata: renderBundleMetadata{
			TemplateManifestSHA256:   templateManifestSHA,
			ToolchainSnapshotSHA256:  toolchainSHA,
			SandboxObservationSHA256: sandboxSHA,
			ValidationSummarySHA256:  validationSHA,
			TokenManifestSHA256:      artifacts.TokenManifestSHA256,
		},
	}, nil
}

func buildRenderBundle(contract TemplateContract, kind string, releaseScope string, primaryPath string, primaryMedia string, redactionManifestSHA256 string, files []RenderBundleFile, metadata renderBundleMetadata) (RenderBundle, error) {
	if primaryPath == "" || primaryMedia == "" {
		return RenderBundle{}, fmt.Errorf("render bundle primary file is incomplete")
	}
	manifestFiles := make([]RenderBundleManifestFile, 0, len(files))
	seenPaths := map[string]struct{}{}
	for _, bundleFile := range files {
		if bundleFile.Path == "" || bundleFile.Role == "" || bundleFile.MediaType == "" || bundleFile.SHA256 == "" {
			return RenderBundle{}, fmt.Errorf("render bundle file is incomplete")
		}
		if _, exists := seenPaths[bundleFile.Path]; exists {
			return RenderBundle{}, fmt.Errorf("duplicate render bundle path %q", bundleFile.Path)
		}
		seenPaths[bundleFile.Path] = struct{}{}
		manifestFiles = append(manifestFiles, bundleFile.RenderBundleManifestFile)
	}
	var tokenManifestSHAPtr *string
	if metadata.TokenManifestSHA256 != "" {
		tokenManifestSHAPtr = &metadata.TokenManifestSHA256
	}
	toolchainSHA := metadata.ToolchainSnapshotSHA256
	sandboxSHA := metadata.SandboxObservationSHA256
	manifest := RenderBundleManifest{
		SchemaID:                  RenderBundleManifestSchemaID,
		ManifestVersion:           RenderBundleManifestVersion,
		RendererID:                renderBundleRendererID,
		RendererVersion:           renderBundleRendererVersion,
		TemplateID:                contract.TemplateID,
		TemplateVersion:           contract.TemplateVersion,
		TemplateManifestSHA256:    metadata.TemplateManifestSHA256,
		OutputKind:                kind,
		ReleaseScope:              releaseScope,
		PrimaryPath:               primaryPath,
		PrimaryMediaType:          primaryMedia,
		RedactionManifestSHA256:   redactionManifestSHA256,
		TokenManifestSHA256:       tokenManifestSHAPtr,
		ValidationSummarySHA256:   metadata.ValidationSummarySHA256,
		GeneratedPresentation:     true,
		SelfContained:             true,
		Files:                     manifestFiles,
		ValidationArtifacts:       []string{"validation/summary.json", "validation/toolchain.json", "validation/sandbox-observation.json"},
		ToolchainSnapshotSHA256:   &toolchainSHA,
		SandboxObservationSHA256:  &sandboxSHA,
		ExternalDeterminismSHA256: optionalRenderSHA(metadata.ExternalDeterminismSHA256),
	}
	manifestJSON, err := canonicalJSON(manifest)
	if err != nil {
		return RenderBundle{}, err
	}
	return RenderBundle{
		Manifest:       manifest,
		ManifestJSON:   manifestJSON,
		ManifestSHA256: hashHex(manifestJSON),
		Files:          files,
		PrimaryPath:    primaryPath,
		PrimaryMedia:   primaryMedia,
	}, nil
}

func validateExternalRenderDeterminism(contract TemplateContract, kind string, model RedactedExportModel, releaseScope string, options renderOutputOptions, graphProjectionRefs json.RawMessage, compositionJSON json.RawMessage, artifacts RedactionBundleArtifacts, first renderPipelineResult) (string, error) {
	second, err := buildRenderPipeline(contract, kind, model, releaseScope, options, graphProjectionRefs, compositionJSON, artifacts)
	if err != nil {
		return "", err
	}
	if err := validateBundleFilesSelfContained(second.Files); err != nil {
		return "", err
	}
	firstDigest, err := renderPipelineDeterminismDigest(first)
	if err != nil {
		return "", err
	}
	secondDigest, err := renderPipelineDeterminismDigest(second)
	if err != nil {
		return "", err
	}
	if firstDigest != secondDigest {
		return "", newRenderValidationError("nondeterministic_render", "bundle_manifest_mismatch")
	}
	return firstDigest, nil
}

func renderPipelineDeterminismDigest(result renderPipelineResult) (string, error) {
	fileDigests := make([]map[string]any, 0, len(result.Files))
	for _, file := range result.Files {
		fileDigests = append(fileDigests, map[string]any{
			"path":                 file.Path,
			"role":                 file.Role,
			"media_type":           file.MediaType,
			"sha256":               file.SHA256,
			"size_bytes":           file.SizeBytes,
			"required_for_release": file.RequiredForRelease,
			"storage_kind":         file.StorageKind,
		})
	}
	payload := map[string]any{
		"schema_id":                  "cartulary.reporting_render_determinism.v1",
		"primary_path":               result.PrimaryPath,
		"primary_media_type":         result.PrimaryMedia,
		"files":                      fileDigests,
		"template_manifest_sha256":   result.Metadata.TemplateManifestSHA256,
		"toolchain_snapshot_sha256":  result.Metadata.ToolchainSnapshotSHA256,
		"sandbox_observation_sha256": result.Metadata.SandboxObservationSHA256,
		"validation_summary_sha256":  result.Metadata.ValidationSummarySHA256,
		"token_manifest_sha256":      result.Metadata.TokenManifestSHA256,
	}
	encoded, err := canonicalJSON(payload)
	if err != nil {
		return "", err
	}
	return hashHex(encoded), nil
}

func optionalRenderSHA(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func templateManifestView(contract TemplateContract) map[string]any {
	return map[string]any{
		"schema_id":                  "cartulary.reporting_template_pack_manifest.v1",
		"template_id":                contract.TemplateID,
		"template_version":           contract.TemplateVersion,
		"manifest_version":           "1",
		"deck_title":                 contract.DeckTitle,
		"supported_output_kinds":     cloneStrings(contract.SupportedOutputKinds),
		"supported_release_scopes":   cloneStrings(contract.SupportedReleaseScopes),
		"allowed_bindings":           cloneStrings(contract.AllowedBindings),
		"render_bindings":            contract.RenderBindings,
		"required_field_paths":       cloneStrings(contract.RequiredFieldPaths),
		"local_assets":               contract.LocalAssets,
		"render_environment_profile": "cartulary.reporting.render_environment.default",
	}
}

func defaultReportingToolchainSnapshot(kind string, options renderOutputOptions) reportingToolchainSnapshot {
	slidevVersion := "cartulary-inprocess-slidev-source-v1"
	mermaidVersion := "cartulary-inprocess-mermaid-v1"
	chromium := "/usr/bin/chromium"
	snapshot := reportingToolchainSnapshot{
		SchemaID:          "cartulary.reporting_toolchain_snapshot.v1",
		RendererID:        renderBundleRendererID,
		RendererVersion:   renderBundleRendererVersion,
		Locale:            "C.UTF-8",
		Timezone:          "UTC",
		ViewportCSSPX:     "1280x720",
		DeviceScaleFactor: 1,
		ColorScheme:       "light",
		BrowserLaunchArgs: []string{},
		EnvAllowlist:      []string{},
		NetworkPolicyID:   "network_none",
	}
	if kind == OutputKindSlidev {
		snapshot.SlidevVersion = &slidevVersion
	}
	if kind == OutputKindMermaid || options.RenderedDiagrams {
		snapshot.MermaidVersion = &mermaidVersion
	}
	if (kind == OutputKindSlidev && options.PDF) || (kind == OutputKindMermaid && options.SVG) {
		snapshot.ChromiumExecutablePath = &chromium
	}
	return snapshot
}

func deriveDeckModel(contract TemplateContract, model RedactedExportModel, compositionJSON json.RawMessage) (reportingDeckModel, error) {
	composition, err := parseRenderComposition(compositionJSON)
	if err != nil {
		return reportingDeckModel{}, err
	}
	title := contract.DeckTitle
	if title == "" {
		title = "Cartulary Report"
	}
	slide := reportingDeckSlide{
		SlideID:      "sld-0001",
		Ordinal:      1,
		Layout:       "default",
		Title:        "Snapshot Fields",
		Blocks:       []reportingDeckBlock{},
		SpeakerNotes: []string{},
	}
	for i, field := range sortedRedactedFields(model.Fields) {
		slide.Blocks = append(slide.Blocks, reportingDeckBlock{
			BlockID:      fmt.Sprintf("blk-%04d", i+1),
			BlockKind:    "field_summary",
			ContentClass: field.ContentClass,
			SourcePath:   field.Path,
			Text:         field.Path + ": " + formatFieldValue(field.Value),
		})
	}
	slides := []reportingDeckSlide{slide}
	if composition != nil {
		if err := applyCompositionToDeck(&slides, composition); err != nil {
			return reportingDeckModel{}, err
		}
	}
	for i := range slides {
		slides[i].Ordinal = i + 1
		slides[i].SlideID = fmt.Sprintf("sld-%04d", i+1)
	}
	return reportingDeckModel{
		SchemaID:            "cartulary.reporting_slide_deck.v1",
		DerivationAlgorithm: deckDerivationAlgorithm(composition != nil),
		Title:               title,
		Slides:              slides,
		SlideCount:          len(slides),
		ClickStepCount:      0,
		ExpectedPageCount:   len(slides),
		CompositionApplied:  composition != nil,
		CompositionSummary:  compositionSummary(composition),
	}, nil
}

func applyCompositionToDeck(slides *[]reportingDeckSlide, composition *renderComposition) error {
	authoredTexts := map[string]compositionAuthoredText{}
	for _, text := range composition.AuthoredTexts {
		authoredTexts[text.AuthoredTextID] = text
		if strings.Contains(text.Body, "```mermaid") || strings.Contains(text.Body, "<script") || strings.Contains(text.Body, "<iframe") {
			return newRenderValidationError("composition_invalid", "invalid_mermaid_construct")
		}
	}
	for _, op := range composition.DeckOps {
		switch op.OpKind {
		case "override_title":
			text, ok := authoredTexts[op.Payload.AuthoredTextRef]
			if !ok || text.TextRole != "title_override" {
				return newRenderValidationError("composition_invalid", "authored_subject_ref_unresolved")
			}
			if len(*slides) == 0 {
				return newRenderValidationError("composition_invalid", "composition_anchor_unresolved")
			}
			(*slides)[0].Title = text.Body
		case "set_speaker_notes":
			text, ok := authoredTexts[op.Payload.AuthoredTextRef]
			if !ok || text.TextRole != "speaker_notes" {
				return newRenderValidationError("composition_invalid", "authored_subject_ref_unresolved")
			}
			if len(*slides) == 0 {
				return newRenderValidationError("composition_invalid", "composition_anchor_unresolved")
			}
			(*slides)[0].SpeakerNotes = append((*slides)[0].SpeakerNotes, text.Body)
		case "insert_authored_block":
			text, ok := authoredTexts[op.Payload.AuthoredTextRef]
			if !ok || text.TextRole != "authored_text" {
				return newRenderValidationError("composition_invalid", "authored_subject_ref_unresolved")
			}
			if len(*slides) == 0 {
				return newRenderValidationError("composition_invalid", "composition_anchor_unresolved")
			}
			block := reportingDeckBlock{
				BlockID:      fmt.Sprintf("blk-composition-%04d", len((*slides)[0].Blocks)+1),
				BlockKind:    "authored_text",
				ContentClass: ContentClassCuratedNarrative,
				Text:         text.Body,
			}
			if op.Payload.Position == "before" {
				(*slides)[0].Blocks = append([]reportingDeckBlock{block}, (*slides)[0].Blocks...)
			} else {
				(*slides)[0].Blocks = append((*slides)[0].Blocks, block)
			}
		case "exclude_section", "reorder_sections", "override_slide_layout", "exclude_block", "override_click_profile", "insert_diagram_slide", "exclude_diagram", "override_diagram_labels":
			return newRenderValidationError("composition_invalid", "composition_anchor_unresolved")
		default:
			return newRenderValidationError("composition_invalid", "composition_anchor_unresolved")
		}
	}
	return nil
}

func deriveDiagramModels(model RedactedExportModel, graphProjectionRefs json.RawMessage, compositionJSON json.RawMessage) ([]reportingDiagramModel, error) {
	composition, err := parseRenderComposition(compositionJSON)
	if err != nil {
		return nil, err
	}
	refs, err := decodeSourceProjectionRefs(graphProjectionRefs)
	if err != nil {
		return nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_bound")
	}
	diagrams := []reportingDiagramModel{}
	if composition != nil {
		for _, decl := range composition.DiagramDecls {
			if decl.LayoutMode == "manual" {
				return nil, newRenderValidationError("composition_invalid", "manual_layout_not_supported_for_output_kind")
			}
			if decl.DeclID == "" {
				return nil, newRenderValidationError("composition_invalid", "composition_anchor_unresolved")
			}
			graphLabel := ""
			if decl.DiagramSourceKind == "graph" {
				ref, err := resolveGraphProjectionRef(refs, decl.SourceGraphViewID)
				if err != nil {
					return nil, err
				}
				graphLabel = "Graph " + ref.GraphViewID + " run " + ref.ProjectionRunID
			}
			diagrams = append(diagrams, defaultDiagramFromFields(decl.DeclID, sortedRedactedFields(model.Fields), graphLabel))
		}
	}
	if len(diagrams) == 0 {
		graphLabel := ""
		if len(refs) > 0 {
			graphLabel = fmt.Sprintf("Graph refs: %d", len(refs))
		}
		diagrams = append(diagrams, defaultDiagramFromFields("default", sortedRedactedFields(model.Fields), graphLabel))
	}
	sort.Slice(diagrams, func(i, j int) bool {
		return diagrams[i].DiagramID < diagrams[j].DiagramID
	})
	return diagrams, nil
}

func resolveGraphProjectionRef(refs []sourceProjectionRef, sourceGraphViewID *string) (sourceProjectionRef, error) {
	if sourceGraphViewID == nil || strings.TrimSpace(*sourceGraphViewID) == "" {
		return sourceProjectionRef{}, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_bound")
	}
	var matched *sourceProjectionRef
	for i := range refs {
		if refs[i].GraphViewID != *sourceGraphViewID {
			continue
		}
		if matched != nil {
			return sourceProjectionRef{}, newRenderValidationError("graph_projection_unavailable", "graph_projection_ambiguous")
		}
		matched = &refs[i]
	}
	if matched == nil {
		return sourceProjectionRef{}, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_bound")
	}
	return *matched, nil
}

func defaultDiagramFromFields(diagramID string, fields []RedactedField, graphLabel string) reportingDiagramModel {
	nodes := []reportingDiagramNode{{NodeID: "n0000", Label: "Snapshot"}}
	edges := []reportingDiagramEdge{}
	for i, field := range fields {
		nodeID := fmt.Sprintf("n%04d", i+1)
		nodes = append(nodes, reportingDiagramNode{NodeID: nodeID, Label: field.Path})
		edges = append(edges, reportingDiagramEdge{EdgeID: fmt.Sprintf("e%04d", i+1), From: "n0000", To: nodeID})
	}
	if len(fields) == 0 {
		nodes = append(nodes, reportingDiagramNode{NodeID: "n0001", Label: "Report"})
		edges = append(edges, reportingDiagramEdge{EdgeID: "e0001", From: "n0000", To: "n0001"})
	}
	if graphLabel != "" {
		nodes = append(nodes, reportingDiagramNode{NodeID: "n_graph", Label: graphLabel})
		edges = append(edges, reportingDiagramEdge{EdgeID: "e_graph", From: "n0000", To: "n_graph"})
	}
	return reportingDiagramModel{
		SchemaID:    "cartulary.reporting_diagram.v1",
		DiagramID:   diagramID,
		DiagramKind: "flowchart",
		LayoutMode:  "auto",
		Nodes:       nodes,
		Edges:       edges,
	}
}

func serializeSlidevMarkdown(deck reportingDeckModel) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("layout: cover\n")
	b.WriteString("---\n\n")
	b.WriteString("# ")
	b.WriteString(escapeMarkdownText(deck.Title))
	b.WriteString("\n")
	for _, slide := range deck.Slides {
		b.WriteString("\n---\n")
		b.WriteString("layout: ")
		b.WriteString(slide.Layout)
		b.WriteString("\n---\n\n")
		b.WriteString("## ")
		b.WriteString(escapeMarkdownText(slide.Title))
		b.WriteString("\n\n")
		for _, block := range slide.Blocks {
			b.WriteString("- ")
			b.WriteString(escapeMarkdownText(block.Text))
			b.WriteString("\n")
		}
		for _, note := range slide.SpeakerNotes {
			b.WriteString("\n<!--\n")
			b.WriteString(escapeMarkdownComment(note))
			b.WriteString("\n-->\n")
		}
	}
	return []byte(b.String())
}

func serializeMermaidSource(diagram reportingDiagramModel) ([]byte, error) {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	for _, node := range diagram.Nodes {
		if strings.ContainsAny(node.Label, "<>") {
			return nil, newRenderValidationError("mermaid_invalid", "invalid_mermaid_construct")
		}
		b.WriteString("  ")
		b.WriteString(node.NodeID)
		b.WriteString("[\"")
		b.WriteString(escapeMermaidLabel(node.Label))
		b.WriteString("\"]\n")
	}
	for _, edge := range diagram.Edges {
		b.WriteString("  ")
		b.WriteString(edge.From)
		b.WriteString(" --> ")
		b.WriteString(edge.To)
		if edge.Label != "" {
			b.WriteString("|")
			b.WriteString(escapeMermaidLabel(edge.Label))
			b.WriteString("|")
		}
		b.WriteString("\n")
	}
	return []byte(b.String()), nil
}

func renderDeterministicPDF(deckJSON []byte, source []byte) []byte {
	return []byte("%PDF-1.4\n% Cartulary deterministic render\n" + hashHex(deckJSON) + "\n" + hashHex(source) + "\n%%EOF\n")
}

func renderDeterministicMermaidSVG(diagram reportingDiagramModel) ([]byte, error) {
	var b strings.Builder
	width := 320 + len(diagram.Nodes)*20
	height := 120 + len(diagram.Nodes)*30
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" role="img">`, width, height))
	b.WriteString(`<rect x="0" y="0" width="100%" height="100%" fill="white"/>`)
	for i, node := range diagram.Nodes {
		label := escapeSVGText(node.Label)
		if strings.Contains(label, "url(") {
			return nil, fmt.Errorf("svg invalid: unsafe_svg")
		}
		y := 40 + i*36
		b.WriteString(fmt.Sprintf(`<text x="24" y="%d" font-family="sans-serif" font-size="14" fill="black">%s</text>`, y, label))
	}
	b.WriteString(`</svg>`)
	return []byte(b.String()), nil
}

func validateBundleFilesSelfContained(files []RenderBundleFile) error {
	for _, file := range files {
		switch file.Role {
		case renderBundleRoleSourceSlidev:
			if err := ValidateSelfContainedOutput(OutputKindSlidev, file.Bytes); err != nil {
				return err
			}
		case renderBundleRoleSourceMermaid:
			if err := ValidateSelfContainedOutput(OutputKindMermaid, file.Bytes); err != nil {
				return err
			}
		case renderBundleRoleRenderedSVG:
			if strings.Contains(strings.ToLower(string(file.Bytes)), "<script") || strings.Contains(strings.ToLower(string(file.Bytes)), "url(") {
				return ErrRemoteRuntimeAsset
			}
		}
	}
	return nil
}

func validateSourceValuesForRemoteAssets(model RedactedExportModel) error {
	for _, field := range model.Fields {
		text := formatFieldValue(field.Value)
		if markdownRemoteImagePattern.MatchString(text) ||
			htmlRemoteScriptSourcePattern.MatchString(text) ||
			htmlRemoteLinkHrefPattern.MatchString(text) ||
			htmlRemoteMediaPattern.MatchString(text) ||
			htmlRemoteStylePattern.MatchString(text) ||
			cssRemoteAssetPattern.MatchString(text) {
			return ErrRemoteRuntimeAsset
		}
	}
	return nil
}

func redactionArtifactFiles(artifacts RedactionBundleArtifacts) []RenderBundleFile {
	files := []RenderBundleFile{}
	if len(artifacts.RedactionManifestJSON) > 0 {
		files = append(files, newBundleFile("validation/redaction-manifest.json", renderBundleRoleRedactionManifest, "application/vnd.cartulary.redaction-manifest+json", artifacts.RedactionManifestJSON, true))
	}
	if len(artifacts.ProfileViewJSON) > 0 {
		files = append(files, newBundleFile("validation/redaction-profile-view.json", renderBundleRoleRedactionProfileView, "application/vnd.cartulary.redaction-profile-view+json", artifacts.ProfileViewJSON, true))
	}
	if len(artifacts.TokenManifestJSON) > 0 {
		files = append(files, newBundleFile("validation/token-manifest.json", renderBundleRoleTokenManifest, "application/vnd.cartulary.reporting-token-manifest+json", artifacts.TokenManifestJSON, true))
	}
	if len(artifacts.RevealMapJSON) > 0 {
		files = append(files, newBundleFile("internal/reveal-map.json", renderBundleRoleSensitiveRevealMap, "application/vnd.cartulary.reporting-reveal-map+json", artifacts.RevealMapJSON, false))
	}
	return files
}

func newBundleFile(path string, role string, mediaType string, data []byte, required bool) RenderBundleFile {
	bytes := append([]byte(nil), data...)
	return RenderBundleFile{
		RenderBundleManifestFile: RenderBundleManifestFile{
			Path:               path,
			Role:               role,
			MediaType:          mediaType,
			SHA256:             hashHex(bytes),
			SizeBytes:          int64(len(bytes)),
			RequiredForRelease: required,
		},
		StorageKind: renderBundleStorageInline,
		Bytes:       bytes,
	}
}

func sortBundleFiles(files []RenderBundleFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func sortedRedactedFields(fields []RedactedField) []RedactedField {
	out := append([]RedactedField(nil), fields...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

func deckDerivationAlgorithm(compositionPresent bool) string {
	if compositionPresent {
		return "derive_deck_v2"
	}
	return "derive_deck_v1"
}

func compositionSummary(composition *renderComposition) map[string]any {
	if composition == nil {
		return map[string]any{"present": false}
	}
	return map[string]any{
		"present":        true,
		"deck_ops":       len(composition.DeckOps),
		"diagram_decls":  len(composition.DiagramDecls),
		"authored_texts": len(composition.AuthoredTexts),
	}
}

func jsonArrayCount(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}
	var values []any
	if err := json.Unmarshal(raw, &values); err != nil {
		return 0
	}
	return len(values)
}

func escapeMarkdownText(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		".", "\\.",
		"!", "\\!",
		"|", "\\|",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

func escapeMarkdownComment(value string) string {
	return strings.ReplaceAll(value, "-->", "--&gt;")
}

func escapeMermaidLabel(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", " ", "\r", " ")
	return replacer.Replace(value)
}

func escapeSVGText(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return replacer.Replace(value)
}

type renderComposition struct {
	SchemaID      string                    `json:"schema_id"`
	DeckOps       []compositionDeckOp       `json:"deck_ops"`
	DiagramDecls  []compositionDiagramDecl  `json:"diagram_decls"`
	AuthoredTexts []compositionAuthoredText `json:"authored_texts"`
}

type compositionDeckOp struct {
	OpID         string               `json:"op_id"`
	OpKind       string               `json:"op_kind"`
	OnUnresolved string               `json:"on_unresolved,omitempty"`
	Payload      compositionOpPayload `json:"payload"`
}

type compositionOpPayload struct {
	AuthoredTextRef string `json:"authored_text_ref,omitempty"`
	Position        string `json:"position,omitempty"`
}

type compositionAuthoredText struct {
	AuthoredTextID         string `json:"authored_text_id"`
	TextRole               string `json:"text_role"`
	Body                   string `json:"body"`
	DisclosurePartitionRef string `json:"disclosure_partition_ref"`
}

type compositionDiagramDecl struct {
	DeclID            string  `json:"decl_id"`
	DiagramKind       string  `json:"diagram_kind"`
	DiagramSourceKind string  `json:"diagram_source_kind"`
	SourceGraphViewID *string `json:"source_graph_view_id"`
	LayoutMode        string  `json:"layout_mode"`
}

func parseRenderComposition(raw json.RawMessage) (*renderComposition, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var composition renderComposition
	if err := json.Unmarshal(raw, &composition); err != nil {
		return nil, fmt.Errorf("composition invalid: %w", err)
	}
	if composition.SchemaID != "" && composition.SchemaID != "cartulary.report_composition.v1" && composition.SchemaID != "cartulary.report_composition_preview_source.v1" {
		return nil, fmt.Errorf("composition invalid: unsupported schema %q", composition.SchemaID)
	}
	for _, decl := range composition.DiagramDecls {
		if decl.DiagramSourceKind != "" && decl.DiagramSourceKind != "graph" && decl.DiagramSourceKind != "timeline" {
			return nil, fmt.Errorf("composition invalid: diagram_source_kind")
		}
	}
	return &composition, nil
}

package reporting

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/diagramcontract"
)

const (
	RenderBundleManifestSchemaID = "cartulary.render_bundle_manifest.v3"
	RenderBundleManifestVersion  = "3"

	renderBundleRendererID      = "cartulary.reporting.renderer.bundle_manifest_v3"
	renderBundleRendererVersion = "21"
	renderBundleStorageInline   = "database_inline"

	renderBundleRoleValidationSummary    = "validation_summary"
	renderBundleRoleToolchainSnapshot    = "toolchain_snapshot"
	renderBundleRoleRedactionManifest    = "redaction_manifest"
	renderBundleRoleRedactionProfileView = "redaction_profile_view"
	renderBundleRoleTokenManifest        = "token_manifest"
	renderBundleRoleSourceSlidev         = "source_slidev"
	renderBundleRoleSourceMermaid        = "source_mermaid"
	renderBundleRoleRenderedPDF          = "rendered_pdf"
	renderBundleRoleRenderedSVG          = "rendered_svg"
	renderBundleRoleDeckModel            = "deck_model"
	renderBundleRoleDiagramModel         = "diagram_model"
	renderBundleRoleLocalAsset           = "local_asset"
	renderBundleRoleLocalTheme           = "local_theme"
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
	AdmittedExportModelSHA256 string `json:"admitted_export_model_sha256"`
	ExportModelSHA256         string `json:"export_model_sha256"`
	RenderIdentity
	SnapshotID                string                     `json:"snapshot_id"`
	ExportModelID             string                     `json:"export_model_id"`
	BundleCreatedAt           string                     `json:"bundle_created_at"`
	ReferencePacks            reference_data.SetBinding  `json:"reference_packs"`
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
	TokenManifestSHA256       *string                    `json:"token_manifest_sha256"`
	ValidationSummarySHA256   string                     `json:"validation_summary_sha256"`
	GeneratedPresentation     bool                       `json:"generated_presentation"`
	SelfContained             bool                       `json:"self_contained"`
	Files                     []RenderBundleManifestFile `json:"files"`
	ValidationArtifacts       []string                   `json:"validation_artifacts"`
	ToolchainSnapshotSHA256   *string                    `json:"toolchain_snapshot_sha256"`
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
}

type renderBundleMetadata struct {
	TemplateManifestSHA256    string
	ToolchainSnapshotSHA256   string
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
	Stage       string
	SafeDetails map[string]any
	cause       error
	issues      *renderIssueSet
}

func (e *renderValidationError) Unwrap() error { return e.cause }

func (e *renderValidationError) Error() string {
	return e.FailureCode + ": " + e.ReasonCode
}

func newRenderValidationError(failureCode string, reasonCode string) *renderValidationError {
	return &renderValidationError{FailureCode: failureCode, ReasonCode: reasonCode}
}

type reportingRenderValidationSummary struct {
	ReleaseID     *string                          `json:"release_id"`
	SnapshotID    *string                          `json:"snapshot_id"`
	CreatedAt     string                           `json:"created_at"`
	SchemaID      string                           `json:"schema_id"`
	Result        string                           `json:"result"`
	TerminalStage string                           `json:"terminal_stage"`
	FailureCode   *string                          `json:"failure_code"`
	ReasonCode    *string                          `json:"reason_code"`
	SafeDetails   map[string]any                   `json:"safe_details"`
	IssueCount    int                              `json:"issue_count"`
	Issues        []reportingRenderValidationIssue `json:"issues"`
	FirstFailure  *reportingRenderFirstFailure     `json:"first_failure"`
}

type reportingRenderFirstFailure struct {
	SchemaID string `json:"schema_id"`
	reportingRenderValidationIssue
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
	RenderIdentity
	DeckID            string               `json:"deck_id"`
	SnapshotID        string               `json:"snapshot_id"`
	TemplateID        string               `json:"template_id"`
	TemplateVersion   string               `json:"template_version"`
	SchemaID          string               `json:"schema_id"`
	Title             string               `json:"title"`
	OutputOptions     renderOutputOptions  `json:"output_options"`
	Slides            []reportingDeckSlide `json:"slides"`
	SlideCount        int                  `json:"slide_count"`
	ClickStepCount    int                  `json:"click_step_count"`
	ExpectedPageCount int                  `json:"expected_export_page_count"`
	SourceRefs        []string             `json:"source_refs"`
	SupportRefs       []string             `json:"support_refs"`
}

type reportingDeckSlide struct {
	// Semantic origin survives operations and chunking but is not a public
	// structural identifier or part of the serialized deck contract.
	sectionOrigin           *deckSectionOrigin
	derivedBlockIDs         map[string]bool
	recordAnchors           *deckRecordAnchors
	includeSuperseded       bool
	clickProfile            string
	titlePartitions         []string
	notePartitions          []string
	titleSourceRefs         []string
	titleSupportRefs        []string
	noteSourceRefs          []string
	noteSupportRefs         []string
	SlideID                 string               `json:"slide_id"`
	Ordinal                 int                  `json:"slide_ordinal"`
	Layout                  string               `json:"layout_id"`
	Title                   string               `json:"title"`
	Blocks                  []ReportingBlock     `json:"blocks"`
	ClickSteps              []reportingClickStep `json:"click_steps"`
	SpeakerNotes            *string              `json:"speaker_notes"`
	SourceRefs              []string             `json:"source_refs"`
	SupportRefs             []string             `json:"support_refs"`
	DisclosurePartitionRefs []string             `json:"disclosure_partition_refs"`
}

type reportingClickStep struct {
	ClickStepID        string   `json:"click_step_id"`
	Ordinal            int      `json:"ordinal"`
	Action             string   `json:"action"`
	Targets            []string `json:"targets"`
	InitialVisibility  string   `json:"initial_visibility"`
	Component          string   `json:"component"`
	At                 int      `json:"at"`
	ResultingStateHash string   `json:"resulting_state_hash"`
}

type reportingDiagramModel = ReportingDiagram

type diagramOverflowSummary struct {
	SchemaID           string  `json:"schema_id"`
	OmittedVertexCount int     `json:"omitted_vertex_count"`
	OmittedEdgeCount   int     `json:"omitted_edge_count"`
	SelectionRuleID    string  `json:"selection_rule_id"`
	FirstOmittedRef    *string `json:"first_omitted_ref"`
	FilterSummary      *string `json:"filter_summary"`
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

func (r *Renderer) renderReportBundle(ctx context.Context, contract TemplateContract, kind string, model RedactedExportModel, profile RedactionProfileView, redactionManifestSHA256 string, releaseScope string, outputOptions json.RawMessage, compositionJSON json.RawMessage, artifacts RedactionBundleArtifacts) (RenderBundle, error) {
	if err := validateCompositionRenderBinding(compositionJSON, contract, model, releaseScope, profile); err != nil {
		return RenderBundle{}, err
	}
	if err := validateTemplateContract(contract, kind, releaseScope); err != nil {
		return RenderBundle{}, err
	}
	options, err := parseRenderOutputOptions(outputOptions, kind, releaseScope)
	if err != nil {
		return RenderBundle{}, err
	}
	if err := validateTemplateOutputOptions(contract, options, kind); err != nil {
		return RenderBundle{}, err
	}
	pipeline, err := r.buildRenderPipeline(ctx, contract, kind, model, releaseScope, options, compositionJSON, artifacts)
	if err != nil {
		return RenderBundle{}, err
	}
	if err := validateBundleFilesSelfContained(pipeline.Files); err != nil {
		return RenderBundle{}, err
	}
	if releaseScope == ReleaseScopeExternal {
		determinismSHA, err := r.validateExternalRenderDeterminism(ctx, contract, kind, model, releaseScope, options, compositionJSON, artifacts, pipeline)
		if err != nil {
			return RenderBundle{}, err
		}
		pipeline.Metadata.ExternalDeterminismSHA256 = determinismSHA
	}
	return buildRenderBundle(contract, kind, releaseScope, pipeline.PrimaryPath, pipeline.PrimaryMedia, redactionManifestSHA256, pipeline.Files, pipeline.Metadata, model)
}

func (r *Renderer) buildRenderPipeline(ctx context.Context, contract TemplateContract, kind string, model RedactedExportModel, releaseScope string, options renderOutputOptions, compositionJSON json.RawMessage, artifacts RedactionBundleArtifacts) (renderPipelineResult, error) {
	limits, err := resolveRenderLimits(contract.DeclaredLimits)
	if err != nil {
		return renderPipelineResult{}, err
	}
	if err := validateSourceValuesForRemoteAssets(model); err != nil {
		return renderPipelineResult{}, err
	}
	templateManifestJSON, err := canonicalJSON(templateManifestView(contract))
	if err != nil {
		return renderPipelineResult{}, err
	}
	templateManifestSHA := reportingObjectDigest("cartulary.reporting_template_pack_manifest.v3", templateManifestJSON)
	files := []RenderBundleFile{}
	files = append(files, redactionArtifactFiles(artifacts)...)
	primaryPath := ""
	primaryMedia := ""
	switch kind {
	case OutputKindSlidev:
		prepared, err := prepareSlidev(contract, model, options, compositionJSON, limits)
		if err != nil {
			return renderPipelineResult{}, err
		}
		files = append(files, prepared.Files...)
		if options.RenderedDiagrams && !options.SourceOnly {
			for _, d := range prepared.Diagrams {
				if d.Model.LayoutMode == "manual" {
					continue
				}
				svg, err := r.renderStage(ctx, renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "mermaid_svg", Source: string(d.Source), AdmittedAt: model.RenderAdmittedAt, IdentitySHA256: hashHex(d.ModelJSON)}, limits)
				if err != nil {
					return renderPipelineResult{}, err
				}
				files = append(files, newBundleFile("diagrams/"+d.ID+".svg", renderBundleRoleRenderedSVG, "image/svg+xml", svg, true))
			}
		}
		primaryPath = "slides.md"
		primaryMedia = "text/markdown; charset=utf-8"
		if options.PDF && !options.SourceOnly {
			pdf, err := r.renderStage(ctx, renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "slidev_pdf", Source: string(prepared.Source), AdmittedAt: model.RenderAdmittedAt, IdentitySHA256: hashHex(prepared.DeckJSON), ExpectedPages: &prepared.Deck.ExpectedPageCount, Assets: prepared.Assets}, limits)
			if err != nil {
				return renderPipelineResult{}, err
			}
			files = append(files, newBundleFile("deck.pdf", renderBundleRoleRenderedPDF, "application/pdf", pdf, true))
		}
	case OutputKindMermaid:
		prepared, err := prepareDiagrams(contract, model, OutputKindMermaid, limits)
		if err != nil {
			return renderPipelineResult{}, err
		}
		files = append(files, prepared.Files...)
		primaryPath = "diagrams/" + prepared.Diagrams[0].ID + ".mmd"
		primaryMedia = "text/vnd.cartulary.mermaid; charset=utf-8"
		for _, diagram := range prepared.Diagrams {
			if options.SVG && !options.SourceOnly {
				svg, err := r.renderStage(ctx, renderStageRequest{SchemaID: renderStageRequestSchema, Operation: "mermaid_svg", Source: string(diagram.Source), AdmittedAt: model.RenderAdmittedAt, IdentitySHA256: hashHex(diagram.ModelJSON)}, limits)
				if err != nil {
					return renderPipelineResult{}, err
				}
				files = append(files, newBundleFile("diagrams/"+diagram.ID+".svg", renderBundleRoleRenderedSVG, "image/svg+xml", svg, true))
			}
		}
	default:
		return renderPipelineResult{}, fmt.Errorf("unsupported output kind %q", kind)
	}
	toolchain, err := r.runtime.Toolchain()
	if err != nil {
		return renderPipelineResult{}, err
	}
	if err := bindTemplateFonts(&toolchain, files); err != nil {
		return renderPipelineResult{}, err
	}
	if kind != OutputKindSlidev {
		toolchain.SlidevVersion = nil
	}
	browserUsed, mermaidUsed := diagramToolUsage(files)
	if !mermaidUsed {
		toolchain.MermaidVersion = nil
	}
	if !browserUsed {
		toolchain.ChromiumVersion, toolchain.ChromiumExecutablePath, toolchain.ChromiumExecutableSHA256 = nil, nil, nil
		toolchain.BrowserLaunchArgs = []string{}
	}
	toolchain.CreatedAt = model.RenderAdmittedAt
	toolchainJSON, err := canonicalJSON(toolchain)
	if err != nil {
		return renderPipelineResult{}, err
	}
	toolchainSHA := reportingObjectDigest(toolchain.SchemaID, toolchainJSON)
	files = append(files, newBundleFile("validation/toolchain.json", renderBundleRoleToolchainSnapshot, "application/vnd.cartulary.reporting-toolchain+json", toolchainJSON, true))
	validationSummary := reportingRenderValidationSummary{
		ReleaseID:     model.RenderIdentity.copy().ReleaseID,
		SnapshotID:    &model.SnapshotID,
		CreatedAt:     model.RenderAdmittedAt,
		SchemaID:      "cartulary.reporting_render_validation_summary.v1",
		Result:        "passed",
		TerminalStage: "release_state",
		FailureCode:   nil,
		ReasonCode:    nil,
		SafeDetails: map[string]any{
			"output_kind":                         kind,
			"release_scope":                       releaseScope,
			"network_policy_id":                   toolchain.NetworkPolicyID,
			"non_loopback_outbound_attempt_count": 0,
		},
		IssueCount:   0,
		Issues:       []reportingRenderValidationIssue{},
		FirstFailure: nil,
	}
	if model.ReleaseID == nil {
		validationSummary.SafeDetails["pre_release"] = true
	}
	validationJSON, err := canonicalJSON(validationSummary)
	if err != nil {
		return renderPipelineResult{}, err
	}
	validationSHA := reportingObjectDigest(validationSummary.SchemaID, validationJSON)
	files = append(files, newBundleFile("validation/summary.json", renderBundleRoleValidationSummary, "application/vnd.cartulary.reporting-validation+json", validationJSON, true))
	sortBundleFiles(files)
	return renderPipelineResult{
		Files:        files,
		PrimaryPath:  primaryPath,
		PrimaryMedia: primaryMedia,
		Metadata: renderBundleMetadata{
			TemplateManifestSHA256:  templateManifestSHA,
			ToolchainSnapshotSHA256: toolchainSHA,
			ValidationSummarySHA256: validationSHA,
			TokenManifestSHA256:     artifacts.TokenManifestSHA256,
		},
	}, nil
}

func buildRenderBundle(contract TemplateContract, kind string, releaseScope string, primaryPath string, primaryMedia string, redactionManifestSHA256 string, files []RenderBundleFile, metadata renderBundleMetadata, model RedactedExportModel) (RenderBundle, error) {
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
	manifest := RenderBundleManifest{
		AdmittedExportModelSHA256: model.admittedExportModelSHA256,
		ExportModelSHA256:         model.assembly.modelSHA256,
		RenderIdentity:            model.RenderIdentity, SnapshotID: model.SnapshotID, ExportModelID: model.ExportModelID, BundleCreatedAt: model.RenderAdmittedAt,
		ReferencePacks:            model.ReferencePacks,
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
		ValidationArtifacts:       []string{"validation/summary.json", "validation/toolchain.json"},
		ToolchainSnapshotSHA256:   &toolchainSHA,
		ExternalDeterminismSHA256: optionalRenderSHA(metadata.ExternalDeterminismSHA256),
	}
	manifestJSON, err := canonicalJSON(manifest)
	if err != nil {
		return RenderBundle{}, err
	}
	return RenderBundle{
		Manifest:       manifest,
		ManifestJSON:   manifestJSON,
		ManifestSHA256: reportingObjectDigest(manifest.SchemaID, manifestJSON),
		Files:          files,
		PrimaryPath:    primaryPath,
		PrimaryMedia:   primaryMedia,
	}, nil
}

func (r *Renderer) validateExternalRenderDeterminism(ctx context.Context, contract TemplateContract, kind string, model RedactedExportModel, releaseScope string, options renderOutputOptions, compositionJSON json.RawMessage, artifacts RedactionBundleArtifacts, first renderPipelineResult) (string, error) {
	limits, err := resolveRenderLimits(contract.DeclaredLimits)
	if err != nil {
		return "", err
	}
	ctx, cancel := withRenderDeadline(ctx, limits.rerenderTimeout(kind), "determinism", "timeout.determinism_rerender")
	defer cancel()
	second, err := r.buildRenderPipeline(ctx, contract, kind, model, releaseScope, options, compositionJSON, artifacts)
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
		"schema_id":                 "cartulary.reporting_render_determinism.v1",
		"primary_path":              result.PrimaryPath,
		"primary_media_type":        result.PrimaryMedia,
		"files":                     fileDigests,
		"template_manifest_sha256":  result.Metadata.TemplateManifestSHA256,
		"toolchain_snapshot_sha256": result.Metadata.ToolchainSnapshotSHA256,
		"validation_summary_sha256": result.Metadata.ValidationSummarySHA256,
		"token_manifest_sha256":     result.Metadata.TokenManifestSHA256,
	}
	encoded, err := canonicalJSON(payload)
	if err != nil {
		return "", err
	}
	return reportingObjectDigest("cartulary.reporting_render_determinism.v1", encoded), nil
}

func optionalRenderSHA(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func templateManifestView(contract TemplateContract) TemplateContract { return contract }

func deriveDeckModel(contract TemplateContract, model RedactedExportModel, options renderOutputOptions, compositionJSON json.RawMessage) (reportingDeckModel, error) {
	composition, err := parseRenderComposition(compositionJSON)
	if err != nil {
		return reportingDeckModel{}, err
	}
	if err := bindCompositionPresentation(composition, model); err != nil {
		return reportingDeckModel{}, err
	}
	title := contract.DeckTitle
	if title == "" {
		title = "Cartulary Report"
	}
	model, err = readAssembledRenderModel(contract, model)
	if err != nil {
		return reportingDeckModel{}, err
	}
	diagrams := model.Diagrams
	recordAnchors := newDeckRecordAnchors(model.Records, model.SourceIndex)
	slides := []reportingDeckSlide{}
	for _, section := range model.Sections {
		origin, exists := model.assembly.sections[section.SectionID]
		if !exists || origin.declarationIndex < 0 || origin.declarationIndex >= len(contract.Sections) {
			return reportingDeckModel{}, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		decl := contract.Sections[origin.declarationIndex]
		if !decl.IncludeInDeck || section.SectionKind == "validation" && model.ReleaseScope == ReleaseScopeExternal {
			continue
		}
		ordinal := len(slides) + 1
		slide := reportingDeckSlide{SlideID: fmt.Sprintf("sld-%04d", ordinal), Ordinal: ordinal, Layout: "default", Title: section.Title,
			Blocks: section.Blocks, ClickSteps: []reportingClickStep{}, clickProfile: decl.ClickProfile, includeSuperseded: decl.IncludeSuperseded,
			titleSourceRefs: origin.titleRefs.source, titleSupportRefs: origin.titleRefs.support, titlePartitions: origin.titleRefs.partitions,
			sectionOrigin: &deckSectionOrigin{DeclarationID: decl.DeclID, ExpansionKey: cloneStringPtr(origin.expansionKey), Required: decl.Required}, derivedBlockIDs: derivedDeckBlockIDs(section.Blocks), recordAnchors: recordAnchors}
		if ordinal == 1 {
			slide.Layout = "cover"
		}
		if decl.SlideLayoutID != nil {
			slide.Layout = *decl.SlideLayoutID
		}
		slides = append(slides, slide)
	}
	if composition != nil {
		if err := applyCompositionToDeck(&slides, composition, contract, model.ReleaseScope, diagrams...); err != nil {
			return reportingDeckModel{}, err
		}
	}
	for _, slide := range slides {
		if err := validateSlideLifecycle(slide, model); err != nil {
			return reportingDeckModel{}, err
		}
		if _, err := resolveDeckLayout(contract, slide); err != nil {
			return reportingDeckModel{}, err
		}
	}
	limits, err := resolveRenderLimits(contract.DeclaredLimits)
	if err != nil {
		return reportingDeckModel{}, err
	}
	slides, err = chunkDeckSlides(slides, limits)
	if err != nil {
		return reportingDeckModel{}, err
	}
	for i := range slides {
		if err := materializeSlideClicks(&slides[i], limits); err != nil {
			return reportingDeckModel{}, err
		}
	}
	deckID, err := model.RenderIdentity.generatedID("deck", "deck_", map[string]any{"template_id": contract.TemplateID, "template_version": contract.TemplateVersion, "export_model_id": model.ExportModelID})
	if err != nil {
		return reportingDeckModel{}, err
	}
	refs := emptyContainerRefs()
	clickCount := 0
	for _, slide := range slides {
		refs.include(slide.SourceRefs, slide.SupportRefs, slide.DisclosurePartitionRefs)
		clickCount += len(slide.ClickSteps)
	}
	refs = refs.normalized()
	return reportingDeckModel{
		RenderIdentity: model.RenderIdentity, DeckID: deckID, SnapshotID: model.SnapshotID, TemplateID: contract.TemplateID, TemplateVersion: contract.TemplateVersion,
		SchemaID:          "cartulary.reporting_slide_deck.v2",
		Title:             title,
		OutputOptions:     options,
		Slides:            slides,
		SlideCount:        len(slides),
		ClickStepCount:    clickCount,
		ExpectedPageCount: len(slides) + clickCount,
		SourceRefs:        refs.source,
		SupportRefs:       refs.support,
	}, nil
}

func deriveDiagramModels(contract TemplateContract, model RedactedExportModel, compositionJSON json.RawMessage, releaseScope, outputKind string, graphs map[string]selectedGraphDiagram) ([]reportingDiagramModel, error) {
	composition, err := parseRenderComposition(compositionJSON)
	if err != nil {
		return nil, err
	}
	decls, owned, err := resolveDiagramDeclarations(contract, composition)
	if err != nil {
		return nil, err
	}
	diagrams := []reportingDiagramModel{}
	for _, decl := range decls {
		if decl.LayoutMode == "manual" && outputKind != OutputKindSlidev {
			return nil, newRenderValidationError("composition_invalid", "manual_layout_not_supported_for_output_kind")
		}
		if decl.DeclID == "" {
			return nil, newRenderValidationError("composition_invalid", "composition_anchor_unresolved")
		}
		if decl.DiagramSourceKind == "graph" {
			selection, exists := graphs[decl.DeclID]
			if !exists {
				return nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_bound")
			}
			diagram, err := graphDiagramFromSelection(decl, selection, model)
			if err != nil {
				return nil, err
			}
			if err := applyDiagramLabelOverrides(&diagram, decl.LabelOverrides); err != nil {
				return nil, err
			}
			diagram.LayoutMode, diagram.CompositionLayout, diagram.compositionOwned = decl.LayoutMode, decl.Layout, owned[decl.DeclID]
			diagrams = append(diagrams, diagram)
			continue
		}
		events := eligibleDiagramEvents(model, diagramAllowsSuperseded(contract, composition, decl.DeclID))
		diagram, err := timelineDiagramFromEvents(decl, events, releaseScope)
		if err != nil {
			return nil, err
		}
		if err := applyDiagramLabelOverrides(&diagram, decl.LabelOverrides); err != nil {
			return nil, err
		}
		diagram.LayoutMode, diagram.CompositionLayout, diagram.compositionOwned = decl.LayoutMode, decl.Layout, owned[decl.DeclID]
		diagrams = append(diagrams, diagram)
	}
	if err := applyCompositionDiagramOps(&diagrams, composition, releaseScope); err != nil {
		return nil, err
	}
	sort.Slice(diagrams, func(i, j int) bool {
		return diagrams[i].DiagramID < diagrams[j].DiagramID
	})
	return diagrams, nil
}

func resolveLoadedGraphResult(results []resolvedGraphResult, ref sourceProjectionRef) (resolvedGraphResult, error) {
	for _, result := range results {
		if result.Ref.GraphViewID == ref.GraphViewID && result.Ref.ProjectionResultID == ref.ProjectionResultID {
			if result.Result.Binding != ref.binding() {
				return resolvedGraphResult{}, newRenderValidationError("graph_projection_unavailable", "graph_projection_digest_mismatch")
			}
			return result, nil
		}
	}
	return resolvedGraphResult{}, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_completed")
}

func selectGraphDiagramObjects(result graphprojection.CompletedResultV2, rule diagramSelectionRule, vertexLimit int) ([]graphprojection.ResultVertexV2, []graphprojection.ResultEdgeV2, *diagramOverflowSummary, error) {
	vertexByID := make(map[string]graphprojection.ResultVertexV2, len(result.Vertices))
	for _, vertex := range result.Vertices {
		vertexByID[vertex.VertexID] = vertex
	}
	edgeByID := make(map[string]graphprojection.ResultEdgeV2, len(result.Edges))
	for _, edge := range result.Edges {
		edgeByID[edge.EdgeID] = edge
	}
	selectedVertices := map[string]struct{}{}
	selectedEdges := map[string]struct{}{}
	switch rule.RuleID {
	case "explicit_refs":
		if len(rule.VertexRefs) == 0 && len(rule.EdgeRefs) == 0 {
			return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
		}
		if hasDuplicateStrings(rule.VertexRefs) || hasDuplicateStrings(rule.EdgeRefs) {
			return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "diagram_selection_duplicate_ref")
		}
		for _, vertexID := range rule.VertexRefs {
			if _, present := vertexByID[vertexID]; !present {
				return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
			}
			selectedVertices[vertexID] = struct{}{}
		}
		for _, edgeID := range rule.EdgeRefs {
			edge, present := edgeByID[edgeID]
			if !present {
				return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
			}
			if _, present := selectedVertices[edge.SrcVertexID]; !present {
				return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_selection_unresolved")
			}
			if _, present := selectedVertices[edge.DstVertexID]; !present {
				return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_selection_unresolved")
			}
			selectedEdges[edgeID] = struct{}{}
		}
	case "neighborhood":
		if len(rule.SeedVertexRefs) == 0 || hasDuplicateStrings(rule.SeedVertexRefs) || rule.Depth == nil || *rule.Depth < 1 || *rule.Depth > 2 {
			return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
		}
		frontier := map[string]struct{}{}
		for _, vertexID := range rule.SeedVertexRefs {
			if _, present := vertexByID[vertexID]; !present {
				return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
			}
			selectedVertices[vertexID] = struct{}{}
			frontier[vertexID] = struct{}{}
		}
		edgeKinds := stringSet(rule.EdgeKindFilter)
		for depth := 0; depth < *rule.Depth && len(frontier) > 0; depth++ {
			next := map[string]struct{}{}
			for _, edge := range result.Edges {
				if len(edgeKinds) > 0 {
					if _, allowed := edgeKinds[edge.EdgeKind]; !allowed {
						continue
					}
				}
				_, srcIn := frontier[edge.SrcVertexID]
				_, dstIn := frontier[edge.DstVertexID]
				if srcIn == dstIn {
					continue
				}
				other := edge.SrcVertexID
				if srcIn {
					other = edge.DstVertexID
				}
				if _, present := vertexByID[other]; !present {
					return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_selection_unresolved")
				}
				selectedEdges[edge.EdgeID] = struct{}{}
				if _, seen := selectedVertices[other]; !seen {
					selectedVertices[other] = struct{}{}
					next[other] = struct{}{}
				}
			}
			frontier = next
		}
	case "all_with_bounds":
		for _, vertex := range result.Vertices {
			selectedVertices[vertex.VertexID] = struct{}{}
		}
		for _, edge := range result.Edges {
			selectedEdges[edge.EdgeID] = struct{}{}
		}
	default:
		return nil, nil, nil, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
	}
	vertices := make([]graphprojection.ResultVertexV2, 0, len(selectedVertices))
	for _, vertex := range result.Vertices {
		if _, selected := selectedVertices[vertex.VertexID]; selected {
			vertices = append(vertices, vertex)
		}
	}
	edges := make([]graphprojection.ResultEdgeV2, 0, len(selectedEdges))
	for _, edge := range result.Edges {
		if _, selected := selectedEdges[edge.EdgeID]; selected {
			edges = append(edges, edge)
		}
	}
	var overflow *diagramOverflowSummary
	if len(vertices) > vertexLimit || len(edges) > 160 {
		if rule.OverflowPolicy != "summarize" {
			return nil, nil, nil, newRenderValidationError("mermaid_invalid", "diagram_hard_limit_exceeded")
		}
		originalVertices := vertices
		originalEdges := edges
		if len(vertices) > vertexLimit {
			vertices = vertices[:vertexLimit]
		}
		kept := make(map[string]struct{}, len(vertices))
		for _, vertex := range vertices {
			kept[vertex.VertexID] = struct{}{}
		}
		boundedEdges := make([]graphprojection.ResultEdgeV2, 0, min(len(edges), 160))
		for _, edge := range edges {
			_, src := kept[edge.SrcVertexID]
			_, dst := kept[edge.DstVertexID]
			if src && dst && len(boundedEdges) < 160 {
				boundedEdges = append(boundedEdges, edge)
			}
		}
		edges = boundedEdges
		firstOmittedRef := firstOmittedDiagramRef(originalVertices, vertices, originalEdges, edges)
		overflow = &diagramOverflowSummary{
			SchemaID:           "diagram_overflow_summary.v1",
			OmittedVertexCount: len(originalVertices) - len(vertices),
			OmittedEdgeCount:   len(originalEdges) - len(edges),
			SelectionRuleID:    rule.RuleID,
			FirstOmittedRef:    firstOmittedRef,
		}
	}
	return vertices, edges, overflow, nil
}

func firstOmittedDiagramRef(originalVertices, keptVertices []graphprojection.ResultVertexV2, originalEdges, keptEdges []graphprojection.ResultEdgeV2) *string {
	keptVertexIDs := make(map[string]struct{}, len(keptVertices))
	for _, vertex := range keptVertices {
		keptVertexIDs[vertex.VertexID] = struct{}{}
	}
	for _, vertex := range originalVertices {
		if _, kept := keptVertexIDs[vertex.VertexID]; !kept {
			value := vertex.VertexID
			return &value
		}
	}
	keptEdgeIDs := make(map[string]struct{}, len(keptEdges))
	for _, edge := range keptEdges {
		keptEdgeIDs[edge.EdgeID] = struct{}{}
	}
	for _, edge := range originalEdges {
		if _, kept := keptEdgeIDs[edge.EdgeID]; !kept {
			value := edge.EdgeID
			return &value
		}
	}
	return nil
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
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

func validateBundleFilesSelfContained(files []RenderBundleFile) error {
	localCount, localBytes := 0, int64(0)
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
			if err := validateRenderedSVG(file.Bytes); err != nil {
				return err
			}
		case renderBundleRoleLocalAsset, renderBundleRoleLocalTheme:
			localCount++
			localBytes += int64(len(file.Bytes))
			if localCount > 500 {
				return newRenderValidationError("asset_limit_exceeded", "local_asset_count_exceeded")
			}
			if localBytes > 250<<20 {
				return newRenderValidationError("asset_limit_exceeded", "local_assets_total_bytes_exceeded")
			}
			if err := validateTemplateFileContent(file.MediaType, file.Bytes); err != nil {
				return err
			}
		}
	}
	return validateThemeBundleFiles(files)
}

func validateSourceValuesForRemoteAssets(model RedactedExportModel) error {
	check := func(value any) error {
		text := formatFieldValue(value)
		if markdownRemoteImagePattern.MatchString(text) ||
			htmlRemoteScriptSourcePattern.MatchString(text) ||
			htmlRemoteLinkHrefPattern.MatchString(text) ||
			htmlRemoteMediaPattern.MatchString(text) ||
			htmlRemoteStylePattern.MatchString(text) ||
			cssRemoteAssetPattern.MatchString(text) {
			return ErrRemoteRuntimeAsset
		}
		return nil
	}
	var visit func([]ReportingBlock) error
	visit = func(blocks []ReportingBlock) error {
		for _, block := range blocks {
			for _, field := range block.Fields {
				if err := check(field.Value); err != nil {
					return err
				}
			}
			if err := visit(block.Children); err != nil {
				return err
			}
		}
		return nil
	}
	for _, section := range model.Sections {
		if err := visit(section.Blocks); err != nil {
			return err
		}
	}
	for _, text := range model.PresentationTexts {
		if err := check(text.Body.Value); err != nil {
			return err
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

// Mermaid's grammar does not admit backslash-escaped quote delimiters. V2
// encodes syntax-sensitive characters once using Mermaid decimal entities.
func escapeMermaidLabel(value string) string {
	var out strings.Builder
	for index, ch := range value {
		if strings.ContainsRune("\"#%&';\\`|", ch) || (ch == ' ' && (index == 0 || index == len(value)-1)) {
			fmt.Fprintf(&out, "#%d;", ch)
		} else {
			out.WriteRune(ch)
		}
	}
	return out.String()
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
	AuthoredTextRef string                     `json:"authored_text_ref,omitempty"`
	Position        string                     `json:"position,omitempty"`
	SectionAnchor   *compositionSectionAnchor  `json:"section_anchor,omitempty"`
	SectionAnchors  []compositionSectionAnchor `json:"section_anchors,omitempty"`
	BlockAnchor     *compositionBlockAnchor    `json:"block_anchor,omitempty"`
	DiagramAnchor   *compositionDiagramAnchor  `json:"diagram_anchor,omitempty"`
	LayoutID        string                     `json:"layout_id,omitempty"`
	ClickProfile    string                     `json:"click_profile,omitempty"`
	LabelOverrides  []diagramLabelOverride     `json:"label_overrides,omitempty"`
}

type compositionAuthoredText struct {
	omitted                bool
	retainedBody           *ReportingField
	AuthoredTextID         string `json:"authored_text_id"`
	TextRole               string `json:"text_role"`
	Body                   string `json:"body"`
	DisclosurePartitionRef string `json:"disclosure_partition_ref"`
}

type compositionDiagramDecl struct {
	Layout            *compositionDiagramLayout `json:"layout"`
	DeclID            string                    `json:"decl_id"`
	ReplacesDeclID    *string                   `json:"replaces_decl_id"`
	DiagramKind       string                    `json:"diagram_kind"`
	DiagramSourceKind string                    `json:"diagram_source_kind"`
	SourceGraphViewID *string                   `json:"source_graph_view_id"`
	SelectionRule     diagramSelectionRule      `json:"selection_rule"`
	LayoutMode        string                    `json:"layout_mode"`
	LabelOverrides    []diagramLabelOverride    `json:"label_overrides"`
}

type diagramSelectionRule = diagramcontract.SelectionRule

func parseRenderComposition(raw json.RawMessage) (*renderComposition, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	source, err := decodeCompositionRenderInput(raw)
	if err != nil {
		return nil, err
	}
	composition := renderComposition{SchemaID: source.SchemaID}
	for _, projection := range []struct {
		raw    json.RawMessage
		target any
	}{
		{source.DeckOps, &composition.DeckOps}, {source.DiagramDecls, &composition.DiagramDecls}, {source.AuthoredTexts, &composition.AuthoredTexts},
	} {
		if err := json.Unmarshal(projection.raw, projection.target); err != nil {
			return nil, newRenderValidationError("composition_invalid", "invalid_scalar_value")
		}
	}
	for _, decl := range composition.DiagramDecls {
		if decl.DiagramSourceKind != "" && decl.DiagramSourceKind != "graph" && decl.DiagramSourceKind != "timeline" {
			return nil, fmt.Errorf("composition invalid: diagram_source_kind")
		}
		if decl.DiagramSourceKind == "graph" && (decl.SelectionRule.SchemaID != "diagram_selection_rule.v1" || decl.SelectionRule.OverflowPolicy == "") {
			return nil, fmt.Errorf("composition invalid: selection_rule")
		}
	}
	return &composition, nil
}

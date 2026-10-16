package reporting

import "encoding/json"

const templateManifestSchema = "cartulary.reporting_template_pack_manifest.v3"

// The admitted template is the manifest itself. No private binding inventory or
// undeclared assets are projected beside it under the same schema identity.
type TemplateContract struct {
	SchemaID                   string                     `json:"schema_id"`
	TemplateID                 string                     `json:"template_id"`
	TemplateVersion            string                     `json:"template_version"`
	ManifestVersion            string                     `json:"manifest_version"`
	DeckTitle                  string                     `json:"deck_title"`
	Author                     string                     `json:"author"`
	Theme                      string                     `json:"theme"`
	SupportedOutputKinds       []string                   `json:"supported_output_kinds"`
	SupportedReleaseScopes     []string                   `json:"supported_release_scopes"`
	AllowedExportModelBindings []string                   `json:"allowed_export_model_bindings"`
	Sections                   []TemplateSection          `json:"sections"`
	Layouts                    []TemplateLayout           `json:"layouts"`
	NarrativeSlots             []TemplateNarrativeSlot    `json:"narrative_slots"`
	Assets                     []TemplateAsset            `json:"assets"`
	RenderProfiles             []TemplateRenderProfile    `json:"render_profiles"`
	SupportedOutputOptions     []TemplateOutputOptions    `json:"supported_output_options"`
	DeclaredLimits             map[string]json.RawMessage `json:"declared_limits"`
	ExtensionFeatures          []TemplateExtensionFeature `json:"extension_features"`
}

type TemplateSection struct {
	DeclID            string                       `json:"decl_id"`
	SectionKind       string                       `json:"section_kind"`
	Title             string                       `json:"title"`
	Expansion         string                       `json:"expansion"`
	Bindings          []string                     `json:"bindings"`
	SlideLayoutID     *string                      `json:"slide_layout_id,omitempty"`
	ClickProfile      string                       `json:"click_profile"`
	IncludeInDeck     bool                         `json:"include_in_deck"`
	Required          bool                         `json:"required"`
	IncludeSuperseded bool                         `json:"include_superseded"`
	TruncateLabels    bool                         `json:"truncate_labels"`
	SelectionRuleID   *string                      `json:"selection_rule_id"`
	DiagramDecls      []TemplateDiagramDeclaration `json:"diagram_decls"`
}

type TemplateLayout struct {
	LayoutID          string   `json:"layout_id"`
	SlidevLayout      string   `json:"slidev_layout"`
	AllowedBlockKinds []string `json:"allowed_block_kinds"`
}

type TemplateNarrativeSlot struct {
	SlotID        string `json:"slot_id"`
	SectionDeclID string `json:"section_decl_id"`
	Binding       string `json:"binding"`
	Required      bool   `json:"required"`
	MaxChars      int    `json:"max_chars"`
}

type TemplateAsset struct {
	SchemaID           string `json:"schema_id"`
	AssetID            string `json:"asset_id"`
	AssetRole          string `json:"asset_role"`
	BundlePath         string `json:"bundle_path"`
	SourcePath         string `json:"source_path"`
	MediaType          string `json:"media_type"`
	SHA256             string `json:"sha256"`
	ByteSize           int64  `json:"byte_size"`
	RequiredForRelease bool   `json:"required_for_release"`
}

type TemplateRenderProfile struct {
	RenderEnvironmentProfileID string `json:"render_environment_profile_id"`
	ViewportCSSPX              string `json:"viewport_css_px"`
	DeviceScaleFactor          int    `json:"device_scale_factor"`
	ColorScheme                string `json:"color_scheme"`
	Locale                     string `json:"locale"`
}

type TemplateOutputOptions struct {
	SchemaID         string `json:"schema_id"`
	OutputKind       string `json:"output_kind"`
	PDF              bool   `json:"pdf"`
	SVG              bool   `json:"svg"`
	PNG              bool   `json:"png"`
	PPTX             bool   `json:"pptx"`
	RenderedDiagrams bool   `json:"rendered_diagrams"`
}

type TemplateExtensionFeature struct {
	FeatureID     string  `json:"feature_id"`
	Required      bool    `json:"required"`
	PayloadSHA256 *string `json:"payload_sha256"`
	PayloadPath   *string `json:"payload_path"`
}

type TemplateDiagramDeclaration struct {
	SchemaID          string               `json:"schema_id"`
	DeclID            string               `json:"decl_id"`
	DiagramKind       string               `json:"diagram_kind"`
	DiagramSourceKind string               `json:"diagram_source_kind"`
	SourceGraphViewID *string              `json:"source_graph_view_id"`
	SelectionRule     diagramSelectionRule `json:"selection_rule"`
	Required          bool                 `json:"required"`
}

func ResolveTemplateContract(id string, version string) (TemplateContract, bool) {
	if id != DefaultTemplateID || version != DefaultTemplateVersion {
		return TemplateContract{}, false
	}
	blocks := supportedReportingBlockKinds()
	return TemplateContract{
		SchemaID: templateManifestSchema, TemplateID: id, TemplateVersion: version, ManifestVersion: "3",
		DeckTitle: "Incident Report", Author: "", Theme: "default", DeclaredLimits: map[string]json.RawMessage{},
		SupportedOutputKinds: supportedOutputKinds(), SupportedReleaseScopes: supportedReleaseScopes(),
		AllowedExportModelBindings: []string{"$.sections[*]"},
		Sections:                   []TemplateSection{{DeclID: "snapshot", SectionKind: "appendix", Title: "Snapshot export", Expansion: "none", Bindings: []string{"$.sections[*]"}, ClickProfile: "none", IncludeInDeck: true, Required: true, DiagramDecls: []TemplateDiagramDeclaration{}}},
		Layouts: []TemplateLayout{
			{LayoutID: "cover", SlidevLayout: "cover", AllowedBlockKinds: append([]string{}, blocks...)},
			{LayoutID: "default", SlidevLayout: "default", AllowedBlockKinds: append([]string{}, blocks...)},
			{LayoutID: "section", SlidevLayout: "section", AllowedBlockKinds: append([]string{}, blocks...)},
		},
		NarrativeSlots: []TemplateNarrativeSlot{}, Assets: []TemplateAsset{},
		RenderProfiles: []TemplateRenderProfile{{RenderEnvironmentProfileID: "cartulary.reporting.local.v1", ViewportCSSPX: "1280x720", DeviceScaleFactor: 1, ColorScheme: "light", Locale: "C.UTF-8"}},
		SupportedOutputOptions: []TemplateOutputOptions{
			{SchemaID: "template_output_option_support.v1", OutputKind: OutputKindMermaid, SVG: true, RenderedDiagrams: true},
			{SchemaID: "template_output_option_support.v1", OutputKind: OutputKindSlidev, PDF: true, RenderedDiagrams: true},
		},
		ExtensionFeatures: []TemplateExtensionFeature{},
	}, true
}

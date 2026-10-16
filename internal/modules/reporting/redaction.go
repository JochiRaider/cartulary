package reporting

import (
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	DefaultRedactionProfileVersion = "5"
	InternalRedactionProfileID     = "cartulary.redaction.internal"
	ExternalRedactionProfileID     = "cartulary.redaction.external"
	TokenizedRedactionProfileID    = "cartulary.redaction.tokenized_review"

	RedactionProfileSchemaID       = "cartulary.redaction_profile.v3"
	RedactionProfileViewSchemaID   = "cartulary.reporting_redaction_profile_view.v2"
	RedactionManifestSchemaID      = "cartulary.redaction_manifest.v5"
	RedactionTokenManifestSchemaID = "cartulary.reporting_token_manifest.v3"
	RedactionRevealMapSchemaID     = "cartulary.reporting_token_reveal_map.v3"

	ActionAllow    = "allow"
	ActionDrop     = "drop"
	ActionMask     = "mask"
	ActionTruncate = "truncate"
	ActionStub     = "stub"
	ActionHash     = "hash"

	DefaultMaskText     = "[REDACTED]"
	DefaultStubText     = "[STUB]"
	DefaultTruncateMark = "[TRUNCATED]"

	ContentClassCuratedNarrative    = "curated_narrative"
	ContentClassDerivedAnalytic     = "derived_analytic"
	ContentClassSourceEvidence      = "source_evidence"
	ContentClassWorkingMaterial     = "working_material"
	ContentClassPresentationText    = "presentation_text"
	ContentClassSupportReference    = "support_reference"
	ContentClassValidation          = "validation"
	ContentClassTemplateBoilerplate = "template_boilerplate"
	ReportingSectionSchemaID        = "cartulary.reporting_section.v2"
	ReportingBlockSchemaID          = "cartulary.reporting_block.v3"
)

var (
	ErrInvalidRedactionProfile   = errors.New("reporting: invalid redaction profile")
	ErrRedactionValidation       = errors.New("reporting: redaction validation failed")
	ErrUndeclaredTemplateBinding = errors.New("reporting: undeclared template binding")
	ErrMissingRequiredField      = errors.New("reporting: missing required template field")
	ErrRemoteRuntimeAsset        = errors.New("reporting: rendered output requires remote runtime asset")
)

var (
	htmlRemoteScriptSourcePattern = regexp.MustCompile(`(?is)<\s*script\b[^>]*\bsrc\s*=\s*(?:"[^"]*(?:https?:)?//[^"]*"|'[^']*(?:https?:)?//[^']*'|[^\s>]*(?:https?:)?//[^\s>]*)`)
	htmlRemoteLinkHrefPattern     = regexp.MustCompile(`(?is)<\s*link\b[^>]*\bhref\s*=\s*(?:"[^"]*(?:https?:)?//[^"]*"|'[^']*(?:https?:)?//[^']*'|[^\s>]*(?:https?:)?//[^\s>]*)`)
	htmlRemoteMediaPattern        = regexp.MustCompile(`(?is)<\s*(?:img|audio|video|source|track|iframe|embed|object)\b[^>]*\b(?:src|srcset|poster|data)\s*=\s*(?:"[^"]*(?:https?:)?//[^"]*"|'[^']*(?:https?:)?//[^']*'|[^\s>]*(?:https?:)?//[^\s>]*)`)
	htmlRemoteStylePattern        = regexp.MustCompile(`(?is)<[^>]*\bstyle\s*=\s*(?:"[^"]*(?:@import\s+(?:url\(\s*)?['"]?(?:https?:)?//|url\(\s*['"]?(?:https?:)?//)[^"]*"|'[^']*(?:@import\s+(?:url\(\s*)?['"]?(?:https?:)?//|url\(\s*['"]?(?:https?:)?//)[^']*')`)
	cssRemoteAssetPattern         = regexp.MustCompile(`(?is)(?:@import\s+(?:url\(\s*)?['"]?(?:https?:)?//|url\(\s*['"]?(?:https?:)?//)`)
	markdownRemoteImagePattern    = regexp.MustCompile(`(?is)!\[[^\]]*\]\(\s*(?:https?:)?//`)
)

type RedactionProfile struct {
	ProfileStatus                  string              `json:"profile_status"`
	AllowedReleaseScopes           []string            `json:"allowed_release_scopes"`
	IncludeSuperseded              bool                `json:"include_superseded"`
	NeutralTokenFamily             bool                `json:"neutral_token_family"`
	SchemaID                       string              `json:"schema_id"`
	ProfileID                      string              `json:"profile_id"`
	Version                        string              `json:"version"`
	AllowedDisclosurePartitionRefs []string            `json:"allowed_disclosure_partition_refs,omitempty"`
	AllowAuthoredPresentationText  bool                `json:"allow_authored_presentation_text"`
	DefaultAction                  RedactionActionSpec `json:"default_action"`
	Rules                          []RedactionRule     `json:"rules"`
}

type RedactionRule struct {
	RuleID       string              `json:"rule_id"`
	Path         *string             `json:"path,omitempty"`
	ContentClass *string             `json:"content_class,omitempty"`
	Action       RedactionActionSpec `json:"action"`
}

type RedactionActionSpec struct {
	TokenBacked     bool    `json:"token_backed"`
	Type            string  `json:"type"`
	ReplacementText *string `json:"replacement_text,omitempty"`
	StubText        *string `json:"stub_text,omitempty"`
	MaxChars        *int    `json:"max_chars,omitempty"`
}

type SnapshotContent struct {
	ReferencePacks               reference_data.SetBinding      `json:"reference_packs"`
	IncidentID                   string                         `json:"incident_id"`
	SnapshotID                   string                         `json:"snapshot_id"`
	SnapshotAt                   time.Time                      `json:"snapshot_at"`
	SourceChangeSetHighWatermark string                         `json:"source_change_set_high_watermark"`
	SnapshotBoundaryKind         *string                        `json:"snapshot_boundary_kind"`
	DerivationVersion            string                         `json:"derivation_version"`
	ExportModelGeneratorID       string                         `json:"export_model_generator_id"`
	ExportModelGeneratorVersion  string                         `json:"export_model_generator_version"`
	Sections                     []ReportingSection             `json:"sections"`
	Records                      []ReportingRecordSummary       `json:"records"`
	Relationships                []ReportingRelationshipSummary `json:"relationships"`
	TimelineEvents               []ReportingTimelineEvent       `json:"timeline_events"`
	Subjects                     []TokenizableSubject           `json:"subjects"`
	Diagrams                     []ReportingDiagram             `json:"diagrams"`
	Assets                       []ReportingAssetDeclaration    `json:"assets"`
	SupportIndex                 []ReportingSupportRef          `json:"support_index"`
	SourceIndex                  []ReportingSourceRef           `json:"source_index"`
	ValidationSummary            ReportingExportModelValidation `json:"validation_summary"`
}

type SnapshotModel struct {
	SourceFacts     exportprovider.SourceFacts `json:"source_facts"`
	SchemaID        string                     `json:"schema_id"`
	SnapshotModelID string                     `json:"snapshot_model_id"`
	SnapshotContent
}

type ExportModel struct {
	PresentationTexts []ReportingPresentationText `json:"presentation_texts"`
	SchemaID          string                      `json:"schema_id"`
	ExportModelID     string                      `json:"export_model_id"`
	SnapshotModelID   string                      `json:"snapshot_model_id"`
	RenderIdentity
	ExportModelCreatedAt   string   `json:"export_model_created_at"`
	ReleaseScope           string   `json:"release_scope"`
	RecipientPartitionRefs []string `json:"recipient_partition_refs"`
	SnapshotContent
}

type ExportField struct {
	lifecycle               lifecycleRestriction
	SourceRecordID          string   `json:"source_record_id"`
	SourceRecordType        string   `json:"source_record_type"`
	Path                    string   `json:"path"`
	ContentClass            string   `json:"content_class"`
	SourceFamily            string   `json:"source_family,omitempty"`
	Value                   any      `json:"value,omitempty"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs,omitempty"`
	SupportRefs             []string `json:"support_refs,omitempty"`
	RawBlobSource           bool     `json:"raw_blob_source,omitempty"`
	OpaqueBinary            bool     `json:"opaque_binary,omitempty"`
	GeneratedPresentation   bool     `json:"generated_presentation,omitempty"`
}

type ReportingSection struct {
	SchemaID                string                       `json:"schema_id"`
	SectionID               string                       `json:"section_id"`
	SectionKind             string                       `json:"section_kind"`
	Title                   string                       `json:"title"`
	OrderingKey             string                       `json:"ordering_key"`
	Blocks                  []ReportingBlock             `json:"blocks"`
	SourceRefs              []string                     `json:"source_refs"`
	SupportRefs             []string                     `json:"support_refs"`
	DisclosurePartitionRefs []string                     `json:"disclosure_partition_refs"`
	ContentClassSummary     ReportingContentClassSummary `json:"content_class_summary"`
	SectionValidation       ReportingSectionValidation   `json:"section_validation"`
}

type ReportingBlock struct {
	SchemaID                string           `json:"schema_id"`
	BlockID                 string           `json:"block_id"`
	BlockKind               string           `json:"block_kind"`
	BlockOrdinal            int              `json:"block_ordinal"`
	ParentBlockID           *string          `json:"parent_block_id"`
	ContentClass            string           `json:"content_class"`
	Fields                  []ReportingField `json:"fields"`
	Children                []ReportingBlock `json:"children"`
	SourceRefs              []string         `json:"source_refs"`
	SupportRefs             []string         `json:"support_refs"`
	DisclosurePartitionRefs []string         `json:"disclosure_partition_refs"`
}

type ReportingField struct {
	SchemaID                string   `json:"schema_id"`
	FieldKey                string   `json:"field_key"`
	DisplayLabel            *string  `json:"display_label"`
	FieldOrdinal            int      `json:"field_ordinal"`
	SourceValueState        string   `json:"source_value_state"`
	RedactedValueState      string   `json:"redacted_value_state"`
	Value                   any      `json:"value"`
	RawValueSHA256          *string  `json:"raw_value_sha256"`
	SourceRefs              []string `json:"source_refs"`
	SupportRefs             []string `json:"support_refs"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs"`
}

type ReportingRecordSummary struct {
	SchemaID                string           `json:"schema_id"`
	RecordID                string           `json:"record_id"`
	RecordType              string           `json:"record_type"`
	SourceRecordRef         SourceRecordRef  `json:"source_record_ref"`
	DisplayName             *string          `json:"display_name"`
	DeletedState            string           `json:"deleted_state"`
	Fields                  []ReportingField `json:"fields"`
	SourceRefs              []string         `json:"source_refs"`
	SupportRefs             []string         `json:"support_refs"`
	DisclosurePartitionRefs []string         `json:"disclosure_partition_refs"`
}

type SourceRecordRef struct {
	SchemaID         string  `json:"schema_id"`
	SourceFamily     string  `json:"source_family"`
	SourceRecordID   string  `json:"source_record_id"`
	SourceSnapshotID string  `json:"source_snapshot_id"`
	SourceRefID      *string `json:"source_ref_id"`
}

type ReportingRelationshipSummary struct {
	SchemaID                string                  `json:"schema_id"`
	RelationshipID          string                  `json:"relationship_id"`
	RelationshipKind        string                  `json:"relationship_kind"`
	SrcRecordRef            RelationshipEndpointRef `json:"src_record_ref"`
	DstRecordRef            RelationshipEndpointRef `json:"dst_record_ref"`
	Direction               string                  `json:"direction"`
	Confidence              *string                 `json:"confidence"`
	SourceRefs              []string                `json:"source_refs"`
	SupportRefs             []string                `json:"support_refs"`
	DisclosurePartitionRefs []string                `json:"disclosure_partition_refs"`
}

type RelationshipEndpointRef struct {
	SchemaID         string          `json:"schema_id"`
	EndpointRole     string          `json:"endpoint_role"`
	SourceRecordRef  SourceRecordRef `json:"source_record_ref"`
	DisplayRef       *string         `json:"display_ref"`
	StableSubjectRef *string         `json:"stable_subject_ref"`
}

type ReportingTimelineEvent struct {
	SchemaID                string                `json:"schema_id"`
	RecordID                string                `json:"record_id"`
	ActivitySortKey         string                `json:"activity_sort_key"`
	ActivitySortTS          *string               `json:"activity_sort_ts"`
	ActivityParseState      string                `json:"activity_parse_state"`
	ActivityPrecisionRank   int                   `json:"activity_precision_rank"`
	DateEnteredSortKey      *string               `json:"date_entered_sort_key"`
	RecordCreatedAt         string                `json:"record_created_at"`
	DisplayTimes            ReportingDisplayTimes `json:"display_times"`
	Fields                  []ReportingField      `json:"fields"`
	SourceRefs              []string              `json:"source_refs"`
	SupportRefs             []string              `json:"support_refs"`
	DisclosurePartitionRefs []string              `json:"disclosure_partition_refs"`
}

type ReportingDisplayTimes struct {
	PrimaryDisplay    *string `json:"primary_display"`
	PrecisionLabel    *string `json:"precision_label"`
	SourceTextDisplay *string `json:"source_text_display"`
}

type TokenizableSubject struct {
	SchemaID                    string   `json:"schema_id"`
	StableSubjectRef            string   `json:"stable_subject_ref"`
	SubjectFamily               string   `json:"subject_family"`
	SourceRecordID              *string  `json:"source_record_id"`
	SourceRecordType            *string  `json:"source_record_type"`
	EntityMentionID             *string  `json:"entity_mention_id"`
	CanonicalDisplayValueSHA256 string   `json:"canonical_display_value_sha256"`
	DisclosurePartitionRefs     []string `json:"disclosure_partition_refs"`
	SourceRefs                  []string `json:"source_refs"`
	SupportRefs                 []string `json:"support_refs"`
}

type ReportingDiagram struct {
	compositionOwned        bool
	SourceRefs              []string                  `json:"source_refs"`
	SupportRefs             []string                  `json:"support_refs"`
	DisclosurePartitionRefs []string                  `json:"disclosure_partition_refs"`
	CompositionLayout       *compositionDiagramLayout `json:"composition_layout"`
	SchemaID                string                    `json:"schema_id"`
	DiagramID               string                    `json:"diagram_id"`
	DiagramKind             string                    `json:"diagram_kind"`
	DiagramSourceKind       string                    `json:"diagram_source_kind"`
	LayoutMode              string                    `json:"layout_mode"`
	SourceProjectionRef     *sourceProjectionRef      `json:"source_projection_ref"`
	SelectionRule           diagramSelectionRule      `json:"selection_rule"`
	IncludedVertexRefs      []string                  `json:"included_vertex_refs"`
	IncludedEdgeRefs        []string                  `json:"included_edge_refs"`
	OverflowSummary         *diagramOverflowSummary   `json:"overflow_summary"`
	Nodes                   []reportingDiagramNode    `json:"nodes"`
	Edges                   []reportingDiagramEdge    `json:"edges"`
}

type ReportingAssetDeclaration struct {
	SchemaID                string   `json:"schema_id"`
	AssetID                 string   `json:"asset_id"`
	AssetKind               string   `json:"asset_kind"`
	BundlePath              string   `json:"bundle_path"`
	MediaType               string   `json:"media_type"`
	SHA256                  string   `json:"sha256"`
	ByteSize                int64    `json:"byte_size"`
	DeclaredBy              string   `json:"declared_by"`
	RequiredForRelease      bool     `json:"required_for_release"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs"`
}

type ReportingSupportRef struct {
	SchemaID                string   `json:"schema_id"`
	SupportRefID            string   `json:"support_ref_id"`
	SupportKind             string   `json:"support_kind"`
	SupportTargetRef        string   `json:"support_target_ref"`
	SourceRefID             *string  `json:"source_ref_id"`
	SourceSnapshotID        string   `json:"source_snapshot_id"`
	SupportRole             string   `json:"support_role"`
	CustodyState            string   `json:"custody_state"`
	SourceSummary           *string  `json:"source_summary"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs"`
}

type ReportingExportModelValidation struct {
	SchemaID   string                     `json:"schema_id"`
	Result     string                     `json:"result"`
	IssueCount int                        `json:"issue_count"`
	Issues     []ReportingValidationIssue `json:"issues"`
}

type ReportingValidationIssue struct {
	Stage           string `json:"stage"`
	Severity        string `json:"severity"`
	ExportModelPath string `json:"export_model_path"`
	FailureCode     string `json:"failure_code"`
	ReasonCode      string `json:"reason_code"`
}

type ReportingContentClassSummary struct {
	SourceEvidence      int `json:"source_evidence"`
	CuratedNarrative    int `json:"curated_narrative"`
	DerivedAnalytic     int `json:"derived_analytic"`
	WorkingMaterial     int `json:"working_material"`
	PresentationText    int `json:"presentation_text"`
	SupportReference    int `json:"support_reference"`
	Validation          int `json:"validation"`
	TemplateBoilerplate int `json:"template_boilerplate"`
}

type ReportingSectionValidation struct {
	Result     string `json:"result"`
	IssueCount int    `json:"issue_count"`
}

type RedactedExportModel struct {
	ExportModel
	assembly                  *renderModelAssembly
	admittedPresentationTexts map[string]string
	// The decision index retains pre-removal addresses for evidence. Deck
	// derivation reads only the reconstructed retained section tree.
	admittedFieldPaths    map[string]string
	lifecycleRestrictions map[string]lifecycleRestriction
	// Frozen dimension identities and selected token labels are private
	// derivation inputs, never an alternative case-content projection.
	admittedPartyIDs          []string
	subjectDisplayTokens      map[string]string
	admittedExportModelSHA256 string
	Fields                    []RedactedField `json:"-"`
}

type RedactedField struct {
	Path                    string   `json:"export_model_path"`
	ContentClass            string   `json:"content_class"`
	SourceFamily            string   `json:"source_family,omitempty"`
	Value                   any      `json:"value,omitempty"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs,omitempty"`
	SupportRefs             []string `json:"support_refs,omitempty"`
	RawBlobSource           bool     `json:"raw_blob_source,omitempty"`
	OpaqueBinary            bool     `json:"opaque_binary,omitempty"`
	GeneratedPresentation   bool     `json:"generated_presentation,omitempty"`
}

type RedactionManifest struct {
	RenderIdentity
	SnapshotID string `json:"snapshot_id"`
	CreatedAt  string `json:"created_at"`

	SchemaID           string                       `json:"schema_id"`
	ProfileID          string                       `json:"redaction_profile_id"`
	ProfileVersion     string                       `json:"redaction_profile_version"`
	ProfileSHA256      string                       `json:"redaction_profile_sha256"`
	Entries            []RedactionManifestEntry     `json:"entries"`
	StructuralRemovals []StructuralRedactionRemoval `json:"structural_removals"`
	ContainerDecisions []ContainerRedactionDecision `json:"container_decisions"`
}

type RedactionManifestEntry struct {
	Path                        string            `json:"export_model_path"`
	ContentClass                string            `json:"content_class"`
	Action                      string            `json:"action"`
	RuleID                      string            `json:"rule_id"`
	RulePrecedence              string            `json:"rule_precedence"`
	SelectedRuleTrace           SelectedRuleTrace `json:"selected_rule_trace"`
	DisclosurePartitionHandling string            `json:"partition_handling"`
	DisclosurePartitionRefs     []string          `json:"disclosure_partition_refs"`
	Outcome                     string            `json:"outcome"`
	TokenID                     *string           `json:"token_id"`
}

type RedactionTokenManifest struct {
	RenderIdentity
	SnapshotID string `json:"snapshot_id"`
	CreatedAt  string `json:"created_at"`

	SchemaID                string                        `json:"schema_id"`
	ProfileSHA256           string                        `json:"redaction_profile_sha256"`
	RedactionManifestSHA256 string                        `json:"redaction_manifest_sha256"`
	Entries                 []RedactionTokenManifestEntry `json:"entries"`
}

type RedactionTokenManifestEntry struct {
	TokenID                     string   `json:"token_id"`
	DisplayToken                string   `json:"display_token"`
	StableSubjectRef            string   `json:"stable_subject_ref"`
	SubjectFamily               string   `json:"subject_family"`
	SourceRecordID              *string  `json:"source_record_id"`
	EntityMentionID             *string  `json:"entity_mention_id"`
	CanonicalDisplayValueSHA256 string   `json:"canonical_display_value_sha256"`
	RecipientPartitionRefs      []string `json:"recipient_partition_refs"`
}

type RedactionRevealMap struct {
	RenderIdentity
	SnapshotID string `json:"snapshot_id"`
	CreatedAt  string `json:"created_at"`

	SchemaID            string                    `json:"schema_id"`
	TokenManifestSHA256 string                    `json:"token_manifest_sha256"`
	ProfileSHA256       string                    `json:"redaction_profile_sha256"`
	Entries             []RedactionRevealMapEntry `json:"entries"`
}

type RedactionRevealMapEntry struct {
	TokenID                     string   `json:"token_id"`
	DisplayToken                string   `json:"display_token"`
	StableSubjectRef            string   `json:"stable_subject_ref"`
	SubjectFamily               string   `json:"subject_family"`
	SourceRecordID              *string  `json:"source_record_id"`
	EntityMentionID             *string  `json:"entity_mention_id"`
	CanonicalDisplayValue       string   `json:"canonical_display_value"`
	CanonicalDisplayValueSHA256 string   `json:"canonical_display_value_sha256"`
	RecipientPartitionRefs      []string `json:"recipient_partition_refs"`
}

type RedactionResult struct {
	Model               RedactedExportModel
	Manifest            RedactionManifest
	ProfileView         RedactionProfileView
	TokenManifest       *RedactionTokenManifest
	RevealMap           *RedactionRevealMap
	ModelSHA256         string
	ModelJSON           []byte
	ManifestSHA256      string
	ProfileViewSHA256   string
	TokenManifestSHA256 string
	RevealMapSHA256     string
	ManifestJSON        []byte
	ProfileViewJSON     []byte
	TokenManifestJSON   []byte
	RevealMapJSON       []byte
}

type IncidentMetadataSnapshot struct {
	ID           string
	Title        string
	Description  *string
	Status       string
	Severity     *string
	TLP          *string
	CurrentPhase *string
	Version      int64
}

func BuildSnapshotModel(incident IncidentMetadataSnapshot, snapshotID string, snapshotAt time.Time, watermark string, source SnapshotSource) (SnapshotModel, string, error) {
	return buildSnapshotModel(context.Background(), incident, snapshotID, snapshotAt, watermark, source)
}

func buildSnapshotModel(ctx context.Context, incident IncidentMetadataSnapshot, snapshotID string, snapshotAt time.Time, watermark string, source SnapshotSource) (SnapshotModel, string, error) {
	fields := []ExportField{
		{
			Path:         "/incident/title",
			ContentClass: ContentClassCuratedNarrative,
			SourceFamily: "incident_metadata",
			Value:        incident.Title,
			SupportRefs: []string{
				"/incident/status",
			},
		},
		{
			Path:         "/incident/status",
			ContentClass: ContentClassDerivedAnalytic,
			SourceFamily: "incident_metadata",
			Value:        incident.Status,
		},
	}
	fields = append(fields, source.Fields...)
	if incident.Description != nil {
		fields = append(fields, ExportField{
			Path:                    "/incident/description",
			ContentClass:            ContentClassSourceEvidence,
			SourceFamily:            "incident_metadata",
			Value:                   *incident.Description,
			DisclosurePartitionRefs: []string{"public"},
		})
	}
	if incident.Severity != nil {
		fields = append(fields, ExportField{
			Path:         "/incident/severity",
			ContentClass: ContentClassDerivedAnalytic,
			SourceFamily: "incident_metadata",
			Value:        *incident.Severity,
		})
	}
	if incident.TLP != nil {
		fields = append(fields, ExportField{
			Path:         "/incident/tlp",
			ContentClass: ContentClassDerivedAnalytic,
			SourceFamily: "incident_metadata",
			Value:        *incident.TLP,
		})
	}
	if incident.CurrentPhase != nil {
		fields = append(fields, ExportField{
			Path:         "/incident/current_phase",
			ContentClass: ContentClassDerivedAnalytic,
			SourceFamily: "incident_metadata",
			Value:        *incident.CurrentPhase,
		})
	}
	return buildStructuredSnapshotModelContext(ctx, incident.ID, snapshotID, snapshotAt, watermark, SnapshotSource{SourceFacts: source.SourceFacts, Fields: fields, Timeline: source.Timeline})
}

func buildStructuredSnapshotModel(incidentID string, snapshotID string, snapshotAt time.Time, watermark string, source SnapshotSource) (SnapshotModel, string, error) {
	return buildStructuredSnapshotModelContext(context.Background(), incidentID, snapshotID, snapshotAt, watermark, source)
}

func buildStructuredSnapshotModelContext(ctx context.Context, incidentID string, snapshotID string, snapshotAt time.Time, watermark string, source SnapshotSource) (SnapshotModel, string, error) {
	fields := cloneExportFields(source.Fields)
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	supportTargets := map[string]supportTarget{}
	sourceOwners := map[string]sourceOwner{}
	for i, field := range fields {
		if err := ctx.Err(); err != nil {
			return SnapshotModel{}, "", err
		}
		if field.SourceFamily != "incident_metadata" && (field.SourceRecordID == "" || field.SourceRecordType == "") {
			return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
		if field.Path == "" {
			return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		if err := validateSourceContentClass(field.ContentClass); err != nil {
			return SnapshotModel{}, "", err
		}
		if field.RawBlobSource || field.OpaqueBinary || field.GeneratedPresentation {
			return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		if i > 0 && fields[i-1].Path == field.Path {
			return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
		}
		refs, targets, err := materializeSupportReferences(field.SupportRefs, incidentID, snapshotID)
		if err != nil {
			return SnapshotModel{}, "", err
		}
		fields[i].SupportRefs = refs
		for id, target := range targets {
			supportTargets[id] = target
		}
	}
	section := ReportingSection{
		SchemaID:                ReportingSectionSchemaID,
		SectionID:               "sec-0001",
		SectionKind:             "appendix",
		Title:                   "Snapshot export",
		OrderingKey:             "0001",
		Blocks:                  make([]ReportingBlock, 0, len(fields)),
		SourceRefs:              []string{},
		SupportRefs:             []string{},
		DisclosurePartitionRefs: []string{},
		SectionValidation: ReportingSectionValidation{
			Result: "passed",
		},
	}
	type sourceRecordKey struct{ family, id string }
	recordsByKey := map[sourceRecordKey]reportingSourceGroup{}
	relationships := []ReportingRelationshipSummary{}
	timelineEvents := []ReportingTimelineEvent{}
	sortFacts := map[string]exportprovider.TimelineSortFact{}
	for _, fact := range source.Timeline {
		if err := ctx.Err(); err != nil {
			return SnapshotModel{}, "", err
		}
		if fact.Validate() != nil {
			return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "timeline_sort_key_missing")
		}
		if _, exists := sortFacts[fact.RecordID]; exists {
			return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "duplicate_stable_id")
		}
		sortFacts[fact.RecordID] = fact.Clone()
	}
	supportIndexByID := map[string]ReportingSupportRef{}
	projectedPaths := map[string]bool{}
	sectionRefs := emptyContainerRefs()

	for index, field := range fields {
		if err := ctx.Err(); err != nil {
			return SnapshotModel{}, "", err
		}
		reportingFields, sourceRefs, err := projectReportingSourceFields(field)
		if err != nil {
			return SnapshotModel{}, "", err
		}
		for _, ref := range sourceRefs {
			if projectedPaths[ref] {
				return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
			}
			projectedPaths[ref] = true
			owner := sourceOwner{field.SourceFamily, field.SourceRecordID}
			if owner.family == "incident_metadata" {
				owner.id = incidentID
			}
			sourceOwners[ref] = owner
		}
		supportRefs := reportingFields[0].SupportRefs
		blockID := fmt.Sprintf("sec-0001.b%04d", index+1)
		block := ReportingBlock{
			SchemaID:                ReportingBlockSchemaID,
			BlockID:                 blockID,
			BlockKind:               "paragraph",
			BlockOrdinal:            index + 1,
			ContentClass:            field.ContentClass,
			Fields:                  reportingFields,
			Children:                []ReportingBlock{},
			SourceRefs:              sourceRefs,
			SupportRefs:             supportRefs,
			DisclosurePartitionRefs: reportingFields[0].DisclosurePartitionRefs,
		}
		section.Blocks = append(section.Blocks, block)
		sectionRefs.include(sourceRefs, supportRefs, block.DisclosurePartitionRefs)

		recordID := field.SourceRecordID
		if recordID != "" {
			sourceRef := SourceRecordRef{
				SchemaID:         "source_record_ref.v1",
				SourceFamily:     field.SourceFamily,
				SourceRecordID:   recordID,
				SourceSnapshotID: snapshotID,
			}
			recordKey := sourceRecordKey{field.SourceFamily, recordID}
			record := recordsByKey[recordKey]
			if record.RecordID != "" && record.RecordType != field.SourceRecordType {
				return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "duplicate_stable_id")
			}
			if record.RecordID == "" {
				record = reportingSourceGroup{
					RecordID:                recordID,
					RecordType:              field.SourceRecordType,
					SourceRecordRef:         sourceRef,
					Fields:                  []ReportingField{},
					SourceRefs:              []string{},
					SupportRefs:             []string{},
					DisclosurePartitionRefs: append([]string{}, reportingFields[0].DisclosurePartitionRefs...),
				}
			}
			record.Fields = append(record.Fields, reportingFields...)
			record.SourceRefs = appendUniqueStrings(record.SourceRefs, sourceRefs...)
			record.SupportRefs = appendUniqueStrings(record.SupportRefs, supportRefs...)
			record.DisclosurePartitionRefs = orderedDisclosurePartitions(append(record.DisclosurePartitionRefs, field.DisclosurePartitionRefs...))
			recordsByKey[recordKey] = record

		}
		for _, supportRef := range supportRefs {
			target, exists := supportTargets[supportRef]
			if !exists {
				return SnapshotModel{}, "", invalidSupportRef()
			}
			supportID := supportRef
			if _, exists := supportIndexByID[supportID]; !exists {
				supportIndexByID[supportID] = ReportingSupportRef{
					SchemaID:                "cartulary.reporting_support_ref.v1",
					SupportRefID:            supportID,
					SupportKind:             target.kind,
					SupportTargetRef:        target.ref,
					SourceSnapshotID:        snapshotID,
					SupportRole:             "corroborating",
					CustodyState:            "not_applicable",
					DisclosurePartitionRefs: []string{},
				}
			}
		}
		if field.SourceFamily == "record_link" {
			relationships = append(relationships, relationshipSummaryFromField(field, snapshotID, ReportingField{SourceRefs: sourceRefs, SupportRefs: supportRefs}))
		}
	}
	var summaryErr error
	sectionRefs = sectionRefs.normalized()
	section.SourceRefs, section.SupportRefs, section.DisclosurePartitionRefs = sectionRefs.source, sectionRefs.support, sectionRefs.partitions
	section.ContentClassSummary, summaryErr = summarizeReportingBlocks(section.Blocks)
	if summaryErr != nil {
		return SnapshotModel{}, "", summaryErr
	}
	groups := make([]reportingSourceGroup, 0, len(recordsByKey))
	subjects := make([]TokenizableSubject, 0, len(recordsByKey))
	for key, record := range recordsByKey {
		if err := ctx.Err(); err != nil {
			return SnapshotModel{}, "", err
		}
		if err := record.normalize(); err != nil {
			return SnapshotModel{}, "", err
		}
		groups = append(groups, record)
		kind, err := snapshotRecordSubjectKind(record)
		if err != nil {
			return SnapshotModel{}, "", err
		}
		if kind != "" {
			// Derive once from the complete source record. A later field must
			// neither replace its binding nor discard earlier disclosure labels.
			subject, err := materializeReportingSubject(record, kind)
			if err != nil {
				return SnapshotModel{}, "", err
			}
			subjects = append(subjects, subject)
		}
		if key.family == "timeline_event" && record.RecordType == "timeline_event" {
			fact, exists := sortFacts[record.RecordID]
			if !exists {
				return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "timeline_sort_key_missing")
			}
			delete(sortFacts, record.RecordID)
			timelineEvents = append(timelineEvents, ReportingTimelineEvent{
				SchemaID: "cartulary.reporting_timeline_event.v1",
				RecordID: fact.RecordID, ActivitySortKey: timelineSortKey(fact),
				ActivitySortTS: fact.ActivitySortTS, ActivityParseState: fact.ActivityParseState,
				ActivityPrecisionRank: fact.ActivityPrecisionRank, DateEnteredSortKey: fact.DateEnteredSortKey,
				RecordCreatedAt:         fact.RecordCreatedAt,
				DisplayTimes:            ReportingDisplayTimes{},
				Fields:                  append([]ReportingField(nil), record.Fields...),
				SourceRefs:              append([]string{}, record.SourceRefs...),
				SupportRefs:             append([]string{}, record.SupportRefs...),
				DisclosurePartitionRefs: append([]string{}, record.DisclosurePartitionRefs...),
			})
		}
	}
	if len(sortFacts) != 0 {
		return SnapshotModel{}, "", newRenderValidationError("export_model_invalid", "dangling_source_ref")
	}
	records, err := coreRecordSummaries(groups)
	if err != nil {
		return SnapshotModel{}, "", err
	}
	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].RelationshipID < relationships[j].RelationshipID
	})
	sort.Slice(timelineEvents, func(i, j int) bool {
		return compareTimelineEvents(timelineEvents[i], timelineEvents[j], false, false) < 0
	})
	sort.Slice(subjects, func(i, j int) bool {
		return subjects[i].StableSubjectRef < subjects[j].StableSubjectRef
	})
	supportIndex := make([]ReportingSupportRef, 0, len(supportIndexByID))
	for _, support := range supportIndexByID {
		supportIndex = append(supportIndex, support)
	}
	sort.Slice(supportIndex, func(i, j int) bool {
		return supportIndex[i].SupportRefID < supportIndex[j].SupportRefID
	})
	snapshotIdentity, err := snapshotModelID(snapshotID, DerivationVersion, snapshotAt)
	if err != nil {
		return SnapshotModel{}, "", err
	}
	createdAt := snapshotAt.UTC()
	model := SnapshotModel{
		SourceFacts:     source.SourceFacts.Clone(),
		SchemaID:        SnapshotModelSchemaID,
		SnapshotModelID: snapshotIdentity,
		SnapshotContent: SnapshotContent{
			IncidentID: incidentID,
			SnapshotID: snapshotID,
			SnapshotAt: createdAt,

			SourceChangeSetHighWatermark: watermark,
			DerivationVersion:            DerivationVersion,

			ExportModelGeneratorID:      "cartulary.reporting.materializer",
			ExportModelGeneratorVersion: "1",

			Sections:       []ReportingSection{section},
			Records:        records,
			Relationships:  relationships,
			TimelineEvents: timelineEvents,
			Subjects:       subjects,
			Diagrams:       []ReportingDiagram{},
			Assets:         []ReportingAssetDeclaration{},
			SupportIndex:   supportIndex,
			ValidationSummary: ReportingExportModelValidation{
				SchemaID: "cartulary.reporting_export_model_validation.v1",
				Result:   "passed",
				Issues:   []ReportingValidationIssue{},
			},
		},
	}
	if err := materializeSnapshotSourceIndex(&model.SnapshotContent, sourceOwners, groups); err != nil {
		return SnapshotModel{}, "", err
	}
	if err := snapshotLifecycleFields(&model.SnapshotContent, true); err != nil {
		return SnapshotModel{}, "", err
	}
	if err := ctx.Err(); err != nil {
		return SnapshotModel{}, "", err
	}
	if err := validateSnapshotModelIdentity(model); err != nil {
		return SnapshotModel{}, "", err
	}
	encoded, err := canonicalJSON(model)
	if err != nil {
		return SnapshotModel{}, "", err
	}
	return model, reportingObjectDigest(model.SchemaID, encoded), nil
}

func (model SnapshotContent) RedactionFields() ([]ExportField, error) {
	if err := validateSnapshotContent(model); err != nil {
		return nil, err
	}
	fields := []ExportField{}
	for _, section := range model.Sections {
		part, err := redactionFieldsFromBlocks(section.Blocks)
		if err != nil {
			return nil, err
		}
		fields = append(fields, part...)
	}
	groups, err := snapshotSourceGroups(model)
	if err != nil {
		return nil, err
	}
	byPath := map[string]reportingSourceGroup{}
	for _, record := range groups {
		if record.SourceRecordRef.SourceSnapshotID != model.SnapshotID {
			return nil, newRenderValidationError("export_model_invalid", "snapshot_ref_mismatch")
		}
		for _, path := range record.SourceRefs {
			if _, exists := byPath[path]; exists {
				return nil, newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
			}
			byPath[path] = record
		}
	}
	index := snapshotSourceIndex(model)
	for i := range fields {
		record, exists := byPath[fields[i].Path]
		if !exists {
			ref, found := index[fields[i].Path]
			if !found || ref.SourceFamily != "incident_metadata" || ref.SourceRecordID != model.IncidentID {
				return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			fields[i].SourceFamily = "incident_metadata"
			continue
		}
		fields[i].SourceFamily = record.SourceRecordRef.SourceFamily
		fields[i].SourceRecordID = record.SourceRecordRef.SourceRecordID
		fields[i].SourceRecordType = record.RecordType
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Path < fields[j].Path })
	return fields, nil
}

func redactionFieldsFromBlocks(blocks []ReportingBlock) ([]ExportField, error) {
	fields := []ExportField{}
	for _, block := range blocks {
		if err := validateSourceContentClass(block.ContentClass); err != nil {
			return nil, err
		}
		for _, field := range block.Fields {
			if _, valid := admitReportingFieldValue(field.Value, block.BlockKind == "paragraph"); !valid {
				return nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
			}
			if len(field.SourceRefs) == 0 || field.SourceRefs[0] == "" {
				return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			fields = append(fields, ExportField{
				Path: field.SourceRefs[0], ContentClass: block.ContentClass,
				Value: field.Value, DisclosurePartitionRefs: append([]string{}, field.DisclosurePartitionRefs...), SupportRefs: cloneStrings(field.SupportRefs),
			})
		}
		children, err := redactionFieldsFromBlocks(block.Children)
		if err != nil {
			return nil, err
		}
		fields = append(fields, children...)
	}
	return fields, nil
}

func validateSourceContentClass(class string) error {
	if class == "" {
		return newRenderValidationError("content_class_missing", "content_class_missing")
	}
	if !validContentClass(class) {
		return newRenderValidationError("export_model_invalid", "invalid_scalar_value")
	}
	return nil
}

func validContentClass(class string) bool {
	switch class {
	case ContentClassCuratedNarrative, ContentClassDerivedAnalytic, ContentClassSourceEvidence, ContentClassWorkingMaterial:
		return true
	default:
		return false
	}
}

func cloneExportFields(fields []ExportField) []ExportField {
	if len(fields) == 0 {
		return nil
	}
	out := make([]ExportField, len(fields))
	for i, field := range fields {
		out[i] = field
		out[i].DisclosurePartitionRefs = orderedDisclosurePartitions(field.DisclosurePartitionRefs)
		out[i].SupportRefs = cloneStrings(field.SupportRefs)
	}
	return out
}

func appendUniqueStrings(values []string, additions ...string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		seen[value] = struct{}{}
	}
	for _, value := range additions {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		values = append(values, value)
		seen[value] = struct{}{}
	}
	sort.Strings(values)
	return values
}

func sourceValueState(value any) string {
	if value == nil {
		return "null"
	}
	return "present"
}

// Subject family is resolved from the complete assembled source record. A
// provider may contribute disjoint field sets without repeating identity facts.
func snapshotRecordSubjectKind(record reportingSourceGroup) (string, error) {
	switch record.SourceRecordRef.SourceFamily {
	case "host", "identity", "party":
		return record.SourceRecordRef.SourceFamily, nil
	case "entity_mention":
		binding, err := reportingMentionSource(record)
		if err != nil {
			return "", err
		}
		if binding.ResolvedRecordID == nil {
			return binding.Family, nil
		}
	}
	return "", nil
}

func relationshipSummaryFromField(field ExportField, snapshotID string, reportingField ReportingField) ReportingRelationshipSummary {
	relationshipID := field.SourceRecordID
	value, _ := field.Value.(map[string]any)
	srcID, _ := value["src_record_id"].(string)
	dstID, _ := value["dst_record_id"].(string)
	kind, _ := value["link_type"].(string)
	if kind == "" {
		kind = "record_link"
	}
	return ReportingRelationshipSummary{
		SchemaID:         "cartulary.reporting_relationship_summary.v1",
		RelationshipID:   relationshipID,
		RelationshipKind: kind,
		SrcRecordRef: RelationshipEndpointRef{
			SchemaID:     "relationship_endpoint_ref.v1",
			EndpointRole: "src",
			SourceRecordRef: SourceRecordRef{
				SchemaID:         "source_record_ref.v1",
				SourceFamily:     "record_envelope",
				SourceRecordID:   srcID,
				SourceSnapshotID: snapshotID,
			},
		},
		DstRecordRef: RelationshipEndpointRef{
			SchemaID:     "relationship_endpoint_ref.v1",
			EndpointRole: "dst",
			SourceRecordRef: SourceRecordRef{
				SchemaID:         "source_record_ref.v1",
				SourceFamily:     "record_envelope",
				SourceRecordID:   dstID,
				SourceSnapshotID: snapshotID,
			},
		},
		Direction:               "directed",
		SourceRefs:              cloneStrings(reportingField.SourceRefs),
		SupportRefs:             append([]string{}, reportingField.SupportRefs...),
		DisclosurePartitionRefs: orderedDisclosurePartitions(field.DisclosurePartitionRefs),
	}
}

func sourceFamilyForPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "incident":
		return "incident_metadata"
	case "record_envelopes":
		return "record_envelope"
	case "timeline":
		return "timeline_event"
	case "hosts":
		return "host"
	case "identities":
		return "identity"
	case "parties":
		return "party"
	case "evidence":
		return "evidence"
	case "task_requests":
		return "task_request"
	case "decisions":
		return "decision"
	case "notes":
		return "note"
	case "findings":
		return "finding_hypothesis"
	case "comm_log":
		return "comm_log"
	case "handoffs":
		return "handoff"
	case "status_reviews":
		return "status_review"
	case "lessons":
		return "lesson"
	case "relationships":
		return "record_link"
	case "tags":
		return "record_tag"
	case "entity_mentions":
		return "entity_mention"
	default:
		return ""
	}
}

// ResolveRedactionProfile materializes the exact profile against the retained
// snapshot. Subject rules are explicit paths, never implicit mask behavior.
func ResolveRedactionProfile(id string, version string, recipientPartitionRefs []string, snapshot SnapshotContent) (RedactionProfile, string, error) {
	profile, err := selectedRedactionProfile(id, version, recipientPartitionRefs)
	if err != nil {
		return RedactionProfile{}, "", err
	}
	return materializeRedactionProfile(profile, snapshot)
}

func selectedRedactionProfile(id, version string, recipientPartitionRefs []string) (RedactionProfile, error) {
	if version != DefaultRedactionProfileVersion {
		return RedactionProfile{}, ErrInvalidRedactionProfile
	}
	switch id {
	case InternalRedactionProfileID:
		return internalRedactionProfile(), nil
	case ExternalRedactionProfileID:
		return externalRedactionProfile(recipientPartitionRefs), nil
	case TokenizedRedactionProfileID:
		return tokenizedReviewRedactionProfile(), nil
	default:
		return RedactionProfile{}, ErrInvalidRedactionProfile
	}
}

// Materialize only the retained authored policy. A queued attempt never resolves
// a selector against a newer registry or substitutes a current policy document.
func materializeRedactionProfile(profile RedactionProfile, snapshot SnapshotContent) (RedactionProfile, string, error) {
	id := profile.ProfileID
	if id == TokenizedRedactionProfileID {
		modelPaths, err := redactionFieldModelPaths(snapshot)
		if err != nil {
			return RedactionProfile{}, "", err
		}
		subjectIndex, err := redactionSubjectIndex(snapshot)
		if err != nil {
			return RedactionProfile{}, "", err
		}
		fields, err := snapshot.RedactionFields()
		if err != nil {
			return RedactionProfile{}, "", err
		}
		for _, field := range fields {
			subjects := subjectIndex[field.Path]
			if len(subjects) == 0 || field.Value == nil {
				continue
			}
			if len(subjects) != 1 {
				return RedactionProfile{}, "", newRenderValidationError("token_manifest_invalid", "token_subject_not_unique")
			}
			path := modelPaths[field.Path]
			if path == "" {
				return RedactionProfile{}, "", newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			profile.Rules = append(profile.Rules, RedactionRule{
				RuleID: "tokenized-review-subject-" + hashHex([]byte(path)), Path: &path,
				Action: RedactionActionSpec{Type: ActionMask, TokenBacked: true},
			})
		}
	}
	if id == ExternalRedactionProfileID || id == TokenizedRedactionProfileID {
		fields, err := snapshot.RedactionFields()
		if err != nil {
			return RedactionProfile{}, "", err
		}
		paths, err := redactionFieldModelPaths(snapshot)
		if err != nil {
			return RedactionProfile{}, "", err
		}
		sources := reportingFieldSources(snapshot)
		for _, field := range fields {
			// Null is an explicit source value, not text to mask or truncate.
			// Working material retains the external profile's drop decision.
			if sources[field.Path].Field.SourceValueState != "null" || id == ExternalRedactionProfileID && field.ContentClass == ContentClassWorkingMaterial {
				continue
			}
			path := paths[field.Path]
			profile.Rules = append(profile.Rules, RedactionRule{
				RuleID: "null-value-allow-" + hashHex([]byte(path)), Path: &path,
				Action: RedactionActionSpec{Type: ActionAllow},
			})
		}
	}
	if id == InternalRedactionProfileID || id == TokenizedRedactionProfileID {
		allowed, err := effectiveDisclosurePartitions(snapshot, ReleaseScopeInternalReview, nil)
		if err != nil {
			return RedactionProfile{}, "", err
		}
		profile.AllowedDisclosurePartitionRefs = allowed
	}
	containers, err := redactionTreeCandidates(snapshot)
	if err != nil {
		return RedactionProfile{}, "", err
	}
	// Built-ins transform values, retaining their typed envelopes. Freeze that
	// policy as exact rules; class masking must not implicitly rewrite a block.
	for _, candidate := range containers {
		if candidate.kind == "field" {
			continue
		}
		path := candidate.path
		profile.Rules = append(profile.Rules, RedactionRule{RuleID: "container-allow-" + hashHex([]byte(path)), Path: &path, Action: RedactionActionSpec{Type: ActionAllow}})
	}
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		return RedactionProfile{}, "", err
	}
	return profile, digest, nil
}

func ValidateRedactionProfile(profile RedactionProfile) (string, error) {
	if profile.SchemaID != RedactionProfileSchemaID || !validReportingIdentifier(profile.ProfileID) || !validReportingIdentifier(profile.Version) {
		return "", fmt.Errorf("%w: missing identity", ErrInvalidRedactionProfile)
	}
	if err := validateProfilePolicy(profile.ProfileStatus, profile.AllowedReleaseScopes); err != nil {
		return "", err
	}
	if err := validateAction(profile.DefaultAction); err != nil {
		return "", err
	}
	if err := validateDisclosurePartitionSet(profile.AllowedDisclosurePartitionRefs); err != nil {
		return "", err
	}
	ruleIDs := map[string]bool{"profile_default": true}
	paths := map[string]string{}
	classes := map[string]string{}
	for _, rule := range profile.Rules {
		if !validReportingIdentifier(rule.RuleID) || ruleIDs[rule.RuleID] {
			return "", fmt.Errorf("%w: missing or duplicate rule id", ErrInvalidRedactionProfile)
		}
		ruleIDs[rule.RuleID] = true
		if err := validateAction(rule.Action); err != nil {
			return "", err
		}
		hasPath := rule.Path != nil
		hasClass := rule.ContentClass != nil
		if hasPath == hasClass {
			return "", fmt.Errorf("%w: rule %s must select exactly one target", ErrInvalidRedactionProfile, rule.RuleID)
		}
		if hasPath {
			if !validExactExportModelPath(*rule.Path) {
				return "", profilePolicyError()
			}
			if previous, ok := paths[*rule.Path]; ok {
				return "", fmt.Errorf("%w: duplicate path rule %s conflicts with %s", ErrInvalidRedactionProfile, rule.RuleID, previous)
			}
			paths[*rule.Path] = rule.RuleID
		}
		if hasClass {
			if !validReportingContentClass(*rule.ContentClass) {
				return "", fmt.Errorf("%w: invalid content class selector", ErrInvalidRedactionProfile)
			}
			if previous, ok := classes[*rule.ContentClass]; ok {
				return "", fmt.Errorf("%w: duplicate class rule %s conflicts with %s", ErrInvalidRedactionProfile, rule.RuleID, previous)
			}
			classes[*rule.ContentClass] = rule.RuleID
		}
	}
	encoded, err := canonicalJSON(profile)
	if err != nil {
		return "", err
	}
	return reportingObjectDigest(profile.SchemaID, encoded), nil
}

func RedactExportModel(model ExportModel, facts exportprovider.SourceFacts, profile RedactionProfile, profileSHA256 string, sourceExportSHA256 string, releaseScope string, recipientPartitionRefs []string) (RedactionResult, error) {
	return redactExportModel(context.Background(), model, facts, profile, profileSHA256, sourceExportSHA256, releaseScope, recipientPartitionRefs, renderSelectionInput{})
}

func redactExportModel(ctx context.Context, model ExportModel, facts exportprovider.SourceFacts, profile RedactionProfile, profileSHA256 string, sourceExportSHA256 string, releaseScope string, recipientPartitionRefs []string, composition renderSelectionInput) (RedactionResult, error) {
	if err := ctx.Err(); err != nil {
		return RedactionResult{}, err
	}
	if digest, err := ValidateRedactionProfile(profile); err != nil || digest != profileSHA256 {
		return RedactionResult{}, fmt.Errorf("%w: profile digest mismatch", ErrInvalidRedactionProfile)
	}
	profileView, profileViewJSON, profileViewSHA, err := BuildRedactionProfileView(profile, profileSHA256)
	if err != nil {
		return RedactionResult{}, err
	}
	fields, err := model.RedactionFields()
	if err != nil {
		return RedactionResult{}, err
	}
	fields, err = assignSubjectFieldPartitions(ctx, model.SnapshotContent, facts, fields, releaseScope)
	if err != nil {
		return RedactionResult{}, err
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	modelPaths, err := redactionFieldModelPaths(model.SnapshotContent)
	if err != nil {
		return RedactionResult{}, err
	}
	if err := validateDeletedRecordBindings(facts, composition); err != nil {
		return RedactionResult{}, err
	}
	if err := validateNarrativeSourceBindings(composition.template, model.SnapshotContent); err != nil {
		return RedactionResult{}, err
	}
	restrictions := map[string]lifecycleRestriction{}
	for _, field := range fields {
		restrictions[field.Path] = field.lifecycle
	}
	bindings, err := templateFieldBindings(model.SnapshotContent, modelPaths)
	if err != nil {
		return RedactionResult{}, err
	}
	fields, err = selectReachableSourceFields(ctx, model, facts, fields, bindings, profile, releaseScope, recipientPartitionRefs, composition)
	if err != nil {
		return RedactionResult{}, err
	}
	eligibleSource := eligibleSourceTree(model.SnapshotContent, fields)
	fields, containers, err := selectContainerRedactions(ctx, eligibleSource, fields, profileView)
	if err != nil {
		return RedactionResult{}, err
	}
	effective, err := effectiveDisclosurePartitions(model.SnapshotContent, releaseScope, recipientPartitionRefs)
	if err != nil {
		return RedactionResult{}, err
	}
	selected, err := selectRedactionFields(fields, modelPaths, profileView, releaseScope, effective)
	if err != nil {
		return RedactionResult{}, err
	}
	sourceFields := reportingFieldSources(model.SnapshotContent)
	authoredFields, authored, err := prepareCompositionTextFields(composition, model, selected, sourceFields, modelPaths)
	if err != nil {
		return RedactionResult{}, err
	}
	for path, source := range composition.graphs.sources {
		if _, exists := modelPaths[path]; exists {
			return RedactionResult{}, newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
		}
		modelPaths[path], sourceFields[path] = path, source
	}
	assets, assetSelected, err := selectTemplateAssets(composition.template, modelPaths, profileView, releaseScope, effective)
	if err != nil {
		return RedactionResult{}, err
	}
	if err := validateRedactionPathBindings(modelPaths, profileView, model.SnapshotContent); err != nil {
		return RedactionResult{}, err
	}
	authoredSelected, err := selectRedactionFields(authoredFields, modelPaths, profileView, releaseScope, effective)
	if err != nil {
		return RedactionResult{}, err
	}
	selected = append(selected, authoredSelected...)
	graphSelected, err := selectRedactionFields(composition.graphs.fields, modelPaths, profileView, releaseScope, effective)
	if err != nil {
		return RedactionResult{}, err
	}
	selected = append(selected, graphSelected...)
	selected = append(selected, assetSelected...)
	subjectBindings := map[string][]TokenizableSubject{}
	for path, text := range authored {
		subjectBindings[path] = text.subjects
	}
	tokens, err := assignDisplayTokens(model, selected, profile.NeutralTokenFamily, subjectBindings)
	if err != nil {
		return RedactionResult{}, err
	}
	if err := resolveCompositionTextValues(composition, model, selected, tokens, sourceFields, authored); err != nil {
		return RedactionResult{}, err
	}
	identity, err := retainedExportIdentity(model)
	if err != nil {
		return RedactionResult{}, err
	}
	identity.ReleaseScope = releaseScope
	identity.RecipientPartitionRefs = orderedDisclosurePartitions(recipientPartitionRefs)
	redacted := RedactedExportModel{
		ExportModel: identity, admittedPresentationTexts: map[string]string{},
		admittedFieldPaths: modelPaths, lifecycleRestrictions: restrictions,
		admittedExportModelSHA256: sourceExportSHA256,
		Fields:                    []RedactedField{},
	}
	redacted.admittedPartyIDs = make([]string, 0, len(facts.Parties))
	for _, party := range facts.Parties {
		redacted.admittedPartyIDs = append(redacted.admittedPartyIDs, party.PartyID)
	}
	redacted.subjectDisplayTokens = map[string]string{}
	for _, token := range tokens {
		redacted.subjectDisplayTokens[token.StableSubjectRef] = token.DisplayToken
	}
	entries := make([]RedactionManifestEntry, 0, len(fields))
	retainedFields := map[string]ReportingField{}
	for _, selection := range selected {
		if err := ctx.Err(); err != nil {
			return RedactionResult{}, err
		}
		field, ruleID, action, selectedTrace := selection.Field, selection.RuleID, selection.Action, selection.Trace
		value, include, outcome := selection.Value, selection.Include, selection.Outcome
		authoredText, isAuthored := authored[field.Path]
		if isAuthored {
			encoded, err := canonicalJSON(authoredText.text)
			if err != nil {
				return RedactionResult{}, err
			}
			redacted.admittedPresentationTexts[field.Path] = hashHex(encoded)
		}
		var tokenID *string
		if action.TokenBacked {
			token := tokens[field.Path]
			id := token.TokenID
			tokenID = &id
			value, include, outcome = token.DisplayToken, true, "tokenized"
		}
		if asset, isAsset := assets[field.Path]; include && isAsset {
			redacted.Assets = append(redacted.Assets, asset)
		} else if include {
			source, exists := sourceFields[field.Path]
			if !exists {
				return RedactionResult{}, newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			retained, err := retainedReportingField(source.Field, value, outcome, selectedTrace.RulePrecedence, source.Multiline)
			if err != nil {
				return RedactionResult{}, err
			}
			retained.DisclosurePartitionRefs = orderedDisclosurePartitions(field.DisclosurePartitionRefs)
			if isAuthored {
				redacted.PresentationTexts = append(redacted.PresentationTexts, ReportingPresentationText{SchemaID: "cartulary.reporting_presentation_text.v1", AuthoredTextID: authoredText.text.AuthoredTextID, TextRole: authoredText.text.TextRole, ContentClass: ContentClassPresentationText, Body: retained})
			} else if _, graphField := composition.graphs.sources[field.Path]; !graphField {
				retainedFields[field.Path] = retained
			}
			value = retained.Value
		}
		// Source identity has finished its role in subject resolution. Public
		// artifacts and token uses address the pre-removal model, never a
		// private source selector or a path reconstructed by the renderer.
		field.Path = modelPaths[field.Path]

		if include {
			redacted.Fields = append(redacted.Fields, RedactedField{
				Path:                    field.Path,
				ContentClass:            field.ContentClass,
				SourceFamily:            field.SourceFamily,
				Value:                   value,
				DisclosurePartitionRefs: orderedDisclosurePartitions(field.DisclosurePartitionRefs),
				SupportRefs:             cloneStrings(field.SupportRefs),
				RawBlobSource:           field.RawBlobSource,
				OpaqueBinary:            field.OpaqueBinary,
				GeneratedPresentation:   field.GeneratedPresentation,
			})

		}
		entries = append(entries, RedactionManifestEntry{
			Path:                        field.Path,
			ContentClass:                field.ContentClass,
			Action:                      action.Type,
			RuleID:                      ruleID,
			RulePrecedence:              selectedTrace.RulePrecedence,
			SelectedRuleTrace:           selectedTrace,
			DisclosurePartitionHandling: disclosurePartitionHandling(field, outcome),
			DisclosurePartitionRefs:     orderedDisclosurePartitions(field.DisclosurePartitionRefs),
			Outcome:                     outcome,
			TokenID:                     tokenID,
		})
	}
	sort.Slice(redacted.Fields, func(i, j int) bool { return redacted.Fields[i].Path < redacted.Fields[j].Path })
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	sortAssetDeclarations(redacted.Assets)
	sort.Slice(redacted.PresentationTexts, func(i, j int) bool {
		return redacted.PresentationTexts[i].AuthoredTextID < redacted.PresentationTexts[j].AuthoredTextID
	})
	redacted.Sections, err = projectRetainedSections(ctx, explicitlyRetainedTree(eligibleSource, containers).Sections, retainedFields)
	if err != nil {
		return RedactionResult{}, err
	}
	redacted.TimelineEvents, err = projectRetainedTimeline(ctx, model.TimelineEvents, retainedFields)
	if err != nil {
		return RedactionResult{}, err
	}
	redacted.Records, err = projectRetainedRecords(ctx, model.Records, retainedFields)
	if err != nil {
		return RedactionResult{}, err
	}
	if err := projectRetainedMetadata(ctx, model, &redacted, retainedFields, effective, profileView); err != nil {
		return RedactionResult{}, err
	}
	redactedBytes, err := canonicalJSON(redacted)
	if err != nil {
		return RedactionResult{}, err
	}
	modelSHA := reportingObjectDigest(redacted.SchemaID, redactedBytes)
	qualifyRetainedContainers(containers, redacted.Fields)
	removals, err := structuralRedactionRemovals(eligibleSource, structuralDecisionEntries(entries, containers))
	if err != nil {
		return RedactionResult{}, err
	}
	manifest := RedactionManifest{
		RenderIdentity: model.RenderIdentity, SnapshotID: model.SnapshotID, CreatedAt: model.RenderAdmittedAt,
		ContainerDecisions: containers,
		SchemaID:           RedactionManifestSchemaID,
		ProfileID:          profile.ProfileID,
		ProfileVersion:     profile.Version,
		ProfileSHA256:      profileSHA256,
		Entries:            entries,
		StructuralRemovals: removals,
	}
	if err := ValidateRedactionResult(redacted, manifest, releaseScope); err != nil {
		return RedactionResult{}, err
	}
	manifestBytes, err := canonicalJSON(manifest)
	if err != nil {
		return RedactionResult{}, err
	}
	manifestSHA := reportingObjectDigest(manifest.SchemaID, manifestBytes)
	tokenManifest, tokenManifestJSON, tokenManifestSHA, revealMap, revealMapJSON, revealMapSHA, err := buildTokenArtifacts(model, tokens, manifestSHA, profileSHA256, recipientPartitionRefs)
	if err != nil {
		return RedactionResult{}, err
	}
	if tokenManifest != nil {
		if err := validateTokenManifestBindings(*tokenManifest, manifest, manifestSHA); err != nil {
			return RedactionResult{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return RedactionResult{}, err
	}
	return RedactionResult{
		Model:               redacted,
		Manifest:            manifest,
		ProfileView:         profileView,
		TokenManifest:       tokenManifest,
		RevealMap:           revealMap,
		ModelSHA256:         modelSHA,
		ModelJSON:           redactedBytes,
		ManifestSHA256:      manifestSHA,
		ProfileViewSHA256:   profileViewSHA,
		TokenManifestSHA256: tokenManifestSHA,
		RevealMapSHA256:     revealMapSHA,
		ManifestJSON:        manifestBytes,
		ProfileViewJSON:     profileViewJSON,
		TokenManifestJSON:   tokenManifestJSON,
		RevealMapJSON:       revealMapJSON,
	}, nil
}

func ValidateRedactionResult(model RedactedExportModel, manifest RedactionManifest, releaseScope string) error {
	if validateRedactionManifest(manifest) != nil {
		return fmt.Errorf("%w: manifest_identity", ErrRedactionValidation)
	}
	fieldPaths := map[string]RedactedField{}
	for _, field := range model.Fields {
		if !validExactExportModelPath(field.Path) {
			return fmt.Errorf("%w: invalid_model_path", ErrRedactionValidation)
		}
		if fieldPaths[field.Path].Path != "" {
			return fmt.Errorf("%w: duplicate_redacted_path", ErrRedactionValidation)
		}
		fieldPaths[field.Path] = field
		if releaseScope == ReleaseScopeExternal {
			if field.RawBlobSource {
				return fmt.Errorf("%w: external_raw_blob_source", ErrRedactionValidation)
			}
			if field.OpaqueBinary {
				return fmt.Errorf("%w: external_opaque_binary", ErrRedactionValidation)
			}
			if field.ContentClass == ContentClassWorkingMaterial {
				return fmt.Errorf("%w: external_working_material", ErrRedactionValidation)
			}
			if field.ContentClass == ContentClassCuratedNarrative && len(field.SupportRefs) == 0 {
				return fmt.Errorf("%w: external_curated_narrative_missing_support_refs", ErrRedactionValidation)
			}
			if externalDirectSourceFamilyBlocked(field.SourceFamily, field.ContentClass) {
				return fmt.Errorf("%w: external_direct_source_family", ErrRedactionValidation)
			}
		}
	}
	manifestPaths := map[string]struct{}{}
	for _, entry := range manifest.Entries {
		if !validExactExportModelPath(entry.Path) || entry.ContentClass == "" || entry.Action == "" || entry.RuleID == "" {
			return fmt.Errorf("%w: manifest_entry_incomplete", ErrRedactionValidation)
		}
		if !validSelectedRuleTrace(entry.SelectedRuleTrace) || entry.SelectedRuleTrace.RuleID != entry.RuleID {
			return fmt.Errorf("%w: manifest_entry_trace_incomplete", ErrRedactionValidation)
		}
		if _, ok := manifestPaths[entry.Path]; ok {
			return fmt.Errorf("%w: duplicate_manifest_path", ErrRedactionValidation)
		}
		manifestPaths[entry.Path] = struct{}{}
		if field, exists := fieldPaths[entry.Path]; exists && (entry.Outcome == "dropped" || field.ContentClass != entry.ContentClass || !sameReportingValue(field.DisclosurePartitionRefs, entry.DisclosurePartitionRefs)) {
			return fmt.Errorf("%w: manifest_outcome_mismatch", ErrRedactionValidation)
		}
		if _, ok := fieldPaths[entry.Path]; !ok && !strings.HasPrefix(entry.Outcome, "dropped") {
			return fmt.Errorf("%w: manifest_outcome_mismatch", ErrRedactionValidation)
		}
	}
	for path := range fieldPaths {
		if _, exists := manifestPaths[path]; !exists {
			return fmt.Errorf("%w: missing_field_decision", ErrRedactionValidation)
		}
	}
	return nil
}

func fieldDisclosureAllowed(allowed map[string]bool, field ExportField) bool {
	for _, partition := range field.DisclosurePartitionRefs {
		if !allowed[partition] {
			return false
		}
	}
	return true
}

func externalDirectSourceFamilyBlocked(sourceFamily string, contentClass string) bool {
	if contentClass == ContentClassWorkingMaterial {
		return true
	}
	switch sourceFamily {
	case "task_request", "decision", "finding_hypothesis", "comm_log", "handoff", "status_review", "lesson", "note":
		return contentClass != ContentClassCuratedNarrative && contentClass != ContentClassDerivedAnalytic
	default:
		return false
	}
}

func validateAction(action RedactionActionSpec) error {
	if action.TokenBacked && (action.Type != ActionMask && action.Type != ActionStub || action.ReplacementText != nil || action.StubText != nil || action.MaxChars != nil) {
		return fmt.Errorf("%w: invalid token-backed action", ErrInvalidRedactionProfile)
	}
	// Parameters belong to one action only. Accepting ignored parameters makes
	// profile identity depend on values that never affect the selected outcome.
	if (action.ReplacementText != nil && action.Type != ActionMask) ||
		(action.StubText != nil && action.Type != ActionStub) ||
		(action.MaxChars != nil && action.Type != ActionTruncate) {
		return fmt.Errorf("%w: action parameter mismatch", ErrInvalidRedactionProfile)
	}
	for _, literal := range []*string{action.ReplacementText, action.StubText} {
		if literal != nil && !validReportingText(*literal, false) {
			return fmt.Errorf("%w: invalid action literal", ErrInvalidRedactionProfile)
		}
	}
	switch action.Type {
	case ActionAllow, ActionDrop, ActionMask, ActionStub:
		return nil
	case ActionTruncate:
		if action.MaxChars == nil || *action.MaxChars < 1 || *action.MaxChars > 240 {
			return fmt.Errorf("%w: truncate.max_chars", ErrInvalidRedactionProfile)
		}
		return nil
	case ActionHash:
		return fmt.Errorf("%w: hash action is reserved", ErrInvalidRedactionProfile)
	default:
		return fmt.Errorf("%w: unsupported action %q", ErrInvalidRedactionProfile, action.Type)
	}
}

func applyRedactionAction(value any, action RedactionActionSpec) (any, bool, string, error) {
	switch action.Type {
	case ActionAllow:
		return value, true, "unchanged", nil
	case ActionDrop:
		return nil, false, "dropped", nil
	case ActionMask:
		replacement := DefaultMaskText
		if action.ReplacementText != nil {
			replacement = *action.ReplacementText
		}
		return replacement, true, "masked", nil
	case ActionTruncate:
		if action.MaxChars == nil || *action.MaxChars < 1 {
			return nil, false, "", fmt.Errorf("%w: truncate.max_chars", ErrInvalidRedactionProfile)
		}
		text := formatFieldValue(value)
		runes := []rune(text)
		if len(runes) <= *action.MaxChars {
			return text, true, "truncated", nil
		}
		return string(runes[:*action.MaxChars]) + DefaultTruncateMark, true, "truncated", nil
	case ActionStub:
		stub := DefaultStubText
		if action.StubText != nil {
			stub = *action.StubText
		}
		return stub, true, "stubbed", nil
	case ActionHash:
		return nil, false, "", fmt.Errorf("%w: hash action is reserved", ErrInvalidRedactionProfile)
	default:
		return nil, false, "", fmt.Errorf("%w: unsupported action %q", ErrInvalidRedactionProfile, action.Type)
	}
}

func disclosurePartitionHandling(field ExportField, outcome string) string {
	if outcome == "dropped" {
		for _, partition := range field.DisclosurePartitionRefs {
			if partition == "blocked" {
				return "blocked_removed"
			}
		}
		return "dropped"
	}
	if outcome == "unchanged" {
		return "releasable"
	}
	return "redacted"
}

func (contract TemplateContract) SupportsOutputKind(kind string) bool {
	for _, supported := range contract.SupportedOutputKinds {
		if kind == supported {
			return true
		}
	}
	return false
}

func (contract TemplateContract) SupportsReleaseScope(scope string) bool {
	for _, supported := range contract.SupportedReleaseScopes {
		if scope == supported {
			return true
		}
	}
	return false
}

func ValidateSelfContainedOutput(kind string, output []byte) error {
	rendered := string(output)
	if cssRemoteAssetPattern.MatchString(rendered) ||
		htmlRemoteScriptSourcePattern.MatchString(rendered) ||
		htmlRemoteLinkHrefPattern.MatchString(rendered) ||
		htmlRemoteMediaPattern.MatchString(rendered) ||
		htmlRemoteStylePattern.MatchString(rendered) {
		return ErrRemoteRuntimeAsset
	}
	switch kind {
	case OutputKindSlidev:
		if markdownRemoteImagePattern.MatchString(rendered) {
			return ErrRemoteRuntimeAsset
		}
	}
	return nil
}

func formatFieldValue(value any) string {
	switch value.(type) {
	case map[string]any, []any:
		encoded, err := canonicalJSON(value)
		if err == nil {
			return string(encoded)
		}
	}
	return fmt.Sprint(value)
}

func canonicalJSON(value any) ([]byte, error) {
	return encodeCanonicalJSON(value)
}

var encodeCanonicalJSON = marshalReportingJSON

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	sort.Strings(out)
	return out
}

func internalRedactionProfile() RedactionProfile {
	return RedactionProfile{
		ProfileStatus:        "adopted_current",
		AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview},
		SchemaID:             RedactionProfileSchemaID,
		ProfileID:            InternalRedactionProfileID,
		Version:              DefaultRedactionProfileVersion,
		AllowedDisclosurePartitionRefs: []string{
			"public",
			"internal_only",
		},
		DefaultAction: RedactionActionSpec{
			Type: ActionAllow,
		},
		Rules: []RedactionRule{},
	}
}

func tokenizedReviewRedactionProfile() RedactionProfile {
	sourceClass := ContentClassSourceEvidence
	return RedactionProfile{
		ProfileStatus:        "adopted_current",
		AllowedReleaseScopes: []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview},
		SchemaID:             RedactionProfileSchemaID,
		ProfileID:            TokenizedRedactionProfileID,
		NeutralTokenFamily:   true,
		Version:              DefaultRedactionProfileVersion,
		AllowedDisclosurePartitionRefs: []string{
			"public",
			"internal_only",
		},
		DefaultAction: RedactionActionSpec{
			Type: ActionAllow,
		},
		Rules: []RedactionRule{
			{
				RuleID:       "tokenized-review-source-subject-mask",
				ContentClass: &sourceClass,
				Action: RedactionActionSpec{
					Type: ActionMask,
				},
			},
		},
	}
}

func externalRedactionProfile(recipientPartitionRefs []string) RedactionProfile {
	sourceClass := ContentClassSourceEvidence
	workingClass := ContentClassWorkingMaterial
	maxChars := 120
	return RedactionProfile{
		ProfileStatus:                  "adopted_current",
		AllowedReleaseScopes:           []string{ReleaseScopeExternal},
		SchemaID:                       RedactionProfileSchemaID,
		ProfileID:                      ExternalRedactionProfileID,
		Version:                        DefaultRedactionProfileVersion,
		AllowedDisclosurePartitionRefs: orderedDisclosurePartitions(append([]string{"public"}, recipientPartitionRefs...)),
		DefaultAction: RedactionActionSpec{
			Type: ActionAllow,
		},
		Rules: []RedactionRule{
			{
				RuleID:       "external-source-evidence-truncate",
				ContentClass: &sourceClass,
				Action: RedactionActionSpec{
					Type:     ActionTruncate,
					MaxChars: &maxChars,
				},
			},
			{
				RuleID:       "external-working-material-drop",
				ContentClass: &workingClass,
				Action: RedactionActionSpec{
					Type: ActionDrop,
				},
			},
		},
	}
}

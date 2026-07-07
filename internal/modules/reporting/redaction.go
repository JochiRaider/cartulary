package reporting

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	InternalRedactionProfileID  = "cartulary.redaction.internal"
	ExternalRedactionProfileID  = "cartulary.redaction.external"
	TokenizedRedactionProfileID = "cartulary.redaction.tokenized_review"

	RedactionProfileSchemaID       = "cartulary.redaction_profile.v1"
	RedactionProfileViewSchemaID   = "cartulary.redaction_profile_view.v1"
	RedactionManifestSchemaID      = "cartulary.redaction_manifest.v1"
	RedactionTokenManifestSchemaID = "cartulary.redaction_token_manifest.v1"
	RedactionRevealMapSchemaID     = "cartulary.redaction_reveal_map.v1"

	ActionAllow    = "allow"
	ActionDrop     = "drop"
	ActionMask     = "mask"
	ActionTruncate = "truncate"
	ActionStub     = "stub"
	ActionHash     = "hash"

	DefaultMaskText     = "[REDACTED]"
	DefaultStubText     = "[STUB]"
	DefaultTruncateMark = "[TRUNCATED]"

	ContentClassCuratedNarrative = "curated_narrative"
	ContentClassDerivedAnalytic  = "derived_analytic"
	ContentClassSourceEvidence   = "source_evidence"
	ContentClassWorkingMaterial  = "working_material"
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
	Type            string  `json:"type"`
	ReplacementText *string `json:"replacement_text,omitempty"`
	StubText        *string `json:"stub_text,omitempty"`
	MaxChars        *int    `json:"max_chars,omitempty"`
}

type ExportModel struct {
	SchemaID                     string                         `json:"schema_id"`
	ExportModelID                string                         `json:"export_model_id"`
	IncidentID                   string                         `json:"incident_id"`
	SnapshotID                   string                         `json:"snapshot_id"`
	SnapshotAt                   time.Time                      `json:"snapshot_at"`
	RenderAdmittedAt             time.Time                      `json:"render_admitted_at"`
	SourceChangeSetHighWatermark string                         `json:"source_change_set_high_watermark"`
	SnapshotBoundaryKind         *string                        `json:"snapshot_boundary_kind"`
	DerivationVersion            string                         `json:"derivation_version"`
	ExportModelCreatedAt         time.Time                      `json:"export_model_created_at"`
	ExportModelGeneratorID       string                         `json:"export_model_generator_id"`
	ExportModelGeneratorVersion  string                         `json:"export_model_generator_version"`
	ReleaseScope                 string                         `json:"release_scope"`
	RecipientPartitionRefs       []string                       `json:"recipient_partition_refs"`
	Sections                     []ReportingSection             `json:"sections"`
	Records                      []ReportingRecordSummary       `json:"records"`
	Relationships                []ReportingRelationshipSummary `json:"relationships"`
	TimelineEvents               []ReportingTimelineEvent       `json:"timeline_events"`
	Subjects                     []TokenizableSubject           `json:"subjects"`
	Diagrams                     []ReportingDiagram             `json:"diagrams"`
	Assets                       []ReportingAssetDeclaration    `json:"assets"`
	SupportIndex                 []ReportingSupportRef          `json:"support_index"`
	ValidationSummary            ReportingExportModelValidation `json:"validation_summary"`
	Fields                       []ExportField                  `json:"-"`
}

type ExportField struct {
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
	SchemaID                    string           `json:"schema_id"`
	BlockID                     string           `json:"block_id"`
	BlockKind                   string           `json:"block_kind"`
	BlockOrdinal                int              `json:"block_ordinal"`
	ParentBlockID               *string          `json:"parent_block_id"`
	SplitFromBlockID            *string          `json:"split_from_block_id"`
	ContentClass                string           `json:"content_class"`
	AggregateOnlyNonIdentifying bool             `json:"aggregate_only_non_identifying"`
	AggregatePolicyID           *string          `json:"aggregate_policy_id"`
	ContributorCount            *int             `json:"contributor_count"`
	ExcludedFieldKeys           []string         `json:"excluded_field_keys"`
	Fields                      []ReportingField `json:"fields"`
	Children                    []ReportingBlock `json:"children"`
	SourceRefs                  []string         `json:"source_refs"`
	SupportRefs                 []string         `json:"support_refs"`
	DisclosurePartitionRefs     []string         `json:"disclosure_partition_refs"`
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
	SchemaID                string           `json:"schema_id"`
	TimelineEventID         string           `json:"timeline_event_id"`
	SourceRecordRef         SourceRecordRef  `json:"source_record_ref"`
	ActivitySortTS          *string          `json:"activity_sort_ts"`
	ActivitySortState       string           `json:"activity_sort_state"`
	DisplayTimes            map[string]any   `json:"display_times"`
	Fields                  []ReportingField `json:"fields"`
	SourceRefs              []string         `json:"source_refs"`
	SupportRefs             []string         `json:"support_refs"`
	DisclosurePartitionRefs []string         `json:"disclosure_partition_refs"`
}

type TokenizableSubject struct {
	SchemaID                string           `json:"schema_id"`
	StableSubjectRef        string           `json:"stable_subject_ref"`
	SubjectKind             string           `json:"subject_kind"`
	SourceRecordRef         *SourceRecordRef `json:"source_record_ref"`
	DisplayName             *string          `json:"display_name"`
	DisclosurePartitionRefs []string         `json:"disclosure_partition_refs"`
}

type ReportingDiagram struct {
	SchemaID  string `json:"schema_id"`
	DiagramID string `json:"diagram_id"`
}

type ReportingAssetDeclaration struct {
	SchemaID string `json:"schema_id"`
	AssetID  string `json:"asset_id"`
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
	CaseFact            int `json:"case_fact"`
	DerivedSummary      int `json:"derived_summary"`
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
	SchemaID                     string          `json:"schema_id"`
	IncidentID                   string          `json:"incident_id"`
	SnapshotAt                   time.Time       `json:"snapshot_at"`
	SourceChangeSetHighWatermark string          `json:"source_change_set_high_watermark"`
	DerivationVersion            string          `json:"derivation_version"`
	Fields                       []RedactedField `json:"fields"`
}

type RedactedField struct {
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

type RedactionManifest struct {
	SchemaID               string                   `json:"schema_id"`
	ProfileID              string                   `json:"profile_id"`
	ProfileVersion         string                   `json:"profile_version"`
	ProfileSHA256          string                   `json:"profile_sha256"`
	ProfileViewSHA256      string                   `json:"profile_view_sha256"`
	TokenManifestSHA256    *string                  `json:"token_manifest_sha256,omitempty"`
	RecipientPartitionRefs []string                 `json:"recipient_partition_refs,omitempty"`
	SourceExportSHA256     string                   `json:"source_export_model_sha256"`
	RedactedExportSHA256   string                   `json:"redacted_export_model_sha256"`
	Entries                []RedactionManifestEntry `json:"entries"`
}

type RedactionManifestEntry struct {
	Path                        string            `json:"path"`
	ContentClass                string            `json:"content_class"`
	SourceFamily                string            `json:"source_family,omitempty"`
	Action                      string            `json:"action"`
	RuleID                      string            `json:"rule_id"`
	ProfileID                   string            `json:"profile_id"`
	ProfileVersion              string            `json:"profile_version"`
	ProfileSHA256               string            `json:"profile_sha256"`
	SelectedRuleTrace           SelectedRuleTrace `json:"selected_rule_trace"`
	DisclosurePartitionHandling string            `json:"disclosure_partition_handling"`
	DisclosurePartitionRefs     []string          `json:"disclosure_partition_refs,omitempty"`
	SupportRefs                 []string          `json:"support_refs,omitempty"`
	Outcome                     string            `json:"outcome"`
}

type SelectedRuleTrace struct {
	SchemaID          string  `json:"schema_id"`
	SelectionKind     string  `json:"selection_kind"`
	SelectorValue     *string `json:"selector_value,omitempty"`
	RuleID            string  `json:"rule_id"`
	Action            string  `json:"action"`
	PrecedenceOrdinal int     `json:"precedence_ordinal"`
}

type RedactionProfileView struct {
	SchemaID                       string              `json:"schema_id"`
	ProfileID                      string              `json:"profile_id"`
	ProfileVersion                 string              `json:"profile_version"`
	ProfileSHA256                  string              `json:"profile_sha256"`
	AllowedDisclosurePartitionRefs []string            `json:"allowed_disclosure_partition_refs,omitempty"`
	AllowAuthoredPresentationText  bool                `json:"allow_authored_presentation_text"`
	DefaultAction                  RedactionActionView `json:"default_action"`
	Rules                          []RedactionRuleView `json:"rules"`
}

type RedactionRuleView struct {
	RuleID        string              `json:"rule_id"`
	SelectorKind  string              `json:"selector_kind"`
	SelectorValue string              `json:"selector_value"`
	Action        RedactionActionView `json:"action"`
}

type RedactionActionView struct {
	Type     string `json:"type"`
	MaxChars *int   `json:"max_chars,omitempty"`
}

type RedactionTokenManifest struct {
	SchemaID             string                        `json:"schema_id"`
	SourceExportSHA256   string                        `json:"source_export_model_sha256"`
	RedactedExportSHA256 string                        `json:"redacted_export_model_sha256"`
	ProfileSHA256        string                        `json:"profile_sha256"`
	ReleaseScope         string                        `json:"release_scope"`
	Entries              []RedactionTokenManifestEntry `json:"entries"`
}

type RedactionTokenManifestEntry struct {
	TokenID                 string   `json:"token_id"`
	DisplayToken            string   `json:"display_token"`
	StableSubjectRef        string   `json:"stable_subject_ref"`
	SubjectKind             string   `json:"subject_kind"`
	Paths                   []string `json:"paths"`
	ContentClasses          []string `json:"content_classes"`
	SourceFamilies          []string `json:"source_families"`
	Actions                 []string `json:"actions"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs,omitempty"`
}

type RedactionRevealMap struct {
	SchemaID             string                    `json:"schema_id"`
	Sensitivity          string                    `json:"sensitivity"`
	TokenManifestSHA256  string                    `json:"token_manifest_sha256"`
	SourceExportSHA256   string                    `json:"source_export_model_sha256"`
	RedactedExportSHA256 string                    `json:"redacted_export_model_sha256"`
	Entries              []RedactionRevealMapEntry `json:"entries"`
}

type RedactionRevealMapEntry struct {
	TokenID                 string   `json:"token_id"`
	DisplayToken            string   `json:"display_token"`
	StableSubjectRef        string   `json:"stable_subject_ref"`
	SubjectKind             string   `json:"subject_kind"`
	Path                    string   `json:"path"`
	Action                  string   `json:"action"`
	RuleID                  string   `json:"rule_id"`
	OriginalValue           any      `json:"original_value"`
	OriginalValueSHA256     string   `json:"original_value_sha256"`
	DisclosurePartitionRefs []string `json:"disclosure_partition_refs,omitempty"`
}

type RedactionResult struct {
	Model               RedactedExportModel
	Manifest            RedactionManifest
	ProfileView         RedactionProfileView
	TokenManifest       *RedactionTokenManifest
	RevealMap           *RedactionRevealMap
	ModelSHA256         string
	ManifestSHA256      string
	ProfileViewSHA256   string
	TokenManifestSHA256 string
	RevealMapSHA256     string
	ManifestJSON        []byte
	ProfileViewJSON     []byte
	TokenManifestJSON   []byte
	RevealMapJSON       []byte
}

type TemplateAsset struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type TemplateBinding struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

type TemplateContract struct {
	TemplateID             string            `json:"template_id"`
	TemplateVersion        string            `json:"template_version"`
	DeckTitle              string            `json:"deck_title"`
	SupportedOutputKinds   []string          `json:"supported_output_kinds"`
	SupportedReleaseScopes []string          `json:"supported_release_scopes"`
	AllowedBindings        []string          `json:"allowed_bindings"`
	RenderBindings         []TemplateBinding `json:"render_bindings"`
	RequiredFieldPaths     []string          `json:"required_field_paths"`
	LocalAssets            []TemplateAsset   `json:"local_assets"`
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

func BuildExportModel(incident IncidentMetadataSnapshot, snapshotAt time.Time, watermark string, workbookFields []ExportField) (ExportModel, string, error) {
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
	fields = append(fields, workbookFields...)
	if incident.Description != nil {
		fields = append(fields, ExportField{
			Path:                    "/incident/description",
			ContentClass:            ContentClassSourceEvidence,
			SourceFamily:            "incident_metadata",
			Value:                   *incident.Description,
			DisclosurePartitionRefs: []string{"public_summary"},
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
	return buildStructuredExportModel(incident.ID, "", snapshotAt, watermark, ReleaseScopeInternalReview, nil, fields)
}

func buildStructuredExportModel(incidentID string, snapshotID string, snapshotAt time.Time, watermark string, releaseScope string, recipientPartitionRefs []string, fields []ExportField) (ExportModel, string, error) {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	fields = cloneExportFields(fields)
	section := ReportingSection{
		SchemaID:                "cartulary.reporting_section.v1",
		SectionID:               "sec-0001",
		SectionKind:             "appendix",
		Title:                   "Snapshot export",
		OrderingKey:             "0001",
		Blocks:                  make([]ReportingBlock, 0, len(fields)),
		SourceRefs:              []string{},
		SupportRefs:             []string{},
		DisclosurePartitionRefs: []string{},
		ContentClassSummary: ReportingContentClassSummary{
			CaseFact: len(fields),
		},
		SectionValidation: ReportingSectionValidation{
			Result: "passed",
		},
	}
	recordsByKey := map[string]ReportingRecordSummary{}
	relationships := []ReportingRelationshipSummary{}
	timelineEvents := []ReportingTimelineEvent{}
	subjectsByRef := map[string]TokenizableSubject{}
	supportIndexByID := map[string]ReportingSupportRef{}

	for index, field := range fields {
		sourceRefs := []string{field.Path}
		supportRefs := cloneStrings(field.SupportRefs)
		blockID := fmt.Sprintf("sec-0001.b%04d", index+1)
		reportingField := ReportingField{
			SchemaID:                "cartulary.reporting_field.v1",
			FieldKey:                fieldKeyFromExportPath(field.Path),
			DisplayLabel:            displayLabelFromExportPath(field.Path),
			FieldOrdinal:            1,
			SourceValueState:        sourceValueState(field.Value),
			RedactedValueState:      "unchanged",
			Value:                   field.Value,
			RawValueSHA256:          nil,
			SourceRefs:              sourceRefs,
			SupportRefs:             supportRefs,
			DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
		}
		block := ReportingBlock{
			SchemaID:                "cartulary.reporting_block.v1",
			BlockID:                 blockID,
			BlockKind:               "paragraph",
			BlockOrdinal:            index + 1,
			ContentClass:            field.ContentClass,
			Fields:                  []ReportingField{reportingField},
			Children:                []ReportingBlock{},
			SourceRefs:              sourceRefs,
			SupportRefs:             supportRefs,
			DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
			ExcludedFieldKeys:       []string{},
		}
		section.Blocks = append(section.Blocks, block)
		section.SourceRefs = appendUniqueStrings(section.SourceRefs, sourceRefs...)
		section.SupportRefs = appendUniqueStrings(section.SupportRefs, supportRefs...)

		recordID, ok := exportRecordID(field.Path)
		if ok {
			refID := field.Path
			sourceRef := SourceRecordRef{
				SchemaID:         "source_record_ref.v1",
				SourceFamily:     field.SourceFamily,
				SourceRecordID:   recordID,
				SourceSnapshotID: snapshotID,
				SourceRefID:      &refID,
			}
			recordKey := field.SourceFamily + ":" + recordID
			record := recordsByKey[recordKey]
			if record.SchemaID == "" {
				displayName := displayNameFromValue(field.Value)
				record = ReportingRecordSummary{
					SchemaID:                "cartulary.reporting_record_summary.v1",
					RecordID:                recordID,
					RecordType:              field.SourceFamily,
					SourceRecordRef:         sourceRef,
					DisplayName:             displayName,
					DeletedState:            "active",
					Fields:                  []ReportingField{},
					SourceRefs:              []string{},
					SupportRefs:             []string{},
					DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
				}
			}
			record.Fields = append(record.Fields, reportingField)
			record.SourceRefs = appendUniqueStrings(record.SourceRefs, sourceRefs...)
			record.SupportRefs = appendUniqueStrings(record.SupportRefs, supportRefs...)
			recordsByKey[recordKey] = record
			if field.SourceFamily == "timeline_event" {
				timelineEvents = append(timelineEvents, ReportingTimelineEvent{
					SchemaID:                "cartulary.reporting_timeline_event.v1",
					TimelineEventID:         recordID,
					SourceRecordRef:         sourceRef,
					ActivitySortState:       "unresolved",
					DisplayTimes:            map[string]any{"primary": nil},
					Fields:                  []ReportingField{reportingField},
					SourceRefs:              sourceRefs,
					SupportRefs:             supportRefs,
					DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
				})
			}
			if subjectKind := subjectKindForSourceFamily(field.SourceFamily); subjectKind != "" {
				stableRef := subjectKind + ":" + recordID
				subjectsByRef[stableRef] = TokenizableSubject{
					SchemaID:                "cartulary.tokenizable_subject.v1",
					StableSubjectRef:        stableRef,
					SubjectKind:             subjectKind,
					SourceRecordRef:         &sourceRef,
					DisplayName:             displayNameFromValue(field.Value),
					DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
				}
			}
		}
		for _, supportRef := range supportRefs {
			supportID := supportRefID(supportRef)
			if _, exists := supportIndexByID[supportID]; !exists {
				supportIndexByID[supportID] = ReportingSupportRef{
					SchemaID:                "cartulary.reporting_support_ref.v1",
					SupportRefID:            supportID,
					SupportKind:             supportKindFromRef(supportRef),
					SupportTargetRef:        supportTargetFromRef(supportRef),
					SourceSnapshotID:        snapshotID,
					SupportRole:             "corroborating",
					CustodyState:            "not_applicable",
					DisclosurePartitionRefs: []string{},
				}
			}
		}
		if field.SourceFamily == "record_link" {
			relationships = append(relationships, relationshipSummaryFromField(field, snapshotID, reportingField))
		}
	}
	records := make([]ReportingRecordSummary, 0, len(recordsByKey))
	for _, record := range recordsByKey {
		sort.Slice(record.Fields, func(i, j int) bool {
			return record.Fields[i].FieldKey < record.Fields[j].FieldKey
		})
		for i := range record.Fields {
			record.Fields[i].FieldOrdinal = i + 1
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].RecordType == records[j].RecordType {
			return records[i].RecordID < records[j].RecordID
		}
		return records[i].RecordType < records[j].RecordType
	})
	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].RelationshipID < relationships[j].RelationshipID
	})
	sort.Slice(timelineEvents, func(i, j int) bool {
		return timelineEvents[i].TimelineEventID < timelineEvents[j].TimelineEventID
	})
	subjects := make([]TokenizableSubject, 0, len(subjectsByRef))
	for _, subject := range subjectsByRef {
		subjects = append(subjects, subject)
	}
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
	createdAt := snapshotAt.UTC()
	model := ExportModel{
		SchemaID:                     ExportModelSchemaID,
		ExportModelID:                exportModelID(snapshotID, incidentID, DerivationVersion, createdAt),
		IncidentID:                   incidentID,
		SnapshotID:                   snapshotID,
		SnapshotAt:                   createdAt,
		RenderAdmittedAt:             createdAt,
		SourceChangeSetHighWatermark: watermark,
		DerivationVersion:            DerivationVersion,
		ExportModelCreatedAt:         createdAt,
		ExportModelGeneratorID:       "cartulary.reporting.materializer",
		ExportModelGeneratorVersion:  "1",
		ReleaseScope:                 releaseScope,
		RecipientPartitionRefs:       cloneStrings(recipientPartitionRefs),
		Sections:                     []ReportingSection{section},
		Records:                      records,
		Relationships:                relationships,
		TimelineEvents:               timelineEvents,
		Subjects:                     subjects,
		Diagrams:                     []ReportingDiagram{},
		Assets:                       []ReportingAssetDeclaration{},
		SupportIndex:                 supportIndex,
		ValidationSummary: ReportingExportModelValidation{
			SchemaID: "cartulary.reporting_export_model_validation.v1",
			Result:   "passed",
			Issues:   []ReportingValidationIssue{},
		},
		Fields: fields,
	}
	encoded, err := canonicalJSON(model)
	if err != nil {
		return ExportModel{}, "", err
	}
	return model, hashHex(encoded), nil
}

func (model ExportModel) CompatibilityFields() []ExportField {
	return model.RedactionFields()
}

func (model ExportModel) RedactionFields() []ExportField {
	fields := []ExportField{}
	for _, section := range model.Sections {
		fields = append(fields, redactionFieldsFromBlocks(section.Blocks)...)
	}
	if len(fields) > 0 {
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Path < fields[j].Path
		})
		return fields
	}
	if len(model.Fields) > 0 {
		return cloneExportFields(model.Fields)
	}
	return nil
}

func redactionFieldsFromBlocks(blocks []ReportingBlock) []ExportField {
	fields := []ExportField{}
	for _, block := range blocks {
		for _, field := range block.Fields {
			path := ""
			if len(field.SourceRefs) > 0 {
				path = field.SourceRefs[0]
			}
			if path == "" {
				path = "/" + field.FieldKey
			}
			fields = append(fields, ExportField{
				Path:                    path,
				ContentClass:            contentClassForRedactionBlock(block, path),
				SourceFamily:            sourceFamilyForPath(path),
				Value:                   field.Value,
				DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
				SupportRefs:             cloneStrings(field.SupportRefs),
			})
		}
		fields = append(fields, redactionFieldsFromBlocks(block.Children)...)
	}
	return fields
}

func contentClassForRedactionBlock(block ReportingBlock, path string) string {
	if strings.TrimSpace(block.ContentClass) != "" {
		return block.ContentClass
	}
	return legacyContentClassForPath(path)
}

func cloneExportFields(fields []ExportField) []ExportField {
	if len(fields) == 0 {
		return nil
	}
	out := make([]ExportField, len(fields))
	for i, field := range fields {
		out[i] = field
		out[i].DisclosurePartitionRefs = cloneStrings(field.DisclosurePartitionRefs)
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

func exportModelID(snapshotID string, incidentID string, derivationVersion string, createdAt time.Time) string {
	input := fmt.Sprintf("cartulary.reporting_export_model_id.v1\n%s\n%s\n%s\n%s", snapshotID, incidentID, derivationVersion, createdAt.UTC().Format(time.RFC3339Nano))
	return "expm_" + hashHex([]byte(input))
}

func fieldKeyFromExportPath(path string) string {
	key := strings.Trim(path, "/")
	key = strings.ReplaceAll(key, "/", ".")
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, "-", "_")
	if key == "" {
		return "value"
	}
	return key
}

func displayLabelFromExportPath(path string) *string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[len(parts)-1] == "" {
		return nil
	}
	label := strings.ReplaceAll(parts[0], "_", " ")
	if len(parts) > 1 {
		label = strings.ReplaceAll(parts[0], "_", " ") + " " + parts[len(parts)-1]
	}
	return &label
}

func sourceValueState(value any) string {
	if value == nil {
		return "null"
	}
	return "present"
}

func exportRecordID(path string) (string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		return "", false
	}
	if parts[0] == "incident" {
		return "", false
	}
	return parts[1], true
}

func displayNameFromValue(value any) *string {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	for _, key := range []string{"display_name", "title", "summary", "activity_synopsis_text", "statement"} {
		if text, ok := object[key].(string); ok && strings.TrimSpace(text) != "" {
			return &text
		}
	}
	return nil
}

func subjectKindForSourceFamily(sourceFamily string) string {
	switch sourceFamily {
	case "host":
		return "host"
	case "identity":
		return "identity"
	case "party":
		return "party"
	case "entity_mention":
		return "unresolved_mention"
	default:
		return ""
	}
}

func supportRefID(ref string) string {
	return "sup_" + hashHex([]byte(ref))[:24]
}

func supportKindFromRef(ref string) string {
	switch {
	case strings.HasPrefix(ref, "/record_envelopes/"):
		return "source_record"
	case strings.HasPrefix(ref, "/evidence/"):
		return "evidence_item"
	case strings.HasPrefix(ref, "/timeline/"):
		return "timeline_event"
	case strings.HasPrefix(ref, "/relationships/"):
		return "relationship"
	default:
		return "source_record"
	}
}

func supportTargetFromRef(ref string) string {
	parts := strings.Split(strings.Trim(ref, "/"), "/")
	if len(parts) != 2 {
		return "record:unknown:" + hashHex([]byte(ref))[:16]
	}
	switch parts[0] {
	case "record_envelopes":
		return "record:record_envelope:" + parts[1]
	case "evidence":
		return "evidence:" + parts[1]
	case "timeline":
		return "timeline:" + parts[1]
	case "relationships":
		return "relationship:" + parts[1]
	default:
		return "record:" + sourceFamilyForPath(ref) + ":" + parts[1]
	}
}

func relationshipSummaryFromField(field ExportField, snapshotID string, reportingField ReportingField) ReportingRelationshipSummary {
	relationshipID, _ := exportRecordID(field.Path)
	value, _ := field.Value.(map[string]any)
	srcID, _ := value["src_record_id"].(string)
	dstID, _ := value["dst_record_id"].(string)
	kind, _ := value["link_type"].(string)
	if kind == "" {
		kind = "record_link"
	}
	srcRefID := "/record_envelopes/" + srcID
	dstRefID := "/record_envelopes/" + dstID
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
				SourceRefID:      &srcRefID,
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
				SourceRefID:      &dstRefID,
			},
		},
		Direction:               "directed",
		SourceRefs:              cloneStrings(reportingField.SourceRefs),
		SupportRefs:             cloneStrings(reportingField.SupportRefs),
		DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
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
		return parts[0]
	}
}

func legacyContentClassForPath(path string) string {
	switch sourceFamilyForPath(path) {
	case "incident_metadata":
		if strings.HasPrefix(path, "/incident/description") {
			return ContentClassSourceEvidence
		}
		if strings.HasPrefix(path, "/incident/title") {
			return ContentClassCuratedNarrative
		}
		return ContentClassDerivedAnalytic
	case "record_envelope", "host", "identity", "record_link", "record_tag":
		return ContentClassDerivedAnalytic
	case "timeline_event", "party", "evidence", "entity_mention":
		return ContentClassSourceEvidence
	case "finding_hypothesis":
		return ContentClassCuratedNarrative
	default:
		return ContentClassWorkingMaterial
	}
}

type RedactionProfileRegistry struct{}

func DefaultRedactionProfileRegistry() RedactionProfileRegistry {
	return RedactionProfileRegistry{}
}

func (RedactionProfileRegistry) Resolve(incidentID string, id string, version string, recipientPartitionRefs []string) (RedactionProfile, string, error) {
	_ = incidentID
	var profile RedactionProfile
	switch {
	case id == InternalRedactionProfileID && version == "1":
		profile = internalRedactionProfile()
	case id == ExternalRedactionProfileID && version == "1":
		profile = externalRedactionProfile(recipientPartitionRefs)
	case id == TokenizedRedactionProfileID && version == "1":
		profile = tokenizedReviewRedactionProfile()
	default:
		return RedactionProfile{}, "", ErrInvalidRedactionProfile
	}
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		return RedactionProfile{}, "", err
	}
	return profile, digest, nil
}

func ResolveRedactionProfile(id string, version string, recipientPartitionRefs []string) (RedactionProfile, string, error) {
	return DefaultRedactionProfileRegistry().Resolve("", id, version, recipientPartitionRefs)
}

func ValidateRedactionProfile(profile RedactionProfile) (string, error) {
	if profile.SchemaID != RedactionProfileSchemaID || strings.TrimSpace(profile.ProfileID) == "" || strings.TrimSpace(profile.Version) == "" {
		return "", fmt.Errorf("%w: missing identity", ErrInvalidRedactionProfile)
	}
	if err := validateAction(profile.DefaultAction); err != nil {
		return "", err
	}
	if err := validateDisclosurePartitionSet(profile.AllowedDisclosurePartitionRefs); err != nil {
		return "", err
	}
	paths := map[string]string{}
	classes := map[string]string{}
	for _, rule := range profile.Rules {
		if strings.TrimSpace(rule.RuleID) == "" {
			return "", fmt.Errorf("%w: missing rule id", ErrInvalidRedactionProfile)
		}
		if err := validateAction(rule.Action); err != nil {
			return "", err
		}
		hasPath := rule.Path != nil && strings.TrimSpace(*rule.Path) != ""
		hasClass := rule.ContentClass != nil && strings.TrimSpace(*rule.ContentClass) != ""
		if hasPath == hasClass {
			return "", fmt.Errorf("%w: rule %s must select exactly one target", ErrInvalidRedactionProfile, rule.RuleID)
		}
		if hasPath {
			if previous, ok := paths[*rule.Path]; ok {
				return "", fmt.Errorf("%w: duplicate path rule %s conflicts with %s", ErrInvalidRedactionProfile, rule.RuleID, previous)
			}
			paths[*rule.Path] = rule.RuleID
		}
		if hasClass {
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
	return hashHex(encoded), nil
}

func BuildRedactionProfileView(profile RedactionProfile, profileSHA256 string) (RedactionProfileView, []byte, string, error) {
	rules := make([]RedactionRuleView, 0, len(profile.Rules))
	for _, rule := range profile.Rules {
		view := RedactionRuleView{
			RuleID: rule.RuleID,
			Action: redactionActionView(rule.Action),
		}
		switch {
		case rule.Path != nil:
			view.SelectorKind = "path"
			view.SelectorValue = *rule.Path
		case rule.ContentClass != nil:
			view.SelectorKind = "content_class"
			view.SelectorValue = *rule.ContentClass
		}
		rules = append(rules, view)
	}
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].SelectorKind == rules[j].SelectorKind {
			if rules[i].SelectorValue == rules[j].SelectorValue {
				return rules[i].RuleID < rules[j].RuleID
			}
			return rules[i].SelectorValue < rules[j].SelectorValue
		}
		return rules[i].SelectorKind < rules[j].SelectorKind
	})
	view := RedactionProfileView{
		SchemaID:                       RedactionProfileViewSchemaID,
		ProfileID:                      profile.ProfileID,
		ProfileVersion:                 profile.Version,
		ProfileSHA256:                  profileSHA256,
		AllowedDisclosurePartitionRefs: cloneStrings(profile.AllowedDisclosurePartitionRefs),
		AllowAuthoredPresentationText:  profile.AllowAuthoredPresentationText,
		DefaultAction:                  redactionActionView(profile.DefaultAction),
		Rules:                          rules,
	}
	encoded, err := canonicalJSON(view)
	if err != nil {
		return RedactionProfileView{}, nil, "", err
	}
	return view, encoded, hashHex(encoded), nil
}

func redactionActionView(action RedactionActionSpec) RedactionActionView {
	return RedactionActionView{
		Type:     action.Type,
		MaxChars: action.MaxChars,
	}
}

func RedactExportModel(model ExportModel, profile RedactionProfile, profileSHA256 string, sourceExportSHA256 string, releaseScope string, recipientPartitionRefs []string) (RedactionResult, error) {
	profileView, profileViewJSON, profileViewSHA, err := BuildRedactionProfileView(profile, profileSHA256)
	if err != nil {
		return RedactionResult{}, err
	}
	pathRules := map[string]RedactionRule{}
	classRules := map[string]RedactionRule{}
	for _, rule := range profile.Rules {
		if rule.Path != nil {
			pathRules[*rule.Path] = rule
			continue
		}
		if rule.ContentClass != nil {
			classRules[*rule.ContentClass] = rule
		}
	}
	fields := model.RedactionFields()
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	redacted := RedactedExportModel{
		SchemaID:                     "cartulary.redacted_export_model.v1",
		IncidentID:                   model.IncidentID,
		SnapshotAt:                   model.SnapshotAt.UTC(),
		SourceChangeSetHighWatermark: model.SourceChangeSetHighWatermark,
		DerivationVersion:            model.DerivationVersion,
		Fields:                       []RedactedField{},
	}
	entries := make([]RedactionManifestEntry, 0, len(fields))
	tokenEntriesByID := map[string]*RedactionTokenManifestEntry{}
	revealEntries := []RedactionRevealMapEntry{}
	for _, field := range fields {
		ruleID, action, selectedTrace := selectRedactionRule(field, profile.DefaultAction, pathRules, classRules)
		value, include, outcome, err := applyRedactionAction(field.Value, action)
		if err != nil {
			return RedactionResult{}, err
		}
		tokenCandidate := false
		tokenID := ""
		displayToken := ""
		stableSubjectRef := ""
		subjectKind := ""
		if action.Type == ActionMask || action.Type == ActionStub {
			if ref, kind, ok := tokenizableSubjectRefForField(field); ok {
				tokenCandidate = true
				stableSubjectRef = ref
				subjectKind = kind
				tokenID, displayToken = deriveDisplayToken(sourceExportSHA256, profileSHA256, releaseScope, stableSubjectRef)
				value = displayToken
				include = true
				outcome = "tokenized"
			}
		}
		if releaseScope == ReleaseScopeExternal && include && !fieldDisclosureAllowed(profile, field) {
			value = nil
			include = false
			outcome = "dropped_disclosure_partition"
			action.Type = ActionDrop
			selectedTrace.Action = action.Type
			tokenCandidate = false
		}
		if include {
			redacted.Fields = append(redacted.Fields, RedactedField{
				Path:                    field.Path,
				ContentClass:            field.ContentClass,
				SourceFamily:            field.SourceFamily,
				Value:                   value,
				DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
				SupportRefs:             cloneStrings(field.SupportRefs),
				RawBlobSource:           field.RawBlobSource,
				OpaqueBinary:            field.OpaqueBinary,
				GeneratedPresentation:   field.GeneratedPresentation,
			})
			if tokenCandidate {
				recordTokenManifestEntry(tokenEntriesByID, tokenID, displayToken, stableSubjectRef, subjectKind, field, action.Type)
				revealEntry, err := redactionRevealMapEntry(tokenID, displayToken, stableSubjectRef, subjectKind, field, action.Type, ruleID)
				if err != nil {
					return RedactionResult{}, err
				}
				revealEntries = append(revealEntries, revealEntry)
			}
		}
		entries = append(entries, RedactionManifestEntry{
			Path:                        field.Path,
			ContentClass:                field.ContentClass,
			SourceFamily:                field.SourceFamily,
			Action:                      action.Type,
			RuleID:                      ruleID,
			ProfileID:                   profile.ProfileID,
			ProfileVersion:              profile.Version,
			ProfileSHA256:               profileSHA256,
			SelectedRuleTrace:           selectedTrace,
			DisclosurePartitionHandling: disclosurePartitionHandling(field, include),
			DisclosurePartitionRefs:     cloneStrings(field.DisclosurePartitionRefs),
			SupportRefs:                 cloneStrings(field.SupportRefs),
			Outcome:                     outcome,
		})
	}
	redactedBytes, err := canonicalJSON(redacted)
	if err != nil {
		return RedactionResult{}, err
	}
	modelSHA := hashHex(redactedBytes)
	tokenManifest, tokenManifestJSON, tokenManifestSHA, revealMap, revealMapJSON, revealMapSHA, err := buildTokenArtifacts(tokenEntriesByID, revealEntries, sourceExportSHA256, modelSHA, profileSHA256, releaseScope)
	if err != nil {
		return RedactionResult{}, err
	}
	var tokenManifestSHAPtr *string
	if tokenManifestSHA != "" {
		tokenManifestSHAPtr = &tokenManifestSHA
	}
	manifest := RedactionManifest{
		SchemaID:               RedactionManifestSchemaID,
		ProfileID:              profile.ProfileID,
		ProfileVersion:         profile.Version,
		ProfileSHA256:          profileSHA256,
		ProfileViewSHA256:      profileViewSHA,
		TokenManifestSHA256:    tokenManifestSHAPtr,
		RecipientPartitionRefs: cloneStrings(recipientPartitionRefs),
		SourceExportSHA256:     sourceExportSHA256,
		RedactedExportSHA256:   modelSHA,
		Entries:                entries,
	}
	if err := ValidateRedactionResult(redacted, manifest, releaseScope); err != nil {
		return RedactionResult{}, err
	}
	manifestBytes, err := canonicalJSON(manifest)
	if err != nil {
		return RedactionResult{}, err
	}
	return RedactionResult{
		Model:               redacted,
		Manifest:            manifest,
		ProfileView:         profileView,
		TokenManifest:       tokenManifest,
		RevealMap:           revealMap,
		ModelSHA256:         modelSHA,
		ManifestSHA256:      hashHex(manifestBytes),
		ProfileViewSHA256:   profileViewSHA,
		TokenManifestSHA256: tokenManifestSHA,
		RevealMapSHA256:     revealMapSHA,
		ManifestJSON:        manifestBytes,
		ProfileViewJSON:     profileViewJSON,
		TokenManifestJSON:   tokenManifestJSON,
		RevealMapJSON:       revealMapJSON,
	}, nil
}

func selectRedactionRule(field ExportField, defaultAction RedactionActionSpec, pathRules map[string]RedactionRule, classRules map[string]RedactionRule) (string, RedactionActionSpec, SelectedRuleTrace) {
	if rule, ok := pathRules[field.Path]; ok {
		selector := field.Path
		return rule.RuleID, rule.Action, SelectedRuleTrace{
			SchemaID:          "cartulary.selected_redaction_rule_trace.v1",
			SelectionKind:     "path",
			SelectorValue:     &selector,
			RuleID:            rule.RuleID,
			Action:            rule.Action.Type,
			PrecedenceOrdinal: 1,
		}
	}
	if rule, ok := classRules[field.ContentClass]; ok {
		selector := field.ContentClass
		return rule.RuleID, rule.Action, SelectedRuleTrace{
			SchemaID:          "cartulary.selected_redaction_rule_trace.v1",
			SelectionKind:     "content_class",
			SelectorValue:     &selector,
			RuleID:            rule.RuleID,
			Action:            rule.Action.Type,
			PrecedenceOrdinal: 2,
		}
	}
	return "profile_default", defaultAction, SelectedRuleTrace{
		SchemaID:          "cartulary.selected_redaction_rule_trace.v1",
		SelectionKind:     "profile_default",
		RuleID:            "profile_default",
		Action:            defaultAction.Type,
		PrecedenceOrdinal: 3,
	}
}

func tokenizableSubjectRefForField(field ExportField) (string, string, bool) {
	sourceFamily := field.SourceFamily
	if sourceFamily == "" {
		sourceFamily = sourceFamilyForPath(field.Path)
	}
	subjectKind := subjectKindForSourceFamily(sourceFamily)
	if subjectKind == "" {
		return "", "", false
	}
	recordID, ok := exportRecordID(field.Path)
	if !ok {
		return "", "", false
	}
	return subjectKind + ":" + recordID, subjectKind, true
}

func deriveDisplayToken(sourceExportSHA256 string, profileSHA256 string, releaseScope string, stableSubjectRef string) (string, string) {
	input := strings.Join([]string{
		"cartulary.reporting.derive_display_token.v1",
		sourceExportSHA256,
		profileSHA256,
		releaseScope,
		stableSubjectRef,
	}, "\n")
	digest := hashHex([]byte(input))
	return "rtok_" + digest[:24], "SUBJECT-" + strings.ToUpper(digest[:12])
}

func recordTokenManifestEntry(entries map[string]*RedactionTokenManifestEntry, tokenID string, displayToken string, stableSubjectRef string, subjectKind string, field ExportField, action string) {
	entry, ok := entries[tokenID]
	if !ok {
		entry = &RedactionTokenManifestEntry{
			TokenID:          tokenID,
			DisplayToken:     displayToken,
			StableSubjectRef: stableSubjectRef,
			SubjectKind:      subjectKind,
			Paths:            []string{},
			ContentClasses:   []string{},
			SourceFamilies:   []string{},
			Actions:          []string{},
		}
		entries[tokenID] = entry
	}
	entry.Paths = appendUniqueStrings(entry.Paths, field.Path)
	entry.ContentClasses = appendUniqueStrings(entry.ContentClasses, field.ContentClass)
	entry.SourceFamilies = appendUniqueStrings(entry.SourceFamilies, field.SourceFamily)
	entry.Actions = appendUniqueStrings(entry.Actions, action)
	entry.DisclosurePartitionRefs = appendUniqueStrings(entry.DisclosurePartitionRefs, field.DisclosurePartitionRefs...)
}

func redactionRevealMapEntry(tokenID string, displayToken string, stableSubjectRef string, subjectKind string, field ExportField, action string, ruleID string) (RedactionRevealMapEntry, error) {
	encodedOriginal, err := canonicalJSON(field.Value)
	if err != nil {
		return RedactionRevealMapEntry{}, err
	}
	return RedactionRevealMapEntry{
		TokenID:                 tokenID,
		DisplayToken:            displayToken,
		StableSubjectRef:        stableSubjectRef,
		SubjectKind:             subjectKind,
		Path:                    field.Path,
		Action:                  action,
		RuleID:                  ruleID,
		OriginalValue:           field.Value,
		OriginalValueSHA256:     hashHex(encodedOriginal),
		DisclosurePartitionRefs: cloneStrings(field.DisclosurePartitionRefs),
	}, nil
}

func buildTokenArtifacts(entriesByID map[string]*RedactionTokenManifestEntry, revealEntries []RedactionRevealMapEntry, sourceExportSHA256 string, redactedExportSHA256 string, profileSHA256 string, releaseScope string) (*RedactionTokenManifest, []byte, string, *RedactionRevealMap, []byte, string, error) {
	if len(entriesByID) == 0 {
		return nil, nil, "", nil, nil, "", nil
	}
	tokenEntries := make([]RedactionTokenManifestEntry, 0, len(entriesByID))
	for _, entry := range entriesByID {
		tokenEntries = append(tokenEntries, *entry)
	}
	sort.Slice(tokenEntries, func(i, j int) bool {
		if tokenEntries[i].StableSubjectRef == tokenEntries[j].StableSubjectRef {
			return tokenEntries[i].TokenID < tokenEntries[j].TokenID
		}
		return tokenEntries[i].StableSubjectRef < tokenEntries[j].StableSubjectRef
	})
	sort.Slice(revealEntries, func(i, j int) bool {
		if revealEntries[i].StableSubjectRef == revealEntries[j].StableSubjectRef {
			return revealEntries[i].Path < revealEntries[j].Path
		}
		return revealEntries[i].StableSubjectRef < revealEntries[j].StableSubjectRef
	})
	tokenManifest := RedactionTokenManifest{
		SchemaID:             RedactionTokenManifestSchemaID,
		SourceExportSHA256:   sourceExportSHA256,
		RedactedExportSHA256: redactedExportSHA256,
		ProfileSHA256:        profileSHA256,
		ReleaseScope:         releaseScope,
		Entries:              tokenEntries,
	}
	tokenManifestJSON, err := canonicalJSON(tokenManifest)
	if err != nil {
		return nil, nil, "", nil, nil, "", err
	}
	tokenManifestSHA := hashHex(tokenManifestJSON)
	revealMap := RedactionRevealMap{
		SchemaID:             RedactionRevealMapSchemaID,
		Sensitivity:          "internal_sensitive",
		TokenManifestSHA256:  tokenManifestSHA,
		SourceExportSHA256:   sourceExportSHA256,
		RedactedExportSHA256: redactedExportSHA256,
		Entries:              revealEntries,
	}
	revealMapJSON, err := canonicalJSON(revealMap)
	if err != nil {
		return nil, nil, "", nil, nil, "", err
	}
	return &tokenManifest, tokenManifestJSON, tokenManifestSHA, &revealMap, revealMapJSON, hashHex(revealMapJSON), nil
}

func ValidateRedactionResult(model RedactedExportModel, manifest RedactionManifest, releaseScope string) error {
	if manifest.SchemaID != RedactionManifestSchemaID ||
		strings.TrimSpace(manifest.ProfileSHA256) == "" ||
		strings.TrimSpace(manifest.ProfileViewSHA256) == "" {
		return fmt.Errorf("%w: manifest_identity", ErrRedactionValidation)
	}
	if len(manifest.Entries) == 0 {
		return fmt.Errorf("%w: manifest_empty", ErrRedactionValidation)
	}
	fieldPaths := map[string]RedactedField{}
	for _, field := range model.Fields {
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
		if entry.Path == "" || entry.ContentClass == "" || entry.Action == "" || entry.RuleID == "" {
			return fmt.Errorf("%w: manifest_entry_incomplete", ErrRedactionValidation)
		}
		if entry.SelectedRuleTrace.SchemaID != "cartulary.selected_redaction_rule_trace.v1" ||
			entry.SelectedRuleTrace.RuleID == "" ||
			entry.SelectedRuleTrace.Action == "" ||
			entry.SelectedRuleTrace.PrecedenceOrdinal < 1 {
			return fmt.Errorf("%w: manifest_entry_trace_incomplete", ErrRedactionValidation)
		}
		if _, ok := manifestPaths[entry.Path]; ok {
			return fmt.Errorf("%w: duplicate_manifest_path", ErrRedactionValidation)
		}
		manifestPaths[entry.Path] = struct{}{}
		if _, ok := fieldPaths[entry.Path]; !ok && !strings.HasPrefix(entry.Outcome, "dropped") {
			return fmt.Errorf("%w: manifest_outcome_mismatch", ErrRedactionValidation)
		}
	}
	return nil
}

func validateDisclosurePartitionSet(values []string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: empty disclosure partition", ErrInvalidRedactionProfile)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("%w: duplicate disclosure partition %q", ErrInvalidRedactionProfile, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func fieldDisclosureAllowed(profile RedactionProfile, field ExportField) bool {
	if len(field.DisclosurePartitionRefs) == 0 {
		return true
	}
	allowed := map[string]struct{}{}
	for _, partition := range profile.AllowedDisclosurePartitionRefs {
		allowed[partition] = struct{}{}
	}
	for _, partition := range field.DisclosurePartitionRefs {
		if _, ok := allowed[partition]; !ok {
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
	switch action.Type {
	case ActionAllow, ActionDrop, ActionMask, ActionStub:
		return nil
	case ActionTruncate:
		if action.MaxChars == nil || *action.MaxChars < 1 || *action.MaxChars > 4096 {
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
		return value, true, "allowed", nil
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
			return text, true, "allowed", nil
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

func disclosurePartitionHandling(field ExportField, include bool) string {
	if len(field.DisclosurePartitionRefs) == 0 {
		return "none"
	}
	if !include {
		return "dropped_with_field"
	}
	return "retained_with_field"
}

const defaultReportCSS = "body{font-family:system-ui,sans-serif;margin:2rem;}main{max-width:72rem;}dt{font-weight:600;}dd{margin:0 0 0.75rem 0;}"

func ResolveTemplateContract(id string, version string) (TemplateContract, bool) {
	if id != DefaultTemplateID || version != DefaultTemplateVersion {
		return TemplateContract{}, false
	}
	return TemplateContract{
		TemplateID:             DefaultTemplateID,
		TemplateVersion:        DefaultTemplateVersion,
		DeckTitle:              "Incident Report",
		SupportedOutputKinds:   supportedOutputKinds(),
		SupportedReleaseScopes: supportedReleaseScopes(),
		AllowedBindings:        []string{"fields", "redaction_manifest"},
		RenderBindings: []TemplateBinding{
			{Name: "fields"},
			{Name: "redaction_manifest"},
		},
		RequiredFieldPaths: []string{"/incident/status", "/incident/title"},
		LocalAssets: []TemplateAsset{
			{Path: "templates/cartulary.report.default/1/report.css", SHA256: hashHex([]byte(defaultReportCSS))},
		},
	}, true
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

func validateTemplateContract(contract TemplateContract, kind string, model RedactedExportModel, releaseScope string) error {
	if contract.TemplateID == "" || contract.TemplateVersion == "" {
		return fmt.Errorf("template contract identity is incomplete")
	}
	if strings.ContainsAny(contract.DeckTitle, "\r\n") || len([]rune(contract.DeckTitle)) > 120 {
		return fmt.Errorf("template deck title is invalid")
	}
	if !contract.SupportsOutputKind(kind) {
		return fmt.Errorf("template %s@%s does not support output kind %q", contract.TemplateID, contract.TemplateVersion, kind)
	}
	if !contract.SupportsReleaseScope(releaseScope) {
		return fmt.Errorf("template %s@%s does not support release scope %q", contract.TemplateID, contract.TemplateVersion, releaseScope)
	}
	allowed := map[string]struct{}{}
	for _, name := range contract.AllowedBindings {
		allowed[name] = struct{}{}
	}
	for _, binding := range contract.RenderBindings {
		if _, ok := allowed[binding.Name]; !ok {
			return fmt.Errorf("%w: %s", ErrUndeclaredTemplateBinding, binding.Name)
		}
	}
	paths := map[string]struct{}{}
	for _, field := range model.Fields {
		paths[field.Path] = struct{}{}
	}
	for _, path := range contract.RequiredFieldPaths {
		if _, ok := paths[path]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingRequiredField, path)
		}
	}
	for _, asset := range contract.LocalAssets {
		if strings.HasPrefix(asset.Path, "http://") || strings.HasPrefix(asset.Path, "https://") || strings.HasPrefix(asset.Path, "//") || asset.Path == "" || asset.SHA256 == "" {
			return fmt.Errorf("template asset is not a local integrity-checked asset")
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

var encodeCanonicalJSON = json.Marshal

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
		SchemaID:  RedactionProfileSchemaID,
		ProfileID: InternalRedactionProfileID,
		Version:   "1",
		AllowedDisclosurePartitionRefs: []string{
			"public_summary",
			"working_material",
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
		SchemaID:  RedactionProfileSchemaID,
		ProfileID: TokenizedRedactionProfileID,
		Version:   "1",
		AllowedDisclosurePartitionRefs: []string{
			"public_summary",
			"working_material",
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
		SchemaID:                       RedactionProfileSchemaID,
		ProfileID:                      ExternalRedactionProfileID,
		Version:                        "1",
		AllowedDisclosurePartitionRefs: canonicalStringSet(append([]string{"public_summary"}, recipientPartitionRefs...)),
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

func canonicalStringSet(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		seen[trimmed] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

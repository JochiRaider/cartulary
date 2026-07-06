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
	InternalRedactionProfileID = "cartulary.redaction.internal"
	ExternalRedactionProfileID = "cartulary.redaction.external"

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
	SchemaID                     string        `json:"schema_id"`
	IncidentID                   string        `json:"incident_id"`
	SnapshotAt                   time.Time     `json:"snapshot_at"`
	SourceChangeSetHighWatermark string        `json:"source_change_set_high_watermark"`
	DerivationVersion            string        `json:"derivation_version"`
	Fields                       []ExportField `json:"fields"`
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
	RecipientPartitionRefs []string                 `json:"recipient_partition_refs,omitempty"`
	SourceExportSHA256     string                   `json:"source_export_model_sha256"`
	RedactedExportSHA256   string                   `json:"redacted_export_model_sha256"`
	Entries                []RedactionManifestEntry `json:"entries"`
}

type RedactionManifestEntry struct {
	Path                        string   `json:"path"`
	ContentClass                string   `json:"content_class"`
	SourceFamily                string   `json:"source_family,omitempty"`
	Action                      string   `json:"action"`
	RuleID                      string   `json:"rule_id"`
	ProfileID                   string   `json:"profile_id"`
	ProfileVersion              string   `json:"profile_version"`
	ProfileSHA256               string   `json:"profile_sha256"`
	DisclosurePartitionHandling string   `json:"disclosure_partition_handling"`
	DisclosurePartitionRefs     []string `json:"disclosure_partition_refs,omitempty"`
	SupportRefs                 []string `json:"support_refs,omitempty"`
	Outcome                     string   `json:"outcome"`
}

type RedactionResult struct {
	Model          RedactedExportModel
	Manifest       RedactionManifest
	ModelSHA256    string
	ManifestSHA256 string
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
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	model := ExportModel{
		SchemaID:                     ExportModelSchemaID,
		IncidentID:                   incident.ID,
		SnapshotAt:                   snapshotAt.UTC(),
		SourceChangeSetHighWatermark: watermark,
		DerivationVersion:            DerivationVersion,
		Fields:                       fields,
	}
	encoded, err := canonicalJSON(model)
	if err != nil {
		return ExportModel{}, "", err
	}
	return model, hashHex(encoded), nil
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
	if profile.SchemaID != "cartulary.redaction_profile.v1" || strings.TrimSpace(profile.ProfileID) == "" || strings.TrimSpace(profile.Version) == "" {
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

func RedactExportModel(model ExportModel, profile RedactionProfile, profileSHA256 string, sourceExportSHA256 string, releaseScope string, recipientPartitionRefs []string) (RedactionResult, error) {
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
	fields := make([]ExportField, len(model.Fields))
	copy(fields, model.Fields)
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
	for _, field := range fields {
		ruleID := "profile_default"
		action := profile.DefaultAction
		if rule, ok := pathRules[field.Path]; ok {
			ruleID = rule.RuleID
			action = rule.Action
		} else if rule, ok := classRules[field.ContentClass]; ok {
			ruleID = rule.RuleID
			action = rule.Action
		}
		value, include, outcome, err := applyRedactionAction(field.Value, action)
		if err != nil {
			return RedactionResult{}, err
		}
		if releaseScope == ReleaseScopeExternal && include && !fieldDisclosureAllowed(profile, field) {
			value = nil
			include = false
			outcome = "dropped_disclosure_partition"
			action.Type = ActionDrop
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
	manifest := RedactionManifest{
		SchemaID:               "cartulary.redaction_manifest.v1",
		ProfileID:              profile.ProfileID,
		ProfileVersion:         profile.Version,
		ProfileSHA256:          profileSHA256,
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
		Model:          redacted,
		Manifest:       manifest,
		ModelSHA256:    modelSHA,
		ManifestSHA256: hashHex(manifestBytes),
	}, nil
}

func ValidateRedactionResult(model RedactedExportModel, manifest RedactionManifest, releaseScope string) error {
	if manifest.SchemaID != "cartulary.redaction_manifest.v1" || strings.TrimSpace(manifest.ProfileSHA256) == "" {
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

func RenderOutput(contract TemplateContract, kind string, model RedactedExportModel, manifest RedactionManifest, releaseScope string) ([]byte, string, error) {
	if err := validateTemplateContract(contract, kind, model, releaseScope); err != nil {
		return nil, "", err
	}
	var output []byte
	var mediaType string
	switch kind {
	case OutputKindHTML:
		output, mediaType = renderHTML(model, manifest), "text/html; charset=utf-8"
	case OutputKindMarkdown:
		output, mediaType = renderMarkdown(model), "text/markdown; charset=utf-8"
	case OutputKindSlidev:
		output, mediaType = renderMarkdown(model), "text/markdown; charset=utf-8"
	case OutputKindMermaid:
		output, mediaType = []byte("flowchart TD\n  snapshot[Snapshot] --> report[Report]\n"), "text/vnd.mermaid; charset=utf-8"
	case OutputKindReenactment:
		if releaseScope == ReleaseScopeExternal {
			return nil, "", fmt.Errorf("reenactment output is not eligible for external release")
		}
		output, mediaType = []byte("{\"schema_id\":\"cartulary.reenactment_stub.v1\",\"events\":[]}\n"), "application/json; charset=utf-8"
	default:
		return nil, "", fmt.Errorf("unsupported output kind %q", kind)
	}
	if err := ValidateSelfContainedOutput(kind, output); err != nil {
		return nil, "", err
	}
	return output, mediaType, nil
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
	case OutputKindMarkdown, OutputKindSlidev:
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

func renderHTML(model RedactedExportModel, manifest RedactionManifest) []byte {
	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>Incident Report</title></head><body>")
	b.WriteString("<main data-schema=\"cartulary.rendered_report.v1\">")
	b.WriteString("<h1>Incident Report</h1>")
	b.WriteString("<dl>")
	for _, field := range model.Fields {
		b.WriteString("<dt>")
		b.WriteString(escapeHTML(field.Path))
		b.WriteString("</dt><dd>")
		b.WriteString(escapeHTML(formatFieldValue(field.Value)))
		b.WriteString("</dd>")
	}
	b.WriteString("</dl><section data-redaction-manifest-sha256=\"")
	b.WriteString(escapeHTML(hashManifestIdentity(manifest)))
	b.WriteString("\"></section></main></body></html>")
	return []byte(b.String())
}

func renderMarkdown(model RedactedExportModel) []byte {
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

func hashManifestIdentity(manifest RedactionManifest) string {
	encoded, err := canonicalJSON(manifest)
	if err != nil {
		return ""
	}
	return hashHex(encoded)
}

func escapeHTML(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&#39;")
	return replacer.Replace(value)
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
		SchemaID:  "cartulary.redaction_profile.v1",
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

func externalRedactionProfile(recipientPartitionRefs []string) RedactionProfile {
	sourceClass := ContentClassSourceEvidence
	workingClass := ContentClassWorkingMaterial
	maxChars := 120
	return RedactionProfile{
		SchemaID:                       "cartulary.redaction_profile.v1",
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

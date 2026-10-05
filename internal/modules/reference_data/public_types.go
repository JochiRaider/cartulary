package reference_data

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"slices"
)

// These DTOs are the Reference Data owner surface. Format representations stay
// private and are explicitly translated without changing canonical fields.
type PackSetMember struct {
	Key            string `json:"pack_key"`
	Version        string `json:"pack_version"`
	ManifestSHA256 string `json:"manifest_sha256"`
	PayloadSHA256  string `json:"payload_sha256"`
	Contract       string `json:"pack_contract_version"`
	ProfileID      string `json:"content_profile_id"`
	ProfileVersion string `json:"content_profile_version"`
}
type PackSet struct {
	SchemaID string          `json:"schema_id"`
	ID       string          `json:"pack_set_id"`
	SHA256   string          `json:"pack_set_sha256"`
	Members  []PackSetMember `json:"members"`
}
type PackSourceArtifact struct {
	Ref     string `json:"source_ref"`
	Version string `json:"source_version"`
	SHA256  string `json:"sha256"`
}
type PackLicense struct {
	Expression     string               `json:"expression"`
	ListVersion    string               `json:"license_list_version"`
	Redistribution string               `json:"redistribution"`
	Notices        []string             `json:"notice_paths"`
	Bindings       []PackLicenseBinding `json:"license_ref_bindings"`
}
type PackLicenseBinding struct {
	Ref  string `json:"license_ref"`
	Path string `json:"notice_path"`
}
type Evaluation struct {
	Type                   string  `json:"indicator_type_id"`
	Kind                   string  `json:"value_kind"`
	Raw                    string  `json:"raw_value"`
	Valid                  bool    `json:"valid"`
	Code                   *string `json:"validation_code"`
	Display                *string `json:"display_value"`
	Normalized             *string `json:"normalized_value"`
	Defanged               *string `json:"defanged_value"`
	DedupeKey              *string `json:"dedupe_key"`
	NormalizationAlgorithm string  `json:"normalization_algorithm_id"`
	ValidationAlgorithm    string  `json:"validation_algorithm_id"`
	DefangAlgorithm        string  `json:"defang_algorithm_id"`
	DedupeAlgorithm        string  `json:"dedupe_algorithm_id"`
}
type IncidentBundleVersionReference struct {
	PackSetMember
	DistributionKind    string `json:"distribution_kind"`
	VerificationMethod  string `json:"verification_method"`
	SourceProfileID     string `json:"source_profile_id"`
	SourceProfileSHA256 string `json:"source_profile_sha256"`
}
type IncidentBundleReferences struct {
	SchemaID string                           `json:"schema_id"`
	Sets     []PackSet                        `json:"sets"`
	Versions []IncidentBundleVersionReference `json:"versions"`
}
type ValidationDetails struct {
	LimitID            *string `json:"limit_id"`
	ExpectedToken      *string `json:"expected_token"`
	ActualToken        *string `json:"actual_token"`
	RelatedPackKey     *string `json:"related_pack_key"`
	RelatedPackVersion *string `json:"related_pack_version"`
}
type ValidationIssue struct {
	ID         string            `json:"issue_id"`
	CheckID    string            `json:"check_id"`
	Severity   string            `json:"severity"`
	Phase      string            `json:"phase"`
	Code       string            `json:"code"`
	ReasonCode *string           `json:"reason_code"`
	Path       string            `json:"path"`
	EntryID    *string           `json:"entry_id"`
	Details    ValidationDetails `json:"safe_details"`
}
type ValidationSummary struct {
	SchemaID       string            `json:"schema_id"`
	Result         string            `json:"result"`
	PrimaryIssueID *string           `json:"primary_issue_id"`
	Truncated      bool              `json:"issues_truncated"`
	Total          int               `json:"total_issue_count"`
	Retained       int               `json:"retained_issue_count"`
	Issues         []ValidationIssue `json:"issues"`
}

func formatMembers(members []PackSetMember) []packformat.SetMember {
	if members == nil {
		return nil
	}
	result := make([]packformat.SetMember, len(members))
	for i, m := range members {
		result[i] = packformat.SetMember(m)
	}
	return result
}
func packSetFromFormat(value packformat.Set) PackSet {
	var members []PackSetMember
	if value.Members != nil {
		members = make([]PackSetMember, len(value.Members))
		for i, m := range value.Members {
			members[i] = PackSetMember(m)
		}
	}
	return PackSet{SchemaID: value.SchemaID, ID: value.ID, SHA256: value.SHA256, Members: members}
}
func (value PackSet) format() packformat.Set {
	return packformat.Set{SchemaID: value.SchemaID, ID: value.ID, SHA256: value.SHA256, Members: formatMembers(value.Members)}
}
func buildPackSet(members []PackSetMember) (PackSet, error) {
	value, err := packformat.BuildSet(formatMembers(members))
	return packSetFromFormat(value), err
}
func packLicenseFromFormat(value packformat.License) PackLicense {
	var bindings []PackLicenseBinding
	if value.Bindings != nil {
		bindings = make([]PackLicenseBinding, len(value.Bindings))
		for i, binding := range value.Bindings {
			bindings[i] = PackLicenseBinding(binding)
		}
	}
	return PackLicense{Expression: value.Expression, ListVersion: value.ListVersion, Redistribution: value.Redistribution, Notices: slices.Clone(value.Notices), Bindings: bindings}
}
func sourceArtifactsFromFormat(values []packformat.SourceArtifact) []PackSourceArtifact {
	if values == nil {
		return nil
	}
	result := make([]PackSourceArtifact, len(values))
	for i, value := range values {
		result[i] = PackSourceArtifact(value)
	}
	return result
}
func portableVersionFromFormat(value packformat.PortableVersion) IncidentBundleVersionReference {
	return IncidentBundleVersionReference{PackSetMember: PackSetMember(value.SetMember), DistributionKind: value.DistributionKind, VerificationMethod: value.VerificationMethod, SourceProfileID: value.SourceProfileID, SourceProfileSHA256: value.SourceProfileSHA256}
}
func (value IncidentBundleVersionReference) format() packformat.PortableVersion {
	return packformat.PortableVersion{SetMember: packformat.SetMember(value.PackSetMember), DistributionKind: value.DistributionKind, VerificationMethod: value.VerificationMethod, SourceProfileID: value.SourceProfileID, SourceProfileSHA256: value.SourceProfileSHA256}
}
func portableReferencesFromFormat(value packformat.PortableReferences) IncidentBundleReferences {
	result := IncidentBundleReferences{SchemaID: value.SchemaID}
	if value.Sets != nil {
		result.Sets = make([]PackSet, len(value.Sets))
		for i, set := range value.Sets {
			result.Sets[i] = packSetFromFormat(set)
		}
	}
	if value.Versions != nil {
		result.Versions = make([]IncidentBundleVersionReference, len(value.Versions))
		for i, version := range value.Versions {
			result.Versions[i] = portableVersionFromFormat(version)
		}
	}
	return result
}
func (value IncidentBundleReferences) format() packformat.PortableReferences {
	result := packformat.PortableReferences{SchemaID: value.SchemaID}
	if value.Sets != nil {
		result.Sets = make([]packformat.Set, len(value.Sets))
		for i, set := range value.Sets {
			result.Sets[i] = set.format()
		}
	}
	if value.Versions != nil {
		result.Versions = make([]packformat.PortableVersion, len(value.Versions))
		for i, version := range value.Versions {
			result.Versions[i] = version.format()
		}
	}
	return result
}
func summaryFromFormat(value *packformat.ValidationSummary) *ValidationSummary {
	if value == nil {
		return nil
	}
	result := &ValidationSummary{SchemaID: value.SchemaID, Result: value.Result, PrimaryIssueID: value.PrimaryIssueID, Truncated: value.Truncated, Total: value.Total, Retained: value.Retained}
	if value.Issues != nil {
		result.Issues = make([]ValidationIssue, len(value.Issues))
		for i, issue := range value.Issues {
			result.Issues[i] = ValidationIssue{ID: issue.ID, CheckID: issue.CheckID, Severity: issue.Severity, Phase: issue.Phase, Code: issue.Code, ReasonCode: issue.ReasonCode, Path: issue.Path, EntryID: issue.EntryID, Details: ValidationDetails(issue.Details)}
		}
	}
	return result
}

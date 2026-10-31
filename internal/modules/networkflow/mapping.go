package networkflow

import (
	"encoding/json"
	"fmt"
	"strings"

	generatedmapping "github.com/JochiRaider/cartulary/internal/gen/networkflowmapping"
	norm "github.com/JochiRaider/cartulary/internal/gen/networkflowunicode"
)

const (
	targetKindNetworkFlowTable = "network_flow_table"
	targetTableSchemaID        = "cartulary.network_flow_table.v1"

	mappingCandidateSchemaID = "cartulary.network_flow.mapping_candidate.v1"

	unknownColumnPolicyPreserve = "preserve_unmapped_raw"
	unknownColumnPolicyReject   = "reject_unmapped_columns"
	unknownColumnPolicyIgnore   = "ignore_unmapped_columns"

	mappingKindSourceColumn        = "source_column"
	mappingKindIgnoredSourceColumn = "ignored_source_column"
	mappingKindSystemDerivation    = "system_derivation"

	transformTimestampProfile = "timestamp_profile_v1"
	transformIPLiteral        = "ip_literal_v1"
	transformPortNumber       = "port_number_v1"
	transformProtocol         = "protocol_number_or_token_v1"
	transformUint64Decimal    = "uint64_decimal_string_v1"
	transformTrimASCIISpace   = "trim_ascii_space_v1"

	emptyPolicyInvalid = "empty_string_is_invalid"
	emptyPolicyNull    = "empty_string_is_null"

	fieldFlowStartUTC         = "network_flow.flow_start_utc"
	fieldFlowEndUTC           = "network_flow.flow_end_utc"
	fieldSrcIP                = "network_flow.src_ip"
	fieldDstIP                = "network_flow.dst_ip"
	fieldSrcPort              = "network_flow.src_port"
	fieldDstPort              = "network_flow.dst_port"
	fieldIPProtocol           = "network_flow.ip_protocol"
	fieldBytesCount           = "network_flow.bytes_count"
	fieldPacketsCount         = "network_flow.packets_count"
	fieldExporterID           = "network_flow.exporter_id"
	fieldInputInterface       = "network_flow.input_interface"
	fieldOutputInterface      = "network_flow.output_interface"
	fieldTCPFlags             = "network_flow.tcp_flags"
	fieldApplicationLabel     = "network_flow.application_label"
	fieldObservationSourceRef = "network_flow.observation_source_ref"
)

type timestampProfile struct {
	SchemaID                                   string  `json:"schema_id"`
	Mode                                       string  `json:"mode"`
	Precision                                  string  `json:"precision"`
	Timezone                                   *string `json:"timezone,omitempty"`
	TimezoneRulesetID                          *string `json:"timezone_ruleset_id,omitempty"`
	AmbiguousLocalTimePolicy                   *string `json:"ambiguous_local_time_policy,omitempty"`
	LocalTimeGapPolicy                         *string `json:"local_time_gap_policy,omitempty"`
	NetFlowExportTimeColumnOrdinal             *int    `json:"netflow_export_time_column_ordinal,omitempty"`
	NetFlowExportTimeMode                      *string `json:"netflow_export_time_mode,omitempty"`
	NetFlowExporterUptimeAtExportColumnOrdinal *int    `json:"netflow_exporter_uptime_at_export_column_ordinal,omitempty"`
}

type sourceColumnDescriptor struct {
	SourceColumnOrdinal           int          `json:"source_column_ordinal"`
	RawHeaderText                 string       `json:"raw_header_text"`
	NormalizedHeaderForSuggestion string       `json:"normalized_header_for_suggestion"`
	RawHeaderSHA256               string       `json:"raw_header_sha256"`
	SampleValues                  []safeSample `json:"sample_values"`
	DetectedEmptyCount            int          `json:"detected_empty_count"`
}

type safeSample struct {
	SafeSample     *string `json:"safe_sample"`
	RawValueSHA256 *string `json:"raw_value_sha256"`
}

type fieldMapping struct {
	MappingKind         string `json:"mapping_kind"`
	FieldKey            string `json:"field_key,omitempty"`
	SourceColumnOrdinal int    `json:"source_column_ordinal,omitempty"`
	TransformID         string `json:"transform_id,omitempty"`
	EmptyValuePolicy    string `json:"empty_value_policy,omitempty"`
	Combinability       string `json:"combinability,omitempty"`
	IgnoreReason        string `json:"ignore_reason,omitempty"`
	DerivationID        string `json:"derivation_id,omitempty"`
}

// Candidate policies retain their wire presence until candidate-to-approved
// conversion. An absent value requests a default; explicit null or "" does not.
type candidateFieldMapping struct {
	MappingKind         string          `json:"mapping_kind"`
	FieldKey            string          `json:"field_key,omitempty"`
	SourceColumnOrdinal int             `json:"source_column_ordinal,omitempty"`
	TransformID         string          `json:"transform_id,omitempty"`
	EmptyValuePolicy    json.RawMessage `json:"empty_value_policy,omitempty"`
	Combinability       string          `json:"combinability,omitempty"`
	IgnoreReason        string          `json:"ignore_reason,omitempty"`
	DerivationID        string          `json:"derivation_id,omitempty"`
}

func (candidate candidateFieldMapping) withPolicy(policy string) fieldMapping {
	return fieldMapping{
		MappingKind: candidate.MappingKind, FieldKey: candidate.FieldKey,
		SourceColumnOrdinal: candidate.SourceColumnOrdinal, TransformID: candidate.TransformID,
		EmptyValuePolicy: policy, Combinability: candidate.Combinability,
		IgnoreReason: candidate.IgnoreReason, DerivationID: candidate.DerivationID,
	}
}

type mappingCandidate struct {
	TargetKind          string                  `json:"target_kind"`
	TargetTableSchemaID string                  `json:"target_table_schema_id"`
	SourceProfileID     string                  `json:"source_profile_id"`
	ParserProfileID     string                  `json:"parser_profile_id,omitempty"`
	UnknownColumnPolicy string                  `json:"unknown_column_policy,omitempty"`
	DisplayNameOverride *string                 `json:"display_name_override,omitempty"`
	TimestampProfile    timestampProfile        `json:"timestamp_profile"`
	FieldMappings       []candidateFieldMapping `json:"field_mappings"`
}

type approvedMapping struct {
	TargetKind          string                   `json:"target_kind"`
	TargetTableSchemaID string                   `json:"target_table_schema_id"`
	SourceProfileID     string                   `json:"source_profile_id"`
	ParserProfileID     string                   `json:"parser_profile_id"`
	UnknownColumnPolicy string                   `json:"unknown_column_policy"`
	DisplayNameOverride *string                  `json:"display_name_override,omitempty"`
	TimestampProfile    timestampProfile         `json:"timestamp_profile"`
	SourceColumns       []sourceColumnDescriptor `json:"source_columns"`
	FieldMappings       []fieldMapping           `json:"field_mappings"`
}

func materializeApprovedMapping(raw json.RawMessage, sourceColumns []sourceColumnDescriptor) (approvedMapping, error) {
	candidate, err := decodeMappingCandidate(raw)
	if err != nil {
		return approvedMapping{}, err
	}
	if candidate.DisplayNameOverride != nil {
		normalized, err := normalizeExplicitDisplayName(*candidate.DisplayNameOverride)
		if err != nil {
			return approvedMapping{}, err
		}
		candidate.DisplayNameOverride = &normalized
	}
	fields, err := materializeCandidateFieldMappings(candidate.SourceProfileID, candidate.FieldMappings)
	if err != nil {
		return approvedMapping{}, err
	}

	approved := approvedMapping{
		TargetKind:          candidate.TargetKind,
		TargetTableSchemaID: candidate.TargetTableSchemaID,
		SourceProfileID:     candidate.SourceProfileID,
		ParserProfileID:     candidate.ParserProfileID,
		UnknownColumnPolicy: candidate.UnknownColumnPolicy,
		DisplayNameOverride: candidate.DisplayNameOverride,
		TimestampProfile:    candidate.TimestampProfile,
		SourceColumns:       append([]sourceColumnDescriptor(nil), sourceColumns...),
		FieldMappings:       fields,
	}
	if err := validateApprovedMapping(approved); err != nil {
		return approvedMapping{}, err
	}
	return approved, nil
}

func marshalApprovedMapping(mapping approvedMapping) json.RawMessage {
	data, err := json.Marshal(mapping)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

func materializeTimestampProfile(profile timestampProfile) timestampProfile {
	defaults := defaultMappingSourceProfile().DefaultTimestampProfile
	if profile.SchemaID == "" {
		profile.SchemaID = defaults.SchemaID
	}
	if profile.Mode == "" {
		profile.Mode = defaults.Mode
	}
	if profile.Precision == "" {
		profile.Precision = defaultTimestampPrecision(profile.Mode)
	}
	if profile.Mode == "rfc3339" {
		if profile.AmbiguousLocalTimePolicy == nil {
			profile.AmbiguousLocalTimePolicy = stringPtr(defaults.AmbiguousLocalTimePolicy)
		}
		if profile.LocalTimeGapPolicy == nil {
			profile.LocalTimeGapPolicy = stringPtr(defaults.LocalTimeGapPolicy)
		}
		if profile.Timezone != nil && *profile.Timezone != "" && *profile.Timezone != "UTC" && profile.TimezoneRulesetID == nil {
			profile.TimezoneRulesetID = stringPtr("tzdb-2026c")
		}
	}
	return profile
}

func defaultTimestampPrecision(mode string) string {
	switch mode {
	case "epoch_seconds":
		return "seconds"
	case "epoch_milliseconds", "netflow_sys_uptime_milliseconds", "cisco_sna_datetime_v1":
		return "milliseconds"
	default:
		return "microseconds"
	}
}

func materializeCandidateFieldMappings(profileID string, input []candidateFieldMapping) ([]fieldMapping, error) {
	mappings := make([]fieldMapping, 0, len(input)+1)
	hasDerivation := false
	for _, candidate := range input {
		mapping := candidate.withPolicy("")
		if candidate.MappingKind == mappingKindSourceColumn {
			if len(candidate.EmptyValuePolicy) != 0 {
				if err := json.Unmarshal(candidate.EmptyValuePolicy, &mapping.EmptyValuePolicy); err != nil || mapping.EmptyValuePolicy == "" {
					return nil, &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "invalid_empty_value_policy", FieldKey: candidate.FieldKey}
				}
			}
			if len(candidate.EmptyValuePolicy) == 0 || mapping.EmptyValuePolicy == "profile_default" {
				mapping.EmptyValuePolicy = defaultEmptyPolicyForField(profileID, candidate.FieldKey)
			}
		}
		if mapping.MappingKind == mappingKindSystemDerivation && mapping.FieldKey == fieldObservationSourceRef {
			hasDerivation = true
		}
		mappings = append(mappings, mapping)
	}
	if !hasDerivation {
		derivation := observationSourceDerivation()
		mappings = append(mappings, fieldMapping{
			MappingKind:   mappingKindSystemDerivation,
			FieldKey:      derivation.FieldKey,
			DerivationID:  derivation.DerivationID,
			Combinability: derivation.Combinability,
		})
	}
	return mappings, nil
}

func validateApprovedMapping(mapping approvedMapping) error {
	if mapping.TargetKind != generatedmapping.Registry.TargetKind || mapping.TargetTableSchemaID != generatedmapping.Registry.TargetTableSchemaID {
		return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "variant_member_conflict"}
	}
	profile, ok := mappingSourceProfile(mapping.SourceProfileID)
	if !ok || profile.ConformanceStatus != "required_v1" {
		return &mappingValidationError{Code: "network_flow_unsupported_source_profile", ReasonCode: "unsupported_source_profile"}
	}
	if mapping.ParserProfileID != profile.ParserProfileID {
		return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "mapping_kind_unavailable"}
	}
	if !slicesContains(profile.SupportedUnknownColumnPolicies, mapping.UnknownColumnPolicy) {
		return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
	}
	if len(mapping.SourceColumns) == 0 {
		return &mappingValidationError{Code: "network_flow_invalid_header", ReasonCode: "empty_header"}
	}
	sourceOrdinals := map[int]struct{}{}
	for index, column := range mapping.SourceColumns {
		if column.SourceColumnOrdinal != index+1 {
			return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
		}
		sourceOrdinals[column.SourceColumnOrdinal] = struct{}{}
	}
	if !slicesContains(profile.SupportedTimestampModes, mapping.TimestampProfile.Mode) {
		return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "variant_member_conflict"}
	}
	if err := validateTimestampProfile(mapping.TimestampProfile, len(mapping.SourceColumns)); err != nil {
		return err
	}
	byField := map[string]int{}
	byOrdinal := map[int]int{}
	hasSystemDerivation := false
	for _, fieldMapping := range mapping.FieldMappings {
		switch fieldMapping.MappingKind {
		case mappingKindSourceColumn:
			if _, ok := sourceOrdinals[fieldMapping.SourceColumnOrdinal]; !ok {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column", FieldKey: fieldMapping.FieldKey}
			}
			if !sourceMappableField(mapping.SourceProfileID, fieldMapping.FieldKey) {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "field_not_supported_by_profile", FieldKey: fieldMapping.FieldKey}
			}
			if fieldMapping.TransformID != defaultTransformForField(mapping.SourceProfileID, fieldMapping.FieldKey) {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "transform_target_mismatch", FieldKey: fieldMapping.FieldKey}
			}
			profileField, _ := mappingProfileField(mapping.SourceProfileID, fieldMapping.FieldKey)
			if !slicesContains(profileField.AllowedEmptyValuePolicies, fieldMapping.EmptyValuePolicy) {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "invalid_empty_value_policy", FieldKey: fieldMapping.FieldKey}
			}
			byField[fieldMapping.FieldKey]++
			byOrdinal[fieldMapping.SourceColumnOrdinal]++
		case mappingKindIgnoredSourceColumn:
			if _, ok := sourceOrdinals[fieldMapping.SourceColumnOrdinal]; !ok {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
			}
			byOrdinal[fieldMapping.SourceColumnOrdinal]++
		case mappingKindSystemDerivation:
			derivation := observationSourceDerivation()
			if fieldMapping.FieldKey != derivation.FieldKey || fieldMapping.DerivationID != derivation.DerivationID || fieldMapping.Combinability != derivation.Combinability {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "system_derivation_missing"}
			}
			hasSystemDerivation = true
			byField[fieldMapping.FieldKey]++
		default:
			return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "mapping_kind_unavailable"}
		}
	}
	for _, fieldKey := range requiredFields(mapping.SourceProfileID) {
		if byField[fieldKey] == 0 {
			return &mappingValidationError{Code: "network_flow_mapping_required", ReasonCode: "required_field_unmapped", FieldKey: fieldKey}
		}
	}
	if !hasSystemDerivation {
		return &mappingValidationError{Code: "network_flow_mapping_required", ReasonCode: "system_derivation_missing", FieldKey: fieldObservationSourceRef}
	}
	for fieldKey, count := range byField {
		if count > 1 {
			return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "target_field_duplicated", FieldKey: fieldKey}
		}
	}
	for ordinal, count := range byOrdinal {
		if count > 1 {
			return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "source_column_reused", FieldKey: fmt.Sprint(ordinal)}
		}
	}
	if mapping.TimestampProfile.Mode == "netflow_sys_uptime_milliseconds" {
		exportOrdinal := *mapping.TimestampProfile.NetFlowExportTimeColumnOrdinal
		uptimeOrdinal := *mapping.TimestampProfile.NetFlowExporterUptimeAtExportColumnOrdinal
		if exportOrdinal == uptimeOrdinal {
			return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "timestamp_column_reused"}
		}
		for _, fieldKey := range []string{fieldFlowStartUTC, fieldFlowEndUTC} {
			ordinal := sourceFieldOrdinal(mapping.FieldMappings, fieldKey)
			if ordinal == exportOrdinal || ordinal == uptimeOrdinal {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "timestamp_column_reused", FieldKey: fieldKey}
			}
		}
	}
	if mapping.UnknownColumnPolicy == unknownColumnPolicyIgnore || mapping.UnknownColumnPolicy == unknownColumnPolicyReject {
		for _, column := range mapping.SourceColumns {
			if byOrdinal[column.SourceColumnOrdinal] == 0 {
				return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
			}
		}
	}
	return nil
}

func defaultTransformForField(profileID, fieldKey string) string {
	if field, ok := mappingProfileField(profileID, fieldKey); ok {
		return field.TransformID
	}
	return ""
}

func defaultEmptyPolicyForField(profileID, fieldKey string) string {
	if field, ok := mappingProfileField(profileID, fieldKey); ok {
		return field.EmptyValuePolicy
	}
	return ""
}

func sourceMappableField(profileID, fieldKey string) bool {
	field, ok := mappingProfileField(profileID, fieldKey)
	return ok && (field.Requirement == "required" || field.Requirement == "optional_map_when_present")
}

func requiredFields(profileID string) []string {
	fields := make([]string, 0)
	profile, _ := mappingSourceProfile(profileID)
	for _, field := range profile.Fields {
		if field.Requirement == "required" {
			fields = append(fields, field.FieldKey)
		}
	}
	return fields
}

func mappingSourceProfile(sourceProfileID string) (generatedmapping.SourceProfile, bool) {
	for _, profile := range generatedmapping.Registry.SourceProfiles {
		if profile.SourceProfileID == sourceProfileID {
			return profile, true
		}
	}
	return generatedmapping.SourceProfile{}, false
}

func defaultMappingSourceProfile() generatedmapping.SourceProfile {
	if len(generatedmapping.Registry.SourceProfiles) == 0 {
		return generatedmapping.SourceProfile{}
	}
	return generatedmapping.Registry.SourceProfiles[0]
}

func mappingProfileField(profileID, fieldKey string) (generatedmapping.Field, bool) {
	profile, _ := mappingSourceProfile(profileID)
	for _, field := range profile.Fields {
		if field.FieldKey == fieldKey {
			return field, true
		}
	}
	return generatedmapping.Field{}, false
}

func observationSourceDerivation() generatedmapping.SystemDerivation {
	if len(generatedmapping.Registry.SystemDerivations) == 0 {
		return generatedmapping.SystemDerivation{}
	}
	return generatedmapping.Registry.SystemDerivations[0]
}

func slicesContains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func sourceAliasMatchKey(input string) string {
	value := trimUnicodeWhitespace(norm.NFC.String(input))
	var b strings.Builder
	for _, r := range value {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

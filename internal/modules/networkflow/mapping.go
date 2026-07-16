package networkflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	generatedmapping "github.com/JochiRaider/cartulary/internal/gen/networkflowmapping"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	TargetKindNetworkFlowTable = "network_flow_table"
	TargetTableSchemaID        = "cartulary.network_flow_table.v1"

	MappingCandidateSchemaID = "cartulary.network_flow.mapping_candidate.v1"

	UnknownColumnPolicyPreserve = "preserve_unmapped_raw"
	UnknownColumnPolicyReject   = "reject_unmapped_columns"
	UnknownColumnPolicyIgnore   = "ignore_unmapped_columns"

	MappingKindSourceColumn        = "source_column"
	MappingKindIgnoredSourceColumn = "ignored_source_column"
	MappingKindSystemDerivation    = "system_derivation"

	TransformTimestampProfile = "timestamp_profile_v1"
	TransformIPLiteral        = "ip_literal_v1"
	TransformPortNumber       = "port_number_v1"
	TransformProtocol         = "protocol_number_or_token_v1"
	TransformUint64Decimal    = "uint64_decimal_string_v1"
	TransformTrimASCIISpace   = "trim_ascii_space_v1"

	EmptyPolicyInvalid = "empty_string_is_invalid"
	EmptyPolicyNull    = "empty_string_is_null"

	FieldFlowStartUTC         = "network_flow.flow_start_utc"
	FieldFlowEndUTC           = "network_flow.flow_end_utc"
	FieldSrcIP                = "network_flow.src_ip"
	FieldDstIP                = "network_flow.dst_ip"
	FieldSrcPort              = "network_flow.src_port"
	FieldDstPort              = "network_flow.dst_port"
	FieldIPProtocol           = "network_flow.ip_protocol"
	FieldBytesCount           = "network_flow.bytes_count"
	FieldPacketsCount         = "network_flow.packets_count"
	FieldExporterID           = "network_flow.exporter_id"
	FieldInputInterface       = "network_flow.input_interface"
	FieldOutputInterface      = "network_flow.output_interface"
	FieldTCPFlags             = "network_flow.tcp_flags"
	FieldApplicationLabel     = "network_flow.application_label"
	FieldObservationSourceRef = "network_flow.observation_source_ref"
)

type TimestampProfile struct {
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

type SourceColumnDescriptor struct {
	SourceColumnOrdinal           int          `json:"source_column_ordinal"`
	RawHeaderText                 string       `json:"raw_header_text"`
	NormalizedHeaderForSuggestion string       `json:"normalized_header_for_suggestion"`
	RawHeaderSHA256               string       `json:"raw_header_sha256"`
	SampleValues                  []SafeSample `json:"sample_values"`
	DetectedEmptyCount            int          `json:"detected_empty_count"`
}

type SafeSample struct {
	SafeSample     *string `json:"safe_sample"`
	RawValueSHA256 *string `json:"raw_value_sha256"`
}

type FieldMapping struct {
	MappingKind         string `json:"mapping_kind"`
	FieldKey            string `json:"field_key,omitempty"`
	SourceColumnOrdinal int    `json:"source_column_ordinal,omitempty"`
	TransformID         string `json:"transform_id,omitempty"`
	EmptyValuePolicy    string `json:"empty_value_policy,omitempty"`
	Combinability       string `json:"combinability,omitempty"`
	IgnoreReason        string `json:"ignore_reason,omitempty"`
	DerivationID        string `json:"derivation_id,omitempty"`
}

type MappingCandidate struct {
	TargetKind          string           `json:"target_kind"`
	TargetTableSchemaID string           `json:"target_table_schema_id"`
	SourceProfileID     string           `json:"source_profile_id"`
	ParserProfileID     string           `json:"parser_profile_id,omitempty"`
	UnknownColumnPolicy string           `json:"unknown_column_policy,omitempty"`
	DisplayNameOverride *string          `json:"display_name_override,omitempty"`
	TimestampProfile    TimestampProfile `json:"timestamp_profile"`
	FieldMappings       []FieldMapping   `json:"field_mappings"`
}

type ApprovedMapping struct {
	TargetKind          string                   `json:"target_kind"`
	TargetTableSchemaID string                   `json:"target_table_schema_id"`
	SourceProfileID     string                   `json:"source_profile_id"`
	ParserProfileID     string                   `json:"parser_profile_id"`
	UnknownColumnPolicy string                   `json:"unknown_column_policy"`
	DisplayNameOverride *string                  `json:"display_name_override,omitempty"`
	TimestampProfile    TimestampProfile         `json:"timestamp_profile"`
	SourceColumns       []SourceColumnDescriptor `json:"source_columns"`
	FieldMappings       []FieldMapping           `json:"field_mappings"`
}

func MaterializeApprovedMapping(raw json.RawMessage, sourceColumns []SourceColumnDescriptor) (ApprovedMapping, error) {
	candidate, err := decodeMappingCandidate(raw)
	if err != nil {
		return ApprovedMapping{}, err
	}
	candidate = materializeCandidateDefaults(candidate)
	approved := ApprovedMapping{
		TargetKind:          candidate.TargetKind,
		TargetTableSchemaID: candidate.TargetTableSchemaID,
		SourceProfileID:     candidate.SourceProfileID,
		ParserProfileID:     candidate.ParserProfileID,
		UnknownColumnPolicy: candidate.UnknownColumnPolicy,
		DisplayNameOverride: candidate.DisplayNameOverride,
		TimestampProfile:    candidate.TimestampProfile,
		SourceColumns:       append([]SourceColumnDescriptor(nil), sourceColumns...),
		FieldMappings:       materializeFieldMappings(candidate.FieldMappings),
	}
	if err := validateApprovedMapping(approved); err != nil {
		return ApprovedMapping{}, err
	}
	return approved, nil
}

func DecodeApprovedMapping(raw json.RawMessage) (ApprovedMapping, error) {
	if err := validateTimestampProfileJSONShape(raw, true); err != nil {
		return ApprovedMapping{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var mapping ApprovedMapping
	if err := decoder.Decode(&mapping); err != nil {
		return ApprovedMapping{}, &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "variant_member_conflict"}
	}
	mapping = materializeApprovedDefaults(mapping)
	if err := validateApprovedMapping(mapping); err != nil {
		return ApprovedMapping{}, err
	}
	return mapping, nil
}

func MarshalApprovedMapping(mapping ApprovedMapping) json.RawMessage {
	data, err := json.Marshal(mapping)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

func decodeMappingCandidate(raw json.RawMessage) (MappingCandidate, error) {
	if err := validateTimestampProfileJSONShape(raw, false); err != nil {
		return MappingCandidate{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var candidate MappingCandidate
	if err := decoder.Decode(&candidate); err != nil {
		return MappingCandidate{}, &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "variant_member_conflict"}
	}
	return candidate, nil
}

func validateTimestampProfileJSONShape(raw json.RawMessage, approved bool) error {
	conflict := func() error {
		return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "variant_member_conflict"}
	}
	top, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(raw))
	if err != nil {
		return conflict()
	}
	profileRaw, ok := top["timestamp_profile"]
	if !ok || bytes.Equal(profileRaw, []byte("null")) {
		return conflict()
	}
	profile, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(profileRaw))
	if err != nil {
		return conflict()
	}
	stringMember := func(name string) (string, bool) {
		value, exists := profile[name]
		if !exists || bytes.Equal(value, []byte("null")) {
			return "", false
		}
		var decoded string
		if json.Unmarshal(value, &decoded) != nil {
			return "", false
		}
		return decoded, true
	}
	schemaID, schemaOK := stringMember("schema_id")
	mode, modeOK := stringMember("mode")
	if !schemaOK || schemaID != timestampProfileSchemaID || !modeOK {
		return conflict()
	}
	allowed := map[string]struct{}{"schema_id": {}, "mode": {}, "precision": {}}
	required := map[string]struct{}{"schema_id": {}, "mode": {}}
	if approved {
		required["precision"] = struct{}{}
	}
	switch mode {
	case "rfc3339":
		for _, member := range []string{"timezone", "timezone_ruleset_id", "ambiguous_local_time_policy", "local_time_gap_policy"} {
			allowed[member] = struct{}{}
			required[member] = struct{}{}
		}
	case "epoch_seconds", "epoch_milliseconds":
	case "netflow_sys_uptime_milliseconds":
		for _, member := range []string{"netflow_export_time_column_ordinal", "netflow_export_time_mode", "netflow_exporter_uptime_at_export_column_ordinal"} {
			allowed[member] = struct{}{}
			required[member] = struct{}{}
		}
	default:
		return conflict()
	}
	for member, value := range profile {
		if _, ok := allowed[member]; !ok {
			return conflict()
		}
		if bytes.Equal(value, []byte("null")) && !(mode == "rfc3339" && (member == "timezone" || member == "timezone_ruleset_id")) {
			return conflict()
		}
	}
	for member := range required {
		if _, ok := profile[member]; !ok {
			return conflict()
		}
	}
	return nil
}

func materializeCandidateDefaults(candidate MappingCandidate) MappingCandidate {
	profile := defaultMappingSourceProfile()
	if candidate.TargetKind == "" {
		candidate.TargetKind = generatedmapping.Registry.TargetKind
	}
	if candidate.TargetTableSchemaID == "" {
		candidate.TargetTableSchemaID = generatedmapping.Registry.TargetTableSchemaID
	}
	if candidate.SourceProfileID == "" {
		candidate.SourceProfileID = profile.SourceProfileID
	}
	if candidate.ParserProfileID == "" {
		candidate.ParserProfileID = profile.ParserProfileID
	}
	if candidate.UnknownColumnPolicy == "" {
		candidate.UnknownColumnPolicy = profile.DefaultUnknownColumnPolicy
	}
	candidate.TimestampProfile = materializeTimestampProfile(candidate.TimestampProfile)
	return candidate
}

func materializeApprovedDefaults(mapping ApprovedMapping) ApprovedMapping {
	mapping.TimestampProfile = materializeTimestampProfile(mapping.TimestampProfile)
	return mapping
}

func materializeTimestampProfile(profile TimestampProfile) TimestampProfile {
	defaults := defaultMappingSourceProfile().DefaultTimestampProfile
	if profile.SchemaID == "" {
		profile.SchemaID = defaults.SchemaID
	}
	if profile.Mode == "" {
		profile.Mode = defaults.Mode
	}
	if profile.Precision == "" {
		switch profile.Mode {
		case "epoch_seconds":
			profile.Precision = "seconds"
		case "epoch_milliseconds", "netflow_sys_uptime_milliseconds":
			profile.Precision = "milliseconds"
		default:
			profile.Precision = "microseconds"
		}
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

func materializeFieldMappings(input []FieldMapping) []FieldMapping {
	mappings := make([]FieldMapping, 0, len(input)+1)
	hasDerivation := false
	for _, mapping := range input {
		materialized := materializeFieldMapping(mapping)
		if materialized.MappingKind == MappingKindSystemDerivation && materialized.FieldKey == FieldObservationSourceRef {
			hasDerivation = true
		}
		mappings = append(mappings, materialized)
	}
	if !hasDerivation {
		derivation := observationSourceDerivation()
		mappings = append(mappings, FieldMapping{
			MappingKind:   MappingKindSystemDerivation,
			FieldKey:      derivation.FieldKey,
			DerivationID:  derivation.DerivationID,
			Combinability: derivation.Combinability,
		})
	}
	return mappings
}

func materializeFieldMapping(mapping FieldMapping) FieldMapping {
	switch mapping.MappingKind {
	case MappingKindSourceColumn:
		if mapping.Combinability == "" {
			mapping.Combinability = "single_source_only"
		}
		if mapping.TransformID == "" {
			mapping.TransformID = defaultTransformForField(mapping.FieldKey)
		}
		if mapping.EmptyValuePolicy == "" || mapping.EmptyValuePolicy == "profile_default" {
			mapping.EmptyValuePolicy = defaultEmptyPolicyForField(mapping.FieldKey)
		}
	case MappingKindIgnoredSourceColumn:
		if mapping.IgnoreReason == "" {
			mapping.IgnoreReason = "user_ignored"
		}
	case MappingKindSystemDerivation:
		derivation := observationSourceDerivation()
		if mapping.FieldKey == "" {
			mapping.FieldKey = derivation.FieldKey
		}
		if mapping.DerivationID == "" {
			mapping.DerivationID = derivation.DerivationID
		}
		if mapping.Combinability == "" {
			mapping.Combinability = derivation.Combinability
		}
	}
	return mapping
}

func validateApprovedMapping(mapping ApprovedMapping) error {
	if mapping.TargetKind != generatedmapping.Registry.TargetKind || mapping.TargetTableSchemaID != generatedmapping.Registry.TargetTableSchemaID {
		return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "variant_member_conflict"}
	}
	profile, ok := mappingSourceProfile(mapping.SourceProfileID)
	if !ok || profile.ConformanceStatus != "required_v1" {
		return &MappingValidationError{Code: "network_flow_unsupported_source_profile", ReasonCode: "unsupported_source_profile"}
	}
	if mapping.ParserProfileID != profile.ParserProfileID {
		return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "mapping_kind_unavailable"}
	}
	if !slicesContains(profile.SupportedUnknownColumnPolicies, mapping.UnknownColumnPolicy) {
		return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
	}
	if len(mapping.SourceColumns) == 0 {
		return &MappingValidationError{Code: "network_flow_invalid_header", ReasonCode: "empty_header"}
	}
	sourceOrdinals := map[int]struct{}{}
	for index, column := range mapping.SourceColumns {
		if column.SourceColumnOrdinal != index+1 {
			return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
		}
		sourceOrdinals[column.SourceColumnOrdinal] = struct{}{}
	}
	if err := validateTimestampProfile(mapping.TimestampProfile, len(mapping.SourceColumns)); err != nil {
		return err
	}
	byField := map[string]int{}
	byOrdinal := map[int]int{}
	hasSystemDerivation := false
	for _, fieldMapping := range mapping.FieldMappings {
		switch fieldMapping.MappingKind {
		case MappingKindSourceColumn:
			if _, ok := sourceOrdinals[fieldMapping.SourceColumnOrdinal]; !ok {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column", FieldKey: fieldMapping.FieldKey}
			}
			if !sourceMappableField(fieldMapping.FieldKey) {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "field_not_supported_by_profile", FieldKey: fieldMapping.FieldKey}
			}
			if fieldMapping.TransformID != defaultTransformForField(fieldMapping.FieldKey) {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "transform_target_mismatch", FieldKey: fieldMapping.FieldKey}
			}
			if fieldMapping.EmptyValuePolicy != EmptyPolicyInvalid && fieldMapping.EmptyValuePolicy != EmptyPolicyNull {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "invalid_empty_value_policy", FieldKey: fieldMapping.FieldKey}
			}
			byField[fieldMapping.FieldKey]++
			byOrdinal[fieldMapping.SourceColumnOrdinal]++
		case MappingKindIgnoredSourceColumn:
			if _, ok := sourceOrdinals[fieldMapping.SourceColumnOrdinal]; !ok {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
			}
			byOrdinal[fieldMapping.SourceColumnOrdinal]++
		case MappingKindSystemDerivation:
			derivation := observationSourceDerivation()
			if fieldMapping.FieldKey != derivation.FieldKey || fieldMapping.DerivationID != derivation.DerivationID || fieldMapping.Combinability != derivation.Combinability {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "system_derivation_missing"}
			}
			hasSystemDerivation = true
			byField[fieldMapping.FieldKey]++
		default:
			return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "mapping_kind_unavailable"}
		}
	}
	for _, fieldKey := range requiredCiscoFields() {
		if byField[fieldKey] == 0 {
			return &MappingValidationError{Code: "network_flow_mapping_required", ReasonCode: "required_field_unmapped", FieldKey: fieldKey}
		}
	}
	if !hasSystemDerivation {
		return &MappingValidationError{Code: "network_flow_mapping_required", ReasonCode: "system_derivation_missing", FieldKey: FieldObservationSourceRef}
	}
	for fieldKey, count := range byField {
		if count > 1 {
			return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "target_field_duplicated", FieldKey: fieldKey}
		}
	}
	for ordinal, count := range byOrdinal {
		if count > 1 {
			return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "source_column_reused", FieldKey: fmt.Sprint(ordinal)}
		}
	}
	if mapping.TimestampProfile.Mode == "netflow_sys_uptime_milliseconds" {
		exportOrdinal := *mapping.TimestampProfile.NetFlowExportTimeColumnOrdinal
		uptimeOrdinal := *mapping.TimestampProfile.NetFlowExporterUptimeAtExportColumnOrdinal
		if exportOrdinal == uptimeOrdinal {
			return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "timestamp_column_reused"}
		}
		for _, fieldKey := range []string{FieldFlowStartUTC, FieldFlowEndUTC} {
			ordinal := sourceFieldOrdinal(mapping.FieldMappings, fieldKey)
			if ordinal == exportOrdinal || ordinal == uptimeOrdinal {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "timestamp_column_reused", FieldKey: fieldKey}
			}
		}
	}
	if mapping.UnknownColumnPolicy == UnknownColumnPolicyIgnore || mapping.UnknownColumnPolicy == UnknownColumnPolicyReject {
		for _, column := range mapping.SourceColumns {
			if byOrdinal[column.SourceColumnOrdinal] == 0 {
				return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "unaccounted_source_column"}
			}
		}
	}
	return nil
}

func defaultTransformForField(fieldKey string) string {
	if field, ok := mappingRegistryField(fieldKey); ok {
		return field.TransformID
	}
	return ""
}

func defaultEmptyPolicyForField(fieldKey string) string {
	if field, ok := mappingRegistryField(fieldKey); ok {
		return field.EmptyValuePolicy
	}
	return ""
}

func sourceMappableField(fieldKey string) bool {
	field, ok := mappingRegistryField(fieldKey)
	return ok && (field.Requirement == "required" || field.Requirement == "optional_map_when_present")
}

func requiredCiscoFields() []string {
	fields := make([]string, 0)
	for _, field := range defaultMappingSourceProfile().Fields {
		if field.Requirement == "required" {
			fields = append(fields, field.FieldKey)
		}
	}
	return fields
}

func SuggestCiscoSNAMapping(sourceColumns []SourceColumnDescriptor) []FieldMapping {
	used := map[int]struct{}{}
	mappings := []FieldMapping{}
	for _, fieldKey := range append(requiredCiscoFields(), FieldInputInterface, FieldOutputInterface) {
		ordinal := firstAliasOrdinal(fieldKey, sourceColumns, used)
		if ordinal == 0 {
			continue
		}
		used[ordinal] = struct{}{}
		mappings = append(mappings, materializeFieldMapping(FieldMapping{
			MappingKind:         MappingKindSourceColumn,
			FieldKey:            fieldKey,
			SourceColumnOrdinal: ordinal,
		}))
	}
	sort.SliceStable(mappings, func(i, j int) bool {
		return mappingSortKey(mappings[i]) < mappingSortKey(mappings[j])
	})
	return mappings
}

func firstAliasOrdinal(fieldKey string, sourceColumns []SourceColumnDescriptor, used map[int]struct{}) int {
	field, ok := mappingRegistryField(fieldKey)
	if !ok {
		return 0
	}
	for _, column := range sourceColumns {
		if _, ok := used[column.SourceColumnOrdinal]; ok {
			continue
		}
		for _, alias := range field.Aliases {
			if column.NormalizedHeaderForSuggestion == SourceAliasMatchKey(alias) {
				return column.SourceColumnOrdinal
			}
		}
	}
	return 0
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

func mappingRegistryField(fieldKey string) (generatedmapping.Field, bool) {
	for _, field := range defaultMappingSourceProfile().Fields {
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

func SourceAliasMatchKey(input string) string {
	value := trimUnicodeWhitespace(input)
	var b strings.Builder
	for _, r := range value {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

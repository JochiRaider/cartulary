package networkflow

import (
	"bytes"
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

// Nested objects remain raw until their own closed variant has been admitted.
// The typed approved representation and canonical serializer do not own admission.
type mappingJSON struct {
	TargetKind          string          `json:"target_kind"`
	TargetTableSchemaID string          `json:"target_table_schema_id"`
	SourceProfileID     string          `json:"source_profile_id"`
	ParserProfileID     string          `json:"parser_profile_id"`
	UnknownColumnPolicy string          `json:"unknown_column_policy"`
	DisplayNameOverride *string         `json:"display_name_override"`
	TimestampProfile    json.RawMessage `json:"timestamp_profile"`
	SourceColumns       json.RawMessage `json:"source_columns"`
	FieldMappings       json.RawMessage `json:"field_mappings"`
}

func mappingShapeConflict() error {
	return &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "variant_member_conflict"}
}

func decodeMappingCandidate(raw json.RawMessage) (mappingCandidate, error) {
	var wire mappingJSON
	object, err := decodeClosedImportObject(raw,
		"target_kind target_table_schema_id source_profile_id timestamp_profile field_mappings",
		"parser_profile_id unknown_column_policy display_name_override", "", &wire)
	if err != nil {
		return mappingCandidate{}, mappingShapeConflict()
	}
	profile, err := decodeMappingTimestampProfile(wire.TimestampProfile, false)
	if err != nil {
		return mappingCandidate{}, err
	}
	fields, err := decodeMappingFieldRows(wire.FieldMappings, false)
	if err != nil {
		return mappingCandidate{}, err
	}
	// Only absence requests these defaults. Empty or null input cannot be repaired.
	if _, exists := object["parser_profile_id"]; !exists {
		wire.ParserProfileID = parserProfileRFC4180HeaderedCSV
	}
	if _, exists := object["unknown_column_policy"]; !exists {
		wire.UnknownColumnPolicy = unknownColumnPolicyPreserve
	}
	return mappingCandidate{
		TargetKind: wire.TargetKind, TargetTableSchemaID: wire.TargetTableSchemaID,
		SourceProfileID: wire.SourceProfileID, ParserProfileID: wire.ParserProfileID,
		UnknownColumnPolicy: wire.UnknownColumnPolicy, DisplayNameOverride: wire.DisplayNameOverride,
		TimestampProfile: profile, FieldMappings: fields,
	}, nil
}

// Stored approvals and both preview copies enter through this reader. No default
// materialization is permitted here, even when a fingerprint still matches.
func decodeApprovedMapping(raw json.RawMessage, limits EffectiveLimits) (approvedMapping, error) {
	var wire mappingJSON
	if _, err := decodeClosedImportObject(raw,
		"target_kind target_table_schema_id source_profile_id parser_profile_id unknown_column_policy timestamp_profile source_columns field_mappings",
		"display_name_override", "", &wire); err != nil {
		return approvedMapping{}, mappingShapeConflict()
	}
	profile, err := decodeMappingTimestampProfile(wire.TimestampProfile, true)
	if err != nil {
		return approvedMapping{}, err
	}
	fields, err := decodeApprovedFieldMappings(wire.FieldMappings)
	if err != nil {
		return approvedMapping{}, err
	}
	columns, err := decodeSourceColumns(wire.SourceColumns, limits)
	if err != nil {
		return approvedMapping{}, mappingShapeConflict()
	}
	mapping := approvedMapping{
		TargetKind: wire.TargetKind, TargetTableSchemaID: wire.TargetTableSchemaID,
		SourceProfileID: wire.SourceProfileID, ParserProfileID: wire.ParserProfileID,
		UnknownColumnPolicy: wire.UnknownColumnPolicy, DisplayNameOverride: wire.DisplayNameOverride,
		TimestampProfile: profile, SourceColumns: columns, FieldMappings: fields,
	}
	if err := validateApprovedMapping(mapping); err != nil {
		return approvedMapping{}, err
	}
	if mapping.DisplayNameOverride != nil {
		normalized, err := normalizeExplicitDisplayName(*mapping.DisplayNameOverride)
		if err != nil || normalized != *mapping.DisplayNameOverride {
			return approvedMapping{}, mappingShapeConflict()
		}
	}
	return mapping, nil
}

func decodeMappingFieldRows(raw json.RawMessage, approved bool) ([]candidateFieldMapping, error) {
	var rows []json.RawMessage
	if json.Unmarshal(raw, &rows) != nil || rows == nil {
		return nil, mappingShapeConflict()
	}
	fields := make([]candidateFieldMapping, 0, len(rows))
	for _, row := range rows {
		object, err := strictjson.DecodeUnicodeObject(bytes.NewReader(row))
		if err != nil {
			return nil, mappingShapeConflict()
		}
		var kind *string
		if json.Unmarshal(object["mapping_kind"], &kind) != nil || kind == nil {
			return nil, mappingShapeConflict()
		}
		var required, optional, nullable string
		switch *kind {
		case mappingKindSourceColumn:
			required = "mapping_kind field_key source_column_ordinal transform_id combinability"
			if approved {
				required += " empty_value_policy"
			} else {
				optional = "empty_value_policy"
				// Retain every explicit policy value for the field-specific policy
				// error, after exact names and variant membership have been checked.
				nullable = "empty_value_policy"
			}
		case mappingKindIgnoredSourceColumn:
			required = "mapping_kind source_column_ordinal ignore_reason"
		case mappingKindSystemDerivation:
			required = "mapping_kind field_key derivation_id combinability"
		default:
			return nil, &mappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: "mapping_kind_unavailable"}
		}
		var field candidateFieldMapping
		if _, err := decodeClosedImportObject(row, required, optional, nullable, &field); err != nil {
			return nil, mappingShapeConflict()
		}
		if field.MappingKind != mappingKindSystemDerivation && field.SourceColumnOrdinal < 1 {
			return nil, mappingShapeConflict()
		}
		if field.MappingKind == mappingKindIgnoredSourceColumn {
			if field.IgnoreReason != "user_ignored" && field.IgnoreReason != "policy_accounting" {
				return nil, mappingShapeConflict()
			}
		} else if field.Combinability != "single_source_only" {
			return nil, mappingShapeConflict()
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func decodeApprovedFieldMappings(raw json.RawMessage) ([]fieldMapping, error) {
	rows, err := decodeMappingFieldRows(raw, true)
	if err != nil {
		return nil, err
	}
	fields := make([]fieldMapping, 0, len(rows))
	for _, row := range rows {
		field := row.withPolicy("")
		if row.MappingKind == mappingKindSourceColumn {
			if json.Unmarshal(row.EmptyValuePolicy, &field.EmptyValuePolicy) != nil {
				return nil, mappingShapeConflict()
			}
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func decodeMappingTimestampProfile(raw json.RawMessage, approved bool) (timestampProfile, error) {
	object, err := strictjson.DecodeUnicodeObject(bytes.NewReader(raw))
	if err != nil {
		return timestampProfile{}, mappingShapeConflict()
	}
	var mode string
	if json.Unmarshal(object["mode"], &mode) != nil {
		return timestampProfile{}, mappingShapeConflict()
	}
	required, optional, nullable := "schema_id mode", "precision", ""
	if approved {
		required += " precision"
		optional = ""
	}
	switch mode {
	case "rfc3339":
		required += " timezone timezone_ruleset_id ambiguous_local_time_policy local_time_gap_policy"
		nullable = "timezone timezone_ruleset_id"
	case "epoch_seconds", "epoch_milliseconds", "cisco_sna_datetime_v1":
	case "netflow_sys_uptime_milliseconds":
		required += " netflow_export_time_column_ordinal netflow_export_time_mode netflow_exporter_uptime_at_export_column_ordinal"
	default:
		return timestampProfile{}, mappingShapeConflict()
	}
	var profile timestampProfile
	if _, err := decodeClosedImportObject(raw, required, optional, nullable, &profile); err != nil {
		return timestampProfile{}, mappingShapeConflict()
	}
	if _, exists := object["precision"]; !exists && !approved {
		profile.Precision = defaultTimestampPrecision(mode)
	}
	return profile, nil
}

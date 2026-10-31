package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
)

func safeNumericSample(value string) bool {
	if len(value) < 1 || len(value) > 32 {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func numericDiagnosticField(field string) bool {
	switch field {
	case fieldSrcPort, fieldDstPort, fieldIPProtocol, fieldBytesCount, fieldPacketsCount:
		return true
	default:
		return false
	}
}

func validDiagnosticReason(code, reason string) bool {
	var allowed []string
	switch code {
	case "network_flow_csv_field_count_mismatch":
		allowed = []string{"field_count_mismatch"}
	case "network_flow_invalid_timestamp":
		allowed = []string{"missing_or_empty", "invalid_syntax", "precision_exceeded", "out_of_range", "ambiguous_local_time", "nonexistent_local_time", "sys_uptime_invalid", "sys_uptime_wrap_ambiguous"}
	case "network_flow_end_before_start":
		allowed = []string{"end_before_start"}
	case "network_flow_invalid_ip":
		allowed = []string{"missing_or_empty", "invalid_syntax", "noncanonical_value"}
	case "network_flow_invalid_port", "network_flow_invalid_counter":
		allowed = []string{"missing_or_empty", "invalid_syntax", "out_of_range"}
	case "network_flow_invalid_text":
		allowed = []string{"missing_or_empty", "out_of_range"}
	case "network_flow_invalid_protocol":
		allowed = []string{"missing_or_empty", "invalid_syntax", "unknown_token", "out_of_range"}
	case "network_flow_resource_limit_exceeded":
		allowed = []string{"cell_scalar_limit_exceeded"}
	}
	return slices.Contains(allowed, reason)
}

// validateDiagnostic checks owner values without interpreting or exposing source bytes.
// Both preview admission and retained-state admission use this boundary.
func validateDiagnostic(d rejectedRowDiagnostic, sourceSHA256, fingerprint string) error {
	if !validLowerSHA256(sourceSHA256) || !validLowerSHA256(fingerprint) || d.DiagnosticID != diagnosticID(sourceSHA256, fingerprint, d.SourceRowNumber, d.SourceColumnOrdinal, d.FieldKey, d.ErrorCode, d.ReasonCode) {
		return errors.New("network flow diagnostic identity failed semantic validation")
	}
	return validateDiagnosticResource(d)
}

// The owner error callback has no source context. Its resource validation shares
// all shape/privacy rules; preview and retained-state admission additionally bind identity.
func validateDiagnosticResource(d rejectedRowDiagnostic) error {
	invalid := errors.New("network flow diagnostic failed semantic validation")
	if !strings.HasPrefix(d.DiagnosticID, "nfd_") || !validLowerSHA256(strings.TrimPrefix(d.DiagnosticID, "nfd_")) || d.SourceRowNumber < 2 || !validDiagnosticReason(d.ErrorCode, d.ReasonCode) {
		return invalid
	}
	message := "network_flow.diagnostic." + d.ErrorCode + "." + d.ReasonCode
	var args map[string]json.RawMessage
	if d.MessageKey != message || d.Message != message || json.Unmarshal(d.MessageArgs, &args) != nil || args == nil || len(args) != 0 {
		return invalid
	}
	if d.SourceColumnOrdinal == nil {
		if d.RawHeaderSHA256 != nil || d.RawValueSHA256 != nil || d.SafeSample != nil {
			return invalid
		}
	} else if *d.SourceColumnOrdinal < 1 || d.RawHeaderSHA256 == nil || !validLowerSHA256(*d.RawHeaderSHA256) || d.RawValueSHA256 == nil || !validLowerSHA256(*d.RawValueSHA256) {
		return invalid
	}
	if d.FieldKey != nil {
		if _, ok := mappingRegistryField(*d.FieldKey); !ok {
			return invalid
		}
	}
	if d.ErrorCode == "network_flow_csv_field_count_mismatch" && (d.SourceColumnOrdinal != nil || d.FieldKey != nil) {
		return invalid
	}
	if d.ErrorCode != "network_flow_csv_field_count_mismatch" && d.ErrorCode != "network_flow_resource_limit_exceeded" {
		if d.FieldKey == nil || d.SourceColumnOrdinal == nil {
			return invalid
		}
		expected := errorCodeForField(*d.FieldKey)
		if d.ErrorCode == "network_flow_end_before_start" {
			if *d.FieldKey != fieldFlowEndUTC {
				return invalid
			}
		} else if expected != d.ErrorCode {
			return invalid
		}
	}
	if d.SafeSample != nil && (d.FieldKey == nil || !numericDiagnosticField(*d.FieldKey) || !safeNumericSample(*d.SafeSample) || d.RawValueSHA256 == nil || sha256Hex([]byte(*d.SafeSample)) != *d.RawValueSHA256) {
		return invalid
	}
	if d.ErrorCode == "network_flow_resource_limit_exceeded" {
		if d.SafeSample != nil || d.SourceColumnOrdinal == nil {
			return invalid
		}
		if d.LimitName == nil || d.LimitValue == nil || d.ActualValue == nil || *d.LimitValue < 1 {
			return invalid
		}
		if *d.LimitName != "network_flow.max_raw_cell_scalar_length" || *d.ActualValue <= *d.LimitValue {
			return invalid
		}
	} else if d.LimitName != nil || d.LimitValue != nil || d.ActualValue != nil {
		return invalid
	}
	if bytes.Equal(bytes.TrimSpace(d.MessageArgs), []byte("null")) {
		return invalid
	}
	return nil
}

func decodeDiagnosticResource(raw json.RawMessage) (rejectedRowDiagnostic, error) {
	var d rejectedRowDiagnostic
	_, err := decodeClosedImportObject(raw, "diagnostic_id source_row_number source_column_ordinal raw_header_sha256 field_key error_code reason_code safe_sample raw_value_sha256 message_key message_args message limit_name limit_value actual_value", "", "source_column_ordinal raw_header_sha256 field_key safe_sample raw_value_sha256 limit_name limit_value actual_value", &d)
	if err == nil {
		err = validateDiagnosticResource(d)
	}
	return d, err
}

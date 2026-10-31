package networkflow

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

const (
	networkFlowImportOwnerErrorSchemaID = "cartulary.network_flow.import_owner_error.v1"
	networkFlowImportErrorTranslationID = "network_flow_activity.import_error_translation.v1"
)

type networkFlowImportOwnerError struct {
	ownerCode   string
	retryable   bool
	safeDetails map[string]any
}

func (e *networkFlowImportOwnerError) Error() string {
	return "network flow import owner validation failed"
}

func importOwnerError(ownerCode string, safeDetails map[string]any) error {
	if safeDetails == nil {
		safeDetails = map[string]any{}
	}
	return &networkFlowImportOwnerError{
		ownerCode:   ownerCode,
		retryable:   false,
		safeDetails: safeDetails,
	}
}

func (f *importFacade) TranslateImportUnitError(
	err error,
) (imports.ExtensionImportErrorTranslation, bool) {
	var ownerErr *networkFlowImportOwnerError
	if !errors.As(err, &ownerErr) {
		return imports.ExtensionImportErrorTranslation{}, false
	}
	coreReason := "owner_apply_validation_failed"
	switch ownerErr.ownerCode {
	case "network_flow_source_changed":
		coreReason = "source_changed"
	case "network_flow_target_unavailable":
		if ownerErr.safeDetails["reason_code"] == "owner_apply_contract_unavailable" {
			coreReason = "owner_apply_contract_unavailable"
		}
	}
	translated := imports.ExtensionImportErrorTranslation{
		ErrorSchemaID:      networkFlowImportOwnerErrorSchemaID,
		ErrorTranslationID: networkFlowImportErrorTranslationID,
		CoreReasonCode:     coreReason,
		OwnerError: imports.ExtensionImportOwnerError{
			SchemaID:    networkFlowImportOwnerErrorSchemaID,
			OwnerCode:   ownerErr.ownerCode,
			Retryable:   ownerErr.retryable,
			SafeDetails: cloneOwnerSafeDetails(ownerErr.safeDetails),
		},
	}
	if err := f.ValidateImportUnitError(translated.OwnerError); err != nil {
		return imports.ExtensionImportErrorTranslation{}, false
	}
	return translated, true
}

func (f *importFacade) ValidateImportUnitError(ownerErr imports.ExtensionImportOwnerError) error {
	if ownerErr.SchemaID != networkFlowImportOwnerErrorSchemaID ||
		ownerErr.SafeDetails == nil {
		return fmt.Errorf("network flow import owner error schema mismatch")
	}
	switch ownerErr.OwnerCode {
	case "network_flow_no_data_rows",
		"network_flow_source_changed",
		"network_flow_internal_failure":
		if len(ownerErr.SafeDetails) != 0 {
			return fmt.Errorf("network flow import owner error details must be empty")
		}
	case "network_flow_mapping_invalid":
		if !exactSafeDetailKeys(ownerErr.SafeDetails, "reason_code", "field") ||
			!validSafeToken(ownerErr.SafeDetails["reason_code"]) {
			return fmt.Errorf("network flow mapping owner error details are invalid")
		}
		if field, exists := ownerErr.SafeDetails["field"]; exists && !validSafeToken(field) {
			return fmt.Errorf("network flow mapping owner error field is invalid")
		}
	case "network_flow_invalid_display_name":
		if len(ownerErr.SafeDetails) != 3 || !exactSafeDetailKeys(ownerErr.SafeDetails, "reason_code", "max_length", "normalized_length") {
			return fmt.Errorf("network flow display-name owner error shape is invalid")
		}
		maximum, maxOK := ownerErr.SafeDetails["max_length"].(int64)
		length, lengthOK := ownerErr.SafeDetails["normalized_length"].(int64)
		if !maxOK || maximum != 64 || !lengthOK || length < 0 {
			return fmt.Errorf("network flow display-name owner error lengths are invalid")
		}
		switch ownerErr.SafeDetails["reason_code"] {
		case "empty_display_name", "forbidden_control", "display_name_too_long", "duplicate_display_name":
		default:
			return fmt.Errorf("network flow display-name owner error reason is invalid")
		}
	case "network_flow_target_unavailable":
		if !exactSafeDetailKeys(ownerErr.SafeDetails, "reason_code") ||
			!validSafeToken(ownerErr.SafeDetails["reason_code"]) {
			return fmt.Errorf("network flow target owner error details are invalid")
		}
	case "network_flow_all_rows_rejected":
		if err := validateAllRowsRejectedSafeDetails(ownerErr.SafeDetails); err != nil {
			return err
		}
	default:
		return fmt.Errorf("network flow import owner error code is not registered")
	}
	return nil
}

func validateAllRowsRejectedSafeDetails(details map[string]any) error {
	if !exactSafeDetailKeys(
		details,
		"row_count_rejected",
		"diagnostics_truncated",
		"diagnostics_sample",
	) {
		return fmt.Errorf("network flow rejected-row owner error shape is invalid")
	}
	rowCount, ok := details["row_count_rejected"].(int64)
	if !ok || rowCount < 1 {
		return fmt.Errorf("network flow rejected-row count is invalid")
	}
	truncated, ok := details["diagnostics_truncated"].(bool)
	if !ok {
		return fmt.Errorf("network flow rejected-row truncation flag is invalid")
	}
	diagnostics, ok := details["diagnostics_sample"].([]map[string]any)
	if !ok || len(diagnostics) == 0 || len(diagnostics) > 50 {
		return fmt.Errorf("network flow rejected-row diagnostics are invalid")
	}
	var previous rejectedRowDiagnostic
	rows := map[int64]struct{}{}
	for i, resource := range diagnostics {
		raw, err := json.Marshal(resource)
		if err != nil {
			return fmt.Errorf("network flow rejected-row diagnostic cannot be encoded")
		}
		diagnostic, err := decodeDiagnosticResource(raw)
		if err != nil || (i > 0 && compareDiagnostics(previous, diagnostic) >= 0) {
			return fmt.Errorf("network flow rejected-row diagnostic resource or order is invalid")
		}
		previous = diagnostic
		rows[diagnostic.SourceRowNumber] = struct{}{}
	}
	if int64(len(rows)) > rowCount || (!truncated && int64(len(rows)) != rowCount) {
		return fmt.Errorf("network flow rejected-row diagnostic count is invalid")
	}

	return nil
}

func exactSafeDetailKeys(details map[string]any, requiredAndOptional ...string) bool {
	allowed := make(map[string]struct{}, len(requiredAndOptional))
	for _, key := range requiredAndOptional {
		allowed[key] = struct{}{}
	}
	for key := range details {
		if _, ok := allowed[key]; !ok {
			return false
		}
	}
	return true
}

func validSafeToken(value any) bool {
	token, ok := value.(string)
	return ok && token != "" && len(token) <= 128
}

func cloneOwnerSafeDetails(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

const allRejectedDiagnosticSampleLimit = 50

func allRowsRejectedOwnerError(
	rejectedCount int64,
	diagnostics []rejectedRowDiagnostic,
	diagnosticsTruncated bool,
) error {
	sampleCount := len(diagnostics)
	if sampleCount > allRejectedDiagnosticSampleLimit {
		sampleCount = allRejectedDiagnosticSampleLimit
	}
	sample := make([]map[string]any, 0, sampleCount)
	for _, diagnostic := range diagnostics[:sampleCount] {
		sample = append(sample, diagnosticResource(diagnostic))
	}
	return importOwnerError("network_flow_all_rows_rejected", map[string]any{
		"row_count_rejected":    rejectedCount,
		"diagnostics_truncated": diagnosticsTruncated || len(diagnostics) > sampleCount,
		"diagnostics_sample":    sample,
	})
}

// Preview and apply preserve the same typed naming failure without raw input.
func importDisplayNameError(invalid *invalidDisplayNameError) error {
	return importOwnerError("network_flow_invalid_display_name", map[string]any{
		"reason_code":       invalid.ReasonCode,
		"max_length":        int64(64),
		"normalized_length": int64(invalid.NormalizedLength),
	})
}

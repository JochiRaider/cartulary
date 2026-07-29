package networkflow

import (
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
	rowCount, ok := details["row_count_rejected"].(int)
	if !ok || rowCount < 0 {
		return fmt.Errorf("network flow rejected-row count is invalid")
	}
	if _, ok := details["diagnostics_truncated"].(bool); !ok {
		return fmt.Errorf("network flow rejected-row truncation flag is invalid")
	}
	diagnostics, ok := details["diagnostics_sample"].([]map[string]any)
	if !ok || len(diagnostics) > 50 {
		return fmt.Errorf("network flow rejected-row diagnostics are invalid")
	}
	for _, diagnostic := range diagnostics {
		if !exactSafeDetailKeys(
			diagnostic,
			"source_row_number",
			"source_column_ordinal",
			"field_key",
			"error_code",
			"reason_code",
		) {
			return fmt.Errorf("network flow rejected-row diagnostic shape is invalid")
		}
		sourceRow, ok := diagnostic["source_row_number"].(int64)
		if !ok || sourceRow < 1 ||
			!validSafeToken(diagnostic["error_code"]) ||
			!validSafeToken(diagnostic["reason_code"]) {
			return fmt.Errorf("network flow rejected-row diagnostic value is invalid")
		}
		if ordinal, exists := diagnostic["source_column_ordinal"]; exists {
			value, ok := ordinal.(int64)
			if !ok || value < 1 {
				return fmt.Errorf("network flow rejected-row diagnostic ordinal is invalid")
			}
		}
		if field, exists := diagnostic["field_key"]; exists && !validSafeToken(field) {
			return fmt.Errorf("network flow rejected-row diagnostic field is invalid")
		}
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

func allRowsRejectedOwnerError(
	diagnostics []RejectedRowDiagnostic,
	diagnosticsTruncated bool,
) error {
	sampleCount := len(diagnostics)
	if sampleCount > 50 {
		sampleCount = 50
	}
	sample := make([]map[string]any, 0, sampleCount)
	for _, diagnostic := range diagnostics[:sampleCount] {
		item := map[string]any{
			"source_row_number": diagnostic.SourceRowNumber,
			"error_code":        diagnostic.ErrorCode,
			"reason_code":       diagnostic.ReasonCode,
		}
		if diagnostic.SourceColumnOrdinal != nil {
			item["source_column_ordinal"] = *diagnostic.SourceColumnOrdinal
		}
		if diagnostic.FieldKey != nil {
			item["field_key"] = *diagnostic.FieldKey
		}
		sample = append(sample, item)
	}
	return importOwnerError("network_flow_all_rows_rejected", map[string]any{
		"row_count_rejected":    len(diagnostics),
		"diagnostics_truncated": diagnosticsTruncated || len(diagnostics) > sampleCount,
		"diagnostics_sample":    sample,
	})
}

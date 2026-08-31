package imports

import (
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

type importUnitFailureDetail struct {
	ErrorCode  string
	ReasonCode string
	Retryable  bool
	Details    map[string]any
}

type translatedImportUnitError struct {
	failure importUnitFailureDetail
	cause   error
}

func (e *translatedImportUnitError) Error() string {
	return "import owner apply failed"
}

func (e *translatedImportUnitError) Unwrap() error {
	return e.cause
}

type ExtensionImportOwnerError struct {
	SchemaID    string         `json:"schema_id"`
	OwnerCode   string         `json:"owner_code"`
	Retryable   bool           `json:"retryable"`
	SafeDetails map[string]any `json:"safe_details"`
}

type ExtensionImportErrorTranslation struct {
	ErrorSchemaID      string
	ErrorTranslationID string
	CoreReasonCode     string
	OwnerError         ExtensionImportOwnerError
}

func importOwnerCreateFailure(err error) importUnitFailureDetail {
	failure := importUnitFailureDetail{
		ErrorCode:  "import_apply_blocked",
		ReasonCode: "owner_create_validation_failed",
		Retryable:  false,
		Details:    map[string]any{"reason_code": "owner_create_validation_failed"},
	}
	if ownerDetail, ok := ownerfacade.ImportOwnerCreateErrorDetail(err); ok {
		failure.Details["owner_error"] = ownerDetail
	}
	return failure
}

func genericOwnerApplyFailure() importUnitFailureDetail {
	return importUnitFailureDetail{
		ErrorCode:  "import_apply_blocked",
		ReasonCode: "owner_apply_validation_failed",
		Retryable:  false,
		Details:    map[string]any{"reason_code": "owner_apply_validation_failed"},
	}
}

func translateExtensionOwnerFailure(
	target importTarget,
	facade ExtensionImportFacade,
	err error,
) importUnitFailureDetail {
	if common, ok := commonImportApplyFailure(err); ok {
		return common
	}
	fallback := genericOwnerApplyFailure()
	if facade == nil {
		return fallback
	}
	translation, ok := facade.TranslateImportUnitError(err)
	if !ok ||
		target.ErrorSchemaID == "" ||
		target.ErrorTranslationID == "" ||
		translation.ErrorSchemaID != target.ErrorSchemaID ||
		translation.ErrorTranslationID != target.ErrorTranslationID ||
		translation.OwnerError.SchemaID != target.ErrorSchemaID ||
		translation.OwnerError.OwnerCode == "" ||
		!validTranslatedCoreReason(translation.CoreReasonCode) {
		return fallback
	}
	if err := facade.ValidateImportUnitError(translation.OwnerError); err != nil {
		return fallback
	}
	ownerDetail, ok := cloneSafeOwnerDetail(translation.OwnerError)
	if !ok {
		return fallback
	}
	return importUnitFailureDetail{
		ErrorCode:  "import_apply_blocked",
		ReasonCode: translation.CoreReasonCode,
		Retryable:  translation.OwnerError.Retryable,
		Details: map[string]any{
			"reason_code": translation.CoreReasonCode,
			"owner_error": ownerDetail,
		},
	}
}

func commonImportApplyFailure(err error) (importUnitFailureDetail, bool) {
	var applyBlocked *applyBlockedError
	if !errors.As(err, &applyBlocked) || !validCommonImportApplyReason(applyBlocked.ReasonCode) {
		return importUnitFailureDetail{}, false
	}
	return importUnitFailureDetail{
		ErrorCode:  "import_apply_blocked",
		ReasonCode: applyBlocked.ReasonCode,
		Retryable:  false,
		Details:    map[string]any{"reason_code": applyBlocked.ReasonCode},
	}, true
}

func validCommonImportApplyReason(reasonCode string) bool {
	switch reasonCode {
	case "overlapping_units",
		"duplicate_apply_blocked",
		"unit_not_ready",
		"target_view_schema_not_importable",
		"target_kind_not_importable",
		"owner_create_contract_unavailable",
		"owner_apply_contract_unavailable",
		"source_changed":
		return true
	default:
		return false
	}
}

func validTranslatedCoreReason(reasonCode string) bool {
	switch reasonCode {
	case "owner_apply_validation_failed",
		"owner_apply_contract_unavailable",
		"source_changed":
		return true
	default:
		return false
	}
}

func cloneSafeOwnerDetail(ownerError ExtensionImportOwnerError) (map[string]any, bool) {
	encoded, err := json.Marshal(ownerError)
	if err != nil {
		return nil, false
	}
	var cloned map[string]any
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		return nil, false
	}
	return cloned, true
}

func cloneStringAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

package entities

import (
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *httpapi.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &httpapi.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: details,
	}
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func requiredRoleDescription(roles ...string) string {
	if len(roles) == 0 {
		return ""
	}
	if len(roles) == 1 {
		return roles[0]
	}
	return strings.Join(roles, "|")
}

func mergePreconditionFailedError(err *merge.MergePreconditionError) *httpapi.APIError {
	details := map[string]any{}
	if err != nil {
		if err.ReasonCode != "" {
			details["reason_code"] = err.ReasonCode
		}
		for key, value := range err.Details {
			details[key] = value
		}
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "merge_precondition_failed",
		Message: "merge precondition failed",
		Details: details,
	}
}

func mergeRowVersionConflictError(err *merge.MergeRowVersionConflictError) *httpapi.APIError {
	details := map[string]any{}
	if err != nil {
		details["record_id"] = err.RecordID.String()
		details["scope"] = err.Scope
		details["base_row_version"] = err.BaseRowVersion
		details["current_row_version"] = err.CurrentRowVersion
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "row_version_conflict",
		Message: "row version conflict",
		Details: details,
	}
}

func recordLockedError(err *merge.MergeRecordLockedError) *httpapi.APIError {
	details := map[string]any{}
	if err != nil {
		details["record_id"] = err.RecordID.String()
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "record_locked",
		Message: "record locked",
		Details: details,
	}
}

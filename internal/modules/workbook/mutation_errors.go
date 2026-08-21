package workbook

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	return invalidMutationPayloadWithDetails(field, reasonCode, nil)
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *httpapi.APIError {
	details := map[string]any{"field": field, "reason_code": reasonCode}
	for key, value := range extra {
		details[key] = value
	}
	return &httpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_mutation_payload",
		Message: "invalid mutation payload", Details: details,
	}
}

func rowVersionConflictError(details map[string]any) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusConflict, Code: "row_version_conflict",
		Message: "row version conflict", Details: details,
	}
}
